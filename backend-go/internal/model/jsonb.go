package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

// ═══ أعمدة JSONB ═══
//
// sqlx ما يعرف يحط عمود jsonb بحقل Go لحاله — يرجّع []byte ويوقف. وأي
// حقل jsonb بلا نوع يعرف يستقبله يكسر الاستعلام كله بالسكوت (نفس الفخ
// الي وكعنا بيه سابقاً بعمود assignedById).
//
// هذولا نوعين يعرفون يقرون ويكتبون:
//   JSONRaw — نمرّر الـ JSON مثل ما هو للواجهة بلا ما نفهمه (سطور
//             التسليك والحديد: الواجهة تعرف شكلهن وإحنا نخزنهن بس)
//   JSONMap — قاموس مفاتيح/قيم (مواصفات المكوّن: Vmp، Voc، الكفاءة...)

// JSONRaw عمود jsonb ينمرّر كما هو.
type JSONRaw json.RawMessage

func (j JSONRaw) Value() (driver.Value, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return []byte(j), nil
}

func (j *JSONRaw) Scan(src any) error {
	if src == nil {
		*j = JSONRaw("null")
		return nil
	}
	switch v := src.(type) {
	case []byte:
		// نسخة خاصة بينا: الدرايفر يعيد استعمال البفر بين الصفوف، فلو
		// خزّنّا الشريحة نفسها تنقلب بياناتنا مع الصف الجاي.
		b := make([]byte, len(v))
		copy(b, v)
		*j = JSONRaw(b)
		return nil
	case string:
		*j = JSONRaw(v)
		return nil
	}
	return fmt.Errorf("JSONRaw: نوع غير متوقع %T", src)
}

func (j JSONRaw) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

func (j *JSONRaw) UnmarshalJSON(b []byte) error {
	if j == nil {
		return errors.New("JSONRaw: مؤشر فارغ")
	}
	*j = append((*j)[0:0], b...)
	return nil
}

// JSONMap عمود jsonb على شكل قاموس.
type JSONMap map[string]any

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(map[string]any(m))
}

func (m *JSONMap) Scan(src any) error {
	if src == nil {
		*m = JSONMap{}
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("JSONMap: نوع غير متوقع %T", src)
	}
	out := map[string]any{}
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	*m = out
	return nil
}
