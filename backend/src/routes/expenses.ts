import { Router } from 'express'
import { prisma } from '../prisma'
import { requireRole } from '../middleware/requireAuth'

const router = Router()

const expenseInclude = {
  employee: { select: { id: true, name: true, position: true } },
} as const

// GET /api/expenses?employeeId=xxx - list expenses, optionally filtered by employee
router.get('/', async (req, res) => {
  const { employeeId } = req.query
  const expenses = await prisma.expense.findMany({
    where: employeeId ? { employeeId: employeeId as string } : undefined,
    include: expenseInclude,
    orderBy: { createdAt: 'desc' },
  })
  res.json(expenses)
})

// POST /api/expenses - employee submits an expense to be reviewed by finance
router.post('/', async (req, res) => {
  const { employeeId, amount, description } = req.body
  if (!employeeId || amount == null) {
    return res.status(400).json({ error: 'employeeId and amount are required' })
  }
  const expense = await prisma.expense.create({
    data: { employeeId, amount, description: description || null },
    include: expenseInclude,
  })
  res.status(201).json(expense)
})

// PUT /api/expenses/:id/status - finance approves or rejects the expense
router.put('/:id/status', requireRole('ADMIN', 'FINANCE'), async (req, res) => {
  const { status } = req.body
  if (!['APPROVED', 'REJECTED'].includes(status)) {
    return res.status(400).json({ error: 'status must be APPROVED or REJECTED' })
  }
  const expense = await prisma.expense.update({
    where: { id: req.params.id },
    data: { status },
    include: expenseInclude,
  })
  res.json(expense)
})

export default router
