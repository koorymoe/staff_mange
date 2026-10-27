package model

import (
	"strings"
	"sync"
	"time"
)

// ═══ أسماء الوكلاء (دازاي 10-27: «اريد اكدر اعدل كل الاسماء مال الذكائات») ═══
// الاسم الافتراضي مكتوب بالنصوص والبرومبتات؛ Rebrand يبدّله بالاسم الي حدده المالك
// بنقاط الخروج (البرومبتات والإشعارات)، حتى ما نلمس مئات النصوص.

var AgentDefaultNames = map[string]string{"MATRIX": "ماتركس", "IRIS": "ايريس", "QUALITY": "مساعد الجودة", "BOOKING": "مساعد الحجوزات"}

var (
	agentNameMu     sync.Mutex
	agentNameLoader func() map[string]string
	agentNameCache  map[string]string
	agentNameAt     time.Time
)

// SetAgentNameLoader يجيب الأسماء المحفوظة (agent → name) من القاعدة.
func SetAgentNameLoader(f func() map[string]string) {
	agentNameMu.Lock()
	agentNameLoader, agentNameCache = f, nil
	agentNameMu.Unlock()
}

// InvalidateAgentNames بعد ما المالك يغيّر اسم.
func InvalidateAgentNames() {
	agentNameMu.Lock()
	agentNameCache = nil
	agentNameMu.Unlock()
}

// AgentName الاسم الحالي للوكيل (الافتراضي إذا ما انحفظ اسم).
func AgentName(agent string) string {
	agentNameMu.Lock()
	defer agentNameMu.Unlock()
	if agentNameCache == nil || time.Since(agentNameAt) > time.Minute {
		agentNameCache = map[string]string{}
		if agentNameLoader != nil {
			agentNameCache = agentNameLoader()
		}
		agentNameAt = time.Now()
	}
	if n := strings.TrimSpace(agentNameCache[agent]); n != "" {
		return n
	}
	return AgentDefaultNames[agent]
}

// Rebrand يبدّل اسمي ماتركس وايريس الافتراضيين بالاسمين الحاليين.
func Rebrand(s string) string {
	if m := AgentName("MATRIX"); m != "ماتركس" && strings.Contains(s, "ماتركس") {
		s = strings.ReplaceAll(s, "ماتركس", m)
	}
	if i := AgentName("IRIS"); i != "ايريس" {
		s = strings.ReplaceAll(strings.ReplaceAll(s, "إيريس", i), "ايريس", i)
	}
	return s
}
