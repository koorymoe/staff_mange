package handler

import (
	"net/http"
	"strconv"

	"staffmange-api/internal/repository"
)

// BackupHandler مراقبة النسخ الاحتياطية — للمالك وحده.
//
// ⚠️ كل مسارات هذا الهاندلر لازم تنلف بـmiddleware.RequireOwner.
// ما تحطها بأي مجموعة صلاحيات ثانية ولا تنطيها لـADMIN: شرط صريح من
// المالك إن هذا الجزء يبقى إله هو بس بعد ما يسلّم إدارة النظام.
type BackupHandler struct {
	runs *repository.BackupRunRepository
}

func NewBackupHandler(runs *repository.BackupRunRepository) *BackupHandler {
	return &BackupHandler{runs: runs}
}

// GET /api/owner/backups — الحالة + آخر التشغيلات بطلب واحد.
func (h *BackupHandler) Overview(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 30
	}
	health, err := h.runs.Health()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر قراءة حالة النسخ الاحتياطية")
		return
	}
	runs, err := h.runs.List(limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر قراءة سجل النسخ الاحتياطية")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"health": health, "runs": runs})
}
