import { Router } from 'express'
import { prisma } from '../prisma'
import { requireRole } from '../middleware/requireAuth'

const router = Router()

// GET / - list all procurement requests
router.get('/', async (_req, res) => {
  const requests = await prisma.procurementRequest.findMany({
    include: {
      items: true,
      requestedBy: { select: { id: true, name: true, role: true } },
      fulfilledBy: { select: { id: true, name: true } },
      booking: { include: { customer: { select: { id: true, name: true, phone: true } } } },
    },
    orderBy: { createdAt: 'desc' },
  })
  res.json(requests)
})

// GET /stats - procurement statistics
router.get('/stats', async (_req, res) => {
  const now = new Date()
  const startOfMonth = new Date(now.getFullYear(), now.getMonth(), 1)

  const [totalSpent, totalItems, pendingCount, monthlySpent, fulfilled] = await Promise.all([
    prisma.procurementRequest.aggregate({
      _sum: { totalCost: true },
      where: { status: 'FULFILLED' },
    }),
    prisma.procurementItem.count(),
    prisma.procurementRequest.count({ where: { status: 'PENDING' } }),
    prisma.procurementRequest.aggregate({
      _sum: { totalCost: true },
      where: { status: 'FULFILLED', fulfilledAt: { gte: startOfMonth } },
    }),
    prisma.procurementRequest.count({ where: { status: 'FULFILLED' } }),
  ])

  // Monthly breakdown (last 6 months)
  const sixMonthsAgo = new Date(now.getFullYear(), now.getMonth() - 5, 1)
  const monthlyData = await prisma.procurementRequest.findMany({
    where: { status: 'FULFILLED', fulfilledAt: { gte: sixMonthsAgo } },
    select: { totalCost: true, fulfilledAt: true },
  })

  const byMonth: Record<string, number> = {}
  for (const r of monthlyData) {
    if (!r.fulfilledAt) continue
    const key = `${r.fulfilledAt.getFullYear()}-${String(r.fulfilledAt.getMonth() + 1).padStart(2, '0')}`
    byMonth[key] = (byMonth[key] || 0) + (r.totalCost || 0)
  }

  res.json({
    totalSpent: totalSpent._sum.totalCost || 0,
    totalItems,
    pendingCount,
    monthlySpent: monthlySpent._sum.totalCost || 0,
    fulfilledCount: fulfilled,
    byMonth,
  })
})

// POST / - create a new procurement request
router.post('/', async (req, res) => {
  const { requestedById, bookingId, notes, items } = req.body

  // Auto-generate code PR-XXXX
  const last = await prisma.procurementRequest.findFirst({
    orderBy: { code: 'desc' },
    select: { code: true },
  })
  let nextNum = 1
  if (last?.code) {
    const match = last.code.match(/PR-(\d+)/)
    if (match) nextNum = parseInt(match[1], 10) + 1
  }
  const code = `PR-${String(nextNum).padStart(4, '0')}`

  const request = await prisma.procurementRequest.create({
    data: {
      code,
      requestedById,
      bookingId: bookingId || null,
      notes: notes || null,
      items: {
        create: (items || []).map((item: any) => ({
          productName: item.productName,
          quantity: item.quantity,
          unitPrice: item.unitPrice || null,
          totalPrice: item.totalPrice || null,
        })),
      },
    },
    include: {
      items: true,
      requestedBy: { select: { id: true, name: true, role: true } },
      booking: { include: { customer: { select: { id: true, name: true, phone: true } } } },
    },
  })
  res.status(201).json(request)
})

// PUT /:id/status - update status
router.put('/:id/status', requireRole('ADMIN', 'FINANCE'), async (req, res) => {
  const { id } = req.params
  const { status } = req.body

  const request = await prisma.procurementRequest.update({
    where: { id },
    data: { status },
    include: {
      items: true,
      requestedBy: { select: { id: true, name: true, role: true } },
      fulfilledBy: { select: { id: true, name: true } },
      booking: { include: { customer: { select: { id: true, name: true, phone: true } } } },
    },
  })
  res.json(request)
})

// PUT /:id/fulfill - mark as fulfilled with costs
router.put('/:id/fulfill', requireRole('ADMIN', 'FINANCE'), async (req, res) => {
  const { id } = req.params
  const { fulfilledById, totalCost, fulfillmentNotes, items } = req.body

  // Update individual item costs if provided
  if (items && Array.isArray(items)) {
    for (const item of items) {
      await prisma.procurementItem.update({
        where: { id: item.id },
        data: {
          unitPrice: item.unitPrice ?? undefined,
          totalPrice: item.totalPrice ?? undefined,
          fulfilled: item.fulfilled ?? true,
        },
      })
    }
  }

  const request = await prisma.procurementRequest.update({
    where: { id },
    data: {
      status: 'FULFILLED',
      fulfilledById,
      totalCost: totalCost || null,
      fulfillmentNotes: fulfillmentNotes || null,
      fulfilledAt: new Date(),
    },
    include: {
      items: true,
      requestedBy: { select: { id: true, name: true, role: true } },
      fulfilledBy: { select: { id: true, name: true } },
      booking: { include: { customer: { select: { id: true, name: true, phone: true } } } },
    },
  })
  res.json(request)
})

export default router
