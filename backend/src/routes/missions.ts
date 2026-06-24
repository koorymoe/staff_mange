import { Router } from 'express'
import { prisma } from '../prisma'

const router = Router()

// GET / — all missions (optionally filter by stage, leaderId, employeeId)
router.get('/', async (req, res) => {
  const { stage, leaderId, employeeId } = req.query
  const where: Record<string, unknown> = {}
  if (stage) where.stage = stage
  if (leaderId) where.leaderId = leaderId
  if (employeeId) {
    where.OR = [
      { leaderId: employeeId },
      { memberIds: { has: employeeId as string } },
    ]
  }
  const missions = await prisma.mission.findMany({
    where,
    include: {
      booking: {
        include: {
          customer: true,
          service: true,
          assignments: { include: { employee: true } },
        },
      },
      events: { orderBy: { createdAt: 'asc' } },
    },
    orderBy: { createdAt: 'desc' },
  })

  // attach employee names
  const empIds = new Set<string>()
  missions.forEach(m => { empIds.add(m.leaderId); m.memberIds.forEach(id => empIds.add(id)) })
  const employees = await prisma.employee.findMany({
    where: { id: { in: [...empIds] } },
    select: { id: true, name: true, role: true },
  })
  const empMap = Object.fromEntries(employees.map(e => [e.id, e]))

  res.json(missions.map(m => ({
    ...m,
    leader: empMap[m.leaderId] || null,
    members: m.memberIds.map(id => empMap[id] || { id, name: '?' }),
  })))
})

// GET /:id — single mission
router.get('/:id', async (req, res) => {
  const mission = await prisma.mission.findUnique({
    where: { id: req.params.id },
    include: {
      booking: { include: { customer: true, service: true, assignments: { include: { employee: true } } } },
      events: { orderBy: { createdAt: 'asc' } },
    },
  })
  if (!mission) return res.status(404).json({ error: 'المهمة غير موجودة' })
  res.json(mission)
})

// POST / — create mission from a booking
router.post('/', async (req, res) => {
  const { bookingId, leaderId, memberIds, customerLat, customerLng, customerAddress } = req.body
  if (!bookingId || !leaderId) return res.status(400).json({ error: 'bookingId و leaderId مطلوبين' })

  const count = await prisma.mission.count()
  const code = 'MSN-' + String(count + 1).padStart(4, '0')

  const mission = await prisma.mission.create({
    data: {
      code,
      bookingId,
      leaderId,
      memberIds: memberIds || [],
      customerLat: customerLat ? parseFloat(customerLat) : null,
      customerLng: customerLng ? parseFloat(customerLng) : null,
      customerAddress: customerAddress || null,
      stage: 'ASSIGNED',
    },
    include: { booking: { include: { customer: true } } },
  })
  res.json(mission)
})

// PUT /:id/stage — advance mission stage
router.put('/:id/stage', async (req, res) => {
  const { stage, employeeId, lat, lng, note, estimatedMinutes, distanceKm, stopReason } = req.body
  if (!stage || !employeeId) return res.status(400).json({ error: 'stage و employeeId مطلوبين' })

  const mission = await prisma.mission.findUnique({ where: { id: req.params.id } })
  if (!mission) return res.status(404).json({ error: 'المهمة غير موجودة' })

  const data: Record<string, unknown> = { stage }
  const now = new Date()

  switch (stage) {
    case 'MATERIALS_PREP':
      break
    case 'MATERIALS_READY':
      data.materialsReadyAt = now
      break
    case 'EN_ROUTE':
      data.departedAt = now
      if (lat) data.departureLat = parseFloat(lat)
      if (lng) data.departureLng = parseFloat(lng)
      if (estimatedMinutes) data.estimatedMinutes = parseInt(estimatedMinutes)
      if (distanceKm) data.distanceKm = parseFloat(distanceKm)
      break
    case 'ARRIVED':
      data.arrivedAt = now
      if (lat) data.arrivalLat = parseFloat(lat)
      if (lng) data.arrivalLng = parseFloat(lng)
      if (mission.departedAt) {
        data.actualMinutes = Math.round((now.getTime() - mission.departedAt.getTime()) / 60000)
      }
      break
    case 'WORK_STARTED':
      data.workStartedAt = now
      break
    case 'COMPLETED':
      data.completedAt = now
      break
    case 'STOPPED':
      data.stoppedAt = now
      data.stopReason = stopReason || null
      break
  }

  if (note) data.notes = note

  const [updated] = await prisma.$transaction([
    prisma.mission.update({ where: { id: req.params.id }, data }),
    prisma.missionEvent.create({
      data: {
        missionId: req.params.id,
        employeeId,
        action: stage,
        lat: lat ? parseFloat(lat) : null,
        lng: lng ? parseFloat(lng) : null,
        note: note || null,
      },
    }),
  ])

  const full = await prisma.mission.findUnique({
    where: { id: req.params.id },
    include: {
      booking: { include: { customer: true, service: true } },
      events: { orderBy: { createdAt: 'asc' } },
    },
  })
  res.json(full)
})

// GET /my/:employeeId — missions for a specific employee (leader or member)
router.get('/my/:employeeId', async (req, res) => {
  const { employeeId } = req.params
  const missions = await prisma.mission.findMany({
    where: {
      OR: [
        { leaderId: employeeId },
        { memberIds: { has: employeeId } },
      ],
      stage: { notIn: ['COMPLETED', 'STOPPED'] },
    },
    include: {
      booking: { include: { customer: true, service: true } },
      events: { orderBy: { createdAt: 'asc' } },
    },
    orderBy: { createdAt: 'desc' },
  })
  res.json(missions)
})

// GET /monitor/live — live dashboard for monitor
router.get('/monitor/live', async (_req, res) => {
  const activeMissions = await prisma.mission.findMany({
    where: { stage: { notIn: ['COMPLETED', 'STOPPED'] } },
    include: {
      booking: { include: { customer: true, service: true } },
      events: { orderBy: { createdAt: 'desc' }, take: 1 },
    },
    orderBy: { createdAt: 'desc' },
  })

  const empIds = new Set<string>()
  activeMissions.forEach(m => { empIds.add(m.leaderId); m.memberIds.forEach(id => empIds.add(id)) })
  const employees = await prisma.employee.findMany({
    where: { id: { in: [...empIds] } },
    select: { id: true, name: true },
  })
  const empMap = Object.fromEntries(employees.map(e => [e.id, e]))

  const stats = {
    total: activeMissions.length,
    assigned: activeMissions.filter(m => m.stage === 'ASSIGNED').length,
    preparing: activeMissions.filter(m => ['MATERIALS_PREP', 'MATERIALS_READY'].includes(m.stage)).length,
    enRoute: activeMissions.filter(m => m.stage === 'EN_ROUTE').length,
    arrived: activeMissions.filter(m => m.stage === 'ARRIVED').length,
    working: activeMissions.filter(m => m.stage === 'WORK_STARTED').length,
  }

  res.json({
    missions: activeMissions.map(m => ({
      ...m,
      leader: empMap[m.leaderId] || null,
      members: m.memberIds.map(id => empMap[id] || { id, name: '?' }),
    })),
    stats,
  })
})

// GET /reports/performance — performance report
router.get('/reports/performance', async (req, res) => {
  const { from, to } = req.query
  const where: Record<string, unknown> = {
    stage: { in: ['COMPLETED', 'STOPPED'] },
  }
  if (from || to) {
    where.createdAt = {}
    if (from) (where.createdAt as Record<string, unknown>).gte = new Date(from as string)
    if (to) (where.createdAt as Record<string, unknown>).lte = new Date(to as string)
  }

  const missions = await prisma.mission.findMany({ where })

  const empIds = new Set<string>()
  missions.forEach(m => { empIds.add(m.leaderId); m.memberIds.forEach(id => empIds.add(id)) })
  const employees = await prisma.employee.findMany({
    where: { id: { in: [...empIds] } },
    select: { id: true, name: true },
  })

  const report = employees.map(emp => {
    const empMissions = missions.filter(m => m.leaderId === emp.id || m.memberIds.includes(emp.id))
    const completed = empMissions.filter(m => m.stage === 'COMPLETED')
    const stopped = empMissions.filter(m => m.stage === 'STOPPED')
    const onTime = completed.filter(m => m.estimatedMinutes && m.actualMinutes && m.actualMinutes <= m.estimatedMinutes)
    const late = completed.filter(m => m.estimatedMinutes && m.actualMinutes && m.actualMinutes > m.estimatedMinutes)
    const avgDelay = late.length
      ? Math.round(late.reduce((s, m) => s + ((m.actualMinutes || 0) - (m.estimatedMinutes || 0)), 0) / late.length)
      : 0

    return {
      employee: emp,
      totalMissions: empMissions.length,
      completed: completed.length,
      stopped: stopped.length,
      onTime: onTime.length,
      late: late.length,
      avgDelayMinutes: avgDelay,
      compliancePercent: completed.length ? Math.round((onTime.length / completed.length) * 100) : 0,
    }
  })

  res.json(report.sort((a, b) => b.totalMissions - a.totalMissions))
})

export default router
