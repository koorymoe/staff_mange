import { useEffect, useState } from 'react'
import { api, type QualityIssue, type Employee } from '../api'

const STATUS_LABELS: Record<string, string> = {
  OPEN: 'مفتوحة',
  IN_PROGRESS: 'قيد المعالجة',
  RESOLVED: 'تم الحل',
}
const STATUS_COLORS: Record<string, string> = {
  OPEN: 'bg-red-100 text-red-700',
  IN_PROGRESS: 'bg-amber-100 text-amber-700',
  RESOLVED: 'bg-emerald-100 text-emerald-700',
}

export default function QualityPage() {
  const [issues, setIssues] = useState<QualityIssue[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [filter, setFilter] = useState<'ALL' | 'EXECUTION' | 'OVERSIGHT'>('ALL')

  const [category, setCategory] = useState<'EXECUTION' | 'OVERSIGHT'>('EXECUTION')
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [responsible, setResponsible] = useState('')

  const load = () => api.getQualityIssues(filter === 'ALL' ? undefined : filter).then(setIssues)

  // `load` is redefined every render (not memoized) and already reflects `filter`
  // via closure; adding `load` itself as a dep would re-run this effect every render.
  // eslint-disable-next-line react-hooks/exhaustive-deps
  useEffect(() => { load() }, [filter])
  useEffect(() => { api.getEmployees().then(setEmployees) }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!title.trim()) return
    await api.createQualityIssue({ category, title, description: description || undefined, responsibleEmployeeId: responsible || undefined })
    setTitle(''); setDescription(''); setResponsible('')
    load()
  }

  const handleStatus = async (id: string, status: 'IN_PROGRESS' | 'RESOLVED') => {
    await api.updateQualityIssue(id, { status })
    load()
  }

  return (
    <div dir="rtl" className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-brand-900">الجودة</h2>
        <p className="mt-1 text-slate-500">مشاكل التنفيذ الميدانية، ومشاكل الرقابة الإدارية/القيادية والمسؤولين عنها</p>
      </div>

      <form onSubmit={handleSubmit} className="grid grid-cols-1 gap-3 rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)] sm:grid-cols-2">
        <select value={category} onChange={(e) => setCategory(e.target.value as typeof category)} className="rounded-lg border border-slate-300 px-3 py-2">
          <option value="EXECUTION">مشكلة تنفيذية (ميدانية)</option>
          <option value="OVERSIGHT">مشكلة رقابية (إدارية/قيادية)</option>
        </select>
        <select value={responsible} onChange={(e) => setResponsible(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2">
          <option value="">-- المسؤول (اختياري) --</option>
          {employees.map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
        </select>
        <input required placeholder="عنوان المشكلة" value={title} onChange={(e) => setTitle(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2 sm:col-span-2" />
        <textarea placeholder="تفاصيل إضافية (اختياري)" value={description} onChange={(e) => setDescription(e.target.value)} rows={2} className="rounded-lg border border-slate-300 px-3 py-2 sm:col-span-2" />
        <button type="submit" className="sm:col-span-2 rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-5 py-2 font-medium text-white shadow-md">تسجيل المشكلة</button>
      </form>

      <div className="flex gap-2">
        {([
          { key: 'ALL', label: 'الكل' },
          { key: 'EXECUTION', label: 'تنفيذية' },
          { key: 'OVERSIGHT', label: 'رقابية' },
        ] as const).map((t) => (
          <button
            key={t.key}
            onClick={() => setFilter(t.key)}
            className={`rounded-lg px-4 py-2 text-sm font-bold transition-colors ${filter === t.key ? 'bg-brand-600 text-white' : 'bg-white text-slate-600 border border-slate-200 hover:bg-slate-50'}`}
          >
            {t.label}
          </button>
        ))}
      </div>

      <div className="space-y-3">
        {issues.map((issue) => (
          <div key={issue.id} className="rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <div className="flex items-center justify-between">
              <span className={`rounded-full px-2 py-1 text-xs font-bold ${issue.category === 'EXECUTION' ? 'bg-blue-100 text-blue-700' : 'bg-purple-100 text-purple-700'}`}>
                {issue.category === 'EXECUTION' ? 'تنفيذية' : 'رقابية'}
              </span>
              <span className={`rounded-full px-2 py-1 text-xs font-bold ${STATUS_COLORS[issue.status]}`}>
                {STATUS_LABELS[issue.status]}
              </span>
            </div>
            <h3 className="mt-2 font-bold text-slate-800">{issue.title}</h3>
            {issue.description && <p className="mt-1 text-sm text-slate-600">{issue.description}</p>}
            <div className="mt-2 flex flex-wrap items-center gap-4 text-xs text-slate-500">
              <span>المسؤول: {issue.responsibleEmployee?.name || '-'}</span>
              <span>أُبلغ من: {issue.reportedBy?.name || '-'}</span>
              <span>{new Date(issue.createdAt).toLocaleDateString('ar-IQ')}</span>
            </div>
            {issue.status !== 'RESOLVED' && (
              <div className="mt-3 flex gap-2">
                {issue.status === 'OPEN' && (
                  <button onClick={() => handleStatus(issue.id, 'IN_PROGRESS')} className="rounded-lg bg-amber-50 px-3 py-1 text-xs font-bold text-amber-700 hover:bg-amber-100">
                    بدء المعالجة
                  </button>
                )}
                <button onClick={() => handleStatus(issue.id, 'RESOLVED')} className="rounded-lg bg-emerald-50 px-3 py-1 text-xs font-bold text-emerald-700 hover:bg-emerald-100">
                  تم الحل
                </button>
              </div>
            )}
          </div>
        ))}
        {issues.length === 0 && (
          <div className="rounded-xl border border-white bg-white p-10 text-center text-slate-400 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            لا توجد مشاكل مسجلة
          </div>
        )}
      </div>
    </div>
  )
}
