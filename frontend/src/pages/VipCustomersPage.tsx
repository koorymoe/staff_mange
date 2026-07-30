import { useEffect, useState } from 'react'
import { api, type VipCustomer } from '../api'

// قائمة الشخصيات المهمة — لمدير النظام حصراً (الراوت بالسيرفر محمي بـrequireAdmin).
// تعرض تفاصيل الزبون ورقمه وشنو طلب من عدنا ومنو الموظف الي علّمه.
export default function VipCustomersPage() {
  const [items, setItems] = useState<VipCustomer[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState<string | null>(null)

  const load = () => {
    api.getVipCustomers()
      .then(setItems)
      .catch((e) => setError(e instanceof Error ? e.message : 'تعذر جلب القائمة'))
      .finally(() => setLoading(false))
  }
  useEffect(load, [])

  const remove = async (customerId: string) => {
    if (!confirm('إزالة تعليم "شخصية مهمة" عن هذا الزبون؟')) return
    setBusy(customerId)
    try {
      await api.unmarkVipCustomer(customerId)
      load()
    } finally {
      setBusy(null)
    }
  }

  return (
    <div dir="rtl">
      <h2 className="text-2xl font-bold text-brand-900">⭐ الشخصيات المهمة</h2>
      <p className="mt-1 text-slate-500">
        الزبائن الي علّمهم الموظفون كشخصيات مهمة — مع تفاصيل الزبون وشنو طلب من عدنا ومنو الموظف الي علّمه.
      </p>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && <p className="mt-4 rounded-lg bg-red-50 p-4 text-red-600">{error}</p>}

      {!loading && !error && (
        <div className="mt-6 overflow-x-auto rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <table className="w-full min-w-[760px] text-sm">
            <thead>
              <tr className="border-b border-slate-100 text-right text-slate-400">
                <th className="px-4 py-3">الزبون</th>
                <th className="px-4 py-3">رقم الهاتف</th>
                <th className="px-4 py-3">شنو طلب</th>
                <th className="px-4 py-3">رمز الحجز</th>
                <th className="px-4 py-3">علّمه</th>
                <th className="px-4 py-3">التاريخ</th>
                <th className="px-4 py-3">إجراءات</th>
              </tr>
            </thead>
            <tbody>
              {items.map((v) => (
                <tr key={v.id} className="border-b border-slate-50 last:border-0">
                  <td className="px-4 py-3 font-bold text-brand-900">⭐ {v.customerName}</td>
                  <td className="px-4 py-3 font-bold text-brand-700" dir="ltr">{v.customerPhone}</td>
                  <td className="px-4 py-3">{v.requestSummary || '—'}</td>
                  <td className="px-4 py-3 font-mono text-slate-500">{v.bookingCode || '—'}</td>
                  <td className="px-4 py-3">{v.markedByName}</td>
                  <td className="px-4 py-3 text-slate-400">{new Date(v.createdAt).toLocaleDateString('ar-IQ')}</td>
                  <td className="px-4 py-3">
                    <button
                      onClick={() => remove(v.customerId)}
                      disabled={busy === v.customerId}
                      className="rounded-lg bg-red-50 px-3 py-1.5 text-xs font-bold text-red-700 hover:bg-red-100 disabled:opacity-50"
                    >
                      إزالة التعليم
                    </button>
                  </td>
                </tr>
              ))}
              {items.length === 0 && (
                <tr>
                  <td colSpan={7} className="px-4 py-6 text-center text-slate-400">
                    ما اكو زبائن معلّمين كشخصيات مهمة بعد.
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
