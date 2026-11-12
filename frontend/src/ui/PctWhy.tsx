import { useState, type ReactNode } from 'react'
import { createPortal } from 'react-dom'
import { pctOf } from './pct'

// ═══ «كل نسبة بالنظام لازم الها سبب» — دازاي 11-12 ═══
// أي نسبة تنعرض للمستخدم تنلف بهذا: تنضغط وتطلّع شباك بيه
//   - الحسبة بالأرقام الحقيقية (البسط والمقام بأسمائهم)،
//   - شنو تعني النسبة ومنين تجي (أي بيانات بالنظام)،
//   - وتفاصيل اختيارية (قائمة أو جدول) تثبت الرقم.
// قاعدة: ممنوع نسبة بلا شرح. إذا المقام صفر نگول ليش ماكو نسبة.

export type PctParts = {
  /** وصف البسط، مثل «حجوزات خلصت» */
  numLabel: string
  num: number
  /** وصف المقام، مثل «كل حجوزات الأسبوع» */
  denLabel: string
  den: number
}

const fmt = (n: number) => (Number.isInteger(n) ? n.toLocaleString('en-US') : (Math.round(n * 10) / 10).toLocaleString('en-US'))

export default function PctWhy({
  title, value, parts, formula, meaning, source, children, className = '', suffix = '%',
}: {
  /** اسم النسبة، مثل «نجاح الأسبوع» */
  title: string
  /** النسبة المعروضة (إذا ما انطت تنحسب من parts) */
  value?: number | null
  /** بسط/مقام — الأشيع */
  parts?: PctParts
  /** حسبة مو بسط/مقام (معدل موزون، فرق…) — جملة بالأرقام */
  formula?: ReactNode
  /** شنو تعني وشلون تقراها */
  meaning: ReactNode
  /** منين تجي البيانات بالنظام */
  source: ReactNode
  /** تفاصيل تثبت الرقم (اختياري) */
  children?: ReactNode
  className?: string
  suffix?: string
}) {
  const [open, setOpen] = useState(false)
  const v = value !== undefined ? value : parts ? pctOf(parts.num, parts.den) : null
  const shown = v == null ? '—' : `${v}${suffix}`
  return (
    <>
      {/* span مو button: النسبة أحياناً جوّا بطاقة أو صف هو نفسه زر (زر جوّا زر = HTML غلط) */}
      <span role="button" tabIndex={0}
        onClick={(e) => { e.preventDefault(); e.stopPropagation(); setOpen(true) }}
        onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); e.stopPropagation(); setOpen(true) } }}
        title="اضغط حتى تشوف منين جات النسبة" aria-label={`${title} ${shown} — منين جات؟`}
        className={`cursor-help underline decoration-dotted decoration-1 underline-offset-4 ${className}`}>
        {shown}
      </span>
      {/* portal: الشباك برّا البطاقة — لو جوّا رابط، الضغط داخله چان يفتح الرابط */}
      {open && createPortal(
        <div className="fixed inset-0 z-[60] flex items-start justify-center overflow-y-auto bg-black/50 p-4" onClick={(e) => { e.stopPropagation(); setOpen(false) }}>
          <div dir="rtl" role="dialog" aria-label={`منين جات ${title}`} onClick={(e) => e.stopPropagation()}
            className="mt-10 w-full max-w-xl rounded-2xl bg-white p-5 text-right text-sm font-normal text-slate-800 shadow-2xl dark:bg-slate-800 dark:text-slate-100">
            <div className="mb-3 flex items-center justify-between gap-2">
              <h3 className="text-lg font-extrabold">{title}: {shown}</h3>
              <button type="button" onClick={() => setOpen(false)} className="text-2xl leading-none text-slate-400" aria-label="سد">×</button>
            </div>
            <div className="space-y-3">
              <div className="rounded-lg bg-slate-100 px-3 py-2 dark:bg-slate-700" data-testid="pct-formula">
                <p className="mb-0.5 text-[11px] font-bold text-slate-500 dark:text-slate-300">🧮 الحسبة</p>
                {parts ? (
                  parts.den > 0
                    ? <p>{fmt(parts.num)} {parts.numLabel} من أصل {fmt(parts.den)} {parts.denLabel} = <b>{shown}</b></p>
                    : <p>ماكو {parts.denLabel} بعد، فماكو نسبة تنحسب.</p>
                ) : <div>{formula}</div>}
              </div>
              <div>
                <p className="mb-0.5 text-[11px] font-bold text-slate-500 dark:text-slate-300">💡 شنو تعني</p>
                <div>{meaning}</div>
              </div>
              <div>
                <p className="mb-0.5 text-[11px] font-bold text-slate-500 dark:text-slate-300">📂 منين تجي الأرقام</p>
                <div>{source}</div>
              </div>
              {children && <div className="max-h-[50vh] overflow-y-auto border-t border-slate-200 pt-3 dark:border-slate-600">{children}</div>}
            </div>
          </div>
        </div>,
        document.body,
      )}
    </>
  )
}
