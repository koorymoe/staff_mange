import type { ReactNode } from 'react'
import { useSession } from '../session'

// ═══ بوابة مختبر المحاكاة ═══
//
// شرط صريح من صاحب النظام: «هذا أريده يظهر فقط عند المالك، حتى مدير
// النظام ما أريده يظهر عنده إلى أن يكتمل بصورة كاملة».
//
// ⚠️ الواجهة تطبّع دور المالك إلى 'ADMIN' (شوف Layout.tsx) والدور
// الحقيقي يبقى بـactualRole. لهذا الفحص لازم يكون actualRole — لو
// كتبته role === 'OWNER' ما يشوفها المالك نفسه، ولو كتبته
// role === 'ADMIN' يشوفها كل مدير.
//
// ⚠️ وهاي **مو** الحماية الحقيقية: المسارات بالسيرفر محمية بـRequireOwner
// وترجّع 404 لأي حساب ثاني. هاي بس حتى ما تطلع شاشة مكسورة لمن يفتح
// الرابط مباشرة.
//
// ⚠️ ويوم ينفتح المختبر للموظفين، **هذا الملف وحده** الي يتغيّر.
export default function SimGate({ children }: { children: ReactNode }) {
  const { employee } = useSession()
  if (employee?.actualRole !== 'OWNER') {
    return (
      <div className="rounded-2xl bg-white p-8 text-center shadow">
        <p className="text-lg font-bold text-slate-700">الصفحة غير موجودة</p>
      </div>
    )
  }
  return <>{children}</>
}
