import { lazy, Suspense } from 'react'

// ═══ مكتب الجودة — قرار (ع) 10-16 ═══
// «الجودة والشكاوى» طلعت من مكتب المراقب لمكتبها جوّا العمل — نفس المكوّنات.
const QualityAuditView = lazy(() => import('../components/QualityAuditView'))
const ComplaintsPage = lazy(() => import('./ComplaintsPage'))

export default function QualityDeskPage() {
  return (
    <div className="space-y-4" dir="rtl">
      <div className="rounded-2xl bg-gradient-to-l from-[#0f2040] to-[#2c5aad] p-5 text-white shadow">
        <h1 className="text-2xl font-black">⭐ مكتب الجودة</h1>
        <p className="mt-1 text-sm text-blue-100">تواصل مهندس الجودة ويا الزبون وحكمه، والشكاوى الواردة.</p>
      </div>
      <Suspense fallback={<p className="py-10 text-center text-slate-400">جاري التحميل…</p>}>
        <QualityAuditView />
        <ComplaintsPage />
      </Suspense>
    </div>
  )
}
