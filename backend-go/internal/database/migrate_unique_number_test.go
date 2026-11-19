package database

import (
	"strconv"
	"testing"
)

// (دازاي 11-19) كل ترحيل جديد ياخذ رقم ما مستعمل — كاكاشي وأنا أخذنا 0511 بنفس اليوم.
// القديمة الي انطبقت تبقى (أرقامها مسجّلة بالقاعدة)؛ من 0515 وطالع ممنوع التكرار.
const uniqueNumberBaseline = 514

func TestNewMigrationNumbersUnique(t *testing.T) {
	seen := map[int]string{}
	for _, m := range versionedMigrations() {
		g := migNum.FindStringSubmatch(m.Version)
		if g == nil {
			continue
		}
		n, _ := strconv.Atoi(g[1])
		if prev, ok := seen[n]; ok && n > uniqueNumberBaseline {
			t.Errorf("رقم الترحيل %04d مكرر: %s و%s — خذ رقم جديد (آخر رقم +1) بعد ما تسحب avg", n, prev, m.Version)
		}
		seen[n] = m.Version
	}
}
