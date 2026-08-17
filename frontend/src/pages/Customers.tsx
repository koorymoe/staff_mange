import { useEffect, useState } from 'react'
import Pager from '../components/Pager'
import { api, type Customer, type GpsCustomerListItem, type Booking } from '../api'
import { validateCustomerName, validateCustomerPhone } from '../validation'
import { matches } from '../utils/search'
// تسمية الحالة من مصدر واحد للنظام كله
import { bookingStatusLabel, bookingStatusColor } from '../bookingStatus'

function splitFullName(fullName: string): [string, string, string, string] {
  const parts = fullName.trim().split(/\s+/).filter(Boolean)
  return [parts[0] || '', parts[1] || '', parts[2] || '', parts.slice(3).join(' ')]
}

const serviceLabels: Record<string, string> = {
  GPS: 'جي بي اس',
}

export default function Customers() {
  const [tab, setTab] = useState<'all' | 'gps'>('all')
  const [customers, setCustomers] = useState<Customer[]>([])
  const [gpsCustomers, setGpsCustomers] = useState<GpsCustomerListItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [history, setHistory] = useState<Booking[]>([])
  const [historyLoading, setHistoryLoading] = useState(false)
  const [search, setSearch] = useState('')
  // ── ترتيب الشاشة الجديد ──
  const [showAdd, setShowAdd] = useState(false)
  const [locationFilter, setLocationFilter] = useState('')
  const [claimFilter, setClaimFilter] = useState('')
  const [kindFilter, setKindFilter] = useState('')
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState(10)

  const [name, setName] = useState('')
  const [phone, setPhone] = useState('')
  const [location, setLocation] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [message, setMessage] = useState<string | null>(null)

  const [editingCustomer, setEditingCustomer] = useState<Customer | null>(null)
  const [editFirstName, setEditFirstName] = useState('')
  const [editFatherName, setEditFatherName] = useState('')
  const [editGrandfatherName, setEditGrandfatherName] = useState('')
  const [editFamilyName, setEditFamilyName] = useState('')
  const [editPhone, setEditPhone] = useState('')
  const [editSubmitting, setEditSubmitting] = useState(false)
  const [editMessage, setEditMessage] = useState<string | null>(null)
  const editName = [editFirstName, editFatherName, editGrandfatherName, editFamilyName].map((p) => p.trim()).filter(Boolean).join(' ')

  // ═══ البحث صار بالسيرفر ═══
  //
  // كانت الصفحة تنزّل **كل** الزبائن بكل فتحة وتفلترهم بالمتصفح — بـ٥٠٠٠
  // زبون هذا ١٫٤ ميغا كل مرة، وعلى 4G ثواني ينتظرها الموظف بلا فايدة.
  //
  // هسه: نجيب أول ٥٠٠ للتصفح، وأول ما يكتب حرفين نسأل السيرفر. السيرفر
  // يطبّع النص بنفس طريقة الواجهة بالضبط (ar_norm) فالنتيجة وحدة.
  const INITIAL_LIMIT = 500
  const load = () => {
    Promise.all([api.getCustomers({ limit: INITIAL_LIMIT }), api.getCustomersByGpsService()])
      .then(([all, gps]) => {
        setCustomers(all)
        setGpsCustomers(gps)
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  // بحث السيرفر — بتأخير ٣٠٠ملي حتى ما نرسل طلب بكل ضغطة زر، ومع إلغاء
  // الطلب القديم حتى نتيجة متأخرة ما تدوس على نتيجة أحدث.
  useEffect(() => {
    const q = search.trim()
    if (q.length < 2) return
    let alive = true
    const t = setTimeout(() => {
      api.getCustomers({ search: q, limit: 200 })
        .then((rows) => { if (alive) setCustomers(rows) })
        .catch(() => {})
    }, 300)
    return () => { alive = false; clearTimeout(t) }
  }, [search])

  useEffect(load, [])

  const openEdit = (c: Customer) => {
    setEditingCustomer(c)
    const [f, fa, gf, fam] = splitFullName(c.name)
    setEditFirstName(f)
    setEditFatherName(fa)
    setEditGrandfatherName(gf)
    setEditFamilyName(fam)
    setEditPhone(c.phone)
    setEditMessage(null)
  }

  const handleEditSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!editingCustomer) return
    setEditMessage(null)

    const nameError = validateCustomerName(editName)
    if (nameError) {
      setEditMessage(nameError)
      return
    }
    const phoneError = validateCustomerPhone(editPhone)
    if (phoneError) {
      setEditMessage(phoneError)
      return
    }

    setEditSubmitting(true)
    try {
      await api.updateCustomer(editingCustomer.id, { name: editName, phone: editPhone })
      setEditingCustomer(null)
      load()
    } catch (e) {
      setEditMessage(e instanceof Error ? e.message : 'حدث خطأ')
    } finally {
      setEditSubmitting(false)
    }
  }

  useEffect(() => {
    // Guard-clause reset of the booking-history panel when selection is cleared.
    if (!selectedId) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setHistory([])
      return
    }
    setHistoryLoading(true)
    api
      .getBookings({ customerId: selectedId })
      .then(setHistory)
      .finally(() => setHistoryLoading(false))
  }, [selectedId])

  const activeList: Customer[] = tab === 'gps' ? gpsCustomers : customers

  // قائمة المواقع تنبني من الزبائن نفسهم — ماكو جدول مواقع بالنظام،
  // وكتابتها بالإيد تعني فلتر يفوت مواقع جديدة.
  const locationOptions = Array.from(
    new Set(activeList.map((c) => (c.location || '').trim()).filter(Boolean)),
  ).sort().slice(0, 60)

  // البحث يمر بالتطبيع العربي: «احمد» تلكه «أحمد»، و«٠٧٧٠» تلكه «0770»
  const filteredCustomers = activeList
    .filter((c) => (search.trim() ? matches([c.code, c.name, c.phone], search) : true))
    .filter((c) => (locationFilter ? (c.location || '').trim() === locationFilter : true))
    .filter((c) => (kindFilter === 'vip' ? !!c.position : kindFilter === 'normal' ? !c.position : true))
    .filter((c) => (claimFilter === 'false' ? (c.falseClaimCount ?? 0) > 0
      : claimFilter === 'clean' ? (c.falseClaimCount ?? 0) === 0 : true))

  const filtersOn = !!(search.trim() || locationFilter || kindFilter || claimFilter)

  // ⚠️ الصفحة تنحصر بالعدد الموجود: بعد التصفية ممكن تكون واقف بصفحة
  // ٥ وماكو إلا صفحتين، فتشوف جدول فاضي وتظن ماكو زبائن.
  const pageCount = Math.max(1, Math.ceil(filteredCustomers.length / perPage))
  const safePage = Math.min(page, pageCount)
  const pageRows = filteredCustomers.slice((safePage - 1) * perPage, safePage * perPage)

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

  // ── أرقام الرأس ──
  // ⚠️ «جدد هذا الشهر» ينحسب من `createdAt` الحقيقي مو من ترتيب
  // القائمة: الترتيب يتغيّر بالبحث، والرقم ينطلع كذب.
  const monthStart = new Date(); monthStart.setDate(1); monthStart.setHours(0, 0, 0, 0)
  const newThisMonth = customers.filter((c) => c.createdAt && new Date(c.createdAt) >= monthStart).length
  const vipCount = customers.filter((c) => !!c.position).length
  const pct = (n: number) => (customers.length ? Math.round((n / customers.length) * 1000) / 10 : 0)

  return (
    <div dir="rtl">
      {/* ═══ الرأس ═══ */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-brand-50 text-lg sm:h-12 sm:w-12 sm:text-2xl">👥</span>
          <div className="min-w-0">
            <h2 className="text-xl font-black text-[#0f2040] sm:text-2xl">الزبائن</h2>
            <p className="text-[11px] text-slate-500 sm:text-xs">إدارة بيانات الزبائن ومتابعة حجوزاتهم وخدماتهم</p>
          </div>
        </div>
      </div>

      {/* ═══ الأرقام ═══ */}
      <div className="mt-4 grid grid-cols-1 gap-2.5 sm:grid-cols-3 sm:gap-3">
        <StatCard icon="👥" tone="sky" label="كل الزبائن" value={customers.length} hint="100% من إجمالي الزبائن" />
        <StatCard icon="🧑‍💼" tone="emerald" label="جدد هذا الشهر" value={newThisMonth} hint={`${pct(newThisMonth)}% من إجمالي الزبائن`} />
        <StatCard icon="👑" tone="amber" label="شخصيات مهمة / VIP" value={vipCount} hint={`${pct(vipCount)}% من إجمالي الزبائن`} />
      </div>

      {/* ═══ إضافة زبون ═══
          ⚠️ النموذج مطوي: كان يفتح أول الصفحة دائماً ويدفع الجدول
          تحته، والإداري بأغلب دخوله جاي **يدوّر** على زبون مو يضيف. */}
      <div className="mt-4 flex justify-end">
        <button
          onClick={() => setShowAdd((v) => !v)}
          className="rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 px-5 py-2.5 text-sm font-bold text-white shadow-md"
        >
          {showAdd ? '✕ إغلاق' : '＋ إضافة زبون'}
        </button>
      </div>

      <form
        onSubmit={handleSubmit}
        className={`mt-3 grid-cols-1 gap-4 rounded-2xl border border-slate-200 bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)] sm:grid-cols-3 ${showAdd ? 'grid' : 'hidden'}`}
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
          <div className="flex gap-2">
            <button
              onClick={() => setTab('all')}
              className={`rounded-lg px-4 py-2 text-sm font-bold transition-colors ${
                tab === 'all' ? 'bg-brand-600 text-white shadow-md' : 'bg-white text-slate-500 hover:bg-slate-100'
              }`}
            >
              كل الزبائن ({customers.length})
            </button>
            <button
              onClick={() => setTab('gps')}
              className={`rounded-lg px-4 py-2 text-sm font-bold transition-colors ${
                tab === 'gps' ? 'bg-brand-600 text-white shadow-md' : 'bg-white text-slate-500 hover:bg-slate-100'
              }`}
            >
              زبائن الجي بي اس ({gpsCustomers.length})
            </button>
          </div>

          {/* ═══ الفلاتر ═══ */}
          <div className="rounded-2xl border border-slate-200 bg-white p-3 shadow-[0_2px_12px_rgba(15,32,64,0.05)]">
            <div className="grid grid-cols-1 gap-2.5 lg:grid-cols-4">
              <div className="lg:col-span-2">
                <label className="mb-1 block text-[10px] font-bold text-slate-500">بحث</label>
                <input
                  value={search}
                  onChange={(e) => { setSearch(e.target.value); setPage(1) }}
                  placeholder="🔍 ابحث بالاسم أو الكود أو الهاتف..."
                  className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                />
              </div>
              <div>
                <label className="mb-1 block text-[10px] font-bold text-slate-500">الموقع</label>
                <select
                  value={locationFilter}
                  onChange={(e) => { setLocationFilter(e.target.value); setPage(1) }}
                  className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                >
                  <option value="">الكل</option>
                  {locationOptions.map((l) => <option key={l} value={l}>{l}</option>)}
                </select>
              </div>
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <label className="mb-1 block text-[10px] font-bold text-slate-500">التصنيف</label>
                  <select
                    value={kindFilter}
                    onChange={(e) => { setKindFilter(e.target.value); setPage(1) }}
                    className="w-full rounded-lg border border-slate-300 px-2 py-2 text-sm outline-none focus:border-brand-500"
                  >
                    <option value="">الكل</option>
                    <option value="vip">شخصية مهمة</option>
                    <option value="normal">عادي</option>
                  </select>
                </div>
                <div>
                  <label className="mb-1 block text-[10px] font-bold text-slate-500">البلاغات</label>
                  <select
                    value={claimFilter}
                    onChange={(e) => { setClaimFilter(e.target.value); setPage(1) }}
                    className="w-full rounded-lg border border-slate-300 px-2 py-2 text-sm outline-none focus:border-brand-500"
                  >
                    <option value="">الكل</option>
                    <option value="clean">بلا بلاغات</option>
                    <option value="false">عنده بلاغ كاذب</option>
                  </select>
                </div>
              </div>
            </div>
            {filtersOn && (
              <div className="mt-2.5 flex flex-wrap items-center justify-between gap-2">
                <p className="text-[11px] text-slate-500">النتائج: {filteredCustomers.length} من {activeList.length}</p>
                <button
                  onClick={() => { setSearch(''); setLocationFilter(''); setClaimFilter(''); setKindFilter(''); setPage(1) }}
                  className="rounded-lg border border-slate-300 px-3 py-1.5 text-[11px] font-bold text-slate-600 hover:bg-slate-50"
                >
                  ✖ مسح الفلاتر
                </button>
              </div>
            )}
          </div>

          <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <div className="overflow-x-auto">
            <table className="w-full text-right">
              <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                <tr>
                  <th className="px-4 py-3 text-sm font-semibold">الكود</th>
                  <th className="px-4 py-3 text-sm font-semibold">الاسم</th>
                  <th className="px-4 py-3 text-sm font-semibold">الهاتف</th>
                  {tab === 'all' && <th className="px-4 py-3 text-sm font-semibold">الخدمات</th>}
                  {tab === 'all' && <th className="px-4 py-3 text-sm font-semibold">الموقع</th>}
                  {tab === 'gps' && <th className="px-4 py-3 text-sm font-semibold">رقم الجهاز</th>}
                  {tab === 'gps' && <th className="px-4 py-3 text-sm font-semibold">انتهاء الاشتراك</th>}
                  {tab === 'all' && <th className="px-4 py-3 text-sm font-semibold">التصنيف</th>}
                  <th className="px-4 py-3 text-sm font-semibold">الإجراءات</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {pageRows.map((c) => (
                  <>
                  <tr
                    key={c.id}
                    onClick={() => setSelectedId(selectedId === c.id ? null : c.id)}
                    className={`cursor-pointer transition-colors ${
                      selectedId === c.id ? 'bg-brand-50' : 'hover:bg-slate-50'
                    }`}
                  >
                    <td className="px-4 py-3 font-mono text-sm font-semibold text-brand-600">
                      {c.code}
                    </td>
                    <td className="px-4 py-3">{c.name}</td>
                    <td className="px-4 py-3 text-slate-500">{c.phone}</td>
                    {tab === 'all' && (
                      <td className="px-4 py-3">
                        <div className="flex flex-wrap gap-1">
                          {c.services.length === 0 && <span className="text-slate-400">-</span>}
                          {c.services.map((s) => (
                            <span key={s} className="rounded-full bg-brand-50 px-2 py-0.5 text-xs font-semibold text-brand-700">
                              {serviceLabels[s] || s}
                            </span>
                          ))}
                        </div>
                      </td>
                    )}
                    {tab === 'all' && <td className="px-4 py-3 text-slate-500">{c.location || '-'}</td>}
                    {tab === 'gps' && (
                      <td className="px-4 py-3 text-slate-500">{(c as GpsCustomerListItem).deviceId || '-'}</td>
                    )}
                    {tab === 'gps' && (
                      <td className="px-4 py-3 text-slate-500">
                        {(c as GpsCustomerListItem).subscriptionEnd
                          ? new Date((c as GpsCustomerListItem).subscriptionEnd as string).toLocaleDateString('ar-IQ')
                          : '-'}
                      </td>
                    )}
                    {tab === 'all' && (
                      <td className="px-4 py-3">
                        {c.position ? (
                          <span className="rounded-full bg-amber-100 px-2 py-0.5 text-[11px] font-bold text-amber-800">👑 {c.position}</span>
                        ) : (c.falseClaimCount ?? 0) > 0 ? (
                          <span className="rounded-full bg-red-100 px-2 py-0.5 text-[11px] font-bold text-red-700">
                            ⚠️ بلاغ كاذب ×{c.falseClaimCount}
                          </span>
                        ) : (
                          <span className="rounded-full bg-emerald-50 px-2 py-0.5 text-[11px] font-bold text-emerald-700">عادي</span>
                        )}
                      </td>
                    )}
                    <td className="px-4 py-3">
                      {/* 👁 يفتح التفاصيل جوّا الصف — مثل التصميم:
                          الإداري يشوف تفاصيل الزبون بلا ما يفقد مكانه
                          بالجدول ولا ينزّل لآخر الصفحة. */}
                      <div className="flex items-center gap-1.5">
                        <button
                          onClick={(e) => { e.stopPropagation(); setSelectedId(selectedId === c.id ? null : c.id) }}
                          title="عرض التفاصيل"
                          className={`rounded-lg px-2.5 py-1.5 text-xs font-bold transition ${
                            selectedId === c.id ? 'bg-brand-600 text-white' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
                          }`}
                        >
                          👁 تفاصيل
                        </button>
                        <button
                          onClick={(e) => { e.stopPropagation(); openEdit(c) }}
                          title="تعديل"
                          className="rounded-lg bg-slate-100 px-2 py-1.5 text-xs text-slate-600 hover:bg-slate-200"
                        >
                          ✏️
                        </button>
                      </div>
                    </td>
                  </tr>

                  {/* ── صف التفاصيل ── */}
                  {selectedId === c.id && (
                    <tr key={`${c.id}-details`} className="bg-slate-50/70">
                      <td colSpan={9} className="px-4 py-4">
                        <CustomerDetails
                          customer={c}
                          history={history}
                          loading={historyLoading}
                        />
                      </td>
                    </tr>
                  )}
                  </>
                ))}
                {filteredCustomers.length === 0 && (
                  <tr>
                    <td colSpan={9} className="px-4 py-6 text-center text-slate-400">
                      {activeList.length === 0 ? 'لا يوجد زبائن بعد' : 'لا توجد نتائج مطابقة للبحث'}
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
            </div>
          </div>

          {/* ═══ الترقيم ═══ */}
          {/* مكوّن مشترك وية شاشة الحجوزات، حتى ما يفترقن أول تعديل */}
          {filteredCustomers.length > 0 && (
            <Pager
              page={safePage}
              perPage={perPage}
              total={filteredCustomers.length}
              unit="زبون"
              onPage={setPage}
              onPerPage={setPerPage}
            />
          )}
        </div>
      )}

      {editingCustomer && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
          <form
            onSubmit={handleEditSubmit}
            className="w-full max-w-md rounded-2xl bg-white p-6 shadow-xl"
          >
            <h3 className="text-lg font-bold text-brand-900">تعديل بيانات الزبون {editingCustomer.code}</h3>
            <div className="mt-4 flex flex-col gap-4">
              <div>
                <label className="mb-1 block text-sm font-medium text-slate-600">الاسم الرباعي</label>
                <div className="grid grid-cols-2 gap-2">
                  <input
                    required
                    placeholder="الاسم"
                    value={editFirstName}
                    onChange={(e) => setEditFirstName(e.target.value)}
                    className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
                  />
                  <input
                    placeholder="اسم الأب"
                    value={editFatherName}
                    onChange={(e) => setEditFatherName(e.target.value)}
                    className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
                  />
                  <input
                    placeholder="اسم الجد"
                    value={editGrandfatherName}
                    onChange={(e) => setEditGrandfatherName(e.target.value)}
                    className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
                  />
                  <input
                    placeholder="اللقب"
                    value={editFamilyName}
                    onChange={(e) => setEditFamilyName(e.target.value)}
                    className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
                  />
                </div>
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium text-slate-600">رقم الهاتف</label>
                <input
                  required
                  maxLength={11}
                  value={editPhone}
                  onChange={(e) => setEditPhone(e.target.value)}
                  className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
                />
              </div>
              {editMessage && <p className="text-sm text-red-600">{editMessage}</p>}
              <div className="flex gap-2">
                <button
                  type="submit"
                  disabled={editSubmitting}
                  className="flex-1 rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-2 font-medium text-white shadow-md disabled:opacity-50"
                >
                  {editSubmitting ? 'جاري الحفظ...' : 'حفظ التعديل'}
                </button>
                <button
                  type="button"
                  onClick={() => setEditingCustomer(null)}
                  className="rounded-lg bg-slate-100 px-4 py-2 font-medium text-slate-600 hover:bg-slate-200"
                >
                  إلغاء
                </button>
              </div>
            </div>
          </form>
        </div>
      )}
    </div>
  )
}

/* ───── قطع الشاشة ───── */

// بطاقة رقم بالرأس
function StatCard({ icon, tone, label, value, hint }: {
  icon: string; tone: 'sky' | 'emerald' | 'amber'; label: string; value: number; hint: string
}) {
  const tones: Record<string, string> = {
    sky: 'bg-sky-50 text-sky-600',
    emerald: 'bg-emerald-50 text-emerald-600',
    amber: 'bg-amber-50 text-amber-600',
  }
  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-4 shadow-[0_2px_12px_rgba(15,32,64,0.05)]">
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0">
          <p className="truncate text-[11px] font-bold text-slate-500">{label}</p>
          <p className="mt-1 text-3xl font-black text-[#0f2040]">{value}</p>
          <p className="text-[10px] text-slate-400">{hint}</p>
        </div>
        <span className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-xl text-lg ${tones[tone]}`}>{icon}</span>
      </div>
    </div>
  )
}

// ⚠️ `PageBtn` و`pageWindow` انشالن من هنا: صارن جوّا `Pager`
// المشترك، ونسخة ثانية منهن تعني ترقيمين يفترقن أول تعديل.

/* ───── تفاصيل الزبون — تنفتح جوّا الصف ───── */

function CustomerDetails({ customer, history, loading }: {
  customer: Customer; history: Booking[]; loading: boolean
}) {
  const last = history[0]
  const services = Array.from(new Set(history.map((b) => b.service?.name).filter(Boolean) as string[]))
  const crew = last ? last.assignments.map((a) => a.employee.name) : []

  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
      <DetailBox icon="⚠️" title="حالة البلاغات">
        {(customer.falseClaimCount ?? 0) > 0 ? (
          <>
            <p className="font-bold text-red-700">انكشف {customer.falseClaimCount} بلاغ غير صحيح</p>
            {customer.falseClaimNote && <p className="mt-1 text-slate-600">{customer.falseClaimNote}</p>}
            {customer.lastFalseClaimAt && (
              <p className="mt-1 text-[10px] text-slate-400">
                آخر مرة: {new Date(customer.lastFalseClaimAt).toLocaleDateString('ar-IQ')}
              </p>
            )}
          </>
        ) : (
          <p className="text-emerald-700">ماكو بلاغات غير صحيحة على هذا الزبون.</p>
        )}
      </DetailBox>

      <DetailBox icon="🕐" title="آخر حجز">
        {loading ? <p className="text-slate-400">جاري التحميل...</p> : last ? (
          <>
            <p><span className="text-slate-400">الكود: </span><span className="font-mono font-bold">{last.code}</span></p>
            <p><span className="text-slate-400">التاريخ: </span>{new Date(last.createdAt).toLocaleDateString('ar-IQ')}</p>
            {last.scheduledAt && (
              <p><span className="text-slate-400">الموعد: </span>
                {new Date(last.scheduledAt).toLocaleString('ar-IQ', { day: 'numeric', month: 'long', hour: '2-digit', minute: '2-digit' })}
              </p>
            )}
          </>
        ) : <p className="text-slate-400">ماكو حجوزات بعد.</p>}
      </DetailBox>

      <DetailBox icon="👥" title="الكادر المكلّف">
        {loading ? <p className="text-slate-400">جاري التحميل...</p> : crew.length > 0 ? (
          <ul className="space-y-0.5">
            {crew.map((n, i) => <li key={i}>• {n}</li>)}
          </ul>
        ) : <p className="text-slate-400">ما انكلّف كادر بآخر حجز.</p>}
      </DetailBox>

      <DetailBox icon="🔧" title="الخدمات المطلوبة">
        {loading ? <p className="text-slate-400">جاري التحميل...</p> : services.length > 0 ? (
          <ul className="space-y-0.5">
            {services.map((n) => <li key={n}>• {n}</li>)}
          </ul>
        ) : <p className="text-slate-400">ماكو خدمات مسجّلة.</p>}
      </DetailBox>

      {/* ── سجل الحجوزات ── */}
      <div className="sm:col-span-2 xl:col-span-4">
        <p className="mb-2 text-xs font-bold text-[#0f2040]">
          📋 سجل الحجوزات {history.length > 0 && <span className="text-slate-400">({history.length})</span>}
        </p>
        {loading && <p className="text-xs text-slate-400">جاري التحميل...</p>}
        {!loading && history.length === 0 && (
          <p className="rounded-xl border border-dashed border-slate-200 bg-white px-3 py-4 text-center text-xs text-slate-400">
            ماكو حجوزات لهذا الزبون.
          </p>
        )}
        {/* ⚠️ تمرير أفقي جوّا الصندوق: بلا سقف، الحجوزات الكثيرة
            تمدّ الصف وتكسر الجدول كله على الموبايل. */}
        <div className="max-h-56 overflow-y-auto rounded-xl border border-slate-200 bg-white">
          {history.map((b) => (
            <div key={b.id} className="flex flex-wrap items-center gap-2 border-b border-slate-100 px-3 py-2 text-[11px] last:border-0">
              <span className="font-mono font-bold text-brand-600">{b.code}</span>
              <span className="text-slate-600">{b.service?.name || '—'}</span>
              <span className="text-slate-400">{new Date(b.createdAt).toLocaleDateString('ar-IQ')}</span>
              <span className={`mr-auto rounded-full px-2 py-0.5 font-bold ${bookingStatusColor(b.status)}`}>
                {bookingStatusLabel(b.status)}
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

function DetailBox({ icon, title, children }: {
  icon: string; title: string; children: React.ReactNode
}) {
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-3">
      <p className="mb-1.5 border-b border-slate-100 pb-1.5 text-[11px] font-extrabold text-[#0f2040]">
        {icon} {title}
      </p>
      <div className="space-y-0.5 text-[11px] text-slate-700">{children}</div>
    </div>
  )
}
