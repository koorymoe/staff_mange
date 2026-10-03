import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type MatrixDecisions } from '../api'
import OwnerSwitch from '../components/OwnerSwitch'
import MatrixGuideRules from '../components/MatrixGuideRules'
import MatrixProposals from '../components/MatrixProposals'
import { SWITCH_MATRIX_AUTOPILOT, SWITCH_MATRIX_STAFF_EYE } from '../systemSwitches'

// ═══ صندوق قرارات ماتركس — للمالك ومدير النظام ═══
// ثلاث أقسام: شنو سوّى لحاله اليوم (تذكيرات بس، ويه زر «لا تسوي هذا»)،
// وشنو ينتظر قرارك (أحكامه + حجوزات باچر بلا كادر)، والقرار ياخذه
// المدير بالشاشة الأصلية بنفس حراسها.

const todayBaghdad = () => new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' })

function entityLink(type: string): string | null {
  if (type === 'BOOKING') return '/coordinator'
  if (type === 'IT_ASSET') return '/it-assets'
  if (type === 'VEHICLE' || type === 'VEHICLE_DOCUMENT') return '/vehicles'
  if (type === 'EXTRA_TASK') return '/extra-tasks'
  return null
}

// «ليش؟» — الأرقام الي بنى عليها ماتركس فعله، بشكل مقروء.
function WhyBox({ details }: { details: Record<string, unknown> }) {
  const show = (v: unknown): string => {
    if (Array.isArray(v)) return v.map((x) => (typeof x === 'object' && x ? Object.values(x as object).join(' / ') : String(x))).join('، ')
    if (typeof v === 'object' && v) return JSON.stringify(v)
    if (typeof v === 'string' && /^\d{4}-\d{2}-\d{2}T/.test(v)) return new Date(v).toLocaleString('ar-IQ', { timeZone: 'Asia/Baghdad' })
    return String(v)
  }
  return (
    <dl className="mt-1 grid gap-x-3 gap-y-0.5 rounded-lg bg-slate-50 p-2 text-xs text-slate-600 sm:grid-cols-[auto_1fr]">
      {Object.entries(details).map(([k, v]) => (
        <div key={k} className="contents"><dt className="font-bold">{WHY_LABELS[k] || k}</dt><dd className="break-words">{show(v)}</dd></div>
      ))}
    </dl>
  )
}

const WHY_LABELS: Record<string, string> = {
  bookingCodes: 'الحجوزات', items: 'التفاصيل', bookingCode: 'الحجز', expectedMinutes: 'المتوقع (دقيقة)',
  availableMinutes: 'المتاح (دقيقة)', samples: 'عدد العيّنات', basis: 'الأساس', limitedBy: 'المحدِّد',
  customerCode: 'رقم الزبون', latestBooking: 'آخر حجز', factors: 'العلامات', scheduledAt: 'الموعد',
  suggestedLeader: 'المقترح', repairCount: 'عدد التصليحات', repairCost: 'كلفة التصليح', reason: 'السبب',
  cost12m: 'كلفة ١٢ شهر', incidents180d: 'حوادث ١٨٠ يوم', reasons: 'الأسباب', customer: 'الزبون',
  subscriptionEnd: 'نهاية الاشتراك', daysLeft: 'باقي (يوم)', document: 'الوثيقة', vehicle: 'السيارة',
  plate: 'اللوحة', expiryDate: 'تنتهي', title: 'المهمة', dueAt: 'الموعد', daysLate: 'متأخرة (يوم)',
  tool: 'الأداة', available: 'المتوفر', total: 'الكلي', submittedAt: 'رُفعت', days: 'الأيام',
}

const pct = (a: number, b: number) => (b > 0 ? `${Math.round((a / b) * 100)}٪` : '—')

export default function MatrixDecisionsPage() {
  const [day, setDay] = useState(todayBaghdad())
  const [data, setData] = useState<MatrixDecisions | null>(null)
  const [err, setErr] = useState<string | null>(null)
  const [busy, setBusy] = useState<string | null>(null)
  const [open, setOpen] = useState<string | null>(null)

  const [reload, setReload] = useState(0)
  // فلتر «ينتظر قرارك» بموظف واحد — كل الأحكام عليه سوا.
  const [who, setWho] = useState('')
  const load = () => setReload((n) => n + 1)

  // جلب بمكان واحد مع alive: تبديل التاريخ بسرعة ما يخلّي جواب قديم يطمس الجديد.
  useEffect(() => {
    let alive = true
    api.getMatrixDecisions(day)
      .then((d) => { if (alive) { setData(d); setErr(null) } })
      .catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر جلب الصندوق') })
    return () => { alive = false }
  }, [day, reload])

  const undo = async (id: string) => {
    if (!confirm('ماتركس ما راح يعيد هذا الفعل على نفس الشي لمدة ٣٠ يوم. متأكد؟')) return
    setBusy(id)
    try {
      await api.undoMatrixAction(id)
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر الرفض')
    } finally {
      setBusy(null)
    }
  }

  const resume = async (kind: string) => {
    setBusy(kind)
    try {
      await api.resumeMatrixKind(kind)
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر الترجيع')
    } finally {
      setBusy(null)
    }
  }

  const waiting = data ? data.pending.length + data.unstaffed.length + data.escalated.length + data.paused.length : 0

  return (
    <div dir="rtl" className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-2xl font-bold text-brand-900">🤖 صندوق قرارات ماتركس</h2>
          <p className="text-sm text-slate-500">ماتركس يرسل التذكيرات البسيطة لحاله. الفلوس والعقوبات والتكليف تبقى إلك.</p>
        </div>
        <input type="date" value={day} onChange={(e) => setDay(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-1.5 text-sm" />
      </div>

      <OwnerSwitch switchKey={SWITCH_MATRIX_STAFF_EYE} label="عين ماتركس عند الموظفين" hint="إذا انطفت، العين وتوجيهاتها تختفي من شاشات الموظفين. إنت والمدير تبقى عندكم، وماتركس يكمّل يحلّل." />
      <OwnerSwitch switchKey={SWITCH_MATRIX_AUTOPILOT} label="ماتركس ينفّذ التذكيرات لحاله" hint="إذا انطفى، ماتركس يبقى يحلّل ويبلّغك بس، وما يرسل ولا تذكير للموظفين." />

      {err && <p className="rounded-lg bg-red-50 p-3 text-red-600">{err}</p>}
      {!data && !err && <p className="text-slate-400">جاري التحميل…</p>}

      <MatrixProposals onChange={load} />

      {data && (
        <>
          <section className="rounded-2xl border border-amber-200 bg-white p-4">
            <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
              <h3 className="font-extrabold text-[#0f2040]">⏳ ينتظر قرارك <span className="text-sm text-slate-500">({waiting})</span></h3>
              {data.pending.some((p) => p.employeeId) && (
                <select value={who} onChange={(e) => setWho(e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1 text-xs">
                  <option value="">كل الموظفين</option>
                  {[...new Map(data.pending.filter((p) => p.employeeId && p.employeeName).map((p) => [p.employeeId as string, p.employeeName as string])).entries()]
                    .map(([id, name]) => <option key={id} value={id}>{name} ({data.pending.filter((p) => p.employeeId === id).length})</option>)}
                </select>
              )}
            </div>
            {waiting === 0 ? <p className="text-sm text-slate-400">ماكو شي ينتظرك ✅</p> : (
              <ul className="divide-y divide-slate-100 text-sm">
                {data.paused.map((p) => (
                  <li key={`p${p.kind}`} className="flex flex-wrap items-center justify-between gap-2 py-2">
                    <span>⏸️ وقّفت «<b>{data.labels[p.kind] || p.kind}</b>» لأنك {p.reason}. أرجّعه؟</span>
                    <button disabled={busy === p.kind} onClick={() => resume(p.kind)} className="rounded-lg bg-emerald-50 px-3 py-1 text-xs font-bold text-emerald-700 disabled:opacity-50">▶️ رجّعه</button>
                  </li>
                ))}
                {data.escalated.map((a) => {
                  const link = entityLink(a.entityType)
                  return (
                    <li key={`e${a.id}`} className="flex flex-wrap items-center justify-between gap-2 py-2">
                      <div>
                        <p className="font-bold text-slate-800">⬆️ {a.summary}</p>
                        <p className="text-xs text-slate-500">ذكّرت {a.targetLabel || '—'} يوم {new Date(a.createdAt).toLocaleDateString('ar-IQ', { timeZone: 'Asia/Baghdad' })} وبعده ما انحل</p>
                      </div>
                      {link && <Link to={link} className="rounded-lg bg-brand-50 px-3 py-1 text-xs font-bold text-brand-700">افتح ←</Link>}
                    </li>
                  )
                })}
                {data.unstaffed.map((b) => (
                  <li key={`u${b.id}`} className="flex flex-wrap items-center justify-between gap-2 py-2">
                    <span>👷 حجز <b>{b.code}</b> باچر {new Date(b.scheduledAt).toLocaleTimeString('ar-IQ', { timeZone: 'Asia/Baghdad', hour: '2-digit', minute: '2-digit' })} وماكو عليه كادر</span>
                    <Link to="/coordinator" className="rounded-lg bg-brand-50 px-3 py-1 text-xs font-bold text-brand-700">كلّف من التنسيق ←</Link>
                  </li>
                ))}
                {data.pending.filter((p) => !who || p.employeeId === who).map((p) => (
                  <li key={p.id} className="flex flex-wrap items-center justify-between gap-2 py-2">
                    <div>
                      <p className="font-bold text-slate-800">🧠 {p.title}{p.employeeName && <> — <span className="text-brand-700">{p.employeeName}</span></>}</p>
                      <div className="my-1 flex flex-wrap gap-1.5 text-[11px]">
                        {p.employeeId && p.employeeName && <Link to={`/matrix/employee/${p.employeeId}`} className="rounded-full bg-sky-50 px-2 py-0.5 font-bold text-sky-800 hover:bg-sky-100">👤 {p.employeeName} — تقريره ←</Link>}
                        {p.bookingCode && <Link to="/bookings" className="rounded-full bg-amber-50 px-2 py-0.5 font-bold text-amber-800 hover:bg-amber-100">📋 الحجز {p.bookingCode}</Link>}
                        {p.occurredAt && <span className="rounded-full bg-slate-100 px-2 py-0.5 text-slate-600">📅 {new Date(p.occurredAt).toLocaleString('ar-IQ', { timeZone: 'Asia/Baghdad', dateStyle: 'medium', timeStyle: 'short' })}</span>}
                      </div>
                      {p.summary && <p className="text-xs text-slate-500">{p.summary}</p>}
                    </div>
                    <Link to="/monitor" className="rounded-lg bg-brand-50 px-3 py-1 text-xs font-bold text-brand-700">راجع بالصندوق ←</Link>
                  </li>
                ))}
              </ul>
            )}
          </section>

          <section className="rounded-2xl border border-slate-200 bg-white p-4">
            <h3 className="mb-3 font-extrabold text-[#0f2040]">✅ سوّيته لحالي <span className="text-sm text-slate-500">({data.done.length})</span></h3>
            {!data.autopilotEnabled && <p className="mb-2 rounded-lg bg-slate-100 p-2 text-xs text-slate-600">التنفيذ التلقائي مطفي هسه.</p>}
            {data.done.length === 0 ? <p className="text-sm text-slate-400">ماكو أفعال بهذا اليوم.</p> : (
              <ul className="divide-y divide-slate-100 text-sm">
                {data.done.map((a) => {
                  const link = entityLink(a.entityType)
                  return (
                    <li key={a.id} className={`flex flex-wrap items-center justify-between gap-2 py-2 ${a.status === 'UNDONE' ? 'opacity-50' : ''}`}>
                      <div>
                        <p className="text-slate-800">{a.summary}{link && <> · <Link to={link} className="text-brand-700 underline">فتح</Link></>}</p>
                        <p className="text-xs text-slate-500">
                          {data.labels[a.kind] || a.kind} · إلى: {a.targetLabel || '—'} · {new Date(a.createdAt).toLocaleTimeString('ar-IQ', { timeZone: 'Asia/Baghdad', hour: '2-digit', minute: '2-digit' })}
                          {a.resolvedAt && <span className="ms-2 rounded-full bg-emerald-50 px-2 text-emerald-700">✅ انحل</span>}
                          {!a.resolvedAt && a.escalatedAt && <span className="ms-2 rounded-full bg-amber-50 px-2 text-amber-700">⬆️ صعد</span>}
                          {a.details && <button onClick={() => setOpen(open === a.id ? null : a.id)} className="ms-2 text-brand-700 underline">ليش؟</button>}
                        </p>
                        {open === a.id && a.details && <WhyBox details={a.details} />}
                      </div>
                      {a.status === 'DONE'
                        ? <button disabled={busy === a.id} onClick={() => undo(a.id)} className="rounded-lg border border-red-200 px-3 py-1 text-xs font-bold text-red-600 hover:bg-red-50 disabled:opacity-50">🚫 لا تسوي هذا</button>
                        : <span className="text-xs text-red-600">مرفوض</span>}
                    </li>
                  )
                })}
              </ul>
            )}
          </section>

          <section className="rounded-2xl border border-slate-200 bg-white p-4 text-sm">
            <h3 className="mb-3 font-extrabold text-[#0f2040]">🎯 دقة ماتركس <span className="text-xs font-normal text-slate-500">(آخر ٣٠ يوم)</span></h3>
            <div className="grid grid-cols-2 gap-2 md:grid-cols-4">
              {[
                ['كل الأفعال', String(data.accuracy.actions)],
                ['انحلت بعد التذكير', `${data.accuracy.resolved} (${pct(data.accuracy.resolved, data.accuracy.actions)})`],
                ['صعدت إلك', String(data.accuracy.escalated)],
                ['رفضتها', `${data.accuracy.rejected} (${pct(data.accuracy.rejected, data.accuracy.actions)})`],
              ].map(([l, v]) => (
                <div key={l} className="rounded-xl bg-slate-50 p-3"><p className="text-xs text-slate-500">{l}</p><p className="text-lg font-extrabold tabular-nums">{v}</p></div>
              ))}
            </div>
            <p className="mt-2 text-xs text-slate-600">
              ⏱️ توقّعات التأخير: {data.accuracy.delayPredicted} توقّع
              {data.accuracy.delayChecked > 0
                ? <> · انفحص منها {data.accuracy.delayChecked} بعد ما خلصت، وصدق {data.accuracy.delayCorrect} ({pct(data.accuracy.delayCorrect, data.accuracy.delayChecked)})</>
                : <> · بعد ماكو حجوزات خلصت حتى نقيس بيها — الرقم يطلع أول ما ينجزون.</>}
            </p>
          </section>

          <MatrixGuideRules />

          <section className="rounded-2xl border border-slate-200 bg-white p-4 text-sm">
            <h3 className="mb-2 font-extrabold text-[#0f2040]">💡 اقتراحات</h3>
            <div className="flex flex-wrap gap-2">
              <Link to="/sales-opportunities" className="rounded-lg bg-slate-100 px-3 py-1.5">💰 فرص البيع</Link>
              <Link to="/it-stats" className="rounded-lg bg-slate-100 px-3 py-1.5">🔁 الاستبدال</Link>
              <Link to="/weekly-report" className="rounded-lg bg-slate-100 px-3 py-1.5">📊 التقرير الأسبوعي</Link>
            </div>
          </section>
        </>
      )}
    </div>
  )
}
