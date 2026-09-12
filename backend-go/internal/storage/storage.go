// Package storage يخزّن الملفات (صور، وثائق، وصولات) برّا قاعدة البيانات.
//
// ليش؟ كل الصور كانت تنخزن base64 داخل أعمدة نصية بالقاعدة. صورة وحدة
// بميغا تصير ~1.4 ميغا نص، وتنسحب مع كل استعلام يجيب الصف — حتى لو
// الشاشة ما تعرض الصورة. النتيجة: قاعدة تنفخ، نسخ احتياطية ثقيلة،
// واستعلامات بطيئة بلا سبب.
//
// الحل: نخزّن الملف برّا ونحتفظ بالمسار بس. باكند‌ين:
//
//   - R2 (أو أي تخزين متوافق مع S3): إذا انضبطت متغيرات البيئة
//   - القرص المحلي: الافتراضي — يشتغل بلا أي إعداد ولا حساب خارجي
//
// الاثنين ينفّذون نفس الواجهة، والباقي بالنظام ما يعرف أيهم شغّال.
package storage

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"strings"
	"time"
)

// MaxFileBytes أكبر ملف مسموح — 10 ميغا.
const MaxFileBytes = 10 << 20

// Store واجهة التخزين. Put يرجّع المفتاح، وGet يرجّع المحتوى ونوعه.
type Store interface {
	Put(ctx context.Context, key string, data []byte, contentType string) error
	Get(ctx context.Context, key string) (data []byte, contentType string, err error)
	Delete(ctx context.Context, key string) error
	// Kind وصف قصير للباكند — يظهر بسجل الإقلاع حتى نعرف وين ينخزن
	Kind() string
}

// ErrNotFound الملف مو موجود.
var ErrNotFound = errors.New("الملف غير موجود")

// NewKey يبني مفتاح فريد تحت مجلد منطقي (products/, receipts/, ...).
//
// المفتاح عشوائي مو مشتق من اسم الملف: أسماء الملفات من المستخدم ممكن
// تحمل مسارات (../) أو تكشف معلومات، والعشوائي يمنع الاثنين.
func NewKey(folder, contentType string) string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	ext := extensionFor(contentType)
	folder = strings.Trim(strings.TrimSpace(folder), "/")
	if folder == "" {
		folder = "misc"
	}
	return fmt.Sprintf("%s/%s%s", folder, hex.EncodeToString(b[:]), ext)
}

func extensionFor(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "application/pdf":
		return ".pdf"
	}
	if exts, err := mime.ExtensionsByType(contentType); err == nil && len(exts) > 0 {
		return exts[0]
	}
	return ".bin"
}

// AllowedContentTypes الأنواع المسموح رفعها. القائمة بيضاء مو سوداء:
// نسمح للي نعرفه بس، بدل ما نحاول نعدّ كل شي خطر.
var AllowedContentTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/webp":      true,
	"application/pdf": true,
}

// DecodeDataURL يفك سلسلة data:...;base64,... ويرجّع المحتوى ونوعه.
//
// نحتاجها بمكانين: ترحيل البيانات القديمة، وقبول رفع من واجهات لسّه
// ترسل base64.
func DecodeDataURL(s string) (data []byte, contentType string, err error) {
	if !strings.HasPrefix(s, "data:") {
		return nil, "", errors.New("مو data URL")
	}
	comma := strings.IndexByte(s, ',')
	if comma < 0 {
		return nil, "", errors.New("data URL ناقص")
	}
	meta := s[len("data:"):comma]
	payload := s[comma+1:]
	if !strings.HasSuffix(meta, ";base64") {
		return nil, "", errors.New("data URL مو base64")
	}
	contentType = strings.TrimSuffix(meta, ";base64")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	data, err = base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, "", fmt.Errorf("base64 غير صالح: %w", err)
	}
	return data, contentType, nil
}

// SniffContentType يحدد النوع من محتوى الملف نفسه مو من ما يدّعيه
// المستخدم — الترويسة المرسلة تنزوّر بسهولة.
func SniffContentType(data []byte) string {
	if len(data) >= 4 && bytes.Equal(data[:4], []byte("%PDF")) {
		return "application/pdf"
	}
	if len(data) >= 8 && bytes.Equal(data[:8], []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}) {
		return "image/png"
	}
	if len(data) >= 3 && bytes.Equal(data[:3], []byte{0xFF, 0xD8, 0xFF}) {
		return "image/jpeg"
	}
	if len(data) >= 12 && bytes.Equal(data[:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")) {
		return "image/webp"
	}
	return ""
}

// readAllLimited يقرأ بحد أقصى، ويطيح لو تجاوزه — بدل ما نثق
// بـContent-Length الي يرسله العميل.
func readAllLimited(r io.Reader, max int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > max {
		return nil, fmt.Errorf("الملف أكبر من الحد المسموح (%d ميغا)", max>>20)
	}
	return data, nil
}

// ctxTimeout مهلة موحّدة لعمليات التخزين البعيد.
func ctxTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, 30*time.Second)
}

// ReadLimited يقرأ محتوى برفع مع حد أقصى — تُستخدم من طبقة المعالجات.
func ReadLimited(r io.Reader, max int64) ([]byte, error) { return readAllLimited(r, max) }
