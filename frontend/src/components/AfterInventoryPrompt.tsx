import { useEffect, useState } from 'react'
import { api } from '../api'
import MatrixNote from './MatrixNote'

// ═══ جرد العدّة بعد الحجز — طلب (ع) 10-05 ═══
// الفني الي طلع ويا ليدر: بعد ما يخلص الحجز يأشّر عدّته كاملة لو ناقصها شي.
// النقص يوصل لمسؤول المخزن فوراً، وماتركس يذكّر كل يوم لحد ما يجرد.
export default function AfterInventoryPrompt() {
  const [rows, setRows] = useState<{ bookingId: string; bookingCode: string; completedAt: string }[]>([])
  const [missing, setMissing] = useState<Record<string, string>>({})
  const [short, setShort] = useState<Record<string, boolean>>({})
  const [busy, setBusy] = useState<string | null>(null)
  const [err, setErr] = useState('')

  const load = () => { void api.getAfterInventoryPending().then(setRows).catch(() => setRows([])) }
  useEffect(load, [])
  if (rows.length === 0) return null

  const save = async (id: string, complete: boolean) => {
    if (!complete && !(missing[id] ?? '').trim()) { setErr('اكتب شنو الناقص.'); return }
    setBusy(id); setErr('')
    try { await api.saveAfterInventory(id, complete, complete ? undefined : missing[id]); load() }
    catch (e) { setErr(e instanceof Error ? e.message : 'تعذر الحفظ') }
    finally { setBusy(null) }
  }

  return (
    <div dir="rtl" className="mb-4 rounded-2xl border border-sky-200 bg-sky-50/60 p-4">
      <h3 className="mb-2 text-sm font-extrabold text-[#0f2040]">🧰 جرد عدّتك بعد الحجز</h3>
      <MatrixNote className="mb-3">قبل ما تنسى: شوف عدّتك بعد الشغل. إذا أكو شي ناقص اكتبه، يوصل لمسؤول المخزن بنفس اللحظة.</MatrixNote>
      {err && <p className="mb-2 text-xs text-red-600">{err}</p>}
      <div className="space-y-2">
        {rows.map((b) => (
          <div key={b.bookingId} className="rounded-xl border border-slate-200 bg-white p-3">
            <p className="mb-2 text-sm font-bold">{b.bookingCode} <span className="text-[11px] font-normal text-slate-500">· خلص {new Date(b.completedAt).toLocaleDateString('en-GB')}</span></p>
            {short[b.bookingId] ? (
              <div className="flex flex-wrap items-center gap-2">
                <input value={missing[b.bookingId] ?? ''} onChange={(e) => setMissing((m) => ({ ...m, [b.bookingId]: e.target.value }))}
                  placeholder="شنو الناقص؟ مثلاً: ميتر، دريل…" className="min-w-0 flex-1 rounded-lg border border-slate-200 px-2 py-1.5 text-sm" />
                <button type="button" disabled={busy === b.bookingId} onClick={() => void save(b.bookingId, false)}
                  className="rounded-lg bg-amber-600 px-3 py-1.5 text-xs font-bold text-white disabled:opacity-50">بلّغ بالنقص</button>
                <button type="button" onClick={() => setShort((s) => ({ ...s, [b.bookingId]: false }))} className="text-xs text-slate-500">رجوع</button>
              </div>
            ) : (
              <div className="flex flex-wrap gap-2">
                <button type="button" disabled={busy === b.bookingId} onClick={() => void save(b.bookingId, true)}
                  className="rounded-lg bg-emerald-600 px-4 py-1.5 text-xs font-bold text-white disabled:opacity-50">✅ عدّتي كاملة</button>
                <button type="button" onClick={() => setShort((s) => ({ ...s, [b.bookingId]: true }))}
                  className="rounded-lg border border-amber-300 bg-amber-50 px-4 py-1.5 text-xs font-bold text-amber-800">⚠️ أكو شي ناقص</button>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
