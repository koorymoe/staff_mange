import type { EmployeeRole } from './api'

// ═══ أسماء الأدوار وألوانها — مصدر وحيد ═══
//
// 🔴 **ليش انجمعت هنا**: كانت **سبع خرائط منفصلة** يكتبن نفس الأسماء
// بالعربي، كل ملف لحاله (`session.ts` · معالج إضافة موظف · دليل
// الأدوار · إدارة الإحصائيات · صندوق المراقب · المشاريع · الموظفين) —
// وخريطتا ألوان زيادة. ومعناها إن **أي دور جديد يحتاج ١٦ تعديلاً**،
// وأي ملف ينساه يعرض للموظف **رمزاً إنكليزياً** مثل `IT_SUPPORT`
// بمحل اسم دوره.
//
// وهاي مو مخافة نظرية: الأسماء **افترقت فعلاً** وهي متناثرة —
// «مدقق / مراقب» بمحل و«مراقب / مدقق» بمحل و«المراقب» بثالث،
// والمالك مرة «مالك النظام 👑» ومرة «المالك». الموظف نفسه يشوف
// وصفين مختلفين لنفس الدور بشاشتين.
//
// ⚠️ **والنوع `Record<EmployeeRole, string>` مقصود مو تجميلاً**: أي
// دور ينضاف للنوع بـ`api.ts` وما ينضاف له اسم هنا **يوقّف البناء**
// (`npm run build`) — يعني الغلط ينكشف عندنا، مو على شاشة موظف.

/**
 * الاسم الكامل — هذا الافتراضي، ويُستعمل بكل مكان إلا لمّا يكون
 * المحل ضيّقاً فعلاً (شريحة داخل صف، أو رأس عمود بجدول).
 */
export const ROLE_LABELS: Record<EmployeeRole, string> = {
  OWNER: 'مالك النظام 👑',
  ADMIN: 'مدير النظام',
  SALES: 'موظف مبيعات',
  HR_COORDINATOR: 'إداري الكوادر',
  TECHNICIAN: 'فني',
  TECHNICAL: 'تقني',
  PROJECT_MANAGER: 'مدير مشاريع',
  MONITOR: 'مدقق / مراقب',
  FINANCE: 'محاسب',
  GPS_ADMIN: 'مسؤول GPS',
  QUALITY_ENGINEER: 'مهندس جودة',
  ENGINEER: 'مهندس',
  PROCUREMENT_ADMIN: 'إداري الكميات',
  DESIGNER: 'مصمم',
  SERVICE_MANAGER: 'مسؤول خدمة',
  IT_SUPPORT: 'الدعم التقني',
}

/**
 * الاسم المختصر — للشرائح والجداول الضيّقة وبس.
 *
 * ⚠️ **ناقصة بقصد** (`Partial`): الي ما إله اختصار يرجع لاسمه الكامل
 * تلقائياً. فإضافة دور جديد **ما تحتاج** صفاً هنا، وهذا الفرق بين
 * «خريطة ثانية» و«استثناءات معدودة».
 *
 * و«فني / ليدر» مو تجميلاً: دور الفني بنظامنا **يشمل التيم ليدر**
 * (الليدر فني مأشّر `isLeader`)، وجدول الإحصائيات لازم يوضّحها حتى
 * ما يتصور القارئ إن الليدرية طلعوا من الحساب.
 */
const SHORT_LABELS: Partial<Record<EmployeeRole, string>> = {
  OWNER: 'المالك',
  TECHNICIAN: 'فني / ليدر',
  SALES: 'مبيعات',
  FINANCE: 'حسابات',
  MONITOR: 'رقابة',
  HR_COORDINATOR: 'كوادر',
  PROCUREMENT_ADMIN: 'مخازن',
  IT_SUPPORT: 'IT',
}

/**
 * اسم الدور، ومعاه أمان للقيمة الي ما نعرفها.
 *
 * ⚠️ **ليش يرجّع الرمز نفسه لمّا ما يعرفه**: القيمة المجهولة هنا
 * معناها **الخادم أحدث من هالبناء** (دور انضاف وما وصلت الواجهة).
 * وبهالحالة الرمز `IT_SUPPORT` يفيد المالك أكثر من «غير معروف» —
 * يعرف منه إن السيرفر يحتاج «سحب وبناء»، بدل ما يفتّش بالفراغ.
 * والفراغ **أسوأ** من الاثنين: يخلي الصف يبين ناقصاً بلا سبب.
 */
export function roleLabel(role: string | null | undefined): string {
  if (!role) return '—'
  return ROLE_LABELS[role as EmployeeRole] ?? role
}

/** نفسها بالمختصر — ويرجع للكامل إذا ماكو اختصار. */
export function roleLabelShort(role: string | null | undefined): string {
  if (!role) return '—'
  return SHORT_LABELS[role as EmployeeRole] ?? roleLabel(role)
}

/**
 * الأدوار الي تنمنح لموظف من الواجهة.
 *
 * ⚠️ **تُشتق من الخريطة مو مكتوبة بالإيد**: قائمة مكتوبة بالإيد تنسى
 * الدور الجديد، فيصير دور موجود بالنظام وماكو طريق يمنحه.
 *
 * - `OWNER` محجوز لحساب المالك الوحيد المزروع بقاعدة البيانات —
 *   محد يمنحه من الواجهة، ولا مدير النظام نفسه.
 * - `GPS_ADMIN` ما ينمنح بعد الآن: الجي بي اس صارت **خدمة** بين
 *   خدمات، ومسؤوليتها تنمنح بصلاحية `gps_system` مو بدور منفصل.
 *   (يبقى بالخريطة لأن اكو موظفون قدام عليه الدور فعلاً، ولازم
 *   يبين اسمهم صح.)
 */
export const ASSIGNABLE_ROLES: EmployeeRole[] =
  (Object.keys(ROLE_LABELS) as EmployeeRole[]).filter((r) => r !== 'OWNER' && r !== 'GPS_ADMIN')

/**
 * الأدوار الي تنمنح **كدور ثانٍ**.
 *
 * 🔴 **بلا OWNER وبلا ADMIN**: دورهما يفتح كل شي — الفلوس والصلاحيات
 * والحذف. ومنحه «كدور ثاني» يعني ترقية كاملة **ما تبين** بقائمة
 * الموظفين، لأن الدور الأساسي يبقى «فني». وهذا بالضبط شكل تسريب
 * الصلاحية الصامت.
 *
 * ⚠️ والخادم يمنعهما بعد (`SanitizeSecondaryRoles`) — هاي القائمة
 * **راحة للمستخدم مو حماية**: أي قيد بالواجهة وحدها يُتخطّى بطلب
 * مباشر على المسار.
 */
export const SECONDARY_ASSIGNABLE_ROLES: EmployeeRole[] =
  ASSIGNABLE_ROLES.filter((r) => r !== 'ADMIN')

/** ألوان شريحة الدور بقائمة الموظفين. */
export const ROLE_CHIP_COLORS: Record<EmployeeRole, { bg: string; text: string; dot: string }> = {
  OWNER: { bg: 'bg-yellow-50', text: 'text-yellow-800', dot: 'bg-yellow-500' },
  ADMIN: { bg: 'bg-amber-50', text: 'text-amber-700', dot: 'bg-amber-500' },
  SALES: { bg: 'bg-emerald-50', text: 'text-emerald-700', dot: 'bg-emerald-500' },
  HR_COORDINATOR: { bg: 'bg-violet-50', text: 'text-violet-700', dot: 'bg-violet-500' },
  TECHNICIAN: { bg: 'bg-sky-50', text: 'text-sky-700', dot: 'bg-sky-500' },
  TECHNICAL: { bg: 'bg-blue-50', text: 'text-blue-700', dot: 'bg-blue-500' },
  PROJECT_MANAGER: { bg: 'bg-rose-50', text: 'text-rose-700', dot: 'bg-rose-500' },
  MONITOR: { bg: 'bg-cyan-50', text: 'text-cyan-700', dot: 'bg-cyan-500' },
  FINANCE: { bg: 'bg-lime-50', text: 'text-lime-700', dot: 'bg-lime-500' },
  GPS_ADMIN: { bg: 'bg-indigo-50', text: 'text-indigo-700', dot: 'bg-indigo-500' },
  QUALITY_ENGINEER: { bg: 'bg-fuchsia-50', text: 'text-fuchsia-700', dot: 'bg-fuchsia-500' },
  ENGINEER: { bg: 'bg-teal-50', text: 'text-teal-700', dot: 'bg-teal-500' },
  PROCUREMENT_ADMIN: { bg: 'bg-orange-50', text: 'text-orange-700', dot: 'bg-orange-500' },
  DESIGNER: { bg: 'bg-pink-50', text: 'text-pink-700', dot: 'bg-pink-500' },
  SERVICE_MANAGER: { bg: 'bg-teal-50', text: 'text-teal-700', dot: 'bg-teal-500' },
  IT_SUPPORT: { bg: 'bg-slate-100', text: 'text-slate-700', dot: 'bg-slate-600' },
}

const NEUTRAL_CHIP = { bg: 'bg-slate-50', text: 'text-slate-700', dot: 'bg-slate-500' }

export function roleChipColor(role: string | null | undefined) {
  return ROLE_CHIP_COLORS[role as EmployeeRole] ?? NEUTRAL_CHIP
}

/**
 * تدرّج صورة الموظف بترويسة القائمة.
 *
 * ⚠️ خمسة أدوار كانت **ناقصة** من هالخريطة فتطلع بالأزرق الافتراضي —
 * ومنهن **المالك**، يعني حساب (ع) نفسه جان بلا لون خاص. صار ذهبياً
 * الآن، وياه التقني والمصمم ومسؤول الخدمة وإداري الكميات.
 */
export const ROLE_GRADIENTS: Record<EmployeeRole, string> = {
  OWNER: 'from-yellow-500 to-amber-600',
  ADMIN: 'from-amber-500 to-orange-600',
  SALES: 'from-emerald-500 to-teal-600',
  HR_COORDINATOR: 'from-violet-500 to-purple-600',
  TECHNICIAN: 'from-sky-500 to-blue-600',
  TECHNICAL: 'from-blue-500 to-indigo-600',
  PROJECT_MANAGER: 'from-rose-500 to-pink-600',
  MONITOR: 'from-cyan-500 to-teal-600',
  FINANCE: 'from-lime-500 to-green-600',
  GPS_ADMIN: 'from-indigo-500 to-blue-600',
  QUALITY_ENGINEER: 'from-fuchsia-500 to-purple-600',
  ENGINEER: 'from-teal-500 to-cyan-700',
  PROCUREMENT_ADMIN: 'from-orange-500 to-amber-600',
  DESIGNER: 'from-pink-500 to-rose-600',
  SERVICE_MANAGER: 'from-teal-500 to-emerald-600',
  IT_SUPPORT: 'from-slate-500 to-slate-700',
}

export function roleGradient(role: string | null | undefined): string {
  return ROLE_GRADIENTS[role as EmployeeRole] ?? 'from-blue-500 to-indigo-600'
}
