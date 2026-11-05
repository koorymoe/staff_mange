import { Suspense, type ReactNode } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { useSession } from '../session'
import { isPathAllowed } from './navTree'

// ═══ ٧و (دازاي 11-05): الشاشات المكررة تصير شاشة وحدة بتبويبات ═══
// كل تبويب = شاشته القديمة نفسها (ما ننسخ محتوى)، ومساره القديم يبقى هو
// رابطه: فالقائمة والروابط الداخلية وحارس الرابط ما يتغيّرون، والتبويب يطلع
// بس للي يگدر يفتح مساره اليوم (isPathAllowed — نفس حارس القائمة).
export type HubTab = { path: string; label: string; hint?: string; render: () => ReactNode }

export default function TabbedHub({ tabs }: { tabs: HubTab[] }) {
  const { employee, permissions, gpsServiceId } = useSession()
  const { pathname } = useLocation()
  const navigate = useNavigate()
  const ctx = { employee, permissions, gpsServiceId }
  const here = tabs.find((t) => pathname.replace(/\/+$/, '').endsWith(t.path)) ?? tabs[0]
  // التبويب الحالي يطلع دائماً (حارس المسار سمحله يوصل)، والباقي بس لو مسموح
  const visible = tabs.filter((t) => t === here || employee?.actualRole === 'OWNER' || isPathAllowed(t.path, ctx) !== false)
  return (
    <div dir="rtl" className="space-y-4">
      {visible.length > 1 && (
        <div role="tablist" className="flex flex-col gap-2 sm:flex-row">
          {visible.map((t) => {
            const on = t === here
            return (
              <button key={t.path} type="button" role="tab" aria-selected={on}
                onClick={() => { if (!on) navigate(t.path) }}
                className={`flex-1 rounded-2xl px-4 py-3 text-right transition ${on ? 'bg-[#0f2040] text-white shadow' : 'bg-white text-slate-700 ring-1 ring-slate-200 hover:bg-slate-50'}`}>
                <span className="block text-base font-extrabold">{t.label}</span>
                {t.hint && <span className={`block text-xs ${on ? 'text-blue-100' : 'text-slate-400'}`}>{t.hint}</span>}
              </button>
            )
          })}
        </div>
      )}
      <Suspense fallback={<p className="py-10 text-center text-slate-400">جاري التحميل…</p>}>
        {here.render()}
      </Suspense>
    </div>
  )
}
