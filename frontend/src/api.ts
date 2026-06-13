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

export type EmployeeRole =
  | 'ADMIN'
  | 'SALES'
  | 'HR_COORDINATOR'
  | 'TECHNICIAN'
  | 'PROJECT_MANAGER'
  | 'MONITOR'
  | 'FINANCE'

export interface Employee {
  id: string
  name: string
  certificate: string | null
  position: string | null
  phone: string | null
  status: 'ACTIVE' | 'INACTIVE'
  role: EmployeeRole
  onDuty: boolean
  username: string | null
  skills: EmployeeSkill[]
}

export interface BookingAssignment {
  id: string
  role: 'TECH_1' | 'TECH_2' | 'TECH_3'
  employee: Employee
}

export interface Booking {
  id: string
  code: string
  sequenceNumber: number | null
  customer: Customer
  service: Service | null
  notes: string | null
  vehicleType: string | null
  priority: 'NORMAL' | 'URGENT'
  status: 'PENDING' | 'CONFIRMED' | 'COMPLETED' | 'CANCELLED'
  transferToProjects: boolean
  confirmedByName: string | null
  adminNotes: string | null
  assignedVehicle: string | null
  completedAt: string | null
  completionNotes: string | null
  amountCollected: number | null
  amountVerified: boolean
  assignments: BookingAssignment[]
  createdAt: string
}

export interface Stats {
  totalCustomers: number
  totalBookings: number
  confirmedBookings: number
  completedBookings: number
  totalRevenue: number
  salesStats: { employeeId: string; name: string; totalTransferred: number; confirmed: number }[]
  technicianStats: { employeeId: string; name: string; totalAssigned: number; completed: number }[]
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
  createService: (data: { name: string; category?: string }) =>
    request<Service>('/services', { method: 'POST', body: JSON.stringify(data) }),

  getEmployees: () => request<Employee[]>('/employees'),
  createEmployee: (
    data: Pick<Employee, 'name' | 'certificate' | 'position' | 'phone'> & {
      username?: string
      password?: string
    },
  ) => request<Employee>('/employees', { method: 'POST', body: JSON.stringify(data) }),
  login: (username: string, password: string) =>
    request<Employee>('/auth/login', { method: 'POST', body: JSON.stringify({ username, password }) }),
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

  getBookings: (status?: Booking['status']) =>
    request<Booking[]>(`/bookings${status ? `?status=${status}` : ''}`),
  createBooking: (data: {
    customerId: string
    serviceId?: string
    notes?: string
    vehicleType?: string
    priority?: 'NORMAL' | 'URGENT'
    transferEmployeeId?: string
  }) => request<Booking>('/bookings', { method: 'POST', body: JSON.stringify(data) }),
  confirmBooking: (
    id: string,
    data: { confirmedByName: string; adminNotes?: string; transferToProjects: boolean },
  ) => request<Booking>(`/bookings/${id}/confirm`, { method: 'PUT', body: JSON.stringify(data) }),
  assignTechnician: (
    id: string,
    data: { employeeId: string; role: 'TECH_1' | 'TECH_2' | 'TECH_3'; assignedVehicle?: string },
  ) => request<Booking>(`/bookings/${id}/assign`, { method: 'PUT', body: JSON.stringify(data) }),
  completeBooking: (id: string, data: { completionNotes?: string; amountCollected?: number }) =>
    request<Booking>(`/bookings/${id}/complete`, { method: 'PUT', body: JSON.stringify(data) }),
  verifyAmount: (id: string) =>
    request<Booking>(`/bookings/${id}/verify`, { method: 'PUT', body: JSON.stringify({}) }),
  getStats: () => request<Stats>('/stats'),

  updateEmployee: (
    id: string,
    data: Partial<Pick<Employee, 'role' | 'onDuty' | 'status' | 'name' | 'position'>> & {
      username?: string
      password?: string
    },
  ) => request<Employee>(`/employees/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
}
