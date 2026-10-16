import { lazy, Suspense, useState } from 'react'
import { useSession } from '../session'

// ═══ مكتب الجودة — قرار (ع) 10-16 ═══
// شغلتين مختلفات: «متابعة الجودة» (نتواصل ويا الزبون ونتأكد الشغل تم وهو راضي)
// و«الشكاوى» (زبون اشتكى — مهندس الجودة يتابعها لحد ما تخلص، والمراقب يدقق).
// نفس الصفحة للمراقب ولمهندس الجودة، وكل وحدة بخيارها فوك — مو وحدة جوّا الثانية.
const QualityAuditView = lazy(() => import('../components/QualityAuditView'))
const ComplaintsPage = lazy(() => import('./ComplaintsPage'))
// (ع) 10-16: «الي يطلع للمراقب يختلف عن الي يطلع لمهندس الجودة» — المهندس
// يشتغل (يتواصل ويا الزبائن ويحكم)، والمراقب يدقق (شنو سوّى المهندس).
const QualityFollowUpsPage = lazy(() => import('./QualityFollowUpsPage'))

type View = 'quality' | 'complaints'

export default function QualityDeskPage() {
  const [view, setView] = useState<View>('quality')
  const { employee, permissions } = useSession()
  const role = employee?.role
  const isAuditor = role === 'ADMIN' || role === 'MONITOR' || (role !== 'QUALITY_ENGINEER' && permissions.includes('monitoring'))
  const btn = (v: View, label: string, hint: string) => (
    <button type="button" onClick={() => setView(v)}
      className={`flex-1 rounded-2xl px-4 py-3 text-right transition ${view === v ? 'bg-[#0f2040] text-white shadow' : 'bg-white text-slate-700 ring-1 ring-slate-200 hover:bg-slate-50'}`}>
      <span className="block text-base font-extrabold">{label}</span>
      <span className={`block text-xs ${view === v ? 'text-blue-100' : 'text-slate-400'}`}>{hint}</span>
    </button>
  )
  return (
    <div className="space-y-4" dir="rtl">
      <div className="rounded-2xl bg-gradient-to-l from-[#0f2040] to-[#2c5aad] p-5 text-white shadow">
        <h1 className="text-2xl font-black">⭐ مكتب الجودة</h1>
        <p className="mt-1 text-sm text-blue-100">متابعة رضا الزبون بعد الشغل، والشكاوى الي لازم مهندس الجودة يخلصها.</p>
      </div>
      <div className="flex flex-col gap-2 sm:flex-row">
        {btn('quality', '⭐ متابعة الجودة', 'مهندس الجودة تواصل ويا الزبون وتأكد الشغل تم وهو راضي')}
        {btn('complaints', '⚠️ الشكاوى', 'الشكاوى المفتوحة — هل مهندس الجودة تابعها؟ وشنو صار بيها')}
      </div>
      <Suspense fallback={<p className="py-10 text-center text-slate-400">جاري التحميل…</p>}>
        {view === 'quality' ? (isAuditor ? <QualityAuditView /> : <QualityFollowUpsPage embedded />) : <ComplaintsPage />}
      </Suspense>
    </div>
  )
}
