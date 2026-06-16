import { useEffect, useState } from 'react'
import { api, type Customer, type Booking } from '../api'
import { validateCustomerName, validateCustomerPhone } from '../validation'

const statusLabels: Record<string, string> = {
  PENDING: 'بانتظار التثبيت',
  CONFIRMED: 'مثبت',
  IN_PROGRESS: 'جاري التنفيذ',
  COMPLETED: 'منجز',
  CANCELLED: 'ملغى',
}

const statusColors: Record<string, string> = {
  PENDING: 'bg-amber-100 text-amber-700',
  CONFIRMED: 'bg-blue-100 text-blue-700',
  IN_PROGRESS: 'bg-orange-100 text-orange-700',
  COMPLETED: 'bg-emerald-100 text-emerald-700',
  CANCELLED: 'bg-red-100 text-red-700',
}

export default function Customers() {
  const [customers, setCustomers] = useState<Customer[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [history, setHistory] = useState<Booking[]>([])
  const [historyLoading, setHistoryLoading] = useState(false)

  const [name, setName] = useState('')
  const [phone, setPhone] = useState('')
  const [location, setLocation] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [message, setMessage] = useState<string | null>(null)

  const load = () => {
    setLoading(true)
    api
      .getCustomers()
      .then(setCustomers)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  useEffect(() => {
    if (!selectedId) {
      setHistory([])
      return
    }
    setHistoryLoading(true)
    api
      .getBookings({ customerId: selectedId })
      .then(setHistory)
      .finally(() => setHistoryLoading(false))
  }, [selectedId])

  const selectedCustomer = customers.find((c) => c.id === selectedId) || null

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setMessage(null)

    const nameError = validateCustomerName(name)
    if (nameError) {
      setMessage(nameError)
      return
    }
    const phoneError = validateCustomerPhone(phone)
    if (phoneError) {
      setMessage(phoneError)
      return
    }

    setSubmitting(true)
    try {
      const result = await api.createCustomer({ name, phone, location: location || undefined })
      setMessage(
        result.existed
          ? `الزبون موجود مسبقاً برقم ${result.code}`
          : `تم إنشاء زبون جديد برقم ${result.code}`,
      )
      setName('')
      setPhone('')
      setLocation('')
      load()
    } catch (e) {
      setMessage(e instanceof Error ? e.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">الزبائن</h2>
      <p className="mt-1 text-slate-500">
        كل زبون يحصل على كود ثابت يبقى مرتبطاً به في كل تعاملاته مع الشركة.
      </p>

      <form
        onSubmit={handleSubmit}
        className="mt-6 grid grid-cols-1 gap-4 rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)] sm:grid-cols-3"
      >
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">
            اسم الزبون / المؤسسة (الاسم الرباعي)
          </label>
          <input
            required
            placeholder="مثال: محمد علي حسن جاسم"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">رقم الهاتف (11 رقم)</label>
          <input
            required
            placeholder="07XXXXXXXXX"
            maxLength={11}
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">الموقع</label>
          <input
            value={location}
            onChange={(e) => setLocation(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>
        <div className="sm:col-span-3">
          <button
            type="submit"
            disabled={submitting}
            className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-2 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30 disabled:opacity-50"
          >
            {submitting ? 'جاري الحفظ...' : 'حفظ الزبون'}
          </button>
          {message && <span className="mr-4 text-sm text-brand-700">{message}</span>}
        </div>
      </form>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && (
        <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">
          تعذر الاتصال بالخادم: {error}
        </p>
      )}

      {!loading && !error && (
        <div className="mt-6 flex flex-col gap-6">
          <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <table className="w-full text-right">
              <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                <tr>
                  <th className="px-4 py-3 text-sm font-semibold">الكود</th>
                  <th className="px-4 py-3 text-sm font-semibold">الاسم</th>
                  <th className="px-4 py-3 text-sm font-semibold">الهاتف</th>
                  <th className="px-4 py-3 text-sm font-semibold">الموقع</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {customers.map((c) => (
                  <tr
                    key={c.id}
                    onClick={() => setSelectedId(c.id)}
                    className={`cursor-pointer transition-colors ${
                      selectedId === c.id ? 'bg-brand-50' : 'hover:bg-slate-50'
                    }`}
                  >
                    <td className="px-4 py-3 font-mono text-sm font-semibold text-brand-600">
                      {c.code}
                    </td>
                    <td className="px-4 py-3">{c.name}</td>
                    <td className="px-4 py-3 text-slate-500">{c.phone}</td>
                    <td className="px-4 py-3 text-slate-500">{c.location || '-'}</td>
                  </tr>
                ))}
                {customers.length === 0 && (
                  <tr>
                    <td colSpan={4} className="px-4 py-6 text-center text-slate-400">
                      لا يوجد زبائن بعد
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          <div className="rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            {!selectedCustomer && (
              <p className="text-slate-400">اختر زبوناً من القائمة لعرض بياناته وسجل طلباته.</p>
            )}
            {selectedCustomer && (
              <div>
                <h3 className="text-lg font-bold text-brand-800">{selectedCustomer.name}</h3>
                <div className="mt-2 grid grid-cols-1 gap-2 text-sm sm:grid-cols-3">
                  <p>
                    <span className="text-slate-500">الكود: </span>
                    <span className="font-mono font-semibold text-brand-600">
                      {selectedCustomer.code}
                    </span>
                  </p>
                  <p>
                    <span className="text-slate-500">الهاتف: </span>
                    {selectedCustomer.phone}
                  </p>
                  <p>
                    <span className="text-slate-500">الموقع: </span>
                    {selectedCustomer.location || '-'}
                  </p>
                </div>

                <h4 className="mt-5 font-bold text-brand-800">طلبات وحجوزات الزبون</h4>
                {historyLoading && <p className="mt-2 text-slate-400">جاري التحميل...</p>}
                {!historyLoading && (
                  <div className="mt-3 divide-y divide-slate-100">
                    {history.map((b) => (
                      <div key={b.id} className="py-3 text-sm">
                        <div className="flex items-center justify-between">
                          <span className="font-mono font-semibold text-brand-600">{b.code}</span>
                          <span
                            className={`rounded-full px-2 py-1 text-xs font-bold ${statusColors[b.status]}`}
                          >
                            {statusLabels[b.status] || b.status}
                          </span>
                        </div>
                        <div className="mt-1 grid grid-cols-1 gap-1 text-slate-600 sm:grid-cols-2">
                          <p>
                            <span className="text-slate-400">الخدمة المطلوبة: </span>
                            {b.service?.name || '-'}
                          </p>
                          <p>
                            <span className="text-slate-400">الموظف الذي سجل الطلب: </span>
                            {b.transferEmployee?.name || '-'}
                          </p>
                          <p>
                            <span className="text-slate-400">السيارة المخصصة: </span>
                            {b.assignedVehicle || '-'}
                          </p>
                          <p>
                            <span className="text-slate-400">التاريخ: </span>
                            {new Date(b.createdAt).toLocaleDateString('ar-IQ')}
                          </p>
                        </div>
                      </div>
                    ))}
                    {history.length === 0 && (
                      <p className="py-4 text-center text-slate-400">لا توجد طلبات لهذا الزبون</p>
                    )}
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
