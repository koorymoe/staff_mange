// ═══ «هذا الحجز مو هنا — موجود بالمكان الفلاني» ═══
//
// «من نريد نبحث بأي خانة عن رقم الحجز، إذا جان الموظف يبحث بالمكان
// الخطأ يطلعله من النظام إن هذا الحجز بالمكان الفلاني».
//
// ⚠️⚠️ **الفرق بين «ماكو نتيجة» و«مو هنا»**: الموظف الي يدوّر على حجز
// بشاشة الحجوزات وما يلگاه **يستنتج إنه مو موجود** — فيسأل زميله، أو
// يسوي حجزاً ثانياً بنفس الشغل. والنظام يعرف إنه بالأرشيف وساكت.
//
// ⚠️ ويظهر **بس لمن يفشل البحث المحلي**: شريط يطلع مع كل بحث يصير
// ضجيجاً، والموظف يتعلّم يتجاهله — فما ينفع لمن يحتاجه فعلاً.

import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'

import { api, type LocateResult } from '../api'
import BookingCodeChip from './BookingCodeChip'

interface Props {
  /** نص البحث الي كتبه الموظف. */
  query: string
  /** عدد النتائج المحلية — الشريط يظهر بس إذا صفر. */
  localCount: number
  /** مسار الشاشة الحالية — ما ننصح بمكان إحنا واگفين بيه. */
  currentRoute: string
}

export default function LocateHint({ query, localCount, currentRoute }: Props) {
  const [hit, setHit] = useState<LocateResult | null>(null)
  const q = query.trim()

  useEffect(() => {
    // ⚠️ ما نسأل الخادم إلا لمن يفشل البحث المحلي **ويكون النص يشبه
    // كوداً**: نداء بكل حرف يكتبه الموظف يعني عشرات النداءات لبحث
    // عادي عن اسم زبون.
    if (localCount > 0 || q.length < 4) return
    let alive = true
    const t = setTimeout(() => {
      api.locate(q).then((r) => { if (alive) setHit(r.found ? r : null) }).catch(() => { if (alive) setHit(null) })
    }, 400)
    return () => { alive = false; clearTimeout(t) }
  }, [q, localCount])

  // ⚠️ الشرط بالرسم مو بالتأثير: تصفير الحالة جوّا `useEffect` يسبّب
  // دورة رسم زيادة بكل حرف يُكتب. والنتيجة الي ما تخصّ البحث الحالي
  // تنتجاهل هنا بلا ما تنمسح.
  if (localCount > 0 || q.length < 4) return null
  if (!hit || !hit.found || hit.code?.toLowerCase() !== q.toLowerCase()) return null
  const sameScreen = hit.route === currentRoute

  return (
    <div className="mb-3 rounded-2xl border-2 border-sky-200 bg-sky-50 px-4 py-3">
      <p className="text-sm font-bold text-sky-900">
        🔎 الحجز <b className="font-mono"><BookingCodeChip code={hit.code} /></b> موجود — بس مو بهذي الشاشة.
      </p>
      {hit.hint && <p className="mt-1 text-[12.5px] leading-relaxed text-sky-800">{hit.hint}</p>}
      <div className="mt-2 flex flex-wrap gap-2">
        {!sameScreen && hit.route && (
          <Link to={hit.route}
            className="rounded-lg bg-sky-600 px-3 py-1.5 text-xs font-bold text-white hover:bg-sky-700">
            روح لـ«{hit.where}» ←
          </Link>
        )}
        {/* ⚠️ ونقول **بأي طابور** فاتورته: أكثر بحث عن حجز سببه
            فاتورته، و«موجود بالحجوزات» ما يكفي جواباً. */}
        {hit.invoiceStage && hit.invoiceRoute !== currentRoute && (
          <Link to={hit.invoiceRoute ?? '/leader-invoices'}
            className="rounded-lg border border-sky-300 bg-white px-3 py-1.5 text-xs font-bold text-sky-700 hover:bg-sky-100">
            🧾 {hit.invoiceStage} ←
          </Link>
        )}
      </div>
    </div>
  )
}
