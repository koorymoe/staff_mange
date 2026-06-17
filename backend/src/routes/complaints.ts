import { Router } from 'express'
import { prisma } from '../prisma'

const router = Router()

// GET / - list all complaints with relations
router.get('/', async (_req, res) => {
  const complaints = await prisma.complaint.findMany({
    include: { customer: true, booking: true, createdBy: true, assignedTo: true },
    orderBy: { createdAt: 'desc' },
  })
  res.json(complaints)
})

// POST / - create complaint
router.post('/', async (req, res) => {
  const { customerId, bookingId, description, createdByEmployeeId } = req.body
  if (!customerId || !description || !createdByEmployeeId) {
    return res.status(400).json({ error: 'customerId, description, and createdByEmployeeId are required' })
  }

  const complaint = await prisma.complaint.create({
    data: { customerId, bookingId, description, createdByEmployeeId },
    include: { customer: true, booking: true, createdBy: true },
  })
  res.status(201).json(complaint)
})

// PUT /:id - update complaint
router.put('/:id', async (req, res) => {
  const { status, assignedToEmployeeId, resolution } = req.body
  const complaint = await prisma.complaint.update({
    where: { id: req.params.id },
    data: { status, assignedToEmployeeId, resolution },
    include: { customer: true, booking: true, createdBy: true, assignedTo: true },
  })
  res.json(complaint)
})

// PUT /:id/resolve - mark resolved
router.put('/:id/resolve', async (req, res) => {
  const { resolution } = req.body
  const complaint = await prisma.complaint.update({
    where: { id: req.params.id },
    data: { status: 'RESOLVED', resolution, resolvedAt: new Date() },
    include: { customer: true, booking: true, createdBy: true, assignedTo: true },
  })
  res.json(complaint)
})

export default router
