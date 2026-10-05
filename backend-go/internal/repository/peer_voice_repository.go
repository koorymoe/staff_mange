package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ صوت الموظفين + المشاكل الوظيفية — البيانات ═══

type PeerVoiceRepository struct{ db *sqlx.DB }

func NewPeerVoiceRepository(db *sqlx.DB) *PeerVoiceRepository { return &PeerVoiceRepository{db: db} }

type PeerCheckin struct {
	ID         string    `db:"id" json:"id"`
	EmployeeID string    `db:"employeeId" json:"employeeId"`
	Week       time.Time `db:"week" json:"week"`
	Mood       *int      `db:"mood" json:"mood"`
	Worry      *string   `db:"worry" json:"worry"`
	Suggestion *string   `db:"suggestion" json:"suggestion"`
	Nothing    bool      `db:"nothing" json:"nothing"`
	Urgent     bool      `db:"urgent" json:"urgent"`
	CreatedAt  time.Time `db:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time `db:"updatedAt" json:"updatedAt"`
}

type PeerNote struct {
	ID            string     `db:"id" json:"id"`
	CheckinID     string     `db:"checkinId" json:"checkinId"`
	AuthorID      string     `db:"authorId" json:"authorId"`
	AuthorName    string     `db:"authorName" json:"authorName"`
	TargetID      string     `db:"targetId" json:"targetId"`
	TargetName    string     `db:"targetName" json:"targetName"`
	Score         int        `db:"score" json:"score"`
	Kind          string     `db:"kind" json:"kind"`
	Text          *string    `db:"text" json:"text"`
	NeedsFollowup bool       `db:"needsFollowup" json:"needsFollowup"`
	FWhat         *string    `db:"fWhat" json:"fWhat"`
	FWhy          *string    `db:"fWhy" json:"fWhy"`
	FWish         *string    `db:"fWish" json:"fWish"`
	Suggestion    *string    `db:"suggestion" json:"suggestion"`
	Accepted      *bool      `db:"accepted" json:"accepted"`
	AcceptNote    *string    `db:"acceptNote" json:"acceptNote"`
	Severity      string     `db:"severity" json:"severity"`
	ReportedAt    *time.Time `db:"reportedAt" json:"reportedAt"`
	CreatedAt     time.Time  `db:"createdAt" json:"createdAt"`
}

const noteSelect = `SELECT n.id, n."checkinId", n."authorId", a.name AS "authorName", n."targetId", t.name AS "targetName",
	n.score, n.kind, n.text, n."needsFollowup", n."fWhat", n."fWhy", n."fWish", n.suggestion, n.accepted, n."acceptNote",
	n.severity, n."reportedAt", n."createdAt"
	FROM "PeerNote" n JOIN "Employee" a ON a.id = n."authorId" JOIN "Employee" t ON t.id = n."targetId"`

func (r *PeerVoiceRepository) Checkin(employeeID string, week time.Time) (*PeerCheckin, error) {
	var c PeerCheckin
	if err := r.db.Get(&c, `SELECT * FROM "PeerCheckin" WHERE "employeeId" = $1 AND week = $2`, employeeID, week); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *PeerVoiceRepository) UpsertCheckin(employeeID string, week time.Time, mood *int, worry, suggestion *string, nothing, urgent bool) (string, error) {
	var id string
	err := r.db.Get(&id, `INSERT INTO "PeerCheckin" ("employeeId", week, mood, worry, suggestion, nothing, urgent)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT ("employeeId", week) DO UPDATE SET mood = EXCLUDED.mood, worry = EXCLUDED.worry, suggestion = EXCLUDED.suggestion,
		  nothing = EXCLUDED.nothing, urgent = EXCLUDED.urgent, "updatedAt" = now()
		RETURNING id`, employeeID, week, mood, worry, suggestion, nothing, urgent)
	return id, err
}

// ReplaceNotes يمسح ملاحظات الأسبوع الي بعدها ما تعمّقت ويكتب الجديدة.
// (الملاحظة الي انرفع تقريرها تبقى — ما ننمسح شي وصل للمراقب.)
func (r *PeerVoiceRepository) DropDraftNotes(checkinID string) {
	_, _ = r.db.Exec(`DELETE FROM "PeerNote" WHERE "checkinId" = $1 AND "reportedAt" IS NULL`, checkinID)
}

func (r *PeerVoiceRepository) AddNote(checkinID, authorID, targetID string, score int, kind string, text *string, needs bool, severity string) (string, error) {
	var id string
	err := r.db.Get(&id, `INSERT INTO "PeerNote" ("checkinId", "authorId", "targetId", score, kind, text, "needsFollowup", severity)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`, checkinID, authorID, targetID, score, kind, text, needs, severity)
	return id, err
}

func (r *PeerVoiceRepository) Note(id string) (*PeerNote, error) {
	var n PeerNote
	if err := r.db.Get(&n, noteSelect+` WHERE n.id = $1`, id); err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *PeerVoiceRepository) NotesOfCheckin(checkinID string) ([]PeerNote, error) {
	rows := []PeerNote{}
	err := r.db.Select(&rows, noteSelect+` WHERE n."checkinId" = $1 ORDER BY n."createdAt"`, checkinID)
	return rows, err
}

func (r *PeerVoiceRepository) SaveFollowup(id string, what, why, wish, suggestion *string, severity string) error {
	_, err := r.db.Exec(`UPDATE "PeerNote" SET "fWhat" = $2, "fWhy" = $3, "fWish" = $4, suggestion = $5,
		severity = CASE WHEN severity = 'URGENT' THEN 'URGENT' ELSE $6 END WHERE id = $1`, id, what, why, wish, suggestion, severity)
	return err
}

func (r *PeerVoiceRepository) Accept(id string, accepted bool, note *string) error {
	_, err := r.db.Exec(`UPDATE "PeerNote" SET accepted = $2, "acceptNote" = $3, "reportedAt" = now() WHERE id = $1`, id, accepted, note)
	return err
}

func (r *PeerVoiceRepository) MarkReported(id string) {
	_, _ = r.db.Exec(`UPDATE "PeerNote" SET "reportedAt" = COALESCE("reportedAt", now()) WHERE id = $1`, id)
}

type Colleague struct {
	ID     string `db:"id" json:"id"`
	Name   string `db:"name" json:"name"`
	Worked bool   `db:"worked" json:"worked"`
}

// Colleagues كل الموظفين الفعّالين (غيره)، والي اشتغل وياهم بآخر ٧ أيام أول.
func (r *PeerVoiceRepository) Colleagues(me string) ([]Colleague, error) {
	rows := []Colleague{}
	err := r.db.Select(&rows, `
		WITH mine AS (
			SELECT b.id FROM "Booking" b
			WHERE COALESCE(b."scheduledAt", b."createdAt") > now() - interval '7 days'
			  AND (b."projectSupervisorId" = $1 OR EXISTS (SELECT 1 FROM "BookingAssignment" a WHERE a."bookingId" = b.id AND a."employeeId" = $1))
		), mates AS (
			SELECT a."employeeId" AS id FROM "BookingAssignment" a JOIN mine ON mine.id = a."bookingId"
			UNION SELECT b."projectSupervisorId" FROM "Booking" b JOIN mine ON mine.id = b.id WHERE b."projectSupervisorId" IS NOT NULL
		)
		SELECT e.id, e.name, (e.id IN (SELECT id FROM mates)) AS worked
		FROM "Employee" e WHERE e.status = 'ACTIVE' AND e.id <> $1 AND e.role::text <> 'OWNER'
		ORDER BY (e.id IN (SELECT id FROM mates)) DESC, e.name`, me)
	return rows, err
}

func (r *PeerVoiceRepository) ActiveEmployee(id string) bool {
	var ok bool
	_ = r.db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "Employee" WHERE id = $1 AND status = 'ACTIVE')`, id)
	return ok
}

// ── للمدير والمراقب ──

type MoodWeek struct {
	Week  time.Time `db:"week" json:"week"`
	Avg   *float64  `db:"avg" json:"avg"`
	Count int       `db:"n" json:"count"`
}

func (r *PeerVoiceRepository) MoodByWeek(weeks int) ([]MoodWeek, error) {
	rows := []MoodWeek{}
	err := r.db.Select(&rows, `SELECT week, avg(mood)::float8 AS avg, count(*)::int AS n FROM "PeerCheckin"
		WHERE week > current_date - make_interval(weeks => $1) GROUP BY week ORDER BY week`, weeks)
	return rows, err
}

type PeerScore struct {
	EmployeeID string   `db:"employeeId" json:"employeeId"`
	Name       string   `db:"name" json:"name"`
	Avg        float64  `db:"avg" json:"avg"`
	Count      int      `db:"n" json:"count"`
	Problems   int      `db:"problems" json:"problems"`
	PrevAvg    *float64 `db:"prevAvg" json:"prevAvg"`
}

// Scores شكد زملاؤه راضين منه: آخر ٤ أسابيع مقابل الـ٤ الي قبلها.
func (r *PeerVoiceRepository) Scores() ([]PeerScore, error) {
	rows := []PeerScore{}
	err := r.db.Select(&rows, `
		SELECT n."targetId" AS "employeeId", e.name, avg(n.score)::float8 AS avg, count(*)::int AS n,
		       count(*) FILTER (WHERE n.kind = 'PROBLEM' OR n.score <= 2)::int AS problems,
		       (SELECT avg(p.score)::float8 FROM "PeerNote" p WHERE p."targetId" = n."targetId"
		         AND p."createdAt" BETWEEN now() - interval '56 days' AND now() - interval '28 days') AS "prevAvg"
		FROM "PeerNote" n JOIN "Employee" e ON e.id = n."targetId"
		WHERE n."createdAt" > now() - interval '28 days'
		GROUP BY n."targetId", e.name ORDER BY avg(n.score)`)
	return rows, err
}

type Tension struct {
	AID   string `db:"aId" json:"aId"`
	AName string `db:"aName" json:"aName"`
	BID   string `db:"bId" json:"bId"`
	BName string `db:"bName" json:"bName"`
	AToB  int    `db:"aToB" json:"aToB"`
	BToA  int    `db:"bToA" json:"bToA"`
}

// Tensions «توتر متبادل»: كل واحد قيّم الثاني سيء (١–٢ أو مشكلة) بآخر ١٤ يوم.
func (r *PeerVoiceRepository) Tensions() ([]Tension, error) {
	rows := []Tension{}
	err := r.db.Select(&rows, `
		WITH bad AS (
			SELECT "authorId", "targetId", min(score) AS s FROM "PeerNote"
			WHERE "createdAt" > now() - interval '14 days' AND (score <= 2 OR kind = 'PROBLEM')
			GROUP BY 1, 2
		)
		SELECT x."authorId" AS "aId", ea.name AS "aName", x."targetId" AS "bId", eb.name AS "bName", x.s AS "aToB", y.s AS "bToA"
		FROM bad x JOIN bad y ON y."authorId" = x."targetId" AND y."targetId" = x."authorId"
		JOIN "Employee" ea ON ea.id = x."authorId" JOIN "Employee" eb ON eb.id = x."targetId"
		WHERE x."authorId" < x."targetId"`)
	return rows, err
}

// Notes ملاحظات آخر N يوم — للمدير كلها، وللمراقب بس الي انرفع تقريرها (التعمّق).
func (r *PeerVoiceRepository) Notes(days int, reportedOnly bool) ([]PeerNote, error) {
	rows := []PeerNote{}
	q := noteSelect + ` WHERE n."createdAt" > now() - make_interval(days => $1)`
	if reportedOnly {
		q += ` AND n."reportedAt" IS NOT NULL AND n."needsFollowup"`
	}
	err := r.db.Select(&rows, q+` ORDER BY CASE n.severity WHEN 'URGENT' THEN 0 WHEN 'ATTENTION' THEN 1 ELSE 2 END, n."createdAt" DESC LIMIT 300`, days)
	return rows, err
}

type CheckinRow struct {
	PeerCheckin
	Name string `db:"name" json:"name"`
}

func (r *PeerVoiceRepository) Checkins(days int) ([]CheckinRow, error) {
	rows := []CheckinRow{}
	err := r.db.Select(&rows, `SELECT c.*, e.name FROM "PeerCheckin" c JOIN "Employee" e ON e.id = c."employeeId"
		WHERE c."createdAt" > now() - make_interval(days => $1) AND (c.worry IS NOT NULL OR c.suggestion IS NOT NULL OR c.urgent)
		ORDER BY c.urgent DESC, c."createdAt" DESC LIMIT 200`, days)
	return rows, err
}

// ── المشاكل الوظيفية ──

type WorkplaceIssue struct {
	ID           string     `db:"id" json:"id"`
	Title        string     `db:"title" json:"title"`
	PartyAID     string     `db:"partyAId" json:"partyAId"`
	PartyAName   string     `db:"partyAName" json:"partyAName"`
	PartyBID     *string    `db:"partyBId" json:"partyBId"`
	PartyBName   *string    `db:"partyBName" json:"partyBName"`
	OpenedByName *string    `db:"openedByName" json:"openedByName"`
	Description  string     `db:"description" json:"description"`
	Severity     string     `db:"severity" json:"severity"`
	Status       string     `db:"status" json:"status"`
	AISummary    *string    `db:"aiSummary" json:"aiSummary"`
	Decision     *string    `db:"decision" json:"decision"`
	DecidedBy    *string    `db:"decidedByName" json:"decidedByName"`
	DecidedAt    *time.Time `db:"decidedAt" json:"decidedAt"`
	FollowUpAt   *time.Time `db:"followUpAt" json:"followUpAt"`
	SourceNoteID *string    `db:"sourceNoteId" json:"sourceNoteId"`
	CreatedAt    time.Time  `db:"createdAt" json:"createdAt"`
}

const issueSelect = `SELECT i.id, i.title, i."partyAId", a.name AS "partyAName", i."partyBId", b.name AS "partyBName", o.name AS "openedByName",
	i.description, i.severity, i.status, i."aiSummary", i.decision, d.name AS "decidedByName", i."decidedAt", i."followUpAt",
	i."sourceNoteId", i."createdAt"
	FROM "WorkplaceIssue" i JOIN "Employee" a ON a.id = i."partyAId" LEFT JOIN "Employee" b ON b.id = i."partyBId"
	LEFT JOIN "Employee" o ON o.id = i."openedById" LEFT JOIN "Employee" d ON d.id = i."decidedById"`

func (r *PeerVoiceRepository) Issues() ([]WorkplaceIssue, error) {
	rows := []WorkplaceIssue{}
	err := r.db.Select(&rows, issueSelect+` ORDER BY CASE i.status WHEN 'OPEN' THEN 0 WHEN 'IN_PROGRESS' THEN 1 ELSE 2 END, i."createdAt" DESC LIMIT 200`)
	return rows, err
}

func (r *PeerVoiceRepository) Issue(id string) (*WorkplaceIssue, error) {
	var w WorkplaceIssue
	if err := r.db.Get(&w, issueSelect+` WHERE i.id = $1`, id); err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *PeerVoiceRepository) CreateIssue(title, a string, b *string, openedBy, desc, severity string, source *string) (string, error) {
	var id string
	err := r.db.Get(&id, `INSERT INTO "WorkplaceIssue" (title, "partyAId", "partyBId", "openedById", description, severity, "sourceNoteId")
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, $7) RETURNING id`, title, a, b, openedBy, desc, severity, source)
	return id, err
}

type IssueEntry struct {
	ID         string    `db:"id" json:"id"`
	AuthorName *string   `db:"authorName" json:"authorName"`
	AuthorID   *string   `db:"authorId" json:"authorId"`
	Kind       string    `db:"kind" json:"kind"`
	Text       string    `db:"text" json:"text"`
	CreatedAt  time.Time `db:"createdAt" json:"createdAt"`
}

func (r *PeerVoiceRepository) Entries(issueID string) ([]IssueEntry, error) {
	rows := []IssueEntry{}
	err := r.db.Select(&rows, `SELECT x.id, e.name AS "authorName", x."authorId", x.kind, x.text, x."createdAt"
		FROM "WorkplaceIssueEntry" x LEFT JOIN "Employee" e ON e.id = x."authorId" WHERE x."issueId" = $1 ORDER BY x."createdAt"`, issueID)
	return rows, err
}

func (r *PeerVoiceRepository) AddEntry(issueID string, authorID *string, kind, text string) error {
	_, err := r.db.Exec(`INSERT INTO "WorkplaceIssueEntry" ("issueId", "authorId", kind, text) VALUES ($1, $2, $3, $4)`, issueID, authorID, kind, text)
	if err == nil {
		_, _ = r.db.Exec(`UPDATE "WorkplaceIssue" SET "updatedAt" = now(), status = CASE WHEN status = 'OPEN' THEN 'IN_PROGRESS' ELSE status END WHERE id = $1`, issueID)
	}
	return err
}

func (r *PeerVoiceRepository) SetSummary(id, summary string) {
	_, _ = r.db.Exec(`UPDATE "WorkplaceIssue" SET "aiSummary" = $2, "updatedAt" = now() WHERE id = $1`, id, summary)
}

func (r *PeerVoiceRepository) Decide(id, decision, by, status string, followUp *time.Time) error {
	_, err := r.db.Exec(`UPDATE "WorkplaceIssue" SET decision = $2, "decidedById" = $3, "decidedAt" = now(), status = $4,
		"followUpAt" = $5, "updatedAt" = now() WHERE id = $1`, id, decision, by, status, followUp)
	return err
}

// FollowupsFor مشاكل الموظف طرف بيها، وحان وقت متابعتها، وبعده ما جاوب بعد الموعد.
func (r *PeerVoiceRepository) FollowupsFor(me string) ([]WorkplaceIssue, error) {
	rows := []WorkplaceIssue{}
	err := r.db.Select(&rows, issueSelect+` WHERE (i."partyAId" = $1 OR i."partyBId" = $1) AND i.status <> 'RESOLVED'
		AND i."followUpAt" IS NOT NULL AND i."followUpAt" <= now()
		AND NOT EXISTS (SELECT 1 FROM "WorkplaceIssueEntry" x WHERE x."issueId" = i.id AND x."authorId" = $1
		                AND x.kind = 'FOLLOWUP' AND x."createdAt" >= i."followUpAt")`, me)
	return rows, err
}

func (r *PeerVoiceRepository) IsParty(issueID, me string) bool {
	var ok bool
	_ = r.db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "WorkplaceIssue" WHERE id = $1 AND ("partyAId" = $2 OR "partyBId" = $2))`, issueID, me)
	return ok
}

// Context سياق الطرفين للتحليل: حجوزات مشتركة، شكاوى، رضا زملائهم.
type IssueContext struct {
	SharedBookings int      `db:"shared" json:"sharedBookings"`
	AComplaints    int      `db:"aComplaints" json:"aComplaints"`
	BComplaints    int      `db:"bComplaints" json:"bComplaints"`
	APeerAvg       *float64 `db:"aPeer" json:"aPeerAvg"`
	BPeerAvg       *float64 `db:"bPeer" json:"bPeerAvg"`
}

func (r *PeerVoiceRepository) Context(a string, b *string) IssueContext {
	var c IssueContext
	bid := ""
	if b != nil {
		bid = *b
	}
	_ = r.db.Get(&c, `SELECT
		(SELECT count(DISTINCT x."bookingId") FROM "BookingAssignment" x JOIN "BookingAssignment" y ON y."bookingId" = x."bookingId"
		  WHERE x."employeeId" = $1 AND y."employeeId" = $2 AND x."createdAt" > now() - interval '90 days')::int AS shared,
		(SELECT count(*) FROM "Complaint" WHERE "relatedEmployeeId" = $1 AND "createdAt" > now() - interval '90 days')::int AS "aComplaints",
		(SELECT count(*) FROM "Complaint" WHERE "relatedEmployeeId" = $2 AND "createdAt" > now() - interval '90 days')::int AS "bComplaints",
		(SELECT avg(score)::float8 FROM "PeerNote" WHERE "targetId" = $1 AND "createdAt" > now() - interval '60 days') AS "aPeer",
		(SELECT avg(score)::float8 FROM "PeerNote" WHERE "targetId" = $2 AND "createdAt" > now() - interval '60 days') AS "bPeer"`, a, bid)
	return c
}
