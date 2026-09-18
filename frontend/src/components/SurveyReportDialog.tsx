import { useState } from 'react'
import { api, type Booking, type BookingSurveyReport } from '../api'
import BookingCodeChip from './BookingCodeChip'

// ═══ نتائج زيارة معاينة («كشف») ═══
//
// حجز الكشف ما يحتاج فاتورة ولا تقرير عمل — بس المعلومة الي طلع
// الكادر يجمعها (شنو يريد الزبون، شنو المساحة) لازم توصل، وإلا
// المعاينة تصير رحلة ضايعة والزبون يتصل يسأل «شنو صار؟».
export default function SurveyReportDialog({
  booking,
  onDone,
  onClose,
}: {
  booking: Booking
  onDone: (report: BookingSurveyReport) => void
  onClose: () => void
}) {
  const [customerWants, setCustomerWants] = useState('')
  const [siteArea, setSiteArea] = useState('')
  const [siteDetails, setSiteDetails] = useState('')
  const [otherNotes, setOtherNotes] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const submit = async () => {
    if (customerWants.trim().length < 3) return setError('اكتب شنو يريد الزبون')
    setBusy(true); setError(null)
    try {
      const report = await api.submitSurveyReport(booking.id, {
        customerWants: customerWants.trim(),
        siteAreaSqm: siteArea ? Number(siteArea) : undefined,
        siteDetails: siteDetails.trim() || undefined,
        otherNotes: otherNotes.trim() || undefined,
      })
      onDone(report)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'تعذر تسجيل نتائج المعاينة')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-2xl bg-white p-5 shadow-xl">
        <h3 className="text-lg font-bold text-[#0f2040]">
          🔍 نتائج المعاينة — حجز <BookingCodeChip code={booking.code} />
        </h3>
        <p className="mt-1 text-xs text-slate-500">
          هذا حجز كشف — بلا فاتورة وبلا تقرير عمل، بس نتائج المعاينة لازم توصل.
        </p>

        <label className="mt-4 block text-sm font-bold text-slate-700">
          شنو يريد الزبون؟ <span className="text-red-500">*</span>
        </label>
        <textarea
          value={customerWants}
          onChange={(e) => setCustomerWants(e.target.value)}
          rows={3}
          placeholder="مثال: يريد كاميرات خارجية للباب الرئيسي وسور الحديقة"
          className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
        />

        <label className="mt-3 block text-sm font-bold text-slate-700">مساحة الموقع (م²)</label>
        <input
          type="number"
          value={siteArea}
          onChange={(e) => setSiteArea(e.target.value)}
          className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
        />

        <label className="mt-3 block text-sm font-bold text-slate-700">تفاصيل الموقع</label>
        <input
          value={siteDetails}
          onChange={(e) => setSiteDetails(e.target.value)}
          placeholder="مثال: دورين، بيت زاوية، الكهرباء عمود قريب"
          className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
        />

        <label className="mt-3 block text-sm font-bold text-slate-700">ملاحظات أخرى</label>
        <textarea
          value={otherNotes}
          onChange={(e) => setOtherNotes(e.target.value)}
          rows={2}
          className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
        />

        {error && <p className="mt-3 rounded-lg bg-red-50 p-2 text-sm font-bold text-red-700">{error}</p>}

        <div className="mt-4 flex gap-2">
          <button
            onClick={submit}
            disabled={busy}
            className="flex-1 rounded-lg bg-gradient-to-l from-teal-500 to-teal-700 px-4 py-2.5 text-sm font-bold text-white disabled:opacity-50"
          >
            {busy ? 'جاري التسجيل...' : '🔍 سجّل نتائج المعاينة'}
          </button>
          <button onClick={onClose} className="rounded-lg border border-slate-300 px-4 py-2.5 text-sm font-bold text-slate-600">
            إلغاء
          </button>
        </div>
      </div>
    </div>
  )
}
