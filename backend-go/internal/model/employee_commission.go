package model

import "time"

// EmployeeCommission عمولة موظف واحد (ليدر أو فني) محسوبة تلقائياً عند إنشاء
// فاتورة ليدر واحدة — لا تحتاج تشغيل يدوي منفصل.
//
// حساب الليدر (EmployeeID على الفاتورة، لازم isLeader=true):
//   - executionCommission = executionCost * 0.4
//   - salesCommission = totalProfit * 0.5  حيث totalProfit = مجموع (quantity *
//     profitPerUnit) عبر كل بنود المواد بالفاتورة. (تفسير حاسم مالياً: كلام
//     المالك "نسبة الأرباح مال المبيع تضرب في 0.5" يقصد بيه مبلغ الربح الفعلي
//     المشتق أصلاً من هامش سعر البيع مقابل سعر الجملة لكل مادة — مو نسبة مئوية
//     مجردة تُضرب بمبلغ آخر. تطبيق totalProfit * 0.5 مباشرة هو القراءة
//     الاقتصادية المنطقية ويطابق "فلوس الربح × 0.5".)
//   - totalCommission = executionCommission + salesCommission
//
// حساب الفني العادي (أي موظف مربوط بنفس الحجز عبر BookingAssignment بدور
// TECH_1/TECH_2/TECH_3 غير الليدر نفسه):
//   - executionCommission = executionCost * 0.3 (كاملة لكل فني، بدون تقسيم
//     بينهم حتى لو تعددوا — مؤكد صراحة من المالك)
//   - salesCommission = 0
//   - totalCommission = executionCommission
type EmployeeCommission struct {
	ID                  string    `db:"id" json:"id"`
	EmployeeID          string    `db:"employeeId" json:"employeeId"`
	LeaderInvoiceID     string    `db:"leaderInvoiceId" json:"leaderInvoiceId"`
	Role                string    `db:"role" json:"role"` // LEADER | TECHNICIAN
	ExecutionCommission float64   `db:"executionCommission" json:"executionCommission"`
	SalesCommission     float64   `db:"salesCommission" json:"salesCommission"`
	TotalCommission     float64   `db:"totalCommission" json:"totalCommission"`
	CreatedAt           time.Time `db:"createdAt" json:"createdAt"`
}

const (
	CommissionRoleLeader     = "LEADER"
	CommissionRoleTechnician = "TECHNICIAN"
)

// leaderExecutionCommissionRate نسبة عمولة الليدر من تكلفة التنفيذ.
const leaderExecutionCommissionRate = 0.4

// leaderSalesCommissionRate نسبة عمولة الليدر من إجمالي ربح المواد بالفاتورة.
const leaderSalesCommissionRate = 0.5

// technicianExecutionCommissionRate نسبة عمولة الفني العادي من تكلفة التنفيذ —
// كاملة لكل فني على حدة، بدون تقسيم بين الفنيين المتعددين بنفس الحجز.
const technicianExecutionCommissionRate = 0.3

// CalculateLeaderCommission يحسب عمولة الليدر من تكلفة التنفيذ وإجمالي ربح
// بنود المواد بالفاتورة (quantity * profitPerUnit لكل بند).
func CalculateLeaderCommission(executionCost float64, materials []LeaderInvoiceMaterialItem) (executionCommission, salesCommission, totalProfit float64) {
	for _, m := range materials {
		totalProfit += m.Quantity * m.ProfitPerUnit
	}
	executionCommission = executionCost * leaderExecutionCommissionRate
	salesCommission = totalProfit * leaderSalesCommissionRate
	return executionCommission, salesCommission, totalProfit
}

// CalculateTechnicianCommission يحسب عمولة فني عادي (غير الليدر) على نفس الحجز.
func CalculateTechnicianCommission(executionCost float64) float64 {
	return executionCost * technicianExecutionCommissionRate
}
