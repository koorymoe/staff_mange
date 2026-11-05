import { useEffect, useId, useRef, type ReactNode } from 'react'

export function Dialog({
  title,
  children,
  cancel,
  busy,
}: {
  title: string
  children: ReactNode
  cancel: () => void
  busy: boolean
}) {
  const id = useId()
  const ref = useRef<HTMLElement>(null)
  useEffect(() => {
    const previous = document.activeElement as HTMLElement | null
    const overflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    const first = ref.current?.querySelector<HTMLElement>(
      'textarea, input, button',
    )
    first?.focus()
    return () => {
      document.body.style.overflow = overflow
      if (previous?.isConnected) previous.focus()
    }
  }, [])
  return (
    <div
      className="access-overlay"
      onKeyDown={(e) => {
        if (e.key === 'Escape' && !busy) {
          e.stopPropagation()
          cancel()
        }
        if (e.key === 'Tab') {
          const items = [
            ...(ref.current?.querySelectorAll<HTMLElement>(
              'button:not(:disabled), input:not(:disabled), textarea:not(:disabled), select:not(:disabled), [tabindex="0"]',
            ) ?? []),
          ]
          const first = items[0],
            last = items.at(-1)
          if (e.shiftKey && document.activeElement === first) {
            e.preventDefault()
            last?.focus()
          } else if (!e.shiftKey && document.activeElement === last) {
            e.preventDefault()
            first?.focus()
          }
        }
      }}
    >
      <section
        ref={ref}
        className="access-dialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby={id}
      >
        <header>
          <h3 id={id}>{title}</h3>
          <button
            type="button"
            aria-label="غلق النافذة"
            disabled={busy}
            onClick={cancel}
          >
            ×
          </button>
        </header>
        {children}
      </section>
    </div>
  )
}
