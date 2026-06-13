import { useEffect, useState } from 'react'
import { api, type Service } from '../api'

export default function Services() {
  const [services, setServices] = useState<Service[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [name, setName] = useState('')
  const [category, setCategory] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const load = () => {
    setLoading(true)
    api
      .getServices()
      .then(setServices)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    try {
      await api.createService({ name, category: category || undefined })
      setName('')
      setCategory('')
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">الخدمات</h2>
      <p className="mt-1 text-slate-500">قائمة الخدمات التي تقدمها الشركة.</p>

      <form
        onSubmit={handleAdd}
        className="mt-6 grid grid-cols-1 gap-4 rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)] sm:grid-cols-3"
      >
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">اسم الخدمة</label>
          <input
            required
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">التصنيف (اختياري)</label>
          <input
            value={category}
            onChange={(e) => setCategory(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>
        <div className="flex items-end">
          <button
            type="submit"
            disabled={submitting}
            className="w-full rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-2 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30 disabled:opacity-50"
          >
            {submitting ? 'جاري الحفظ...' : 'إضافة خدمة'}
          </button>
        </div>
      </form>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && (
        <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">
          تعذر الاتصال بالخادم: {error}
        </p>
      )}

      {!loading && !error && (
        <div className="mt-6 grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {services.map((service) => (
            <div
              key={service.id}
              className="rounded-lg border border-white bg-white px-4 py-3 shadow-[0_4px_20px_rgba(15,32,64,0.06)]"
            >
              <span className="font-medium text-brand-800">{service.name}</span>
              {service.category && (
                <span className="mr-2 text-sm text-slate-400">({service.category})</span>
              )}
            </div>
          ))}
          {services.length === 0 && <p className="text-slate-400">لا توجد خدمات بعد.</p>}
        </div>
      )}
    </div>
  )
}
