import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type TodayBoardData, type FinanceSummary } from '../api'
import EmployeeAvatar from './EmployeeAvatar'
import { useSession } from '../session'
import { canDrill, drillTo } from '../utils/drill'
import { pollVisible } from '../utils/poll'

// ═══ لوحة اليوم ═══
//
// «خل نضيف ملاعيب وترتيبات وزينة للنظام».
//
// الإداري يفتح النظام الصبح ويسأل نفس الأسئلة كل يوم: شكد حجز اليوم؟
// منو بالميدان هسه؟ شكد خلّصوا؟ وشنو ينتظرني أني؟ وچان لازم يفتح
// أربع شاشات حتى يجاوبهن.
//
// ⚠️ اللوحة مقسومة قسمين مقصودين:
//   • «شنو صاير اليوم» — أرقام تخبّر (تنقرا ولا تنضغط).
//   • «شغلي اليوم» — أرقام **تنتظر تصرّف**، وكل وحدة تفتح محطتها.
// خلطهن يخلّي الإداري يشوف عشر أرقام ما يعرف أيها يطالبه بشي.

const NUMBERS = 'tabular-nums'

// تصميم (ع) 10-13: العنوان ملوّن فوگ، الرقم كبير، سطر تحته، الأيقونة
// بمربع ملوّن، وسهم ‹ لمن البطاقة تودّي لتفاصيلها.
function StatCard({ icon, label, value, tone, hint, to }: {
  icon: string; label: string; value: React.ReactNode; tone: 'blue' | 'amber' | 'emerald' | 'violet'; hint?: string; to?: string
}) {
  const tones = {
    blue: { card: 'border-sky-200 bg-sky-50/70', ink: 'text-sky-800', chip: 'bg-sky-100' },
    amber: { card: 'border-amber-200 bg-amber-50/70', ink: 'text-amber-800', chip: 'bg-amber-100' },
    emerald: { card: 'border-emerald-200 bg-emerald-50/70', ink: 'text-emerald-800', chip: 'bg-emerald-100' },
    violet: { card: 'border-violet-200 bg-violet-50/70', ink: 'text-violet-800', chip: 'bg-violet-100' },
  }[tone]
  const body = (
    <>
      <div className="min-w-0 flex-1">
        <p className={`text-[12px] font-extrabold ${tones.ink}`}>{label}</p>
        <p className={`mt-0.5 text-3xl font-black leading-tight text-[#0f2040] ${NUMBERS}`}>{value}</p>
        {hint && <p className="text-[11px] text-slate-500">{hint}</p>}
      </div>
      <div className="flex flex-col items-center gap-2">
        <span className={`grid h-11 w-11 place-items-center rounded-xl text-xl ${tones.chip}`}>{icon}</span>
        {to && <span className="grid h-6 w-6 place-items-center rounded-full bg-white text-xs text-slate-500 shadow-sm">‹</span>}
      </div>
    </>
  )
  const cls = `flex items-center justify-between gap-3 rounded-2xl border px-4 py-3 ${tones.card}`
  return to ? <Link to={to} className={`${cls} transition hover:-translate-y-0.5 hover:shadow-md`}>{body}</Link> : <div className={cls}>{body}</div>
}

/** بطاقة شغل — تنضغط وتروح لمحطتها. */
function TaskCard({ icon, label, value, to, hint }: {
  icon: string; label: string; value: number; to?: string; hint: string
}) {
  // ⚠️ الصفر ما يطلع: «ماكو شي ينتظرك» أحسن من صف بطاقات بأصفار
  // تخلّي العين تتعوّد تتجاهلها — فلمن يصير رقم حقيقي ما تنتبهله.
  if (!value) return null
  const cls = 'flex items-center gap-3 rounded-2xl border border-slate-200 bg-white px-4 py-3 transition hover:border-brand-300 hover:shadow-sm'
  const body = (
    <>
      <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-brand-50 text-lg">{icon}</span>
      <div className="min-w-0 flex-1">
        <p className="text-[13px] font-extrabold text-[#0f2040]">{label}</p>
        <p className="text-[11px] text-slate-400">{hint}</p>
      </div>
      <span className={`rounded-xl bg-[#0f2040] px-2.5 py-1 text-sm font-black text-white ${NUMBERS}`}>{value}</span>
    </>
  )
  return to ? <Link to={to} className={cls}>{body}</Link> : <div className={cls}>{body}</div>
}

const MONTH_AR = ['يناير', 'فبراير', 'مارس', 'أبريل', 'مايو', 'يونيو', 'يوليو', 'أغسطس', 'سبتمبر', 'أكتوبر', 'نوفمبر', 'ديسمبر']

/** مخطط خطي — بلا مكتبة، نفس قاعدة النظام (SVG يدوي، صفر تبعيات جديدة).
 *
 * ⚠️ نسبة المقارنة محسوبة من نفس الـ١٤ يوم المتوفرة (أول ٧ مقابل
 * آخر ٧) — ماكو مصدر بيانات لفترة سابقة منفصلة بعد، فهذا أقرب رقم
 * صادق نقدر نطلّعه بلا تعديل بالخادم. */
function MiniChart({ points }: { points: { day: string; count: number }[] }) {
  const width = 700
  const height = 220
  const padX = 28
  const padTop = 34
  const padBottom = 34
  const innerW = width - padX * 2
  const innerH = height - padTop - padBottom
  const max = Math.max(1, ...points.map((p) => p.count))
  const n = points.length || 1

  const coords = points.map((p, i) => ({
    x: padX + (n === 1 ? innerW / 2 : (innerW * i) / (n - 1)),
    y: padTop + innerH - (p.count / max) * innerH,
    ...p,
  }))

  const linePath = coords.map((c, i) => `${i === 0 ? 'M' : 'L'}${c.x.toFixed(1)},${c.y.toFixed(1)}`).join(' ')
  const areaPath = coords.length
    ? `${linePath} L${coords[coords.length - 1].x.toFixed(1)},${(padTop + innerH).toFixed(1)} L${coords[0].x.toFixed(1)},${(padTop + innerH).toFixed(1)} Z`
    : ''

  const total = points.reduce((s, p) => s + p.count, 0)
  const half = Math.floor(n / 2)
  const firstHalf = points.slice(0, half).reduce((s, p) => s + p.count, 0)
  const secondHalf = points.slice(half).reduce((s, p) => s + p.count, 0)
  const trendPct = firstHalf > 0 ? Math.round(((secondHalf - firstHalf) / firstHalf) * 100) : null

  return (
    <div>
      {trendPct !== null && (
        <div className={`mb-2 inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-bold ${
          trendPct >= 0 ? 'bg-emerald-50 text-emerald-700' : 'bg-rose-50 text-rose-700'
        }`}>
          <span>{trendPct >= 0 ? '↗' : '↘'}</span>
          <span>{trendPct >= 0 ? '+' : ''}{trendPct}٪</span>
          <span className="font-normal text-slate-400">مقارنة بالفترة السابقة</span>
        </div>
      )}
      <svg viewBox={`0 0 ${width} ${height}`} className="w-full" style={{ height: 180 }}>
        <defs>
          <linearGradient id="bookingTrendFill" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#2c5aad" stopOpacity="0.25" />
            <stop offset="100%" stopColor="#2c5aad" stopOpacity="0" />
          </linearGradient>
        </defs>
        {/* خطوط الشبكة الأفقية */}
        {[0, 0.5, 1].map((f) => (
          <line key={f} x1={padX} x2={width - padX}
            y1={padTop + innerH * f} y2={padTop + innerH * f}
            stroke="#e2e8f0" strokeDasharray="4 4" strokeWidth={1} />
        ))}
        {areaPath && <path d={areaPath} fill="url(#bookingTrendFill)" />}
        {linePath && <path d={linePath} fill="none" stroke="#2c5aad" strokeWidth={2.5} strokeLinejoin="round" strokeLinecap="round" />}
        {coords.map((c) => {
          const d = new Date(`${c.day}T00:00:00`)
          return (
            <g key={c.day}>
              <circle cx={c.x} cy={c.y} r={4} fill="#fff" stroke="#2c5aad" strokeWidth={2.5}>
                <title>{`${d.toLocaleDateString('ar-IQ', { weekday: 'long', day: 'numeric', month: 'long' })}: ${c.count} حجز`}</title>
              </circle>
              <text x={c.x} y={c.y - 12} textAnchor="middle" fontSize="12" fontWeight="700" fill="#1a3a5c">{c.count}</text>
              <text x={c.x} y={height - padBottom + 18} textAnchor="middle" fontSize="10" fill="#94a3b8">
                {d.getDate()} {MONTH_AR[d.getMonth()]}
              </text>
            </g>
          )
        })}
      </svg>
      <div className="mt-1 flex justify-end">
        <span className="inline-flex items-center gap-1.5 rounded-xl bg-slate-50 px-3 py-1.5 text-xs font-bold text-slate-600">
          📅 المجموع <b className={`text-sm text-[#0f2040] ${NUMBERS}`}>{total}</b> حجز
        </span>
      </div>
    </div>
  )
}

export default function TodayBoard({ finance }: { finance?: FinanceSummary | null } = {}) {
  const [d, setD] = useState<TodayBoardData | null>(null)
  // 🔴 (ع) 10-16 «الرقم = القائمة»: كل بطاقة تفتح نفس الحجوزات الي عدّتها
  const { employee, permissions } = useSession()
  const dl = (key: string) => (canDrill(employee, permissions) ? drillTo(key) : undefined)

  useEffect(() => {
    let alive = true
    const load = () => api.getTodayBoard().then((x) => { if (alive) setD(x) }).catch(() => {})
    const t = setTimeout(load, 0)
    // يتجدد كل دقيقتين — الأرقام تتغيّر والإداري ما يعيد تحميل الصفحة
    const stop = pollVisible(load, 120_000)
    return () => { alive = false; clearTimeout(t); stop() }
  }, [])

  if (!d) return null

  const totalWaiting = d.needsContact + d.needsCrew + d.needsPaper + d.needsFinish

  return (
    <div className="space-y-4">
      {/* ═══ شنو صاير اليوم ═══ */}
      <div>
        <h3 className="mb-2 text-sm font-extrabold text-[#0f2040]">📅 شنو صاير اليوم</h3>
        <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
          <StatCard icon="✨" label="حجوزات وصلت اليوم" value={d.newToday} tone="amber" hint="حجز جديد انسجّل اليوم — اضغط تشوفهن" to={dl('new_today')} />
          <StatCard icon="🏁" label="خلّصوا اليوم" value={d.completedToday} tone="emerald" hint="تأشّرت «منجز» اليوم — اضغط تشوفهن" to={dl('completed_today')} />
          {/* ⚠️ نفس `d.inField` (حجوزات IN_PROGRESS) — تسمية بس، البيانات ما تغيّرت */}
          <StatCard icon="🚚" label="بالميدان هسه" value={d.inField} tone="violet" hint="الكادر انطلق أو وصل اليوم وبعده ما خلص" to={dl('in_field')} />
          <StatCard icon="🗓️" label="حجوزات اليوم" value={d.bookingsToday} tone="blue" hint="موعدها اليوم — اضغط تشوفهن" to={dl('today_scheduled')} />
        </div>
      </div>

      {/* ═══ شغلي اليوم ═══ */}
      <div>
        <h3 className="mb-2 text-sm font-extrabold text-[#0f2040]">
          🎯 شغلي اليوم
          {totalWaiting === 0 && <span className="mr-2 text-xs font-normal text-emerald-600">— ماكو شي ينتظرك ✓</span>}
        </h3>

        {/* ⚠️ «نجاح الأسبوع»/«تدقيق الاستلام»/«مهام لسه» — أرقام تخبّر
            مو تنتظر تصرّف، فهي أقرب لصف «شنو صاير» لكن مطلوبة هنا
            بالتصميم. الصفر المزيّف أسوأ من فراغ: «—» لو ماكو حجوزات
            هذا الأسبوع أصلاً. */}
        <div className="mb-2.5 grid grid-cols-1 gap-2.5 sm:grid-cols-3">
          <StatCard icon="📈" label="نجاح الأسبوع"
            value={d.weekTotal > 0 ? `${Math.round((d.weekCompleted / d.weekTotal) * 100)}٪` : '—'}
            tone="emerald" to={dl('week')}
            hint={d.weekTotal > 0 ? `خلص ${d.weekCompleted} من ${d.weekTotal} حجز انسجّل آخر ٧ أيام — اضغط تشوف التفاصيل` : 'ماكو حجوزات آخر ٧ أيام'} />
          {finance != null && (
            <StatCard icon="💵" label="تدقيق استلام المبالغ"
              value={finance.unverifiedCount} tone="amber" hint="حجوزات منجزة المحاسب بعده ما دقّق مبلغها — من كل الأيام" to={dl('unverified')} />
          )}
          <StatCard icon="🎯" label="عدد المهام لسه" value={totalWaiting} tone="blue"
            hint="مجموع البطاقات الأربع جوّا (تواصل + كادر + ورق + إكمال) — من كل الفترات، مو اليوم بس" />
        </div>

        {totalWaiting > 0 && (
          <div className="grid grid-cols-1 gap-2.5 sm:grid-cols-2">
            <TaskCard icon="📞" label="ينتظرون تواصل وتثبيت" value={d.needsContact}
              to={dl('needs_contact')} hint="انسجّلت وما أحد ثبّتها ويا الزبون بعد" />
            <TaskCard icon="👥" label="ينتظرون كادراً وموعداً" value={d.needsCrew}
              to={dl('needs_crew')} hint="انثبّتت ويا الزبون وما انحطلها كادر" />
            <TaskCard icon="🧾" label="منجزة وناقصها ورق" value={d.needsPaper}
              to={dl('needs_paper')} hint="الشغل خلص وناقص فاتورة أو تقرير — هاي الي تجيب الغرامات" />
            <TaskCard icon="🔄" label="تحتاج موعد إكمال" value={d.needsFinish}
              to={dl('needs_finish')} hint="طلع الكادر وأنجز جزء — يحتاج موعد ثاني" />
          </div>
        )}
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {/* ═══ لوحة الشرف ═══ */}
        {d.topCrew.length > 0 && (
          <div className="rounded-2xl border border-amber-200 bg-gradient-to-br from-amber-50 to-white p-4">
            <h3 className="text-sm font-extrabold text-amber-900">🏆 أكثر الكوادر طلعات — هذا الشهر</h3>
            {/* ⚠️ المقياس **الطلعات** مو الحجوزات: الحجز الي ياخذ أربع
                أيام أربع طلعات، وكل طلعة إلها كادرها. نفس المقياس الي
                صلّحنا بيه الإنتاجية. */}
            <div className="mt-3 space-y-2">
              {d.topCrew.map((c, i) => (
                <div key={c.employeeId} className="flex items-center gap-3 rounded-xl bg-white px-3 py-2 shadow-sm">
                  <span className="text-lg">{['🥇', '🥈', '🥉'][i] || '⭐'}</span>
                  <EmployeeAvatar name={c.name} photoUrl={c.photoUrl} size="sm" />
                  <span className="flex-1 truncate text-[13px] font-bold text-[#0f2040]">{c.name}</span>
                  <span className={`text-xs font-black text-amber-700 ${NUMBERS}`}>{c.visits} طلعة</span>
                  {c.done > 0 && <span className="text-[10px] text-emerald-600">{c.done} خلّصها</span>}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* ═══ حركة الحجوزات ═══ */}
        <div className="rounded-2xl border border-slate-200 bg-white p-4">
          <h3 className="text-sm font-extrabold text-[#0f2040]">📊 حركة الحجوزات — آخر ١٤ يوم</h3>
          <div className="mt-3">
            <MiniChart points={d.last14} />
          </div>
        </div>
      </div>
    </div>
  )
}
