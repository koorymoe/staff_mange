import { useEffect, useState } from 'react'
import { api, type LeaveRequest, type LeaveStatus } from '../api'

/**
 * الإجازات.
 *
 * تبويبان: «طلباتي» لأي موظف، و«الموافقات» للي عنده صلاحية البت.
 * التوجيه يتحدد من شفت الموظف نفسه — طلبات الصباحي تروح لإداري الشفت
 * الصباحي، والمسائي لإداري المسائي. مدير النظام والمالك يوافقون على الكل.
 */

const STATUS_STYLE: Record<LeaveStatus, string> = {
  PENDING: 'bg-amber-100 text-amber-800',
  APPROVED: 'bg-emerald-100 text-emerald-800',
  REJECTED: 'bg-red-100 text-red-800',
  CANCELLED: 'bg-slate-200 text-slate-600',
}

const fmt = (d: string) => new Date(d).toLocaleDateString('ar-IQ', { year: 'numeric', month: 'long', day: 'numeric' })
const iso = (d: Date) => d.toISOString().slice(0, 10)

export default function LeavesPage() {
  const [tab, setTab] = useState<'mine' | 'inbox'>('mine')
  const [mine, setMine] = useState<LeaveRequest[]>([])
  const [inbox, setInbox] = useState<LeaveRequest[]>([])
  const [canApprove, setCanApprove] = useState(false)
  const [loading, setLoading] = useState(true)

  const [open, setOpen] = useState(false)
  const [start, setStart] = useState('')
  const [end, setEnd] = useState('')
  const [reason, setReason] = useState('')
  const [busy, setBusy] = useState(false)
  const [note, setNote] = useState('')
  const [noteFor, setNoteFor] = useState<string | null>(null)

  // نسأل pending-count أول — يرجع 200 للكل ويقول هل هذا الشخص مخوّل.
  // بهذا الشكل الموظف العادي ما يشوف 403 بالكونسول أصلاً.
  const load = () => {
    api.getMyLeaves().then(setMine).catch(() => setMine([]))
    api.getLeavePendingCount()
      .then(({ canApprove: ok }) => {
        setCanApprove(ok)
        if (!ok) { setInbox([]); return }
        return api.getLeaveInbox().then(setInbox)
      })
      .catch(() => setCanApprove(false))
      .finally(() => setLoading(false))
  }
  useEffect(load, [])

  // أقرب تاريخ مسموح: بعد يومين — نفس القيد الي يفرضه السيرفر.
  // يُحسب مرة وحدة عند أول رندر (مو بكل رندر) حتى ما يتغير تحت إيد المستخدم.
  const [minDate] = useState(() => iso(new Date(Date.now() + 2 * 24 * 60 * 60 * 1000)))

  const submit = async () => {
    if (!start) { alert('اختر تاريخ الإجازة'); return }
    setBusy(true)
    try {
      await api.createLeave({ startDate: start, endDate: end || undefined, reason: reason || null })
      setOpen(false); setStart(''); setEnd(''); setReason('')
      load()
      alert('انرفع طلب الإجازة — راح يوصلك إشعار بالقرار')
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر تقديم الطلب')
    } finally { setBusy(false) }
  }

  const decide = async (l: LeaveRequest, approve: boolean) => {
    if (!approve && noteFor !== l.id) { setNoteFor(l.id); setNote(''); return }
    setBusy(true)
    try {
      await api.decideLeave(l.id, approve, note || undefined)
      setNoteFor(null); setNote('')
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر تنفيذ القرار')
    } finally { setBusy(false) }
  }

  const cancel = async (l: LeaveRequest) => {
    if (!confirm('تسحب طلب الإجازة؟')) return
    try { await api.cancelLeave(l.id); load() }
    catch (e) { alert(e instanceof Error ? e.message : 'خطأ') }
  }

  const card = (l: LeaveRequest, forApprover: boolean) => (
    <div key={l.id} className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          {forApprover && (
            <p className="font-bold text-slate-800">
              {l.employeeName}
              {l.jobTitle && <span className="mr-2 text-xs font-normal text-slate-400">{l.jobTitle}</span>}
            </p>
          )}
          <p className={forApprover ? 'mt-1 text-sm text-slate-600' : 'font-bold text-slate-800'}>
            {fmt(l.startDate)}
            {l.endDate !== l.startDate && <> ← {fmt(l.endDate)}</>}
            <span className="mr-2 text-xs text-slate-400">({l.days} يوم)</span>
          </p>
          {l.reason && <p className="mt-1 text-sm text-slate-600">📝 {l.reason}</p>}
          <p className="mt-1 text-xs text-slate-400">{l.routeLabel}</p>
          {l.decidedByName && (
            <p className="mt-1 text-xs text-slate-500">
              البت: {l.decidedByName}
              {l.decisionNote && ` — ${l.decisionNote}`}
            </p>
          )}
        </div>
        <span className={`rounded-full px-3 py-1 text-xs font-bold ${STATUS_STYLE[l.status]}`}>{l.statusLabel}</span>
      </div>

      {forApprover && l.status === 'PENDING' && (
        <div className="mt-3">
          {noteFor === l.id && (
            <input
              value={note} onChange={(e) => setNote(e.target.value)} placeholder="سبب الرفض (اختياري)"
              className="mb-2 w-full rounded-lg border border-gray-300 px-3 py-2 text-right text-sm outline-none focus:border-brand-500"
            />
          )}
          <div className="flex gap-2">
            <button disabled={busy} onClick={() => decide(l, true)}
              className="rounded-lg bg-emerald-600 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-700 disabled:opacity-50">✔ موافقة</button>
            <button disabled={busy} onClick={() => decide(l, false)}
              className="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-50">
              {noteFor === l.id ? 'تأكيد الرفض' : '✕ رفض'}
            </button>
          </div>
        </div>
      )}
      {!forApprover && l.status === 'PENDING' && (
        <button onClick={() => cancel(l)}
          className="mt-3 rounded-lg border border-slate-300 px-4 py-2 text-sm font-medium text-slate-600 hover:bg-slate-50">
          سحب الطلب
        </button>
      )}
    </div>
  )

  if (loading) return <div dir="rtl" className="p-8 text-center text-slate-500">جاري التحميل...</div>

  const pending = inbox.filter((l) => l.status === 'PENDING')
  const list = tab === 'mine' ? mine : inbox

  return (
    <div dir="rtl" className="space-y-6">
      <div className="rounded-2xl p-6 shadow-sm" style={{ backgroundColor: '#1a3a5c' }}>
        <h1 className="text-2xl font-bold text-white">🏖️ الإجازات</h1>
        <p className="mt-1 text-sm text-blue-200">
          طلب الإجازة يُقدَّم قبل يومين على الأقل، ويروح لإداري شفتك — ويوصلك إشعار بالقرار.
        </p>
      </div>

      <div className="flex flex-wrap items-center gap-2">
        <button onClick={() => setTab('mine')}
          className={`rounded-xl px-5 py-2.5 text-sm font-bold ${tab === 'mine' ? 'text-white' : 'bg-white text-slate-600'}`}
          style={tab === 'mine' ? { backgroundColor: '#1a3a5c' } : undefined}>طلباتي ({mine.length})</button>
        {canApprove && (
          <button onClick={() => setTab('inbox')}
            className={`rounded-xl px-5 py-2.5 text-sm font-bold ${tab === 'inbox' ? 'text-white' : 'bg-white text-slate-600'}`}
            style={tab === 'inbox' ? { backgroundColor: '#1a3a5c' } : undefined}>
            الموافقات {pending.length > 0 && <span className="mr-1 rounded-full bg-red-500 px-2 py-0.5 text-xs text-white">{pending.length}</span>}
          </button>
        )}
        <button onClick={() => { setOpen(true); setStart(minDate) }}
          className="mr-auto rounded-xl bg-emerald-600 px-5 py-2.5 text-sm font-bold text-white hover:bg-emerald-700">
          ➕ طلب إجازة
        </button>
      </div>

      <div className="space-y-3">
        {list.map((l) => card(l, tab === 'inbox'))}
        {list.length === 0 && (
          <p className="rounded-2xl bg-white p-8 text-center text-slate-400 shadow-sm">
            {tab === 'mine' ? 'ما قدّمت ولا طلب إجازة' : 'ماكو طلبات تنتظر قرارك'}
          </p>
        )}
      </div>

      {open && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={() => setOpen(false)}>
          <div dir="rtl" className="w-full max-w-md rounded-2xl bg-white p-6 shadow-xl" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-lg font-bold" style={{ color: '#1a3a5c' }}>طلب إجازة</h3>
            <p className="mt-1 rounded-lg bg-slate-50 p-3 text-xs text-slate-600">
              الإجازة تُطلب قبل يومين على الأقل من تاريخها.
            </p>

            <label className="mt-4 mb-1 block text-sm font-medium text-slate-600">من تاريخ *</label>
            <input type="date" min={minDate} value={start} onChange={(e) => setStart(e.target.value)}
              className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />

            <label className="mt-3 mb-1 block text-sm font-medium text-slate-600">إلى تاريخ (اتركه فاضي لو يوم واحد)</label>
            <input type="date" min={start || minDate} value={end} onChange={(e) => setEnd(e.target.value)}
              className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />

            <label className="mt-3 mb-1 block text-sm font-medium text-slate-600">السبب</label>
            <textarea value={reason} onChange={(e) => setReason(e.target.value)} rows={3}
              className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />

            <div className="mt-4 flex gap-3">
              <button disabled={busy} onClick={submit}
                className="flex-1 rounded-lg px-4 py-3 font-medium text-white disabled:opacity-50" style={{ backgroundColor: '#1a3a5c' }}>
                {busy ? 'جاري الإرسال...' : 'قدّم الطلب'}
              </button>
              <button onClick={() => setOpen(false)} className="rounded-lg border border-slate-300 px-4 py-3 font-medium text-slate-700">إلغاء</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
