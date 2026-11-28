import { useEffect, useMemo, useState } from 'react'
import { apiRequest } from '../api'
import { notify } from '../ui/notify'

// ═══ 🧾 خيارات الفاتورة لكل خدمة — تجريبي (ل) 11-24 دازاي ═══
// المالك ومدير النظام: لكل خدمة خانات (الاسم، الوحدة، سعر الوحدة) وأقل سعر للشغل.
// «تنضاف بس ما تنطلق» — الفواتير الحقيقية بعدها بالطريقة القديمة؛ هنا حاسبة تجربة بس.

type Opt = { id?: string; label: string; unit: string; unitPrice: number }
type Tpl = { serviceId: string; serviceName: string; minWork: number; minTiers: number[]; options: Opt[] }
type Quote = { lines: { label: string; unit: string; qty: number; price: number; total: number }[]; sum: number; minWork: number; work: number; materials: number; total: number; minUsed: boolean }

const iqd = (v: number) => `${Math.round(v).toLocaleString('en-US')} د.ع`
const inp = 'rounded-lg border border-slate-300 p-2 text-sm'

export default function ServiceInvoiceOptionsPage() {
  const [all, setAll] = useState<Tpl[] | null>(null)
  const [err, setErr] = useState('')
  const [sel, setSel] = useState('')
  const [q, setQ] = useState('')
  const [opts, setOpts] = useState<Opt[]>([])
  const [tiers, setTiers] = useState<number[]>([])
  const [busy, setBusy] = useState(false)
  const [qty, setQty] = useState<Record<string, number>>({})
  const [materials, setMaterials] = useState(0)
  const [quote, setQuote] = useState<Quote | null>(null)

  useEffect(() => { apiRequest<Tpl[]>('/service-invoice-options').then(setAll).catch((e: Error) => setErr(e.message)) }, [])
  const cur = useMemo(() => all?.find((t) => t.serviceId === sel) ?? null, [all, sel])
  const pick = (t: Tpl) => { setSel(t.serviceId); setOpts(t.options.map((o) => ({ ...o }))); setTiers([...(t.minTiers ?? [])]); setQty({}); setMaterials(0); setQuote(null) }
  const dirty = !!cur && (JSON.stringify(tiers) !== JSON.stringify(cur.minTiers ?? []) || JSON.stringify(opts.map(({ label, unit, unitPrice }) => [label, unit, unitPrice])) !== JSON.stringify(cur.options.map(({ label, unit, unitPrice }) => [label, unit, unitPrice])))

  if (err) return <p role="alert" className="text-rose-700">{err}</p>
  if (!all) return <p className="py-10 text-center text-slate-400">جاري التحميل…</p>
  const shown = all.filter((t) => t.serviceName.includes(q.trim()))
  const setO = (i: number, p: Partial<Opt>) => setOpts((xs) => xs.map((x, k) => (k === i ? { ...x, ...p } : x)))
  const move = (i: number, d: -1 | 1) => setOpts((xs) => { const j = i + d; if (j < 0 || j >= xs.length) return xs; const c = [...xs]; [c[i], c[j]] = [c[j], c[i]]; return c })

  const save = async () => {
    if (!cur) return
    setBusy(true)
    try {
      const next = await apiRequest<Tpl[]>(`/service-invoice-options/${cur.serviceId}`, { method: 'PUT', body: JSON.stringify({ minWork: 0, minTiers: tiers.filter((x) => x > 0), options: opts }) })
      setAll(next)
      const t = next.find((x) => x.serviceId === cur.serviceId)
      if (t) pick(t)
      notify('انحفظت خيارات الخدمة ✅', 'success')
    } catch (e) { notify(e instanceof Error ? e.message : 'تعذر الحفظ') } finally { setBusy(false) }
  }
  const test = async () => {
    if (!cur) return
    try { setQuote(await apiRequest<Quote>(`/service-invoice-options/${cur.serviceId}/quote`, { method: 'POST', body: JSON.stringify({ qty, materials }) })) } catch (e) { notify(e instanceof Error ? e.message : 'تعذر') }
  }

  return (
    <div className="mx-auto max-w-6xl space-y-4" data-testid="service-invoice-options">
      <header>
        <h1 className="text-2xl font-extrabold text-[#0f2040]">🧾 خيارات الفاتورة لكل خدمة</h1>
        <p className="mt-1 rounded-xl bg-amber-50 px-4 py-2 text-sm font-bold text-amber-900">🧪 تجريبي — الفواتير الحقيقية بعدها بالطريقة القديمة. هنا تبني الخيارات وتجرّب الحساب، ولمن نتأكد نطلقها.</p>
      </header>
      <div className="grid gap-4 md:grid-cols-[18rem_1fr]">
        <aside className="space-y-2 rounded-2xl bg-white p-3 shadow-sm">
          <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="ابحث عن خدمة…" className={`${inp} w-full`} />
          <ul className="max-h-[70vh] space-y-1 overflow-y-auto">
            {shown.map((t) => (
              <li key={t.serviceId}>
                <button type="button" onClick={() => pick(t)} className={`flex w-full items-center justify-between rounded-lg px-3 py-2 text-start text-sm ${sel === t.serviceId ? 'bg-[#0f2040] text-white' : 'hover:bg-slate-100'}`}>
                  <span className="font-bold">{t.serviceName}</span>
                  <span className={`rounded-full px-2 text-xs ${t.options.length ? 'bg-emerald-100 text-emerald-800' : 'bg-slate-100 text-slate-500'}`}>{t.options.length || '—'}</span>
                </button>
              </li>
            ))}
          </ul>
        </aside>
        {!cur ? <p className="rounded-2xl bg-white p-10 text-center text-slate-500 shadow-sm">اختار خدمة من القائمة</p> : (
          <section className="space-y-4">
            <div className="space-y-3 rounded-2xl bg-white p-4 shadow-sm">
              <h2 className="text-lg font-bold text-slate-800">{cur.serviceName}</h2>
              {opts.length === 0 && <p className="text-sm text-slate-500">ماكو خيارات بعد — أضف أول خانة.</p>}
              {opts.map((o, i) => (
                <div key={i} className="grid grid-cols-[auto_1fr_7rem_8rem_auto] items-center gap-2" data-testid="sio-row">
                  <span className="flex flex-col">
                    <button type="button" aria-label="لفوگ" onClick={() => move(i, -1)} className="text-xs text-slate-400 hover:text-slate-700">▲</button>
                    <button type="button" aria-label="لجوه" onClick={() => move(i, 1)} className="text-xs text-slate-400 hover:text-slate-700">▼</button>
                  </span>
                  <input value={o.label} onChange={(e) => setO(i, { label: e.target.value })} placeholder="اسم الخانة (مثلاً: تسليك كيبل)" maxLength={80} className={inp} />
                  <input value={o.unit} onChange={(e) => setO(i, { unit: e.target.value })} placeholder="الوحدة (متر…)" maxLength={20} className={inp} />
                  <label className="flex items-center gap-1"><input type="number" min={0} value={o.unitPrice} onChange={(e) => setO(i, { unitPrice: Number(e.target.value) })} className={`${inp} w-full`} dir="ltr" /><small className="text-slate-500">د.ع</small></label>
                  <button type="button" onClick={() => setOpts((xs) => xs.filter((_, k) => k !== i))} className="rounded-lg px-2 py-1 text-rose-600 hover:bg-rose-50" aria-label="شيل الخانة">✕</button>
                </div>
              ))}
              <button type="button" onClick={() => setOpts((xs) => [...xs, { label: '', unit: 'قطعة', unitPrice: 0 }])} className="rounded-lg border border-dashed border-slate-400 px-3 py-1.5 text-sm font-bold text-slate-700">+ أضف خانة</button>
              <div className="space-y-2 rounded-xl bg-[#0f2040]/5 px-4 py-3" data-testid="sio-tiers">
                <b className="text-slate-800">أقل سعر للشغل (درجات)</b>
                <p className="text-xs text-slate-500">الفاتورة تطلع بأصغر درجة توصل لمجموع الخانات. مثلاً 14,000 و 24,000: إذا المجموع 10,000 ← 14,000، وإذا 16,000 ← 24,000، وإذا فات آخر درجة ينحسب مثل ما هو. المواد تنضاف فوگه.</p>
                {tiers.map((v, i) => (
                  <div key={i} className="flex items-center gap-2">
                    <span className="w-16 text-sm text-slate-600">درجة {i + 1}</span>
                    <input type="number" min={0} value={v || ''} onChange={(e) => setTiers((xs) => xs.map((x, k) => (k === i ? Number(e.target.value) : x)))} className={`${inp} w-36`} dir="ltr" />
                    <small className="text-slate-500">د.ع</small>
                    <button type="button" onClick={() => setTiers((xs) => xs.filter((_, k) => k !== i))} className="rounded-lg px-2 py-1 text-rose-600 hover:bg-rose-50" aria-label="شيل الدرجة">✕</button>
                  </div>
                ))}
                <button type="button" onClick={() => setTiers((xs) => [...xs, 0])} className="rounded-lg border border-dashed border-slate-400 px-3 py-1.5 text-sm font-bold text-slate-700">+ أضف درجة</button>
              </div>
              <button type="button" disabled={busy || !dirty} onClick={() => void save()} className="rounded-xl bg-[#0f2040] px-6 py-2.5 font-bold text-white disabled:opacity-40">{busy ? 'يحفظ…' : 'احفظ'}</button>
            </div>

            <div className="space-y-3 rounded-2xl bg-white p-4 shadow-sm">
              <h3 className="font-bold text-slate-800">🧪 جرّب الحساب {dirty && <small className="text-xs font-normal text-amber-700">(احفظ أول حتى تجرّب التعديلات)</small>}</h3>
              {cur.options.map((o) => (
                <label key={o.id} className="flex items-center justify-between gap-3 text-sm">
                  <span>{o.label} <small className="text-slate-500">({iqd(o.unitPrice)} / {o.unit})</small></span>
                  <input type="number" min={0} value={qty[o.id!] ?? ''} onChange={(e) => setQty({ ...qty, [o.id!]: Number(e.target.value) })} placeholder={`كم ${o.unit}`} className={`${inp} w-28`} dir="ltr" />
                </label>
              ))}
              <label className="flex items-center justify-between gap-3 text-sm"><span>المواد (اختياري)</span><input type="number" min={0} value={materials || ''} onChange={(e) => setMaterials(Number(e.target.value))} className={`${inp} w-28`} dir="ltr" /></label>
              <button type="button" disabled={!cur.options.length} onClick={() => void test()} className="rounded-xl bg-emerald-700 px-5 py-2 font-bold text-white disabled:opacity-40">احسب</button>
              {quote && (
                <div className="rounded-xl bg-slate-50 p-3 text-sm" data-testid="sio-quote">
                  {quote.lines.map((l) => <p key={l.label} className="flex justify-between"><span>{l.label} × {l.qty} {l.unit}</span><b dir="ltr">{iqd(l.total)}</b></p>)}
                  <p className="mt-2 flex justify-between border-t pt-2"><span>مجموع الخانات</span><b dir="ltr">{iqd(quote.sum)}</b></p>
                  <p className="flex justify-between"><span>الشغل {quote.minUsed && <small className="text-amber-700">(طلعت درجة {iqd(quote.minWork)})</small>}</span><b dir="ltr">{iqd(quote.work)}</b></p>
                  {quote.materials > 0 && <p className="flex justify-between"><span>المواد</span><b dir="ltr">{iqd(quote.materials)}</b></p>}
                  <p className="mt-1 flex justify-between text-base text-[#0f2040]"><span className="font-bold">المجموع</span><b dir="ltr">{iqd(quote.total)}</b></p>
                </div>
              )}
            </div>
          </section>
        )}
      </div>
    </div>
  )
}
