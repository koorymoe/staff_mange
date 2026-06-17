import { Router } from 'express'
import { prisma } from '../prisma'

const router = Router()

// ── Personal Tools ──────────────────────────────────────────────────────────

router.get('/personal', async (req, res) => {
  const where = req.query.employeeId ? { employeeId: req.query.employeeId as string } : {}
  const tools = await prisma.personalTool.findMany({ where, include: { employee: true }, orderBy: { createdAt: 'desc' } })
  res.json(tools)
})

router.post('/personal', async (req, res) => {
  const { employeeId, name, barcode } = req.body
  const tool = await prisma.personalTool.create({ data: { employeeId, name, barcode } })
  res.status(201).json(tool)
})

router.put('/personal/:id', async (req, res) => {
  const { status, checkedOut } = req.body
  const tool = await prisma.personalTool.update({ where: { id: req.params.id }, data: { status, checkedOut } })
  res.json(tool)
})

router.delete('/personal/:id', async (req, res) => {
  await prisma.personalTool.delete({ where: { id: req.params.id } })
  res.json({ success: true })
})

// ── Vehicle Tools ───────────────────────────────────────────────────────────

router.get('/vehicle', async (req, res) => {
  const where = req.query.vehicleId ? { vehicleId: req.query.vehicleId as string } : {}
  const tools = await prisma.vehicleTool.findMany({ where, orderBy: { createdAt: 'desc' } })
  res.json(tools)
})

router.post('/vehicle', async (req, res) => {
  const tool = await prisma.vehicleTool.create({ data: req.body })
  res.status(201).json(tool)
})

router.put('/vehicle/:id', async (req, res) => {
  const tool = await prisma.vehicleTool.update({ where: { id: req.params.id }, data: req.body })
  res.json(tool)
})

router.delete('/vehicle/:id', async (req, res) => {
  await prisma.vehicleTool.delete({ where: { id: req.params.id } })
  res.json({ success: true })
})

// ── On-demand Tools ─────────────────────────────────────────────────────────

router.get('/ondemand', async (_req, res) => {
  const tools = await prisma.onDemandTool.findMany({ orderBy: { name: 'asc' } })
  res.json(tools)
})

router.post('/ondemand', async (req, res) => {
  const tool = await prisma.onDemandTool.create({ data: req.body })
  res.status(201).json(tool)
})

router.put('/ondemand/:id', async (req, res) => {
  const tool = await prisma.onDemandTool.update({ where: { id: req.params.id }, data: req.body })
  res.json(tool)
})

// ── Tool Requests ───────────────────────────────────────────────────────────

router.get('/requests', async (req, res) => {
  const where = req.query.employeeId ? { employeeId: req.query.employeeId as string } : {}
  const requests = await prisma.toolRequest.findMany({
    where,
    include: { employee: true, tool: true, approvedBy: true },
    orderBy: { createdAt: 'desc' },
  })
  res.json(requests)
})

router.post('/requests', async (req, res) => {
  const { employeeId, toolId } = req.body
  const request = await prisma.toolRequest.create({
    data: { employeeId, toolId },
    include: { employee: true, tool: true },
  })
  res.status(201).json(request)
})

router.put('/requests/:id/approve', async (req, res) => {
  const { approvedById } = req.body
  const request = await prisma.toolRequest.update({
    where: { id: req.params.id },
    data: { status: 'APPROVED', approvedById },
    include: { employee: true, tool: true, approvedBy: true },
  })
  res.json(request)
})

router.put('/requests/:id/reject', async (req, res) => {
  const request = await prisma.toolRequest.update({
    where: { id: req.params.id },
    data: { status: 'REJECTED' },
    include: { employee: true, tool: true },
  })
  res.json(request)
})

router.put('/requests/:id/return', async (req, res) => {
  const request = await prisma.toolRequest.update({
    where: { id: req.params.id },
    data: { status: 'RETURNED', returnedAt: new Date() },
    include: { tool: true },
  })

  // Increment available quantity on the tool
  await prisma.onDemandTool.update({
    where: { id: request.toolId },
    data: { availableQuantity: { increment: 1 } },
  })

  res.json(request)
})

export default router
