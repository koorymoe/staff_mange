import { useEffect, useState } from 'react'
import { api, type Customer, type GpsCustomerListItem, type Booking } from '../api'
import { validateCustomerName, validateCustomerPhone } from '../validation'
import { matches } from '../utils/search'

function splitFullName(fullName: string): [string, string, string, string] {
  const parts = fullName.trim().split(/\s+/).filter(Boolean)
  return [parts[0] || '', parts[1] || '', parts[2] || '', parts.slice(3).join(' ')]
}

const serviceLabels: Record<string, string> = {
  GPS: 'جي بي اس',
}

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
  const [tab, setTab] = useState<'all' | 'gps'>('all')
  const [customers, setCustomers] = useState<Customer[]>([])
  const [gpsCustomers, setGpsCustomers] = useState<GpsCustomerListItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [history, setHistory] = useState<Booking[]>([])
  const [historyLoading, setHistoryLoading] = useState(false)
  const [search, setSearch] = useState('')

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

  const load = () => {
    Promise.all([api.getCustomers(), api.getCustomersByGpsService()])
      .then(([all, gps]) => {
        setCustomers(all)
        setGpsCustomers(gps)
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

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

  const selectedCustomer = customers.find((c) => c.id === selectedId) || null

  const activeList: Customer[] = tab === 'gps' ? gpsCustomers : customers
  // البحث يمر بالتطبيع العربي: «احمد» تلكه «أحمد»، و«٠٧٧٠» تلكه «0770»
  const filteredCustomers = search.trim()
    ? activeList.filter((c) => matches([c.code, c.name, c.phone], search))
    : activeList

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

          <div className="relative">
            <input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="دور بكود الزبون (CUST-00001)، الاسم، أو رقم الهاتف..."
              className="w-full rounded-xl border border-slate-300 bg-white px-4 py-2.5 outline-none transition-colors focus:border-brand-500 focus:ring-1 focus:ring-brand-500/20"
            />
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
                  <th className="px-4 py-3 text-sm font-semibold">تعديل</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {filteredCustomers.map((c) => (
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
                    <td className="px-4 py-3">
                      <button
                        onClick={(e) => {
                          e.stopPropagation()
                          openEdit(c)
                        }}
                        className="rounded-lg bg-slate-100 px-3 py-1 text-xs font-semibold text-slate-600 hover:bg-slate-200"
                      >
                        تعديل
                      </button>
                    </td>
                  </tr>
                ))}
                {filteredCustomers.length === 0 && (
                  <tr>
                    <td colSpan={6} className="px-4 py-6 text-center text-slate-400">
                      {activeList.length === 0 ? 'لا يوجد زبائن بعد' : 'لا توجد نتائج مطابقة للبحث'}
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
            </div>
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
                    <span className="text-slate-500">آخر موقع محفوظ: </span>
                    {selectedCustomer.location || '-'}
                    {selectedCustomer.mapLatitude != null && selectedCustomer.mapLongitude != null && (
                      <a
                        href={`https://www.openstreetmap.org/?mlat=${selectedCustomer.mapLatitude}&mlon=${selectedCustomer.mapLongitude}#map=17/${selectedCustomer.mapLatitude}/${selectedCustomer.mapLongitude}`}
                        target="_blank"
                        rel="noreferrer"
                        className="mr-2 font-semibold text-brand-600 underline"
                      >
                        عرض على الخريطة
                      </a>
                    )}
                  </p>
                </div>

                {!historyLoading && (
                  <div className="mt-4 grid grid-cols-2 gap-3 sm:grid-cols-4">
                    <div className="rounded-xl bg-slate-50 p-3 text-center">
                      <p className="text-xl font-bold text-brand-700">{history.length}</p>
                      <p className="text-xs text-slate-500">إجمالي الحجوزات</p>
                    </div>
                    <div className="rounded-xl bg-slate-50 p-3 text-center">
                      <p className="text-xl font-bold text-emerald-600">
                        {history.filter((b) => b.status === 'COMPLETED').length}
                      </p>
                      <p className="text-xs text-slate-500">زيارات مكتملة</p>
                    </div>
                    <div className="rounded-xl bg-slate-50 p-3 text-center">
                      <p className="text-xl font-bold text-blue-600">
                        {history.filter((b) => b.status === 'IN_PROGRESS').length}
                      </p>
                      <p className="text-xs text-slate-500">قيد التنفيذ</p>
                    </div>
                    <div className="rounded-xl bg-slate-50 p-3 text-center">
                      <p className="text-xl font-bold text-amber-600">
                        {history.filter((b) => b.status === 'PENDING').length}
                      </p>
                      <p className="text-xs text-slate-500">بانتظار التثبيت</p>
                    </div>
                  </div>
                )}

                <h4 className="mt-5 font-bold text-brand-800">أرشيف طلبات وعناوين الزبون</h4>
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
                          <p className="sm:col-span-2">
                            <span className="text-slate-400">العنوان: </span>
                            {b.address || '-'}
                            {b.mapLatitude != null && b.mapLongitude != null && (
                              <a
                                href={`https://www.openstreetmap.org/?mlat=${b.mapLatitude}&mlon=${b.mapLongitude}#map=17/${b.mapLatitude}/${b.mapLongitude}`}
                                target="_blank"
                                rel="noreferrer"
                                className="mr-2 font-semibold text-brand-600 underline"
                              >
                                عرض على الخريطة
                              </a>
                            )}
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
