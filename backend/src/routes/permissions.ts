import { Router } from 'express'
import { prisma } from '../prisma'

const router = Router()

const defaultPermissions = [
  { name: 'gps_system', label: 'نظام GPS' },
  { name: 'quotation_system', label: 'نظام عروض الأسعار' },
  { name: 'kpi_management', label: 'تقييم الأداء (KPI)' },
  { name: 'inventory', label: 'جرد الأدوات' },
  { name: 'complaints', label: 'الشكاوى' },
  { name: 'edit_employee_profile', label: 'تعديل ملف الموظف (الراتب/الدوام/الإجازات)' },
  { name: 'view_bookings', label: 'عرض الحجوزات' },
  { name: 'manage_customers', label: 'إدارة العملاء' },
  { name: 'sales_booking', label: 'إنشاء حجز جديد' },
  { name: 'coordinator', label: 'تنسيق الحجوزات' },
  { name: 'finance', label: 'المالية' },
  { name: 'expenses', label: 'المصاريف' },
  { name: 'project_management', label: 'إدارة المشاريع' },
]

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

export default router
