import { useCallback, useEffect, useState } from 'react'
import { Navigate, Route, Routes } from 'react-router-dom'
import { api, getToken, setToken, type Account } from './api'
import BottomNav from './components/BottomNav'
import AuthPage from './pages/AuthPage'
import HomePage from './pages/HomePage'
import ChatPage from './pages/ChatPage'
import BookingsPage from './pages/BookingsPage'
import AccountPage from './pages/AccountPage'

export default function App() {
  const [me, setMe] = useState<Account | null>(null)
  const [ready, setReady] = useState(() => !getToken())
  const [badge, setBadge] = useState(0)

  useEffect(() => {
    if (!getToken()) return
    api.me().then(setMe).catch(() => setToken(null)).finally(() => setReady(true))
  }, [])
  useEffect(() => {
    const out = () => setMe(null)
    window.addEventListener('amani-logout', out)
    return () => window.removeEventListener('amani-logout', out)
  }, [])

  const onSession = useCallback((token: string, acc: Account) => { setToken(token); setMe(acc) }, [])
  const logout = () => { setToken(null); setMe(null) }

  if (!ready) return <div className="grid h-full place-items-center text-slate-400">…</div>
  if (!me) return <AuthPage onSession={onSession} />

  return (
    <div className="mx-auto min-h-full max-w-lg pb-24">
      <Routes>
        <Route path="/" element={<HomePage me={me} onBadge={setBadge} />} />
        <Route path="/matrix" element={<ChatPage me={me} />} />
        <Route path="/bookings" element={<BookingsPage />} />
        <Route path="/account" element={<AccountPage me={me} onChange={setMe} onLogout={logout} />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
      <BottomNav badge={badge} />
    </div>
  )
}
