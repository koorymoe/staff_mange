// موديل جدول "CostEmployee" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

// (ل) 11-24 دازاي: «تكاليف الشركة» — إدخال يدوي من الإعدادات (المالك بس).
// كل موظف يختاره المالك: نوعه (يدخّل فلوس / إداري / إداري حجوزات)، شفته، راتبه، ساعاته، أيام شغله وإجازاته.
func init() {
	register(Table{
		Name: "CostEmployee",
		Columns: []Column{
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true, Default: "'EARNER'::text"},
			{Name: "shift", Type: "text"},
			{Name: "salary", Type: "numeric", NotNull: true, Default: "0"},
			// (ل) 11-24 دازاي: «الأساسي 350 والباقي حسب الهدف» — الراتب المتوقع = الأساسي + (الراتب − الأساسي) × نسبة الوصول
			{Name: "baseSalary", Type: "numeric", NotNull: true, Default: "350000"},
			{Name: "hoursPerDay", Type: "numeric", NotNull: true, Default: "8"},
			{Name: "daysPerMonth", Type: "numeric", NotNull: true, Default: "26"},
			{Name: "leaveDays", Type: "numeric", NotNull: true, Default: "0"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "CostEmployee_pkey", Def: "PRIMARY KEY (\"employeeId\")"},
			{Name: "CostEmployee_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "cost_employee_kind_ok", Def: "CHECK ((kind = ANY (ARRAY['EARNER'::text, 'ADMIN'::text, 'BOOKING_ADMIN'::text])))"},
			{Name: "cost_employee_shift_ok", Def: "CHECK (((shift IS NULL) OR (shift = ANY (ARRAY['MORNING'::text, 'EVENING'::text]))))"},
			{Name: "cost_employee_numbers_ok", Def: "CHECK (((salary >= (0)::numeric) AND (\"hoursPerDay\" > (0)::numeric) AND (\"daysPerMonth\" > (0)::numeric) AND (\"leaveDays\" >= (0)::numeric)))"},
		},
	})
}
