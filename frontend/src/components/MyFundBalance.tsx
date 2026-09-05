import { useEffect, useState } from 'react'
import { api, type EmployeeFundBalance, type EmployeeFundLine } from '../api'

/**
 * رصيد الدوار مال الموظف — يظهر باللوحة الرئيسية حتى يشوف شكد مطلوب منه
 * للمحاسب، ويقدر يسوّي من نفس المكان.
 *
 * ما يظهر إلا إذا الموظف فعلاً أخذ من الدوار — ما نزحم لوحة موظف ما إله علاقة.
 *
 * ═══ ⚠️ التسوية صارت **لكل دوار على حِدة** ═══
 *
 * قبل، الشاشة تعرض رقماً واحداً مجموعاً وترفع تسوية وحدة، وتخمّن الدوار
 * من **آخر عملية تسليم**:
 *
 *     const lastDisburse = txns.find((t) => t.kind === 'DISBURSE')
 *     await api.submitFundSettlement({ fundId: lastDisburse.fundId, ... })
 *
 * يعني موظف أخذ ١٠ آلاف من دوار الطاقة و٢٠ ألف من دوار الشعبة، تسويته
 * كلها تنختم بالدوار الي دفع آخر مرة — فترجع الـ٣٠ ألف كلها لدوار واحد.
 * دوار يطلع عنده فائض وهمي وآخر عليه دَين وهمي، **للأبد**.
 *
 * هسه: بطاقة لكل دوار برصيده، والتسوية تنرفع لدوارها هي، والخادم يتحقق
 * من رصيد **ذاك الدوار** لا من المجموع.
 */
const money = (n: number) => n.toLocaleString('en-US') + ' د.ع'

export default function MyFundBalance() {
  const [bal, setBal] = useState<EmployeeFundBalance | null>(null)
  // الدوار الي دا يتسوّى الآن — بدل التخمين
  const [target, setTarget] = useState<EmployeeFundLine | null>(null)
  const [spent, setSpent] = useState('')
  const [returned, setReturned] = useState('')
  const [notes, setNotes] = useState('')
  const [receipt, setReceipt] = useState<string>('')
  const [busy, setBusy] = useState(false)

  const load = () => {
    api.getMyFundBalance().then(setBal).catch(() => setBal(null))
  }
  useEffect(load, [])

  // ما نعرض شي لموظف ما أخذ من الدوار أبداً
  if (!bal || (bal.totalTaken === 0 && bal.outstanding === 0)) return null

  const funds = bal.funds || []
  // الدوارات الي لسه عليه منهن شي
  const openFunds = funds.filter((f) => f.outstanding > 0.001)

  const pickReceipt = (file: File) => {
    const reader = new FileReader()
    reader.onload = () => setReceipt(String(reader.result || ''))
    reader.readAsDataURL(file)
  }

  const startSettle = (f: EmployeeFundLine) => {
    setTarget(f); setSpent(''); setReturned(''); setNotes(''); setReceipt('')
  }

  const submit = async () => {
    if (!target) return
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
    // ⚠️ ما نجمع أكثر من رصيد **هذا الدوار** — الخادم يرفضها هم،
    // بس نمنعها هنا حتى ما يضيع وقت الموظف بتعبئة نموذج كامل.
    if (s + r > target.outstanding + 0.001) {
      alert(`المبلغ أكبر من الي بيدك من ${target.fundName} (${money(target.outstanding)})`)
      return
    }
    setBusy(true)
    try {
      await api.submitFundSettlement({
        fundId: target.fundId, spentAmount: s, returnedAmount: r,
        receiptImage: receipt || null, notes: notes || null,
      })
      setTarget(null)
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

        {/* ⚠️ بطاقة لكل دوار — التسوية تنرفع لدوارها، حتى ترجع الفلوس
            للجزء الي طلعت منه بالضبط. */}
        {openFunds.length > 0 && (
          <div className="mt-4 space-y-2">
            {openFunds.length > 1 && (
              <p className="text-xs font-bold text-amber-900">
                عندك رصيد من {openFunds.length} دوارات — سوّي كل دوار لحاله
              </p>
            )}
            {openFunds.map((f) => (
              <div key={f.fundId}
                className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-amber-200 bg-white px-4 py-3">
                <div>
                  <p className="text-sm font-bold" style={{ color: 'var(--t-title)' }}>{f.fundName}</p>
                  <p className="mt-0.5 text-xs text-slate-600">
                    بيدك {money(f.outstanding)} · استلمت منه {money(f.totalTaken)}
                    {f.pendingSettlements > 0 && ` · ${f.pendingSettlements} تسوية تنتظر التدقيق`}
                  </p>
                </div>
                <button onClick={() => startSettle(f)}
                  className="rounded-xl bg-amber-600 px-4 py-2.5 text-sm font-bold text-white hover:bg-amber-700">
                  🧾 سوّي هذا الدوار
                </button>
              </div>
            ))}
          </div>
        )}
      </div>

      {target && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={() => setTarget(null)}>
          <div dir="rtl" className="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-2xl bg-white p-6 shadow-xl" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-lg font-bold" style={{ color: 'var(--t-title)' }}>تسوية — {target.fundName}</h3>
            <p className="mt-1 text-sm text-slate-600">
              الي بيدك من هذا الدوار: <span className="font-bold">{money(target.outstanding)}</span>
            </p>
            <p className="mt-2 rounded-lg bg-slate-50 p-3 text-xs text-slate-600">
              اكتب شكد صرفت وشكد راجع رجّعته، وارفع صورة الوصل. رصيدك ما يتصفّر إلا بعد ما المحاسب يدقّق ويوافق.
              {' '}⚠️ هاي التسوية تخص <span className="font-bold">{target.fundName}</span> وحده — الفلوس ترجع إله هو.
            </p>

            <label className="mt-4 mb-1 block text-sm font-medium text-slate-600">المبلغ المصروف (بالوصل)</label>
            <input type="number" value={spent} onChange={(e) => setSpent(e.target.value)}
              className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />

            <label className="mt-3 mb-1 block text-sm font-medium text-slate-600">المبلغ المرتجع (كاش)</label>
            <input type="number" value={returned} onChange={(e) => setReturned(e.target.value)}
              className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />

            {(Number(spent) || 0) + (Number(returned) || 0) > 0 && (
              <p className={`mt-2 text-sm font-medium ${
                Math.abs((Number(spent) || 0) + (Number(returned) || 0) - target.outstanding) < 0.01
                  ? 'text-emerald-700' : 'text-amber-700'}`}>
                المجموع: {money((Number(spent) || 0) + (Number(returned) || 0))} من أصل {money(target.outstanding)}
              </p>
            )}

            {/* ═══ صورة الوصل: كامرة **أو** من الاستوديو ═══
                ⚠️ كان `capture="environment"` — وهاي تجبر المتصفح يفتح
                الكامرة **مباشرة** وما تنطي خيار الاستوديو إطلاقاً.
                يعني الموظف الي صوّر الوصل قبل ساعة، أو استلم صورته
                بالواتساب، ما يكدر يرفعها — لازم يمسك الورقة ويصوّرها
                من جديد. وإذا الورقة ضاعت، ما يكدر يسوّي حسابه أصلاً.

                بلا capture، المتصفح يعرض الخيارين (كامرة/استوديو). */}
            <label className="mt-3 mb-1 block text-sm font-medium text-slate-600">
              صورة الوصل <span className="text-xs font-normal text-slate-400">— صوّرها هسه أو اختارها من الاستوديو</span>
            </label>
            <input type="file" accept="image/*"
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
              <button onClick={() => setTarget(null)} className="rounded-lg border border-slate-300 px-4 py-3 font-medium text-slate-700">إلغاء</button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}
