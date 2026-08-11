import { useEffect, useState } from 'react'
import { api, type EmployeeFundBalance, type RevolvingFundTxn } from '../api'

/**
 * رصيد الدوار مال الموظف — يظهر باللوحة الرئيسية حتى يشوف شكد مطلوب منه
 * للمحاسب، ويقدر يسوّي من نفس المكان.
 *
 * ما يظهر إلا إذا الموظف فعلاً أخذ من الدوار — ما نزحم لوحة موظف ما إله علاقة.
 */
const money = (n: number) => n.toLocaleString('en-US') + ' د.ع'

export default function MyFundBalance() {
  const [bal, setBal] = useState<EmployeeFundBalance | null>(null)
  const [txns, setTxns] = useState<RevolvingFundTxn[]>([])
  const [open, setOpen] = useState(false)
  const [spent, setSpent] = useState('')
  const [returned, setReturned] = useState('')
  const [notes, setNotes] = useState('')
  const [receipt, setReceipt] = useState<string>('')
  const [busy, setBusy] = useState(false)

  const load = () => {
    api.getMyFundBalance().then(setBal).catch(() => setBal(null))
    api.getMyFundTransactions().then(setTxns).catch(() => setTxns([]))
  }
  useEffect(load, [])

  // ما نعرض شي لموظف ما أخذ من الدوار أبداً
  if (!bal || (bal.totalTaken === 0 && bal.outstanding === 0)) return null

  const pickReceipt = (file: File) => {
    const reader = new FileReader()
    reader.onload = () => setReceipt(String(reader.result || ''))
    reader.readAsDataURL(file)
  }

  const submit = async () => {
    const s = Number(spent) || 0
    const r = Number(returned) || 0
    if (s + r <= 0) { alert('اكتب المبلغ المصروف أو المرتجع'); return }
    if (s > 0 && !receipt) { alert('لازم ترفع صورة الوصل للمبلغ المصروف'); return }
    // ═══ بيان الصرف إجباري ═══
    // الوصل يثبت إن الفلوس انصرفت بس ما يگول **على شنو**. بدونه
    // المحاسب لازم يتصل يسأل، وبعد أسبوع الموظف ما يتذكر.
    // ⚠️ الطول بالحروف مو بالبايتات — العربي حرفه بايتين وأكثر.
    if (s > 0 && [...notes.trim()].length < 5) {
      alert('اكتب على شنو انصرفت الفلوس — بيان مختصر يكفي')
      return
    }
    // الدوار الي أخذ منه — ناخذه من آخر عملية تسليم
    const lastDisburse = txns.find((t) => t.kind === 'DISBURSE')
    if (!lastDisburse) { alert('ماكو عملية تسليم مسجّلة'); return }
    setBusy(true)
    try {
      await api.submitFundSettlement({
        fundId: lastDisburse.fundId, spentAmount: s, returnedAmount: r,
        receiptImage: receipt || null, notes: notes || null,
      })
      setOpen(false); setSpent(''); setReturned(''); setNotes(''); setReceipt('')
      load()
      alert('انرفعت التسوية — تنتظر تدقيق المحاسب')
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر رفع التسوية')
    } finally { setBusy(false) }
  }

  const settled = bal.outstanding <= 0

  return (
    <>
      <div className={`rounded-2xl border-2 p-5 ${settled ? 'border-emerald-200 bg-emerald-50' : 'border-amber-300 bg-amber-50'}`}>
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <p className={`text-sm font-bold ${settled ? 'text-emerald-800' : 'text-amber-900'}`}>💵 رصيد الدوار</p>
            <p className={`mt-1 text-2xl font-bold ${settled ? 'text-emerald-700' : 'text-amber-900'}`}>
              {settled ? 'ما عليك شي' : `مطلوب منك ${money(bal.outstanding)}`}
            </p>
            <p className="mt-1 text-xs text-slate-600">
              استلمت {money(bal.totalTaken)} · سوّيت {money(bal.totalSettled)}
              {bal.pendingSettlements > 0 && ` · ${bal.pendingSettlements} تسوية تنتظر تدقيق المحاسب`}
            </p>
          </div>
          {!settled && (
            <button onClick={() => setOpen(true)}
              className="rounded-xl bg-amber-600 px-5 py-3 text-sm font-bold text-white hover:bg-amber-700">
              🧾 سوّي حسابك
            </button>
          )}
        </div>
      </div>

      {open && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={() => setOpen(false)}>
          <div dir="rtl" className="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-2xl bg-white p-6 shadow-xl" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-lg font-bold" style={{ color: '#1a3a5c' }}>تسوية الدوار</h3>
            <p className="mt-1 text-sm text-slate-600">الي بيدك: <span className="font-bold">{money(bal.outstanding)}</span></p>
            <p className="mt-2 rounded-lg bg-slate-50 p-3 text-xs text-slate-600">
              اكتب شكد صرفت وشكد راجع رجّعته، وارفع صورة الوصل. رصيدك ما يتصفّر إلا بعد ما المحاسب يدقّق ويوافق.
            </p>

            <label className="mt-4 mb-1 block text-sm font-medium text-slate-600">المبلغ المصروف (بالوصل)</label>
            <input type="number" value={spent} onChange={(e) => setSpent(e.target.value)}
              className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />

            <label className="mt-3 mb-1 block text-sm font-medium text-slate-600">المبلغ المرتجع (كاش)</label>
            <input type="number" value={returned} onChange={(e) => setReturned(e.target.value)}
              className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />

            {(Number(spent) || 0) + (Number(returned) || 0) > 0 && (
              <p className={`mt-2 text-sm font-medium ${
                Math.abs((Number(spent) || 0) + (Number(returned) || 0) - bal.outstanding) < 0.01
                  ? 'text-emerald-700' : 'text-amber-700'}`}>
                المجموع: {money((Number(spent) || 0) + (Number(returned) || 0))} من أصل {money(bal.outstanding)}
              </p>
            )}

            <label className="mt-3 mb-1 block text-sm font-medium text-slate-600">صورة الوصل</label>
            <input type="file" accept="image/*" capture="environment"
              onChange={(e) => { const f = e.target.files?.[0]; if (f) pickReceipt(f) }}
              className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right" />
            {receipt && <img src={receipt} alt="الوصل" className="mt-2 max-h-48 w-full rounded-lg border border-slate-200 object-contain" />}

            <label className="mt-3 mb-1 block text-sm font-medium text-slate-600">
              على شنو انصرفت الفلوس؟ <span className="text-red-600">*</span>
            </label>
            <textarea value={notes} onChange={(e) => setNotes(e.target.value)} rows={2}
              placeholder="مثال: شريت كيبل ٥٠ متر وموصلات لحجز B120"
              className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />

            <div className="mt-4 flex gap-3">
              <button disabled={busy} onClick={submit}
                className="flex-1 rounded-lg px-4 py-3 font-medium text-white disabled:opacity-50" style={{ backgroundColor: '#1a3a5c' }}>
                {busy ? 'جاري الرفع...' : 'ارفع التسوية'}
              </button>
              <button onClick={() => setOpen(false)} className="rounded-lg border border-slate-300 px-4 py-3 font-medium text-slate-700">إلغاء</button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}
