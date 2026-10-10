package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"staffmange-api/internal/model"
)

type SolarRepository struct {
	db *sqlx.DB
}

func NewSolarRepository(db *sqlx.DB) *SolarRepository {
	return &SolarRepository{db: db}
}

// ═══════════════ المكوّنات ═══════════════

func (r *SolarRepository) ListComponents(category string) ([]model.SolarComponent, error) {
	rows := []model.SolarComponent{}
	err := r.db.Select(&rows, `
		SELECT * FROM "SolarComponent"
		WHERE ($1 = '' OR category = $1)
		ORDER BY category, name`, category)
	return rows, err
}

func (r *SolarRepository) CreateComponent(req model.SaveSolarComponentRequest) (*model.SolarComponent, error) {
	id := uuid.NewString()
	_, err := r.db.Exec(`
		INSERT INTO "SolarComponent" (id, name, category, quantity, price, "minStock", specs, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		id, req.Name, req.Category, req.Quantity, req.Price, req.MinStock, req.Specs, req.Notes)
	if err != nil {
		return nil, err
	}
	return r.FindComponent(id)
}

func (r *SolarRepository) UpdateComponent(id string, req model.SaveSolarComponentRequest) (*model.SolarComponent, error) {
	_, err := r.db.Exec(`
		UPDATE "SolarComponent"
		SET name = $2, category = $3, quantity = $4, price = $5,
		    "minStock" = $6, specs = $7, notes = $8, "updatedAt" = now()
		WHERE id = $1`,
		id, req.Name, req.Category, req.Quantity, req.Price, req.MinStock, req.Specs, req.Notes)
	if err != nil {
		return nil, err
	}
	return r.FindComponent(id)
}

func (r *SolarRepository) FindComponent(id string) (*model.SolarComponent, error) {
	var c model.SolarComponent
	if err := r.db.Get(&c, `SELECT * FROM "SolarComponent" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	return &c, nil
}

// DeleteComponent يمنع محو مكوّن مستعمل بمنظومة.
//
// القيد بقاعدة البيانات (ON DELETE RESTRICT) يمنعها أصلاً، بس رسالته
// إنكليزية وما تفهّم المستخدم شنو صار — فنسأل أول ونرد بالعربي.
func (r *SolarRepository) DeleteComponent(id string) error {
	var used int
	if err := r.db.Get(&used, `
		SELECT COUNT(*) FROM "SolarSystem"
		WHERE "panelId" = $1 OR "inverterId" = $1 OR "batteryId" = $1 OR "boardId" = $1
	`, id); err != nil {
		return err
	}
	if used > 0 {
		return fmt.Errorf("ما نكدر نمحيها: مستعملة بـ %d منظومة", used)
	}
	_, err := r.db.Exec(`DELETE FROM "SolarComponent" WHERE id = $1`, id)
	return err
}

// ═══════════════ المنظومات ═══════════════

func (r *SolarRepository) ListSystems(brand string) ([]model.SolarSystem, error) {
	rows := []model.SolarSystem{}
	if err := r.db.Select(&rows, `
		SELECT * FROM "SolarSystem"
		WHERE ($1 = '' OR brand = $1)
		ORDER BY brand, "createdAt" DESC`, brand); err != nil {
		return nil, err
	}
	if err := r.hydrateSystems(toSystemPointers(rows)); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *SolarRepository) FindSystem(id string) (*model.SolarSystem, error) {
	var s model.SolarSystem
	if err := r.db.Get(&s, `SELECT * FROM "SolarSystem" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	if err := r.hydrateSystems([]*model.SolarSystem{&s}); err != nil {
		return nil, err
	}
	return &s, nil
}

func toSystemPointers(rows []model.SolarSystem) []*model.SolarSystem {
	out := make([]*model.SolarSystem, len(rows))
	for i := range rows {
		out[i] = &rows[i]
	}
	return out
}

// hydrateSystems يعبّي مكوّنات كل منظومة ويحسب سعرها ونواقصها.
//
// استعلام واحد لكل المنظومات مو استعلام لكل منظومة: شاشة الكتالوك تعرض
// عشرات المنظومات، ولو سألنا المخزن أربع مرات لكل وحدة صارت مئات
// الاستعلامات بكل فتحة صفحة.
func (r *SolarRepository) hydrateSystems(systems []*model.SolarSystem) error {
	if len(systems) == 0 {
		return nil
	}
	idSet := map[string]bool{}
	add := func(id *string) {
		if id != nil && *id != "" {
			idSet[*id] = true
		}
	}
	for _, s := range systems {
		add(s.PanelID)
		add(s.InverterID)
		add(s.BatteryID)
		add(s.BoardID)
	}
	byID := map[string]model.SolarComponent{}
	if len(idSet) > 0 {
		ids := make([]string, 0, len(idSet))
		for id := range idSet {
			ids = append(ids, id)
		}
		rows := []model.SolarComponent{}
		if err := r.db.Select(&rows, `SELECT * FROM "SolarComponent" WHERE id = ANY($1)`, pq.Array(ids)); err != nil {
			return err
		}
		for _, c := range rows {
			byID[c.ID] = c
		}
	}

	pick := func(id *string) *model.SolarComponent {
		if id == nil || *id == "" {
			return nil
		}
		if c, ok := byID[*id]; ok {
			return &c
		}
		return nil
	}

	for _, s := range systems {
		s.Panel = pick(s.PanelID)
		s.Inverter = pick(s.InverterID)
		s.Battery = pick(s.BatteryID)
		s.Board = pick(s.BoardID)
		s.Price = SolarSystemPrice(s)
		s.Shortages = SolarSystemShortages(s)
	}
	return nil
}

// SolarSystemPrice يحسب سعر المنظومة من أسعار المخزن الحالية.
//
// السعر ما ينخزن بالكتالوك عمداً: سعر اللوح والإنفيرتر يتغير بالسوق، ولو
// خزّنّاه صار الكتالوك يعرض أسعار السنة الماضية بثقة كاملة.
func SolarSystemPrice(s *model.SolarSystem) model.SolarPrice {
	p := model.SolarPrice{
		Wiring:   s.WiringTotalCost,
		Iron:     s.IronTotalCost,
		Install:  s.InstallPrice,
		Program:  s.ProgramPrice,
		Warranty: s.WarrantyPrice,
	}
	if s.Panel != nil {
		p.Panels = s.Panel.Price * float64(s.PanelQty)
	}
	if s.Inverter != nil {
		p.Inverters = s.Inverter.Price * float64(s.InverterQty)
	}
	if s.Battery != nil {
		p.Batteries = s.Battery.Price * float64(s.BatteryQty)
	}
	if s.Board != nil {
		p.Board = s.Board.Price
	}
	p.Components = p.Panels + p.Inverters + p.Batteries
	p.Total = p.Components + p.Board + p.Wiring + p.Iron + p.Install + p.Program + p.Warranty
	return p
}

// SolarSystemShortages يرجّع المكوّنات الي ما تكفي لتجهيز المنظومة.
//
// «ما تكفي» نوعين ولازم نفرّقهن: مكوّن موجود بس كميته أقل من المطلوب،
// ومكوّن انمحى أو ما انربط أصلاً — الثاني أخطر لأن المنظومة تبين كاملة
// وهي ناقصة قطعة.
func SolarSystemShortages(s *model.SolarSystem) []model.SolarShortage {
	out := []model.SolarShortage{}
	check := func(id *string, c *model.SolarComponent, need int) {
		if id == nil || *id == "" {
			return
		}
		if need <= 0 {
			return
		}
		if c == nil {
			out = append(out, model.SolarShortage{ComponentID: *id, Name: "مكوّن محذوف", Required: need, Missing: true})
			return
		}
		if c.Quantity < need {
			out = append(out, model.SolarShortage{
				ComponentID: c.ID, Name: c.Name, Required: need, Available: c.Quantity,
			})
		}
	}
	check(s.PanelID, s.Panel, s.PanelQty)
	check(s.InverterID, s.Inverter, s.InverterQty)
	check(s.BatteryID, s.Battery, s.BatteryQty)
	check(s.BoardID, s.Board, 1)
	return out
}

func (r *SolarRepository) SaveSystem(id string, req model.SaveSolarSystemRequest, byEmployeeID string) (*model.SolarSystem, error) {
	wiring, wiringCost := marshalWiring(req.WiringDetails)
	iron, ironCost := marshalIron(req.IronDetails)

	if id == "" {
		id = uuid.NewString()
		_, err := r.db.Exec(`
			INSERT INTO "SolarSystem" (
				id, brand, model, capacity,
				"panelId", "panelQty", "inverterId", "inverterQty",
				"batteryId", "batteryQty", "boardId",
				"wiringDetails", "wiringTotalCost", "ironDetails", "ironTotalCost",
				"installPrice", "programPrice", "warrantyPrice", notes, "createdById")
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)`,
			id, req.Brand, req.Model, req.Capacity,
			req.PanelID, req.PanelQty, req.InverterID, req.InverterQty,
			req.BatteryID, req.BatteryQty, req.BoardID,
			wiring, wiringCost, iron, ironCost,
			req.InstallPrice, req.ProgramPrice, req.WarrantyPrice, req.Notes, nullIfEmpty(byEmployeeID))
		if err != nil {
			return nil, err
		}
		return r.FindSystem(id)
	}

	_, err := r.db.Exec(`
		UPDATE "SolarSystem" SET
			brand = $2, model = $3, capacity = $4,
			"panelId" = $5, "panelQty" = $6, "inverterId" = $7, "inverterQty" = $8,
			"batteryId" = $9, "batteryQty" = $10, "boardId" = $11,
			"wiringDetails" = $12, "wiringTotalCost" = $13,
			"ironDetails" = $14, "ironTotalCost" = $15,
			"installPrice" = $16, "programPrice" = $17, "warrantyPrice" = $18,
			notes = $19, "updatedAt" = now()
		WHERE id = $1`,
		id, req.Brand, req.Model, req.Capacity,
		req.PanelID, req.PanelQty, req.InverterID, req.InverterQty,
		req.BatteryID, req.BatteryQty, req.BoardID,
		wiring, wiringCost, iron, ironCost,
		req.InstallPrice, req.ProgramPrice, req.WarrantyPrice, req.Notes)
	if err != nil {
		return nil, err
	}
	return r.FindSystem(id)
}

// marshalWiring يحسب كلفة التسليك بالسيرفر مو ياخذها من الواجهة.
// الواجهة تحسبها للعرض، بس الرقم الي ينخزن لازم يجي من نفس السطور
// المخزونة — وإلا صار عدنا مجموع ما يطابق تفاصيله.
func marshalWiring(lines []model.SolarWiringLine) (model.JSONRaw, float64) {
	total := 0.0
	kept := []model.SolarWiringLine{}
	for _, l := range lines {
		if l.Length <= 0 || l.Price <= 0 {
			continue
		}
		kept = append(kept, l)
		total += l.Length * l.Price
	}
	b, _ := json.Marshal(kept)
	return model.JSONRaw(b), total
}

func marshalIron(lines []model.SolarIronLine) (model.JSONRaw, float64) {
	total := 0.0
	kept := []model.SolarIronLine{}
	for _, l := range lines {
		if l.Qty <= 0 || l.Price <= 0 {
			continue
		}
		kept = append(kept, l)
		total += l.Qty * l.Price
	}
	b, _ := json.Marshal(kept)
	return model.JSONRaw(b), total
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// DeleteSystem يمنع محو منظومة انجهزت لزبون — الأرشيف لازم يضل مفهوم.
func (r *SolarRepository) DeleteSystem(id string) error {
	var used int
	if err := r.db.Get(&used, `SELECT COUNT(*) FROM "SolarInstallation" WHERE "systemId" = $1`, id); err != nil {
		return err
	}
	if used > 0 {
		return fmt.Errorf("ما نكدر نمحيها: مجهّزة لـ %d زبون", used)
	}
	_, err := r.db.Exec(`DELETE FROM "SolarSystem" WHERE id = $1`, id)
	return err
}

// ═══════════════ التجهيز — أهم عملية بالنظام ═══════════════

// ProcessSystem يجهّز منظومة لزبون: يخصم مكوّناتها من المخزن ويسجّل
// التركيب ويدخل الزبون بدورة المتابعة.
//
// كلها بمعاملة وحدة، ومحمية بقفل صف (FOR UPDATE) على كل مكوّن. بدون
// القفل، لو موظفين جهّزوا منظومتين بنفس اللحظة يقرون نفس الكمية
// ويكتبونها الاثنين — فينخصم لوح واحد بدل اثنين، والمخزن يكذب.
//
// والنظام القديم جان أسوأ: يخصم بـ max(0, current - required)، يعني لو
// المخزن ما يكفي يخصم للصفر ويكمّل بهدوء وكأن كل شي تمام. هنا نوقف
// ونعلم منو الناقص.
func (r *SolarRepository) ProcessSystem(systemID string, req model.ProcessSolarSystemRequest, byEmployeeID string) (*model.SolarInstallation, error) {
	installDate, err := time.Parse("2006-01-02", req.InstallDate)
	if err != nil {
		return nil, errors.New("تاريخ التركيب مو صحيح")
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var sys model.SolarSystem
	if err := tx.Get(&sys, `SELECT * FROM "SolarSystem" WHERE id = $1`, systemID); err != nil {
		return nil, errors.New("المنظومة مو موجودة")
	}

	// المطلوب من كل مكوّن. البورد دائماً واحد — نفس منطق النظام القديم.
	needs := []solarNeed{}
	addNeed := func(id *string, qty int) {
		if id != nil && *id != "" && qty > 0 {
			needs = append(needs, solarNeed{*id, qty})
		}
	}
	addNeed(sys.PanelID, sys.PanelQty)
	addNeed(sys.InverterID, sys.InverterQty)
	addNeed(sys.BatteryID, sys.BatteryQty)
	addNeed(sys.BoardID, 1)

	// نقفل بترتيب ثابت (بالـ id) حتى معاملتين متزامنتين ما تتقافلن
	// على بعض: وحدة تمسك اللوح وتنتظر البطارية، والثانية بالعكس.
	sortNeeds(needs)

	components := map[string]model.SolarComponent{}
	shortages := []model.SolarShortage{}
	for _, n := range needs {
		var c model.SolarComponent
		if err := tx.Get(&c, `SELECT * FROM "SolarComponent" WHERE id = $1 FOR UPDATE`, n.id); err != nil {
			shortages = append(shortages, model.SolarShortage{ComponentID: n.id, Name: "مكوّن محذوف", Required: n.qty, Missing: true})
			continue
		}
		if c.Quantity < n.qty {
			shortages = append(shortages, model.SolarShortage{
				ComponentID: c.ID, Name: c.Name, Required: n.qty, Available: c.Quantity,
			})
			continue
		}
		components[c.ID] = c
	}
	if len(shortages) > 0 {
		return nil, shortageError(shortages)
	}

	for _, n := range needs {
		if _, err := tx.Exec(`
			UPDATE "SolarComponent" SET quantity = quantity - $2, "updatedAt" = now()
			WHERE id = $1`, n.id, n.qty); err != nil {
			return nil, err
		}
	}

	// الزبون: الموجود ينربط، والجديد ينسجّل بدفتر زبائن الشركة نفسه
	customerID := ""
	if req.CustomerID != nil && *req.CustomerID != "" {
		customerID = *req.CustomerID
	} else {
		customerID = uuid.NewString()
		// عمود العنوان بجدول الزبائن اسمه location مو address —
		// نستعمل عمودهم مو نضيف عمود ثاني بنفس المعنى
		if _, err := tx.Exec(`
			INSERT INTO "Customer" (id, name, phone, location)
			VALUES ($1, $2, $3, NULLIF($4, ''))`,
			customerID, req.CustomerName, req.CustomerPhone, req.CustomerAddress); err != nil {
			return nil, err
		}
	}

	// نسخة السعر وقت البيع — أسعار المخزن تتغير، والي انباع انباع
	sys.Panel = ptrComponent(components, sys.PanelID)
	sys.Inverter = ptrComponent(components, sys.InverterID)
	sys.Battery = ptrComponent(components, sys.BatteryID)
	sys.Board = ptrComponent(components, sys.BoardID)
	price := SolarSystemPrice(&sys)
	breakdown, _ := json.Marshal(price)

	id := uuid.NewString()
	followUp := installDate.AddDate(0, 0, model.SolarFollowUpDays)
	if _, err := tx.Exec(`
		INSERT INTO "SolarInstallation" (
			id, "systemId", "customerId", "installDate", "followUpAt",
			"totalPrice", "priceBreakdown", notes, "createdById")
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		id, systemID, customerID, installDate, followUp,
		price.Total, model.JSONRaw(breakdown), req.Notes, nullIfEmpty(byEmployeeID)); err != nil {
		return nil, err
	}

	payload, _ := json.Marshal(map[string]any{"systemId": systemID, "customerId": customerID, "total": price.Total})
	if _, err := tx.Exec(`
		INSERT INTO "SolarLog" (id, kind, details, payload, "employeeId")
		VALUES ($1, 'PROCESS_SYSTEM', $2, $3, $4)`,
		uuid.NewString(),
		fmt.Sprintf("تجهيز %s %s للزبون %s", sys.Brand, sys.Capacity, req.CustomerName),
		model.JSONRaw(payload), nullIfEmpty(byEmployeeID)); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.FindInstallation(id)
}

func ptrComponent(m map[string]model.SolarComponent, id *string) *model.SolarComponent {
	if id == nil || *id == "" {
		return nil
	}
	if c, ok := m[*id]; ok {
		return &c
	}
	return nil
}

// solarNeed كم قطعة نحتاج من مكوّن معيّن لتجهيز منظومة.
type solarNeed struct {
	id  string
	qty int
}

func sortNeeds(needs []solarNeed) {
	for i := 1; i < len(needs); i++ {
		for j := i; j > 0 && needs[j].id < needs[j-1].id; j-- {
			needs[j], needs[j-1] = needs[j-1], needs[j]
		}
	}
}

func shortageError(shortages []model.SolarShortage) error {
	msg := "ما نكدر نجهّز المنظومة، المخزن ما يكفي:"
	for _, s := range shortages {
		if s.Missing {
			msg += fmt.Sprintf("\n• %s (مو موجود بالمخزن)", s.Name)
			continue
		}
		msg += fmt.Sprintf("\n• %s — مطلوب %d والمتوفر %d", s.Name, s.Required, s.Available)
	}
	return errors.New(msg)
}

// ═══════════════ التركيبات والمتابعة ═══════════════

func (r *SolarRepository) ListInstallations(status string) ([]model.SolarInstallation, error) {
	rows := []model.SolarInstallation{}
	if err := r.db.Select(&rows, `
		SELECT * FROM "SolarInstallation"
		WHERE ($1 = '' OR status = $1)
		ORDER BY "installDate" DESC`, status); err != nil {
		return nil, err
	}
	return rows, r.hydrateInstallations(rows)
}

func (r *SolarRepository) FindInstallation(id string) (*model.SolarInstallation, error) {
	var inst model.SolarInstallation
	if err := r.db.Get(&inst, `SELECT * FROM "SolarInstallation" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	one := []model.SolarInstallation{inst}
	if err := r.hydrateInstallations(one); err != nil {
		return nil, err
	}
	return &one[0], nil
}

func (r *SolarRepository) hydrateInstallations(rows []model.SolarInstallation) error {
	if len(rows) == 0 {
		return nil
	}
	customerIDs := make([]string, 0, len(rows))
	systemIDs := make([]string, 0, len(rows))
	empIDs := []string{}
	for _, i := range rows {
		customerIDs = append(customerIDs, i.CustomerID)
		systemIDs = append(systemIDs, i.SystemID)
		if i.ContactedByID != nil && *i.ContactedByID != "" {
			empIDs = append(empIDs, *i.ContactedByID)
		}
	}

	customers := map[string]model.Customer{}
	crows := []model.Customer{}
	if err := r.db.Select(&crows, `SELECT * FROM "Customer" WHERE id = ANY($1)`, pq.Array(customerIDs)); err != nil {
		return err
	}
	for _, c := range crows {
		customers[c.ID] = c
	}

	systems := map[string]model.SolarSystem{}
	srows := []model.SolarSystem{}
	if err := r.db.Select(&srows, `SELECT * FROM "SolarSystem" WHERE id = ANY($1)`, pq.Array(systemIDs)); err != nil {
		return err
	}
	if err := r.hydrateSystems(toSystemPointers(srows)); err != nil {
		return err
	}
	for _, s := range srows {
		systems[s.ID] = s
	}

	names := map[string]string{}
	if len(empIDs) > 0 {
		type row struct {
			ID   string `db:"id"`
			Name string `db:"name"`
		}
		erows := []row{}
		if err := r.db.Select(&erows, `SELECT id, name FROM "Employee" WHERE id = ANY($1)`, pq.Array(empIDs)); err != nil {
			return err
		}
		for _, e := range erows {
			names[e.ID] = e.Name
		}
	}

	today := time.Now().Truncate(24 * time.Hour)
	for i := range rows {
		if c, ok := customers[rows[i].CustomerID]; ok {
			cc := c
			rows[i].Customer = &cc
		}
		if s, ok := systems[rows[i].SystemID]; ok {
			ss := s
			rows[i].System = &ss
		}
		if rows[i].ContactedByID != nil {
			if n, ok := names[*rows[i].ContactedByID]; ok {
				nn := n
				rows[i].ContactedByName = &nn
			}
		}
		if rows[i].Status == "PENDING" && !rows[i].FollowUpAt.After(today) {
			rows[i].DueForFollowUp = true
			rows[i].DaysOverdue = int(today.Sub(rows[i].FollowUpAt).Hours() / 24)
		}
	}
	return nil
}

// MarkContacted يأشّر إنه الزبون انتصل بيه بعد التركيب.
func (r *SolarRepository) MarkContacted(id, byEmployeeID, notes string) (*model.SolarInstallation, error) {
	if _, err := r.db.Exec(`
		UPDATE "SolarInstallation"
		SET status = 'CONTACTED', "contactedAt" = now(),
		    "contactedById" = $2, "contactNotes" = NULLIF($3, '')
		WHERE id = $1`, id, nullIfEmpty(byEmployeeID), notes); err != nil {
		return nil, err
	}
	return r.FindInstallation(id)
}

// ═══════════════ الأرقام ═══════════════

func (r *SolarRepository) Stats() (*model.SolarStats, error) {
	var s model.SolarStats
	err := r.db.Get(&s, `
		SELECT
			(SELECT COUNT(*) FROM "SolarSystem")                                     AS "systemCount",
			(SELECT COUNT(*) FROM "SolarComponent")                                  AS "componentCount",
			(SELECT COALESCE(SUM(quantity * price), 0) FROM "SolarComponent")        AS "inventoryValue",
			(SELECT COUNT(*) FROM "SolarComponent"
			  WHERE quantity <= "minStock" AND quantity > 0)                         AS "lowStockCount",
			(SELECT COUNT(*) FROM "SolarComponent" WHERE quantity = 0)               AS "outOfStockCount",
			(SELECT COUNT(*) FROM "SolarInstallation")                               AS "processedCount",
			(SELECT COUNT(DISTINCT "customerId") FROM "SolarInstallation")           AS "customerCount",
			(SELECT COUNT(*) FROM "SolarInstallation"
			  WHERE status = 'PENDING' AND "followUpAt" <= baghdad_today())             AS "dueFollowUpCount",
			(SELECT COUNT(*) FROM "SolarInstallation" WHERE status = 'CONTACTED')    AS "contactedCount",
			(SELECT COUNT(*) FROM "SolarInstallation"
			  WHERE date_trunc('month', "installDate") = date_trunc('month', baghdad_today()))
			                                                                         AS "installedThisMonth",
			(SELECT COALESCE(SUM("wiringTotalCost"), 0) FROM "SolarSystem")          AS "totalWiring",
			(SELECT COALESCE(SUM("ironTotalCost"), 0)   FROM "SolarSystem")          AS "totalIron",
			(SELECT COALESCE(SUM("installPrice"), 0)    FROM "SolarSystem")          AS "totalInstall",
			(SELECT COALESCE(SUM("programPrice"), 0)    FROM "SolarSystem")          AS "totalProgram"`)
	return &s, err
}

// LowStock يرجّع المواد الي وصلت حدها الأدنى — لبانر التنبيهات.
func (r *SolarRepository) LowStock() ([]model.SolarComponent, error) {
	rows := []model.SolarComponent{}
	err := r.db.Select(&rows, `
		SELECT * FROM "SolarComponent"
		WHERE quantity <= "minStock"
		ORDER BY quantity, name`)
	return rows, err
}

// SystemTotalPrice يرجّع سعر المنظومة الكلي بأسعار المخزن الحالية.
// يستعملها حجز الطاقة الشمسية حتى السعر المقدّر يجي محسوب مو مكتوب.
func (r *SolarRepository) SystemTotalPrice(systemID string) (float64, error) {
	sys, err := r.FindSystem(systemID)
	if err != nil {
		return 0, err
	}
	return sys.Price.Total, nil
}
