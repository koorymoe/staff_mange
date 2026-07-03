import { Router } from 'express'
import { prisma } from '../prisma'
import { requireRole } from '../middleware/requireAuth'

const router = Router()

// GET / - list all quotations with items and creator
router.get('/', async (_req, res) => {
  const quotations = await prisma.quotation.findMany({
    include: { items: true, createdByEmployee: { select: { id: true, name: true } } },
    orderBy: { createdAt: 'desc' },
  })
  res.json(quotations)
})

// GET /:id - single quotation with items
router.get('/:id', async (req, res) => {
  const quotation = await prisma.quotation.findUnique({
    where: { id: req.params.id },
    include: { items: true, createdByEmployee: { select: { id: true, name: true } } },
  })
  if (!quotation) return res.status(404).json({ error: 'Quotation not found' })
  res.json(quotation)
})

// POST / - create quotation with items
router.post('/', async (req, res) => {
  const {
    customerName,
    customerPhone,
    customerAddress,
    projectName,
    items,
    discountPercent,
    notes,
    duration,
    createdByEmployeeId,
  } = req.body

  // Auto-generate quotation number AM-YYYY-XXXX
  const year = new Date().getFullYear()
  const lastQuotation = await prisma.quotation.findFirst({
    where: { quotationNumber: { startsWith: `AM-${year}-` } },
    orderBy: { quotationNumber: 'desc' },
  })

  let seq = 1
  if (lastQuotation) {
    const lastSeq = parseInt(lastQuotation.quotationNumber.split('-')[2], 10)
    seq = lastSeq + 1
  }
  const quotationNumber = `AM-${year}-${String(seq).padStart(4, '0')}`

  // Calculate totals
  const grandTotal = (items as { quantity: number; unitPrice: number }[]).reduce(
    (sum, item) => sum + item.quantity * item.unitPrice,
    0
  )
  const discountValue = grandTotal * ((discountPercent || 0) / 100)
  const netTotal = grandTotal - discountValue

  const quotation = await prisma.quotation.create({
    data: {
      quotationNumber,
      customerName,
      customerPhone,
      customerAddress,
      projectName,
      discountPercent: discountPercent || 0,
      discountValue,
      grandTotal,
      netTotal,
      notes,
      duration,
      createdByEmployeeId,
      items: {
        create: (items as { description: string; unit: string; quantity: number; unitPrice: number }[]).map(item => ({
          productName: item.description,
          unit: item.unit,
          quantity: item.quantity,
          unitPrice: item.unitPrice,
          totalPrice: item.quantity * item.unitPrice,
        })),
      },
    },
    include: { items: true, createdByEmployee: { select: { id: true, name: true } } },
  })
  res.status(201).json(quotation)
})

// PUT /:id - update quotation
router.put('/:id', async (req, res) => {
  const { items, ...data } = req.body

  if (items) {
    // Recalculate totals
    const grandTotal = (items as { quantity: number; unitPrice: number }[]).reduce(
      (sum, item) => sum + item.quantity * item.unitPrice,
      0
    )
    const discountValue = grandTotal * ((data.discountPercent || 0) / 100)
    const netTotal = grandTotal - discountValue

    const quotation = await prisma.$transaction(async tx => {
      await tx.quotationItem.deleteMany({ where: { quotationId: req.params.id } })
      return tx.quotation.update({
        where: { id: req.params.id },
        data: {
          ...data,
          grandTotal,
          discountValue,
          netTotal,
          items: {
            create: (items as { description: string; unit: string; quantity: number; unitPrice: number }[]).map(item => ({
              productName: item.description,
              unit: item.unit,
              quantity: item.quantity,
              unitPrice: item.unitPrice,
              totalPrice: item.quantity * item.unitPrice,
            })),
          },
        },
        include: { items: true, createdByEmployee: { select: { id: true, name: true } } },
      })
    })
    return res.json(quotation)
  }

  const quotation = await prisma.quotation.update({
    where: { id: req.params.id },
    data,
    include: { items: true, createdByEmployee: { select: { id: true, name: true } } },
  })
  res.json(quotation)
})

// DELETE /:id - delete quotation
router.delete('/:id', requireRole('ADMIN'), async (req, res) => {
  await prisma.quotation.delete({ where: { id: req.params.id } })
  res.json({ success: true })
})

export default router
