// ═══ نسخ نص للحافظة ═══
//
// «الموضوع متعب من اكعد احدد وانسخ والصق» — زر نسخ بجنب كل كود حجز.
//
// ⚠️ ماكو ولا دالة نسخ بكل المشروع قبل هذا الملف (صفر
// `navigator.clipboard`) — فتنبنى مرة وحدة وتخدم كل مكان.
//
// ⚠️ و`navigator.clipboard` **ما تشتغل إلا على HTTPS أو localhost**،
// وترمي لو المستخدم رفض الإذن. فبديل `execCommand` موجود حتى الزر
// ما يصير ميتاً على أي جهاز — زر ينضغط وما ينسخ أسوأ من ماكو زر.
export async function copyText(text: string): Promise<boolean> {
  const value = String(text ?? '').trim()
  if (!value) return false
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(value)
      return true
    }
  } catch {
    // ننزل للبديل — ما نرجّع فشل قبل ما نجرّب
  }
  try {
    const ta = document.createElement('textarea')
    ta.value = value
    // برّا الشاشة حتى ما تقفز الصفحة عند التركيز
    ta.style.position = 'fixed'
    ta.style.top = '-1000px'
    ta.setAttribute('readonly', '')
    document.body.appendChild(ta)
    ta.select()
    const ok = document.execCommand('copy')
    document.body.removeChild(ta)
    return ok
  } catch {
    return false
  }
}
