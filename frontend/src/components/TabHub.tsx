import { Suspense, type ReactNode } from 'react'
import { useSearchParams } from 'react-router-dom'

// ═══ شاشة وحدة بتبويبات — قرار (ع) 10-08 ═══
// «ماريد هذا العدد من الخيارات بالقائمة — بواجهة وحدة ومن الواجهة نتنقل».
// كل تبويب يفتح نفس الشاشة القديمة بلا نسخ منطقها، والتبويب ينحفظ بالرابط (?tab=).
export type HubTab = { id: string; label: string; icon: string; show: boolean; render: () => ReactNode }

export default function TabHub({ title, subtitle, icon, tabs }: { title: string; subtitle: string; icon: string; tabs: HubTab[] }) {
  const [params, setParams] = useSearchParams()
  const shown = tabs.filter((t) => t.show)
  const cur = shown.find((t) => t.id === params.get('tab')) ?? shown[0]
  return (
    <div dir="rtl" className="space-y-4">
      <div className="rounded-2xl bg-gradient-to-l from-[#0f2040] to-[#2c5aad] p-5 text-white shadow">
        <h1 className="text-2xl font-extrabold">{icon} {title}</h1>
        <p className="mt-1 text-sm text-blue-100">{subtitle}</p>
      </div>
      <div className="flex gap-2 overflow-x-auto pb-1">
        {shown.map((t) => (
          <button key={t.id} type="button" onClick={() => setParams({ tab: t.id }, { replace: true })}
            className={`shrink-0 whitespace-nowrap rounded-xl border px-4 py-2 text-sm font-bold transition ${
              cur?.id === t.id ? 'border-[#2c5aad] bg-[#2c5aad] text-white shadow' : 'border-slate-200 bg-white text-slate-600 hover:bg-slate-50'
            }`}>
            {t.icon} {t.label}
          </button>
        ))}
      </div>
      <Suspense fallback={<p className="text-sm text-slate-400">جاري التحميل…</p>}>
        {cur ? cur.render() : <p className="rounded-xl bg-white p-6 text-center text-sm text-slate-400">ماكو شي مسموح إلك هنا.</p>}
      </Suspense>
    </div>
  )
}
