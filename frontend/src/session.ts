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

export const hasMonitorAccess = (role?: string, permissions: string[] = []) =>
  role === 'ADMIN' || role === 'MONITOR' || permissions.includes('monitoring')

export const hasAuditAccess = (role?: string, permissions: string[] = []) =>
  role === 'ADMIN' || role === 'MONITOR' || permissions.includes('auditing')

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
export const canAuditFinance = (role?: string, permissions: string[] = []) =>
  role === 'ADMIN' || role === 'FINANCE' || permissions.includes('finance_audit')

// ⚠️ **أسماء الأدوار مو هنا**: انتقلن لـ`roleLabels.ts` وياهن ألوان
// الأدوار — مصدر وحيد بنوع محكوم، حتى الدور الجديد يوقّف البناء
// إذا انساه أحد بدل ما يعرض رمزاً إنكليزياً لموظف.
