import type { ReactNode } from 'react'
import { useSession } from '../session'

// يمنع الوصول إلا لـ ADMIN أو موظف عنده الصلاحية المذكورة، حتى لو كتب رابط الصفحة مباشرة
//
// ⚠️ `anyOf` — بدائل مقبولة غير الصلاحية الأساسية. الحاجة إلها: صلاحية
// **الوحدة** الي «تفتح الوحدة كاملة بكل صفحاتها» تفتح البند بالقائمة
// الجانبية، وبعدين الصفحة ترفض لأنها تطلب صلاحية الصفحة وحدها — فالموظف
// يشوف البند ويضغطه ويلگه «غير مصرح لك». البند الي ينضغط وينرفض أسوأ
// من بند ما موجود.
export default function RequirePermission(
  { permission, anyOf, children }: { permission: string; anyOf?: string[]; children: ReactNode },
) {
  const { employee, permissions } = useSession()

  const allowed = employee?.role === 'ADMIN'
    || permissions.includes(permission)
    || (anyOf ?? []).some((p) => permissions.includes(p))

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
