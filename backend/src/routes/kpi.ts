import { Router } from 'express'
import { prisma } from '../prisma'

const router = Router()

// GET / - list all evaluations with employee and evaluator names
router.get('/', async (_req, res) => {
  const evaluations = await prisma.kpiEvaluation.findMany({
    include: { employee: true, evaluator: true },
    orderBy: { createdAt: 'desc' },
  })
  res.json(evaluations)
})

// GET /employee/:employeeId - get evaluations for specific employee
router.get('/employee/:employeeId', async (req, res) => {
  const evaluations = await prisma.kpiEvaluation.findMany({
    where: { employeeId: req.params.employeeId },
    include: { employee: true, evaluator: true },
    orderBy: { createdAt: 'desc' },
  })
  res.json(evaluations)
})

// POST / - create evaluation
router.post('/', async (req, res) => {
  const { employeeId, evaluatorId, points, reason } = req.body
  if (!employeeId || !evaluatorId || points === undefined) {
    return res.status(400).json({ error: 'employeeId, evaluatorId, and points are required' })
  }

  const deductionAmount = points * 10000

  const evaluation = await prisma.kpiEvaluation.create({
    data: { employeeId, evaluatorId, points, reason, deductionAmount },
    include: { employee: true, evaluator: true },
  })
  res.status(201).json(evaluation)
})

// DELETE /:id - delete evaluation
router.delete('/:id', async (req, res) => {
  await prisma.kpiEvaluation.delete({ where: { id: req.params.id } })
  res.json({ success: true })
})

export default router
