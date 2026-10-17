import type { ReactNode } from 'react'
import { useSession } from '../session'

// يمنع أي موظف غير ADMIN من الوصول للصفحة حتى لو كتب رابطها مباشرة بالمتصفح
export default function RequireAdmin({ children }: { children: ReactNode }) {
  const { employee } = useSession()

  if (employee?.role !== 'ADMIN') {
    return (
      <div className="flex flex-col items-center justify-center gap-3 rounded-2xl border border-red-200 bg-red-50 p-10 text-center">
        <span className="text-4xl">🚫</span>
        <p className="text-lg font-bold text-red-700">غير مصرح لك بالوصول لهذه الصفحة</p>
        <p className="text-sm text-red-500">هذه الصفحة مخصصة لمدير النظام فقط.</p>
      </div>
    )
  }

  return <>{children}</>
}
