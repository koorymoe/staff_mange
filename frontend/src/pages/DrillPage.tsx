import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { api, type DrillResult } from '../api'
import { useSession } from '../session'
import { canDrill } from '../utils/drill'

// ═══ «الرقم = القائمة» — قرار (ع) 10-16 ═══
// صفحة وحدة لكل بطاقة رقم باللوحة: العنوان نفس اسم البطاقة، والعدد
// نفس رقمها، والشرح يگول شنو هذا الرقم ومن يا فترة. الخادم يعدّ ويطلّع
// القائمة بنفس الشرط (drill_repository.go)، فالرقمان ما يفترقان.
export default function DrillPage() {
  const { key = '' } = useParams()
  const navigate = useNavigate()
  const { employee, permissions } = useSession()
  const allowed = canDrill(employee, permissions)
  const [d, setD] = useState<DrillResult | null>(null)
  const [err, setErr] = useState('')
  const [q, setQ] = useState('')

  useEffect(() => {
    if (!allowed) return
    let alive = true
    api.getDrill(key).then((x) => { if (alive) { setD(x); setErr('') } })
      .catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر جلب القائمة') })
    return () => { alive = false }
  }, [key, allowed])

  if (!allowed) return <p className="py-16 text-center text-slate-500">هاي القائمة مو ضمن صلاحياتك.</p>

  const rows = (d?.rows ?? []).filter((r) => !q.trim()
    || [r.code, r.customer, r.phone, r.service, r.crew, r.responsible, r.station].some((v) => v?.includes(q.trim())))
  const showMissing = key === 'needs_coord' || key === 'needs_crew'

  return (
    <div className="space-y-4" dir="rtl">
      <div className="rounded-2xl bg-gradient-to-l from-[#0f2040] to-[#2c5aad] p-5 text-white shadow">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <button onClick={() => navigate(-1)} className="mb-2 rounded-lg bg-white/15 px-3 py-1 text-xs font-bold hover:bg-white/25">→ رجوع</button>
            <h1 className="text-2xl font-black">{d?.title ?? '…'}</h1>
            {d && <p className="mt-1 max-w-3xl text-sm text-blue-100">{d.explain}</p>}
          </div>
          {d && (
            <div className="rounded-xl bg-white/15 px-5 py-3 text-center">
              <p className="text-3xl font-black tabular-nums">{d.count}</p>
              <p className="text-[11px] text-blue-100">حجز</p>
            </div>
          )}
        </div>
        {d?.extra && d.extra.total > 0 && (
          <p className="mt-3 rounded-xl bg-white/10 px-3 py-2 text-sm">
            📈 نجاح الأسبوع: خلص <b>{d.extra.completed}</b> من أصل <b>{d.extra.total}</b> حجز = <b>{Math.round((d.extra.completed / d.extra.total) * 100)}٪</b>
            {' '}— الباقي ({d.extra.total - d.extra.completed}) تشوف حالة كل واحد بعمود «وين واكف».
          </p>
        )}
      </div>

      {err && <p className="rounded-xl bg-red-50 px-4 py-3 text-sm text-red-700">{err}</p>}
      {!d && !err && <p className="py-16 text-center text-slate-400">جاري التحميل…</p>}

      {d && (
        <div className="rounded-2xl border border-slate-200 bg-white p-3">
          <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="بحث بالكود، الزبون، الهاتف، الكادر…"
            className="mb-3 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-brand-500" />
          {rows.length === 0 ? (
            <p className="py-10 text-center text-slate-500">{d.count === 0 ? 'ماكو ولا حجز — الرقم صفر فعلاً.' : 'ماكو نتيجة للبحث.'}</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-right text-[13px]">
                <thead className="bg-slate-50 text-[11px] text-slate-500">
                  <tr>
                    <th className="p-2">الحجز</th><th className="p-2">الزبون</th><th className="p-2">الخدمة</th>
                    <th className="p-2">الموعد</th><th className="p-2">وين واكف</th>
                    {showMissing && <th className="p-2">الناقص</th>}
                    <th className="p-2">الكادر</th><th className="p-2">المسؤول</th><th className="p-2">صارله</th>
                  </tr>
                </thead>
                <tbody>
                  {rows.map((r) => (
                    <tr key={r.id} className="border-t border-slate-100 hover:bg-sky-50/50">
                      <td className="p-2"><Link to={`/bookings?focus=${r.id}`} className="font-bold text-sky-700 underline">{r.code}</Link></td>
                      <td className="p-2">{r.customer ?? '—'}{r.phone && <span className="block text-[11px] text-slate-400" dir="ltr">{r.phone}</span>}</td>
                      <td className="p-2">{r.service ?? '—'}</td>
                      <td className="p-2 tabular-nums">{r.scheduledAt ?? <span className="text-amber-700">بلا موعد</span>}</td>
                      <td className="p-2"><span className="rounded-full bg-slate-100 px-2 py-0.5 text-[11px] font-bold">{r.station}</span></td>
                      {showMissing && <td className="p-2 text-amber-800">{r.missing ?? '—'}</td>}
                      <td className="p-2">{r.crew ?? '—'}</td>
                      <td className="p-2">{r.responsible ?? '—'}</td>
                      <td className="p-2 tabular-nums">{r.ageDays == null ? '—' : r.ageDays === 0 ? 'اليوم' : `${r.ageDays} يوم`}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {d.rows.length < d.count && <p className="mt-2 text-[11px] text-slate-400">معروض أقدم {d.rows.length} من {d.count}.</p>}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
