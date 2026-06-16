import { Router } from 'express'
import { prisma } from '../prisma'

const router = Router()

const bookingInclude = {
  customer: true,
  service: true,
  transferEmployee: true,
  projectSupervisor: true,
  confirmedByEmployee: true,
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
// body: { confirmedByName, adminNotes?, transferToProjects, quotedPrice?, address?, scheduledAt? }
router.put('/:id/confirm', async (req, res) => {
  const { confirmedByName, confirmedByEmployeeId, adminNotes, transferToProjects, quotedPrice, address, scheduledAt } = req.body

  const booking = await prisma.booking.update({
    where: { id: req.params.id },
    data: {
      status: 'CONFIRMED',
      confirmedByName,
      confirmedByEmployeeId: confirmedByEmployeeId || undefined,
      adminNotes,
      transferToProjects: Boolean(transferToProjects),
      quotedPrice: quotedPrice !== undefined && quotedPrice !== '' ? Number(quotedPrice) : undefined,
      address: address !== undefined ? address : undefined,
      scheduledAt: scheduledAt ? new Date(scheduledAt) : undefined,
    },
    include: bookingInclude,
  })
  res.json(booking)
})

// PUT /api/bookings/:id/details - HR coordinator updates price/address/vehicle after confirmation
// body: { quotedPrice?, address?, assignedVehicle? }
router.put('/:id/details', async (req, res) => {
  const { quotedPrice, address, assignedVehicle } = req.body

  const booking = await prisma.booking.update({
    where: { id: req.params.id },
    data: {
      quotedPrice: quotedPrice !== undefined && quotedPrice !== '' ? Number(quotedPrice) : null,
      address: address !== undefined ? address : undefined,
      assignedVehicle: assignedVehicle !== undefined ? assignedVehicle : undefined,
    },
    include: bookingInclude,
  })
  res.json(booking)
})

// PUT /api/bookings/:id/schedule - HR coordinator sets/changes the appointment time
// If the booking already has a scheduled time, the new time becomes a pending request
// that requires the monitor's approval before it takes effect.
// body: { scheduledAt }
router.put('/:id/schedule', async (req, res) => {
  const { id } = req.params
  const { scheduledAt } = req.body

  if (!scheduledAt) return res.status(400).json({ error: 'scheduledAt is required' })

  const existing = await prisma.booking.findUnique({ where: { id } })
  if (!existing) return res.status(404).json({ error: 'Booking not found' })

  const booking = await prisma.booking.update({
    where: { id },
    data: existing.scheduledAt
      ? { pendingScheduledAt: new Date(scheduledAt) }
      : { scheduledAt: new Date(scheduledAt) },
    include: bookingInclude,
  })
  res.json(booking)
})

// PUT /api/bookings/:id/schedule/approve - monitor approves a pending reschedule request
router.put('/:id/schedule/approve', async (req, res) => {
  const { id } = req.params
  const existing = await prisma.booking.findUnique({ where: { id } })
  if (!existing?.pendingScheduledAt) {
    return res.status(400).json({ error: 'لا يوجد طلب تعديل موعد بانتظار الموافقة' })
  }

  const booking = await prisma.booking.update({
    where: { id },
    data: { scheduledAt: existing.pendingScheduledAt, pendingScheduledAt: null },
    include: bookingInclude,
  })
  res.json(booking)
})

// PUT /api/bookings/:id/schedule/reject - monitor rejects a pending reschedule request
router.put('/:id/schedule/reject', async (req, res) => {
  const { id } = req.params
  const booking = await prisma.booking.update({
    where: { id },
    data: { pendingScheduledAt: null },
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

// PUT /api/bookings/:id/supervisor - HR coordinator optionally assigns a supervisor to the dispatch
// body: { employeeId: string | null }
router.put('/:id/supervisor', async (req, res) => {
  const { id } = req.params
  const { employeeId } = req.body

  if (employeeId) {
    const employee = await prisma.employee.findUnique({ where: { id: employeeId } })
    if (!employee || employee.role !== 'PROJECT_MANAGER') {
      return res.status(400).json({ error: 'يجب أن يكون المشرف من مديري المشاريع' })
    }
  }

  const updated = await prisma.booking.update({
    where: { id },
    data: { projectSupervisorId: employeeId || null },
    include: bookingInclude,
  })
  res.json(updated)
})

// PUT /api/bookings/:id/start - technician acknowledges receipt and starts the booking
router.put('/:id/start', async (req, res) => {
  const booking = await prisma.booking.update({
    where: { id: req.params.id },
    data: { status: 'IN_PROGRESS' },
    include: bookingInclude,
  })
  res.json(booking)
})

// PUT /api/bookings/:id/complete - technician marks the booking as done
// body: { completionNotes?, amountCollected?, advancePaid? }
router.put('/:id/complete', async (req, res) => {
  const { completionNotes, amountCollected, advancePaid } = req.body

  const booking = await prisma.booking.update({
    where: { id: req.params.id },
    data: {
      status: 'COMPLETED',
      completedAt: new Date(),
      completionNotes,
      amountCollected: amountCollected !== undefined && amountCollected !== '' ? Number(amountCollected) : undefined,
      advancePaid: advancePaid !== undefined && advancePaid !== '' ? Number(advancePaid) : undefined,
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
