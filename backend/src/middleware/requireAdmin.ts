import type { Request, Response, NextFunction } from 'express'
import { prisma } from '../prisma'

// تحقق مبدئي: يمنع تنفيذ العملية إلا إذا كان صاحب الطلب موظف بدور ADMIN
// ملاحظة: هذا تحقق على مستوى التطبيق وليس بديلاً عن نظام توثيق كامل (JWT/جلسات)
// وهذا ضمن نطاق تحسينات الأمان (Cyber Security) اللاحقة المتفق عليها
export async function requireAdmin(req: Request<any>, res: Response, next: NextFunction) {
  const employeeId = req.header('x-employee-id')
  if (!employeeId) return res.status(401).json({ error: 'يجب تسجيل الدخول' })

  const employee = await prisma.employee.findUnique({ where: { id: employeeId } })
  if (!employee || employee.role !== 'ADMIN') {
    return res.status(403).json({ error: 'هذه العملية مخصصة لمدير النظام فقط' })
  }

  next()
}
