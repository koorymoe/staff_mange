import { useEffect, useState, type ReactNode } from 'react'
import { api, type Employee } from '../api'

// ═══ «تقرير تنفيذ المشروع» — نموذج المشرف الأصلي (ع 10-16) ═══
// «هاي الطريقه الصحيحه مال التقارير… الليدر يطلعله نفس هذا النموذج،
// والمراقب يطلعله نفس الترتيب، نفس المدخلات». نفس الأقسام والترقيم
// والتلميحات الي بملف «المشرف 1.xlsx» — لتقرير الحجز ولتقرير المشروع اليومي.
// الليدر يكتب بـExecutionSheetForm، والمراقب يقرا بـExecutionSheetView —
// نفس التعريف (SECTIONS) للاثنين حتى ما يفترق الترتيب.

import { type SheetCrewRow, type ExecutionSheetData, CHECKS, PROJECT_STATUS, PROGRAMMING, RATINGS, HOURS, MINUTES, DAYS, DATA_ROWS, today, inp } from '../utils/executionSheet'

const Hint = ({ children }: { children: ReactNode }) => <span className="block text-[11px] font-bold text-white/90">{children}</span>
const Title = ({ children }: { children: ReactNode }) => (
  <div className="rounded-t-lg bg-slate-200 px-3 py-2 text-center text-sm font-black text-[#0f2040]">{children}</div>
)

export function ExecutionSheetForm({ value, onChange }: { value: ExecutionSheetData; onChange: (v: ExecutionSheetData) => void }) {
  const [employees, setEmployees] = useState<Employee[]>([])
  useEffect(() => { api.getEmployees().then((e) => setEmployees(e.filter((x) => x.status === 'ACTIVE'))).catch(() => {}) }, [])
  const s = value
  const set = <K extends keyof ExecutionSheetData>(k: K, patch: Partial<ExecutionSheetData[K]>) =>
    onChange({ ...s, [k]: { ...(s[k] as object), ...patch } } as ExecutionSheetData)
  const setCrew = (i: number, patch: Partial<SheetCrewRow>) => onChange({ ...s, crew: s.crew.map((r, j) => (j === i ? { ...r, ...patch } : r)) })

  const dataInput = (row: typeof DATA_ROWS[number]) => {
    const v = s.data[row.key]
    const ch = (x: string) => set('data', { [row.key]: x } as Partial<ExecutionSheetData['data']>)
    if (row.kind === 'days') return <select className={inp} value={v} onChange={(e) => ch(e.target.value)}>{DAYS.map((d) => <option key={d}>{d}</option>)}</select>
    if (row.kind === 'status') return <select className={inp} value={v} onChange={(e) => ch(e.target.value)}><option value="">— اختار —</option>{PROJECT_STATUS.map((d) => <option key={d}>{d}</option>)}</select>
    if (row.kind === 'prog') return <select className={inp} value={v} onChange={(e) => ch(e.target.value)}><option value="">— اختار —</option>{PROGRAMMING.map((d) => <option key={d}>{d}</option>)}</select>
    if (row.kind === 'date') return <input type="date" className={inp} value={v} onChange={(e) => ch(e.target.value)} />
    if (row.kind === 'num') return <input type="number" min={0} className={inp} value={v} onChange={(e) => ch(e.target.value)} />
    return <input className={inp} value={v} onChange={(e) => ch(e.target.value)} />
  }

  return (
    <div className="space-y-3 text-right" dir="rtl">
      <SheetHeader s={s} />
      <div className="overflow-hidden rounded-lg border border-slate-300">
        <Title>وقت التنفيذ</Title>
        {([['start', 'وقت البدء بالعمل'], ['end', 'وقت انتهاء العمل الكلي']] as const).map(([k, label], i) => (
          <div key={k} className="grid grid-cols-[2rem_1fr] border-t border-slate-200 sm:grid-cols-[2rem_1fr_6rem_6rem_1fr]">
            <span className="grid place-items-center bg-slate-50 text-sm font-bold">{i + 1}</span>
            <span className="px-2 py-2 text-sm font-bold">{label}</span>
            <select className={`${inp} m-1`} value={s.time[`${k}H`]} onChange={(e) => set('time', { [`${k}H`]: e.target.value })}><option value="">الساعة "24"</option>{HOURS.map((h) => <option key={h}>{h}</option>)}</select>
            <select className={`${inp} m-1`} value={s.time[`${k}M`]} onChange={(e) => set('time', { [`${k}M`]: e.target.value })}><option value="">الدقيقة "60"</option>{MINUTES.map((m) => <option key={m}>{m}</option>)}</select>
            <input className={`${inp} m-1`} placeholder="اذكر الوقت بالتحديد" value={s.time[`${k}Note`]} onChange={(e) => set('time', { [`${k}Note`]: e.target.value })} />
          </div>
        ))}
      </div>

      <div className="overflow-hidden rounded-lg border border-slate-300">
        <Title>بيانات التنفيذ</Title>
        {DATA_ROWS.map((row, i) => (
          <div key={row.key} className="grid grid-cols-[2rem_1fr] border-t border-slate-200 sm:grid-cols-[2rem_12rem_1fr_14rem]">
            <span className="grid place-items-center bg-slate-50 text-sm font-bold">{i + 1}</span>
            <span className={`px-2 py-2 text-sm font-bold ${row.key === 'systemType' ? 'bg-yellow-200' : ''}`}>{row.label}</span>
            <div className="col-span-2 p-1 sm:col-span-1">{dataInput(row)}</div>
            {row.hint ? <div className="col-span-2 bg-cyan-500 px-2 py-1.5 sm:col-span-1"><Hint>{row.hint}</Hint></div> : <div className="hidden sm:block" />}
          </div>
        ))}
      </div>

      <div className="overflow-hidden rounded-lg border border-slate-300">
        <Title>تقارير الفنيين و الكادر</Title>
        {s.crew.map((r, i) => (
          <div key={i} className="grid grid-cols-2 gap-1 border-t border-slate-200 p-1 sm:grid-cols-[4rem_1fr_9rem_9rem_5rem_8rem_2rem]">
            <span className="self-center text-sm font-bold">الفني {i + 1}</span>
            <select className={inp} value={r.employeeId} onChange={(e) => setCrew(i, { employeeId: e.target.value, name: employees.find((x) => x.id === e.target.value)?.name || '' })}>
              <option value="">— اختار —</option>
              {r.employeeId && !employees.some((x) => x.id === r.employeeId) && <option value={r.employeeId}>{r.name}</option>}
              {employees.map((x) => <option key={x.id} value={x.id}>{x.name}</option>)}
            </select>
            <label className="text-[11px] text-slate-500">من<input type="date" className={inp} value={r.from} onChange={(e) => setCrew(i, { from: e.target.value })} /></label>
            <label className="text-[11px] text-slate-500">الى<input type="date" className={inp} value={r.to} onChange={(e) => setCrew(i, { to: e.target.value })} /></label>
            <label className="text-[11px] text-slate-500">ايام<select className={inp} value={r.days} onChange={(e) => setCrew(i, { days: e.target.value })}>{DAYS.map((d) => <option key={d}>{d}</option>)}</select></label>
            <label className="text-[11px] text-slate-500">التقييم<select className={inp} value={r.rating} onChange={(e) => setCrew(i, { rating: e.target.value })}><option value="">—</option>{RATINGS.map((d) => <option key={d}>{d}</option>)}</select></label>
            <button type="button" title="شيل" onClick={() => onChange({ ...s, crew: s.crew.filter((_, j) => j !== i) })} className="text-red-500">✕</button>
          </div>
        ))}
        {s.crew.length < 7 && (
          <button type="button" onClick={() => onChange({ ...s, crew: [...s.crew, { employeeId: '', name: '', from: today(), to: today(), days: '1', rating: '' }] })}
            className="m-2 rounded-lg border border-dashed border-slate-400 px-3 py-1 text-xs font-bold text-slate-600">➕ ضيف فني</button>
        )}
      </div>

      <div className="overflow-hidden rounded-lg border border-slate-300">
        <Title>اكمال المشروع</Title>
        {CHECKS.map((c) => (
          <label key={c.key} className="grid cursor-pointer grid-cols-[1fr_2.5rem] border-t border-slate-200 sm:grid-cols-[14rem_2.5rem_1fr]">
            <span className="px-2 py-2 text-sm font-bold">{c.label}</span>
            <span className="grid place-items-center"><input type="checkbox" className="h-5 w-5" checked={s.checks[c.key]} onChange={(e) => onChange({ ...s, checks: { ...s.checks, [c.key]: e.target.checked } })} /></span>
            <span className="col-span-2 bg-cyan-500 px-2 py-1.5 sm:col-span-1"><Hint>{c.hint}</Hint></span>
          </label>
        ))}
      </div>

      <div className="grid gap-2 sm:grid-cols-2">
        <label className="block rounded-lg border border-slate-300"><span className="block bg-cyan-500 px-2 py-1.5 text-center text-sm font-black text-white">اقتراح تنصيب من الزبون ( حسب رغبته )</span>
          <textarea rows={3} className="w-full p-2 text-sm outline-none" value={s.suggestion} onChange={(e) => onChange({ ...s, suggestion: e.target.value })} /></label>
        <label className="block rounded-lg border border-slate-300"><span className="block bg-cyan-500 px-2 py-1.5 text-center text-sm font-black text-white">المبلغ المستلم من الزبون</span>
          <input type="number" min={0} className="w-full p-2 text-sm outline-none" value={s.amountReceived} onChange={(e) => onChange({ ...s, amountReceived: e.target.value })} placeholder="معلومة للمراقب — الحسابات تبقى بالفاتورة" /></label>
      </div>
    </div>
  )
}

function SheetHeader({ s }: { s: ExecutionSheetData }) {
  const h = s.header
  return (
    <div className="overflow-hidden rounded-lg border border-slate-300 text-sm">
      <div className="flex flex-wrap items-center justify-between gap-2 bg-slate-100 px-3 py-2">
        <b className="text-[#0f2040]">تقرير تنفيذ المشروع</b>
        <span>الموظف المسؤول: <b>{h.responsible || '—'}</b></span>
      </div>
      <div className="grid grid-cols-2 gap-px bg-slate-200 sm:grid-cols-6">
        {([['التاريخ', h.date], ['اسم الزبون', h.customer], ['رقم الهاتف', h.phone], ['موقع المشروع', h.location], ['نوع العمل', h.workType], ['كود الزبون المخصص', h.customerCode]] as const).map(([k, v]) => (
          <div key={k} className="bg-white p-2"><span className="block text-[11px] text-slate-500">{k}</span><b dir={k === 'رقم الهاتف' ? 'ltr' : undefined}>{v || '—'}</b></div>
        ))}
      </div>
    </div>
  )
}

/** عرض المراقب — نفس الترتيب ونفس المدخلات، للقراءة بس */
export function ExecutionSheetView({ sheet }: { sheet: ExecutionSheetData }) {
  const s = sheet
  const dash = (v?: string) => (v && String(v).trim() ? v : '—')
  return (
    <div className="space-y-3 text-right" dir="rtl">
      <SheetHeader s={s} />
      <div className="overflow-hidden rounded-lg border border-slate-300 text-sm">
        <Title>وقت التنفيذ</Title>
        {([['وقت البدء بالعمل', s.time.startH, s.time.startM, s.time.startNote], ['وقت انتهاء العمل الكلي', s.time.endH, s.time.endM, s.time.endNote]] as const).map(([l, hh, mm, n], i) => (
          <div key={l} className="grid grid-cols-[2rem_1fr_6rem_1fr] border-t border-slate-200">
            <span className="grid place-items-center bg-slate-50 font-bold">{i + 1}</span><span className="p-2 font-bold">{l}</span>
            <span className="p-2 tabular-nums" dir="ltr">{hh && mm ? `${hh}:${mm}` : '—'}</span><span className="p-2 text-slate-600">{dash(n)}</span>
          </div>
        ))}
      </div>
      <div className="overflow-hidden rounded-lg border border-slate-300 text-sm">
        <Title>بيانات التنفيذ</Title>
        {DATA_ROWS.map((row, i) => (
          <div key={row.key} className="grid grid-cols-[2rem_12rem_1fr] border-t border-slate-200">
            <span className="grid place-items-center bg-slate-50 font-bold">{i + 1}</span>
            <span className={`p-2 font-bold ${row.key === 'systemType' ? 'bg-yellow-200' : ''}`}>{row.label}</span>
            <span className={`p-2 ${row.key === 'projectStatus' && s.data.projectStatus && s.data.projectStatus !== PROJECT_STATUS[0] ? 'font-bold text-red-700' : ''}`}>{dash(s.data[row.key])}</span>
          </div>
        ))}
      </div>
      <div className="overflow-x-auto rounded-lg border border-slate-300 text-sm">
        <Title>تقارير الفنيين و الكادر</Title>
        <table className="w-full"><thead className="bg-slate-50 text-[11px] text-slate-500"><tr><th className="p-1.5">#</th><th className="p-1.5">الموظف</th><th className="p-1.5">من</th><th className="p-1.5">الى</th><th className="p-1.5">ايام</th><th className="p-1.5">التقييم</th></tr></thead>
          <tbody>{s.crew.length === 0 ? <tr><td colSpan={6} className="p-2 text-center text-slate-400">—</td></tr> : s.crew.map((r, i) => (
            <tr key={i} className="border-t border-slate-100"><td className="p-1.5">الفني {i + 1}</td><td className="p-1.5 font-bold">{dash(r.name)}</td><td className="p-1.5 tabular-nums">{dash(r.from)}</td><td className="p-1.5 tabular-nums">{dash(r.to)}</td><td className="p-1.5">{dash(r.days)}</td><td className="p-1.5">{dash(r.rating)}</td></tr>
          ))}</tbody></table>
      </div>
      <div className="overflow-hidden rounded-lg border border-slate-300 text-sm">
        <Title>اكمال المشروع</Title>
        {CHECKS.map((c) => (
          <div key={c.key} className="grid grid-cols-[1fr_3rem] border-t border-slate-200">
            <span className="p-2 font-bold">{c.label}</span>
            <span className={`grid place-items-center font-black ${s.checks?.[c.key] ? 'text-emerald-600' : 'text-red-600'}`}>{s.checks?.[c.key] ? '✓' : '✗'}</span>
          </div>
        ))}
      </div>
      <div className="grid gap-2 text-sm sm:grid-cols-2">
        <div className="rounded-lg border border-slate-300"><span className="block bg-cyan-500 px-2 py-1 text-center font-black text-white">اقتراح تنصيب من الزبون</span><p className="p-2">{dash(s.suggestion)}</p></div>
        <div className="rounded-lg border border-slate-300"><span className="block bg-cyan-500 px-2 py-1 text-center font-black text-white">المبلغ المستلم من الزبون</span><p className="p-2 tabular-nums">{s.amountReceived ? `${Number(s.amountReceived).toLocaleString('en-US')} د.ع` : '—'}</p></div>
      </div>
    </div>
  )
}
