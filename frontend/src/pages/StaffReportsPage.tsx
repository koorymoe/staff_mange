import { useState } from 'react'
import { Link } from 'react-router-dom'
import { apiRequest } from '../api'
import { useSession, isAdminOrOwner, hasRole } from '../session'
import { useApi } from '../utils/useApi'
import { ReportAnalysisBox } from '../components/ReportAnalysisBox'
import { DeleteButton } from '../ui/DeleteButton'
import { card } from '../ui/classes'
import { num, todayBaghdad } from '../utils/format'
import { swallow } from '../utils/swallow'

// ═══ 🧾 تقارير الموظفين (دازاي 11-02 — تعديلات ٩.١) ═══
// تقارير الإنجاز اليومية (منو كتب، ومنو داوم وما كتب) + استمارات معلومات
// الموظفين ويا الفروقات عن ملفهم (للكوادر والمالك بس — بيها رواتب).

type Daily = { day: string; reports: { id: string; employeeId: string; employeeName: string; body: string; late: boolean; createdAt: string; matchStatus: string | null; matchNote: string | null; matchVerdict: string | null; matchReview: string | null }[]; missing: { employeeId: string; employeeName: string }[] }
type Survey = { employeeId: string; employeeName: string; major: string; degree: string; selfSalaryIQD: number | null; fileSalary: number | null; brought: number | null; monthlyHours: number | null; currentJob: string; companyRating: number; companyNote: string; contractEnd: string | null; fileContract: string | null; mismatch: string | null; reviewedAt: string | null }
type Surveys = { surveys: Survey[]; pending: { employeeId: string; employeeName: string }[] }

// (ك) 11-21 زينا: تبويبات بإطار، والمختار إطاره ذهبي (مثل التصميم)
const tabCls = (on: boolean) => `rounded-2xl border bg-white px-6 py-3 text-base font-bold transition ${on ? 'border-amber-400 text-[#0f2040] ring-2 ring-amber-300/60' : 'border-slate-200 text-slate-600 hover:border-slate-300'}`

const yesterday = () => { const d = new Date(todayBaghdad() + 'T12:00:00'); d.setDate(d.getDate() - 1); return d.toISOString().slice(0, 10) }

export default function StaffReportsPage() {
  const { employee, permissions } = useSession()
  const canSurveys = isAdminOrOwner(employee) || hasRole(employee, 'HR_COORDINATOR') || permissions.includes('staff_management')
  const [tab, setTab] = useState<'daily' | 'surveys' | 'negligence'>('daily')
  return (
    <div dir="rtl" className="space-y-4">
      <h1 className="text-2xl font-extrabold text-[#0f2040]">🧾 تقارير الموظفين</h1>
      <div className="flex gap-2">
        <button onClick={() => setTab('daily')} className={tabCls(tab === 'daily')}>📝 تقارير الإنجاز اليومية</button>
        <button onClick={() => setTab('negligence')} className={tabCls(tab === 'negligence')}>📒 سجل الإهمال</button>
        {canSurveys && <button onClick={() => setTab('surveys')} className={tabCls(tab === 'surveys')}>🧾 استمارات المعلومات</button>}
      </div>
      {tab === 'daily' ? <DailyTab /> : tab === 'negligence' ? <NegligenceTab /> : canSurveys && <SurveysTab />}
    </div>
  )
}

function DailyTab() {
  const [day, setDay] = useState(yesterday())
  const [reload, setReload] = useState(0)
  // (ك) 11-21 زينا: حذف التقرير للمالك بس (الباك إند يرفض غيره)
  const isOwner = useSession().employee?.actualRole === 'OWNER'
  const { data, error } = useApi(() => apiRequest<Daily>(`/daily-reports?day=${day}`), [day, reload])
  return (
    <div className={`${card} space-y-3`}>
      <div className="flex flex-wrap items-center gap-3">
        <input type="date" value={day} onChange={(e) => e.target.value && setDay(e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1.5 text-sm" />
        {data && <span className="text-sm text-slate-500">كتبوا: <b>{data.reports.length}</b> · داوموا وما كتبوا: <b className={data.missing.length ? 'text-red-600' : ''}>{data.missing.length}</b></span>}
      </div>
      {error && <p className="text-sm text-red-600">{error}</p>}
      {data && data.missing.length > 0 && (
        <div className="rounded-xl bg-red-50 p-3 text-sm text-red-800"><b>داوموا وما كتبوا تقريرهم:</b> {data.missing.map((m) => m.employeeName).join('، ')} <span className="text-xs">(تنقفل عليهم الشاشات لحد ما يكتبون)</span></div>
      )}
      <div className="divide-y divide-slate-100">
        {data?.reports.map((r) => (
          <div key={r.id} className="py-2.5">
            <p className="text-sm font-bold text-[#0f2040]"><Link to={`/employee-dossier?id=${r.employeeId}`} className="hover:underline">{r.employeeName}</Link> {r.late && <span className="mr-1 rounded-full bg-amber-50 px-2 py-0.5 text-[11px] font-bold text-amber-700 ring-1 ring-amber-200">متأخر</span>}
              {isOwner && <DeleteButton className="float-left" confirm={`تحذف تقرير ${r.employeeName}؟ ينحذف وياه تحليل ماتركس، وما يرجع.`}
                onConfirm={async () => { await apiRequest(`/daily-reports/${r.id}`, { method: 'DELETE' }); setReload((n) => n + 1) }} />}</p>
            <p className="whitespace-pre-wrap text-sm text-slate-700">{r.body}</p>
            <MatchRow r={r} onDone={() => setReload((n) => n + 1)} />
            <ReportAnalysisBox reportId={r.id} />
          </div>
        ))}
        {data && data.reports.length === 0 && <p className="py-4 text-center text-sm text-slate-400">ماكو تقارير لهذا اليوم.</p>}
      </div>
    </div>
  )
}

// (ك) 11-21 زينا: أيقونات خط بسيطة لرأس جدول الاستمارات (ماكو مكتبة أيقونات بالمشروع)
const ICONS: Record<string, string> = {
  user: 'M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8Zm-7 9a7 7 0 0 1 14 0',
  users: 'M9 11a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7Zm-6 9a6 6 0 0 1 12 0m1-9a3 3 0 1 0 0-6m2 15a5 5 0 0 0-3-4.6',
  cap: 'M2 9l10-5 10 5-10 5L2 9Zm4 2v5c2 2 10 2 12 0v-5M22 9v6',
  bag: 'M4 8h16v11H4zM9 8V5h6v3M4 13h16',
  coins: 'M12 7c4 0 7-1 7-2.5S16 2 12 2 5 3 5 4.5 8 7 12 7Zm-7-2.5v5C5 11 8 12 12 12s7-1 7-2.5v-5M5 9.5v5C5 16 8 17 12 17s7-1 7-2.5v-5M5 14.5v5C5 21 8 22 12 22s7-1 7-2.5v-5',
  clock: 'M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18Zm0-13v5l3 2',
  cal: 'M4 6h16v14H4zM4 10h16M8 3v4m8-4v4',
  star: 'M12 3l2.7 5.6 6.1.9-4.4 4.3 1 6.1L12 17l-5.4 2.9 1-6.1L3.2 9.5l6.1-.9L12 3Z',
  pulse: 'M3 12h4l2-5 4 10 2-5h6',
  chart: 'M4 20V10m6 10V4m6 16v-7m4 7H2',
}
const Ico = ({ n, className = 'h-4 w-4' }: { n: string; className?: string }) => (
  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" className={className} aria-hidden="true"><path d={ICONS[n]} /></svg>
)

// (ك) 11-21 زينا: الهدف — الفني أو الليدر يدخّل بحجوزاته ١٣٠٪ من راتبه (راتبه + ٣٠٪)
const TARGET = 1.3
function broughtCell(s: Survey) {
  if (s.brought === null) return <span className="text-slate-400">—</span>
  const sal = s.fileSalary ?? 0
  const pct = sal > 0 ? Math.round((s.brought / sal) * 100) : null
  const ok = sal > 0 && s.brought >= sal * TARGET
  return (
    <div className="space-y-1.5" title={sal > 0 ? `الهدف: ${num(Math.round(sal * TARGET))} (راتبه + ٣٠٪)` : 'ماكو راتب بالملف'}>
      <div className="text-base font-bold text-slate-800" dir="ltr">{num(Math.round(s.brought))}</div>
      {pct === null
        ? <span className="text-xs text-slate-400">ماكو راتب بالملف</span>
        : <span className={`inline-block whitespace-nowrap rounded-full px-3 py-1 text-xs font-bold ring-1 ${ok ? 'bg-emerald-50 text-emerald-700 ring-emerald-300' : 'bg-red-50 text-red-700 ring-red-200'}`}>
            {ok ? '✓' : '✗'} {pct}٪ من راتبه
          </span>}
    </div>
  )
}

function SurveysTab() {
  const [reload, setReload] = useState(0)
  const { data, error } = useApi(() => apiRequest<Surveys>('/profile-surveys'), [reload])
  const review = async (id: string) => { await apiRequest(`/profile-surveys/${id}/review`, { method: 'POST' }).catch(swallow); setReload((n) => n + 1) }
  const heads: [string, string][] = [['', '#'], ['user', 'الموظف'], ['cap', 'التخصص / الشهادة'], ['bag', 'الوظيفة'], ['coins', 'الراتب (د.ع)'], ['chart', 'دخّل هالشهر'], ['clock', 'ساعات العقد (ساعة)'], ['cal', 'تاريخ العقد'], ['star', 'رأيه بالشركة'], ['pulse', 'الحالة']]
  const dash = <span className="text-slate-400">—</span>
  return (
    <div className={`${card} space-y-4`}>
      {error && <p className="text-sm text-red-600">{error}</p>}
      {data && data.pending.length > 0 && (
        <div className="flex items-center gap-3 rounded-2xl border border-slate-200 bg-slate-50 px-5 py-4 text-base text-slate-700">
          <span className="text-amber-500"><Ico n="users" className="h-8 w-8" /></span>
          <p><b className="text-lg text-[#0f2040]">بعدهم ما عبّوا ({data.pending.length}):</b> {data.pending.map((p) => p.employeeName).join('، ')}</p>
        </div>
      )}
      <div className="overflow-x-auto rounded-2xl border border-slate-200" data-testid="surveys-table">
        <table className="w-full min-w-[1200px] text-[15px]">
          <thead className="bg-slate-50 text-sm font-bold text-[#0f2040]">
            <tr>
              {heads.map(([ic, h]) => (
                <th key={h} className="px-3 py-5 text-center">
                  <span className="inline-flex items-center justify-center gap-2">{ic && <span className="text-amber-500"><Ico n={ic} className="h-6 w-6" /></span>}<span className="whitespace-nowrap">{h}</span></span>
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-200">
            {data?.surveys.map((s, i) => {
              const flag = !!s.mismatch && !s.reviewedAt
              return (
                <tr key={s.employeeId} className={`align-middle ${flag ? 'bg-red-50' : 'hover:bg-slate-50'}`}>
                  <td className="px-3 py-5 text-center"><span className="inline-flex h-9 w-9 items-center justify-center rounded-full bg-slate-100 text-sm ring-1 ring-slate-200 font-bold text-slate-600">{i + 1}</span></td>
                  <td className="px-3 py-5"><span className="flex items-center gap-3 whitespace-nowrap text-base font-extrabold text-[#0f2040]"><span className="shrink-0 text-amber-500"><Ico n="user" className="h-7 w-7" /></span>{s.employeeName}</span></td>
                  <td className="px-3 py-5 text-center text-slate-700">{s.major} / {s.degree}</td>
                  <td className="px-3 py-5 text-center">{s.currentJob ? <span className="inline-block rounded-full bg-blue-50 px-4 py-1.5 text-sm font-bold text-blue-700 ring-1 ring-blue-200">{s.currentJob}</span> : dash}</td>
                  <td className="px-3 py-5 text-center text-base font-bold leading-relaxed text-slate-800" dir="ltr" title="هو / الملف">
                    <div>{s.selfSalaryIQD !== null ? num(s.selfSalaryIQD) : '—'} /</div>
                    <div>{s.fileSalary !== null ? num(s.fileSalary) : '—'}</div>
                  </td>
                  {/* (ك) 11-21 زينا: شكد دخّل فلوس هالشهر من حجوزاته (ليدر 60٪، فني 40٪) — والهدف ٣٠٪ فوك راتبه */}
                  <td className="px-3 py-5 text-center">{broughtCell(s)}</td>
                  <td className="px-3 py-5 text-center text-base font-bold text-slate-700">{s.monthlyHours ?? dash}</td>
                  <td className="px-3 py-5 text-center" title="هو / الملف">
                    {s.contractEnd || s.fileContract
                      ? <span className="inline-block whitespace-nowrap rounded-xl bg-slate-100 px-4 py-2 text-base font-bold ring-1 ring-slate-200 text-slate-700" dir="ltr">{s.contractEnd ?? '—'}{s.fileContract && s.fileContract !== s.contractEnd ? ` / ${s.fileContract}` : ''}</span>
                      : <span className="text-slate-400">—/—</span>}
                  </td>
                  <td className="px-3 py-5 text-center">
                    <div className="text-2xl leading-none tracking-tight text-amber-400" aria-label={`${s.companyRating} من ٥`}>{'★'.repeat(s.companyRating)}</div>
                    {s.companyNote && <div className="mt-1.5 text-sm text-slate-500">{s.companyNote}</div>}
                  </td>
                  <td className="px-3 py-5 text-center">
                    {flag && (
                      <>
                        <div className="mb-2 text-sm font-bold text-amber-600">⚠️ {s.mismatch}</div>
                        <button type="button" onClick={() => void review(s.employeeId)} className="rounded-xl border border-slate-300 bg-white px-4 py-1.5 text-sm font-bold text-slate-700 hover:bg-slate-50">✓ طابقت</button>
                      </>
                    )}
                    {!flag && <span className="inline-block rounded-full bg-emerald-50 px-5 py-2 text-sm font-bold text-emerald-700 ring-1 ring-emerald-300">✓ {s.reviewedAt ? 'انطابقت' : 'طابق'}</span>}
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
      {data && data.surveys.length === 0 && <p className="py-4 text-center text-sm text-slate-400">بعد ماكو استمارات معبّاية.</p>}
    </div>
  )
}

type Neg = { id: string; employeeName: string; day: string; kind: string; detail: string; reason: string; compensatedAt: string | null }

// ٩.٢ سجل الإهمال: كل مقيّم فاته تقييم — يتسجّل بالتاريخ، ويا سببه ومتى عوّض.
function NegligenceTab() {
  const { data, error } = useApi(() => apiRequest<Neg[]>('/negligence'), [])
  return (
    <div className={`${card} space-y-2`}>
      <p className="text-xs text-slate-500">كل مقيّم فاته يقيّم الي عليه بيومه ينكتب هنا. ينزل تقييمه هو بماتركس (عوّض = نص، ما عوّض = صفر) — بلا خصم فلوس ولا نقاط انضباط.</p>
      {error && <p className="text-sm text-red-600">{error}</p>}
      <div className="divide-y divide-slate-100">
        {data?.map((n) => (
          <div key={n.id} className="flex flex-wrap items-start gap-2 py-2 text-sm">
            <span className="w-24 shrink-0 font-mono text-xs text-slate-500" dir="ltr">{n.day}</span>
            <span className="w-40 shrink-0 font-bold text-[#0f2040]">{n.employeeName}</span>
            <span className="flex-1 text-slate-700">{n.detail}{n.reason && <span className="block text-xs text-slate-500">السبب: {n.reason}</span>}</span>
            {n.compensatedAt
              ? <span className="rounded-full bg-amber-50 px-2 py-0.5 text-[11px] font-bold text-amber-700 ring-1 ring-amber-200">عوّض متأخر</span>
              : <span className="rounded-full bg-red-50 px-2 py-0.5 text-[11px] font-bold text-red-700 ring-1 ring-red-200">ما عوّض</span>}
          </div>
        ))}
        {data && data.length === 0 && <p className="py-6 text-center text-emerald-700">✅ ماكو إهمال مسجّل</p>}
      </div>
    </div>
  )
}

type Report = Daily['reports'][number]
type MatchDetail = { status: string; note: string; claimedCodes: string[]; missingCodes: string[]; unclaimedCodes: string[]; activity: { at: string; text: string; count: number }[] }

const MATCH_BADGE: Record<string, [string, string]> = {
  MATCH: ['✅ يطابق الشغل المسجّل', 'bg-emerald-50 text-emerald-700 ring-emerald-200'],
  PARTIAL: ['⚠️ يطابق جزئياً', 'bg-amber-50 text-amber-700 ring-amber-200'],
  NO_TRACE: ['❌ بلا أثر بالنظام — ما ينحسب لحد ما تحكم', 'bg-red-50 text-red-700 ring-red-200'],
}

// ٩.٤ ماتركس يطابق التقرير ويا الشغل المسجّل — والمراقب يحكم على الفرق.
function MatchRow({ r, onDone }: { r: Report; onDone: () => void }) {
  const [open, setOpen] = useState(false)
  const [note, setNote] = useState('')
  const [err, setErr] = useState('')
  const { data } = useApi(() => apiRequest<MatchDetail>(`/daily-reports/${r.id}/match`), [r.id], { enabled: open })
  const status = data?.status ?? r.matchStatus
  if (!status) return <button onClick={() => setOpen(true)} className="mt-1 text-xs text-slate-400 hover:underline">🔎 طابق ويا الشغل المسجّل</button>
  const [label, cls] = MATCH_BADGE[status] ?? [status, 'bg-slate-50 text-slate-600 ring-slate-200']
  const review = async (verdict: string) => {
    setErr('')
    try { await apiRequest(`/daily-reports/${r.id}/match-review`, { method: 'POST', body: JSON.stringify({ verdict, note }) }); onDone() } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر الحفظ') }
  }
  return (
    <div className="mt-1.5 text-xs">
      <button onClick={() => setOpen((o) => !o)} className={`rounded-full px-2 py-0.5 font-bold ring-1 ${cls}`}>{label} {open ? '▴' : '▾'}</button>
      {r.matchVerdict && <span className={`mr-2 font-bold ${r.matchVerdict === 'JUSTIFIED' ? 'text-emerald-700' : 'text-red-700'}`}>المراقب: {r.matchVerdict === 'JUSTIFIED' ? 'مبرّر' : 'مو مبرّر'}{r.matchReview && ` — ${r.matchReview}`}</span>}
      {open && (
        <div className="mt-2 space-y-1.5 rounded-xl bg-slate-50 p-3 text-slate-700">
          <p className="font-bold">{data?.note ?? r.matchNote}</p>
          {data && data.missingCodes.length > 0 && <p>ذكرها وما لمسها: <b dir="ltr">{data.missingCodes.join('، ')}</b></p>}
          {data && data.unclaimedCodes.length > 0 && <p>اشتغل عليها وما ذكرها: <b dir="ltr">{data.unclaimedCodes.join('، ')}</b></p>}
          {data && <div><b>المسجّل بالنظام:</b>{data.activity.length === 0 ? <span className="text-slate-400"> ماكو</span> : (
            <ul className="mt-1 list-inside list-disc">{data.activity.map((a, i) => <li key={i}><span dir="ltr" className="font-mono text-slate-400">{a.at}</span> {a.text}{a.count > 1 && ` ×${a.count}`}</li>)}</ul>
          )}</div>}
          {status !== 'MATCH' && !r.matchVerdict && (
            <div className="flex flex-wrap items-center gap-2 pt-1">
              <input value={note} onChange={(e) => setNote(e.target.value)} placeholder="ملاحظة (اختياري)" className="min-w-40 flex-1 rounded-lg border border-slate-300 px-2 py-1" />
              <button onClick={() => void review('JUSTIFIED')} className="rounded-lg bg-emerald-600 px-3 py-1 font-bold text-white">✓ مبرّر</button>
              <button onClick={() => void review('UNJUSTIFIED')} className="rounded-lg bg-red-600 px-3 py-1 font-bold text-white">✗ مو مبرّر</button>
            </div>
          )}
          {err && <p className="text-red-600">{err}</p>}
        </div>
      )}
    </div>
  )
}
