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
export function Modal({ children, title, onClose, wide }: { children: React.ReactNode; title: string; onClose: () => void; wide?: boolean }) {
  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center bg-black/50 overflow-y-auto py-8 px-4" onClick={onClose}>
      <div className={`bg-white rounded-2xl shadow-xl w-full ${wide ? 'max-w-2xl' : 'max-w-md'} p-6`} onClick={e => e.stopPropagation()}>
        <div className="flex items-center justify-between border-b pb-3 mb-4">
          <h3 className="text-lg font-bold text-[var(--color-brand-900)]">{title}</h3>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-700 text-xl">✕</button>
        </div>
        {children}
      </div>
    </div>
  )
}

export function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="block">
      <span className="text-xs text-gray-500 font-medium">{label}</span>
      <div className="mt-1">{children}</div>
    </label>
  )
}
