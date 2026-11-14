import { useEffect, useState } from 'react'
import { NOTIFY_MS, onNotify, pendingToasts, type Toast } from './notify'

// ينحط مرة وحدة بـ main.tsx يم DialogHost — يرسم تنبيهات notify() مرتبة فوگ بعض (دازاي 11-15)
const KIND_CLS: Record<Toast['kind'], string> = {
  error: 'bg-red-600 text-white dark:bg-red-700',
  success: 'bg-emerald-600 text-white dark:bg-emerald-700',
  info: 'bg-[#0f2040] text-white dark:bg-slate-700',
}

export function NotifyHost() {
  const [list, setList] = useState<Toast[]>(() => pendingToasts.splice(0))
  useEffect(() => {
    onNotify((t) => setList((l) => [...l, t]))
    return () => onNotify(null)
  }, [])
  const drop = (id: number) => setList((l) => l.filter((t) => t.id !== id))
  if (!list.length) return null
  return (
    <div dir="rtl" className="pointer-events-none fixed left-1/2 top-4 z-[210] flex w-[min(92vw,28rem)] -translate-x-1/2 flex-col gap-2">
      {list.map((t) => <ToastItem key={t.id} t={t} drop={drop} />)}
    </div>
  )
}

function ToastItem({ t, drop }: { t: Toast; drop: (id: number) => void }) {
  useEffect(() => {
    const h = window.setTimeout(() => drop(t.id), NOTIFY_MS)
    return () => window.clearTimeout(h)
  }, [t.id, drop])
  return (
    <div role={t.kind === 'error' ? 'alert' : 'status'} onClick={() => drop(t.id)}
      className={`pointer-events-auto cursor-pointer whitespace-pre-line rounded-xl px-4 py-3 text-sm font-bold shadow-2xl ${KIND_CLS[t.kind]}`}>
      {t.message}
    </div>
  )
}
