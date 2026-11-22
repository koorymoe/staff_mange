import { useContext } from 'react'
import { useNavigate } from 'react-router-dom'
import { GlobalBackContext } from './backContext'

// ═══ → رجوع — زر رجوع واحد بكل النظام (دازاي 11-15) ═══
// يرجع للصفحة الي جاي منها؛ وإذا فتح الرابط مباشرة (ماكو تاريخ) يروح لـ`to`.
// نفس الشكل بكل مكان — فوگ الخلفية الفاتحة وفوگ الترويسة الغامقة.
export function BackLink({ to = '/', label = '→ رجوع', className = '' }: { to?: string; label?: string; className?: string }) {
  const navigate = useNavigate()
  const global = useContext(GlobalBackContext)
  const back = () => {
    const idx = (window.history.state as { idx?: number } | null)?.idx ?? 0
    if (idx > 0) navigate(-1)
    else navigate(to)
  }
  if (global) return null
  return (
    <button type="button" onClick={back}
      className={`back-btn inline-flex items-center gap-1 rounded-lg border border-slate-300 bg-white/90 px-3 py-1 text-xs font-bold text-slate-700 hover:bg-white ${className}`}>
      {label}
    </button>
  )
}

// زر الرجوع العام — يطلع فوگ كل شاشة بالنظام عدا الرئيسية.
export function GlobalBack({ path }: { path: string }) {
  const navigate = useNavigate()
  if (path === '/' || path === '') return null
  const back = () => {
    const idx = (window.history.state as { idx?: number } | null)?.idx ?? 0
    if (idx > 0) navigate(-1)
    else navigate('/')
  }
  return (
    <div className="mb-3 flex">
      <button type="button" onClick={back}
        className="back-btn inline-flex items-center gap-1 rounded-lg border border-slate-300 bg-white/90 px-3 py-1 text-xs font-bold text-slate-700 hover:bg-white">
        → رجوع
      </button>
    </div>
  )
}
