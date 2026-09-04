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

// ═══ ليش المرحلة تنحسب من الحجز ═══
//
// «هاي ليش مو صحيحة؟ ما أعتقد مربوطة — لأن عندي مهام مكتملة وعندي
// هواي أمور، ليش هيج كله صفرة؟».
//
// وهو محق: كل العدّادات صفر إلا «تم الإسناد». والسبب إن مرحلة المهمة
// (`Mission.stage`) ما تتقدّم إلا لمن أحد يضغط أزرار **شاشة المهام**
// نفسها (`PUT /missions/{id}/stage`). بس الكادر بالواقع يشتغل على
// **الحجز**: يأشّر المواد جاهزة، ويوصل، ويبدي، ويخلّص — وكل هذي
// تنكتب بجدول الحجوزات، وما تلمس المهمة أبداً.
//
// فالمهمة تقعد بـ«تم الإسناد» للأبد حتى لو حجزها منجز من أسبوع.
// الشاشة ما كانت مكسورة — كانت تعرض بصدق حقلاً محد يحدّثه.
//
// الحل: المرحلة **تنحسب** من أوقات الحجز الحقيقية، وناخذ الأبعد بين
// الاثنين (مرحلة المهمة المسجّلة، والمرحلة المستنتجة من الحجز).
//
// ⚠️ «الأبعد» مقصودة: أزرار شاشة المهام تبقى شغّالة (مرحلة «بالطريق»
// مثلاً ماكو إلها مقابل بالحجز، فلازم تنحفظ من المهمة)، والحجز يدفع
// المرحلة للأمام لمن يسبقها. وبأي حال ما ترجع المرحلة للورا.
const missionStageExpr = `
	CASE GREATEST(
		CASE m.stage
			WHEN 'ASSIGNED' THEN 0 WHEN 'MATERIALS_PREP' THEN 1 WHEN 'MATERIALS_READY' THEN 2
			WHEN 'EN_ROUTE' THEN 3 WHEN 'ARRIVED' THEN 4 WHEN 'WORK_STARTED' THEN 5
			WHEN 'COMPLETED' THEN 6 WHEN 'STOPPED' THEN 7 ELSE 0 END,
		CASE
			WHEN b.status = 'CANCELLED' THEN 7
			WHEN b.status = 'COMPLETED' OR b."completedAt" IS NOT NULL THEN 6
			WHEN b."startedAt" IS NOT NULL OR b.status = 'IN_PROGRESS' THEN 5
			WHEN b."arrivedAt" IS NOT NULL THEN 4
			WHEN b."materialsReadyAt" IS NOT NULL THEN 2
			ELSE 0 END
	)
		WHEN 0 THEN 'ASSIGNED' WHEN 1 THEN 'MATERIALS_PREP' WHEN 2 THEN 'MATERIALS_READY'
		WHEN 3 THEN 'EN_ROUTE' WHEN 4 THEN 'ARRIVED' WHEN 5 THEN 'WORK_STARTED'
		WHEN 6 THEN 'COMPLETED' ELSE 'STOPPED' END`

// missionSelect يجيب أعمدة المهمة مع المرحلة المحسوبة، والأوقات
// المكمّلة من الحجز.
//
// ⚠️ الأوقات تنكمّل بـ`COALESCE` مو تنكتب فوگ: الوقت المسجّل بالمهمة
// (لو أحد استعمل أزرارها) أدق لأنه ينسجّل بلحظته، ووقت الحجز يعبّي
// الفراغ بس. بدون هذا الخط الزمني بالبطاقة يطلع فاضي مع إن الشغل صار.
const missionSelect = `
	SELECT m.*,
	       ` + missionStageExpr + ` AS stage,
	       COALESCE(m."materialsReadyAt", b."materialsReadyAt") AS "materialsReadyAt",
	       COALESCE(m."arrivedAt", b."arrivedAt") AS "arrivedAt",
	       COALESCE(m."workStartedAt", b."startedAt") AS "workStartedAt",
	       COALESCE(m."completedAt", b."completedAt") AS "completedAt"
	FROM "Mission" m
	JOIN "Booking" b ON b.id = m."bookingId"`

func (r *MissionRepository) List(stage, leaderID, employeeID string) ([]model.Mission, error) {
	query := missionSelect + ` WHERE 1=1`
	args := []any{}
	if stage != "" {
		args = append(args, stage)
		query += fmt.Sprintf(` AND `+missionStageExpr+` = $%d`, len(args))
	}
	if leaderID != "" {
		args = append(args, leaderID)
		query += fmt.Sprintf(` AND m."leaderId" = $%d`, len(args))
	}
	if employeeID != "" {
		args = append(args, employeeID, employeeID)
		query += fmt.Sprintf(` AND (m."leaderId" = $%d OR $%d = ANY(m."memberIds"))`, len(args)-1, len(args))
	}
	query += ` ORDER BY m."createdAt" DESC`

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
	err := r.db.Get(&m, missionSelect+` WHERE m.id = $1`, id)
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

// ExistsForBooking هل الحجز عنده مهمة أصلاً.
//
// ضروري لأن التكليف ينعاد: الإداري يبدّل فني، أو يضيف ثاني للكادر،
// أو يعيد تكليف نفس الواحد بالغلط. بلا هذا الفحص كل ضغطة تخلق مهمة
// جديدة لنفس الحجز، وشاشة التتبع تمتلئ نسخ مكررة ما إلها معنى.
func (r *MissionRepository) ExistsForBooking(bookingID string) (bool, error) {
	var n int
	err := r.db.Get(&n, `SELECT COUNT(*) FROM "Mission" WHERE "bookingId" = $1`, bookingID)
	return n > 0, err
}

// SyncCrew يحدّث ليدر المهمة وأعضاءها لمن يتغيّر الكادر بعد الإنشاء.
//
// ⚠️ ما نلمس المرحلة (stage): مهمة وصلت «بالطريق» وانضاف إلها فني
// ما ترجع «تم الإسناد» — الشغل ماشي، بس الكادر توسّع.
func (r *MissionRepository) SyncCrew(bookingID, leaderID string, memberIDs []string) error {
	_, err := r.db.Exec(`
		UPDATE "Mission"
		SET "leaderId" = $2, "memberIds" = $3, "updatedAt" = now()
		WHERE "bookingId" = $1
	`, bookingID, leaderID, pq.Array(memberIDs))
	return err
}

// BackfillFromAssignments يخلق مهام للحجوزات الشغّالة الي عندها كادر
// مكلّف وما إلها مهمة.
//
// ⚠️ ليش نحتاجها: التوليد التلقائي ينشبك بلحظة **التكليف**. يعني كل
// الحجوزات الي انكلّفت قبل هذا التصليح تضل بلا مهمة للأبد، وشاشة
// التتبع تبقى فاضية لحد ما ينكلّف حجز جديد. الشركة شغّالة من زمان،
// فالشاشة راح تظل تكذب على المراقب.
//
// نقتصر على الشغّالة (مو منجزة ولا ملغاة ولا مؤرشفة): المنجزة خلصت
// وماكو شي يتتبّع بيها، وإنشاء مهام إلها يملأ الشاشة بتاريخ ميت.
//
// الليدر: المؤشّر «تيم ليدر» بالكادر، وإلا أول مكلّف.
// يرجّع عدد الي انخلق.
func (r *MissionRepository) BackfillFromAssignments() (int, error) {
	res, err := r.db.Exec(`
		WITH candidates AS (
			SELECT b.id  AS booking_id,
			       b.address,
			       b."mapLatitude"  AS lat,
			       b."mapLongitude" AS lng,
			       ARRAY_AGG(ba."employeeId" ORDER BY e."isLeader" DESC, ba."createdAt") AS member_ids
			FROM "Booking" b
			JOIN "BookingAssignment" ba ON ba."bookingId" = b.id
			JOIN "Employee" e           ON e.id = ba."employeeId"
			LEFT JOIN "Mission" m       ON m."bookingId" = b.id
			WHERE m.id IS NULL
			  AND b."archivedAt" IS NULL
			  AND NOT ` + BookingDeletePendingSQL("b") + `
			  AND b.status NOT IN ('COMPLETED', 'CANCELLED')
			GROUP BY b.id, b.address, b."mapLatitude", b."mapLongitude"
		), numbered AS (
			SELECT c.*,
			       (SELECT COUNT(*) FROM "Mission") + ROW_NUMBER() OVER (ORDER BY c.booking_id) AS seq
			FROM candidates c
		)
		INSERT INTO "Mission" (id, code, "bookingId", "leaderId", "memberIds",
		                       "customerLat", "customerLng", "customerAddress", stage, "updatedAt")
		SELECT gen_random_uuid()::text,
		       'MSN-' || LPAD(seq::text, 4, '0'),
		       booking_id,
		       member_ids[1],
		       member_ids,
		       lat, lng, address,
		       'ASSIGNED', now()
		FROM numbered
	`)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
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
	// ⚠️⚠️ چان `SELECT * FROM "Mission"` بلا اسم مستعار، والشرط تحته
	// يشير لـ`m."leaderId"` — SQL غلط، فالمسار **يرجّع خطأ دائماً**.
	// يعني كل فني وليدر يفتح «تتبع المهام» ما تجيه ولا مهمة، والشاشة
	// تطلعله فارغة بلا ما يعرف السبب. (المالك والمدير ما انتبهوا لأن
	// مسارهم غير: /missions مو /missions/my.)
	//
	// ونستعمل missionSelect مو استعلاماً خاصاً: هو الي يحسب المرحلة
	// الحقيقية (missionStageExpr) ويجيب أوقات الحجز — و`SELECT *`
	// چان يرجّع m.stage الخام حتى لو انصلّح الاسم المستعار.
	missions := []model.Mission{}
	err := r.db.Select(&missions, missionSelect+`
		WHERE (m."leaderId" = $1 OR $1 = ANY(m."memberIds"))
		  AND `+missionStageExpr+` NOT IN ('COMPLETED', 'STOPPED')
		ORDER BY m."createdAt" DESC
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
`+missionSelect+` WHERE `+missionStageExpr+` NOT IN ('COMPLETED', 'STOPPED') ORDER BY m."createdAt" DESC
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
	query := missionSelect + ` WHERE ` + missionStageExpr + ` IN ('COMPLETED', 'STOPPED')`
	args := []any{}
	if from != nil {
		args = append(args, *from)
		query += fmt.Sprintf(` AND m."createdAt" >= $%d`, len(args))
	}
	if to != nil {
		args = append(args, *to)
		query += fmt.Sprintf(` AND m."createdAt" <= $%d`, len(args))
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
