import { useEffect, useState } from 'react'
import { api, type Announcement } from '../api'

/**
 * شريط الإعلانات المتحرك.
 *
 * يظهر لكل موظف أول ما يفتح النظام، ويعيد نفسه باستمرار. الموظف
 * يقدر يخفيه — بس الإخفاء ينخزن بـsessionStorage مو localStorage،
 * يعني يرجع يظهرله بعد تسجيل خروج ودخول، مثل ما انطلب بالضبط.
 */

const DISMISS_KEY = 'announcements-dismissed'

export default function AnnouncementTicker() {
  const [items, setItems] = useState<Announcement[]>([])
  const [hidden, setHidden] = useState(() => sessionStorage.getItem(DISMISS_KEY) === '1')

  // نعيد الجلب كل دقيقة: بدونها الإعلان الجديد ما يوصل إلا للي يفتح
  // النظام بعده — والموظفين الي شغّالين أصلاً يضلون ما يشوفونه.
  useEffect(() => {
    if (hidden) return
    let alive = true
    const fetch = () => {
      api.getAnnouncements()
        .then((rows) => { if (alive) setItems(rows) })
        .catch(() => { if (alive) setItems([]) })
    }
    fetch()
    const timer = setInterval(fetch, 60_000)
    return () => { alive = false; clearInterval(timer) }
  }, [hidden])

  if (hidden || items.length === 0) return null

  // نكرر النص مرتين حتى الحركة تبقى متصلة بلا فراغ بالنص
  const line = items.map((a) => a.body).join('   ★   ')

  return (
    <div
      dir="rtl"
      className="relative flex items-center gap-3 overflow-hidden border-b border-amber-300/40 px-4 py-2"
      style={{ background: 'linear-gradient(90deg, #1a3a5c, #2c5aad)' }}
    >
      <span className="shrink-0 rounded-full bg-amber-400 px-3 py-0.5 text-xs font-bold text-amber-950">
        📢 إعلان
      </span>

      <div className="relative flex-1 overflow-hidden">
        <div className="ticker-track whitespace-nowrap text-sm font-medium text-white">
          <span className="px-8">{line}</span>
          <span className="px-8">{line}</span>
        </div>
      </div>

      <button
        onClick={() => { sessionStorage.setItem(DISMISS_KEY, '1'); setHidden(true) }}
        title="يخفيه لحد ما تسجل خروج وترجع تدخل"
        className="shrink-0 rounded-lg px-2 py-0.5 text-xs text-blue-200 hover:bg-white/10 hover:text-white"
      >
        ✕
      </button>

      <style>{`
        .ticker-track {
          display: inline-flex;
          animation: amani-ticker 30s linear infinite;
        }
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
