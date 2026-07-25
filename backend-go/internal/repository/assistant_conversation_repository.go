package repository

import (
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// AssistantConversationRepository يخزن ويسترجع كل محادثات المساعد الذكي مع
// الموظفين العاديين (نقطة /api/assistant/ask) — أساس لوحة مراجعة المالك.
type AssistantConversationRepository struct {
	db *sqlx.DB
}

func NewAssistantConversationRepository(db *sqlx.DB) *AssistantConversationRepository {
	return &AssistantConversationRepository{db: db}
}

// Create يسجل تبادل رسالة/جواب وحد. لا يرجع خطأ حرج للمتصل — الخدمة اللي تستدعيها
// هي اللي تقرر تتجاهل الخطأ وتكمل (اللوغ ما لازم يكسر تجربة الموظف).
func (r *AssistantConversationRepository) Create(employeeID, message, reply string) error {
	_, err := r.db.Exec(`
		INSERT INTO "AssistantConversation" (id, "employeeId", message, reply)
		VALUES (gen_random_uuid()::text, $1, $2, $3)
	`, employeeID, message, reply)
	return err
}

// AssistantConversationFilter فلاتر اختيارية لقائمة المحادثات — كلها اختيارية.
type AssistantConversationFilter struct {
	EmployeeID string
	From       *time.Time
	To         *time.Time
	Limit      int
	Offset     int
}

// ListAll يرجع المحادثات مرتبة (الأحدث أولاً) مع اسم الموظف، مع دعم فلترة
// بالموظف/التاريخ وتقسيم صفحات بسيط (limit/offset). يرجع أيضاً العدد الكلي
// المطابق للفلتر (بدون limit/offset) حتى تقدر الواجهة تعرف إذا يبقى "تحميل أكثر".
func (r *AssistantConversationRepository) ListAll(f AssistantConversationFilter) ([]model.AssistantConversation, int, error) {
	var conditions []string
	var args []any
	argN := 1

	if f.EmployeeID != "" {
		conditions = append(conditions, `"employeeId" = $`+strconv.Itoa(argN))
		args = append(args, f.EmployeeID)
		argN++
	}
	if f.From != nil {
		conditions = append(conditions, `"createdAt" >= $`+strconv.Itoa(argN))
		args = append(args, *f.From)
		argN++
	}
	if f.To != nil {
		conditions = append(conditions, `"createdAt" <= $`+strconv.Itoa(argN))
		args = append(args, *f.To)
		argN++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM "AssistantConversation" ` + where
	if err := r.db.Get(&total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	query := `SELECT * FROM "AssistantConversation" ` + where +
		` ORDER BY "createdAt" DESC LIMIT $` + strconv.Itoa(argN) + ` OFFSET $` + strconv.Itoa(argN+1)
	args = append(args, limit, offset)

	rows := []model.AssistantConversation{}
	if err := r.db.Select(&rows, query, args...); err != nil {
		return nil, 0, err
	}

	// نجيب أسماء الموظفين المشاركين بدفعة وحدة بدل استعلام لكل صف
	idSet := map[string]bool{}
	for _, row := range rows {
		idSet[row.EmployeeID] = true
	}
	if len(idSet) > 0 {
		ids := make([]string, 0, len(idSet))
		for id := range idSet {
			ids = append(ids, id)
		}
		briefs := []model.EmployeeBrief{}
		query, args, err := sqlx.In(`SELECT id, name, position FROM "Employee" WHERE id IN (?)`, ids)
		if err == nil {
			query = r.db.Rebind(query)
			if err := r.db.Select(&briefs, query, args...); err == nil {
				byID := make(map[string]model.EmployeeBrief, len(briefs))
				for _, b := range briefs {
					byID[b.ID] = b
				}
				for i := range rows {
					if b, ok := byID[rows[i].EmployeeID]; ok {
						brief := b
						rows[i].Employee = &brief
					}
				}
			}
		}
	}

	return rows, total, nil
}

// ListEmployeesWithConversations يرجع قائمة الموظفين المميزين اللي عندهم
// محادثة وحدة عالأقل مع المساعد — لقائمة فلتر بواجهة المالك.
func (r *AssistantConversationRepository) ListEmployeesWithConversations() ([]model.EmployeeBrief, error) {
	rows := []model.EmployeeBrief{}
	err := r.db.Select(&rows, `
		SELECT DISTINCT e.id, e.name, e.position
		FROM "Employee" e
		JOIN "AssistantConversation" ac ON ac."employeeId" = e.id
		ORDER BY e.name
	`)
	return rows, err
}

