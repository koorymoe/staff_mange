import { useState } from 'react'
import { apiRequest } from '../api'
import type { AnnexPart } from '../utils/projectQuotePrint'

// «⚙️ قوالب الملحق» (دازاي 11-06): تعديل قوالب الملحق الفني، و«نسخ كقالب جديد» حتى تسوي
// قالب لنوع مشروع ثاني (كاميرات، شبكات…) من قالب موجود.
export type AnnexTemplate = { id: string; name: string; parts: AnnexPart[] }
type Draft = { id: string; name: string; parts: AnnexPart[] }

const input = 'w-full rounded-lg border border-slate-300 px-2 py-1.5 text-sm'
const clone = (p: AnnexPart[]): AnnexPart[] => p.map((x) => ({ title: x.title, items: x.items.map((i) => ({ ...i })) }))

export default function AnnexTemplatesManager({ templates, onClose, onSaved }: { templates: AnnexTemplate[]; onClose: () => void; onSaved: () => void }) {
  const [draft, setDraft] = useState<Draft | null>(templates[0] ? { ...templates[0], parts: clone(templates[0].parts) } : null)
  const [msg, setMsg] = useState<{ t: string; ok: boolean } | null>(null)
  const [busy, setBusy] = useState(false)

  const pick = (t: AnnexTemplate) => { setDraft({ ...t, parts: clone(t.parts) }); setMsg(null) }
  const setParts = (f: (p: AnnexPart[]) => AnnexPart[]) => setDraft((d) => (d ? { ...d, parts: f(d.parts) } : d))
  const setItem = (pi: number, ii: number, patch: Partial<{ title: string; text: string }>) =>
    setParts((ps) => ps.map((p, i) => (i !== pi ? p : { ...p, items: p.items.map((it, j) => (j === ii ? { ...it, ...patch } : it)) })))

  const save = async () => {
    if (!draft) return
    setBusy(true)
    setMsg(null)
    try {
      const r = await apiRequest<{ id: string }>('/quotation-annex-templates', { method: 'PUT', body: JSON.stringify(draft) })
      setDraft({ ...draft, id: r.id })
      setMsg({ t: '✅ انحفظ القالب', ok: true })
      onSaved()
    } catch (e) {
      setMsg({ t: e instanceof Error ? e.message : 'تعذر الحفظ', ok: false })
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/40 p-4" role="dialog" aria-label="قوالب الملحق">
      <div className="w-full max-w-3xl space-y-3 rounded-2xl bg-white p-4 shadow-xl">
        <div className="flex items-center justify-between">
          <h2 className="text-base font-bold text-slate-700">⚙️ قوالب الملحق</h2>
          <button type="button" onClick={onClose} className="rounded-lg bg-slate-100 px-3 py-1 text-sm">سد</button>
        </div>
        <div className="flex flex-wrap gap-2">
          {templates.map((t) => (
            <button key={t.id} type="button" onClick={() => pick(t)}
              className={`rounded-lg px-3 py-1 text-sm ${draft?.id === t.id ? 'bg-slate-800 text-white' : 'bg-slate-100'}`}>{t.name}</button>
          ))}
          {draft && (
            <button type="button" onClick={() => { setDraft({ id: '', name: `${draft.name} (نسخة)`, parts: clone(draft.parts) }); setMsg({ t: 'هاي نسخة جديدة — غيّر اسمها واحفظ', ok: true }) }}
              className="rounded-lg bg-amber-50 px-3 py-1 text-sm text-amber-800">📄 نسخ كقالب جديد</button>
          )}
        </div>
        {!draft ? <p className="text-sm text-slate-500">ماكو قوالب.</p> : (
          <>
            <label className="block text-sm">اسم القالب
              <input className={input} value={draft.name} onChange={(e) => setDraft({ ...draft, name: e.target.value })} />
            </label>
            {draft.parts.map((p, pi) => (
              <div key={pi} className="space-y-2 rounded-xl border border-slate-200 p-3">
                <div className="flex gap-1">
                  <input className={`${input} font-bold`} value={p.title} placeholder="عنوان الجزء"
                    onChange={(e) => setParts((ps) => ps.map((x, i) => (i === pi ? { ...x, title: e.target.value } : x)))} />
                  <button type="button" onClick={() => setParts((ps) => ps.filter((_, i) => i !== pi))} className="rounded-lg bg-rose-50 px-2 text-xs text-rose-700">شيل الجزء</button>
                </div>
                {p.items.map((it, ii) => (
                  <div key={ii} className="space-y-1">
                    <div className="flex gap-1">
                      <input className={input} value={it.title} onChange={(e) => setItem(pi, ii, { title: e.target.value })} />
                      <button type="button" aria-label="شيل البند" onClick={() => setParts((ps) => ps.map((x, i) => (i !== pi ? x : { ...x, items: x.items.filter((_, j) => j !== ii) })))} className="rounded-lg bg-rose-50 px-2 text-xs text-rose-700">✕</button>
                    </div>
                    <textarea rows={3} className={input} value={it.text} onChange={(e) => setItem(pi, ii, { text: e.target.value })} />
                  </div>
                ))}
                <button type="button" onClick={() => setParts((ps) => ps.map((x, i) => (i !== pi ? x : { ...x, items: [...x.items, { title: 'بند جديد', text: '' }] })))} className="rounded-lg bg-slate-100 px-2 py-1 text-xs font-bold">＋ بند</button>
              </div>
            ))}
            <button type="button" onClick={() => setParts((ps) => [...ps, { title: 'جزء جديد', items: [{ title: 'بند جديد', text: '' }] }])} className="rounded-lg bg-slate-100 px-2 py-1 text-xs font-bold">＋ جزء</button>
            <div className="flex items-center gap-2">
              <button type="button" disabled={busy} onClick={() => void save()} className="rounded-lg bg-emerald-600 px-4 py-1.5 text-sm font-bold text-white disabled:opacity-50">{busy ? '…' : draft.id ? '💾 احفظ التعديل' : '💾 احفظ كقالب جديد'}</button>
              {msg && <span className={`text-sm ${msg.ok ? 'text-emerald-700' : 'text-rose-700'}`}>{msg.t}</span>}
            </div>
          </>
        )}
      </div>
    </div>
  )
}
