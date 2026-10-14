import { useLocation, useNavigate } from 'react-router-dom'
import MatrixEye from './MatrixEye'

// زر ماتركس العائم — المساعد الذكي بكل الشاشات
export default function MatrixFab() {
  const nav = useNavigate()
  const { pathname } = useLocation()
  if (pathname === '/matrix' || pathname.startsWith('/request')) return null
  return (
    <button aria-label="اسأل ماتركس" onClick={() => nav('/matrix')}
      className="fixed bottom-[80px] left-3 z-30 rounded-full shadow-[0_8px_24px_-6px_rgba(13,59,140,.55)] ring-4 ring-white transition active:scale-90">
      <MatrixEye size={48} />
    </button>
  )
}
