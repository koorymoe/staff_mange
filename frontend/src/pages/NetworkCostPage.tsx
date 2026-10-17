import { useEffect, useMemo, useState } from 'react'
import { api, type NetworkPriceItem, type NetworkCostResponse } from '../api'
import { useSession } from '../session'

// ═══ حساب كلفة الشبكات ═══
//
// قبلها الليدر يحسب فاتورة الشبكات براسه أو بالتلفون، وكل واحد يطلع
// برقم غير — ما كان بالنظام ولا سعر شبكات (حتى الاكسل ما بيه).
//
// ⚠️ الأسعار تجي من السيرفر مو مكتوبة هنا: صاحب العمل يعدّلها من
// «تعديل الأسعار» بدون ما نغيّر كود وننشر. والحساب نفسه يصير
// بالسيرفر، حتى نسخة قديمة من الواجهة ما تطلّع رقم مختلف.
//
// الفقرات وأنماطها:
//   تسليك كيبل  (TIERED)  لحد ٢٠ متر ١٢٬٠٠٠، وكل متر زايد ١٬٤٠٠
//   تنصيب راوتر (FLAT)    ١٧٬٠٠٠ للجهاز
//   تنظيم الراك (BRACKET) سعر البورت حسب حجم السويتج — سطر لكل سويتج
export default function NetworkCostPage() {
  const { employee } = useSession()
  const isAdmin = employee?.role === 'ADMIN'

  const [items, setItems] = useState<NetworkPriceItem[]>([])
  const [qty, setQty] = useState<Record<string, string>>({})
  const [discount, setDiscount] = useState('0')
  const [result, setResult] = useState<NetworkCostResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    api.getNetworkCostItems()
      .then(setItems)
      .catch((e) => setError(e instanceof Error ? e.message : 'تعذر جلب قائمة الأسعار'))
  }, [])

  // الرقم الفارغ صفر مو NaN — Number('') = 0 بس Number('ابc') = NaN
  const numOf = (v: string) => {
    const n = Number(v)
    return Number.isFinite(n) && n > 0 ? n : 0
  }

  const filled = useMemo(
    () => items.filter((it) => numOf(qty[it.id] || '')> 0),
    [items, qty],
  )

  const calculate = async () => {
    if (filled.length === 0) { setError('دخّل كمية وحدة على الأقل'); return }
    setLoading(true)
    setError(null)
    try {
      setResult(await api.calculateNetworkCost({
        lines: filled.map((it) => ({ itemId: it.id, quantity: numOf(qty[it.id]) })),
        discount: numOf(discount),
      }))
    } catch (e) {
      setError(e instanceof Error ? e.message : 'تعذر الحساب')
      setResult(null)
    } finally {
      setLoading(false)
    }
  }

  const num = (v: number) => v.toLocaleString('en-US')

  // شرح التسعيرة بسطر واحد تحت كل فقرة — الفني لازم يعرف ليش طلع الرقم
  const describe = (it: NetworkPriceItem) => {
    if (it.pricingMode === 'TIERED') {
      return `لحد ${num(it.includedQty)} ${it.unit} → ${num(it.basePrice)}، وكل ${it.unit} زايد ${num(it.extraPerUnit)}`
    }
    if (it.pricingMode === 'BRACKET') {
      return it.brackets
        .map((b) => (b.upTo > 0 ? `لحد ${num(b.upTo)}` : 'وأكثر') + ` → ${num(b.unitPrice)} لل${it.unit}`)
        .join(' • ')
    }
    return `${num(it.basePrice)} لل${it.unit} الواحد`
  }

  return (
    <div dir="rtl" className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-brand-900">حساب كلفة الشبكات</h2>
        <p className="mt-1 text-slate-500">
          دخّل الكميات وبس. الأسعار محفوظة بالنظام والحساب يصير بالسيرفر — نفس الرقم لكل واحد يفتح الاستمارة.
        </p>
      </div>

      {error && <p className="rounded-lg bg-red-50 p-4 text-red-600">{error}</p>}

      <div className="space-y-3 rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        {items.length === 0 && !error && <p className="text-slate-400">ماكو فقرات تسعيرة مفعّلة.</p>}

        {items.map((it) => (
          <div key={it.id} className="grid items-center gap-3 rounded-lg border border-slate-200 p-3 sm:grid-cols-[1fr_160px]">
            <div>
              <p className="font-bold text-[#0f2040]">{it.label}</p>
              <p className="text-[11px] text-slate-500">{describe(it)}</p>
              {it.note && <p className="text-[11px] text-slate-400">{it.note}</p>}
            </div>
            <div className="flex items-center gap-2">
              <input
                type="number"
                min={0}
                value={qty[it.id] ?? ''}
                onChange={(e) => setQty((p) => ({ ...p, [it.id]: e.target.value }))}
                placeholder="0"
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
              />
              <span className="shrink-0 text-xs text-slate-500">{it.unit}</span>
            </div>
          </div>
        ))}

        <div className="grid items-center gap-3 rounded-lg bg-slate-50 p-3 sm:grid-cols-[1fr_160px]">
          <p className="font-bold text-slate-700">الخصم</p>
          <input
            type="number"
            min={0}
            value={discount}
            onChange={(e) => setDiscount(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
          />
        </div>

        <button
          onClick={calculate}
          disabled={loading}
          className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-5 py-2.5 text-sm font-bold text-white disabled:opacity-50"
        >
          {loading ? 'جاري الحساب...' : 'احسب'}
        </button>
      </div>

      {result && (
        <div className="rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <h3 className="mb-3 text-lg font-bold text-[#0f2040]">النتيجة</h3>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-slate-200 text-right text-xs text-slate-500">
                  <th className="py-2">الفقرة</th>
                  <th className="py-2">الكمية</th>
                  <th className="py-2">التفصيل</th>
                  <th className="py-2">المبلغ</th>
                </tr>
              </thead>
              <tbody>
                {result.lines.map((ln) => (
                  <tr key={ln.itemId} className="border-b border-slate-100">
                    <td className="py-2 font-bold">{ln.label}</td>
                    <td className="py-2">{num(ln.quantity)} {ln.unit}</td>
                    <td className="py-2 text-xs text-slate-500">
                      {ln.extraPart > 0
                        ? `${num(ln.basePart)} مقطوعة + ${num(ln.extraQty)} × ${num(ln.extraPart / ln.extraQty)}`
                        : ln.unitPrice > 0
                          ? `${num(ln.quantity)} × ${num(ln.unitPrice)}`
                          : '—'}
                    </td>
                    <td className="py-2 font-bold">{num(ln.total)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <div className="mt-4 space-y-1 text-sm">
            <p>المجموع: <span className="font-bold">{num(result.subtotal)}</span></p>
            {result.discount > 0 && <p className="text-red-600">الخصم: {num(result.discount)}</p>}
            <p className="text-lg font-extrabold text-[#0f2040]">المبلغ النهائي: {num(result.finalAmount)} د.ع</p>
          </div>
        </div>
      )}

      {isAdmin && (
        <p className="text-xs text-slate-500">
          تعديل الأسعار من صفحة <a href="network-prices" className="font-bold text-brand-700 underline">تعديل أسعار الشبكات</a> — المالك ومدير النظام بس.
        </p>
      )}
    </div>
  )
}
