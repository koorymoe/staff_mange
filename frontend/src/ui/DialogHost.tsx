import { useEffect, useState } from 'react'
import { onDialog, pendingDialogs, type DialogReq } from './dialogs'

/** ينحط مرة وحدة بـ main.tsx — يعرض طابور askConfirm/askInput واحد واحد */
export function DialogHost() {
  const [queue, setQueue] = useState<DialogReq[]>(() => pendingDialogs.splice(0))
  useEffect(() => {
    onDialog((r) => setQueue((q) => [...q, r]))
    return () => onDialog(null)
  }, [])
  const cur = queue[0]
  if (!cur) return null
  return <Dialog key={cur.id} req={cur} done={() => setQueue((q) => q.slice(1))} />
}

function Dialog({ req, done }: { req: DialogReq; done: () => void }) {
  const [text, setText] = useState(req.kind === 'input' ? (req.defaultValue ?? '') : '')
  const close = (ok: boolean) => {
    if (req.kind === 'confirm') req.resolve(ok)
    else req.resolve(ok ? text : null)
    done()
  }
  return (
    <div dir="rtl" className="fixed inset-0 z-[200] flex items-center justify-center bg-black/40 p-4" onKeyDown={(e) => {
      if (e.key === 'Escape') close(false)
      if (e.key === 'Enter' && !(req.kind === 'input' && req.multiline)) close(true)
    }}>
      <div role="dialog" aria-modal="true" className="w-full max-w-md rounded-2xl bg-white p-5 shadow-2xl dark:bg-slate-800">
        <p className="whitespace-pre-wrap text-sm leading-7 text-slate-800 dark:text-slate-100">{req.message}</p>
        {req.kind === 'input' && (req.multiline
          ? <textarea autoFocus value={text} onChange={(e) => setText(e.target.value)} rows={3} className="mt-3 w-full rounded-lg border border-slate-300 p-2 text-sm dark:bg-slate-700" />
          : <input autoFocus value={text} onChange={(e) => setText(e.target.value)} className="mt-3 w-full rounded-lg border border-slate-300 p-2 text-sm dark:bg-slate-700" />)}
        <div className="mt-4 flex justify-start gap-2">
          <button autoFocus={req.kind === 'confirm'} onClick={() => close(true)}
            className={`rounded-lg px-4 py-2 text-sm font-bold text-white ${req.kind === 'confirm' && req.danger ? 'bg-red-600' : 'bg-[#0f2040]'}`}>
            {req.kind === 'confirm' ? (req.okLabel ?? 'تأكيد') : 'موافق'}
          </button>
          <button onClick={() => close(false)} className="rounded-lg border border-slate-300 px-4 py-2 text-sm font-bold text-slate-600 dark:text-slate-200">إلغاء</button>
        </div>
      </div>
    </div>
  )
}
