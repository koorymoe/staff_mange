import { Router } from 'express'
import { prisma } from '../prisma'

const router = Router()

// ── GPS Customers ───────────────────────────────────────────────────────────

router.get('/customers', async (_req, res) => {
  const customers = await prisma.gpsCustomer.findMany({ orderBy: { name: 'asc' } })
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
  const devices = await prisma.gpsDevice.findMany({
    include: { customer: true, employee: true, simCard: true },
    orderBy: { createdAt: 'desc' },
  })
  res.json(devices)
})

router.post('/devices', async (req, res) => {
  const device = await prisma.gpsDevice.create({
    data: req.body,
    include: { customer: true, employee: true, simCard: true },
  })
  res.status(201).json(device)
})

router.put('/devices/:id', async (req, res) => {
  const device = await prisma.gpsDevice.update({
    where: { id: req.params.id },
    data: req.body,
    include: { customer: true, employee: true, simCard: true },
  })
  res.json(device)
})

// ── Renewals ────────────────────────────────────────────────────────────────

router.get('/renewals', async (_req, res) => {
  const renewals = await prisma.gpsRenewal.findMany({
    include: { customer: true, device: true },
    orderBy: { createdAt: 'desc' },
  })
  res.json(renewals)
})

router.post('/renewals', async (req, res) => {
  const renewal = await prisma.gpsRenewal.create({
    data: req.body,
    include: { customer: true, device: true },
  })
  res.status(201).json(renewal)
})

router.put('/renewals/:id', async (req, res) => {
  const renewal = await prisma.gpsRenewal.update({
    where: { id: req.params.id },
    data: req.body,
    include: { customer: true, device: true },
  })
  res.json(renewal)
})

// ── Maintenance ─────────────────────────────────────────────────────────────

router.get('/maintenance', async (_req, res) => {
  const records = await prisma.gpsMaintenance.findMany({
    include: { customer: true, device: true, employee: true },
    orderBy: { createdAt: 'desc' },
  })
  res.json(records)
})

router.post('/maintenance', async (req, res) => {
  const record = await prisma.gpsMaintenance.create({
    data: req.body,
    include: { customer: true, device: true, employee: true },
  })
  res.status(201).json(record)
})

router.put('/maintenance/:id', async (req, res) => {
  const record = await prisma.gpsMaintenance.update({
    where: { id: req.params.id },
    data: req.body,
    include: { customer: true, device: true, employee: true },
  })
  res.json(record)
})

// ── Settings (prices) ───────────────────────────────────────────────────────

router.get('/settings', async (_req, res) => {
  const settings = await prisma.gpsSetting.findMany()
  res.json(settings)
})

router.put('/settings', async (req, res) => {
  const { id, ...data } = req.body
  const settings = await prisma.gpsSetting.upsert({
    where: { id: id || 'default' },
    update: data,
    create: { id: id || 'default', ...data },
  })
  res.json(settings)
})

// ── Stats ───────────────────────────────────────────────────────────────────

router.get('/stats', async (_req, res) => {
  const [devicesByStatus, totalDevices, activeSubscriptions] = await Promise.all([
    prisma.gpsDevice.groupBy({ by: ['status'], _count: { id: true } }),
    prisma.gpsDevice.count(),
    prisma.gpsRenewal.count({ where: { status: 'ACTIVE' } }),
  ])

  res.json({
    devicesByStatus: devicesByStatus.map(d => ({ status: d.status, count: d._count.id })),
    totalDevices,
    activeSubscriptions,
  })
})

export default router
