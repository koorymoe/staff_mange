import { useState, useEffect } from 'react'
import { api } from '../../api'

export default function GpsCustomers() {
  const [customers, setCustomers] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [showForm, setShowForm] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  const [fullName, setFullName] = useState('')
  const [fatherName, setFatherName] = useState('')
  const [grandfatherName, setGrandfatherName] = useState('')
  const [phone, setPhone] = useState('')
  const [address, setAddress] = useState('')
  const [governorate, setGovernorate] = useState('')

  const load = () => {
    setLoading(true)
    api.getGpsCustomers()
      .then(setCustomers)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    try {
      await api.createGpsCustomer({ fullName, fatherName, grandfatherName, phone, address, governorate })
      setFullName(''); setFatherName(''); setGrandfatherName(''); setPhone(''); setAddress(''); setGovernorate('')
      setShowForm(false)
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  const filtered = customers.filter((c) =>
    (c.fullName || '').includes(search) || (c.phone || '').includes(search)
  )

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">زبائن GPS</h2>

      <div className="mt-4 flex flex-wrap items-center gap-4">
        <input
          placeholder="بحث بالاسم أو الهاتف..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-64 rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
        />
        <button
          onClick={() => setShowForm(!showForm)}
          className="rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md transition-all hover:shadow-lg"
        >
          {showForm ? 'إلغاء' : 'إضافة زبون جديد'}
        </button>
      </div>

      {showForm && (
        <form onSubmit={handleSubmit} className="mt-4 grid grid-cols-1 gap-4 rounded-xl bg-white p-6 shadow-md sm:grid-cols-3">
          <div>
            <label className="mb-1 block text-sm font-medium text-slate-600">الاسم الكامل</label>
            <input required value={fullName} onChange={(e) => setFullName(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500" />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-slate-600">اسم الأب</label>
            <input value={fatherName} onChange={(e) => setFatherName(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500" />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-slate-600">اسم الجد</label>
            <input value={grandfatherName} onChange={(e) => setGrandfatherName(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500" />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-slate-600">الهاتف</label>
            <input required value={phone} onChange={(e) => setPhone(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500" />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-slate-600">العنوان</label>
            <input value={address} onChange={(e) => setAddress(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500" />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-slate-600">المحافظة</label>
            <input value={governorate} onChange={(e) => setGovernorate(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500" />
          </div>
          <div className="sm:col-span-3">
            <button type="submit" disabled={submitting} className="rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md disabled:opacity-50">
              {submitting ? 'جاري الحفظ...' : 'حفظ الزبون'}
            </button>
          </div>
        </form>
      )}

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">تعذر الاتصال بالخادم: {error}</p>}

      {!loading && !error && (
        <div className="mt-6 overflow-hidden rounded-xl bg-white shadow-md">
          <table className="w-full text-right">
            <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
              <tr>
                <th className="px-4 py-3 text-sm font-semibold">الاسم الكامل</th>
                <th className="px-4 py-3 text-sm font-semibold">الهاتف</th>
                <th className="px-4 py-3 text-sm font-semibold">المحافظة</th>
                <th className="px-4 py-3 text-sm font-semibold">التاريخ</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {filtered.map((c) => (
                <tr key={c.id} className="hover:bg-slate-50">
                  <td className="px-4 py-3 font-medium">{c.fullName}</td>
                  <td className="px-4 py-3 text-slate-600">{c.phone}</td>
                  <td className="px-4 py-3 text-slate-600">{c.governorate || '-'}</td>
                  <td className="px-4 py-3 text-slate-500 text-sm">{c.createdAt ? new Date(c.createdAt).toLocaleDateString('ar-IQ') : '-'}</td>
                </tr>
              ))}
              {filtered.length === 0 && (
                <tr><td colSpan={4} className="px-4 py-6 text-center text-slate-400">لا يوجد زبائن</td></tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
