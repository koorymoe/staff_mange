import { Router } from 'express'
import { prisma } from '../prisma'

const router = Router()

// GET / - list all permissions
router.get('/', async (_req, res) => {
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
