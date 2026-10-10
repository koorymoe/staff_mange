import { useState, useEffect } from 'react'
import { api, type GpsCustomer, type GpsDeviceRequest, type GpsRenewalRequest, type GpsMaintenanceRequest } from '../../api'
import { useSession } from '../../session'

type ActiveForm = null | 'purchase' | 'renewal' | 'maintenance'
type RecentRequest =
  | (GpsDeviceRequest & { type: 'device' })
  | (GpsRenewalRequest & { type: 'renewal' })
  | (GpsMaintenanceRequest & { type: 'maintenance' })

export default function GpsEmployee() {
  const { employee } = useSession()
  const [customers, setCustomers] = useState<GpsCustomer[]>([])
  const [devices, setDevices] = useState<GpsDeviceRequest[]>([])
  const [recentRequests, setRecentRequests] = useState<RecentRequest[]>([])
  const [assignedInstalls, setAssignedInstalls] = useState<GpsDeviceRequest[]>([])
  const [loading, setLoading] = useState(true)
  const [activeForm, setActiveForm] = useState<ActiveForm>(null)
  const [submitting, setSubmitting] = useState(false)

  // Purchase form
  const [customerId, setCustomerId] = useState('')
  const [purchaseType, setPurchaseType] = useState<'DEVICE_SIM' | 'DEVICE_ONLY'>('DEVICE_SIM')
  const [subscriptionType, setSubscriptionType] = useState<'THREE_MONTHS' | 'SIX_MONTHS' | 'YEARLY'>('THREE_MONTHS')
  const [notes, setNotes] = useState('')

  // Renewal form
  const [renewCustomerId, setRenewCustomerId] = useState('')
  const [renewDeviceId, setRenewDeviceId] = useState('')
  const [renewSubType, setRenewSubType] = useState('THREE_MONTHS')

  // Maintenance form
  const [maintCustomerId, setMaintCustomerId] = useState('')
  const [problemDescription, setProblemDescription] = useState('')

  const load = async () => {
    try {
      const [c, d, r, m] = await Promise.all([
        api.getGpsCustomers(),
        api.getGpsDevices(),
        api.getGpsRenewals(),
        api.getGpsMaintenance(),
      ])
      setCustomers(c)
      setDevices(d)
      const myId = employee?.id
      setAssignedInstalls(
        d.filter((x) => x.assignedTechnician?.id === myId && x.status !== 'REJECTED')
          .sort((a, b) => new Date(a.scheduledAt || a.createdAt).getTime() - new Date(b.scheduledAt || b.createdAt).getTime())
      )
      const allRequests: RecentRequest[] = [
        ...d.filter((x) => x.employeeId === myId).map((x): RecentRequest => ({ ...x, type: 'device' })),
        ...r.filter((x) => x.employeeId === myId).map((x): RecentRequest => ({ ...x, type: 'renewal' })),
        ...m.filter((x) => x.employeeId === myId).map((x): RecentRequest => ({ ...x, type: 'maintenance' })),
      ].sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()).slice(0, 10)
      setRecentRequests(allRequests)
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    }
    setLoading(false)
  }

  // `load` is async and only sets state after its awaits resolve (false positive for
  // set-state-in-effect); `load` is redefined every render (not memoized), so adding
  // it as a dep would re-run this effect every render — we intentionally load once.
  // eslint-disable-next-line react-hooks/set-state-in-effect, react-hooks/exhaustive-deps
  useEffect(() => { load() }, [])

  const handlePurchase = async () => {
    if (!customerId || !employee) return alert('اختر الزبون')
    setSubmitting(true)
    try {
      await api.createGpsDevice({ customerId, employeeId: employee.id, purchaseType, subscriptionType, notes })
      setCustomerId(''); setNotes(''); setActiveForm(null)
      load()
    } catch (e) { alert(e instanceof Error ? e.message : 'حدث خطأ') }
    setSubmitting(false)
  }

  const handleRenewal = async () => {
    if (!renewCustomerId || !renewDeviceId || !employee) return alert('اختر الزبون والجهاز')
    setSubmitting(true)
    try {
      await api.createGpsRenewal({ customerId: renewCustomerId, deviceRequestId: renewDeviceId, employeeId: employee.id, subscriptionType: renewSubType })
      setRenewCustomerId(''); setRenewDeviceId(''); setActiveForm(null)
      load()
    } catch (e) { alert(e instanceof Error ? e.message : 'حدث خطأ') }
    setSubmitting(false)
  }

  const handleMaintenance = async () => {
    if (!maintCustomerId || !employee) return alert('اختر الزبون')
    if (!problemDescription) return alert('اكتب وصف المشكلة')
    setSubmitting(true)
    try {
      await api.createGpsMaintenance({ customerId: maintCustomerId, employeeId: employee.id, problemDescription })
      setMaintCustomerId(''); setProblemDescription(''); setActiveForm(null)
      load()
    } catch (e) { alert(e instanceof Error ? e.message : 'حدث خطأ') }
    setSubmitting(false)
  }

  const statusLabel: Record<string, string> = {
    PENDING: 'قيد الانتظار', APPROVED: 'موافق عليه', REJECTED: 'مرفوض',
    DELIVERED: 'تم التسليم', IN_PROGRESS: 'قيد المعالجة', COMPLETED: 'مكتمل',
  }
  const statusColor: Record<string, { color: string; bg: string }> = {
    PENDING: { color: 'var(--t-warning)', bg: '#fffbeb' },
    APPROVED: { color: 'var(--t-info)', bg: '#eff6ff' },
    REJECTED: { color: 'var(--t-danger)', bg: '#fef2f2' },
    DELIVERED: { color: 'var(--t-success)', bg: '#f0fdf4' },
    IN_PROGRESS: { color: 'var(--t-violet)', bg: '#f5f3ff' },
    COMPLETED: { color: 'var(--t-success)', bg: '#f0fdf4' },
  }
  const typeLabel: Record<string, string> = { device: 'شراء جهاز', renewal: 'تجديد', maintenance: 'صيانة' }
  const subLabel: Record<string, string> = { THREE_MONTHS: '3 أشهر', SIX_MONTHS: '6 أشهر', YEARLY: 'سنوي' }

  if (loading) return <div className="flex items-center justify-center p-12"><div className="text-lg text-slate-500">جاري التحميل...</div></div>

  return (
    <div dir="rtl">
      {/* Header */}
      <div className="mb-6 rounded-2xl p-6 shadow-sm" style={{ backgroundColor: '#1a3a5c' }}>
        <h1 className="text-2xl font-bold text-white">👋 مرحباً {employee?.name || ''}</h1>
        <p className="mt-1 text-blue-200 text-sm">لوحة موظف GPS - اختر الخدمة المطلوبة</p>
      </div>

      {/* Assigned Installs */}
      {assignedInstalls.length > 0 && (
        <div className="mb-6 rounded-2xl bg-white p-6 shadow-sm">
          <h3 className="mb-4 text-lg font-bold" style={{ color: 'var(--t-title)' }}>🛠️ تركيباتي المجدولة</h3>
          <div className="flex flex-col gap-3">
            {assignedInstalls.map(d => (
              <div key={d.id} className="flex items-center justify-between rounded-xl border border-slate-100 p-4">
                <div>
                  <p className="font-bold text-slate-800">{d.customer?.fullName}</p>
                  <p className="text-sm text-slate-500">📞 {d.customer?.phone} · 📍 {d.customer?.address}</p>
                </div>
                <div className="text-left">
                  <span className="rounded-full px-3 py-1 text-xs font-bold" style={{ backgroundColor: statusColor[d.status]?.bg, color: statusColor[d.status]?.color }}>
                    {statusLabel[d.status] || d.status}
                  </span>
                  <p className="mt-1 text-sm font-bold text-amber-700">
                    {d.scheduledAt ? new Date(d.scheduledAt).toLocaleString('ar-IQ') : 'بدون موعد محدد'}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Service Cards */}
      <div className="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-3">
        {[
          { key: 'purchase' as const, icon: '📱', title: 'شراء جهاز', desc: 'طلب جهاز تتبع جديد للزبون', color: 'var(--t-title)' },
          { key: 'renewal' as const, icon: '🔄', title: 'تجديد اشتراك', desc: 'تجديد اشتراك زبون حالي', color: 'var(--t-success)' },
          { key: 'maintenance' as const, icon: '🔧', title: 'طلب صيانة', desc: 'إرسال طلب صيانة لجهاز', color: 'var(--t-warning)' },
        ].map(card => (
          <button
            key={card.key}
            onClick={() => setActiveForm(activeForm === card.key ? null : card.key)}
            className={`rounded-2xl bg-white p-6 shadow-sm text-right transition-all hover:shadow-md ${activeForm === card.key ? 'ring-2' : ''}`}
            style={activeForm === card.key ? { borderColor: card.color } : {}}
          >
            <span className="text-4xl">{card.icon}</span>
            <h3 className="mt-3 text-lg font-bold" style={{ color: card.color }}>{card.title}</h3>
            <p className="mt-1 text-sm text-slate-500">{card.desc}</p>
          </button>
        ))}
      </div>

      {/* Purchase Form */}
      {activeForm === 'purchase' && (
        <div className="mb-6 rounded-2xl bg-white p-6 shadow-sm">
          <h3 className="mb-4 text-lg font-bold" style={{ color: 'var(--t-title)' }}>📱 طلب شراء جهاز</h3>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">الزبون *</label>
              <select value={customerId} onChange={e => setCustomerId(e.target.value)} className="w-full rounded-xl border border-slate-200 px-4 py-3 outline-none">
                <option value="">اختر الزبون</option>
                {customers.map(c => <option key={c.id} value={c.id}>{c.fullName} - {c.phone}</option>)}
              </select>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">نوع الشراء</label>
              <select value={purchaseType} onChange={e => setPurchaseType(e.target.value as 'DEVICE_SIM' | 'DEVICE_ONLY')} className="w-full rounded-xl border border-slate-200 px-4 py-3 outline-none">
                <option value="DEVICE_SIM">جهاز + شريحة</option>
                <option value="DEVICE_ONLY">جهاز فقط</option>
              </select>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">نوع الاشتراك</label>
              <select value={subscriptionType} onChange={e => setSubscriptionType(e.target.value as 'THREE_MONTHS' | 'SIX_MONTHS' | 'YEARLY')} className="w-full rounded-xl border border-slate-200 px-4 py-3 outline-none">
                <option value="THREE_MONTHS">3 أشهر</option>
                <option value="SIX_MONTHS">6 أشهر</option>
                <option value="YEARLY">سنوي</option>
              </select>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">ملاحظات</label>
              <input value={notes} onChange={e => setNotes(e.target.value)} className="w-full rounded-xl border border-slate-200 px-4 py-3 outline-none" />
            </div>
          </div>
          <button onClick={handlePurchase} disabled={submitting} className="mt-4 rounded-2xl px-8 py-3 text-sm font-bold text-white shadow-sm disabled:opacity-50" style={{ backgroundColor: '#1a3a5c' }}>
            {submitting ? 'جاري الإرسال...' : '📤 إرسال الطلب'}
          </button>
        </div>
      )}

      {/* Renewal Form */}
      {activeForm === 'renewal' && (
        <div className="mb-6 rounded-2xl bg-white p-6 shadow-sm">
          <h3 className="mb-4 text-lg font-bold text-green-700">🔄 طلب تجديد اشتراك</h3>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">الزبون *</label>
              <select value={renewCustomerId} onChange={e => setRenewCustomerId(e.target.value)} className="w-full rounded-xl border border-slate-200 px-4 py-3 outline-none">
                <option value="">اختر الزبون</option>
                {customers.map(c => <option key={c.id} value={c.id}>{c.fullName}</option>)}
              </select>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">الجهاز *</label>
              <select value={renewDeviceId} onChange={e => setRenewDeviceId(e.target.value)} className="w-full rounded-xl border border-slate-200 px-4 py-3 outline-none">
                <option value="">اختر الجهاز</option>
                {devices.filter(d => d.customerId === renewCustomerId && d.status === 'DELIVERED').map(d => (
                  <option key={d.id} value={d.id}>جهاز #{d.id.slice(-6)}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">نوع الاشتراك</label>
              <select value={renewSubType} onChange={e => setRenewSubType(e.target.value)} className="w-full rounded-xl border border-slate-200 px-4 py-3 outline-none">
                <option value="THREE_MONTHS">3 أشهر</option>
                <option value="SIX_MONTHS">6 أشهر</option>
                <option value="YEARLY">سنوي</option>
              </select>
            </div>
          </div>
          <button onClick={handleRenewal} disabled={submitting} className="mt-4 rounded-2xl bg-green-700 px-8 py-3 text-sm font-bold text-white shadow-sm disabled:opacity-50">
            {submitting ? 'جاري الإرسال...' : '📤 إرسال طلب التجديد'}
          </button>
        </div>
      )}

      {/* Maintenance Form */}
      {activeForm === 'maintenance' && (
        <div className="mb-6 rounded-2xl bg-white p-6 shadow-sm">
          <h3 className="mb-4 text-lg font-bold text-yellow-700">🔧 طلب صيانة</h3>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">الزبون *</label>
              <select value={maintCustomerId} onChange={e => setMaintCustomerId(e.target.value)} className="w-full rounded-xl border border-slate-200 px-4 py-3 outline-none">
                <option value="">اختر الزبون</option>
                {customers.map(c => <option key={c.id} value={c.id}>{c.fullName}</option>)}
              </select>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">وصف المشكلة *</label>
              <textarea value={problemDescription} onChange={e => setProblemDescription(e.target.value)} rows={3} className="w-full rounded-xl border border-slate-200 px-4 py-3 outline-none" placeholder="اكتب وصف المشكلة بالتفصيل..." />
            </div>
          </div>
          <button onClick={handleMaintenance} disabled={submitting} className="mt-4 rounded-2xl bg-yellow-600 px-8 py-3 text-sm font-bold text-white shadow-sm disabled:opacity-50">
            {submitting ? 'جاري الإرسال...' : '📤 إرسال طلب الصيانة'}
          </button>
        </div>
      )}

      {/* Recent Requests Table */}
      <div className="rounded-2xl bg-white shadow-sm overflow-hidden">
        <div className="p-5 border-b border-slate-100">
          <h3 className="text-lg font-bold" style={{ color: 'var(--t-title)' }}>📋 آخر الطلبات</h3>
        </div>
        <div className="overflow-x-auto">
        <table className="w-full text-right">
          <thead>
            <tr style={{ backgroundColor: '#1a3a5c' }}>
              <th className="px-4 py-3 text-sm font-semibold text-white">النوع</th>
              <th className="px-4 py-3 text-sm font-semibold text-white">الزبون</th>
              <th className="px-4 py-3 text-sm font-semibold text-white">التفاصيل</th>
              <th className="px-4 py-3 text-sm font-semibold text-white">الحالة</th>
              <th className="px-4 py-3 text-sm font-semibold text-white">التاريخ</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {recentRequests.map((r) => {
              const sc = statusColor[r.status] || { color: 'var(--t-muted)', bg: '#f8fafc' }
              return (
                <tr key={r.id} className="hover:bg-slate-50">
                  <td className="px-4 py-3 text-sm font-medium">{typeLabel[r.type]}</td>
                  <td className="px-4 py-3 text-sm">{r.customer?.fullName || '-'}</td>
                  <td className="px-4 py-3 text-sm text-slate-500">
                    {r.type === 'device' ? (r.purchaseType === 'DEVICE_SIM' ? 'جهاز + شريحة' : 'جهاز فقط') :
                     r.type === 'renewal' ? subLabel[r.subscriptionType] || '-' :
                     (r.problemDescription || '-').slice(0, 40)}
                  </td>
                  <td className="px-4 py-3">
                    <span className="rounded-full px-3 py-1 text-xs font-bold" style={{ backgroundColor: sc.bg, color: sc.color }}>
                      {statusLabel[r.status] || r.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-slate-500">{r.createdAt ? new Date(r.createdAt).toLocaleDateString('ar-IQ') : '-'}</td>
                </tr>
              )
            })}
            {recentRequests.length === 0 && (
              <tr><td colSpan={5} className="p-8 text-center text-slate-400">لا يوجد طلبات بعد</td></tr>
            )}
          </tbody>
        </table>
        </div>
      </div>
    </div>
  )
}
