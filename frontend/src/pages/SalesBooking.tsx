import { useEffect, useState } from 'react'
import { api, type Service } from '../api'
import { useSession } from '../session'

export default function SalesBooking() {
  const { employee } = useSession()
  const [services, setServices] = useState<Service[]>([])

  const [name, setName] = useState('')
  const [phone, setPhone] = useState('')
  const [serviceId, setServiceId] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [message, setMessage] = useState<string | null>(null)

  useEffect(() => {
    api.getServices().then(setServices)
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    setMessage(null)
    try {
      const customer = await api.createCustomer({ name, phone })
      const booking = await api.createBooking({
        customerId: customer.id,
        serviceId: serviceId || undefined,
        transferEmployeeId: employee?.id,
      })
      setMessage(
        `تم إنشاء الحجز بنجاح. كود الزبون: ${customer.code} - كود الحجز: ${booking.code}`,
      )
      setName('')
      setPhone('')
      setServiceId('')
    } catch (err) {
      setMessage(err instanceof Error ? err.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">حجز جديد (موظف المبيعات)</h2>
      <p className="mt-1 text-slate-500">
        سجل اسم الزبون، رقم هاتفه، والخدمة التي يطلبها. الإداري راح يكمل باقي البيانات ويثبت الحجز.
      </p>

      <form
        onSubmit={handleSubmit}
        className="mt-6 grid grid-cols-1 gap-4 rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)] sm:grid-cols-2"
      >
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">اسم الزبون</label>
          <input
            required
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">رقم الهاتف</label>
          <input
            required
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>
        <div className="sm:col-span-2">
          <label className="mb-1 block text-sm font-medium text-slate-600">
            هل يوجد خدمة محددة يطلبها الزبون؟
          </label>
          <select
            value={serviceId}
            onChange={(e) => setServiceId(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          >
            <option value="">-- بدون خدمة محددة --</option>
            {services.map((s) => (
              <option key={s.id} value={s.id}>
                {s.name}
              </option>
            ))}
          </select>
        </div>
        <div className="sm:col-span-2">
          <button
            type="submit"
            disabled={submitting}
            className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-2 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30 disabled:opacity-50"
          >
            {submitting ? 'جاري الحفظ...' : 'إرسال الحجز'}
          </button>
          {message && <p className="mt-3 text-sm text-brand-700">{message}</p>}
        </div>
      </form>
    </div>
  )
}
