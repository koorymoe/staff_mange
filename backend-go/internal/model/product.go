package model

import (
	"strings"
	"time"
)

// توفر المنتج: موجود بمخزن الشركة، أو ينطلب من المجهز وقت الحاجة.
const (
	ProductInStock  = "IN_STOCK"  // متوفر داخل الشركة
	ProductOnDemand = "ON_DEMAND" // يطلب عند الحاجة
)

var ProductAvailabilityLabels = map[string]string{
	ProductInStock:  "متوفر داخل الشركة",
	ProductOnDemand: "يطلب عند الحاجة",
}

func ValidProductAvailability(v string) bool {
	_, ok := ProductAvailabilityLabels[v]
	return ok
}

type Product struct {
	ID             string    `db:"id" json:"id"`
	Name           string    `db:"name" json:"name"`
	Unit           *string   `db:"unit" json:"unit"`
	DefaultPrice   *float64  `db:"defaultPrice" json:"defaultPrice"`
	WholesalePrice *float64  `db:"wholesalePrice" json:"wholesalePrice"` // سعر الجملة — لحساب هامش الربح الحقيقي
	ImageBase64    *string   `db:"imageBase64" json:"imageBase64"`
	CreatedAt      time.Time `db:"createdAt" json:"createdAt"`

	// ⚠️ أي عمود تضيفه لجدول Product لازم يكون إله حقل هنا، لأن الاستعلام
	// يستخدم SELECT * — وبدونه sqlx يطيح بـ missing destination name.
	Availability string  `db:"availability" json:"availability"`
	Specs        *string `db:"specs" json:"specs"`  // المواصفات
	Source       *string `db:"source" json:"source"` // المصدر (المجهز/بلد المنشأ)
	ModelName    *string `db:"modelName" json:"modelName"` // الموديل
	ServiceID    *string `db:"serviceId" json:"serviceId"` // تصنيف الموظف — هو المعتمد

	ServiceName       *string `db:"serviceName" json:"serviceName"`
	AvailabilityLabel string  `db:"-" json:"availabilityLabel"`

	// اقتراح النظام: يُحسب من اسم المنتج مقابل أسماء الخدمات ومهاراتها.
	// يبقى اقتراح بس — المعتمد تصنيف الموظف، والاقتراح يظهر إله حتى
	// يوافق عليه أو يختار غيره.
	SuggestedServiceID   *string `db:"-" json:"suggestedServiceId"`
	SuggestedServiceName *string `db:"-" json:"suggestedServiceName"`
}

type CreateProductRequest struct {
	Name           string   `json:"name"`
	Unit           *string  `json:"unit"`
	DefaultPrice   *float64 `json:"defaultPrice"`
	WholesalePrice *float64 `json:"wholesalePrice"`
	ImageBase64    *string  `json:"imageBase64"`
	Availability   *string  `json:"availability"`
	ServiceID      *string  `json:"serviceId"`
	Specs          *string  `json:"specs"`
	Source         *string  `json:"source"`
	ModelName      *string  `json:"modelName"`
}

type UpdateProductRequest struct {
	Name           *string  `json:"name"`
	Unit           *string  `json:"unit"`
	DefaultPrice   *float64 `json:"defaultPrice"`
	WholesalePrice *float64 `json:"wholesalePrice"`
	ImageBase64    *string  `json:"imageBase64"`
	Availability   *string  `json:"availability"`
	ServiceID      *string  `json:"serviceId"`
	Specs          *string  `json:"specs"`
	Source         *string  `json:"source"`
	ModelName      *string  `json:"modelName"`
	// ClearService يخلي الموظف يشيل التصنيف كلياً — بدونها ما نميّز بين
	// «ما غيّر التصنيف» و«يريد يفضّيه».
	ClearService bool `json:"clearService"`
}

// ServiceHint خدمة بأسماء مهاراتها — مدخل مطابقة اقتراح النظام.
type ServiceHint struct {
	ID    string
	Name  string
	Terms []string // اسم الخدمة + أسماء مهاراتها
}

// SuggestServiceFor يخمّن خدمة المنتج من اسمه.
//
// المطابقة بالكلمات مو بالحرف: نقسم اسم المنتج والمصطلح لكلمات، وكل
// كلمة مشتركة تحسب نقطة. المصطلح الي ينطابق كامل ياخذ ترجيح إضافي،
// لأن «كاميرا حرارية» أدل من «كاميرا» لحالها.
//
// النتيجة اقتراح بس — الموظف هو الي يعتمد أو يغيّر.
func SuggestServiceFor(productName string, hints []ServiceHint) (id, name string, ok bool) {
	words := arabicWords(productName)
	if len(words) == 0 {
		return "", "", false
	}
	bestScore := 0
	var best *ServiceHint
	for i := range hints {
		score := 0
		for _, term := range hints[i].Terms {
			tw := arabicWords(term)
			if len(tw) == 0 {
				continue
			}
			hit := 0
			for _, w := range tw {
				if matchesAny(w, words) {
					hit++
				}
			}
			if hit == len(tw) {
				hit += len(tw) // المصطلح انطابق كامل
			}
			if hit > score {
				score = hit
			}
		}
		if score > bestScore {
			bestScore = score
			best = &hints[i]
		}
	}
	if best == nil {
		return "", "", false
	}
	return best.ID, best.Name, true
}

func matchesAny(w string, words []string) bool {
	for _, pw := range words {
		if w == pw {
			return true
		}
		// الاحتواء للكلمات الطويلة بس، حتى ما تنطابق كلمات قصيرة بالغلط
		if len([]rune(w)) >= 4 && strings.Contains(pw, w) {
			return true
		}
		if len([]rune(pw)) >= 4 && strings.Contains(w, pw) {
			return true
		}
	}
	return false
}

// arabicWords يقسم النص لكلمات، يشيل «ال» الملتصقة حتى «الحرارة» تطابق
// «حرارة»، ويهمل الكلمات القصيرة (حروف الجر وأدوات الربط).
func arabicWords(s string) []string {
	out := []string{}
	for _, f := range strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == '-' || r == '/' || r == '،' || r == ','
	}) {
		f = strings.TrimPrefix(f, "ال")
		if len([]rune(f)) >= 3 {
			out = append(out, f)
		}
	}
	return out
}
