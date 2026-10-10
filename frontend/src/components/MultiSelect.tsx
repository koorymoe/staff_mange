import { useEffect, useRef, useState } from 'react'

// ═══ قائمة منسدلة باختيار متعدد ═══
//
// «الخدمات بالحجز الجديد ماريدها هيج — أريدها قائمة منسدلة أفتحها
// وأسدها وأختار منها العدد الي يعجبني».
//
// ⚠️ ليش مو `<select multiple>` الجاهزة؟ لأنها بالمتصفح تحتاج
// Ctrl+نقرة حتى تختار أكثر من وحدة، وبالموبايل تطلع بشكل مختلف كل
// نظام. الإداري الي يحجز من تلفونه ما راح يعرف يختار خدمتين، فينتهي
// بحجز ناقص خدمة.
//
// وليش مو قائمة مربّعات مفتوحة (الي كانت)؟ لأن الخدمات صارن ٢٠+،
// فالقائمة تاخذ نص الصفحة وتدفن باقي خانات الحجز تحتها.

export interface MultiSelectOption {
  id: string
  name: string
}

export default function MultiSelect({
  options, selected, onChange, placeholder = 'اختر من القائمة', searchable = true, emptyText = 'ماكو خيارات',
}: {
  options: MultiSelectOption[]
  selected: string[]
  onChange: (ids: string[]) => void
  placeholder?: string
  searchable?: boolean
  emptyText?: string
}) {
  const [open, setOpen] = useState(false)
  const [q, setQ] = useState('')
  const boxRef = useRef<HTMLDivElement>(null)

  // ⚠️ الضغط برّا يسدّها: بدونه تبقى مفتوحة وتغطي الخانات الي تحتها،
  // والإداري يظن الصفحة علّقت.
  useEffect(() => {
    if (!open) return
    const onDown = (e: MouseEvent) => {
      if (boxRef.current && !boxRef.current.contains(e.target as Node)) setOpen(false)
    }
    const onEsc = (e: KeyboardEvent) => { if (e.key === 'Escape') setOpen(false) }
    document.addEventListener('mousedown', onDown)
    document.addEventListener('keydown', onEsc)
    return () => {
      document.removeEventListener('mousedown', onDown)
      document.removeEventListener('keydown', onEsc)
    }
  }, [open])

  const toggle = (id: string) =>
    onChange(selected.includes(id) ? selected.filter((x) => x !== id) : [...selected, id])

  const shown = q.trim()
    ? options.filter((o) => o.name.toLowerCase().includes(q.trim().toLowerCase()))
    : options

  const picked = options.filter((o) => selected.includes(o.id))

  return (
    <div ref={boxRef} className="relative">
      {/* الزر — يعرض المختار مو «٣ مختارة» بس: الإداري لازم يتأكد
          شنو اختار قبل ما يحفظ بلا ما يعيد فتح القائمة. */}
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className={`flex w-full items-center gap-2 rounded-xl border px-3 py-2.5 text-right text-sm transition ${
          open ? 'border-brand-500 ring-2 ring-brand-100' : 'border-slate-300 hover:border-slate-400'
        }`}
      >
        <span className="flex-1 min-w-0">
          {picked.length === 0 ? (
            <span className="text-slate-400">{placeholder}</span>
          ) : (
            <span className="flex flex-wrap gap-1">
              {picked.map((o) => (
                <span key={o.id} className="rounded-lg bg-brand-50 px-2 py-0.5 text-[11px] font-bold text-brand-700">
                  {o.name}
                </span>
              ))}
            </span>
          )}
        </span>
        {picked.length > 0 && (
          <span className="shrink-0 rounded-full bg-brand-600 px-2 py-0.5 text-[10px] font-black text-white">
            {picked.length}
          </span>
        )}
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"
          strokeLinecap="round" strokeLinejoin="round"
          className={`shrink-0 text-slate-400 transition-transform duration-200 ${open ? 'rotate-180' : ''}`}>
          <polyline points="6 9 12 15 18 9" />
        </svg>
      </button>

      {open && (
        <div className="absolute z-30 mt-1 w-full rounded-xl border border-slate-200 bg-white shadow-xl">
          {searchable && options.length > 8 && (
            <div className="border-b border-slate-100 p-2">
              <input
                autoFocus
                value={q}
                onChange={(e) => setQ(e.target.value)}
                placeholder="🔍 دوّر بالاسم..."
                className="w-full rounded-lg border border-slate-200 px-3 py-1.5 text-xs outline-none focus:border-brand-500"
              />
            </div>
          )}

          {/* ⚠️ ارتفاع محدود مع تمرير: ٢٠ خدمة بلا سقف تطلع قائمة
              أطول من الشاشة وما تكدر توصل آخرها بالموبايل. */}
          <div className="max-h-64 overflow-y-auto p-1">
            {shown.length === 0 && (
              <p className="px-3 py-6 text-center text-xs text-slate-400">
                {q ? 'ماكو نتيجة بهذا البحث' : emptyText}
              </p>
            )}
            {shown.map((o) => {
              const checked = selected.includes(o.id)
              return (
                <label
                  key={o.id}
                  className={`flex cursor-pointer items-center gap-2.5 rounded-lg px-3 py-2 text-sm transition-colors ${
                    checked ? 'bg-brand-50 font-bold text-brand-800' : 'text-slate-600 hover:bg-slate-50'
                  }`}
                >
                  <input type="checkbox" checked={checked} onChange={() => toggle(o.id)} className="h-4 w-4" />
                  <span className="flex-1">{o.name}</span>
                </label>
              )
            })}
          </div>

          <div className="flex items-center justify-between gap-2 border-t border-slate-100 p-2">
            <span className="text-[11px] text-slate-500">اختير {selected.length}</span>
            <div className="flex gap-1.5">
              {selected.length > 0 && (
                <button
                  type="button"
                  onClick={() => onChange([])}
                  className="rounded-lg border border-slate-300 px-2.5 py-1 text-[11px] font-bold text-slate-600 hover:bg-slate-50"
                >
                  مسح الكل
                </button>
              )}
              <button
                type="button"
                onClick={() => setOpen(false)}
                className="rounded-lg bg-brand-600 px-3 py-1 text-[11px] font-bold text-white hover:bg-brand-700"
              >
                تم
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
