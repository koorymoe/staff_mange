import { useEffect, useMemo, useState } from 'react'
import { api, type Employee, type StaffRequest } from '../api'
import { useSession } from '../session'
import { useSaveGuard } from '../useSaveGuard'
import { matches } from '../utils/search'
import PageHeader from '../components/PageHeader'
import StatTile from '../components/StatTile'
import EmptyState from '../components/EmptyState'
import SaveError from '../components/SaveError'
import SearchBar from '../components/SearchBar'

const statusLabel: Record<string, string> = {
  PENDING: 'قيد الانتظار',
  APPROVED: 'تمت الموافقة',
  REJECTED: 'مرفوض',
  FULFILLED: 'تم التلبية',
}
const STATUS_ORDER = ['PENDING', 'APPROVED', 'FULFILLED', 'REJECTED'] as const
const statusTone = {
  PENDING: 'warning', APPROVED: 'info', FULFILLED: 'success', REJECTED: 'danger',
} as const
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
  // چانت الأخطاء تطلع بـalert() بأربع مواضع — تقطع الشغل وما تنقرا
  // بالموبايل. نفس حارس الحفظ الي تستعمله باقي الشاشات.
  const guard = useSaveGuard()
  const [query, setQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('')

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
      // ⚠️ چان fetch خام بهيدر مبني بالإيد — ما تسري عليه معالجة
      // الجلسة والأخطاء الموحّدة. والدالة موجودة أصلاً بـapi.ts.
      jobs.push(api.getProjectsBrief().then(setProjects).catch(() => {}))
    }
    Promise.all(jobs).finally(() => setLoading(false))
  }
  useEffect(load, [canRequest])

  const toggleEmployee = (id: string) =>
    setSelectedIds(prev => (prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id]))

  const submit = async () => {
    if (selectedIds.length === 0) return guard.run('إرسال الطلب', async () => { throw new Error('اختر موظف واحد على الأقل') })
    if (!neededAt) return guard.run('إرسال الطلب', async () => { throw new Error('حدد وقت الحاجة للكادر') })
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
      guard.run('إرسال الطلب', async () => { throw e })
    } finally {
      setSaving(false)
    }
  }

  const setStatus = async (id: string, status: 'APPROVED' | 'REJECTED' | 'FULFILLED') => {
    const ok = await guard.run('تحديث حالة الطلب', () =>
      api.updateStaffRequestStatus(id, status))
    if (ok !== undefined) load()
  }

  const counts = useMemo(() => {
    const c: Record<string, number> = {}
    for (const r of requests) c[r.status] = (c[r.status] ?? 0) + 1
    return c
  }, [requests])

  // ⚠️ البحث بـmatches مو includes: يعالج الهمزة والتاء المربوطة
  // والأرقام الهندية، فـ«احمد» تلگي «أحمد».
  const shown = requests
    .filter(r => !statusFilter || r.status === statusFilter)
    .filter(r => !query.trim() || matches(
      [r.requester?.name, r.projectName, r.notes, ...r.employees.map(e => e.name)], query))

  return (
    <div className="space-y-4">
      <SaveError message={guard.error} onClose={guard.clear} />

      <PageHeader
        title="👷 طلبات الكادر"
        subtitle="طلبات إدارة المشاريع للفنيين — من الطلب للموافقة للتلبية"
        aside={canRequest && (
          <button onClick={() => setShowForm(f => !f)}
            className="rounded-xl bg-white/20 px-5 py-2.5 text-sm font-bold text-white hover:bg-white/30">
            {showForm ? 'إغلاق' : '+ طلب كادر جديد'}
          </button>
        )}
      />

      {/* بطاقات العدّ — وكل وحدة ترشّح القائمة، حتى الرقم يودّي لشغله */}
      <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        {STATUS_ORDER.map((k) => (
          <StatTile
            key={k}
            label={statusLabel[k]}
            value={counts[k] ?? 0}
            tone={statusTone[k]}
            hint={statusFilter === k ? 'مفعّلة — اضغط للإلغاء' : 'اضغط للتصفية'}
            onClick={() => setStatusFilter(statusFilter === k ? '' : k)}
          />
        ))}
      </div>

      <SearchBar
        value={query}
        onChange={setQuery}
        placeholder="ابحث باسم مقدّم الطلب أو المشروع أو الموظف..."
      >
        {statusFilter && (
          <button onClick={() => setStatusFilter('')}
            className="rounded-xl border px-4 py-2.5 text-sm font-medium"
            style={{ borderColor: 'var(--bd-line)', color: 'var(--t-body)' }}>
            امسح التصفية ✕
          </button>
        )}
      </SearchBar>

      {isHandler && (
        <p className="rounded-xl bg-blue-50 p-3 text-sm text-blue-800">
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
      {!loading && shown.length === 0 && (
        <div className="rounded-2xl border" style={{ backgroundColor: 'var(--sf-card)', borderColor: 'var(--bd-line)' }}>
          <EmptyState
            title={requests.length === 0 ? 'لا توجد طلبات كادر بعد' : 'ماكو طلب يطابق البحث أو التصفية'}
            reason={requests.length === 0
              ? 'الطلبات تجي من إدارة المشاريع لمن تحتاج فنيين لموقع.'
              : `اكو ${requests.length} طلب بالمجموع — امسح البحث أو التصفية حتى تشوفهن.`}
          />
        </div>
      )}

      <div className="flex flex-col gap-3">
        {shown.map(req => (
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
                    className="rounded-lg bg-emerald-500 px-4 py-2 text-xs font-bold text-white hover:bg-emerald-700">تم التلبية ✅</button>
                )}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
