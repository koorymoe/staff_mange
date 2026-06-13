import { Router } from 'express'
import { prisma } from '../prisma'

const router = Router()

// GET /api/services - list all services with their skills
router.get('/', async (_req, res) => {
  const services = await prisma.service.findMany({
    include: { skills: { orderBy: { name: 'asc' } } },
    orderBy: { name: 'asc' },
  })
  res.json(services)
})

// POST /api/services - add a new service
router.post('/', async (req, res) => {
  const { name, category } = req.body
  if (!name) return res.status(400).json({ error: 'name is required' })

  const service = await prisma.service.create({
    data: { name, category: category || null },
  })
  res.status(201).json(service)
})

export default router
