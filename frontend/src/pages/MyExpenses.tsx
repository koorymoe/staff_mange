import { useEffect, useState } from 'react'
import { api, type Expense } from '../api'
import { useSession } from '../session'

const statusLabels: Record<string, string> = {
  PENDING: 'بانتظار الاعتماد',
  APPROVED: 'تم الاعتماد',
  REJECTED: 'مرفوض',
}

const statusColors: Record<string, string> = {
  PENDING: 'bg-amber-100 text-amber-700',
  APPROVED: 'bg-emerald-100 text-emerald-700',
  REJECTED: 'bg-red-100 text-red-700',
}

export default function MyExpenses() {
  const { employee } = useSession()
  const [expenses, setExpenses] = useState<Expense[]>([])
  const [amount, setAmount] = useState('')
  const [description, setDescription] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const load = () => {
    if (!employee) return
    api.getExpenses(employee.id).then(setExpenses)
  }

  useEffect(load, [employee?.id])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!employee) return
    setSubmitting(true)
    try {
      await api.createExpense({
        employeeId: employee.id,
        amount: Number(amount),
        description: description || undefined,
      })
      setAmount('')
      setDescription('')
      load()
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">مصاريفي</h2>
      <p className="mt-1 text-slate-500">
        سجل أي مبلغ صرفته من جيبك الخاص لإنجاز مهمة، ليتم اعتماده واسترجاعه من قبل المحاسب.
      </p>

      <form
        onSubmit={handleSubmit}
        className="mt-6 grid grid-cols-1 gap-4 rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)] sm:grid-cols-3"
      >
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">المبلغ</label>
          <input
            required
            type="number"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>
        <div className="sm:col-span-2">
          <label className="mb-1 block text-sm font-medium text-slate-600">سبب الصرف</label>
          <input
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="مثال: شراء قطعة غيار، وقود..."
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>
        <div className="sm:col-span-3">
          <button
            type="submit"
            disabled={submitting}
            className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-2 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30 disabled:opacity-50"
          >
            {submitting ? 'جاري الإرسال...' : 'إرسال للمحاسب'}
          </button>
        </div>
      </form>

      <div className="mt-6 overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        <table className="w-full text-right">
          <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
            <tr>
              <th className="px-4 py-3 text-sm font-semibold">المبلغ</th>
              <th className="px-4 py-3 text-sm font-semibold">السبب</th>
              <th className="px-4 py-3 text-sm font-semibold">التاريخ</th>
              <th className="px-4 py-3 text-sm font-semibold">الحالة</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {expenses.map((e) => (
              <tr key={e.id}>
                <td className="px-4 py-3 font-bold text-brand-800">{e.amount.toLocaleString()}</td>
                <td className="px-4 py-3 text-slate-600">{e.description || '-'}</td>
                <td className="px-4 py-3 text-slate-500">
                  {new Date(e.createdAt).toLocaleDateString('ar-IQ')}
                </td>
                <td className="px-4 py-3">
                  <span className={`rounded-full px-2 py-1 text-xs font-bold ${statusColors[e.status]}`}>
                    {statusLabels[e.status]}
                  </span>
                </td>
              </tr>
            ))}
            {expenses.length === 0 && (
              <tr>
                <td colSpan={4} className="px-4 py-6 text-center text-slate-400">
                  لا توجد مصاريف مسجلة
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
