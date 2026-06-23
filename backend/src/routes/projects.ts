import { Router } from 'express'
import { prisma } from '../prisma'

const router = Router()

// المراحل بالترتيب
const STAGES = [
  '1. اتصال بالزبون',
  '2. مرحلة الكشف',
  '3. عرض السعر',
  '4. البدء بالتنفيذ',
  '✅ مكتمل',
  '❌ مرفوض',
]

function computeStats(projects: { stage: string }[]) {
  const stats = { اتصال: 0, كشف: 0, سعر: 0, تنفيذ: 0, مكتمل: 0, مرفوض: 0 }
  for (const p of projects) {
    const s = p.stage || ''
    if (s.includes('اتصال')) stats['اتصال']++
    else if (s.includes('كشف')) stats['كشف']++
    else if (s.includes('سعر')) stats['سعر']++
    else if (s.includes('تنفيذ')) stats['تنفيذ']++
    else if (s.includes('مكتمل')) stats['مكتمل']++
    else if (s.includes('مرفوض')) stats['مرفوض']++
  }
  return stats
}

// GET / - list all projects with stage stats
router.get('/', async (_req, res) => {
  const projects = await prisma.project.findMany({ orderBy: { createdAt: 'desc' } })
  res.json({ projects, stats: computeStats(projects), stages: STAGES })
})

// POST / - create a new project (always starts at first stage)
router.post('/', async (req, res) => {
  const { name, rep, phone, location, workType, refPerson, priority, deliveryDate } = req.body
  if (!name) return res.status(400).json({ error: 'اسم المؤسسة مطلوب' })
  const count = await prisma.project.count()
  const code = 'PRJ-' + String(count + 1).padStart(4, '0')
  const project = await prisma.project.create({
    data: {
      code,
      name,
      rep: rep || null,
      phone: phone || null,
      location: location || null,
      workType: workType || null,
      refPerson: refPerson || null,
      priority: priority || 'عادي',
      deliveryDate: deliveryDate || null,
      stage: STAGES[0],
    },
  })
  res.json(project)
})

// PUT /:id - update / move a project (stage transitions, survey, rejection...)
router.put('/:id', async (req, res) => {
  const { name, rep, phone, location, workType, refPerson, stage, price, staff, time, task, priority, deliveryDate, survey, sentToGroup } = req.body
  const data: Record<string, unknown> = {}
  if (name !== undefined) data.name = name
  if (rep !== undefined) data.rep = rep
  if (phone !== undefined) data.phone = phone
  if (location !== undefined) data.location = location
  if (workType !== undefined) data.workType = workType
  if (refPerson !== undefined) data.refPerson = refPerson
  if (stage !== undefined) data.stage = stage
  if (price !== undefined) data.price = price
  if (staff !== undefined) data.staff = staff
  if (time !== undefined) data.time = time
  if (task !== undefined) data.task = task
  if (priority !== undefined) data.priority = priority
  if (deliveryDate !== undefined) data.deliveryDate = deliveryDate
  if (survey !== undefined) data.survey = survey
  if (sentToGroup !== undefined) data.sentToGroup = sentToGroup
  const project = await prisma.project.update({ where: { id: req.params.id }, data })
  res.json(project)
})

// DELETE /:id
router.delete('/:id', async (req, res) => {
  await prisma.project.delete({ where: { id: req.params.id } })
  res.json({ ok: true })
})

export default router
