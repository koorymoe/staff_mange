import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { api, type Customer, type Service, type SolarSystem } from '../api'
import { useSession } from '../session'
import MultiSelect from '../components/MultiSelect'
import { validateCustomerName, validateCustomerPhone } from '../validation'
import { LocationPicker } from '../components/MapLazy'

type BookingType = 'REGULAR' | 'MAINTENANCE' | 'INTERNAL' | 'SOLAR'
type Urgency = 'ASAP' | 'BY_PRIORITY' | 'SPECIFIC_DATE'
type MaintenanceType = 'EXECUTION_ERROR' | 'DEVICE_ISSUE' | 'UPKEEP'

// يوزّع اسم كامل محفوظ بحقل واحد (من زبون موجود مسبقاً) على الخانات الأربع —
// الخانة الأخيرة تاخذ أي كلمات زايدة عن 4 حتى ما تضيع لو الاسم أطول من المتوقع.
function splitFullName(fullName: string): [string, string, string, string] {
  const parts = fullName.trim().split(/\s+/).filter(Boolean)
  return [
    parts[0] || '',
    parts[1] || '',
    parts[2] || '',
    parts.slice(3).join(' '),
  ]
}

function SectionHeader({ num, title }: { num: number; title: string }) {
  return (
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
}

export default function SalesBooking() {
  const { employee, permissions } = useSession()
  // حجز داخل الشركة للإداري فما فوق — نفس قيد السيرفر بالضبط
  const canInternal = employee?.role === 'ADMIN' || employee?.role === 'OWNER' ||
    employee?.role === 'HR_COORDINATOR' || employee?.role === 'MONITOR' ||
    permissions.includes('booking_internal')
  const [services, setServices] = useState<Service[]>([])

  const [bookingType, setBookingType] = useState<BookingType | null>(null)
  // ═══ حجز طاقة شمسية ═══
  // المنظومة اختيارية: الزبون أحياناً يريد منظومة ولسه ما قرر أي وحدة،
  // فالمبيعات يسجّل الحجز والمنسّق يحدد المنظومة بعد المعاينة.
  const [solarSystems, setSolarSystems] = useState<SolarSystem[]>([])
  const [solarSystemId, setSolarSystemId] = useState('')
  const [solarMonthlyKwh, setSolarMonthlyKwh] = useState('')
  const [firstName, setFirstName] = useState('')
  const [fatherName, setFatherName] = useState('')
  const [grandfatherName, setGrandfatherName] = useState('')
  const [familyName, setFamilyName] = useState('')
  const name = [firstName, fatherName, grandfatherName, familyName].map((p) => p.trim()).filter(Boolean).join(' ')
  const [phone, setPhone] = useState('')
  // تصحيح بيانات زبون قديم: أسماء النظام القديم غريبة أو ناقصة، فالموظف
  // لازم يقدر يصلّحها من نفس شاشة الحجز — وتنحفظ على سجل الزبون مو
  // على الحجز بس، وإلا الحجز الجديد يطلع بالاسم القديم الغلط.
  const [fixCustomer, setFixCustomer] = useState(false)
  // حجز داخل الشركة: الشغل لموظف من موظفينا، فنسأل عن معلوماته
  // وموافقة مسؤوله بدل معلومات الزبون الخارجي.
  const [intName, setIntName] = useState('')
  const [intPhone, setIntPhone] = useState('')
  const [intDept, setIntDept] = useState('')
  const [intApproved, setIntApproved] = useState(false)
  // خدمات متعددة: الزبون ممكن يطلب أكثر من منظومة بنفس الحجز
  // (مثلاً منظومة صوت + كاميرات). أول خدمة تنعتبر الرئيسية.
  const [serviceIds, setServiceIds] = useState<string[]>([])
  // تفاصيل الأجهزة — تنطلب بس للخدمات المؤشّرة، والسيرفر يفرضها هم
  const [deviceCount, setDeviceCount] = useState('')
  const [gpsVehicleType, setGpsVehicleType] = useState('')
  // الخدمة المختارة تطلب تفاصيل أجهزة؟ العلم يجي من السيرفر مو مكتوب
  // بالكود — صاحب العمل يأشّر أي خدمة ثانية بلا نشر نسخة جديدة.
  const needsDeviceInfo = services.some((s) => serviceIds.includes(s.id) && s.requiresDeviceInfo)
  // رابط الموقع (كوكل ماب) — بديل عن التحديد على الخريطة، نفس فكرة الموردين
  const [locationUrl, setLocationUrl] = useState('')
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
  const [usedSavedLocation, setUsedSavedLocation] = useState(false)

  const [searchParams] = useSearchParams()

  useEffect(() => {
    api.getServices().then(setServices)
  }, [])

  // إذا وصلنا من "تحويل شكوى/متابعة جودة لحجز جديد" نعبّي بيانات الزبون تلقائياً
  useEffect(() => {
    const customerId = searchParams.get('customerId')
    if (!customerId) return
    api.getCustomers().then((customers) => {
      const c = customers.find((cust) => cust.id === customerId)
      if (c) {
        setPhone(c.phone)
        const [f, fa, gf, fam] = splitFullName(c.name)
        setFirstName(f)
        setFatherName(fa)
        setGrandfatherName(gf)
        setFamilyName(fam)
      }
    })
  }, [searchParams])

  useEffect(() => {
    // Guard-clause reset when the typed phone isn't a valid lookup key yet.
    if (!/^\d{11}$/.test(phone.trim())) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
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

  // إذا الزبون موجود مسبقاً وعنده موقع محفوظ من حجز سابق، نحمّله تلقائياً
  // (بس إذا الموظف لسه ما بدأ يعبي الموقع بنفسه) حتى ما يعيد نفس الشغل، مع خيار تغييره.
  useEffect(() => {
    if (!existingCustomer) return
    if (addressDesc.trim() || mapPoint) return
    // Pre-filling the address form from the looked-up customer's saved location is a
    // derived-state sync from a prop (existingCustomer), not a fetch.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    if (existingCustomer.location) setAddressDesc(existingCustomer.location)
    if (existingCustomer.mapLatitude != null && existingCustomer.mapLongitude != null) {
      setMapPoint({ lat: existingCustomer.mapLatitude, lng: existingCustomer.mapLongitude })
    }
    if (existingCustomer.location || existingCustomer.mapLatitude != null) setUsedSavedLocation(true)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [existingCustomer])

  const nameError = nameTouched ? validateCustomerName(name) : null
  const phoneError = phoneTouched ? validateCustomerPhone(phone) : null

  const handlePhoneChange = (value: string) => {
    const digitsOnly = value.replace(/\D/g, '').slice(0, 11)
    setPhone(digitsOnly)
  }

  // نجيب الكتالوك بس لمن ينختار النوع — ما نحمّله لكل موظف مبيعات
  // يفتح الصفحة وهو أصلاً راح يسوي حجز عادي.
  useEffect(() => {
    if (bookingType !== 'SOLAR' || solarSystems.length > 0) return
    api.getSolarSystems().then(setSolarSystems).catch(() => {})
  }, [bookingType, solarSystems.length])

  const buildNotesString = () => {
    const parts: string[] = []
    const typeLabel = bookingType === 'REGULAR' ? 'حجز عادي'
    : bookingType === 'SOLAR' ? 'حجز طاقة شمسية'
      : bookingType === 'INTERNAL' ? 'حجز داخل الشركة' : 'حجز صيانة'
    parts.push(`[نوع: ${typeLabel}]`)

    if (bookingType === 'REGULAR' && urgency) {
      const urgencyLabels: Record<Urgency, string> = {
        ASAP: 'أسرع وقت ممكن',
        BY_PRIORITY: 'حسب الأولوية',
        SPECIFIC_DATE: `تاريخ محدد: ${specificDate}`,
      }
      parts.push(`[الأولوية: ${urgencyLabels[urgency]}]`)
    }

    if (bookingType === 'SOLAR') {
      const sys = solarSystems.find((x) => x.id === solarSystemId)
      if (sys) parts.push(`[المنظومة: ${sys.brand} ${sys.capacity} — ${sys.model}]`)
      if (solarMonthlyKwh) parts.push(`[استهلاك الزبون الشهري: ${solarMonthlyKwh} كيلو واط/ساعة]`)
    }

    if (bookingType === 'MAINTENANCE' && maintenanceType) {
      const maintenanceTypeLabels: Record<MaintenanceType, string> = {
        EXECUTION_ERROR: 'خطأ تنفيذ',
        DEVICE_ISSUE: 'مشكلة جهاز',
        UPKEEP: 'إدامة',
      }
      const mtLabel = maintenanceTypeLabels[maintenanceType]
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

    // الحجز الداخلي ما بيه زبون — بياناته اسم الموظف وقسمه، وينفحصن
    // تحت. فحص الاسم الرباعي هنا جان يوقف الحجز الداخلي بلا سبب.
    if (bookingType !== 'INTERNAL') {
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
    if (bookingType === 'INTERNAL') {
      if (!intName.trim() || !intDept.trim()) {
        setMessage('اكتب اسم الموظف الثلاثي والقسم الي يشتغل بيه')
        return
      }
      if (intName.trim().split(/\s+/).length < 3) {
        setMessage('اسم الموظف لازم يكون ثلاثي (٣ أسماء على الأقل)')
        return
      }
      if (!/^\d{11}$/.test(intPhone.trim())) {
        setMessage('رقم هاتف الموظف لازم يكون ١١ رقم')
        return
      }
    }
    if (bookingType === 'MAINTENANCE' && !maintenanceType) {
      setMessage('يرجى اختيار نوع الصيانة')
      return
    }

    setSubmitting(true)
    try {
      // الحجز الداخلي ما بيه زبون، بس السجل لازمه جهة يتعلّق بيها
      // (وإلا ما ينفتح حجز أصلاً) — فنستعملها بيانات الموظف الطالب
      // نفسه: اسمه الثلاثي ورقمه. رقمه هو المفتاح، فلو حجز مرة ثانية
      // ينربط بنفس السجل ويطلع تاريخه كله سوه.
      const customer =
        bookingType === 'INTERNAL'
          ? await api.createCustomer({ name: intName.trim(), phone: intPhone.trim() })
          : await api.createCustomer({ name, phone })
      // الزبون القديم: createCustomer يرجّعه بالاسم القديم ويتجاهل الي
      // كتبناه. فلو الموظف اختار يصحّح، نحدّث سجله فعلياً — الاسم
      // والعنوان والموقع ورابط الخريطة — حتى الحجز والي بعده يطلعون
      // بالاسم الصحيح.
      if (fixCustomer && existingCustomer) {
        await api.updateCustomer(customer.id, {
          name,
          phone,
          location: addressDesc.trim() || null,
          mapLatitude: mapPoint.lat,
          mapLongitude: mapPoint.lng,
          locationUrl: locationUrl.trim() || null,
        })
      }
      const booking = await api.createBooking({
        customerId: customer.id,
        serviceId: serviceIds[0] || undefined,
        serviceIds: serviceIds.length ? serviceIds : undefined,
        // ⚠️ Number('') = صفر مو فاضي — بلا الفحص يمشي صفر للسيرفر
        // وينرفض برسالة «عدد الأجهزة مطلوب» والموظف كاتبه فعلاً.
        deviceCount: needsDeviceInfo && deviceCount ? Number(deviceCount) : undefined,
        vehicleType: needsDeviceInfo && gpsVehicleType.trim() ? gpsVehicleType.trim() : undefined,
        locationUrl: locationUrl.trim() || undefined,
        transferEmployeeId: employee?.id,
        notes: buildNotesString() || undefined,
        address: addressDesc.trim(),
        mapLatitude: mapPoint.lat,
        mapLongitude: mapPoint.lng,
        bookingType: bookingType ?? undefined,
        // المنظومة والاستهلاك — السعر المقدّر ينحسب بالسيرفر من الكتالوك
        solarSystemId: bookingType === 'SOLAR' && solarSystemId ? solarSystemId : undefined,
        solarMonthlyKwh: bookingType === 'SOLAR' && solarMonthlyKwh ? Number(solarMonthlyKwh) : undefined,
        internalEmployeeName: bookingType === 'INTERNAL' ? intName.trim() : undefined,
        internalEmployeePhone: bookingType === 'INTERNAL' ? intPhone.trim() : undefined,
        internalDepartment: bookingType === 'INTERNAL' ? intDept.trim() : undefined,
        internalApproved: bookingType === 'INTERNAL' ? intApproved : undefined,
      })
      setSuccess({ customerCode: customer.code, bookingCode: booking.code })
      setFirstName('')
      setFatherName('')
      setGrandfatherName('')
      setFamilyName('')
      setPhone('')
      setServiceIds([])
      setLocationUrl('')
      setNotes('')
      setAddressDesc('')
      setMapPoint(null)
      setUsedSavedLocation(false)
      setFixCustomer(false)
      setIntName(''); setIntPhone(''); setIntDept(''); setIntApproved(false)
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
            <button
              type="button"
              onClick={() => { setBookingType('SOLAR'); setMaintenanceType(null); setRemembersCrew(false) }}
              className={`flex flex-col items-center gap-3 rounded-2xl border-2 p-6 transition-all ${
                bookingType === 'SOLAR'
                  ? 'border-amber-500 bg-amber-50 shadow-md'
                  : 'border-slate-200 bg-white hover:border-amber-300 hover:bg-amber-50/50'
              }`}
            >
              <span className="text-4xl">☀️</span>
              <span className="text-lg font-bold text-slate-800">حجز طاقة شمسية</span>
              <span className="text-xs text-slate-500">زبون يريد منظومة — السعر يجي من الكتالوك تلقائياً</span>
            </button>
            {canInternal && (
              <button
                type="button"
                onClick={() => { setBookingType('INTERNAL'); setUrgency(null); setSpecificDate(''); setMaintenanceType(null) }}
                className={`flex flex-col items-center gap-3 rounded-2xl border-2 p-6 transition-all ${
                  bookingType === 'INTERNAL'
                    ? 'border-emerald-500 bg-emerald-50 shadow-md'
                    : 'border-slate-200 bg-white hover:border-emerald-300 hover:bg-emerald-50/50'
                }`}
              >
                <span className="text-4xl">🏢</span>
                <span className="text-lg font-bold text-slate-800">حجز داخل الشركة</span>
                <span className="text-xs text-slate-500">شغل لموظف من موظفينا — إحصائياته تروح للشغل الداخلي</span>
              </button>
            )}
          </div>
        </div>

        {/* Step 2: Customer Info — حجز داخل الشركة ماكو بيه زبون أصلاً،
            الشغل لموظف من موظفينا، فتنشال الخانة كاملة ويحل محلها
            كارت «معلومات الموظف الطالب» تحت. */}
        {bookingType !== 'INTERNAL' && (
        <div className="rounded-2xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <SectionHeader num={2} title="معلومات الزبون" />
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div className="sm:col-span-2">
              <label className="mb-1 block text-sm font-medium text-slate-600">اسم الزبون الرباعي</label>
              <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
                <input
                  required
                  placeholder="الاسم"
                  value={firstName}
                  onChange={(e) => setFirstName(e.target.value)}
                  onBlur={() => setNameTouched(true)}
                  className={`w-full rounded-xl border px-4 py-2.5 outline-none transition-colors focus:border-brand-500 focus:ring-1 focus:ring-brand-500/20 ${
                    nameError ? 'border-red-400' : 'border-slate-300'
                  }`}
                />
                <input
                  required
                  placeholder="اسم الأب"
                  value={fatherName}
                  onChange={(e) => setFatherName(e.target.value)}
                  onBlur={() => setNameTouched(true)}
                  className={`w-full rounded-xl border px-4 py-2.5 outline-none transition-colors focus:border-brand-500 focus:ring-1 focus:ring-brand-500/20 ${
                    nameError ? 'border-red-400' : 'border-slate-300'
                  }`}
                />
                <input
                  required
                  placeholder="اسم الجد"
                  value={grandfatherName}
                  onChange={(e) => setGrandfatherName(e.target.value)}
                  onBlur={() => setNameTouched(true)}
                  className={`w-full rounded-xl border px-4 py-2.5 outline-none transition-colors focus:border-brand-500 focus:ring-1 focus:ring-brand-500/20 ${
                    nameError ? 'border-red-400' : 'border-slate-300'
                  }`}
                />
                <input
                  required
                  placeholder="اللقب"
                  value={familyName}
                  onChange={(e) => setFamilyName(e.target.value)}
                  onBlur={() => setNameTouched(true)}
                  className={`w-full rounded-xl border px-4 py-2.5 outline-none transition-colors focus:border-brand-500 focus:ring-1 focus:ring-brand-500/20 ${
                    nameError ? 'border-red-400' : 'border-slate-300'
                  }`}
                />
              </div>
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
                <div className="mt-1.5 flex flex-wrap items-center gap-2">
                  <span className="rounded-full bg-amber-100 px-3 py-1 text-xs font-bold text-amber-800">
                    زبون قديم
                    {existingCustomer.previousBookingsCount != null &&
                      ` — له ${existingCustomer.previousBookingsCount} حجز سابق`}
                  </span>
                  <span className="text-xs text-brand-700">
                    {existingCustomer.name} (كود: {existingCustomer.code})
                  </span>
                  {!fixCustomer ? (
                    <button
                      type="button"
                      onClick={() => {
                        const [f, fa, gf, fam] = splitFullName(existingCustomer.name)
                        setFirstName(f); setFatherName(fa); setGrandfatherName(gf); setFamilyName(fam)
                        setFixCustomer(true)
                      }}
                      className="rounded-lg bg-brand-50 px-3 py-1 text-xs font-bold text-brand-700 hover:bg-brand-100"
                    >
                      ✏️ صحّح بياناته
                    </button>
                  ) : (
                    <span className="rounded-full bg-emerald-100 px-3 py-1 text-xs font-bold text-emerald-800">
                      راح تنحفظ البيانات الجديدة على سجل الزبون
                    </span>
                  )}
                </div>
              )}
              {!phoneError && !existingCustomer && /^\d{11}$/.test(phone.trim()) && (
                <span className="mt-1.5 inline-block rounded-full bg-green-100 px-3 py-1 text-xs font-bold text-green-800">
                  زبون جديد
                </span>
              )}
            </div>
          </div>
        </div>
        )}

        {/* حجز داخل الشركة: هذي هي «معلومات الزبون» مالته — الشغل لموظف
            من موظفينا، فنسأل عن اسمه الثلاثي بخانة وحدة وقسمه، مو عن
            اسم زبون رباعي. */}
        {bookingType === 'INTERNAL' && (
          <div className="rounded-2xl border border-emerald-200 bg-emerald-50/40 p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <SectionHeader num={2} title="🏢 معلومات الموظف الطالب" />
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div className="sm:col-span-2">
                <label className="mb-1 block text-sm font-medium text-slate-600">اسم الموظف الثلاثي *</label>
                <input value={intName} onChange={(e) => setIntName(e.target.value)}
                  placeholder="مثال: أحمد علي حسين"
                  className="w-full rounded-xl border border-slate-300 px-4 py-2.5 outline-none focus:border-emerald-500" />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium text-slate-600">القسم الي يشتغل بيه *</label>
                <input value={intDept} onChange={(e) => setIntDept(e.target.value)}
                  placeholder="مثال: وحدة التقنيين"
                  className="w-full rounded-xl border border-slate-300 px-4 py-2.5 outline-none focus:border-emerald-500" />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium text-slate-600">رقم هاتفه *</label>
                <input value={intPhone} onChange={(e) => setIntPhone(e.target.value.replace(/\D/g, '').slice(0, 11))}
                  dir="ltr" inputMode="numeric" placeholder="07XXXXXXXXX"
                  className="w-full rounded-xl border border-slate-300 px-4 py-2.5 outline-none focus:border-emerald-500" />
                <p className="mt-1 text-[11px] text-slate-400">
                  رقم اتصال واحد لازم — الفني يحتاجه يوصله يوم الشغل.
                </p>
              </div>
              <div className="sm:col-span-2">
                <label className="flex items-center gap-2 rounded-xl border border-slate-300 bg-white px-4 py-2.5 text-sm">
                  <input type="checkbox" checked={intApproved} onChange={(e) => setIntApproved(e.target.checked)} />
                  تم الحصول على موافقة مسؤوله
                </label>
              </div>
            </div>
          </div>
        )}

        {/* Step 3: Address */}
        <div className="rounded-2xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <SectionHeader num={3} title="العنوان والموقع" />
          {usedSavedLocation && (
            <div className="mb-4 flex items-center justify-between rounded-xl border border-blue-200 bg-blue-50 px-4 py-2.5">
              <p className="text-xs font-semibold text-blue-700">
                تم تحميل آخر موقع محفوظ لهذا الزبون من حجز سابق — عدّله إذا تغير عنوانه
              </p>
              <button
                type="button"
                onClick={() => { setAddressDesc(''); setMapPoint(null); setUsedSavedLocation(false) }}
                className="shrink-0 rounded-lg bg-white px-3 py-1 text-xs font-bold text-blue-700 shadow-sm hover:bg-blue-100"
              >
                تغيير الموقع
              </button>
            </div>
          )}
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
            {/* رابط الموقع — بديل عن التحديد على الخريطة، نفس فكرة الموردين */}
            <div className="mt-4">
              <label className="mb-1 block text-sm font-medium text-slate-600">
                أو رابط الموقع (من كوكل ماب)
              </label>
              <input
                value={locationUrl}
                onChange={(e) => setLocationUrl(e.target.value)}
                placeholder="https://maps.app.goo.gl/..."
                dir="ltr"
                className="w-full rounded-xl border border-slate-300 px-4 py-2.5 text-sm outline-none transition-colors focus:border-brand-500"
              />
              <p className="mt-1 text-[11px] text-slate-400">
                إذا حطيت رابط ما تحتاج تأشّر على الخريطة.
              </p>
            </div>
          </div>
        </div>

        {/* Step 4: Service */}
        <div className="rounded-2xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <SectionHeader num={4} title="الخدمات المطلوبة" />
          <p className="mb-3 text-xs text-slate-400">
            تكدر تختار أكثر من خدمة لنفس الزبون (مثلاً منظومة صوت + كاميرات).
          </p>
          {/* ⚠️ قائمة منسدلة بدل ٢٠+ مربّع مفتوح: القائمة المفتوحة كانت
              تاخذ نص الصفحة وتدفن باقي خانات الحجز تحتها، والإداري
              يضطر ينزّل ويطلع بين الخدمة والموعد. */}
          <MultiSelect
            options={services.map((s) => ({ id: s.id, name: s.name }))}
            selected={serviceIds}
            onChange={setServiceIds}
            placeholder="اضغط لاختيار الخدمات..."
            emptyText="ماكو خدمات مسجّلة بالنظام"
          />
          {/* ═══ حقول إجبارية حسب الخدمة ═══
              الخدمة المؤشّرة requiresDeviceInfo (جي بي اس) تطلب عدد
              الأجهزة ونوع المركبة. قبل، حجز الجي بي اس ينوصل للفني
              بلا ولا معلومة عن الأجهزة، ويكتشفها بموقع الزبون.
              ⚠️ السيرفر يرفضها هم — الواجهة تنخدع. */}
          {needsDeviceInfo && (
            <div className="mt-4 rounded-xl border border-amber-200 bg-amber-50/70 p-4">
              <p className="mb-3 text-sm font-bold text-amber-900">
                📡 هذي الخدمة تطلب تفاصيل الأجهزة — إجبارية
              </p>
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <div>
                  <label className="mb-1 block text-xs font-medium text-slate-600">عدد الأجهزة *</label>
                  <input
                    required
                    type="number"
                    min={1}
                    value={deviceCount}
                    onChange={(e) => setDeviceCount(e.target.value)}
                    placeholder="مثال: 3"
                    className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                  />
                </div>
                <div>
                  <label className="mb-1 block text-xs font-medium text-slate-600">نوع المركبة *</label>
                  <input
                    required
                    value={gpsVehicleType}
                    onChange={(e) => setGpsVehicleType(e.target.value)}
                    placeholder="مثال: شاحنة / صالون / بيك أب"
                    className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                  />
                </div>
              </div>
            </div>
          )}

          {serviceIds.length > 1 && (
            <p className="mt-3 rounded-lg bg-brand-50 px-3 py-2 text-xs text-brand-700">
              انختارت {serviceIds.length} خدمات — الخدمة الأولى «
              {services.find((x) => x.id === serviceIds[0])?.name}» تنعتبر الرئيسية.
            </p>
          )}
        </div>

        {/* حجز طاقة شمسية: المنظومة المتفق عليها واستهلاك الزبون */}
        {bookingType === 'SOLAR' && (
          <div className="rounded-2xl border-2 border-amber-200 bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <SectionHeader num={5} title="☀️ تفاصيل المنظومة الشمسية" />

            <div className="mb-4 rounded-xl bg-amber-50 p-3 text-xs text-amber-800">
              السعر المقدّر ينحسب تلقائياً من مكوّنات المنظومة بأسعار المخزن —
              ما تحتاج تكتبه، ويوصل للمنسّق جاهز.
            </div>

            <label className="mb-1 block text-sm font-medium text-slate-600">
              المنظومة المتفق عليها (اختيارية — تكدر تتركها ويحددها المنسّق بعد المعاينة)
            </label>
            <select
              value={solarSystemId}
              onChange={(e) => setSolarSystemId(e.target.value)}
              className="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm outline-none focus:border-amber-500"
            >
              <option value="">— لسه ما تحددت —</option>
              {solarSystems.map((sys) => (
                <option key={sys.id} value={sys.id}>
                  {sys.brand} · {sys.capacity} · {sys.model} — {Math.round(sys.price.total).toLocaleString('en-US')} د.ع
                </option>
              ))}
            </select>

            {(() => {
              const sys = solarSystems.find((x) => x.id === solarSystemId)
              if (!sys) return null
              return (
                <div className="mt-3 rounded-xl border border-amber-200 bg-amber-50/60 p-3 text-xs">
                  <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
                    <Info label="الألواح" value={sys.panel ? `${sys.panelQty} × ${sys.panel.name}` : '—'} />
                    <Info label="الإنفيرتر" value={sys.inverter ? `${sys.inverterQty} × ${sys.inverter.name}` : '—'} />
                    <Info label="البطاريات" value={sys.battery ? `${sys.batteryQty} × ${sys.battery.name}` : '—'} />
                    <Info label="السعر التقديري" value={`${Math.round(sys.price.total).toLocaleString('en-US')} د.ع`} />
                  </div>
                  {sys.shortages.length > 0 && (
                    <div className="mt-2 rounded-lg bg-red-100 p-2 font-bold text-red-700">
                      ⚠️ انتبه: مكوّنات هذي المنظومة ما متوفرة كاملة بالمخزن حالياً —
                      لا توعد الزبون بموعد قريب قبل ما تتأكد من الإداري.
                    </div>
                  )}
                </div>
              )
            })()}

            <label className="mb-1 mt-4 block text-sm font-medium text-slate-600">
              استهلاك الزبون الشهري (كيلو واط/ساعة) — اختياري
            </label>
            <input
              type="number"
              min={0}
              value={solarMonthlyKwh}
              onChange={(e) => setSolarMonthlyKwh(e.target.value)}
              placeholder="مثال: 500 — يساعد الإداري يتأكد إن السعة مناسبة"
              className="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm outline-none focus:border-amber-500"
            />
            {Number(solarMonthlyKwh) > 0 && (
              <p className="mt-2 text-xs text-slate-500">
                بهذا الاستهلاك، السعة المناسبة تقريباً{' '}
                <b className="text-amber-700">
                  {Math.ceil(Number(solarMonthlyKwh) / 30 / 4.5)} كيلو واط
                </b>{' '}
                (محسوبة على ٤٫٥ ساعة شمس فعّالة باليوم)
              </p>
            )}
          </div>
        )}

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
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
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
              <button
                type="button"
                onClick={() => { setMaintenanceType('UPKEEP'); setRemembersCrew(false) }}
                className={`flex items-start gap-3 rounded-2xl border-2 p-5 text-right transition-all ${
                  maintenanceType === 'UPKEEP'
                    ? 'border-emerald-500 bg-emerald-50 shadow-md'
                    : 'border-slate-200 bg-white hover:border-emerald-300'
                }`}
              >
                <span className="mt-0.5 h-3 w-3 shrink-0 rounded-full bg-emerald-500" />
                <div>
                  <span className="block text-sm font-bold text-slate-800">إدامة</span>
                  <span className="block text-xs text-slate-500">صيانة دورية (مثل غسل الألواح الشمسية)</span>
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

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <div className="text-[10px] text-slate-500">{label}</div>
      <div className="font-bold text-slate-800">{value}</div>
    </div>
  )
}
