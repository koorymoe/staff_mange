import { useState } from 'react'
import { api, type Booking, type BookingProgressReport } from '../api'
import BookingCodeChip from './BookingCodeChip'

// ═══ الإنجاز الجزئي — «خلصنا جزء والباقي باچر» ═══
//
// قبل هاي الشاشة، الليدر آخر اليوم عنده خيارين وكلاهما غلط: يأشّر «تم
// الإنجاز» على شغل ناقص (فتطلع فاتورة على شي ما انخلص)، أو «توقف
// العمل» (فيبين وكأن الشغل فشل).
//
// والمعلومة الأهم كانت تضيع بالحالتين: **وين وصلوا؟** فالكادر الي
// يطلع باچر يبدي من الصفر ويسأل الزبون «شنو سووا أمس؟».
//
// لهذا «شنو انخلص» و«شنو باقي» إلزاميين — بدونهم التقرير ما ينفع أحد،
// وهو كل سبب وجود هاي الميزة.
export default function PartialCompleteDialog({
  booking,
  onDone,
  onClose,
}: {
  booking: Booking
  onDone: (report: BookingProgressReport) => void
  onClose: () => void
}) {
  const [workDone, setWorkDone] = useState('')
  const [remaining, setRemaining] = useState('')
  const [percent, setPercent] = useState(50)
  const [blockers, setBlockers] = useState('')
  const [materials, setMaterials] = useState('')
  const [amount, setAmount] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const submit = async () => {
    if (workDone.trim().length < 3) return setError('اكتب شنو انخلص اليوم')
    if (remaining.trim().length < 3) return setError('اكتب شنو باقي — الكادر الجاي يعتمد عليها')
    setBusy(true); setError(null)
    try {
      const report = await api.partialCompleteBooking(booking.id, {
        workDone: workDone.trim(),
        remainingWork: remaining.trim(),
        percentDone: percent,
        blockers: blockers.trim() || undefined,
        materialsUsed: materials.trim() || undefined,
        amountCollected: amount ? Number(amount) : undefined,
      })
      onDone(report)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'تعذر تسجيل الإنجاز الجزئي')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-2xl bg-white p-5 shadow-xl">
        <h3 className="text-lg font-bold text-[#0f2040]">
          🔄 إنجاز جزئي — حجز <BookingCodeChip code={booking.code} />
        </h3>
        <p className="mt-1 text-xs text-slate-500">
          الحجز راح يرجع لإداري الحجوزات حتى ينسّق يوم جديد، وتقريرك هذا يوصل للكادر الي يكمّل.
        </p>

        <label className="mt-4 block text-sm font-bold text-slate-700">
          شنو انخلص اليوم؟ <span className="text-red-500">*</span>
        </label>
        <textarea
          value={workDone}
          onChange={(e) => setWorkDone(e.target.value)}
          rows={3}
          placeholder="مثال: انسلكت ٦ كاميرات بالطابق الأول، وانركّب الراك والـDVR"
          className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
        />

        <label className="mt-3 block text-sm font-bold text-slate-700">
          شنو باقي؟ <span className="text-red-500">*</span>
        </label>
        <textarea
          value={remaining}
          onChange={(e) => setRemaining(e.target.value)}
          rows={3}
          placeholder="مثال: باقي ٤ كاميرات بالطابق الثاني + البرمجة والتسليم للزبون"
          className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
        />

        <label className="mt-3 block text-sm font-bold text-slate-700">
          شكد النسبة المنجزة؟ <span className="text-brand-700">{percent}٪</span>
        </label>
        <input
          type="range"
          min={1}
          max={99}
          value={percent}
          onChange={(e) => setPercent(Number(e.target.value))}
          className="mt-1 w-full"
        />
        <p className="text-[11px] text-slate-400">
          ١٠٠٪ يعني خلص — بهذي الحالة استعمل «تم الإنجاز» مو الإنجاز الجزئي.
        </p>

        <label className="mt-3 block text-sm font-bold text-slate-700">أكو شي عرقل الشغل؟</label>
        <input
          value={blockers}
          onChange={(e) => setBlockers(e.target.value)}
          placeholder="مثال: الكهرباء مقطوعة، أو الزبون ما فتح غرفة السيرفر"
          className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
        />

        <label className="mt-3 block text-sm font-bold text-slate-700">المواد الي انستهلكت</label>
        <input
          value={materials}
          onChange={(e) => setMaterials(e.target.value)}
          placeholder="مثال: ٨٠ متر كيبل، ٦ قواعد"
          className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
        />

        <label className="mt-3 block text-sm font-bold text-slate-700">مبلغ مستلم اليوم (إن وجد)</label>
        <input
          type="number"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
        />

        {error && <p className="mt-3 rounded-lg bg-red-50 p-2 text-sm font-bold text-red-700">{error}</p>}

        <div className="mt-4 flex gap-2">
          <button
            onClick={submit}
            disabled={busy}
            className="flex-1 rounded-lg bg-gradient-to-l from-amber-500 to-amber-700 px-4 py-2.5 text-sm font-bold text-white disabled:opacity-50"
          >
            {busy ? 'جاري التسجيل...' : '🔄 سجّل الإنجاز الجزئي'}
          </button>
          <button onClick={onClose} className="rounded-lg border border-slate-300 px-4 py-2.5 text-sm font-bold text-slate-600">
            إلغاء
          </button>
        </div>
      </div>
    </div>
  )
}
