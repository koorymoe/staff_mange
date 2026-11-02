import { useState } from 'react'
import { Link } from 'react-router-dom'
import { apiRequest } from '../api'
import { useSession, isAdminOrOwner, hasRole } from '../session'
import { useApi } from '../utils/useApi'
import { card } from '../ui/classes'
import { num, todayBaghdad } from '../utils/format'

// ═══ 🧾 تقارير الموظفين (دازاي 11-02 — تعديلات ٩.١) ═══
// تقارير الإنجاز اليومية (منو كتب، ومنو داوم وما كتب) + استمارات معلومات
// الموظفين ويا الفروقات عن ملفهم (للكوادر والمالك بس — بيها رواتب).

type Daily = { day: string; reports: { id: string; employeeId: string; employeeName: string; body: string; late: boolean; createdAt: string }[]; missing: { employeeId: string; employeeName: string }[] }
type Survey = { employeeId: string; employeeName: string; major: string; degree: string; selfSalaryIQD: number | null; fileSalary: number | null; monthlyHours: number | null; currentJob: string; companyRating: number; companyNote: string; contractEnd: string | null; fileContract: string | null; mismatch: string | null; reviewedAt: string | null }
type Surveys = { surveys: Survey[]; pending: { employeeId: string; employeeName: string }[] }

const yesterday = () => { const d = new Date(todayBaghdad() + 'T12:00:00'); d.setDate(d.getDate() - 1); return d.toISOString().slice(0, 10) }

export default function StaffReportsPage() {
  const { employee, permissions } = useSession()
  const canSurveys = isAdminOrOwner(employee) || hasRole(employee, 'HR_COORDINATOR') || permissions.includes('staff_management')
  const [tab, setTab] = useState<'daily' | 'surveys' | 'negligence'>('daily')
  return (
    <div dir="rtl" className="space-y-4">
      <h1 className="text-2xl font-extrabold text-[#0f2040]">🧾 تقارير الموظفين</h1>
      <div className="flex gap-2">
        <button onClick={() => setTab('daily')} className={`rounded-xl px-4 py-2 text-sm font-bold ${tab === 'daily' ? 'bg-[#0f2040] text-white' : 'bg-white text-slate-600'}`}>📝 تقارير الإنجاز اليومية</button>
        <button onClick={() => setTab('negligence')} className={`rounded-xl px-4 py-2 text-sm font-bold ${tab === 'negligence' ? 'bg-[#0f2040] text-white' : 'bg-white text-slate-600'}`}>📒 سجل الإهمال</button>
        {canSurveys && <button onClick={() => setTab('surveys')} className={`rounded-xl px-4 py-2 text-sm font-bold ${tab === 'surveys' ? 'bg-[#0f2040] text-white' : 'bg-white text-slate-600'}`}>🧾 استمارات المعلومات</button>}
      </div>
      {tab === 'daily' ? <DailyTab /> : tab === 'negligence' ? <NegligenceTab /> : canSurveys && <SurveysTab />}
    </div>
  )
}

function DailyTab() {
  const [day, setDay] = useState(yesterday())
  const { data, error } = useApi(() => apiRequest<Daily>(`/daily-reports?day=${day}`), [day])
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
            <p className="text-sm font-bold text-[#0f2040]"><Link to={`/employee-dossier?id=${r.employeeId}`} className="hover:underline">{r.employeeName}</Link> {r.late && <span className="mr-1 rounded-full bg-amber-50 px-2 py-0.5 text-[11px] font-bold text-amber-700 ring-1 ring-amber-200">متأخر</span>}</p>
            <p className="whitespace-pre-wrap text-sm text-slate-700">{r.body}</p>
          </div>
        ))}
        {data && data.reports.length === 0 && <p className="py-4 text-center text-sm text-slate-400">ماكو تقارير لهذا اليوم.</p>}
      </div>
    </div>
  )
}

function SurveysTab() {
  const [reload, setReload] = useState(0)
  const { data, error } = useApi(() => apiRequest<Surveys>('/profile-surveys'), [reload])
  const review = async (id: string) => { await apiRequest(`/profile-surveys/${id}/review`, { method: 'POST' }).catch(() => {}); setReload((n) => n + 1) }
  return (
    <div className={`${card} space-y-3`}>
      {error && <p className="text-sm text-red-600">{error}</p>}
      {data && data.pending.length > 0 && <p className="rounded-xl bg-slate-50 p-3 text-sm text-slate-600"><b>بعدهم ما عبّوا ({data.pending.length}):</b> {data.pending.map((p) => p.employeeName).join('، ')}</p>}
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="text-xs text-slate-500"><tr><th className="p-2 text-right">الموظف</th><th>التخصص / الشهادة</th><th>الوظيفة</th><th>الراتب (هو / الملف)</th><th>ساعات</th><th>العقد (هو / الملف)</th><th>رأيه بالشركة</th><th></th></tr></thead>
          <tbody>
            {data?.surveys.map((s) => (
              <tr key={s.employeeId} className={`border-t border-slate-100 align-top ${s.mismatch && !s.reviewedAt ? 'bg-red-50/60' : ''}`}>
                <td className="p-2 font-bold">{s.employeeName}</td>
                <td className="p-2">{s.major} / {s.degree}</td>
                <td className="p-2">{s.currentJob}</td>
                <td className="p-2" dir="ltr">{s.selfSalaryIQD !== null ? num(s.selfSalaryIQD) : '—'} / {s.fileSalary !== null ? num(s.fileSalary) : '—'}</td>
                <td className="p-2 text-center">{s.monthlyHours ?? '—'}</td>
                <td className="p-2" dir="ltr">{s.contractEnd ?? '—'} / {s.fileContract ?? '—'}</td>
                <td className="p-2">{'⭐'.repeat(s.companyRating)}{s.companyNote && <div className="text-xs text-slate-500">{s.companyNote}</div>}</td>
                <td className="p-2">
                  {s.mismatch && <div className="mb-1 text-xs font-bold text-red-700">⚠️ {s.mismatch}</div>}
                  {s.mismatch && !s.reviewedAt && <button onClick={() => void review(s.employeeId)} className="rounded-lg border border-slate-300 px-2 py-1 text-xs font-bold">✓ طابقت</button>}
                  {s.reviewedAt && <span className="text-xs text-emerald-700">✓ انطابقت</span>}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
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
