import { useState, useEffect } from 'react'
import { api, type GpsRenewalRequest } from '../../api'

function formatDate(d: string) {
  if (!d) return '-'
  try { return new Date(d).toLocaleDateString('ar-IQ') } catch { return d }
}

function getSubscriptionLabel(t: string) {
  switch (t) {
    case 'THREE_MONTHS': case '3_months': return '3 أشهر'
    case 'SIX_MONTHS': case '6_months': return '6 أشهر'
    case 'YEARLY': case 'yearly': return 'سنوي'
    default: return t || '-'
  }
}

function getSubDays(t: string) {
  if (t === 'THREE_MONTHS' || t === '3_months') return 90
  if (t === 'SIX_MONTHS' || t === '6_months') return 180
  return 365
}

export default function GpsRenewals() {
  const [renewals, setRenewals] = useState<GpsRenewalRequest[]>([])
  const [loading, setLoading] = useState(true)
  const [selected, setSelected] = useState<GpsRenewalRequest | null>(null)
  const [approving, setApproving] = useState(false)

  async function load() {
    try {
      const data = await api.getGpsRenewals()
      setRenewals((data || []).filter((r) => r.status === 'PENDING' || r.status === 'pending'))
    } catch (e) { console.error(e) }
    setLoading(false)
  }
  // `load` is async and only sets state after its awaits resolve; false positive.
  // eslint-disable-next-line react-hooks/set-state-in-effect
  useEffect(() => { load() }, [])

  async function handleApprove(renewal: GpsRenewalRequest) {
    setApproving(true)
    try {
      const days = getSubDays(renewal.subscriptionType)
      const now = new Date()
      const currentEnd = renewal.currentEnd ? new Date(renewal.currentEnd) : now
      const startFrom = currentEnd > now ? currentEnd : now
      const newEnd = new Date(startFrom)
      newEnd.setDate(newEnd.getDate() + days)

      await api.updateGpsRenewal(renewal.id, {
        status: 'APPROVED',
        newEndDate: newEnd.toISOString(),
      })
      // Also update the device request if available
      if (renewal.deviceRequestId) {
        await api.updateGpsDevice(renewal.deviceRequestId, {
          subscriptionEnd: newEnd.toISOString().split('T')[0],
        })
      }
      printRenewalInvoice(renewal)
      setSelected(null)
      load()
    } catch (e) { console.error(e); alert('حدث خطأ') }
    setApproving(false)
  }

  function printRenewalInvoice(renewal: GpsRenewalRequest) {
    const customerName = renewal.customer
      ? `${renewal.customer.fullName} ${renewal.customer.fatherName} ${renewal.customer.grandfatherName}`
      : '-'
    const days = getSubDays(renewal.subscriptionType)
    const now = new Date()
    const end = renewal.currentEnd ? new Date(renewal.currentEnd) : now
    const startFrom = end > now ? end : now
    const newEnd = new Date(startFrom)
    newEnd.setDate(newEnd.getDate() + days)
    const newEndStr = newEnd.toLocaleDateString('ar-IQ')
    const subLabel = getSubscriptionLabel(renewal.subscriptionType)

    const makeCopy = (copyLabel: string) => `
      <div style="width:190mm;min-height:130mm;box-sizing:border-box;font-family:'Segoe UI',Arial,sans-serif;direction:rtl;border:1.5px solid #b8974a;border-radius:6px;overflow:hidden;display:flex;flex-direction:column;">
        <div style="flex:1;padding:10px 14px 10px 8px;">
          <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:8px;">
            <span style="font-size:12px;font-weight:bold;color:#1a3a5c;">فاتورة تجديد اشتراك</span>
            <span style="font-size:10px;color:#aaa;">نسخة: ${copyLabel}</span>
          </div>
          <table style="width:100%;border-collapse:collapse;font-size:12px;">
            <tr><td style="padding:4px 6px;color:#444;width:38%;border-bottom:1px solid #f0e8d0;">اسم الزبون</td><td style="padding:4px 6px;font-weight:bold;color:#1a3a5c;border-bottom:1px solid #f0e8d0;">${customerName}</td></tr>
            <tr style="background:#fffdf5;"><td style="padding:4px 6px;color:#444;border-bottom:1px solid #f0e8d0;">رقم الهاتف</td><td style="padding:4px 6px;font-weight:bold;border-bottom:1px solid #f0e8d0;">${renewal.customer?.phone || '-'}</td></tr>
            <tr><td style="padding:4px 6px;color:#444;border-bottom:1px solid #f0e8d0;">العنوان</td><td style="padding:4px 6px;border-bottom:1px solid #f0e8d0;">${renewal.customer?.address || '-'}</td></tr>
            <tr style="background:#fffdf5;"><td style="padding:4px 6px;color:#444;border-bottom:1px solid #f0e8d0;">نوع الاشتراك</td><td style="padding:4px 6px;font-weight:bold;border-bottom:1px solid #f0e8d0;">${subLabel}</td></tr>
            <tr><td style="padding:4px 6px;color:#444;border-bottom:1px solid #f0e8d0;">تاريخ الانتهاء السابق</td><td style="padding:4px 6px;border-bottom:1px solid #f0e8d0;">${renewal.currentEnd ? formatDate(renewal.currentEnd) : '-'}</td></tr>
            <tr style="background:#fffdf5;"><td style="padding:4px 6px;color:#444;">تاريخ الانتهاء الجديد</td><td style="padding:4px 6px;font-weight:bold;color:#16a34a;">${newEndStr}</td></tr>
          </table>
          <div style="margin-top:12px;display:flex;justify-content:space-between;font-size:11px;color:#999;border-top:1px solid #f0e8d0;padding-top:8px;">
            <span>توقيع الزبون: _______________</span>
            <span>توقيع المسؤول: _______________</span>
          </div>
        </div>
      </div>
    `

    const win = window.open('', '_blank')
    if (!win) return
    win.document.write(`
      <html><head><title>فاتورة تجديد</title>
      <style>@page { size: A4 portrait; margin: 8mm; } body { margin: 0; padding: 0; display: flex; flex-direction: column; gap: 8mm; align-items: center; background: white; }</style></head>
      <body>
        ${makeCopy('الزبون')}
        <div style="border-top:2px dashed #ccc;width:100%;margin:2mm 0;"></div>
        ${makeCopy('الشركة')}
        <script>window.onload = function(){ window.print(); window.close(); }</script>
      </body></html>
    `)
    win.document.close()
  }

  return (
    <div dir="rtl">
      <h1 className="text-2xl font-bold mb-6" style={{ color: '#1a3a5c' }}>طلبات تجديد الاشتراك 🔄</h1>
      {loading ? (
        <div className="text-center py-20" style={{ color: '#9ca3af' }}>جارٍ التحميل...</div>
      ) : renewals.length === 0 ? (
        <div className="text-center py-16 rounded-2xl" style={{ background: 'white', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
          <div className="text-5xl mb-4">✅</div>
          <p className="text-lg" style={{ color: '#9ca3af' }}>لا توجد طلبات تجديد معلقة</p>
        </div>
      ) : (
        <div className="rounded-2xl overflow-hidden" style={{ background: 'white', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
          <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr style={{ background: '#f8fafc' }}>
                <th className="px-4 py-3 text-right font-semibold" style={{ color: '#4b5563' }}>اسم الزبون</th>
                <th className="px-4 py-3 text-right font-semibold" style={{ color: '#4b5563' }}>الهاتف</th>
                <th className="px-4 py-3 text-right font-semibold" style={{ color: '#4b5563' }}>تاريخ انتهاء الاشتراك</th>
                <th className="px-4 py-3 text-right font-semibold" style={{ color: '#4b5563' }}>النوع المطلوب</th>
                <th className="px-4 py-3 text-right font-semibold" style={{ color: '#4b5563' }}>تاريخ الطلب</th>
                <th className="px-4 py-3 text-right font-semibold" style={{ color: '#4b5563' }}>إجراء</th>
              </tr>
            </thead>
            <tbody>
              {renewals.map((r, i) => (
                <tr key={r.id} style={{ background: i % 2 === 0 ? 'white' : '#fafafa' }}>
                  <td className="px-4 py-3 font-medium" style={{ color: '#1f2937' }}>
                    {r.customer ? `${r.customer.fullName} ${r.customer.fatherName} ${r.customer.grandfatherName}` : '-'}
                  </td>
                  <td className="px-4 py-3" style={{ color: '#4b5563' }}>{r.customer?.phone || '-'}</td>
                  <td className="px-4 py-3" style={{ color: '#4b5563' }}>{r.currentEnd ? formatDate(r.currentEnd) : '-'}</td>
                  <td className="px-4 py-3" style={{ color: '#4b5563' }}>{getSubscriptionLabel(r.subscriptionType)}</td>
                  <td className="px-4 py-3" style={{ color: '#6b7280' }}>{r.createdAt ? formatDate(r.createdAt) : '-'}</td>
                  <td className="px-4 py-3">
                    <button onClick={() => setSelected(r)}
                      className="text-sm px-4 py-1.5 rounded-lg font-medium"
                      style={{ background: '#f8fafc', color: '#1a3a5c', border: '1px solid #e2e8f0', cursor: 'pointer' }}>
                      موافقة ومعالجة ←
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          </div>
        </div>
      )}

      {selected && (
        <div className="fixed inset-0 z-50 flex items-center justify-center" style={{ background: 'rgba(0,0,0,0.5)' }}>
          <div className="rounded-2xl w-full max-w-lg m-4" style={{ background: 'white', boxShadow: '0 25px 50px -12px rgba(0,0,0,0.25)' }}>
            <div className="p-5 flex items-center justify-between" style={{ borderBottom: '1px solid #e5e7eb' }}>
              <h2 className="text-xl font-bold" style={{ color: '#1a3a5c' }}>تأكيد التجديد</h2>
              <button onClick={() => setSelected(null)} style={{ color: '#9ca3af', fontSize: '1.5rem', background: 'none', border: 'none', cursor: 'pointer' }}>✕</button>
            </div>
            <div className="p-6 text-sm" style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div className="grid grid-cols-2 gap-3 p-4 rounded-xl" style={{ background: '#f9fafb' }}>
                <div><span style={{ color: '#6b7280' }}>الاسم: </span><span className="font-semibold">{selected.customer ? `${selected.customer.fullName} ${selected.customer.fatherName} ${selected.customer.grandfatherName}` : '-'}</span></div>
                <div><span style={{ color: '#6b7280' }}>الهاتف: </span><span className="font-semibold">{selected.customer?.phone || '-'}</span></div>
                <div><span style={{ color: '#6b7280' }}>العنوان: </span><span className="font-semibold">{selected.customer?.address || '-'}</span></div>
                <div><span style={{ color: '#6b7280' }}>الاشتراك المطلوب: </span><span className="font-semibold">{getSubscriptionLabel(selected.subscriptionType)}</span></div>
                <div><span style={{ color: '#6b7280' }}>تاريخ انتهاء الحالي: </span><span className="font-semibold" style={{ color: '#dc2626' }}>{selected.currentEnd ? formatDate(selected.currentEnd) : '-'}</span></div>
              </div>
              <div className="text-sm rounded-xl p-4" style={{ background: '#f0fdf4', border: '1px solid #bbf7d0', color: '#166534' }}>
                سيتم تجديد الاشتراك تلقائياً وطباعة الفاتورة عند الموافقة
              </div>
            </div>
            <div className="p-5 flex gap-3" style={{ borderTop: '1px solid #e5e7eb' }}>
              <button onClick={() => handleApprove(selected)} disabled={approving}
                className="flex-1 py-3 rounded-lg font-semibold text-white"
                style={{ background: '#1a3a5c', border: 'none', cursor: 'pointer' }}>
                {approving ? 'جارٍ المعالجة...' : '✅ موافقة وطباعة فاتورة'}
              </button>
              <button onClick={() => setSelected(null)}
                className="px-6 py-3 rounded-lg font-medium"
                style={{ border: '1px solid #e5e7eb', color: '#4b5563', background: 'white', cursor: 'pointer' }}>
                إغلاق
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
