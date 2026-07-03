import { Router } from 'express'
import { prisma } from '../prisma'
import { requireRole } from '../middleware/requireAuth'

const router = Router()

// GET / - list all evaluations with employee and evaluator names
router.get('/', async (_req, res) => {
  const evaluations = await prisma.kpiEvaluation.findMany({
    include: { employee: { select: { id: true, name: true } }, evaluator: { select: { id: true, name: true } } },
    orderBy: { createdAt: 'desc' },
  })
  res.json(evaluations)
})

// GET /employee/:employeeId - get evaluations for specific employee
router.get('/employee/:employeeId', async (req, res) => {
  const evaluations = await prisma.kpiEvaluation.findMany({
    where: { employeeId: req.params.employeeId },
    include: { employee: { select: { id: true, name: true } }, evaluator: { select: { id: true, name: true } } },
    orderBy: { createdAt: 'desc' },
  })
  res.json(evaluations)
})

// POST / - create evaluation
router.post('/', requireRole('ADMIN', 'MONITOR'), async (req, res) => {
  const { employeeId, evaluatorId, points, reason } = req.body
  if (!employeeId || !evaluatorId || points === undefined) {
    return res.status(400).json({ error: 'employeeId, evaluatorId, and points are required' })
  }

  const deductionAmount = points * 10000

  const evaluation = await prisma.kpiEvaluation.create({
    data: { employeeId, evaluatorId, points, reason, deductionAmount },
    include: { employee: { select: { id: true, name: true } }, evaluator: { select: { id: true, name: true } } },
  })
  res.status(201).json(evaluation)
})

// DELETE /:id - delete evaluation
router.delete('/:id', requireRole('ADMIN'), async (req, res) => {
  await prisma.kpiEvaluation.delete({ where: { id: req.params.id } })
  res.json({ success: true })
})

export default router
