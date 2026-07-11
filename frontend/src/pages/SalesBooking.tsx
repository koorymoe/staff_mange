import { useEffect, useState } from 'react'
import { api, type Customer, type Service } from '../api'
import { useSession } from '../session'
import { validateCustomerName, validateCustomerPhone } from '../validation'
import LocationPicker from '../components/LocationPicker'

type BookingType = 'REGULAR' | 'MAINTENANCE'
type Urgency = 'ASAP' | 'BY_PRIORITY' | 'SPECIFIC_DATE'
type MaintenanceType = 'EXECUTION_ERROR' | 'DEVICE_ISSUE'

export default function SalesBooking() {
  const { employee } = useSession()
  const [services, setServices] = useState<Service[]>([])

  const [bookingType, setBookingType] = useState<BookingType | null>(null)
  const [name, setName] = useState('')
  const [phone, setPhone] = useState('')
  const [serviceId, setServiceId] = useState('')
  const [notes, setNotes] = useState('')
  const [addressDesc, setAddressDesc] = useState('')
  const [mapPoint, setMapPoint] = useState<{ lat: number; lng: number } | null>(null)
  const [urgency, setUrgency] = useState<Urgency | null>(null)
  const [specificDate, setSpecificDate] = useState('')
  const [maintenanceType, setMaintenanceType] = useState<MaintenanceType | null>(null)
  const [remembersCrew, setRemembersCrew] = useState(false)
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

  const buildNotesString = () => {
    const parts: string[] = []
    const typeLabel = bookingType === 'REGULAR' ? 'حجز عادي' : 'حجز صيانة'
    parts.push(`[نوع: ${typeLabel}]`)

    if (bookingType === 'REGULAR' && urgency) {
      const urgencyLabels: Record<Urgency, string> = {
        ASAP: 'أسرع وقت ممكن',
        BY_PRIORITY: 'حسب الأولوية',
        SPECIFIC_DATE: `تاريخ محدد: ${specificDate}`,
      }
      parts.push(`[الأولوية: ${urgencyLabels[urgency]}]`)
    }

    if (bookingType === 'MAINTENANCE' && maintenanceType) {
      const mtLabel = maintenanceType === 'EXECUTION_ERROR' ? 'خطأ تنفيذ' : 'مشكلة جهاز'
      let mtExtra = mtLabel
      if (maintenanceType === 'EXECUTION_ERROR') {
        mtExtra += remembersCrew ? ' - يتذكر الكادر المنفذ' : ' - لا يتذكر الكادر المنفذ'
      }
      parts.push(`[نوع الصيانة: ${mtExtra}]`)
    }

    if (notes.trim()) parts.push(notes.trim())

    return parts.join(' ')
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setMessage(null)
    setSuccess(null)
    setNameTouched(true)
    setPhoneTouched(true)

    if (!bookingType) {
      setMessage('يرجى اختيار نوع الحجز')
      return
    }

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

    if (!addressDesc.trim()) {
      setMessage('يرجى إدخال وصف الموقع')
      return
    }
    if (!mapPoint) {
      setMessage('يرجى تحديد موقع الزبون على الخريطة')
      return
    }

    if (bookingType === 'REGULAR' && !urgency) {
      setMessage('يرجى اختيار مستوى الأولوية')
      return
    }
    if (bookingType === 'REGULAR' && urgency === 'SPECIFIC_DATE' && !specificDate) {
      setMessage('يرجى تحديد التاريخ')
      return
    }
    if (bookingType === 'MAINTENANCE' && !maintenanceType) {
      setMessage('يرجى اختيار نوع الصيانة')
      return
    }

    setSubmitting(true)
    try {
      const customer = await api.createCustomer({ name, phone })
      const booking = await api.createBooking({
        customerId: customer.id,
        serviceId: serviceId || undefined,
        transferEmployeeId: employee?.id,
        notes: buildNotesString() || undefined,
        address: addressDesc.trim(),
        mapLatitude: mapPoint.lat,
        mapLongitude: mapPoint.lng,
      })
      setSuccess({ customerCode: customer.code, bookingCode: booking.code })
      setName('')
      setPhone('')
      setServiceId('')
      setNotes('')
      setAddressDesc('')
      setMapPoint(null)
      setBookingType(null)
      setUrgency(null)
      setSpecificDate('')
      setMaintenanceType(null)
      setRemembersCrew(false)
      setNameTouched(false)
      setPhoneTouched(false)
      setExistingCustomer(null)
    } catch (err) {
      setMessage(err instanceof Error ? err.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  const SectionHeader = ({ num, title }: { num: number; title: string }) => (
    <div className="flex items-center gap-3 mb-4">
      <span className="flex h-7 w-7 items-center justify-center rounded-full bg-gradient-to-l from-brand-500 to-brand-800 text-xs font-bold text-white">
        {num}
      </span>
      <div className="flex items-center gap-2">
        <span className="h-2 w-2 rounded-full bg-brand-500" />
        <h3 className="text-base font-semibold text-brand-900">{title}</h3>
      </div>
    </div>
  )

  return (
    <div dir="rtl" className="space-y-6">
      {/* Header */}
      <div className="rounded-2xl bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-5 text-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        <h2 className="text-2xl font-bold">حجز جديد</h2>
        <p className="mt-1 text-sm text-white/80">
          سجل اسم الزبون، رقم هاتفه، والخدمة التي يطلبها. الإداري راح يكمل باقي البيانات ويثبت الحجز.
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Step 1: Booking Type */}
        <div className="rounded-2xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <SectionHeader num={1} title="نوع الحجز" />
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <button
              type="button"
              onClick={() => { setBookingType('REGULAR'); setMaintenanceType(null); setRemembersCrew(false) }}
              className={`flex flex-col items-center gap-3 rounded-2xl border-2 p-6 transition-all ${
                bookingType === 'REGULAR'
                  ? 'border-blue-500 bg-blue-50 shadow-md'
                  : 'border-slate-200 bg-white hover:border-blue-300 hover:bg-blue-50/50'
              }`}
            >
              <span className="text-4xl">📅</span>
              <span className="text-lg font-bold text-slate-800">حجز عادي</span>
              <span className="text-xs text-slate-500">خدمة تركيب أو خدمة جديدة</span>
            </button>
            <button
              type="button"
              onClick={() => { setBookingType('MAINTENANCE'); setUrgency(null); setSpecificDate('') }}
              className={`flex flex-col items-center gap-3 rounded-2xl border-2 p-6 transition-all ${
                bookingType === 'MAINTENANCE'
                  ? 'border-amber-500 bg-amber-50 shadow-md'
                  : 'border-slate-200 bg-white hover:border-amber-300 hover:bg-amber-50/50'
              }`}
            >
              <span className="text-4xl">🔧</span>
              <span className="text-lg font-bold text-slate-800">حجز صيانة</span>
              <span className="text-xs text-slate-500">صيانة أو إصلاح جهاز مركب</span>
            </button>
          </div>
        </div>

        {/* Step 2: Customer Info */}
        <div className="rounded-2xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <SectionHeader num={2} title="معلومات الزبون" />
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">اسم الزبون الرباعي</label>
              <input
                required
                placeholder="مثال: محمد علي حسن جاسم"
                value={name}
                onChange={(e) => setName(e.target.value)}
                onBlur={() => setNameTouched(true)}
                className={`w-full rounded-xl border px-4 py-2.5 outline-none transition-colors focus:border-brand-500 focus:ring-1 focus:ring-brand-500/20 ${
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
                className={`w-full rounded-xl border px-4 py-2.5 outline-none transition-colors focus:border-brand-500 focus:ring-1 focus:ring-brand-500/20 ${
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
          </div>
        </div>

        {/* Step 3: Address */}
        <div className="rounded-2xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <SectionHeader num={3} title="العنوان والموقع" />
          <div className="grid grid-cols-1 gap-4">
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">وصف الموقع وأقرب نقطة دالة</label>
              <textarea
                required
                value={addressDesc}
                onChange={(e) => setAddressDesc(e.target.value)}
                placeholder="مثال: بغداد - الكرادة - قرب مطعم ..."
                rows={2}
                className="w-full rounded-xl border border-slate-300 px-4 py-2.5 outline-none transition-colors focus:border-brand-500 focus:ring-1 focus:ring-brand-500/20"
              />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">تحديد الموقع على الخريطة</label>
              <LocationPicker value={mapPoint} onChange={setMapPoint} />
            </div>
          </div>
        </div>

        {/* Step 4: Service */}
        <div className="rounded-2xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <SectionHeader num={4} title="الخدمة المطلوبة" />
          <select
            value={serviceId}
            onChange={(e) => setServiceId(e.target.value)}
            className="w-full rounded-xl border border-slate-300 px-4 py-2.5 outline-none transition-colors focus:border-brand-500 focus:ring-1 focus:ring-brand-500/20"
          >
            <option value="">-- بدون خدمة محددة --</option>
            {services.map((s) => (
              <option key={s.id} value={s.id}>
                {s.name}
              </option>
            ))}
          </select>
        </div>

        {/* Step 5: Urgency / Maintenance Type */}
        {bookingType === 'REGULAR' && (
          <div className="rounded-2xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <SectionHeader num={5} title="مستوى الأولوية" />
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
              {([
                { value: 'ASAP' as Urgency, label: 'أسرع وقت ممكن', desc: 'يحتاج الخدمة بشكل عاجل', color: 'red' },
                { value: 'BY_PRIORITY' as Urgency, label: 'حسب الأولوية', desc: 'حسب ترتيب الطلبات', color: 'blue' },
                { value: 'SPECIFIC_DATE' as Urgency, label: 'تاريخ محدد', desc: 'يريد تاريخ معين', color: 'emerald' },
              ]).map((opt) => {
                const selected = urgency === opt.value
                const borderColor = selected
                  ? opt.color === 'red' ? 'border-red-500 bg-red-50'
                  : opt.color === 'blue' ? 'border-blue-500 bg-blue-50'
                  : 'border-emerald-500 bg-emerald-50'
                  : 'border-slate-200 bg-white hover:border-slate-300'
                const accentDot = opt.color === 'red' ? 'bg-red-500' : opt.color === 'blue' ? 'bg-blue-500' : 'bg-emerald-500'
                return (
                  <button
                    key={opt.value}
                    type="button"
                    onClick={() => setUrgency(opt.value)}
                    className={`flex items-start gap-3 rounded-2xl border-2 p-4 text-right transition-all ${borderColor}`}
                  >
                    <span className={`mt-1 h-3 w-3 shrink-0 rounded-full ${accentDot}`} />
                    <div>
                      <span className="block text-sm font-bold text-slate-800">{opt.label}</span>
                      <span className="block text-xs text-slate-500">{opt.desc}</span>
                    </div>
                  </button>
                )
              })}
            </div>
            {urgency === 'SPECIFIC_DATE' && (
              <div className="mt-4">
                <label className="mb-1 block text-sm font-medium text-slate-600">حدد التاريخ</label>
                <input
                  type="date"
                  value={specificDate}
                  onChange={(e) => setSpecificDate(e.target.value)}
                  className="w-full rounded-xl border border-slate-300 px-4 py-2.5 outline-none transition-colors focus:border-brand-500 focus:ring-1 focus:ring-brand-500/20 sm:w-auto"
                />
              </div>
            )}
          </div>
        )}

        {bookingType === 'MAINTENANCE' && (
          <div className="rounded-2xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <SectionHeader num={5} title="نوع الصيانة" />
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              <button
                type="button"
                onClick={() => setMaintenanceType('EXECUTION_ERROR')}
                className={`flex items-start gap-3 rounded-2xl border-2 p-5 text-right transition-all ${
                  maintenanceType === 'EXECUTION_ERROR'
                    ? 'border-orange-500 bg-orange-50 shadow-md'
                    : 'border-slate-200 bg-white hover:border-orange-300'
                }`}
              >
                <span className="mt-0.5 h-3 w-3 shrink-0 rounded-full bg-orange-500" />
                <div>
                  <span className="block text-sm font-bold text-slate-800">خطأ تنفيذ</span>
                  <span className="block text-xs text-slate-500">مشكلة ناتجة عن التركيب</span>
                </div>
              </button>
              <button
                type="button"
                onClick={() => { setMaintenanceType('DEVICE_ISSUE'); setRemembersCrew(false) }}
                className={`flex items-start gap-3 rounded-2xl border-2 p-5 text-right transition-all ${
                  maintenanceType === 'DEVICE_ISSUE'
                    ? 'border-purple-500 bg-purple-50 shadow-md'
                    : 'border-slate-200 bg-white hover:border-purple-300'
                }`}
              >
                <span className="mt-0.5 h-3 w-3 shrink-0 rounded-full bg-purple-500" />
                <div>
                  <span className="block text-sm font-bold text-slate-800">مشكلة جهاز</span>
                  <span className="block text-xs text-slate-500">عطل في الجهاز نفسه</span>
                </div>
              </button>
            </div>
            {maintenanceType === 'EXECUTION_ERROR' && (
              <div className="mt-4 rounded-xl border border-orange-200 bg-orange-50/50 p-4">
                <p className="mb-3 text-sm font-medium text-slate-700">هل تتذكر الكادر المنفذ؟</p>
                <div className="flex gap-3">
                  <button
                    type="button"
                    onClick={() => setRemembersCrew(true)}
                    className={`rounded-xl px-5 py-2 text-sm font-medium transition-all ${
                      remembersCrew
                        ? 'bg-brand-600 text-white shadow-md'
                        : 'bg-white text-slate-600 border border-slate-300 hover:border-brand-400'
                    }`}
                  >
                    نعم
                  </button>
                  <button
                    type="button"
                    onClick={() => setRemembersCrew(false)}
                    className={`rounded-xl px-5 py-2 text-sm font-medium transition-all ${
                      !remembersCrew
                        ? 'bg-brand-600 text-white shadow-md'
                        : 'bg-white text-slate-600 border border-slate-300 hover:border-brand-400'
                    }`}
                  >
                    لا
                  </button>
                </div>
              </div>
            )}
          </div>
        )}

        {/* Step 6: Notes */}
        <div className="rounded-2xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <SectionHeader num={6} title="ملاحظات إضافية" />
          <textarea
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            placeholder="أي تفاصيل إضافية يطلبها الزبون..."
            rows={3}
            className="w-full rounded-xl border border-slate-300 px-4 py-2.5 outline-none transition-colors focus:border-brand-500 focus:ring-1 focus:ring-brand-500/20"
          />
        </div>

        {/* Submit */}
        <div className="rounded-2xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <button
            type="submit"
            disabled={submitting}
            className="w-full rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 text-lg font-bold text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30 disabled:opacity-50 sm:w-auto"
          >
            {submitting ? 'جاري الحفظ...' : 'إرسال الحجز'}
          </button>
          {message && (
            <div className="mt-4 rounded-xl border border-red-200 bg-red-50 p-3">
              <p className="text-sm text-red-700">{message}</p>
            </div>
          )}
          {success && (
            <div className="mt-4 rounded-xl border border-emerald-200 bg-emerald-50 p-5 text-emerald-800">
              <p className="text-lg font-bold">تم إنشاء الحجز بنجاح ✅</p>
              <div className="mt-3 grid grid-cols-2 gap-3 text-sm">
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
