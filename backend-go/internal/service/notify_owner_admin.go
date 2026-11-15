package service

import (
	"staffmange-api/internal/repository"
)

// notifyOwnerAndAdmin يرسل للمالك والمدير **كل واحد بروحه**.
//
// 🔴 چان النمط: «أرسل للمالك، وإذا فشل ارجع» — فإذا فشل إرسال المالك ما
// يوصل للمدير، والخدمة الدورية حاجزة علامة الأسبوع قبل الإرسال، فالملخص
// يضيع لذاك الأسبوع. هسه فشل واحد ما يمنع الثاني، والخطأ يرجع مجمّع.
func notifyOwnerAndAdmin(n *repository.NotificationRepository, notifType, msg string) error {
	// صف واحد للشخص حتى لو طابق الدورين.
	return n.CreateForRoles([]string{"OWNER", "ADMIN"}, notifType, msg)
}
