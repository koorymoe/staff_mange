package model

import "time"

type Skill struct {
	ID        string `db:"id" json:"id"`
	Name      string `db:"name" json:"name"`
	ServiceID string `db:"serviceId" json:"serviceId"`
	// Category محور ثاني غير الخدمة: فنية / سلامة / إدارية. جاي من نظام
	// الطاقة الشمسية — «السلامة المهنية» و«خدمة العملاء» مهارات ما تخص
	// خدمة وحدة، تخص كل الخدمات.
	// ⚠️ عمود بالجدول → لازم حقل هنا، وإلا SELECT * يفشل بالسكوت.
	Category    string    `db:"category" json:"category"`
	Description *string   `db:"description" json:"description"`
	CreatedAt   time.Time `db:"createdAt" json:"createdAt"`
}

type Service struct {
	ID       string  `db:"id" json:"id"`
	Name     string  `db:"name" json:"name"`
	Category *string `db:"category" json:"category"`
	// Division: "ENGINEERING" (افتراضي، كل الخدمات القديمة) أو "DECOR" (المهن
	// السبعة الجديدة: حدادة/نجارة/صباغة/سيراميك/لبخ/تأسيس ماء ومجاري/جبس بورد) —
	// نفس تقسيم model.Employee.Division، يحدد أي كتالوج مهارات يظهر لأي شعبة.
	Division  string    `db:"division" json:"division"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
	Skills    []Skill   `db:"-" json:"skills"`
}

type CreateServiceRequest struct {
	Name     string  `json:"name"`
	Category *string `json:"category"`
	// Division: شعبة الخدمة — "ENGINEERING" (الشد/الهندسية) أو "DECOR".
	// هي الي تقرر مهارات الخدمة تطلع لمنو: موظف الديكور ما يشوف مهارات
	// خدمة هندسية أبداً، والعكس. لو ما انبعثت، تنحسب هندسية (سلوك
	// النظام قبل ما ينضاف السؤال).
	Division *string `json:"division"`
}

type CreateSkillRequest struct {
	Name string `json:"name"`
}
