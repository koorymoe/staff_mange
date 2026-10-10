package service

import (
	"encoding/json"
	"os"
	"testing"

	"staffmange-api/internal/model"
)

type fuzzLine struct {
	Count   int     `json:"count"`
	Install float64 `json:"install"`
	H       int     `json:"h"`
	Wired   bool    `json:"wired"`
	WH      int     `json:"wh"`
	M       int     `json:"m"`
	N       float64 `json:"N"`
	Prog    float64 `json:"prog"`
}
type fuzzCase struct {
	Lines    []fuzzLine `json:"lines"`
	E58      int        `json:"e58"`
	Expected int64      `json:"expected"`
}

// TestExcelParityFuzz يقارن محرك النظام بمحاكي مستقل مكتوب من صيغ الاكسل
// مباشرة (مو من كود النظام)، على 200 فاتورة عشوائية فيها تسليك وبرمجة
// وارتفاعات وأطوال كيبل مختلفة. أي اختلاف بأي حالة = فشل.
func TestExcelParityFuzz(t *testing.T) {
	// الحالات مولّدة من محاكي مستقل يطبّق صيغ الاكسل حرفياً
	// (testdata/excel_sim.py) — محفوظة بالمستودع حتى الفحص يتكرر دائماً.
	raw, err := os.ReadFile("testdata/excel_parity_cases.json")
	if err != nil {
		t.Fatalf("ملف حالات المطابقة مفقود: %v", err)
	}
	var cases []fuzzCase
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	bad := 0
	for i, c := range cases {
		catalog := []model.SystemPriceCatalog{}
		items := []model.ExecutionCostItem{}
		for j, l := range c.Lines {
			itemName := "بند" + string(rune('A'+j))
			wireName := "كيبل" + string(rune('A'+j))
			progName := "برمجة" + string(rune('A'+j))
			catalog = append(catalog,
				model.SystemPriceCatalog{SystemName: "س", ItemName: itemName, Category: "install", Value: l.Install},
				model.SystemPriceCatalog{SystemName: "س", ItemName: wireName, Category: "wiring", Value: l.N},
				model.SystemPriceCatalog{SystemName: "س", ItemName: progName, Category: "programming", Value: l.Prog},
			)
			it := model.ExecutionCostItem{
				SystemName: "س", ItemName: itemName, Count: l.Count, HeightMeters: l.H,
			}
			if l.Wired {
				it.WiringItemName = wireName
				it.WiringHeightMeters = l.WH
				it.CableLengthMeters = l.M
			}
			if l.Prog > 0 {
				it.ProgrammingItem = progName
			}
			items = append(items, it)
		}
		got, _, _, err := CalculateExecutionCostDetailed(items, catalog, c.E58)
		if err != nil {
			t.Fatalf("حالة %d: %v", i, err)
		}
		if got != c.Expected {
			bad++
			if bad <= 5 {
				t.Errorf("حالة %d: الاكسل %d، النظام %d", i, c.Expected, got)
			}
		}
	}
	if bad > 0 {
		t.Fatalf("اختلفت %d حالة من %d", bad, len(cases))
	}
	t.Logf("مطابق بكل %d حالة ✔", len(cases))
}
