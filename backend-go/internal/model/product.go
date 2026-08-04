package model

import "time"

// حاجة المنتج: هل نحتاج نوفّره أو لا.
//
// القيم بقت مثل ما هي (IN_STOCK/ON_DEMAND) حتى ما تنكسر البيانات
// الموجودة، بس معناها صار «الحاجة» مو «التوفر»:
//   ON_DEMAND = يحتاج نوفّره   |   IN_STOCK = ما يحتاج
const (
	ProductInStock  = "IN_STOCK"  // ما يحتاج توفير
	ProductOnDemand = "ON_DEMAND" // يحتاج توفير
)

var ProductAvailabilityLabels = map[string]string{
	ProductInStock:  "ما يحتاج توفير",
	ProductOnDemand: "يحتاج توفير",
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
	ServiceID    *string `db:"serviceId" json:"serviceId"`     // تصنيف الموظف — هو المعتمد
	ServiceText  *string `db:"serviceText" json:"serviceText"` // الخدمة كما كتبها التقني
	CreatedByID  *string `db:"createdById" json:"createdById"`

	ServiceName       *string `db:"serviceName" json:"serviceName"`
	CreatedByName     *string `db:"createdByName" json:"createdByName"`
	AvailabilityLabel string  `db:"-" json:"availabilityLabel"`

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
	ServiceText    *string  `json:"serviceText"`
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
	ServiceText    *string  `json:"serviceText"`
	// ClearService يخلي الموظف يشيل التصنيف كلياً — بدونها ما نميّز بين
	// «ما غيّر التصنيف» و«يريد يفضّيه».
	ClearService bool `json:"clearService"`
}
