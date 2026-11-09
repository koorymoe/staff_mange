package service

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ═══ 📜 سجل تحديثات النظام (دازاي 11-06، اقتراح كمال) ═══
// changelog.md مضمّن بالبرنامج، فيتحدث ويا كل نشر. ماتركس وكمال يجاوبون منه على
// «شنو انضاف بالنظام؟». ما بيه أسماء ولا أرقام حساسة، فمسموح لكل المستخدمين.

//go:embed changelog.md
var changelogMD string

type changelogDay struct {
	Date  time.Time
	Items []string
}

// parseChangelog يقرا «## YYYY-MM-DD» وتحته أسطر «- …».
func parseChangelog(md string) []changelogDay {
	var days []changelogDay
	for _, line := range strings.Split(md, "\n") {
		line = strings.TrimSpace(line)
		if d, ok := strings.CutPrefix(line, "## "); ok {
			if t, err := time.Parse("2006-01-02", strings.TrimSpace(d)); err == nil {
				days = append(days, changelogDay{Date: t})
			}
			continue
		}
		if item, ok := strings.CutPrefix(line, "- "); ok && len(days) > 0 {
			days[len(days)-1].Items = append(days[len(days)-1].Items, item)
		}
	}
	return days
}

// ChangelogText التحديثات خلال آخر sinceDays يوم (من now)، أو الي بيها search.
func ChangelogText(now time.Time, sinceDays int, search string) string {
	if sinceDays <= 0 {
		sinceDays = 14
	}
	cut := now.AddDate(0, 0, -sinceDays)
	search = strings.ToLower(strings.TrimSpace(search))
	var b strings.Builder
	for _, d := range parseChangelog(changelogMD) {
		if search == "" && d.Date.Before(cut.Truncate(24*time.Hour)) {
			continue
		}
		var items []string
		for _, it := range d.Items {
			if search == "" || strings.Contains(strings.ToLower(it), search) {
				items = append(items, it)
			}
		}
		if len(items) == 0 {
			continue
		}
		fmt.Fprintf(&b, "%s:\n", d.Date.Format("2006-01-02"))
		for _, it := range items {
			fmt.Fprintf(&b, "- %s\n", it)
		}
	}
	if b.Len() == 0 {
		return "ماكو تحديثات مسجّلة بهالفترة أو بهالكلمة."
	}
	return b.String()
}

// RegisterChangelogTool يضيف أداة system_changelog للوكيل.
func (a *MatrixAgent) RegisterChangelogTool() {
	a.AddTool(changelogTool())
}

func changelogTool() agentTool {
	return agentTool{
		Name: "system_changelog",
		Desc: "سجل تحديثات النظام: شنو انضاف أو تغيّر أو تصلّح، حسب التاريخ. استعملها من أحد يسأل «شنو الجديد بالنظام؟» أو «شوكت انضافت الميزة الفلانية؟». since_days (افتراضي ١٤) أو search بكلمة.",
		Props: map[string]any{
			"since_days": map[string]any{"type": "integer", "description": "كم يوم لورا (افتراضي ١٤)"},
			"search":     map[string]any{"type": "string", "description": "كلمة تدوّر عليها بكل السجل (اختياري)"},
		},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			days := 14
			switch v := in["since_days"].(type) {
			case float64:
				days = int(v)
			case string:
				days, _ = strconv.Atoi(v)
			}
			search, _ := in["search"].(string)
			return ChangelogText(time.Now(), days, search), nil
		},
	}
}
