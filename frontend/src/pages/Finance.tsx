import { useEffect, useState } from 'react'
import { api, type Booking, type Expense } from '../api'

export default function Finance() {
  const [bookings, setBookings] = useState<Booking[]>([])
  const [expenses, setExpenses] = useState<Expense[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [expanded, setExpanded] = useState<Record<string, boolean>>({})
  const [filter, setFilter] = useState<'all' | 'pending' | 'verified'>('all')
  const [search, setSearch] = useState('')

  const load = () => {
    Promise.all([
      api.getBookings({ status: 'COMPLETED' }),
      api.getExpenses(),
    ])
      .then(([b, e]) => {
        setBookings(b)
        setExpenses(e)
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  // مبلغ الفاتورة الي يكتبه المحاسب لكل حجز — لازم للحجوزات القديمة
  // المستوردة بقيم صفر: يفتح فاتورة النظام القديم ويكتب سعرها هنا.
  const [invoiceAmounts, setInvoiceAmounts] = useState<Record<string, string>>({})
  const [auditBusy, setAuditBusy] = useState<string | null>(null)

  const amountFor = (b: Booking) => {
    const typed = invoiceAmounts[b.id]
    if (typed !== undefined && typed !== '') return Number(typed)
    return b.amountCollected ?? 0
  }

  const doAudit = async (b: Booking, action: 'VERIFY' | 'MISMATCH' | 'PRICE_ERROR') => {
    const typed = invoiceAmounts[b.id]
    const amount = typed !== undefined && typed !== '' ? Number(typed) : undefined

    if (action === 'VERIFY' && amountFor(b) <= 0) {
      alert('اكتب المبلغ من الفاتورة أول — ما ينفع تدقق حجز بلا مبلغ')
      return
    }
    let note: string | undefined
    if (action !== 'VERIFY') {
      const answer = prompt(action === 'MISMATCH'
        ? 'شنو الفرق بالضبط؟ (يروح للرقابة والجودة)'
        : 'شنو الغلط بالسعر؟ (يروح للرقابة والإداري)')
      if (answer === null) return
      note = answer.trim() || undefined
    }

    setAuditBusy(b.id)
    try {
      const res = await api.auditBooking(b.id, { action, amountCollected: amount, note })
      if (action === 'VERIFY') {
        setBookings((prev) => prev.map((x) => (x.id === b.id ? (res as Booking) : x)))
      } else {
        alert('انسجّل البلاغ وانوجّه للمعني — الحجز يبقى غير مدقق لحد ما ينحسم')
        if (amount !== undefined) {
          setBookings((prev) => prev.map((x) => (x.id === b.id ? { ...x, amountCollected: amount } : x)))
        }
      }
      setInvoiceAmounts((prev) => ({ ...prev, [b.id]: '' }))
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر تنفيذ التدقيق')
    } finally {
      setAuditBusy(null)
    }
  }

  const toggle = (id: string) =>
    setExpanded((prev) => ({ ...prev, [id]: !prev[id] }))

  // بحث بكود الحجز، كود الزبون، رقم هاتفه، أو اسمه
  const matchesSearch = (b: Booking) => {
    const q = search.trim().toLowerCase()
    if (!q) return true
    return [b.code, b.customer?.code, b.customer?.phone, b.customer?.name]
      .some((v) => (v || '').toString().toLowerCase().includes(q))
  }
  const filtered = bookings.filter((b) => {
    if (!matchesSearch(b)) return false
    if (filter === 'pending') return !b.amountVerified
    if (filter === 'verified') return b.amountVerified
    return true
  })

  const pendingCount = bookings.filter((b) => !b.amountVerified).length
  const verifiedCount = bookings.filter((b) => b.amountVerified).length

  // Get expenses for the expense-responsible employee of a booking
  const getBookingExpenses = (b: Booking): Expense[] => {
    if (!b.expenseResponsibleId) return []
    return expenses.filter(
      (e) => e.employeeId === b.expenseResponsibleId && e.status === 'APPROVED',
    )
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">تدقيق الحسابات</h2>
      <p className="mt-1 text-slate-500">
        مراجعة المبالغ المستلمة من الحجوزات المنجزة وتدقيقها.
      </p>

      {/* Stats bar */}
      <div className="mt-4 grid grid-cols-3 gap-3">
        <button
          onClick={() => setFilter('all')}
          className={`rounded-xl border p-3 text-center transition-all ${
            filter === 'all'
              ? 'border-brand-500 bg-brand-50 shadow-sm'
              : 'border-slate-200 bg-white'
          }`}
        >
          <p className="text-2xl font-bold text-brand-900">{bookings.length}</p>
          <p className="text-xs text-slate-500">الكل</p>
        </button>
        <button
          onClick={() => setFilter('pending')}
          className={`rounded-xl border p-3 text-center transition-all ${
            filter === 'pending'
              ? 'border-amber-400 bg-amber-50 shadow-sm'
              : 'border-slate-200 bg-white'
          }`}
        >
          <p className="text-2xl font-bold text-amber-600">{pendingCount}</p>
          <p className="text-xs text-slate-500">بانتظار التدقيق</p>
        </button>
        <button
          onClick={() => setFilter('verified')}
          className={`rounded-xl border p-3 text-center transition-all ${
            filter === 'verified'
              ? 'border-emerald-400 bg-emerald-50 shadow-sm'
              : 'border-slate-200 bg-white'
          }`}
        >
          <p className="text-2xl font-bold text-emerald-600">{verifiedCount}</p>
          <p className="text-xs text-slate-500">تم التدقيق</p>
        </button>
      </div>

      <div className="mt-6 rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="🔍 بحث بكود الحجز، كود الزبون، رقم الهاتف، أو اسم الزبون..."
          className="w-full rounded-lg border border-slate-300 px-4 py-2.5 outline-none focus:border-brand-500"
        />
        {search.trim() && (
          <p className="mt-2 text-xs text-slate-500">النتائج: {filtered.length} حجز</p>
        )}
      </div>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && (
        <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">
          تعذر الاتصال بالخادم: {error}
        </p>
      )}

      <div className="mt-6 flex flex-col gap-4">
        {filtered.map((b) => {
          const isOpen = expanded[b.id] ?? false
          const cartTotal = (b.cartItems ?? []).reduce(
            (sum, c) => sum + c.totalPrice,
            0,
          )
          const bookingExpenses = getBookingExpenses(b)
          const expensesTotal = bookingExpenses.reduce(
            (sum, e) => sum + e.amount,
            0,
          )
          const totalCollected =
            (b.amountCollected || 0) + (b.advancePaid || 0)
          const expectedTotal =
            (b.quotedPrice || 0) + cartTotal + expensesTotal
          const diff = totalCollected - expectedTotal

          return (
            <div
              key={b.id}
              className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]"
            >
              {/* Header - always visible */}
              <button
                onClick={() => toggle(b.id)}
                className="flex w-full items-center justify-between p-4 text-start transition-colors hover:bg-slate-50"
              >
                <div className="flex flex-wrap items-center gap-2">
                  <span className="font-mono text-sm font-semibold text-brand-600">
                    {b.code}
                  </span>
                  <span className="text-sm font-medium text-brand-800">
                    {b.customer?.name}
                  </span>
                  {b.service && (
                    <span className="rounded-md bg-slate-100 px-2 py-0.5 text-xs text-slate-600">
                      {b.service.name}
                    </span>
                  )}
                  {/* التاريخ وموعد الانتهاء ظاهرين بالسطر — الباقي داخل التفاصيل */}
                  <span className="rounded-md bg-blue-50 px-2 py-0.5 text-xs font-medium text-blue-700">
                    📅 الحجز: {b.scheduledAt ? new Date(b.scheduledAt).toLocaleDateString('ar-IQ') : new Date(b.createdAt).toLocaleDateString('ar-IQ')}
                  </span>
                  <span className={`rounded-md px-2 py-0.5 text-xs font-medium ${
                    b.completedAt ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-100 text-slate-500'
                  }`}>
                    🏁 الانتهاء: {b.completedAt ? new Date(b.completedAt).toLocaleDateString('ar-IQ') : 'لم ينتهِ بعد'}
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <span
                    className={`rounded-full px-3 py-1 text-xs font-bold ${
                      b.amountVerified
                        ? 'bg-emerald-100 text-emerald-700'
                        : 'bg-amber-100 text-amber-700'
                    }`}
                  >
                    {b.amountVerified ? 'تم التدقيق' : 'بانتظار التدقيق'}
                  </span>
                  <svg
                    className={`h-5 w-5 text-slate-400 transition-transform ${isOpen ? 'rotate-180' : ''}`}
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    strokeWidth={2}
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      d="M19 9l-7 7-7-7"
                    />
                  </svg>
                </div>
              </button>

              {/* Expanded details */}
              {isOpen && (
                <div className="border-t border-slate-100 p-4">
                  {/* Booking info */}
                  <div className="grid grid-cols-1 gap-2 text-sm sm:grid-cols-2">
                    <InfoRow label="رقم الزبون" value={b.customer ? `CUST-${String(b.customer.customerCode).padStart(5, '0')}` : 'زبون غير معروف'} />
                    <InfoRow label="هاتف الزبون" value={b.customer?.phone || '-'} />
                    <InfoRow
                      label="تاريخ الحجز"
                      value={new Date(b.createdAt).toLocaleDateString('ar-IQ', { year: 'numeric', month: 'long', day: 'numeric' })}
                    />
                    <InfoRow
                      label="موعد التنفيذ"
                      value={b.scheduledAt ? new Date(b.scheduledAt).toLocaleString('ar-IQ') : 'غير محدد'}
                    />
                    <InfoRow
                      label="تاريخ الإنجاز"
                      value={
                        b.completedAt
                          ? new Date(b.completedAt).toLocaleDateString('ar-IQ', {
                              year: 'numeric',
                              month: 'long',
                              day: 'numeric',
                            })
                          : 'غير محدد'
                      }
                    />
                    <InfoRow label="العنوان" value={b.address || 'غير محدد'} />
                    {b.completionNotes && (
                      <div className="col-span-full">
                        <InfoRow label="ملاحظات الفني" value={b.completionNotes} />
                      </div>
                    )}
                  </div>

                  {/* Cart Items */}
                  {(b.cartItems ?? []).length > 0 && (
                    <div className="mt-4">
                      <h4 className="mb-2 text-sm font-bold text-brand-900">
                        المواد المستخدمة
                      </h4>
                      <div className="overflow-hidden rounded-lg border border-slate-200">
                        <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                          <thead className="bg-slate-50 text-slate-600">
                            <tr>
                              <th className="p-2 text-start font-medium">المنتج</th>
                              <th className="p-2 text-center font-medium">الكمية</th>
                              <th className="p-2 text-center font-medium">سعر الوحدة</th>
                              <th className="p-2 text-center font-medium">المجموع</th>
                            </tr>
                          </thead>
                          <tbody>
                            {(b.cartItems ?? []).map((item) => (
                              <tr
                                key={item.id}
                                className="border-t border-slate-100"
                              >
                                <td className="p-2 text-slate-700">
                                  {item.productName}
                                  {item.notes && (
                                    <span className="mr-1 text-xs text-slate-400">
                                      ({item.notes})
                                    </span>
                                  )}
                                </td>
                                <td className="p-2 text-center text-slate-600">
                                  {item.quantity}
                                </td>
                                <td className="p-2 text-center text-slate-600">
                                  {item.unitPrice.toLocaleString()}
                                </td>
                                <td className="p-2 text-center font-medium text-slate-800">
                                  {item.totalPrice.toLocaleString()}
                                </td>
                              </tr>
                            ))}
                            <tr className="border-t-2 border-slate-300 bg-slate-50">
                              <td
                                colSpan={3}
                                className="p-2 text-start font-bold text-slate-700"
                              >
                                مجموع المواد
                              </td>
                              <td className="p-2 text-center font-bold text-brand-700">
                                {cartTotal.toLocaleString()}
                              </td>
                            </tr>
                          </tbody>
                        </table>
                        </div>
                      </div>
                    </div>
                  )}

                  {/* Leader Expenses */}
                  {bookingExpenses.length > 0 && (
                    <div className="mt-4">
                      <h4 className="mb-2 text-sm font-bold text-brand-900">
                        مصاريف الليدر
                        {b.expenseResponsible && (
                          <span className="mr-2 text-xs font-normal text-slate-500">
                            ({b.expenseResponsible.name})
                          </span>
                        )}
                      </h4>
                      <div className="overflow-hidden rounded-lg border border-slate-200">
                        <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                          <thead className="bg-slate-50 text-slate-600">
                            <tr>
                              <th className="p-2 text-start font-medium">الوصف</th>
                              <th className="p-2 text-center font-medium">المبلغ</th>
                            </tr>
                          </thead>
                          <tbody>
                            {bookingExpenses.map((exp) => (
                              <tr
                                key={exp.id}
                                className="border-t border-slate-100"
                              >
                                <td className="p-2 text-slate-700">
                                  {exp.description || 'بدون وصف'}
                                </td>
                                <td className="p-2 text-center text-slate-600">
                                  {exp.amount.toLocaleString()}
                                </td>
                              </tr>
                            ))}
                            <tr className="border-t-2 border-slate-300 bg-slate-50">
                              <td className="p-2 text-start font-bold text-slate-700">
                                مجموع المصاريف
                              </td>
                              <td className="p-2 text-center font-bold text-brand-700">
                                {expensesTotal.toLocaleString()}
                              </td>
                            </tr>
                          </tbody>
                        </table>
                        </div>
                      </div>
                    </div>
                  )}

                  {/* Cost Breakdown */}
                  <div className="mt-4 rounded-lg border border-brand-200 bg-brand-50/50 p-4">
                    <h4 className="mb-3 text-sm font-bold text-brand-900">
                      ملخص التكاليف
                    </h4>
                    <div className="space-y-2 text-sm">
                      <SummaryRow
                        label="التكلفة المقدرة (الإداري)"
                        value={b.quotedPrice}
                      />
                      {cartTotal > 0 && (
                        <SummaryRow
                          label="مجموع المواد المستخدمة"
                          value={cartTotal}
                        />
                      )}
                      {expensesTotal > 0 && (
                        <SummaryRow
                          label="مصاريف الليدر"
                          value={expensesTotal}
                        />
                      )}
                      <div className="border-t border-brand-200 pt-2">
                        <SummaryRow
                          label="الإجمالي المتوقع"
                          value={expectedTotal}
                          bold
                        />
                      </div>
                      <div className="border-t border-brand-200 pt-2">
                        <SummaryRow
                          label="الدفعة المقدمة"
                          value={b.advancePaid ?? 0}
                        />
                        <SummaryRow
                          label="المبلغ المستلم (الفني)"
                          value={b.amountCollected}
                        />
                        <SummaryRow
                          label="إجمالي المحصّل"
                          value={totalCollected}
                          bold
                        />
                      </div>
                      {expectedTotal > 0 && (
                        <div className="border-t border-brand-200 pt-2">
                          {diff === 0 ? (
                            <p className="text-sm font-medium text-emerald-600">
                              المبلغ مطابق للتكلفة المتوقعة
                            </p>
                          ) : diff > 0 ? (
                            <p className="text-sm font-bold text-emerald-600">
                              زيادة بمقدار {diff.toLocaleString()}
                            </p>
                          ) : (
                            <p className="text-sm font-bold text-red-600">
                              نقص بمقدار {Math.abs(diff).toLocaleString()}
                            </p>
                          )}
                        </div>
                      )}
                    </div>
                  </div>

                  {/* التدقيق: مبلغ الفاتورة إجباري، أو بلاغ خطأ ينوجّه للمعني */}
                  {!b.amountVerified && (
                    <div className="mt-4 rounded-xl border border-slate-200 bg-slate-50 p-4">
                      <label className="mb-1 block text-xs font-bold text-slate-600">
                        المبلغ حسب الفاتورة *
                      </label>
                      <input
                        type="number" min="0" inputMode="numeric"
                        value={invoiceAmounts[b.id] ?? ''}
                        onChange={(e) => setInvoiceAmounts((prev) => ({ ...prev, [b.id]: e.target.value }))}
                        placeholder={b.amountCollected ? String(b.amountCollected) : 'اكتب المبلغ من الفاتورة'}
                        className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                      />
                      {(b.amountCollected ?? 0) === 0 && (
                        <p className="mt-1 text-xs text-amber-700">
                          ⚠️ هذا الحجز بلا مبلغ (مستورد من النظام القديم) — افتح فاتورته واكتب سعرها.
                        </p>
                      )}

                      <div className="mt-3 grid grid-cols-1 gap-2 sm:grid-cols-3">
                        <button
                          disabled={auditBusy === b.id}
                          onClick={() => doAudit(b, 'VERIFY')}
                          className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-2.5 text-sm font-medium text-white shadow-md transition-all hover:shadow-lg disabled:opacity-50"
                        >
                          ✔ مطابق — أكّد التدقيق
                        </button>
                        <button
                          disabled={auditBusy === b.id}
                          onClick={() => doAudit(b, 'MISMATCH')}
                          className="rounded-lg bg-amber-500 px-4 py-2.5 text-sm font-medium text-white hover:bg-amber-600 disabled:opacity-50"
                        >
                          ⚠️ غير مطابق
                        </button>
                        <button
                          disabled={auditBusy === b.id}
                          onClick={() => doAudit(b, 'PRICE_ERROR')}
                          className="rounded-lg bg-red-600 px-4 py-2.5 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-50"
                        >
                          ✕ خطأ بالسعر
                        </button>
                      </div>
                      <p className="mt-2 text-[11px] text-slate-400">
                        «غير مطابق» يروح للرقابة والجودة · «خطأ بالسعر» يروح للرقابة والإداري
                      </p>
                    </div>
                  )}
                </div>
              )}
            </div>
          )
        })}
        {!loading && bookings.length === 0 && (
          <p className="text-slate-400">لا توجد حجوزات منجزة بعد.</p>
        )}
      </div>
    </div>
  )
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <p className="text-slate-600">
      <span className="text-slate-400">{label}: </span>
      {value}
    </p>
  )
}

function SummaryRow({
  label,
  value,
  bold,
}: {
  label: string
  value: number | null | undefined
  bold?: boolean
}) {
  return (
    <div className="flex items-center justify-between">
      <span className={`text-slate-600 ${bold ? 'font-bold' : ''}`}>
        {label}
      </span>
      <span
        className={`font-mono ${bold ? 'text-base font-bold text-brand-900' : 'text-slate-800'}`}
      >
        {value != null ? value.toLocaleString() : 'غير محدد'}
      </span>
    </div>
  )
}
