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
  // ═══ أخبار اليوم ═══
  //
  // «الأخبار لازم تضل تظهر — مثلاً أخبار اليوم: كم شخص تغرم، كم شخص
  // رجعتله النقاط مالته؟»
  //
  // ⚠️ **أرقام بلا أسماء**: الشريط يشوفه كل موظفي الشركة، ونشر اسم
  // المغرَّم على الجميع عقوبة ثانية ما طلبها أحد.
  const [news, setNews] = useState<{ penalized: number; restored: number } | null>(null)

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
    const fetchNews = () => {
      api.getDisciplineTodayHeadline()
        .then((n) => { if (alive) setNews(n) })
        // ⚠️ فشل الأخبار ما يطفّي إعلانات المدير — كل واحد منفصل
        .catch(() => { if (alive) setNews(null) })
    }
    fetch()
    fetchNews()
    const timer = setInterval(() => { fetch(); fetchNews() }, 60_000)
    return () => { alive = false; clearInterval(timer) }
  }, [])

  // بنود أخبار اليوم — تدور بالشريط جنب إعلانات المدير.
  //
  // ⚠️ **«ماكو» مو «٠»**: صفر مغرَّمين خبر زين، بس عرضه كرقم كل يوم
  // ضجيج. فالبند ما يطلع إلا لمن يكون أكبر من صفر — والسطر يبقى
  // صادقاً بالحالتين.
  const newsLines: string[] = []
  if (news) {
    if (news.penalized > 0) newsLines.push(`📉 أخبار اليوم: ${news.penalized} موظف انخصمت نقاط انضباطه`)
    if (news.restored > 0) newsLines.push(`📈 أخبار اليوم: ${news.restored} موظف رجعتله نقاطه`)
  }

  if (items.length === 0 && newsLines.length === 0) return null

  // كل إعلان بند مستقل مفصول بنجمة — تكدر تضيف أكثر من موضوع
  // وكلهم يمرّون بالشريط واحد ورا الثاني.
  const line = [...items.map((a) => a.body), ...newsLines].join('   ★   ')

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
        📢 إعلان{items.length + newsLines.length > 1 ? ` (${items.length + newsLines.length})` : ''}
      </span>

      <div className="ticker-frame flex-1">
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
        /* ═══ الشريط الدوّار ═══
           «من تجي تظهر مرة وحدة وتوصل نصف الكلام يختفي الكلام كله
            ويبقى بس الشريط الأزرق».

           ⚠️⚠️ **سببان اجتمعوا**، وقستهما بالمتصفح:

           ① العرض: المسار چان \`inline-flex\` بلا عرض، فعرض صندوقه
              ينقص لعرض الحاوية — و\`translateX(-50%)\` تنحسب من **عرض
              الصندوق مو عرض النص**، فالإزاحة ما تساوي نسخة كاملة.
              \`width: max-content\` يخليها نسخة بالضبط.

           ② ⚠️ **والاتجاه** — وهذا الي چان يخفي النص كلياً: بـRTL
              المحتوى الفايض يطلع **لليسار**، فالمسار يبدي أصلاً برّه
              الشاشة (قست \`x = -9170\` وعرض الإطار 1101 — مخفي بالكامل
              بكل لحظة قستها). فالحل: **نثبّت المسار على الحافة اليمنى**
              (\`position:absolute; right:0\`) حتى تبدي النسخة الأولى
              مرئية، وندفعه **لليمين** بنسخة كاملة — فالنسخة الثانية
              تحل محل الأولى بالضبط والدوران ينوصل بلا فراغ. */
        .ticker-frame { position: relative; height: 1.5rem; overflow: hidden; }
        .ticker-track {
          position: absolute;
          top: 0;
          right: 0;
          display: inline-flex;
          width: max-content;
          animation-name: amani-ticker;
          animation-timing-function: linear;
          animation-iteration-count: infinite;
        }
        /* يوقف بمرور الماوس — حتى يكدر يقرا إعلان طويل */
        .ticker-track:hover { animation-play-state: paused; }
        @keyframes amani-ticker {
          from { transform: translateX(0); }
          to { transform: translateX(50%); }
        }
        /* ⚠️ بلا حركة: النص يبقى **مقروءاً** مو مخفياً — الي يطفّي
           الحركة بجهازه لازم يقرا الإعلان بعد. */
        @media (prefers-reduced-motion: reduce) {
          .ticker-track { animation: none; position: static; width: auto; }
        }
      `}</style>
    </div>
  )
}
