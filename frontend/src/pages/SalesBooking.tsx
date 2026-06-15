import { useEffect, useState } from 'react'
import { api, type Customer, type Service } from '../api'
import { useSession } from '../session'
import { validateCustomerName, validateCustomerPhone } from '../validation'

export default function SalesBooking() {
  const { employee } = useSession()
  const [services, setServices] = useState<Service[]>([])

  const [name, setName] = useState('')
  const [phone, setPhone] = useState('')
  const [serviceId, setServiceId] = useState('')
  const [notes, setNotes] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [message, setMessage] = useState<string | null>(null)
  const [success, setSuccess] = useState<{ customerCode: string; bookingCode: string } | null>(null)
  const [nameTouched, setNameTouched] = useState(false)
  const [phoneTouched, setPhoneTouched] = useState(false)
  const [existingCustomer, setExistingCustomer] = useState<Customer | null>(null)

  useEffect(() => {
    api.getServices().then(setServices)
  }, [])

  useEffect(() => {
    if (!/^\d{11}$/.test(phone.trim())) {
      setExistingCustomer(null)
      return
    }
    let active = true
    api.lookupCustomer(phone.trim()).then((c) => {
      if (active) setExistingCustomer(c)
    })
    return () => {
      active = false
    }
  }, [phone])

  const nameError = nameTouched ? validateCustomerName(name) : null
  const phoneError = phoneTouched ? validateCustomerPhone(phone) : null

  const handlePhoneChange = (value: string) => {
    const digitsOnly = value.replace(/\D/g, '').slice(0, 11)
    setPhone(digitsOnly)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setMessage(null)
    setSuccess(null)
    setNameTouched(true)
    setPhoneTouched(true)

    const nameValidationError = validateCustomerName(name)
    if (nameValidationError) {
      setMessage(nameValidationError)
      return
    }
    const phoneValidationError = validateCustomerPhone(phone)
    if (phoneValidationError) {
      setMessage(phoneValidationError)
      return
    }

    setSubmitting(true)
    try {
      const customer = await api.createCustomer({ name, phone })
      const booking = await api.createBooking({
        customerId: customer.id,
        serviceId: serviceId || undefined,
        transferEmployeeId: employee?.id,
        notes: notes || undefined,
      })
      setSuccess({ customerCode: customer.code, bookingCode: booking.code })
      setName('')
      setPhone('')
      setServiceId('')
      setNotes('')
      setNameTouched(false)
      setPhoneTouched(false)
      setExistingCustomer(null)
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
          <label className="mb-1 block text-sm font-medium text-slate-600">اسم الزبون (الاسم الرباعي)</label>
          <input
            required
            placeholder="مثال: محمد علي حسن جاسم"
            value={name}
            onChange={(e) => setName(e.target.value)}
            onBlur={() => setNameTouched(true)}
            className={`w-full rounded-lg border px-3 py-2 outline-none focus:border-brand-500 ${
              nameError ? 'border-red-400' : 'border-slate-300'
            }`}
          />
          {nameError && <p className="mt-1 text-xs text-red-600">{nameError}</p>}
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">رقم الهاتف (11 رقم)</label>
          <input
            required
            placeholder="07XXXXXXXXX"
            inputMode="numeric"
            maxLength={11}
            value={phone}
            onChange={(e) => handlePhoneChange(e.target.value)}
            onBlur={() => setPhoneTouched(true)}
            className={`w-full rounded-lg border px-3 py-2 outline-none focus:border-brand-500 ${
              phoneError ? 'border-red-400' : 'border-slate-300'
            }`}
          />
          {phoneError && <p className="mt-1 text-xs text-red-600">{phoneError}</p>}
          {!phoneError && existingCustomer && (
            <p className="mt-1 text-xs text-brand-700">
              هذا الزبون مسجل مسبقاً: {existingCustomer.name} (كود: {existingCustomer.code})
            </p>
          )}
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
          <label className="mb-1 block text-sm font-medium text-slate-600">
            ملاحظات (اختياري)
          </label>
          <textarea
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            placeholder="أي تفاصيل إضافية يطلبها الزبون..."
            rows={2}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>
        <div className="sm:col-span-2">
          <button
            type="submit"
            disabled={submitting}
            className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-2 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30 disabled:opacity-50"
          >
            {submitting ? 'جاري الحفظ...' : 'إرسال الحجز'}
          </button>
          {message && <p className="mt-3 text-sm text-red-600">{message}</p>}
          {success && (
            <div className="mt-4 rounded-lg border border-emerald-200 bg-emerald-50 p-4 text-emerald-800">
              <p className="font-bold">تم إنشاء الحجز بنجاح ✅</p>
              <div className="mt-2 grid grid-cols-2 gap-2 text-sm">
                <p>
                  <span className="text-emerald-600">كود الزبون: </span>
                  <span className="font-mono font-bold">{success.customerCode}</span>
                </p>
                <p>
                  <span className="text-emerald-600">كود الحجز: </span>
                  <span className="font-mono font-bold">{success.bookingCode}</span>
                </p>
              </div>
            </div>
          )}
        </div>
      </form>
    </div>
  )
}
