import { useEffect, useState } from 'react'
import { api, type Employee, type StaffRequest } from '../api'
import { useSession } from '../session'

const statusLabel: Record<string, string> = {
  PENDING: 'قيد الانتظار',
  APPROVED: 'تمت الموافقة',
  REJECTED: 'مرفوض',
  FULFILLED: 'تم التلبية',
}
const statusStyle: Record<string, string> = {
  PENDING: 'bg-amber-50 text-amber-700',
  APPROVED: 'bg-blue-50 text-blue-700',
  REJECTED: 'bg-red-50 text-red-600',
  FULFILLED: 'bg-emerald-50 text-emerald-700',
}

export default function StaffRequestsPage() {
  const { employee, permissions } = useSession()
  const isHandler = employee?.role === 'ADMIN' || employee?.role === 'HR_COORDINATOR'
  const canRequest = employee?.role === 'ADMIN' || permissions.includes('project_management')

  const [requests, setRequests] = useState<StaffRequest[]>([])
  const [technicians, setTechnicians] = useState<Employee[]>([])
  const [projects, setProjects] = useState<{ id: string; name: string; code: string }[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [saving, setSaving] = useState(false)

  // نموذج الطلب
  const [selectedIds, setSelectedIds] = useState<string[]>([])
  const [neededAt, setNeededAt] = useState('')
  const [durationHours, setDurationHours] = useState('8')
  const [notes, setNotes] = useState('')
  const [projectId, setProjectId] = useState('')

  const load = () => {
    const jobs: Promise<unknown>[] = [api.getStaffRequests().then(setRequests)]
    if (canRequest) {
      jobs.push(api.getEmployees().then(all => setTechnicians(all.filter(e => e.role === 'TECHNICIAN' && e.status === 'ACTIVE'))))
      const token = localStorage.getItem('authToken')
      jobs.push(
        fetch(`${import.meta.env.VITE_API_URL || 'http://localhost:4000/api'}/projects`, {
          headers: token ? { Authorization: `Bearer ${token}` } : {},
        })
          .then(r => (r.ok ? r.json() : null))
          .then(d => setProjects(d?.projects || []))
          .catch(() => {})
      )
    }
    Promise.all(jobs).finally(() => setLoading(false))
  }
  useEffect(load, [])

  const toggleEmployee = (id: string) =>
    setSelectedIds(prev => (prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id]))

  const submit = async () => {
    if (selectedIds.length === 0) return alert('اختر موظف واحد على الأقل')
    if (!neededAt) return alert('حدد وقت الحاجة للكادر')
    setSaving(true)
    try {
      await api.createStaffRequest({
        projectId: projectId || null,
        neededAt: new Date(neededAt).toISOString(),
        durationHours: Number(durationHours) || 8,
        notes: notes || null,
        employeeIds: selectedIds,
      })
      setShowForm(false)
      setSelectedIds([]); setNeededAt(''); setDurationHours('8'); setNotes(''); setProjectId('')
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر إرسال الطلب')
    } finally {
      setSaving(false)
    }
  }

  const setStatus = async (id: string, status: 'APPROVED' | 'REJECTED' | 'FULFILLED') => {
    try {
      await api.updateStaffRequestStatus(id, status)
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر تحديث الحالة')
    }
  }

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <h2 className="text-2xl font-bold text-brand-900">طلبات الكادر 👷</h2>
        {canRequest && (
          <button onClick={() => setShowForm(f => !f)}
            className="rounded-xl bg-brand-500 px-5 py-2.5 text-sm font-bold text-white hover:bg-brand-600">
            {showForm ? 'إغلاق' : '+ طلب كادر جديد'}
          </button>
        )}
      </div>

      {isHandler && (
        <p className="mb-4 rounded-xl bg-blue-50 p-3 text-sm text-blue-800">
          الطلبات الواردة من إدارة المشاريع — وافق عليها وبلّغ الموظفين، وبعد ما يلتحقون بالموقع علّمها "تم التلبية".
        </p>
      )}

      {showForm && canRequest && (
        <div className="mb-6 rounded-2xl bg-white p-6 shadow-sm">
          <h3 className="mb-4 text-lg font-bold text-brand-900">طلب كادر جديد</h3>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
            <div>
              <label className="mb-1 block text-xs font-bold text-slate-500">موعد الحاجة للكادر *</label>
              <input type="datetime-local" value={neededAt} onChange={e => setNeededAt(e.target.value)}
                className="w-full rounded-xl border border-slate-200 px-3 py-2.5 text-sm outline-none focus:border-brand-500" />
            </div>
            <div>
              <label className="mb-1 block text-xs font-bold text-slate-500">المدة (ساعات)</label>
              <input type="number" min="1" value={durationHours} onChange={e => setDurationHours(e.target.value)}
                className="w-full rounded-xl border border-slate-200 px-3 py-2.5 text-sm outline-none focus:border-brand-500" />
            </div>
            <div>
              <label className="mb-1 block text-xs font-bold text-slate-500">المشروع (اختياري)</label>
              <select value={projectId} onChange={e => setProjectId(e.target.value)}
                className="w-full rounded-xl border border-slate-200 px-3 py-2.5 text-sm outline-none focus:border-brand-500">
                <option value="">— بدون ربط بمشروع —</option>
                {projects.map(p => <option key={p.id} value={p.id}>{p.code} — {p.name}</option>)}
              </select>
            </div>
          </div>

          <label className="mb-1 mt-4 block text-xs font-bold text-slate-500">الموظفون المطلوبون (من كادر الشد) *</label>
          <div className="grid max-h-52 grid-cols-2 gap-2 overflow-y-auto sm:grid-cols-3">
            {technicians.map(t => (
              <label key={t.id} className={`flex cursor-pointer items-center gap-2 rounded-xl border px-3 py-2 text-sm transition-colors ${
                selectedIds.includes(t.id) ? 'border-brand-500 bg-brand-50 font-bold text-brand-700' : 'border-slate-200 text-slate-600'
              }`}>
                <input type="checkbox" checked={selectedIds.includes(t.id)} onChange={() => toggleEmployee(t.id)} />
                {t.name}{t.position ? ` — ${t.position}` : ''}
              </label>
            ))}
            {technicians.length === 0 && <p className="col-span-full text-sm text-slate-400">لا يوجد فنيون فعالون حالياً</p>}
          </div>

          <label className="mb-1 mt-4 block text-xs font-bold text-slate-500">ملاحظات</label>
          <textarea value={notes} onChange={e => setNotes(e.target.value)} rows={2} placeholder="مثال: يحتاجون معدات سحب كيبل..."
            className="w-full rounded-xl border border-slate-200 px-3 py-2.5 text-sm outline-none focus:border-brand-500" />

          <button onClick={submit} disabled={saving}
            className="mt-4 rounded-xl bg-brand-500 px-8 py-3 text-sm font-bold text-white hover:bg-brand-600 disabled:opacity-50">
            {saving ? 'جاري الإرسال...' : '📤 إرسال الطلب لإدارة الكوادر'}
          </button>
        </div>
      )}

      {loading && <p className="py-16 text-center text-slate-400">جاري التحميل...</p>}
      {!loading && requests.length === 0 && (
        <div className="rounded-2xl bg-white py-16 text-center shadow-sm">
          <div className="mb-3 text-5xl">📭</div>
          <p className="text-slate-400">لا توجد طلبات كادر بعد</p>
        </div>
      )}

      <div className="flex flex-col gap-3">
        {requests.map(req => (
          <div key={req.id} className="rounded-2xl bg-white p-5 shadow-sm">
            <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
              <div className="flex items-center gap-3">
                <span className={`rounded-full px-3 py-1 text-xs font-bold ${statusStyle[req.status]}`}>{statusLabel[req.status]}</span>
                {req.projectName && <span className="rounded-full bg-violet-50 px-3 py-1 text-xs font-bold text-violet-700">🏗️ {req.projectName}</span>}
              </div>
              <span className="text-xs text-slate-400">قدمه: {req.requester?.name || '—'} · {new Date(req.createdAt).toLocaleDateString('ar-IQ')}</span>
            </div>

            <div className="mb-2 flex flex-wrap gap-4 text-sm text-slate-600">
              <span>📅 الموعد: <b className="text-brand-900">{new Date(req.neededAt).toLocaleString('ar-IQ')}</b></span>
              <span>⏱️ المدة: <b className="text-brand-900">{req.durationHours} ساعة</b></span>
            </div>

            <div className="mb-2 flex flex-wrap gap-1.5">
              {req.employees.map(e => (
                <span key={e.id} className="rounded-full bg-slate-100 px-3 py-1 text-xs font-bold text-slate-700">👤 {e.name}</span>
              ))}
            </div>

            {req.notes && <p className="mb-2 rounded-xl bg-slate-50 p-3 text-sm text-slate-600">{req.notes}</p>}
            {req.handledBy && (
              <p className="text-xs text-slate-400">عالجه: {req.handledBy.name} {req.handledAt ? '· ' + new Date(req.handledAt).toLocaleString('ar-IQ') : ''}</p>
            )}

            {isHandler && (req.status === 'PENDING' || req.status === 'APPROVED') && (
              <div className="mt-3 flex gap-2 border-t border-slate-100 pt-3">
                {req.status === 'PENDING' && (
                  <>
                    <button onClick={() => setStatus(req.id, 'APPROVED')}
                      className="rounded-lg bg-blue-500 px-4 py-2 text-xs font-bold text-white hover:bg-blue-600">موافقة ✔</button>
                    <button onClick={() => setStatus(req.id, 'REJECTED')}
                      className="rounded-lg border border-red-200 px-4 py-2 text-xs font-bold text-red-600 hover:bg-red-50">رفض ✖</button>
                  </>
                )}
                {req.status === 'APPROVED' && (
                  <button onClick={() => setStatus(req.id, 'FULFILLED')}
                    className="rounded-lg bg-emerald-500 px-4 py-2 text-xs font-bold text-white hover:bg-emerald-600">تم التلبية ✅</button>
                )}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
