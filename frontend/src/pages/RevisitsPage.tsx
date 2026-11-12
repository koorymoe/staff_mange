import { useEffect, useState } from 'react'
import { api, type RevisitReport } from '../api'
import BookingCodeChip from '../components/BookingCodeChip'
import PctWhy from '../ui/PctWhy'

// ═══ 🔁 الرجعات — (ع) 11-07 ═══
// حجز ثاني لنفس الزبون ونفس الخدمة خلال ٣٠ يوم من الإنجاز = الشغل الأول ما انعمل
// صح. أهم مقياس لجودة الفني. ⚠️ للمتابعة والتدريب بس — ماكو أي خصم يطلع منه.
const RANGES = [30, 90, 180]

export default function RevisitsPage() {
  const [days, setDays] = useState(90)
  const [rep, setRep] = useState<RevisitReport | null>(null)
  const [err, setErr] = useState('')
  useEffect(() => {
    let alive = true
    api.getRevisits(days)
      .then((r) => { if (alive) { setRep(r); setErr('') } })
      .catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر') })
    return () => { alive = false }
  }, [days])
  const rate = rep && rep.jobs ? Math.round((1000 * rep.revisits) / rep.jobs) / 10 : 0

  return (
    <div dir="rtl" className="space-y-4">
      <div>
        <h2 className="text-2xl font-bold text-brand-900">🔁 الرجعات</h2>
        <p className="text-sm text-slate-500">
          حجز رجع لنفس الزبون ونفس الخدمة خلال {rep?.windowDays ?? 30} يوم من الإنجاز — يعني الشغل الأول يمكن ما انعمل صح.
          تنحسب على ليدر الحجز الأول. للمتابعة والتدريب، مو للخصم.
        </p>
      </div>
      <div className="flex gap-2">
        {RANGES.map((d) => (
          <button key={d} type="button" onClick={() => { if (d !== days) { setRep(null); setDays(d) } }}
            className={`rounded-full px-3 py-1 text-xs font-bold ${days === d ? 'bg-brand-700 text-white' : 'bg-slate-100 text-slate-600'}`}>آخر {d} يوم</button>
        ))}
      </div>
      {err && <p className="text-sm text-red-600">{err}</p>}
      {!rep && !err && <p className="text-slate-400">جاري الحساب…</p>}
      {rep && (
        <>
          <div className="grid grid-cols-3 gap-2">
            <div className="rounded-2xl bg-white p-4 shadow-sm"><p className="text-xs text-slate-500">حجوزات انجزت</p><p className="text-2xl font-extrabold">{rep.jobs}</p></div>
            <div className="rounded-2xl bg-white p-4 shadow-sm"><p className="text-xs text-slate-500">رجعت</p><p className="text-2xl font-extrabold text-amber-700">{rep.revisits}</p></div>
            <div className="rounded-2xl bg-white p-4 shadow-sm"><p className="text-xs text-slate-500">نسبة الرجعة</p><p className="text-2xl font-extrabold"><PctWhy title="نسبة الرجعة" value={rate} suffix="٪" parts={{ numLabel: 'حجز رجع', num: rep.revisits, denLabel: 'حجز انجز', den: rep.jobs }}
              meaning={<>من كل 100 حجز انجز، شكد واحد رجع الزبون فتح حجز ثاني لنفس الخدمة خلال {rep.windowDays} يوم — يعني الشغل الأول يمكن ما انعمل صح. كل ما تكون أقل أحسن.</>}
              source={<>جدول الحجوزات: كل حجز حالته «مكتمل» انجز خلال المدة المختارة (بدون الاستيراد القديم OLD- والشغل الداخلي والمحذوف). وبعدها ندور حجز ثاني لنفس الزبون ونفس الخدمة انفتح خلال {rep.windowDays} يوم من الإنجاز ومو ملغي.</>} /></p></div>
          </div>

          <div className="overflow-x-auto rounded-2xl bg-white shadow-sm">
            <table className="w-full text-sm">
              <thead><tr className="border-b border-slate-100 text-xs text-slate-500"><th className="p-3 text-right">الليدر</th><th className="p-3">انجز</th><th className="p-3">رجع</th><th className="p-3">النسبة</th></tr></thead>
              <tbody>
                {rep.leaders.length === 0 && <tr><td colSpan={4} className="p-6 text-center text-slate-400">ماكو حجوزات منجزة بهالمدة</td></tr>}
                {rep.leaders.map((l) => (
                  <tr key={l.leaderId} className="border-b border-slate-50">
                    <td className="p-3 font-bold">{l.leaderName}</td>
                    <td className="p-3 text-center tabular-nums">{l.jobs}</td>
                    <td className="p-3 text-center tabular-nums">{l.revisits}</td>
                    <td className="p-3 text-center"><span className={`rounded-full px-2 py-0.5 text-xs font-bold ${l.rate >= 15 ? 'bg-red-100 text-red-700' : l.rate >= 7 ? 'bg-amber-100 text-amber-800' : 'bg-emerald-100 text-emerald-700'}`}><PctWhy title={`نسبة رجعة ${l.leaderName}`} value={l.rate} suffix="٪" parts={{ numLabel: 'حجز رجع', num: l.revisits, denLabel: 'حجز انجز وهو ليدره', den: l.jobs }}
                      meaning={<>من الحجوزات الي هذا الليدر كان ليدرها وانجزت، شكد منها رجع الزبون عليها خلال {rep.windowDays} يوم. فوق 15٪ أحمر، من 7٪ لـ15٪ أصفر، أقل من 7٪ أخضر.</>}
                      source={<>جدول الحجوزات: كل حجز حالته «مكتمل» انجز خلال المدة المختارة (بدون الاستيراد القديم OLD- والشغل الداخلي والمحذوف). وبعدها ندور حجز ثاني لنفس الزبون ونفس الخدمة انفتح خلال {rep.windowDays} يوم من الإنجاز ومو ملغي.</>} /></span></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {rep.items.length > 0 && (
            <div className="space-y-2">
              <p className="text-sm font-extrabold text-brand-900">الحجوزات الي رجعت</p>
              {rep.items.map((it) => (
                <div key={it.firstId} className="flex flex-wrap items-center gap-2 rounded-xl bg-white p-3 text-sm shadow-sm">
                  <BookingCodeChip code={it.firstCode} />
                  <span className="text-slate-400">← رجع بعد {it.gapDays} يوم →</span>
                  <BookingCodeChip code={it.secondCode} />
                  <span className="text-slate-600">{it.serviceName ?? '—'}</span>
                  <span className="text-slate-500">· {it.customerName ?? ''}</span>
                  <span className="ms-auto text-xs text-slate-500">الليدر: {it.leaderName ?? '—'}</span>
                </div>
              ))}
            </div>
          )}
        </>
      )}
    </div>
  )
}
