import type { ReactNode } from 'react'
import Modal from '../../ui/Modal'

// (دازاي 11-15) صارت النافذة الموحّدة (ui/Modal) — بستايل الصلاحيات الخاص (access-overlay)،
// ونفس السلوك: Esc والتركيز الدوّار وقفل الغلق وقت الحفظ.
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
  return (
    <Modal title={title} onClose={cancel} busy={busy} closeOnBackdrop={false} overlayClassName="access-overlay" panelClassName="access-dialog">
      {children}
    </Modal>
  )
}
