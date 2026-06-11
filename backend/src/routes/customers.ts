import { Router } from 'express'
import { prisma } from '../prisma'

const router = Router()

const formatCode = (customerCode: number) => `CUST-${String(customerCode).padStart(5, '0')}`

// GET /api/customers - list customers
router.get('/', async (_req, res) => {
  const customers = await prisma.customer.findMany({ orderBy: { customerCode: 'asc' } })
  res.json(customers.map((c) => ({ ...c, code: formatCode(c.customerCode) })))
})

// GET /api/customers/lookup?phone=xxx - find existing customer by phone
router.get('/lookup', async (req, res) => {
  const { phone } = req.query
  if (typeof phone !== 'string') {
    return res.status(400).json({ error: 'phone is required' })
  }

  const customer = await prisma.customer.findUnique({ where: { phone } })
  if (!customer) return res.status(404).json({ error: 'Customer not found' })
  res.json({ ...customer, code: formatCode(customer.customerCode) })
})

// POST /api/customers - find existing customer by phone, or create a new one
router.post('/', async (req, res) => {
  const { name, phone, location } = req.body
  if (!name || !phone) {
    return res.status(400).json({ error: 'name and phone are required' })
  }

  const existing = await prisma.customer.findUnique({ where: { phone } })
  if (existing) {
    return res.json({ ...existing, code: formatCode(existing.customerCode), existed: true })
  }

  const customer = await prisma.customer.create({ data: { name, phone, location } })
  res.status(201).json({ ...customer, code: formatCode(customer.customerCode), existed: false })
})

export default router
