import { Router } from 'express'
import { prisma } from '../prisma'

const router = Router()

// GET /training/materials/mine?employeeId=xxx - materials for services assigned to this employee
router.get('/materials/mine', async (req, res) => {
  const { employeeId } = req.query
  if (typeof employeeId !== 'string') return res.status(400).json({ error: 'employeeId is required' })

  const assignments = await prisma.employeeTrainingAssignment.findMany({
    where: { employeeId },
    include: { service: true },
  })
  const serviceIds = assignments.map(a => a.serviceId)

  const materials = serviceIds.length
    ? await prisma.trainingMaterial.findMany({
        where: { serviceId: { in: serviceIds } },
        include: { service: true },
        orderBy: [{ serviceId: 'asc' }, { order: 'asc' }],
      })
    : []

  res.json({ services: assignments.map(a => a.service), materials })
})

// GET /training/assignments/:employeeId - services assigned to an employee for training
router.get('/assignments/:employeeId', async (req, res) => {
  const assignments = await prisma.employeeTrainingAssignment.findMany({
    where: { employeeId: req.params.employeeId },
    include: { service: true },
  })
  res.json(assignments.map(a => a.service))
})

// PUT /training/assignments/:employeeId - replace training service assignments
// body: { serviceIds: string[] }
router.put('/assignments/:employeeId', async (req, res) => {
  const { employeeId } = req.params
  const { serviceIds } = req.body as { serviceIds: string[] }
  if (!Array.isArray(serviceIds)) return res.status(400).json({ error: 'serviceIds must be an array' })

  await prisma.$transaction([
    prisma.employeeTrainingAssignment.deleteMany({ where: { employeeId } }),
    prisma.employeeTrainingAssignment.createMany({
      data: serviceIds.map(serviceId => ({ employeeId, serviceId })),
    }),
  ])

  const assignments = await prisma.employeeTrainingAssignment.findMany({
    where: { employeeId },
    include: { service: true },
  })
  res.json(assignments.map(a => a.service))
})

// GET /training/materials?serviceId=xxx - materials for a service (admin management)
router.get('/materials', async (req, res) => {
  const { serviceId } = req.query
  const materials = await prisma.trainingMaterial.findMany({
    where: serviceId ? { serviceId: String(serviceId) } : undefined,
    include: { service: true },
    orderBy: [{ serviceId: 'asc' }, { order: 'asc' }],
  })
  res.json(materials)
})

// POST /training/materials - add a training material
router.post('/materials', async (req, res) => {
  const { serviceId, title, url, type, order } = req.body
  if (!serviceId || !title || !url) return res.status(400).json({ error: 'serviceId, title and url are required' })

  const material = await prisma.trainingMaterial.create({
    data: { serviceId, title, url, type: type || 'VIDEO', order: order ?? 0 },
    include: { service: true },
  })
  res.status(201).json(material)
})

// PUT /training/materials/:id
router.put('/materials/:id', async (req, res) => {
  const { title, url, type, order } = req.body
  const material = await prisma.trainingMaterial.update({
    where: { id: req.params.id },
    data: { title, url, type, order },
    include: { service: true },
  })
  res.json(material)
})

// DELETE /training/materials/:id
router.delete('/materials/:id', async (req, res) => {
  await prisma.trainingMaterial.delete({ where: { id: req.params.id } })
  res.status(204).end()
})

export default router
