import { Router } from 'express'
import { prisma } from '../prisma'
import { requireRole } from '../middleware/requireAuth'

const router = Router()

// GET / - list all products
router.get('/', async (_req, res) => {
  const products = await prisma.product.findMany({ orderBy: { name: 'asc' } })
  res.json(products)
})

// POST / - create product
router.post('/', async (req, res) => {
  const product = await prisma.product.create({ data: req.body })
  res.status(201).json(product)
})

// PUT /:id - update product (يعدل السعر الافتراضي لكل من يستخدم هذا المنتج بالكتالوگ، لذا محصور بالمدير)
router.put('/:id', requireRole('ADMIN'), async (req, res) => {
  const product = await prisma.product.update({ where: { id: req.params.id }, data: req.body })
  res.json(product)
})

// DELETE /:id - delete product
router.delete('/:id', requireRole('ADMIN'), async (req, res) => {
  await prisma.product.delete({ where: { id: req.params.id } })
  res.json({ success: true })
})

export default router
