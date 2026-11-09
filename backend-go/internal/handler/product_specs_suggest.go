package handler

import (
	"encoding/json"
	"net/http"
	"strings"
)

// ProductSpecsSuggest (دازاي 11-06): ماتركس يقترح مواصفات منتج من اسمه وموديله.
// المواصفات مو بيانات شخصية، والموظف يراجعها قبل الحفظ.
type ProductSpecsSuggest struct {
	Ask func(system, task string) (string, error)
}

const specsSuggestSystem = `أنت تكتب مواصفات تقنية لمنتج (كاميرات، شبكات، أجهزة أمنية...).
اكتب كل مواصفة بسطر بصيغة «العنوان: القيمة»، بالعربي مع المصطلحات التقنية كما هي.
اكتب فقط الي تعرفه بثقة عن هذا الموديل بالضبط. إذا ما تعرف الموديل ارجع كلمة «ماعرف» بس. لا تخترع.
بلا مقدمة وبلا شرح، أسطر المواصفات فقط، بحد أقصى 20 سطر.`

const specsMaxLines = 20

// CleanSpecs ينظف جواب الذكاء: يشيل التعداد والأسطر الفاضية والمقدمات، ويحدّ بـ20 سطر.
func CleanSpecs(raw string) string {
	out := []string{}
	for _, ln := range strings.Split(raw, "\n") {
		ln = strings.TrimSpace(strings.Trim(strings.TrimSpace(ln), "-*•`#"))
		ln = strings.TrimSpace(strings.ReplaceAll(ln, "**", ""))
		i := strings.IndexAny(ln, ":：")
		if i <= 0 || strings.TrimSpace(ln[i+1:]) == "" { // لازم «عنوان: قيمة» — المقدمة «تفضل:» تنشال
			continue
		}
		out = append(out, ln)
		if len(out) == specsMaxLines {
			break
		}
	}
	return strings.Join(out, "\n")
}

func (h *ProductSpecsSuggest) Suggest(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name      string `json:"name"`
		ModelName string `json:"modelName"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غلط")
		return
	}
	in.Name, in.ModelName = strings.TrimSpace(in.Name), strings.TrimSpace(in.ModelName)
	if in.Name == "" && in.ModelName == "" {
		WriteError(w, http.StatusBadRequest, "اكتب اسم المنتج أو الموديل")
		return
	}
	if h.Ask == nil {
		WriteError(w, http.StatusServiceUnavailable, "الذكاء مو مفعّل")
		return
	}
	ans, err := h.Ask(specsSuggestSystem, "المنتج: "+in.Name+"\nالموديل: "+in.ModelName)
	if err != nil {
		WriteError(w, http.StatusBadGateway, "ماتركس ما رد، جرب بعدين")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"specs": CleanSpecs(ans)})
}
