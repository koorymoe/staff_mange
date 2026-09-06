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

export const roleLabels: Record<string, string> = {
  ADMIN: 'مدير النظام',
  SALES: 'موظف مبيعات',
  HR_COORDINATOR: 'إداري الكوادر',
  TECHNICIAN: 'فني',
  PROJECT_MANAGER: 'مدير مشاريع',
  MONITOR: 'مدقق / مراقب',
  FINANCE: 'محاسب',
  GPS_ADMIN: 'مسؤول GPS',
  QUALITY_ENGINEER: 'مهندس جودة',
  ENGINEER: 'مهندس',
  PROCUREMENT_ADMIN: 'إداري الكميات',
  DESIGNER: 'مصمم',
  SERVICE_MANAGER: 'مسؤول خدمة',
  TECHNICAL: 'تقني',
  OWNER: 'مالك النظام 👑',
}
