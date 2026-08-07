import { useEffect, useState } from 'react'
import { api, type Expense, type Booking } from '../api'
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
  const [activeBookings, setActiveBookings] = useState<Booking[]>([])
  const [canExpense, setCanExpense] = useState(false)

  const load = () => {
    if (!employee) return
    api.getExpenses(employee.id).then(setExpenses)
    // حجوزاته الشغالة بس. جان يسحب كل أرشيف الشركة (بكل زبائنه
    // وتعييناته) ويفلتره بالمتصفح حتى يلكه حجوزاته — يعني كل ما كبر
    // الأرشيف صارت صفحة مصاريفه أبطأ، وهو أصلاً ما إله علاقة بالباقي.
    api.getBookings({ assignedTo: 'me' }).then(bookings => {
      const active = bookings.filter(b =>
        (b.status === 'CONFIRMED' || b.status === 'IN_PROGRESS') &&
        b.assignments.some(a => a.employee.id === employee.id)
      )
      setActiveBookings(active)

      const isResponsible = active.some(b => {
        if (b.expenseResponsibleId === employee.id) return true
        if (!b.expenseResponsibleId) {
          const isAlone = b.assignments.length === 1 && b.assignments[0].employee.id === employee.id && !b.projectSupervisor
          if (isAlone) return true
          if (employee.isLeader && !b.projectSupervisor) return true
        }
        return false
      })
      setCanExpense(isResponsible || employee.isLeader)
    })
  }

  useEffect(load, [employee])

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

  const responsibleBookings = activeBookings.filter(b =>
    b.expenseResponsibleId === employee?.id ||
    (!b.expenseResponsibleId && (
      (b.assignments.length === 1 && b.assignments[0].employee.id === employee?.id && !b.projectSupervisor) ||
      (employee?.isLeader && b.projectSupervisor?.id === employee?.id)
    ))
  )

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">مصاريفي</h2>
      <p className="mt-1 text-slate-500">
        سجل أي مبلغ صرفته من جيبك الخاص لإنجاز مهمة، ليتم اعتماده واسترجاعه من قبل المحاسب.
      </p>

      {responsibleBookings.length > 0 && (
        <div className="mt-4 rounded-xl border border-emerald-200 bg-emerald-50 p-4">
          <h3 className="text-sm font-bold text-emerald-800">أنت المسؤول عن المصاريف في:</h3>
          <ul className="mt-2 list-inside list-disc text-sm text-emerald-700">
            {responsibleBookings.map(b => (
              <li key={b.id}>
                {b.code} — {b.customer?.name} ({b.service?.name || 'بدون خدمة'})
              </li>
            ))}
          </ul>
        </div>
      )}

      {!canExpense && activeBookings.length > 0 && (
        <div className="mt-4 rounded-xl border border-amber-200 bg-amber-50 p-4">
          <p className="text-sm text-amber-800">
            أنت مكلف بحجوزات لكن المسؤول عن المصاريف شخص آخر (تيم ليدر). إذا تحتاج تصرف مبلغ، راجع التيم ليدر.
          </p>
        </div>
      )}

      {canExpense && (
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
      )}

      <div className="mt-6 overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        <div className="overflow-x-auto">
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
    </div>
  )
}
