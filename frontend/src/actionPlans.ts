// (ك) 11-28 زينا: «معالجة الحالات الحرجة» — خطط للموظف الي عنده مشكلة بالشغل
import { apiRequest } from './api'

export type APProblem = 'SLOW' | 'ACCURACY' | 'TIME_WASTE' | 'BEHAVIOR' | 'OTHER'
export type APSource = 'MATRIX' | 'MANAGER' | 'OWNER'
export type APStatus = 'OPEN' | 'RESOLVED' | 'NOT_RESOLVED'
export type ActionPlan = {
  id: string; employeeId: string; employeeName: string; position: string | null
  problem: APProblem; details: string; source: APSource; target: string; steps: string
  startDate: string; dueDate: string; daysLeft: number; status: APStatus
  outcome: string | null; nextAction: string | null; createdBy: string; createdAt: string
  closedBy: string | null; closedAt: string | null; notes: number
}
export type APNote = { id: string; note: string; by: string; createdAt: string }

export const PROBLEMS: Record<APProblem, string> = {
  SLOW: 'بطء بالعمل', ACCURACY: 'قلة دقة', TIME_WASTE: 'تضييع وقت', BEHAVIOR: 'سلوك والتزام', OTHER: 'غير ذلك',
}
export const SOURCES: Record<APSource, string> = { MATRIX: 'تقرير ماتركس', MANAGER: 'المدير', OWNER: 'المالك' }

const post = <T,>(path: string, body: unknown) => apiRequest<T>(path, { method: 'POST', body: JSON.stringify(body) })
export const actionPlansApi = {
  list: () => apiRequest<ActionPlan[]>('/action-plans'),
  employees: () => apiRequest<{ id: string; name: string; jobTitle: string | null }[]>('/action-plans/employees'),
  notes: (id: string) => apiRequest<APNote[]>(`/action-plans/${id}/notes`),
  create: (b: { employeeId: string; problem: APProblem; source: APSource; details: string; target: string; steps: string; dueDate: string }) =>
    post<{ id: string }>('/action-plans', b),
  addNote: (id: string, note: string) => post<{ ok: boolean }>(`/action-plans/${id}/notes`, { note }),
  extend: (id: string, dueDate: string, reason: string) => post<{ ok: boolean }>(`/action-plans/${id}/extend`, { dueDate, reason }),
  close: (id: string, status: 'RESOLVED' | 'NOT_RESOLVED', outcome: string, nextAction: string) =>
    post<{ ok: boolean }>(`/action-plans/${id}/close`, { status, outcome, nextAction }),
}
