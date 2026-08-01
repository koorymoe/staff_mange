import { useEffect, useMemo, useState } from 'react'

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:4000/api'

async function request<T>(path: string): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    // مفتاح التوكن بالنظام هو authToken — استخدام 'token' كان يخلي الصفحة
    // ترجع "يجب تسجيل الدخول" رغم إن المستخدم داخل فعلاً
    headers: { Authorization: `Bearer ${localStorage.getItem('authToken') || ''}` },
  })
  if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || 'خطأ بالاتصال')
  return res.json()
}

interface ProjectRow {
  id: string; code: string; name: string; stage: string
  workType: string | null; priority: string
  priceRaw: string | null; priceValue: number | null
  createdByName: string | null; responsibleName: string | null
  surveyorName: string | null; delegatedToName: string | null
  hasSurvey: boolean; createdAt: string
}

interface EmployeeRow {
  employeeId: string; name: string; role: string
  addedCount: number
  surveyAssignedCount: number; surveyFilledCount: number
  responsibleCount: number
  delegationsReceived: number; currentlyDelegated: number
  responsibleValue: number; completedCount: number
}

interface StatsResponse {
  overview: {
    totalProjects: number; totalValue: number; pricedProjects: number
    averageValue: number; completedCount: number; rejectedCount: number
    activeCount: number; delegatedCount: number; surveysFilled: number
    completedValue: number; inProgressValue: number
    stageBreakdown: Record<string, number>
  }
  projects: ProjectRow[]
  employees: EmployeeRow[]
}

const money = (v: number) => Math.round(v).toLocaleString('en-US')

const STAGE_ORDER = ['اتصال', 'كشف', 'سعر', 'عقد', 'تنفيذ', 'مكتمل', 'مرفوض'] as const
const STAGE_ICON: Record<string, string> = {
  اتصال: '📞', كشف: '🔍', سعر: '💰', عقد: '📄', تنفيذ: '🛠️', مكتمل: '✅', مرفوض: '❌',
}

export default function ProjectStatisticsPage() {
  const [data, setData] = useState<StatsResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [tab, setTab] = useState<'projects' | 'employees'>('projects')
  const [search, setSearch] = useState('')
  // ترتيب جدول الموظفين حسب العمود الي يهم المدير هالوقت
  const [empSort, setEmpSort] = useState<keyof EmployeeRow>('responsibleCount')

  useEffect(() => {
    request<StatsResponse>('/projects/statistics')
      .then(setData)
      .catch((e) => setError(e instanceof Error ? e.message : 'تعذر جلب الإحصائيات'))
      .finally(() => setLoading(false))
  }, [])

  const projects = useMemo(() => {
    if (!data) return []
    const q = search.trim().toLowerCase()
    if (!q) return data.projects
    return data.projects.filter((p) =>
      [p.code, p.name, p.stage, p.workType, p.responsibleName, p.surveyorName, p.delegatedToName, p.createdByName]
        .some((v) => (v || '').toString().toLowerCase().includes(q)))
  }, [data, search])

  const employees = useMemo(() => {
    if (!data) return []
    return [...data.employees].sort((a, b) => Number(b[empSort]) - Number(a[empSort]))
  }, [data, empSort])

  if (loading) return <p className="p-6 text-slate-400" dir="rtl">جاري التحميل...</p>
  if (error) return <p className="m-6 rounded-lg bg-red-50 p-4 text-red-600" dir="rtl">{error}</p>
  if (!data) return null

  const ov = data.overview
  const maxStage = Math.max(1, ...STAGE_ORDER.map((s) => ov.stageBreakdown[s] || 0))

  return (
    <div dir="rtl" className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-[var(--color-brand-900)]">📊 إحصائيات المشاريع</h1>
        <p className="mt-1 text-slate-500">القيمة المالية لكل مشروع، ودور كل موظف داخل المشاريع.</p>
      </div>

      {/* أرقام عامة */}
      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        <Card label="عدد المشاريع" value={ov.totalProjects.toLocaleString()} tone="brand" />
        <Card label="القيمة الإجمالية" value={`${money(ov.totalValue)} د.ع`} tone="green"
          hint={`${ov.pricedProjects} مشروع مُسعَّر من ${ov.totalProjects}`} />
        <Card label="معدّل قيمة المشروع" value={`${money(ov.averageValue)} د.ع`} tone="amber" />
        <Card label="مشاريع مُسلَّمة لموظفين" value={ov.delegatedCount.toLocaleString()} tone="violet" />
        <Card label="قيد التنفيذ" value={ov.activeCount.toLocaleString()} tone="blue"
          hint={`بقيمة ${money(ov.inProgressValue)} د.ع`} />
        <Card label="مكتملة" value={ov.completedCount.toLocaleString()} tone="green"
          hint={`بقيمة ${money(ov.completedValue)} د.ع`} />
        <Card label="مرفوضة" value={ov.rejectedCount.toLocaleString()} tone="slate" />
        <Card label="استمارات كشف مملوءة" value={ov.surveysFilled.toLocaleString()} tone="brand" />
      </div>

      {/* توزيع المراحل */}
      <div className="rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        <h3 className="mb-3 font-bold text-[var(--color-brand-900)]">توزيع المشاريع على المراحل</h3>
        <div className="space-y-2">
          {STAGE_ORDER.map((s) => {
            const n = ov.stageBreakdown[s] || 0
            return (
              <div key={s} className="flex items-center gap-3 text-sm">
                <span className="w-24 shrink-0 text-slate-600">{STAGE_ICON[s]} {s}</span>
                <div className="h-3 flex-1 overflow-hidden rounded-full bg-slate-100">
                  <div className="h-full rounded-full bg-[var(--color-brand-500)]"
                    style={{ width: `${(n / maxStage) * 100}%` }} />
                </div>
                <span className="w-10 text-left font-bold text-slate-700">{n}</span>
              </div>
            )
          })}
        </div>
      </div>

      {/* تبويبات */}
      <div className="flex gap-2">
        <button onClick={() => setTab('projects')}
          className={`rounded-lg px-4 py-2 text-sm font-medium ${tab === 'projects' ? 'bg-[var(--color-brand-500)] text-white' : 'border bg-white text-slate-600'}`}>
          كل مشروع وقيمته ({data.projects.length})
        </button>
        <button onClick={() => setTab('employees')}
          className={`rounded-lg px-4 py-2 text-sm font-medium ${tab === 'employees' ? 'bg-[var(--color-brand-500)] text-white' : 'border bg-white text-slate-600'}`}>
          إحصائيات الموظفين ({data.employees.length})
        </button>
      </div>

      {tab === 'projects' && (
        <div className="rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="🔍 بحث بالكود، الاسم، المرحلة، أو اسم الموظف..."
            className="mb-4 w-full rounded-lg border border-slate-300 px-4 py-2 text-sm outline-none focus:border-[var(--color-brand-500)]"
          />
          <div className="overflow-x-auto">
            <table className="w-full min-w-[900px] text-right text-sm">
              <thead>
                <tr className="border-b text-xs text-slate-500">
                  <th className="p-2">الكود</th>
                  <th className="p-2">المشروع</th>
                  <th className="p-2">المرحلة</th>
                  <th className="p-2">نوع العمل</th>
                  <th className="p-2">القيمة المالية</th>
                  <th className="p-2">المسؤول</th>
                  <th className="p-2">منفّذ الكشف</th>
                  <th className="p-2">مُسلَّم إلى</th>
                  <th className="p-2">أضافه</th>
                </tr>
              </thead>
              <tbody>
                {projects.map((p) => (
                  <tr key={p.id} className="border-b border-slate-50 hover:bg-slate-50/60">
                    <td className="p-2 text-xs text-slate-400">{p.code}</td>
                    <td className="p-2 font-medium text-[var(--color-brand-900)]">{p.name}</td>
                    <td className="p-2 text-xs">{p.stage}</td>
                    <td className="p-2 text-xs text-slate-500">{p.workType || '—'}</td>
                    <td className="p-2 font-bold text-emerald-700">
                      {p.priceValue != null ? `${money(p.priceValue)} د.ع` : <span className="text-slate-300">غير مُسعَّر</span>}
                    </td>
                    <td className="p-2 text-xs">{p.responsibleName || '—'}</td>
                    <td className="p-2 text-xs">
                      {p.surveyorName || '—'}
                      {p.surveyorName && (
                        <span className={p.hasSurvey ? 'text-emerald-600' : 'text-amber-600'}>
                          {p.hasSurvey ? ' ✅' : ' ⏳'}
                        </span>
                      )}
                    </td>
                    <td className="p-2 text-xs text-violet-700">{p.delegatedToName || '—'}</td>
                    <td className="p-2 text-xs text-slate-500">{p.createdByName || '—'}</td>
                  </tr>
                ))}
                {projects.length === 0 && (
                  <tr><td colSpan={9} className="p-6 text-center text-slate-400">لا توجد نتائج</td></tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {tab === 'employees' && (
        <div className="rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <div className="mb-4 flex flex-wrap items-center gap-2 text-xs">
            <span className="text-slate-500">ترتيب حسب:</span>
            {([
              ['responsibleCount', 'مشاريع مسؤول عنها'],
              ['surveyAssignedCount', 'كشوفات'],
              ['delegationsReceived', 'مشاريع استلمها'],
              ['addedCount', 'مشاريع أضافها'],
              ['responsibleValue', 'القيمة المالية'],
            ] as [keyof EmployeeRow, string][]).map(([k, label]) => (
              <button key={k} onClick={() => setEmpSort(k)}
                className={`rounded-lg px-3 py-1.5 ${empSort === k ? 'bg-[var(--color-brand-500)] text-white' : 'border bg-white text-slate-600'}`}>
                {label}
              </button>
            ))}
          </div>
          <div className="overflow-x-auto">
            <table className="w-full min-w-[860px] text-right text-sm">
              <thead>
                <tr className="border-b text-xs text-slate-500">
                  <th className="p-2">الموظف</th>
                  <th className="p-2">أضاف / رحّل</th>
                  <th className="p-2">طلع كشف</th>
                  <th className="p-2">كشوفات مملوءة</th>
                  <th className="p-2">مسؤول عن</th>
                  <th className="p-2">منها مكتملة</th>
                  <th className="p-2">استلم مشروع</th>
                  <th className="p-2">مُسلَّم إله حالياً</th>
                  <th className="p-2">قيمة مشاريعه</th>
                </tr>
              </thead>
              <tbody>
                {employees.map((e) => (
                  <tr key={e.employeeId} className="border-b border-slate-50 hover:bg-slate-50/60">
                    <td className="p-2 font-medium text-[var(--color-brand-900)]">{e.name}</td>
                    <td className="p-2">{e.addedCount || '—'}</td>
                    <td className="p-2">{e.surveyAssignedCount || '—'}</td>
                    <td className="p-2">
                      {e.surveyAssignedCount
                        ? <span className={e.surveyFilledCount === e.surveyAssignedCount ? 'text-emerald-700' : 'text-amber-600'}>
                            {e.surveyFilledCount} / {e.surveyAssignedCount}
                          </span>
                        : '—'}
                    </td>
                    <td className="p-2 font-bold">{e.responsibleCount || '—'}</td>
                    <td className="p-2 text-emerald-700">{e.completedCount || '—'}</td>
                    <td className="p-2 text-violet-700">{e.delegationsReceived || '—'}</td>
                    <td className="p-2 text-violet-700">{e.currentlyDelegated || '—'}</td>
                    <td className="p-2 font-bold text-emerald-700">
                      {e.responsibleValue ? `${money(e.responsibleValue)} د.ع` : '—'}
                    </td>
                  </tr>
                ))}
                {employees.length === 0 && (
                  <tr><td colSpan={9} className="p-6 text-center text-slate-400">ما اكو موظف مرتبط بمشروع لحد الآن</td></tr>
                )}
              </tbody>
            </table>
          </div>
          <p className="mt-3 text-[11px] text-slate-400">
            «كشوفات مملوءة» = من الكشوفات المُسندة إله، شكد واحد انملت استمارته فعلاً.
            «قيمة مشاريعه» = مجموع أسعار المشاريع الي هو المسؤول عنها.
          </p>
        </div>
      )}
    </div>
  )
}

const TONES: Record<string, string> = {
  brand: 'bg-[var(--color-brand-500)]',
  green: 'bg-emerald-600',
  amber: 'bg-amber-500',
  violet: 'bg-violet-600',
  blue: 'bg-blue-600',
  slate: 'bg-slate-500',
}

function Card({ label, value, tone, hint }: { label: string; value: string; tone: string; hint?: string }) {
  return (
    <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
      <div className={`h-1.5 ${TONES[tone] || TONES.brand}`} />
      <div className="p-4">
        <p className="text-xs text-slate-500">{label}</p>
        <p className="mt-1 text-2xl font-bold text-[var(--color-brand-900)]">{value}</p>
        {hint && <p className="mt-1 text-[11px] text-slate-400">{hint}</p>}
      </div>
    </div>
  )
}
