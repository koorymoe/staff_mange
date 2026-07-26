import { useEffect, useState } from 'react'
import { api, type LeaderInvoice } from '../api'

// قائمة بسيطة لعرض فواتير الليدر السابقة (كل الفواتير أو حسب الموظف).
export default function LeaderInvoicesListPage() {
  const [invoices, setInvoices] = useState<LeaderInvoice[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.getLeaderInvoices().then(setInvoices).finally(() => setLoading(false))
  }, [])

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">فواتير الليدر</h2>
      <p className="mt-1 text-slate-500">كل فواتير التنفيذ التي أنشأها الليدرز عبر النظام.</p>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}

      {!loading && (
        <div className="mt-6 overflow-x-auto rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <table className="w-full min-w-[720px] text-sm">
            <thead>
              <tr className="border-b border-slate-100 text-right text-slate-400">
                <th className="px-4 py-3">كود المحاسبة</th>
                <th className="px-4 py-3">الزبون</th>
                <th className="px-4 py-3">المنظومات</th>
                <th className="px-4 py-3">تكاليف التنفيذ</th>
                <th className="px-4 py-3">مجموع المواد</th>
                <th className="px-4 py-3">المجموع الصافي</th>
                <th className="px-4 py-3">التاريخ</th>
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
                  <td className="px-4 py-3 text-slate-400">
                    {new Date(inv.createdAt).toLocaleDateString('ar-IQ')}
                  </td>
                </tr>
              ))}
              {invoices.length === 0 && (
                <tr>
                  <td colSpan={7} className="px-4 py-6 text-center text-slate-400">
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
