import { useEffect, useState } from 'react'
import { api, type ManagerRequest } from '../api'
import { PRIORITY, STATUS } from '../utils/managerRequest'

// ═══ 📨 الطلبات المرفوعة للمدير — (ع) 11-04 ═══
// المفتوح أول والعاجل فوق. الحسم يوصل للموظف كإشعار، والرفض لازم بسبب.
function Row({ m, onDone }: { m: ManagerRequest; onDone: () => void }) {
  const [reply, setReply] = useState('')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const act = async (status: 'DONE' | 'REJECTED') => {
    setBusy(true); setErr('')
    try { await api.handleManagerRequest(m.id, status, reply); onDone() }
    catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') }
    finally { setBusy(false) }
  }
  return (
    <div className={`space-y-2 rounded-2xl border bg-white p-4 ${m.status === 'OPEN' && m.priority === 'URGENT' ? 'border-red-300' : 'border-slate-200'}`}>
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span className="font-bold text-slate-800">{m.title}</span>
        <span className="flex gap-1 text-xs">
          <span className={`rounded-full px-2 py-0.5 font-bold ${PRIORITY[m.priority].cls}`}>{PRIORITY[m.priority].label}</span>
          <span className={`rounded-full px-2 py-0.5 font-bold ${STATUS[m.status].cls}`}>{STATUS[m.status].label}</span>
        </span>
      </div>
      <p className="text-xs text-slate-500">{m.requesterName} · {new Date(m.createdAt).toLocaleString('en-GB', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' })}{m.source === 'MATRIX' && ' · عن طريق ماتركس'}</p>
      {m.details && <p className="whitespace-pre-wrap text-sm text-slate-700">{m.details}</p>}
      {m.context && <pre className="whitespace-pre-wrap rounded-lg bg-slate-50 p-2 font-sans text-xs text-slate-600">📊 {m.context}</pre>}
      {m.status === 'OPEN' ? (
        <div className="flex flex-wrap gap-2">
          <input value={reply} onChange={(e) => setReply(e.target.value)} placeholder="ردّك للموظف (إجباري للرفض)" className="min-w-0 flex-1 rounded-lg border border-slate-200 px-2 py-1 text-sm" />
          <button type="button" disabled={busy} onClick={() => void act('DONE')} className="rounded-lg bg-emerald-700 px-3 py-1 text-xs font-bold text-white disabled:opacity-50">✅ انعمل</button>
          <button type="button" disabled={busy} onClick={() => void act('REJECTED')} className="rounded-lg bg-red-600 px-3 py-1 text-xs font-bold text-white disabled:opacity-50">❌ رفض</button>
        </div>
      ) : (
        <p className="text-xs text-slate-500">{m.handledByName && `حسمه ${m.handledByName}`}{m.reply && ` — ${m.reply}`}</p>
      )}
      {err && <p className="text-xs text-red-600">{err}</p>}
    </div>
  )
}

export default function ManagerRequestsPage() {
  const [rows, setRows] = useState<ManagerRequest[] | null>(null)
  const [err, setErr] = useState('')
  const [reload, setReload] = useState(0)
  useEffect(() => {
    let alive = true
    api.getManagerRequests()
      .then((r) => { if (alive) setRows(r) })
      .catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر') })
    return () => { alive = false }
  }, [reload])
  const open = rows?.filter((r) => r.status === 'OPEN').length ?? 0
  return (
    <div dir="rtl" className="space-y-4">
      <div>
        <h2 className="text-2xl font-bold text-brand-900">📨 طلبات للمدير</h2>
        <p className="text-sm text-slate-500">طلبات رفعها الموظفين (من الشاشة أو عن طريق ماتركس) ويا أرقامها. {rows && `مفتوح: ${open}`}</p>
      </div>
      {err && <p className="text-sm text-red-600">{err}</p>}
      {rows === null && !err && <p className="text-slate-400">جاري التحميل…</p>}
      {rows && rows.length === 0 && <p className="rounded-xl bg-white p-8 text-center text-slate-400">ماكو طلبات.</p>}
      {rows?.map((m) => <Row key={m.id} m={m} onDone={() => setReload((n) => n + 1)} />)}
    </div>
  )
}
