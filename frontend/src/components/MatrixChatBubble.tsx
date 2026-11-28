import { Suspense, lazy, useState } from 'react'
import { Link } from 'react-router-dom'
import { useSession, isAdminOrOwner } from '../session'

// ═══ 💬 ماتركس كفقاعة دردشة (دازاي 11-28) ═══
// «محادثة ماتركس تكون بعلامة دردشة جوه اسفل يسار الشاشه لكل موظف عنده صلاحية…
//  ومايطلع بلقائمة الجانبيه — لان ماخذ حيز». تنفتح نافذة صغيرة فوق الشاشة الحالية.
const MatrixChat = lazy(() => import('./MatrixChat'))

export default function MatrixChatBubble() {
  const { employee, permissions } = useSession()
  const [open, setOpen] = useState(false)
  if (!employee || !(isAdminOrOwner(employee) || permissions.includes('agent_matrix'))) return null
  return (
    <>
      {open && (
        <div dir="rtl" className="fixed bottom-24 left-4 z-50 flex h-[75vh] w-[min(440px,calc(100vw-2rem))] flex-col overflow-hidden rounded-2xl border border-slate-200 bg-[var(--sf-card)] shadow-2xl">
          <div className="flex items-center justify-between bg-[#0f2040] px-4 py-2.5 text-white">
            <b className="text-sm">🤖 ماتركس</b>
            <span className="flex items-center gap-3 text-xs">
              <Link to="/matrix-chat" onClick={() => setOpen(false)} className="underline opacity-80 hover:opacity-100">صفحة كاملة</Link>
              <button type="button" aria-label="سدّ" onClick={() => setOpen(false)} className="text-lg leading-none">✕</button>
            </span>
          </div>
          <div className="min-h-0 flex-1 overflow-y-auto p-2">
            <Suspense fallback={<p className="p-6 text-center text-slate-400">جاري التحميل…</p>}>
              <MatrixChat />
            </Suspense>
          </div>
        </div>
      )}
      <button type="button" onClick={() => setOpen((o) => !o)} aria-label="محادثة ماتركس" title="محادثة ماتركس"
        className="fixed bottom-6 left-4 z-50 flex h-14 w-14 items-center justify-center rounded-full bg-[#0f2040] text-2xl text-white shadow-xl ring-4 ring-white/70 transition hover:scale-105">
        {open ? '✕' : '💬'}
      </button>
    </>
  )
}
