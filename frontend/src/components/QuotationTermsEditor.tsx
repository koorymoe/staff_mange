import { useEffect, useState } from 'react'
import { api } from '../api'
import { askChoice } from '../utils/dialog'

// ⚙️ الشروط والأحكام بعرض السعر — قرار (ع) 10-08: صاحب صلاحية عرض السعر
// يضيف ويعدّل ويحذف ويرتّب. العروض المطبوعة قبل التعديل تبقى بنصها.
export default function QuotationTermsEditor({ onClose }: { onClose: () => void }) {
  const [terms, setTerms] = useState<string[] | null>(null)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  useEffect(() => { api.getQuotationTerms().then((r) => setTerms(r.map((t) => t.text))).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر')) }, [])

  const set = (i: number, v: string) => setTerms((t) => (t ?? []).map((x, j) => (j === i ? v : x)))
  const move = (i: number, d: number) => setTerms((t) => {
    const a = [...(t ?? [])]; const j = i + d
    if (j < 0 || j >= a.length) return a
    ;[a[i], a[j]] = [a[j], a[i]]; return a
  })
  const remove = async (i: number) => {
    if (await askChoice('تحذف هالشرط؟', [['yes', '🗑️ إي احذفه']])) setTerms((t) => (t ?? []).filter((_, j) => j !== i))
  }
  const save = async () => {
    setBusy(true); setErr('')
    try { await api.saveQuotationTerms(terms ?? []); onClose() } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر الحفظ') } finally { setBusy(false) }
  }

  return (
    <div dir="rtl" className="fixed inset-0 z-[90] flex items-center justify-center bg-black/50 p-4" onClick={onClose}>
      <div className="max-h-[90vh] w-full max-w-2xl overflow-y-auto rounded-2xl bg-white p-5 shadow-xl" onClick={(e) => e.stopPropagation()}>
        <h3 className="text-lg font-bold text-[#0f2040]">⚙️ الشروط والأحكام بعرض السعر</h3>
        <p className="mt-1 text-xs text-slate-500">تطلع بكل عرض سعر جديد بنفس هالترتيب. العروض الي انطبعت قبل تبقى مثل ما هي.</p>
        {terms === null ? <p className="mt-4 text-slate-400">جاري التحميل…</p> : (
          <div className="mt-4 space-y-2">
            {terms.map((t, i) => (
              <div key={i} className="flex items-start gap-2 rounded-xl border border-slate-200 p-2">
                <span className="mt-2 w-6 text-center text-sm font-bold text-slate-500">{i + 1}.</span>
                <textarea value={t} onChange={(e) => set(i, e.target.value)} rows={2} className="min-w-0 flex-1 rounded-lg border border-slate-200 p-2 text-sm" />
                <div className="flex flex-col gap-1">
                  <button type="button" onClick={() => move(i, -1)} disabled={i === 0} className="rounded bg-slate-100 px-2 text-xs disabled:opacity-30" title="لفوك">▲</button>
                  <button type="button" onClick={() => move(i, 1)} disabled={i === terms.length - 1} className="rounded bg-slate-100 px-2 text-xs disabled:opacity-30" title="لجوّه">▼</button>
                  <button type="button" onClick={() => void remove(i)} className="rounded bg-red-50 px-2 text-xs text-red-600" title="حذف">🗑️</button>
                </div>
              </div>
            ))}
            <button type="button" onClick={() => setTerms((t) => [...(t ?? []), ''])} className="w-full rounded-xl border-2 border-dashed border-slate-300 py-2 text-sm font-bold text-slate-600">➕ أضف شرط</button>
          </div>
        )}
        {err && <p className="mt-2 text-sm text-red-600">{err}</p>}
        <div className="mt-4 flex justify-end gap-2">
          <button type="button" onClick={onClose} className="rounded-lg border border-slate-200 px-4 py-2 text-sm">إلغاء</button>
          <button type="button" disabled={busy || terms === null} onClick={() => void save()} className="rounded-lg bg-[#1a237e] px-4 py-2 text-sm font-bold text-white disabled:opacity-50">💾 احفظ</button>
        </div>
      </div>
    </div>
  )
}
