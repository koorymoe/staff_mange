import { Router } from 'express'
import { prisma } from '../prisma'

const router = Router()

// GET /booking/:bookingId - get cart items for a booking
router.get('/booking/:bookingId', async (req, res) => {
  const items = await prisma.cartItem.findMany({
    where: { bookingId: req.params.bookingId },
    orderBy: { createdAt: 'asc' },
  })
  res.json(items)
})

// POST /booking/:bookingId - add item
router.post('/booking/:bookingId', async (req, res) => {
  const { productName, quantity, unitPrice, notes } = req.body
  if (!productName || quantity === undefined || unitPrice === undefined) {
    return res.status(400).json({ error: 'productName, quantity, and unitPrice are required' })
  }

  const totalPrice = quantity * unitPrice

  const item = await prisma.cartItem.create({
    data: {
      bookingId: req.params.bookingId,
      productName,
      quantity,
      unitPrice,
      totalPrice,
      notes,
    },
  })
  res.status(201).json(item)
})

// PUT /:id - update item
router.put('/:id', async (req, res) => {
  const { productName, quantity, unitPrice, notes } = req.body
  const data: Record<string, unknown> = { productName, quantity, unitPrice, notes }

  if (quantity !== undefined && unitPrice !== undefined) {
    data.totalPrice = quantity * unitPrice
  }

  const item = await prisma.cartItem.update({
    where: { id: req.params.id },
    data,
  })
  res.json(item)
})

// DELETE /:id - delete item
router.delete('/:id', async (req, res) => {
  await prisma.cartItem.delete({ where: { id: req.params.id } })
  res.json({ success: true })
})

export default router
