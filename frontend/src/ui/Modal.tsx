import { useEffect, useId, useRef, type ReactNode } from 'react'

// ═══ (دازاي 11-15) النافذة الموحّدة ═══
// چانت ١٠٩ نافذة مكتوبة بالإيد + مكوّنين عامّين متنافسين. هسه وحدة:
//   • خلفية معتمة + لوحة بالنص، عنوان وزر ×
//   • Esc يسدّ، والضغط برّا اللوحة يسدّ (إلا إذا busy — حتى ما ينقطع حفظ بنصّه)
//   • السكرول جوّا النافذة، والصفحة وراها ما تتحرك
//   • التركيز يدور جوّا النافذة (Tab) ويرجع للزر الي فتحها بعد الغلق
//   • الأحجام sm/md/lg/xl، وألوانها bg-white/slate فالوضع الليلي يقلبها لحاله
export type ModalSize = 'sm' | 'md' | 'lg' | 'xl'
const SIZE: Record<ModalSize, string> = { sm: 'max-w-sm', md: 'max-w-md', lg: 'max-w-2xl', xl: 'max-w-4xl' }

// (دازاي 11-15) كومة النوافذ المفتوحة: Esc يسدّ الي فوك بس، مو كل النوافذ الي وراها
// (مثلاً askConfirm فوك نافذة تعديل — Esc يلغي السؤال وتبقى النافذة)
const stack: symbol[] = []

const FOCUSABLE = 'button:not(:disabled), input:not(:disabled), textarea:not(:disabled), select:not(:disabled), a[href], [tabindex="0"]'

export default function Modal({
  title, onClose, children, size = 'md', busy = false, closeOnBackdrop = true, footer,
  overlayClassName, panelClassName, bodyClassName = 'p-5', autoFocus = true,
  dismissible = true, showClose = true, layer = 'z-[60]',
}: {
  title?: ReactNode
  onClose: () => void
  children: ReactNode
  size?: ModalSize
  busy?: boolean
  closeOnBackdrop?: boolean
  footer?: ReactNode
  // للشاشات الي عدها ستايل خاص (مثل الصلاحيات) — يبدّل الافتراضي كله
  overlayClassName?: string
  panelClassName?: string
  bodyClassName?: string
  autoFocus?: boolean
  // (دازاي 11-15) dismissible={false}: قرار لازم يتاخذ — لا Esc ولا ضغطة برّا تسدّها
  dismissible?: boolean
  // (دازاي 11-15) showClose={false}: إذا ماكو طريق غلق أصلاً، نخفي ×
  showClose?: boolean
  // (دازاي 11-15) طبقة الـz للنوافذ الي لازم تطلع فوق غيرها (بوابات، نوافذ فوق نوافذ)
  layer?: string
}) {
  const id = useId()
  const ref = useRef<HTMLElement>(null)
  const closeRef = useRef(onClose)
  const busyRef = useRef(busy)
  const dismissRef = useRef(dismissible)
  useEffect(() => { closeRef.current = onClose; busyRef.current = busy; dismissRef.current = dismissible })

  useEffect(() => {
    const me = Symbol('modal')
    stack.push(me)
    const previous = document.activeElement as HTMLElement | null
    const overflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    if (autoFocus) {
      const first = ref.current?.querySelector<HTMLElement>('textarea, input, select') ?? ref.current?.querySelector<HTMLElement>('button')
      first?.focus()
    }
    // ⚠️ Esc على الـdocument مو على اللوحة: لو التركيز طلع برّا، Esc بعده يسدّ
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && stack.at(-1) === me && !busyRef.current && dismissRef.current) { e.stopPropagation(); closeRef.current() }
    }
    document.addEventListener('keydown', onKey)
    return () => {
      document.removeEventListener('keydown', onKey)
      stack.splice(stack.indexOf(me), 1)
      document.body.style.overflow = overflow
      if (previous?.isConnected) previous.focus()
    }
  }, [autoFocus])

  const trap = (e: React.KeyboardEvent) => {
    if (e.key !== 'Tab') return
    const items = [...(ref.current?.querySelectorAll<HTMLElement>(FOCUSABLE) ?? [])]
    const first = items[0], last = items.at(-1)
    if (e.shiftKey && document.activeElement === first) { e.preventDefault(); last?.focus() }
    else if (!e.shiftKey && document.activeElement === last) { e.preventDefault(); first?.focus() }
  }

  return (
    <div
      dir="rtl"
      className={overlayClassName ?? `fixed inset-0 ${layer} flex items-start justify-center overflow-y-auto bg-black/50 px-4 py-8 sm:items-center`}
      onMouseDown={(e) => { if (e.target === e.currentTarget && closeOnBackdrop && dismissible && !busy) onClose() }}
      onKeyDown={trap}
    >
      <section
        ref={ref}
        role="dialog"
        aria-modal="true"
        aria-labelledby={title ? id : undefined}
        className={panelClassName ?? `flex max-h-[calc(100vh-4rem)] w-full ${SIZE[size]} flex-col overflow-hidden rounded-2xl border border-slate-200 bg-white text-right shadow-xl`}
      >
        {title != null && (
          <header className={panelClassName ? undefined : 'flex items-center justify-between gap-3 border-b border-slate-100 px-5 py-3'}>
            <h3 id={id} className={panelClassName ? undefined : 'text-lg font-bold text-[#0f2040]'}>{title}</h3>
            {showClose && <button
              type="button"
              aria-label="غلق النافذة"
              disabled={busy}
              onClick={onClose}
              className={panelClassName ? undefined : 'grid h-8 w-8 place-items-center rounded-lg text-xl text-slate-400 hover:bg-slate-100 hover:text-slate-700 disabled:opacity-40'}
            >
              ×
            </button>}
          </header>
        )}
        {panelClassName ? children : <div className={`min-h-0 flex-1 overflow-y-auto ${bodyClassName}`}>{children}</div>}
        {footer && <footer className={panelClassName ? undefined : 'flex flex-wrap justify-end gap-2 border-t border-slate-100 px-5 py-3'}>{footer}</footer>}
      </section>
    </div>
  )
}
