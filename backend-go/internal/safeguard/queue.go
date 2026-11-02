package safeguard

import (
	"log"
	"os"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// ═══ 🚦 طوابير المهام الخلفية (دازاي 11-02: «الضغط والتعليك والتأخير») ═══
//
// قبل: ٣٥ مهمة خلفية (تقييم ماتركس، التحليل، التقارير، الكنسات…) كل وحدة
// تشتغل بوقتها بلا تنسيق، ويا طلبات الموظفين على نفس مستودع اتصالات القاعدة.
// لمن تتصادف ثلاث أربع مهام ثگيلة بنفس الدقيقة، طلبات الموظفين تنطر دورها
// بالقاعدة — وهذا «التعليك».
//
// هسه كل دورة تدخل طابور:
//   - طابور خفيف (المهام الي تتكرر بأقل من ١٠ دقايق): عمّالين.
//   - طابور ثگيل (الباقي): عمّال واحد افتراضياً — وحدة بعد وحدة.
//   - **الأولوية للموظفين**: المهمة الثگيلة تنطر لحد ما تخف الطلبات الشغالة
//     (أكثر من حد معيّن = السيرفر مشغول بالناس). وحتى ما تنحبس للأبد،
//     بعد دقيقتين انتظار تشتغل على كل حال.
//
// العدد قابل للتعديل بمتغيرات البيئة: BG_HEAVY_WORKERS و BG_LIGHT_WORKERS
// و BG_YIELD_INFLIGHT.

func envInt(k string, def int) int {
	if v, err := strconv.Atoi(os.Getenv(k)); err == nil && v > 0 {
		return v
	}
	return def
}

var (
	heavySlots    = make(chan struct{}, envInt("BG_HEAVY_WORKERS", 1))
	lightSlots    = make(chan struct{}, envInt("BG_LIGHT_WORKERS", 2))
	yieldInFlight = int64(envInt("BG_YIELD_INFLIGHT", 6))
	inFlight      int64
)

const (
	lightEvery   = 10 * time.Minute
	maxYieldWait = 2 * time.Minute
	slowJobLog   = 20 * time.Second
)

// RequestStarted / RequestDone — طلب موظف بدا/خلص (من middleware.Metrics).
func RequestStarted() { atomic.AddInt64(&inFlight, 1) }
func RequestDone()    { atomic.AddInt64(&inFlight, -1) }

// InFlight عدد طلبات الموظفين الشغالة هسه.
func InFlight() int64 { return atomic.LoadInt64(&inFlight) }

// waitForQuiet ينطر لحد ما طلبات الموظفين تخف، بحد أقصى.
func waitForQuiet(max time.Duration) {
	deadline := time.Now().Add(max)
	for InFlight() > yieldInFlight && time.Now().Before(deadline) {
		time.Sleep(300 * time.Millisecond)
	}
}

// JobStat حالة مهمة خلفية — تطلع بلوحة الأمان للمالك.
type JobStat struct {
	Name         string    `json:"name"`
	Lane         string    `json:"lane"`
	Runs         int64     `json:"runs"`
	Failures     int64     `json:"failures"`
	Running      bool      `json:"running"`
	Queued       bool      `json:"queued"`
	LastStart    time.Time `json:"lastStart"`
	LastMs       int64     `json:"lastMs"`
	LastWaitMs   int64     `json:"lastWaitMs"`
	MaxMs        int64     `json:"maxMs"`
	EverySeconds int64     `json:"everySeconds"`
}

var (
	statsMu sync.Mutex
	stats   = map[string]*JobStat{}
)

func stat(name string, f func(*JobStat)) {
	statsMu.Lock()
	defer statsMu.Unlock()
	s, ok := stats[name]
	if !ok {
		s = &JobStat{Name: name}
		stats[name] = s
	}
	f(s)
}

// Jobs لقطة من كل المهام الخلفية — الأبطأ أول.
func Jobs() []JobStat {
	statsMu.Lock()
	defer statsMu.Unlock()
	out := make([]JobStat, 0, len(stats))
	for _, s := range stats {
		out = append(out, *s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].MaxMs > out[j].MaxMs })
	return out
}

// runQueued دورة وحدة عبر الطابور.
func runQueued(name string, every time.Duration, fn func()) {
	slots, lane := heavySlots, "heavy"
	if every < lightEvery {
		slots, lane = lightSlots, "light"
	}
	queuedAt := time.Now()
	stat(name, func(s *JobStat) { s.Lane, s.Queued, s.EverySeconds = lane, true, int64(every.Seconds()) })
	slots <- struct{}{}
	if lane == "heavy" {
		waitForQuiet(maxYieldWait)
	}
	start := time.Now()
	stat(name, func(s *JobStat) {
		s.Queued, s.Running, s.LastStart = false, true, start
		s.LastWaitMs = start.Sub(queuedAt).Milliseconds()
	})
	ok := Run(name, fn)
	<-slots
	d := time.Since(start)
	stat(name, func(s *JobStat) {
		s.Running = false
		s.Runs++
		if !ok {
			s.Failures++
		}
		s.LastMs = d.Milliseconds()
		if s.LastMs > s.MaxMs {
			s.MaxMs = s.LastMs
		}
	})
	if d > slowJobLog {
		log.Printf("[queue] «%s» (%s) أخذت %s — انتظرت بالطابور %s", name, lane, d.Round(time.Second), start.Sub(queuedAt).Round(time.Second))
	}
}
