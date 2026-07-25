import { useState, useEffect, useRef } from 'react'
import { api, type GpsDeviceRequest } from '../../api'

// Some legacy records store purchaseType/subscriptionType in lowercase, so we widen
// those two fields to `string` here rather than reusing the strict literal unions.
type DeviceRequest = Omit<GpsDeviceRequest, 'purchaseType' | 'subscriptionType'> & {
  purchaseType: string
  subscriptionType: string
}

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

export default function GpsDevices() {
  const [requests, setRequests] = useState<DeviceRequest[]>([])
  const [loading, setLoading] = useState(true)
  const [selected, setSelected] = useState<DeviceRequest | null>(null)
  const [delivering, setDelivering] = useState(false)
  const [activationDate, setActivationDate] = useState('')
  const [deviceChecks, setDeviceChecks] = useState({ checked: false, activated: false, delivered: false })
  const printRef = useRef<HTMLDivElement>(null)

  async function load() {
    try {
      const data = await api.getGpsDevices()
      setRequests(((data || []) as DeviceRequest[]).filter((d) => d.status === 'PENDING'))
    } catch (e) { console.error(e) }
    setLoading(false)
  }
  // `load` is async and only sets state after its awaits resolve; false positive.
  // eslint-disable-next-line react-hooks/set-state-in-effect
  useEffect(() => { load() }, [])

  async function handleDeliver(req: DeviceRequest) {
    if (req.purchaseType === 'DEVICE_ONLY') {
      if (!deviceChecks.checked || !deviceChecks.activated || !deviceChecks.delivered) {
        alert('يرجى تأكيد جميع الخطوات أولاً'); return
      }
    } else if (req.subscriptionType && !activationDate) {
      alert('يرجى إدخال تاريخ التفعيل أولاً'); return
    }
    setDelivering(true)
    try {
      const days = getSubDays(req.subscriptionType)
      const endDate = activationDate ? (() => { const d = new Date(activationDate); d.setDate(d.getDate() + days); return d.toISOString().split('T')[0] })() : undefined
      await api.updateGpsDevice(req.id, {
        status: 'DELIVERED',
        activationDate: activationDate || undefined,
        subscriptionEnd: endDate,
      })
      setSelected(null)
      setActivationDate('')
      setDeviceChecks({ checked: false, activated: false, delivered: false })
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ أثناء التفعيل')
    } finally { setDelivering(false) }
  }

  function handlePrint(req: DeviceRequest) {
    const customerName = req.customer ? `${req.customer.fullName} ${req.customer.fatherName} ${req.customer.grandfatherName}` : '-'
    const subLabel = req.subscriptionType ? getSubscriptionLabel(req.subscriptionType) : 'لا يوجد'
    const purchaseDate = formatDate(req.createdAt)
    const actDate = activationDate ? new Date(activationDate).toLocaleDateString('ar-IQ') : '-'
    const expDate = activationDate && req.subscriptionType ? (() => {
      const days = getSubDays(req.subscriptionType)
      const d = new Date(activationDate); d.setDate(d.getDate() + days)
      return d.toLocaleDateString('ar-IQ')
    })() : '-'
    const price = req.price ? req.price.toLocaleString() + ' د.ع' : '-'

    const makeCopy = (copyLabel: string) => `
      <div style="width:100%;height:100%;box-sizing:border-box;font-family:'Segoe UI',Arial,sans-serif;direction:rtl;background:white;">
        <div style="border:2px solid #b8974a;border-radius:6px;overflow:hidden;height:100%;display:flex;flex-direction:column;">
          <div style="flex:1;padding:8px 14px 8px 8px;overflow:hidden;">
            <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:6px;">
              <span style="font-size:12px;font-weight:bold;color:#1a3a5c;">فاتورة اشتراك GPS</span>
              <span style="font-size:9px;background:#1a3a5c;color:white;padding:2px 8px;border-radius:10px;">نسخة ${copyLabel}</span>
            </div>
            <table style="width:100%;border-collapse:collapse;font-size:10px;">
              <tr><td style="padding:3px 6px;color:#444;width:38%;border-bottom:1px solid #f0e8d0;background:#f8f4ee;font-weight:600;">اسم الزبون</td><td style="padding:3px 6px;font-weight:bold;color:#1a3a5c;border-bottom:1px solid #f0e8d0;">${customerName}</td></tr>
              <tr><td style="padding:3px 6px;color:#444;border-bottom:1px solid #f0e8d0;background:#f8f4ee;font-weight:600;">رقم الهاتف</td><td style="padding:3px 6px;font-weight:bold;border-bottom:1px solid #f0e8d0;">${req.customer?.phone || '-'}</td></tr>
              <tr><td style="padding:3px 6px;color:#444;border-bottom:1px solid #f0e8d0;background:#f8f4ee;font-weight:600;">العنوان</td><td style="padding:3px 6px;border-bottom:1px solid #f0e8d0;">${req.customer?.address || '-'}</td></tr>
              <tr><td style="padding:3px 6px;color:#444;border-bottom:1px solid #f0e8d0;background:#f8f4ee;font-weight:600;">نوع الاشتراك</td><td style="padding:3px 6px;font-weight:bold;border-bottom:1px solid #f0e8d0;">${subLabel}</td></tr>
              <tr><td style="padding:3px 6px;color:#444;border-bottom:1px solid #f0e8d0;background:#f8f4ee;font-weight:600;">المبلغ المدفوع</td><td style="padding:3px 6px;font-weight:bold;color:#b8974a;font-size:12px;border-bottom:1px solid #f0e8d0;">${price}</td></tr>
              <tr><td style="padding:3px 6px;color:#444;border-bottom:1px solid #f0e8d0;background:#f8f4ee;font-weight:600;">تاريخ الشراء</td><td style="padding:3px 6px;border-bottom:1px solid #f0e8d0;">${purchaseDate}</td></tr>
              <tr><td style="padding:3px 6px;color:#444;border-bottom:1px solid #f0e8d0;background:#f8f4ee;font-weight:600;">تاريخ التفعيل</td><td style="padding:3px 6px;font-weight:bold;color:#16a34a;border-bottom:1px solid #f0e8d0;">${actDate}</td></tr>
              <tr><td style="padding:3px 6px;color:#444;background:#f8f4ee;font-weight:600;">تاريخ الانتهاء</td><td style="padding:3px 6px;font-weight:bold;color:#dc2626;">${expDate}</td></tr>
            </table>
            <div style="margin-top:8px;display:flex;justify-content:space-between;font-size:9px;color:#999;border-top:1px dashed #c9a84c;padding-top:6px;">
              <span>توقيع الزبون: _______________</span>
              <span>توقيع المسؤول: _______________</span>
            </div>
          </div>
        </div>
      </div>
    `
    const win = window.open('', '_blank')
    if (!win) return
    win.document.write(`
      <html><head><title>فاتورة</title>
      <style>@page { size: A4 portrait; margin: 5mm; } html, body { margin: 0; padding: 0; background: white; } .page { display: flex; flex-direction: column; gap: 3mm; } .copy { flex: 1; } .divider { border: none; border-top: 2px dashed #b8974a; }</style></head>
      <body><div class="page"><div class="copy">${makeCopy('الزبون')}</div><hr class="divider" /><div class="copy">${makeCopy('الشركة')}</div></div>
      <script>window.onload = function(){ window.print(); window.close(); }</script></body></html>
    `)
    win.document.close()
  }

  const typeLabel = (t: string) => (t === 'DEVICE_SIM' || t === 'device_sim') ? 'جهاز + SIM كارد' : 'جهاز فقط'

  return (
    <div dir="rtl">
      <h1 className="text-2xl font-bold mb-6" style={{ color: '#1a3a5c' }}>الطلبات المعلقة 📋</h1>
      {loading ? (
        <div className="text-center py-20" style={{ color: '#9ca3af' }}>جارٍ التحميل...</div>
      ) : requests.length === 0 ? (
        <div className="text-center py-16 rounded-2xl" style={{ background: 'white', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
          <div className="text-5xl mb-4">✅</div>
          <p className="text-lg" style={{ color: '#9ca3af' }}>لا توجد طلبات معلقة</p>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
          {requests.map(req => (
            <div key={req.id} className="rounded-xl p-5 flex items-center justify-between"
              style={{ background: 'white', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
              <div style={{ flex: 1 }}>
                <div className="flex items-center gap-3 mb-1">
                  <span className="font-bold" style={{ color: '#1f2937' }}>{req.customer ? `${req.customer.fullName} ${req.customer.fatherName} ${req.customer.grandfatherName}` : '-'}</span>
                  <span className="text-xs px-2 py-0.5 rounded-full" style={{ background: '#f5f3ff', color: '#7c3aed' }}>{typeLabel(req.purchaseType)}</span>
                  {req.subscriptionType && <span className="text-xs px-2 py-0.5 rounded-full" style={{ background: '#eff6ff', color: '#1d4ed8' }}>{getSubscriptionLabel(req.subscriptionType)}</span>}
                </div>
                <div className="flex gap-4 text-sm" style={{ color: '#6b7280' }}>
                  <span>📞 {req.customer?.phone}</span>
                  <span>📍 {req.customer?.address}</span>
                  <span>👤 {req.employee?.name || '-'}</span>
                  <span>📅 {formatDate(req.createdAt)}</span>
                </div>
              </div>
              <button onClick={() => { setSelected(req); setActivationDate(''); setDeviceChecks({ checked: false, activated: false, delivered: false }) }}
                className="text-sm px-4 py-2 mr-4 rounded-lg font-medium"
                style={{ background: '#f8fafc', color: '#1a3a5c', border: '1px solid #e2e8f0', cursor: 'pointer' }}>
                مراجعة ←
              </button>
            </div>
          ))}
        </div>
      )}

      {selected && (
        <div className="fixed inset-0 z-50 flex items-center justify-center" style={{ background: 'rgba(0,0,0,0.5)' }}>
          <div className="rounded-2xl w-full max-w-3xl m-4" style={{ background: 'white', boxShadow: '0 25px 50px -12px rgba(0,0,0,0.25)', maxHeight: '90vh', overflowY: 'auto' }}>
            <div className="p-5 flex items-center justify-between" style={{ borderBottom: '1px solid #e5e7eb', position: 'sticky', top: 0, background: 'white', zIndex: 1 }}>
              <h2 className="text-xl font-bold" style={{ color: '#1a3a5c' }}>مراجعة الطلب</h2>
              <button onClick={() => setSelected(null)} style={{ color: '#9ca3af', fontSize: '1.5rem', background: 'none', border: 'none', cursor: 'pointer' }}>✕</button>
            </div>
            <div className="p-6" style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
              <div className="grid grid-cols-2 gap-3 text-sm">
                <InfoRow label="الاسم الكامل" value={`${selected.customer?.fullName || ''} ${selected.customer?.fatherName || ''} ${selected.customer?.grandfatherName || ''}`} />
                <InfoRow label="رقم الهاتف" value={selected.customer?.phone || '-'} />
                <InfoRow label="العنوان" value={selected.customer?.address || '-'} />
                <InfoRow label="نوع الشراء" value={typeLabel(selected.purchaseType)} />
                {selected.subscriptionType && <InfoRow label="نوع الاشتراك" value={getSubscriptionLabel(selected.subscriptionType)} />}
                {selected.price ? <InfoRow label="المبلغ" value={`${selected.price.toLocaleString()} د.ع`} /> : null}
                <InfoRow label="تاريخ الطلب" value={formatDate(selected.createdAt)} />
              </div>

              {(selected.purchaseType === 'DEVICE_ONLY' || selected.purchaseType === 'device_only') ? (
                <div className="rounded-xl p-4" style={{ background: '#f0fdf4', border: '1px solid #bbf7d0' }}>
                  <p className="text-sm font-bold mb-3" style={{ color: '#166534' }}>تأكيد تسليم الجهاز</p>
                  {[
                    { key: 'checked', label: '✅ تم فحص الجهاز' },
                    { key: 'activated', label: '✅ تم تفعيل الجهاز' },
                    { key: 'delivered', label: '✅ تم تسليم الجهاز' },
                  ].map(({ key, label }) => (
                    <label key={key} className="flex items-center gap-3 cursor-pointer mb-2">
                      <input type="checkbox" checked={deviceChecks[key as keyof typeof deviceChecks]}
                        onChange={e => setDeviceChecks(prev => ({ ...prev, [key]: e.target.checked }))}
                        className="w-5 h-5 rounded" />
                      <span className="text-sm font-medium" style={{ color: '#374151' }}>{label}</span>
                    </label>
                  ))}
                </div>
              ) : selected.subscriptionType ? (
                <div className="rounded-xl p-4" style={{ background: '#eff6ff', border: '1px solid #bfdbfe' }}>
                  <label className="block text-sm font-bold mb-2" style={{ color: '#1e40af' }}>📅 تاريخ التفعيل (مطلوب) *</label>
                  <input type="date" className="w-full text-sm rounded-xl px-4 py-2.5" style={{ border: '1px solid #e5e7eb', outline: 'none' }}
                    value={activationDate} onChange={e => setActivationDate(e.target.value)} />
                  {activationDate && (
                    <div className="mt-3 grid grid-cols-2 gap-2 text-sm">
                      <div className="rounded-lg p-2 text-center" style={{ background: 'white' }}>
                        <div className="text-xs" style={{ color: '#6b7280' }}>تاريخ التفعيل</div>
                        <div className="font-bold" style={{ color: '#15803d' }}>{new Date(activationDate).toLocaleDateString('ar-IQ')}</div>
                      </div>
                      <div className="rounded-lg p-2 text-center" style={{ background: 'white' }}>
                        <div className="text-xs" style={{ color: '#6b7280' }}>تاريخ الانتهاء</div>
                        <div className="font-bold" style={{ color: '#dc2626' }}>{(() => {
                          const days = getSubDays(selected.subscriptionType)
                          const d = new Date(activationDate); d.setDate(d.getDate() + days)
                          return d.toLocaleDateString('ar-IQ')
                        })()}</div>
                      </div>
                    </div>
                  )}
                </div>
              ) : null}

              {selected.customer?.idCardFrontUrl && (
                <div>
                  <h4 className="font-semibold mb-3" style={{ color: '#1a3a5c' }}>وثائق الهوية</h4>
                  <div className="grid grid-cols-2 gap-3">
                    {selected.customer.idCardFrontUrl && <DocImage url={selected.customer.idCardFrontUrl} label="الهوية - أمامي" />}
                    {selected.customer.idCardBackUrl && <DocImage url={selected.customer.idCardBackUrl} label="الهوية - خلفي" />}
                    {selected.customer.residenceCardFrontUrl && <DocImage url={selected.customer.residenceCardFrontUrl} label="السكن - أمامي" />}
                    {selected.customer.residenceCardBackUrl && <DocImage url={selected.customer.residenceCardBackUrl} label="السكن - خلفي" />}
                  </div>
                </div>
              )}

              {selected.invoicePhotoUrl && (
                <div>
                  <h4 className="font-semibold mb-3" style={{ color: '#1a3a5c' }}>صورة الفاتورة</h4>
                  <img src={selected.invoicePhotoUrl} alt="الفاتورة" className="w-full rounded-xl" style={{ maxHeight: '16rem', objectFit: 'contain', border: '1px solid #e5e7eb' }} />
                </div>
              )}
            </div>
            <div className="p-5 flex gap-3" style={{ borderTop: '1px solid #e5e7eb', position: 'sticky', bottom: 0, background: 'white' }}>
              <button onClick={() => handleDeliver(selected)}
                disabled={delivering || ((selected.purchaseType === 'DEVICE_ONLY' || selected.purchaseType === 'device_only') && (!deviceChecks.checked || !deviceChecks.activated || !deviceChecks.delivered))}
                className="flex-1 py-3 rounded-lg font-semibold text-white"
                style={{ background: '#1a3a5c', border: 'none', cursor: 'pointer' }}>
                {delivering ? 'جارٍ التفعيل...' : 'تفعيل الجهاز ✅'}
              </button>
              <button onClick={() => handlePrint(selected)}
                className="px-4 py-3 rounded-lg font-medium"
                style={{ background: '#f8fafc', color: '#1a3a5c', border: '1px solid #e2e8f0', cursor: 'pointer' }}>
                🖨️ طباعة
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
      <div ref={printRef} />
    </div>
  )
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return <div className="flex gap-2"><span className="font-semibold" style={{ color: '#6b7280', whiteSpace: 'nowrap' }}>{label}:</span><span style={{ color: '#1f2937' }}>{value}</span></div>
}

function DocImage({ url, label }: { url: string; label: string }) {
  return (
    <div className="rounded-lg overflow-hidden" style={{ border: '1px solid #e5e7eb' }}>
      <p className="text-xs p-2" style={{ color: '#6b7280', background: '#f9fafb', borderBottom: '1px solid #e5e7eb' }}>{label}</p>
      <img src={url} alt={label} className="w-full" style={{ objectFit: 'contain', maxHeight: '9rem' }} />
    </div>
  )
}
