// ═══ مسارات التصنيف ═══
//
// الترتيب كان ينبني على **الدور** (e.role). يعني الموظف الي دوره
// «محاسب» بس ينطوه صلاحيات إدارة الكوادر وتنسيق الحجوزات — يشتغل
// شغل المنسّقين كل يوم — ما يظهر بتصنيفهم أبداً، وينقارن بمحاسبين
// ما يشتغلون شغله.
//
// النظام يحكم بالاسم المكتوب بملفه مو بالشغل الي يسويه. وهاي تخلي
// التصنيف يكذب من الطرفين: منسّق شاطر ما يطلع بالقائمة، وواحد ما
// يمسّ التنسيق يتصدّرها.
//
// هسه التصنيف ينبني على **الصلاحية**: منو عنده صلاحية «تنسيق
// الحجوزات» ينقارن بمنسّقي الحجوزات — مهما كان اسم دوره.
//
// ⚠️ والموظف يظهر بأكثر من مسار إذا يشتغل أكثر من شغلة. هذا مو خلل،
// هذا واقعه — والتصنيف الواحد كان يخفيه.

export type RankingTrack = {
  /** اسم الصلاحية بقاعدة البيانات */
  permission: string
  /** اسم المسار بالواجهة */
  label: string
  icon: string
  /**
   * ⚠️ استثناء صريح لمسار المبيعات — صاحب النظام رفض مبدأ "نفس
   * الشغل" هنا تحديداً: صلاحية `sales_booking` يحملها أصلاً إداري
   * الكوادر ومهندس الجودة بحكم دورهم (`RoleDefaultPermissions`)،
   * فترتيب المبيعات كان يخلطهم مع موظفي المبيعات الحقيقيين. لمّن
   * `strictRoles` موجودة، المسار يرجع لترتيب **عائلة** الأدوار
   * (`RoleLeaderboard`) بدل ترتيب الصلاحية — بس لهذا المسار، بقية
   * المسارات تبقى «حسب الشغل» متل ما هي.
   */
  strictRoles?: string[]
}

/**
 * ═══ عوائل التقييم — بقرار (ع) ═══
 *
 * «لازم التقييمات تنفصل: موظفين المبيعات يتصنفون وحد، والتقنيين
 * ومسؤولي الخدمات وحد، والفنيين والليدريه وحد، والمحاسب والمراقب
 * وحد، وموظفين الكميات وحد، والمهندسين وحد».
 *
 * والخمسة الي ما ذكرهن انصنّفن بقراره: مدير المشاريع والمصمم مع
 * **المهندسين** · IT ومسؤول GPS مع **الحسابات والرقابة** · وإداري
 * الكوادر **عائلة وحده**.
 *
 * ⚠️ العائلة تنرسل للخادم مفصولة بفاصلة، والخادم يرشّح
 * `e.role = ANY(...)` — فالمحاسب والمراقب يطلعون بنفس اللوحة فعلاً
 * مو بلوحتين متشابهتين.
 */
export const RANKING_FAMILIES: { label: string; icon: string; roles: string[] }[] = [
  { label: 'المبيعات', icon: '🤝', roles: ['SALES'] },
  { label: 'الفنيون والليدريه', icon: '🔧', roles: ['TECHNICIAN'] },
  { label: 'الحسابات والرقابة', icon: '💰', roles: ['FINANCE', 'MONITOR', 'IT_SUPPORT', 'GPS_ADMIN'] },
  { label: 'المهندسون', icon: '📐', roles: ['ENGINEER', 'QUALITY_ENGINEER', 'PROJECT_MANAGER', 'DESIGNER'] },
  { label: 'الكوادر', icon: '👥', roles: ['HR_COORDINATOR'] },
]

/** عائلة الموظف حسب دوره — أو null لو دوره مو مصنّفاً. */
export function familyForRole(role?: string): { label: string; icon: string; roles: string[] } | null {
  if (!role) return null
  return RANKING_FAMILIES.find((f) => f.roles.includes(role)) ?? null
}

/**
 * المسارات المتاحة، بترتيب الأولوية.
 *
 * ⚠️ ما نحط كل الصلاحيات هنا — بس الي إلها **شغل يومي يتقاس**.
 * صلاحية مثل «شوف بعين الموظف» ما إلها إنتاج ينقارن، وإضافتها
 * تخلق تصنيف بموظف واحد ما يعني شي.
 */
export const RANKING_TRACKS: RankingTrack[] = [
  { permission: 'coordinator',        label: 'تنسيق الحجوزات',   icon: '🗂️' },
  { permission: 'staff_management',   label: 'إدارة الكوادر',    icon: '👥' },
  { permission: 'sales_booking',      label: 'المبيعات',         icon: '🤝', strictRoles: ['SALES'] },
  { permission: 'quality_control',    label: 'الجودة',           icon: '⭐' },
  // ⚠️ الحسابات والمراقبة عائلة **وحدة** بقراره — والصلاحية وحدها
  // ما تكفي: المراقب يحمل `finance` بافتراضي دوره، فترتيب «الحسابات»
  // بالصلاحية چان يخلط الاثنين بلوحة والمراقبة بلوحة ثانية.
  { permission: 'finance',            label: 'الحسابات والرقابة', icon: '💰',
    strictRoles: ['FINANCE', 'MONITOR', 'IT_SUPPORT', 'GPS_ADMIN'] },
  { permission: 'procurement',        label: 'المشتريات',        icon: '📦' },
  { permission: 'gps_system',         label: 'نظام GPS',         icon: '📡' },
  { permission: 'project_management', label: 'إدارة المشاريع',   icon: '🏗️' },
  { permission: 'vehicle_management', label: 'المركبات',         icon: '🚗' },
  { permission: 'solar_system',       label: 'الطاقة الشمسية',   icon: '☀️' },
  { permission: 'inventory',          label: 'المخازن والجرد',   icon: '🧰' },
  { permission: 'monitoring',         label: 'الحسابات والرقابة', icon: '👁️',
    strictRoles: ['FINANCE', 'MONITOR', 'IT_SUPPORT', 'GPS_ADMIN'] },
]

/**
 * ═══ مسارات الموظف: عائلته أولاً، وبعدها شغله ═══
 *
 * 🔴 **ليش الدور قبل الصلاحية**: الترشيح بالصلاحية وحدها هو **بيت
 * العلّة** الي شكى منها (ع). إداري الكوادر يحمل `sales_booking`
 * بافتراضي دوره، فچان يطلعله تبويب «المبيعات» — وهو مو من المبيعات
 * أصلاً. ونفس الشي المراقب يحمل `finance`.
 *
 * فعائلة **دوره** تجي أولاً ودائماً، وتُشال المسارات الي تكرّرها
 * (المبيعات · الحسابات · المراقبة) حتى ما يطلع نفس الموظف بلوحتين
 * لنفس المفهوم.
 *
 * ⚠️ وباقي المسارات تبقى «حسب الشغل» مثل ما هي: الي يشتغل تنسيقاً
 * أو مخازن أو مشتريات ينقارن بأصحاب نفس الشغل مهما كان دوره — هذا
 * مبدأ موجود وصحيح، والعوائل ما تلغيه.
 */
const FAMILY_OWNED_PERMISSIONS = ['sales_booking', 'finance', 'monitoring']

export function tracksFor(permissions: string[], role?: string): RankingTrack[] {
  const family = familyForRole(role)
  const byWork = RANKING_TRACKS.filter(
    (t) => permissions.includes(t.permission) && !FAMILY_OWNED_PERMISSIONS.includes(t.permission),
  )
  if (!family) return byWork
  return [
    { permission: '__family__', label: family.label, icon: family.icon, strictRoles: family.roles },
    ...byWork,
  ]
}
