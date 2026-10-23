import { useState, type ReactNode } from 'react'
import PageHeader from './PageHeader'

// ═══ مدخل موحّد بتبويبات ═══
//
// «ما أريدهن يكونن هيج بالقائمة الجانبية، أريدهن بواجهة وحدة».
//
// شاشات نفس الشغلة كانت بنوداً منفصلة بالقائمة، فالموظف يفتح بنداً
// ويشتغل عليه وينسى الثاني. هذا المكوّن يلمّهن بمدخل واحد.
//
// ⚠️ **ما ينسخ منطق ولا شاشة**: كل تبويب يعرض شاشته الأصلية كما هي
// بـ`embedded`. نسختان لنفس الشغلة تفترقن أول تعديل.
//
// ⚠️ والتبويب المخفي **ما ينبنى** (مو ينخفى بـCSS): كل شاشة تجيب
// بياناتها من السيرفر، وبناؤهن كلهن يعني أضعاف النداءات بكل فتحة.
export type ShellTab = {
  key: string
  label: string
  icon?: string
  /** ما يطلع إذا `false` — التبويب الي ينضغط وينرفض أسوأ من تبويب ما موجود. */
  show?: boolean
  render: () => ReactNode
}

export default function TabsShell({
  title, subtitle, tabs,
}: {
  title: string
  subtitle?: string
  tabs: ShellTab[]
}) {
  const shown = tabs.filter((t) => t.show !== false)
  const [tab, setTab] = useState(shown[0]?.key ?? '')
  // حالة محسوبة مو مخزّنة: لو انسحبت صلاحية وانخفى تبويبه، ما نبقى
  // واقفين على مفتاح ميّت ونعرض فراغاً.
  const active = shown.find((t) => t.key === tab) ?? shown[0]

  return (
    <div dir="rtl">
      <div className="mb-6">
        <PageHeader title={title} subtitle={subtitle} />
      </div>

      {shown.length > 1 && (
        <div className="sticky top-0 z-30 mb-4 grid grid-cols-2 gap-1.5 rounded-2xl border border-slate-200 bg-white/95 p-1.5 shadow-[0_2px_12px_rgba(15,32,64,0.08)] backdrop-blur sm:inline-flex sm:gap-2">
          {shown.map((t) => (
            <button
              key={t.key}
              onClick={() => setTab(t.key)}
              className={`rounded-xl px-3 py-2 text-[11px] font-extrabold transition sm:px-5 sm:text-xs ${
                active?.key === t.key
                  ? 'bg-gradient-to-l from-brand-500 to-brand-800 text-white shadow-md'
                  : 'text-slate-600 hover:bg-slate-50'
              }`}
            >
              {t.icon} {t.label}
            </button>
          ))}
        </div>
      )}

      {active?.render()}
    </div>
  )
}
