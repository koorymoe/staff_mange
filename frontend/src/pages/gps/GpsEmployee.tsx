import { useState, useEffect } from 'react'
import { api } from '../../api'
import { useSession } from '../../session'

type Tab = 'new' | 'delivery' | 'renewal' | 'maintenance'

export default function GpsEmployee() {
  const { employee } = useSession()
  const [tab, setTab] = useState<Tab>('new')
  const [customers, setCustomers] = useState<any[]>([])
  const [devices, setDevices] = useState<any[]>([])
  const [renewals, setRenewals] = useState<any[]>([])
  const [maintenance, setMaintenance] = useState<any[]>([])
  const [loading, setLoading] = useState(true)

  // New device form
  const [customerId, setCustomerId] = useState('')
  const [purchaseType, setPurchaseType] = useState<'DEVICE_SIM' | 'DEVICE_ONLY'>('DEVICE_SIM')
  const [subscriptionType, setSubscriptionType] = useState<'THREE_MONTHS' | 'SIX_MONTHS' | 'YEARLY'>('THREE_MONTHS')
  const [notes, setNotes] = useState('')

  // Maintenance form
  const [maintCustomerId, setMaintCustomerId] = useState('')
  const [problemDescription, setProblemDescription] = useState('')

  // Renewal form
  const [renewCustomerId, setRenewCustomerId] = useState('')
  const [renewDeviceId, setRenewDeviceId] = useState('')
  const [renewSubType, setRenewSubType] = useState<'THREE_MONTHS' | 'SIX_MONTHS' | 'YEARLY'>('THREE_MONTHS')

  const load = async () => {
    setLoading(true)
    try {
      const [c, d, r, m] = await Promise.all([
        api.getGpsCustomers(),
        api.getGpsDevices(),
        api.getGpsRenewals(),
        api.getGpsMaintenance(),
      ])
      setCustomers(c)
      const myId = employee?.id
      setDevices(d.filter((x: any) => x.employeeId === myId))
      setRenewals(r.filter((x: any) => x.employeeId === myId))
      setMaintenance(m.filter((x: any) => x.employeeId === myId))
    } catch (e: any) {
      alert(e.message)
    }
    setLoading(false)
  }

  useEffect(() => { load() }, [])

  const handleNewDevice = async () => {
    if (!customerId || !employee) return alert('اختر الزبون')
    try {
      await api.createGpsDevice({ customerId, employeeId: employee.id, purchaseType, subscriptionType, notes } as any)
      setCustomerId(''); setNotes('')
      load()
    } catch (e: any) { alert(e.message) }
  }

  const handleNewMaintenance = async () => {
    if (!maintCustomerId || !employee) return alert('اختر الزبون')
    if (!problemDescription) return alert('اكتب وصف المشكلة')
    try {
      await api.createGpsMaintenance({ customerId: maintCustomerId, employeeId: employee.id, problemDescription } as any)
      setMaintCustomerId(''); setProblemDescription('')
      load()
    } catch (e: any) { alert(e.message) }
  }

  const handleNewRenewal = async () => {
    if (!renewCustomerId || !renewDeviceId || !employee) return alert('اختر الزبون والجهاز')
    try {
      await api.createGpsRenewal({ customerId: renewCustomerId, deviceRequestId: renewDeviceId, employeeId: employee.id, subscriptionType: renewSubType } as any)
      setRenewCustomerId(''); setRenewDeviceId('')
      load()
    } catch (e: any) { alert(e.message) }
  }

  const subTypeLabel: Record<string, string> = {
    THREE_MONTHS: '3 أشهر',
    SIX_MONTHS: '6 أشهر',
    YEARLY: 'سنوي',
  }

  const statusLabel: Record<string, string> = {
    PENDING: 'قيد الانتظار',
    APPROVED: 'موافق عليه',
    REJECTED: 'مرفوض',
    DELIVERED: 'تم التسليم',
    IN_PROGRESS: 'قيد المعالجة',
    COMPLETED: 'مكتمل',
  }

  const tabs: { key: Tab; label: string }[] = [
    { key: 'new', label: 'طلب جديد' },
    { key: 'delivery', label: 'تسليم' },
    { key: 'renewal', label: 'تجديد' },
    { key: 'maintenance', label: 'صيانة' },
  ]

  if (loading) return <div className="p-8 text-center text-lg">جاري التحميل...</div>

  return (
    <div>
      <h1 className="mb-6 text-2xl font-bold text-brand-800">لوحة موظف GPS</h1>

      <div className="mb-6 flex gap-2">
        {tabs.map(t => (
          <button
            key={t.key}
            onClick={() => setTab(t.key)}
            className={`rounded-lg px-6 py-3 text-sm font-bold transition-all ${
              tab === t.key ? 'bg-brand-600 text-white shadow-lg' : 'bg-white text-brand-700 border border-brand-200 hover:bg-brand-50'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {tab === 'new' && (
        <div className="rounded-xl bg-white p-6 shadow-md">
          <h2 className="mb-4 text-lg font-bold text-brand-700">طلب جهاز جديد</h2>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">الزبون</label>
              <select value={customerId} onChange={e => setCustomerId(e.target.value)} className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right">
                <option value="">اختر الزبون</option>
                {customers.map(c => <option key={c.id} value={c.id}>{c.fullName} - {c.phone}</option>)}
              </select>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">نوع الشراء</label>
              <select value={purchaseType} onChange={e => setPurchaseType(e.target.value as any)} className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right">
                <option value="DEVICE_SIM">جهاز + شريحة</option>
                <option value="DEVICE_ONLY">جهاز فقط</option>
              </select>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">نوع الاشتراك</label>
              <select value={subscriptionType} onChange={e => setSubscriptionType(e.target.value as any)} className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right">
                <option value="THREE_MONTHS">3 أشهر</option>
                <option value="SIX_MONTHS">6 أشهر</option>
                <option value="YEARLY">سنوي</option>
              </select>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">ملاحظات</label>
              <input value={notes} onChange={e => setNotes(e.target.value)} className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right" />
            </div>
          </div>
          <button onClick={handleNewDevice} className="mt-4 rounded-lg bg-brand-600 px-8 py-3 text-white font-bold hover:bg-brand-700 transition-colors">
            إرسال الطلب
          </button>

          <h3 className="mt-8 mb-3 text-md font-bold text-gray-700">طلباتي</h3>
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-right text-xs font-bold text-gray-600">الزبون</th>
                  <th className="px-4 py-3 text-right text-xs font-bold text-gray-600">النوع</th>
                  <th className="px-4 py-3 text-right text-xs font-bold text-gray-600">الاشتراك</th>
                  <th className="px-4 py-3 text-right text-xs font-bold text-gray-600">الحالة</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {devices.map((d: any) => (
                  <tr key={d.id}>
                    <td className="px-4 py-3 text-sm">{d.customer?.fullName}</td>
                    <td className="px-4 py-3 text-sm">{d.purchaseType === 'DEVICE_SIM' ? 'جهاز + شريحة' : 'جهاز فقط'}</td>
                    <td className="px-4 py-3 text-sm">{subTypeLabel[d.subscriptionType]}</td>
                    <td className="px-4 py-3 text-sm">
                      <span className={`rounded-full px-3 py-1 text-xs font-bold ${
                        d.status === 'PENDING' ? 'bg-yellow-100 text-yellow-800' :
                        d.status === 'APPROVED' ? 'bg-blue-100 text-blue-800' :
                        d.status === 'DELIVERED' ? 'bg-green-100 text-green-800' :
                        'bg-red-100 text-red-800'
                      }`}>{statusLabel[d.status]}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {tab === 'delivery' && (
        <div className="rounded-xl bg-white p-6 shadow-md">
          <h2 className="mb-4 text-lg font-bold text-brand-700">طلبات التسليم</h2>
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-right text-xs font-bold text-gray-600">الزبون</th>
                  <th className="px-4 py-3 text-right text-xs font-bold text-gray-600">النوع</th>
                  <th className="px-4 py-3 text-right text-xs font-bold text-gray-600">الحالة</th>
                  <th className="px-4 py-3 text-right text-xs font-bold text-gray-600">تم التفعيل</th>
                  <th className="px-4 py-3 text-right text-xs font-bold text-gray-600">تم التسليم</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {devices.filter((d: any) => d.status === 'APPROVED' || d.status === 'DELIVERED').map((d: any) => (
                  <tr key={d.id}>
                    <td className="px-4 py-3 text-sm">{d.customer?.fullName}</td>
                    <td className="px-4 py-3 text-sm">{d.purchaseType === 'DEVICE_SIM' ? 'جهاز + شريحة' : 'جهاز فقط'}</td>
                    <td className="px-4 py-3 text-sm">{statusLabel[d.status]}</td>
                    <td className="px-4 py-3 text-sm">{d.isActivated ? 'نعم' : 'لا'}</td>
                    <td className="px-4 py-3 text-sm">{d.isDelivered ? 'نعم' : 'لا'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {tab === 'renewal' && (
        <div className="rounded-xl bg-white p-6 shadow-md">
          <h2 className="mb-4 text-lg font-bold text-brand-700">طلب تجديد</h2>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-3 mb-4">
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">الزبون</label>
              <select value={renewCustomerId} onChange={e => setRenewCustomerId(e.target.value)} className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right">
                <option value="">اختر الزبون</option>
                {customers.map(c => <option key={c.id} value={c.id}>{c.fullName}</option>)}
              </select>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">الجهاز</label>
              <select value={renewDeviceId} onChange={e => setRenewDeviceId(e.target.value)} className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right">
                <option value="">اختر الجهاز</option>
                {devices.filter((d: any) => d.customerId === renewCustomerId && d.status === 'DELIVERED').map((d: any) => (
                  <option key={d.id} value={d.id}>جهاز #{d.id.slice(-6)}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">نوع الاشتراك</label>
              <select value={renewSubType} onChange={e => setRenewSubType(e.target.value as any)} className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right">
                <option value="THREE_MONTHS">3 أشهر</option>
                <option value="SIX_MONTHS">6 أشهر</option>
                <option value="YEARLY">سنوي</option>
              </select>
            </div>
          </div>
          <button onClick={handleNewRenewal} className="rounded-lg bg-brand-600 px-8 py-3 text-white font-bold hover:bg-brand-700 transition-colors">
            إرسال طلب التجديد
          </button>

          <h3 className="mt-8 mb-3 text-md font-bold text-gray-700">طلبات التجديد السابقة</h3>
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-right text-xs font-bold text-gray-600">الزبون</th>
                  <th className="px-4 py-3 text-right text-xs font-bold text-gray-600">نوع الاشتراك</th>
                  <th className="px-4 py-3 text-right text-xs font-bold text-gray-600">الحالة</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {renewals.map((r: any) => (
                  <tr key={r.id}>
                    <td className="px-4 py-3 text-sm">{r.customer?.fullName}</td>
                    <td className="px-4 py-3 text-sm">{subTypeLabel[r.subscriptionType]}</td>
                    <td className="px-4 py-3 text-sm">
                      <span className={`rounded-full px-3 py-1 text-xs font-bold ${
                        r.status === 'PENDING' ? 'bg-yellow-100 text-yellow-800' :
                        r.status === 'APPROVED' ? 'bg-green-100 text-green-800' :
                        'bg-red-100 text-red-800'
                      }`}>{statusLabel[r.status]}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {tab === 'maintenance' && (
        <div className="rounded-xl bg-white p-6 shadow-md">
          <h2 className="mb-4 text-lg font-bold text-brand-700">طلب صيانة</h2>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 mb-4">
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">الزبون</label>
              <select value={maintCustomerId} onChange={e => setMaintCustomerId(e.target.value)} className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right">
                <option value="">اختر الزبون</option>
                {customers.map(c => <option key={c.id} value={c.id}>{c.fullName}</option>)}
              </select>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">وصف المشكلة</label>
              <textarea value={problemDescription} onChange={e => setProblemDescription(e.target.value)} className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right" rows={3} />
            </div>
          </div>
          <button onClick={handleNewMaintenance} className="rounded-lg bg-brand-600 px-8 py-3 text-white font-bold hover:bg-brand-700 transition-colors">
            إرسال طلب الصيانة
          </button>

          <h3 className="mt-8 mb-3 text-md font-bold text-gray-700">طلبات الصيانة السابقة</h3>
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-right text-xs font-bold text-gray-600">الزبون</th>
                  <th className="px-4 py-3 text-right text-xs font-bold text-gray-600">المشكلة</th>
                  <th className="px-4 py-3 text-right text-xs font-bold text-gray-600">الحالة</th>
                  <th className="px-4 py-3 text-right text-xs font-bold text-gray-600">ملاحظات الإدارة</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {maintenance.map((m: any) => (
                  <tr key={m.id}>
                    <td className="px-4 py-3 text-sm">{m.customer?.fullName}</td>
                    <td className="px-4 py-3 text-sm">{m.problemDescription}</td>
                    <td className="px-4 py-3 text-sm">
                      <span className={`rounded-full px-3 py-1 text-xs font-bold ${
                        m.status === 'PENDING' ? 'bg-yellow-100 text-yellow-800' :
                        m.status === 'IN_PROGRESS' ? 'bg-blue-100 text-blue-800' :
                        'bg-green-100 text-green-800'
                      }`}>{statusLabel[m.status]}</span>
                    </td>
                    <td className="px-4 py-3 text-sm">{m.adminNotes || '-'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}
