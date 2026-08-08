import { useState, useEffect } from 'react'
import { api, type GpsCustomer } from '../../api'
import { matches } from '../../utils/search'

export default function GpsCustomers() {
  const [customers, setCustomers] = useState<GpsCustomer[]>([])
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
    api.getGpsCustomers()
      .then(setCustomers)
      .catch((e) => setError(e instanceof Error ? e.message : 'حدث خطأ'))
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
    matches([c.fullName, c.phone], search)
  )

  return (
    <div dir="rtl">
      {/* Header */}
      <div className="mb-6 rounded-2xl p-6 shadow-sm" style={{ backgroundColor: '#1a3a5c' }}>
        <h1 className="text-2xl font-bold text-white">👥 إدارة الزبائن</h1>
        <p className="mt-1 text-blue-200 text-sm">إضافة وإدارة زبائن نظام GPS</p>
      </div>

      {/* Search + Add */}
      <div className="mb-6 flex flex-wrap items-center gap-4">
        <div className="relative flex-1">
          <input
            placeholder="🔍 بحث بالاسم أو الهاتف..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full rounded-2xl border border-slate-200 bg-white px-5 py-3 shadow-sm outline-none focus:border-blue-400"
          />
        </div>
        <button
          onClick={() => setShowForm(!showForm)}
          className="rounded-2xl px-6 py-3 text-sm font-bold text-white shadow-sm transition-all hover:opacity-90"
          style={{ backgroundColor: '#1a3a5c' }}
        >
          {showForm ? '✕ إلغاء' : '➕ إضافة زبون جديد'}
        </button>
      </div>

      {/* Add Form Modal */}
      {showForm && (
        <div className="mb-6 rounded-2xl bg-white p-6 shadow-sm">
          <h3 className="mb-4 text-lg font-bold" style={{ color: '#1a3a5c' }}>➕ زبون جديد</h3>
          <form onSubmit={handleSubmit} className="grid grid-cols-1 gap-4 sm:grid-cols-3">
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">الاسم الكامل *</label>
              <input required value={fullName} onChange={(e) => setFullName(e.target.value)} className="w-full rounded-xl border border-slate-200 px-4 py-3 outline-none focus:border-blue-400" />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">اسم الأب</label>
              <input value={fatherName} onChange={(e) => setFatherName(e.target.value)} className="w-full rounded-xl border border-slate-200 px-4 py-3 outline-none focus:border-blue-400" />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">اسم الجد</label>
              <input value={grandfatherName} onChange={(e) => setGrandfatherName(e.target.value)} className="w-full rounded-xl border border-slate-200 px-4 py-3 outline-none focus:border-blue-400" />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">الهاتف *</label>
              <input required value={phone} onChange={(e) => setPhone(e.target.value)} className="w-full rounded-xl border border-slate-200 px-4 py-3 outline-none focus:border-blue-400" />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">العنوان</label>
              <input value={address} onChange={(e) => setAddress(e.target.value)} className="w-full rounded-xl border border-slate-200 px-4 py-3 outline-none focus:border-blue-400" />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">المحافظة</label>
              <input value={governorate} onChange={(e) => setGovernorate(e.target.value)} className="w-full rounded-xl border border-slate-200 px-4 py-3 outline-none focus:border-blue-400" />
            </div>
            <div className="sm:col-span-3">
              <button type="submit" disabled={submitting} className="rounded-2xl px-8 py-3 text-sm font-bold text-white shadow-sm disabled:opacity-50" style={{ backgroundColor: '#1a3a5c' }}>
                {submitting ? 'جاري الحفظ...' : '💾 حفظ الزبون'}
              </button>
            </div>
          </form>
        </div>
      )}

      {loading && <div className="flex items-center justify-center p-12"><div className="text-lg text-slate-500">جاري التحميل...</div></div>}
      {error && <div className="rounded-2xl bg-red-50 p-6 text-red-600 shadow-sm">{error}</div>}

      {/* Customer Cards */}
      {!loading && !error && (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {filtered.map((c) => (
            <div key={c.id} className="rounded-2xl bg-white p-5 shadow-sm transition-all hover:shadow-md">
              <div className="flex items-start justify-between">
                <div>
                  <h3 className="text-lg font-bold" style={{ color: '#1a3a5c' }}>{c.fullName}</h3>
                  {c.fatherName && <p className="text-sm text-slate-500">{c.fatherName} {c.grandfatherName || ''}</p>}
                </div>
                <span className="text-2xl">👤</span>
              </div>
              <div className="mt-3 space-y-1">
                <p className="text-sm"><span className="text-slate-500">📞 الهاتف:</span> <span className="font-medium">{c.phone}</span></p>
                {c.address && <p className="text-sm"><span className="text-slate-500">📍 العنوان:</span> {c.address}</p>}
                {c.governorate && <p className="text-sm"><span className="text-slate-500">🏙️ المحافظة:</span> {c.governorate}</p>}
              </div>
              {c.createdAt && (
                <p className="mt-3 text-xs text-slate-400">تاريخ التسجيل: {new Date(c.createdAt).toLocaleDateString('ar-IQ')}</p>
              )}
            </div>
          ))}
          {filtered.length === 0 && (
            <div className="col-span-full rounded-2xl bg-white p-8 text-center text-slate-400 shadow-sm">لا يوجد زبائن</div>
          )}
        </div>
      )}
    </div>
  )
}
