package database

// ═══ دور «العلاقات العامة» — طلب (ع) 10-08 ═══
// الإعلام والعلاقات العامة كلهم يشوفون المشاريع المرحّلة للتصوير. الصلاحيات
// الافتراضية للدور بـmodel/permission.go (media + unit_pr).
// ⚠️ قيمة enum جديدة بترحيل لوحدها (نفس درس 0278).
func publicRelationsMigrations() []Migration {
	return []Migration{
		{
			Version: "0332_role_public_relations",
			SQL:     `ALTER TYPE "EmployeeRole" ADD VALUE IF NOT EXISTS 'PUBLIC_RELATIONS';`,
		},
	}
}
