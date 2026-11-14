import { useState } from 'react'
import { askConfirm } from './dialogs'

// ═══ 🗑️ زر حذف واحد بكل النظام (دازاي 11-15) ═══
// نفس الشكل ونفس الكلمة «🗑️ حذف»، ودائماً يسأل قبل — وينقفل لحد ما يخلص الحذف حتى ما تنضغط مرتين.
//   <DeleteButton confirm="تحذف المورد «س»؟" onConfirm={() => api.deleteX(id)} />
export function DeleteButton({ confirm, onConfirm, label = '🗑️ حذف', className = '', disabled }: {
  confirm: string
  onConfirm: () => Promise<unknown> | unknown
  label?: string
  /** بس للمكان (موقع/مسافة) — اللون والشكل ثابتين */
  className?: string
  disabled?: boolean
}) {
  const [busy, setBusy] = useState(false)
  const click = async () => {
    if (busy) return
    if (!(await askConfirm(confirm, { danger: true, okLabel: 'حذف' }))) return
    setBusy(true)
    try { await onConfirm() } finally { setBusy(false) }
  }
  return (
    <button type="button" onClick={() => void click()} disabled={disabled || busy} aria-busy={busy}
      className={`rounded-lg border border-red-300 bg-white px-2.5 py-1 text-xs font-bold text-red-600 hover:bg-red-50 disabled:opacity-50 dark:border-red-500/60 dark:bg-transparent dark:text-red-400 dark:hover:bg-red-950/40 ${className}`}>
      {busy ? 'جاري الحذف…' : label}
    </button>
  )
}
