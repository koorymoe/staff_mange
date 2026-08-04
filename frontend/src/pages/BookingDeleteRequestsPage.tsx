import { useEffect, useState } from 'react'
import { api, type BookingDeleteRequest } from '../api'

/**
 * طلبات حذف الحجوزات.
 *
 * الإداري يطلب من صفحة الحجوزات، والمراقب أو مدير النظام يبت هنا.
 * الحذف ما يترد، فما يصير بضغطة زر وحدة من شخص واحد.
 */

const STATUS_STYLE: Record<string, string> = {
  PENDING: 'bg-amber-100 text-amber-800',
  APPROVED: 'bg-red-100 text-red-800',
  REJECTED: 'bg-slate-200 text-slate-600',
}

export default function BookingDeleteRequestsPage() {
  const [rows, setRows] = useState<BookingDeleteRequest[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState<string | null>(null)
  const [noteFor, setNoteFor] = useState<string | null>(null)
  const [note, setNote] = useState('')

  const load = () => {
    api.getBookingDeleteRequests()
      .then(setRows)
      .catch((e) => setError(e instanceof Error ? e.message : 'تعذر جلب الطلبات'))
      .finally(() => setLoading(false))
  }
  useEffect(load, [])

  const decide = async (r: BookingDeleteRequest, approve: boolean) => {
    if (approve && !confirm(`تأكيد حذف الحجز ${r.bookingCode}؟ الحذف ما يترد.`)) return
    if (!approve && noteFor !== r.id) { setNoteFor(r.id); setNote(''); return }
    setBusy(r.id)
    try {
      await api.decideBookingDelete(r.id, approve, note || undefined)
      setNoteFor(null); setNote('')
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر تنفيذ القرار')
    } finally {
      setBusy(null)
    }
  }

  const pending = rows.filter((r) => r.status === 'PENDING')

  return (
    <div dir="rtl" className="space-y-6">
      <div className="rounded-2xl p-6 shadow-sm" style={{ backgroundColor: '#1a3a5c' }}>
        <h1 className="text-2xl font-bold text-white">🗑️ طلبات حذف الحجوزات</h1>
        <p className="mt-1 text-sm text-blue-200">
          الحجوزات التجريبية والملغاة تنشال بموافقتك — الموافقة تحذف الحجز نهائياً وما ترجع.
        </p>
      </div>

      {loading && <p className="text-center text-slate-400">جاري التحميل...</p>}
      {error && <p className="rounded-lg bg-red-50 p-4 text-red-600">{error}</p>}

      {!loading && !error && (
        <div className="space-y-3">
          {pending.length === 0 && (
            <p className="rounded-2xl bg-white p-8 text-center text-slate-400 shadow-sm">
              ماكو طلبات تنتظر قرارك
            </p>
          )}

          {rows.map((r) => (
            <div key={r.id} className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <p className="font-bold text-slate-800">
                    {r.bookingCode}
                    <span className="mr-2 text-sm font-normal text-slate-500">{r.customerName}</span>
                  </p>
                  <p className="mt-1 text-sm text-slate-600">📝 {r.reason}</p>
                  <p className="mt-1 text-xs text-slate-400">
                    طلبه: {r.requestedByName} · حالة الحجز: {r.bookingStatus}
                    {r.decidedByName && ` · البت: ${r.decidedByName}`}
                    {r.decisionNote && ` — ${r.decisionNote}`}
                  </p>
                </div>
                <span className={`rounded-full px-3 py-1 text-xs font-bold ${STATUS_STYLE[r.status]}`}>
                  {r.statusLabel}
                </span>
              </div>

              {r.status === 'PENDING' && (
                <div className="mt-3">
                  {noteFor === r.id && (
                    <input
                      value={note} onChange={(e) => setNote(e.target.value)} placeholder="سبب الرفض (اختياري)"
                      className="mb-2 w-full rounded-lg border border-gray-300 px-3 py-2 text-right text-sm outline-none focus:border-brand-500"
                    />
                  )}
                  <div className="flex gap-2">
                    <button disabled={busy === r.id} onClick={() => decide(r, true)}
                      className="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-50">
                      🗑️ وافق واحذف
                    </button>
                    <button disabled={busy === r.id} onClick={() => decide(r, false)}
                      className="rounded-lg border border-slate-300 px-4 py-2 text-sm font-medium text-slate-600 hover:bg-slate-50 disabled:opacity-50">
                      {noteFor === r.id ? 'تأكيد الرفض' : '✕ ارفض الطلب'}
                    </button>
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
