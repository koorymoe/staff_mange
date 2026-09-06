import { useEffect, useState } from 'react'
import { api, type Service, type TrainingMaterial } from '../api'
import { useSession } from '../session'

function toEmbedUrl(url: string) {
  const yt = url.match(/(?:youtube\.com\/watch\?v=|youtu\.be\/)([\w-]+)/)
  if (yt) return `https://www.youtube.com/embed/${yt[1]}`
  return url
}

export default function TrainingPage() {
  const { employee } = useSession()
  const [services, setServices] = useState<Service[]>([])
  const [materials, setMaterials] = useState<TrainingMaterial[]>([])
  const [activeServiceId, setActiveServiceId] = useState<string | null>(null)
  const [activeMaterial, setActiveMaterial] = useState<TrainingMaterial | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!employee) return
    api.getMyTraining(employee.id).then(({ services, materials }) => {
      setServices(services)
      setMaterials(materials)
      if (services[0]) setActiveServiceId(services[0].id)
    }).finally(() => setLoading(false))
  }, [employee])

  const visibleMaterials = materials.filter(m => m.serviceId === activeServiceId)

  return (
    <div>
      <div className="rounded-2xl bg-gradient-to-l from-[#0f2040] to-[#2c5aad] p-6 text-white shadow-lg">
        <h2 className="text-2xl font-bold">مرحباً {employee?.name} 👋</h2>
        <p className="mt-1 text-blue-100">أنت الآن في مرحلة التدريب. تابع المواد التدريبية المخصصة لك أدناه.</p>
      </div>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}

      {!loading && services.length === 0 && (
        <div className="mt-6 rounded-2xl border border-amber-200 bg-amber-50 p-6 text-center text-amber-700">
          لم يتم تحديد أي تدريب لك بعد. الرجاء التواصل مع الإدارة.
        </div>
      )}

      {!loading && services.length > 0 && (
        <div className="mt-6 grid gap-6 lg:grid-cols-[240px_1fr]">
          {/* Service tabs */}
          <div className="flex flex-col gap-2">
            {services.map(s => (
              <button
                key={s.id}
                onClick={() => { setActiveServiceId(s.id); setActiveMaterial(null) }}
                className={`rounded-xl px-4 py-3 text-right text-sm font-bold transition-all ${
                  activeServiceId === s.id
                    ? 'bg-brand-500 text-white shadow-md'
                    : 'bg-white text-slate-600 border border-slate-200 hover:bg-slate-50'
                }`}
              >
                {s.name}
              </button>
            ))}
          </div>

          {/* Materials */}
          <div>
            {activeMaterial ? (
              <div className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
                <button onClick={() => setActiveMaterial(null)} className="mb-3 text-sm font-bold text-brand-600 hover:underline">
                  ← رجوع للقائمة
                </button>
                <h3 className="mb-3 text-lg font-bold text-brand-900">{activeMaterial.title}</h3>
                {activeMaterial.type === 'VIDEO' ? (
                  <div className="aspect-video overflow-hidden rounded-xl bg-black">
                    <iframe
                      src={toEmbedUrl(activeMaterial.url)}
                      className="h-full w-full"
                      allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
                      allowFullScreen
                    />
                  </div>
                ) : (
                  <a href={activeMaterial.url} target="_blank" rel="noreferrer"
                    className="inline-block rounded-lg bg-brand-500 px-4 py-2 text-sm font-bold text-white hover:bg-brand-600">
                    فتح المستند
                  </a>
                )}
              </div>
            ) : (
              <div className="flex flex-col gap-3">
                {visibleMaterials.length === 0 && (
                  <p className="rounded-xl border border-slate-200 bg-white p-6 text-center text-slate-400">
                    لا توجد مواد تدريبية لهذه الخدمة بعد.
                  </p>
                )}
                {visibleMaterials.map(m => (
                  <button key={m.id} onClick={() => setActiveMaterial(m)}
                    className="flex items-center gap-3 rounded-xl border border-slate-200 bg-white p-4 text-right shadow-sm transition hover:border-brand-300 hover:shadow-md">
                    <div className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-brand-50 text-brand-600">
                      {m.type === 'VIDEO' ? '▶' : '📄'}
                    </div>
                    <div className="flex-1">
                      <p className="font-bold text-brand-900">{m.title}</p>
                      <p className="text-xs text-slate-400">{m.type === 'VIDEO' ? 'فيديو تدريبي' : 'مستند'}</p>
                    </div>
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
