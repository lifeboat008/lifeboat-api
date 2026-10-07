package api

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	ledger "github.com/lifeboat008/lifeboat-ledger"
	protocol "github.com/lifeboat008/lifeboat-protocol"
	"github.com/stellar/go-stellar-sdk/keypair"
	_ "modernc.org/sqlite"
)

type Actor struct {
	ID    string `json:"id"`
	Role  string `json:"role"`
	Token string `json:"token"`
}

type Server struct {
	db       *sql.DB
	actors   []Actor
	payments *ledger.Service
	now      func() time.Time
}

func NewServer(databasePath string, actors []Actor, payments *ledger.Service) (*Server, error) {
	if databasePath == "" || payments == nil || len(actors) == 0 {
		return nil, errors.New("database, actors, and payment service are required")
	}
	for _, actor := range actors {
		if actor.ID == "" || len(actor.Token) < 24 || !validRole(actor.Role) {
			return nil, errors.New("actor id, role, and a strong token are required")
		}
	}
	seenTokens := make(map[string]struct{}, len(actors))
	for _, actor := range actors {
		if _, duplicate := seenTokens[actor.Token]; duplicate {
			return nil, errors.New("actor tokens must be unique")
		}
		seenTokens[actor.Token] = struct{}{}
	}
	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	for _, statement := range []string{
		`PRAGMA foreign_keys=ON`, `PRAGMA busy_timeout=5000`,
		`CREATE TABLE IF NOT EXISTS projects (id TEXT PRIMARY KEY, body BLOB NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS plans (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, body BLOB NOT NULL, total INTEGER NOT NULL, reserved INTEGER NOT NULL DEFAULT 0, paid INTEGER NOT NULL DEFAULT 0)`,
		`CREATE TABLE IF NOT EXISTS evidence (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, delivery_id TEXT NOT NULL UNIQUE, body BLOB NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS claims (id TEXT PRIMARY KEY, plan_id TEXT NOT NULL, evidence_id TEXT NOT NULL UNIQUE, body BLOB NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS decisions (claim_id TEXT PRIMARY KEY, body BLOB NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS payments (claim_id TEXT PRIMARY KEY, idempotency_key TEXT NOT NULL UNIQUE, body BLOB NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS rescue_tasks (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, body BLOB NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS audit (seq INTEGER PRIMARY KEY AUTOINCREMENT, occurred_at TEXT NOT NULL, actor_id TEXT NOT NULL, action TEXT NOT NULL, subject_id TEXT NOT NULL)`,
	} {
		if _, err = db.Exec(statement); err != nil {
			db.Close()
			return nil, err
		}
	}
	return &Server{db: db, actors: actors, payments: payments, now: time.Now}, nil
}

func (s *Server) Close() error { return s.db.Close() }

func validRole(role string) bool {
	switch role {
	case "steward", "sponsor", "maintainer", "reviewer", "payer", "github":
		return true
	}
	return false
}

func (s *Server) hasActor(id, role string) bool {
	for _, actor := range s.actors {
		if actor.ID == id && actor.Role == role {
			return true
		}
	}
	return false
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /v1/projects", s.createProject)
	mux.HandleFunc("POST /v1/plans", s.createPlan)
	mux.HandleFunc("GET /v1/plans/{id}/budget", s.getBudget)
	mux.HandleFunc("POST /v1/evidence", s.createEvidence)
	mux.HandleFunc("POST /v1/claims", s.createClaim)
	mux.HandleFunc("GET /v1/claims/{id}", s.getClaim)
	mux.HandleFunc("POST /v1/claims/{id}/decision", s.decideClaim)
	mux.HandleFunc("POST /v1/claims/{id}/payment", s.payClaim)
	mux.HandleFunc("POST /v1/rescue-tasks", s.createRescueTask)
	mux.HandleFunc("GET /v1/audit/{id}", s.getAudit)
	return mux
}

func (s *Server) authorize(w http.ResponseWriter, r *http.Request, role string) (Actor, bool) {
	provided := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	for _, actor := range s.actors {
		if actor.Role == role && subtle.ConstantTimeCompare([]byte(provided), []byte(actor.Token)) == 1 {
			return actor, true
		}
	}
	problem(w, http.StatusUnauthorized, "valid bearer token for "+role+" is required")
	return Actor{}, false
}

func decode(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		problem(w, http.StatusBadRequest, "invalid JSON request")
		return false
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		problem(w, http.StatusBadRequest, "one JSON object is required")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func problem(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
func encode(value any) []byte { data, _ := json.Marshal(value); return data }
func audit(tx *sql.Tx, actor, action, subject string) error {
	_, err := tx.Exec(`INSERT INTO audit(occurred_at,actor_id,action,subject_id) VALUES(?,?,?,?)`, time.Now().UTC().Format(time.RFC3339Nano), actor, action, subject)
	return err
}

func (s *Server) insertAudited(ctx context.Context, actor, action, subject, statement string, args ...any) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(statement, args...); err != nil {
		return err
	}
	if err = audit(tx, actor, action, subject); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	actor, ok := s.authorize(w, r, "steward")
	if !ok {
		return
	}
	var project protocol.Project
	if !decode(w, r, &project) {
		return
	}
	if project.StewardID != actor.ID || project.State != protocol.ProjectActive || project.Validate() != nil {
		problem(w, 400, "invalid or unauthorized project")
		return
	}
	err := s.insertAudited(r.Context(), actor.ID, "project_created", project.ID, `INSERT INTO projects(id,body) VALUES(?,?)`, project.ID, encode(project))
	if err != nil {
		problem(w, 409, "project already exists or could not be stored")
		return
	}
	writeJSON(w, 201, project)
}

func (s *Server) createPlan(w http.ResponseWriter, r *http.Request) {
	actor, ok := s.authorize(w, r, "sponsor")
	if !ok {
		return
	}
	var plan protocol.Plan
	if !decode(w, r, &plan) {
		return
	}
	if plan.SponsorID != actor.ID || !s.hasActor(plan.ReviewerID, "reviewer") || plan.Validate(s.now()) != nil {
		problem(w, 400, "invalid or unauthorized plan")
		return
	}
	var exists int
	if err := s.db.QueryRow(`SELECT 1 FROM projects WHERE id=?`, plan.ProjectID).Scan(&exists); err != nil {
		problem(w, 404, "project not found")
		return
	}
	err := s.insertAudited(r.Context(), actor.ID, "plan_created", plan.ID, `INSERT INTO plans(id,project_id,body,total) VALUES(?,?,?,?)`, plan.ID, plan.ProjectID, encode(plan), plan.BudgetStroops)
	if err != nil {
		problem(w, 409, "plan already exists or could not be stored")
		return
	}
	writeJSON(w, 201, plan)
}

func (s *Server) getBudget(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.authorizeAny(w, r, "steward", "sponsor", "reviewer", "maintainer"); !ok {
		return
	}
	var budget protocol.Budget
	err := s.db.QueryRow(`SELECT total,reserved,paid FROM plans WHERE id=?`, r.PathValue("id")).Scan(&budget.TotalStroops, &budget.ReservedStroops, &budget.PaidStroops)
	if err != nil {
		problem(w, 404, "plan not found")
		return
	}
	writeJSON(w, 200, map[string]any{"budget": budget, "remaining_stroops": budget.Remaining()})
}

func (s *Server) authorizeAny(w http.ResponseWriter, r *http.Request, roles ...string) (Actor, bool) {
	provided := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	for _, actor := range s.actors {
		for _, role := range roles {
			if actor.Role == role && subtle.ConstantTimeCompare([]byte(provided), []byte(actor.Token)) == 1 {
				return actor, true
			}
		}
	}
	problem(w, 401, "valid bearer token is required")
	return Actor{}, false
}

func (s *Server) createEvidence(w http.ResponseWriter, r *http.Request) {
	actor, ok := s.authorize(w, r, "github")
	if !ok {
		return
	}
	var evidence protocol.Evidence
	if !decode(w, r, &evidence) {
		return
	}
	if evidence.Validate() != nil {
		problem(w, 400, "invalid evidence")
		return
	}
	var data []byte
	if err := s.db.QueryRow(`SELECT body FROM projects WHERE id=?`, evidence.ProjectID).Scan(&data); err != nil {
		problem(w, 404, "project not found")
		return
	}
	var project protocol.Project
	_ = json.Unmarshal(data, &project)
	if project.InstallationID != evidence.InstallationID {
		problem(w, 403, "installation does not match project")
		return
	}
	err := s.insertAudited(r.Context(), actor.ID, "evidence_recorded", evidence.ID, `INSERT INTO evidence(id,project_id,delivery_id,body) VALUES(?,?,?,?)`, evidence.ID, evidence.ProjectID, evidence.DeliveryID, encode(evidence))
	if err != nil {
		problem(w, 409, "duplicate evidence delivery or id")
		return
	}
	writeJSON(w, 201, evidence)
}

func (s *Server) createClaim(w http.ResponseWriter, r *http.Request) {
	actor, ok := s.authorize(w, r, "maintainer")
	if !ok {
		return
	}
	var claim protocol.Claim
	if !decode(w, r, &claim) {
		return
	}
	var planData, evidenceData []byte
	if s.db.QueryRow(`SELECT body FROM plans WHERE id=?`, claim.PlanID).Scan(&planData) != nil || s.db.QueryRow(`SELECT body FROM evidence WHERE id=?`, claim.EvidenceID).Scan(&evidenceData) != nil {
		problem(w, 404, "plan or evidence not found")
		return
	}
	var plan protocol.Plan
	var evidence protocol.Evidence
	_ = json.Unmarshal(planData, &plan)
	_ = json.Unmarshal(evidenceData, &evidence)
	_, addressErr := keypair.ParseAddress(claim.DestinationAccount)
	if claim.MaintainerID != actor.ID || claim.State != protocol.ClaimSubmitted || claim.Validate(plan, s.now()) != nil || evidence.ProjectID != claim.ProjectID || evidence.Kind != claim.WorkType || addressErr != nil {
		problem(w, 400, "invalid or unauthorized claim")
		return
	}
	err := s.insertAudited(r.Context(), actor.ID, "claim_submitted", claim.ID, `INSERT INTO claims(id,plan_id,evidence_id,body) VALUES(?,?,?,?)`, claim.ID, claim.PlanID, claim.EvidenceID, encode(claim))
	if err != nil {
		problem(w, 409, "claim already exists or could not be stored")
		return
	}
	writeJSON(w, 201, claim)
}

func (s *Server) getClaim(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.authorizeAny(w, r, "steward", "sponsor", "maintainer", "reviewer", "payer"); !ok {
		return
	}
	var data []byte
	if err := s.db.QueryRow(`SELECT body FROM claims WHERE id=?`, r.PathValue("id")).Scan(&data); err != nil {
		problem(w, 404, "claim not found")
		return
	}
	var claim protocol.Claim
	_ = json.Unmarshal(data, &claim)
	response := map[string]any{"claim": claim}
	var decisionData []byte
	if s.db.QueryRow(`SELECT body FROM decisions WHERE claim_id=?`, claim.ID).Scan(&decisionData) == nil {
		var decision protocol.Decision
		_ = json.Unmarshal(decisionData, &decision)
		response["decision"] = decision
	}
	var paymentData []byte
	if s.db.QueryRow(`SELECT body FROM payments WHERE claim_id=?`, claim.ID).Scan(&paymentData) == nil {
		var payment paymentRecord
		_ = json.Unmarshal(paymentData, &payment)
		response["transaction_hash"] = payment.Prepared.TransactionHash
		if payment.Receipt.Confirmed {
			response["receipt"] = payment.Receipt
		}
	}
	writeJSON(w, 200, response)
}

func (s *Server) decideClaim(w http.ResponseWriter, r *http.Request) {
	actor, ok := s.authorize(w, r, "reviewer")
	if !ok {
		return
	}
	var decision protocol.Decision
	if !decode(w, r, &decision) {
		return
	}
	if decision.ClaimID != r.PathValue("id") || decision.ReviewerID != actor.ID {
		problem(w, 403, "reviewer or claim mismatch")
		return
	}
	decision.DecidedAt = s.now().UTC()
	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil {
		problem(w, 500, "database unavailable")
		return
	}
	defer tx.Rollback()
	var claimData, planData []byte
	if tx.QueryRow(`SELECT body FROM claims WHERE id=?`, decision.ClaimID).Scan(&claimData) != nil {
		problem(w, 404, "claim not found")
		return
	}
	var claim protocol.Claim
	_ = json.Unmarshal(claimData, &claim)
	if tx.QueryRow(`SELECT body FROM plans WHERE id=?`, claim.PlanID).Scan(&planData) != nil {
		problem(w, 500, "plan not found")
		return
	}
	var plan protocol.Plan
	_ = json.Unmarshal(planData, &plan)
	if decision.Validate(claim, plan) != nil {
		problem(w, 409, "claim cannot be decided by this reviewer")
		return
	}
	if decision.Approved {
		result, err := tx.Exec(`UPDATE plans SET reserved=reserved+? WHERE id=? AND total-paid-reserved>=?`, claim.AmountStroops, plan.ID, claim.AmountStroops)
		if err != nil {
			problem(w, 500, "budget update failed")
			return
		}
		count, _ := result.RowsAffected()
		if count != 1 {
			problem(w, 409, "insufficient unreserved budget")
			return
		}
		claim.State = protocol.ClaimApproved
	} else {
		claim.State = protocol.ClaimRejected
	}
	if _, err = tx.Exec(`UPDATE claims SET body=? WHERE id=?`, encode(claim), claim.ID); err != nil {
		problem(w, 500, "claim update failed")
		return
	}
	if _, err = tx.Exec(`INSERT INTO decisions(claim_id,body) VALUES(?,?)`, claim.ID, encode(decision)); err != nil {
		problem(w, 409, "claim already decided")
		return
	}
	if err = audit(tx, actor.ID, "claim_decided", claim.ID); err != nil {
		problem(w, 500, "audit failed")
		return
	}
	if err = tx.Commit(); err != nil {
		problem(w, 500, "decision commit failed")
		return
	}
	writeJSON(w, 200, map[string]any{"claim": claim, "decision": decision})
}

type paymentRecord struct {
	Prepared ledger.Prepared `json:"prepared"`
	Receipt  ledger.Receipt  `json:"receipt"`
}

func (s *Server) payClaim(w http.ResponseWriter, r *http.Request) {
	actor, ok := s.authorize(w, r, "payer")
	if !ok {
		return
	}
	key := r.Header.Get("Idempotency-Key")
	if len(key) < 12 || len(key) > 128 {
		problem(w, 400, "Idempotency-Key must be 12-128 characters")
		return
	}
	claimID := r.PathValue("id")
	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil {
		problem(w, 500, "database unavailable")
		return
	}
	defer tx.Rollback()
	var claimData, recordData []byte
	var storedKey string
	if tx.QueryRow(`SELECT body FROM claims WHERE id=?`, claimID).Scan(&claimData) != nil {
		problem(w, 404, "claim not found")
		return
	}
	var claim protocol.Claim
	_ = json.Unmarshal(claimData, &claim)
	err = tx.QueryRow(`SELECT idempotency_key,body FROM payments WHERE claim_id=?`, claimID).Scan(&storedKey, &recordData)
	var record paymentRecord
	if err == nil {
		if key != storedKey {
			problem(w, 409, "claim already has another payment key")
			return
		}
		_ = json.Unmarshal(recordData, &record)
	} else if errors.Is(err, sql.ErrNoRows) {
		if claim.State != protocol.ClaimApproved {
			problem(w, 409, "claim is not approved")
			return
		}
		prepared, prepareErr := s.payments.Prepare(r.Context(), claim)
		if prepareErr != nil {
			problem(w, 502, "testnet payment preparation failed")
			return
		}
		record.Prepared = prepared
		if _, err = tx.Exec(`INSERT INTO payments(claim_id,idempotency_key,body) VALUES(?,?,?)`, claimID, key, encode(record)); err != nil {
			problem(w, 409, "payment already exists")
			return
		}
		claim.State = protocol.ClaimPaymentPending
		if _, err = tx.Exec(`UPDATE claims SET body=? WHERE id=?`, encode(claim), claimID); err != nil {
			problem(w, 500, "claim update failed")
			return
		}
		if err = audit(tx, actor.ID, "payment_prepared", claimID); err != nil {
			problem(w, 500, "audit failed")
			return
		}
	} else {
		problem(w, 500, "payment lookup failed")
		return
	}
	if err = tx.Commit(); err != nil {
		problem(w, 500, "payment commit failed")
		return
	}
	if claim.State == protocol.ClaimPaid {
		writeJSON(w, 200, record.Receipt)
		return
	}
	// Reconcile before resubmitting the exact signed envelope after any retry.
	receipt, err := s.payments.Lookup(r.Context(), record.Prepared.TransactionHash)
	if errors.Is(err, ledger.ErrNotFound) {
		receipt, err = s.payments.Submit(r.Context(), record.Prepared)
	}
	if err != nil {
		problem(w, 202, "submission state uncertain; retry with the same key")
		return
	}
	if !receipt.Confirmed {
		writeJSON(w, 202, receipt)
		return
	}
	if receipt.TransactionHash != record.Prepared.TransactionHash || receipt.Network != "testnet" {
		problem(w, 502, "receipt does not match prepared transaction")
		return
	}
	if err = s.settle(r.Context(), actor.ID, claimID, record, receipt); err != nil {
		problem(w, 500, "payment confirmed but settlement needs retry")
		return
	}
	writeJSON(w, 200, receipt)
}

func (s *Server) settle(ctx context.Context, actorID, claimID string, record paymentRecord, receipt ledger.Receipt) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var data []byte
	if err = tx.QueryRow(`SELECT body FROM claims WHERE id=?`, claimID).Scan(&data); err != nil {
		return err
	}
	var claim protocol.Claim
	if err = json.Unmarshal(data, &claim); err != nil {
		return err
	}
	if claim.State == protocol.ClaimPaid {
		return nil
	}
	if claim.State != protocol.ClaimPaymentPending {
		return errors.New("claim is not pending")
	}
	result, err := tx.Exec(`UPDATE plans SET reserved=reserved-?, paid=paid+? WHERE id=? AND reserved>=?`, claim.AmountStroops, claim.AmountStroops, claim.PlanID, claim.AmountStroops)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count != 1 {
		return errors.New("reserved budget missing")
	}
	claim.State = protocol.ClaimPaid
	record.Receipt = receipt
	if _, err = tx.Exec(`UPDATE claims SET body=? WHERE id=?`, encode(claim), claim.ID); err != nil {
		return err
	}
	if _, err = tx.Exec(`UPDATE payments SET body=? WHERE claim_id=?`, encode(record), claim.ID); err != nil {
		return err
	}
	if err = audit(tx, actorID, "payment_settled", claimID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Server) createRescueTask(w http.ResponseWriter, r *http.Request) {
	actor, ok := s.authorize(w, r, "steward")
	if !ok {
		return
	}
	var task protocol.RescueTask
	if !decode(w, r, &task) {
		return
	}
	var data []byte
	if s.db.QueryRow(`SELECT body FROM projects WHERE id=?`, task.ProjectID).Scan(&data) != nil {
		problem(w, 404, "project not found")
		return
	}
	var project protocol.Project
	_ = json.Unmarshal(data, &project)
	task.OpenedAt = s.now().UTC()
	if task.StewardID != actor.ID || !s.hasActor(task.ReviewerID, "reviewer") || task.Validate(project) != nil {
		problem(w, 400, "invalid or unauthorized rescue task")
		return
	}
	project.State = protocol.ProjectRescueOpen
	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil {
		problem(w, 500, "database unavailable")
		return
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`INSERT INTO rescue_tasks(id,project_id,body) VALUES(?,?,?)`, task.ID, task.ProjectID, encode(task)); err != nil {
		problem(w, 409, "rescue task already exists")
		return
	}
	if _, err = tx.Exec(`UPDATE projects SET body=? WHERE id=?`, encode(project), project.ID); err != nil {
		problem(w, 500, "project update failed")
		return
	}
	if err = audit(tx, actor.ID, "rescue_opened", task.ID); err != nil {
		problem(w, 500, "audit failed")
		return
	}
	if err = tx.Commit(); err != nil {
		problem(w, 500, "rescue commit failed")
		return
	}
	writeJSON(w, 201, task)
}

type AuditEvent struct {
	Sequence   int64  `json:"sequence"`
	OccurredAt string `json:"occurred_at"`
	ActorID    string `json:"actor_id"`
	Action     string `json:"action"`
	SubjectID  string `json:"subject_id"`
}

func (s *Server) getAudit(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.authorizeAny(w, r, "steward", "sponsor", "reviewer", "payer"); !ok {
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT seq,occurred_at,actor_id,action,subject_id FROM audit WHERE subject_id=? ORDER BY seq`, r.PathValue("id"))
	if err != nil {
		problem(w, 500, "audit unavailable")
		return
	}
	defer rows.Close()
	events := []AuditEvent{}
	for rows.Next() {
		var event AuditEvent
		if err := rows.Scan(&event.Sequence, &event.OccurredAt, &event.ActorID, &event.Action, &event.SubjectID); err != nil {
			problem(w, 500, "audit unavailable")
			return
		}
		events = append(events, event)
	}
	if rows.Err() != nil {
		problem(w, 500, "audit unavailable")
		return
	}
	writeJSON(w, 200, events)
}
