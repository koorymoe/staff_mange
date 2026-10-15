import { useEffect, useState } from 'react'
import { purchaseApi, type ProcurementRequest, type PurchaseAlert, type PurchaseControlRow, type PurchasePriceRow, type SupplierReviewInput } from '../api'

// ═══ رقابة المشتريات — قرار (ع) 10-15 ═══
// المشتري يسجّل الشراء ويقيّم المورد ببنود (أولها السعر) · المحاسب يثبّت ·
// سجل أسعار لكل مادة · وتنبيهات ماتركس: «اشترى من الأغلى» و«المورد رفع سعره».

const iqd = (n: number) => `${Math.round(n).toLocaleString('en-US')} د.ع`
const day = (s: string) => new Date(s).toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' })
const CRITERIA: [keyof Omit<SupplierReviewInput, 'note'>, string][] = [
  ['price', '💰 السعر'], ['quality', '✅ الجودة'], ['speed', '⚡ سرعة التجهيز'], ['punctuality', '⏰ الالتزام بالموعد'], ['dealing', '🤝 التعامل'],
]

interface Props {
  requests: ProcurementRequest[]
  canBuy: boolean // إداري الكميات / المدير — يقيّم المورد
  canVerify: boolean // المحاسب / المدير — يثبّت
  canWatch: boolean // المالك / المدير / المراقب / المحاسب — التنبيهات
  canDecide: boolean // المدير — يحسم التنبيه
  reviewFor: ProcurementRequest | null
  onReviewClose: () => void
}

export default function PurchaseControlPanel({ requests, canBuy, canVerify, canWatch, canDecide, reviewFor, onReviewClose }: Props) {
  const [tab, setTab] = useState<'todo' | 'prices' | 'alerts'>('todo')
  const [control, setControl] = useState<PurchaseControlRow[]>([])
  const [alerts, setAlerts] = useState<PurchaseAlert[]>([])
  const [q, setQ] = useState('')
  const [prices, setPrices] = useState<PurchasePriceRow[]>([])
  const [reload, setReload] = useState(0)
  const [reviewing, setReviewing] = useState<ProcurementRequest | null>(null)
  const open = reviewFor ?? reviewing

  useEffect(() => {
    let alive = true
    purchaseApi.control().then((r) => { if (alive) setControl(r) }).catch(() => {})
    if (canWatch) purchaseApi.alerts().then((r) => { if (alive) setAlerts(r) }).catch(() => {})
    return () => { alive = false }
  }, [reload, canWatch, requests])

  useEffect(() => {
    let alive = true
    const t = setTimeout(() => { purchaseApi.prices(q).then((r) => { if (alive) setPrices(r) }).catch(() => {}) }, 300)
    return () => { alive = false; clearTimeout(t) }
  }, [q, reload])

  const byId = new Map(control.map((c) => [c.id, c]))
  const fulfilled = requests.filter((r) => r.status === 'FULFILLED' && r.supplierId && byId.has(r.id))
  const toReview = canBuy ? fulfilled.filter((r) => !byId.get(r.id)!.myReview) : []
  const toVerify = canVerify ? fulfilled.filter((r) => !byId.get(r.id)!.verifiedAt) : []
  const openAlerts = alerts.filter((a) => !a.reviewedAt)
  const totalLoss = openAlerts.reduce((s, a) => s + a.loss, 0)

  // سجل الأسعار مجمّع لكل مادة
  const groups = new Map<string, PurchasePriceRow[]>()
  for (const p of prices) groups.set(p.key, [...(groups.get(p.key) ?? []), p])

  const verify = async (id: string) => {
    try { await purchaseApi.verify(id); setReload((n) => n + 1) } catch (e) { alert(e instanceof Error ? e.message : 'تعذر التثبيت') }
  }
  const decide = async (a: PurchaseAlert) => {
    const note = prompt('ملاحظتك على التنبيه (مثلاً: الأرخص ما چان متوفر · تم التنبيه للمشتري):')
    if (!note?.trim()) return
    try { await purchaseApi.reviewAlert(a.id, note.trim()); setReload((n) => n + 1) } catch (e) { alert(e instanceof Error ? e.message : 'تعذر الحفظ') }
  }

  const tabs: [typeof tab, string][] = [['todo', `📋 بانتظاري (${toReview.length + toVerify.length})`], ['prices', '📊 سجل الأسعار']]
  if (canWatch) tabs.push(['alerts', `🚨 تنبيهات ماتركس (${openAlerts.length})`])

  return (
    <div className="rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
      <div className="mb-3 flex flex-wrap items-center gap-2">
        <h2 className="ml-auto text-lg font-bold text-[#0f2040]">🛡️ رقابة المشتريات</h2>
        {tabs.map(([k, label]) => (
          <button key={k} onClick={() => setTab(k)} className={`rounded-lg px-3 py-1.5 text-sm font-bold ${tab === k ? (k === 'alerts' ? 'bg-red-600 text-white' : 'bg-[#2c5aad] text-white') : 'bg-slate-100 text-slate-600'}`}>{label}</button>
        ))}
      </div>

      {tab === 'todo' && (
        <div className="space-y-2">
          {toReview.length === 0 && toVerify.length === 0 && <p className="text-sm text-slate-400">ماكو شي ينتظرك ✓</p>}
          {toReview.map((r) => (
            <div key={'rv' + r.id} className="flex flex-wrap items-center gap-2 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm">
              <span className="font-bold">{r.code}</span> · {r.supplier?.name} · {r.totalCost != null ? iqd(r.totalCost) : ''}
              <button onClick={() => setReviewing(r)} className="mr-auto rounded-lg bg-amber-500 px-3 py-1 font-bold text-white">⭐ قيّم المورد</button>
            </div>
          ))}
          {toVerify.map((r) => (
            <div key={'vf' + r.id} className="flex flex-wrap items-center gap-2 rounded-lg border border-sky-200 bg-sky-50 px-3 py-2 text-sm">
              <span className="font-bold">{r.code}</span> · {r.supplier?.name} · اشتراه {r.fulfilledBy?.name ?? '—'} · {r.totalCost != null ? iqd(r.totalCost) : 'بلا مبلغ'}
              {byId.get(r.id)!.alerts > 0 && <span className="rounded bg-red-100 px-2 py-0.5 text-xs font-bold text-red-700">🚨 {byId.get(r.id)!.alerts} تنبيه سعر</span>}
              <button onClick={() => verify(r.id)} className="mr-auto rounded-lg bg-sky-600 px-3 py-1 font-bold text-white">✓ تثبيت المحاسب</button>
            </div>
          ))}
        </div>
      )}

      {tab === 'prices' && (
        <div className="space-y-3">
          <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="دوّر على مادة (كيبل، كاميرا...)" className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
          {groups.size === 0 && <p className="text-sm text-slate-400">ماكو أسعار مسجّلة لهذي المادة.</p>}
          {[...groups.entries()].slice(0, 50).map(([key, rows]) => {
            const min = Math.min(...rows.map((r) => r.lastPrice))
            return (
              <div key={key} className="rounded-lg border border-slate-200">
                <p className="border-b border-slate-100 bg-slate-50 px-3 py-1.5 text-sm font-bold">{rows[0].product}</p>
                <table className="w-full text-sm">
                  <thead><tr className="text-xs text-slate-400"><th className="px-3 text-right">المورد</th><th>آخر سعر</th><th>أقل سعر</th><th>مرات</th><th>آخر شراء</th></tr></thead>
                  <tbody>
                    {rows.map((r) => (
                      <tr key={r.supplier} className="border-t border-slate-100">
                        <td className="px-3 py-1">{r.supplier} {r.lastPrice === min && rows.length > 1 && <span className="text-xs font-bold text-emerald-600">الأرخص</span>}</td>
                        <td className={`text-center font-bold ${r.lastPrice > min * 1.1 ? 'text-red-600' : ''}`}>{iqd(r.lastPrice)}</td>
                        <td className="text-center">{iqd(r.minPrice)}</td>
                        <td className="text-center">{r.times}</td>
                        <td className="text-center">{day(r.lastAt)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )
          })}
        </div>
      )}

      {tab === 'alerts' && (
        <div className="space-y-2">
          {openAlerts.length > 0 && <p className="rounded-lg bg-red-50 px-3 py-2 text-sm font-bold text-red-700">مجموع الخسارة بالتنبيهات المفتوحة: {iqd(totalLoss)}</p>}
          {alerts.length === 0 && <p className="text-sm text-slate-400">ماكو تنبيهات — المشتريات بأسعار طبيعية ✓</p>}
          {alerts.map((a) => (
            <div key={a.id} className={`rounded-lg border px-3 py-2 text-sm ${a.reviewedAt ? 'border-slate-200 bg-slate-50 text-slate-500' : 'border-red-200 bg-red-50'}`}>
              <p className="font-bold">
                {a.kind === 'PRICIER' ? '🚨 اشترى من الأغلى' : '📈 المورد رفع سعره'} · {a.code} · «{a.product}» ×{a.quantity}
              </p>
              <p className="mt-0.5">
                {a.kind === 'PRICIER'
                  ? <>انشترت من <b>{a.supplier}</b> بـ<b>{iqd(a.unitPrice)}</b> — وموجودة عند <b>{a.refSupplier}</b> بـ<b>{iqd(a.refPrice)}</b></>
                  : <><b>{a.supplier}</b> رفع السعر من <b>{iqd(a.refPrice)}</b> إلى <b>{iqd(a.unitPrice)}</b></>}
                {' · '}الفرق <b className="text-red-700">{iqd(a.loss)}</b> · المشتري: {a.buyer ?? '—'} · {day(a.createdAt)}
              </p>
              {a.reviewedAt ? <p className="mt-1 text-xs">✓ انحسم: {a.reviewNote}</p>
                : canDecide && <button onClick={() => decide(a)} className="mt-1 rounded-lg bg-white px-3 py-1 text-xs font-bold text-red-700 ring-1 ring-red-300">احسم التنبيه</button>}
            </div>
          ))}
        </div>
      )}

      {open && <ReviewModal req={open} onClose={() => { setReviewing(null); onReviewClose() }} onSaved={() => { setReviewing(null); onReviewClose(); setReload((n) => n + 1) }} />}
    </div>
  )
}

function ReviewModal({ req, onClose, onSaved }: { req: ProcurementRequest; onClose: () => void; onSaved: () => void }) {
  const [v, setV] = useState<SupplierReviewInput>({ price: 0, quality: 0, speed: 0, punctuality: 0, dealing: 0 })
  const [note, setNote] = useState('')
  const [busy, setBusy] = useState(false)
  const ready = CRITERIA.every(([k]) => v[k] > 0)
  const save = async () => {
    setBusy(true)
    try { await purchaseApi.review(req.id, { ...v, note: note.trim() || undefined }); onSaved() } catch (e) { alert(e instanceof Error ? e.message : 'تعذر الحفظ') } finally { setBusy(false) }
  }
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" dir="rtl">
      <div className="w-full max-w-md space-y-3 rounded-2xl bg-white p-5">
        <p className="text-lg font-bold">⭐ قيّم المورد: {req.supplier?.name}</p>
        <p className="text-xs text-slate-500">طلب {req.code} — التقييم يساعد ماتركس يعرف منو أرخص وأحسن مورد لكل مادة.</p>
        {CRITERIA.map(([k, label]) => (
          <div key={k} className="flex items-center justify-between">
            <span className="text-sm font-medium">{label}</span>
            <div className="flex gap-1">
              {[1, 2, 3, 4, 5].map((n) => (
                <button key={n} onClick={() => setV({ ...v, [k]: n })} className={`text-2xl ${n <= v[k] ? 'text-amber-400' : 'text-slate-200'}`}>★</button>
              ))}
            </div>
          </div>
        ))}
        <textarea value={note} onChange={(e) => setNote(e.target.value)} placeholder="ملاحظة (اختياري)" rows={2} className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
        <div className="flex gap-2">
          <button disabled={!ready || busy} onClick={save} className="flex-1 rounded-lg bg-[#2c5aad] py-2 font-bold text-white disabled:opacity-50">حفظ التقييم</button>
          <button onClick={onClose} className="rounded-lg px-4 py-2 text-sm text-slate-600 hover:bg-slate-100">بعدين</button>
        </div>
      </div>
    </div>
  )
}
