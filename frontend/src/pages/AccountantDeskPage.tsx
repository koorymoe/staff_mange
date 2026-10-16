import { lazy, Suspense } from 'react'

// ═══ مكتب المحاسب — قرار (ع) 10-16 ═══
// «الفواتير شغل المحاسب — بدل ما موجودة بمكتب المراقب اريد فدشي اسمه مكتب
// المحاسب جوّا العمل». نفس شاشة الفواتير (ما ننسخها)، وتطلع لكل من عنده
// صلاحية الحسابات — مو لأن دوره مراقب.
const LeaderInvoicesListPage = lazy(() => import('./LeaderInvoicesListPage'))

export default function AccountantDeskPage() {
  return (
    <div className="space-y-4" dir="rtl">
      <div className="rounded-2xl bg-gradient-to-l from-[#0f2040] to-[#2c5aad] p-5 text-white shadow">
        <h1 className="text-2xl font-black">🧾 مكتب المحاسب</h1>
        <p className="mt-1 text-sm text-blue-100">فواتير الليدرية: تدقيق، اعتماد، ربط رقم الفاتورة المحاسبية، وتعديل المبالغ.</p>
      </div>
      <Suspense fallback={<p className="py-10 text-center text-slate-400">جاري التحميل…</p>}>
        <LeaderInvoicesListPage embedded />
      </Suspense>
    </div>
  )
}
