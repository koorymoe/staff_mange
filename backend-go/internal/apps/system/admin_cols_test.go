package system

import "testing"

func TestOwnerOnlyCols(t *testing.T) {
	for _, k := range []string{"salary", "baseSalary", "commissionRate", "price"} {
		if !ownerOnlyCol(k) {
			t.Errorf("%s should be owner-only", k)
		}
	}
	if ownerOnlyCol("name") {
		t.Error("name is not owner-only")
	}
	if !ownerOnlyEmployeeCol("Employee", "role") || !ownerOnlyEmployeeCol("Employee", "status") || ownerOnlyEmployeeCol("Task", "status") {
		t.Error("employee role/status guard wrong")
	}
}
