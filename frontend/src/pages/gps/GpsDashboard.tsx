import { useState, useEffect } from 'react'
import { api } from '../../api'

interface SubData {
  id: string
  subscriptionEnd: string
  subscriptionType: string
  activationDate?: string | null
  createdAt?: string | null
  gpsNumber?: string
  customer: { fullName: string; fatherName: string; grandfatherName: string; phone: string; address: string } | null
  simCard?: { simNumber: string } | null
}

interface SimCard {
  id: string
  simNumber: string
  operator: string
  status: string
  currentCustomerName: string | null
}

function formatDate(d: string) {
  if (!d) return '-'
  try { return new Date(d).toLocaleDateString('ar-IQ') } catch { return d }
}

function getRemainingDays(endDate: string): number {
  const end = new Date(endDate)
  const now = new Date()
  return Math.ceil((end.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
}

function getSubscriptionStatus(endDate: string) {
  const days = getRemainingDays(endDate)
  if (days > 30) return 'active'
  if (days >= 0) return 'expiring_soon'
  if (days >= -40) return 'expired_40'
  return 'expired_80'
}

function getStatusLabel(status: string) {
  switch (status) {
    case 'active': return 'نشط'
    case 'expiring_soon': return 'قارب الانتهاء'
    case 'expired_40': return 'منتهي +40'
    case 'expired_80': return 'منتهي +80'
    default: return status
  }
}

function getStatusColor(status: string) {
  switch (status) {
    case 'active': return 'bg-green-50 text-green-700'
    case 'expiring_soon': return 'bg-yellow-50 text-yellow-600'
    case 'expired_40': return 'bg-orange-50 text-orange-600'
    case 'expired_80': return 'bg-red-50 text-red-600'
    default: return 'bg-gray-50 text-gray-600'
  }
}

function getSubscriptionLabel(t: string) {
  switch (t) {
    case 'THREE_MONTHS': case '3_months': return '3 أشهر'
    case 'SIX_MONTHS': case '6_months': return '6 أشهر'
    case 'YEARLY': case 'yearly': return 'سنوي'
    default: return t || '-'
  }
}

export default function GpsDashboard() {
  const [tab, setTab] = useState<'overview' | 'subscriptions' | 'assets'>('overview')
  const [subs, setSubs] = useState<SubData[]>([])
  const [sims, setSims] = useState<SimCard[]>([])
  const [loading, setLoading] = useState(true)
  const [selected, setSelected] = useState<SubData | null>(null)
  const [subFilter, setSubFilter] = useState('all')
  const [totalCustomers, setTotalCustomers] = useState(0)

  useEffect(() => {
    async function load() {
      try {
        const [devices, simsData, customers] = await Promise.all([
          api.getGpsDevices(),
          api.getSimCards(),
          api.getGpsCustomers(),
        ])
        const delivered = (devices || []).filter((d: any) => d.status === 'DELIVERED' && d.subscriptionEnd)
        setSubs(delivered.map((d: any) => ({
          id: d.id,
          subscriptionEnd: d.subscriptionEnd,
          subscriptionType: d.subscriptionType,
          activationDate: d.activationDate,
          createdAt: d.createdAt,
          gpsNumber: d.gpsNumber,
          customer: d.customer ? { fullName: d.customer.fullName, fatherName: d.customer.fatherName, grandfatherName: d.customer.grandfatherName, phone: d.customer.phone, address: d.customer.address } : null,
          simCard: d.simCard,
        })))
        setSims((simsData || []).map((s: any) => ({ id: s.id, simNumber: s.simNumber, operator: s.operator, status: s.status, currentCustomerName: s.currentCustomerName })))
        setTotalCustomers((customers || []).length)
      } catch (e) { console.error(e) }
      setLoading(false)
    }
    load()
  }, [])

  const active = subs.filter(s => getSubscriptionStatus(s.subscriptionEnd) === 'active').length
  const expiring = subs.filter(s => getSubscriptionStatus(s.subscriptionEnd) === 'expiring_soon').length
  const exp40 = subs.filter(s => getSubscriptionStatus(s.subscriptionEnd) === 'expired_40').length
  const exp80 = subs.filter(s => getSubscriptionStatus(s.subscriptionEnd) === 'expired_80').length
  const simsAvailable = sims.filter(s => s.status === 'AVAILABLE' || s.status === 'available').length
  const simsInUse = sims.filter(s => s.status === 'IN_USE' || s.status === 'in_use').length

  const sorted = [...subs].sort((a, b) => getRemainingDays(a.subscriptionEnd) - getRemainingDays(b.subscriptionEnd))
  const filtered = subFilter === 'all' ? sorted : sorted.filter(s => getSubscriptionStatus(s.subscriptionEnd) === subFilter)
  const expiringSoon = sorted.filter(s => { const d = getRemainingDays(s.subscriptionEnd); return d >= 0 && d <= 30 })

  return (
    <div dir="rtl">
      <div className="mb-6">
        <h1 className="text-2xl font-bold" style={{ color: '#1a3a5c' }}>لوحة المراقبة</h1>
        <p className="text-gray-500 text-sm mt-1">نظرة شاملة على جميع العمليات</p>
      </div>

      <div className="flex gap-2 mb-6 overflow-x-auto pb-1">
        {([['overview', '📊 نظرة عامة'], ['subscriptions', '📡 الاشتراكات'], ['assets', '🏢 أصول الشركة']] as const).map(([v, l]) => (
          <button key={v} onClick={() => setTab(v as any)}
            className="px-4 py-2.5 rounded-xl text-sm font-semibold whitespace-nowrap"
            style={{ background: tab === v ? '#1a3a5c' : 'white', color: tab === v ? 'white' : '#64748b', boxShadow: '0 1px 4px rgba(0,0,0,0.08)', flexShrink: 0 }}>
            {l}
          </button>
        ))}
      </div>

      {tab === 'overview' && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
          <div className="grid grid-cols-2 lg:grid-cols-5 gap-4">
            {[
              { label: 'إجمالي الزبائن', value: totalCustomers, color: '#1a3a5c', bg: '#eff6ff', icon: '👥' },
              { label: 'اشتراكات نشطة', value: active, color: '#16a34a', bg: '#f0fdf4', icon: '✅' },
              { label: 'تنتهي قريباً', value: expiring, color: '#d97706', bg: '#fffbeb', icon: '⚠️' },
              { label: 'منتهية +40 يوم', value: exp40, color: '#ea580c', bg: '#fff7ed', icon: '🔔' },
              { label: 'منتهية +80 يوم', value: exp80, color: '#dc2626', bg: '#fef2f2', icon: '🚨' },
            ].map(card => (
              <div key={card.label} className="rounded-2xl p-5" style={{ background: card.bg, border: `1px solid ${card.color}22`, boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
                <div className="flex items-center justify-between mb-3">
                  <span className="text-2xl">{card.icon}</span>
                  <span className="text-3xl font-bold" style={{ color: card.color }}>{loading ? '...' : card.value}</span>
                </div>
                <p className="text-sm font-semibold" style={{ color: card.color }}>{card.label}</p>
              </div>
            ))}
          </div>

          <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
            {[
              { label: 'الطلبات المعلقة', icon: '📋', color: '#7c3aed' },
              { label: 'تجديد اشتراك', icon: '🔄', color: '#2d6a4f' },
              { label: 'إضافة زبون قديم', icon: '📂', color: '#1a3a5c' },
              { label: 'إعدادات الأسعار', icon: '⚙️', color: '#64748b' },
            ].map(a => (
              <button key={a.label} onClick={() => alert(a.label)}
                className="rounded-2xl p-4 text-right flex items-center gap-3"
                style={{ background: 'white', boxShadow: '0 1px 3px rgba(0,0,0,0.05)', border: 'none', cursor: 'pointer' }}>
                <span className="text-2xl">{a.icon}</span>
                <span className="text-sm font-semibold" style={{ color: a.color }}>{a.label}</span>
              </button>
            ))}
          </div>

          {expiringSoon.length > 0 && (
            <div style={{ background: '#fffbeb', border: '1px solid #fde68a', borderRadius: '1rem', padding: '1.25rem' }}>
              <h3 className="font-bold mb-3" style={{ color: '#92400e' }}>⚠️ تنتهي خلال 30 يوم ({expiringSoon.length} زبون)</h3>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                {expiringSoon.slice(0, 5).map(s => {
                  const days = getRemainingDays(s.subscriptionEnd)
                  return (
                    <div key={s.id} onClick={() => { setSelected(s); setTab('subscriptions') }}
                      className="flex items-center justify-between"
                      style={{ background: 'white', borderRadius: '0.75rem', padding: '0.75rem 1rem', cursor: 'pointer', boxShadow: '0 1px 2px rgba(0,0,0,0.05)' }}>
                      <div>
                        <span className="font-semibold text-sm" style={{ color: '#1f2937' }}>
                          {s.customer ? `${s.customer.fullName} ${s.customer.fatherName}` : '-'}
                        </span>
                        <span className="text-xs mr-2" style={{ color: '#9ca3af' }}>{s.customer?.phone}</span>
                      </div>
                      <span className="font-bold text-sm" style={{ color: days <= 7 ? '#dc2626' : '#d97706' }}>
                        {days} يوم
                      </span>
                    </div>
                  )
                })}
              </div>
            </div>
          )}

          <div className="grid grid-cols-3 gap-4">
            <div className="rounded-2xl p-5 text-center" style={{ background: 'white', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
              <div className="text-3xl font-bold" style={{ color: '#1f2937' }}>{sims.length}</div>
              <div className="text-sm mt-1" style={{ color: '#6b7280' }}>إجمالي SIM</div>
            </div>
            <div className="rounded-2xl p-5 text-center" style={{ background: '#f0fdf4', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
              <div className="text-3xl font-bold" style={{ color: '#15803d' }}>{simsAvailable}</div>
              <div className="text-sm mt-1" style={{ color: '#16a34a' }}>SIM متاح</div>
            </div>
            <div className="rounded-2xl p-5 text-center" style={{ background: '#eff6ff', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
              <div className="text-3xl font-bold" style={{ color: '#1d4ed8' }}>{simsInUse}</div>
              <div className="text-sm mt-1" style={{ color: '#2563eb' }}>SIM مستخدم</div>
            </div>
          </div>
        </div>
      )}

      {tab === 'subscriptions' && (
        <div className="flex gap-6">
          <div style={{ flex: selected ? undefined : 1, display: selected ? 'none' : 'block' }} className="lg:!block lg:!flex-1">
            <div className="flex gap-2 mb-4 flex-wrap">
              {[['all','الكل'],['active','نشط'],['expiring_soon','قارب الانتهاء'],['expired_40','منتهي 40'],['expired_80','منتهي 80']].map(([v,l]) => (
                <button key={v} onClick={() => setSubFilter(v)}
                  className="px-3 py-1.5 rounded-full text-xs font-semibold"
                  style={{ background: subFilter === v ? '#1a3a5c' : 'white', color: subFilter === v ? 'white' : '#64748b', boxShadow: '0 1px 4px rgba(0,0,0,0.08)', border: 'none', cursor: 'pointer' }}>
                  {l}
                </button>
              ))}
            </div>

            {loading ? (
              <div className="text-center py-20" style={{ color: '#9ca3af' }}>جارٍ التحميل...</div>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                {filtered.map(sub => {
                  const status = getSubscriptionStatus(sub.subscriptionEnd)
                  const days = getRemainingDays(sub.subscriptionEnd)
                  const isSelected = selected?.id === sub.id
                  const subTypeDays = sub.subscriptionType === 'YEARLY' || sub.subscriptionType === 'yearly' ? 365 : (sub.subscriptionType === 'SIX_MONTHS' || sub.subscriptionType === '6_months') ? 180 : 90
                  return (
                    <div key={sub.id} onClick={() => setSelected(isSelected ? null : sub)}
                      className="rounded-xl px-4 py-3"
                      style={{ background: 'white', boxShadow: '0 1px 3px rgba(0,0,0,0.05)', cursor: 'pointer', borderRight: isSelected ? '3px solid #1a3a5c' : '3px solid transparent' }}>
                      <div className="flex items-center justify-between">
                        <div style={{ flex: 1, minWidth: 0 }}>
                          <div className="flex items-center gap-2 mb-1">
                            <span className="font-semibold text-sm" style={{ color: '#1f2937', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                              {sub.customer ? `${sub.customer.fullName} ${sub.customer.fatherName} ${sub.customer.grandfatherName}` : '-'}
                            </span>
                            <span className={`text-xs px-2 py-0.5 rounded-full ${getStatusColor(status)}`}>{getStatusLabel(status)}</span>
                          </div>
                          <div className="flex gap-3 text-xs" style={{ color: '#9ca3af' }}>
                            <span>📞 {sub.customer?.phone || '-'}</span>
                            {sub.gpsNumber && <span>📦 {sub.gpsNumber}</span>}
                          </div>
                        </div>
                        <div className="text-left mr-3" style={{ flexShrink: 0 }}>
                          <div className="font-bold text-base" style={{ color: days < 0 ? '#dc2626' : days < 30 ? '#d97706' : '#16a34a' }}>
                            {days < 0 ? `${Math.abs(days)}-` : days}
                          </div>
                          <div className="text-xs" style={{ color: '#9ca3af' }}>يوم</div>
                        </div>
                      </div>
                      <div className="mt-2" style={{ height: '4px', background: '#f3f4f6', borderRadius: '9999px', overflow: 'hidden' }}>
                        <div style={{
                          height: '100%', borderRadius: '9999px',
                          width: `${Math.max(0, Math.min(100, (days / subTypeDays) * 100))}%`,
                          background: days < 0 ? '#dc2626' : days < 30 ? '#d97706' : '#16a34a'
                        }} />
                      </div>
                    </div>
                  )
                })}
                {filtered.length === 0 && <div className="text-center py-12" style={{ color: '#9ca3af' }}>لا توجد اشتراكات</div>}
              </div>
            )}
          </div>

          {selected && (
            <div style={{ width: '100%', maxWidth: '20rem', flexShrink: 0 }}>
              <div className="rounded-2xl p-5" style={{ background: 'white', boxShadow: '0 1px 3px rgba(0,0,0,0.05)', position: 'sticky', top: '1rem' }}>
                <div className="flex items-center justify-between mb-4">
                  <h3 className="font-bold" style={{ color: '#1a3a5c' }}>تفاصيل الزبون</h3>
                  <button onClick={() => setSelected(null)} style={{ color: '#9ca3af', fontSize: '1.25rem', background: 'none', border: 'none', cursor: 'pointer' }}>✕</button>
                </div>
                <div className="text-center mb-4 pb-4" style={{ borderBottom: '1px solid #e5e7eb' }}>
                  <div className="mx-auto mb-2 flex items-center justify-center text-2xl font-bold text-white" style={{ width: '4rem', height: '4rem', borderRadius: '9999px', background: '#1a3a5c' }}>
                    {selected.customer?.fullName?.charAt(0) || '?'}
                  </div>
                  <h4 className="font-bold" style={{ color: '#1f2937' }}>
                    {selected.customer ? `${selected.customer.fullName} ${selected.customer.fatherName} ${selected.customer.grandfatherName}` : '-'}
                  </h4>
                </div>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem', fontSize: '0.875rem' }}>
                  {[
                    ['📞 الهاتف', selected.customer?.phone || '-'],
                    ['📍 العنوان', selected.customer?.address || '-'],
                    ['📦 رقم الجهاز', selected.gpsNumber || '-'],
                    ['📋 نوع الاشتراك', getSubscriptionLabel(selected.subscriptionType)],
                    ['📅 تاريخ التفعيل', selected.activationDate ? formatDate(selected.activationDate) : '-'],
                    ['⏰ تاريخ الانتهاء', formatDate(selected.subscriptionEnd)],
                  ].map(([label, value]) => (
                    <div key={label as string} className="flex justify-between items-start gap-2">
                      <span style={{ color: '#6b7280', flexShrink: 0 }}>{label}</span>
                      <span className="font-semibold text-left" style={{ color: '#1f2937' }}>{value}</span>
                    </div>
                  ))}
                </div>
                <div className="mt-4 p-3 rounded-xl text-center" style={{ background: getRemainingDays(selected.subscriptionEnd) < 0 ? '#fef2f2' : getRemainingDays(selected.subscriptionEnd) < 30 ? '#fffbeb' : '#f0fdf4' }}>
                  <div className="text-2xl font-bold" style={{ color: getRemainingDays(selected.subscriptionEnd) < 0 ? '#dc2626' : getRemainingDays(selected.subscriptionEnd) < 30 ? '#d97706' : '#16a34a' }}>
                    {Math.abs(getRemainingDays(selected.subscriptionEnd))} يوم
                  </div>
                  <div className="text-xs mt-1" style={{ color: '#6b7280' }}>
                    {getRemainingDays(selected.subscriptionEnd) < 0 ? 'منذ انتهاء الاشتراك' : 'متبقي على الانتهاء'}
                  </div>
                </div>
                <button onClick={() => alert('تجديد الاشتراك')} className="mt-4 text-sm font-semibold text-white w-full py-2.5 rounded-xl" style={{ background: '#1a3a5c', border: 'none', cursor: 'pointer' }}>
                  🔄 تجديد الاشتراك
                </button>
              </div>
            </div>
          )}
        </div>
      )}

      {tab === 'assets' && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
          <div className="grid grid-cols-3 gap-4">
            <div className="rounded-2xl p-5 text-center" style={{ background: 'white', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
              <div className="text-3xl font-bold" style={{ color: '#1f2937' }}>{sims.length}</div>
              <div className="text-sm mt-1" style={{ color: '#6b7280' }}>إجمالي SIM كارتات</div>
            </div>
            <div className="rounded-2xl p-5 text-center" style={{ background: '#f0fdf4', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
              <div className="text-3xl font-bold" style={{ color: '#15803d' }}>{simsAvailable}</div>
              <div className="text-sm mt-1" style={{ color: '#16a34a' }}>متاح للتوزيع</div>
            </div>
            <div className="rounded-2xl p-5 text-center" style={{ background: '#eff6ff', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
              <div className="text-3xl font-bold" style={{ color: '#1d4ed8' }}>{simsInUse}</div>
              <div className="text-sm mt-1" style={{ color: '#2563eb' }}>مستخدم حالياً</div>
            </div>
          </div>

          <div className="rounded-2xl overflow-hidden" style={{ background: 'white', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
            <div className="flex items-center justify-between p-5" style={{ borderBottom: '1px solid #e5e7eb' }}>
              <h3 className="font-bold" style={{ color: '#1a3a5c' }}>SIM كارتات الشركة</h3>
            </div>
            <div style={{ overflowX: 'auto' }}>
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-right" style={{ background: '#f9fafb', color: '#6b7280' }}>
                    <th className="px-4 py-3 font-semibold">رقم الخط</th>
                    <th className="px-4 py-3 font-semibold">المشغل</th>
                    <th className="px-4 py-3 font-semibold">الحالة</th>
                    <th className="px-4 py-3 font-semibold">الزبون</th>
                  </tr>
                </thead>
                <tbody>
                  {sims.slice(0, 15).map((sim, i) => (
                    <tr key={sim.id} style={{ background: i % 2 === 0 ? 'white' : '#fafafa', borderTop: '1px solid #f3f4f6' }}>
                      <td className="px-4 py-3 font-medium">{sim.simNumber}</td>
                      <td className="px-4 py-3" style={{ color: '#4b5563' }}>{sim.operator}</td>
                      <td className="px-4 py-3">
                        {(sim.status === 'AVAILABLE' || sim.status === 'available')
                          ? <span className="text-xs px-2 py-0.5 rounded-full" style={{ background: '#f0fdf4', color: '#15803d' }}>متاح</span>
                          : <span className="text-xs px-2 py-0.5 rounded-full" style={{ background: '#eff6ff', color: '#1d4ed8' }}>مستخدم</span>}
                      </td>
                      <td className="px-4 py-3" style={{ color: '#6b7280' }}>{sim.currentCustomerName || '-'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {sims.length === 0 && <div className="text-center py-8" style={{ color: '#9ca3af' }}>لا توجد SIM كارتات</div>}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
