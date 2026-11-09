import { useEffect, useRef, useState } from 'react'
import '../pages/OrgPage.css'
export default function InlineQuestion({
  message,
  reasonRequired,
  finish,
}: {
  message: string
  reasonRequired: boolean
  finish: (result: string | null) => void
}) {
  const ref = useRef<HTMLDialogElement>(null)
  const [reason, setReason] = useState('')
  useEffect(() => {
    const previous = document.activeElement as HTMLElement | null,
      dialog = ref.current,
      overflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    dialog?.showModal()
    return () => {
      dialog?.close()
      document.body.style.overflow = overflow
      if (previous?.isConnected) previous.focus()
    }
  }, [])
  return (
    <dialog
      ref={ref}
      className="org-ui org-dialog"
      dir="rtl"
      aria-label="تأكيد الإجراء"
      onCancel={(e) => {
        e.preventDefault()
        finish(null)
      }}
    >
      <header className="org-toolbar">
        <h2>تأكيد الإجراء</h2>
        <button aria-label="إغلاق التأكيد" onClick={() => finish(null)}>
          ✕
        </button>
      </header>
      <p>{message}</p>
      <form
        onSubmit={(e) => {
          e.preventDefault()
          if (!reasonRequired || reason.trim())
            finish(reasonRequired ? reason.trim() : 'yes')
        }}
      >
        {reasonRequired && (
          <label>
            السبب
            <textarea
              required
              rows={3}
              value={reason}
              onChange={(e) => setReason(e.target.value)}
            />
          </label>
        )}
        <div className="org-actions">
          <button
            type="button"
            autoFocus={!reasonRequired}
            onClick={() => finish(null)}
          >
            تراجع
          </button>
          <button
            className="org-danger"
            type="submit"
            disabled={reasonRequired && !reason.trim()}
          >
            تأكيد
          </button>
        </div>
      </form>
    </dialog>
  )
}
