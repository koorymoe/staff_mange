import { useState, useEffect } from 'react'
import { useAutoRefresh } from '../useAutoRefresh'
import PrivacyPolicyGate from './PrivacyPolicyGate'
import EmployeeAvatar from './EmployeeAvatar'
import { NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { api, type Employee, type EmployeeRole } from '../api'
import { SessionContext, roleLabels, hasGpsSkill } from '../session'
import { ensureFileToken } from '../api'
import Login from '../pages/Login'
import CommandApp from '../command/CommandApp'
import TrainingPage from '../pages/TrainingPage'
import AssistantWidget from './AssistantWidget'
import ManagerAssistantChat from './ManagerAssistantChat'
import SettingsPanel from './SettingsPanel'

// وجهة كل نوع إشعار: ضغطة على الإشعار توديك للشاشة الي تخصه بدل ما
// تدوّر عليها. نوع مو موجود بالخريطة = الإشعار يتأشر مقروء وبس.
const notifTargets: Record<string, string> = {
  leave_request: '/leaves',
  leave_preliminary: '/leaves',
  leave_decision: '/leaves',
  booking_delete_request: '/booking-delete-requests',
  booking_confirmed: '/bookings',
  booking_returned_to_crew: '/bookings',
  audit_issue: '/audit-issues',
  kpi_deduction: '/kpi',
  kpi_leaderboard: '/kpi',
  authz_violation: '/owner-security',
  job_duration_overrun: '/missions',
  // ═══ قرارات الطلبات ═══
  // الإشعار بلا وجهة يخلي الموظف يقرا «انوافق على طلبك» ويضغط عليه
  // وما يصير شي — فيضطر يدوّر على الشاشة بنفسه.
  tool_request_decision: '/my-inventory',
  procurement_decision: '/procurement',
  staff_request_decision: '/staff-requests',
  booking_delete_decision: '/bookings',
  employee_letter_decision: '/letters',
  extra_task: '/my-extra-tasks',
  extra_task_done: '/extra-tasks',
  extra_task_cancelled: '/my-extra-tasks',
}
import AnnouncementTicker from './AnnouncementTicker'

export interface NavItem {
  to: string
  label: string
  icon: React.ReactNode
  end?: boolean
  roles?: EmployeeRole[]
  permission?: string
  anyPermission?: string[]
  leaderOnly?: boolean
  gpsSkillOnly?: boolean
  // صلاحية ظهور الوحدة كاملة: منحها للموظف يفتحله الوحدة وكل صفحاتها،
  // بغض النظر عن صلاحياته التفصيلية — هذا معنى "أنطيه الصلاحية ويشوف".
  unitPermission?: string
  // شغل الميدان مو إدارة: العنصر ينحجب عن الفني والتيم ليدر مهما جانت
  // صلاحياتهم. الليدر عنده صلاحيات إدارية (مشاريع، طلبات مواد) بس محلها
  // مجموعة «العمل» مالته — مو باب «الإدارة».
  hideFromFieldStaff?: boolean
  // unlockPermission: صلاحية **تفتح** العنصر وما تقيّده أبداً.
  //
  // الفرق عن permission: هذيچ شرط (ما عندك الصلاحية = ما تشوف)، وهاي
  // مفتاح (عندك الصلاحية = تشوف، مهما كان دورك). كل عنصر جان إله roles
  // بس ما جان ينفتح بأي منح — وهاي المشكلة الي شكه منها صاحب العمل:
  // «صلاحية من أنطيها لأحد يلا تظهر إله».
  unlockPermission?: string
  // ownerOnly: الاستثناء الوحيد الي المنح ما يكسره — شاشات المالك.
  ownerOnly?: boolean
  // hideForRoles: منع صريح لدور معيّن — يشتغل **قبل** كل شي، حتى قبل
  // «الصلاحية الممنوحة تفتح العنصر». استعماله الوحيد: نفس الشاشة
  // موجودة بمحل أنسب لهذا الدور، فما نريدها تتكرر بالقائمة مرتين.
  hideForRoles?: EmployeeRole[]
  // fieldStaffOnly: عكس hideFromFieldStaff — صندوق شغل الميدان. ما يطلع
  // لأي دور مكتبي مهما انمنحت له صلاحيات، لأن الصلاحية الوحدة (مثل
  // execution_cost) جانت تفتح أبناءه فيطلع للمحاسب «العمل» **مرتين**:
  // واحدة مالته وواحدة مال الفني. مدير النظام ما ينتأثر.
  fieldStaffOnly?: boolean
  children?: NavItem[]
  divider?: boolean
  /** عنوان قسم بلا خطوط جانبية — «التنقل الرئيسي» بأعلى القائمة */
  plain?: boolean
}

const I = ({ d }: { d: string }) => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d={d} /></svg>
)

// ═══ قائمة الفني العادي ═══
// الفني ما يشوف إلا شغله: الرئيسية، الحضور، الإجازات، تصنيفي، جرد
// أدواته، ومهامه. كل شي غير هذا يختفي عنه — حتى لو انفتح بصلاحية
// جماعية أو انضاف عنصر جديد للقائمة بعدين.
const TECHNICIAN_NAV = [
  '/', '/attendance', '/leaves', '/my-ranking', '/my-tasks', '/my-extra-tasks', '/my-inventory', '/privacy-policy',
]

export const navItems: NavItem[] = [
  // عنوان القسم — يفصل التنقل عن رأس القائمة (الشعار وبطاقة الموظف)،
  // فالعين تعرف وين تبدي بدل ما تلگه كتلة أزرار ملزوقة بالبطاقة.
  { to: '/nav-main-label', label: 'التنقل الرئيسي', icon: <></>, divider: true, plain: true },
  { to: '/', label: 'الرئيسية', end: true, icon: <I d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z M9 22V12h6v10" /> },
  { to: '/attendance', label: 'جدول دوامي', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg> },
  // ⚠️ مهامي الإضافية بلا قيد دور ولا صلاحية — هاي **مهام الموظف
  // نفسه**، مو مهام غيره. كانت محصورة بشاشة «مهامي» (فنيين فقط)،
  // يعني المدير يوجّه مهمة لإداري أو محاسب وما توصله إلا كإشعار
  // يضيع. المهمة الي ما إلها مكان ثابت تنعرض بيه تنتنسى.
  //
  // أما **توجيه** المهام لغيره فيحتاج صلاحية extra_tasks_assign.
  { to: '/my-extra-tasks', label: 'مهامي الإضافية', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M9 5H7a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2h-2"/><rect x="9" y="3" width="6" height="4" rx="1"/><path d="m9 14 2 2 4-4"/></svg> },
  // ⚠️ «حساب الكلفة» انشال من القائمة العلوية: كان يطلع مرتين —
  // مرة فوگ ومرة داخل «العمل» بعنوانين مختلفين لنفس الشاشة، فالليدر
  // يحتار أي وحدة يفتح.
  //
  // صار مدخل واحد داخل «العمل»، والشاشة نفسها تسأل: استفسار لو
  // مربوط بحجز. الفرق قرار **جوّا الشاشة** مو رابطين بالقائمة.
  // الإجازات: مع نظام الحضور — أي موظف يقدّم طلبه من هنا، والمخوّل يشوف صندوق الموافقات
  // تصنيفي: صفحة شخصية عامة لكل الأدوار — لازم تبقى بمستوى مستقل بره "الإدارة"،
  // لأنه الفني/الليدر ما عندهم وصول لأي شي ثاني بالإدارة، فتضل قائمة فاضية
  // بالنسبة الهم لو حطيناها جوه.
  // مجموعة «تصنيفي»: تصنيف الموظف نفسه، وتحته تقييم فريقه إذا كان ليدر.
  // لمن ما يكون بيها إلا «تصنيفي» تنفك المجموعة وتطلع الشاشة مباشرة —
  // فباقي الأدوار ما ينتغيّر عندهم شي.
  // ═══ «التقييم» بند واحد ═══
  // كانت مجموعة تنفتح على «تصنيفي» و«تقييم الأداء». صارت بند واحد
  // يفتح الشاشة، والاختيار بين «تقييمي وتصنيفي» و«تقييم فريقي» من
  // فوگ بالواجهة — نفس نمط «مهامي» و«الجرد».
  { to: '/my-ranking', label: 'التقييم', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>,
    roles: ['ADMIN', 'SALES', 'HR_COORDINATOR', 'TECHNICIAN', 'MONITOR', 'FINANCE', 'GPS_ADMIN', 'QUALITY_ENGINEER', 'PROCUREMENT_ADMIN', 'TECHNICAL'] },

  // ── الإدارة ──
  {
    // مجموعة الإدارة بدون قيد أدوار — ظهورها يعتمد على أبنائها (كل ابن مقيّد بدوره/صلاحيته)،
    // حتى أي موظف ينمنح صلاحية إدارية (مثل إدارة المشاريع) توصله من دون تغيير دوره.
    // ما تطلع لكادر الميدان أبداً — الفني ما إله شغل بيها، والتيم ليدر
    // شغله الإداري (المشاريع، طلبات المواد) محله مجموعة «العمل» مالته.
    hideFromFieldStaff: true,
    to: '/admin-group', label: 'العمل', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9"/></svg>,
    children: [
      {
        // بدون قيد أدوار هنا أيضاً — نفس مبدأ مجموعة "الإدارة" الأعلى: قيد على المجموعة
        // الوسيطة يمنع ظهورها بالكامل حتى لو ابن معيّن مسموح لدور/صلاحية موظف غير
        // مذكور بهذي القائمة (مثال حقيقي: PROCUREMENT_ADMIN مع صلاحية "جرد الأدوات").
        to: '/mgmt-employees', label: 'إدارة الموظفين', icon: <></>,
        children: [
          // ═══ الوصول بالصلاحية مو بالدور ═══
          //
          // «إداري الكوادر» كان ياخذ هاي الشاشات لمجرد إن دوره اسمه
          // هيچ — بلا ما ينطيه أحد صلاحية. وهاي هي المشكلة نفسها الي
          // تظهر بالتصنيف: النظام يحكم بالمسمّى مو بالشغل.
          //
          // شلناه من قوائم الأدوار. صاحب الصلاحية يشوفها، والي ما عنده
          // ما يشوفها — مهما كان اسم دوره.
          //
          // ⚠️ ADMIN و MONITOR بقوا: هذني إدارة النظام نفسه، وشيلهم
          // يقفل الباب على الي ينطي الصلاحيات أصلاً.
          { to: '/employees', label: 'إدارة الكوادر', icon: <></>, roles: ['ADMIN', 'MONITOR'], permission: 'staff_management' },
          // الصلاحيات جانت مدفونة جوّا مجموعة «إدارة الصلاحيات» — يعني
          // خمس مستويات للوصول لشاشة وحدة. المجموعة انشالت والشاشتين
          // صعدن هنا مباشرة.
          { to: '/permissions', label: 'الصلاحيات', icon: <></>, roles: ['ADMIN'] },
          { to: '/permission-preview', label: '🔎 شوف بعين الموظف', icon: <></>, roles: ['ADMIN'] },
          { to: '/kpi', label: 'نقاط الكي بي اي', icon: <></>, roles: ['ADMIN', 'MONITOR'], permission: 'kpi_management' },
          // تقييم الأداء (منفصل عن الكي بي اي) — تيم ليدرات الفرق
          { to: '/performance-review', label: '⭐ تقييم الأداء', icon: <></>, roles: ['ADMIN', 'MONITOR'], unlockPermission: 'performance_review' },
          // طلبات الكادر الواردة من إدارة المشاريع
          { to: '/staff-requests', label: 'طلبات الكادر', icon: <></>, roles: ['ADMIN', 'MONITOR'], unlockPermission: 'staff_requests' },
          // ⚠️ الإجازات انشالت من القائمة العلوية: الموظف يطلبها من
          // «جدول دوامي» مباشرة. بس المدير لازم يضل يوصل صندوق الطلبات
          // حتى يوافق — بلا هذا المدخل الطلبات تنتراكم وماكو منو يشوفها.
          { to: '/leaves', label: '🗓️ طلبات الإجازات', icon: <></>, roles: ['ADMIN', 'OWNER', 'MONITOR'], anyPermission: ['leave_approve_morning', 'leave_approve_evening'] },
          { to: '/stats', label: 'إحصائيات الموظفين', icon: <></>, roles: ['ADMIN', 'MONITOR'], permission: 'staff_stats' },
          { to: '/employee-stats', label: 'إحصائيات الموظفين الشهرية', icon: <></>, roles: ['ADMIN'], unlockPermission: 'employee_stats' },
        ],
      },
      // إدارة الإحصائيات — عنصر مستقل مباشر تحت "الإدارة"، مو داخل إدارة
      // الموظفين، حصراً لمدير النظام.
      { to: '/stats-management', label: 'إدارة الإحصائيات', icon: <></>, roles: ['ADMIN'], unlockPermission: 'employee_stats' },
      // ═══ «إدارة العمل» انفكّت ═══
      // كانت مجموعة جوّا مجموعة: تضغط «الإدارة» فتنفتح «إدارة العمل»
      // فتنفتح الشاشات — ثلاث ضغطات. هسه محتوياتها مباشرة تحت «العمل».
      //
      // ═══ ودمج الحجوزات ═══
      // «الحجوزات» و«حجز جديد» و«تنسيق الحجوزات» صارن **بند واحد**،
      // والثلاثة خيارات من فوگ بالشاشة. الثلاثة نفس الشغلة: حجز
      // تسجّله، تنسّقه، وتتابعه.
      { to: '/bookings', label: '📋 الحجوزات', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR', 'FINANCE'], anyPermission: ['view_bookings', 'coordinator', 'sales_booking'], unlockPermission: 'view_bookings' },
      { to: '/customers', label: 'العملاء', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'], permission: 'manage_customers' },
      // ⚠️ «📅 الحجوزات المؤجلة» انشالت: نفس حجوزاتها تطلع بسلّة
      // «حجوزات مؤجّلة» جوّا «ما وصلت للتنفيذ» بشاشة الحجوزات، مع
      // نفس إمكانية تحديد موعد جديد.
      { to: '/bookings-archive', label: 'أرشيف الحجوزات', icon: <></>, roles: ['ADMIN', 'MONITOR'], permission: 'bookings_archive' },
      // ⚠️ «🔄 حجوزات تحتاج إكمال» انشالت: صارت محطة جوّا شاشة
      // الحجوزات بين «مكلّف» و«تم الإنجاز» — «ما أريدها بالقائمة
      // الجانبية، أريدها تطلع بجانب الحجوزات وحجوزات بانتظار
      // التثبيت».
      // ⚠️ «ما وصلت للتنفيذ» انشالت من القائمة: نفس الحجوزات تطلع
      // بشاشة «الحجوزات» بفلاترها، وبند ثاني إلها يعني نفس الحجز
      // بمكانين والإداري يشتغل على وحدة وينسى الثانية.
      // توجيه شغل لموظف — نفس صلاحية إدارة الكوادر
      { to: '/extra-tasks', label: '📋 توجيه المهام الإضافية', icon: <></>, roles: ['ADMIN'], permission: 'extra_tasks_assign' },
      // ⚠️ المالك ومدير النظام بس — تحليل سلوك موظف بيد زميله يتحول لسلاح.
      { to: '/ai-insights', label: '🧠 مؤشرات الذكاء الاصطناعي', icon: <></>, roles: ['ADMIN'] },
      // دليل الأدوار — يوضّح منو يوصل لوين، فمحله عند من يوزّع الصلاحيات
      { to: '/roles-guide', label: '📋 دليل الأدوار والصلاحيات', icon: <></>, roles: ['ADMIN'] },
      { to: '/solar', label: '☀️ الطاقة الشمسية', icon: <></>, roles: ['ADMIN', 'OWNER', 'MONITOR', 'TECHNICIAN', 'SERVICE_MANAGER'], permission: 'solar_system' },
      { to: '/training-programs', label: '🎓 برامج التدريب', icon: <></>, roles: ['ADMIN', 'OWNER'], permission: 'training_manage' },
          { to: '/missions', label: 'تتبع المهام', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'], permission: 'mission_tracking' },
          // الشكاوى ومتابعة الجودة جانن تحت «إدارة الموظفين» — وهنّ شغل
          // على الزبون مو على ملف الموظف. محلهن هنا مع باقي شغل العمل.
          { to: '/complaints', label: '⚠️ الشكاوى', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'], permission: 'complaints' },
          { to: '/quality-follow-ups', label: 'متابعة الجودة', icon: <></>, roles: ['ADMIN', 'MONITOR', 'QUALITY_ENGINEER'], permission: 'quality_control' },
          // شاشة البت بطلبات حذف الحجوزات. كانت مدفونة جوّا وحدة العلاقات
          // العامة، ووحدة كاملة تنحجب عن أي واحد ما عنده صلاحية الوحدة —
          // فالمالك كان يوصله إشعار الطلب وما يلكه مكان يوافق بيه. محلها
          // المنطقي هنا: هي قرار على حجز.
      { to: '/booking-delete-requests', label: '🗑️ طلبات حذف الحجوزات', icon: <></>, roles: ['ADMIN', 'OWNER', 'MONITOR'], permission: 'booking_delete_approve' },
      {
        // الجي بي اس صارت خدمة بتحكم صلاحية "gps_system" — مو دور وظيفي منفصل،
        // فأي موظف عنده هذي الصلاحية (مسؤول خدمة الجي بي اس أو المراقب) يشوفها.
        to: '/mgmt-services', label: 'إدارة الخدمات', icon: <></>,
        children: [
          { to: '/gps', label: 'نظام GPS', icon: <></>, permission: 'gps_system' },
          { to: '/gps/requests', label: 'طلبات GPS المعلقة', icon: <></>, permission: 'gps_system' },
          // متابعة التجديد تخص مسؤول الجي بي اس ومهندس الجودة سوا
          { to: '/gps/follow-up', label: '🔄 متابعة تجديد الاشتراكات', icon: <></>, anyPermission: ['gps_system', 'quality_control'] },
          // شاشة الشرائح كانت موجودة بالمسارات بس بلا رابط بالقائمة —
          // يعني تحرير الشريحة وحرقها ما كان يوصلهن أحد إلا بكتابة الرابط
          { to: '/gps/sims', label: '📶 شرائح GPS', icon: <></>, permission: 'gps_system' },
          { to: '/gps/renewals-review', label: 'طلبات تجديد GPS', icon: <></>, permission: 'gps_system' },
          { to: '/gps/maintenance-review', label: 'طلبات صيانة GPS', icon: <></>, permission: 'gps_system' },
          { to: '/service-managers', label: 'مسؤولو الخدمات', icon: <></>, roles: ['ADMIN'], unlockPermission: 'service_managers' },
          // أسعار الشبكات — تنعدّل من الشاشة لأنها لسه تتبني وتتغيّر
          { to: '/network-prices', label: 'أسعار الشبكات', icon: <></>, roles: ['ADMIN'], unlockPermission: 'network_prices' },
        ],
      },
      {
        // إدارة المشاريع صارت صلاحية: أي موظف عنده project_management يشوفها بغض النظر عن دوره
        to: '/mgmt-projects', label: 'إدارة المشاريع', icon: <></>,
        children: [
          { to: '/projects', label: '🏗️ المشاريع', icon: <></>, anyPermission: ['project_management', 'project_create_only'], unlockPermission: 'project_management' },
          { to: '/project-work-types', label: 'إعدادات: أنواع الأعمال', icon: <></>, permission: 'project_management' },
          { to: '/project-statistics', label: '📊 إحصائيات المشاريع', icon: <></>, permission: 'project_management' },
      { to: '/checklists', label: 'الكشوفات', icon: <></>, permission: 'project_management' },
          { to: '/staff-requests', label: 'طلبات الكادر', icon: <></>, permission: 'project_management' },
          { to: '/quotations', label: 'عروض الأسعار', icon: <></>, anyPermission: ['quotation_create', 'quotation_edit_own', 'quotation_manage_all', 'quotation_system'] },
          { to: '/products', label: '📦 المنتجات', icon: <></>, anyPermission: ['quotation_manage_all', 'quotation_system'] },
        ],
      },
      {
        to: '/mgmt-finance', label: 'إدارة الحسابات', icon: <></>,
        // ⚠️ محجوبة عن المحاسب: محتواها **نفسه** محتوى «العمل» مالته
        // كلمة بكلمة، فكان يشوف نفس التسع شاشات مرتين — مرة بالإدارة
        // ومرة بالعمل. «ماريد الأوامر تتكرر… أحتاج واحد مو ثنينهن».
        // للمدير والمراقب تبقى مثل ما هي: هذول ما عندهم «العمل» مال
        // المحاسب أصلاً، فما يتكرر عندهم شي.
        hideForRoles: ['FINANCE'],
        children: [
          { to: '/finance', label: 'تدقيق الحسابات', icon: <></>, roles: ['ADMIN', 'FINANCE', 'MONITOR'], permission: 'finance' },
      { to: '/daily-audit', label: '📅 التدقيق اليومي', icon: <></>, roles: ['ADMIN', 'FINANCE', 'MONITOR'], permission: 'finance' },
          // فواتير الليدر تترحّل للمحاسب بتفاصيلها حتى يدققها ويعتمدها
          { to: '/revolving-fund', label: '💵 الدوار', icon: <></>, permission: 'revolving_fund' },
      { to: '/audit-issues', label: '💸 بلاغات أخطاء التدقيق', icon: <></>, roles: ['ADMIN', 'MONITOR', 'QUALITY_ENGINEER', 'FINANCE'], unlockPermission: 'audit_issues' },
      // موجودة بالقائمة الرئيسية كمان — منحطة هنا لأن محلها المنطقي الحسابات
      { to: '/leader-invoices/new', label: '🧮 حساب الكلفة', icon: <></>, permission: 'execution_cost' },
          { to: '/gps-install-costs', label: '🔧 حساب تكاليف الشد', icon: <></>, roles: ['ADMIN', 'FINANCE'], unlockPermission: 'gps_install_costs' },
          // شاشة مراجعة كل الفواتير — للمحاسب والمراقب والمدير والمالك.
          // الليدر إله بنده الخاص تحت (يشوف فواتيره هو بس).
          { to: '/leader-invoices', label: '🧾 فواتير الليدر', icon: <></>, roles: ['ADMIN', 'FINANCE', 'MONITOR'], unlockPermission: 'leader_invoices_view' },
          { to: '/expenses', label: 'إدارة المصاريف', icon: <></>, roles: ['ADMIN', 'FINANCE'], unlockPermission: 'expenses_manage' },
        ],
      },
      {
        to: '/mgmt-procurement', label: 'المشتريات والمخازن', icon: <></>,
        children: [
          { to: '/procurement', label: 'طلبات المواد', icon: <></>, roles: ['ADMIN', 'MONITOR', 'PROJECT_MANAGER', 'TECHNICIAN', 'PROCUREMENT_ADMIN'], permission: 'procurement' },
          // جرد الأدوات جان تحت «إدارة الموظفين» — وهو مخزن مو ملف موظف
          { to: '/inventory', label: 'جرد الأدوات', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR', 'PROCUREMENT_ADMIN'], permission: 'inventory' },
          { to: '/suppliers', label: 'الموردون', icon: <></>, anyPermission: ['suppliers_management'] },
        ],
      },
      {
        // إدارة المركبات ما تظل معزولة بره قائمة الإدارة — صارت مجموعة فرعية هنا
        to: '/mgmt-vehicles', label: 'المركبات والأسطول', icon: <></>,
        children: [
          { to: '/vehicles', label: 'إدارة المركبات', icon: <></>, permission: 'vehicle_management' },
          // لوحة الأسطول جانت تحت «وحدة المشتريات والمخازن» — ما إلها علاقة
          { to: '/fleet-dashboard', label: 'لوحة تحكم الأسطول', icon: <></>, permission: 'vehicle_management' },
        ],
      },
    ],
  },

  // صندوق المراقب أول شي بباب المراقبة: هو الي «يوصله» شغل، وبقية
  // الشاشات تفرّج بس.
  { to: '/monitor-inbox', label: '👁️ صندوق المراقب', icon: <></>, roles: ['ADMIN', 'MONITOR'], permission: 'monitoring' },
  { to: '/monitor', label: 'لوحة المراقبة', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>, roles: ['ADMIN', 'MONITOR'], permission: 'monitoring' },
  // "خريطة المواقع" انشالت من القائمة — الفني هسه يشوف طريق مهمته مباشرة
  // من صفحة "مهامي" (بوب-أب داخل نفس الصفحة، بدون تحويل لصفحة ثانية).
  // مدير المشاريع مدير مو فني: ما عنده مهام تنستلم ولا تقييم ولا تصنيف ولا تقارير عمل
  {
    // مجموعة "العمل" للفني/الليدر — مهامه اليومية ومصاريفه وتقاريره وفواتيره
    fieldStaffOnly: true,
    to: '/tech-work-group', label: 'العمل', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>,
    children: [
      // ⚠️ «التقارير» و«فواتير الليدر» انشالن من هنا عن قصد — صاروا
      // خيارين جوّا شاشة «مهامي» نفسها.
      //
      // الثلاثة نفس الشغل: حجزك، وفاتورة حجزك، وتقرير حجزك. لما كانوا
      // ثلاث بنود بالقائمة، الفني يخلص الحجز ويطلع للقائمة ويدور على
      // البند الثاني حتى يكمّل نفس الشغلة — وأغلبهم ما كانوا يدورون.
      //
      // (المحاسب والمراقب والمدير إلهم بندهم الخاص بمجموعة الحسابات
      // — ذاك يشوف فواتير **الكل**، وهذا يشوف فواتيره هو.)
      { to: '/my-tasks', label: 'مهامي', icon: <></>, roles: ['TECHNICIAN', 'TECHNICAL'] },
      // ⚠️ «مصاريفي» انشالت من هنا — صارت خيار جوّا «مهامي» ويّا
      // الحجوزات والفواتير والتقارير. الأربعة يخصّون نفس الشغل:
      // حجزك، وفاتورة حجزك، وتقرير حجزك، ومصاريف حجزك.
      //
      // (مدير المشاريع ما عنده «مهامي»، فيبقى بنده هنا — بدونه
      // يفقد الوصول لمصاريفه نهائياً.)
      { to: '/my-expenses', label: 'مصاريفي', icon: <></>, roles: ['PROJECT_MANAGER'] },
      // حسبتان مختلفتان بنفس المحرك:
      //  • «استفسار زبون» = رقم بس، ما ينحفظ ولا ينربط بحجز — للزبون
      //    الي يسأل عن السعر قبل ما يحجز.
      //  • «حساب كلفة زبون (حجز)» = نفس الحساب بس مربوط بالحجز الي
      //    راح يطلع له الليدر، ويترحّل فاتورة للمحاسب.
      // مدخل واحد للاثنين — الشاشة نفسها تسأل: استفسار لو مربوط بحجز
      { to: '/leader-invoices/new', label: '🧮 حساب الكلفة', icon: <></>, roles: ['TECHNICIAN', 'TECHNICAL'], leaderOnly: true, permission: 'execution_cost' },
      // ⚠️ «حساب كلفة الكاميرات» و«حساب كلفة الشبكات» انشالن من هنا —
      // الثلاثة كانوا ثلاث بنود متشابهة الأسماء بالقائمة، والليدر
      // يحتار أي وحدة يفتح.
      //
      // صاروا **خيارين جوّا «حساب الكلفة»** بخانة اختيار نوع العمل:
      // تختار كاميرات أو شبكات فيوديك لحاسبتها. وهذا أصح منطقياً
      // بعد — الاثنين تسعيرتهن **شرائح** (الشبكات: ١٢٬٠٠٠ لحد ٢٠م
      // وبعدها للمتر؛ الكاميرات: شيت مستقل بمعادلة غير)، فما ينفع
      // ينحسبن بجدول المنظومات العادي.
      //
      // (المحاسب والمدير إلهم بندهم الخاص بمجموعة الحسابات — ذاك
      // مدخل تسعير مستقل مو مربوط بشغل ليدر معيّن.)
      // صيانة الأجهزة العامة: حصراً للتيم ليدر (شيت "صيانة الاجهزة")
      { to: '/device-maintenance', label: 'صيانة الأجهزة', icon: <></>, roles: ['TECHNICIAN', 'TECHNICAL'], leaderOnly: true, unlockPermission: 'device_maintenance' },
      // طلبات المواد — شغل ميدان مو إدارة، فمحلها هنا مو باب «الإدارة»
      { to: '/procurement', label: 'طلبات المواد', icon: <></>, roles: ['TECHNICIAN', 'TECHNICAL'], permission: 'procurement' },
      // ═══ «الجرد» بند واحد ═══
      // كان مجموعة تنفتح على بندين («جرد أدواتي» و«جرد أدوات فريقي»)،
      // يعني ضغطتين حتى توصل لشغلة وحدة. صار بند واحد يفتح الشاشة،
      // والاختيار بين «جرد عدتي» و«جرد أدوات فريقي» من فوگ بالواجهة.
      { to: '/my-inventory', label: 'الجرد', icon: <></>, roles: ['TECHNICIAN', 'TECHNICAL'] },
      // ═══ «المشاريع» بند واحد ═══
      // كانت مجموعة تنفتح على بندين. صارت بند واحد يفتح الشاشة،
      // والاختيار بين «كل المشاريع» و«الموجّهة لي» وزر «إضافة مشروع»
      // من فوگ بالواجهة — نفس نمط «مهامي» و«الجرد» و«التقييم».
      //
      // ⚠️ المدخل ‎/my-projects‎ لأن الفني الي ما عنده إدارة مشاريع
      // ما يشوف إلا الموجّهة له، والشاشة تفتح عليها مباشرة. ومن عنده
      // الإدارة يلگه الخيارين من فوگ.
      { to: '/my-projects', label: '🏗️ المشاريع', icon: <></>, roles: ['TECHNICIAN', 'TECHNICAL'], unlockPermission: 'my_projects' },
    ],
  },
  { to: '/gps/employee', label: 'لوحتي GPS', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/></svg>, roles: ['TECHNICIAN'], gpsSkillOnly: true },

  // ═══ فاصل: تحته "الوحدات" — كل وحدة إدارية تجمع محتوياتها تحت باب واحد.
  // الجودة/مراجعة تقارير العمل صارت بس داخل "وحدة الجودة والسلامة المهنية"،
  // وتدقيق تنسيق الحجوزات صار بس داخل "وحدة الرقابة" — ما ضلوا عناصر
  // منفصلة هنا حتى ما تتكرر بالقائمة.
  // ملاحظة: وحدات "الإعلام والعلاقات العامة" و"التصميم" و"التقنيات (IT)" ما
  // ضفناها لأنه ما عندها صفحات مبنية بالنظام بعد — تحتاج طلب منفصل لبنائها.
  // المشاريع الموجّهة لي: أي موظف ينوجّهله مشروع يشوفه هنا بكل مراحله — بدون
  // ما ننطيه صلاحية إدارة المشاريع العامة. الصفحة تطلع فاضية لو ماكو شي.
  { to: '/my-projects', label: 'المشاريع الموجّهة لي', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><path d="M14 2v6h6"/><path d="M9 15l2 2 4-4"/></svg>, unlockPermission: 'my_projects', hideForRoles: ['FINANCE'] },

  // للمحاسب ما بقى فوگ الفاصل ولا شي (الإدارة انحجبت عنه لأنها مكرّرة)،
  // فالفاصل يصير خط يفصل الفراغ عن «العمل» — ضجيج بلا معنى.
  { to: '/units-divider', label: '── الوحدات ──', icon: <></>, divider: true, hideForRoles: ['FINANCE'] },

  {
    // وحدة الخدمة: استقبال وتنسيق طلبات الزبائن وتنفيذها
    to: '/unit-service', label: 'وحدة الخدمة', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>,
    unitPermission: 'unit_service',
    children: [
      { to: '/sales', label: 'حجز جديد', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'QUALITY_ENGINEER'], permission: 'sales_booking' },
      { to: '/customers', label: 'العملاء', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'], permission: 'manage_customers' },
      { to: '/bookings', label: 'الحجوزات', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR', 'FINANCE'], permission: 'view_bookings' },
      { to: '/coordinator', label: 'تنسيق الحجوزات', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'], permission: 'coordinator' },
      { to: '/bookings-archive', label: 'أرشيف الحجوزات', icon: <></>, roles: ['ADMIN', 'MONITOR'], permission: 'bookings_archive' },
      { to: '/solar', label: '☀️ الطاقة الشمسية', icon: <></>, roles: ['ADMIN', 'OWNER', 'MONITOR', 'TECHNICIAN', 'SERVICE_MANAGER'], permission: 'solar_system' },
      { to: '/training-programs', label: '🎓 برامج التدريب', icon: <></>, roles: ['ADMIN', 'OWNER'], permission: 'training_manage' },
      { to: '/missions', label: 'تتبع المهام', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'], permission: 'mission_tracking' },
    ],
  },
  {
    // وحدة التقنيين: خمس فقرات منفصلة بالقائمة الجانبية (كل ضغطة تفتح صفحتها
    // هي بس، مو كل الفقرات مع بعض) — إدارة المعارض / المنتجات / الخدمات /
    // مفردات التدريب (ترحّلت من وحدة التدريب المستقلة السابقة) / معرض الأعمال.
    // permission: 'unit_technicians' على المجموعة نفسها (مو بس على الأبناء) —
    // بوابة إضافية فوق بوابات الأبناء، حتى الوحدة كاملة تختفي عن أي حد ما
    // يملك هذي الصلاحية تحديداً، بغض النظر عن أي صلاحيات ثانية عنده (كان
    // content_technician لحاله كافي يفتحها بالغلط لأي حد يملكها من مكان ثاني).
    to: '/unit-technicians', label: 'وحدة التقنيين', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/></svg>,
    unitPermission: 'unit_technicians',
    permission: 'unit_technicians',
    children: [
      { to: '/exhibitions', label: 'إدارة المعارض', icon: <></>, permission: 'unit_technicians' },
      // شغلتان مختلفتان: «إضافة منتج» تضيفه لكتالوج النظام مباشرة
      // بكل مزاياه (التوفر، الخدمة، المواصفات، المصدر، الموديل)،
      // و«طلبات المنتجات» اقتراح ينتظر موافقة المدير.
      { to: '/products', label: '📦 المنتجات', icon: <></>, permission: 'unit_technicians' },
      { to: '/product-requests', label: 'طلبات المنتجات', icon: <></>, permission: 'unit_technicians' },
      { to: '/service-studies', label: 'دراسات الخدمات', icon: <></>, permission: 'unit_technicians' },
      { to: '/training-management', label: 'مفردات التدريب', icon: <></>, permission: 'content_technician' },
      { to: '/tech-showcase', label: 'معرض الأعمال', icon: <></>, permission: 'content_technician' },
    ],
  },
  // وحدتان فارغتان مؤقتاً (بانتظار تحديد آلية العمل والصلاحيات المطلوبة لكل
  // وحدة) — تطلعان لمدير النظام بس، وتفتحان صفحة "قريباً" بدل ما تختفيان بالكامل.
  {
    to: '/unit-design-group', label: 'وحدة التصميم', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 2a4.5 4.5 0 0 0 0 9 4.5 4.5 0 0 1 0 9"/></svg>, roles: ['ADMIN'],
    unitPermission: 'unit_design',
    children: [
      { to: '/design-forms/quick-add', label: 'إضافة سؤال', icon: <></>, roles: ['ADMIN'], unlockPermission: 'design_forms' },
      { to: '/design-forms', label: 'فورمة التصميم', icon: <></>, roles: ['ADMIN'], unlockPermission: 'design_forms' },
    ],
  },
  {
    // وحدة الإعلام والعلاقات العامة: علاقات الشركة مع زبائنها — الشخصيات
    // المهمة (VIP) الي يأشّرها الموظفون، وسياسة الخصوصية المعلنة.
    to: '/unit-pr', label: 'وحدة الإعلام والعلاقات العامة', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M11 5L6 9H2v6h4l5 4V5z"/><path d="M15.5 8.5a5 5 0 0 1 0 7"/></svg>,
    unitPermission: 'unit_pr',
    children: [
      { to: '/vip-customers', label: '⭐ الشخصيات المهمة', icon: <></>, permission: 'vip_manual_add' },
    ],
  },

  {
    to: '/unit-quality', label: 'وحدة الجودة والسلامة المهنية', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z"/></svg>,
    unitPermission: 'unit_quality',
    children: [
      { to: '/quality', label: 'الجودة', icon: <></>, permission: 'quality_control' },
      { to: '/work-reports-review', label: 'مراجعة تقارير العمل', icon: <></>, anyPermission: ['monitoring', 'quality_control'] },
      { to: '/quality-follow-ups', label: 'متابعة الجودة', icon: <></>, roles: ['ADMIN', 'MONITOR', 'QUALITY_ENGINEER'], permission: 'quality_control' },
    ],
  },
  {
    // وحدة الرقابة: كل أدوات المراقبة والتدقيق العام بمكان واحد
    to: '/unit-monitoring', label: 'وحدة الرقابة', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>,
    unitPermission: 'unit_monitoring',
    children: [
      { to: '/monitor-inbox', label: '👁️ صندوق المراقب', icon: <></>, roles: ['ADMIN', 'MONITOR'], permission: 'monitoring' },
      { to: '/monitor', label: 'لوحة المراقبة', icon: <></>, roles: ['ADMIN', 'MONITOR'], permission: 'monitoring' },
      { to: '/crew-bookings-audit', label: 'تدقيق تنسيق الحجوزات', icon: <></>, permission: 'crew_management' },
      { to: '/complaints', label: '⚠️ الشكاوى', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'], permission: 'complaints' },
      // إدارة سياسة الخصوصية (إضافة/تعديل النقاط) — صلاحية مستقلة تماماً عن
      // قراءتها. القراءة متاحة لكل موظف من الرابط فوق "تسجيل الخروج".
      { to: '/privacy-policy', label: '🔒 إدارة سياسة الخصوصية', icon: <></>, permission: 'privacy_policy_manage' },
    ],
  },
  {
    to: '/unit-procurement', label: 'وحدة المشتريات والمخازن', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M3 3h18v18H3z"/></svg>,
    unitPermission: 'unit_procurement',
    children: [
      { to: '/procurement', label: 'طلبات المواد', icon: <></>, roles: ['ADMIN', 'MONITOR', 'PROJECT_MANAGER', 'TECHNICIAN', 'PROCUREMENT_ADMIN'], permission: 'procurement' },
      { to: '/suppliers', label: 'الموردون', icon: <></>, anyPermission: ['suppliers_management'] },
      { to: '/inventory', label: 'جرد الأدوات', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR', 'PROCUREMENT_ADMIN'], permission: 'inventory' },
      // إدارة المركبات تظهر هنا كمان (مو بس بمجموعتها) — المخازن مسؤولة عنها
      { to: '/vehicles', label: 'إدارة المركبات', icon: <></>, permission: 'vehicle_management' },
      // لوحة تحكم الأسطول محلها هنا — المخازن هيه المسؤولة عن الأسطول
    ],
  },
  {
    // ⚠️ الاسم «العمل» مو «وحدة الحسابات» — بطلب صاحب العمل: المحاسب
    // ما يريد باب اسمه وحدة ويگعد يدور جوّاه؛ يريد شغله كله تحت «العمل».
    // صلاحية الوحدة نفسها (unit_finance) ما انتغيّرت، فالمنوحين ما
    // ينتأثرون — الي انتغيّر العنوان ومحتوياته بس.
    // ما تتصادم وية «العمل» مال الفني (فوگ): كل أبناء ذيچ مقيّدين
    // بـTECHNICIAN/TECHNICAL، فما تطلع للمحاسب أبداً.
    to: '/unit-finance', label: 'العمل', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M12 1v22M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>,
    unitPermission: 'unit_finance',
    children: [
      { to: '/finance', label: 'تدقيق الحسابات', icon: <></>, roles: ['ADMIN', 'FINANCE', 'MONITOR'], permission: 'finance' },
      { to: '/daily-audit', label: '📅 التدقيق اليومي', icon: <></>, roles: ['ADMIN', 'FINANCE', 'MONITOR'], permission: 'finance' },
      { to: '/revolving-fund', label: '💵 الدوار', icon: <></>, permission: 'revolving_fund' },
      { to: '/audit-issues', label: '💸 بلاغات أخطاء التدقيق', icon: <></>, roles: ['ADMIN', 'MONITOR', 'QUALITY_ENGINEER', 'FINANCE'], unlockPermission: 'audit_issues' },
      // موجودة بالقائمة الرئيسية كمان — منحطة هنا لأن محلها المنطقي الحسابات
      { to: '/leader-invoices/new', label: '🧮 حساب الكلفة', icon: <></>, permission: 'execution_cost' },
      { to: '/gps-install-costs', label: '🔧 حساب تكاليف الشد', icon: <></>, roles: ['ADMIN', 'FINANCE'], unlockPermission: 'gps_install_costs' },
      // الشبكات والكاميرات: حاسبات مستقلة بمعادلات خاصة. جانن بمجموعة
      // الميدان بس، فالمحاسب ما يوصلهن — وهو الي يطلع فاتورة الشبكات.
      { to: '/network-cost', label: '🌐 حساب كلفة الشبكات', icon: <></>, roles: ['ADMIN', 'FINANCE'], permission: 'execution_cost' },
      { to: '/camera-cost', label: '📷 حساب كلفة الكاميرات', icon: <></>, roles: ['ADMIN', 'FINANCE'], permission: 'execution_cost' },
      // شاشة مراجعة كل الفواتير — للمحاسب والمراقب والمدير والمالك.
      // الليدر إله بنده الخاص تحت (يشوف فواتيره هو بس).
      { to: '/leader-invoices', label: '🧾 فواتير الليدر', icon: <></>, roles: ['ADMIN', 'FINANCE', 'MONITOR'], unlockPermission: 'leader_invoices_view' },
      { to: '/expenses', label: 'إدارة المصاريف', icon: <></>, roles: ['ADMIN', 'FINANCE'], unlockPermission: 'expenses_manage' },
      // المشاريع الموجّهة لي: محلها هنا للمحاسب — والنسخة العامة فوگ
      // منحجوبة عنه بـhideForRoles حتى ما تتكرر.
      { to: '/my-projects', label: 'المشاريع الموجّهة لي', icon: <></>, unlockPermission: 'my_projects' },
    ],
  },
  {
    to: '/unit-hr', label: 'وحدة الكوادر التنفيذية', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="8.5" cy="7" r="4"/></svg>,
    unitPermission: 'unit_hr',
    children: [
      { to: '/employees', label: 'إدارة الكوادر', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'], permission: 'staff_management' },
      { to: '/kpi', label: 'نقاط الكي بي اي', icon: <></>, roles: ['ADMIN', 'MONITOR'], permission: 'kpi_management' },
      { to: '/staff-requests', label: 'طلبات الكادر', icon: <></>, roles: ['HR_COORDINATOR'], unlockPermission: 'staff_requests' },
      { to: '/performance-review', label: '⭐ تقييم الأداء', icon: <></>, roles: ['HR_COORDINATOR'], unlockPermission: 'performance_review' },
    ],
  },
  {
    to: '/unit-projects', label: 'وحدة إدارة المشاريع', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="3" width="18" height="18" rx="2"/></svg>,
    unitPermission: 'unit_projects',
    children: [
      { to: '/projects', label: '🏗️ المشاريع', icon: <></>, anyPermission: ['project_management', 'project_create_only'], unlockPermission: 'project_management' },
          { to: '/project-work-types', label: 'إعدادات: أنواع الأعمال', icon: <></>, permission: 'project_management' },
      { to: '/project-statistics', label: '📊 إحصائيات المشاريع', icon: <></>, permission: 'project_management' },
      { to: '/checklists', label: 'الكشوفات', icon: <></>, permission: 'project_management' },
      { to: '/quotations', label: 'عروض الأسعار', icon: <></>, anyPermission: ['quotation_create', 'quotation_edit_own', 'quotation_manage_all', 'quotation_system'] },
      { to: '/products', label: '📦 المنتجات', icon: <></>, anyPermission: ['quotation_manage_all', 'quotation_system'] },
    ],
  },

  // ── اختصارات سريعة (أهم إجراءات المبيعات) ──
  { to: '/sales', label: 'حجز جديد', icon: <I d="M12 5v14M5 12h14" />, roles: ['SALES'], unlockPermission: 'create_booking' },
  { to: '/complaints', label: '⚠️ الشكاوى', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M12 9v4M12 17h.01M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/></svg>, roles: ['SALES'] },

  // ── مجموعة GPS (تلم كل طلبات الـ GPS الخاصة بالمبيعات تحت باب وحد) ──
  {
    to: '/gps-group', label: 'GPS', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/></svg>,
    roles: ['SALES'],
    children: [
      { to: '/gps/purchase', label: 'طلب GPS جديد', icon: <></>, roles: ['SALES'], unlockPermission: 'gps_requests' },
      { to: '/gps/delivery', label: 'تسليم أجهزة GPS', icon: <></>, roles: ['SALES'], unlockPermission: 'gps_requests' },
      { to: '/gps/renewal', label: 'طلب تجديد GPS', icon: <></>, roles: ['SALES'], unlockPermission: 'gps_requests' },
      { to: '/gps/maintenance-request', label: 'طلب صيانة GPS', icon: <></>, roles: ['SALES'], unlockPermission: 'gps_requests' },
    ],
  },
]

const loadStoredEmployee = (): Employee | null => {
  const raw = localStorage.getItem('currentEmployee')
  if (!raw) return null
  try { return JSON.parse(raw) as Employee } catch { return null }
}

function hasActiveChild(item: NavItem, pathname: string): boolean {
  if (!item.children) return pathname === item.to || pathname.startsWith(item.to + '/')
  return item.children.some(c => hasActiveChild(c, pathname))
}

const roleColors: Record<string, string> = {
  ADMIN: 'from-amber-500 to-orange-600',
  SALES: 'from-emerald-500 to-teal-600',
  HR_COORDINATOR: 'from-violet-500 to-purple-600',
  TECHNICIAN: 'from-sky-500 to-blue-600',
  PROJECT_MANAGER: 'from-rose-500 to-pink-600',
  MONITOR: 'from-cyan-500 to-teal-600',
  FINANCE: 'from-lime-500 to-green-600',
  GPS_ADMIN: 'from-indigo-500 to-blue-600',
  QUALITY_ENGINEER: 'from-fuchsia-500 to-purple-600',
  ENGINEER: 'from-teal-500 to-cyan-700',
}

// ═══ منو يشوف شنو — مصدر وحيد ═══
//
// نقية بلا حالة، حتى نقدر نسألها «شنو راح يشوفه فلان؟» بدون ما نسجّل
// دخوله. شاشة «شوف بعين الموظف» تناديها بنفسها، فما تنحرف عن الواقع
// أبداً — نفس الدالة الي ترسم القائمة الحقيقية.
export type NavContext = {
  employee: Employee | null | undefined
  permissions: string[]
  gpsServiceId: string | null
}

export function isNavVisible(item: NavItem, ctx: NavContext, unitGranted = false): boolean {
  const role = ctx.employee?.role
  const hasMonitor = role === 'MONITOR' || ctx.permissions.includes('monitoring')
  const hasAudit = ctx.permissions.includes('auditing')
  // الفني والتقني العادي (مو ليدر) — قائمتهم مقفلة على شغلهم.
  // التقني نفس الفني بالميدان، بس يتولى أكثر من خدمة — فما إله شغل
  // بالشاشات الإدارية مثل ما ما إله شغل الفني.
  const isPlainTechnician = (role === 'TECHNICIAN' || role === 'TECHNICAL') && !ctx.employee?.isLeader
  // unitGranted: صحيح لما يكون الموظف عنده صلاحية الوحدة الي هذا العنصر
  // داخلها — وقتها كل شي جوّا الوحدة يظهر له بدون فحص صلاحيات تفصيلية.
  
    // منع صريح لدور — **قبل كل شي**، حتى قبل الفواصل وقاعدة «المنح
    // يفتح»: الفاصل يرجّع true بلا فحص، والصلاحية ترجّع العنصر وتخلي
    // الشاشة تتكرر بمحلين بقائمة نفس الموظف.
    if (item.hideForRoles && role && item.hideForRoles.includes(role as EmployeeRole)) return false

    if (item.divider) return true
    if (item.fieldStaffOnly && role !== 'ADMIN' && !(role === 'TECHNICIAN' || role === 'TECHNICAL')) return false

    // ═══ قاعدة تعلو على كل شي: الصلاحية الممنوحة صراحةً تفتح العنصر ═══
    //
    // قبل، منح الصلاحية ما جان ينفع لأن أربع بوابات تشتغل **قبله**:
    // قائمة الفني البيضاء، وحجب كادر الميدان، وشرط الدور، وبوابة الوحدة.
    // فالمدير ينطي الصلاحية والموظف يضل ما يشوف الشاشة — وهاي بالضبط
    // شكوى صاحب العمل «صلاحية من أنطيها لأحد يلا تظهر إله».
    //
    // ⚠️ الاستثناء الوحيد ownerOnly: شاشات المالك ما تنفتح بالمنح أبداً.
    if (!item.ownerOnly) {
      const grantedExplicitly =
        (!!item.unlockPermission && ctx.permissions.includes(item.unlockPermission)) ||
        (!!item.permission && ctx.permissions.includes(item.permission)) ||
        (!!item.anyPermission && item.anyPermission.some((p) => ctx.permissions.includes(p))) ||
        (!!item.unitPermission && ctx.permissions.includes(item.unitPermission))
      if (grantedExplicitly) return true
    }
    if (item.ownerOnly && ctx.employee?.actualRole !== 'OWNER') return false
    // كادر الميدان (فني أو ليدر): باب «الإدارة» ما يطلع لهم أبداً. شغل
    // الليدر الإداري موجود بمجموعة «العمل» مالته.
    if (item.hideFromFieldStaff && (role === 'TECHNICIAN' || role === 'TECHNICAL')) return false
    // الوحدات: بوابة صارمة. الوحدة ما تطلع أبداً إلا لمن عنده صلاحية الوحدة
    // نفسها (أو مدير النظام). قبل، الوحدة جانت تطلع لمجرد إنه ابن واحد جوّاها
    // مسموح بصلاحية عامة — فصار الموظف يشوف وحدات مو إلها علاقة بشغله ويتكرر
    // نفس المحتوى مرتين (مرة بـ"الإدارة" ومرة بالوحدة).
    // القاعدة المطلوبة: "الإدارة" = كل شي الموظف عنده صلاحيته مهما كانت وحدته،
    // و"الوحدات" = بس الوحدة الي انمنحت له صراحةً.
    if (item.unitPermission && role !== 'ADMIN' && !ctx.permissions.includes(item.unitPermission)) {
      return false
    }
    // الفني العادي: قائمة مقفلة على شغله. العنصر الي إله رابط (مو
    // مجموعة) لازم يكون بالقائمة المسموحة، أو ينفتح بصلاحية منحها
    // المدير بيده — مو بصلاحية جات تلقائياً مع الدور.
    if (isPlainTechnician && !item.children && !item.divider && !unitGranted) {
      const path = (item.to || '').split('?')[0]
      if (!TECHNICIAN_NAV.includes(path)) return false
    }
    const granted =
      unitGranted ||
      (!!item.unitPermission && (role === 'ADMIN' || ctx.permissions.includes(item.unitPermission)))
    if (!granted) {
      // الصلاحية الممنوحة فعلياً تكفي بحالها. قبل، العنصر كان يشترط الدور
      // *و* الصلاحية سوه — فلو منحت إداري الكميات صلاحية "عرض الحجوزات"
      // تضل مخفية عنه لأن دوره مو بقائمة الأدوار المسموحة. هذا خالف معنى
      // منح الصلاحية أصلاً، وكان يخلي صلاحيات كثيرة "ما تنطبق".
      const hasOwnPermission =
        (!!item.permission && ctx.permissions.includes(item.permission)) ||
        (!!item.anyPermission && item.anyPermission.some((p) => ctx.permissions.includes(p)))

      if (!hasOwnPermission) {
        if (item.roles && role && !item.roles.includes(role as EmployeeRole)) {
          // ⚠️ التيم ليدر فني قبل كل شي — لازم يشوف شاشات الميدان
          // (مهامي، جردي، حساب الكلفة) حتى لو دوره مو TECHNICIAN.
          // بدون هذا كان لازم نخليه «ليدر **وفني**» بنفس الوقت حتى
          // يطلعله الحجز، وهذا مو معقول: الفاتورة يمّه والشغل يمّه.
          const leaderFieldItem = !!ctx.employee?.isLeader && item.roles.includes('TECHNICIAN')
          if (!leaderFieldItem && !((hasMonitor || hasAudit) && item.roles.includes('MONITOR'))) return false
        }
        if (item.permission && role !== 'ADMIN' && !ctx.permissions.includes(item.permission)) return false
        if (item.anyPermission && role !== 'ADMIN' && !item.anyPermission.some((p) => ctx.permissions.includes(p))) return false
      }
      if (item.leaderOnly && !ctx.employee?.isLeader && role !== 'ADMIN') return false
      if (item.gpsSkillOnly && role !== 'ADMIN' && !hasGpsSkill(ctx.employee ?? null, ctx.gpsServiceId)) return false
      // الفني العادي بس ينمنع — الليدر يشوفه (نفس قيد السيرفر بالضبط)
    }
    if (item.children) return item.children.some((c) => isNavVisible(c, ctx, granted))
    return true
  return true
}

export default function Layout() {
  const [employee, setEmployeeState] = useState<Employee | null>(loadStoredEmployee)
  const [employeePermissions, setEmployeePermissions] = useState<string[]>([])
  const [gpsServiceId, setGpsServiceId] = useState<string | null>(null)
  // هل عند الموظف المتدرب مواد فعلاً؟ null = لسّه ما فحصنا.
  // بدونها ينحبس بشاشة تدريب فارغة ما بيها شي يدرسه.
  const [traineeHasMaterials, setTraineeHasMaterials] = useState<boolean | null>(null)
  const [expandedGroups, setExpandedGroups] = useState<Record<string, boolean>>({})
  const [collapsed, setCollapsed] = useState(false)
  const [mobileOpen, setMobileOpen] = useState(false)
  const [notifications, setNotifications] = useState<import('../api').Notification[]>([])
  const [unreadCount, setUnreadCount] = useState(0)
  const [notifOpen, setNotifOpen] = useState(false)
  const [settingsOpen, setSettingsOpen] = useState(false)
  const location = useLocation()
  const navigate = useNavigate()

  // Closing the mobile nav on route change is a one-line UI reset tied to router
  // navigation, not data fetching; safe to keep as a synchronous effect.
  // eslint-disable-next-line react-hooks/set-state-in-effect
  useEffect(() => { setMobileOpen(false) }, [location.pathname])

  // وسم عرض الملفات: ينجاب مرة عند الدخول ويتجدد كل 10 دقائق. بدونه
  // الصور المخزّنة برّا القاعدة ترجع 401 لأن <img> ما يرسل ترويسة
  // Authorization.
  useEffect(() => {
    if (!employee) return
    void ensureFileToken()
    const timer = setInterval(() => { void ensureFileToken() }, 10 * 60 * 1000)
    return () => clearInterval(timer)
  }, [employee])

  // نفحص إذا اكو مواد تدريبية منشورة لهذا المتدرب. ما اكو مواد = ما
  // اكو تدريب، فيدخل النظام عادي بدل ما ينحبس بشاشة فارغة.
  useEffect(() => {
    // Resetting the flag when the trainee state changes is derived-state sync, not a fetch.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    if (!employee?.isTrainee) { setTraineeHasMaterials(null); return }
    let alive = true
    api.getMyTraining(employee.id)
      .then((res) => { if (alive) setTraineeHasMaterials((res?.materials?.length ?? 0) > 0) })
      // فشل الطلب ما يصير يحبس الموظف — نخليه يدخل عادي
      .catch(() => { if (alive) setTraineeHasMaterials(false) })
    return () => { alive = false }
  }, [employee])

  // تحديث تلقائي كل نص ساعة + معالجة رجوع الموظف للتبويب بعد غياب طويل
  useAutoRefresh(!!employee)

  const toggleSidebar = () => {
    if (typeof window !== 'undefined' && window.innerWidth < 1024) setMobileOpen((o) => !o)
    else setCollapsed((c) => !c)
  }

  const setEmployee = (emp: Employee | null) => {
    // المالك (OWNER) لازم يشوف ويسوي كل شي مدير النظام (ADMIN) يسويه، وأكثر —
    // نطبّع role إلى 'ADMIN' حتى كل شرط role === 'ADMIN' بالواجهة يشتغل له
    // تلقائياً بدون تعديل كل مكان، ونحفظ الدور الحقيقي بـ actualRole للعرض
    // وللتحقق الحصري بصفحة المراقبة الخلفية.
    const normalized = emp && emp.role === 'OWNER' ? { ...emp, actualRole: 'OWNER' as const, role: 'ADMIN' as const } : emp
    setEmployeeState(normalized)
    if (normalized) {
      localStorage.setItem('currentEmployee', JSON.stringify(normalized))
    } else {
      localStorage.removeItem('currentEmployee')
      localStorage.removeItem('authToken')
    }
  }

  // ═══ حفظ صورة الموظف ═══
  //
  // «أضيف صورة بدل الحرف، ومن أضغط عليها تنفتح».
  //
  // ⚠️ نحدّث الجلسة بالجواب الراجع من السيرفر مو بالقيمة الي دزّيناها:
  // لو السيرفر عدّل شي (أو رفض) تبقى الواجهة تعرض الحقيقة مو أمنيتنا.
  const savePhoto = async (url: string | null) => {
    if (!employee) return
    try {
      const updated = await api.updateEmployee(employee.id, { photoUrl: url ?? '' })
      setEmployee(updated)
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر حفظ الصورة')
    }
  }

  // نتحقق من هوية الموظف الحقيقية من السيرفر مرة وحدة عند فتح النظام —
  // هذا يصحح تلقائياً أي بيانات جلسة قديمة/معدَّلة (مثلاً بأدوات المطورين
  // بالمتصفح) بقيت محفوظة بذاكرة المتصفح المحلية من قبل، ويسجل خروج
  // الحساب تلقائياً إذا صار موقوف (SUSPENDED) بينما الجلسة القديمة لسه مفتوحة
  useEffect(() => {
    if (!employee) return
    api.getMe()
      .then((fresh) => setEmployee(fresh))
      .catch(() => setEmployee(null))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    // Clearing notifications when the user logs out is a guard-clause reset of local
    // state tied to the `employee` dependency, not a fetch; safe to keep as-is.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    if (!employee) { setNotifications([]); setUnreadCount(0); return }
    const load = () => api.getNotifications().then((r) => { setNotifications(r.notifications); setUnreadCount(r.unreadCount) }).catch(() => {})
    load()
    const interval = setInterval(load, 30000)
    return () => clearInterval(interval)
  }, [employee])

  const handleNotifClick = async (n: import('../api').Notification) => {
    if (!n.read) {
      try {
        await api.markNotificationRead(n.id)
        setNotifications((prev) => prev.map((x) => (x.id === n.id ? { ...x, read: true } : x)))
        setUnreadCount((c) => Math.max(0, c - 1))
      } catch { /* ignore */ }
    }
    // الإشعار بلا وجهة يخلي الموظف يدوّر بيده على الشاشة الي تخص الخبر —
    // وأحياناً ما يلكها أصلاً. كل نوع يوديه لمحله مباشرة.
    const target = notifTargets[n.type]
    if (target) {
      setNotifOpen(false)
      navigate(target)
    }
  }

  const handleMarkAllRead = async () => {
    try {
      await api.markAllNotificationsRead()
      setNotifications((prev) => prev.map((x) => ({ ...x, read: true })))
      setUnreadCount(0)
    } catch { /* ignore */ }
  }

  useEffect(() => {
    // Guard-clause reset when logged out; not part of the fetch itself.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    if (!employee) { setEmployeePermissions([]); return }
    api.getEmployeePermissions(employee.id)
      .then((perms) => setEmployeePermissions(perms.map((p) => p.name)))
      .catch(() => setEmployeePermissions([]))
  }, [employee])

  useEffect(() => {
    // Guard-clause reset when logged out; not part of the fetch itself.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    if (!employee) { setGpsServiceId(null); return }
    api.getServices()
      .then((services) => setGpsServiceId(services.find((s) => s.name === 'GPS')?.id || null))
      .catch(() => setGpsServiceId(null))
  }, [employee])

  useEffect(() => {
    const autoExpand = (items: NavItem[]) => {
      items.forEach((item) => {
        if (item.children && hasActiveChild(item, location.pathname)) {
          setExpandedGroups((prev) => ({ ...prev, [item.label]: true }))
          autoExpand(item.children)
        }
      })
    }
    autoExpand(navItems)
  }, [location.pathname])

  if (!employee) {
    return (
      <SessionContext.Provider value={{ employee, setEmployee, permissions: employeePermissions, gpsServiceId }}>
        <Login />
      </SessionContext.Provider>
    )
  }

  // ═══ مركز القيادة ═══
  // نفس اليوزر دخل بباسورد ثاني → نظام ثاني بالكامل (فكرة PPSK).
  // ⚠️ الفصل الحقيقي بالسيرفر (توكن القيادة ينرفض على مسارات
  // الموظفين) — هذا الشرط للعرض بس، مو حماية.
  if (api.currentRealm() === 'command') {
    return (
      <CommandApp
        onExit={() => {
          localStorage.removeItem('authToken')
          localStorage.removeItem('authRealm')
          window.location.reload()
        }}
      />
    )
  }

  // موظف قيد التدريب: يشوف صفحة التدريب فقط، بدون أي وصول لباقي النظام.
  //
  // بس بشرط: لازم يكون المدير ناشر مواد تدريبية فعلاً. قبل، مجرد ما
  // ينتأشر الموظف «متدرب» ينحبس بشاشة مكتوب بيها «لا توجد مواد تدريبية
  // لهذه الخدمة بعد» — ما يشوف مهامه ولا يسجّل دوام ولا يسوي ولا شي،
  // وما عنده حتى شنو يدرس. حبس بلا فايدة.
  //
  // هسه: ما اكو مواد = ما اكو تدريب، والموظف يدخل النظام عادي. الحبس
  // يشتغل بس لمن المدير يحدد المواد — وهو صاحب القرار.
  if (employee.isTrainee && traineeHasMaterials === true) {
    return (
      <SessionContext.Provider value={{ employee, setEmployee, permissions: employeePermissions, gpsServiceId }}>
        <div dir="rtl" className="min-h-screen bg-[#f0f4f9]">
          <header className="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-slate-200/60 bg-white/80 backdrop-blur-xl px-8">
            <span className="text-lg font-extrabold text-[#0f2040] tracking-tight">نظام شركة الأماني — التدريب</span>
            <button onClick={() => setEmployee(null)}
              className="rounded-lg px-3 py-1.5 text-sm font-bold text-red-500 hover:bg-red-50">
              تسجيل الخروج
            </button>
          </header>
          <main className="p-3 sm:p-5 lg:p-8">
            <TrainingPage />
          </main>
        </div>
      </SessionContext.Provider>
    )
  }

  const role = employee?.role
  const isVisible = (item: NavItem, unitGranted = false): boolean =>
    isNavVisible(item, { employee, permissions: employeePermissions, gpsServiceId }, unitGranted)
  // الفني العادي (مو ليدر): قائمته مسطّحة — تنستعمل بترتيب العرض تحت
  const isPlainTechnician = (role === 'TECHNICIAN' || role === 'TECHNICAL') && !employee?.isLeader

  // ═══ تنظيف الشجرة قبل العرض ═══
  //
  // القائمة كانت تتبنى مثل ما هي مكتوبة بالكود، فطلعت:
  //  • ٣٣ شاشة مكررة — نفس الصفحة تحت «الإدارة» ومرة ثانية تحت «الوحدات»
  //    (تدقيق الحسابات، الدوار، المشاريع، الحجوزات... كلها مرتين)
  //  • مجموعات بولد واحد: «إدارة المركبات ← إدارة المركبات»
  //  • تفرعات لخمس مستويات للوصول لشاشة وحدة
  //
  // بدل ما نحذف بالإيد عنصر عنصر (ويرجع الخلل أول ما ينضاف عنصر جديد)،
  // ننضّف بقاعدتين عامتين تشتغلن لحالهن على أي عنصر ينضاف بالمستقبل:
  //   ١. الشاشة تظهر بأول محل توصلها بيه بس — التكرار بعدها ينشال.
  //   ٢. المجموعة الي ما بقى بيها إلا ولد واحد تنفك، والولد يطلع بمحلها.
  type PrunedItem = Omit<NavItem, 'children'> & { granted: boolean; children?: PrunedItem[] }

  // insideUnit: هل احنا جوّا وحدة إدارية؟ الوحدات مستثناة من قاعدة
  // «الشاشة تظهر بأول محل بس» — لأن محتواها كله موجود أصلاً تحت
  // «الإدارة»، فالقاعدة جانت تفرّغ الوحدات وتشيلها كاملة عن مدير
  // النظام. الوحدة باب مستقل يلمّ شغل قسم معيّن، مو تكرار.
  const prune = (items: NavItem[], unitGranted: boolean, seen: Set<string>, depth = 0, insideUnit = false): PrunedItem[] => {
    const out: PrunedItem[] = []
    for (const item of items) {
      if (!isVisible(item, unitGranted)) continue
      const granted =
        unitGranted ||
        (!!item.unitPermission && (role === 'ADMIN' || employeePermissions.includes(item.unitPermission)))

      if (item.divider) { out.push({ ...item, children: undefined, granted }); continue }

      if (item.children) {
        const kids = prune(item.children, granted, seen, depth + 1, insideUnit || !!item.unitPermission)
        if (!kids.length) continue
        // مجموعة بولد واحد = تفرع بلا فايدة: نطلّع الولد محلها.
        // بس مو بالمستوى الأول — هناك الولد يطلع يتيم بلا عنوان يدل
        // على وين هو (مثلاً «تقييم الأداء» طايح جنب «سياسة الخصوصية»).
        // قائمة الفني العادي مسطّحة بالكامل: مجموعات المستوى الأول تنفك
        // وأولادها يطلعون مباشرة. عنده ست شاشات بس، فأي تفرع فوقهن
        // ضغطة زايدة بلا فايدة.
        if (isPlainTechnician && depth === 0) { out.push(...kids); continue }
        if (depth > 0 && kids.length === 1 && !kids[0].divider) { out.push(kids[0]); continue }
        // بالمستوى الأول ما ننفك عموماً (الولد يطلع يتيم بلا عنوان يدل
        // على وين هو). الاستثناء: المجموعة الي ما بقى بيها إلا نفس شاشة
        // عنوانها — مثل «تصنيفي ← تصنيفي» لمن الموظف مو ليدر. هذي
        // تفرع على نفسها، تنفك.
        if (depth === 0 && kids.length === 1 && kids[0].to === item.to) { out.push(kids[0]); continue }
        out.push({ ...item, granted, children: kids })
        continue
      }

      // نفس الشاشة ما تتكرر — أول محل يوصلها هو محلها.
      // المفتاح يشمل الـ query لأن نفس الصفحة تشتغل شغلتين مختلفتين حسبه:
      // ‎/leader-invoices/new?mode=estimate‎ حساب استفساري ما ينحفظ، و
      // ‎/leader-invoices/new‎ فاتورة مربوطة بحجز.
      const key = item.to || ''
      if (!insideUnit) {
        if (key && seen.has(key)) continue
        if (key) seen.add(key)
      }
      out.push({ ...item, children: undefined, granted })
    }
    // فاصل ما يتبعه ولا عنصر (مثلاً "── الوحدات ──" وكل الوحدات انشالت
    // لأن محتواها ظاهر فوق) ما إله معنى
    return out.filter((it, idx, arr) => {
      if (!it.divider) return true
      const next = arr[idx + 1]
      return !!next && !next.divider
    })
  }

  const visibleItems = prune(navItems, false, new Set<string>())

  const toggle = (label: string) => setExpandedGroups((p) => ({ ...p, [label]: !p[label] }))

  const gradientClass = roleColors[employee.role] || 'from-blue-500 to-indigo-600'

  // unitGranted ينتقل للأولاد: لما الموظف عنده صلاحية الوحدة، كل صفحاتها
  // تنعرض له بدون فحص صلاحياتها التفصيلية.
  // العنصر وصلنا منضّف من prune — ما نفلتر ولا نحسب صلاحيات هنا من جديد.
  const renderNavItem = (item: PrunedItem, depth: number = 0): React.ReactNode => {

    if (item.divider) {
      // عنوان قسم بسيط (بلا خطوط) مقابل الفاصل الي يفصل مجموعتين
      if (item.plain) {
        if (collapsed) return <div key={item.label} className="my-2 h-px bg-white/10" />
        return (
          <p key={item.label} className="mb-1 mt-2.5 px-2 text-[9.5px] font-bold tracking-wide text-white/35">
            {item.label}
          </p>
        )
      }
      return (
        <div key={item.label} className="my-2 flex items-center gap-2 px-1">
          <span className="h-px flex-1 bg-white/20" />
          {!collapsed && <span className="text-[11px] font-bold text-white/50">{item.label}</span>}
          <span className="h-px flex-1 bg-white/20" />
        </div>
      )
    }

    if (item.children) {
      const kids = item.children
      const open = expandedGroups[item.label]
      const active = hasActiveChild(item, location.pathname)

      if (depth === 0) {
        return (
          <div key={item.label}>
            <button
              onClick={() => toggle(item.label)}
              className={`group relative flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-[13px] font-bold transition-all duration-300 ${
                active
                  ? 'bg-gradient-to-l from-[#2563eb] to-[#1e40af] text-white shadow-[0_0_18px_rgba(59,130,246,0.45)] ring-1 ring-sky-400/40'
                  : 'bg-white/[0.04] text-blue-100/70 hover:bg-white/[0.08] hover:text-white hover:shadow-[0_0_12px_rgba(59,130,246,0.18)]'
              }`}
            >
              {active && <span className="absolute left-0 top-1/2 h-5 w-[3px] -translate-y-1/2 rounded-r-full bg-sky-300 shadow-[0_0_10px_2px_rgba(125,211,252,0.9)]" />}
              {!collapsed && (
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"
                  className={`ml-0.5 transition-transform duration-300 ${open ? 'rotate-180' : ''}`} style={{ flexShrink: 0 }}>
                  <polyline points="6 9 12 15 18 9"/>
                </svg>
              )}
              {!collapsed && <span className="flex-1 text-right">{item.label}</span>}
              {/* الأيقونة بمربّع — نفس التصميم: صندوق صغير بحافة دائرية
                  يميّز العنصر النشط ويخلي الأيقونات على خط واحد. */}
              <span className={`flex h-7 w-7 shrink-0 items-center justify-center rounded-md transition-all duration-300 ${
                active ? 'bg-white/25 text-white' : 'bg-white/[0.07] text-blue-100/70 group-hover:bg-white/[0.12]'
              }`}>{item.icon}</span>
            </button>
            {open && !collapsed && (
              <div className="mt-1 flex flex-col gap-1 rounded-lg bg-black/25 p-1 animate-in">
                {kids.map(child => renderNavItem(child, 1))}
              </div>
            )}
          </div>
        )
      }

      return (
        <div key={item.label}>
          <button
            onClick={() => toggle(item.label)}
            className={`group flex w-full items-center gap-2 rounded-lg px-3 py-2 text-[13px] font-bold transition-all duration-200 ${
              active ? 'bg-white/[0.08] text-white' : 'text-slate-400 hover:bg-white/[0.04] hover:text-slate-200'
            }`}
          >
            <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round"
              className={`transition-transform duration-300 ${open ? 'rotate-180' : ''}`} style={{ flexShrink: 0 }}>
              <polyline points="6 9 12 15 18 9"/>
            </svg>
            <span className="flex-1 text-right">{item.label}</span>
            {active && <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 animate-pulse" style={{ flexShrink: 0 }}/>}
          </button>
          {open && (
            <div className="mr-4 flex flex-col gap-0.5 border-r border-white/[0.06] pr-2">
              {kids.map(child => renderNavItem(child, 2))}
            </div>
          )}
        </div>
      )
    }

    // Leaf
    if (depth === 0) {
      return (
        <NavLink key={item.to} to={item.to} end={item.end}
          className={({ isActive }) =>
            `group relative flex items-center gap-2 rounded-lg px-2 py-1.5 text-[13px] font-bold transition-all duration-300 ${
              isActive
                ? 'bg-gradient-to-l from-[#2563eb] to-[#1e40af] text-white shadow-[0_0_18px_rgba(59,130,246,0.45)] ring-1 ring-sky-400/40'
                : 'bg-white/[0.04] text-blue-100/70 hover:bg-white/[0.08] hover:text-white hover:shadow-[0_0_12px_rgba(59,130,246,0.18)]'
            }`
          }>
          {({ isActive }) => (
            <>
              {/* شريط أزرق على الحافة اليسرى للعنصر النشط */}
              {/* شعاع الإنارة على الحافة: يمشي مع العنصر النشط */}
              {isActive && <span className="absolute left-0 top-1/2 h-5 w-[3px] -translate-y-1/2 rounded-r-full bg-sky-300 shadow-[0_0_10px_2px_rgba(125,211,252,0.9)]" />}
              {!collapsed && <span className="flex-1 text-right">{item.label}</span>}
              <span className={`flex h-7 w-7 shrink-0 items-center justify-center rounded-md transition-all duration-300 ${
                isActive ? 'bg-white/25 text-white' : 'bg-white/[0.07] text-blue-100/70 group-hover:bg-white/[0.12]'
              }`}>{item.icon}</span>
            </>
          )}
        </NavLink>
      )
    }

    // Leaf مباشرة تحت مجموعة رئيسية (زي "الإدارة") — نفس شكل عناوين المجموعات
    // الشقيقة (بولد وأبيض) حتى ما توهم إنها ابن تابع لمجموعة ثانية فوقها.
    if (depth === 1) {
      return (
        <NavLink key={item.to} to={item.to} end={item.end}
          className={({ isActive }) =>
            `group flex items-center gap-2 rounded-lg px-3 py-2 text-[13px] font-bold transition-all duration-200 ${
              isActive ? 'bg-white/[0.08] text-white' : 'text-slate-300 hover:bg-white/[0.04] hover:text-white'
            }`
          }>
          {({ isActive }) => (
            <>
              <span className="flex-1 text-right">{item.label}</span>
              {isActive && <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 animate-pulse" style={{ flexShrink: 0 }}/>}
            </>
          )}
        </NavLink>
      )
    }

    // ── بند داخل مجموعة ──
    // نقطة زرقاء صغيرة بدل الخط الجانبي: تربط البند بمجموعته بلمحة
    // وتخلي العنصر النشط يبيّن بلا ما ياخذ لون كامل يزاحم عنوان
    // المجموعة فوقه.
    return (
      <NavLink key={item.to} to={item.to} end={item.end}
        className={({ isActive }) =>
          `group flex items-center gap-2 rounded-md px-2.5 py-1.5 text-right text-[12px] transition-all duration-300 ${
            isActive
              ? 'bg-sky-500/25 font-bold text-white shadow-[0_0_12px_rgba(56,189,248,0.35)] ring-1 ring-sky-400/30'
              : 'bg-white/[0.03] font-medium text-blue-100/55 hover:bg-white/[0.08] hover:text-blue-100/85'
          }`
        }>
        {({ isActive }) => (
          <>
            <span className={`h-1.5 w-1.5 shrink-0 rounded-full transition-all duration-300 ${
              isActive ? 'bg-sky-300 shadow-[0_0_8px_2px_rgba(125,211,252,0.8)]' : 'bg-white/20 group-hover:bg-white/40'
            }`} />
            <span className="flex-1 text-right">{item.label}</span>
          </>
        )}
      </NavLink>
    )
  }

  return (
    <SessionContext.Provider value={{ employee, setEmployee, permissions: employeePermissions, gpsServiceId }}>
      {/* سياسة الخصوصية: تنعرض أول دخول ولما تنضاف نقاط جديدة */}
      <PrivacyPolicyGate />
      <div dir="ltr" className="app-shell flex bg-[#f0f4f9]">

        {/* ===== Main Area ===== */}
        <div dir="rtl" className="flex min-w-0 flex-1 flex-col">
          {/* Top Header — Glass effect */}
          <header className="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-slate-200/60 bg-white/80 backdrop-blur-xl px-3 sm:px-5 lg:px-8">
            <div className="flex items-center gap-2 sm:gap-3">
              <button
                onClick={toggleSidebar}
                className="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600"
              >
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="15" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/>
                </svg>
              </button>
              <div className="hidden h-8 w-px bg-slate-200 sm:block"/>
              <span className="truncate text-sm font-extrabold text-[#0f2040] tracking-tight sm:text-lg">نظام شركة الأماني</span>
            </div>
            <div className="flex items-center gap-2 sm:gap-3">
              <div className="relative">
                <button
                  onClick={() => setSettingsOpen((o) => !o)}
                  className="flex h-10 w-10 items-center justify-center rounded-full text-slate-400 transition-all hover:bg-slate-100 hover:text-slate-600"
                  title="الإعدادات"
                >
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <circle cx="12" cy="12" r="3"/>
                    <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.6a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09A1.65 1.65 0 0 0 15 4.6a1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>
                  </svg>
                </button>
                {settingsOpen && <SettingsPanel onClose={() => setSettingsOpen(false)} />}
              </div>
              <div className="relative">
                <button
                  onClick={() => setNotifOpen((o) => !o)}
                  className="relative flex h-10 w-10 items-center justify-center rounded-xl text-slate-400 transition-all hover:bg-slate-100 hover:text-slate-600"
                >
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9M13.73 21a2 2 0 0 1-3.46 0"/>
                  </svg>
                  {unreadCount > 0 && (
                    <span className="absolute -top-0.5 -left-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-red-500 px-1 text-[10px] font-bold text-white">
                      {unreadCount > 9 ? '9+' : unreadCount}
                    </span>
                  )}
                </button>
                {notifOpen && (
                  <>
                    <div className="fixed inset-0 z-40" onClick={() => setNotifOpen(false)} />
                    <div className="absolute left-0 top-12 z-50 w-80 max-w-[90vw] overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-xl">
                      <div className="flex items-center justify-between border-b border-slate-100 px-4 py-3">
                        <span className="text-sm font-bold text-slate-800">الإشعارات</span>
                        {unreadCount > 0 && (
                          <button onClick={handleMarkAllRead} className="text-xs font-medium text-brand-600 hover:underline">تحديد الكل كمقروء</button>
                        )}
                      </div>
                      <div className="max-h-96 overflow-y-auto">
                        {notifications.length === 0 && (
                          <p className="px-4 py-6 text-center text-sm text-slate-400">ماكو إشعارات</p>
                        )}
                        {notifications.map((n) => (
                          <button
                            key={n.id}
                            onClick={() => handleNotifClick(n)}
                            className={`block w-full border-b border-slate-50 px-4 py-3 text-right text-sm transition-colors hover:bg-slate-50 ${n.read ? 'text-slate-500' : 'bg-brand-50/50 font-medium text-slate-800'}`}
                          >
                            <p>{n.message}</p>
                            <div className="mt-1 flex items-center justify-between">
                              <span className="text-[11px] text-slate-400">{new Date(n.createdAt).toLocaleString('ar-IQ')}</span>
                              {/* سهم يبيّن إن الإشعار ينفتح على شاشة */}
                              {notifTargets[n.type] && (
                                <span className="text-[11px] font-bold text-brand-600">افتحها ←</span>
                              )}
                            </div>
                          </button>
                        ))}
                      </div>
                    </div>
                  </>
                )}
              </div>
              <div className="flex items-center gap-2 rounded-xl bg-slate-50/80 px-2 py-1.5 transition-colors hover:bg-slate-100 sm:gap-3 sm:px-3">
                <div className="hidden text-left sm:block">
                  <p className="text-sm font-bold text-slate-800">{employee.name}</p>
                  <p className="text-[11px] text-slate-400">{roleLabels[employee.actualRole || employee.role]}</p>
                </div>
                <div className={`relative flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-xl bg-gradient-to-br ${gradientClass} text-sm font-bold text-white shadow-md sm:h-10 sm:w-10`}>
                  {employee.attendanceIcon || employee.name.charAt(0)}
                  <span className="absolute -bottom-0.5 -left-0.5 h-3 w-3 rounded-full border-2 border-white bg-emerald-500"/>
                </div>
              </div>
            </div>
          </header>

          {/* شريط الإعلانات — يشوفه كل موظف تحت الهيدر مباشرة */}
          <AnnouncementTicker />

          {/* Content */}
          <main className="min-w-0 flex-1 overflow-x-hidden overflow-y-auto p-3 sm:p-5 lg:p-8">
            <Outlet />
          </main>
        </div>

        {/* Mobile backdrop */}
        {mobileOpen && (
          <div onClick={() => setMobileOpen(false)} className="fixed inset-0 z-40 bg-black/40 lg:hidden" />
        )}

        {/* ===== Right Sidebar — Premium ===== */}
        <aside
          dir="rtl"
          // ⚠️ isolate + overflow-hidden: هالة الضوء تحت مطلقة الموضع،
          // وبدونهن تطلع برّا القائمة وتغطي محتوى الصفحة.
          className={`app-sidebar glossy-dark isolate fixed inset-y-0 right-0 z-50 flex flex-col overflow-hidden bg-[#0f2040] transition-transform duration-300 ease-in-out lg:sticky lg:top-0 lg:z-auto lg:h-auto lg:translate-x-0 ${
            mobileOpen ? 'translate-x-0' : 'translate-x-full'
          }`}
          style={{ width: collapsed ? 72 : 270, minWidth: collapsed ? 72 : 270 }}
        >
          {/* ═══ هالة الإنارة ═══
              ضوء أزرق خفيف ينزل من أعلى القائمة، يعطيها العمق الي
              بالتصميم بدل الكحلي المسطّح.
              ⚠️ pointer-events-none: طبقة زينة، وبدونها تبلع ضغطات
              أول بندين بالقائمة. */}
          <div aria-hidden className="pointer-events-none absolute inset-x-0 top-0 -z-10 h-72 bg-[radial-gradient(120%_70%_at_50%_0%,rgba(56,120,255,0.28),transparent_70%)]" />

          {/* Logo */}
          <div className={`flex items-center ${collapsed ? 'justify-center' : 'justify-center gap-3'} px-4 py-5`}>
            {/* شعار الشركة الحقيقي بدل المكعّب العام.
                ⚠️ على خلفية بيضا: الشعار أزرق غامق، وعلى خلفية القائمة
                الكحلية ما يبيّن — نفس اللونين تقريباً. */}
            <div className="relative flex h-11 w-11 items-center justify-center rounded-xl bg-white shadow-lg shadow-blue-900/40">
              <img src="/favicon.png?v=3" alt="شعار شركة الأماني" className="h-8 w-8 object-contain" />
              <span className="absolute -top-1 -left-1 h-3 w-3 rounded-full border-2 border-[#0f2040] bg-emerald-400"/>
            </div>
            {!collapsed && (
              <div>
                <p className="text-sm font-extrabold text-white tracking-tight">الأماني</p>
                <p className="text-[10px] text-blue-300/50 font-medium">Management System</p>
              </div>
            )}
          </div>

          {/* User card */}
          {!collapsed ? (
            <div className="mx-3 mb-4 overflow-hidden rounded-2xl bg-gradient-to-br from-white/[0.08] to-white/[0.03] p-4 shadow-[inset_0_0_0_1px_rgba(255,255,255,0.06)]">
              <div className="flex items-center gap-3">
                <div className="flex-1 text-right">
                  <p className="text-sm font-bold text-white">{employee.name}</p>
                  <p className="text-[11px] text-blue-300/60">{roleLabels[employee.actualRole || employee.role]}</p>
                </div>
                {/* ═══ صورة الموظف ═══
                    «أضيف صورة بدل الحرف، ومن أضغط عليها تنفتح».
                    ⚠️ كل واحد يبدّل **صورته هو** من هنا — مو صور
                    غيره: هاي بطاقته الشخصية بالقائمة. */}
                <div className="relative">
                  <EmployeeAvatar
                    name={employee.name}
                    photoUrl={employee.photoUrl}
                    size="md"
                    rounded="xl"
                    canEdit
                    onPhotoChange={savePhoto}
                  />
                  <span className="pointer-events-none absolute -bottom-0.5 -left-0.5 h-3 w-3 rounded-full border-2 border-[#0f2040] bg-emerald-400"/>
                </div>
              </div>
              <div className="mt-3 flex items-center gap-2">
                <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 animate-pulse"/>
                <span className="text-[10px] text-emerald-400/80 font-medium">متصل الآن</span>
                <span className={`mr-auto rounded-full bg-gradient-to-l ${gradientClass} px-2.5 py-0.5 text-[10px] font-bold text-white shadow-sm`}>
                  {roleLabels[employee.actualRole || employee.role]}
                </span>
              </div>
            </div>
          ) : (
            <div className="mx-auto mb-3">
              <div className="relative">
                <EmployeeAvatar name={employee.name} photoUrl={employee.photoUrl} size="md" rounded="xl" />
                <span className="pointer-events-none absolute -bottom-0.5 -left-0.5 h-2.5 w-2.5 rounded-full border-2 border-[#0f2040] bg-emerald-400"/>
              </div>
            </div>
          )}

          {/* Divider */}
          <div className="mx-4 mb-2 h-px bg-gradient-to-l from-transparent via-white/10 to-transparent"/>

          {/* Nav */}
          <nav className="flex flex-1 flex-col gap-1 overflow-y-auto px-2.5 pb-3 scrollbar-thin">
            {visibleItems.map(item => renderNavItem(item, 0))}
            {employee.actualRole === 'OWNER' && (
              <NavLink
                to="/owner-security"
                className={({ isActive }) =>
                  `mt-1 flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors ${
                    isActive ? 'bg-white/10 text-white' : 'text-white/60 hover:bg-white/5 hover:text-white'
                  }`
                }
              >
                👁️ لوحة المراقبة الخلفية
              </NavLink>
            )}
            {/* النسخ الاحتياطية: للمالك وحده — actualRole مو role، لأن
                role يتطبّع لـ'ADMIN' فوق وهذا يكشفها لكل مدير */}
            {/* ⚠️ «الطلبات» انشالت من هنا — صارت بطاقة بالواجهة
                الرئيسية جنب «جردي». الموظف يفتح النظام ويشوفها
                قدامه، بدل ما ينزّل بالقائمة الجانبية يدوّر عليها.
                (المسار ‎/letters‎ باقي شغّال لمن عنده رابط مباشر.) */}
            {employee.actualRole === 'OWNER' && (
              <NavLink
                to="/command-code"
                className={({ isActive }) =>
                  `mt-1 flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors ${
                    isActive ? 'bg-white/10 text-white' : 'text-white/60 hover:bg-white/5 hover:text-white'
                  }`
                }
              >
                🔐 رمز مركز القيادة
              </NavLink>
            )}
            {employee.actualRole === 'OWNER' && (
              <NavLink
                to="/owner-backups"
                className={({ isActive }) =>
                  `mt-1 flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors ${
                    isActive ? 'bg-white/10 text-white' : 'text-white/60 hover:bg-white/5 hover:text-white'
                  }`
                }
              >
                🔐 النسخ الاحتياطية
              </NavLink>
            )}
            {employee.actualRole === 'OWNER' && (
              <NavLink
                to="/assistant-conversations"
                className={({ isActive }) =>
                  `mt-1 flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors ${
                    isActive ? 'bg-white/10 text-white' : 'text-white/60 hover:bg-white/5 hover:text-white'
                  }`
                }
              >
                💬 محادثات المساعد الذكي
              </NavLink>
            )}
          </nav>

          {/* سياسة الخصوصية — قراءة متاحة لكل موظف بدون أي صلاحية، ثابتة فوق
              زر تسجيل الخروج. إضافة/تعديل النقاط شي ثاني تماماً: صلاحية
              "privacy_policy_manage" وتظهر داخل وحدة الرقابة. */}
          <div className="mx-3 mt-1">
            <NavLink
              to="/privacy-policy"
              className={({ isActive }) =>
                `flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors ${
                  isActive ? 'bg-white/10 text-white' : 'text-white/60 hover:bg-white/5 hover:text-white'
                }`
              }
            >
              {collapsed ? '🔒' : '🔒 سياسة الخصوصية'}
            </NavLink>
          </div>

          {/* لوحة الإعلانات — ثابتة بالأسفل تحت سياسة الخصوصية، مو داخل
              أي وحدة، لأنها تخص المالك ومدير النظام وحدهم. */}
          {(role === 'ADMIN' || role === 'OWNER') && (
            <div className="mx-3 mt-1">
              <NavLink
                to="/announcements"
                className={({ isActive }) =>
                  `flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors ${
                    isActive ? 'bg-white/10 text-white' : 'text-white/60 hover:bg-white/5 hover:text-white'
                  }`
                }
              >
                {collapsed ? '📢' : '📢 لوحة الإعلانات'}
              </NavLink>
            </div>
          )}

          {/* Logout */}
          <div className="mx-3 mb-3 mt-1">
            <div className="h-px bg-gradient-to-l from-transparent via-white/10 to-transparent mb-3"/>
            <button onClick={() => setEmployee(null)}
              className="group flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium text-red-300/50 transition-all duration-200 hover:bg-red-500/10 hover:text-red-300">
              {!collapsed && <span className="flex-1 text-right">تسجيل الخروج</span>}
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" className="opacity-60 group-hover:opacity-100 transition-opacity">
                <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4M16 17l5-5-5-5M21 12H9"/>
              </svg>
            </button>
          </div>
        </aside>
      </div>
      {(employee?.role === 'ADMIN' || employee?.role === 'MONITOR' || employee?.actualRole === 'OWNER')
        ? <ManagerAssistantChat />
        : <AssistantWidget />}
    </SessionContext.Provider>
  )
}
