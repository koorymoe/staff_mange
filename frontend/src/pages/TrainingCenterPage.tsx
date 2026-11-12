import { Suspense, useEffect, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { api, type Employee, type TcInsight, type TcPlan } from '../api'
import { ROLE_LABELS, roleLabel } from '../roleLabels'
import { card } from '../ui/classes'
import { lazyRetry } from '../utils/staleReload'
import { CurriculumTree, MatchingTab, TrainingReportsTab } from './TrainingMatching'
import { askConfirm } from '../ui/dialogs'
import { swallow } from '../utils/swallow'

// ═══ مركز التدريب — قرار (ع) 10-13: «شي إعجازي بالتدريب» ═══
// ماتركس يحدد منو لازم يتدرّب من شغله الحقيقي (شكاوى، توقف، بطء، خصومات،
// مهارات، اختبار) — بالأرقام مو بالتخمين، وكل سبب مكتوب. ويا المسارات
// التخصصية، الخطة الشهرية والتفتيش، الجدد ومشرفيهم، المنهج لكل دور، بنك
// الأسئلة (يدوي + ماتركس يقترح والمدرب يعتمد) والاختبار الشهري وتقريره.
// بلا خصم نقاط — تقرير وتدريب.

const TrainingPrograms = lazyRetry(() => import('./TrainingPrograms'))

type Tab = 'need' | 'notes' | 'tracks' | 'plan' | 'new' | 'curriculum' | 'matching' | 'reports' | 'programs'
const TABS: { id: Tab; icon: string; label: string }[] = [
  { id: 'need', icon: '🎯', label: 'الواجب تدريبهم' },
  { id: 'notes', icon: '⚡', label: 'ملاحظات ماتركس (السرعة والعمل)' },
  { id: 'tracks', icon: '🧭', label: 'المسارات التخصصية' },
  { id: 'plan', icon: '📅', label: 'الخطة الشهرية والتفتيش' },
  { id: 'new', icon: '🌱', label: 'الموظفين الجدد' },
  { id: 'curriculum', icon: '📚', label: 'مفردات المنهج' },
  // (ل) 11-12 دازاي: بنك الأسئلة والاختبار ما نريده — بداله الماتشينغ وتقارير التدريب
  { id: 'matching', icon: '🎯', label: 'الماتشينغ' },
  { id: 'reports', icon: '📋', label: 'تقارير التدريب' },
  { id: 'programs', icon: '🎓', label: 'البرامج التدريبية' },
]
const ROLES = Object.keys(ROLE_LABELS).filter((r) => r !== 'ADMIN' && r !== 'OWNER')
const btn = 'rounded-lg bg-[#0f2040] px-4 py-2 text-sm font-bold text-white hover:bg-[#1e3f7a] disabled:opacity-40'
const input = 'rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-[#2c5aad]'
const thisMonth = () => new Date().toISOString().slice(0, 7)
const err = (e: unknown) => alert(e instanceof Error ? e.message : 'تعذّر')

export default function TrainingCenterPage() {
  // ٧و (11-05): «برامج التدريب» انشال من القائمة — رابطه القديم يفتح هنا على تبويبه
  const { pathname } = useLocation()
  const [tab, setTab] = useState<Tab>(pathname.replace(/\/+$/, '').endsWith('/training-programs') ? 'programs' : 'need')
  const [insights, setInsights] = useState<TcInsight[] | null>(null)
  const [emps, setEmps] = useState<Employee[]>([])
  useEffect(() => {
    api.tcInsights().then(setInsights).catch(() => setInsights([]))
    api.getEmployees().then((r) => setEmps(r.filter((e) => e.status === 'ACTIVE'))).catch(swallow)
  }, [])
  const need = (insights ?? []).filter((x) => x.score > 0)

  return (
    <div dir="rtl" className="w-full space-y-4">
      <div className="relative overflow-hidden rounded-2xl bg-gradient-to-l from-[#0f2040] to-[#1a3a6c] p-6 text-white">
        <span aria-hidden className="pointer-events-none absolute -left-4 top-1/2 -translate-y-1/2 text-[120px] opacity-10">🎓</span>
        <h1 className="text-2xl font-extrabold">🎓 مركز التدريب</h1>
        <p className="mt-1 text-sm text-blue-200/80">ماتركس يحدد منو يحتاج تدريب من شغله الحقيقي — والخطة والتفتيش والاختبارات بمكان واحد</p>
        <div className="mt-4 flex flex-wrap gap-3">
          {[
            ['🎯', 'يحتاجون تدريب', need.length],
            ['🔴', 'أولوية عالية', need.filter((x) => x.score >= 6).length],
            ['🌱', 'موظفين جدد', (insights ?? []).filter((x) => x.daysIn <= 90).length],
            ['👥', 'تحت المتابعة', (insights ?? []).length],
          ].map(([i, l, v]) => (
            <span key={l as string} className="rounded-xl bg-white/10 px-4 py-2">
              <span className="block text-[11px] text-blue-200/80">{i} {l}</span>
              <b className="text-xl">{insights ? v : '…'}</b>
            </span>
          ))}
        </div>
      </div>

      <div className="flex flex-wrap gap-2 rounded-2xl bg-white p-2 shadow-sm">
        {TABS.map((t) => (
          <button key={t.id} onClick={() => setTab(t.id)}
            className={`rounded-xl px-3.5 py-2 text-xs font-bold transition ${tab === t.id ? 'bg-[#0f2040] text-white shadow' : 'text-slate-600 hover:bg-slate-100'}`}>
            {t.icon} {t.label}
          </button>
        ))}
      </div>

      {tab === 'need' && <NeedTab rows={need} loading={!insights} />}
      {tab === 'notes' && <NotesTab rows={insights} />}
      {tab === 'tracks' && <TracksTab />}
      {tab === 'plan' && <PlanTab emps={emps} need={need} />}
      {tab === 'new' && <NewTab rows={(insights ?? []).filter((x) => x.daysIn <= 90)} emps={emps} />}
      {tab === 'curriculum' && <CurriculumTree />}
      {tab === 'matching' && <MatchingTab emps={emps} />}
      {tab === 'reports' && <TrainingReportsTab />}
      {tab === 'programs' && <Suspense fallback={<p className="text-slate-400">جاري التحميل…</p>}><TrainingPrograms embedded /></Suspense>}
    </div>
  )
}

function Priority({ score }: { score: number }) {
  const [t, c] = score >= 6 ? ['عالية', 'bg-red-100 text-red-700'] : score >= 3 ? ['متوسطة', 'bg-amber-100 text-amber-700'] : ['خفيفة', 'bg-sky-100 text-sky-700']
  return <span className={`rounded-full px-2.5 py-0.5 text-[11px] font-extrabold ${c}`}>أولوية {t} · {score}</span>
}

function NeedTab({ rows, loading }: { rows: TcInsight[]; loading: boolean }) {
  if (loading) return <p className="py-8 text-center text-slate-400">ماتركس يحسب…</p>
  if (rows.length === 0) return <p className={`${card} text-center text-slate-500`}>ماكو أحد يحتاج تدريب حسب شغل آخر ٦٠ يوم ✅</p>
  return (
    <div className="space-y-2">
      <p className="text-xs text-slate-500">🧮 النقاط من آخر ٦٠ يوم: شكوى ×٣ · توقف ×٢ · خصم كي بي اي ×٢ · أبطأ من زملائه ٤٠٪+ = ٣ · مهارات أقل من ٣ = ٢ · رسوب بالاختبار = ٣ · جديد = ١. كل سبب مكتوب تحت اسمه.</p>
      <div className="grid gap-3 lg:grid-cols-2">
        {rows.map((x) => (
          <div key={x.id} className={card}>
            <div className="flex flex-wrap items-center justify-between gap-2">
              <Link to={`/matrix/employee/${x.id}`} className="font-extrabold text-[#0f2040] hover:underline">{x.name}</Link>
              <div className="flex items-center gap-2"><span className="text-[11px] text-slate-500">{roleLabel(x.role)}</span><Priority score={x.score} /></div>
            </div>
            <ul className="mt-2 space-y-1 text-sm text-slate-700">{x.reasons.map((r) => <li key={r}>• {r}</li>)}</ul>
            {x.strengths.length > 0 && <p className="mt-2 text-xs text-emerald-700">💪 {x.strengths.join(' · ')}</p>}
          </div>
        ))}
      </div>
    </div>
  )
}

function NotesTab({ rows }: { rows: TcInsight[] | null }) {
  if (!rows) return <p className="py-8 text-center text-slate-400">جاري التحميل…</p>
  return (
    <div className={`${card} overflow-x-auto`}>
      <p className="mb-2 text-xs text-slate-500">مدة الشغلة = من بدء العمل لحد الإنجاز (آخر ٦٠ يوم)، ومعدل الفريق = نفس الخدمات الي اشتغلها.</p>
      <table className="w-full min-w-[900px] text-sm">
        <thead><tr className="border-b text-[12px] text-slate-500">
          {['الموظف', 'الدور', 'حجوزات', 'مدته', 'معدل الفريق', 'شكاوى', 'توقف', 'جزئي', 'مهارات', 'آخر اختبار', 'ملاحظة ماتركس'].map((h) => <th key={h} className="py-2 text-right font-bold">{h}</th>)}
        </tr></thead>
        <tbody>
          {rows.map((x) => {
            const ratio = x.avgMinutes && x.teamMinutes ? x.avgMinutes / x.teamMinutes : null
            return (
              <tr key={x.id} className="border-b border-slate-50">
                <td className="py-2 font-bold text-[#0f2040]">{x.name}</td>
                <td className="text-xs">{roleLabel(x.role)}</td>
                <td className="tabular-nums">{x.jobs}</td>
                <td className={`tabular-nums font-bold ${ratio && ratio >= 1.4 ? 'text-red-600' : ratio && ratio <= 0.8 ? 'text-emerald-700' : ''}`}>{x.avgMinutes ? `${Math.round(x.avgMinutes)} د` : '—'}</td>
                <td className="tabular-nums text-slate-500">{x.teamMinutes ? `${Math.round(x.teamMinutes)} د` : '—'}</td>
                <td className={x.complaints ? 'font-bold text-red-600' : 'text-slate-400'}>{x.complaints}</td>
                <td className={x.stops ? 'font-bold text-amber-600' : 'text-slate-400'}>{x.stops}</td>
                <td className="text-slate-500">{x.partials}</td>
                <td>{x.skills}</td>
                <td>{x.lastExamPct != null ? `${x.lastExamPct}%` : '—'}</td>
                <td className="text-xs">{[...x.strengths.map((s) => `💪 ${s}`), ...x.reasons.slice(0, 2).map((r) => `⚠️ ${r}`)].join(' · ') || 'مستقر'}</td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}

function TracksTab() {
  const [rows, setRows] = useState<Awaited<ReturnType<typeof api.tcTracks>> | null>(null)
  useEffect(() => { api.tcTracks().then(setRows).catch(() => setRows([])) }, [])
  if (!rows) return <p className="py-8 text-center text-slate-400">جاري التحميل…</p>
  return (
    <div className="space-y-3">
      {(['ENGINEERING', 'DECOR'] as const).map((d) => (
        <div key={d}>
          <h3 className="mb-2 font-extrabold text-[#0f2040]">{d === 'DECOR' ? '🪵 الديكور' : '⚙️ الشعبة الهندسية'}</h3>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            {rows.filter((t) => (t.division === 'DECOR') === (d === 'DECOR')).map((t) => {
              const ready = t.materials > 0 && t.skills > 0
              return (
                <div key={t.id} className={`${card} ${ready ? '' : 'border-amber-200'}`}>
                  <p className="font-extrabold text-[#0f2040]">{t.name}</p>
                  <div className="mt-2 grid grid-cols-2 gap-1 text-xs text-slate-600">
                    <span>📚 مواد: <b>{t.materials}</b></span><span>🧩 مهارات: <b>{t.skills}</b></span>
                    <span>💪 يتقنها: <b>{t.masters}</b></span><span>🎓 برامج: <b>{t.programs}</b></span>
                  </div>
                  {!ready && <p className="mt-2 text-[11px] font-bold text-amber-700">⚠️ البيئة ناقصة — {t.materials === 0 ? 'ماكو مواد تدريب' : 'ماكو مهارات معرّفة'}</p>}
                  {t.masters === 0 && t.skills > 0 && <p className="mt-1 text-[11px] font-bold text-red-600">محد يتقن هالخدمة — أولوية تدريب</p>}
                </div>
              )
            })}
          </div>
        </div>
      ))}
      <p className="text-xs text-slate-500">المواد والمهارات تنضاف من «إدارة الخدمات»، والبرنامج من تبويب «البرامج التدريبية» ويا مهاراته.</p>
    </div>
  )
}

function PlanTab({ emps, need }: { emps: Employee[]; need: TcInsight[] }) {
  const [month, setMonth] = useState(thisMonth())
  const [rows, setRows] = useState<TcPlan[]>([])
  const [f, setF] = useState({ title: '', role: '', plannedDate: '', trainerId: '', employeeIds: [] as string[] })
  const [insp, setInsp] = useState<Record<string, string>>({})
  const [reload, setReload] = useState(0)
  useEffect(() => { api.tcPlans(month).then(setRows).catch(() => setRows([])) }, [month, reload])
  const add = async () => {
    try { await api.tcAddPlan({ ...f, month }); setF({ title: '', role: '', plannedDate: '', trainerId: '', employeeIds: [] }); setReload((x) => x + 1) } catch (e) { err(e) }
  }
  const toggle = (id: string) => setF((x) => ({ ...x, employeeIds: x.employeeIds.includes(id) ? x.employeeIds.filter((y) => y !== id) : [...x.employeeIds, id] }))
  const pool = f.role ? emps.filter((e) => e.role === f.role) : emps
  return (
    <div className="space-y-3">
      <div className={card}>
        <div className="mb-3 flex flex-wrap items-center gap-2">
          <h3 className="font-extrabold text-[#0f2040]">➕ تدريب جديد بخطة الشهر</h3>
          <input type="month" value={month} onChange={(e) => setMonth(e.target.value)} className={input} />
        </div>
        <div className="grid gap-2 sm:grid-cols-4">
          <input value={f.title} onChange={(e) => setF({ ...f, title: e.target.value })} placeholder="عنوان التدريب (مثلاً: ضبط كاميرات IP)" className={`${input} sm:col-span-2`} />
          <select value={f.role} onChange={(e) => setF({ ...f, role: e.target.value, employeeIds: [] })} className={input}>
            <option value="">كل الأدوار</option>{ROLES.map((r) => <option key={r} value={r}>{roleLabel(r)}</option>)}
          </select>
          <input type="date" value={f.plannedDate} onChange={(e) => setF({ ...f, plannedDate: e.target.value })} className={input} />
          <select value={f.trainerId} onChange={(e) => setF({ ...f, trainerId: e.target.value })} className={input}>
            <option value="">المدرب…</option>{emps.map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
          </select>
        </div>
        <p className="mt-3 text-xs font-bold text-slate-600">المتدربين ({f.employeeIds.length}) — <button onClick={() => setF((x) => ({ ...x, employeeIds: need.filter((n) => !x.role || n.role === x.role).map((n) => n.id) }))} className="text-[#2c5aad] underline">🎯 ضيف الي ماتركس يرشحهم</button></p>
        <div className="mt-1 flex max-h-32 flex-wrap gap-1 overflow-y-auto">
          {pool.map((e) => (
            <button key={e.id} onClick={() => toggle(e.id)} className={`rounded-full px-2.5 py-1 text-[11px] font-bold ${f.employeeIds.includes(e.id) ? 'bg-[#2c5aad] text-white' : 'bg-slate-100 text-slate-600'}`}>{e.name}</button>
          ))}
        </div>
        <button disabled={!f.title.trim()} onClick={() => void add()} className={`${btn} mt-3`}>حفظ بالخطة</button>
      </div>
      {rows.length === 0 ? <p className={`${card} text-center text-slate-500`}>ماكو تدريب بخطة {month}.</p> : rows.map((p) => (
        <div key={p.id} className={card}>
          <div className="flex flex-wrap items-center justify-between gap-2">
            <p className="font-extrabold text-[#0f2040]">{p.title} {p.role && <span className="text-xs font-normal text-slate-500">· {roleLabel(p.role)}</span>}</p>
            <span className="text-xs text-slate-500">📅 {p.plannedDate ? new Date(p.plannedDate).toLocaleDateString('ar-IQ') : 'بلا موعد'} · 👨‍🏫 {p.trainerName || '—'}</span>
          </div>
          <p className="mt-1 text-xs text-slate-600">👥 {p.employeeNames.join('، ') || '—'}</p>
          {p.inspectionResult ? (
            <p className={`mt-2 rounded-lg px-3 py-2 text-xs font-bold ${p.inspectionResult === 'DONE' ? 'bg-emerald-50 text-emerald-700' : p.inspectionResult === 'PARTIAL' ? 'bg-amber-50 text-amber-700' : 'bg-red-50 text-red-700'}`}>
              🔍 التفتيش: {p.inspectionResult === 'DONE' ? 'انعمل كامل' : p.inspectionResult === 'PARTIAL' ? 'انعمل جزئي' : 'ما انعمل'} — {p.inspectorName}{p.inspectionNotes ? ` · «${p.inspectionNotes}»` : ''}
            </p>
          ) : (
            <div className="mt-2 flex flex-wrap items-center gap-2">
              <input value={insp[p.id] ?? ''} onChange={(e) => setInsp({ ...insp, [p.id]: e.target.value })} placeholder="ملاحظة التفتيش" className={`${input} min-w-[200px] flex-1 text-xs`} />
              {(['DONE', 'PARTIAL', 'NOT_DONE'] as const).map((r) => (
                <button key={r} onClick={() => api.tcInspectPlan(p.id, r, insp[p.id] || '').then(() => setReload((x) => x + 1)).catch(err)}
                  className={`rounded-lg px-3 py-1.5 text-xs font-bold text-white ${r === 'DONE' ? 'bg-emerald-700' : r === 'PARTIAL' ? 'bg-amber-600' : 'bg-red-600'}`}>
                  {r === 'DONE' ? '✓ انعمل' : r === 'PARTIAL' ? '◐ جزئي' : '✗ ما انعمل'}
                </button>
              ))}
              <button onClick={async () => { if ((await askConfirm('تحذف هالتدريب من الخطة؟'))) void api.tcDeletePlan(p.id).then(() => setReload((x) => x + 1)) }} className="text-xs text-slate-400">حذف</button>
            </div>
          )}
        </div>
      ))}
    </div>
  )
}

function NewTab({ rows, emps }: { rows: TcInsight[]; emps: Employee[] }) {
  const [mentors, setMentors] = useState<Record<string, { mentorId: string | null; notes: string | null }>>({})
  const [draft, setDraft] = useState<Record<string, { mentorId: string; notes: string }>>({})
  useEffect(() => { api.tcMentors().then((r) => setMentors(Object.fromEntries(r.map((m) => [m.employeeId, m])))).catch(swallow) }, [])
  if (rows.length === 0) return <p className={`${card} text-center text-slate-500`}>ماكو موظفين جدد (آخر ٩٠ يوم).</p>
  return (
    <div className="space-y-2">
      <p className="text-xs text-slate-500">كل موظف جديد إله مشرف تدريب يتابعه أول ٩٠ يوم. المنحنى الأسبوعي بلوحة المتابعة ← «منحنى الموظفين الجدد».</p>
      {rows.map((x) => {
        const d = draft[x.id] ?? { mentorId: mentors[x.id]?.mentorId ?? '', notes: mentors[x.id]?.notes ?? '' }
        return (
          <div key={x.id} className={`${card} flex flex-wrap items-center gap-3`}>
            <div className="min-w-[180px] flex-1">
              <p className="font-extrabold text-[#0f2040]">{x.name} <span className="text-xs font-normal text-slate-500">· {roleLabel(x.role)}</span></p>
              <p className="text-[11px] text-slate-500">يومه {x.daysIn} · الأسبوع {Math.min(13, Math.ceil(x.daysIn / 7) || 1)} من ١٣ · {x.jobs} حجز · {x.complaints} شكوى</p>
            </div>
            <select value={d.mentorId} onChange={(e) => setDraft({ ...draft, [x.id]: { ...d, mentorId: e.target.value } })} className={input}>
              <option value="">المشرف عليه…</option>{emps.filter((e) => e.id !== x.id).map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
            </select>
            <input value={d.notes} onChange={(e) => setDraft({ ...draft, [x.id]: { ...d, notes: e.target.value } })} placeholder="ملاحظات المتابعة" className={`${input} min-w-[200px] flex-1`} />
            <button onClick={() => api.tcSetMentor(x.id, d.mentorId, d.notes).then(() => setMentors({ ...mentors, [x.id]: d })).catch(err)} className={btn}>حفظ</button>
          </div>
        )
      })}
    </div>
  )
}

