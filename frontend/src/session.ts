import { createContext, useContext } from 'react'
import type { Employee } from './api'

export interface Session {
  employee: Employee | null
  setEmployee: (emp: Employee | null) => void
  permissions: string[]
}

export const SessionContext = createContext<Session>({
  employee: null,
  setEmployee: () => {},
  permissions: [],
})

export const useSession = () => useContext(SessionContext)

export const roleLabels: Record<string, string> = {
  ADMIN: 'مدير النظام',
  SALES: 'موظف مبيعات',
  HR_COORDINATOR: 'إداري الكوادر',
  TECHNICIAN: 'فني',
  PROJECT_MANAGER: 'مدير مشاريع',
  MONITOR: 'مراقب',
  FINANCE: 'محاسب',
  GPS_ADMIN: 'مسؤول GPS',
  GPS_ENGINEER: 'مهندس GPS',
  QUALITY_ENGINEER: 'مهندس جودة',
}
