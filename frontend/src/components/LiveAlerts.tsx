import { useEffect, useRef, useState } from 'react'
import { api } from '../api'

// ═══ التنبيهات الحيّة ═══
//
// «أريد توصل إشعارات لأجهزتهم وإشعارات للحاسبات».
//
// الموظف چان لازم يفتح النظام ويضغط الجرس حتى يعرف إن أكو شي جديد.
// يعني الإشعار موجود من ساعتين ومحد شافه — والحجز العاجل يقعد ينتظر.
//
// ثلاث طبقات، كل وحدة تشتغل لحالها لو الي فوگها ما اشتغلت:
//   ١) **إشعار النظام** (Notification API) — يطلع بزاوية الشاشة حتى
//      لو النظام مو مفتوح بالتبويب الحالي. يشتغل بالحاسبة وبأندرويد
//      لمن يكون النظام مضاف للشاشة الرئيسية.
//   ٢) **صوت خفيف** — الموظف الي شغّال بورقة قدّامه ما يشوف الشاشة.
//   ٣) **عدّاد بعنوان التبويب** — حتى لو التبويب بالخلفية، يشوف
//      «(٣) نظام الأماني» بشريط المتصفح.
//
// ⚠️ ما نطلب إذن الإشعارات بأول ثانية: طلب يطلع للموظف قبل ما يفهم
// شنو النظام ينرفض بالغالب، وبعدها ما تكدر تطلبه مرة ثانية. نعرض زراً
// هادئاً وهو يقرر.
//
// ⚠️ والصوت ما ينشغّل إلا بعد أول تفاعل من المستخدم: المتصفحات تمنع
// الصوت التلقائي، ومحاولة تشغيله بلا تفاعل تطلّع خطأ بالكونسول بكل
// دورة فحص.

const SEEN_KEY = 'lastSeenNotificationAt'

/** صوت قصير بلا ملف — نولّده بالمتصفح (بدون تحميل أي شي). */
function beep() {
  try {
    const Ctx = window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext
    const ctx = new Ctx()
    const osc = ctx.createOscillator()
    const gain = ctx.createGain()
    osc.connect(gain); gain.connect(ctx.destination)
    osc.frequency.value = 880
    gain.gain.setValueAtTime(0.0001, ctx.currentTime)
    gain.gain.exponentialRampToValueAtTime(0.08, ctx.currentTime + 0.02)
    gain.gain.exponentialRampToValueAtTime(0.0001, ctx.currentTime + 0.35)
    osc.start()
    osc.stop(ctx.currentTime + 0.36)
    setTimeout(() => ctx.close().catch(() => {}), 600)
  } catch { /* الصوت ترف — فشله ما يجوز يوقف شي */ }
}

export default function LiveAlerts() {
  const [permission, setPermission] = useState<NotificationPermission>(
    typeof Notification !== 'undefined' ? Notification.permission : 'denied',
  )
  const baseTitle = useRef(document.title)
  const interacted = useRef(false)

  useEffect(() => {
    const mark = () => { interacted.current = true }
    window.addEventListener('pointerdown', mark, { once: true })
    window.addEventListener('keydown', mark, { once: true })
    return () => {
      window.removeEventListener('pointerdown', mark)
      window.removeEventListener('keydown', mark)
    }
  }, [])

  useEffect(() => {
    let alive = true

    const check = async () => {
      try {
        const { notifications, unreadCount } = await api.getNotifications()
        if (!alive) return

        // عدّاد بعنوان التبويب — يشتغل بأي متصفح وبلا أي إذن
        document.title = unreadCount > 0 ? `(${unreadCount}) ${baseTitle.current}` : baseTitle.current

        const lastSeen = localStorage.getItem(SEEN_KEY)
        // ⚠️ «الجديد» = الي وصل بعد آخر مرة نبّهنا عنه، مو «غير
        // مقروء»: بدونها ننبّه على نفس الإشعار كل دورة فحص لحد ما
        // يفتحه — يعني صوت كل دقيقة على شي شافه من زمان.
        const fresh = notifications.filter((n) => !n.read && (!lastSeen || n.createdAt > lastSeen))
        if (fresh.length === 0) return

        localStorage.setItem(SEEN_KEY, fresh[0].createdAt)

        if (permission === 'granted' && typeof Notification !== 'undefined') {
          const top = fresh[0]
          new Notification('نظام شركة الأماني', {
            body: fresh.length > 1 ? `${top.message}\n(و${fresh.length - 1} إشعار ثاني)` : top.message,
            icon: '/staff_mange/favicon.png',
            tag: 'staffmange',
          })
        }
        if (interacted.current) beep()
      } catch { /* الشبكة تقطع — التنبيه ترف، ما يجوز يطلّع خطأ للموظف */ }
    }

    const t = setTimeout(check, 3000)
    const iv = setInterval(check, 45_000)
    return () => { alive = false; clearTimeout(t); clearInterval(iv) }
  }, [permission])

  // ⚠️ ما نعرض الزر إلا لمن يكون الإذن ما انطلب بعد: عرضه بعد الرفض
  // إلحاح، وبعد الموافقة ضجيج.
  if (typeof Notification === 'undefined' || permission !== 'default') return null

  return (
    <button
      onClick={() => Notification.requestPermission().then(setPermission)}
      className="rounded-lg border border-brand-200 bg-brand-50 px-2.5 py-1 text-[11px] font-bold text-brand-700 hover:bg-brand-100"
      title="خلي إشعارات النظام توصلك على الجهاز حتى لو النظام مو مفتوح قدّامك"
    >
      🔔 فعّل الإشعارات
    </button>
  )
}
