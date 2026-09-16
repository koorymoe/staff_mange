// ═══ فترة الحجز — صباحي لو مسائي ═══
//
// (ع): «أكو شغلة اسمها تعديل الحجز، أريد خيار وياهن يطلع — نحدد
// صباحي لو مسائي».
//
// ⚠️ **ترجع null لما ماكو فترة، ما تكتب «—»**: أغلب الحجوزات
// القديمة بلا فترة، ولو كل صف كتب «—» يصير الجدول مليان شرطات
// ويدفن الي محدَّد فعلاً. الي إله فترة يبان، والباقي ما ياخذ محلاً.
//
// والشريحة **ما تخمّن من الساعة**: موعد ٤ العصر ممكن يكون آخر
// الشفت الصباحي أو أول المسائي، والتخمين هنا يوجّه كادراً غلط.

const SHIFTS: Record<string, { label: string; cls: string }> = {
  MORNING: { label: '🌅 صباحي', cls: 'bg-amber-50 text-amber-800 border-amber-200' },
  EVENING: { label: '🌙 مسائي', cls: 'bg-indigo-50 text-indigo-800 border-indigo-200' },
}

export default function ShiftBadge({
  shift,
  className = '',
}: {
  shift?: 'MORNING' | 'EVENING' | null
  className?: string
}) {
  const meta = shift ? SHIFTS[shift] : undefined
  if (!meta) return null
  return (
    <span className={`rounded-md border px-1.5 py-0.5 text-[11px] font-bold ${meta.cls} ${className}`}>
      {meta.label}
    </span>
  )
}
