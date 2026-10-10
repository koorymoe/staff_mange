import { useEffect, useState } from 'react'
import { api, type NetworkPriceItem, type NetworkBracket } from '../api'

// ═══ تعديل أسعار الشبكات — المالك ومدير النظام بس ═══
//
// تسعيرة الشبكات لسه تتبني: فقرات راح تنضاف وأرقام راح تتغيّر. لو
// خليناها بالكود، كل تغيير سعر يريد تعديل ونشر ويتأخر أيام. من هنا
// يتغيّر السعر ويشتغل بالحساب الجاي مباشرة.
//
// ⚠️ التعطيل مو محو: الفواتير القديمة تشير للفقرة، ومحوها يخلي
// فاتورة منجزة بلا تفسير لمبلغها.

type Draft = {
  id?: string
  label: string
  unit: string
  pricingMode: 'FLAT' | 'TIERED' | 'BRACKET'
  basePrice: string
  includedQty: string
  extraPerUnit: string
  brackets: { upTo: string; unitPrice: string }[]
  note: string
  sortOrder: number
}

const emptyDraft = (): Draft => ({
  label: '', unit: 'قطعة', pricingMode: 'FLAT',
  basePrice: '', includedQty: '', extraPerUnit: '',
  brackets: [{ upTo: '', unitPrice: '' }], note: '', sortOrder: 0,
})

const toDraft = (it: NetworkPriceItem): Draft => ({
  id: it.id,
  label: it.label,
  unit: it.unit,
  pricingMode: it.pricingMode,
  basePrice: String(it.basePrice),
  includedQty: String(it.includedQty),
  extraPerUnit: String(it.extraPerUnit),
  brackets: it.brackets.length ? it.brackets.map((b) => ({ upTo: String(b.upTo), unitPrice: String(b.unitPrice) })) : [{ upTo: '', unitPrice: '' }],
  note: it.note || '',
  sortOrder: it.sortOrder,
})

export default function NetworkPricesPage() {
  const [items, setItems] = useState<NetworkPriceItem[]>([])
  const [draft, setDraft] = useState<Draft | null>(null)
  const [msg, setMsg] = useState<string | null>(null)
  const [err, setErr] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)

  const load = () => api.getNetworkPrices().then(setItems).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر الجلب'))
  useEffect(() => { void load() }, [])

  // Number('') = 0 — فالحقل الفارغ ينحسب صفر بالسكوت لو ما نتحقق
  const numOf = (v: string) => {
    const n = Number(v)
    return Number.isFinite(n) ? n : 0
  }

  const save = async () => {
    if (!draft) return
    setBusy(true); setErr(null); setMsg(null)
    const payload = {
      label: draft.label.trim(),
      unit: draft.unit.trim() || 'قطعة',
      pricingMode: draft.pricingMode,
      basePrice: numOf(draft.basePrice),
      includedQty: numOf(draft.includedQty),
      extraPerUnit: numOf(draft.extraPerUnit),
      brackets: draft.pricingMode === 'BRACKET'
        ? draft.brackets
            .filter((b) => numOf(b.unitPrice) > 0)
            .map((b): NetworkBracket => ({ upTo: numOf(b.upTo), unitPrice: numOf(b.unitPrice) }))
        : [],
      note: draft.note.trim() || null,
      sortOrder: draft.sortOrder,
    }
    try {
      if (draft.id) await api.updateNetworkPrice(draft.id, payload)
      else await api.createNetworkPrice(payload)
      setMsg('انحفظ ✓')
      setDraft(null)
      await load()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر الحفظ')
    } finally { setBusy(false) }
  }

  const deactivate = async (it: NetworkPriceItem) => {
    if (!window.confirm(`تعطيل «${it.label}»؟\n\nما ينمحي — يختفي من الاستمارة بس والفواتير القديمة تبقى مفهومة.`)) return
    try { await api.deactivateNetworkPrice(it.id); await load() }
    catch (e) { setErr(e instanceof Error ? e.message : 'تعذر التعطيل') }
  }

  const num = (v: number) => v.toLocaleString('en-US')

  return (
    <div dir="rtl" className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-brand-900">أسعار الشبكات</h2>
        <p className="mt-1 text-slate-500">
          التغيير هنا ينطبق على كل حساب جاي مباشرة. المالك ومدير النظام بس.
        </p>
      </div>

      {err && <p className="rounded-lg bg-red-50 p-4 text-red-600">{err}</p>}
      {msg && <p className="rounded-lg bg-emerald-50 p-4 font-bold text-emerald-700">{msg}</p>}

      <button
        onClick={() => setDraft(emptyDraft())}
        className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-2 text-sm font-bold text-white"
      >
        + فقرة تسعيرة جديدة
      </button>

      {draft && (
        <div className="space-y-3 rounded-xl border border-brand-200 bg-brand-50/40 p-5">
          <h3 className="font-bold text-[#0f2040]">{draft.id ? 'تعديل فقرة' : 'فقرة جديدة'}</h3>
          <div className="grid gap-3 sm:grid-cols-3">
            <label className="text-xs font-bold text-slate-600">
              اسم الفقرة
              <input value={draft.label} onChange={(e) => setDraft({ ...draft, label: e.target.value })}
                className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
            </label>
            <label className="text-xs font-bold text-slate-600">
              الوحدة (متر، بورت، جهاز...)
              <input value={draft.unit} onChange={(e) => setDraft({ ...draft, unit: e.target.value })}
                className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
            </label>
            <label className="text-xs font-bold text-slate-600">
              نمط التسعير
              <select value={draft.pricingMode}
                onChange={(e) => setDraft({ ...draft, pricingMode: e.target.value as Draft['pricingMode'] })}
                className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm">
                <option value="FLAT">سعر ثابت للوحدة</option>
                <option value="TIERED">مقطوعة + زيادة بعد كمية</option>
                <option value="BRACKET">سعر الوحدة حسب الحجم (شرائح)</option>
              </select>
            </label>
          </div>

          {draft.pricingMode === 'FLAT' && (
            <label className="block text-xs font-bold text-slate-600">
              سعر الوحدة
              <input type="number" value={draft.basePrice} onChange={(e) => setDraft({ ...draft, basePrice: e.target.value })}
                className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm sm:w-56" />
            </label>
          )}

          {draft.pricingMode === 'TIERED' && (
            <div className="grid gap-3 sm:grid-cols-3">
              <label className="text-xs font-bold text-slate-600">
                المبلغ المقطوع
                <input type="number" value={draft.basePrice} onChange={(e) => setDraft({ ...draft, basePrice: e.target.value })}
                  className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
              </label>
              <label className="text-xs font-bold text-slate-600">
                الكمية المشمولة بيه
                <input type="number" value={draft.includedQty} onChange={(e) => setDraft({ ...draft, includedQty: e.target.value })}
                  className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
              </label>
              <label className="text-xs font-bold text-slate-600">
                سعر كل وحدة زايدة
                <input type="number" value={draft.extraPerUnit} onChange={(e) => setDraft({ ...draft, extraPerUnit: e.target.value })}
                  className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
              </label>
            </div>
          )}

          {draft.pricingMode === 'BRACKET' && (
            <div className="space-y-2">
              <p className="text-xs text-slate-500">
                كل شريحة: «لحد كم وحدة» و«سعر الوحدة». خلي آخر شريحة بـ«لحد» = ٠ يعني «وأكثر».
              </p>
              {draft.brackets.map((b, i) => (
                <div key={i} className="flex items-center gap-2">
                  <input type="number" placeholder="لحد (٠ = وأكثر)" value={b.upTo}
                    onChange={(e) => setDraft({ ...draft, brackets: draft.brackets.map((x, j) => j === i ? { ...x, upTo: e.target.value } : x) })}
                    className="w-40 rounded-lg border border-slate-300 px-3 py-2 text-sm" />
                  <input type="number" placeholder="سعر الوحدة" value={b.unitPrice}
                    onChange={(e) => setDraft({ ...draft, brackets: draft.brackets.map((x, j) => j === i ? { ...x, unitPrice: e.target.value } : x) })}
                    className="w-40 rounded-lg border border-slate-300 px-3 py-2 text-sm" />
                  <button onClick={() => setDraft({ ...draft, brackets: draft.brackets.filter((_, j) => j !== i) })}
                    className="text-sm text-red-600">حذف</button>
                </div>
              ))}
              <button onClick={() => setDraft({ ...draft, brackets: [...draft.brackets, { upTo: '', unitPrice: '' }] })}
                className="text-sm font-bold text-brand-700">+ شريحة</button>
            </div>
          )}

          <label className="block text-xs font-bold text-slate-600">
            ملاحظة تظهر للفني
            <input value={draft.note} onChange={(e) => setDraft({ ...draft, note: e.target.value })}
              className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
          </label>

          <div className="flex gap-2">
            <button onClick={save} disabled={busy}
              className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-2 text-sm font-bold text-white disabled:opacity-50">
              {busy ? 'جاري الحفظ...' : 'حفظ'}
            </button>
            <button onClick={() => setDraft(null)} className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-bold text-slate-600">
              إلغاء
            </button>
          </div>
        </div>
      )}

      <div className="space-y-2">
        {items.map((it) => (
          <div key={it.id} className={`flex flex-wrap items-center justify-between gap-3 rounded-xl border p-4 ${it.active ? 'border-white bg-white' : 'border-slate-200 bg-slate-50 opacity-70'}`}>
            <div>
              <p className="font-bold text-[#0f2040]">
                {it.label} {!it.active && <span className="text-xs font-normal text-slate-500">(معطّلة)</span>}
              </p>
              <p className="text-xs text-slate-500">
                {it.pricingMode === 'TIERED' && `لحد ${num(it.includedQty)} ${it.unit} → ${num(it.basePrice)}، وكل ${it.unit} زايد ${num(it.extraPerUnit)}`}
                {it.pricingMode === 'FLAT' && `${num(it.basePrice)} لل${it.unit}`}
                {it.pricingMode === 'BRACKET' && it.brackets.map((b) => (b.upTo > 0 ? `لحد ${num(b.upTo)}` : 'وأكثر') + ` → ${num(b.unitPrice)}`).join(' • ')}
              </p>
            </div>
            <div className="flex gap-2">
              <button onClick={() => setDraft(toDraft(it))} className="rounded-lg border border-slate-300 px-3 py-1.5 text-xs font-bold text-slate-700">تعديل</button>
              {it.active && (
                <button onClick={() => deactivate(it)} className="rounded-lg border border-red-300 px-3 py-1.5 text-xs font-bold text-red-600">تعطيل</button>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
