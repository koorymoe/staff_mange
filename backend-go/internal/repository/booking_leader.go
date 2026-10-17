package repository

// ═══ ليدر الحجز وكادره — تعريف واحد (ع 10-16) ═══
// الدالتان بقاعدة البيانات (migration 0371). أي شاشة أو عدّاد يسأل «منو
// ليدر هذا الحجز؟» أو «منو اشتغل عليه؟» يستعمل هذني — مو شرطاً يكتبه بيده،
// لأن چان عدنا أكثر من ٨ تعريفات تتناقض.

// BookingLeaderSQL معرّف ليدر الحجز (أو NULL). qualifier = اسم جدول الحجز بالاستعلام.
func BookingLeaderSQL(qualifier string) string { return `booking_leader_id(` + qualifier + `.id)` }

// BookingCrewHasSQL شرط: الموظف (تعبير SQL) من كادر الحجز.
func BookingCrewHasSQL(qualifier, employeeExpr string) string {
	return employeeExpr + ` IN (SELECT booking_crew_ids(` + qualifier + `.id))`
}

// BookingWorkedBySQL شرط: الموظف (تعبير SQL) اشتغل على الحجز — مكلّف، أو
// مشرفه الليدر، أو ليدر/عضو بمهمته. ⚠️ نفس booking_crew_ids (0371) بالضبط،
// بس بصيغة EXISTS تستعمل الفهارس — للعدّادات الي تمر على حجوزات هواية.
func BookingWorkedBySQL(qualifier, employeeExpr string) string {
	q, e := qualifier, employeeExpr
	return `(EXISTS (SELECT 1 FROM "BookingAssignment" bw WHERE bw."bookingId" = ` + q + `.id AND bw."employeeId" = ` + e + `)
		OR (` + q + `."projectSupervisorId" = ` + e + ` AND EXISTS (SELECT 1 FROM "Employee" bwe WHERE bwe.id = ` + e + ` AND bwe."isLeader"))
		OR EXISTS (SELECT 1 FROM "Mission" bwm WHERE bwm."bookingId" = ` + q + `.id AND (bwm."leaderId" = ` + e + ` OR ` + e + ` = ANY(bwm."memberIds"))))`
}
