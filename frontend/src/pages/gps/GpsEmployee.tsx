import { useState, useEffect } from 'react'
import { api } from '../../api'

function formatDate(d: string) {
  if (!d) return '-'
  try { return new Date(d).toLocaleDateString('ar-IQ') } catch { return d }
}

export default function GpsEmployee() {
  const [myRequests, setMyRequests] = useState<any[]>([])
  const [loadingReqs, setLoadingReqs] = useState(true)

  useEffect(() => {
    async function load() {
      try {
        const data = await api.getGpsDevices()
        setMyRequests((data || []).slice(0, 10))
      } catch (e) { console.error(e) }
      setLoadingReqs(false)
    }
    load()
  }, [])

  const services = [
    {
      title: 'شراء جهاز',
      description: 'تسجيل طلب شراء جهاز GPS جديد',
      icon: '🛒',
      color: '#1a3a5c',
    },
    {
      title: 'تجديد اشتراك',
      description: 'إرسال طلب تجديد اشتراك لزبون موجود',
      icon: '🔄',
      color: '#2d6a4f',
    },
    {
      title: 'صيانة',
      description: 'تسجيل طلب صيانة جهاز',
      icon: '🔧',
      color: '#7b3f00',
    },
  ]

  function getStatusBadge(status: string) {
    if (status === 'DELIVERED' || status === 'delivered')
      return <span className="text-xs px-2 py-0.5 rounded-full" style={{ background: '#f0fdf4', color: '#16a34a' }}>تم التسليم ✅</span>
    return <span className="text-xs px-2 py-0.5 rounded-full" style={{ background: '#fffbeb', color: '#d97706' }}>في الانتظار</span>
  }

  const typeLabel = (t: string) => {
    if (t === 'DEVICE_SIM' || t === 'device_sim') return 'جهاز + SIM'
    if (t === 'DEVICE_ONLY' || t === 'device_only') return 'جهاز فقط'
    return t
  }

  return (
    <div dir="rtl">
      <div className="mb-8">
        <h1 className="text-2xl font-bold" style={{ color: '#1a3a5c' }}>مرحباً بك في نظام الأماني 👋</h1>
        <p className="mt-1" style={{ color: '#6b7280' }}>اختر الخدمة التي تريد تقديمها للزبون</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {services.map(service => (
          <div
            key={service.title}
            onClick={() => alert(`${service.title} - سيتم إضافة هذه الصفحة قريباً`)}
            className="rounded-2xl overflow-hidden"
            style={{ background: 'white', boxShadow: '0 1px 3px rgba(0,0,0,0.05)', cursor: 'pointer', transition: 'box-shadow 0.3s' }}
            onMouseEnter={e => (e.currentTarget.style.boxShadow = '0 10px 25px -5px rgba(0,0,0,0.1)')}
            onMouseLeave={e => (e.currentTarget.style.boxShadow = '0 1px 3px rgba(0,0,0,0.05)')}
          >
            <div style={{ height: '0.5rem', background: service.color }} />
            <div className="p-6">
              <div
                className="flex items-center justify-center text-2xl mb-4"
                style={{ width: '3.5rem', height: '3.5rem', borderRadius: '0.75rem', background: service.color + '15' }}
              >
                {service.icon}
              </div>
              <h3 className="text-lg font-bold mb-2" style={{ color: service.color }}>{service.title}</h3>
              <p className="text-sm" style={{ color: '#6b7280' }}>{service.description}</p>
            </div>
            <div className="px-6 pb-5">
              <div
                className="w-full py-2.5 rounded-lg text-white text-sm font-semibold text-center"
                style={{ background: service.color }}
              >
                ابدأ الآن ←
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className="mt-10">
        <h2 className="text-xl font-bold mb-4" style={{ color: '#1a3a5c' }}>طلباتي الأخيرة</h2>
        {loadingReqs ? (
          <div className="text-center py-8 rounded-2xl" style={{ background: 'white', boxShadow: '0 1px 3px rgba(0,0,0,0.05)', color: '#9ca3af' }}>جارٍ التحميل...</div>
        ) : myRequests.length === 0 ? (
          <div className="text-center py-8 rounded-2xl" style={{ background: 'white', boxShadow: '0 1px 3px rgba(0,0,0,0.05)', color: '#9ca3af' }}>لا توجد طلبات بعد</div>
        ) : (
          <div className="rounded-2xl overflow-hidden" style={{ background: 'white', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
            <table className="w-full text-sm">
              <thead>
                <tr style={{ background: '#f8fafc' }}>
                  <th className="px-4 py-3 text-right font-semibold" style={{ color: '#4b5563' }}>اسم الزبون</th>
                  <th className="px-4 py-3 text-right font-semibold" style={{ color: '#4b5563' }}>نوع الشراء</th>
                  <th className="px-4 py-3 text-right font-semibold" style={{ color: '#4b5563' }}>الحالة</th>
                  <th className="px-4 py-3 text-right font-semibold" style={{ color: '#4b5563' }}>التاريخ</th>
                </tr>
              </thead>
              <tbody>
                {myRequests.map((req, i) => (
                  <tr key={req.id} style={{ background: i % 2 === 0 ? 'white' : '#fafafa' }}>
                    <td className="px-4 py-3 font-medium" style={{ color: '#1f2937' }}>
                      {req.customer ? `${req.customer.fullName} ${req.customer.fatherName} ${req.customer.grandfatherName}` : '-'}
                    </td>
                    <td className="px-4 py-3" style={{ color: '#4b5563' }}>{typeLabel(req.purchaseType)}</td>
                    <td className="px-4 py-3">{getStatusBadge(req.status)}</td>
                    <td className="px-4 py-3" style={{ color: '#6b7280' }}>{req.createdAt ? formatDate(req.createdAt) : '-'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
