import { useEffect, useState } from 'react'
import { api, apiRequest, fileUrl, storeFile, thumbUrl } from '../api'
import { notify } from '../ui/notify'
import { fmtDateTime } from '../utils/date'

// ═══ 🧰 عهدة الأدوات (ل) 11-25 دازاي ═══
// «الموظف بنفسه يختار الأداة ووقت أخذها، تبقى بعهدته، ويسوي تسليم ويرفع صورة، واني أدققها،
//  أو اني أختار الموظف في حال هو ما گدر».

type Item = {
  id: string; toolId: string; toolName: string; employeeId: string; employeeName: string; takenAt: string | null; selfCheckout: boolean; byName: string | null
  returnRequestedAt: string | null; returnPhoto: string | null; returnNote: string | null; verifyNote: string | null
}
type Resp = { mine: Item[]; tools: { id: string; name: string; available: number }[]; canManage: boolean; all?: Item[] }
const nowLocal = () => new Date(Date.now() - new Date().getTimezoneOffset() * 60000).toISOString().slice(0, 16)
const inp = 'rounded-lg border border-slate-300 p-2 text-sm'

export default function ToolCustodyPage() {
  const [d, setD] = useState<Resp | null>(null)
  const [err, setErr] = useState('')
  const [take, setTake] = useState({ toolId: '', takenAt: nowLocal(), note: '', employeeId: '' })
  const [emps, setEmps] = useState<{ id: string; name: string }[]>([])
  const [busy, setBusy] = useState('')
  const [ret, setRet] = useState<Record<string, { file?: File; note: string }>>({})
  const [rej, setRej] = useState<Record<string, string>>({})
  useEffect(() => { apiRequest<Resp>('/tool-custody').then(setD).catch((e: Error) => setErr(e.message)) }, [])
  useEffect(() => { if (d?.canManage && !emps.length) api.getEmployees().then((es) => setEmps(es.map((e) => ({ id: e.id, name: e.name })))).catch(() => {}) }, [d?.canManage, emps.length])
  if (err) return <p role="alert" className="text-rose-700">{err}</p>
  if (!d) return <p className="py-10 text-center text-slate-400">جاري التحميل…</p>

  const run = async (key: string, fn: () => Promise<Resp>, ok: string) => {
    setBusy(key)
    try { setD(await fn()); notify(ok, 'success'); return true } catch (e) { notify(e instanceof Error ? e.message : 'تعذر'); return false } finally { setBusy('') }
  }
  const doTake = async () => {
    const body = { toolId: take.toolId, takenAt: new Date(take.takenAt).toISOString(), note: take.note, employeeId: take.employeeId }
    if (await run('take', () => apiRequest<Resp>('/tool-custody/take', { method: 'POST', body: JSON.stringify(body) }), 'انسجّلت بالعهدة ✅')) setTake({ toolId: '', takenAt: nowLocal(), note: '', employeeId: '' })
  }
  const doReturn = async (it: Item) => {
    const r = ret[it.id]
    if (!r?.file) { notify('اختار صورة الأداة وهي مسلّمة'); return }
    setBusy(it.id)
    try {
      const photo = await storeFile(r.file, 'tool-returns')
      setD(await apiRequest<Resp>(`/tool-custody/${it.id}/return`, { method: 'POST', body: JSON.stringify({ photo, note: r.note }) }))
      notify('انرسل التسليم — ينتظر التدقيق ✅', 'success')
    } catch (e) { notify(e instanceof Error ? e.message : 'تعذر') } finally { setBusy('') }
  }
  const verify = (it: Item, ok: boolean) => run(it.id, () => apiRequest<Resp>(`/tool-custody/${it.id}/verify`, { method: 'POST', body: JSON.stringify({ ok, note: ok ? '' : rej[it.id] ?? '' }) }), ok ? 'رجعت للمخزن ✅' : 'رجعت لعهدته')
  const pending = (d.all ?? []).filter((x) => x.returnRequestedAt)
  const held = (d.all ?? []).filter((x) => !x.returnRequestedAt)

  return (
    <div className="mx-auto max-w-5xl space-y-4" data-testid="tool-custody">
      <header>
        <h1 className="text-2xl font-extrabold text-[#0f2040]">🧰 عهدة الأدوات</h1>
        <p className="text-sm text-slate-500">خذ الأداة بنفسك وسجّل وقت أخذها — تبقى بعهدتك لحد ما تسلّمها بصورة ويتدقق التسليم.</p>
      </header>

      <section className="grid gap-2 rounded-2xl bg-white p-4 shadow-sm sm:grid-cols-[1fr_12rem_1fr_auto]">
        <h2 className="font-bold sm:col-span-4">➕ آخذ أداة</h2>
        {d.canManage && (
          <select value={take.employeeId} onChange={(e) => setTake({ ...take, employeeId: e.target.value })} className={`${inp} sm:col-span-4`}>
            <option value="">— لنفسي —</option>
            {emps.map((e) => <option key={e.id} value={e.id}>عن: {e.name}</option>)}
          </select>
        )}
        <select value={take.toolId} onChange={(e) => setTake({ ...take, toolId: e.target.value })} className={inp}>
          <option value="">اختار الأداة…</option>
          {d.tools.map((t) => <option key={t.id} value={t.id}>{t.name} ({t.available} متوفر)</option>)}
        </select>
        <input type="datetime-local" value={take.takenAt} onChange={(e) => setTake({ ...take, takenAt: e.target.value })} className={inp} aria-label="وقت الأخذ" />
        <input value={take.note} onChange={(e) => setTake({ ...take, note: e.target.value })} placeholder="ملاحظة (مثلاً: لحجز B1234)" className={inp} />
        <button type="button" disabled={!take.toolId || busy === 'take'} onClick={() => void doTake()} className="rounded-lg bg-[#0f2040] px-4 py-2 text-sm font-bold text-white disabled:opacity-40">سجّل</button>
      </section>

      <section className="space-y-2 rounded-2xl bg-white p-4 shadow-sm">
        <h2 className="font-bold">📦 عهدتي ({d.mine.length})</h2>
        {d.mine.length === 0 && <p className="text-sm text-slate-500">ماكو أدوات بعهدتك.</p>}
        {d.mine.map((it) => (
          <div key={it.id} className="space-y-2 rounded-xl bg-slate-50 p-3">
            <p className="flex flex-wrap justify-between gap-2 text-sm"><b>{it.toolName}</b><span className="text-slate-500">أخذتها {it.takenAt ? fmtDateTime(it.takenAt) : '—'}{it.byName ? ` · سجّلها ${it.byName}` : ''}</span></p>
            {it.verifyNote && !it.returnRequestedAt && <p className="rounded-lg bg-rose-50 px-3 py-1 text-xs text-rose-800">التسليم انرفض: {it.verifyNote} — سلّمها من جديد</p>}
            {it.returnRequestedAt ? <p className="text-xs font-bold text-amber-700">⏳ سلّمتها {fmtDateTime(it.returnRequestedAt)} — تنتظر التدقيق</p> : (
              <div className="flex flex-wrap items-center gap-2">
                <input type="file" accept="image/*" capture="environment" onChange={(e) => setRet({ ...ret, [it.id]: { note: ret[it.id]?.note ?? '', file: e.target.files?.[0] } })} className="text-xs" />
                <input value={ret[it.id]?.note ?? ''} onChange={(e) => setRet({ ...ret, [it.id]: { ...ret[it.id], note: e.target.value } })} placeholder="ملاحظة التسليم" className={`${inp} flex-1`} />
                <button type="button" disabled={busy === it.id || !ret[it.id]?.file} onClick={() => void doReturn(it)} className="rounded-lg bg-emerald-600 px-4 py-2 text-sm font-bold text-white disabled:opacity-40">📸 سلّم</button>
              </div>
            )}
          </div>
        ))}
      </section>

      {d.canManage && (
        <>
          <section className="space-y-2 rounded-2xl bg-white p-4 shadow-sm">
            <h2 className="font-bold">🔍 تسليمات تنتظر تدقيقك ({pending.length})</h2>
            {pending.length === 0 && <p className="text-sm text-slate-500">ماكو تسليمات معلّقة.</p>}
            {pending.map((it) => (
              <div key={it.id} className="flex flex-wrap gap-3 rounded-xl bg-amber-50 p-3">
                {it.returnPhoto && <a href={fileUrl(it.returnPhoto)} target="_blank" rel="noreferrer"><img src={thumbUrl(it.returnPhoto, 160)} alt="صورة التسليم" className="h-24 w-24 rounded-lg object-cover" /></a>}
                <div className="min-w-0 flex-1 space-y-1 text-sm">
                  <p><b>{it.toolName}</b> — {it.employeeName}</p>
                  <p className="text-xs text-slate-600">أخذها {it.takenAt ? fmtDateTime(it.takenAt) : '—'} · سلّمها {fmtDateTime(it.returnRequestedAt!)}{it.returnNote ? ` · «${it.returnNote}»` : ''}</p>
                  <div className="flex flex-wrap gap-2">
                    <button type="button" disabled={busy === it.id} onClick={() => void verify(it, true)} className="rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-bold text-white">✅ صحيح — ترجع للمخزن</button>
                    <input value={rej[it.id] ?? ''} onChange={(e) => setRej({ ...rej, [it.id]: e.target.value })} placeholder="سبب الرفض" className={`${inp} flex-1 py-1 text-xs`} />
                    <button type="button" disabled={busy === it.id || !(rej[it.id] ?? '').trim()} onClick={() => void verify(it, false)} className="rounded-lg bg-slate-200 px-3 py-1.5 text-xs font-bold text-slate-700 disabled:opacity-40">🚫 مو صحيح</button>
                  </div>
                </div>
              </div>
            ))}
          </section>
          <section className="space-y-2 rounded-2xl bg-white p-4 shadow-sm">
            <h2 className="font-bold">👥 كل العهد المفتوحة ({held.length})</h2>
            <table className="w-full text-sm">
              <thead className="bg-slate-50 text-xs text-slate-600"><tr>{['الأداة', 'بعهدة', 'من وكت', 'سجّلها'].map((h) => <th key={h} className="p-2 text-start">{h}</th>)}</tr></thead>
              <tbody className="divide-y">{held.map((it) => <tr key={it.id}><td className="p-2 font-bold">{it.toolName}</td><td>{it.employeeName}</td><td>{it.takenAt ? fmtDateTime(it.takenAt) : '—'}</td><td>{it.selfCheckout ? 'بنفسه' : it.byName ?? '—'}</td></tr>)}</tbody>
            </table>
          </section>
        </>
      )}
    </div>
  )
}
