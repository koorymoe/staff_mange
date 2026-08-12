import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, type DailyAuditReport, type DailyAuditRow } from '../api'
import { useSession } from '../session'

/**
 * التدقيق اليومي.
 *
 * الفرق عن «تدقيق الحسابات» العام: هذا يمشي بيوم واحد. المحاسب يحدد
 * تاريخ ويشوف حجوزات ذاك اليوم **المنجزة** بمبالغها وحالة تدقيقها،
 * ويعرف من الصبح شكد المفروض يجمع.
 *
 * التدقيق نفسه (المبلغ وخيارات الخطأ) يصير من نفس الشاشة بالضبط،
 * و«مطابق» ترحّل فاتورة الحجز لطابور الاعتماد بشاشة فواتير الليدر.
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

/**
 * بطاقة رقم.
 *
 * اللون يجي شريط فوگ + خلفية مغسولة بنفس اللون بدل ما يكون بالرقم بس —
 * حتى المحاسب يميّز الأربعة بلمحة عين وهو يمرّ عليهن، بلا ما يقرا
 * العناوين وحدة وحدة.
 */
function Tile({ label, value, hint, color, tint, icon }: {
  label: string; value: string; hint?: string; color: string; tint: string; icon: string
}) {
  return (
    <div
      className="group relative overflow-hidden rounded-2xl bg-white shadow-sm ring-1 ring-slate-100 transition hover:-translate-y-0.5 hover:shadow-md"
      style={{ background: `linear-gradient(180deg, ${tint} 0%, #ffffff 62%)` }}
    >
      <span className="absolute inset-x-0 top-0 h-1" style={{ backgroundColor: color }} />
      <div className="p-5">
        <p className="flex items-center gap-1.5 text-sm text-slate-500">
          <span aria-hidden>{icon}</span>
          {label}
        </p>
        <p className="mt-1.5 text-2xl font-black tracking-tight" style={{ color }}>{value}</p>
        {hint && <p className="mt-1 text-xs text-slate-400">{hint}</p>}
      </div>
    </div>
  )
}

export default function DailyAuditPage() {
  const navigate = useNavigate()
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
      {/* الرأس: تدرّج بدل اللون المسطّح + هالة خفيفة، ونفس المعالجة
          بشاشة بلاغات الأخطاء حتى شاشات المحاسب تبين عائلة وحدة. */}
      <div
        className="relative overflow-hidden rounded-2xl p-6 shadow-md"
        style={{ background: 'linear-gradient(135deg, #1a3a5c 0%, #24507e 55%, #2f6ba8 100%)' }}
      >
        <span
          aria-hidden
          className="pointer-events-none absolute -left-16 -top-24 h-64 w-64 rounded-full opacity-20"
          style={{ background: 'radial-gradient(circle, #c8a45a 0%, transparent 70%)' }}
        />
        <div className="relative flex flex-wrap items-end justify-between gap-4">
          <div>
            <h1 className="text-2xl font-black text-white">📅 التدقيق اليومي</h1>
            <p className="mt-1 max-w-xl text-sm text-blue-100">
              حجوزات اليوم <b className="text-white">المنجزة</b> بمبالغها — شنو انجمع، شنو باقي،
              وشكد المفروض يجي.
            </p>
          </div>
          {rep && (
            <div className="rounded-xl bg-white/10 px-4 py-2 text-center ring-1 ring-white/20 backdrop-blur">
              <p className="text-2xl font-black leading-none text-white">{rep.rows.length}</p>
              <p className="mt-1 text-[11px] text-blue-100">حجز بالقائمة</p>
            </div>
          )}
        </div>
      </div>

      {/* شريط التاريخ لزج: المحاسب ينزل بقائمة طويلة ويبقى يكدر يبدّل
          اليوم بلا ما يرجع لفوگ. */}
      <div className="sticky top-2 z-10 flex flex-wrap items-center gap-3 rounded-2xl border border-slate-100 bg-white/90 p-4 shadow-sm backdrop-blur">
        <label className="text-sm font-bold text-slate-600">📆 التاريخ</label>
        <input type="date" value={date} onChange={(e) => e.target.value && setDate(e.target.value)}
          className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none transition focus:border-brand-500 focus:ring-2 focus:ring-brand-100" />
        {date !== todayStr() && (
          <button onClick={() => setDate(todayStr())}
            className="rounded-lg border border-brand-500 px-4 py-2 text-sm font-bold text-brand-700 transition hover:bg-brand-50">
            ارجع لليوم
          </button>
        )}
        {/* «وين ألگه الفواتير الي بانتظار الاعتماد؟» — الفاتورة الي
            تنتأشر «مطابق» هنا تروح لطابور الاعتماد بشاشة فواتير الليدر،
            فالطريق لازم يكون من نفس المكان مو بالقائمة الجانبية. */}
        <a
          href="#/leader-invoices"
          onClick={(e) => { e.preventDefault(); navigate('/leader-invoices') }}
          className="mr-auto rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-2 text-sm font-bold text-white shadow-sm transition hover:opacity-90"
        >
          🧾 الفواتير بانتظار الاعتماد ←
        </a>
      </div>

      {!rep && <p className="py-10 text-center text-slate-400">جاري التحميل...</p>}

      {rep && (
        <>
          {/* المجاميع الأربعة */}
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <Tile label="١) المبالغ المستلمة" value={money(rep.collectedTotal)}
              hint={`من ${rep.completedCount} حجز مكتمل`} color="#15803d" tint="#ecfdf5" icon="💰" />
            <Tile label="٢) ما تم تدقيقه" value={money(rep.notVerifiedTotal)}
              hint="لسه بانتظار قرارك" color="#b45309" tint="#fffbeb" icon="⏳" />
            <Tile label="٣) كل المبالغ (مدقق + غير مدقق)" value={money(rep.allAmountsTotal)}
              hint={`المدقق منها: ${money(rep.verifiedTotal)}`} color="#1a3a5c" tint="#eff6ff" icon="🧮" />
            <Tile label="٤) الإجمالي المتوقع لليوم" value={money(rep.expectedTotal)}
              hint="من فواتير الليدرز وتقديرات الإداري" color="#a67c2e" tint="#fefce8" icon="🎯" />
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
            {/* ما عاد نعرض «غير مكتملة» — الحجز الي ما انجز ما يوصل هنا
                أصلاً. الرقم كان يخلي المحاسب يدور على صفوف مو موجودة. */}
            {rep.issuesCount > 0 && (
              <span className="rounded-full bg-red-100 px-3 py-1 font-bold text-red-700">
                محوّلة للرقابة: {rep.issuesCount}
              </span>
            )}
          </div>

          <div className="space-y-3">
            {rep.rows.map((row) => (
              <div
                key={row.id}
                /* شريط جانبي بلون الحالة: أحمر إذا محوّل للرقابة، أخضر
                   إذا انتدقق، وكهرماني إذا لسه بالطابور. المحاسب يعرف
                   وين شغله بلمحة بدل ما يقرا كل شارة. */
                className={`relative overflow-hidden rounded-xl border bg-white p-4 pr-5 shadow-sm transition hover:shadow-md ${
                  row.openIssues > 0 ? 'border-red-200' : row.amountVerified ? 'border-emerald-200' : 'border-amber-200'
                }`}
              >
                <span
                  aria-hidden
                  className={`absolute inset-y-0 right-0 w-1.5 ${
                    row.openIssues > 0 ? 'bg-red-400' : row.amountVerified ? 'bg-emerald-400' : 'bg-amber-400'
                  }`}
                />
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <p className="font-bold text-slate-800">
                      <span className="rounded-lg bg-slate-100 px-2 py-0.5 font-mono text-sm text-slate-700">{row.code}</span>
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
                    {/* القائمة صارت **منجزة بس** (فلتر بالسيرفر)، فكل صف
                        هنا إله مبلغ مستلم وينتدقق فعلاً — ما عاد يصير صف
                        بلا أزرار حكم. */}
                    {row.amountVerified ? (
                      <span className="rounded-full bg-emerald-100 px-3 py-1 text-xs font-bold text-emerald-800">✔ مدقق</span>
                    ) : (
                      <span className="rounded-full bg-amber-100 px-3 py-1 text-xs font-bold text-amber-800">بانتظار التدقيق</span>
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

                {!row.amountVerified && (
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
              /* الفراغ لازم يفسّر نفسه: القائمة منجزة بس، فـ«ماكو حجوزات»
                 لحاله يخلي المحاسب يظن إنه النظام ما جاب البيانات. */
              <div className="rounded-2xl border border-dashed border-slate-200 bg-white p-10 text-center shadow-sm">
                <p className="text-4xl">🗓️</p>
                <p className="mt-3 font-bold text-slate-600">ماكو حجز منجز بهذا اليوم</p>
                <p className="mx-auto mt-1 max-w-sm text-sm text-slate-400">
                  القائمة تعرض الحجوزات <b>المنجزة</b> بس — هيّه الي عدها مبلغ مستلم ينتدقق.
                  الحجوزات الي لسه بالتنفيذ تطلع هنا يوم ما تنجز.
                </p>
              </div>
            )}
          </div>
        </>
      )}
    </div>
  )
}
