import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type TodayBoardData } from '../api'
import EmployeeAvatar from './EmployeeAvatar'

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

function StatCard({ icon, label, value, tone }: {
  icon: string; label: string; value: number; tone: 'blue' | 'amber' | 'emerald' | 'violet'
}) {
  const tones = {
    blue: 'border-sky-200 bg-sky-50/70 text-sky-800',
    amber: 'border-amber-200 bg-amber-50/70 text-amber-800',
    emerald: 'border-emerald-200 bg-emerald-50/70 text-emerald-800',
    violet: 'border-violet-200 bg-violet-50/70 text-violet-800',
  }
  return (
    <div className={`flex items-center justify-between rounded-2xl border px-4 py-3 ${tones[tone]}`}>
      <span className="text-xl opacity-70">{icon}</span>
      <div className="text-left">
        <p className="text-[11px] font-bold opacity-80">{label}</p>
        <p className={`text-2xl font-black leading-tight ${NUMBERS}`}>{value}</p>
      </div>
    </div>
  )
}

/** بطاقة شغل — تنضغط وتروح لمحطتها. */
function TaskCard({ icon, label, value, to, hint }: {
  icon: string; label: string; value: number; to: string; hint: string
}) {
  // ⚠️ الصفر ما يطلع: «ماكو شي ينتظرك» أحسن من صف بطاقات بأصفار
  // تخلّي العين تتعوّد تتجاهلها — فلمن يصير رقم حقيقي ما تنتبهله.
  if (!value) return null
  return (
    <Link
      to={to}
      className="flex items-center gap-3 rounded-2xl border border-slate-200 bg-white px-4 py-3 transition hover:border-brand-300 hover:shadow-sm"
    >
      <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-brand-50 text-lg">{icon}</span>
      <div className="min-w-0 flex-1">
        <p className="text-[13px] font-extrabold text-[#0f2040]">{label}</p>
        <p className="text-[11px] text-slate-400">{hint}</p>
      </div>
      <span className={`rounded-xl bg-[#0f2040] px-2.5 py-1 text-sm font-black text-white ${NUMBERS}`}>{value}</span>
    </Link>
  )
}

/** مخطط بسيط — أعمدة، بلا مكتبة. */
function MiniChart({ points }: { points: { day: string; count: number }[] }) {
  const max = Math.max(1, ...points.map((p) => p.count))
  return (
    <div className="flex h-24 items-end gap-1">
      {points.map((p) => {
        const h = Math.round((p.count / max) * 100)
        const d = new Date(`${p.day}T00:00:00`)
        return (
          <div key={p.day} className="group flex flex-1 flex-col items-center gap-1">
            <div className="flex w-full flex-1 items-end">
              {/* ⚠️ اليوم الفاضي يبقى عمود بارتفاع أدنى: العمود
                  المفقود يخلّي العين تظن إن اليوم مو موجود أصلاً. */}
              <div
                className={`w-full rounded-t-md transition-all ${p.count ? 'bg-brand-500 group-hover:bg-brand-700' : 'bg-slate-200'}`}
                style={{ height: `${Math.max(h, 4)}%` }}
                title={`${d.toLocaleDateString('ar-IQ', { weekday: 'long', day: 'numeric', month: 'long' })}: ${p.count} حجز`}
              />
            </div>
            <span className="text-[8px] text-slate-400">{d.getDate()}</span>
          </div>
        )
      })}
    </div>
  )
}

export default function TodayBoard() {
  const [d, setD] = useState<TodayBoardData | null>(null)

  useEffect(() => {
    let alive = true
    const load = () => api.getTodayBoard().then((x) => { if (alive) setD(x) }).catch(() => {})
    const t = setTimeout(load, 0)
    // يتجدد كل دقيقتين — الأرقام تتغيّر والإداري ما يعيد تحميل الصفحة
    const iv = setInterval(load, 120_000)
    return () => { alive = false; clearTimeout(t); clearInterval(iv) }
  }, [])

  if (!d) return null

  const totalWaiting = d.needsContact + d.needsCrew + d.needsPaper + d.needsFinish

  return (
    <div className="space-y-4">
      {/* ═══ شنو صاير اليوم ═══ */}
      <div>
        <h3 className="mb-2 text-sm font-extrabold text-[#0f2040]">📅 شنو صاير اليوم</h3>
        <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
          <StatCard icon="🗓️" label="حجوزات اليوم" value={d.bookingsToday} tone="blue" />
          <StatCard icon="🚚" label="بالميدان هسه" value={d.inField} tone="violet" />
          <StatCard icon="🏁" label="خلّصوا اليوم" value={d.completedToday} tone="emerald" />
          <StatCard icon="✨" label="حجوزات وصلت اليوم" value={d.newToday} tone="amber" />
        </div>
      </div>

      {/* ═══ شغلي اليوم ═══ */}
      <div>
        <h3 className="mb-2 text-sm font-extrabold text-[#0f2040]">
          🎯 شغلي اليوم
          {totalWaiting === 0 && <span className="mr-2 text-xs font-normal text-emerald-600">— ماكو شي ينتظرك ✓</span>}
        </h3>
        {totalWaiting > 0 && (
          <div className="grid grid-cols-1 gap-2.5 sm:grid-cols-2">
            <TaskCard icon="📞" label="ينتظرون تواصل وتثبيت" value={d.needsContact}
              to="/bookings" hint="حجوزات مرحّلة وما أحد حچى وية زبونها" />
            <TaskCard icon="👥" label="ينتظرون كادراً وموعداً" value={d.needsCrew}
              to="/bookings" hint="مثبّتة وما عليها كادر بعد" />
            <TaskCard icon="🧾" label="منجزة وناقصها ورق" value={d.needsPaper}
              to="/bookings" hint="فاتورة أو تقرير — وهاي الي تجيب الغرامات" />
            <TaskCard icon="🔄" label="تحتاج موعد إكمال" value={d.needsFinish}
              to="/bookings" hint="طلع الكادر وما خلّص الشغل" />
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
          <p className="mt-2 text-[11px] text-slate-400">
            المجموع: <b className="text-slate-600">{d.last14.reduce((s, p) => s + p.count, 0)}</b> حجز
          </p>
        </div>
      </div>
    </div>
  )
}
