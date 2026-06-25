import { Router } from 'express'
import { prisma } from '../prisma'

const router = Router()

const defaultPermissions = [
  { name: 'staff_management', label: 'إدارة الكوادر' },
  { name: 'edit_employee_profile', label: 'تعديل ملف الموظف (الراتب/الدوام/الإجازات)' },
  { name: 'kpi_management', label: 'تقييم الأداء (KPI)' },
  { name: 'inventory', label: 'جرد الأدوات' },
  { name: 'complaints', label: 'الشكاوى' },
  { name: 'sales_booking', label: 'إنشاء حجز جديد' },
  { name: 'manage_customers', label: 'إدارة العملاء' },
  { name: 'view_bookings', label: 'عرض الحجوزات' },
  { name: 'coordinator', label: 'تنسيق الحجوزات' },
  { name: 'manage_services', label: 'الخدمات' },
  { name: 'mission_tracking', label: 'تتبع المهام' },
  { name: 'gps_system', label: 'نظام GPS' },
  { name: 'project_management', label: 'إدارة المشاريع' },
  { name: 'quotation_system', label: 'نظام عروض الأسعار' },
  { name: 'finance', label: 'المالية' },
  { name: 'expenses', label: 'المصاريف' },
  { name: 'procurement', label: 'المشتريات' },
]

const roleDefaultPermissions: Record<string, string[]> = {
  ADMIN: [],
  SALES: ['sales_booking', 'manage_customers', 'complaints', 'quotation_system'],
  HR_COORDINATOR: ['staff_management', 'edit_employee_profile', 'coordinator', 'manage_customers', 'view_bookings', 'manage_services', 'inventory', 'complaints', 'mission_tracking'],
  TECHNICIAN: ['expenses'],
  PROJECT_MANAGER: ['project_management', 'expenses', 'mission_tracking'],
  MONITOR: ['staff_management', 'kpi_management', 'view_bookings', 'manage_customers', 'manage_services', 'mission_tracking', 'inventory', 'complaints', 'finance'],
  FINANCE: ['finance', 'view_bookings'],
  GPS_ADMIN: ['gps_system'],
  GPS_ENGINEER: ['gps_system'],
  QUALITY_ENGINEER: [],
}

let seeded = false
async function ensurePermissions() {
  if (seeded) return
  for (const perm of defaultPermissions) {
    await prisma.permission.upsert({
      where: { name: perm.name },
      update: { label: perm.label },
      create: perm,
    })
  }
  seeded = true
}

// GET / - list all permissions
router.get('/', async (_req, res) => {
  await ensurePermissions()
  const permissions = await prisma.permission.findMany({ orderBy: { name: 'asc' } })
  res.json(permissions)
})

// GET /employee/:employeeId - get permissions for an employee
router.get('/employee/:employeeId', async (req, res) => {
  const perms = await prisma.employeePermission.findMany({
    where: { employeeId: req.params.employeeId },
    include: { permission: true },
  })
  res.json(perms.map(p => p.permission))
})

// PUT /employee/:employeeId - set permissions for an employee (replace all)
// body: { permissionIds: string[] }
router.put('/employee/:employeeId', async (req, res) => {
  const { employeeId } = req.params
  const { permissionIds } = req.body as { permissionIds: string[] }

  await prisma.$transaction([
    prisma.employeePermission.deleteMany({ where: { employeeId } }),
    prisma.employeePermission.createMany({
      data: permissionIds.map(permissionId => ({ employeeId, permissionId })),
    }),
  ])

  const perms = await prisma.employeePermission.findMany({
    where: { employeeId },
    include: { permission: true },
  })
  res.json(perms.map(p => p.permission))
})

// POST /employee/:employeeId/apply-defaults - apply default permissions for employee's role
router.post('/employee/:employeeId/apply-defaults', async (req, res) => {
  const { employeeId } = req.params
  const employee = await prisma.employee.findUnique({ where: { id: employeeId } })
  if (!employee) return res.status(404).json({ error: 'Employee not found' })

  const defaults = roleDefaultPermissions[employee.role] || []
  if (!defaults.length) return res.json([])

  await ensurePermissions()
  const allPerms = await prisma.permission.findMany()
  const defaultPermIds = allPerms
    .filter(p => defaults.includes(p.name))
    .map(p => p.id)

  const existing = await prisma.employeePermission.findMany({
    where: { employeeId },
    select: { permissionId: true },
  })
  const existingIds = new Set(existing.map(e => e.permissionId))
  const newIds = defaultPermIds.filter(id => !existingIds.has(id))

  if (newIds.length > 0) {
    await prisma.employeePermission.createMany({
      data: newIds.map(permissionId => ({ employeeId, permissionId })),
    })
  }

  const perms = await prisma.employeePermission.findMany({
    where: { employeeId },
    include: { permission: true },
  })
  res.json(perms.map(p => p.permission))
})

// GET /role-defaults - get the default permissions map for each role
router.get('/role-defaults', (_req, res) => {
  res.json(roleDefaultPermissions)
})

export default router
