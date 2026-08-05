import { useEffect, useState } from 'react'
import { api, type Announcement } from '../api'

/**
 * شريط الإعلانات المتحرك.
 *
 * يظهر لكل موظف طول ما اكو إعلان فعّال، وما يتوقف أبداً إلا لما
 * المدير يوقفه من «لوحة الإعلانات».
 *
 * ⚠️ ماكو زر إخفاء للموظف — وهذا مقصود. قبل كان اكو زر ✕ يخزّن
 * الإخفاء بـsessionStorage، وsessionStorage ما ينمسح بتحديث الصفحة
 * (ينمسح بس لما ينسد التبويب). يعني الموظف يضغط ✕ مرة وحدة، وبعدها
 * يحدّث الصفحة ألف مرة وما يطلعله ولا إعلان — وهو يحسب النظام خربان.
 * والمطلوب أصلاً إن الإعلان ما يوقف إلا بيد المدير.
 */

// مفتاح الإخفاء القديم — ننظّفه حتى الي ضغط ✕ سابقاً ترجعله الإعلانات
const LEGACY_DISMISS_KEY = 'announcements-dismissed'

export default function AnnouncementTicker() {
  const [items, setItems] = useState<Announcement[]>([])

  // نعيد الجلب كل دقيقة: بدونها الإعلان الجديد ما يوصل إلا للي يفتح
  // النظام بعده — والموظفين الي شغّالين أصلاً يضلون ما يشوفونه.
  useEffect(() => {
    sessionStorage.removeItem(LEGACY_DISMISS_KEY)
    let alive = true
    const fetch = () => {
      api.getAnnouncements()
        .then((rows) => { if (alive) setItems(rows ?? []) })
        .catch(() => { if (alive) setItems([]) })
    }
    fetch()
    const timer = setInterval(fetch, 60_000)
    return () => { alive = false; clearInterval(timer) }
  }, [])

  if (items.length === 0) return null

  // كل إعلان بند مستقل مفصول بنجمة — تكدر تضيف أكثر من موضوع
  // وكلهم يمرّون بالشريط واحد ورا الثاني.
  const line = items.map((a) => a.body).join('   ★   ')

  // المدة تتناسب مع الطول: النص الطويل ما يمر بنفس سرعة القصير
  // وإلا ما ينقرأ.
  const durationSec = Math.max(20, Math.round(line.length / 4))

  return (
    <div
      dir="rtl"
      className="relative flex items-center gap-3 overflow-hidden border-b border-amber-300/40 px-4 py-2"
      style={{ background: 'linear-gradient(90deg, #1a3a5c, #2c5aad)' }}
    >
      <span className="shrink-0 rounded-full bg-amber-400 px-3 py-0.5 text-xs font-bold text-amber-950">
        📢 إعلان{items.length > 1 ? ` (${items.length})` : ''}
      </span>

      <div className="relative flex-1 overflow-hidden">
        {/* نسختين من النص حتى الدوران يبقى متصل بلا فراغ بالنص */}
        <div
          className="ticker-track whitespace-nowrap text-sm font-medium text-white"
          style={{ animationDuration: `${durationSec}s` }}
        >
          <span className="px-8">{line}</span>
          <span className="px-8">{line}</span>
        </div>
      </div>

      <style>{`
        .ticker-track {
          display: inline-flex;
          animation-name: amani-ticker;
          animation-timing-function: linear;
          animation-iteration-count: infinite;
        }
        /* يوقف بمرور الماوس — حتى يكدر يقرا إعلان طويل */
        .ticker-track:hover { animation-play-state: paused; }
        @keyframes amani-ticker {
          from { transform: translateX(-50%); }
          to { transform: translateX(0); }
        }
        @media (prefers-reduced-motion: reduce) {
          .ticker-track { animation: none; }
        }
      `}</style>
    </div>
  )
}
