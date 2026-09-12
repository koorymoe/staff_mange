import { useEffect, useState } from 'react'
import { api, type DeviceMaintenanceTicket } from '../api'

const statusLabels: Record<string, string> = {
  NEW: 'جديدة',
  IN_PROGRESS: 'قيد الصيانة',
  DELIVERED: 'تم التسليم',
}

const statusColors: Record<string, string> = {
  NEW: 'bg-amber-100 text-amber-700',
  IN_PROGRESS: 'bg-blue-100 text-blue-700',
  DELIVERED: 'bg-green-100 text-green-700',
}

export default function DeviceMaintenancePage() {
  const [tickets, setTickets] = useState<DeviceMaintenanceTicket[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const [appointmentDate, setAppointmentDate] = useState('')
  const [customerCode, setCustomerCode] = useState('')
  const [deviceTypeName, setDeviceTypeName] = useState('')
  const [problem, setProblem] = useState('')
  const [deviceSerial, setDeviceSerial] = useState('')

  const load = async () => {
    try {
      const list = await api.getDeviceMaintenanceTickets()
      setTickets(list)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'حدث خطأ')
    } finally {
      setLoading(false)
    }
  }

  // eslint-disable-next-line react-hooks/set-state-in-effect
  useEffect(() => { load() }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const code = parseInt(customerCode, 10)
    if (!code || !deviceTypeName.trim() || !problem.trim()) {
      alert('كود الزبون ونوع الجهاز والمشكلة حقول مطلوبة')
      return
    }
    setSubmitting(true)
    try {
      await api.createDeviceMaintenanceTicket({
        appointmentDate: appointmentDate || null,
        customerCode: code,
        deviceTypeName,
        problem,
        deviceSerial: deviceSerial || null,
      })
      setAppointmentDate('')
      setCustomerCode('')
      setDeviceTypeName('')
      setProblem('')
      setDeviceSerial('')
      await load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر حفظ التذكرة')
    } finally {
      setSubmitting(false)
    }
  }

  const markReceived = async (id: string) => {
    try {
      await api.updateDeviceMaintenanceTicket(id, { markReceived: true })
      await load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر التحديث')
    }
  }

  const markDelivered = async (id: string) => {
    try {
      await api.updateDeviceMaintenanceTicket(id, { markDelivered: true })
      await load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر التحديث')
    }
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">صيانة الأجهزة</h2>
      <p className="mt-1 text-slate-500">استلام وتسليم أجهزة الزبائن للصيانة (كاميرا، إنذار، بصمة...)</p>

      <form onSubmit={handleSubmit} className="mt-6 grid grid-cols-1 gap-4 rounded-xl bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)] sm:grid-cols-2">
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">تاريخ موعد الزبون</label>
          <input type="date" value={appointmentDate} onChange={(e) => setAppointmentDate(e.target.value)}
            className="w-full rounded-lg border border-slate-200 px-3 py-2" />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">كود الزبون *</label>
          <input type="number" value={customerCode} onChange={(e) => setCustomerCode(e.target.value)}
            placeholder="مثال: 14" className="w-full rounded-lg border border-slate-200 px-3 py-2" />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">نوع واسم الجهاز *</label>
          <input value={deviceTypeName} onChange={(e) => setDeviceTypeName(e.target.value)}
            placeholder="مثال: كاميرا هيكفيجن" className="w-full rounded-lg border border-slate-200 px-3 py-2" />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">سيريال الجهاز</label>
          <input value={deviceSerial} onChange={(e) => setDeviceSerial(e.target.value)}
            className="w-full rounded-lg border border-slate-200 px-3 py-2" />
        </div>
        <div className="sm:col-span-2">
          <label className="mb-1 block text-sm font-medium text-slate-600">المشكلة *</label>
          <textarea value={problem} onChange={(e) => setProblem(e.target.value)} rows={2}
            className="w-full rounded-lg border border-slate-200 px-3 py-2" />
        </div>
        <div className="sm:col-span-2">
          <button type="submit" disabled={submitting}
            className="w-full rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 py-3 font-bold text-white shadow-md disabled:opacity-50">
            {submitting ? 'جارٍ الحفظ...' : 'حفظ تذكرة الصيانة'}
          </button>
        </div>
      </form>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">تعذر الاتصال بالخادم: {error}</p>}

      {!loading && !error && (
        <div className="mt-6 space-y-3">
          {tickets.length === 0 ? (
            <div className="rounded-xl bg-white p-12 text-center text-slate-400 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              لا توجد تذاكر صيانة بعد
            </div>
          ) : (
            tickets.map((t) => (
              <div key={t.id} className="rounded-xl bg-white p-4 shadow-[0_2px_10px_rgba(15,32,64,0.05)]">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <div>
                    <div className="font-bold text-brand-900">{t.deviceTypeName} — {t.customer?.name || '—'}</div>
                    <div className="text-xs text-slate-400 mt-0.5">
                      فاتورة: {t.invoiceNumber} · كود الزبون: {t.customer?.code || '—'} · هاتف: {t.customer?.phone || '—'}
                    </div>
                  </div>
                  <span className={`rounded-full px-3 py-1 text-xs font-bold ${statusColors[t.status]}`}>
                    {statusLabels[t.status]}
                  </span>
                </div>
                <p className="mt-2 text-sm text-slate-600">المشكلة: {t.problem}</p>
                {t.deviceSerial && <p className="mt-1 text-xs text-slate-400">سيريال: {t.deviceSerial}</p>}
                <div className="mt-3 flex gap-2">
                  {!t.receivedAt && (
                    <button onClick={() => markReceived(t.id)}
                      className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-bold text-white">استلام الجهاز من الزبون</button>
                  )}
                  {t.receivedAt && !t.deliveredAt && (
                    <button onClick={() => markDelivered(t.id)}
                      className="rounded-lg bg-green-700 px-4 py-2 text-sm font-bold text-white">تسليم الجهاز للزبون</button>
                  )}
                </div>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  )
}
