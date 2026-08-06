package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"staffmange-api/internal/model"
)

type MissionRepository struct {
	db          *sqlx.DB
	bookingRepo *BookingRepository
}

func NewMissionRepository(db *sqlx.DB, bookingRepo *BookingRepository) *MissionRepository {
	return &MissionRepository{db: db, bookingRepo: bookingRepo}
}

func (r *MissionRepository) loadEmployeeBrief(id string) *model.EmployeeBrief {
	var brief model.EmployeeBrief
	if err := r.db.Get(&brief, `SELECT id, name FROM "Employee" WHERE id = $1`, id); err != nil {
		return nil
	}
	return &brief
}

func (r *MissionRepository) hydrate(m *model.Mission, withBooking, withEvents bool) error {
	if withBooking {
		booking, err := r.bookingRepo.FindByID(m.BookingID)
		if err == nil {
			m.Booking = booking
		}
	}
	if withEvents {
		events := []model.MissionEvent{}
		if err := r.db.Select(&events, `SELECT * FROM "MissionEvent" WHERE "missionId" = $1 ORDER BY "createdAt" ASC`, m.ID); err == nil {
			m.Events = events
		}
	}
	return nil
}

// attachBookingsAndEvents يعبّي حجوزات المهام وأحداثها بعدد استعلامات ثابت.
//
// hydrate جانت تنادي FindByID لكل مهمة — وكل نداء يسوي حزمة استعلامات
// كاملة للحجز (زبون، خدمات، كادر، سجل الجدولة). يعني قائمة ٢٠٠ مهمة =
// مئات الرحلات لقاعدة البيانات. هنا نجيب كل الحجوزات بضربة وحدة.
//
// latestEventOnly: شاشة المتابعة تحتاج آخر حدث بس، وشاشة التفاصيل تحتاج
// كل الأحداث — نفس السلوك السابق بالضبط، بس بلا حلقة استعلامات.
func (r *MissionRepository) attachBookingsAndEvents(missions []model.Mission, withEvents, latestEventOnly bool) error {
	if len(missions) == 0 {
		return nil
	}
	bookingIDs := make([]string, 0, len(missions))
	missionIDs := make([]string, 0, len(missions))
	for i := range missions {
		bookingIDs = append(bookingIDs, missions[i].BookingID)
		missionIDs = append(missionIDs, missions[i].ID)
	}

	bookings, err := r.bookingRepo.FindByIDs(bookingIDs)
	if err != nil {
		return err
	}
	for i := range missions {
		if b := bookings[missions[i].BookingID]; b != nil {
			missions[i].Booking = b
		}
	}

	if !withEvents {
		return nil
	}
	// مهمة بلا أحداث لازم ترجع قائمة فاضية مو null — هيك جانت ترجع قبل،
	// والواجهة تعتمد عليها.
	for i := range missions {
		missions[i].Events = []model.MissionEvent{}
	}
	events := []model.MissionEvent{}
	query := `SELECT * FROM "MissionEvent" WHERE "missionId" = ANY($1) ORDER BY "missionId", "createdAt" ASC`
	if latestEventOnly {
		// DISTINCT ON يطلّع آخر حدث لكل مهمة بضربة وحدة — نفس ما تسويه
		// ORDER BY "createdAt" DESC LIMIT 1 لكل مهمة على حدة.
		query = `SELECT DISTINCT ON ("missionId") * FROM "MissionEvent"
			WHERE "missionId" = ANY($1) ORDER BY "missionId", "createdAt" DESC`
	}
	if err := r.db.Select(&events, query, pq.Array(missionIDs)); err != nil {
		return err
	}
	byMission := map[string][]model.MissionEvent{}
	for _, e := range events {
		byMission[e.MissionID] = append(byMission[e.MissionID], e)
	}
	for i := range missions {
		if list := byMission[missions[i].ID]; list != nil {
			missions[i].Events = list
		}
	}
	return nil
}

func (r *MissionRepository) attachLeaderAndMembers(missions []model.Mission) {
	empIDs := map[string]bool{}
	for _, m := range missions {
		empIDs[m.LeaderID] = true
		for _, id := range m.MemberIDs {
			empIDs[id] = true
		}
	}
	briefs := map[string]model.EmployeeBrief{}
	for id := range empIDs {
		if b := r.loadEmployeeBrief(id); b != nil {
			briefs[id] = *b
		}
	}
	for i := range missions {
		if b, ok := briefs[missions[i].LeaderID]; ok {
			missions[i].Leader = &b
		}
		members := make([]model.EmployeeBrief, 0, len(missions[i].MemberIDs))
		for _, id := range missions[i].MemberIDs {
			if b, ok := briefs[id]; ok {
				members = append(members, b)
			} else {
				members = append(members, model.EmployeeBrief{ID: id, Name: "?"})
			}
		}
		missions[i].Members = members
	}
}

func (r *MissionRepository) List(stage, leaderID, employeeID string) ([]model.Mission, error) {
	query := `SELECT * FROM "Mission" WHERE 1=1`
	args := []any{}
	if stage != "" {
		args = append(args, stage)
		query += fmt.Sprintf(` AND stage = $%d`, len(args))
	}
	if leaderID != "" {
		args = append(args, leaderID)
		query += fmt.Sprintf(` AND "leaderId" = $%d`, len(args))
	}
	if employeeID != "" {
		args = append(args, employeeID, employeeID)
		query += fmt.Sprintf(` AND ("leaderId" = $%d OR $%d = ANY("memberIds"))`, len(args)-1, len(args))
	}
	query += ` ORDER BY "createdAt" DESC`

	missions := []model.Mission{}
	if err := r.db.Select(&missions, query, args...); err != nil {
		return nil, err
	}
	if err := r.attachBookingsAndEvents(missions, true, false); err != nil {
		return nil, err
	}
	r.attachLeaderAndMembers(missions)
	return missions, nil
}

func (r *MissionRepository) FindByID(id string) (*model.Mission, error) {
	var m model.Mission
	err := r.db.Get(&m, `SELECT * FROM "Mission" WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := r.hydrate(&m, true, true); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MissionRepository) CountAll() (int, error) {
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM "Mission"`)
	return count, err
}

func (r *MissionRepository) Create(code, bookingID, leaderID string, memberIDs []string, customerLat, customerLng *float64, customerAddress *string) (*model.Mission, error) {
	var m model.Mission
	err := r.db.Get(&m, `
		INSERT INTO "Mission" (id, code, "bookingId", "leaderId", "memberIds", "customerLat", "customerLng", "customerAddress", stage, "updatedAt")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, $7, 'ASSIGNED', now())
		RETURNING *
	`, code, bookingID, leaderID, pq.Array(memberIDs), customerLat, customerLng, customerAddress)
	if err != nil {
		return nil, err
	}
	if err := r.hydrate(&m, true, false); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MissionRepository) UpdateStage(id, stage string, fields map[string]any) (*model.Mission, error) {
	setClauses := []string{`stage = $1`, `"updatedAt" = now()`}
	args := []any{stage}
	for col, val := range fields {
		args = append(args, val)
		setClauses = append(setClauses, fmt.Sprintf(`"%s" = $%d`, col, len(args)))
	}
	query := `UPDATE "Mission" SET `
	for i, c := range setClauses {
		if i > 0 {
			query += `, `
		}
		query += c
	}
	query += fmt.Sprintf(` WHERE id = $%d RETURNING *`, len(args)+1)
	args = append(args, id)

	var m model.Mission
	if err := r.db.Get(&m, query, args...); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MissionRepository) CreateEvent(missionID, employeeID, action string, lat, lng *float64, note *string) error {
	_, err := r.db.Exec(`
		INSERT INTO "MissionEvent" (id, "missionId", "employeeId", action, lat, lng, note)
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6)
	`, missionID, employeeID, action, lat, lng, note)
	return err
}

func (r *MissionRepository) ListForEmployee(employeeID string) ([]model.Mission, error) {
	missions := []model.Mission{}
	err := r.db.Select(&missions, `
		SELECT * FROM "Mission"
		WHERE ("leaderId" = $1 OR $1 = ANY("memberIds")) AND stage NOT IN ('COMPLETED', 'STOPPED')
		ORDER BY "createdAt" DESC
	`, employeeID)
	if err != nil {
		return nil, err
	}
	if err := r.attachBookingsAndEvents(missions, true, false); err != nil {
		return nil, err
	}
	return missions, nil
}

func (r *MissionRepository) ListActive() ([]model.Mission, error) {
	missions := []model.Mission{}
	err := r.db.Select(&missions, `
		SELECT * FROM "Mission" WHERE stage NOT IN ('COMPLETED', 'STOPPED') ORDER BY "createdAt" DESC
	`)
	if err != nil {
		return nil, err
	}
	if err := r.attachBookingsAndEvents(missions, true, true); err != nil {
		return nil, err
	}
	r.attachLeaderAndMembers(missions)
	return missions, nil
}

func (r *MissionRepository) ListForPerformanceReport(from, to *string) ([]model.Mission, error) {
	query := `SELECT * FROM "Mission" WHERE stage IN ('COMPLETED', 'STOPPED')`
	args := []any{}
	if from != nil {
		args = append(args, *from)
		query += fmt.Sprintf(` AND "createdAt" >= $%d`, len(args))
	}
	if to != nil {
		args = append(args, *to)
		query += fmt.Sprintf(` AND "createdAt" <= $%d`, len(args))
	}
	missions := []model.Mission{}
	err := r.db.Select(&missions, query, args...)
	return missions, err
}

func (r *MissionRepository) LoadEmployeeBriefsByIDs(ids []string) (map[string]model.EmployeeBrief, error) {
	briefs := map[string]model.EmployeeBrief{}
	if len(ids) == 0 {
		return briefs, nil
	}
	rows := []model.EmployeeBrief{}
	query, args, err := sqlx.In(`SELECT id, name FROM "Employee" WHERE id IN (?)`, ids)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	if err := r.db.Select(&rows, query, args...); err != nil {
		return nil, err
	}
	for _, row := range rows {
		briefs[row.ID] = row
	}
	return briefs, nil
}
