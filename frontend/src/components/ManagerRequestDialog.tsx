import { useEffect, useState } from 'react'
import { api, type ManagerRequest } from '../api'
import { PRIORITY, STATUS } from '../utils/managerRequest'

// ═══ 📨 رفع طلب للمدير — (ع) 11-04 ═══
// الطلب يوصل للمدير والمالك كإشعار ويبقى مسجّل لحد ما ينحسم. أرقام المخازن
// (اختيارية) يحسبها الخادم نفسه حتى الرقم الي يوصل المدير حقيقي.
export default function ManagerRequestDialog({ onClose, withInventory }: { onClose: () => void; withInventory: boolean }) {
  const [title, setTitle] = useState('')
  const [details, setDetails] = useState('')
  const [priority, setPriority] = useState<ManagerRequest['priority']>('NORMAL')
  const [attach, setAttach] = useState(withInventory)
  const [ctx, setCtx] = useState<string | null>(null)
  const [mine, setMine] = useState<ManagerRequest[]>([])
  const [busy, setBusy] = useState(false)
  const [msg, setMsg] = useState('')
  const [reload, setReload] = useState(0)

  useEffect(() => {
    let alive = true
    api.getMyManagerRequests().then((r) => { if (alive) setMine(r) }).catch(() => {})
    return () => { alive = false }
  }, [reload])
  useEffect(() => {
    if (!attach || !withInventory) return
    let alive = true
    api.getManagerRequestInventoryContext().then((r) => { if (alive) setCtx(r.context) }).catch(() => { if (alive) setCtx('') })
    return () => { alive = false }
  }, [attach, withInventory])

  const send = async () => {
    setBusy(true); setMsg('')
    try {
      await api.createManagerRequest({ title, details, priority, attachInventory: attach && withInventory })
      setTitle(''); setDetails(''); setPriority('NORMAL')
      setMsg('✅ انرفع الطلب ووصل إشعار للمدير.')
      setReload((n) => n + 1)
    } catch (e) { setMsg(e instanceof Error ? e.message : 'تعذر') }
    finally { setBusy(false) }
  }

  return (
    <div dir="rtl" className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={onClose}>
      <div className="max-h-[90vh] w-full max-w-lg space-y-3 overflow-y-auto rounded-2xl bg-white p-5 shadow-xl" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-bold text-brand-900">📨 رفع طلب للمدير</h3>
          <button type="button" onClick={onClose} className="text-slate-400 hover:text-slate-700">✕</button>
        </div>
        <input value={title} onChange={(e) => setTitle(e.target.value)} maxLength={200} placeholder="العنوان (مثلاً: نحتاج موافقة على شراء دريلات)" className="w-full rounded-lg border border-slate-200 p-2 text-sm" />
        <textarea value={details} onChange={(e) => setDetails(e.target.value)} rows={4} placeholder="شنو المطلوب بالضبط؟" className="w-full rounded-lg border border-slate-200 p-2 text-sm" />
        <div className="flex flex-wrap gap-2">
          {(Object.keys(PRIORITY) as ManagerRequest['priority'][]).map((p) => (
            <button key={p} type="button" onClick={() => setPriority(p)}
              className={`rounded-full px-3 py-1 text-xs font-bold ${priority === p ? 'ring-2 ring-brand-500 ' : ''}${PRIORITY[p].cls}`}>{PRIORITY[p].label}</button>
          ))}
        </div>
        {withInventory && (
          <div className="rounded-xl border border-slate-200 p-2">
            <label className="flex items-center gap-2 text-sm font-bold"><input type="checkbox" checked={attach} onChange={(e) => setAttach(e.target.checked)} /> أرفق أرقام المخازن</label>
            {attach && <pre className="mt-2 whitespace-pre-wrap font-sans text-xs text-slate-600">{ctx === null ? '…' : ctx || 'تعذر جلب الأرقام — الطلب ينرفع بدونها.'}</pre>}
          </div>
        )}
        {msg && <p className="text-sm font-bold text-slate-700">{msg}</p>}
        <button type="button" disabled={busy || title.trim().length < 3} onClick={() => void send()}
          className="w-full rounded-lg bg-brand-700 py-2 text-sm font-bold text-white disabled:opacity-50">{busy ? 'جاري…' : 'ارفع الطلب'}</button>

        {mine.length > 0 && (
          <div className="space-y-1.5 border-t border-slate-100 pt-3">
            <p className="text-sm font-bold text-slate-700">طلباتي للمدير</p>
            {mine.map((m) => (
              <div key={m.id} className="rounded-lg bg-slate-50 p-2 text-xs">
                <div className="flex flex-wrap items-center justify-between gap-1">
                  <b className="text-slate-800">{m.title}</b>
                  <span className={`rounded-full px-2 py-0.5 font-bold ${STATUS[m.status].cls}`}>{STATUS[m.status].label}</span>
                </div>
                <span className="text-slate-400">{new Date(m.createdAt).toLocaleDateString('en-GB')}{m.source === 'MATRIX' && ' · عن طريق ماتركس'}</span>
                {m.reply && <p className="mt-1 text-slate-700">رد المدير: {m.reply}</p>}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
