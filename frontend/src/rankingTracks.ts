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
   * `strictRole` موجودة، المسار يرجع لترتيب الدور الصارم
   * (`RoleLeaderboard`) بدل ترتيب الصلاحية — بس لهذا المسار، بقية
   * المسارات تبقى «حسب الشغل» متل ما هي.
   */
  strictRole?: string
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
  { permission: 'sales_booking',      label: 'المبيعات',         icon: '🤝', strictRole: 'SALES' },
  { permission: 'quality_control',    label: 'الجودة',           icon: '⭐' },
  { permission: 'finance',            label: 'الحسابات',         icon: '💰' },
  { permission: 'procurement',        label: 'المشتريات',        icon: '📦' },
  { permission: 'gps_system',         label: 'نظام GPS',         icon: '📡' },
  { permission: 'project_management', label: 'إدارة المشاريع',   icon: '🏗️' },
  { permission: 'vehicle_management', label: 'المركبات',         icon: '🚗' },
  { permission: 'solar_system',       label: 'الطاقة الشمسية',   icon: '☀️' },
  { permission: 'inventory',          label: 'المخازن والجرد',   icon: '🧰' },
  { permission: 'monitoring',         label: 'المراقبة',         icon: '👁️' },
]

/** المسارات الي ينتمي إلها الموظف حسب صلاحياته الفعلية. */
export function tracksFor(permissions: string[]): RankingTrack[] {
  return RANKING_TRACKS.filter((t) => permissions.includes(t.permission))
}
