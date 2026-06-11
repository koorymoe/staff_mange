import { Router } from 'express'
import { prisma } from '../prisma'

const router = Router()

// GET /api/employees - list employees with their skills
router.get('/', async (_req, res) => {
  const employees = await prisma.employee.findMany({
    include: { skills: { include: { service: true } } },
    orderBy: { name: 'asc' },
  })
  res.json(employees)
})

// GET /api/employees/match?serviceId=xxx - employees able to perform a given service
router.get('/match', async (req, res) => {
  const { serviceId } = req.query
  if (typeof serviceId !== 'string') {
    return res.status(400).json({ error: 'serviceId is required' })
  }

  const employees = await prisma.employee.findMany({
    where: {
      status: 'ACTIVE',
      skills: { some: { serviceId, canPerform: true } },
    },
    include: { skills: { include: { service: true } } },
    orderBy: { name: 'asc' },
  })
  res.json(employees)
})

// GET /api/employees/:id
router.get('/:id', async (req, res) => {
  const employee = await prisma.employee.findUnique({
    where: { id: req.params.id },
    include: { skills: { include: { service: true } } },
  })
  if (!employee) return res.status(404).json({ error: 'Employee not found' })
  res.json(employee)
})

// POST /api/employees - create employee
router.post('/', async (req, res) => {
  const { name, certificate, position, phone } = req.body
  if (!name) return res.status(400).json({ error: 'name is required' })

  const employee = await prisma.employee.create({
    data: { name, certificate, position, phone },
  })
  res.status(201).json(employee)
})

// PUT /api/employees/:id - update employee basic info
router.put('/:id', async (req, res) => {
  const { name, certificate, position, phone, status } = req.body
  const employee = await prisma.employee.update({
    where: { id: req.params.id },
    data: { name, certificate, position, phone, status },
  })
  res.json(employee)
})

// PUT /api/employees/:id/skills - replace employee skill set
// body: { skills: [{ serviceId: string, canPerform: boolean }] }
router.put('/:id/skills', async (req, res) => {
  const { id } = req.params
  const { skills } = req.body as { skills: { serviceId: string; canPerform: boolean }[] }

  if (!Array.isArray(skills)) {
    return res.status(400).json({ error: 'skills must be an array' })
  }

  await prisma.$transaction([
    prisma.employeeSkill.deleteMany({ where: { employeeId: id } }),
    prisma.employeeSkill.createMany({
      data: skills.map((s) => ({
        employeeId: id,
        serviceId: s.serviceId,
        canPerform: s.canPerform,
      })),
    }),
  ])

  const employee = await prisma.employee.findUnique({
    where: { id },
    include: { skills: { include: { service: true } } },
  })
  res.json(employee)
})

export default router
