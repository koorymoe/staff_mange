import { useEffect, useState } from 'react'
import { api, type RevolvingFund, type RevolvingFundTxn, type EmployeeFundBalance, type Employee, type ProcurementRequest, fileUrl } from '../api'
import { useSession } from '../session'

/**
 * الدوار — شاشة المحاسب.
 *
 * الدورة: المحاسب يغذّي الدوار → يسلّم موظف مبلغ (ينزل من الدوار ويطلع دَين
 * برقبة الموظف) → الموظف يشتري ويرفع صورة الوصل ويرجّع الباقي → المحاسب
 * يدقّق ويوافق → وقتها بس رصيد الموظف يتصفّر والمرتجع يرجع للدوار.
 */

const money = (n: number) => n.toLocaleString('en-US') + ' د.ع'

type Tab = 'funds' | 'balances' | 'review' | 'discharge' | 'log'

export default function RevolvingFundPage() {
  // قيمة مبالغ الدوار نفسها مو شغل يومي: تعديل الرصيد والتغذية لمدير
  // النظام والمالك، أو لمن ينطونه صلاحية revolving_fund_amount صراحةً.
  // الباقي (الصرف والتدقيق) يبقى بصلاحية revolving_fund مثل ما هو.
  const { permissions, employee } = useSession()
  const canEditAmount = employee?.role === 'ADMIN' || employee?.role === 'OWNER' || permissions.includes('revolving_fund_amount')
  // التخريج يرجّع فلوس للدوار — صلاحية مستقلة عن التدقيق اليومي
  const canDischarge = employee?.role === 'ADMIN' || employee?.role === 'OWNER'
    || employee?.role === 'FINANCE' || permissions.includes('fund_discharge')

  const [tab, setTab] = useState<Tab>('funds')
  const [funds, setFunds] = useState<RevolvingFund[]>([])
  const [balances, setBalances] = useState<EmployeeFundBalance[]>([])
  const [pending, setPending] = useState<RevolvingFundTxn[]>([])
  const [log, setLog] = useState<RevolvingFundTxn[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  // نماذج
  const [topupFor, setTopupFor] = useState<RevolvingFund | null>(null)
  const [topupAmount, setTopupAmount] = useState('')
  const [editFor, setEditFor] = useState<RevolvingFund | null>(null)
  const [editBalance, setEditBalance] = useState('')
  const [disburseOpen, setDisburseOpen] = useState(false)
  const [dFund, setDFund] = useState('')
  const [dEmployee, setDEmployee] = useState('')
  const [dAmount, setDAmount] = useState('')
  const [dNotes, setDNotes] = useState('')
  const [dRequestId, setDRequestId] = useState('')
  const [openRequests, setOpenRequests] = useState<ProcurementRequest[]>([])
  const [accounts, setAccounts] = useState<string[]>([])
  const [dischargeFor, setDischargeFor] = useState<RevolvingFundTxn | null>(null)
  const [dischargeAccount, setDischargeAccount] = useState('')
  const [dischargeNote, setDischargeNote] = useState('')
  const [viewReceipt, setViewReceipt] = useState<RevolvingFundTxn | null>(null)
  const [reviewNote, setReviewNote] = useState('')
  const [busy, setBusy] = useState(false)

  const load = () => {
    Promise.all([
      api.getFunds(),
      api.getFundBalances(),
      api.getFundTransactions({ status: 'PENDING' }),
      api.getFundTransactions(),
      api.getEmployees().catch(() => [] as Employee[]),
      api.getProcurementRequests().catch(() => [] as ProcurementRequest[]),
      api.getDischargeAccounts().catch(() => [] as string[]),
    ])
      .then(([f, b, p, l, e, reqs, accs]) => {
        setFunds(f); setBalances(b); setPending(p); setLog(l); setEmployees(e)
        setOpenRequests((reqs ?? []).filter((r) => r.status === 'PENDING' || r.status === 'IN_PROGRESS'))
        setAccounts(accs ?? [])
      })
      .catch((e) => setError(e instanceof Error ? e.message : 'تعذر تحميل بيانات الدوار'))
      .finally(() => setLoading(false))
  }
  useEffect(load, [])

  const doTopup = async () => {
    if (!topupFor || !Number(topupAmount)) return
    setBusy(true)
    try { await api.topupFund(topupFor.id, Number(topupAmount)); setTopupFor(null); setTopupAmount(''); load() }
    catch (e) { alert(e instanceof Error ? e.message : 'خطأ') } finally { setBusy(false) }
  }

  const doEdit = async () => {
    if (!editFor) return
    setBusy(true)
    try { await api.updateFund(editFor.id, { balance: Number(editBalance) }); setEditFor(null); load() }
    catch (e) { alert(e instanceof Error ? e.message : 'خطأ') } finally { setBusy(false) }
  }

  const doDisburse = async () => {
    if (!dFund || !dEmployee || !Number(dAmount)) { alert('اختر الدوار والموظف واكتب المبلغ'); return }
    setBusy(true)
    try {
      await api.disburseFund({
        fundId: dFund, employeeId: dEmployee, amount: Number(dAmount),
        requestId: dRequestId || null, notes: dNotes || null,
      })
      setDisburseOpen(false); setDAmount(''); setDNotes(''); setDEmployee(''); setDRequestId('')
      load()
    } catch (e) { alert(e instanceof Error ? e.message : 'خطأ') } finally { setBusy(false) }
  }

  const doReview = async (t: RevolvingFundTxn, approve: boolean) => {
    if (!approve && !reviewNote.trim() && !confirm('ترفض بدون سبب؟')) return
    setBusy(true)
    try { await api.reviewFundSettlement(t.id, approve, reviewNote || undefined); setViewReceipt(null); setReviewNote(''); load() }
    catch (e) { alert(e instanceof Error ? e.message : 'خطأ') } finally { setBusy(false) }
  }

  const doDischarge = async () => {
    if (!dischargeFor || !dischargeAccount.trim()) { alert('حدد على أي حساب انخرجت المادة'); return }
    setBusy(true)
    try {
      await api.dischargeSettlement(dischargeFor.id, { account: dischargeAccount.trim(), note: dischargeNote || null })
      setDischargeFor(null); setDischargeAccount(''); setDischargeNote(''); load()
    } catch (e) { alert(e instanceof Error ? e.message : 'خطأ') } finally { setBusy(false) }
  }

  if (loading) return <div dir="rtl" className="p-8 text-center text-slate-500">جاري التحميل...</div>
  if (error) return <div dir="rtl" className="rounded-xl bg-red-50 p-6 text-center text-red-700">{error}</div>

  const totalOutstanding = balances.reduce((s, b) => s + b.outstanding, 0)
  // التسويات المدققة الي بيها مصروف وما انخرجت — الدوار ناطر فلوسها
  const awaitingDischarge = log.filter((t) => t.awaitingDischarge)
  const totalAwaiting = funds.reduce((s, f) => s + (f.awaitingDischargeTotal || 0), 0)

  return (
    <div dir="rtl" className="space-y-6">
      <div className="rounded-2xl p-6 shadow-sm" style={{ backgroundColor: '#1a3a5c' }}>
        <h1 className="text-2xl font-bold text-white">💵 الدوار</h1>
        <p className="mt-1 text-sm text-blue-200">
          مبلغ دوّار للعمل. الموظف ياخذ، يشتري، يرفع صورة الوصل، ويرجّع الباقي — ورصيده ما يتصفّر إلا بموافقتك.
        </p>
      </div>

      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        {funds.map((f) => (
          <div key={f.id} className="rounded-2xl bg-white p-5 shadow-sm">
            <p className="text-sm font-medium text-slate-500">{f.name}</p>
            <p className="mt-1 text-2xl font-bold" style={{ color: 'var(--t-title)' }}>{money(f.balance)}</p>
            <p className="mt-1 text-xs text-amber-700">بيد الموظفين: {money(f.outstandingTotal)}</p>
            {f.awaitingDischargeTotal > 0 && (
              <p className="mt-0.5 text-xs text-violet-700">ناطر التخريج: {money(f.awaitingDischargeTotal)}</p>
            )}
          </div>
        ))}
        <div className="rounded-2xl border-2 border-amber-300 bg-amber-50 p-5">
          <p className="text-sm font-medium text-amber-800">إجمالي المطلوب من الموظفين</p>
          <p className="mt-1 text-2xl font-bold text-amber-900">{money(totalOutstanding)}</p>
        </div>
        <div className="rounded-2xl border-2 border-red-300 bg-red-50 p-5">
          <p className="text-sm font-medium text-red-800">تسويات تنتظر تدقيقك</p>
          <p className="mt-1 text-2xl font-bold text-red-900">{pending.length}</p>
        </div>
        {/* الي الدوار ناطره منك: مصروف مدقّق وما انخرّج بعد */}
        <div className="rounded-2xl border-2 border-violet-300 bg-violet-50 p-5">
          <p className="text-sm font-medium text-violet-800">الدوار ناطر تخريجك</p>
          <p className="mt-1 text-2xl font-bold text-violet-900">{money(totalAwaiting)}</p>
          <p className="mt-0.5 text-xs text-violet-700">{awaitingDischarge.length} مادة ما انخرجت</p>
        </div>
      </div>

      <div className="flex flex-wrap gap-2">
        {([['funds', 'الدوارات'], ['balances', 'أرصدة الموظفين'], ['review', `التدقيق (${pending.length})`], ['discharge', `التخريج (${awaitingDischarge.length})`], ['log', 'سجل الحركات']] as [Tab, string][]).map(([k, label]) => (
          <button key={k} onClick={() => setTab(k)}
            className={`rounded-xl px-5 py-2.5 text-sm font-bold transition-colors ${tab === k ? 'text-white' : 'bg-white text-slate-600 hover:bg-slate-50'}`}
            style={tab === k ? { backgroundColor: '#1a3a5c' } : undefined}>
            {label}
          </button>
        ))}
      </div>

      {tab === 'funds' && (
        <div className="space-y-4">
          <button onClick={() => { setDisburseOpen(true); setDFund(funds[0]?.id || '') }}
            className="rounded-xl bg-emerald-700 px-6 py-3 font-bold text-white hover:bg-emerald-800">
            ➕ تسليم مبلغ لموظف
          </button>
          {funds.map((f) => (
            <div key={f.id} className="rounded-2xl bg-white p-6 shadow-sm">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <h3 className="text-lg font-bold" style={{ color: 'var(--t-title)' }}>{f.name}</h3>
                  <p className="mt-1 text-sm text-slate-500">الرصيد الحالي: <span className="font-bold text-slate-800">{money(f.balance)}</span></p>
                  <p className="text-sm text-slate-500">بيد الموظفين وما انتسوّى: <span className="font-bold text-amber-700">{money(f.outstandingTotal)}</span></p>
                  <p className="text-sm text-slate-500">ناطر تخريج المحاسب: <span className="font-bold text-violet-700">{money(f.awaitingDischargeTotal || 0)}</span></p>
                </div>
                {canEditAmount && (
                  <div className="flex gap-2">
                    <button onClick={() => { setTopupFor(f); setTopupAmount('') }}
                      className="rounded-lg bg-emerald-700 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-800">تغذية</button>
                    <button onClick={() => { setEditFor(f); setEditBalance(String(f.balance)) }}
                      className="rounded-lg border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50">تعديل المبلغ</button>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {tab === 'balances' && (
        <div className="overflow-hidden rounded-2xl bg-white shadow-sm">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-slate-200 text-right">
              <thead style={{ backgroundColor: '#1a3a5c' }} className="text-white">
                <tr>
                  <th className="px-4 py-3 text-sm font-semibold">الموظف</th>
                  <th className="px-4 py-3 text-sm font-semibold">استلم</th>
                  <th className="px-4 py-3 text-sm font-semibold">سوّى</th>
                  <th className="px-4 py-3 text-sm font-semibold">مطلوب منه</th>
                  <th className="px-4 py-3 text-sm font-semibold">تسويات معلّقة</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {balances.map((b) => (
                  <tr key={b.employeeId} className="hover:bg-slate-50">
                    <td className="px-4 py-3 font-medium">{b.employeeName}<span className="block text-xs text-slate-400">{b.jobTitle}</span></td>
                    <td className="px-4 py-3 text-slate-600">{money(b.totalTaken)}</td>
                    <td className="px-4 py-3 text-slate-600">{money(b.totalSettled)}</td>
                    {/* ⚠️ مطلوب منه — **مفصَّلاً لكل دوار**.
                        الرقم المجموع وحده يخلّي المحاسب ما يعرف وين
                        ترجع الفلوس: موظف «عليه ٣٠ ألف» ممكن تكون ١٠
                        من دوار الطاقة و٢٠ من دوار الشعبة، وكل مبلغ
                        يرجع لدواره هو. */}
                    <td className={`px-4 py-3 font-bold ${b.outstanding > 0 ? 'text-red-600' : 'text-emerald-600'}`}>
                      {money(b.outstanding)}
                      {(b.funds || []).filter((f) => f.outstanding > 0.001).length > 1 && (
                        <span className="mt-1 block space-y-0.5 text-xs font-normal text-slate-500">
                          {(b.funds || []).filter((f) => f.outstanding > 0.001).map((f) => (
                            <span key={f.fundId} className="block">{f.fundName}: {money(f.outstanding)}</span>
                          ))}
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3">{b.pendingSettlements > 0
                      ? <span className="rounded-full bg-amber-100 px-3 py-1 text-xs font-bold text-amber-800">{b.pendingSettlements}</span>
                      : <span className="text-slate-400">—</span>}</td>
                  </tr>
                ))}
                {balances.length === 0 && <tr><td colSpan={5} className="p-8 text-center text-slate-400">ماكو ولا موظف ساحب من الدوار</td></tr>}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {tab === 'review' && (
        <div className="space-y-3">
          {pending.map((t) => (
            <div key={t.id} className="rounded-2xl border-2 border-amber-300 bg-amber-50 p-5">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <p className="font-bold text-amber-900">{t.employeeName}</p>
                  <p className="mt-1 text-sm text-slate-600">
                    صرف <span className="font-bold">{money(t.spentAmount)}</span> · رجّع <span className="font-bold">{money(t.returnedAmount)}</span>
                    {' · '}المجموع <span className="font-bold">{money(t.spentAmount + t.returnedAmount)}</span>
                  </p>
                  <p className="text-xs text-slate-500">{t.fundName} · {new Date(t.createdAt).toLocaleDateString('ar-IQ')}</p>
                  {t.notes && <p className="mt-1 text-sm text-slate-700">📝 {t.notes}</p>}
                </div>
                <div className="flex gap-2">
                  {t.receiptImage && (
                    <button onClick={() => { setViewReceipt(t); setReviewNote('') }}
                      className="rounded-lg bg-white px-4 py-2 text-sm font-medium text-slate-700 shadow-sm hover:bg-slate-50">
                      🧾 شوف الوصل
                    </button>
                  )}
                  <button disabled={busy} onClick={() => doReview(t, true)}
                    className="rounded-lg bg-emerald-700 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-800 disabled:opacity-50">✔ وافق</button>
                  <button disabled={busy} onClick={() => doReview(t, false)}
                    className="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-50">✕ ارفض</button>
                </div>
              </div>
            </div>
          ))}
          {pending.length === 0 && <p className="rounded-2xl bg-white p-8 text-center text-slate-400 shadow-sm">ماكو تسويات تنتظر التدقيق</p>}
        </div>
      )}

      {/* ═══ التخريج: الخطوة الي ترجّع المبلغ المصروف للدوار ═══
          الموظف رجّع الزايد بالتسوية، والمبلغ الي انصرف فعلاً يبقى ناقص
          من الدوار لحد ما المحاسب يخرّج المادة ويحدد على أي حساب. */}
      {tab === 'discharge' && (
        <div className="space-y-3">
          <div className="rounded-2xl bg-white p-4 text-sm text-slate-600 shadow-sm">
            هاي المبالغ انصرفت فعلاً وانوافق عليها — الدوار ناقصها لحد ما
            تخرّج المادة على النظام المحاسبي وتحدد على أي حساب انحسبت.
            بضغطة «تم التخريج» يرجع <b>المبلغ المصروف</b> للدوار (المرتجع
            رجع من قبل بالتدقيق).
          </div>
          {awaitingDischarge.map((t) => (
            <div key={t.id} className="rounded-2xl border-2 border-violet-300 bg-violet-50 p-5">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <p className="font-bold text-violet-900">{t.employeeName}</p>
                  <p className="mt-1 text-sm text-slate-600">
                    الدوار ناطر <span className="font-bold text-violet-800">{money(t.spentAmount)}</span>
                    {t.returnedAmount > 0 && (
                      <span className="text-slate-500"> · رجّع {money(t.returnedAmount)} بالتدقيق</span>
                    )}
                  </p>
                  <p className="text-xs text-slate-500">{t.fundName} · {new Date(t.createdAt).toLocaleDateString('ar-IQ')}</p>
                  {t.notes && <p className="mt-1 text-sm text-slate-700">📝 {t.notes}</p>}
                </div>
                <div className="flex gap-2">
                  {t.receiptImage && (
                    <button onClick={() => { setViewReceipt(t); setReviewNote('') }}
                      className="rounded-lg bg-white px-4 py-2 text-sm font-medium text-slate-700 shadow-sm hover:bg-slate-50">
                      🧾 شوف الوصل
                    </button>
                  )}
                  {canDischarge ? (
                    <button disabled={busy}
                      onClick={() => { setDischargeFor(t); setDischargeAccount(''); setDischargeNote('') }}
                      className="rounded-lg bg-violet-600 px-4 py-2 text-sm font-bold text-white hover:bg-violet-700 disabled:opacity-50">
                      📤 تم التخريج
                    </button>
                  ) : (
                    <span className="self-center text-xs text-slate-400">التخريج بيد المحاسب</span>
                  )}
                </div>
              </div>
            </div>
          ))}
          {awaitingDischarge.length === 0 && (
            <p className="rounded-2xl bg-white p-8 text-center text-slate-400 shadow-sm">
              ماكو مادة ناطرة التخريج — الدوار مكتمل ✔
            </p>
          )}
        </div>
      )}

      {/* نافذة التخريج: الحساب يتكتب أو ينتخب من الحسابات المستخدمة سابقاً */}
      {dischargeFor && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" dir="rtl">
          <div className="w-full max-w-md space-y-4 rounded-2xl bg-white p-6 shadow-xl">
            <h3 className="text-lg font-bold" style={{ color: 'var(--t-title)' }}>تخريج مادة — {dischargeFor.employeeName}</h3>
            <p className="rounded-lg bg-violet-50 p-3 text-sm text-violet-900">
              راح يرجع للدوار <b>{money(dischargeFor.spentAmount)}</b>
            </p>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-700">على أي حساب انخرجت؟</label>
              <input
                list="discharge-accounts"
                value={dischargeAccount}
                onChange={(e) => setDischargeAccount(e.target.value)}
                placeholder="اكتب الحساب أو اختاره من القائمة"
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-violet-500"
              />
              <datalist id="discharge-accounts">
                {accounts.map((a) => <option key={a} value={a} />)}
              </datalist>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-700">ملاحظة (اختياري)</label>
              <input
                value={dischargeNote}
                onChange={(e) => setDischargeNote(e.target.value)}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-violet-500"
              />
            </div>
            <div className="flex justify-end gap-2">
              <button onClick={() => setDischargeFor(null)}
                className="rounded-lg px-4 py-2 text-sm text-slate-600 hover:bg-slate-100">إلغاء</button>
              <button disabled={busy} onClick={doDischarge}
                className="rounded-lg bg-violet-600 px-6 py-2 text-sm font-bold text-white hover:bg-violet-700 disabled:opacity-50">
                {busy ? 'جاري...' : 'تأكيد التخريج'}
              </button>
            </div>
          </div>
        </div>
      )}

      {tab === 'log' && (
        <div className="overflow-hidden rounded-2xl bg-white shadow-sm">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-slate-200 text-right">
              <thead style={{ backgroundColor: '#1a3a5c' }} className="text-white">
                <tr>
                  <th className="px-4 py-3 text-sm font-semibold">التاريخ</th>
                  <th className="px-4 py-3 text-sm font-semibold">النوع</th>
                  <th className="px-4 py-3 text-sm font-semibold">الدوار</th>
                  <th className="px-4 py-3 text-sm font-semibold">الموظف</th>
                  <th className="px-4 py-3 text-sm font-semibold">المبلغ</th>
                  <th className="px-4 py-3 text-sm font-semibold">الحالة</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {log.map((t) => (
                  <tr key={t.id} className="hover:bg-slate-50">
                    <td className="px-4 py-3 text-sm text-slate-500">{new Date(t.createdAt).toLocaleDateString('ar-IQ')}</td>
                    <td className="px-4 py-3 text-sm">{t.kindLabel}</td>
                    <td className="px-4 py-3 text-sm text-slate-600">{t.fundName}</td>
                    <td className="px-4 py-3 text-sm">{t.employeeName || '—'}</td>
                    <td className="px-4 py-3 text-sm font-medium">
                      {t.kind === 'SETTLEMENT' ? `صرف ${money(t.spentAmount)} · رجّع ${money(t.returnedAmount)}` : money(t.amount)}
                    </td>
                    <td className="px-4 py-3">
                      <span className={`rounded-full px-3 py-1 text-xs font-bold ${
                        t.status === 'APPROVED' ? 'bg-emerald-100 text-emerald-800'
                        : t.status === 'PENDING' ? 'bg-amber-100 text-amber-800' : 'bg-red-100 text-red-800'}`}>
                        {t.statusLabel}
                      </span>
                    </td>
                  </tr>
                ))}
                {log.length === 0 && <tr><td colSpan={6} className="p-8 text-center text-slate-400">ماكو حركات</td></tr>}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* تغذية */}
      {topupFor && (
        <Modal onClose={() => setTopupFor(null)} title={`تغذية ${topupFor.name}`}>
          <input type="number" value={topupAmount} onChange={(e) => setTopupAmount(e.target.value)}
            placeholder="المبلغ" className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />
          <ModalActions busy={busy} onConfirm={doTopup} onCancel={() => setTopupFor(null)} confirmLabel="تغذية" />
        </Modal>
      )}

      {/* تعديل المبلغ */}
      {editFor && (
        <Modal onClose={() => setEditFor(null)} title={`تعديل رصيد ${editFor.name}`}>
          <input type="number" value={editBalance} onChange={(e) => setEditBalance(e.target.value)}
            className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />
          <ModalActions busy={busy} onConfirm={doEdit} onCancel={() => setEditFor(null)} confirmLabel="حفظ" />
        </Modal>
      )}

      {/* تسليم لموظف */}
      {disburseOpen && (
        <Modal onClose={() => setDisburseOpen(false)} title="تسليم مبلغ لموظف">
          <label className="mb-1 block text-sm font-medium text-slate-600">الدوار</label>
          <select value={dFund} onChange={(e) => setDFund(e.target.value)}
            className="mb-3 w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500">
            {funds.map((f) => <option key={f.id} value={f.id}>{f.name} — رصيد {money(f.balance)}</option>)}
          </select>
          {/* التغذية على طلب مادة: اختيار الطلب يعبّي اسم الموظف الطالب
              والمبلغ التقريبي تلقائياً — بدل ما يدوّر المحاسب بيده. */}
          <label className="mb-1 block text-sm font-medium text-slate-600">على أي طلب؟ (اختياري)</label>
          <select value={dRequestId}
            onChange={(e) => {
              const id = e.target.value
              setDRequestId(id)
              const req = openRequests.find((r) => r.id === id)
              if (req) {
                setDEmployee(req.requestedBy.id)
                const est = req.items.reduce((sum, i) => sum + (i.unitPrice || 0) * i.quantity, 0)
                if (est > 0 && !dAmount) setDAmount(String(est))
              }
            }}
            className="mb-3 w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500">
            <option value="">بدون طلب — تسليم مباشر</option>
            {openRequests.map((r) => (
              <option key={r.id} value={r.id}>
                {r.code} · {r.requestedBy.name} · {r.items.map((i) => i.productName).join('، ').slice(0, 40)}
              </option>
            ))}
          </select>
          <label className="mb-1 block text-sm font-medium text-slate-600">الموظف</label>
          <select value={dEmployee} onChange={(e) => setDEmployee(e.target.value)}
            className="mb-3 w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500">
            <option value="">اختر الموظف</option>
            {employees.map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
          </select>
          <label className="mb-1 block text-sm font-medium text-slate-600">المبلغ</label>
          <input type="number" value={dAmount} onChange={(e) => setDAmount(e.target.value)}
            className="mb-3 w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />
          <label className="mb-1 block text-sm font-medium text-slate-600">ملاحظات</label>
          <textarea value={dNotes} onChange={(e) => setDNotes(e.target.value)} rows={2}
            className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />
          <ModalActions busy={busy} onConfirm={doDisburse} onCancel={() => setDisburseOpen(false)} confirmLabel="تسليم" />
        </Modal>
      )}

      {/* عرض الوصل + التدقيق */}
      {viewReceipt && (
        <Modal onClose={() => setViewReceipt(null)} title={`وصل ${viewReceipt.employeeName}`} wide>
          <p className="mb-3 text-sm text-slate-600">
            المبلغ المصروف: <span className="font-bold">{money(viewReceipt.spentAmount)}</span>
            {' · '}المرتجع: <span className="font-bold">{money(viewReceipt.returnedAmount)}</span>
          </p>
          {viewReceipt.receiptImage && (
            <img src={fileUrl(viewReceipt.receiptImage)} alt="صورة الوصل"
              className="mb-3 max-h-[50vh] w-full rounded-lg border border-slate-200 object-contain" />
          )}
          <textarea value={reviewNote} onChange={(e) => setReviewNote(e.target.value)} rows={2}
            placeholder="ملاحظة التدقيق (اختياري)"
            className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />
          <div className="mt-4 flex gap-3">
            <button disabled={busy} onClick={() => doReview(viewReceipt, true)}
              className="flex-1 rounded-lg bg-emerald-700 px-4 py-3 font-medium text-white disabled:opacity-50">✔ وافق وصفّر رصيده</button>
            <button disabled={busy} onClick={() => doReview(viewReceipt, false)}
              className="rounded-lg bg-red-600 px-4 py-3 font-medium text-white disabled:opacity-50">✕ ارفض</button>
          </div>
        </Modal>
      )}
    </div>
  )
}

function Modal({ title, children, onClose, wide }: { title: string; children: React.ReactNode; onClose: () => void; wide?: boolean }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={onClose}>
      <div dir="rtl" className={`w-full ${wide ? 'max-w-2xl' : 'max-w-md'} rounded-2xl bg-white p-6 shadow-xl`} onClick={(e) => e.stopPropagation()}>
        <h3 className="mb-4 text-lg font-bold" style={{ color: 'var(--t-title)' }}>{title}</h3>
        {children}
      </div>
    </div>
  )
}

function ModalActions({ busy, onConfirm, onCancel, confirmLabel }: { busy: boolean; onConfirm: () => void; onCancel: () => void; confirmLabel: string }) {
  return (
    <div className="mt-4 flex gap-3">
      <button disabled={busy} onClick={onConfirm} className="flex-1 rounded-lg px-4 py-3 font-medium text-white disabled:opacity-50" style={{ backgroundColor: '#1a3a5c' }}>
        {busy ? 'جاري...' : confirmLabel}
      </button>
      <button onClick={onCancel} className="rounded-lg border border-slate-300 px-4 py-3 font-medium text-slate-700">إلغاء</button>
    </div>
  )
}
