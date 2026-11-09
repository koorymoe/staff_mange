package model

import "time"

// ProjectWorkType نوع عمل واحد يظهر بقائمة "نوع العمل" عند إنشاء/تعديل مشروع —
// قابل للإضافة والحذف من إعدادات وحدة إدارة المشاريع، بدل قائمة ثابتة بالكود.
type ProjectWorkType struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
}

type CreateProjectWorkTypeRequest struct {
	Name string `json:"name"`
}
