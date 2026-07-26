package model

import "time"

// TeamInventoryToolCatalog قائمة الأدوات/المعدات المطلوبة الأساسية لجرد الفريق
// (شيت "جرد العدد") — بريمة جدارية، مفتاح صباغ... إلخ. قائمة عامة واحدة يشترك
// فيها كل جرد فريق (مو خاصة بموظف وحد مثل PersonalTool).
type TeamInventoryToolCatalog struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
}

// TeamInventoryCheck جلسة جرد فريق: ليدر + موظفين (لغاية اثنين) يختارهم الليدر
// وقت تشغيل الجرد، مطابق لعمود "الليدر"/"الموظف الأول"/"الموظف الثاني" بالشيت.
type TeamInventoryCheck struct {
	ID          string    `db:"id" json:"id"`
	LeaderID    string    `db:"leaderId" json:"leaderId"`
	Employee1ID *string   `db:"employee1Id" json:"employee1Id"`
	Employee2ID *string   `db:"employee2Id" json:"employee2Id"`
	CreatedAt   time.Time `db:"createdAt" json:"createdAt"`

	Leader    *EmployeeBrief              `db:"-" json:"leader"`
	Employee1 *EmployeeBrief              `db:"-" json:"employee1"`
	Employee2 *EmployeeBrief              `db:"-" json:"employee2"`
	Items     []TeamInventoryCheckItem    `db:"-" json:"items"`
}

// PersonRole يميز صاحب الحالة (present/reason) لكل أداة داخل جلسة الجرد — الليدر
// أو الموظف الأول أو الثاني (نفس ترتيب أعمدة الشيت).
const (
	PersonRoleLeader    = "LEADER"
	PersonRoleEmployee1 = "EMPLOYEE1"
	PersonRoleEmployee2 = "EMPLOYEE2"
)

// ShortageReason القيم الثابتة لسبب النقص — مطابقة حرفياً لقائمة الشيت المنسدلة.
const (
	ShortageReasonForgotten = "FORGOTTEN"    // نسيان في مكان معين
	ShortageReasonDamaged   = "DAMAGED"      // يجب جلب القطعة المتلوفة "تلف"
	ShortageReasonUnknown   = "UNKNOWN"      // لا اعرف
)

var ValidShortageReasons = map[string]bool{
	ShortageReasonForgotten: true,
	ShortageReasonDamaged:   true,
	ShortageReasonUnknown:   true,
}

type TeamInventoryCheckItem struct {
	ID         string  `db:"id" json:"id"`
	CheckID    string  `db:"checkId" json:"checkId"`
	ToolName   string  `db:"toolName" json:"toolName"`
	PersonRole string  `db:"personRole" json:"personRole"`
	Present    bool    `db:"present" json:"present"`
	Reason     *string `db:"reason" json:"reason"`
}

// TeamInventoryItemInput مدخل واحد (أداة + دور شخص) وقت إنشاء جلسة الجرد.
type TeamInventoryItemInput struct {
	ToolName   string  `json:"toolName"`
	PersonRole string  `json:"personRole"`
	Present    bool    `json:"present"`
	Reason     *string `json:"reason"`
}

type CreateTeamInventoryCheckRequest struct {
	Employee1ID *string                  `json:"employee1Id"`
	Employee2ID *string                  `json:"employee2Id"`
	Items       []TeamInventoryItemInput `json:"items"`
}

type CreateTeamInventoryToolRequest struct {
	Name string `json:"name"`
}
