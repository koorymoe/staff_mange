import type { Employee, EmployeeRole } from '../api'
import { hasGpsSkill } from '../session'

// ═══ شجرة القائمة ومنو يشوف شنو ═══
//
// ⚠️ بملف لحاله مو داخل `Layout.tsx`: التحديث السريع بالتطوير
// (Fast Refresh) ينكسر لو الملف صدّر مكوّناً وثوابت وياه — يعني كل
// تعديل بالقائمة چان يعيد تحميل الصفحة كلها بدل ما يحدّث المكوّن.
//
// وفصلها إله فايدة ثانية: شاشة «شوف النظام بعين الموظف» وشاشة دليل
// الأدوار تستوردنّ من هنا، بلا ما تسحبن معاهن كل الـLayout.

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

// ⚠️ دالة صغيرة بحرف صغير مو مكوّن بحرف كبير: الملف هذا يصدّر
// بيانات (شجرة القائمة) مو مكوّنات، ووجود مكوّن وياها يكسر التحديث
// السريع بالتطوير.
const icon = (d: string) => (
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
  { to: '/', label: 'الرئيسية', end: true, icon: icon('M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z M9 22V12h6v10') },
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
          // ⚠️⚠️ أربعة بنود (إدارة الكوادر · نقاط الكي بي اي · تقييم
          // الأداء · طلبات الكادر) صارت بند واحد يفتح مكتباً بتبويبات
          // — نفس نمط «مكتب المراقب» بالضبط. الشرط هنا **اتحاد OR**
          // لشروط الأربعة الأصلية: ADMIN/MONITOR بالدور، أو أي وحدة
          // من الصلاحيات الأربعة ممنوحة (`permission`/`unlockPermission`
          // يتصرفان نفس التصرف بـ`isNavVisible` — تفتح العنصر بغض
          // النظر عن الدور)، فـ`anyPermission` تغطي الاثنين سوا.
          // والشاشات الأربع القديمة تبقى بمساراتها (`/employees` وغيرها)
          // — المكتب يضمّها بخاصية `embedded`، ما ينسخها.
          {
            to: '/staff-management-desk', label: 'إدارة الموظفين', icon: <></>,
            roles: ['ADMIN', 'MONITOR'],
            anyPermission: ['staff_management', 'kpi_management', 'performance_review', 'staff_requests'],
          },
          // الصلاحيات جانت مدفونة جوّا مجموعة «إدارة الصلاحيات» — يعني
          // خمس مستويات للوصول لشاشة وحدة. المجموعة انشالت والشاشتين
          // صعدن هنا مباشرة.
          { to: '/permissions', label: 'الصلاحيات', icon: <></>, roles: ['ADMIN'] },
          { to: '/permission-preview', label: '🔎 شوف بعين الموظف', icon: <></>, roles: ['ADMIN'] },
          // ⚠️ الإجازات انشالت من القائمة العلوية: الموظف يطلبها من
          // «جدول دوامي» مباشرة. بس المدير لازم يضل يوصل صندوق الطلبات
          // حتى يوافق — بلا هذا المدخل الطلبات تنتراكم وماكو منو يشوفها.
          { to: '/leaves', label: '🗓️ طلبات الإجازات', icon: <></>, roles: ['ADMIN', 'OWNER', 'MONITOR'], anyPermission: ['leave_approve_morning', 'leave_approve_evening'] },
          // ⚠️ ADMIN بس — StatsPage.tsx نفسها تقفل على غير ADMIN
          // (role !== 'ADMIN')، وكان المراقب يشوف الرابط هنا ويفتح
          // دائماً على «غير مصرح». والخادم (`GET /api/stats`) صار
          // محمياً بـ`requireAdmin` نفسه — القفلان يطابقان القائمة.
          { to: '/stats', label: 'إحصائيات الموظفين', icon: <></>, roles: ['ADMIN'] },
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
  // ⚠️ المكتب **أول** بند: هو الباب الي نريد المراقب يدخل منه، وحطّه
  // بالآخر يخلّيه يفتح الأبواب القديمة بالعادة ولا يشوفه أبداً.
  { to: '/monitor-desk', label: '🗂️ مكتب المراقب', icon: <></>, roles: ['ADMIN', 'MONITOR'], permission: 'monitoring' },
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
      { to: '/monitor-desk', label: '🗂️ مكتب المراقب', icon: <></>, roles: ['ADMIN', 'MONITOR'], permission: 'monitoring' },
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
  { to: '/sales', label: 'حجز جديد', icon: icon('M12 5v14M5 12h14'), roles: ['SALES'], unlockPermission: 'create_booking' },
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

  // ═══ مختبر المحاكاة — للمالك وحده ═══
  //
  // «هذا أريده يظهر فقط عند المالك، حتى مدير النظام ما أريده يظهر عنده
  // إلى أن يكتمل بصورة كاملة».
  //
  // ⚠️ `ownerOnly` هي الوحيدة الي **المنح ما يكسرها**: قاعدة «الصلاحية
  // الممنوحة تفتح العنصر» تستثنيها صراحةً (شوف isNavVisible). والصلاحيات
  // ما تگدر تخفي شي عن مدير النظام أصلاً — كل وسائط الحماية تمرّره.
  //
  // ⚠️ وهاي **أول استعمال** لهالعلامة بالنظام — انفحصت بحساب مدير حقيقي.
  {
    to: '/simulator-lab', label: '🧪 مختبر المحاكاة',
    icon: icon('M9 3v2m6-2v2M5 8h14M6 8v11a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V8M10 12v5m4-5v5'),
    ownerOnly: true,
  },
  {
    to: '/simulator-lab/workbench', label: '🧰 مساحة عمل المحاكاة',
    icon: icon('M4 7h16M4 12h16M4 17h10'),
    ownerOnly: true,
  },
]


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

// ═══ عزل شغل المراقب الأساسي عن الصلاحيات الإضافية ═══
//
// «هذا المراقب اني منطي هواي صلاحيات... بس جاي يتهيه بينهن وبين
// الشغل الرئيسي مالته» — صاحب النظام يمنح المراقب صلاحيات تشغيلية
// زايدة عن دوره (تنسيق الحجوزات، طلبات حذف الحجوزات، ...)، وصار
// ما يميّز وين شغله الأساسي (الإشراف) ينتهي ووين يبدي شغله كمنفّذ
// لصلاحية ممنوحة. القرار: فصل بصري بس — بلا لمس أي صلاحية ولا حارس.
//
// ⚠️ نسخة طبق الأصل عن `RoleDefaultPermissions["MONITOR"]`
// (`internal/model/permission.go:100`) — أي تعديل هناك يلزم تعديلها هنا.
const MONITOR_CORE_PERMISSIONS = new Set([
  'staff_management', 'edit_employee_profile', 'kpi_management',
  'view_bookings', 'manage_customers', 'manage_services', 'mission_tracking',
  'inventory', 'complaints', 'finance', 'monitoring', 'auditing',
  'quality_control', 'gps_system',
])

// صلاحيات العنصر الي **تقيّد فعلاً** — بعكس `unlockPermission` الي
// «تفتح ولا تقيّد» (تعليق `NavItem.unlockPermission` أعلاه): عنصر
// إله `unlockPermission` بس وموجود بـ`roles` المراقب يظهر له دائماً
// بغض النظر عن الصلاحية، فما نعتبرها شرط فتح حقيقي هنا.
function structuralPermissions(item: NavItem): string[] {
  const list: string[] = []
  if (item.permission) list.push(item.permission)
  if (item.anyPermission) list.push(...item.anyPermission)
  if (item.unitPermission) list.push(item.unitPermission)
  return list
}

/**
 * isExtraForMonitor: هل هذا العنصر ظاهر للمراقب **بسبب صلاحية ممنوحة
 * زيادة**، مو بشغله الاعتيادي؟
 *
 * ⚠️ ما نقارن بمصفوفة صلاحيات افتراضية بالتجريد (`RoleDefaultPermissions`
 * بالخادم **ما تُطبَّق تلقائياً** — موظف حقيقي ممكن يوصل يشتغل بدون
 * حتى صلاحياته الأساسية مطبَّقة، وقتها أي مقارنة "لو عنده الافتراضي
 * بس" تعطي نتيجة مضلِّلة). بدلها: نفحص **شروط فتح العنصر نفسه**
 * (`permission`/`anyPermission`/`unitPermission`، لا `unlockPermission`
 * الي ما يقيّد أصلاً): لو أي شرط منها صلاحية أساسية، العنصر جزء من
 * شغله الاعتيادي — ولو كل الشروط صلاحيات زايدة، فهو ظاهر له بسببها بس.
 */
export function isExtraForMonitor(item: NavItem, ctx: NavContext, unitGranted = false): boolean {
  if (ctx.employee?.role !== 'MONITOR') return false
  if (!isNavVisible(item, ctx, unitGranted)) return false
  const structural = structuralPermissions(item)
  if (structural.length === 0) return false
  return !structural.some((p) => MONITOR_CORE_PERMISSIONS.has(p))
}

/**
 * collectMonitorExtraLinks: كل شاشة ظاهرة للمراقب بصلاحية إضافية —
 * دفعة وحدة، مرتّبة كما تظهر بالقائمة. تستعملها الرئيسية حتى تعرض
 * **نفس** الروابط الي عليها شارة «إضافية» بالقائمة الجانبية، بلا
 * قائمة يدوية ثانية تفترق عن الأولى بأول تعديل.
 */
export function collectMonitorExtraLinks(items: NavItem[], ctx: NavContext, unitGranted = false): NavItem[] {
  if (ctx.employee?.role !== 'MONITOR') return []
  const seen = new Set<string>()
  const out: NavItem[] = []
  const walk = (list: NavItem[], granted: boolean) => {
    for (const item of list) {
      if (item.divider) continue
      if (!isNavVisible(item, ctx, granted)) continue
      const childGranted =
        granted ||
        (!!item.unitPermission && (ctx.employee?.role === 'ADMIN' || ctx.permissions.includes(item.unitPermission)))
      if (item.children) { walk(item.children, childGranted); continue }
      if (!item.to || seen.has(item.to)) continue
      if (isExtraForMonitor(item, ctx, granted)) { seen.add(item.to); out.push(item) }
    }
  }
  walk(items, unitGranted)
  return out
}

