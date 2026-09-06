import { useEffect, useRef, useState } from 'react'

// ═══ قائمة إجراءات الصف («•••») ═══
//
// ⚠️ انبنت لأن زر «•••» بشاشة الشكاوى **چان ميّتاً بالعمل**: يفتح
// لوحة التفاصيل بأسفل الصفحة، فالمستخدم يضغط ويشوف ماكو شي يتحرّك
// كدامه ويستنتج إن الزر خربان. زر يسوي شي بعيد عن مكان الضغط =
// زر ما يسوي شي، من ناحية المستخدم.
//
// فالقائمة تنفتح **مكانها** جنب الزر، وكل إجراء مكتوب باسمه.
//
// ماكو مكوّن قائمة منسدلة بالمشروع، فانبنى مرة وحدة حتى يخدم باقي
// الجداول بدل ما ينتنسخ.

export type RowAction = {
  label: string
  icon?: string
  onClick: () => void
  /** إجراء خطر (حذف/رفض) — ينلوّن أحمر */
  danger?: boolean
  hidden?: boolean
}

export default function RowActions({ actions, label = 'إجراءات' }: {
  actions: RowAction[]
  label?: string
}) {
  const [open, setOpen] = useState(false)
  const boxRef = useRef<HTMLDivElement>(null)

  // ⚠️ الإغلاق بالضغط برّا **وبـEsc**: قائمة تنفتح وما تنغلق إلا
  // باختيار إجراء تجبر المستخدم يختار شي ما يريده.
  useEffect(() => {
    if (!open) return
    const onDown = (e: MouseEvent) => {
      if (boxRef.current && !boxRef.current.contains(e.target as Node)) setOpen(false)
    }
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') setOpen(false) }
    document.addEventListener('mousedown', onDown)
    document.addEventListener('keydown', onKey)
    return () => {
      document.removeEventListener('mousedown', onDown)
      document.removeEventListener('keydown', onKey)
    }
  }, [open])

  const shown = actions.filter((a) => !a.hidden)
  if (shown.length === 0) return null

  return (
    <div ref={boxRef} className="relative inline-block">
      <button
        type="button"
        aria-haspopup="menu"
        aria-expanded={open}
        aria-label={label}
        onClick={(e) => { e.stopPropagation(); setOpen((o) => !o) }}
        className="rounded-lg px-3 py-1 text-xs font-bold transition-colors"
        style={{
          backgroundColor: open ? 'var(--sf-info)' : 'var(--sf-sunken)',
          color: 'var(--t-body)',
        }}
      >
        •••
      </button>

      {open && (
        <div
          role="menu"
          // ⚠️ z عالية: القائمة تنفتح فوق صفوف الجدول الي تحتها،
          // وبلاها تنقص نصها ورا الصف التالي.
          className="absolute z-50 mt-1 min-w-[11rem] overflow-hidden rounded-xl border shadow-lg"
          style={{
            insetInlineEnd: 0,
            backgroundColor: 'var(--sf-card)',
            borderColor: 'var(--bd-line)',
          }}
        >
          {shown.map((a) => (
            <button
              key={a.label}
              role="menuitem"
              type="button"
              onClick={(e) => { e.stopPropagation(); setOpen(false); a.onClick() }}
              className="flex w-full items-center gap-2 px-3 py-2 text-right text-xs font-medium transition-colors hover:brightness-95"
              style={{
                color: a.danger ? 'var(--t-danger)' : 'var(--t-body)',
                backgroundColor: 'var(--sf-card)',
              }}
            >
              {a.icon && <span>{a.icon}</span>}
              <span>{a.label}</span>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
