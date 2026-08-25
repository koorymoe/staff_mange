package service

import (
	"fmt"

	"staffmange-api/internal/model"
)

// ── محرك شيت "حساب تكلفة التنفيذ" (استمارة كاميرات المراقبة) ──
//
// هذا شيت مستقل عن "تكاليف المشروع" وله معادلة مختلفة تماماً: سعر أساس مأخوذ
// من شريحة طول الكيبل لكل كاميرا، ثم ثلاث طبقات ضرب متتالية (نوع المكان، نوع
// المنظومة، ارتفاع الكاميرا)، ثم تنجمع الأعمال الإضافية وينطرح الخصم.
//
// المعادلات منقولة حرفياً من خلايا J/K/L/M و D18 بالشيت.

// أسعار الأعمال الإضافية — من معادلة D18 بالشيت.
const (
	extraScreenLarge43Price  = 15000 // تثبيت شاشة حجم 43 واكبر
	extraScreenSmall43Price  = 7500  // تثبيت شاشة اصغر من 43
	extraRackPrice           = 15000 // تثبيت راك
	extraBoardPrice          = 7500  // تثبيت بورد
	extraIpCameraChangePrice = 15000 // تغيير أو إضافة IP كاميرا
	extraVipInternetPerM     = 400   // مد كيبل انترنيت بالمتر VIP
	extraNormalInternetPerM  = 200   // مد كيبل انترنيت بالمتر عادي
)

// CameraCostNote الملاحظة المكتوبة بالشيت نفسه (B17) — تنعرض بالواجهة وبالطباعة.
const CameraCostNote = "في حالة هناك اختلاف بين البيانات المدخلة والعمل الفعلي يضاف أجور العمل الاضافية الى الفاتورة"

// CameraPlaceTypes / CameraSystemTypes القوائم المنسدلة المسموحة (من data validation
// بالشيت) — الواجهة تستعملها حتى ما تختلف الكتابة عن المعادلة.
var CameraPlaceTypes = []string{"منزل سكني", "محل تجاري", "مدرسة او شركة", "مصنع او معمل"}
var CameraSystemTypes = []string{"ANLOGE", "IP"}

// cableTierPrice شريحة السعر حسب طول الكيبل — من معادلة J بالشيت:
// صفر متر = 7500، وأقل من 10 = 10000 ... و49 فما فوق = 17000.
// ملاحظة مهمة: "ما مدخل شي" غير "صفر" — الفاضي يعطي 0 والصفر يعطي 7500.
func cableTierPrice(meters float64, entered bool) float64 {
	if !entered {
		return 0
	}
	switch {
	case meters == 0:
		return 7500
	case meters < 10:
		return 10000
	case meters < 20:
		return 12000
	case meters < 30:
		return 13000
	case meters < 40:
		return 14000
	case meters < 50:
		return 15000
	default: // >= 50
		return 17000
	}
}

// cameraPlaceMultiplier مضاعف نوع المكان — من معادلة K بالشيت.
func cameraPlaceMultiplier(placeType string) float64 {
	switch placeType {
	case "منزل سكني":
		return 1
	case "محل تجاري":
		return 0.95
	case "مدرسة او شركة":
		return 1.3
	case "مصنع او معمل":
		return 1.4
	default:
		return 1
	}
}

// cameraSystemMultiplier مضاعف نوع المنظومة — من معادلة L بالشيت.
func cameraSystemMultiplier(systemType string) float64 {
	if systemType == "IP" {
		return 1.2
	}
	return 1 // ANLOGE
}

// cameraHeightMultiplier مضاعف ارتفاع الكاميرا — من معادلة M بالشيت.
func cameraHeightMultiplier(above3m bool) float64 {
	if above3m {
		return 1.1
	}
	return 1
}

// CalculateCameraCost ينفّذ استمارة "حساب تكلفة التنفيذ" كاملة.
func CalculateCameraCost(req model.CameraCostRequest) (*model.CameraCostResponse, error) {
	if len(req.Rows) == 0 {
		return nil, fmt.Errorf("لا يمكن الحساب بدون صف كاميرا واحد على الأقل")
	}
	if req.PlaceType == "" {
		return nil, fmt.Errorf("نوع المكان مطلوب")
	}
	if req.SystemType == "" {
		return nil, fmt.Errorf("نوع المنظومة مطلوب")
	}

	placeMult := cameraPlaceMultiplier(req.PlaceType)
	sysMult := cameraSystemMultiplier(req.SystemType)

	resp := &model.CameraCostResponse{
		Rows: make([]model.CameraCostRowResult, 0, len(req.Rows)),
		Note: CameraCostNote,
	}

	for i, row := range req.Rows {
		// J: شريحة الكيبل العادي + 1.2 × شريحة كيبل الـVIP.
		// "مدخل" يعني القيمة > 0 أو مساوية صفر بشكل صريح — بالواجهة الصف
		// الفاضي تماماً (بلا عادي ولا VIP) ما ينحسب.
		normalEntered := row.NormalCableMeters > 0
		vipEntered := row.VipCableMeters > 0
		if !normalEntered && !vipEntered {
			continue // صف فاضي — مثل ما الشيت يعطيه 0
		}
		base := cableTierPrice(row.NormalCableMeters, normalEntered) +
			1.2*cableTierPrice(row.VipCableMeters, vipEntered)

		afterPlace := base * placeMult
		afterSystem := afterPlace * sysMult
		heightMult := cameraHeightMultiplier(row.HeightAbove3m)
		total := afterSystem * heightMult

		// I: الكاميرا تنعد بالعدد الإجمالي إذا إلها طول كيبل
		counts := row.NormalCableMeters > 0 || row.VipCableMeters > 0
		if counts {
			resp.CameraCount++
		}
		resp.CamerasTotal += total

		resp.Rows = append(resp.Rows, model.CameraCostRowResult{
			Index:            i + 1,
			BasePrice:        base,
			PlaceMultiplier:  placeMult,
			AfterPlace:       afterPlace,
			SystemMultiplier: sysMult,
			AfterSystem:      afterSystem,
			HeightMultiplier: heightMult,
			Total:            total,
			CountsAsCamera:   counts,
		})
	}

	// الأعمال الإضافية — من معادلة D18 بالشيت (سعر ثابت × العدد، والبرمجة/غيرها مبالغ يدوية)
	e := req.Extras
	resp.ExtrasTotal = float64(e.ScreenLarge43Count)*extraScreenLarge43Price +
		float64(e.ScreenSmall43Count)*extraScreenSmall43Price +
		float64(e.RackCount)*extraRackPrice +
		float64(e.BoardCount)*extraBoardPrice +
		float64(e.IpCameraChangeCount)*extraIpCameraChangePrice +
		float64(e.VipInternetMeters)*extraVipInternetPerM +
		float64(e.NormalInternetM)*extraNormalInternetPerM +
		e.ProgrammingAmount + e.OtherAmount

	resp.Discount = req.Discount
	resp.FinalAmount = resp.CamerasTotal + resp.ExtrasTotal - req.Discount
	if resp.FinalAmount < 0 {
		resp.FinalAmount = 0 // الخصم ما يخلي الفاتورة بالسالب
	}
	return resp, nil
}
