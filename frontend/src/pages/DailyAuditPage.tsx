import { useEffect, useState } from 'react'
import { api, type DailyAuditReport, type DailyAuditRow } from '../api'
import { useSession } from '../session'

/**
 * التدقيق اليومي.
 *
 * الفرق عن «تدقيق الحسابات» العام: هذا يمشي بيوم واحد. المحاسب يحدد
 * تاريخ، يشوف كل حجوزات ذاك اليوم — المكتملة وغير المكتملة، المدققة
 * والمحوّلة للرقابة — ويعرف من الصبح شكد المفروض يجمع.
 *
 * التدقيق نفسه (المبلغ وخيارات الخطأ) يصير من نفس الشاشة بالضبط.
 */

const money = (n: number) => n.toLocaleString('en-IQ') + ' د.ع'

function todayStr() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

const STATUS_LABEL: Record<string, string> = {
  COMPLETED: 'مكتمل',
  PENDING: 'معلّق',
  CONFIRMED: 'مثبّت',
  IN_PROGRESS: 'قيد التنفيذ',
  CANCELLED: 'ملغى',
}

function Tile({ label, value, hint, color }: { label: string; value: string; hint?: string; color: string }) {
  return (
    <div className="rounded-2xl bg-white p-5 shadow-sm">
      <p className="text-sm text-slate-500">{label}</p>
      <p className="mt-1 text-2xl font-bold" style={{ color }}>{value}</p>
      {hint && <p className="mt-1 text-xs text-slate-400">{hint}</p>}
    </div>
  )
}

export default function DailyAuditPage() {
  const [date, setDate] = useState(todayStr())
  const [rep, setRep] = useState<DailyAuditReport | null>(null)
  const [amounts, setAmounts] = useState<Record<string, string>>({})
  const [busy, setBusy] = useState<string | null>(null)
  // إرجاع الحجز للتدقيق: صلاحية مدير النظام حصراً (المالك يتطبّع لمدير
  // بالجلسة، فتشمله تلقائياً).
  const { employee } = useSession()
  const isAdmin = employee?.role === 'ADMIN'

  const load = (d: string) => {
    api.getDailyAudit(d).then(setRep).catch(() => setRep(null))
  }
  useEffect(() => { load(date) }, [date])

  // ═══ إرجاع الحجز للتدقيق ═══
  // التدقيق جان قرار نهائي ما إله رجعة: أول ما ينضغط «مطابق» تختفي
  // خانة التدقيق كاملة، فأي غلط بالمبلغ يبقى محبوس بالسجل وما ينصلّح.
  // مدير النظام حصراً يكدر يفتحه من جديد ويرجع يدققه.
  const unverify = async (row: DailyAuditRow) => {
    if (!window.confirm(`ترجّع الحجز ${row.code} للتدقيق من جديد؟`)) return
    setBusy(row.id)
    try {
      await api.unverifyBooking(row.id)
      load(date)
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر إرجاع الحجز للتدقيق')
    } finally {
      setBusy(null)
    }
  }

  const audit = async (row: DailyAuditRow, action: 'VERIFY' | 'MISMATCH' | 'PRICE_ERROR') => {
    const typed = amounts[row.id]
    // بلا مبلغ مكتوب: ننزل على المعتمد (فاتورة الليدر أو تقدير الإداري)
    const amount = typed !== undefined && typed !== ''
      ? Number(typed)
      : (row.collected > 0 ? undefined : row.expectedAmount || undefined)

    let note: string | undefined
    if (action !== 'VERIFY') {
      const answer = prompt(action === 'MISMATCH'
        ? 'شنو الفرق بالضبط؟ (يروح للرقابة والجودة)'
        : 'شنو الغلط بالسعر؟ (يروح للرقابة والإداري)')
      if (answer === null) return
      note = answer.trim() || undefined
    }
    setBusy(row.id)
    try {
      await api.auditBooking(row.id, { action, amountCollected: amount, note })
      setAmounts((prev) => ({ ...prev, [row.id]: '' }))
      load(date)
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر التدقيق')
    } finally {
      setBusy(null)
    }
  }

  return (
    <div dir="rtl" className="space-y-6">
      <div className="rounded-2xl p-6 shadow-sm" style={{ backgroundColor: '#1a3a5c' }}>
        <h1 className="text-2xl font-bold text-white">📅 التدقيق اليومي</h1>
        <p className="mt-1 text-sm text-blue-200">
          حدّد التاريخ وشوف كل حجوزات ذاك اليوم — شنو انجمع، شنو باقي، وشكد المفروض يجي اليوم.
        </p>
      </div>

      <div className="flex flex-wrap items-center gap-3 rounded-2xl bg-white p-4 shadow-sm">
        <label className="text-sm text-slate-600">التاريخ</label>
        <input type="date" value={date} onChange={(e) => e.target.value && setDate(e.target.value)}
          className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500" />
        {date !== todayStr() && (
          <button onClick={() => setDate(todayStr())}
            className="rounded-lg border border-brand-500 px-4 py-2 text-sm font-bold text-brand-700 hover:bg-brand-50">
            اليوم
          </button>
        )}
      </div>

      {!rep && <p className="py-10 text-center text-slate-400">جاري التحميل...</p>}

      {rep && (
        <>
          {/* المجاميع الأربعة */}
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <Tile label="١) المبالغ المستلمة" value={money(rep.collectedTotal)}
              hint={`من ${rep.completedCount} حجز مكتمل`} color="#15803d" />
            <Tile label="٢) ما تم تدقيقه" value={money(rep.notVerifiedTotal)}
              hint="لسه بانتظار قرارك" color="#b45309" />
            <Tile label="٣) كل المبالغ (مدقق + غير مدقق)" value={money(rep.allAmountsTotal)}
              hint={`المدقق منها: ${money(rep.verifiedTotal)}`} color="#1a3a5c" />
            <Tile label="٤) الإجمالي المتوقع لليوم" value={money(rep.expectedTotal)}
              hint="من فواتير الليدرز وتقديرات الإداري" color="#c8a45a" />
          </div>

          {/* عدّاد التقدم: المبلغ يزيد من الصفر لحد الإجمالي المتوقع
              كل ما المحاسب يدقق حجز — حتى يعرف وين واصل بدون ما يحسب */}
          {rep.expectedTotal > 0 && (
            <div className="rounded-2xl bg-white p-5 shadow-sm">
              <div className="mb-2 flex flex-wrap items-baseline justify-between gap-2">
                <span className="text-sm font-bold text-slate-700">تقدّم التدقيق اليوم</span>
                <span className="text-sm text-slate-500">
                  <b style={{ color: '#15803d' }}>{money(rep.verifiedTotal)}</b>
                  {' '}من أصل{' '}
                  <b style={{ color: '#c8a45a' }}>{money(rep.expectedTotal)}</b>
                </span>
              </div>
              <div className="h-5 overflow-hidden rounded-full bg-slate-100">
                <div
                  className="h-full rounded-full transition-all duration-500"
                  style={{
                    width: `${Math.min(100, Math.round((rep.verifiedTotal / rep.expectedTotal) * 100))}%`,
                    background: 'linear-gradient(90deg, #15803d, #22c55e)',
                  }}
                />
              </div>
              <div className="mt-2 flex flex-wrap justify-between gap-2 text-xs text-slate-500">
                <span>
                  {Math.min(100, Math.round((rep.verifiedTotal / rep.expectedTotal) * 100))}% مدقق
                </span>
                <span>
                  باقي: <b className="text-amber-700">{money(Math.max(0, rep.expectedTotal - rep.verifiedTotal))}</b>
                </span>
              </div>
            </div>
          )}

          <div className="flex flex-wrap gap-3 text-sm">
            <span className="rounded-full bg-emerald-100 px-3 py-1 font-bold text-emerald-800">
              مكتملة: {rep.completedCount}
            </span>
            <span className="rounded-full bg-slate-200 px-3 py-1 font-bold text-slate-700">
              غير مكتملة: {rep.pendingCount}
            </span>
            {rep.issuesCount > 0 && (
              <span className="rounded-full bg-red-100 px-3 py-1 font-bold text-red-700">
                محوّلة للرقابة: {rep.issuesCount}
              </span>
            )}
          </div>

          <div className="space-y-3">
            {rep.rows.map((row) => (
              <div key={row.id} className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <p className="font-bold text-slate-800">
                      {row.code}
                      <span className="mr-2 text-sm font-normal text-slate-500">{row.customerName || 'زبون غير معروف'}</span>
                    </p>
                    <p className="mt-1 text-xs text-slate-400" dir="ltr">{row.customerPhone}</p>
                    <p className="mt-1 text-sm text-slate-600">
                      {row.serviceName} · <span className="text-slate-400">{STATUS_LABEL[row.status] || row.status}</span>
                    </p>
                    <p className="mt-2 text-sm text-slate-600">
                      المستلم: <b>{money(row.collected)}</b>
                      {row.invoiceTotal != null
                        ? <> · فاتورة الليدر: <b>{money(row.invoiceTotal)}</b> <span className="text-xs text-slate-400">({row.invoiceCode})</span></>
                        : <> · <span className="text-amber-700">ماكو فاتورة ليدر — المعتمد تقدير الإداري: <b>{money(row.quotedPrice)}</b></span></>}
                    </p>
                  </div>
                  <div className="flex flex-col items-end gap-1">
                    {/* ⚠️ الشارة كانت **تكذب**: تكتب «بانتظار التدقيق» على
                        حجز لسه ما انجز — يعني ماكو فلوس انستلمت وماكو شي
                        ينتدقق أصلاً. المحاسب يشوف طابور شغل ما يكدر
                        يسويه، ويحس إنه متأخر بشي مو بيده.
                        وهاي بالضبط الي شكّه صاحب العمل: «ليش وحدة بيها
                        خيارات المطابقة والبقية ما بيهن؟» */}
                    {row.amountVerified ? (
                      <span className="rounded-full bg-emerald-100 px-3 py-1 text-xs font-bold text-emerald-800">✔ مدقق</span>
                    ) : row.status === 'COMPLETED' ? (
                      <span className="rounded-full bg-amber-100 px-3 py-1 text-xs font-bold text-amber-800">بانتظار التدقيق</span>
                    ) : (
                      <span className="rounded-full bg-slate-200 px-3 py-1 text-xs font-bold text-slate-600">
                        لسه ما انجز — ما ينتدقق
                      </span>
                    )}
                    {row.amountVerified && isAdmin && (
                      <button
                        type="button"
                        disabled={busy === row.id}
                        onClick={() => unverify(row)}
                        className="rounded-lg border border-amber-300 bg-amber-50 px-2.5 py-1 text-[11px] font-bold text-amber-800 hover:bg-amber-100 disabled:opacity-50"
                      >
                        ↩ إرجاع للتدقيق
                      </button>
                    )}
                    {row.openIssues > 0 && (
                      <span className="rounded-full bg-red-100 px-3 py-1 text-xs font-bold text-red-700">
                        محوّل للرقابة
                      </span>
                    )}
                  </div>
                </div>

                {/* الحجز الي ما انجز: نگول ليش ماكو أزرار بدل ما نتركه
                    فاضي والمستخدم يظن النظام مكسور. */}
                {!row.amountVerified && row.status !== 'COMPLETED' && (
                  <p className="mt-3 rounded-xl border border-slate-200 bg-slate-50 px-3 py-2 text-xs text-slate-500">
                    ⏳ التدقيق يصير بعد إنجاز الحجز — الحالة هسه: <b>{STATUS_LABEL[row.status] || row.status}</b>.
                    ماكو مبلغ مستلم حتى ينتدقق.
                  </p>
                )}

                {!row.amountVerified && row.status === 'COMPLETED' && (
                  <div className="mt-3 rounded-xl border border-slate-200 bg-slate-50 p-3">
                    <input
                      type="number" min="0" inputMode="numeric"
                      value={amounts[row.id] ?? ''}
                      onChange={(e) => setAmounts((prev) => ({ ...prev, [row.id]: e.target.value }))}
                      placeholder={`المبلغ حسب الفاتورة${row.expectedAmount ? ` (المعتمد: ${row.expectedAmount})` : ''}`}
                      className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                    />
                    <div className="mt-2 grid grid-cols-1 gap-2 sm:grid-cols-3">
                      <button disabled={busy === row.id} onClick={() => audit(row, 'VERIFY')}
                        className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-3 py-2 text-sm font-medium text-white disabled:opacity-50">
                        ✔ مطابق
                      </button>
                      <button disabled={busy === row.id} onClick={() => audit(row, 'MISMATCH')}
                        className="rounded-lg bg-amber-500 px-3 py-2 text-sm font-medium text-white disabled:opacity-50">
                        ⚠️ غير مطابق
                      </button>
                      <button disabled={busy === row.id} onClick={() => audit(row, 'PRICE_ERROR')}
                        className="rounded-lg bg-red-600 px-3 py-2 text-sm font-medium text-white disabled:opacity-50">
                        ✕ خطأ بالسعر
                      </button>
                    </div>
                  </div>
                )}
              </div>
            ))}
            {rep.rows.length === 0 && (
              <p className="rounded-2xl bg-white p-8 text-center text-slate-400 shadow-sm">
                ماكو حجوزات بهذا التاريخ
              </p>
            )}
          </div>
        </>
      )}
    </div>
  )
}
