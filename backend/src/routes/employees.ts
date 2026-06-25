import { Router } from 'express'
import bcrypt from 'bcryptjs'
import { prisma } from '../prisma'

const router = Router()

const stripPassword = <T extends { password?: string | null }>(employee: T) => {
  const { password, ...rest } = employee
  return rest
}

// GET /api/employees - list employees with their skills
router.get('/', async (_req, res) => {
  const employees = await prisma.employee.findMany({
    include: { skills: { include: { skill: { include: { service: true } } } } },
    orderBy: { name: 'asc' },
  })
  res.json(employees.map(stripPassword))
})

// GET /api/employees/supervisors - team leaders eligible to supervise a dispatch
router.get('/supervisors', async (_req, res) => {
  const employees = await prisma.employee.findMany({
    where: { AND: [{ status: 'ACTIVE' }, { OR: [{ role: 'PROJECT_MANAGER' }, { isLeader: true }] }] },
    orderBy: { name: 'asc' },
  })
  res.json(employees.map(stripPassword))
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
      onDuty: true,
      role: 'TECHNICIAN',
    },
    include: { skills: { include: { skill: { include: { service: true } } } } },
    orderBy: { name: 'asc' },
  })

  res.json(employees.map(emp => ({
    ...stripPassword(emp),
    hasRequiredSkill: emp.skills.some(s => s.canPerform && s.skill.serviceId === serviceId),
  })))
})

// GET /api/employees/:id
router.get('/:id', async (req, res) => {
  const employee = await prisma.employee.findUnique({
    where: { id: req.params.id },
    include: { skills: { include: { skill: { include: { service: true } } } } },
  })
  if (!employee) return res.status(404).json({ error: 'Employee not found' })
  res.json(stripPassword(employee))
})

// POST /api/employees - create employee
router.post('/', async (req, res) => {
  const { name, certificate, position, phone, username, password } = req.body
  if (!name) return res.status(400).json({ error: 'name is required' })

  if (username) {
    const existing = await prisma.employee.findUnique({ where: { username } })
    if (existing) return res.status(409).json({ error: 'اسم المستخدم مستخدم من قبل' })
  }

  const employee = await prisma.employee.create({
    data: {
      name,
      certificate,
      position,
      phone,
      username: username || null,
      password: password ? bcrypt.hashSync(password, 10) : null,
    },
  })
  res.status(201).json(stripPassword(employee))
})

// PUT /api/employees/:id - update employee basic info
router.put('/:id', async (req, res) => {
  const {
    name,
    certificate,
    position,
    phone,
    status,
    role,
    onDuty,
    username,
    password,
    hasDrivingLicense,
    hasSafetyCertificate,
    isLeader,
  } = req.body

  if (username) {
    const existing = await prisma.employee.findUnique({ where: { username } })
    if (existing && existing.id !== req.params.id) {
      return res.status(409).json({ error: 'اسم المستخدم مستخدم من قبل' })
    }
  }

  const employee = await prisma.employee.update({
    where: { id: req.params.id },
    data: {
      name,
      certificate,
      position,
      phone,
      status,
      role,
      onDuty,
      username: username === undefined ? undefined : username || null,
      password: password ? bcrypt.hashSync(password, 10) : undefined,
      hasDrivingLicense,
      hasSafetyCertificate,
      isLeader,
    },
  })
  res.json(stripPassword(employee))
})

// PUT /api/employees/:id/skills - replace employee skill set
// body: { skills: [{ skillId: string, canPerform: boolean }] }
router.put('/:id/skills', async (req, res) => {
  const { id } = req.params
  const { skills } = req.body as { skills: { skillId: string; canPerform: boolean }[] }

  if (!Array.isArray(skills)) {
    return res.status(400).json({ error: 'skills must be an array' })
  }

  await prisma.$transaction([
    prisma.employeeSkill.deleteMany({ where: { employeeId: id } }),
    prisma.employeeSkill.createMany({
      data: skills.map((s) => ({
        employeeId: id,
        skillId: s.skillId,
        canPerform: s.canPerform,
      })),
    }),
  ])

  const employee = await prisma.employee.findUnique({
    where: { id },
    include: { skills: { include: { skill: { include: { service: true } } } } },
  })
  res.json(stripPassword(employee!))
})

export default router
