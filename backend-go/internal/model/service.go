package model

import "time"

type Skill struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	ServiceID string    `db:"serviceId" json:"serviceId"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
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
}

type CreateSkillRequest struct {
	Name string `json:"name"`
}
