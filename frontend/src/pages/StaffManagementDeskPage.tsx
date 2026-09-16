// ═══ مكتب إدارة الموظفين ═══
//
// «أربعة بنود بالقائمة الجانبية تاخذ حيّز — أريدها بند واحد، ومن
// أدخله أشوف نفس الأربعة كتبويبات بشاشة واحدة أتنقّل بينهن».
//
// ⚠️⚠️ **المكتب يضمّ الشاشات — ما ينسخها.** كل قسم هو **الشاشة
// نفسها** بخاصية `embedded` تخفي ترويستها بس — نفس نمط `MonitorDeskPage.tsx`
// بالضبط. نسخ المحتوى يعني إن أول تصحيح يوصل نسخة وينسى الثانية.
//
// ⚠️ والشاشات القديمة **تبقى بمساراتها**: اكو روابط محفوظة بمتصفحات
// الموظفين، وحذف مسار يعني صفحة بيضاء بلا تفسير.

import { lazy, Suspense, useState } from 'react'

import { isNavVisible, type NavItem } from '../components/navTree'
import { useSession } from '../session'

const Employees = lazy(() => import('./Employees'))
const KpiPage = lazy(() => import('./KpiPage'))
const PerformanceReviewPage = lazy(() => import('./PerformanceReviewPage'))
const StaffRequestsPage = lazy(() => import('./StaffRequestsPage'))

type SectionId = 'employees' | 'kpi' | 'performance' | 'staffRequests'

interface Section {
  id: SectionId
  label: string
  icon: string
  /** نفس شرط الظهور الموجود أصلاً بـ`navTree.tsx` لكل بند — حرفياً. */
  navItem: NavItem
}

// ⚠️ نسخة طبق الأصل عن عناصر `navTree.tsx:128,134,136,138` — أي
// تعديل هناك (منح/سحب صلاحية) يلزم تعديله هنا هم، حتى المكتب ما
// يفتح شي القائمة تمنعه، ولا يقفل شي القائمة تسمح فيه.
const SECTIONS: Section[] = [
  {
    id: 'employees', label: 'إدارة الكوادر', icon: '👥',
    navItem: { to: '/employees', label: 'إدارة الكوادر', icon: <></>, roles: ['ADMIN', 'MONITOR'], permission: 'staff_management' },
  },
  {
    id: 'kpi', label: 'نقاط الكي بي اي', icon: '⭐',
    navItem: { to: '/kpi', label: 'نقاط الكي بي اي', icon: <></>, roles: ['ADMIN', 'MONITOR'], permission: 'kpi_management' },
  },
  {
    id: 'performance', label: 'تقييم الأداء', icon: '📊',
    navItem: { to: '/performance-review', label: 'تقييم الأداء', icon: <></>, roles: ['ADMIN', 'MONITOR'], unlockPermission: 'performance_review' },
  },
  {
    id: 'staffRequests', label: 'طلبات الكادر', icon: '👷',
    navItem: { to: '/staff-requests', label: 'طلبات الكادر', icon: <></>, roles: ['ADMIN', 'MONITOR'], unlockPermission: 'staff_requests' },
  },
]

export default function StaffManagementDeskPage() {
  const { employee, permissions, gpsServiceId } = useSession()
  const ctx = { employee, permissions, gpsServiceId }

  // ⚠️ القسم الي ما عنده صلاحيته **ما يظهر أصلاً** — نفس تحذير
  // `MonitorDeskPage.tsx`: قسم فاضٍ بلا تفسير يخلّي الموظف يظن
  // النظام مكسوراً.
  const shown = SECTIONS.filter((s) => isNavVisible(s.navItem, ctx))
  // ⚠️ بلا تهيئة بحالة `undefined` تنتظر وصول الصلاحيات: `active`
  // يبقى فارغاً لحد أول ضغطة، والسطر تحت يرجع لأول قسم ظاهر فعلياً
  // (`shown[0]`) — بلا Effect يحدّث حالة، وبلا وميض على تبويب مخفي.
  const [active, setActive] = useState<SectionId | undefined>(undefined)
  const cur = shown.find((s) => s.id === active) ?? shown[0]

  return (
    <div dir="rtl" className="space-y-4">
      <div>
        <h1 className="text-2xl font-black text-[#0f2040]">🗂️ إدارة الموظفين</h1>
        <p className="mt-1 text-sm text-slate-500">
          إدارة الكوادر ونقاط الكي بي اي وتقييم الأداء وطلبات الكادر — كل شغلك بمكان واحد.
        </p>
      </div>

      {shown.length === 0 && (
        <p className="rounded-xl border border-slate-200 bg-slate-50 px-4 py-6 text-center text-sm text-slate-400">
          ماكو شاشة بهذي الوحدة تناسب صلاحياتك الحالية.
        </p>
      )}

      {shown.length > 0 && (
        <>
          <div className="flex flex-wrap gap-2">
            {shown.map((s) => (
              <button
                key={s.id}
                onClick={() => setActive(s.id)}
                className={`rounded-xl border px-4 py-2 text-sm font-bold transition-all ${
                  active === s.id
                    ? 'border-transparent bg-gradient-to-l from-brand-500 to-brand-800 text-white shadow'
                    : 'border-slate-200 bg-white text-slate-600 hover:border-brand-300 hover:text-brand-700'
                }`}
              >
                {s.icon} {s.label}
              </button>
            ))}
          </div>

          <Suspense fallback={<p className="py-16 text-center text-slate-400">جاري التحميل…</p>}>
            {cur?.id === 'employees' && <Employees embedded />}
            {cur?.id === 'kpi' && <KpiPage embedded />}
            {cur?.id === 'performance' && <PerformanceReviewPage embedded />}
            {cur?.id === 'staffRequests' && <StaffRequestsPage embedded />}
          </Suspense>
        </>
      )}
    </div>
  )
}
