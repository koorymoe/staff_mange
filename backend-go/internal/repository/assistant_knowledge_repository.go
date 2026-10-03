package repository

import (
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// AssistantKnowledgeRepository يخزن ويسترجع "المعرفة المتعلّمة" من محادثات
// المساعد الذكي — أساس ذاكرة المساعد طويلة الأمد (RAG بسيط بدون متجهات).
type AssistantKnowledgeRepository struct {
	db *sqlx.DB
}

func NewAssistantKnowledgeRepository(db *sqlx.DB) *AssistantKnowledgeRepository {
	return &AssistantKnowledgeRepository{db: db}
}

// Create يخزن سطر معرفة وحد. learnedFromEmployeeID اختياري (فاضي = "" يعني NULL).
func (r *AssistantKnowledgeRepository) Create(topic, content, learnedFromEmployeeID string) error {
	var employeeIDArg any
	if learnedFromEmployeeID != "" {
		employeeIDArg = learnedFromEmployeeID
	}
	_, err := r.db.Exec(`
		INSERT INTO "AssistantKnowledge" (id, topic, content, "learnedFromEmployeeId")
		VALUES (gen_random_uuid()::text, $1, $2, $3)
	`, topic, content, employeeIDArg)
	return err
}

// stopWords كلمات عربية شائعة جداً (أدوات ربط/ضمائر) نستبعدها من كلمات البحث
// حتى الاستعلام يركّز على الكلمات ذات المعنى بس.
var knowledgeStopWords = map[string]bool{
	"من": true, "الى": true, "إلى": true, "على": true, "عن": true, "في": true,
	"هذا": true, "هذه": true, "ذلك": true, "التي": true, "الذي": true,
	"مع": true, "او": true, "أو": true, "و": true, "ما": true, "لا": true,
	"شنو": true, "شلون": true, "هل": true, "كل": true, "بس": true, "انا": true,
	"أنا": true, "انت": true, "أنت": true, "هو": true, "هي": true,
}

// ExtractKeywords يستخرج كلمات بحث بسيطة من رسالة الموظف — يقسمها على الفراغات،
// يشيل الكلمات القصيرة جداً وكلمات الوقف الشائعة. لا حاجة لتحليل لغوي معقّد،
// المطلوب بس مطابقة كلمات مفتاحية تقريبية (keyword overlap) مو دقة تامة.
func ExtractKeywords(message string) []string {
	fields := strings.Fields(message)
	seen := map[string]bool{}
	var out []string
	for _, f := range fields {
		w := strings.Trim(f, ".,،؟!?؛:\"'()[]{}")
		if len([]rune(w)) < 3 {
			continue
		}
		if knowledgeStopWords[w] {
			continue
		}
		if seen[w] {
			continue
		}
		seen[w] = true
		out = append(out, w)
		if len(out) >= 6 {
			break
		}
	}
	return out
}

// SearchRelevant يرجع أحدث سطور معرفة (حتى limit) موضوعها أو محتواها يحتوي
// أي وحدة من الكلمات المفتاحية المستخرجة من رسالة الموظف — تطابق تقريبي
// بسيط (ILIKE) كافي هنا، مو محتاجين بحث نصي كامل أو ترتيب دلالي متطور.
func (r *AssistantKnowledgeRepository) SearchRelevant(keywords []string, limit int) ([]model.AssistantKnowledge, error) {
	rows := []model.AssistantKnowledge{}
	if len(keywords) == 0 {
		return rows, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 8
	}

	var conditions []string
	var args []any
	argN := 1
	for _, kw := range keywords {
		conditions = append(conditions, `(topic ILIKE $`+strconv.Itoa(argN)+` OR content ILIKE $`+strconv.Itoa(argN)+`)`)
		args = append(args, "%"+kw+"%")
		argN++
	}

	query := `SELECT * FROM "AssistantKnowledge" WHERE ` + strings.Join(conditions, " OR ") +
		` ORDER BY "createdAt" DESC LIMIT $` + strconv.Itoa(argN)
	args = append(args, limit)

	err := r.db.Select(&rows, query, args...)
	return rows, err
}
