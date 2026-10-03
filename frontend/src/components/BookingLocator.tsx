import { useEffect, useState } from 'react'
import { api } from '../api'
import BookingCodeChip from './BookingCodeChip'

// ═══ «وين هذا الحجز؟» ═══
//
// «يجي الموظف يبحث عن الحجز — وين يبحث؟ بالحجوزات المثبتة، والحجز
// بعده ما متثبت. أريد النظام يساعد الموظف: من يبحث بكود حجز بمكان
// غلط، يكله هذا الحجز بفلان مكان، ابحث عنه هناك».
//
// وهاي أذكى من تسريع البحث: المشكلة مو إن الموظف ما لگاه، هي إنه ما
// يعرف **وين يدوّر**. عشر محطات وهو يفتحهن وحدة وحدة، وأغلب الوقت
// ينتهي يظن الحجز انحذف ويسجّله من جديد — فيصير حجزين لنفس الزبون.
//
// ⚠️ ما نطلع إلا لمن يكون البحث **ما لگه شي بالمحطة الحالية**: مساعدة
// تطلع وانت لاگي طلبك تصير إزعاج.

export default function BookingLocator({ term, currentStation }: {
  term: string
  /** اسم المحطة الحالية — حتى ما نگول «موجود هنا» وهو مو هنا */
  currentStation: string
}) {
  const [found, setFound] = useState<{ code: string; customerName: string | null; station: string }[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    const q = term.trim()
    let alive = true
    const t = setTimeout(() => {
      // أقل من حرفين يجيب نص الشركة — ننتظر شوية
      if (q.length < 2) { setFound([]); return }
      setLoading(true)
      api.locateBookings(q)
        .then((r) => { if (alive) setFound(r) })
        .catch(() => { if (alive) setFound([]) })
        .finally(() => { if (alive) setLoading(false) })
    }, 350)
    return () => { alive = false; clearTimeout(t) }
  }, [term])

  if (term.trim().length < 2) return null
  if (loading) return <p className="mt-3 text-xs text-slate-400">جاري البحث بباقي المحطات...</p>

  if (found.length === 0) {
    return (
      <div className="mt-3 rounded-xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm">
        <p className="font-bold text-slate-700">ماكو حجز بهذا البحث — لا هنا ولا بأي محطة ثانية.</p>
        <p className="mt-0.5 text-xs text-slate-500">
          تأكد من الكود أو رقم الهاتف. وإذا الحجز قديم، يمكن يكون بالأرشيف.
        </p>
      </div>
    )
  }

  return (
    <div className="mt-3 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3">
      <p className="text-sm font-extrabold text-amber-900">
        🔎 لگيناه — بس مو بهاي المحطة
      </p>
      <p className="mt-0.5 text-xs text-amber-800">
        انت تدوّر بـ«{currentStation}»، وهذا وين هو فعلاً:
      </p>
      <div className="mt-2 space-y-1.5">
        {found.map((f) => (
          <div key={f.code} className="flex flex-wrap items-center gap-2 rounded-lg bg-white px-3 py-2">
            <span className="font-mono text-xs font-black text-brand-700"><BookingCodeChip code={f.code} /></span>
            {f.customerName && <span className="text-xs font-bold text-slate-700">{f.customerName}</span>}
            <span className="mr-auto rounded-full bg-amber-100 px-2.5 py-0.5 text-[11px] font-black text-amber-900">
              📍 {f.station}
            </span>
          </div>
        ))}
      </div>
      <p className="mt-2 text-[11px] text-amber-700">
        افتح المحطة المذكورة من الخيارات فوق — الحجز بيها.
      </p>
    </div>
  )
}
