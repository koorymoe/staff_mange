import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type FinanceSummary, type MonitorReview, type TodayBoardData } from '../api'
import { useSession } from '../session'
import { isPathAllowed } from './navTree'

// ═══ الواجهة الرئيسية للمراقب — بتصميم (ع) 10-04 ═══
// ترحيب وساعة، أربع بطاقات، حركة الحجوزات، أهم التنبيهات (صندوق المراقب)،
// أفضل الموظفين، أنشطة سريعة، وإحصائيات. كل رقم من النظام ويودّي لشاشته.
// (بلا عين بالترحيب — طلب (ع)).

const card = 'rounded-2xl border border-slate-200 bg-white p-4 shadow-sm'
const fmt = (n: number) => n.toLocaleString('en-US')

function Clock() {
  const [now, setNow] = useState(() => new Date())
  useEffect(() => {
    const t = window.setInterval(() => setNow(new Date()), 30_000)
    return () => window.clearInterval(t)
  }, [])
  return (
    <div className="rounded-2xl bg-white/10 px-6 py-4 text-center ring-1 ring-white/20 backdrop-blur">
      <p className="text-3xl font-extrabold tabular-nums">{now.toLocaleTimeString('ar-IQ', { timeZone: 'Asia/Baghdad', hour: '2-digit', minute: '2-digit' })}</p>
      <p className="mt-1 text-xs opacity-80">{now.toLocaleDateString('ar-IQ', { timeZone: 'Asia/Baghdad', weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })}</p>
    </div>
  )
}

function StatCard({ label, value, to, tone, icon }: { label: string; value: number | string; to: string | null; tone: 'blue' | 'green' | 'amber' | 'red'; icon: string }) {
  const t = {
    blue: 'border-sky-200 bg-sky-50/60 text-sky-700',
    green: 'border-emerald-200 bg-emerald-50/60 text-emerald-700',
    amber: 'border-amber-200 bg-amber-50/60 text-amber-700',
    red: 'border-red-200 bg-red-50/60 text-red-700',
  }[tone]
  const body = (
    <>
      <div>
        <p className="text-sm font-bold text-slate-700">{label}</p>
        <p className="mt-1 text-3xl font-extrabold tabular-nums">{value}</p>
        {to && <p className="mt-2 text-xs font-bold opacity-80 group-hover:underline">← عرض التفاصيل</p>}
      </div>
      <span className="grid h-12 w-12 place-items-center rounded-full bg-white text-2xl shadow-sm">{icon}</span>
    </>
  )
  const cls = `group flex items-center justify-between rounded-2xl border p-4 transition ${t}`
  return to ? <Link to={to} className={`${cls} hover:shadow-md`}>{body}</Link> : <div className={cls}>{body}</div>
}

function Chart({ points }: { points: { day: string; count: number }[] }) {
  if (points.length < 2) return <p className="text-sm text-slate-400">ماكو بيانات كافية.</p>
  const W = 600, H = 170, P = 24
  const max = Math.max(1, ...points.map((p) => p.count))
  const x = (i: number) => P + (i * (W - 2 * P)) / (points.length - 1)
  const y = (v: number) => H - P - (v * (H - 2 * P)) / max
  const line = points.map((p, i) => `${i ? 'L' : 'M'}${x(i)} ${y(p.count)}`).join(' ')
  return (
    <svg viewBox={`0 0 ${W} ${H}`} className="w-full">
      <path d={`${line} L${x(points.length - 1)} ${H - P} L${x(0)} ${H - P}Z`} fill="#3b82f6" opacity="0.12" />
      <path d={line} fill="none" stroke="#2563eb" strokeWidth="2.5" />
      {points.map((p, i) => (
        <g key={p.day}>
          <circle cx={x(i)} cy={y(p.count)} r="3.5" fill="#fff" stroke="#2563eb" strokeWidth="2" />
          <text x={x(i)} y={y(p.count) - 8} textAnchor="middle" fontSize="11" fontWeight="700" fill="#1e3a8a">{p.count}</text>
          {i % 2 === 0 && <text x={x(i)} y={H - 6} textAnchor="middle" fontSize="10" fill="#94a3b8">{p.day.slice(5)}</text>}
        </g>
      ))}
    </svg>
  )
}

const ago = (s: string) => {
  const m = Math.round((Date.now() - new Date(s).getTime()) / 60000)
  if (m < 60) return `منذ ${m} د`
  if (m < 1440) return `منذ ${Math.round(m / 60)} ساعة`
  return `منذ ${Math.round(m / 1440)} يوم`
}

export default function MonitorHome() {
  const { employee, permissions, gpsServiceId } = useSession()
  const [board, setBoard] = useState<TodayBoardData | null>(null)
  const [fin, setFin] = useState<FinanceSummary | null>(null)
  const [summary, setSummary] = useState<{ customerCount: number } | null>(null)
  const [alerts, setAlerts] = useState<MonitorReview[]>([])
  const [pending, setPending] = useState<number | null>(null)

  useEffect(() => {
    let alive = true
    api.getTodayBoard().then((x) => { if (alive) setBoard(x) }).catch(() => {})
    api.getFinanceSummary().then((x) => { if (alive) setFin(x) }).catch(() => {})
    api.getDashboardSummary().then((x) => { if (alive) setSummary(x) }).catch(() => {})
    api.getMonitorReviews({ status: 'PENDING', limit: 30 })
      .then((rows) => { if (alive) setAlerts([...rows].sort((a, b) => Number(b.urgent) - Number(a.urgent)).slice(0, 4)) }).catch(() => {})
    api.getMonitorReviewCounts().then((c) => { if (alive) setPending(c.reduce((s, x) => s + x.count, 0)) }).catch(() => {})
    return () => { alive = false }
  }, [])

  const ctx = { employee, permissions, gpsServiceId }
  // رابط بس إذا يگدر يفتحه — وإلا البطاقة رقم بلا رابط.
  const linkIf = (to: string) => (isPathAllowed(to, ctx) === false ? null : to)
  // الأنشطة السريعة: تطلع بس الي يگدر يفتحها (نفس حارس الروابط) — زر يودّي لـ«ما عندك صلاحية» أسوأ من غيابه.
  const tiles = [
    { to: '/work-reports-review', icon: '📈', label: 'التقارير' },
    { to: '/my-extra-tasks', icon: '📋', label: 'المهام الموكلة' },
    { to: '/missions', icon: '📅', label: 'حجوزات اليوم' },
    { to: '/attendance', icon: '🕐', label: 'جدول الدوام' },
    { to: '/monitor-inbox', icon: '📝', label: 'ملاحظات المراقب' },
    { to: '/complaints', icon: '💬', label: 'شكاوى العملاء' },
  ].filter((t) => isPathAllowed(t.to, ctx) !== false)

  const change = useMemo(() => {
    const p = board?.last14 ?? []
    if (p.length < 14) return null
    const a = p.slice(0, 7).reduce((s, x) => s + x.count, 0), b = p.slice(7).reduce((s, x) => s + x.count, 0)
    return a > 0 ? Math.round(((b - a) / a) * 100) : null
  }, [board])
  const success = board && board.weekTotal > 0 ? Math.round((board.weekCompleted / board.weekTotal) * 100) : null

  return (
    <div dir="rtl" className="space-y-4">
      <section className="flex flex-wrap items-center justify-between gap-4 rounded-3xl bg-gradient-to-l from-[#0f2040] via-[#1e3a8a] to-[#1e40af] p-6 text-white shadow-lg">
        <div>
          <span className="inline-flex items-center gap-1.5 rounded-full bg-white/15 px-3 py-0.5 text-xs"><i className="h-2 w-2 rounded-full bg-emerald-400" /> متصل</span>
          <h2 className="mt-2 text-3xl font-extrabold">مرحباً، {employee?.name}</h2>
          <p className="mt-1 text-sm opacity-80">هذه نظرة سريعة على عملك اليوم</p>
        </div>
        <Clock />
      </section>

      <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        <StatCard label="حجوزات اليوم" value={board ? fmt(board.bookingsToday) : '…'} to={linkIf('/missions')} tone="blue" icon="📅" />
        <StatCard label="مكتملة اليوم" value={board ? fmt(board.completedToday) : '…'} to={linkIf('/bookings') && '/bookings?tab=done'} tone="green" icon="✅" />
        <StatCard label="قيد التنفيذ" value={board ? fmt(board.inField) : '…'} to={linkIf('/missions')} tone="amber" icon="⏳" />
        <StatCard label="تحتاج متابعة" value={pending ?? '…'} to={linkIf('/monitor-inbox')} tone="red" icon="❗" />
      </div>

      <div className="grid gap-4 lg:grid-cols-5">
        <section className={`${card} lg:col-span-3`}>
          <div className="mb-2 flex items-center justify-between">
            <h3 className="font-extrabold text-[#0f2040]">حركة الحجوزات — آخر ١٤ يوم</h3>
            {change != null && <span className={`rounded-full px-2.5 py-0.5 text-xs font-bold ${change >= 0 ? 'bg-emerald-50 text-emerald-700' : 'bg-red-50 text-red-700'}`}>{change >= 0 ? '↗' : '↘'} {Math.abs(change)}% مقارنة بالأسبوع السابق</span>}
          </div>
          <Chart points={board?.last14 ?? []} />
        </section>

        <section className={`${card} lg:col-span-2`}>
          <h3 className="mb-3 font-extrabold text-[#0f2040]">⚠️ أهم الملاحظات والتنبيهات</h3>
          {alerts.length === 0 ? <p className="text-sm text-slate-400">ماكو تنبيهات معلّقة ✅</p> : (
            <ul className="divide-y divide-slate-100">
              {alerts.map((a) => (
                <li key={a.id} className="flex items-center justify-between gap-2 py-2.5">
                  <div className="min-w-0">
                    <p className="truncate text-sm font-bold text-slate-800">{a.title}</p>
                    <p className="truncate text-xs text-slate-500">{a.summary ?? ''}</p>
                  </div>
                  <div className="shrink-0 text-left">
                    <span className={`rounded-full px-2 py-0.5 text-[11px] font-bold ${a.urgent ? 'bg-red-50 text-red-700' : 'bg-amber-50 text-amber-700'}`}>{a.urgent ? 'عاجل' : 'مطلوب متابعة'}</span>
                    <p className="mt-0.5 text-[10px] text-slate-400">{ago(a.createdAt)}</p>
                  </div>
                </li>
              ))}
            </ul>
          )}
          <Link to="/monitor-inbox" className="mt-2 inline-block text-xs font-bold text-brand-600 hover:underline">← عرض جميع التنبيهات</Link>
        </section>
      </div>

      <div className="grid gap-4 lg:grid-cols-3">
        <section className={card}>
          <h3 className="mb-3 font-extrabold text-[#0f2040]">🏆 أفضل الموظفين هذا الشهر</h3>
          {(board?.topCrew ?? []).length === 0 ? <p className="text-sm text-slate-400">ماكو بيانات بعد.</p> : (
            <ul className="space-y-3">
              {(board?.topCrew ?? []).slice(0, 3).map((c, i) => {
                const pct = c.visits > 0 ? Math.round((c.done / c.visits) * 100) : 0
                return (
                  <li key={c.employeeId} className="flex items-center gap-3">
                    <span className="text-lg">{['🥇', '🥈', '🥉'][i]}</span>
                    <Link to={`/matrix/employee/${c.employeeId}`} className="w-28 truncate text-sm font-bold text-slate-800 hover:underline">{c.name}</Link>
                    <div className="h-2 flex-1 rounded-full bg-slate-100"><div className="h-2 rounded-full bg-gradient-to-l from-emerald-500 to-sky-500" style={{ width: `${pct}%` }} /></div>
                    <span className="w-16 text-left text-xs text-slate-500"><b className="text-slate-800">{c.visits}</b> حجز · {pct}%</span>
                  </li>
                )
              })}
            </ul>
          )}
        </section>

        <section className={card}>
          <h3 className="mb-3 font-extrabold text-[#0f2040]">⚡ أنشطة سريعة</h3>
          <div className="grid grid-cols-3 gap-2">
            {tiles.map((t) => (
              <Link key={t.to} to={t.to} className="flex flex-col items-center gap-1 rounded-xl border border-slate-200 bg-slate-50 p-3 text-center text-xs font-bold text-slate-700 transition hover:border-brand-300 hover:bg-white">
                <span className="text-xl">{t.icon}</span>{t.label}
              </Link>
            ))}
          </div>
        </section>

        <section className={card}>
          <h3 className="mb-3 font-extrabold text-[#0f2040]">📊 إحصائيات سريعة</h3>
          <ul className="divide-y divide-slate-100 text-sm">
            <li className="flex items-center justify-between py-2.5"><span>👥 عدد العملاء</span><b className="text-lg">{summary ? fmt(summary.customerCount) : '…'}</b></li>
            <li className="flex items-center justify-between py-2.5"><span>💵 إجمالي الإيرادات</span><b className="text-lg">{fin ? `${fmt(Math.round(fin.totalCollected))} د.ع` : '…'}</b></li>
            <li className="flex items-center justify-between py-2.5"><span>🎯 نسبة النجاح (أسبوع)</span><b className="text-lg">{success != null ? `${success}%` : '—'}</b></li>
          </ul>
        </section>
      </div>
    </div>
  )
}
