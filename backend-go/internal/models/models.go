// Package models — 🏴‍☠️ ون بيس: **الموديل هو مصدر الحقيقة لقاعدة البيانات** (مثل جانغو).
//
// كل جدول إله ملف بـinternal/models (t_<اسم>.go): أعمدته بأنواعها، والقيود، والفهارس.
// الشغل من هسه:
//   - تريد عمود/جدول/فهرس جديد؟ ضيفه **هنا** بالموديل.
//   - `go run ./cmd/makemigration` يقارن الموديلات ويا القاعدة ويكتب التحديث لحاله
//     (internal/database/schema_auto.go). الحذف ما يتولّد أبداً — يطلع تنبيه وينتظر موافقة.
//   - الاختبار (models drift) يفشل إذا القاعدة والموديلات مو متطابقين بأي شي.
//
// البداية (11-10) انولدت من القاعدة نفسها بـ`go run ./cmd/modelgen` — مرة وحدة.
package models

import (
	"sort"
	"sync"
)

// Column عمود: النوع بصيغة Postgres نفسها (format_type)، فالمقارنة حرفية.
type Column struct {
	Name    string
	Type    string
	NotNull bool
	Default string // تعبير SQL كما هو ("" = بلا افتراضي)
}

// Constraint قيد بالاسم وتعريفه (pg_get_constraintdef): PRIMARY KEY / UNIQUE / FOREIGN KEY / CHECK.
type Constraint struct {
	Name string
	Def  string
}

// Table موديل جدول.
type Table struct {
	Name        string
	Columns     []Column
	Constraints []Constraint
	Indexes     []string // تعريف كامل (CREATE INDEX …) — بس الي مو وراها قيد
}

// Enum نوع خاص (CREATE TYPE … AS ENUM).
type Enum struct {
	Name   string
	Values []string
}

// Object شي بالقاعدة غير الجداول: إضافة، تسلسل، دالة، view، trigger — بتعريفه الكامل.
// Key يميّزه (الدالة باسمها ومعاملاتها، الـtrigger بجدوله واسمه).
type Object struct {
	Kind string // ext | seq | fn | view | trg
	Key  string
	Def  string
}

var (
	mu      sync.Mutex
	tables  = map[string]Table{}
	enums   = map[string]Enum{}
	objects = map[string]Object{}
)

func registerObject(o Object) {
	mu.Lock()
	defer mu.Unlock()
	objects[o.Kind+":"+o.Key] = o
}

// Objects كل الأشياء غير الجداول مرتبة.
func Objects() []Object {
	mu.Lock()
	defer mu.Unlock()
	out := make([]Object, 0, len(objects))
	for _, o := range objects {
		out = append(out, o)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Kind+out[i].Key < out[j].Kind+out[j].Key })
	return out
}

func register(t Table) {
	mu.Lock()
	defer mu.Unlock()
	if _, dup := tables[t.Name]; dup {
		panic("models: جدول مكرر " + t.Name)
	}
	tables[t.Name] = t
}

func registerEnum(e Enum) {
	mu.Lock()
	defer mu.Unlock()
	enums[e.Name] = e
}

// Tables كل الموديلات مرتبة بالاسم.
func Tables() []Table {
	mu.Lock()
	defer mu.Unlock()
	out := make([]Table, 0, len(tables))
	for _, t := range tables {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Get موديل جدول بالاسم.
func Get(name string) (Table, bool) {
	mu.Lock()
	defer mu.Unlock()
	t, ok := tables[name]
	return t, ok
}

// Enums كل الأنواع الخاصة مرتبة.
func Enums() []Enum {
	mu.Lock()
	defer mu.Unlock()
	out := make([]Enum, 0, len(enums))
	for _, e := range enums {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// ColumnNames أسماء أعمدة كل الموديلات (لفحص «عمود بلا موديل»).
func ColumnNames() map[string]bool {
	out := map[string]bool{}
	for _, t := range Tables() {
		for _, c := range t.Columns {
			out[c.Name] = true
		}
	}
	return out
}
