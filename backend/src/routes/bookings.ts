import { Router } from 'express'
import { prisma } from '../prisma'

const router = Router()

const bookingInclude = {
  customer: true,
  service: true,
  transferEmployee: true,
  projectSupervisor: true,
  confirmedByEmployee: true,
  expenseResponsible: true,
  assignments: { include: { employee: true } },
  scheduleLogs: { include: { changedBy: { select: { id: true, name: true, role: true } } }, orderBy: { createdAt: 'desc' as const } },
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
  const { quotedPrice, address, assignedVehicle, mapLocation, mapLatitude, mapLongitude, expenseResponsibleId } = req.body

  const booking = await prisma.booking.update({
    where: { id: req.params.id },
    data: {
      quotedPrice: quotedPrice !== undefined && quotedPrice !== '' ? Number(quotedPrice) : null,
      address: address !== undefined ? address : undefined,
      assignedVehicle: assignedVehicle !== undefined ? assignedVehicle : undefined,
      mapLocation: mapLocation !== undefined ? mapLocation : undefined,
      mapLatitude: mapLatitude !== undefined ? (mapLatitude !== null ? Number(mapLatitude) : null) : undefined,
      mapLongitude: mapLongitude !== undefined ? (mapLongitude !== null ? Number(mapLongitude) : null) : undefined,
      expenseResponsibleId: expenseResponsibleId !== undefined ? expenseResponsibleId || null : undefined,
    },
    include: bookingInclude,
  })
  res.json(booking)
})

// PUT /api/bookings/:id/schedule - HR coordinator sets/changes the appointment time
// body: { scheduledAt, changedById }
router.put('/:id/schedule', async (req, res) => {
  const { id } = req.params
  const { scheduledAt, changedById } = req.body

  if (!scheduledAt) return res.status(400).json({ error: 'scheduledAt is required' })

  const existing = await prisma.booking.findUnique({ where: { id } })
  if (!existing) return res.status(404).json({ error: 'Booking not found' })

  const [booking] = await prisma.$transaction([
    prisma.booking.update({
      where: { id },
      data: { scheduledAt: new Date(scheduledAt) },
      include: bookingInclude,
    }),
    ...(changedById ? [prisma.scheduleChangeLog.create({
      data: {
        bookingId: id,
        changedById,
        oldTime: existing.scheduledAt,
        newTime: new Date(scheduledAt),
      },
    })] : []),
  ])
  res.json(booking)
})

// GET /api/bookings/:id/schedule-log - view schedule change history
router.get('/:id/schedule-log', async (req, res) => {
  const logs = await prisma.scheduleChangeLog.findMany({
    where: { bookingId: req.params.id },
    include: { changedBy: { select: { id: true, name: true, role: true } } },
    orderBy: { createdAt: 'desc' },
  })
  res.json(logs)
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

  let skillWarning = false
  if (booking.serviceId) {
    const matchingSkill = await prisma.employeeSkill.findFirst({
      where: { employeeId, canPerform: true, skill: { serviceId: booking.serviceId } },
    })
    if (!matchingSkill) {
      skillWarning = true
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

  const updateData: Record<string, unknown> = {}
  if (assignedVehicle !== undefined) updateData.assignedVehicle = assignedVehicle
  if (skillWarning) {
    const warning = `⚠️ تنبيه: الموظف ${employee.name} لا يمتلك المهارة اللازمة لهذه الخدمة`
    const existing = booking.adminNotes || ''
    if (!existing.includes(warning)) {
      updateData.adminNotes = existing ? `${existing}\n${warning}` : warning
    }
  }

  // تعيين المسؤول عن المصاريف تلقائياً
  if (!booking.expenseResponsibleId) {
    if (employee.isLeader) {
      updateData.expenseResponsibleId = employeeId
    } else {
      // إذا فني واحد فقط بدون تيم ليدر → هو المسؤول
      const allAssignments = await prisma.bookingAssignment.findMany({
        where: { bookingId: id },
        include: { employee: true },
      })
      const hasLeader = allAssignments.some(a => a.employee.isLeader)
      if (!hasLeader && allAssignments.length === 1) {
        updateData.expenseResponsibleId = allAssignments[0].employeeId
      }
    }
  }

  if (Object.keys(updateData).length > 0) {
    await prisma.booking.update({ where: { id }, data: updateData })
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
    if (!employee || (employee.role !== 'PROJECT_MANAGER' && !employee.isLeader)) {
      return res.status(400).json({ error: 'يجب أن يكون تيم ليدر أو مدير مشاريع' })
    }
  }

  const updated = await prisma.booking.update({
    where: { id },
    data: {
      projectSupervisorId: employeeId || null,
      expenseResponsibleId: employeeId || null,
    },
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
