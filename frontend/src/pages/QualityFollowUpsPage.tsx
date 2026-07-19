import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type QualityFollowUp } from '../api'
import { useSession } from '../session'

const statusLabels: Record<QualityFollowUp['status'], string> = {
  PENDING: 'بانتظار التواصل',
  CONTACTED_OK: 'تم التواصل - كله تمام',
  CONTACTED_ISSUE: 'تم التواصل - اكو مشكلة',
  CONVERTED: 'تحول لحجز جديد',
  CLOSED: 'مغلقة',
}

const statusColors: Record<QualityFollowUp['status'], string> = {
  PENDING: 'bg-yellow-100 text-yellow-800',
  CONTACTED_OK: 'bg-green-100 text-green-800',
  CONTACTED_ISSUE: 'bg-red-100 text-red-800',
  CONVERTED: 'bg-blue-100 text-blue-800',
  CLOSED: 'bg-gray-100 text-gray-800',
}

export default function QualityFollowUpsPage() {
  const { employee } = useSession()
  // مهندس الجودة (والأدمن) يتواصلون مع الزبون مباشرة ويشوفون تفاصيله كاملة.
  // المراقب المدقق يشوف بس تقرير عام (كم متابعة قيد الانتظار وكم فيها مشكلة).
  const canContact = employee?.role === 'QUALITY_ENGINEER' || employee?.role === 'ADMIN'

  const [items, setItems] = useState<QualityFollowUp[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [notesDraft, setNotesDraft] = useState<Record<string, string>>({})

  const load = () => {
    setLoading(true)
    api.getQualityFollowUps()
      .then(setItems)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const handleUpdate = async (id: string, status: QualityFollowUp['status']) => {
    try {
      await api.updateQualityFollowUp(id, { status, contactNotes: notesDraft[id] })
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    }
  }

  if (!canContact) {
    const pending = items.filter((i) => i.status === 'PENDING').length
    const issues = items.filter((i) => i.status === 'CONTACTED_ISSUE').length
    return (
      <div>
        <h2 className="text-2xl font-bold text-brand-900">متابعة الجودة بعد الحجوزات</h2>
        <p className="mt-1 text-slate-500">تقرير عام — التواصل مع الزبائن مسؤولية مهندس الجودة.</p>
        {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
        {!loading && (
          <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div className="rounded-xl border border-white bg-white p-6 text-center shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              <p className="text-3xl font-bold text-amber-600">{pending}</p>
              <p className="mt-1 text-sm text-slate-500">بانتظار تواصل مهندس الجودة</p>
            </div>
            <div className="rounded-xl border border-white bg-white p-6 text-center shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              <p className="text-3xl font-bold text-red-600">{issues}</p>
              <p className="mt-1 text-sm text-slate-500">فيهن مشكلة تحتاج متابعة</p>
            </div>
          </div>
        )}
      </div>
    )
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">متابعة الجودة بعد الحجوزات</h2>
      <p className="mt-1 text-slate-500">تواصل مع الزبائن اللي اكتمل حجزهم، وتأكد ما اكو مشاكل.</p>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && (
        <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">تعذر الاتصال بالخادم: {error}</p>
      )}

      {!loading && !error && (
        <div className="mt-6 space-y-4">
          {items.map((item) => (
            <div key={item.id} className="rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div>
                  <p className="font-bold text-brand-900">
                    {item.customer.name} — {item.customer.phone}
                  </p>
                  <p className="text-sm text-slate-500">
                    الحجز: {item.booking.code} — {item.booking.service?.name || 'بدون خدمة محددة'}
                  </p>
                </div>
                <span className={`rounded-full px-3 py-1 text-xs font-bold ${statusColors[item.status]}`}>
                  {statusLabels[item.status]}
                </span>
              </div>

              {item.status === 'PENDING' && (
                <div className="mt-4 space-y-3">
                  <textarea
                    placeholder="ملاحظات التواصل مع الزبون..."
                    value={notesDraft[item.id] || ''}
                    onChange={(e) => setNotesDraft({ ...notesDraft, [item.id]: e.target.value })}
                    rows={2}
                    className="w-full rounded-lg border border-gray-300 px-4 py-2 text-right outline-none focus:border-brand-500"
                  />
                  <div className="flex flex-wrap gap-2">
                    <button
                      onClick={() => handleUpdate(item.id, 'CONTACTED_OK')}
                      className="rounded-lg bg-green-100 px-4 py-2 text-sm font-medium text-green-700 hover:bg-green-200"
                    >
                      تواصلت - كله تمام
                    </button>
                    <button
                      onClick={() => handleUpdate(item.id, 'CONTACTED_ISSUE')}
                      className="rounded-lg bg-red-100 px-4 py-2 text-sm font-medium text-red-700 hover:bg-red-200"
                    >
                      تواصلت - اكو مشكلة
                    </button>
                  </div>
                </div>
              )}

              {item.status === 'CONTACTED_ISSUE' && (
                <div className="mt-4 flex flex-wrap items-center gap-2">
                  <Link
                    to={`/sales?customerId=${item.customer.id}`}
                    onClick={() => handleUpdate(item.id, 'CONVERTED')}
                    className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-2 text-sm font-medium text-white"
                  >
                    تحويل لحجز جديد
                  </Link>
                  <button
                    onClick={() => handleUpdate(item.id, 'CLOSED')}
                    className="rounded-lg border border-slate-300 px-4 py-2 text-sm font-medium text-slate-600"
                  >
                    إغلاق بدون حجز
                  </button>
                </div>
              )}

              {item.contactNotes && (
                <p className="mt-3 rounded-lg bg-slate-50 px-3 py-2 text-sm text-slate-600">
                  {item.contactNotes}
                </p>
              )}
            </div>
          ))}
          {items.length === 0 && (
            <p className="rounded-xl border border-white bg-white p-6 text-center text-slate-400 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              لا توجد متابعات جودة بعد
            </p>
          )}
        </div>
      )}
    </div>
  )
}
