import type { Request, Response, NextFunction } from 'express'
import { verifyToken } from '../lib/jwt'

export interface AuthedRequest extends Request<any> {
  auth?: { employeeId: string; role: string }
}

// يتحقق من صحة توكن JWT بكل طلب. يرفض أي طلب بدون توكن صالح.
export function requireAuth(req: AuthedRequest, res: Response, next: NextFunction) {
  const header = req.header('authorization')
  if (!header?.startsWith('Bearer ')) {
    return res.status(401).json({ error: 'يجب تسجيل الدخول' })
  }

  try {
    const payload = verifyToken(header.slice(7))
    req.auth = { employeeId: payload.employeeId, role: payload.role }
    next()
  } catch {
    return res.status(401).json({ error: 'جلسة الدخول منتهية، الرجاء تسجيل الدخول مجدداً' })
  }
}

// يُستخدم بعد requireAuth — يمنع الوصول إلا لأصحاب الأدوار المذكورة
export function requireRole(...roles: string[]) {
  return (req: AuthedRequest, res: Response, next: NextFunction) => {
    if (!req.auth || !roles.includes(req.auth.role)) {
      return res.status(403).json({ error: 'لا تملك صلاحية الوصول لهذه العملية' })
    }
    next()
  }
}
