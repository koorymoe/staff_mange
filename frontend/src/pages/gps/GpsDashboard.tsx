import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../../api'

export default function GpsDashboard() {
  const navigate = useNavigate()
  const [stats, setStats] = useState<any>(null)
  const [devices, setDevices] = useState<any[]>([])
  const [sims, setSims] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [tab, setTab] = useState<'overview' | 'subscriptions' | 'assets'>('overview')
  const [filter, setFilter] = useState('all')
  const [selectedDevice, setSelectedDevice] = useState<any>(null)

  useEffect(() => {
    Promise.all([api.getGpsStats(), api.getGpsDevices(), api.getSimCards()])
      .then(([s, d, sim]) => { setStats(s); setDevices(d); setSims(sim) })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  const getDaysRemaining = (endDate: string) => {
    if (!endDate) return null
    const end = new Date(endDate)
    const now = new Date()
    return Math.ceil((end.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
  }

  const getSubscriptionStatus = (device: any) => {
    const days = getDaysRemaining(device.subscriptionEndDate)
    if (days === null) return 'unknown'
    if (days > 40) return 'active'
    if (days > 0) return 'expiring'
    if (days > -40) return 'expired40'
    return 'expired80'
  }

  const filteredDevices = devices.filter(d => {
    if (filter === 'all') return true
    return getSubscriptionStatus(d) === filter
  })

  const getProgressPercent = (device: any) => {
    if (!device.activationDate || !device.subscriptionEndDate) return 0
    const start = new Date(device.activationDate).getTime()
    const end = new Date(device.subscriptionEndDate).getTime()
    const now = Date.now()
    const total = end - start
    const elapsed = now - start
    return Math.min(100, Math.max(0, (elapsed / total) * 100))
  }

  const getProgressColor = (device: any) => {
    const status = getSubscriptionStatus(device)
    if (status === 'active') return '#16a34a'
    if (status === 'expiring') return '#d97706'
    if (status === 'expired40') return '#dc2626'
    return '#7c2d12'
  }

  const operatorLabel: Record<string, string> = { ZAIN: 'زين', ASIACELL: 'آسياسيل', KOREK: 'كورك', OTHER: 'أخرى' }

  if (loading) return <div className="flex items-center justify-center p-12"><div className="text-lg text-slate-500">جاري التحميل...</div></div>
  if (error) return <div className="rounded-2xl bg-red-50 p-6 text-red-600 shadow-sm">{error}</div>

  const tabs = [
    { key: 'overview', label: 'نظرة عامة', icon: '📊' },
    { key: 'subscriptions', label: 'الاشتراكات', icon: '📋' },
    { key: 'assets', label: 'أصول الشركة', icon: '📡' },
  ]

  const expiringSoon = devices.filter(d => {
    const days = getDaysRemaining(d.subscriptionEndDate)
    return days !== null && days > 0 && days <= 40
  })

  return (
    <div dir="rtl">
      {/* Header */}
      <div className="mb-6 rounded-2xl p-6 shadow-sm" style={{ backgroundColor: '#1a3a5c' }}>
        <h1 className="text-2xl font-bold text-white">📡 لوحة المراقبة</h1>
        <p className="mt-1 text-blue-200 text-sm">نظام تتبع المركبات GPS</p>
      </div>

      {/* Tabs */}
      <div className="mb-6 flex gap-2">
        {tabs.map(t => (
          <button
            key={t.key}
            onClick={() => setTab(t.key as any)}
            className={`rounded-2xl px-6 py-3 text-sm font-bold transition-all ${
              tab === t.key
                ? 'text-white shadow-sm'
                : 'bg-white text-slate-600 shadow-sm hover:bg-slate-50'
            }`}
            style={tab === t.key ? { backgroundColor: '#1a3a5c' } : {}}
          >
            {t.icon} {t.label}
          </button>
        ))}
      </div>

      {/* Overview Tab */}
      {tab === 'overview' && stats && (
        <div className="space-y-6">
          {/* Stats Grid */}
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-5">
            {[
              { label: 'إجمالي الزبائن', value: stats.totalCustomers ?? 0, icon: '👥', color: '#1a3a5c' },
              { label: 'اشتراكات فعالة', value: stats.activeSubscriptions ?? 0, icon: '✅', color: '#16a34a' },
              { label: 'تنتهي قريباً', value: stats.expiringSoon ?? 0, icon: '⚠️', color: '#d97706' },
              { label: 'منتهي +40 يوم', value: stats.expired40 ?? 0, icon: '🔴', color: '#dc2626' },
              { label: 'منتهي +80 يوم', value: stats.expired80 ?? 0, icon: '⛔', color: '#7c2d12' },
            ].map(card => (
              <div key={card.label} className="rounded-2xl bg-white p-5 shadow-sm">
                <div className="flex items-center justify-between">
                  <span className="text-2xl">{card.icon}</span>
                  <span className="text-3xl font-bold" style={{ color: card.color }}>{card.value}</span>
                </div>
                <p className="mt-2 text-sm font-medium text-slate-600">{card.label}</p>
              </div>
            ))}
          </div>

          {/* Quick Actions */}
          <div className="rounded-2xl bg-white p-6 shadow-sm">
            <h3 className="mb-4 text-lg font-bold" style={{ color: '#1a3a5c' }}>⚡ إجراءات سريعة</h3>
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
              {[
                { label: 'إضافة زبون', icon: '👤', href: '/gps/customers' },
                { label: 'طلب جهاز', icon: '📱', href: '/gps/devices' },
                { label: 'تجديد اشتراك', icon: '🔄', href: '/gps/renewals' },
                { label: 'طلب صيانة', icon: '🔧', href: '/gps/maintenance' },
              ].map(action => (
                <button
                  key={action.label}
                  onClick={() => navigate(action.href)}
                  className="rounded-2xl border-2 p-4 text-center transition-all hover:shadow-md"
                  style={{ borderColor: '#1a3a5c' }}
                >
                  <span className="text-2xl">{action.icon}</span>
                  <p className="mt-2 text-sm font-bold" style={{ color: '#1a3a5c' }}>{action.label}</p>
                </button>
              ))}
            </div>
          </div>

          {/* Expiring Soon Alert */}
          {expiringSoon.length > 0 && (
            <div className="rounded-2xl border-2 border-yellow-300 bg-yellow-50 p-6 shadow-sm">
              <h3 className="mb-3 text-lg font-bold text-yellow-800">⚠️ اشتراكات تنتهي قريباً ({expiringSoon.length})</h3>
              <div className="space-y-2">
                {expiringSoon.slice(0, 5).map(d => (
                  <div key={d.id} className="flex items-center justify-between rounded-xl bg-white p-3">
                    <span className="font-medium">{d.customer?.fullName || 'غير معروف'}</span>
                    <span className="rounded-full bg-yellow-100 px-3 py-1 text-xs font-bold text-yellow-800">
                      {getDaysRemaining(d.subscriptionEndDate)} يوم متبقي
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* SIM Quick Stats */}
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
            {[
              { label: 'إجمالي الشرائح', value: stats.simStats?.total ?? sims.length, icon: '📱', color: '#1a3a5c' },
              { label: 'شرائح متوفرة', value: stats.simStats?.available ?? sims.filter((s: any) => s.status !== 'ACTIVE').length, icon: '🟢', color: '#16a34a' },
              { label: 'شرائح مستخدمة', value: stats.simStats?.inUse ?? sims.filter((s: any) => s.status === 'ACTIVE').length, icon: '🔵', color: '#2563eb' },
            ].map(card => (
              <div key={card.label} className="rounded-2xl bg-white p-5 shadow-sm">
                <div className="flex items-center gap-3">
                  <span className="text-2xl">{card.icon}</span>
                  <div>
                    <p className="text-2xl font-bold" style={{ color: card.color }}>{card.value}</p>
                    <p className="text-sm text-slate-500">{card.label}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Subscriptions Tab */}
      {tab === 'subscriptions' && (
        <div className="space-y-4">
          {/* Filter Buttons */}
          <div className="flex flex-wrap gap-2">
            {[
              { key: 'all', label: 'الكل', color: '#1a3a5c' },
              { key: 'active', label: 'فعال', color: '#16a34a' },
              { key: 'expiring', label: 'ينتهي قريباً', color: '#d97706' },
              { key: 'expired40', label: 'منتهي +40', color: '#dc2626' },
              { key: 'expired80', label: 'منتهي +80', color: '#7c2d12' },
            ].map(f => (
              <button
                key={f.key}
                onClick={() => setFilter(f.key)}
                className={`rounded-2xl px-5 py-2 text-sm font-bold transition-all ${
                  filter === f.key ? 'text-white shadow-sm' : 'bg-white text-slate-600 shadow-sm'
                }`}
                style={filter === f.key ? { backgroundColor: f.color } : {}}
              >
                {f.label} ({f.key === 'all' ? devices.length : devices.filter(d => getSubscriptionStatus(d) === f.key).length})
              </button>
            ))}
          </div>

          <div className="flex gap-4">
            {/* Device List */}
            <div className="flex-1 space-y-3">
              {filteredDevices.map(d => {
                const days = getDaysRemaining(d.subscriptionEndDate)
                const progress = getProgressPercent(d)
                const progressColor = getProgressColor(d)
                return (
                  <div
                    key={d.id}
                    onClick={() => setSelectedDevice(d)}
                    className={`cursor-pointer rounded-2xl bg-white p-4 shadow-sm transition-all hover:shadow-md ${
                      selectedDevice?.id === d.id ? 'ring-2' : ''
                    }`}
                    style={selectedDevice?.id === d.id ? { '--tw-ring-color': '#1a3a5c' } as React.CSSProperties : {}}
                  >
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-bold" style={{ color: '#1a3a5c' }}>{d.customer?.fullName || 'غير معروف'}</span>
                      <span
                        className="rounded-full px-3 py-1 text-xs font-bold"
                        style={{ backgroundColor: progressColor + '20', color: progressColor }}
                      >
                        {days !== null ? (days > 0 ? `${days} يوم متبقي` : `منتهي منذ ${Math.abs(days)} يوم`) : 'غير محدد'}
                      </span>
                    </div>
                    <div className="mb-1 text-xs text-slate-500">
                      {d.subscriptionType === 'THREE_MONTHS' ? '3 أشهر' : d.subscriptionType === 'SIX_MONTHS' ? '6 أشهر' : d.subscriptionType === 'YEARLY' ? 'سنوي' : d.subscriptionType || '-'}
                    </div>
                    <div className="h-2 w-full rounded-full bg-slate-100">
                      <div
                        className="h-full rounded-full transition-all"
                        style={{ width: `${progress}%`, backgroundColor: progressColor }}
                      />
                    </div>
                  </div>
                )
              })}
              {filteredDevices.length === 0 && (
                <div className="rounded-2xl bg-white p-8 text-center text-slate-400 shadow-sm">لا يوجد اشتراكات</div>
              )}
            </div>

            {/* Detail Panel */}
            {selectedDevice && (
              <div className="w-80 rounded-2xl bg-white p-6 shadow-sm">
                <h3 className="mb-4 text-lg font-bold" style={{ color: '#1a3a5c' }}>تفاصيل الاشتراك</h3>
                <div className="space-y-3">
                  <div>
                    <span className="text-xs text-slate-500">الزبون</span>
                    <p className="font-bold">{selectedDevice.customer?.fullName}</p>
                  </div>
                  <div>
                    <span className="text-xs text-slate-500">الهاتف</span>
                    <p>{selectedDevice.customer?.phone || '-'}</p>
                  </div>
                  <div>
                    <span className="text-xs text-slate-500">نوع الاشتراك</span>
                    <p>{selectedDevice.subscriptionType === 'THREE_MONTHS' ? '3 أشهر' : selectedDevice.subscriptionType === 'SIX_MONTHS' ? '6 أشهر' : selectedDevice.subscriptionType === 'YEARLY' ? 'سنوي' : selectedDevice.subscriptionType || '-'}</p>
                  </div>
                  <div>
                    <span className="text-xs text-slate-500">تاريخ التفعيل</span>
                    <p>{selectedDevice.activationDate ? new Date(selectedDevice.activationDate).toLocaleDateString('ar-IQ') : '-'}</p>
                  </div>
                  <div>
                    <span className="text-xs text-slate-500">تاريخ الانتهاء</span>
                    <p>{selectedDevice.subscriptionEndDate ? new Date(selectedDevice.subscriptionEndDate).toLocaleDateString('ar-IQ') : '-'}</p>
                  </div>
                  <div>
                    <span className="text-xs text-slate-500">الحالة</span>
                    <p className="font-bold" style={{ color: getProgressColor(selectedDevice) }}>
                      {(() => {
                        const s = getSubscriptionStatus(selectedDevice)
                        if (s === 'active') return 'فعال'
                        if (s === 'expiring') return 'ينتهي قريباً'
                        if (s === 'expired40') return 'منتهي +40 يوم'
                        if (s === 'expired80') return 'منتهي +80 يوم'
                        return 'غير محدد'
                      })()}
                    </p>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Assets Tab */}
      {tab === 'assets' && (
        <div className="space-y-6">
          {/* SIM Stats */}
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
            <div className="rounded-2xl bg-white p-5 shadow-sm">
              <p className="text-sm text-slate-500">📱 إجمالي الشرائح</p>
              <p className="text-3xl font-bold" style={{ color: '#1a3a5c' }}>{sims.length}</p>
            </div>
            <div className="rounded-2xl bg-white p-5 shadow-sm">
              <p className="text-sm text-slate-500">🟢 متوفرة</p>
              <p className="text-3xl font-bold text-green-600">{sims.filter(s => s.status !== 'ACTIVE').length}</p>
            </div>
            <div className="rounded-2xl bg-white p-5 shadow-sm">
              <p className="text-sm text-slate-500">🔵 مستخدمة</p>
              <p className="text-3xl font-bold text-blue-600">{sims.filter(s => s.status === 'ACTIVE').length}</p>
            </div>
          </div>

          {/* SIM Table */}
          <div className="overflow-hidden rounded-2xl bg-white shadow-sm">
            <div className="overflow-x-auto">
            <table className="w-full text-right">
              <thead>
                <tr style={{ backgroundColor: '#1a3a5c' }}>
                  <th className="px-4 py-3 text-sm font-semibold text-white">رقم الشريحة</th>
                  <th className="px-4 py-3 text-sm font-semibold text-white">المشغل</th>
                  <th className="px-4 py-3 text-sm font-semibold text-white">الحالة</th>
                  <th className="px-4 py-3 text-sm font-semibold text-white">الزبون</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {sims.map(s => (
                  <tr key={s.id} className="hover:bg-slate-50">
                    <td className="px-4 py-3 font-medium">{s.simNumber}</td>
                    <td className="px-4 py-3 text-slate-600">{operatorLabel[s.operator] || s.operator}</td>
                    <td className="px-4 py-3">
                      <span className={`rounded-full px-3 py-1 text-xs font-bold ${s.status === 'ACTIVE' ? 'bg-green-100 text-green-800' : 'bg-slate-100 text-slate-600'}`}>
                        {s.status === 'ACTIVE' ? 'مستخدمة' : 'متوفرة'}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-slate-600">{s.customer?.fullName || '-'}</td>
                  </tr>
                ))}
                {sims.length === 0 && (
                  <tr><td colSpan={4} className="p-8 text-center text-slate-400">لا يوجد شرائح</td></tr>
                )}
              </tbody>
            </table>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
