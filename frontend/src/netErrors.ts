// ═══ مصيدة أخطاء الشبكة ═══
//
// أكثر من ٢٠ شاشة تنادي الـAPI بـ.then() بلا .catch(). لمن يفشل الطلب،
// الوعد المرفوض ما يمسكه أحد: ما تطلع رسالة، ما ينتغير شي بالشاشة،
// والموظف يضل يضغط الزر ويظن النظام «واگف».
//
// الصمت أسوأ من رسالة خطأ: الموظف يعيد إدخال البيانات، أو يظن الشغلة
// انحفظت وهي ما انحفظت.
//
// هذا الملف يمسك أي رفض ما انمسك بالشاشة نفسها ويعرضه بلغة مفهومة.
// مو بديل عن معالجة الخطأ بمكانها — بس شبكة أمان تمنع الفشل الصامت.

/** رسالة «Failed to fetch» مالت المتصفح ما تفيد موظف عربي بشي. */
function humanize(reason: unknown): string {
  const raw = reason instanceof Error ? reason.message : String(reason ?? '')

  // الفشل على مستوى الشبكة: السيرفر ما رد أصلاً (واگف، انترنت مقطوع،
  // أو الجهاز نايم). هذا مو خطأ بالبيانات فما ينفع نعرضه كأنه رفض.
  if (/Failed to fetch|NetworkError|Load failed|ERR_NETWORK/i.test(raw)) {
    return 'ما وصلنا للسيرفر — تأكد من الانترنت وجرّب مرة ثانية بعد ثواني.'
  }
  if (/timeout|aborted/i.test(raw)) {
    return 'الطلب طوّل أكثر من اللازم — جرّب مرة ثانية.'
  }
  return raw || 'صار خطأ غير متوقع.'
}

/** ما نزعج الموظف بنفس الرسالة عشر مرات لو تكرر نفس الفشل. */
let lastShown = ''
let lastShownAt = 0
const QUIET_MS = 4000

export function installNetworkErrorTrap() {
  window.addEventListener('unhandledrejection', (event) => {
    const msg = humanize(event.reason)

    // انتهاء الجلسة إله معالجته الخاصة بـapi.ts (يطلّع المستخدم لصفحة
    // الدخول) — عرض رسالة ثانية فوگه بس يربك.
    if (/يجب تسجيل الدخول/.test(msg)) return

    const now = Date.now()
    if (msg === lastShown && now - lastShownAt < QUIET_MS) return
    lastShown = msg
    lastShownAt = now

    // نطبعه بالكونسول بعد حتى يكدر يشوفه المطور بالتفصيل
    console.error('[unhandled]', event.reason)
    alert(msg)
  })
}
