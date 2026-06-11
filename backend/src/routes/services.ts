import { Router } from 'express'
import { prisma } from '../prisma'

const router = Router()

// GET /api/services - list all services
router.get('/', async (_req, res) => {
  const services = await prisma.service.findMany({ orderBy: { name: 'asc' } })
  res.json(services)
})

export default router
