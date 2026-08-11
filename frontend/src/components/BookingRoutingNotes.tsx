import { useState } from 'react'
import { api, type Booking } from '../api'

// ═══ ملاحظات الترحيل ═══
//
// «الإداري من يرحّل حجز للتيم يريد يكتب ملاحظة مع الحجز حتى يقراها
// الفريق، ونفس الشي للمشاريع يكتب ملاحظة لمدير المشاريع».
//
// قبل، الملاحظة تنقال بالتلفون: «الدرج ضيق جيبوا سلّم قصير»، «الزبون
// ما يرد قبل العصر». والي ما كان بالمكالمة ما يعرفها، وبعد يومين ما
// يتذكرها أحد. هسه تنكتب مع الحجز ويقراها الي يفتحه.
//
// ⚠️ ملاحظتين منفصلات مو وحدة: الكادر يقرا شي (وين، شنو يجيب، متى
// الزبون موجود) ومدير المشاريع يقرا شي ثاني (عرض سعر، عقد، شروط
// الزبون). دمجهن يخلي كل واحد يقرا كلام مو إله ويتجاهل الكل.
//
// ⚠️ الفراغ يمسح الملاحظة — مقصود: الإداري لازم يكدر يشيلها لو انتغيّر
// الوضع، وإلا تبقى ملاحظة قديمة تضلّل الفريق.

type Props = {
  booking: Booking
  /** أي ملاحظة تنعرض: للكادر، للمشاريع، أو الاثنين */
  show?: 'crew' | 'project' | 'both'
  onSaved: (updated: Booking) => void
}

export default function BookingRoutingNotes({ booking, show = 'both', onSaved }: Props) {
  const [crew, setCrew] = useState(booking.crewNotes || '')
  const [project, setProject] = useState(booking.projectNotes || '')
  const [busy, setBusy] = useState<'crew' | 'project' | null>(null)
  const [err, setErr] = useState<string | null>(null)

  const save = async (which: 'crew' | 'project') => {
    setBusy(which)
    setErr(null)
    try {
      const updated =
        which === 'crew'
          ? await api.setBookingCrewNotes(booking.id, crew)
          : await api.setBookingProjectNotes(booking.id, project)
      onSaved(updated)
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر الحفظ')
    } finally {
      setBusy(null)
    }
  }

  const Box = ({
    which, label, hint, value, setValue, savedBy, savedAt, accent,
  }: {
    which: 'crew' | 'project'
    label: string
    hint: string
    value: string
    setValue: (v: string) => void
    savedBy?: string
    savedAt?: string
    accent: string
  }) => (
    <div className={`rounded-xl border p-3 ${accent}`}>
      <label className="mb-1 block text-xs font-bold text-slate-700">{label}</label>
      <textarea
        value={value}
        onChange={(e) => setValue(e.target.value)}
        rows={2}
        placeholder={hint}
        className="w-full resize-y rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
      />
      <div className="mt-1.5 flex items-center justify-between gap-2">
        <span className="text-[11px] text-slate-500">
          {savedBy ? <>كتبها {savedBy}{savedAt && <> · {new Date(savedAt).toLocaleString('en-GB')}</>}</> : 'ما انكتبت بعد'}
        </span>
        <button
          onClick={() => save(which)}
          disabled={busy !== null}
          className="rounded-lg bg-[#0f2040] px-3 py-1.5 text-xs font-bold text-white disabled:opacity-50"
        >
          {busy === which ? 'جاري الحفظ...' : 'احفظ'}
        </button>
      </div>
    </div>
  )

  return (
    <div className="mt-3 grid grid-cols-1 gap-3 sm:grid-cols-2" dir="rtl">
      {(show === 'crew' || show === 'both') && (
        <Box
          which="crew"
          label="📋 ملاحظة للكادر المنفّذ"
          hint="مثال: الدرج ضيق — جيبوا سلّم قصير"
          value={crew}
          setValue={setCrew}
          savedBy={booking.crewNotesByName}
          savedAt={booking.crewNotesAt}
          accent="border-sky-200 bg-sky-50/60"
        />
      )}
      {(show === 'project' || show === 'both') && (
        <Box
          which="project"
          label="🏢 ملاحظة لمدير المشاريع"
          hint="مثال: الزبون يريد عرض سعر قبل التنفيذ"
          value={project}
          setValue={setProject}
          savedBy={booking.projectNotesByName}
          savedAt={booking.projectNotesAt}
          accent="border-violet-200 bg-violet-50/60"
        />
      )}
      {err && <p className="col-span-full rounded-lg bg-red-50 px-3 py-2 text-xs font-bold text-red-700">{err}</p>}
    </div>
  )
}
