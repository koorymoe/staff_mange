import { useEffect, useState } from 'react'
import { api, type Expense } from '../api'

export default function ExpensesReview() {
  const [expenses, setExpenses] = useState<Expense[]>([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState<'all' | 'PENDING' | 'APPROVED' | 'REJECTED'>('all')
  const [actionLoading, setActionLoading] = useState<string | null>(null)

  useEffect(() => {
    api.getExpenses()
      .then(setExpenses)
      .finally(() => setLoading(false))
  }, [])

  const handleAction = async (id: string, status: 'APPROVED' | 'REJECTED') => {
    setActionLoading(id)
    try {
      const updated = await api.updateExpenseStatus(id, status)
      setExpenses(prev => prev.map(e => e.id === updated.id ? updated : e))
    } finally {
      setActionLoading(null)
    }
  }

  const filtered = filter === 'all' ? expenses : expenses.filter(e => e.status === filter)
  const pendingCount = expenses.filter(e => e.status === 'PENDING').length
  const approvedCount = expenses.filter(e => e.status === 'APPROVED').length
  const rejectedCount = expenses.filter(e => e.status === 'REJECTED').length
  const approvedTotal = expenses.filter(e => e.status === 'APPROVED').reduce((s, e) => s + e.amount, 0)
  const pendingTotal = expenses.filter(e => e.status === 'PENDING').reduce((s, e) => s + e.amount, 0)

  const statusLabel: Record<string, string> = { PENDING: 'بانتظار الموافقة', APPROVED: 'معتمد', REJECTED: 'مرفوض' }
  const statusStyle: Record<string, string> = {
    PENDING: 'bg-amber-100 text-amber-700',
    APPROVED: 'bg-emerald-100 text-emerald-700',
    REJECTED: 'bg-red-100 text-red-700',
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">إدارة المصاريف</h2>
      <p className="mt-1 text-slate-500">مراجعة واعتماد مصاريف الموظفين الميدانية.</p>

      {/* Stats */}
      <div className="mt-4 grid grid-cols-2 gap-3 sm:grid-cols-4">
        <button onClick={() => setFilter('all')}
          className={`rounded-xl border p-3 text-center transition-all ${filter === 'all' ? 'border-brand-500 bg-brand-50 shadow-sm' : 'border-slate-200 bg-white'}`}>
          <p className="text-2xl font-bold text-brand-900">{expenses.length}</p>
          <p className="text-xs text-slate-500">الكل</p>
        </button>
        <button onClick={() => setFilter('PENDING')}
          className={`rounded-xl border p-3 text-center transition-all ${filter === 'PENDING' ? 'border-amber-400 bg-amber-50 shadow-sm' : 'border-slate-200 bg-white'}`}>
          <p className="text-2xl font-bold text-amber-600">{pendingCount}</p>
          <p className="text-xs text-slate-500">بانتظار الموافقة</p>
        </button>
        <button onClick={() => setFilter('APPROVED')}
          className={`rounded-xl border p-3 text-center transition-all ${filter === 'APPROVED' ? 'border-emerald-400 bg-emerald-50 shadow-sm' : 'border-slate-200 bg-white'}`}>
          <p className="text-2xl font-bold text-emerald-600">{approvedCount}</p>
          <p className="text-xs text-slate-500">معتمدة</p>
        </button>
        <button onClick={() => setFilter('REJECTED')}
          className={`rounded-xl border p-3 text-center transition-all ${filter === 'REJECTED' ? 'border-red-400 bg-red-50 shadow-sm' : 'border-slate-200 bg-white'}`}>
          <p className="text-2xl font-bold text-red-600">{rejectedCount}</p>
          <p className="text-xs text-slate-500">مرفوضة</p>
        </button>
      </div>

      {/* Totals summary */}
      <div className="mt-3 grid grid-cols-2 gap-3">
        <div className="rounded-xl bg-emerald-50 p-3 text-center">
          <p className="text-lg font-bold text-emerald-700">{approvedTotal.toLocaleString()} <span className="text-xs font-normal">د.ع</span></p>
          <p className="text-[10px] text-emerald-500">إجمالي المعتمد</p>
        </div>
        <div className="rounded-xl bg-amber-50 p-3 text-center">
          <p className="text-lg font-bold text-amber-700">{pendingTotal.toLocaleString()} <span className="text-xs font-normal">د.ع</span></p>
          <p className="text-[10px] text-amber-500">إجمالي المعلق</p>
        </div>
      </div>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}

      {/* Expenses list */}
      <div className="mt-6 flex flex-col gap-3">
        {filtered.map(exp => (
          <div key={exp.id} className="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
            <div className="flex items-center justify-between p-4">
              <div className="flex items-center gap-2">
                {exp.status === 'PENDING' ? (
                  <>
                    <button onClick={() => handleAction(exp.id, 'REJECTED')} disabled={actionLoading === exp.id}
                      className="rounded-lg border border-red-200 bg-red-50 px-4 py-1.5 text-xs font-bold text-red-600 transition hover:bg-red-100 disabled:opacity-50">
                      رفض
                    </button>
                    <button onClick={() => handleAction(exp.id, 'APPROVED')} disabled={actionLoading === exp.id}
                      className="rounded-lg bg-gradient-to-l from-emerald-500 to-emerald-600 px-4 py-1.5 text-xs font-bold text-white shadow-sm transition hover:shadow-md disabled:opacity-50">
                      {actionLoading === exp.id ? 'جاري...' : 'اعتماد'}
                    </button>
                  </>
                ) : (
                  <span className={`rounded-full px-3 py-1 text-xs font-bold ${statusStyle[exp.status]}`}>
                    {statusLabel[exp.status]}
                  </span>
                )}
              </div>
              <div className="flex items-center gap-4 text-right">
                <div>
                  <p className="text-sm font-bold text-brand-900">{exp.employee.name}</p>
                  {exp.employee.position && <p className="text-[10px] text-slate-400">{exp.employee.position}</p>}
                  {/* ⚠️ **كود الحجز جنب الاسم**: «مصروف ٥٠ ألف لفلان»
                      بلا حجز يُعتمد بالثقة مو بالمراجعة — والي يعتمد
                      لازم يعرف على أي شغل ينصرف. */}
                  {exp.bookingCode ? (
                    <p className="mt-0.5 rounded bg-brand-50 px-1.5 py-0.5 text-[10px] font-bold text-brand-700">
                      🗂️ {exp.bookingCode}
                    </p>
                  ) : (
                    <p className="mt-0.5 text-[10px] text-amber-600">مصروف عام بلا حجز</p>
                  )}
                </div>
                <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-brand-50 text-sm font-black text-brand-600">
                  {exp.employee.name.charAt(0)}
                </div>
              </div>
            </div>
            <div className="border-t border-slate-100 bg-slate-50/50 px-4 py-3">
              <div className="flex items-center justify-between">
                <span className="text-xs text-slate-400">
                  {new Date(exp.createdAt).toLocaleDateString('ar-IQ', { year: 'numeric', month: 'long', day: 'numeric' })}
                </span>
                <div className="flex items-center gap-4 text-right">
                  {exp.description && <span className="text-sm text-slate-600">{exp.description}</span>}
                  <span className="text-lg font-black text-brand-900">{exp.amount.toLocaleString()} <span className="text-xs font-normal text-slate-400">د.ع</span></span>
                </div>
              </div>
            </div>
          </div>
        ))}
        {!loading && filtered.length === 0 && (
          <p className="py-8 text-center text-slate-400">لا توجد مصاريف</p>
        )}
      </div>
    </div>
  )
}
