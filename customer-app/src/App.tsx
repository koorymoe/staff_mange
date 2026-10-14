import { useCallback, useEffect, useState } from 'react'
import { Navigate, Route, Routes, useLocation } from 'react-router-dom'
import { api, getToken, setToken, type Account } from './api'
import { clearCache } from './data'
import BottomNav from './components/BottomNav'
import MatrixFab from './components/MatrixFab'
import AuthPage from './pages/AuthPage'
import HomePage from './pages/HomePage'
import ServicesPage from './pages/ServicesPage'
import ServiceDetailPage from './pages/ServiceDetailPage'
import OffersPage from './pages/OffersPage'
import NewRequestPage from './pages/NewRequestPage'
import BookingsPage from './pages/BookingsPage'
import BookingDetailPage from './pages/BookingDetailPage'
import ChatPage from './pages/ChatPage'
import AccountPage, { AboutPage, BranchesPage, EditProfilePage, FavoritesPage, NotificationsPage, PrivacyPage } from './pages/AccountPage'

// الشاشات الي بيها الشريط السفلي (الخمسة الرئيسية)
const TABS = ['/', '/services', '/offers', '/bookings', '/account']

export default function App() {
  const [me, setMe] = useState<Account | null>(null)
  const [ready, setReady] = useState(() => !getToken())
  const [badge, setBadge] = useState(0)
  const { pathname } = useLocation()

  useEffect(() => {
    if (!getToken()) return
    api.me().then(setMe).catch(() => setToken(null)).finally(() => setReady(true))
  }, [])
  useEffect(() => {
    const out = () => setMe(null)
    window.addEventListener('amani-logout', out)
    return () => window.removeEventListener('amani-logout', out)
  }, [])
  useEffect(() => { window.scrollTo(0, 0) }, [pathname])

  const onSession = useCallback((token: string, acc: Account) => { setToken(token); clearCache(); setMe(acc) }, [])
  const logout = () => { setToken(null); clearCache(); setMe(null) }

  if (!ready) return <div className="grid h-full place-items-center bg-navy-deep" />
  if (!me) return <AuthPage onSession={onSession} />

  const tabbed = TABS.includes(pathname)
  return (
    <div className={`mx-auto min-h-full max-w-lg ${tabbed ? 'pb-24' : ''}`}>
      <Routes>
        <Route path="/" element={<HomePage me={me} onBadge={setBadge} />} />
        <Route path="/services" element={<ServicesPage />} />
        <Route path="/services/:id" element={<ServiceDetailPage />} />
        <Route path="/offers" element={<OffersPage />} />
        <Route path="/request" element={<NewRequestPage me={me} />} />
        <Route path="/bookings" element={<BookingsPage />} />
        <Route path="/bookings/:id" element={<BookingDetailPage />} />
        <Route path="/matrix" element={<ChatPage me={me} />} />
        <Route path="/account" element={<AccountPage me={me} onLogout={logout} />} />
        <Route path="/account/edit" element={<EditProfilePage me={me} onChange={setMe} />} />
        <Route path="/account/favorites" element={<FavoritesPage />} />
        <Route path="/account/branches" element={<BranchesPage />} />
        <Route path="/account/notifications" element={<NotificationsPage />} />
        <Route path="/account/about" element={<AboutPage />} />
        <Route path="/account/privacy" element={<PrivacyPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
      {tabbed && <><MatrixFab /><BottomNav badge={badge} /></>}
    </div>
  )
}
