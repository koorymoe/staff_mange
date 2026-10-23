import { useEffect, useState } from 'react'
import { apiRequest } from '../api'

type Meeting = {
  id: string
  kind: 'MEETING' | 'GUEST'
  reason: string
  guestName: string | null
  preferredAt: string | null
  status: 'PENDING' | 'SCHEDULED' | 'DECLINED'
  scheduledAt: string | null
  note: string | null
  createdAt: string
}
export default function MeetingRequests() {
  const [open, setOpen] = useState(false)
  const [rows, setRows] = useState<Meeting[]>([])
  const [kind, setKind] = useState<'MEETING' | 'GUEST'>('MEETING')
  const [reason, setReason] = useState('')
  const [guestName, setGuestName] = useState('')
  const [preferredAt, setPreferredAt] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [reload, setReload] = useState(0)
  useEffect(() => {
    if (!open) return
    let alive = true
    apiRequest<Meeting[]>('/meeting-requests/mine')
      .then((r) => {
        if (alive) setRows(r)
      })
      .catch((e) => {
        if (alive) setError(e.message)
      })
    return () => {
      alive = false
    }
  }, [open, reload])
  const submit = async () => {
    if (busy) return
    setBusy(true)
    setError('')
    try {
      await apiRequest('/meeting-requests', {
        method: 'POST',
        body: JSON.stringify({
          kind,
          reason: reason.trim(),
          ...(kind === 'GUEST' ? { guestName: guestName.trim() } : {}),
          ...(preferredAt ? { preferredAt } : {}),
        }),
      })
      setReason('')
      setGuestName('')
      setPreferredAt('')
      setReload((n) => n + 1)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setBusy(false)
    }
  }
  return (
    <section className="rounded-2xl border border-slate-200 bg-white p-4">
      <button
        className="text-base font-bold text-brand-600"
        onClick={() => setOpen((v) => !v)}
        aria-expanded={open}
      >
        ◷ اطلب مقابلة المدير {open ? '−' : '+'}
      </button>
      {open && (
        <div className="mt-4">
          <form
            className="space-y-3"
            onSubmit={(e) => {
              e.preventDefault()
              void submit()
            }}
          >
            <label className="block">
              نوع الطلب
              <select
                className="inp mt-1 text-base"
                value={kind}
                onChange={(e) => setKind(e.target.value as typeof kind)}
              >
                <option value="MEETING">مقابلة المدير</option>
                <option value="GUEST">جايب ضيف</option>
              </select>
            </label>
            <label className="block">
              السبب
              <textarea
                required
                className="inp mt-1 text-base"
                value={reason}
                onChange={(e) => setReason(e.target.value)}
              />
            </label>
            {kind === 'GUEST' && (
              <label className="block">
                اسم الضيف
                <input
                  required
                  className="inp mt-1 text-base"
                  value={guestName}
                  onChange={(e) => setGuestName(e.target.value)}
                />
              </label>
            )}
            <label className="block">
              وقت يناسبك · اختياري · بغداد
              <input
                className="inp mt-1 text-base"
                type="datetime-local"
                value={preferredAt}
                onChange={(e) => setPreferredAt(e.target.value)}
              />
            </label>
            <button
              disabled={
                busy ||
                !reason.trim() ||
                (kind === 'GUEST' && !guestName.trim())
              }
              className="rounded-xl bg-brand-600 px-4 py-3 font-bold text-white disabled:opacity-50"
            >
              {busy ? 'دا أرسل…' : 'إرسال الطلب'}
            </button>
          </form>
          {error && (
            <p role="alert" className="mt-3 text-red-600">
              {error}
            </p>
          )}
          <h3 className="mt-5 font-bold">طلباتي</h3>
          {!rows.length && (
            <p className="mt-2 text-sm text-slate-500">بعدك ما عندك طلبات.</p>
          )}
          {rows.map((row) => (
            <article
              className="mt-3 rounded-xl border border-slate-200 p-3"
              key={row.id}
            >
              <b>
                {row.kind === 'GUEST'
                  ? `ضيف: ${row.guestName}`
                  : 'مقابلة المدير'}
              </b>
              <p>{row.reason}</p>
              <p className="mt-1 text-sm font-bold">
                {row.status === 'PENDING'
                  ? 'ينتظر'
                  : row.status === 'SCHEDULED'
                    ? `انحدد موعد: ${row.scheduledAt?.replace('T', ' ')}`
                    : `انرفض: ${row.note || 'بدون ملاحظة'}`}
              </p>
              {row.status === 'SCHEDULED' && row.note && (
                <p className="text-sm text-slate-500">{row.note}</p>
              )}
            </article>
          ))}
        </div>
      )}
    </section>
  )
}
