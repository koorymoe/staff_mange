import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type MatrixDecisions } from '../api'
import OwnerSwitch from '../components/OwnerSwitch'
import { SWITCH_MATRIX_AUTOPILOT } from '../systemSwitches'

// ═══ صندوق قرارات ماتركس — للمالك ومدير النظام ═══
// ثلاث أقسام: شنو سوّى لحاله اليوم (تذكيرات بس، ويه زر «لا تسوي هذا»)،
// وشنو ينتظر قرارك (أحكامه + حجوزات باچر بلا كادر)، والقرار ياخذه
// المدير بالشاشة الأصلية بنفس حراسها.

const todayBaghdad = () => new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' })

function entityLink(type: string): string | null {
  if (type === 'BOOKING') return '/coordinator'
  if (type === 'IT_ASSET') return '/it-assets'
  if (type === 'VEHICLE') return '/vehicles'
  return null
}

export default function MatrixDecisionsPage() {
  const [day, setDay] = useState(todayBaghdad())
  const [data, setData] = useState<MatrixDecisions | null>(null)
  const [err, setErr] = useState<string | null>(null)
  const [busy, setBusy] = useState<string | null>(null)

  const [reload, setReload] = useState(0)
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

  return (
    <div dir="rtl" className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-2xl font-bold text-brand-900">🤖 صندوق قرارات ماتركس</h2>
          <p className="text-sm text-slate-500">ماتركس يرسل التذكيرات البسيطة لحاله. الفلوس والعقوبات والتكليف تبقى إلك.</p>
        </div>
        <input type="date" value={day} onChange={(e) => setDay(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-1.5 text-sm" />
      </div>

      <OwnerSwitch switchKey={SWITCH_MATRIX_AUTOPILOT} label="ماتركس ينفّذ التذكيرات لحاله" hint="إذا انطفى، ماتركس يبقى يحلّل ويبلّغك بس، وما يرسل ولا تذكير للموظفين." />

      {err && <p className="rounded-lg bg-red-50 p-3 text-red-600">{err}</p>}
      {!data && !err && <p className="text-slate-400">جاري التحميل…</p>}

      {data && (
        <>
          <section className="rounded-2xl border border-amber-200 bg-white p-4">
            <h3 className="mb-3 font-extrabold text-[#0f2040]">⏳ ينتظر قرارك <span className="text-sm text-slate-500">({data.pending.length + data.unstaffed.length})</span></h3>
            {data.pending.length === 0 && data.unstaffed.length === 0 ? <p className="text-sm text-slate-400">ماكو شي ينتظرك ✅</p> : (
              <ul className="divide-y divide-slate-100 text-sm">
                {data.unstaffed.map((b) => (
                  <li key={`u${b.id}`} className="flex flex-wrap items-center justify-between gap-2 py-2">
                    <span>👷 حجز <b>{b.code}</b> باچر {new Date(b.scheduledAt).toLocaleTimeString('ar-IQ', { timeZone: 'Asia/Baghdad', hour: '2-digit', minute: '2-digit' })} وماكو عليه كادر</span>
                    <Link to="/coordinator" className="rounded-lg bg-brand-50 px-3 py-1 text-xs font-bold text-brand-700">كلّف من التنسيق ←</Link>
                  </li>
                ))}
                {data.pending.map((p) => (
                  <li key={p.id} className="flex flex-wrap items-center justify-between gap-2 py-2">
                    <div>
                      <p className="font-bold text-slate-800">🧠 {p.title}</p>
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
                        <p className="text-xs text-slate-500">{data.labels[a.kind] || a.kind} · إلى: {a.targetLabel || '—'} · {new Date(a.createdAt).toLocaleTimeString('ar-IQ', { timeZone: 'Asia/Baghdad', hour: '2-digit', minute: '2-digit' })}</p>
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
