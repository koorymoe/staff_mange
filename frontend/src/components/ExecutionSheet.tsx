import { useEffect, useState, type ReactNode } from 'react'
import { api, type Booking, type Employee } from '../api'

// ═══ «تقرير تنفيذ المشروع» — نموذج المشرف الأصلي (ع 10-16) ═══
// «هاي الطريقه الصحيحه مال التقارير… الليدر يطلعله نفس هذا النموذج،
// والمراقب يطلعله نفس الترتيب، نفس المدخلات». نفس الأقسام والترقيم
// والتلميحات الي بملف «المشرف 1.xlsx» — لتقرير الحجز ولتقرير المشروع اليومي.
// الليدر يكتب بـExecutionSheetForm، والمراقب يقرا بـExecutionSheetView —
// نفس التعريف (SECTIONS) للاثنين حتى ما يفترق الترتيب.

export interface SheetCrewRow { employeeId: string; name: string; from: string; to: string; days: string; rating: string }
export interface ExecutionSheetData {
  v: 1
  header: { date: string; customer: string; phone: string; location: string; workType: string; customerCode: string; responsible: string }
  time: { startH: string; startM: string; startNote: string; endH: string; endM: string; endNote: string }
  data: {
    workDays: string; landmark: string; startDate: string; endDate: string; systemType: string; devices: string
    wiringMeters: string; extras: string; obstacles: string; projectStatus: string; programming: string; codeGiven: string
  }
  crew: SheetCrewRow[]
  checks: Record<CheckKey, boolean>
  suggestion: string
  amountReceived: string
}

type CheckKey = 'photosDevices' | 'photosWiring' | 'photosController' | 'systemRunning' | 'explained' | 'cleaned' | 'toolsTaken'
const CHECKS: { key: CheckKey; label: string; hint: string }[] = [
  { key: 'photosDevices', label: 'تصوير الاجهزة', hint: 'لا تنسى التصوير كل الاجهزة' },
  { key: 'photosWiring', label: 'تصوير التسليك', hint: 'لا تنسى تصوير الاسلاك' },
  { key: 'photosController', label: 'تصوير الجهاز المسيطر', hint: 'صور الجهاز المسيطرة و الشاشاة ان وجدت' },
  { key: 'systemRunning', label: 'تشغيل المنظومة', hint: 'شغل المنظومة قبل الخروج' },
  { key: 'explained', label: 'شرح المنظومة للزبون', hint: 'اشرح بالتفصيل عن المنظومة' },
  { key: 'cleaned', label: 'تم تنظيف الموقع', hint: 'نظف مكانك بعد اكمال العمل لان انت نظيف' },
  { key: 'toolsTaken', label: 'اخذ جميع المعدات و عدم نسيان شيء', hint: 'معداتك و ادواتك لك لا تنساهم' },
]
export const PROJECT_STATUS = ['تم اكمال العمل 100%', 'انجاز جزئي', 'متوقف'] as const
const PROGRAMMING = ['كملت الاعدادات', 'ما كملت', 'ما تحتاج برمجة']
const RATINGS = ['ممتاز', 'جيد جداً', 'جيد', 'مقبول', 'ضعيف']
const HOURS = Array.from({ length: 24 }, (_, i) => String(i).padStart(2, '0'))
const MINUTES = Array.from({ length: 12 }, (_, i) => String(i * 5).padStart(2, '0'))
const DAYS = Array.from({ length: 30 }, (_, i) => String(i + 1))

// بنود «بيانات التنفيذ» بنفس ترتيب الشيت وتلميحاته
type DataKey = keyof ExecutionSheetData['data']
const DATA_ROWS: { key: DataKey; label: string; hint: string; kind: 'text' | 'num' | 'date' | 'days' | 'system' | 'status' | 'prog' }[] = [
  { key: 'workDays', label: 'عدد ايام العمل', hint: 'عدد الايام العمل بالمضبوط', kind: 'days' },
  { key: 'landmark', label: 'موقع المشروع اقرب نقطة دالة', hint: '', kind: 'text' },
  { key: 'startDate', label: 'تاريخ البدء', hint: 'ادخل الزمن الحقيقي', kind: 'date' },
  { key: 'endDate', label: 'تاريخ الانتهاء', hint: 'يا ساعة كملت', kind: 'date' },
  { key: 'systemType', label: 'نوع المنظومة', hint: 'شوكت خلص البطل', kind: 'system' },
  { key: 'devices', label: 'عدد الاجهزة النهائية', hint: 'كل الاجهزة الي استخدمتها', kind: 'num' },
  { key: 'wiringMeters', label: 'مسافات التسليك', hint: 'الرقم التقريبي بالمتر', kind: 'num' },
  { key: 'extras', label: 'اضافات اخرى', hint: 'اي اضافة من الزبون اذكرها', kind: 'text' },
  { key: 'obstacles', label: 'معرقلات العمل', hint: 'شنو الي اخرك و شنو تحتاج', kind: 'text' },
  { key: 'projectStatus', label: 'حالة المشروع', hint: 'تم اكمال العمل 100%', kind: 'status' },
  { key: 'programming', label: 'برمجة الاجهزة', hint: 'كملت الاعدادات ؟', kind: 'prog' },
  { key: 'codeGiven', label: 'اعطاء الكود للزبون', hint: 'اذكر الكود الخاص بالزبون و يجب ان يكون رقم الهاتف و am', kind: 'text' },
]

const today = () => new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' })

export function emptySheet(b: Booking): ExecutionSheetData {
  const crew: SheetCrewRow[] = []
  const seen = new Set<string>()
  const push = (id?: string | null, name?: string | null) => {
    if (!id || seen.has(id) || crew.length >= 7) return
    seen.add(id); crew.push({ employeeId: id, name: name || '', from: today(), to: today(), days: '1', rating: '' })
  }
  push(b.leaderId, b.leaderName)
  for (const a of b.assignments || []) push(a.employee?.id, a.employee?.name)
  const svc = (b.services && b.services.length ? b.services.map((s) => s.name).join(' + ') : b.service?.name) || ''
  return {
    v: 1,
    header: {
      date: today(), customer: b.customer?.name || '', phone: b.customer?.phone || '',
      location: b.address || '', workType: svc, customerCode: String(b.customer?.customerCode ?? ''),
      responsible: b.leaderName || b.projectSupervisor?.name || '',
    },
    time: { startH: '', startM: '', startNote: '', endH: '', endM: '', endNote: '' },
    data: {
      workDays: '1', landmark: '', startDate: today(), endDate: today(), systemType: svc, devices: '',
      wiringMeters: '', extras: '', obstacles: '', projectStatus: '', programming: '', codeGiven: '',
    },
    crew,
    checks: Object.fromEntries(CHECKS.map((c) => [c.key, false])) as Record<CheckKey, boolean>,
    suggestion: '', amountReceived: '',
  }
}

/** شنو ناقص قبل الإرسال — نفس الخانات الإجبارية بالشيت */
export function sheetMissing(s: ExecutionSheetData): string | null {
  if (!s.time.startH || !s.time.startM) return 'اذكر وقت البدء بالعمل (الساعة والدقيقة)'
  if (!s.time.endH || !s.time.endM) return 'اذكر وقت انتهاء العمل الكلي (الساعة والدقيقة)'
  if (!s.data.projectStatus) return 'اختار حالة المشروع'
  if (s.data.projectStatus !== PROJECT_STATUS[0] && s.data.obstacles.trim().length < 3) return 'اكتب معرقلات العمل — شنو الي أخرك وشنو تحتاج'
  return null
}

const inp = 'w-full rounded-lg border border-slate-300 bg-white px-2 py-1.5 text-sm outline-none focus:border-brand-500'
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

/** يقرا sheet من الخادم (JSON) بأمان */
export function asSheet(raw: unknown): ExecutionSheetData | null {
  return raw && typeof raw === 'object' && (raw as { v?: number }).v === 1 ? (raw as ExecutionSheetData) : null
}
