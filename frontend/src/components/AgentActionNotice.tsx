import { useEffect, useState } from 'react'
import { useSession } from '../session'
import { waApi, allowedWaAgents, type AiAgentAction } from '../waApi'

export default function AgentActionNotice({
  targetId,
  actions,
  onChanged,
}: {
  targetId?: string
  actions?: AiAgentAction[]
  onChanged?: () => void
}) {
  const { employee, permissions } = useSession()
  const allowed = allowedWaAgents(
    employee?.role,
    employee?.actualRole,
    permissions,
  )
  const [rows, setRows] = useState<AiAgentAction[]>([])
  const [updates, setUpdates] = useState<Record<string, AiAgentAction>>({})
  const [selected, setSelected] = useState<AiAgentAction | null>(null)
  const [reason, setReason] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [reload, setReload] = useState(0)
  const scope = allowed.join(',')
  useEffect(() => {
    if (!scope || !targetId || actions) return
    let alive = true
    waApi
      .actions(targetId)
      .then((x) => {
        if (alive) setRows(x)
      })
      .catch((e) => {
        if (alive) setError(e.message)
      })
    return () => {
      alive = false
    }
  }, [scope, targetId, actions, reload])
  if (!scope) return null
  const visible = (actions ?? rows)
    .map((x) => updates[x.id] ?? x)
    .filter((x) => allowed.includes(x.agent))
  const undo = async () => {
    if (!selected || busy || !reason.trim()) return
    setBusy(true)
    setError('')
    try {
      const updated = await waApi.undo(selected.id, reason.trim())
      setUpdates((v) => ({ ...v, [updated.id]: updated }))
      setRows((v) => v.map((x) => (x.id === updated.id ? updated : x)))
      setSelected(null)
      setReload((n) => n + 1)
      onChanged?.()
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setBusy(false)
    }
  }
  return (
    <div className="my-3 space-y-2">
      {error && !selected && (
        <p role="alert" className="text-xs text-red-700">
          تعذر جلب إجراءات الوكيل: {error}
        </p>
      )}
      {visible.map((action) => (
        <div
          key={action.id}
          className="rounded-xl border border-sky-200 bg-sky-50 p-3"
        >
          <div className="flex flex-wrap items-center justify-between gap-2">
            <p className="text-sm font-bold text-sky-800">
              🤖 سواه {action.agentName}
              {action.sandbox ? ' · تجربة' : ''}
              {action.undoneAt ? ' · انلغى الإجراء' : ''}
            </p>
            {!action.undoneAt && action.undoable && (
              <button
                className="rounded-lg border border-red-200 px-3 py-2 text-xs font-bold text-red-700"
                onClick={() => {
                  setSelected(action)
                  setReason('')
                  setError('')
                }}
              >
                إلغاء الإجراء
              </button>
            )}
          </div>
          <p className="mt-1 text-sm text-slate-700">{action.summary}</p>
          <p className="mt-1 text-xs text-slate-500">السبب: {action.reason}</p>
        </div>
      ))}
      {selected && (
        <div className="fixed inset-0 z-[80] overflow-y-auto bg-black/40 p-4">
          <form
            role="dialog"
            aria-modal="true"
            aria-label="إلغاء إجراء الوكيل"
            className="mx-auto mt-12 max-w-md rounded-2xl bg-white p-5"
            onSubmit={(e) => {
              e.preventDefault()
              void undo()
            }}
          >
            <h3 className="text-lg font-bold text-slate-800">
              إلغاء إجراء الوكيل؟
            </h3>
            <p className="mt-2 text-sm text-slate-500">{selected.summary}</p>
            {selected.kind === 'CREATE_BOOKING' && (
              <p className="mt-3 text-sm text-amber-700">
                الإلغاء يرفع طلب حذف عادي للحجز، مو حذف مباشر.
              </p>
            )}
            <label className="mt-4 block text-sm text-slate-600">
              ليش تريد تلغيه؟
              <textarea
                required
                className="inp mt-1 text-base"
                value={reason}
                onChange={(e) => setReason(e.target.value)}
              />
            </label>
            {error && (
              <p role="alert" className="mt-2 text-red-700">
                {error}
              </p>
            )}
            <div className="mt-4 flex gap-2">
              <button
                disabled={busy || !reason.trim()}
                className="rounded-xl bg-red-600 px-4 py-2 font-bold text-white disabled:opacity-50"
              >
                {busy ? 'دا ألغي…' : 'تأكيد الإلغاء'}
              </button>
              <button
                type="button"
                disabled={busy}
                className="rounded-xl border border-slate-200 px-4 py-2 text-slate-700"
                onClick={() => setSelected(null)}
              >
                رجوع
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  )
}
