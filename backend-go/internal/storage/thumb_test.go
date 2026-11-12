package storage

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

// (11-12) الصورة الكبيرة تتصغّر لـJPEG بالعرض المطلوب والطول بالنسبة، والصغيرة ما تتصغّر.
func TestMakeThumb(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 1200, 600))
	for x := 0; x < 1200; x++ {
		src.Set(x, x%600, color.NRGBA{200, 10, 10, 255})
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, src); err != nil {
		t.Fatal(err)
	}
	out, err := MakeThumb(buf.Bytes(), 320)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := jpeg.DecodeConfig(bytes.NewReader(out))
	if err != nil || cfg.Width != 320 || cfg.Height != 160 {
		t.Fatalf("الأبعاد: %+v %v", cfg, err)
	}
	if _, err := MakeThumb(buf.Bytes(), 1600); err != ErrThumbSkip {
		t.Fatal("الأصغر من العرض ما يتصغّر")
	}
	if ThumbKey("designs/a.png", 320) != "thumbs/w320/designs/a.png.jpg" || ThumbWidths[333] {
		t.Fatal("المفتاح/العروض")
	}
}
