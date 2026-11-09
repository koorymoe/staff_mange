import { useEffect, useState } from 'react'
import { api, type ProjectDelay } from '../api'

// ═══ شريط «تجاوز حد المرحلة — اكتب سبب التأخير» (قرار (ع) 10-10) ═══
// طلب واحد للصفحة كلها (مو لكل بطاقة)، ويتحدّث بعد ما ينكتب سبب.
let cache: Promise<ProjectDelay[]> | null = null
const listeners = new Set<() => void>()
const load = () => (cache ??= api.getMyProjectDelays().catch(() => []))
const refresh = () => { cache = null; listeners.forEach((f) => f()) }

export default function ProjectDelayBanner({ projectId }: { projectId: string }) {
  const [d, setD] = useState<ProjectDelay | null>(null)
  const [tick, setTick] = useState(0)
  const [text, setText] = useState('')
  const [editing, setEditing] = useState(false)
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    const f = () => setTick((t) => t + 1)
    listeners.add(f)
    return () => { listeners.delete(f) }
  }, [])
  useEffect(() => {
    let alive = true
    load().then((rows) => { if (alive) setD(rows.find((r) => r.id === projectId) ?? null) })
    return () => { alive = false }
  }, [projectId, tick])

  if (!d) return null
  const save = async () => {
    setBusy(true)
    try { await api.addProjectDelayReason(projectId, text.trim()); setText(''); setEditing(false); refresh() }
    catch (e) { alert(e instanceof Error ? e.message : 'تعذر الحفظ') } finally { setBusy(false) }
  }
  const over = d.daysInStage - d.stageLimit
  return (
    <div className="mt-3 rounded-xl border border-red-200 bg-red-50 p-3 text-sm">
      <p className="font-extrabold text-red-800">
        ⏱️ صارله {d.daysInStage} يوم بمرحلة «{d.stage}» والحد {d.stageLimit} — متجاوز {over} يوم
        {!d.workedToday && <span className="mr-2 rounded bg-red-200 px-1.5 text-xs">ما اشتغل عليه أحد اليوم</span>}
      </p>
      {d.delayReason && !editing ? (
        <p className="mt-1 text-xs text-red-900">
          السبب: <b>{d.delayReason}</b> <span className="text-red-700">— {d.delayReasonBy ?? ''} {d.delayReasonAt ? new Date(d.delayReasonAt).toLocaleString('ar-IQ') : ''}</span>
          <button type="button" onClick={() => setEditing(true)} className="mr-2 font-bold text-red-700 underline">حدّث السبب</button>
        </p>
      ) : (
        <div className="mt-2 flex flex-wrap gap-2">
          <input value={text} onChange={(e) => setText(e.target.value)} placeholder="ليش متأخر؟ (مثلاً: الزبون مسافر، ننتظر مواد، …)"
            className="min-w-[14rem] flex-1 rounded-lg border border-red-200 bg-white px-3 py-1.5 text-sm" />
          <button type="button" disabled={busy || text.trim().length < 5} onClick={() => void save()}
            className="rounded-lg bg-red-600 px-3 py-1.5 text-xs font-bold text-white disabled:opacity-40">احفظ السبب</button>
        </div>
      )}
    </div>
  )
}
