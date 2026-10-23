import { useEffect, useState } from 'react'
import { api, type StaffScore } from '../api'
import RoleSplit, { SummaryCard } from '../components/RoleSplit'

// ═══ 📈 مؤشرات أداء الموظف — طلب (ع) 10-08 ═══
// ماتركس يحلل كل شي عن الموظف وينطي «نسبة اعتمادية»: الحضور والانصراف،
// الجرد والعدّة، الفواتير والتقارير، توجيهات المسؤولين، الـKPI، الشكاوى،
// المشاكل ويا الزملاء، الإجازات. والمهارات وإجازة السوق والليدرية تنعرض
// معلومات بس (ما تدخل بالنسبة حتى ما ينظلم أحد).

const tone = (v: number | null) => (v == null ? 'text-slate-400' : v >= 85 ? 'text-emerald-600' : v >= 65 ? 'text-amber-600' : 'text-red-600')
const bar = (v: number) => (v >= 85 ? 'bg-emerald-500' : v >= 65 ? 'bg-amber-500' : 'bg-red-500')

function Card({ s }: { s: StaffScore }) {
  const [open, setOpen] = useState(false)
  const r = s.reliability
  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <button type="button" onClick={() => setOpen((o) => !o)} className="flex w-full items-center justify-between gap-3 text-right">
        <div className="min-w-0">
          <p className="truncate text-base font-extrabold text-slate-800">{s.name}</p>
          <div className="mt-1 flex flex-wrap gap-1 text-[11px]">
            {s.isLeader && <span className="rounded-full bg-indigo-100 px-2 py-0.5 font-bold text-indigo-700">👑 ليدر</span>}
            <span className={`rounded-full px-2 py-0.5 font-bold ${s.hasLicense ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-100 text-slate-500'}`}>{s.hasLicense ? '🚗 عنده إجازة سوق' : '🚫 بلا إجازة سوق'}</span>
            <span className="rounded-full bg-sky-50 px-2 py-0.5 font-bold text-sky-700">🧰 {s.skills ?? 0} مهارة/خدمة</span>
          </div>
        </div>
        <div className="shrink-0 text-center">
          <p className={`text-3xl font-black tabular-nums ${tone(r)}`}>{r == null ? '—' : `${Math.round(r)}%`}</p>
          <p className="text-[10px] text-slate-400">الاعتمادية</p>
        </div>
      </button>
      <div className="mt-3 space-y-1.5">
        {(open ? s.reliabilityParts : s.reliabilityParts.slice().sort((a, b) => a.pct - b.pct).slice(0, 3)).map((p) => (
          <div key={p.key}>
            <div className="flex justify-between text-xs"><span className="text-slate-600">{p.label}</span><span className="font-bold tabular-nums">{Math.round(p.pct)}%</span></div>
            <div className="h-1.5 rounded-full bg-slate-100"><div className={`h-1.5 rounded-full ${bar(p.pct)}`} style={{ width: `${Math.max(2, p.pct)}%` }} /></div>
            {open && <p className="text-[11px] text-slate-400">{p.detail}</p>}
          </div>
        ))}
      </div>
      <button type="button" onClick={() => setOpen((o) => !o)} className="mt-2 text-xs font-bold text-sky-700">{open ? '▲ أقل' : `▼ كل المؤشرات (${s.reliabilityParts.length})`}</button>
    </div>
  )
}

export default function EmployeeIndicatorsPage() {
  const [rows, setRows] = useState<StaffScore[] | null>(null)
  const [month, setMonth] = useState('')
  const [err, setErr] = useState('')
  useEffect(() => {
    api.getStaffScoreBoard(month).then((b) => setRows(b.staff.filter((s) => s.role !== 'ADMIN' && s.role !== 'OWNER')))
      .catch((e) => setErr(e instanceof Error ? e.message : 'تعذر'))
  }, [month])
  const rated = (rows ?? []).filter((s) => s.reliability != null)
  const avg = rated.length ? rated.reduce((a, s) => a + (s.reliability ?? 0), 0) / rated.length : null

  return (
    <div dir="rtl" className="space-y-4">
      <div className="rounded-2xl bg-gradient-to-l from-[#0f2040] to-indigo-800 p-5 text-white shadow">
        <h2 className="text-2xl font-extrabold">📈 مؤشرات أداء الموظف</h2>
        <p className="mt-1 text-sm text-indigo-100">ماتركس يحلل التزام كل موظف ويطلع نسبة اعتمادية. اضغط على البطاقة حتى تشوف كل مؤشر ومنين جاي.</p>
        <input type="month" value={month} onChange={(e) => setMonth(e.target.value)} className="mt-3 rounded-lg px-2 py-1 text-sm text-slate-800" />
      </div>
      {err && <p className="text-sm text-red-600">{err}</p>}
      {rows === null && !err ? <p className="text-slate-400">جاري التحميل…</p> : (
        <>
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
            <SummaryCard label="موظفين" value={rows?.length ?? 0} />
            <SummaryCard label="معدل الاعتمادية" value={avg == null ? '—' : `${Math.round(avg)}%`} tone={avg == null ? undefined : avg >= 85 ? 'good' : avg >= 65 ? 'mid' : 'bad'} />
            <SummaryCard label="اعتمادية عالية (٨٥+)" value={rated.filter((s) => (s.reliability ?? 0) >= 85).length} tone="good" />
            <SummaryCard label="تحتاج متابعة (<٦٥)" value={rated.filter((s) => (s.reliability ?? 0) < 65).length} tone="bad" />
          </div>
          <RoleSplit items={rows ?? []}>
            {(list) => (
              <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
                {list.slice().sort((a, b) => (a.reliability ?? 101) - (b.reliability ?? 101)).map((s) => <Card key={s.id} s={s} />)}
              </div>
            )}
          </RoleSplit>
        </>
      )}
    </div>
  )
}
