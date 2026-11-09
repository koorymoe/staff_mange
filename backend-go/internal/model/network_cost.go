package model

import "time"

// ═══ تسعيرة الشبكات ═══
//
// فقرة وحدة بقائمة أسعار الشبكات. الأسعار تنخزن بالجدول مو بالكود
// لأنها لسه تتبني وصاحب العمل يعدّلها بنفسه.
type NetworkPriceItem struct {
	ID          string  `db:"id" json:"id"`
	Label       string  `db:"label" json:"label"`
	Unit        string  `db:"unit" json:"unit"`
	PricingMode string  `db:"pricingMode" json:"pricingMode"` // FLAT | TIERED
	BasePrice   float64 `db:"basePrice" json:"basePrice"`
	IncludedQty float64 `db:"includedQty" json:"includedQty"`
	// ExtraPerUnit سعر كل وحدة بعد الكمية المشمولة (TIERED بس)
	ExtraPerUnit float64 `db:"extraPerUnit" json:"extraPerUnit"`
	// BracketsJSON عمود jsonb — ما ينقرا مباشرة بالواجهة، ينفكّ لـBrackets
	BracketsJSON []byte           `db:"brackets" json:"-"`
	Brackets     []NetworkBracket `db:"-" json:"brackets"`
	Note         *string          `db:"note" json:"note"`
	SortOrder    int              `db:"sortOrder" json:"sortOrder"`
	Active       bool             `db:"active" json:"active"`
	UpdatedAt    time.Time        `db:"updatedAt" json:"updatedAt"`
	CreatedAt    time.Time        `db:"createdAt" json:"createdAt"`
}

// NetworkBracket شريحة تسعيرة جملة: كل كمية لحد UpTo سعر وحدتها
// UnitPrice. آخر شريحة UpTo=0 يعني «وأكثر».
type NetworkBracket struct {
	UpTo      float64 `json:"upTo"`
	UnitPrice float64 `json:"unitPrice"`
}

const (
	NetworkPricingFlat    = "FLAT"
	NetworkPricingTiered  = "TIERED"
	NetworkPricingBracket = "BRACKET"
)

// SaveNetworkPriceItemRequest إضافة أو تعديل فقرة تسعيرة.
type SaveNetworkPriceItemRequest struct {
	Label        string  `json:"label"`
	Unit         string  `json:"unit"`
	PricingMode  string  `json:"pricingMode"`
	BasePrice    float64          `json:"basePrice"`
	IncludedQty  float64          `json:"includedQty"`
	ExtraPerUnit float64          `json:"extraPerUnit"`
	Brackets     []NetworkBracket `json:"brackets"`
	Note         *string          `json:"note"`
	SortOrder    int              `json:"sortOrder"`
	Active       *bool            `json:"active"`
}

// NetworkCostLine سطر بالاستمارة: فقرة من القائمة + كمية.
type NetworkCostLine struct {
	ItemID   string  `json:"itemId"`
	Quantity float64 `json:"quantity"`
}

type NetworkCostRequest struct {
	Lines    []NetworkCostLine `json:"lines"`
	Discount float64           `json:"discount"`
}

// NetworkCostLineResult تفصيل السطر — نرجّع المقطوعة والزيادة كل وحدة
// على حدة حتى الزبون يشوف ليش طلع الرقم، مو رقم نهائي بس.
type NetworkCostLineResult struct {
	ItemID    string  `json:"itemId"`
	Label     string  `json:"label"`
	Unit      string  `json:"unit"`
	Quantity  float64 `json:"quantity"`
	BasePart  float64 `json:"basePart"`
	ExtraQty  float64 `json:"extraQty"`
	ExtraPart float64 `json:"extraPart"`
	// UnitPrice سعر الوحدة الي انطبّق فعلاً (BRACKET: سعر الشريحة)
	UnitPrice float64 `json:"unitPrice"`
	Total     float64 `json:"total"`
}

type NetworkCostResponse struct {
	Lines       []NetworkCostLineResult `json:"lines"`
	Subtotal    float64                 `json:"subtotal"`
	Discount    float64                 `json:"discount"`
	FinalAmount float64                 `json:"finalAmount"`
}

// CalculateLine يحسب سطر واحد.
//
// TIERED: المبلغ المقطوع ينحسب كامل حتى لو الكمية أقل من المشمولة —
// هذا مقصود، «لحد ٢٠ متر ١٢ ألف» يعني ١٢ ألف حتى لو الشغلة ٥ أمتار،
// لأن الطلعة والوقت نفسهم.
func (it NetworkPriceItem) CalculateLine(quantity float64) NetworkCostLineResult {
	res := NetworkCostLineResult{
		ItemID: it.ID, Label: it.Label, Unit: it.Unit, Quantity: quantity,
	}
	if quantity <= 0 {
		return res
	}
	switch it.PricingMode {
	case NetworkPricingTiered:
		res.BasePart = it.BasePrice
		if quantity > it.IncludedQty {
			res.ExtraQty = quantity - it.IncludedQty
			res.ExtraPart = res.ExtraQty * it.ExtraPerUnit
		}
	case NetworkPricingBracket:
		res.UnitPrice = it.bracketUnitPrice(quantity)
		res.BasePart = res.UnitPrice * quantity
	default: // FLAT
		res.UnitPrice = it.BasePrice
		res.BasePart = it.BasePrice * quantity
	}
	res.Total = res.BasePart + res.ExtraPart
	return res
}

// bracketUnitPrice يختار سعر الوحدة حسب الشريحة الي توقع بيها الكمية.
//
// الشرائح مرتّبة تصاعدي وآخر وحدة UpTo=0 يعني «وأكثر». إذا ماكو
// شريحة مفتوحة والكمية أكبر من كل الشرائح، نستعمل سعر آخر شريحة —
// أرخص شي لو ننطي صفر بالسكوت ونطلع فاتورة بلا مبلغ.
func (it NetworkPriceItem) bracketUnitPrice(quantity float64) float64 {
	last := 0.0
	for _, b := range it.Brackets {
		last = b.UnitPrice
		if b.UpTo <= 0 || quantity <= b.UpTo {
			return b.UnitPrice
		}
	}
	return last
}
