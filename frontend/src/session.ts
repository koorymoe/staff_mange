import { createContext, useContext } from 'react'
import type { Employee } from './api'

export interface Session {
  employee: Employee | null
  setEmployee: (emp: Employee | null) => void
  permissions: string[]
  gpsServiceId: string | null
}

export const SessionContext = createContext<Session>({
  employee: null,
  setEmployee: () => {},
  permissions: [],
  gpsServiceId: null,
})

// موظف يشد مهارة GPS = عنده مهارة فعالة (canPerform) تحت خدمة GPS، بغض النظر عن دوره الوظيفي
export const hasGpsSkill = (employee: Employee | null, gpsServiceId: string | null) =>
  !!employee && !!gpsServiceId && employee.skills?.some(s => s.canPerform && s.skill.serviceId === gpsServiceId)

export const useSession = () => useContext(SessionContext)

/**
 * hasRole (دازاي 11-01): دوره الأصلي، أو دور فعلي من منصبه بالهيكل (محاسب،
 * إداري كوادر، مدير مشاريع…) — نفس `middleware.HasRole` بالخادم. للفحوصات
 * الي چانت مكتوبة على اسم الدور، حتى صاحب المنصب ياخذ شغل الدور القديم.
 */
/**
 * can (11-04 — محرك الصلاحيات، المرحلة ٢): «يشوف كل صفوف هذا النوع؟» محسوبة
 * بالسيرفر (`authz.SeeAllRules`) — نفس القاعدة الي يطبقها الحارس، فماكو 403 غلط.
 * ترجّع undefined إذا الجلسة قديمة (قبل الحقل) — المنادي يرجع لحسابه القديم.
 */
export const can = (employee: Employee | null | undefined, resource: string): boolean | undefined =>
  employee?.can?.[resource]

export const hasRole = (employee: Employee | null | undefined, ...roles: string[]) =>
  !!employee && (roles.includes(employee.role) || (employee.orgRoles ?? []).some((r) => roles.includes(r)))



/**
 * canAuditFinance: منو يصدر **قرار** التدقيق والمطابقة؟
 *
 * «مرات المحاسب ياخذ إجازة أو يصير عنده ظرف — منو يدقّق الحسابات؟»
 * فصار المفتاح `finance_audit`: المالك ينطيه للموظف الي يريده، وهو
 * الي يشتغل بدل المحاسب.
 *
 * 🔴 **قدرة مو دور**: الشاشات جانت تفحص `role === 'MONITOR'`، فمنح
 * الصلاحية ما جان يغيّر شي بالواجهة. ومنو ما عنده المفتاح **يشوف
 * ولا يعدّل**.
 *
 * ⚠️ ولا تُحسب بـ`useState`: الصلاحيات توصل **بعد** أول رسم (بـ
 * `Layout`)، فأي قيمة أوليّة تنجمّد على `[]` وتضل غلط. قيمة مشتقّة
 * بكل رسم وبس.
 */
export const canAuditFinance = (role?: string, permissions: string[] = [], orgRoles: string[] = []) =>
  // 🔴 `OWNER` جان **ناقصاً**: الخادم يمرّر المالك دائماً
  // (`RequireRole` يتخطاه بلا شرط)، فالواجهة چانت تخفي أزراره
  // وتعرضله «عرض فقط» — يفتح شاشة التدقيق ويحسب إن الشغل ما انفّذ
  // أصلاً، مع إن طلبه ينقبل لو وصل. واجهة تكذب على صاحب النظام.
  role === 'ADMIN' || role === 'OWNER' || role === 'FINANCE' ||
  permissions.includes('finance_audit') || orgRoles.includes('FINANCE')

/**
 * canLinkPartialBooking: منو يقدر يربط حجزين تاريخيين منفصلين كإنجاز
 * جزئي لنفس الشغلة (شغلة طوّلت أكثر من يوم واستوردت كصفوف منفصلة)؟
 *
 * صاحب النظام طلبها صراحة **صلاحية مستقلة يمنحها بعدين لمن يريد**،
 * مو حصراً بدور معيّن — نفس مطابقة `middleware.RequirePermission`
 * بالخادم بالضبط: `ADMIN`/`OWNER` يمرّون دائماً، وغيرهم يحتاج المفتاح
 * `booking_partial_link` بالإيد (بلا افتراض حتى لـ`FINANCE`).
 */
export const canLinkPartialBooking = (role?: string, permissions: string[] = []) =>
  role === 'ADMIN' || role === 'OWNER' || permissions.includes('booking_partial_link')

// ⚠️ **أسماء الأدوار مو هنا**: انتقلن لـ`roleLabels.ts` وياهن ألوان
// الأدوار — مصدر وحيد بنوع محكوم، حتى الدور الجديد يوقّف البناء
// إذا انساه أحد بدل ما يعرض رمزاً إنكليزياً لموظف.

/** (DRY 11-02) مدير النظام أو المالك — بدل `role === 'ADMIN' || actualRole === 'OWNER'` المكتوبة بكل شاشة */
// ⚠️ بالجلسة المالك يتحوّل role=ADMIN + actualRole=OWNER (Layout)، بس صفوف الـAPI
// (قائمة موظفين…) يجي بيها role=OWNER خام — الثلاث حالات محسوبة.
export const isAdminOrOwner = (employee: Pick<Employee, 'role' | 'actualRole'> | null | undefined) =>
  employee?.role === 'ADMIN' || employee?.role === 'OWNER' || employee?.actualRole === 'OWNER'
