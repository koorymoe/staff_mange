package core

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// peekLimit حد قراءة الجسم للفحص — أكبر من أي طلب JSON عادي، وأصغر من رفع ملف.
const peekLimit = 1 << 20

// BodyField يقرا حقل نصي من جسم JSON **ويرجّع الجسم مثل ما چان** للمعالج.
// للقواعد الي الشي بيها مذكور بالجسم مو بالرابط (مثل bookingId). جسم مو JSON أو
// أكبر من الحد = "" (والقاعدة ترفض — يضيّق ما يوسّع).
func BodyField(r *http.Request, field string) string {
	if r.Body == nil {
		return ""
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, peekLimit+1))
	r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(raw), r.Body))
	if err != nil || len(raw) > peekLimit {
		return ""
	}
	var m map[string]json.RawMessage
	if json.Unmarshal(raw, &m) != nil {
		return ""
	}
	var s string
	if json.Unmarshal(m[field], &s) != nil {
		return ""
	}
	return s
}
