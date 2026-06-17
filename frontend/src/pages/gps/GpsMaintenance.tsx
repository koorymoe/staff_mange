import { useState, useEffect } from 'react'
import { api } from '../../api'

const statusBadge: Record<string, string> = {
  PENDING: 'bg-yellow-100 text-yellow-800',
  IN_PROGRESS: 'bg-blue-100 text-blue-800',
  COMPLETED: 'bg-green-100 text-green-800',
}
const statusLabel: Record<string, string> = {
  PENDING: 'قيد الانتظار',
  IN_PROGRESS: 'قيد المعالجة',
  COMPLETED: 'مكتمل',
}

export default function GpsMaintenance() {
  const [requests, setRequests] = useState<any[]>([])
  const [customers, setCustomers] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showForm, setShowForm] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  const [customerId, setCustomerId] = useState('')
  const [problemDescription, setProblemDescription] = useState('')

  const load = () => {
    setLoading(true)
    Promise.all([api.getGpsMaintenance(), api.getGpsCustomers()])
      .then(([r, c]) => { setRequests(r); setCustomers(c) })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    try {
      await api.createGpsMaintenance({ customerId, problemDescription })
      setCustomerId(''); setProblemDescription('')
      setShowForm(false)
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  const handleUpdateStatus = async (id: string, status: string, adminNotes?: string) => {
    try {
      const notes = status !== 'PENDING' ? (adminNotes || prompt('ملاحظات المسؤول:') || '') : ''
      await api.updateGpsMaintenance(id, { status, adminNotes: notes })
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    }
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">طلبات صيانة GPS</h2>

      <div className="mt-4">
        <button onClick={() => setShowForm(!showForm)} className="rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md transition-all hover:shadow-lg">
          {showForm ? 'إلغاء' : 'طلب صيانة جديد'}
        </button>
      </div>

      {showForm && (
        <form onSubmit={handleSubmit} className="mt-4 grid grid-cols-1 gap-4 rounded-xl bg-white p-6 shadow-md sm:grid-cols-2">
          <div>
            <label className="mb-1 block text-sm font-medium text-slate-600">الزبون</label>
            <select required value={customerId} onChange={(e) => setCustomerId(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500">
              <option value="">اختر الزبون</option>
              {customers.map((c) => <option key={c.id} value={c.id}>{c.fullName}</option>)}
            </select>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-slate-600">وصف المشكلة</label>
            <textarea required value={problemDescription} onChange={(e) => setProblemDescription(e.target.value)} rows={3} className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500" />
          </div>
          <div className="sm:col-span-2">
            <button type="submit" disabled={submitting} className="rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md disabled:opacity-50">
              {submitting ? 'جاري الحفظ...' : 'إرسال طلب الصيانة'}
            </button>
          </div>
        </form>
      )}

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">تعذر الاتصال بالخادم: {error}</p>}

      {!loading && !error && (
        <div className="mt-6 overflow-hidden rounded-xl bg-white shadow-md">
          <table className="w-full text-right">
            <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
              <tr>
                <th className="px-4 py-3 text-sm font-semibold">الزبون</th>
                <th className="px-4 py-3 text-sm font-semibold">وصف المشكلة</th>
                <th className="px-4 py-3 text-sm font-semibold">الحالة</th>
                <th className="px-4 py-3 text-sm font-semibold">ملاحظات المسؤول</th>
                <th className="px-4 py-3 text-sm font-semibold">التاريخ</th>
                <th className="px-4 py-3 text-sm font-semibold">إجراءات</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {requests.map((r) => (
                <tr key={r.id} className="hover:bg-slate-50">
                  <td className="px-4 py-3 font-medium">{r.customer?.fullName || '-'}</td>
                  <td className="px-4 py-3 text-slate-600 text-sm">{r.problemDescription || '-'}</td>
                  <td className="px-4 py-3">
                    <span className={`rounded-full px-3 py-1 text-xs font-medium ${statusBadge[r.status] || 'bg-slate-100 text-slate-800'}`}>
                      {statusLabel[r.status] || r.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-slate-500">{r.adminNotes || '-'}</td>
                  <td className="px-4 py-3 text-sm text-slate-500">{r.createdAt ? new Date(r.createdAt).toLocaleDateString('ar-IQ') : '-'}</td>
                  <td className="px-4 py-3">
                    <div className="flex flex-wrap gap-2">
                      {r.status === 'PENDING' && (
                        <button onClick={() => handleUpdateStatus(r.id, 'IN_PROGRESS')} className="rounded-lg bg-blue-600 px-3 py-1 text-xs text-white hover:bg-blue-700">بدء المعالجة</button>
                      )}
                      {r.status === 'IN_PROGRESS' && (
                        <button onClick={() => handleUpdateStatus(r.id, 'COMPLETED')} className="rounded-lg bg-green-600 px-3 py-1 text-xs text-white hover:bg-green-700">إكمال</button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
              {requests.length === 0 && (
                <tr><td colSpan={6} className="px-4 py-6 text-center text-slate-400">لا يوجد طلبات صيانة</td></tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
