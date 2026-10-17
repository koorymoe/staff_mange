//go:build !linux

package handler

// diskUsage نسخة احتياطية لأي نظام غير لينكس (مثلاً وقت التطوير المحلي على
// ويندوز) — syscall.Statfs غير موجود إلا بلينكس، فهذا يرجع أصفار بدل ما
// يفشل البناء. سيرفر الإنتاج دايماً لينكس فيستخدم النسخة الحقيقية.
func diskUsage() (totalGB, usedGB, freeGB float64) {
	return 0, 0, 0
}
