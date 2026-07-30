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
                <th className="px-4 py-3">الزبون</th>
                <th className="px-4 py-3">المنظومات</th>
                <th className="px-4 py-3">تكاليف التنفيذ</th>
                <th className="px-4 py-3">مجموع المواد</th>
                <th className="px-4 py-3">المجموع الصافي</th>
                <th className="px-4 py-3">الحالة</th>
                <th className="px-4 py-3">التاريخ</th>
                {canApprove && <th className="px-4 py-3">إجراءات</th>}
              </tr>
            </thead>
            <tbody>
              {invoices.map((inv) => (
                <tr key={inv.id} className="border-b border-slate-50 last:border-0">
                  <td className="px-4 py-3 font-mono text-brand-700">{inv.accountingCode}</td>
                  <td className="px-4 py-3">{inv.customerName || '—'}</td>
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
                  {canApprove && (
                    <td className="px-4 py-3">
                      {inv.status !== 'APPROVED' && (
                        <button
                          onClick={() => handleApprove(inv.id)}
                          disabled={busyId === inv.id}
                          className="rounded-lg bg-brand-600 px-3 py-1.5 text-xs font-bold text-white hover:bg-brand-700 disabled:opacity-50"
                        >
                          {busyId === inv.id ? 'جاري الاعتماد...' : 'اعتماد'}
                        </button>
                      )}
                    </td>
                  )}
                </tr>
              ))}
              {invoices.length === 0 && (
                <tr>
                  <td colSpan={canApprove ? 8 : 7} className="px-4 py-6 text-center text-slate-400">
                    لا توجد فواتير بعد.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
