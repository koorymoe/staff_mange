package handler

import (
	"testing"

	"staffmange-api/internal/model"
)

func TestCanManageExtraTask(t *testing.T) {
	by := "a"
	task := &model.ExtraTask{AssignedByID: &by}
	if !canManageExtraTask("SUPERVISOR", "a", task) || canManageExtraTask("SUPERVISOR", "b", task) ||
		!canManageExtraTask("OWNER", "b", task) || canManageExtraTask("ADMIN", "a", nil) ||
		canManageExtraTask("SUPERVISOR", "", &model.ExtraTask{}) {
		t.Fatal("assigner guard wrong")
	}
}
