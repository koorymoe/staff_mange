import { useEffect, useState } from 'react'
import { api, type LeaderInvoice } from '../api'
import { useSession } from '../session'

// قائمة بسيطة لعرض فواتير الليدر السابقة (كل الفواتير أو حسب الموظف).
// الفاتورة تضل SUBMITTED (ظاهرة عند الليدر) لين مدير/محاسب يعتمدها لـAPPROVED —
// الليدر ما يقدر يعتمد فاتورته بنفسه (زر "اعتماد" ما يظهر إلا لـADMIN/FINANCE).
export default function LeaderInvoicesListPage() {
  const { employee } = useSession()
  const canApprove = employee?.role === 'ADMIN' || employee?.role === 'FINANCE'
  const [invoices, setInvoices] = useState<LeaderInvoice[]>([])
  const [loading, setLoading] = useState(true)
  const [busyId, setBusyId] = useState<string | null>(null)
  const [details, setDetails] = useState<LeaderInvoice | null>(null)

  const load = () => { api.getLeaderInvoices().then(setInvoices).finally(() => setLoading(false)) }
  useEffect(load, [])

  const handleApprove = async (id: string) => {
    setBusyId(id)
    try {
      await api.approveLeaderInvoice(id)
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر اعتماد الفاتورة')
    } finally {
      setBusyId(null)
    }
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">فواتير الليدر</h2>
      <p className="mt-1 text-slate-500">كل فواتير التنفيذ التي أنشأها الليدرز عبر النظام.</p>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}

      {!loading && (
        <div className="mt-6 overflow-x-auto rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <table className="w-full min-w-[820px] text-sm">
            <thead>
              <tr className="border-b border-slate-100 text-right text-slate-400">
                <th className="px-4 py-3">كود المحاسبة</th>
                <th className="px-4 py-3">الليدر</th>
                <th className="px-4 py-3">الزبون</th>
                <th className="px-4 py-3">المنظومات</th>
                <th className="px-4 py-3">تكاليف التنفيذ</th>
                <th className="px-4 py-3">مجموع المواد</th>
                <th className="px-4 py-3">المجموع الصافي</th>
                <th className="px-4 py-3">الحالة</th>
                <th className="px-4 py-3">التاريخ</th>
                <th className="px-4 py-3">إجراءات</th>
              </tr>
            </thead>
            <tbody>
              {invoices.map((inv) => (
                <tr key={inv.id} className="border-b border-slate-50 last:border-0">
                  <td className="px-4 py-3 font-mono text-brand-700">{inv.accountingCode}</td>
                  <td className="px-4 py-3">
                    <span className="font-bold text-slate-700">{inv.employeeName || '—'}</span>
                    {inv.bookingCode && <div className="text-xs text-slate-400">حجز {inv.bookingCode}</div>}
                  </td>
                  <td className="px-4 py-3">{inv.customerName || inv.booking?.customer?.name || '—'}</td>
                  <td className="px-4 py-3">{inv.systems.join('، ')}</td>
                  <td className="px-4 py-3">{inv.executionCost.toLocaleString()}</td>
                  <td className="px-4 py-3">{inv.materialsTotal.toLocaleString()}</td>
                  <td className="px-4 py-3 font-bold text-brand-800">{inv.netTotal.toLocaleString()}</td>
                  <td className="px-4 py-3">
                    <span className={`rounded-full px-3 py-1 text-xs font-bold ${
                      inv.status === 'APPROVED' ? 'bg-emerald-50 text-emerald-700' : 'bg-amber-50 text-amber-700'
                    }`}>
                      {inv.status === 'APPROVED' ? '✔ معتمدة' : 'بانتظار الاعتماد'}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-slate-400">
                    {new Date(inv.createdAt).toLocaleDateString('ar-IQ')}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex gap-2">
                      <button
                        onClick={() => setDetails(inv)}
                        className="rounded-lg bg-slate-100 px-3 py-1.5 text-xs font-bold text-slate-700 hover:bg-slate-200"
                      >
                        📋 التفاصيل
                      </button>
                      {canApprove && inv.status !== 'APPROVED' && (
                        <button
                          onClick={() => handleApprove(inv.id)}
                          disabled={busyId === inv.id}
                          className="rounded-lg bg-brand-600 px-3 py-1.5 text-xs font-bold text-white hover:bg-brand-700 disabled:opacity-50"
                        >
                          {busyId === inv.id ? 'جاري الاعتماد...' : 'اعتماد'}
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
              {invoices.length === 0 && (
                <tr>
                  <td colSpan={10} className="px-4 py-6 text-center text-slate-400">
                    لا توجد فواتير بعد.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* تفاصيل الفاتورة — كل الي يحتاجه المحاسب بمكان واحد */}
      {details && (
        <div className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/40 p-4" onClick={() => setDetails(null)}>
          <div dir="rtl" className="my-8 w-full max-w-3xl rounded-2xl bg-white p-6 shadow-xl" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-start justify-between">
              <div>
                <h3 className="text-lg font-bold text-brand-900">تفاصيل الفاتورة</h3>
                <p className="font-mono text-sm text-brand-700">{details.accountingCode}</p>
              </div>
              <button onClick={() => setDetails(null)} className="rounded-lg px-3 py-1 text-slate-400 hover:bg-slate-100">✕</button>
            </div>

            <section className="mt-4 rounded-xl bg-slate-50 p-4">
              <h4 className="mb-2 text-sm font-bold text-slate-700">منو رفعها</h4>
              <div className="grid grid-cols-2 gap-2 text-sm">
                <p>الليدر: <b>{details.employeeName || '—'}</b></p>
                <p>الدور: {details.employeeRole || '—'}</p>
                <p dir="ltr" className="text-right">الهاتف: {details.employeePhone || '—'}</p>
                <p>التاريخ: {new Date(details.createdAt).toLocaleString('ar-IQ')}</p>
                <p>الحالة: {details.status === 'APPROVED' ? '✔ معتمدة' : 'بانتظار الاعتماد'}</p>
                {details.approvedByName && <p>اعتمدها: <b>{details.approvedByName}</b></p>}
              </div>
            </section>

            <section className="mt-3 rounded-xl bg-blue-50 p-4">
              <h4 className="mb-2 text-sm font-bold text-blue-900">الزبون والحجز</h4>
              <div className="grid grid-cols-2 gap-2 text-sm">
                <p>الزبون: <b>{details.customerName || details.booking?.customer?.name || '—'}</b></p>
                <p dir="ltr" className="text-right">الهاتف: {details.customerPhone || details.booking?.customer?.phone || '—'}</p>
                <p className="col-span-2">العنوان: {details.customerAddress || details.booking?.address || '—'}</p>
                {details.bookingCode && <p>رمز الحجز: <b>{details.bookingCode}</b></p>}
                {details.booking?.service && <p>الخدمة: {details.booking.service.name}</p>}
                {details.booking && (
                  <>
                    <p>حالة الحجز: {details.booking.status}</p>
                    <p>المستلم بالحجز: <b>{(details.booking.amountCollected ?? 0).toLocaleString()} د.ع</b></p>
                    <p>تقدير الإداري: <b>{(details.booking.quotedPrice ?? 0).toLocaleString()} د.ع</b></p>
                    <p>التدقيق: {details.booking.amountVerified ? '✔ مدقق' : 'بانتظار التدقيق'}</p>
                  </>
                )}
              </div>
              {details.booking?.assignments && details.booking.assignments.length > 0 && (
                <p className="mt-2 text-sm">
                  الكادر المنفّذ: {details.booking.assignments.map((a) => a.employee?.name).filter(Boolean).join('، ')}
                </p>
              )}
            </section>

            <section className="mt-3 rounded-xl border border-slate-200 p-4">
              <h4 className="mb-2 text-sm font-bold text-slate-700">بنود التنفيذ</h4>
              <p className="mb-2 text-xs text-slate-500">المنظومات: {details.systems.join('، ') || '—'} · عدد الأجهزة: {details.totalDeviceCount}</p>
              <table className="w-full text-right text-sm">
                <thead><tr className="text-slate-400">
                  <th className="py-1">البند</th><th className="py-1">العدد</th><th className="py-1">تفاصيل</th>
                </tr></thead>
                <tbody>
                  {details.items.map((it, i) => (
                    <tr key={i} className="border-t border-slate-100">
                      <td className="py-1">{it.itemName} <span className="text-xs text-slate-400">{it.systemName}</span></td>
                      <td className="py-1">{it.count}</td>
                      <td className="py-1 text-xs text-slate-500">
                        {it.heightMeters ? `ارتفاع ${it.heightMeters}م` : ''}
                        {it.cableLengthMeters ? ` · كيبل ${it.cableLengthMeters}م` : ''}
                      </td>
                    </tr>
                  ))}
                  {details.items.length === 0 && <tr><td colSpan={3} className="py-2 text-slate-400">ماكو بنود</td></tr>}
                </tbody>
              </table>
            </section>

            {details.materials.length > 0 && (
              <section className="mt-3 rounded-xl border border-slate-200 p-4">
                <h4 className="mb-2 text-sm font-bold text-slate-700">المواد</h4>
                <table className="w-full text-right text-sm">
                  <thead><tr className="text-slate-400">
                    <th className="py-1">المادة</th><th className="py-1">الكمية</th><th className="py-1">المجموع</th>
                  </tr></thead>
                  <tbody>
                    {details.materials.map((m, i) => (
                      <tr key={i} className="border-t border-slate-100">
                        <td className="py-1">{m.name}</td>
                        <td className="py-1">{m.quantity}</td>
                        <td className="py-1">{m.lineTotal.toLocaleString()}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </section>
            )}

            <section className="mt-3 rounded-xl bg-brand-50 p-4 text-sm">
              <div className="flex justify-between"><span>تكاليف التنفيذ</span><b>{details.executionCost.toLocaleString()} د.ع</b></div>
              <div className="flex justify-between"><span>مجموع المواد</span><b>{details.materialsTotal.toLocaleString()} د.ع</b></div>
              <div className="flex justify-between"><span>الخصم</span><b>{details.discountValue.toLocaleString()} د.ع</b></div>
              <div className="mt-2 flex justify-between border-t border-brand-200 pt-2 text-base">
                <span className="font-bold">المجموع الصافي</span>
                <b className="text-brand-800">{details.netTotal.toLocaleString()} د.ع</b>
              </div>
            </section>
          </div>
        </div>
      )}
    </div>
  )
}
