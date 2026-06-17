import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, type Quotation } from '../api'

const statusConfig: Record<Quotation['status'], { label: string; color: string }> = {
  NEW: { label: 'جديد', color: 'bg-yellow-100 text-yellow-800' },
  SENT: { label: 'مرسل', color: 'bg-blue-100 text-blue-800' },
  ACCEPTED: { label: 'مقبول', color: 'bg-green-100 text-green-800' },
  REJECTED: { label: 'مرفوض', color: 'bg-red-100 text-red-800' },
}

export default function QuotationsPage() {
  const navigate = useNavigate()
  const [quotations, setQuotations] = useState<Quotation[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = () => {
    setLoading(true)
    api.getQuotations()
      .then(setQuotations)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const handleDelete = async (id: string) => {
    if (!confirm('هل أنت متأكد من حذف عرض السعر؟')) return
    try {
      await api.deleteQuotation(id)
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    }
  }

  return (
    <div>
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-brand-900">عروض الأسعار</h2>
        <button
          onClick={() => navigate('/quotations/new')}
          className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-2 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30"
        >
          عرض سعر جديد
        </button>
      </div>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && (
        <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">
          تعذر الاتصال بالخادم: {error}
        </p>
      )}

      {!loading && !error && (
        <div className="mt-6 overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <table className="w-full text-right">
            <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
              <tr>
                <th className="px-4 py-3 text-sm font-semibold">رقم العرض</th>
                <th className="px-4 py-3 text-sm font-semibold">اسم الزبون</th>
                <th className="px-4 py-3 text-sm font-semibold">المشروع</th>
                <th className="px-4 py-3 text-sm font-semibold">المجموع</th>
                <th className="px-4 py-3 text-sm font-semibold">الخصم</th>
                <th className="px-4 py-3 text-sm font-semibold">الصافي</th>
                <th className="px-4 py-3 text-sm font-semibold">الحالة</th>
                <th className="px-4 py-3 text-sm font-semibold">التاريخ</th>
                <th className="px-4 py-3 text-sm font-semibold">إجراءات</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {quotations.map((q) => {
                const st = statusConfig[q.status]
                return (
                  <tr key={q.id} className="hover:bg-slate-50">
                    <td className="px-4 py-3 font-medium">{q.quotationNumber}</td>
                    <td className="px-4 py-3">{q.customerName}</td>
                    <td className="px-4 py-3">{q.projectName || '-'}</td>
                    <td className="px-4 py-3">{q.grandTotal.toLocaleString()} د.ع</td>
                    <td className="px-4 py-3">{q.discountValue.toLocaleString()} د.ع</td>
                    <td className="px-4 py-3 font-bold">{q.netTotal.toLocaleString()} د.ع</td>
                    <td className="px-4 py-3">
                      <span className={`rounded-full px-3 py-1 text-xs font-medium ${st.color}`}>
                        {st.label}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm text-slate-500">
                      {new Date(q.createdAt).toLocaleDateString('ar-IQ')}
                    </td>
                    <td className="px-4 py-3">
                      <button
                        onClick={() => handleDelete(q.id)}
                        className="text-sm text-red-600 hover:text-red-800"
                      >
                        حذف
                      </button>
                    </td>
                  </tr>
                )
              })}
              {quotations.length === 0 && (
                <tr>
                  <td colSpan={9} className="px-4 py-6 text-center text-slate-400">
                    لا توجد عروض أسعار بعد
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
