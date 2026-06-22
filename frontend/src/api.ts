const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:4000/api'

export interface Skill {
  id: string
  name: string
  serviceId: string
}

export interface Service {
  id: string
  name: string
  category: string | null
  skills: Skill[]
}

export interface EmployeeSkill {
  id: string
  skillId: string
  canPerform: boolean
  skill: Skill & { service: Service }
}

export type EmployeeRole =
  | 'ADMIN'
  | 'SALES'
  | 'HR_COORDINATOR'
  | 'TECHNICIAN'
  | 'PROJECT_MANAGER'
  | 'MONITOR'
  | 'FINANCE'
  | 'GPS_ADMIN'
  | 'GPS_ENGINEER'
  | 'QUALITY_ENGINEER'

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
  hasDrivingLicense: boolean
  hasSafetyCertificate: boolean
  skills: EmployeeSkill[]
  hasRequiredSkill?: boolean
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
  scheduledAt: string | null
  pendingScheduledAt: string | null
  customer: Customer
  service: Service | null
  transferEmployee: Employee | null
  projectSupervisor: Employee | null
  confirmedByEmployee: Employee | null
  notes: string | null
  vehicleType: string | null
  priority: 'NORMAL' | 'URGENT'
  status: 'PENDING' | 'CONFIRMED' | 'IN_PROGRESS' | 'COMPLETED' | 'CANCELLED'
  transferToProjects: boolean
  confirmedByName: string | null
  adminNotes: string | null
  assignedVehicle: string | null
  quotedPrice: number | null
  address: string | null
  completedAt: string | null
  completionNotes: string | null
  amountCollected: number | null
  advancePaid: number | null
  amountVerified: boolean
  assignments: BookingAssignment[]
  createdAt: string
}

export interface Stats {
  totals: {
    totalCustomers: number
    totalBookings: number
    pendingBookings: number
    confirmedBookings: number
    completedBookings: number
    cancelledBookings: number
    urgentPending: number
    totalRevenue: number
    unverifiedRevenue: number
  }
  salesStats: { employeeId: string; name: string; totalTransferred: number; confirmed: number; today: number; thisMonth: number }[]
  coordinatorStats: { employeeId: string; name: string; totalConfirmed: number; today: number; thisMonth: number }[]
  technicianStats: {
    employeeId: string
    name: string
    onDuty: boolean
    totalAssigned: number
    completed: number
    revenueHandled: number
  }[]
  serviceBreakdown: { serviceId: string | null; name: string; count: number }[]
  roleCounts: { role: EmployeeRole; count: number }[]
  recentBookings: {
    id: string
    code: string
    status: string
    priority: string
    customerName: string
    serviceName: string | null
    createdAt: string
  }[]
}

export interface Expense {
  id: string
  employeeId: string
  employee: { id: string; name: string; position: string | null }
  amount: number
  description: string | null
  status: 'PENDING' | 'APPROVED' | 'REJECTED'
  createdAt: string
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

export interface Product {
  id: string
  name: string
  unit: string
  defaultPrice: number
  imageBase64?: string
}

export interface QuotationItem {
  id?: string
  productName: string
  unit: string
  quantity: number
  unitPrice: number
  totalPrice: number
}

export interface Quotation {
  id: string
  quotationNumber: string
  customerName: string
  customerPhone: string | null
  customerAddress: string | null
  projectName: string | null
  items: QuotationItem[]
  grandTotal: number
  discountPercent: number
  discountValue: number
  netTotal: number
  duration: string | null
  notes: string | null
  status: 'NEW' | 'SENT' | 'ACCEPTED' | 'REJECTED'
  createdAt: string
}

export interface Permission {
  id: string
  name: string
  label: string
}

export interface KpiEvaluation {
  id: string
  employeeId: string
  employee: { id: string; name: string }
  evaluatorId: string
  evaluator: { id: string; name: string }
  points: number
  reason: string
  deductionAmount: number
  createdAt: string
}

export interface CartItem {
  id: string
  bookingId: string
  productName: string
  quantity: number
  unitPrice: number
  totalPrice: number
  notes: string | null
  createdAt: string
}

export interface PersonalTool {
  id: string
  employeeId: string
  employee?: { id: string; name: string }
  name: string
  barcode: string
  status: 'AVAILABLE' | 'CHECKED_OUT' | 'DAMAGED'
  checkedOut: boolean
}

export interface VehicleTool {
  id: string
  name: string
  barcode: string
  vehicleId: string
  status: 'AVAILABLE' | 'CHECKED_OUT' | 'DAMAGED'
}

export interface OnDemandTool {
  id: string
  name: string
  barcode: string
  totalQuantity: number
  availableQuantity: number
  status: 'AVAILABLE' | 'CHECKED_OUT' | 'DAMAGED'
}

export type ToolRequest = ToolRequestItem
export interface ToolRequestItem {
  id: string
  employeeId: string
  employee?: { id: string; name: string }
  toolId: string
  tool?: OnDemandTool
  status: 'PENDING' | 'APPROVED' | 'REJECTED' | 'RETURNED'
  approvedById: string | null
  requestedAt: string
  returnedAt: string | null
}

export interface Complaint {
  id: string
  customerId: string
  customer: { id: string; name: string; phone: string }
  bookingId: string | null
  description: string
  status: 'NEW' | 'IN_PROGRESS' | 'RESOLVED' | 'CLOSED'
  createdByEmployeeId: string
  createdByEmployee: { id: string; name: string }
  assignedToEmployeeId: string | null
  assignedToEmployee: { id: string; name: string } | null
  resolution: string | null
  createdAt: string
  resolvedAt: string | null
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
  updateEmployeeSkills: (id: string, skills: { skillId: string; canPerform: boolean }[]) =>
    request<Employee>(`/employees/${id}/skills`, {
      method: 'PUT',
      body: JSON.stringify({ skills }),
    }),
  matchEmployees: (serviceId: string) =>
    request<Employee[]>(`/employees/match?serviceId=${serviceId}`),
  getSupervisors: () => request<Employee[]>('/employees/supervisors'),

  getCustomers: () => request<Customer[]>('/customers'),
  lookupCustomer: (phone: string) =>
    request<Customer | null>(`/customers/lookup?phone=${phone}`).catch(() => null),
  createCustomer: (data: { name: string; phone: string; location?: string }) =>
    request<Customer & { existed: boolean }>('/customers', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  getBookings: (params?: { status?: Booking['status']; customerId?: string }) => {
    const query = new URLSearchParams()
    if (params?.status) query.set('status', params.status)
    if (params?.customerId) query.set('customerId', params.customerId)
    const qs = query.toString()
    return request<Booking[]>(`/bookings${qs ? `?${qs}` : ''}`)
  },
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
    data: {
      confirmedByName: string
      confirmedByEmployeeId?: string
      adminNotes?: string
      transferToProjects: boolean
      quotedPrice?: number
      address?: string
      scheduledAt?: string
    },
  ) => request<Booking>(`/bookings/${id}/confirm`, { method: 'PUT', body: JSON.stringify(data) }),
  scheduleBooking: (id: string, scheduledAt: string) =>
    request<Booking>(`/bookings/${id}/schedule`, { method: 'PUT', body: JSON.stringify({ scheduledAt }) }),
  approveReschedule: (id: string) =>
    request<Booking>(`/bookings/${id}/schedule/approve`, { method: 'PUT', body: JSON.stringify({}) }),
  rejectReschedule: (id: string) =>
    request<Booking>(`/bookings/${id}/schedule/reject`, { method: 'PUT', body: JSON.stringify({}) }),
  updateBookingDetails: (
    id: string,
    data: { quotedPrice?: number | null; address?: string; assignedVehicle?: string },
  ) => request<Booking>(`/bookings/${id}/details`, { method: 'PUT', body: JSON.stringify(data) }),
  assignTechnician: (
    id: string,
    data: { employeeId: string; role: 'TECH_1' | 'TECH_2' | 'TECH_3'; assignedVehicle?: string },
  ) => request<Booking>(`/bookings/${id}/assign`, { method: 'PUT', body: JSON.stringify(data) }),
  assignSupervisor: (id: string, employeeId: string | null) =>
    request<Booking>(`/bookings/${id}/supervisor`, {
      method: 'PUT',
      body: JSON.stringify({ employeeId }),
    }),
  completeBooking: (
    id: string,
    data: { completionNotes?: string; amountCollected?: number; advancePaid?: number },
  ) => request<Booking>(`/bookings/${id}/complete`, { method: 'PUT', body: JSON.stringify(data) }),
  startBooking: (id: string) =>
    request<Booking>(`/bookings/${id}/start`, { method: 'PUT', body: JSON.stringify({}) }),
  verifyAmount: (id: string) =>
    request<Booking>(`/bookings/${id}/verify`, { method: 'PUT', body: JSON.stringify({}) }),
  getStats: () => request<Stats>('/stats'),

  getExpenses: (employeeId?: string) =>
    request<Expense[]>(`/expenses${employeeId ? `?employeeId=${employeeId}` : ''}`),
  createExpense: (data: { employeeId: string; amount: number; description?: string }) =>
    request<Expense>('/expenses', { method: 'POST', body: JSON.stringify(data) }),
  updateExpenseStatus: (id: string, status: 'APPROVED' | 'REJECTED') =>
    request<Expense>(`/expenses/${id}/status`, {
      method: 'PUT',
      body: JSON.stringify({ status }),
    }),

  updateEmployee: (
    id: string,
    data: Partial<
      Pick<
        Employee,
        'role' | 'onDuty' | 'status' | 'name' | 'position' | 'hasDrivingLicense' | 'hasSafetyCertificate'
      >
    > & {
      username?: string
      password?: string
    },
  ) => request<Employee>(`/employees/${id}`, { method: 'PUT', body: JSON.stringify(data) }),

  // GPS
  getGpsStats: () => request<any>('/gps/stats'),
  getGpsCustomers: () => request<any[]>('/gps/customers'),
  createGpsCustomer: (data: any) =>
    request<any>('/gps/customers', { method: 'POST', body: JSON.stringify(data) }),
  getGpsDevices: () => request<any[]>('/gps/devices'),
  createGpsDevice: (data: any) =>
    request<any>('/gps/devices', { method: 'POST', body: JSON.stringify(data) }),
  updateGpsDevice: (id: string, data: any) =>
    request<any>(`/gps/devices/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  getSimCards: () => request<any[]>('/gps/sims'),
  createSimCard: (data: any) =>
    request<any>('/gps/sims', { method: 'POST', body: JSON.stringify(data) }),
  updateSimCard: (id: string, data: any) =>
    request<any>(`/gps/sims/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  getGpsRenewals: () => request<any[]>('/gps/renewals'),
  createGpsRenewal: (data: any) =>
    request<any>('/gps/renewals', { method: 'POST', body: JSON.stringify(data) }),
  updateGpsRenewal: (id: string, data: any) =>
    request<any>(`/gps/renewals/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  getGpsMaintenance: () => request<any[]>('/gps/maintenance'),
  createGpsMaintenance: (data: any) =>
    request<any>('/gps/maintenance', { method: 'POST', body: JSON.stringify(data) }),
  updateGpsMaintenance: (id: string, data: any) =>
    request<any>(`/gps/maintenance/${id}`, { method: 'PUT', body: JSON.stringify(data) }),

  // Products
  getProducts: () => request<Product[]>('/products'),
  createProduct: (data: Omit<Product, 'id'> & { imageBase64?: string }) =>
    request<Product>('/products', { method: 'POST', body: JSON.stringify(data) }),
  updateProduct: (id: string, data: Partial<Omit<Product, 'id'>> & { imageBase64?: string }) =>
    request<Product>(`/products/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteProduct: (id: string) =>
    request<void>(`/products/${id}`, { method: 'DELETE' }),

  // Quotations
  getQuotations: () => request<Quotation[]>('/quotations'),
  createQuotation: (data: {
    customerName: string
    customerPhone?: string
    customerAddress?: string
    projectName?: string
    items: Omit<QuotationItem, 'id'>[]
    grandTotal: number
    discountPercent: number
    discountValue: number
    netTotal: number
    duration?: string
    notes?: string
    createdByEmployeeId?: string
  }) => request<Quotation>('/quotations', { method: 'POST', body: JSON.stringify(data) }),
  deleteQuotation: (id: string) =>
    request<void>(`/quotations/${id}`, { method: 'DELETE' }),

  // Permissions
  getPermissions: () => request<Permission[]>('/permissions'),
  getEmployeePermissions: (employeeId: string) =>
    request<Permission[]>(`/permissions/employee/${employeeId}`),
  setEmployeePermissions: (employeeId: string, permissionIds: string[]) =>
    request<Permission[]>(`/permissions/employee/${employeeId}`, {
      method: 'PUT',
      body: JSON.stringify({ permissionIds }),
    }),

  // KPI
  getKpiEvaluations: () => request<KpiEvaluation[]>('/kpi'),
  getEmployeeKpi: (employeeId: string) => request<KpiEvaluation[]>(`/kpi/employee/${employeeId}`),
  createKpiEvaluation: (data: { employeeId: string; evaluatorId: string; points: number; reason: string }) =>
    request<KpiEvaluation>('/kpi', { method: 'POST', body: JSON.stringify(data) }),
  deleteKpiEvaluation: (id: string) => request<void>(`/kpi/${id}`, { method: 'DELETE' }),

  // Cart
  getCartItems: (bookingId: string) => request<CartItem[]>(`/cart/booking/${bookingId}`),
  addCartItem: (bookingId: string, data: { productName: string; quantity: number; unitPrice: number; notes?: string }) =>
    request<CartItem>(`/cart/booking/${bookingId}`, { method: 'POST', body: JSON.stringify(data) }),
  updateCartItem: (id: string, data: Partial<CartItem>) =>
    request<CartItem>(`/cart/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteCartItem: (id: string) => request<void>(`/cart/${id}`, { method: 'DELETE' }),

  // Inventory
  getPersonalTools: (employeeId?: string) =>
    request<PersonalTool[]>(`/inventory/personal${employeeId ? `?employeeId=${employeeId}` : ''}`),
  createPersonalTool: (data: { employeeId: string; name: string; barcode: string }) =>
    request<PersonalTool>('/inventory/personal', { method: 'POST', body: JSON.stringify(data) }),
  updatePersonalTool: (id: string, data: Partial<PersonalTool>) =>
    request<PersonalTool>(`/inventory/personal/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deletePersonalTool: (id: string) => request<void>(`/inventory/personal/${id}`, { method: 'DELETE' }),
  getVehicleTools: (vehicleId?: string) =>
    request<VehicleTool[]>(`/inventory/vehicle${vehicleId ? `?vehicleId=${vehicleId}` : ''}`),
  createVehicleTool: (data: { name: string; barcode: string; vehicleId: string }) =>
    request<VehicleTool>('/inventory/vehicle', { method: 'POST', body: JSON.stringify(data) }),
  updateVehicleTool: (id: string, data: Partial<VehicleTool>) =>
    request<VehicleTool>(`/inventory/vehicle/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteVehicleTool: (id: string) => request<void>(`/inventory/vehicle/${id}`, { method: 'DELETE' }),
  getOnDemandTools: () => request<OnDemandTool[]>('/inventory/ondemand'),
  createOnDemandTool: (data: { name: string; barcode: string; totalQuantity: number }) =>
    request<OnDemandTool>('/inventory/ondemand', { method: 'POST', body: JSON.stringify(data) }),
  updateOnDemandTool: (id: string, data: Partial<OnDemandTool>) =>
    request<OnDemandTool>(`/inventory/ondemand/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  getToolRequests: (employeeId?: string) =>
    request<ToolRequestItem[]>(`/inventory/requests${employeeId ? `?employeeId=${employeeId}` : ''}`),
  createToolRequest: (data: { employeeId: string; toolId: string }) =>
    request<ToolRequestItem>('/inventory/requests', { method: 'POST', body: JSON.stringify(data) }),
  approveToolRequest: (id: string, approvedById: string) =>
    request<ToolRequestItem>(`/inventory/requests/${id}/approve`, { method: 'PUT', body: JSON.stringify({ approvedById }) }),
  rejectToolRequest: (id: string) =>
    request<ToolRequestItem>(`/inventory/requests/${id}/reject`, { method: 'PUT', body: JSON.stringify({}) }),
  returnToolRequest: (id: string) =>
    request<ToolRequestItem>(`/inventory/requests/${id}/return`, { method: 'PUT', body: JSON.stringify({}) }),

  // Complaints
  getComplaints: () => request<Complaint[]>('/complaints'),
  createComplaint: (data: { customerId: string; bookingId?: string; description: string; createdByEmployeeId: string }) =>
    request<Complaint>('/complaints', { method: 'POST', body: JSON.stringify(data) }),
  updateComplaint: (id: string, data: { status?: string; assignedToEmployeeId?: string; resolution?: string }) =>
    request<Complaint>(`/complaints/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  resolveComplaint: (id: string, resolution: string) =>
    request<Complaint>(`/complaints/${id}/resolve`, { method: 'PUT', body: JSON.stringify({ resolution }) }),

  // GPS Settings
  getGpsSettings: () => request<any[]>('/gps/settings'),
  updateGpsSettings: (data: any) =>
    request<any>('/gps/settings', { method: 'PUT', body: JSON.stringify(data) }),
  updateGpsCustomer: (id: string, data: any) =>
    request<any>(`/gps/customers/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
}
