import { useEffect, useState } from 'react'
import { api, type VipCustomer, type Customer } from '../api'

// قائمة الشخصيات المهمة — لمدير النظام حصراً (الراوت بالسيرفر محمي بـrequireAdmin).
// تعرض تفاصيل الزبون ورقمه وشنو طلب من عدنا ومنو الموظف الي علّمه.
export default function VipCustomersPage() {
  const [items, setItems] = useState<VipCustomer[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState<string | null>(null)

  // الإضافة اليدوية: الإداري يدز الرقم، والنظام يطلعله الزبون بمعلوماته
  const [phone, setPhone] = useState('')
  const [found, setFound] = useState<Customer | null>(null)
  const [notFound, setNotFound] = useState(false)
  const [position, setPosition] = useState('')
  const [summary, setSummary] = useState('')
  const [note, setNote] = useState('')
  const [adding, setAdding] = useState(false)
  // الشخصية المهمة مو لازم تكون زبون عدنا — نتعرف عليها بأي مكان
  const [newName, setNewName] = useState('')
  const [newLocation, setNewLocation] = useState('')
  const [newLocationUrl, setNewLocationUrl] = useState('')
  const [boughtFromUs, setBoughtFromUs] = useState(true)

  // نبحث تلقائياً أول ما يكتمل الرقم — بلا زر بحث
  useEffect(() => {
    const p = phone.trim()
    if (p.length < 10) {
      // تفضية النتيجة السابقة خارج جسم الـeffect المتزامن
      queueMicrotask(() => { setFound(null); setNotFound(false) })
      return
    }
    let active = true
    api.lookupCustomer(p)
      .then((c) => { if (active) { setFound(c); setNotFound(!c) } })
      .catch(() => { if (active) { setFound(null); setNotFound(true) } })
    return () => { active = false }
  }, [phone])

  const addManual = async () => {
    // موجود بالنظام → نعلّمه. مو موجود → لازم اسم حتى ننشئ سجله.
    if (!found && !newName.trim()) { alert('اكتب اسم الشخصية أول'); return }
    setAdding(true)
    try {
      await api.markVipCustomer({
        phone: phone.trim(),
        name: found ? undefined : newName.trim(),
        location: found ? undefined : (newLocation.trim() || undefined),
        locationUrl: found ? undefined : (newLocationUrl.trim() || undefined),
        boughtFromUs,
        customerPosition: position.trim() || undefined,
        requestSummary: summary.trim() || undefined,
        note: note.trim() || undefined,
      })
      setPhone(''); setFound(null); setPosition(''); setSummary(''); setNote('')
      setNewName(''); setNewLocation(''); setNewLocationUrl(''); setBoughtFromUs(true)
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذرت الإضافة')
    } finally {
      setAdding(false)
    }
  }

  const load = () => {
    api.getVipCustomers()
      .then(setItems)
      .catch((e) => setError(e instanceof Error ? e.message : 'تعذر جلب القائمة'))
      .finally(() => setLoading(false))
  }
  useEffect(load, [])

  const remove = async (customerId: string) => {
    if (!confirm('إزالة تعليم "شخصية مهمة" عن هذا الزبون؟')) return
    setBusy(customerId)
    try {
      await api.unmarkVipCustomer(customerId)
      load()
    } finally {
      setBusy(null)
    }
  }

  return (
    <div dir="rtl">
      <h2 className="text-2xl font-bold text-brand-900">⭐ الشخصيات المهمة</h2>
      <p className="mt-1 text-slate-500">
        الزبائن الي علّمهم الموظفون كشخصيات مهمة — مع تفاصيل الزبون وشنو طلب من عدنا ومنو الموظف الي علّمه.
      </p>

      {/* إضافة يدوية — للإداري ومدير النظام (السيرفر يفرض القيد كمان) */}
      <div className="mt-6 rounded-xl border border-amber-200 bg-amber-50 p-5">
        <h3 className="font-bold text-amber-900">➕ إضافة شخصية مهمة يدوياً</h3>
        <p className="mt-1 text-xs text-amber-700">
          اكتب رقم الزبون بس — النظام يطلعلك اسمه ومعلوماته كاملة.
        </p>

        <input
          value={phone} onChange={(e) => setPhone(e.target.value)} dir="ltr" inputMode="numeric"
          placeholder="07XXXXXXXXX"
          className="mt-3 w-full max-w-xs rounded-lg border border-amber-300 bg-white px-4 py-2.5 text-right outline-none focus:border-amber-500"
        />

        {notFound && phone.trim().length >= 10 && (
          <div className="mt-3 rounded-xl bg-white p-4 shadow-sm">
            <p className="mb-3 rounded-lg bg-blue-50 p-3 text-sm text-blue-800">
              ما مشترى من عدنا — اكتب معلوماته كاملة وراح ينحفظ بالنظام.
            </p>
            <div className="grid gap-3 sm:grid-cols-2">
              <input value={newName} onChange={(e) => setNewName(e.target.value)}
                placeholder="الاسم الكامل *"
                className="rounded-lg border border-slate-300 px-3 py-2 text-right text-sm outline-none focus:border-amber-500" />
              <input value={position} onChange={(e) => setPosition(e.target.value)}
                placeholder="المنصب (مثال: مدير شركة)"
                className="rounded-lg border border-slate-300 px-3 py-2 text-right text-sm outline-none focus:border-amber-500" />
              <input value={newLocation} onChange={(e) => setNewLocation(e.target.value)}
                placeholder="العنوان"
                className="rounded-lg border border-slate-300 px-3 py-2 text-right text-sm outline-none focus:border-amber-500" />
              <input value={newLocationUrl} onChange={(e) => setNewLocationUrl(e.target.value)} dir="ltr"
                placeholder="رابط الموقع (كوكل ماب)"
                className="rounded-lg border border-slate-300 px-3 py-2 text-right text-sm outline-none focus:border-amber-500" />
              <input value={summary} onChange={(e) => setSummary(e.target.value)}
                placeholder="شنو يخصه / وين تعرفنا عليه"
                className="rounded-lg border border-slate-300 px-3 py-2 text-right text-sm outline-none focus:border-amber-500" />
              <input value={note} onChange={(e) => setNote(e.target.value)}
                placeholder="ملاحظة (اختياري)"
                className="rounded-lg border border-slate-300 px-3 py-2 text-right text-sm outline-none focus:border-amber-500" />
            </div>

            <label className="mt-3 flex items-center gap-2 text-sm text-slate-700">
              <input type="checkbox" checked={boughtFromUs} onChange={(e) => setBoughtFromUs(e.target.checked)} />
              مشترى من عدنا
            </label>

            <button onClick={addManual} disabled={adding || !newName.trim()}
              className="mt-3 rounded-lg bg-amber-600 px-5 py-2.5 text-sm font-bold text-white hover:bg-amber-700 disabled:opacity-50">
              {adding ? 'جاري الإضافة...' : '⭐ أضفه شخصية مهمة'}
            </button>
          </div>
        )}

        {found && (
          <div className="mt-3 rounded-xl bg-white p-4 shadow-sm">
            <p className="font-bold text-slate-800">{found.name}</p>
            <p className="mt-1 text-sm text-slate-600">📞 {found.phone} · كود: {found.code}</p>
            {found.location && <p className="mt-1 text-sm text-slate-600">📍 {found.location}</p>}
            <p className="mt-1 text-xs text-slate-400">
              حجوزاته السابقة: {found.previousBookingsCount ?? 0}
              {found.services?.length ? ` · خدماته: ${found.services.join('، ')}` : ''}
            </p>

            <div className="mt-3 grid gap-3 sm:grid-cols-2">
              <input value={position} onChange={(e) => setPosition(e.target.value)}
                placeholder="منصب الزبون (مثال: مدير شركة)"
                className="rounded-lg border border-slate-300 px-3 py-2 text-right text-sm outline-none focus:border-amber-500" />
              <input value={summary} onChange={(e) => setSummary(e.target.value)}
                placeholder="شنو طلب من عدنا (مثال: شد كاميرات)"
                className="rounded-lg border border-slate-300 px-3 py-2 text-right text-sm outline-none focus:border-amber-500" />
            </div>
            <input value={note} onChange={(e) => setNote(e.target.value)} placeholder="ملاحظة (اختياري)"
              className="mt-3 w-full rounded-lg border border-slate-300 px-3 py-2 text-right text-sm outline-none focus:border-amber-500" />
            <label className="mt-3 flex items-center gap-2 text-sm text-slate-700">
              <input type="checkbox" checked={boughtFromUs} onChange={(e) => setBoughtFromUs(e.target.checked)} />
              مشترى من عدنا
            </label>

            <button onClick={addManual} disabled={adding}
              className="mt-3 rounded-lg bg-amber-600 px-5 py-2.5 text-sm font-bold text-white hover:bg-amber-700 disabled:opacity-50">
              {adding ? 'جاري الإضافة...' : '⭐ علّمه شخصية مهمة'}
            </button>
          </div>
        )}
      </div>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && <p className="mt-4 rounded-lg bg-red-50 p-4 text-red-600">{error}</p>}

      {!loading && !error && (
        <div className="mt-6 overflow-x-auto rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <table className="w-full min-w-[760px] text-sm">
            <thead>
              <tr className="border-b border-slate-100 text-right text-slate-400">
                <th className="px-4 py-3">الزبون</th>
                <th className="px-4 py-3">رقم الهاتف</th>
                <th className="px-4 py-3">المنصب</th>
                <th className="px-4 py-3">مشترى من عدنا</th>
                <th className="px-4 py-3">شنو طلب</th>
                <th className="px-4 py-3">رمز الحجز</th>
                <th className="px-4 py-3">علّمه</th>
                <th className="px-4 py-3">التاريخ</th>
                <th className="px-4 py-3">إجراءات</th>
              </tr>
            </thead>
            <tbody>
              {items.map((v) => (
                <tr key={v.id} className="border-b border-slate-50 last:border-0">
                  <td className="px-4 py-3 font-bold text-brand-900">⭐ {v.customerName}</td>
                  <td className="px-4 py-3 font-bold text-brand-700" dir="ltr">{v.customerPhone}</td>
                  <td className="px-4 py-3">{v.customerPosition || '—'}</td>
                  <td className="px-4 py-3">
                    {v.boughtFromUs === false
                      ? <span className="rounded-full bg-slate-100 px-2 py-0.5 text-xs text-slate-600">لا</span>
                      : <span className="rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-bold text-emerald-700">إي</span>}
                  </td>
                  <td className="px-4 py-3">{v.requestSummary || '—'}</td>
                  <td className="px-4 py-3 font-mono text-slate-500">{v.bookingCode || '—'}</td>
                  <td className="px-4 py-3">{v.markedByName}</td>
                  <td className="px-4 py-3 text-slate-400">{new Date(v.createdAt).toLocaleDateString('ar-IQ')}</td>
                  <td className="px-4 py-3">
                    <button
                      onClick={() => remove(v.customerId)}
                      disabled={busy === v.customerId}
                      className="rounded-lg bg-red-50 px-3 py-1.5 text-xs font-bold text-red-700 hover:bg-red-100 disabled:opacity-50"
                    >
                      إزالة التعليم
                    </button>
                  </td>
                </tr>
              ))}
              {items.length === 0 && (
                <tr>
                  <td colSpan={9} className="px-4 py-6 text-center text-slate-400">
                    ما اكو زبائن معلّمين كشخصيات مهمة بعد.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
