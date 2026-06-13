import { Router } from 'express'
import { prisma } from '../prisma'

const router = Router()

const bookingInclude = {
  customer: true,
  service: true,
  transferEmployee: true,
  assignments: { include: { employee: true } },
} as const

// GET /api/bookings?status=PENDING - list bookings, optionally filtered by status
router.get('/', async (req, res) => {
  const { status, customerId } = req.query
  const bookings = await prisma.booking.findMany({
    where: {
      ...(status ? { status: status as 'PENDING' | 'CONFIRMED' | 'COMPLETED' | 'CANCELLED' } : {}),
      ...(customerId ? { customerId: customerId as string } : {}),
    },
    include: bookingInclude,
    orderBy: { createdAt: 'desc' },
  })
  res.json(bookings)
})

// POST /api/bookings - sales employee registers a new booking request (ما قبل الحجز)
router.post('/', async (req, res) => {
  const { customerId, serviceId, notes, vehicleType, priority, transferEmployeeId } = req.body

  if (!customerId) return res.status(400).json({ error: 'customerId is required' })

  const lastBooking = await prisma.booking.findFirst({
    orderBy: { sequenceNumber: 'desc' },
    select: { sequenceNumber: true },
  })
  const sequenceNumber = (lastBooking?.sequenceNumber || 0) + 1

  const code = `B${sequenceNumber}`
  const booking = await prisma.booking.create({
    data: {
      code,
      sequenceNumber,
      customerId,
      serviceId,
      notes,
      vehicleType,
      priority,
      transferEmployeeId,
    },
    include: bookingInclude,
  })
  res.status(201).json(booking)
})

// PUT /api/bookings/:id/confirm - HR coordinator confirms the booking with the customer
// body: { confirmedByName, adminNotes?, transferToProjects }
router.put('/:id/confirm', async (req, res) => {
  const { confirmedByName, adminNotes, transferToProjects } = req.body

  const booking = await prisma.booking.update({
    where: { id: req.params.id },
    data: {
      status: 'CONFIRMED',
      confirmedByName,
      adminNotes,
      transferToProjects: Boolean(transferToProjects),
    },
    include: bookingInclude,
  })
  res.json(booking)
})

// PUT /api/bookings/:id/assign - HR coordinator assigns a technician to the crew dispatch
// body: { employeeId, role: 'TECH_1' | 'TECH_2' | 'TECH_3' }
router.put('/:id/assign', async (req, res) => {
  const { id } = req.params
  const { employeeId, role, assignedVehicle } = req.body

  if (!employeeId || !role) {
    return res.status(400).json({ error: 'employeeId and role are required' })
  }

  // Verify the employee can perform the booking's service and is currently on duty
  const booking = await prisma.booking.findUnique({ where: { id } })
  if (!booking) return res.status(404).json({ error: 'Booking not found' })

  if (booking.serviceId) {
    const matchingSkill = await prisma.employeeSkill.findFirst({
      where: { employeeId, canPerform: true, skill: { serviceId: booking.serviceId } },
    })
    if (!matchingSkill) {
      return res.status(400).json({ error: 'هذا الموظف لا يمتلك المهارة اللازمة لهذه المهمة' })
    }
  }

  const employee = await prisma.employee.findUnique({ where: { id: employeeId } })
  if (!employee?.onDuty || employee.status !== 'ACTIVE') {
    return res.status(400).json({ error: 'هذا الموظف غير متاح حالياً (خارج الدوام)' })
  }

  await prisma.bookingAssignment.upsert({
    where: { bookingId_role: { bookingId: id, role } },
    update: { employeeId },
    create: { bookingId: id, employeeId, role },
  })

  if (assignedVehicle !== undefined) {
    await prisma.booking.update({
      where: { id },
      data: { assignedVehicle },
    })
  }

  const updated = await prisma.booking.findUnique({
    where: { id },
    include: bookingInclude,
  })
  res.json(updated)
})

// PUT /api/bookings/:id/complete - technician marks the booking as done
// body: { completionNotes?, amountCollected? }
router.put('/:id/complete', async (req, res) => {
  const { completionNotes, amountCollected } = req.body

  const booking = await prisma.booking.update({
    where: { id: req.params.id },
    data: {
      status: 'COMPLETED',
      completedAt: new Date(),
      completionNotes,
      amountCollected: amountCollected !== undefined ? Number(amountCollected) : undefined,
    },
    include: bookingInclude,
  })
  res.json(booking)
})

// PUT /api/bookings/:id/verify - finance employee verifies the collected amount
router.put('/:id/verify', async (req, res) => {
  const booking = await prisma.booking.update({
    where: { id: req.params.id },
    data: { amountVerified: true },
    include: bookingInclude,
  })
  res.json(booking)
})

export default router
