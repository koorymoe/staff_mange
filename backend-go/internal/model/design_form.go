package model

import (
	"time"

	"github.com/lib/pq"
)

// DesignForm استمارة تصميم مستقلة بيها اسمها الخاص ورابط عام (publicToken)
// يُرسَل للزبون مباشرة بدون تسجيل دخول ولا تسريب أي معلومة عن باقي النظام —
// كل استمارة تحمل أسئلتها الخاصة بيها منفصلة عن استمارات ثانية.
type DesignForm struct {
	ID          string    `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	PublicToken string    `db:"publicToken" json:"publicToken"`
	CreatedAt   time.Time `db:"createdAt" json:"createdAt"`
}

type CreateDesignFormRequest struct {
	Name string `json:"name"`
}

// DesignFormQuestion سؤال واحد ضمن استمارة تصميم معيّنة (formId) قابل للتخصيص
// كلياً — المدير يبني الأسئلة يدوياً (نص/رقم/تاريخ/اختيار من متعدد/خيارات
// متعددة/ملف)، النظام بس يوفّر آلية الإضافة/الترتيب، مو محتوى الاستمارة نفسه.
type DesignFormQuestion struct {
	ID        string         `db:"id" json:"id"`
	FormID    string         `db:"formId" json:"formId"`
	Label     string         `db:"label" json:"label"`
	Type      string         `db:"type" json:"type"` // TEXT|TEXTAREA|NUMBER|DATE|SELECT|CHECKBOX|FILE
	Options   pq.StringArray `db:"options" json:"options"`
	Required  bool           `db:"required" json:"required"`
	Order     int            `db:"order" json:"order"`
	CreatedAt time.Time      `db:"createdAt" json:"createdAt"`
}

// DesignFormSubmission جواب زبون كامل على استمارة تصميم معيّنة — الأجوبة تنخزن
// كـJSON (معرّف السؤال → الجواب) حتى تتحمّل أي شكل أسئلة نضيفها بالمستقبل
// بدون تغيير هيكل الجدول.
type DesignFormSubmission struct {
	ID          string    `db:"id" json:"id"`
	FormID      string    `db:"formId" json:"formId"`
	Answers     RawJSON   `db:"answers" json:"answers"`
	SubmittedAt time.Time `db:"submittedAt" json:"submittedAt"`
}

type SubmitDesignFormRequest struct {
	Answers map[string]any `json:"answers"`
}

const (
	DesignFormQuestionText     = "TEXT"
	DesignFormQuestionTextarea = "TEXTAREA"
	DesignFormQuestionNumber   = "NUMBER"
	DesignFormQuestionDate     = "DATE"
	DesignFormQuestionSelect   = "SELECT"
	DesignFormQuestionCheckbox = "CHECKBOX"
	DesignFormQuestionFile     = "FILE"
)

var validDesignFormQuestionTypes = map[string]bool{
	DesignFormQuestionText: true, DesignFormQuestionTextarea: true, DesignFormQuestionNumber: true,
	DesignFormQuestionDate: true, DesignFormQuestionSelect: true, DesignFormQuestionCheckbox: true,
	DesignFormQuestionFile: true,
}

func IsValidDesignFormQuestionType(t string) bool {
	return validDesignFormQuestionTypes[t]
}

type CreateDesignFormQuestionRequest struct {
	FormID   string   `json:"formId"`
	Label    string   `json:"label"`
	Type     string   `json:"type"`
	Options  []string `json:"options"`
	Required bool     `json:"required"`
}

type UpdateDesignFormQuestionRequest struct {
	Label    *string  `json:"label"`
	Type     *string  `json:"type"`
	Options  []string `json:"options"`
	Required *bool    `json:"required"`
}

type ReorderDesignFormQuestionsRequest struct {
	QuestionIDs []string `json:"questionIds"`
}
