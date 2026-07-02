import { Router } from 'express'
import bcrypt from 'bcryptjs'
import { prisma } from '../prisma'
import { signToken, verifyToken } from '../lib/jwt'

const router = Router()

// POST /api/auth/login - body: { username, password }
router.post('/login', async (req, res) => {
  const { username, password } = req.body as { username?: string; password?: string }
  if (!username || !password) {
    return res.status(400).json({ error: 'اسم المستخدم وكلمة المرور مطلوبان' })
  }

  const employee = await prisma.employee.findUnique({
    where: { username },
    include: { skills: { include: { skill: { include: { service: true } } } } },
  })

  if (!employee || !employee.password || !bcrypt.compareSync(password, employee.password)) {
    return res.status(401).json({ error: 'اسم المستخدم أو كلمة المرور غير صحيحة' })
  }

  const token = signToken({ employeeId: employee.id, role: employee.role })
  const { password: _password, ...rest } = employee
  res.json({ ...rest, token })
})

// GET /api/auth/me - يعيد بيانات الموظف الحالي بناءً على التوكن (تستخدمها الواجهة عند فتح الصفحة)
router.get('/me', async (req, res) => {
  const header = req.header('authorization')
  if (!header?.startsWith('Bearer ')) return res.status(401).json({ error: 'غير مسجل الدخول' })

  try {
    const payload = verifyToken(header.slice(7))
    const employee = await prisma.employee.findUnique({
      where: { id: payload.employeeId },
      include: { skills: { include: { skill: { include: { service: true } } } } },
    })
    if (!employee) return res.status(401).json({ error: 'الحساب غير موجود' })
    const { password: _password, ...rest } = employee
    res.json(rest)
  } catch {
    res.status(401).json({ error: 'جلسة الدخول منتهية' })
  }
})

export default router
