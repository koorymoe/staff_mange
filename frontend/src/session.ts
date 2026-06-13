import { createContext, useContext } from 'react'
import type { Employee } from './api'

export interface Session {
  employee: Employee | null
  setEmployee: (emp: Employee | null) => void
}

export const SessionContext = createContext<Session>({
  employee: null,
  setEmployee: () => {},
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
}
