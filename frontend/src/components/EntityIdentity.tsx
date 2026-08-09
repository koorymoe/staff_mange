import type { Booking } from '../api'
import { identityOf, formatWhen, type IdentityFields } from '../utils/identity'

// ═══ رأس الهوية الموحّد ═══
//
// «أي مكان بي فاتورة بي حجز بي تقرير أريد تفاصيل كاملة — الكود مال
// الزبون ومال الحجز، اسم الليدر، وكلشي يحتاجه».
//
// الورقة الي تطلع من النظام لازم تعرّف نفسها. قبل، التقرير يقول
// «تم الإنجاز» وبس — والي يقراه لازم يدور بأي حجز وبأي زبون.
//
// ⚠️ الحقل الفارغ **ينختفي** ما يطلع «-». السطر المليان شرطات يوسّخ
// الورقة ويخلي الي مهم ما ينشاف.

type Props = {
  booking?: Partial<Booking> | null
  /** لو البيانات مو جاية من حجز (فاتورة بزبون حر مثلاً) */
  fields?: IdentityFields
  /** compact للبطاقات، full للتفاصيل والطباعة */
  variant?: 'compact' | 'full'
  className?: string
}

export default function EntityIdentity({ booking, fields, variant = 'compact', className = '' }: Props) {
  const id: IdentityFields = { ...identityOf(booking), ...(fields || {}) }
  const when = formatWhen(id.scheduledAt)

  // كود الحجز وكود الزبون هما المفتاحين — بدونهم الرأس ما يعرّف شي
  const hasAny = id.bookingCode || id.customerCode || id.customerName
  if (!hasAny) return null

  const Chip = ({ children, strong = false }: { children: React.ReactNode; strong?: boolean }) => (
    <span className={strong ? 'font-extrabold text-[#0f2040]' : 'text-slate-600'}>{children}</span>
  )

  return (
    <div
      className={`flex flex-wrap items-center gap-x-2.5 gap-y-1 rounded-lg border border-slate-200 bg-slate-50/70 px-3 py-2 text-xs ${className}`}
      dir="rtl"
    >
      {id.bookingCode && <Chip strong>🔖 {id.bookingCode}</Chip>}
      {id.customerCode && <Chip strong>{id.customerCode}</Chip>}
      {id.customerName && <Chip strong>{id.customerName}</Chip>}
      {id.customerPhone && (
        <a href={`tel:${id.customerPhone}`} className="font-bold text-brand-700 underline">
          📞 {id.customerPhone}
        </a>
      )}
      {id.leaderName ? <Chip>👷 الليدر: {id.leaderName}</Chip> : <span className="text-amber-700">👷 بلا ليدر</span>}

      {variant === 'full' && (
        <>
          {id.serviceName && <Chip>🛠️ {id.serviceName}</Chip>}
          {id.address && <Chip>📍 {id.address}</Chip>}
          {when && <Chip>📅 {when}</Chip>}
        </>
      )}
    </div>
  )
}
