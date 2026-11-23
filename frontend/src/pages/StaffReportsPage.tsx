import { useEffect, useRef, useState } from 'react'
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

type Daily = { day: string; reports: { id: string; employeeId: string; employeeName: string; day: string; body: string; late: boolean; createdAt: string; matchStatus: string | null; matchNote: string | null; matchVerdict: string | null; matchReview: string | null }[]; missing: { employeeId: string; employeeName: string }[] }
type Survey = { employeeId: string; employeeName: string; major: string; degree: string; selfSalaryIQD: number | null; fileSalary: number | null; brought: number | null; monthlyHours: number | null; currentJob: string; companyRating: number; companyNote: string; contractEnd: string | null; fileContract: string | null; mismatch: string | null; reviewedAt: string | null }
type Surveys = { surveys: Survey[]; pending: { employeeId: string; employeeName: string }[] }

// (ك) 11-22 زينا: تبويبات كبيرة بأيقونة، والمختار إطاره أزرق متوهّج (مثل التصميم) — الألوان بـindex.css (.sr-tab)
const tabCls = (on: boolean) => `sr-tab flex items-center justify-center gap-2.5 rounded-2xl border px-8 py-2.5 text-base font-bold transition ${on ? 'sr-tab-on' : ''}`

const yesterday = () => { const d = new Date(todayBaghdad() + 'T12:00:00'); d.setDate(d.getDate() - 1); return d.toISOString().slice(0, 10) }

export default function StaffReportsPage() {
  const { employee, permissions } = useSession()
  const canSurveys = isAdminOrOwner(employee) || hasRole(employee, 'HR_COORDINATOR') || permissions.includes('staff_management')
  const [tab, setTab] = useState<'daily' | 'surveys' | 'negligence'>('daily')
  return (
    <div dir="rtl" className="space-y-4">
      <div className="flex items-center gap-4">
        <span className="sr-logo flex h-16 w-16 shrink-0 items-center justify-center rounded-2xl"><Ico n="clipList" className="h-9 w-9" /></span>
        <div>
          <h1 className="text-3xl font-extrabold text-[#0f2040]">تقارير الموظفين</h1>
          <p className="mt-0.5 text-[15px] text-slate-500">متابعة تقارير الأداء والإنجاز اليومي للموظفين</p>
        </div>
      </div>
      <div className="flex flex-wrap gap-4">
        <button onClick={() => setTab('daily')} className={tabCls(tab === 'daily')}><Ico n="fileCheck" className="h-6 w-6" />تقارير الإنجاز اليومية</button>
        <button onClick={() => setTab('negligence')} className={tabCls(tab === 'negligence')}><span className="text-xl leading-none">📒</span>سجل الإهمال</button>
        {canSurveys && <button onClick={() => setTab('surveys')} className={tabCls(tab === 'surveys')}><Ico n="file" className="h-6 w-6" />استمارات المعلومات</button>}
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
  // (دازاي 11-21) «اريد احذف الزايدات» — زر «كل تقاريره» يعرض تقارير موظف واحد (آخر ٩٠ يوم)
  const [emp, setEmp] = useState<{ id: string; name: string } | null>(null)
  const { data, error } = useApi(() => apiRequest<Daily>(emp ? `/daily-reports?employeeId=${emp.id}` : `/daily-reports?day=${day}`), [day, reload, emp?.id])
  // (ك) 11-22 زينا: تصميم التقارير اليومية — شريط التاريخ والعدّاد، تنبيه الي ما كتبوا، وكل تقرير ببطاقة
  return (
    <div className="sr-panel space-y-4 rounded-3xl border p-5">
      <div className="flex flex-wrap items-center gap-4">
        {emp ? (
          <span className="flex items-center gap-2 text-sm"><b className="text-[#0f2040]">كل تقارير {emp.name}</b> ({data?.reports.length ?? '…'})
            <button type="button" onClick={() => setEmp(null)} className="sr-chip rounded-xl border px-3 py-1.5 text-xs font-bold">↩ رجوع لليوم</button></span>
        ) : (
          <label className="sr-box flex w-full items-center sm:w-72 gap-3 rounded-2xl border px-4 py-2.5">
            <Ico n="cal" className="h-6 w-6 shrink-0" />
            <input type="date" value={day} onChange={(e) => e.target.value && setDay(e.target.value)} className="w-full bg-transparent text-center text-lg font-bold outline-none" />
          </label>
        )}
        {data && !emp && (
          <div className="sr-box mx-auto flex items-center gap-4 rounded-2xl border px-5 py-2.5 text-base font-bold sm:gap-8 sm:px-10 sm:text-lg">
            <span className="flex items-center gap-2"><Ico n="file" className="h-7 w-7" />كتبوا: <b className="text-3xl font-extrabold text-emerald-500">{data.reports.length}</b></span>
            <span className="h-7 w-px bg-current opacity-30" />
            <span className="flex items-center gap-2"><Ico n="file" className="h-7 w-7" />ما كتبوا: <b className={`text-3xl font-extrabold ${data.missing.length ? 'text-red-500' : 'text-emerald-500'}`}>{data.missing.length}</b></span>
          </div>
        )}
      </div>
      {error && <p className="text-sm text-red-600">{error}</p>}
      {data && data.missing.length > 0 && (
        <div className="sr-alert flex items-center gap-3 rounded-2xl border px-4 py-3 text-sm">
          <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-red-500 text-base font-extrabold text-white">!</span>
          <p><b className="text-base">داوموا وما كتبوا تقريرهم:</b> {data.missing.map((m) => m.employeeName).join('، ')} <span className="text-xs opacity-80">(تنقفل عليهم الشاشات لحد ما يكتبون)</span></p>
        </div>
      )}
      <div className="space-y-3">
        {data?.reports.map((r) => (
          <div key={r.id} className="sr-card flex flex-col gap-3 rounded-2xl border p-3 md:flex-row md:items-stretch md:gap-4">
            <div className="flex shrink-0 items-center gap-3 md:w-72 md:items-start md:pt-1">
              <span className="sr-ic flex h-16 w-16 shrink-0 items-center justify-center rounded-2xl"><Ico n="layers" className="h-8 w-8" /></span>
              <div className="min-w-0 space-y-1.5">
                <Link to={`/employee-dossier?id=${r.employeeId}`} className="flex items-center gap-2 text-lg font-bold text-[#0f2040] hover:underline"><Ico n="user" className="h-6 w-6 shrink-0 opacity-80" />{r.employeeName}</Link>
                {emp ? <span className="text-xs text-slate-500">{r.day}</span>
                  : <button type="button" onClick={() => setEmp({ id: r.employeeId, name: r.employeeName })} className="sr-chip inline-flex items-center gap-1.5 rounded-xl border px-2.5 py-1 text-xs font-bold"><Ico n="file" className="h-4 w-4" />كل تقاريره</button>}
              </div>
            </div>
            <div className="sr-sep hidden w-px shrink-0 md:block" />
            <div className="min-w-0 flex-1 space-y-2 py-1">
              <ReportBody body={r.body} late={r.late} />
              <div className="flex flex-wrap items-start gap-3">
                <MatchRow r={r} onDone={() => setReload((n) => n + 1)} />
                <ReportAnalysisBox reportId={r.id} />
              </div>
            </div>
            {isOwner && <>
              <div className="sr-sep hidden w-px shrink-0 md:block" />
              <div className="flex shrink-0 items-center justify-end md:items-start md:px-3 md:pt-1">
                <DeleteButton className="sr-del" label="🗑️ حذف" confirm={`تحذف تقرير ${r.employeeName}؟ ينحذف وياه تحليل ماتركس، وما يرجع.`}
                  onConfirm={async () => { await apiRequest(`/daily-reports/${r.id}`, { method: 'DELETE' }); setReload((n) => n + 1) }} />
              </div>
            </>}
          </div>
        ))}
        {data && data.reports.length === 0 && <p className="py-4 text-center text-sm text-slate-400">{emp ? 'ماكو تقارير إله.' : 'ماكو تقارير لهذا اليوم.'}</p>}
      </div>
    </div>
  )
}

// (ك) 11-22 زينا: نص التقرير — الأسطر الي تبدي بـ«-» أو «•» أو «✓» تطلع بعلامة صح خضرة (مثل التصميم)، والباقي سطر عادي
// (ك) 11-23 زينا: ترتيب التقرير مثل الصورة — أول سطر عنوان (وجنبه «متأخر»)، وإذا كتب التاريخ/وقت الدوام/تفاصيل
// الأعمال تطلع بشريط بأيقونات، وبعدها التفاصيل: «تم…» بعلامة خضرة وخط فاصل، والباقي بنقطة زرقة.
// التقرير الطويل ينطوي على سطرين، و«عرض التقرير كامل» يفتحه.
const META: [RegExp, string][] = [[/^التاريخ\s*:?\s*/, 'cal'], [/^وقت\s+بدء\s+الدوام\s*:?\s*/, 'clock'], [/^تفاصيل\s+الأعمال\s+المنجزة\s*:?\s*/, 'bag']]
const isDone = (l: string) => /^[-•✓*]\s*/.test(l) || /^(تم|تمت|أنجزت|انجزت)\s/.test(l)
function ReportBody({ body, late }: { body: string; late?: boolean }) {
  const [full, setFull] = useState(false)
  const lines = body.split('\n').map((l) => l.trim()).filter(Boolean)
  const meta: { ic: string; text: string }[] = []
  const rest: string[] = []
  for (const l of lines) {
    const m = META.find(([re]) => re.test(l))
    if (m) meta.push({ ic: m[1], text: l.includes(':') ? l : `${l}:` })
    else rest.push(l)
  }
  const title = rest.length > 0 && !isDone(rest[0]) ? rest.shift() : undefined
  const fold = meta.length > 0 && rest.length > 3 && !full
  const shown = fold ? rest.slice(0, 2) : rest
  const allDone = shown.length > 0 && shown.every(isDone)
  return (
    <div className="min-w-0 flex-1 space-y-2 text-[15px] leading-7 text-slate-700">
      {(title || late) && (
        <div className="flex flex-wrap items-center justify-center gap-x-10 gap-y-2">
          {title && <p className="text-base font-extrabold text-[#0f2040]">{title}</p>}
          {late && <span className="inline-flex shrink-0 items-center gap-1.5 rounded-full bg-amber-50 px-3 py-0.5 text-sm font-bold text-amber-700 ring-1 ring-amber-400"><Ico n="clock" className="h-4 w-4" />متأخر</span>}
        </div>
      )}
      {meta.length > 0 && (
        <div className="sr-meta flex flex-wrap items-stretch rounded-xl border">
          {meta.map((x, i) => (
            <span key={i} className={`flex flex-1 items-center justify-center gap-2 px-4 py-1.5 text-sm ${i ? 'sr-meta-sep border-r' : ''}`}><Ico n={x.ic} className="h-5 w-5 shrink-0" />{x.text}</span>
          ))}
        </div>
      )}
      {shown.length > 0 && (
        <div className={allDone ? 'sr-done-list' : 'space-y-0'}>
          {shown.map((l, i) => {
            const t = l.replace(/^[-•✓*]\s*/, '')
            return allDone
              ? <p key={i} className="sr-done flex items-start gap-2"><span className="mt-1.5 flex h-4 w-4 shrink-0 items-center justify-center rounded-full bg-emerald-500 text-[10px] font-extrabold text-white">✓</span>{t}</p>
              : <p key={i} className={`flex items-start gap-2 ${fold ? 'truncate' : ''}`}><span className="mt-3 h-1.5 w-1.5 shrink-0 rounded-full bg-sky-500" /><span className={`min-w-0 ${fold ? 'truncate' : 'whitespace-pre-wrap'}`}>{t}</span></p>
          })}
        </div>
      )}
      {meta.length > 0 && rest.length > 3 && (
        <button type="button" onClick={() => setFull((v) => !v)} className="text-xs font-bold text-sky-600 hover:underline">{full ? 'طوي التقرير ▴' : `عرض التقرير كامل (${rest.length} سطر) ▾`}</button>
      )}
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
  layers: 'M12 3l9 5-9 5-9-5 9-5Zm-9 9l9 5 9-5M3 16l9 5 9-5',
  clip: 'M9 4h6v3H9zM7 5H5v16h14V5h-2M9 12h6m-6 4h4',
  chat: 'M4 5h16v11H9l-5 4V5Z',
  info: 'M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18Zm0-10v6m0-9v.5',
  xc: 'M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18ZM9 9l6 6m0-6-6 6',
  file: 'M6 3h8l4 4v14H6zM14 3v4h4M9 12h6m-6 4h6',
  fileCheck: 'M6 3h8l4 4v14H6zM14 3v4h4M9 14l2 2 4-4',
  clipList: 'M9 3h6v3H9zM7 4.5H5V21h14V4.5h-2M8 11h8M8 15h8M8 19h5',
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
    <div className="space-y-1" title={sal > 0 ? `الهدف: ${num(Math.round(sal * TARGET))} (راتبه + ٣٠٪)` : 'ماكو راتب بالملف'}>
      <div className="text-[13px] font-bold text-slate-800" dir="ltr">{num(Math.round(s.brought))}</div>
      {pct === null
        ? <span className="text-xs text-slate-400">ماكو راتب بالملف</span>
        : <span className={`inline-block whitespace-nowrap rounded-full px-2 py-0.5 text-[11px] font-bold ring-1 ${ok ? 'bg-emerald-50 text-emerald-700 ring-emerald-300' : 'bg-red-50 text-red-700 ring-red-200'}`}>
            {ok ? '✓' : '✗'} {pct}٪ من راتبه
          </span>}
    </div>
  )
}

// (ك) 11-22 زينا: الملاحظة الطويلة تطوّل الصف كله — نطلع كم كلمة بس، و«▾» يفتح مربع منسدل بالكلام كامل
const NOTE_WORDS = 4
function NoteCell({ text }: { text: string }) {
  // المربع «fixed» فوك الصفحة — حتى إطار الجدول (overflow) ما يگصّه
  const [pos, setPos] = useState<{ top: number; left: number } | null>(null)
  const open = pos !== null
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!open) return
    const close = (e: Event) => { if (!ref.current?.contains(e.target as Node)) setPos(null) }
    document.addEventListener('mousedown', close)
    window.addEventListener('scroll', close, true)
    window.addEventListener('resize', close)
    return () => { document.removeEventListener('mousedown', close); window.removeEventListener('scroll', close, true); window.removeEventListener('resize', close) }
  }, [open])
  const toggle = (e: React.MouseEvent<HTMLButtonElement>) => {
    if (open) { setPos(null); return }
    const r = e.currentTarget.getBoundingClientRect()
    const W = 288 // w-72
    setPos({ top: r.bottom + 6, left: Math.max(8, Math.min(window.innerWidth - W - 8, r.left + r.width / 2 - W / 2)) })
  }
  const words = text.trim().split(/\s+/)
  if (words.length <= NOTE_WORDS) return <div className="mt-1 text-[11px] text-slate-500">{text}</div>
  return (
    <div ref={ref} className="relative mt-1 text-[11px] text-slate-500">
      <button type="button" onClick={toggle} aria-expanded={open} title="شوف الكلام كامل"
        className="inline-flex max-w-full items-center gap-1 rounded-md px-1 hover:bg-slate-100">
        <span className="truncate">{words.slice(0, NOTE_WORDS).join(' ')}…</span>
        <span className={`shrink-0 text-xs font-bold text-amber-500 transition-transform ${open ? 'rotate-180' : ''}`}>▾</span>
      </button>
      {pos && (
        <div role="dialog" style={{ top: pos.top, left: pos.left }} className="fixed z-50 max-h-72 w-72 overflow-y-auto whitespace-pre-wrap rounded-xl border border-slate-200 bg-white p-3 text-start text-xs leading-relaxed text-slate-700 shadow-xl">
          {text}
        </div>
      )}
    </div>
  )
}

function SurveysTab() {
  const [reload, setReload] = useState(0)
  const { data, error } = useApi(() => apiRequest<Surveys>('/profile-surveys'), [reload])
  const review = async (id: string) => { await apiRequest(`/profile-surveys/${id}/review`, { method: 'POST' }).catch(swallow); setReload((n) => n + 1) }
  const heads: [string, string][] = [['', '#'], ['user', 'الموظف'], ['cap', 'التخصص / الشهادة'], ['bag', 'الوظيفة'], ['coins', 'الراتب (د.ع)'], ['chart', 'دخّل هالشهر'], ['clock', 'ساعات العقد (ساعة)'], ['cal', 'تاريخ العقد'], ['star', 'رأيه بالشركة'], ['layers', 'الحالة']]
  const dash = <span className="text-slate-400">—</span>
  return (
    <div className={`${card} space-y-4`}>
      {error && <p className="text-sm text-red-600">{error}</p>}
      {data && data.pending.length > 0 && (
        <div className="flex items-center gap-3 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-2.5 text-sm text-slate-700">
          <span className="text-amber-500"><Ico n="users" className="h-5 w-5" /></span>
          <p><b className="text-[#0f2040]">بعدهم ما عبّوا ({data.pending.length}):</b> {data.pending.map((p) => p.employeeName).join('، ')}</p>
        </div>
      )}
      <div className="overflow-x-auto rounded-2xl border border-slate-200" data-testid="surveys-table">
        <table className="w-full min-w-[980px] table-fixed text-[13px]">
          {/* (ك) 11-21 زينا: عرض ثابت لكل عمود — الأسماء والوظائف الطويلة تنكسر جوه عمودها، والجدول ما يطلع برا */}
          <colgroup>{['4%', '14%', '12%', '11%', '10%', '11%', '7%', '9%', '9%', '13%'].map((w, i) => <col key={i} style={{ width: w }} />)}</colgroup>
          <thead className="text-xs font-bold">
            <tr>
              {heads.map(([ic, h]) => (
                <th key={h} className="p-1 text-center">
                  {/* (ك) 11-22 زينا: كل عنوان بمربع لحاله (أزرق بإطار، الأيقونة فوك) — مثل التصميم */}
                  <span className="flex h-[4.25rem] flex-col items-center justify-center gap-1 th-box rounded-xl px-1.5 py-2">
                    {ic && <span className="th-box-ic"><Ico n={ic} className="h-5 w-5" /></span>}
                    {h.includes(' (') && h !== 'الراتب (د.ع)'
                      ? <span className="whitespace-nowrap text-[11px] leading-tight">{h.split(' (')[0]}<small className="block text-[10px] font-semibold opacity-80">({h.split(' (')[1]}</small></span>
                      : <span className="leading-snug">{h}</span>}
                  </span>
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-200">
            {data?.surveys.map((s, i) => {
              const flag = !!s.mismatch && !s.reviewedAt
              return (
                <tr key={s.employeeId} className={`align-middle ${flag ? 'bg-red-50' : 'hover:bg-slate-50'}`}>
                  <td className="px-2 py-2 text-center"><span className="inline-flex h-6 w-6 items-center justify-center rounded-full bg-slate-100 text-[11px] ring-1 ring-slate-200 font-bold text-slate-600">{i + 1}</span></td>
                  <td className="px-2 py-2"><span className="flex items-center gap-1.5 text-[13px] font-bold leading-snug text-[#0f2040]"><span className="shrink-0 text-amber-500"><Ico n="user" className="h-4 w-4" /></span>{s.employeeName}</span></td>
                  <td className="px-2 py-2 text-center text-slate-700">{s.major} / {s.degree}</td>
                  <td className="px-2 py-2 text-center">{s.currentJob ? <span className="inline-block max-w-full rounded-xl bg-blue-50 px-2 py-0.5 text-[11px] font-bold leading-snug text-blue-700 ring-1 ring-blue-200">{s.currentJob}</span> : dash}</td>
                  <td className="px-2 py-2 text-center text-[13px] font-bold leading-snug text-slate-800" dir="ltr" title="هو / الملف">
                    <div>{s.selfSalaryIQD !== null ? num(s.selfSalaryIQD) : '—'} /</div>
                    <div>{s.fileSalary !== null ? num(s.fileSalary) : '—'}</div>
                  </td>
                  {/* (ك) 11-21 زينا: شكد دخّل فلوس هالشهر من حجوزاته (ليدر 60٪، فني 40٪) — والهدف ٣٠٪ فوك راتبه */}
                  <td className="px-2 py-2 text-center">{broughtCell(s)}</td>
                  <td className="px-2 py-2 text-center font-bold text-slate-700">{s.monthlyHours ?? dash}</td>
                  <td className="px-2 py-2 text-center" title="هو / الملف">
                    {s.contractEnd || s.fileContract
                      ? <span className="inline-block whitespace-nowrap rounded-lg bg-slate-100 px-2 py-1 text-xs font-bold ring-1 ring-slate-200 text-slate-700" dir="ltr">{s.contractEnd ?? '—'}{s.fileContract && s.fileContract !== s.contractEnd ? ` / ${s.fileContract}` : ''}</span>
                      : <span className="text-slate-400">—/—</span>}
                  </td>
                  <td className="px-2 py-2 text-center">
                    <div className="text-sm leading-none text-amber-400" aria-label={`${s.companyRating} من ٥`}>{'★'.repeat(s.companyRating)}</div>
                    {s.companyNote && <NoteCell text={s.companyNote} />}
                  </td>
                  <td className="px-2 py-2 text-center">
                    {flag && (
                      <>
                        <div className="mb-1 text-[11px] font-bold leading-snug text-amber-600">⚠️ {s.mismatch}</div>
                        <button type="button" onClick={() => void review(s.employeeId)} className="rounded-lg border border-slate-300 bg-white px-2.5 py-0.5 text-xs font-bold text-slate-700 hover:bg-slate-50">✓ طابقت</button>
                      </>
                    )}
                    {!flag && <span className="inline-block rounded-full bg-emerald-50 px-3 py-0.5 text-xs font-bold text-emerald-700 ring-1 ring-emerald-300">✓ {s.reviewedAt ? 'انطابقت' : 'طابق'}</span>}
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
// (ك) 11-22 زينا: «سجل الإهمال» بجدول مثل الاستمارات — عناوين بمربعات، والتقييم بكبسولة، والحالة بكبسولة ملوّنة.
// detail ينكتب «فاته N تقييم: الأسماء» (rating_duty.go) — نفصله: الجزء الأول بعمود التقييم، والأسماء بالملاحظات.
function splitDetail(d: string): [string, string] {
  const i = d.indexOf(': ')
  return i > 0 ? [d.slice(0, i), d.slice(i + 2)] : ['', d]
}

function NegligenceTab() {
  const { data, error } = useApi(() => apiRequest<Neg[]>('/negligence'), [])
  const heads: [string, string][] = [['', '#'], ['cal', 'التاريخ'], ['user', 'اسم الموظف'], ['clip', 'التقييم'], ['chat', 'الملاحظات / السبب'], ['layers', 'الحالة']]
  return (
    <div className={`${card} space-y-4`}>
      <div className="flex items-center gap-3 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-2.5 text-sm text-slate-700">
        <span className="shrink-0 text-blue-500"><Ico n="info" className="h-5 w-5" /></span>
        <p>كل مقيّم فاته يقيّم الي عليه بيومه ينكتب هنا. ينزل تقييمه هو بماتركس (عوّض = نص، ما عوّض = صفر) — بلا خصم فلوس ولا نقاط انضباط.</p>
      </div>
      {error && <p className="text-sm text-red-600">{error}</p>}
      <div className="overflow-x-auto rounded-2xl border border-slate-200" data-testid="negligence-table">
        <table className="w-full min-w-[980px] table-fixed text-[13px]">
          <colgroup>{['5%', '12%', '18%', '13%', '38%', '14%'].map((w, i) => <col key={i} style={{ width: w }} />)}</colgroup>
          <thead className="text-xs font-bold">
            <tr>
              {heads.map(([ic, h]) => (
                <th key={h} className="p-0 text-center">
                  {/* (ك) 11-22 زينا: شريط عناوين واحد متصل (مثل التصميم) وبطول عناوين الاستمارات */}
                  <span className="th-box flex h-[4.75rem] items-center justify-center gap-2 rounded-none border-x-0 px-1.5 text-[13px]">
                    {ic && <span className="th-box-ic"><Ico n={ic} className="h-5 w-5" /></span>}
                    <span className="leading-snug">{h}</span>
                  </span>
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-200">
            {data?.map((n, i) => {
              const [rating, names] = splitDetail(n.detail)
              return (
                <tr key={n.id} className="h-[3.8rem] align-middle hover:bg-slate-50 [&>td]:border-l [&>td]:border-slate-200 [&>td:last-child]:border-l-0">
                  <td className="px-2 py-2 text-center"><span className="inline-flex h-6 w-6 items-center justify-center rounded-full bg-slate-100 text-[11px] font-bold text-slate-600 ring-1 ring-slate-200">{i + 1}</span></td>
                  <td className="px-2 py-2 text-center"><span className="inline-flex items-center gap-1.5 text-slate-700" dir="ltr"><span className="text-blue-500"><Ico n="cal" className="h-4 w-4" /></span>{n.day}</span></td>
                  <td className="px-2 py-2"><span className="flex items-center gap-1.5 font-bold text-[#0f2040]"><span className="shrink-0 text-amber-500"><Ico n="user" className="h-4 w-4" /></span>{n.employeeName}</span></td>
                  <td className="px-2 py-2 text-center">
                    {rating
                      ? <span className="inline-flex items-center gap-1 whitespace-nowrap rounded-full bg-blue-50 px-2.5 py-0.5 text-[11px] font-bold text-blue-700 ring-1 ring-blue-200"><Ico n="clip" className="h-3.5 w-3.5" />{rating}</span>
                      : <span className="text-slate-400">—</span>}
                  </td>
                  <td className="px-2 py-2 text-slate-700">
                    {names && <div className="leading-snug">{names}</div>}
                    {n.reason && <div className="text-xs leading-snug text-slate-500">السبب: {n.reason}</div>}
                    {!names && !n.reason && <span className="text-slate-400">—</span>}
                  </td>
                  <td className="px-2 py-2 text-center">
                    {n.compensatedAt
                      ? <span className="inline-flex items-center gap-1 whitespace-nowrap rounded-full bg-amber-50 px-2.5 py-0.5 text-[11px] font-bold text-amber-700 ring-1 ring-amber-300"><Ico n="clock" className="h-3.5 w-3.5" />عوّض متأخر</span>
                      : <span className="inline-flex items-center gap-1 whitespace-nowrap rounded-full bg-red-50 px-2.5 py-0.5 text-[11px] font-bold text-red-700 ring-1 ring-red-300"><Ico n="xc" className="h-3.5 w-3.5" />ما عوّض</span>}
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
      {data && data.length === 0 && <p className="py-4 text-center text-emerald-700">✅ ماكو إهمال مسجّل</p>}
    </div>
  )
}

type Report = Daily['reports'][number]
type MatchDetail = { status: string; note: string; claimedCodes: string[]; missingCodes: string[]; unclaimedCodes: string[]; activity: { at: string; text: string; count: number }[] }

const MATCH_BADGE: Record<string, [string, string]> = {
  MATCH: ['✓ يطابق الشغل المسجّل', 'bg-emerald-50 text-emerald-700 ring-emerald-400'],
  PARTIAL: ['⚠️ يطابق جزئياً', 'bg-amber-50 text-amber-700 ring-amber-400'],
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
    <div className="text-xs">
      <button onClick={() => setOpen((o) => !o)} className={`rounded-full px-4 py-1.5 text-sm font-bold ring-1 ${cls}`}>{label} {open ? '▴' : '▾'}</button>
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
