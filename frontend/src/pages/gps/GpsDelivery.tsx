import { useEffect, useState } from 'react'
import { api, type GpsDeviceRequest } from '../../api'

const subLabel = (t: string) => t === 'THREE_MONTHS' ? '3 أشهر' : t === 'SIX_MONTHS' ? '6 أشهر' : 'سنوي'

function printAgreement(req: GpsDeviceRequest) {
  const customerName = `${req.customer.fullName} ${req.customer.fatherName || ''} ${req.customer.grandfatherName || ''}`.trim()
  const today = new Date().toLocaleDateString('ar-IQ')
  const accountNum = req.id.slice(-6).toUpperCase()

  const copy = (label: string) => `
    <div style="width:190mm;box-sizing:border-box;font-family:'Segoe UI',Tahoma,Arial,sans-serif;direction:rtl;border:2px solid #b8974a;border-radius:5px;padding:10px 14px;background:white;">
      <div style="display:flex;justify-content:space-between;align-items:flex-start;margin-bottom:6px;padding-bottom:5px;border-bottom:2px solid #b8974a;">
        <div>
          <div style="font-size:13px;font-weight:bold;color:#0f2040;">شركة الأماني للتجارة العامة والاستثمارات العقارية والوكالات التجارية م.م</div>
          <div style="font-size:9px;color:#888;">تعهد شراء جهاز GPS مع خط</div>
        </div>
        <div style="text-align:left;">
          <span style="font-size:9px;background:#0f2040;color:white;padding:3px 10px;border-radius:10px;">نسخة ${label}</span>
          <div style="font-size:8px;color:#888;margin-top:3px;">رقم: ${accountNum} | ${today}</div>
        </div>
      </div>
      <div style="font-size:9.5px;color:#333;margin-bottom:8px;padding:6px 10px;border:1px solid #e8d8b0;border-radius:4px;background:#fffdf5;line-height:1.7;">
        إني <strong>${customerName}</strong> اشتريت جهاز GPS من <strong>شركة الأماني للتجارة العامة والاستثمارات العقارية والوكالات التجارية محدودة المسؤولية</strong> وأوافق على جميع بنود هذا العقد
      </div>
      <table style="width:100%;border-collapse:collapse;font-size:9.5px;margin-bottom:10px;">
        <tr>
          <td style="padding:5px 7px;border:1px solid #d4b896;background:#f8f4ee;font-weight:600;width:16%;">رقم الهاتف</td>
          <td style="padding:5px 7px;border:1px solid #d4b896;width:18%;">${req.customer.phone || '___'}</td>
          <td style="padding:5px 7px;border:1px solid #d4b896;background:#f8f4ee;font-weight:600;width:16%;">عنوان السكن</td>
          <td style="padding:5px 7px;border:1px solid #d4b896;width:18%;">${req.customer.address || '___'}</td>
          <td style="padding:5px 7px;border:1px solid #d4b896;background:#f8f4ee;font-weight:600;width:16%;">نوع الاشتراك</td>
          <td style="padding:5px 7px;border:1px solid #d4b896;font-weight:bold;width:16%;">${subLabel(req.subscriptionType)}</td>
        </tr>
        <tr>
          <td style="padding:5px 7px;border:1px solid #d4b896;background:#f8f4ee;font-weight:600;">رقم بطاقة السكن</td>
          <td style="padding:5px 7px;border:1px solid #d4b896;">${req.residenceCardNumber || '___________'}</td>
          <td style="padding:5px 7px;border:1px solid #d4b896;background:#f8f4ee;font-weight:600;">ID GPS (رقم الجهاز)</td>
          <td style="padding:5px 7px;border:1px solid #d4b896;font-weight:bold;color:#0f2040;">${req.gpsNumber || '___________'}</td>
          <td style="padding:5px 7px;border:1px solid #d4b896;background:#f8f4ee;font-weight:600;">نوع الشراء</td>
          <td style="padding:5px 7px;border:1px solid #d4b896;">${req.purchaseType === 'DEVICE_SIM' ? 'جهاز + SIM' : 'جهاز فقط'}</td>
        </tr>
      </table>
      <div style="font-weight:bold;font-size:10px;color:#0f2040;border-bottom:1px solid #c9a84c;padding-bottom:3px;margin-bottom:5px;">بنود العقد</div>
      <ol style="margin:0;padding-right:18px;font-size:9.5px;line-height:2.4;color:#222;columns:2;column-gap:18px;">
        <li>يلتزم المشتري بدفع الاشتراك الشهري في موعده وفي حال التأخر يتحمل المسؤولية الكاملة عن أي انقطاع.</li>
        <li>الجهاز ملك للمشتري وعليه المحافظة عليه وفي حال تلفه يتحمل تكلفة الاستبدال كاملاً.</li>
        <li>لا تتحمل الشركة مسؤولية الضرر الناجم عن انقطاع الإنترنت أو الكهرباء أو أي قوة قاهرة.</li>
        <li>يحق للشركة إيقاف الخدمة فوراً في حال مخالفة أي بند من بنود هذا العقد.</li>
        <li>لا يحق للمشتري نقل الجهاز أو الشريحة لشخص آخر دون إشعار الشركة خطياً.</li>
        <li>في حال الرغبة بإلغاء الاشتراك يجب إشعار الشركة قبل أسبوع من تاريخ الاستحقاق.</li>
        <li>يتعهد المشتري بعدم استخدام الجهاز لأغراض غير مشروعة ويخضع العقد لقوانين العراق.</li>
        <li>في حال فقدان الجهاز أو سرقته يجب إبلاغ الشركة فوراً خلال 24 ساعة.</li>
        <li>يلتزم المشتري بعدم فك الجهاز أو التلاعب في برمجته تحت طائلة إلغاء الضمان.</li>
        <li>تلتزم الشركة بتقديم الدعم الفني خلال أوقات الدوام الرسمي فقط.</li>
        <li>يحق للشركة تعديل أسعار الاشتراك مع إشعار المشتري قبل شهر من التعديل.</li>
        <li>يُقر المشتري بأنه اطلع على جميع بنود هذا العقد ووافق عليها طوعاً واختياراً.</li>
      </ol>
      <div style="display:grid;grid-template-columns:1fr 1fr 1fr 1fr;gap:10px;font-size:9px;border-top:1px dashed #c9a84c;padding-top:8px;margin-top:8px;">
        <div style="text-align:center;"><div style="height:24px;border-bottom:1px solid #333;margin-bottom:3px;"></div><div>توقيع المشتري</div></div>
        <div style="text-align:center;"><div style="height:24px;border-bottom:1px solid #333;margin-bottom:3px;"></div><div>اسم الفني</div></div>
        <div style="text-align:center;"><div style="height:24px;border-bottom:1px solid #333;margin-bottom:3px;"></div><div>توقيع مخول</div></div>
        <div style="text-align:center;"><div style="height:24px;border-bottom:1px solid #333;margin-bottom:3px;font-weight:bold;padding-top:5px;">${today}</div><div>التاريخ</div></div>
      </div>
    </div>
  `

  const win = window.open('', '_blank')
  if (!win) return
  win.document.write(`<!DOCTYPE html><html><head><meta charset="utf-8"/><title>تعهد GPS</title>
    <style>
      @page { size: A4 portrait; margin: 5mm; }
      body { margin: 0; padding: 0; background: white; display:flex; flex-direction:column; align-items:center; }
      .cut-line { width:190mm; border-top:1px dashed #999; margin:4mm 0; position:relative; height:0; }
      .cut-line span { position:absolute; right:50%; top:-2.5mm; transform:translateX(50%); background:white; padding:0 4px; font-size:7px; color:#999; }
    </style></head>
    <body>
      ${copy('الزبون')}
      <div class="cut-line"><span>✂ خط القص</span></div>
      ${copy('الشركة')}
      <script>window.onload = function(){ window.print(); }</script>
    </body></html>`)
  win.document.close()
}

export default function GpsDelivery() {
  const [tab, setTab] = useState<'pending' | 'delivered'>('pending')
  const [pending, setPending] = useState<GpsDeviceRequest[]>([])
  const [delivered, setDelivered] = useState<GpsDeviceRequest[]>([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [deliveringId, setDeliveringId] = useState<string | null>(null)

  const load = () => {
    setLoading(true)
    api.getGpsDevices().then(all => {
      setPending(all.filter(r => r.status === 'APPROVED' && !r.isDelivered))
      setDelivered(all.filter(r => r.isDelivered).slice(0, 20))
    }).finally(() => setLoading(false))
  }
  useEffect(load, [])

  const handleDeliver = async (req: GpsDeviceRequest) => {
    setDeliveringId(req.id)
    try {
      await api.updateGpsDevice(req.id, { status: 'DELIVERED', isDelivered: true, deliveredAt: new Date().toISOString() })
      printAgreement(req)
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ أثناء التسليم')
    } finally {
      setDeliveringId(null)
    }
  }

  const filtered = pending.filter(r =>
    !search ||
    r.customer.fullName.includes(search) ||
    r.customer.phone.includes(search) ||
    (r.gpsNumber || '').includes(search)
  )

  const customerFullName = (req: GpsDeviceRequest) => `${req.customer.fullName} ${req.customer.fatherName || ''} ${req.customer.grandfatherName || ''}`.trim()

  return (
    <div className="mx-auto max-w-3xl">
      <h2 className="mb-6 text-2xl font-bold text-brand-900">تسليم أجهزة GPS 📦</h2>

      <div className="mb-5 flex gap-2">
        <button onClick={() => setTab('pending')}
          className={`rounded-xl border-2 px-5 py-2 text-sm font-bold transition-all ${tab === 'pending' ? 'border-brand-500 bg-brand-500 text-white' : 'border-brand-500 text-brand-700'}`}>
          📦 بانتظار التسليم ({pending.length})
        </button>
        <button onClick={() => setTab('delivered')}
          className={`rounded-xl border-2 px-5 py-2 text-sm font-bold transition-all ${tab === 'delivered' ? 'border-brand-500 bg-brand-500 text-white' : 'border-brand-500 text-brand-700'}`}>
          ✅ تم التسليم مؤخراً ({delivered.length})
        </button>
      </div>

      {tab === 'pending' && (
        <>
          <div className="mb-4 flex gap-2">
            <input value={search} onChange={e => setSearch(e.target.value)} placeholder="ابحث باسم الزبون أو الهاتف أو رقم الجهاز..."
              className="flex-1 rounded-xl border border-slate-200 px-4 py-2.5 text-sm" />
          </div>

          {loading && <p className="py-20 text-center text-slate-400">جاري التحميل...</p>}
          {!loading && filtered.length === 0 && <p className="py-20 text-center text-slate-400">لا توجد أجهزة بانتظار التسليم</p>}

          <div className="flex flex-col gap-3">
            {filtered.map(req => (
              <div key={req.id} className="rounded-xl bg-white p-5 shadow-sm">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1">
                    <p className="mb-1 text-lg font-bold text-brand-900">{customerFullName(req)}</p>
                    <div className="mt-2 grid grid-cols-2 gap-x-6 gap-y-1 text-sm text-slate-500">
                      <span>📞 {req.customer.phone}</span>
                      <span>📦 رقم الجهاز: <b>{req.gpsNumber || '-'}</b></span>
                      <span>📋 {subLabel(req.subscriptionType)}</span>
                      {req.subscriptionEnd && <span className="text-red-600">⏰ ينتهي: {new Date(req.subscriptionEnd).toLocaleDateString('ar-IQ')}</span>}
                    </div>
                  </div>
                  <button onClick={() => handleDeliver(req)} disabled={deliveringId === req.id}
                    className="rounded-lg bg-brand-500 px-4 py-2 text-sm font-bold text-white hover:bg-brand-600 disabled:opacity-50">
                    {deliveringId === req.id ? 'جاري...' : 'تسليم وطباعة 🖨️'}
                  </button>
                </div>
              </div>
            ))}
          </div>
        </>
      )}

      {tab === 'delivered' && (
        <div className="flex flex-col gap-3">
          {delivered.length === 0 && <p className="py-20 text-center text-slate-400">لا توجد تسليمات مؤخراً</p>}
          {delivered.map(req => (
            <div key={req.id} className="rounded-xl bg-white p-5 shadow-sm">
              <div className="flex items-start justify-between gap-4">
                <div className="flex-1">
                  <p className="mb-1 text-lg font-bold text-brand-900">{customerFullName(req)}</p>
                  <div className="mt-2 flex gap-4 text-sm text-slate-500">
                    <span>📞 {req.customer.phone}</span>
                    <span>📦 {req.gpsNumber || '-'}</span>
                    {req.deliveredAt && <span className="text-emerald-600">✅ سُلّم: {new Date(req.deliveredAt).toLocaleDateString('ar-IQ')}</span>}
                  </div>
                </div>
                <button onClick={() => printAgreement(req)} className="rounded-lg border border-slate-200 px-4 py-2 text-sm font-bold text-slate-600 hover:bg-slate-50">
                  🖨️ إعادة طباعة
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
