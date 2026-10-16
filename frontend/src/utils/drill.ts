import type { Employee } from '../api'

// ═══ «الرقم = القائمة» — قرار (ع) 10-16 ═══
// كل بطاقة رقم تفتح `/drill/<key>`: نفس الحجوزات الي عدّتها بالضبط.
// ⚠️ canDrill لازم تطابق حارس `GET /api/drill/{key}` بـmain.go حرفياً —
// زر يفتح شي يرفضه الخادم = ٤٠٣ = مخالفة تنسجّل على موظف بريء.
export function canDrill(employee: Employee | null | undefined, permissions: string[]): boolean {
  if (!employee) return false
  const role = employee.actualRole ?? employee.role
  if (['ADMIN', 'OWNER', 'MONITOR', 'HR_COORDINATOR'].includes(role) || employee.role === 'ADMIN') return true
  return ['coordinator', 'view_bookings', 'monitoring'].some((p) => permissions.includes(p))
}

export const drillTo = (key: string) => `/drill/${key}`
