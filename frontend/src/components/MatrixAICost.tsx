import { useEffect, useState } from 'react'
import { api, type AIUsageReport } from '../api'
import MatrixNote from './MatrixNote'
import OwnerSwitch from './OwnerSwitch'

// ═══ كلفة ذكاء ماتركس — طلب (ع) 10-05: «اريد اقل سعر ممكن» ═══
// الكلفة الحقيقية (من توكنات كل نداء لهايكو)، وتوقّع الشهر، ولكل ميزة مفتاح:
// المطفية ترجع لقواعد النظام فالكلفة تنزل بلا ما يوقف شي.
const usd = (n: number) => `$${n < 1 ? n.toFixed(3) : n.toFixed(2)}`
const HINTS: Record<string, string> = {
  PEER: 'مطفي = اقتراحات الفضفضة وتحليل المشاكل الوظيفية تنكتب بالقواعد (مجانية).',
  GUIDE: 'مطفي = ماتركس يستعمل تعليمة المدير المكتوبة كما هي بدل ما يصيغها.',
  JUDGE: 'مطفي = الأحكام تنحسب بالقواعد (أدق شوية أقل، بس مجانية).',
  DISCOVERY: 'مطفي = «ماتركس اكتشف» يطلع بجمل جاهزة بدل الشرح.',
  VOICE: 'مطفي = الفويس يتحول لنص (مجاني على سيرفرك) بلا ملخّص وحساب إنجاز وقاموس لهجة.',
  ASK: 'مطفي = «اسأل ماتركس» يجاوب من الأرقام بس.',
  LEARNING: 'مطفي = توقعات ماتركس بالقواعد.',
  EMPLOYEE_REPORT: 'مطفي = تقرير الموظف بلا الخلاصة المكتوبة.',
}

export default function MatrixAICost() {
  const [r, setR] = useState<AIUsageReport | null>(null)
  useEffect(() => { void api.getAIUsage().then(setR).catch(() => {}) }, [])
  if (!r) return null
  const feats = [...r.features].sort((a, b) => b.cost30 - a.cost30)
  const max = Math.max(...r.days.map((d) => d.cost), 0.0001)
  return (
    <details dir="rtl" className="rounded-2xl border border-amber-200 bg-white p-3 text-slate-800">
      <summary className="cursor-pointer text-sm font-extrabold text-amber-900">💲 كلفة ذكاء ماتركس <span className="text-[11px] font-normal text-slate-500">اليوم {usd(r.today)} · هالشهر {usd(r.month)} · المتوقع للشهر {usd(r.forecast)}</span></summary>
      <MatrixNote className="my-2">
        هاي الكلفة الحقيقية من كل نداء لهايكو (مو تقدير). الحسبة تقريبية بالدولار حسب سعر هايكو ($1 للمليون داخل، $5 للمليون طالع). أي ميزة تطفيها ترجع لقواعد النظام وكلفتها تصير صفر.
      </MatrixNote>
      <div className="mb-3 flex h-16 items-end gap-0.5" title="آخر ٣٠ يوم">
        {[...r.days].reverse().map((d) => (
          <div key={d.day} title={`${d.day}: ${usd(d.cost)}`} className="flex-1 rounded-t bg-amber-400" style={{ height: `${Math.max(2, (d.cost / max) * 100)}%` }} />
        ))}
      </div>
      <div className="space-y-2">
        {feats.map((f) => (
          <div key={f.feature} className="rounded-xl border border-slate-200 p-2 text-xs">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <b className="text-sm">{f.label || f.feature}</b>
              <span>آخر ٣٠ يوم: <b>{usd(f.cost30)}</b> · اليوم {usd(f.costToday)} · {f.calls} نداء</span>
            </div>
            {r.switches[f.feature] && (
              <div className="mt-2"><OwnerSwitch switchKey={r.switches[f.feature]} label="يستعمل هايكو" hint={HINTS[f.feature] ?? ''} /></div>
            )}
          </div>
        ))}
      </div>
    </details>
  )
}
