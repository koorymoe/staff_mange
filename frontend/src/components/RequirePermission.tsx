import type { ReactNode } from 'react'
import { useSession } from '../session'

// يمنع الوصول إلا لـ ADMIN أو موظف عنده الصلاحية المذكورة، حتى لو كتب رابط الصفحة مباشرة
export default function RequirePermission({ permission, children }: { permission: string; children: ReactNode }) {
  const { employee, permissions } = useSession()

  const allowed = employee?.role === 'ADMIN' || permissions.includes(permission)

  if (!allowed) {
    return (
      <div className="flex flex-col items-center justify-center gap-3 rounded-2xl border border-red-200 bg-red-50 p-10 text-center">
        <span className="text-4xl">🚫</span>
        <p className="text-lg font-bold text-red-700">غير مصرح لك بالوصول لهذه الصفحة</p>
        <p className="text-sm text-red-500">تحتاج صلاحية خاصة للوصول لهذه الصفحة، تواصل مع مدير النظام.</p>
      </div>
    )
  }

  return <>{children}</>
}
