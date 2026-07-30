package model

import (
	"time"

	"github.com/lib/pq"
)

// DesignFormQuestion سؤال واحد ضمن نموذج طلب تصميم قابل للتخصيص كلياً (وحدة
// التصميم) — المدير يبني الأسئلة يدوياً (نص/رقم/تاريخ/اختيار من متعدد/خيارات
// متعددة/ملف)، النظام بس يوفّر آلية الإضافة/الترتيب، مو محتوى الاستمارة نفسه.
type DesignFormQuestion struct {
	ID        string         `db:"id" json:"id"`
	Label     string         `db:"label" json:"label"`
	Type      string         `db:"type" json:"type"` // TEXT|TEXTAREA|NUMBER|DATE|SELECT|CHECKBOX|FILE
	Options   pq.StringArray `db:"options" json:"options"`
	Required  bool           `db:"required" json:"required"`
	Order     int            `db:"order" json:"order"`
	CreatedAt time.Time      `db:"createdAt" json:"createdAt"`
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
