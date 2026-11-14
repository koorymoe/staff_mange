import SharedModal from '../../ui/Modal'
import { type ProjectCandidate, candidateLabel } from './shared'

// EmployeeOptions يرسم <optgroup> لكل مجموعة بنفس ترتيب وصولها من السيرفر.
export function EmployeeOptions({ candidates }: { candidates: ProjectCandidate[] }) {
  const groups: { label: string; items: ProjectCandidate[] }[] = []
  candidates.forEach(c => {
    const last = groups[groups.length - 1]
    if (last && last.label === c.groupLabel) last.items.push(c)
    else groups.push({ label: c.groupLabel, items: [c] })
  })
  return (
    <>
      {groups.map(g => (
        <optgroup key={g.label} label={g.label}>
          {g.items.map(c => <option key={c.id} value={c.id}>{candidateLabel(c)}</option>)}
        </optgroup>
      ))}
    </>
  )
}

export function Info({ icon, value }: { icon: string; value: string | null }) {
  return (
    <div className="flex items-center gap-2 bg-gray-50 rounded-lg p-2 text-sm text-gray-700">
      <span>{icon}</span>
      <span className="truncate">{value || '---'}</span>
    </div>
  )
}

export function RejStat({ label, value, color, small }: { label: string; value: string | number; color: string; small?: boolean }) {
  return (
    <div className={`bg-gradient-to-br ${color} text-white rounded-xl p-4 text-center`}>
      <div className={`font-bold ${small ? 'text-lg' : 'text-3xl'} my-1`}>{value}</div>
      <div className="text-xs opacity-90">{label}</div>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Shared UI bits
// ---------------------------------------------------------------------------
// (دازاي 11-15) صارت النافذة الموحّدة (ui/Modal) — نفس الاسم حتى ما تتغيّر الاستدعاءات.
export function Modal({ children, title, onClose, wide }: { children: React.ReactNode; title: string; onClose: () => void; wide?: boolean }) {
  return <SharedModal title={title} onClose={onClose} size={wide ? 'lg' : 'md'} bodyClassName="p-6">{children}</SharedModal>
}

export function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="block">
      <span className="text-xs text-gray-500 font-medium">{label}</span>
      <div className="mt-1">{children}</div>
    </label>
  )
}
