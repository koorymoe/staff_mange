import { useEffect, useState } from 'react'
import { api, type Booking, type BookingCrewInventoryState } from '../api'
import { acceptedBookings } from '../utils/acceptedBookings'
import BookingCodeChip from '../components/BookingCodeChip'

// ═══ جرد أدوات فريقي ═══
//
// «ما أختار فريقي — يطلعلي فريقي الي وياي بهذا الحجز، وجاردين عددهم
// لو لا. هم من ينطون «تم» على أدواتهم تطلع عندي، ويطلع عندي إذا
// عدهم نقص».
//
// ⚠️ الشاشة القديمة كانت **الليدر يجرد عن فريقه**: يختار موظفين من
// قائمة، ويأشّر بنفسه شنو موجود عند كل واحد. وهاي غلط بثلاث نواحي:
//
//   • الليدر ما يعرف شنو بحقيبة الفني — يخمّن، والتخمين ينكتب حقيقة.
//   • المسؤولية تنقلب: الفني ينقصه أداة والسجل يگول الليدر أشّرها.
//   • واختيار الموظفين بالإيد يعني ممكن يختار واحد مو بالحجز أصلاً.
//
// هسه كل واحد يجرد **عدته هو**، والليدر يشوف النتيجة جاهزة. الليدر
// ما يجرد عن أحد — يشوف ويتابع، وأدواته هو يجردها بـ«جرد أدواتي»
// مثل أي أحد.

export default function TeamInventoryCheckPage({ embedded }: { embedded?: boolean } = {}) {
  const [bookings, setBookings] = useState<Booking[]>([])
  const [selected, setSelected] = useState('')
  const [crew, setCrew] = useState<BookingCrewInventoryState[]>([])
  const [loading, setLoading] = useState(true)
  const [err, setErr] = useState<string | null>(null)

  useEffect(() => {
    api.getBookings({ assignedTo: 'me' })
      .then((bs) => {
        const open = acceptedBookings(bs)
        setBookings(open)
        if (open.length > 0) setSelected(open[0].id)
      })
      .catch((e) => setErr(e instanceof Error ? e.message : 'تعذر جلب حجوزاتك'))
      .finally(() => setLoading(false))
  }, [])

  // ⚠️ التصفير جوّا الـpromise مو بجسم الـeffect: تعديل الحالة
  // مباشرة بالجسم يولّد رندر متسلسل.
  useEffect(() => {
    if (!selected) return
    api.getBookingCrewInventory(selected).then(setCrew).catch(() => setCrew([]))
  }, [selected])

  const booking = bookings.find((b) => b.id === selected) || null
  const doneCount = crew.filter((c) => c.checkedAt).length
  const shortCount = crew.filter((c) => c.checkedAt && !c.complete).length

  if (loading) return <p className="p-6 text-center text-slate-400">جارٍ التحميل...</p>

  return (
    <div dir="rtl" className="mx-auto max-w-3xl space-y-4">
      {/* مدمجة جوّا «الجرد»؟ ما نعيد الرأس — الشاشة أصلاً بيها رأس
          وعنوان التبويب، وتكراره يدفع المحتوى للأسفل بلا فايدة. */}
      {!embedded && (
        <div className="flex items-center gap-3">
          <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-brand-50 text-lg sm:h-11 sm:w-11 sm:text-xl">👥</span>
          <div className="min-w-0">
            <h1 className="text-xl font-black text-[#0f2040] sm:text-2xl">جرد أدوات فريقي</h1>
            <p className="text-[11px] text-slate-500 sm:text-xs">
              منو من فريقك جرد عدته قبل ما تطلعون — وشنو الناقص عنده
            </p>
          </div>
        </div>
      )}

      {err && <p className="rounded-xl bg-red-50 px-4 py-3 text-sm font-bold text-red-700">{err}</p>}

      {!booking ? (
        <div className="rounded-2xl border border-slate-200 bg-white p-10 text-center">
          <p className="text-3xl">🗓️</p>
          <p className="mt-2 text-sm font-bold text-slate-600">ما استلمت أي حجز بعد</p>
          <p className="mt-1 text-xs text-slate-400">
            روح لـ«مهامي» واضغط «استلام» على الحجز الي طالع له — ويطلع هنا فريقك وحالة جردهم تلقائياً.
          </p>
        </div>
      ) : (
        <>
          {/* ── الحجز — يطلع لحاله ── */}
          <div className="rounded-2xl border-2 border-brand-200 bg-brand-50/50 p-4">
            <p className="text-[11px] font-bold text-brand-700">🔧 الحجز الي استلمته</p>
            <p className="mt-1 font-black text-[#0f2040]">
              {booking.code && <span className="font-mono"><BookingCodeChip code={booking.code} /> · </span>}
              {booking.customer?.name || 'بدون اسم'}
            </p>
            <div className="mt-1 flex flex-wrap gap-x-3 gap-y-1 text-[11px] text-slate-600">
              {booking.scheduledAt && (
                <span>🕐 {new Date(booking.scheduledAt).toLocaleString('ar-IQ', {
                  day: 'numeric', month: 'long', hour: '2-digit', minute: '2-digit',
                })}</span>
              )}
              {(booking.address || booking.customer?.location) && (
                <span>📍 {booking.address || booking.customer?.location}</span>
              )}
            </div>

            {/* أكثر من حجز؟ أزرار — تشوفهن بلمحة وتبدّل بضغطة */}
            {bookings.length > 1 && (
              <div className="mt-2.5 flex flex-wrap items-center gap-1.5 border-t border-brand-200/60 pt-2.5">
                <span className="text-[10px] text-slate-500">عندك {bookings.length} حجوزات:</span>
                {bookings.map((b) => (
                  <button
                    key={b.id}
                    onClick={() => setSelected(b.id)}
                    className={`rounded-lg px-2 py-1 text-[10px] font-bold transition ${
                      b.id === selected ? 'bg-brand-700 text-white' : 'border border-slate-300 bg-white text-slate-600 hover:bg-slate-50'
                    }`}
                  >
                    {b.code || b.customer?.name || 'حجز'}
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* ── الأرقام ── */}
          <div className="grid grid-cols-3 gap-2.5">
            <Stat label="الفريق" value={crew.length} tone="slate" />
            <Stat label="جردوا" value={doneCount} tone="emerald" />
            <Stat label="عدهم نقص" value={shortCount} tone="red" />
          </div>

          {/* ── الفريق ── */}
          <div className="space-y-2">
            {crew.length === 0 && (
              <p className="rounded-2xl border border-dashed border-slate-200 bg-white px-4 py-8 text-center text-sm text-slate-400">
                ما انكلّف كادر لهذا الحجز بعد.
              </p>
            )}
            {crew.map((c) => (
              <div
                key={c.employeeId}
                className={`rounded-2xl border-2 bg-white p-4 ${
                  !c.checkedAt ? 'border-amber-300'
                    : c.complete ? 'border-emerald-300'
                    : 'border-red-300'
                }`}
              >
                <div className="flex flex-wrap items-center gap-2">
                  <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-slate-100 text-sm font-black text-slate-500">
                    {c.name.charAt(0)}
                  </span>
                  <div className="min-w-0">
                    <p className="text-sm font-bold text-slate-800">{c.name}</p>
                    <p className="text-[10px] text-slate-500">
                      {c.isLeader ? 'ليدر' : c.position || 'فني'}
                    </p>
                  </div>
                  <span className={`mr-auto rounded-full px-3 py-1 text-[11px] font-bold ${
                    !c.checkedAt ? 'bg-amber-100 text-amber-800'
                      : c.complete ? 'bg-emerald-100 text-emerald-700'
                      : 'bg-red-100 text-red-700'
                  }`}>
                    {!c.checkedAt ? '⏳ ما جرد بعد' : c.complete ? '✅ عدته كاملة' : '⚠️ عنده نقص'}
                  </span>
                </div>

                {/* الناقص بالاسم مو «عنده نقص» بس: بلا التفصيل يضطر
                    الليدر يتصل يسأل — ونفس المكالمة الي بنينا الشاشة
                    حتى نلغيها. */}
                {c.missingItems && (
                  <p className="mt-2 rounded-lg bg-red-50 px-3 py-2 text-[11px] text-red-800">
                    الناقص: {c.missingItems}
                  </p>
                )}
                {c.checkedAt && (
                  <p className="mt-1 text-[10px] text-slate-400">
                    جرد {new Date(c.checkedAt).toLocaleString('ar-IQ', {
                      day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit',
                    })}
                  </p>
                )}
              </div>
            ))}
          </div>

          <p className="rounded-xl bg-slate-50 px-3 py-2 text-[11px] leading-relaxed text-slate-500">
            ⓘ كل واحد يجرد عدته هو من شاشة «جرد أدواتي» — وأنت تشوف النتيجة هنا.
            أدواتك انت جردها من نفس الشاشة مثل الباقين.
          </p>
        </>
      )}
    </div>
  )
}

function Stat({ label, value, tone }: { label: string; value: number; tone: 'slate' | 'emerald' | 'red' }) {
  const tones: Record<string, string> = {
    slate: 'text-slate-700',
    emerald: 'text-emerald-700',
    red: 'text-red-700',
  }
  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-3 text-center shadow-sm">
      <p className={`text-2xl font-black ${tones[tone]}`}>{value}</p>
      <p className="text-[10px] text-slate-500">{label}</p>
    </div>
  )
}
