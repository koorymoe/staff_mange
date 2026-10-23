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

// NewFileToken يوقّع وسماً **باسم الموظف** صالحاً لمدة قصيرة.
//
// 🔴 (فحص السكيورتي A-02) چان الوسم وقت انتهاء وبس — ما يخص أحداً،
// فالي يحصّله يقرا كل ملفات النظام. هسه بيه رقم الموظف: الخادم يرفضه
// لو الموظف انوقف أو انحذف، ويطبّق حدود المجلدات حسب صلاحياته هو.
func NewFileToken(secret []byte, employeeID string) string {
	exp := time.Now().Add(fileTokenTTL).Unix()
	payload := employeeID + "~" + strconv.FormatInt(exp, 10)
	return payload + "." + sign(secret, payload)
}

// VerifyFileToken يتأكد من التوقيع ومن إنه ما انتهى، ويرجّع رقم الموظف.
//
// ⚠️ الوسم القديم (وقت بس، بلا موظف) ينقبل لحد ما ينتهي — عمره ١٥ دقيقة
// والخادم ما يصدر غيره بعد اليوم، فالصور المفتوحة وقت التحديث ما تنكسر.
// وهو يرجع employeeID فارغ، فيعدّي بس على المجلدات العامة.
func VerifyFileToken(secret []byte, token string) (string, error) {
	dot := strings.LastIndex(token, ".")
	if dot <= 0 {
		return "", errors.New("وسم غير صالح")
	}
	payload, sig := token[:dot], token[dot+1:]
	// hmac.Equal مقارنة ثابتة الزمن — المقارنة العادية تسرّب التوقيع
	// حرف حرف عبر فروقات التوقيت.
	if !hmac.Equal([]byte(sign(secret, payload)), []byte(sig)) {
		return "", errors.New("توقيع غير صالح")
	}
	employeeID, expStr := "", payload
	if i := strings.LastIndex(payload, "~"); i >= 0 {
		employeeID, expStr = payload[:i], payload[i+1:]
	}
	exp, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		return "", fmt.Errorf("وسم غير صالح: %w", err)
	}
	if time.Now().Unix() > exp {
		return "", errors.New("انتهت صلاحية الوسم")
	}
	return employeeID, nil
}

func sign(secret []byte, payload string) string {
	m := hmac.New(sha256.New, secret)
	m.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}
