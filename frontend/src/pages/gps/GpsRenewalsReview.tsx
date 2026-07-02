import { useEffect, useState } from 'react'
import { api, type GpsRenewalRequest } from '../../api'

const subLabel = (t: string) => t === 'THREE_MONTHS' ? '3 أشهر' : t === 'SIX_MONTHS' ? '6 أشهر' : 'سنوي'
const subDays = (t: string) => t === 'THREE_MONTHS' ? 90 : t === 'SIX_MONTHS' ? 180 : 365

function printRenewal(renewal: GpsRenewalRequest, newEnd: Date) {
  const customerName = `${renewal.customer.fullName} ${renewal.customer.fatherName || ''} ${renewal.customer.grandfatherName || ''}`.trim()
  const newEndStr = newEnd.toLocaleDateString('ar-IQ')
  const oldEndStr = renewal.deviceRequest.subscriptionEnd ? new Date(renewal.deviceRequest.subscriptionEnd).toLocaleDateString('ar-IQ') : '-'

  const copy = (label: string) => `
    <div style="width:190mm;box-sizing:border-box;font-family:'Segoe UI',Tahoma,Arial,sans-serif;direction:rtl;border:1.5px solid #b8974a;border-radius:6px;padding:10px 14px;background:white;">
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:8px;">
        <span style="font-size:13px;font-weight:bold;color:#0f2040;">فاتورة تجديد اشتراك GPS</span>
        <span style="font-size:10px;color:#aaa;">نسخة: ${label}</span>
      </div>
      <table style="width:100%;border-collapse:collapse;font-size:12px;">
        <tr><td style="padding:5px 7px;color:#444;width:38%;border-bottom:1px solid #f0e8d0;">اسم الزبون</td><td style="padding:5px 7px;font-weight:bold;color:#0f2040;border-bottom:1px solid #f0e8d0;">${customerName}</td></tr>
        <tr style="background:#fffdf5;"><td style="padding:5px 7px;color:#444;border-bottom:1px solid #f0e8d0;">رقم الهاتف</td><td style="padding:5px 7px;font-weight:bold;border-bottom:1px solid #f0e8d0;">${renewal.customer.phone}</td></tr>
        <tr><td style="padding:5px 7px;color:#444;border-bottom:1px solid #f0e8d0;">نوع الاشتراك</td><td style="padding:5px 7px;font-weight:bold;border-bottom:1px solid #f0e8d0;">${subLabel(renewal.subscriptionType)}</td></tr>
        <tr style="background:#fffdf5;"><td style="padding:5px 7px;color:#444;border-bottom:1px solid #f0e8d0;">تاريخ الانتهاء السابق</td><td style="padding:5px 7px;border-bottom:1px solid #f0e8d0;">${oldEndStr}</td></tr>
        <tr><td style="padding:5px 7px;color:#444;">تاريخ الانتهاء الجديد</td><td style="padding:5px 7px;font-weight:bold;color:#16a34a;">${newEndStr}</td></tr>
      </table>
      <div style="margin-top:14px;display:flex;justify-content:space-between;font-size:11px;color:#999;border-top:1px solid #f0e8d0;padding-top:8px;">
        <span>توقيع الزبون: _______________</span>
        <span>توقيع المسؤول: _______________</span>
      </div>
    </div>
  `

  const win = window.open('', '_blank')
  if (!win) return
  win.document.write(`<!DOCTYPE html><html><head><meta charset="utf-8"/><title>فاتورة تجديد</title>
    <style>@page { size: A4 portrait; margin: 8mm; } body { margin:0; padding:0; display:flex; flex-direction:column; gap:8mm; align-items:center; background:white; }</style>
    </head><body>
      ${copy('الزبون')}
      <div style="border-top:2px dashed #ccc;width:190mm;"></div>
      ${copy('الشركة')}
      <script>window.onload = function(){ window.print(); }</script>
    </body></html>`)
  win.document.close()
}

export default function GpsRenewalsReview() {
  const [renewals, setRenewals] = useState<GpsRenewalRequest[]>([])
  const [loading, setLoading] = useState(true)
  const [selected, setSelected] = useState<GpsRenewalRequest | null>(null)
  const [approving, setApproving] = useState(false)

  const load = () => {
    setLoading(true)
    api.getGpsRenewals().then(all => setRenewals(all.filter(r => r.status === 'PENDING'))).finally(() => setLoading(false))
  }
  useEffect(load, [])

  const handleApprove = async (renewal: GpsRenewalRequest) => {
    setApproving(true)
    try {
      const oldEnd = renewal.deviceRequest.subscriptionEnd ? new Date(renewal.deviceRequest.subscriptionEnd) : new Date()
      const base = oldEnd > new Date() ? oldEnd : new Date()
      const newEnd = new Date(base.getTime() + subDays(renewal.subscriptionType) * 86400000)

      await api.updateGpsRenewal(renewal.id, { status: 'APPROVED', newEndDate: newEnd.toISOString() })
      await api.updateGpsDevice(renewal.deviceRequestId, {
        subscriptionEnd: newEnd.toISOString(),
        subscriptionType: renewal.subscriptionType,
        subscriptionStatus: 'ACTIVE',
      })

      printRenewal(renewal, newEnd)
      setSelected(null)
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ أثناء المعالجة')
    } finally {
      setApproving(false)
    }
  }

  return (
    <div>
      <h2 className="mb-6 text-2xl font-bold text-brand-900">طلبات تجديد الاشتراك 🔄</h2>

      {loading && <p className="py-20 text-center text-slate-400">جاري التحميل...</p>}
      {!loading && renewals.length === 0 && (
        <div className="rounded-2xl bg-white py-16 text-center shadow-sm">
          <div className="mb-4 text-5xl">✅</div>
          <p className="text-lg text-slate-400">لا توجد طلبات تجديد معلقة</p>
        </div>
      )}

      {!loading && renewals.length > 0 && (
        <div className="overflow-hidden rounded-2xl bg-white shadow-sm">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-slate-50">
                <th className="px-4 py-3 text-right font-semibold text-slate-600">اسم الزبون</th>
                <th className="px-4 py-3 text-right font-semibold text-slate-600">الهاتف</th>
                <th className="px-4 py-3 text-right font-semibold text-slate-600">تاريخ الانتهاء الحالي</th>
                <th className="px-4 py-3 text-right font-semibold text-slate-600">النوع المطلوب</th>
                <th className="px-4 py-3 text-right font-semibold text-slate-600">إجراء</th>
              </tr>
            </thead>
            <tbody>
              {renewals.map((r, i) => (
                <tr key={r.id} className={i % 2 === 0 ? 'bg-white' : 'bg-slate-50/50'}>
                  <td className="px-4 py-3 font-medium text-brand-900">{r.customer.fullName} {r.customer.fatherName} {r.customer.grandfatherName}</td>
                  <td className="px-4 py-3 text-slate-600">{r.customer.phone}</td>
                  <td className="px-4 py-3 text-slate-600">{r.deviceRequest.subscriptionEnd ? new Date(r.deviceRequest.subscriptionEnd).toLocaleDateString('ar-IQ') : '-'}</td>
                  <td className="px-4 py-3 text-slate-600">{subLabel(r.subscriptionType)}</td>
                  <td className="px-4 py-3"><button onClick={() => setSelected(r)} className="rounded-lg bg-brand-500 px-4 py-1.5 text-xs font-bold text-white hover:bg-brand-600">موافقة ومعالجة ←</button></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {selected && (
        <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 p-4">
          <div className="w-full max-w-lg rounded-2xl bg-white shadow-2xl">
            <div className="flex items-center justify-between border-b border-slate-100 p-5">
              <h3 className="text-lg font-bold text-brand-900">تأكيد التجديد</h3>
              <button onClick={() => setSelected(null)} className="text-2xl text-slate-400 hover:text-slate-600">×</button>
            </div>
            <div className="space-y-4 p-6 text-sm">
              <div className="grid grid-cols-2 gap-3 rounded-xl bg-slate-50 p-4">
                <p><span className="text-slate-500">الاسم: </span><b>{selected.customer.fullName} {selected.customer.fatherName} {selected.customer.grandfatherName}</b></p>
                <p><span className="text-slate-500">الهاتف: </span><b>{selected.customer.phone}</b></p>
                <p><span className="text-slate-500">الاشتراك المطلوب: </span><b>{subLabel(selected.subscriptionType)}</b></p>
                <p><span className="text-slate-500">الانتهاء الحالي: </span><b className="text-red-600">{selected.deviceRequest.subscriptionEnd ? new Date(selected.deviceRequest.subscriptionEnd).toLocaleDateString('ar-IQ') : '-'}</b></p>
              </div>
              <div className="rounded-xl bg-emerald-50 p-4 text-emerald-800">سيتم تجديد الاشتراك تلقائياً وطباعة الفاتورة عند الموافقة</div>
            </div>
            <div className="flex gap-3 border-t border-slate-100 p-5">
              <button onClick={() => handleApprove(selected)} disabled={approving} className="flex-1 rounded-lg bg-brand-500 py-3 text-sm font-bold text-white hover:bg-brand-600 disabled:opacity-50">
                {approving ? 'جاري المعالجة...' : '✅ موافقة وطباعة فاتورة'}
              </button>
              <button onClick={() => setSelected(null)} className="rounded-lg border border-slate-200 px-6 py-3 text-sm font-bold text-slate-600 hover:bg-slate-50">إغلاق</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
