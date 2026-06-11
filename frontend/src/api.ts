const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:4000/api'

export interface Service {
  id: string
  name: string
  category: string | null
}

export interface EmployeeSkill {
  id: string
  serviceId: string
  canPerform: boolean
  service: Service
}

export interface Employee {
  id: string
  name: string
  certificate: string | null
  position: string | null
  phone: string | null
  status: 'ACTIVE' | 'INACTIVE'
  skills: EmployeeSkill[]
}

export interface Customer {
  id: string
  customerCode: number
  code: string
  name: string
  phone: string
  location: string | null
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `Request failed: ${res.status}`)
  }
  return res.json()
}

export const api = {
  getServices: () => request<Service[]>('/services'),

  getEmployees: () => request<Employee[]>('/employees'),
  createEmployee: (data: Pick<Employee, 'name' | 'certificate' | 'position' | 'phone'>) =>
    request<Employee>('/employees', { method: 'POST', body: JSON.stringify(data) }),
  updateEmployeeSkills: (id: string, skills: { serviceId: string; canPerform: boolean }[]) =>
    request<Employee>(`/employees/${id}/skills`, {
      method: 'PUT',
      body: JSON.stringify({ skills }),
    }),
  matchEmployees: (serviceId: string) =>
    request<Employee[]>(`/employees/match?serviceId=${serviceId}`),

  getCustomers: () => request<Customer[]>('/customers'),
  createCustomer: (data: { name: string; phone: string; location?: string }) =>
    request<Customer & { existed: boolean }>('/customers', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
}
