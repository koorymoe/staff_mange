import { useState, useEffect } from 'react'
import { api } from '../../api'

const statusConfig: Record<string, { label: string; color: string; bg: string }> = {
  PENDING: { label: 'قيد الانتظار', color: '#d97706', bg: '#fffbeb' },
  APPROVED: { label: 'موافق عليه', color: '#2563eb', bg: '#eff6ff' },
  REJECTED: { label: 'مرفوض', color: '#dc2626', bg: '#fef2f2' },
  DELIVERED: { label: 'تم التسليم', color: '#16a34a', bg: '#f0fdf4' },
  CHECKED: { label: 'تم الفحص', color: '#7c3aed', bg: '#f5f3ff' },
  ACTIVATED: { label: 'مفعّل', color: '#0891b2', bg: '#ecfeff' },
}

export default function GpsDevices() {
  const [devices, setDevices] = useState<any[]>([])
  const [customers, setCustomers] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const [activationDate, setActivationDate] = useState('')

  const load = () => {
    setLoading(true)
    Promise.all([api.getGpsDevices(), api.getGpsCustomers()])
      .then(([d, c]) => { setDevices(d); setCustomers(c) })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const handleAction = async (id: string, action: string, extra?: any) => {
    try {
      await api.updateGpsDevice(id, { action, ...extra })
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    }
  }

  const purchaseLabel: Record<string, string> = { DEVICE_SIM: 'جهاز + شريحة', DEVICE_ONLY: 'جهاز فقط' }
  const subLabel: Record<string, string> = { THREE_MONTHS: '3 أشهر', SIX_MONTHS: '6 أشهر', YEARLY: 'سنوي' }

  if (loading) return <div className="flex items-center justify-center p-12"><div className="text-lg text-slate-500">جاري التحميل...</div></div>
  if (error) return <div className="rounded-2xl bg-red-50 p-6 text-red-600 shadow-sm">{error}</div>

  return (
    <div dir="rtl">
      {/* Header */}
      <div className="mb-6 rounded-2xl p-6 shadow-sm" style={{ backgroundColor: '#1a3a5c' }}>
        <h1 className="text-2xl font-bold text-white">📱 طلبات الأجهزة</h1>
        <p className="mt-1 text-blue-200 text-sm">إدارة طلبات أجهزة التتبع</p>
      </div>

      {/* Device Cards */}
      <div className="space-y-4">
        {devices.map((d) => {
          const sc = statusConfig[d.status] || { label: d.status, color: '#64748b', bg: '#f8fafc' }
          const isExpanded = expandedId === d.id
          return (
            <div key={d.id} className="rounded-2xl bg-white shadow-sm overflow-hidden">
              <div
                className="flex items-center justify-between p-5 cursor-pointer hover:bg-slate-50"
                onClick={() => setExpandedId(isExpanded ? null : d.id)}
              >
                <div className="flex items-center gap-4">
                  <span className="text-2xl">📱</span>
                  <div>
                    <h3 className="font-bold" style={{ color: '#1a3a5c' }}>{d.customer?.fullName || 'غير معروف'}</h3>
                    <p className="text-sm text-slate-500">
                      {purchaseLabel[d.purchaseType] || d.purchaseType} | {subLabel[d.subscriptionType] || d.subscriptionType || '-'}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-3">
                  <span
                    className="rounded-full px-4 py-1 text-xs font-bold"
                    style={{ backgroundColor: sc.bg, color: sc.color }}
                  >
                    {sc.label}
                  </span>
                  <span className="text-slate-400">{isExpanded ? '▲' : '▼'}</span>
                </div>
              </div>

              {isExpanded && (
                <div className="border-t border-slate-100 p-5">
                  <div className="grid grid-cols-2 gap-4 sm:grid-cols-4 mb-4">
                    <div>
                      <span className="text-xs text-slate-500">الموظف</span>
                      <p className="font-medium">{d.employee?.name || '-'}</p>
                    </div>
                    <div>
                      <span className="text-xs text-slate-500">الهاتف</span>
                      <p className="font-medium">{d.customer?.phone || '-'}</p>
                    </div>
                    <div>
                      <span className="text-xs text-slate-500">تاريخ الطلب</span>
                      <p className="font-medium">{d.createdAt ? new Date(d.createdAt).toLocaleDateString('ar-IQ') : '-'}</p>
                    </div>
                    <div>
                      <span className="text-xs text-slate-500">تاريخ التفعيل</span>
                      <p className="font-medium">{d.activationDate ? new Date(d.activationDate).toLocaleDateString('ar-IQ') : '-'}</p>
                    </div>
                  </div>
                  {d.notes && <p className="mb-4 text-sm text-slate-600">📝 {d.notes}</p>}

                  <div className="flex flex-wrap gap-2">
                    {d.status === 'PENDING' && (
                      <>
                        <div className="flex items-center gap-2">
                          <input
                            type="date"
                            value={activationDate}
                            onChange={(e) => setActivationDate(e.target.value)}
                            className="rounded-xl border border-slate-200 px-3 py-2 text-sm"
                          />
                          <button
                            onClick={() => handleAction(d.id, 'approve', { activationDate: activationDate || undefined })}
                            className="rounded-2xl px-5 py-2 text-sm font-bold text-white"
                            style={{ backgroundColor: '#2563eb' }}
                          >
                            ✅ موافقة
                          </button>
                        </div>
                        <button
                          onClick={() => handleAction(d.id, 'reject')}
                          className="rounded-2xl bg-red-600 px-5 py-2 text-sm font-bold text-white"
                        >
                          ❌ رفض
                        </button>
                      </>
                    )}
                    {d.status === 'APPROVED' && (
                      <button onClick={() => handleAction(d.id, 'check')} className="rounded-2xl px-5 py-2 text-sm font-bold text-white" style={{ backgroundColor: '#7c3aed' }}>
                        🔍 فحص
                      </button>
                    )}
                    {d.status === 'CHECKED' && (
                      <button onClick={() => handleAction(d.id, 'activate')} className="rounded-2xl px-5 py-2 text-sm font-bold text-white" style={{ backgroundColor: '#0891b2' }}>
                        ⚡ تفعيل
                      </button>
                    )}
                    {d.status === 'ACTIVATED' && (
                      <button onClick={() => handleAction(d.id, 'deliver')} className="rounded-2xl px-5 py-2 text-sm font-bold text-white" style={{ backgroundColor: '#16a34a' }}>
                        🚚 تسليم
                      </button>
                    )}
                  </div>
                </div>
              )}
            </div>
          )
        })}
        {devices.length === 0 && (
          <div className="rounded-2xl bg-white p-8 text-center text-slate-400 shadow-sm">لا يوجد طلبات أجهزة</div>
        )}
      </div>
    </div>
  )
}
