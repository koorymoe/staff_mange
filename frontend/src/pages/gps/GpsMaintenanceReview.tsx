import { useEffect, useState } from 'react'
import { api, type GpsMaintenanceRequest } from '../../api'

const statusMap: Record<string, { label: string; className: string }> = {
  PENDING: { label: 'معلق', className: 'bg-amber-50 text-amber-700' },
  IN_PROGRESS: { label: 'قيد المعالجة', className: 'bg-violet-50 text-violet-700' },
  COMPLETED: { label: 'مكتمل', className: 'bg-emerald-50 text-emerald-700' },
}

export default function GpsMaintenanceReview() {
  const [requests, setRequests] = useState<GpsMaintenanceRequest[]>([])
  const [loading, setLoading] = useState(true)
  const [selected, setSelected] = useState<GpsMaintenanceRequest | null>(null)
  const [notes, setNotes] = useState('')
  const [saving, setSaving] = useState(false)

  const load = () => {
    api.getGpsMaintenance().then(setRequests).finally(() => setLoading(false))
  }
  useEffect(load, [])

  const handleUpdate = async (status: 'IN_PROGRESS' | 'COMPLETED') => {
    if (!selected) return
    setSaving(true)
    try {
      await api.updateGpsMaintenance(selected.id, {
        status, adminNotes: notes,
        resolvedAt: status === 'COMPLETED' ? new Date().toISOString() : undefined,
      })
      setSelected(null)
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <h2 className="mb-6 text-2xl font-bold text-brand-900">طلبات الصيانة 🔧</h2>

      {loading && <p className="py-20 text-center text-slate-400">جاري التحميل...</p>}
      {!loading && requests.length === 0 && (
        <div className="rounded-2xl bg-white py-16 text-center shadow-sm">
          <p className="text-lg text-slate-400">لا توجد طلبات صيانة</p>
        </div>
      )}

      <div className="flex flex-col gap-3">
        {requests.map(req => {
          const s = statusMap[req.status] || statusMap.PENDING
          return (
            <div key={req.id} className="flex items-center justify-between rounded-xl bg-white p-5 shadow-sm">
              <div className="flex-1">
                <div className="mb-1 flex items-center gap-3">
                  <span className="font-bold text-brand-900">{req.customer.fullName} {req.customer.fatherName}</span>
                  <span className={`rounded-full px-2 py-0.5 text-xs font-semibold ${s.className}`}>{s.label}</span>
                </div>
                <p className="mb-1 text-sm text-slate-500">🔧 {req.problemDescription}</p>
                <div className="flex gap-4 text-xs text-slate-400">
                  <span>📞 {req.customer.phone}</span>
                  <span>👤 {req.employee.name}</span>
                  <span>📅 {new Date(req.createdAt).toLocaleDateString('ar-IQ')}</span>
                </div>
              </div>
              <button onClick={() => { setSelected(req); setNotes(req.adminNotes || '') }}
                className="rounded-lg bg-brand-500 px-4 py-2 text-sm font-bold text-white hover:bg-brand-600">تفاصيل ←</button>
            </div>
          )
        })}
      </div>

      {selected && (
        <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 p-4">
          <div className="w-full max-w-xl rounded-2xl bg-white shadow-2xl">
            <div className="flex items-center justify-between border-b border-slate-100 p-5">
              <h3 className="text-lg font-bold text-brand-900">تفاصيل طلب الصيانة</h3>
              <button onClick={() => setSelected(null)} className="text-2xl text-slate-400 hover:text-slate-600">×</button>
            </div>
            <div className="space-y-4 p-6">
              <div className="grid grid-cols-2 gap-3 rounded-xl bg-slate-50 p-4 text-sm">
                <p><span className="text-slate-500">الاسم: </span><b>{selected.customer.fullName} {selected.customer.fatherName} {selected.customer.grandfatherName}</b></p>
                <p><span className="text-slate-500">الهاتف: </span><b>{selected.customer.phone}</b></p>
                <p><span className="text-slate-500">العنوان: </span><b>{selected.customer.address}</b></p>
                <p><span className="text-slate-500">التاريخ: </span><b>{new Date(selected.createdAt).toLocaleDateString('ar-IQ')}</b></p>
              </div>
              <div>
                <p className="mb-1 text-xs font-bold text-slate-500">وصف المشكلة</p>
                <p className="rounded-lg bg-slate-50 p-3 text-sm text-slate-700">{selected.problemDescription}</p>
              </div>
              <div>
                <label className="mb-1 block text-xs font-bold text-slate-500">ملاحظات الإداري</label>
                <textarea rows={3} value={notes} onChange={e => setNotes(e.target.value)} placeholder="أضف ملاحظاتك..."
                  className="w-full rounded-lg border border-slate-200 px-3 py-2.5 text-sm" />
              </div>
            </div>
            <div className="flex gap-3 border-t border-slate-100 p-5">
              <button onClick={() => handleUpdate('IN_PROGRESS')} disabled={saving} className="flex-1 rounded-lg bg-violet-50 py-2.5 text-sm font-bold text-violet-700 hover:bg-violet-100 disabled:opacity-50">قيد المعالجة</button>
              <button onClick={() => handleUpdate('COMPLETED')} disabled={saving} className="flex-1 rounded-lg bg-emerald-50 py-2.5 text-sm font-bold text-emerald-700 hover:bg-emerald-100 disabled:opacity-50">✅ مكتمل</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
