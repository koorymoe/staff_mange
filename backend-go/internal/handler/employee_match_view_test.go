package handler

import (
	"testing"

	"staffmange-api/internal/model"
)

// /employees/match يمر بنفس فلتر /employees: الراتب ما يطلع لزميل، ومهارة الخدمة تبقى للمنسّق.
func TestMatchViewHidesSalary(t *testing.T) {
	sal, yes := 900000.0, true
	list := []model.Employee{{ID: "other", Name: "زميل", Salary: &sal, HasRequiredSkill: &yes}}
	v := ViewEmployees(list, "TECHNICIAN", "me")[0]
	if _, ok := v["salary"]; ok {
		t.Fatal("الراتب طلع لزميل")
	}
	if v["hasRequiredSkill"] != true {
		t.Fatal("hasRequiredSkill لازم يبقى")
	}
	if _, ok := ViewEmployees(list, "TECHNICIAN", "other")[0]["salary"]; !ok {
		t.Fatal("صاحب الراتب يشوفه")
	}
}
