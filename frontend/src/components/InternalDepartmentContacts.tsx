import { useEffect, useState } from 'react'
import { api, type Department } from '../api'

/* ═══ أرقام مسؤولي القسم بالشغل داخل الشركة ═══
 *
 * (ع): «هذا المفروض يكون بي أرقام الموظفين المسؤولين ع القسم الي
 * احنه راح نشتغل اله».
 *
 * والحجز يصوّر **مسؤولاً واحداً** (الي طلب الحجز) — والقسم ممكن عنده
 * أكثر من مسؤول بالسجل. فالكادر الي رايح يشتغل، أو الي يسوي الفاتورة،
 * يحتاج يتصل — ولو الطالب ما رد يحتاج الثاني.
 *
 * 🔴 «رقم غلط أسوأ من ماكو رقم»: قسم بلا مسؤولين بالسجل يعرض **«—»**،
 * وما نخمّن رقماً ولا نعرض خانة فارغة تشبه رقماً ناقصاً.
 */

// نداء واحد للسجل بكل الجلسة — الصفحة تنفتح بكل حجز داخلي، وسجل
// الأقسام ما يتغيّر بالدقيقة.
let cache: Promise<Department[]> | null = null
const load = () => {
  if (!cache) cache = api.getDepartments().catch(() => [] as Department[])
  return cache
}
export function forgetDepartments(): void { cache = null }

type Props = {
  departmentId?: string | null
  /** اسم القسم المصوَّر بالحجز — يُعرض لما القسم ما يلگى بالسجل */
  departmentName?: string | null
  /** طالب الحجز المصوَّر بالحجز — يبقى ظاهراً دائماً وموسوماً */
  requesterName?: string | null
  requesterPhone?: string | null
}

export default function InternalDepartmentContacts({
  departmentId, departmentName, requesterName, requesterPhone,
}: Props) {
  const [heads, setHeads] = useState<Department['heads']>(undefined)
  const [name, setName] = useState<string | null>(departmentName ?? null)

  useEffect(() => {
    if (!departmentId) return
    let alive = true
    load().then((list) => {
      if (!alive) return
      const d = list.find((x) => x.id === departmentId)
      if (!d) return
      setHeads(d.heads)
      setName(d.name)
    })
    return () => { alive = false }
  }, [departmentId])

  // الطالب أول، وبعده بقية مسؤولي القسم بلا تكرار رقمه.
  const others = (heads ?? []).filter(
    (h) => !!h.phone && h.phone !== (requesterPhone ?? ''),
  )

  return (
    <div className="rounded-xl border p-3" style={{ borderColor: 'var(--bd-line)' }}>
      <p className="mb-2 text-xs font-bold" style={{ color: 'var(--t-muted)' }}>
        📞 مسؤولو القسم {name ? `— ${name}` : ''}
      </p>
      <div className="space-y-1.5">
        {requesterName && (
          <Row label={requesterName} phone={requesterPhone} tag="صاحب الطلب" />
        )}
        {others.map((h) => (
          <Row key={h.id} label={h.name} phone={h.phone} />
        ))}
        {!requesterName && others.length === 0 && (
          <p className="text-sm" style={{ color: 'var(--t-muted)' }}>—</p>
        )}
      </div>
    </div>
  )
}

function Row({ label, phone, tag }: { label: string; phone?: string | null; tag?: string }) {
  return (
    <div className="flex items-center justify-between gap-2 text-sm">
      <span className="flex items-center gap-1.5">
        <span style={{ color: 'var(--t-body)' }}>{label}</span>
        {tag && (
          <span className="rounded-full bg-brand-50 px-2 py-0.5 text-[10px] font-bold text-brand-700">
            {tag}
          </span>
        )}
      </span>
      {phone ? (
        <a href={`tel:${phone}`} className="font-bold text-brand-700 hover:underline" dir="ltr">
          {phone}
        </a>
      ) : (
        <span style={{ color: 'var(--t-muted)' }}>—</span>
      )}
    </div>
  )
}
