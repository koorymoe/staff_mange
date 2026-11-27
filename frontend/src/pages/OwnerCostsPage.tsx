import { useEffect, useMemo, useState } from 'react'
import { thisPeriod } from '../utils/format'
import { api, apiRequest } from '../api'
import { useSession } from '../session'
import { notify } from '../ui/notify'

// ═══ 💼 تكاليف الشركة وأهداف الموظفين — المالك وحده (ل) 11-24 دازاي ═══
// إدخال يدوي: الموظفين (نوع، شفت، راتب، ساعات، أيام، إجازات)، مصاريف الشهر، فترات التوقف، ونسبة ربح الشركة.
// والنظام يحسب لكل موظف يدخّل فلوس: شكد لازم يدخّل بالشهر وباليوم وبالساعة، ووين وصل.

type Kind = 'EARNER' | 'ADMIN' | 'BOOKING_ADMIN'
type Shift = 'MORNING' | 'EVENING' | null
type CostEmp = { employeeId: string; name: string; kind: Kind; shift: Shift; salary: number; baseSalary: number; hoursPerDay: number; daysPerMonth: number; leaveDays: number }
type Expense = { id: string; kind: string; amount: number; note: string | null }
type Down = { id: string; startDate: string; endDate: string; reason: string }
type Target = { employeeId: string; name: string; shift: Shift; salary: number; workDays: number; overhead: number; shiftShare: number; target: number; daily: number; hourly: number; brought: number; remaining: number; forecast: number; pct: number; baseSalary: number; expectedNow: number; expectedSalary: number; dailyNeeded: number; errors: number; fixes: number; rank: number }
type Rep = {
  month: string; marginPct: number; employees: CostEmp[]; expenses: Expense[]; downtime: Down[]; downtimeDays: number
  salariesEarn: number; salariesAdmin: number; salariesShift: Record<string, number>; expensesTotal: number; totalCost: number; totalTarget: number; targets: Target[]
}

const iqd = (v: number) => Math.round(v).toLocaleString('en-US')
const thisMonth = thisPeriod // (11-27) فترة الشركة ٢٦ ← ٢٥
const KIND: Record<Kind, string> = { EARNER: '💰 يدخّل فلوس (فني/ليدر)', ADMIN: '🏢 إداري', BOOKING_ADMIN: '📞 إداري حجوزات' }
const SHIFT: Record<string, string> = { MORNING: 'صباحي', EVENING: 'مسائي' }
const EXP: Record<string, string> = { FUEL: '⛽ كاز', VEHICLE: '🚗 تصليح سيارات', TOOLS: '🧰 عدد جديدة', OTHER: '📦 غيرها' }
const inp = 'rounded-lg border border-slate-300 p-1.5 text-sm'

export default function OwnerCostsPage() {
  const { employee } = useSession()
  const isOwner = employee?.actualRole === 'OWNER'
  const [month, setMonth] = useState(thisMonth())
  const [rep, setRep] = useState<Rep | null>(null)
  const [err, setErr] = useState('')
  const [all, setAll] = useState<{ id: string; name: string; salary?: number | null }[]>([])
  const [edit, setEdit] = useState<CostEmp | null>(null)
  const [exp, setExp] = useState({ kind: 'FUEL', amount: 0, note: '' })
  const [dt, setDt] = useState({ startDate: '', endDate: '', reason: '' })
  const [margin, setMargin] = useState(0)

  const load = (m = month) => apiRequest<Rep>(`/owner/costs?month=${m}`).then((r) => { setRep(r); setMargin(r.marginPct) }).catch((e: Error) => setErr(e.message))
  useEffect(() => {
    if (!isOwner) return
    void load(month)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOwner, month])
  useEffect(() => { if (isOwner) api.getEmployees().then((es) => setAll(es.map((e) => ({ id: e.id, name: e.name, salary: (e as { salary?: number | null }).salary })))).catch(() => {}) }, [isOwner])
  const inList = useMemo(() => new Set(rep?.employees.map((e) => e.employeeId)), [rep])
  if (!isOwner) return <p className="py-10 text-center text-slate-500">الصفحة غير موجودة</p>
  if (err) return <p role="alert" className="text-rose-700">{err}</p>
  if (!rep) return <p className="py-10 text-center text-slate-400">جاري التحميل…</p>

  const call = async (path: string, method: string, body?: unknown) => {
    try {
      setRep(await apiRequest<Rep>(`${path}${path.includes('?') ? '&' : '?'}month=${month}`, { method, body: body ? JSON.stringify(body) : undefined }))
      return true
    } catch (e) { notify(e instanceof Error ? e.message : 'تعذر'); return false }
  }
  const saveEmp = async () => { if (edit && await call(`/owner/costs/employees/${edit.employeeId}`, 'PUT', edit)) { setEdit(null); notify('انحفظ ✅', 'success') } }

  return (
    <div className="mx-auto max-w-6xl space-y-4" data-testid="owner-costs">
      <header className="flex flex-wrap items-center justify-between gap-2">
        <div>
          <h1 className="text-2xl font-extrabold text-[#0f2040]">💼 تكاليف الشركة والأهداف</h1>
          <p className="text-sm text-slate-500">إدخال يدوي. النظام يحسب كل موظف يدخّل فلوس شكد لازم يدخّل حتى تتغطى الرواتب والمصاريف وزيادة.</p>
        </div>
        <input type="month" value={month} onChange={(e) => e.target.value && setMonth(e.target.value)} className={inp} dir="ltr" />
      </header>

      <section className="grid gap-2 sm:grid-cols-5">
        {([['رواتب الي يدخّلون', rep.salariesEarn], ['رواتب الإداريين', rep.salariesAdmin], ['إداريي الحجوزات (بالشفت)', Object.values(rep.salariesShift).reduce((a, b) => a + b, 0)], ['مصاريف الشهر', rep.expensesTotal], ['التكلفة الكلية', rep.totalCost]] as const).map(([k, v], i) => (
          <div key={k} className={`rounded-xl p-3 shadow-sm ${i === 4 ? 'bg-[#0f2040] text-white' : 'bg-white'}`}><p className="text-xs opacity-70">{k}</p><p className="text-lg font-extrabold" dir="ltr">{iqd(v)}</p></div>
        ))}
      </section>

      <section className="space-y-2 rounded-2xl bg-white p-4 shadow-sm">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <h2 className="text-lg font-bold">🎯 أهداف الموظفين — {rep.month}{rep.downtimeDays > 0 && <small className="ms-2 text-xs font-normal text-amber-700">({rep.downtimeDays} يوم توقف محسوب)</small>}</h2>
          <label className="flex items-center gap-2 text-sm"><b>ربح الشركة فوگ التكاليف</b>
            <input type="number" min={0} value={margin} onChange={(e) => setMargin(Number(e.target.value))} className={`${inp} w-20 text-center`} dir="ltr" />٪
            {margin !== rep.marginPct && <button type="button" onClick={() => void call('/owner/costs/margin', 'PUT', { marginPct: margin })} className="rounded-lg bg-[#0f2040] px-3 py-1 text-xs font-bold text-white">احفظ</button>}
          </label>
        </div>
        {rep.targets.length === 0 ? <p className="text-sm text-slate-500">أضف موظفين نوعهم «يدخّل فلوس» حتى تطلع أهدافهم.</p> : (
          <div className="overflow-x-auto">
            <table className="w-full min-w-[1150px] text-sm">
              <thead className="bg-slate-50 text-xs text-slate-600"><tr>{['#', 'الموظف', 'الشفت', 'الراتب', 'حصته من الإداريين والمصاريف', 'من إداري شفته', 'الهدف بالشهر', 'باليوم', 'بالساعة', 'دخّل لحد هسه', 'باقي', 'التوقع لآخر الشهر', 'الراتب المتوقع', 'أخطاء/تصحيح'].map((h) => <th key={h} className="p-2 text-center">{h}</th>)}</tr></thead>
              <tbody className="divide-y">
                {rep.targets.map((t) => (
                  <tr key={t.employeeId} className="text-center">
                    <td className="p-2 text-slate-500">{t.rank}</td>
                    <td className="p-2 text-start font-bold">{t.name}<small className="block text-xs font-normal text-slate-500">{t.workDays} يوم شغل</small></td>
                    <td>{t.shift ? SHIFT[t.shift] : '—'}</td><td dir="ltr">{iqd(t.salary)}</td><td dir="ltr">{iqd(t.overhead)}</td><td dir="ltr">{iqd(t.shiftShare)}</td>
                    <td className="font-extrabold text-[#0f2040]" dir="ltr">{iqd(t.target)}</td><td dir="ltr">{iqd(t.daily)}</td><td dir="ltr">{iqd(t.hourly)}</td>
                    <td dir="ltr"><b className={t.pct >= 100 ? 'text-emerald-700' : t.pct >= 60 ? 'text-amber-700' : 'text-rose-700'}>{iqd(t.brought)}</b> <small className="text-slate-500">({t.pct}٪)</small></td>
                    <td dir="ltr">{iqd(t.remaining)}</td>
                    <td dir="ltr" className={t.forecast >= t.target ? 'text-emerald-700' : 'text-rose-700'}>{iqd(t.forecast)}</td>
                    <td dir="ltr" title={`لحد هسه ${iqd(t.expectedNow)} · الأساسي ${iqd(t.baseSalary)}`}><b>{iqd(t.expectedSalary)}</b></td>
                    <td>{t.errors > 0 ? <span className="text-rose-700">{t.errors}✗</span> : '—'} / {t.fixes > 0 ? <span className="text-emerald-700">{t.fixes}✓</span> : '—'}</td>
                  </tr>
                ))}
              </tbody>
              <tfoot><tr className="bg-slate-50 font-bold"><td className="p-2" colSpan={6}>المجموع</td><td className="text-center" dir="ltr">{iqd(rep.totalTarget)}</td><td colSpan={7} /></tr></tfoot>
            </table>
          </div>
        )}
        <p className="text-xs text-slate-400">الراتب المتوقع = الأساسي + (الراتب − الأساسي) × نسبة الوصول للهدف (على التوقع لآخر الشهر). «دخّل» = نفس «الفلوس الي دخّلها» بالتقييم (حسب نسب «توزيع الفلوس»). التوقع = على هالمنوال لآخر الشهر.</p>
      </section>

      <section className="space-y-2 rounded-2xl bg-white p-4 shadow-sm">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <h2 className="text-lg font-bold">👥 الموظفين بالحساب</h2>
          <select value="" onChange={(e) => { const p = all.find((x) => x.id === e.target.value); if (p) setEdit({ employeeId: p.id, name: p.name, kind: 'EARNER', shift: null, salary: Number(p.salary) || 0, baseSalary: 350000, hoursPerDay: 8, daysPerMonth: 26, leaveDays: 0 }) }} className={inp}>
            <option value="">+ أضف موظف…</option>
            {all.filter((x) => !inList.has(x.id)).map((x) => <option key={x.id} value={x.id}>{x.name}</option>)}
          </select>
        </div>
        {edit && (
          <div className="grid gap-2 rounded-xl bg-amber-50 p-3 text-sm sm:grid-cols-4" data-testid="cost-edit">
            <b className="sm:col-span-4">{edit.name}</b>
            <label>النوع<select value={edit.kind} onChange={(e) => setEdit({ ...edit, kind: e.target.value as Kind })} className={`${inp} mt-1 w-full`}>{(Object.keys(KIND) as Kind[]).map((k) => <option key={k} value={k}>{KIND[k]}</option>)}</select></label>
            <label>الشفت<select value={edit.shift ?? ''} onChange={(e) => setEdit({ ...edit, shift: (e.target.value || null) as Shift })} className={`${inp} mt-1 w-full`}><option value="">—</option><option value="MORNING">صباحي</option><option value="EVENING">مسائي</option></select></label>
            <label>الراتب<input type="number" min={0} value={edit.salary} onChange={(e) => setEdit({ ...edit, salary: Number(e.target.value) })} className={`${inp} mt-1 w-full`} dir="ltr" /></label>
            <label>الراتب الأساسي<input type="number" min={0} value={edit.baseSalary} onChange={(e) => setEdit({ ...edit, baseSalary: Number(e.target.value) })} className={`${inp} mt-1 w-full`} dir="ltr" /></label>
            <label>ساعات اليوم<input type="number" min={1} max={24} value={edit.hoursPerDay} onChange={(e) => setEdit({ ...edit, hoursPerDay: Number(e.target.value) })} className={`${inp} mt-1 w-full`} dir="ltr" /></label>
            <label>أيام الدوام بالشهر<input type="number" min={1} max={31} value={edit.daysPerMonth} onChange={(e) => setEdit({ ...edit, daysPerMonth: Number(e.target.value) })} className={`${inp} mt-1 w-full`} dir="ltr" /></label>
            <label>إجازات بالشهر<input type="number" min={0} value={edit.leaveDays} onChange={(e) => setEdit({ ...edit, leaveDays: Number(e.target.value) })} className={`${inp} mt-1 w-full`} dir="ltr" /></label>
            <p className="self-end text-xs text-slate-600">ساعات الشهر: <b>{edit.hoursPerDay * (edit.daysPerMonth - edit.leaveDays)}</b></p>
            <div className="flex items-end gap-2"><button type="button" onClick={() => void saveEmp()} className="rounded-lg bg-[#0f2040] px-4 py-1.5 font-bold text-white">احفظ</button><button type="button" onClick={() => setEdit(null)} className="text-slate-500">إلغاء</button></div>
          </div>
        )}
        <table className="w-full text-sm">
          <thead className="bg-slate-50 text-xs text-slate-600"><tr>{['الموظف', 'النوع', 'الشفت', 'الراتب', 'الأساسي', 'ساعات/يوم', 'أيام/شهر', 'إجازات', ''].map((h) => <th key={h} className="p-2 text-center">{h}</th>)}</tr></thead>
          <tbody className="divide-y">
            {rep.employees.map((e) => (
              <tr key={e.employeeId} className="text-center">
                <td className="p-2 text-start font-bold">{e.name}</td><td>{KIND[e.kind]}</td><td>{e.shift ? SHIFT[e.shift] : '—'}</td>
                <td dir="ltr">{iqd(e.salary)}</td><td dir="ltr">{iqd(e.baseSalary)}</td><td>{e.hoursPerDay}</td><td>{e.daysPerMonth}</td><td>{e.leaveDays}</td>
                <td className="whitespace-nowrap"><button type="button" onClick={() => setEdit({ ...e })} className="px-2 text-[#0f2040]">✏️</button><button type="button" onClick={() => void call(`/owner/costs/employees/${e.employeeId}`, 'DELETE')} className="px-2 text-rose-600" aria-label="شيل من الحساب">✕</button></td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>

      <div className="grid gap-4 md:grid-cols-2">
        <section className="space-y-2 rounded-2xl bg-white p-4 shadow-sm">
          <h2 className="text-lg font-bold">🧾 مصاريف {rep.month}</h2>
          {rep.expenses.map((x) => (
            <p key={x.id} className="flex items-center justify-between rounded-lg bg-slate-50 px-3 py-2 text-sm"><span>{EXP[x.kind] ?? x.kind}{x.note ? ` · ${x.note}` : ''}</span><span className="flex items-center gap-2"><b dir="ltr">{iqd(x.amount)}</b><button type="button" onClick={() => void call(`/owner/costs/expenses/${x.id}`, 'DELETE')} className="text-rose-600" aria-label="شيل">✕</button></span></p>
          ))}
          <div className="flex flex-wrap gap-2">
            <select value={exp.kind} onChange={(e) => setExp({ ...exp, kind: e.target.value })} className={inp}>{Object.entries(EXP).map(([k, v]) => <option key={k} value={k}>{v}</option>)}</select>
            <input type="number" min={0} value={exp.amount || ''} onChange={(e) => setExp({ ...exp, amount: Number(e.target.value) })} placeholder="المبلغ" className={`${inp} w-32`} dir="ltr" />
            <input value={exp.note} onChange={(e) => setExp({ ...exp, note: e.target.value })} placeholder="ملاحظة" className={`${inp} flex-1`} />
            <button type="button" disabled={!exp.amount} onClick={async () => { if (await call('/owner/costs/expenses', 'POST', { ...exp, month })) setExp({ kind: exp.kind, amount: 0, note: '' }) }} className="rounded-lg bg-[#0f2040] px-3 py-1.5 text-sm font-bold text-white disabled:opacity-40">أضف</button>
          </div>
        </section>
        <section className="space-y-2 rounded-2xl bg-white p-4 shadow-sm">
          <h2 className="text-lg font-bold">⏸️ فترات التوقف</h2>
          <p className="text-xs text-slate-500">مناسبات يتوقف بيها الشغل أو يقل — أيامها تنطرح من أيام شغل الموظف، فما يتحاسب عليها.</p>
          {rep.downtime.map((d) => (
            <p key={d.id} className="flex items-center justify-between rounded-lg bg-slate-50 px-3 py-2 text-sm"><span><b>{d.reason}</b> <small dir="ltr">{d.startDate} → {d.endDate}</small></span><button type="button" onClick={() => void call(`/owner/costs/downtime/${d.id}`, 'DELETE')} className="text-rose-600" aria-label="شيل">✕</button></p>
          ))}
          <div className="flex flex-wrap gap-2">
            <input type="date" value={dt.startDate} onChange={(e) => setDt({ ...dt, startDate: e.target.value })} className={inp} />
            <input type="date" value={dt.endDate} onChange={(e) => setDt({ ...dt, endDate: e.target.value })} className={inp} />
            <input value={dt.reason} onChange={(e) => setDt({ ...dt, reason: e.target.value })} placeholder="السبب (عيد، الأربعين…)" className={`${inp} flex-1`} />
            <button type="button" disabled={!dt.startDate || !dt.endDate || !dt.reason.trim()} onClick={async () => { if (await call('/owner/costs/downtime', 'POST', dt)) setDt({ startDate: '', endDate: '', reason: '' }) }} className="rounded-lg bg-[#0f2040] px-3 py-1.5 text-sm font-bold text-white disabled:opacity-40">أضف</button>
          </div>
        </section>
      </div>
    </div>
  )
}
