import { useState, useEffect } from 'react'
import { api } from '../../api'

const statusBadge: Record<string, string> = {
  PENDING: 'bg-yellow-100 text-yellow-800',
  APPROVED: 'bg-blue-100 text-blue-800',
  REJECTED: 'bg-red-100 text-red-800',
  DELIVERED: 'bg-green-100 text-green-800',
  CHECKED: 'bg-indigo-100 text-indigo-800',
  ACTIVATED: 'bg-teal-100 text-teal-800',
}

const statusLabel: Record<string, string> = {
  PENDING: 'قيد الانتظار',
  APPROVED: 'موافق عليه',
  REJECTED: 'مرفوض',
  DELIVERED: 'تم التسليم',
  CHECKED: 'تم الفحص',
  ACTIVATED: 'مفعّل',
}

export default function GpsDevices() {
  const [devices, setDevices] = useState<any[]>([])
  const [customers, setCustomers] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showForm, setShowForm] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  const [customerId, setCustomerId] = useState('')
  const [purchaseType, setPurchaseType] = useState('DEVICE_SIM')
  const [subscriptionType, setSubscriptionType] = useState('')

  const load = () => {
    setLoading(true)
    Promise.all([api.getGpsDevices(), api.getGpsCustomers()])
      .then(([d, c]) => { setDevices(d); setCustomers(c) })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    try {
      await api.createGpsDevice({ customerId, purchaseType, subscriptionType })
      setCustomerId(''); setPurchaseType('DEVICE_SIM'); setSubscriptionType('')
      setShowForm(false)
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  const handleAction = async (id: string, action: string) => {
    try {
      await api.updateGpsDevice(id, { action })
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    }
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">طلبات الأجهزة</h2>

      <div className="mt-4">
        <button onClick={() => setShowForm(!showForm)} className="rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md transition-all hover:shadow-lg">
          {showForm ? 'إلغاء' : 'طلب جهاز جديد'}
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
            <label className="mb-1 block text-sm font-medium text-slate-600">نوع الشراء</label>
            <select value={purchaseType} onChange={(e) => setPurchaseType(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500">
              <option value="DEVICE_SIM">جهاز + شريحة</option>
              <option value="DEVICE_ONLY">جهاز فقط</option>
            </select>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-slate-600">نوع الاشتراك</label>
            <input value={subscriptionType} onChange={(e) => setSubscriptionType(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500" />
          </div>
          <div className="sm:col-span-3">
            <button type="submit" disabled={submitting} className="rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md disabled:opacity-50">
              {submitting ? 'جاري الحفظ...' : 'إرسال الطلب'}
            </button>
          </div>
        </form>
      )}

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">تعذر الاتصال بالخادم: {error}</p>}

      {!loading && !error && (
        <div className="mt-6 overflow-x-auto rounded-xl bg-white shadow-md">
          <table className="w-full text-right">
            <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
              <tr>
                <th className="px-4 py-3 text-sm font-semibold">الزبون</th>
                <th className="px-4 py-3 text-sm font-semibold">الموظف</th>
                <th className="px-4 py-3 text-sm font-semibold">نوع الشراء</th>
                <th className="px-4 py-3 text-sm font-semibold">الاشتراك</th>
                <th className="px-4 py-3 text-sm font-semibold">الحالة</th>
                <th className="px-4 py-3 text-sm font-semibold">التاريخ</th>
                <th className="px-4 py-3 text-sm font-semibold">إجراءات</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {devices.map((d) => (
                <tr key={d.id} className="hover:bg-slate-50">
                  <td className="px-4 py-3 font-medium">{d.customer?.fullName || '-'}</td>
                  <td className="px-4 py-3 text-slate-600">{d.employee?.name || '-'}</td>
                  <td className="px-4 py-3 text-slate-600">{d.purchaseType === 'DEVICE_SIM' ? 'جهاز + شريحة' : 'جهاز فقط'}</td>
                  <td className="px-4 py-3 text-slate-600">{d.subscriptionType || '-'}</td>
                  <td className="px-4 py-3">
                    <span className={`rounded-full px-3 py-1 text-xs font-medium ${statusBadge[d.status] || 'bg-slate-100 text-slate-800'}`}>
                      {statusLabel[d.status] || d.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-slate-500">{d.createdAt ? new Date(d.createdAt).toLocaleDateString('ar-IQ') : '-'}</td>
                  <td className="px-4 py-3">
                    <div className="flex flex-wrap gap-2">
                      {d.status === 'PENDING' && (
                        <>
                          <button onClick={() => handleAction(d.id, 'approve')} className="rounded-lg bg-blue-600 px-3 py-1 text-xs text-white hover:bg-blue-700">موافقة</button>
                          <button onClick={() => handleAction(d.id, 'reject')} className="rounded-lg bg-red-600 px-3 py-1 text-xs text-white hover:bg-red-700">رفض</button>
                        </>
                      )}
                      {d.status === 'APPROVED' && (
                        <button onClick={() => handleAction(d.id, 'check')} className="rounded-lg bg-indigo-600 px-3 py-1 text-xs text-white hover:bg-indigo-700">فحص</button>
                      )}
                      {d.status === 'CHECKED' && (
                        <button onClick={() => handleAction(d.id, 'activate')} className="rounded-lg bg-teal-600 px-3 py-1 text-xs text-white hover:bg-teal-700">تفعيل</button>
                      )}
                      {d.status === 'ACTIVATED' && (
                        <button onClick={() => handleAction(d.id, 'deliver')} className="rounded-lg bg-green-600 px-3 py-1 text-xs text-white hover:bg-green-700">تسليم</button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
              {devices.length === 0 && (
                <tr><td colSpan={7} className="px-4 py-6 text-center text-slate-400">لا يوجد طلبات أجهزة</td></tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
