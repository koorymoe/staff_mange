package storage

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// وسم وصول قصير العمر للملفات.
//
// المشكلة: الملفات تنعرض داخل <img src="/api/files/..."> والمتصفح ما
// يرسل ترويسة Authorization مع الصور — فالتحقق العادي بالتوكن يرجّع
// 401 وكل الصور تنكسر.
//
// الحل: وسم موقّع قصير العمر ينضاف للرابط. ما نحط توكن الدخول نفسه
// بالرابط لأن الروابط تنحفظ بسجلات السيرفر وترويسة Referer — والوسم
// هذا صلاحيته دقائق ويخص الملفات بس، فلو تسرّب ما ينفع لشي ثاني.
const fileTokenTTL = 15 * time.Minute

// NewFileToken يوقّع وسم صالح لمدة قصيرة.
func NewFileToken(secret []byte) string {
	exp := time.Now().Add(fileTokenTTL).Unix()
	payload := strconv.FormatInt(exp, 10)
	return payload + "." + sign(secret, payload)
}

// VerifyFileToken يتأكد من التوقيع ومن إنه ما انتهى.
func VerifyFileToken(secret []byte, token string) error {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return errors.New("وسم غير صالح")
	}
	// hmac.Equal مقارنة ثابتة الزمن — المقارنة العادية تسرّب التوقيع
	// حرف حرف عبر فروقات التوقيت.
	if !hmac.Equal([]byte(sign(secret, parts[0])), []byte(parts[1])) {
		return errors.New("توقيع غير صالح")
	}
	exp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return fmt.Errorf("وسم غير صالح: %w", err)
	}
	if time.Now().Unix() > exp {
		return errors.New("انتهت صلاحية الوسم")
	}
	return nil
}

func sign(secret []byte, payload string) string {
	m := hmac.New(sha256.New, secret)
	m.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}
