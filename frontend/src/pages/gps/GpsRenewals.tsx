import { useState, useEffect } from 'react'
import { api } from '../../api'

const statusBadge: Record<string, string> = {
  PENDING: 'bg-yellow-100 text-yellow-800',
  APPROVED: 'bg-green-100 text-green-800',
  REJECTED: 'bg-red-100 text-red-800',
}
const statusLabel: Record<string, string> = {
  PENDING: 'قيد الانتظار',
  APPROVED: 'موافق عليه',
  REJECTED: 'مرفوض',
}

export default function GpsRenewals() {
  const [renewals, setRenewals] = useState<any[]>([])
  const [customers, setCustomers] = useState<any[]>([])
  const [devices, setDevices] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showForm, setShowForm] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  const [customerId, setCustomerId] = useState('')
  const [deviceId, setDeviceId] = useState('')
  const [subscriptionType, setSubscriptionType] = useState('')

  const load = () => {
    setLoading(true)
    Promise.all([api.getGpsRenewals(), api.getGpsCustomers(), api.getGpsDevices()])
      .then(([r, c, d]) => { setRenewals(r); setCustomers(c); setDevices(d) })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    try {
      await api.createGpsRenewal({ customerId, deviceRequestId: deviceId, subscriptionType })
      setCustomerId(''); setDeviceId(''); setSubscriptionType('')
      setShowForm(false)
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  const handleAction = async (id: string, status: string) => {
    try {
      await api.updateGpsRenewal(id, { status })
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    }
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">طلبات التجديد</h2>

      <div className="mt-4">
        <button onClick={() => setShowForm(!showForm)} className="rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md transition-all hover:shadow-lg">
          {showForm ? 'إلغاء' : 'طلب تجديد جديد'}
        </button>
      </div>

      {showForm && (
        <form onSubmit={handleSubmit} className="mt-4 grid grid-cols-1 gap-4 rounded-xl bg-white p-6 shadow-md sm:grid-cols-3">
          <div>
            <label className="mb-1 block text-sm font-medium text-slate-600">الزبون</label>
            <select required value={customerId} onChange={(e) => setCustomerId(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500">
              <option value="">اختر الزبون</option>
              {customers.map((c) => <option key={c.id} value={c.id}>{c.fullName}</option>)}
            </select>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-slate-600">طلب الجهاز</label>
            <select required value={deviceId} onChange={(e) => setDeviceId(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500">
              <option value="">اختر الجهاز</option>
              {devices.map((d) => <option key={d.id} value={d.id}>{d.customer?.fullName} - {d.purchaseType}</option>)}
            </select>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-slate-600">نوع الاشتراك</label>
            <input required value={subscriptionType} onChange={(e) => setSubscriptionType(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500" />
          </div>
          <div className="sm:col-span-3">
            <button type="submit" disabled={submitting} className="rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md disabled:opacity-50">
              {submitting ? 'جاري الحفظ...' : 'إرسال طلب التجديد'}
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
                <th className="px-4 py-3 text-sm font-semibold">طلب الجهاز</th>
                <th className="px-4 py-3 text-sm font-semibold">نوع الاشتراك</th>
                <th className="px-4 py-3 text-sm font-semibold">تاريخ الانتهاء الجديد</th>
                <th className="px-4 py-3 text-sm font-semibold">الحالة</th>
                <th className="px-4 py-3 text-sm font-semibold">إجراءات</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {renewals.map((r) => (
                <tr key={r.id} className="hover:bg-slate-50">
                  <td className="px-4 py-3 font-medium">{r.customer?.fullName || '-'}</td>
                  <td className="px-4 py-3 text-slate-600">{r.deviceRequest?.customer?.fullName || '-'}</td>
                  <td className="px-4 py-3 text-slate-600">{r.subscriptionType || '-'}</td>
                  <td className="px-4 py-3 text-sm text-slate-500">{r.newEndDate ? new Date(r.newEndDate).toLocaleDateString('ar-IQ') : '-'}</td>
                  <td className="px-4 py-3">
                    <span className={`rounded-full px-3 py-1 text-xs font-medium ${statusBadge[r.status] || 'bg-slate-100 text-slate-800'}`}>
                      {statusLabel[r.status] || r.status}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    {r.status === 'PENDING' && (
                      <div className="flex gap-2">
                        <button onClick={() => handleAction(r.id, 'APPROVED')} className="rounded-lg bg-green-600 px-3 py-1 text-xs text-white hover:bg-green-700">موافقة</button>
                        <button onClick={() => handleAction(r.id, 'REJECTED')} className="rounded-lg bg-red-600 px-3 py-1 text-xs text-white hover:bg-red-700">رفض</button>
                      </div>
                    )}
                  </td>
                </tr>
              ))}
              {renewals.length === 0 && (
                <tr><td colSpan={6} className="px-4 py-6 text-center text-slate-400">لا يوجد طلبات تجديد</td></tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
