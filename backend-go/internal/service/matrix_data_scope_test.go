package service

import "testing"

// (دازاي 10-29) الموظف يقرا بيانات مجاله بس.
func TestAllowedQuery(t *testing.T) {
	veh := MatrixDataScope([]string{"agent_matrix", "vehicle_management"})
	if veh == nil {
		t.Fatal("صاحب المركبات لازم إله مجال")
	}
	if MatrixDataScope([]string{"agent_matrix"}) != nil {
		t.Fatal("بلا صلاحيات مجال = ماكو قراءة")
	}
	for q, ok := range map[string]bool{
		`SELECT count(*) FROM "VehicleWashLog" WHERE baghdad_date("createdAt") = baghdad_today()`:   true,
		`SELECT v.name, w.id FROM "Vehicle" v JOIN "VehicleWashLog" w ON w."vehicleId" = v.id`:      true,
		`SELECT e.name FROM "VehicleDriver" d JOIN "Employee" e ON e.id = d."employeeId"`:            true,
		`WITH x AS (SELECT * FROM "Vehicle") SELECT count(*) FROM x`:                                true,
		`SELECT * FROM "LeaderInvoice"`:                                                             false,
		`SELECT * FROM "Vehicle" v JOIN "Expense" e ON true`:                                        false,
		`SELECT salary FROM "Employee"`:                                                             false,
		`SELECT * FROM "SecretaryNote"`:                                                             false,
		`SELECT * FROM "EmployeePermission"`:                                                        false,
		`SELECT table_name FROM information_schema.tables`:                                          false,
	} {
		if err := AllowedQuery(q, veh); (err == nil) != ok {
			t.Errorf("%s: err=%v want ok=%v", q, err, ok)
		}
	}
}
