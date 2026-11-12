package storage

import (
	"bytes"
	"errors"
	"image"
	_ "image/gif" // يسجّل فك GIF
	"image/jpeg"
	_ "image/png" // يسجّل فك PNG

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp" // يسجّل فك WebP
)

// ═══ 🖼️ صور مصغّرة (دازاي 11-12: «الثقل بتحميل الصور كلش كلش») ═══
// المخططات المرفوعة مكبّرة (upscayl 4x) — الوحدة عشرات الميگات، والشاشة تعرضها بمربع صغير.
// هسه الشاشة تطلب ?w=… والسيرفر يرجّع JPEG مصغّر (ينحفظ مرة وحدة بـthumbs/)، والأصلية تنفتح بالضغط.

// ThumbWidths العروض المسموحة (حتى ما ينطلب كل رقم وينملأ التخزين).
var ThumbWidths = map[int]bool{160: true, 320: true, 480: true, 800: true}

// thumbMaxPixels فوگه ما نفك الصورة (ذاكرة السيرفر ٤ GB): نرجّع الأصلية.
const thumbMaxPixels = 80_000_000

// قفل: صورتين بنفس الوكت كحد أقصى — فك صورة ضخمة ياكل مئات الميگات رام.
var thumbSem = make(chan struct{}, 2)

var ErrThumbSkip = errors.New("ما تنصغّر")

// ThumbKey مفتاح النسخة المصغّرة بالتخزين.
func ThumbKey(key string, w int) string {
	return "thumbs/w" + itoa(w) + "/" + key + ".jpg"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	b := []byte{}
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// IsThumbable نوع صورة نگدر نصغّره.
func IsThumbable(contentType string) bool {
	switch contentType {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		return true
	}
	return false
}

// MakeThumb يصغّر الصورة لعرض w (الطول بالنسبة) ويرجّع JPEG. الصورة الأصغر من w ترجع
// ErrThumbSkip (الأصلية أخف أصلاً).
func MakeThumb(data []byte, w int) ([]byte, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if cfg.Width <= w || cfg.Width*cfg.Height > thumbMaxPixels {
		return nil, ErrThumbSkip
	}
	thumbSem <- struct{}{}
	defer func() { <-thumbSem }()
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	h := cfg.Height * w / cfg.Width
	if h < 1 {
		h = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	// خلفية بيضاء للـPNG الشفاف (JPEG ما يدعم الشفافية)
	draw.Draw(dst, dst.Bounds(), image.White, image.Point{}, draw.Src)
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	var out bytes.Buffer
	if err := jpeg.Encode(&out, dst, &jpeg.Options{Quality: 80}); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
