package service

import "testing"

// (ل) 11-22 دازاي: تغيير الدور يشيل صلاحيات الدور القديم
func TestSameRoles(t *testing.T) {
	if !sameRoles([]string{"A", "B"}, []string{"B", "A"}) || sameRoles([]string{"A"}, []string{"B"}) || sameRoles([]string{"A"}, []string{"A", "B"}) {
		t.Fatal("sameRoles غلط")
	}
}
