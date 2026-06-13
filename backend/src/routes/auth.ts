import { Router } from 'express'
import bcrypt from 'bcryptjs'
import { prisma } from '../prisma'

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

  const { password: _password, ...rest } = employee
  res.json(rest)
})

export default router
