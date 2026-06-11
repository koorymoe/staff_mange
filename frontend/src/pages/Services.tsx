import { useEffect, useState } from 'react'
import { api, type Service } from '../api'

export default function Services() {
  const [services, setServices] = useState<Service[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    api
      .getServices()
      .then(setServices)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">الخدمات</h2>
      <p className="mt-1 text-slate-500">قائمة الخدمات التي تقدمها الشركة.</p>

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
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
