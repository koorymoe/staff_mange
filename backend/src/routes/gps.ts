import { Router } from 'express'
import { prisma } from '../prisma'

const router = Router()

// ── GPS Customers ───────────────────────────────────────────────────────────

router.get('/customers', async (_req, res) => {
  const customers = await prisma.gpsCustomer.findMany({ orderBy: { fullName: 'asc' } })
  res.json(customers)
})

router.post('/customers', async (req, res) => {
  const customer = await prisma.gpsCustomer.create({ data: req.body })
  res.status(201).json(customer)
})

router.put('/customers/:id', async (req, res) => {
  const customer = await prisma.gpsCustomer.update({ where: { id: req.params.id }, data: req.body })
  res.json(customer)
})

// ── SIM Cards ───────────────────────────────────────────────────────────────

router.get('/sims', async (_req, res) => {
  const sims = await prisma.simCard.findMany({ orderBy: { createdAt: 'desc' } })
  res.json(sims)
})

router.post('/sims', async (req, res) => {
  const sim = await prisma.simCard.create({ data: req.body })
  res.status(201).json(sim)
})

router.put('/sims/:id', async (req, res) => {
  const sim = await prisma.simCard.update({ where: { id: req.params.id }, data: req.body })
  res.json(sim)
})

// ── Device Requests ─────────────────────────────────────────────────────────

router.get('/devices', async (_req, res) => {
  const devices = await prisma.gpsDeviceRequest.findMany({
    include: { customer: true, employee: true, simCard: true },
    orderBy: { createdAt: 'desc' },
  })
  res.json(devices)
})

router.post('/devices', async (req, res) => {
  const device = await prisma.gpsDeviceRequest.create({
    data: req.body,
    include: { customer: true, employee: true, simCard: true },
  })
  res.status(201).json(device)
})

router.put('/devices/:id', async (req, res) => {
  const device = await prisma.gpsDeviceRequest.update({
    where: { id: req.params.id },
    data: req.body,
    include: { customer: true, employee: true, simCard: true },
  })
  res.json(device)
})

// ── Renewals ────────────────────────────────────────────────────────────────

router.get('/renewals', async (_req, res) => {
  const renewals = await prisma.gpsRenewalRequest.findMany({
    include: { customer: true, deviceRequest: true },
    orderBy: { createdAt: 'desc' },
  })
  res.json(renewals)
})

router.post('/renewals', async (req, res) => {
  const renewal = await prisma.gpsRenewalRequest.create({
    data: req.body,
    include: { customer: true, deviceRequest: true },
  })
  res.status(201).json(renewal)
})

router.put('/renewals/:id', async (req, res) => {
  const renewal = await prisma.gpsRenewalRequest.update({
    where: { id: req.params.id },
    data: req.body,
    include: { customer: true, deviceRequest: true },
  })
  res.json(renewal)
})

// ── Maintenance ─────────────────────────────────────────────────────────────

router.get('/maintenance', async (_req, res) => {
  const records = await prisma.gpsMaintenanceRequest.findMany({
    include: { customer: true, employee: true },
    orderBy: { createdAt: 'desc' },
  })
  res.json(records)
})

router.post('/maintenance', async (req, res) => {
  const record = await prisma.gpsMaintenanceRequest.create({
    data: req.body,
    include: { customer: true, employee: true },
  })
  res.status(201).json(record)
})

router.put('/maintenance/:id', async (req, res) => {
  const record = await prisma.gpsMaintenanceRequest.update({
    where: { id: req.params.id },
    data: req.body,
    include: { customer: true, employee: true },
  })
  res.json(record)
})

// ── Settings (prices) ───────────────────────────────────────────────────────

router.get('/settings', async (_req, res) => {
  const settings = await prisma.gpsSubscriptionPrice.findMany()
  res.json(settings)
})

router.put('/settings', async (req, res) => {
  const { id, ...data } = req.body
  const settings = await prisma.gpsSubscriptionPrice.upsert({
    where: { id: id || 'default' },
    update: data,
    create: { id: id || 'default', ...data },
  })
  res.json(settings)
})

// ── Stats ───────────────────────────────────────────────────────────────────

router.get('/stats', async (_req, res) => {
  const [devicesByStatus, totalDevices, totalCustomers, totalSims, availableSims] = await Promise.all([
    prisma.gpsDeviceRequest.groupBy({ by: ['status'], _count: { id: true } }),
    prisma.gpsDeviceRequest.count(),
    prisma.gpsCustomer.count(),
    prisma.simCard.count(),
    prisma.simCard.count({ where: { status: 'AVAILABLE' } }),
  ])

  res.json({
    devicesByStatus: devicesByStatus.map(d => ({ status: d.status, count: d._count.id })),
    totalDevices,
    totalCustomers,
    totalSims,
    availableSims,
    inUseSims: totalSims - availableSims,
  })
})

export default router
