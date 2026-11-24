// ═══ 🔧 «خطأ فني من الكادر السابق» داخل الفاتورة (ل) 11-24 دازاي ═══
// «يجب أن يُذكر بالتقرير أو ملاحظات الفاتورة إنه خطأ فني من الكادر السابق، حتى ينخصم ويُضاف».
// بعد ما تنحفظ الفاتورة ينسجّل تصحيح بـ/rework ويروح للمراقب يوافق (نفس صفحة «الخطأ الفني والتصحيح»).
import { emptyRework, reworkProblem, type ReworkState } from '../utils/rework'

export default function ReworkFlag({ value, onChange }: { value: ReworkState; onChange: (v: ReworkState) => void }) {
  const problem = reworkProblem(value)
  return (
    <div className={`rounded-xl border-2 p-3 ${value.on ? 'border-rose-300 bg-rose-50' : 'border-slate-200 bg-white'}`} data-testid="rework-flag">
      <label className="flex cursor-pointer items-center gap-2 text-sm font-bold text-slate-800">
        <input type="checkbox" className="h-4 w-4" checked={value.on} onChange={(e) => onChange(e.target.checked ? { ...value, on: true } : emptyRework)} />
        🔧 خطأ فني من الكادر السابق (رجعنا نصلّح غلط غيرنا)
      </label>
      {value.on && (
        <div className="mt-2 grid gap-2 sm:grid-cols-[10rem_1fr]">
          <input value={value.orig} onChange={(e) => onChange({ ...value, orig: e.target.value })} placeholder="الحجز الأصلي (B…)" dir="ltr" className="rounded-lg border border-slate-300 px-3 py-2 text-sm" />
          <input value={value.note} onChange={(e) => onChange({ ...value, note: e.target.value })} placeholder="شنو چان الخطأ الفني وشنو صلّحتوا" className="rounded-lg border border-slate-300 px-3 py-2 text-sm" />
          <p className="text-xs text-slate-600 sm:col-span-2">المراقب يوافق، ووقت هذا الحجز × سعر ساعتكم ينضاف إلكم وينخصم من الكادر الأول.{problem && <b className="ms-1 text-rose-700">{problem}</b>}</p>
        </div>
      )}
    </div>
  )
}
