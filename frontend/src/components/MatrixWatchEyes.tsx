import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type MatrixEyesReport } from '../api'
import MatrixNote from './MatrixNote'
import { useSession, isAdminOrOwner } from '../session'
import { swallow } from '../utils/swallow'
import PctWhy from '../ui/PctWhy'

// ═══ عيون الرقابة — المخزن، السيارات، رأي الزبائن، الآيتي، الطلبات المعلّقة ═══
// طلب (ع) 10-05: «لازم اكو رقابه بكل مكان» و«تروح للمراقب ويا تحليل ماتركس».
// كل بند: شنو الغلط بالأرقام، منو يخصه، وشنو يقترح ماتركس.

const ICON: Record<string, string> = { REVENUE: '💰', INTEGRITY: '🧮', STOCK: '📦', VEHICLES: '🚗', CUSTOMERS: '⭐', IT: '🖥️', PENDING: '⏳' }
// (دازاي 11-13) «الطريقة الي جاي تطلع بيها المعلومات صعبة ع الفهم» — كل بند بطاقة مرتبة:
// خط لون الأهمية، عنوان، «شنو صار»، «شنو لازم يصير»، وتحت منو يخصه ومتى وزر الحجز.
// والبنود مقسّمة أقسام حسب الأهمية، المهم فوك.
const SEV: Record<string, { label: string; bar: string; pill: string; head: string }> = {
  HIGH: { label: '🔴 مهم', bar: 'border-r-red-500', pill: 'bg-red-100 text-red-800', head: 'text-red-800' },
  MEDIUM: { label: '🟠 يحتاج متابعة', bar: 'border-r-amber-400', pill: 'bg-amber-100 text-amber-800', head: 'text-amber-800' },
  LOW: { label: '🔵 للعلم', bar: 'border-r-sky-400', pill: 'bg-sky-100 text-sky-800', head: 'text-sky-800' },
}
const SEV_ORDER = ['HIGH', 'MEDIUM', 'LOW'] as const

type EyeItem = MatrixEyesReport['eyes'][number]['items'][number]

function EyeCard({ it }: { it: EyeItem }) {
  const sev = SEV[it.severity] ?? SEV.LOW
  return (
    <div className={`rounded-xl border border-slate-200 border-r-4 bg-white p-3 shadow-sm dark:border-slate-700 dark:bg-slate-800 ${sev.bar}`}>
      <p className="text-sm font-extrabold leading-6 text-slate-900 dark:text-slate-100">{it.title}</p>
      <div className="mt-2 grid gap-1.5 text-xs leading-5">
        {it.detail && <p className="text-slate-700 dark:text-slate-300"><span className="font-bold text-slate-500 dark:text-slate-400">📌 شنو صار: </span>{it.detail}</p>}
        {it.advice && <p className="rounded-lg bg-violet-50 px-2 py-1 text-violet-900 dark:bg-violet-950/40 dark:text-violet-200"><span className="font-bold">✅ شنو لازم يصير: </span>{it.advice}</p>}
      </div>
      <div className="mt-2 flex flex-wrap items-center gap-2 border-t border-slate-100 pt-2 text-[11px] text-slate-500 dark:border-slate-700">
        {it.ownerName && <span className="rounded-full bg-slate-100 px-2 py-0.5">👤 <b className="text-slate-700">{it.ownerName}</b></span>}
        {it.at && <span className="rounded-full bg-slate-100 px-2 py-0.5">📅 {new Date(it.at).toLocaleDateString('en-GB')}</span>}
        {it.bookingId && <Link to={`/bookings?focus=${it.bookingId}`} className="mr-auto rounded-lg bg-[#0f2040] px-2.5 py-1 font-bold text-white hover:bg-[#1a3a6c]">افتح الحجز ←</Link>}
      </div>
    </div>
  )
}

export default function MatrixWatchEyes({ initialTab = 'REVENUE' }: { initialTab?: string } = {}) {
  const [r, setR] = useState<MatrixEyesReport | null>(null)
  const [tab, setTab] = useState(initialTab)
  const [all, setAll] = useState(false)
  const [busy, setBusy] = useState(false)
  const [ran, setRan] = useState('')
  const { employee } = useSession()
  const isBoss = isAdminOrOwner(employee)
  useEffect(() => { void api.getMatrixEyes().then(setR).catch(swallow) }, [])
  // «احسب كلشي وخلي يتاكد من كل شي» — نقاط ماتركس + كل العيون من جديد.
  const checkAll = async () => {
    setBusy(true); setRan('')
    try {
      const sc = await api.runStaffScore().catch(() => ({ added: 0 }))
      const rep = await api.getMatrixEyes()
      setR(rep)
      const high = rep.eyes.reduce((a, e) => a + e.high, 0)
      setRan(`✅ ماتركس حسب ${sc.added} نقطة تقييم جديدة وفحص كل العيون: ${high ? `${high} شغلة مهمة` : 'ماكو شي مهم'}.`)
    } finally { setBusy(false) }
  }
  if (!r) return <p className="text-xs text-slate-400">ماتركس يفحص…</p>
  const eye = r.eyes.find((e) => e.key === tab) ?? r.eyes[0]

  return (
    <div dir="rtl" className="space-y-3 rounded-2xl border border-slate-200 bg-slate-50/60 p-3 text-slate-800 sm:p-4">
      <div>
        <div className="flex flex-wrap items-center justify-between gap-2">
          <p className="text-base font-extrabold text-[#0f2040]">👁️ عيون الرقابة</p>
          {isBoss && <button type="button" disabled={busy} onClick={() => void checkAll()} className="rounded-lg bg-violet-700 px-3 py-1.5 text-xs font-bold text-white disabled:opacity-50">{busy ? 'ماتركس يحسب…' : '🧮 احسب وافحص كلشي هسه'}</button>}
        </div>
        {ran && <p className="mt-1 text-xs font-bold text-emerald-700">{ran}</p>}
        <p className="text-[11px] text-slate-500">الفلوس والإيرادات، تطابق الأرقام، المخزن، السيارات، رأي الزبائن، الآيتي، والطلبات المعلّقة. ماتركس يفحصها كل يوم بالقواعد (بلا كلفة).</p>
      </div>
      <MatrixNote>{r.insights.join(' ')}</MatrixNote>
      <div className="grid grid-cols-2 gap-2 sm:grid-cols-4 lg:grid-cols-7">
        {r.eyes.map((e) => (
          <button key={e.key} type="button" onClick={() => { setTab(e.key); setAll(false) }}
            className={`rounded-xl border p-2.5 text-right transition ${tab === e.key ? 'border-[#0f2040] bg-white shadow' : 'border-slate-200 bg-white/70 hover:bg-white'}`}>
            <p className="text-sm font-bold">{ICON[e.key]} {e.title}</p>
            <p className="text-[11px] text-slate-500">{e.items.length ? `${e.items.length} بند${e.high ? ` · ${e.high} مهم` : ''}` : '✅ ماكو شي'}</p>
          </button>
        ))}
      </div>

      <p className="rounded-lg bg-white px-3 py-2 text-xs text-slate-600 ring-1 ring-slate-200">{eye.summary}</p>
      {eye.items.length === 0 && <p className="rounded-xl bg-emerald-50 p-4 text-center text-sm font-bold text-emerald-700">✅ ماكو شي يحتاج انتباه هنا</p>}
      {SEV_ORDER.map((k) => {
        const group = eye.items.filter((x) => x.severity === k)
        if (!group.length) return null
        const shown = all ? group : group.slice(0, k === 'HIGH' ? 8 : 6)
        return (
          <details key={k} open={k !== 'LOW'} className="group space-y-2">
            <summary className={`flex cursor-pointer list-none items-center gap-2 text-sm font-extrabold ${SEV[k].head}`}>
              <span className={`rounded-full px-2.5 py-0.5 ${SEV[k].pill}`}>{SEV[k].label}</span>
              <span>{group.length} بند</span>
              <span className="text-[11px] font-normal text-slate-400 group-open:hidden">(اضغط حتى تفتحها)</span>
            </summary>
            <div className="grid gap-2 lg:grid-cols-2">
              {shown.map((it, i) => <EyeCard key={i} it={it} />)}
            </div>
            {shown.length < group.length && (
              <button type="button" onClick={() => setAll(true)} className="text-xs font-bold text-sky-700 hover:underline">عرض الكل ({group.length})</button>
            )}
          </details>
        )
      })}

      {tab === 'CUSTOMERS' && r.customers.length > 0 && (
        <div className="rounded-xl border border-slate-200 bg-white p-3">
          <p className="mb-2 text-xs font-extrabold text-slate-700">رضا الزبائن عن كل موظف (آخر ٦٠ يوم){r.teamHappyPct != null && <span className="font-normal text-slate-500"> — الفريق <PctWhy title="رضا الزبائن عن الفريق" value={r.teamHappyPct}
            parts={{ numLabel: 'متابعة الزبون بيها راضي', num: r.customers.reduce((a, c) => a + c.happy, 0), denLabel: 'متابعة (راضي + عنده ملاحظة)', den: r.customers.reduce((a, c) => a + c.happy + c.unhappy, 0) }}
            meaning={<>مجموع كل الموظفين سوا. نقارن بيها كل موظف: إذا واحد أقل من الفريق بـ٢٠ نقطة أو أكثر (وعنده ٥ متابعات أو أكثر) يطلع تنبيه. الشكاوى ما تدخل بهاي النسبة.</>}
            source={<>متابعات الجودة (اتصالات قسم الجودة بالزبون) على الحجوزات الي الموظف ليدرها أو بكادرها، بآخر ٦٠ يوم. «راضي» = تواصل وماكو مشكلة أو تقرير إيجابي؛ «عنده ملاحظة» = تواصل وبيه مشكلة، أو تقرير سلبي، أو الفحص طلّع الزبون صاحب حق.</>} /></span>}</p>
          <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
            {r.customers.map((c) => (
              <div key={c.employeeId} className="rounded-lg border border-slate-200 p-2 text-xs">
                <div className="flex items-center justify-between"><b>{c.name}</b><b style={{ color: c.happyPct == null ? '#94a3b8' : c.happyPct >= 80 ? '#059669' : c.happyPct >= 60 ? '#d97706' : '#dc2626' }}>{c.happyPct == null ? '—' : <PctWhy title={`رضا الزبائن عن ${c.name}`} value={c.happyPct}
                  parts={{ numLabel: 'متابعة الزبون بيها راضي', num: c.happy, denLabel: 'متابعة (راضي + عنده ملاحظة)', den: c.happy + c.unhappy }}
                  meaning={<>كل ما تعلى أحسن: ٨٠% فما فوق أخضر، ٦٠–٧٩ برتقالي، أقل أحمر. الشكاوى ({c.complaints}) تنعرض جوّه بس ما تدخل بالنسبة.{r.teamHappyPct != null ? ` الفريق كله ${r.teamHappyPct}%.` : ''}</>}
                  source={<>متابعات الجودة (اتصالات قسم الجودة بالزبون) على الحجوزات الي الموظف ليدرها أو بكادرها، بآخر ٦٠ يوم. «راضي» = تواصل وماكو مشكلة أو تقرير إيجابي؛ «عنده ملاحظة» = تواصل وبيه مشكلة، أو تقرير سلبي، أو الفحص طلّع الزبون صاحب حق.</>} />}</b></div>
                <p className="text-slate-500">{c.happy} راضي · {c.unhappy} عنده ملاحظة · {c.complaints} شكوى{c.avgRating != null ? ` · تقييم ${c.avgRating.toFixed(1)}` : ''}</p>
                {c.notes.map((n, i) => <p key={i} className="text-slate-600">{n}</p>)}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
