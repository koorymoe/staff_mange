package service

import (
	"fmt"
	"math"

	"staffmange-api/internal/repository"
)

// ═══ بند «الالتزام بتسجيل المراحل» — قرار (ع) 10-15 ═══
// 2: كل المراحل انضغطت بوقتها (بلا تسجيل جماعي ولا بعيد، والفرق ويّا حساب النظام ≤ ٢٠ دقيقة)
// 1: انضغطت بس بيها تسجيل جماعي أو بعيد أو فرق كبير
// 0: نسى (النظام سجّل الوصول بداله) أو ناقصة مراحل
func stageLogPoints(f repository.StageLogFact) (int, string) {
	if f.AutoArrive || f.ArrivedAt == nil {
		return 0, "نسى يضغط «وصلنا» — النظام سجّلها بداله من الـGPS"
	}
	if f.StartedAt == nil || f.CompletedAt == nil {
		return 0, "مراحل ناقصة ما انضغطت"
	}
	var issues []string
	if f.Batch > 0 {
		issues = append(issues, "تسجيل جماعي (أكثر من مرحلة بنفس الدقايق)")
	}
	if f.Far > 0 {
		issues = append(issues, "ضغطة بعيدة عن موقع الزبون")
	}
	if f.AutoArrivedAt != nil {
		if d := math.Abs(f.ArrivedAt.Sub(*f.AutoArrivedAt).Minutes()); d > 20 {
			issues = append(issues, fmt.Sprintf("الوصول انسجّل بفرق %d دقيقة عن وصوله الفعلي", int(d)))
		}
	}
	if f.AutoCompletedAt != nil {
		if d := math.Abs(f.CompletedAt.Sub(*f.AutoCompletedAt).Minutes()); d > 20 {
			issues = append(issues, fmt.Sprintf("الإنجاز انسجّل بفرق %d دقيقة عن طلعته الفعلية", int(d)))
		}
	}
	if len(issues) > 0 {
		r := issues[0]
		for _, x := range issues[1:] {
			r += " · " + x
		}
		return 1, r
	}
	return 2, "كل المراحل انضغطت بوقتها"
}
