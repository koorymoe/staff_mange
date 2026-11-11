import { useEffect, useState } from 'react'
import { api, type OwnerGrants } from '../api'
import { useSession } from '../session'

// ═══ 👑 صلاحيات المالك لمدير النظام — ون بيس (11-10) ═══
// «خلي صلاحية للمالك ينطي صلاحياته لمدير النظام — بعيداً عن صلاحيات الموظفين».
// المالك يختار فعل فعل، لمدير نظام معيّن. الإنطاء نفسه يبقى بيده هو بس،
// وكل إنطاء أو سحب ينسجل بالسجل تحت.

export default function OwnerGrantsPage() {
  const { employee } = useSession()
  const isOwner = employee?.actualRole === 'OWNER'
  const [data, setData] = useState<OwnerGrants | null>(null)
  const [who, setWho] = useState('')
  const [picked, setPicked] = useState<Set<string>>(new Set())
  const [busy, setBusy] = useState(false)
  const [msg, setMsg] = useState('')
  const [error, setError] = useState('')
  const [reload, setReload] = useState(0)
  useEffect(() => {
    if (!isOwner) return
    let alive = true
    api.getOwnerGrants().then((d) => { if (alive) setData(d) }).catch((e: Error) => { if (alive) setError(e.message) })
    return () => { alive = false }
  }, [isOwner, reload])
  const choose = (id: string) => { setWho(id); setPicked(new Set(data?.grants[id] ?? [])); setMsg('') }
  const toggle = (a: string) => setPicked((s) => { const n = new Set(s); if (n.has(a)) n.delete(a); else n.add(a); return n })
  const save = () => {
    setBusy(true); setMsg(''); setError('')
    api.setOwnerGrants(who, [...picked])
      .then((r) => { setMsg(r.changed.length ? `انحفظ — ${r.changed.length} تغيير` : 'ماكو تغيير'); setReload((n) => n + 1) })
      .catch((e: Error) => setError(e.message))
      .finally(() => setBusy(false))
  }
  if (!isOwner) return <p className="p-6 text-center text-slate-500">الصفحة غير موجودة</p>
  if (error && !data) return <p role="alert" className="m-4 rounded-xl bg-red-50 p-3 text-sm font-bold text-red-700">{error}</p>
  if (!data) return <p className="p-6 text-sm text-slate-500">جاري التحميل…</p>
  return (
    <div className="space-y-4 p-4" dir="rtl">
      <header className="space-y-1">
        <h1 className="text-xl font-extrabold text-slate-900">👑 صلاحيات المالك لمدير النظام</h1>
        <p className="text-sm text-slate-600">هاي صلاحياتك إنت (مو صلاحيات الموظفين). تنطي منها لمدير نظام فعل فعل، وتسحبها وقت ما تريد — السحب فوري. إنطاء هالصلاحيات نفسه يبقى بيدك بس.</p>
      </header>
      {data.admins.length === 0 && <p className="rounded-xl bg-white p-4 text-sm text-slate-500 ring-1 ring-slate-200">ماكو مدير نظام فعّال حالياً.</p>}
      <div className="flex flex-wrap gap-2">
        {data.admins.map((a) => (
          <button key={a.id} type="button" onClick={() => choose(a.id)} aria-pressed={who === a.id}
            className={`rounded-xl px-4 py-2 text-sm font-bold ring-1 ${who === a.id ? 'bg-brand-700 text-white ring-brand-700' : 'bg-white text-slate-800 ring-slate-200'}`}>
            {a.name} <span className="ms-1 text-xs opacity-75">({(data.grants[a.id] ?? []).length})</span>
          </button>
        ))}
      </div>
      {who && (
        <section className="space-y-3 rounded-xl bg-white p-4 ring-1 ring-slate-200">
          <div className="flex flex-wrap items-center gap-2">
            <button type="button" onClick={() => setPicked(new Set(data.delegable.map((d) => d.action)))} className="rounded-lg px-3 py-1 text-sm ring-1 ring-slate-300">اختار الكل</button>
            <button type="button" onClick={() => setPicked(new Set())} className="rounded-lg px-3 py-1 text-sm ring-1 ring-slate-300">شيل الكل</button>
            <span className="text-xs text-slate-500">{picked.size} من {data.delegable.length}</span>
          </div>
          <ul className="grid gap-1 sm:grid-cols-2">
            {data.delegable.map((d) => (
              <li key={d.action}>
                <label className="flex cursor-pointer items-start gap-2 rounded-lg p-2 hover:bg-slate-50">
                  <input id={`g-${d.action}`} type="checkbox" checked={picked.has(d.action)} onChange={() => toggle(d.action)} className="mt-1" />
                  <span><span className="block text-sm text-slate-800">{d.label}</span><code className="text-[11px] text-slate-500" dir="ltr">{d.action}</code></span>
                </label>
              </li>
            ))}
          </ul>
          <div className="flex items-center gap-3">
            <button type="button" disabled={busy} onClick={save} className="min-h-11 rounded-xl bg-brand-700 px-6 text-sm font-extrabold text-white disabled:opacity-50">{busy ? '…' : 'احفظ'}</button>
            {msg && <span role="status" className="text-sm font-bold text-emerald-700">{msg}</span>}
            {error && <span role="alert" className="text-sm font-bold text-red-700">{error}</span>}
          </div>
        </section>
      )}
      <section className="space-y-2">
        <h2 className="text-sm font-extrabold text-slate-700">السجل</h2>
        {data.log.length === 0 ? <p className="text-sm text-slate-500">ماكو تغييرات بعد.</p> : (
          <ul className="space-y-1 text-sm">
            {data.log.map((l, i) => (
              <li key={i} className="rounded-lg bg-white p-2 ring-1 ring-slate-100">
                <span className="tabular-nums text-slate-500">{l.at}</span> · <b>{l.by || '—'}</b> {l.granted ? 'نطى' : 'سحب'} <b>{l.employee}</b> <code className="text-[11px]" dir="ltr">{l.action}</code>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  )
}
