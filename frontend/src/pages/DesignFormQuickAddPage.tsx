import { useEffect, useState } from 'react'
import { Navigate, Link } from 'react-router-dom'
import { api, type DesignForm } from '../api'

const PRIMARY = '#47528f'
// ⚠️ نسخة **النص** تنقلب بالوضع الليلي، والأصل يبقى للأسطح:
// نفس اللون يخدم عنواناً غامقاً على أبيض، ورأس جدول كحلي عليه نص أبيض.
// قلب الاثنين سوا يكسر واحداً منهما — نفس فخّ --color-white.
const PRIMARY_TEXT = 'var(--design-ink)'

// اختصار "إضافة سؤال": يودّي مباشرة لأسئلة الفورمة الوحيدة الموجودة، أو يخلّي
// المستخدم يختار من عدة فورمات إذا اكو أكثر من وحدة، أو يوجّهه ينشئ أول
// فورمة إذا ما عنده وحدة بعد.
export default function DesignFormQuickAddPage() {
  const [forms, setForms] = useState<DesignForm[] | null>(null)

  useEffect(() => { api.getDesignForms().then(setForms) }, [])

  if (forms === null) return <p className="mt-6 text-slate-400">جاري التحميل...</p>

  if (forms.length === 0) {
    return (
      <div dir="rtl" className="rounded-xl border border-white bg-white p-8 text-center shadow-sm">
        <p className="text-slate-500">ما عندك أي فورمة بعد — سوّي فورمة جديدة الأول من صفحة "فورمة التصميم" وبعدها تكدر تضيفلها الأسئلة.</p>
        <Link to="/design-forms" className="mt-4 inline-block rounded-lg px-5 py-2 text-sm font-bold text-white" style={{ background: PRIMARY }}>
          فورمة التصميم
        </Link>
      </div>
    )
  }

  if (forms.length === 1) {
    return <Navigate to={`/design-forms/${forms[0].id}`} replace />
  }

  return (
    <div dir="rtl">
      <h2 className="text-2xl font-bold" style={{ color: PRIMARY_TEXT }}>لأي فورمة تريد تضيف السؤال؟</h2>
      <div className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
        {forms.map((f) => (
          <Link key={f.id} to={`/design-forms/${f.id}`} className="rounded-xl border border-white bg-white p-4 font-bold shadow-sm hover:shadow-md" style={{ color: PRIMARY_TEXT }}>
            {f.name}
          </Link>
        ))}
      </div>
    </div>
  )
}
