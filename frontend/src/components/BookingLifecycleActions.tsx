import { useState } from 'react'
import { api, type Booking } from '../api'

// ═══ قرارات الحجز الثلاثة ═══
//
// ثلاث حالات تصير كل يوم وما جان إلها مكان بالنظام:
//
//   الزبون ما رد    → الحجز ينزاح من طابور الشغل ويضل محفوظ
//   الزبون أجّل     → موعد جديد بسبب مكتوب، والتأجيل ينعد
//   الزبون ألغى     → الحجز يختفي من الشاشات ويضل بالأرشيف
//
// كلها بمكوّن واحد حتى تكون نفس الأزرار ونفس السلوك بتنسيق الحجوزات
// وبشاشة الحجوزات — مو نسختين تفترقن أول ما نعدّل وحدة.
export default function BookingLifecycleActions({
  booking,
  onChanged,
  canArchive = false,
}: {
  booking: Booking
  onChanged: (b: Booking) => void
  /** الأرشفة (الحذف) للإداري بس */
  canArchive?: boolean
}) {
  const [open, setOpen] = useState<null | 'postpone' | 'waiting' | 'archive'>(null)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)

  // datetime-local يريد "YYYY-MM-DDTHH:mm" بتوقيت محلي
  const toLocalInput = (iso: string | null) => {
    if (!iso) return ''
    const d = new Date(iso)
    const pad = (n: number) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
  }
  const [when, setWhen] = useState(toLocalInput(booking.scheduledAt))
  const [reason, setReason] = useState('')

  const run = async (fn: () => Promise<Booking>) => {
    setBusy(true)
    setErr(null)
    try {
      onChanged(await fn())
      setOpen(null)
      setReason('')
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر تنفيذ العملية')
    } finally {
      setBusy(false)
    }
  }

  const isWaiting = booking.status === 'WAITING'
  const isArchived = !!booking.archivedAt
  const finished = booking.status === 'COMPLETED' || booking.status === 'CANCELLED'

  if (isArchived) {
    return (
      <div className="mt-2 rounded-lg border border-slate-300 bg-slate-50 p-3">
        <div className="text-xs font-bold text-slate-700">🗄️ هذا الحجز بالأرشيف</div>
        {booking.archiveReason && <div className="mt-1 text-xs text-slate-500">السبب: {booking.archiveReason}</div>}
        {canArchive && (
          <button
            onClick={() => run(() => api.restoreBooking(booking.id))}
            disabled={busy}
            className="mt-2 rounded-lg border border-emerald-300 bg-white px-3 py-1.5 text-xs font-bold text-emerald-700 disabled:opacity-50"
          >
            ↩️ رجّعه للعمل
          </button>
        )}
        {err && <div className="mt-2 text-xs font-bold text-red-600">{err}</div>}
      </div>
    )
  }

  return (
    <div className="mt-2">
      {/* لافتة الانتظار: الحجز محفوظ بس مو بطابور الشغل */}
      {isWaiting && (
        <div className="mb-2 rounded-lg border border-slate-300 bg-slate-50 p-2.5 text-xs">
          <span className="font-bold text-slate-700">⏸️ في الانتظار — الزبون ما رد</span>
          <span className="mr-2 text-slate-500">
            ({booking.contactAttempts} {booking.contactAttempts === 1 ? 'محاولة' : 'محاولات'} اتصال)
          </span>
          {booking.waitingNote && <div className="mt-1 text-slate-500">{booking.waitingNote}</div>}
          <button
            onClick={() => run(() => api.resumeBooking(booking.id))}
            disabled={busy}
            className="mt-2 rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-bold text-white disabled:opacity-50"
          >
            ✅ الزبون رد — رجّعه للطابور
          </button>
        </div>
      )}

      {/* تنبيه التأجيل المتكرر: مرتين عادي، ثلاثة فما فوق علامة على شي غلط */}
      {booking.postponeCount >= 3 && (
        <div className="mb-2 rounded-lg border border-amber-300 bg-amber-50 p-2 text-xs font-bold text-amber-800">
          ⚠️ هذا الحجز تأجل {booking.postponeCount} مرات — يستاهل مراجعة
        </div>
      )}

      {!finished && (
        <div className="flex flex-wrap gap-2">
          <button
            onClick={() => { setOpen('postpone'); setWhen(toLocalInput(booking.scheduledAt)) }}
            className="rounded-lg border border-blue-300 bg-white px-3 py-1.5 text-xs font-bold text-blue-700"
          >
            📅 تأجيل الموعد
          </button>
          {!isWaiting && (
            <button
              onClick={() => setOpen('waiting')}
              className="rounded-lg border border-slate-300 bg-white px-3 py-1.5 text-xs font-bold text-slate-600"
            >
              📞 الزبون ما رد
            </button>
          )}
          {canArchive && (
            <button
              onClick={() => setOpen('archive')}
              className="rounded-lg border border-red-300 bg-white px-3 py-1.5 text-xs font-bold text-red-600"
            >
              🗑️ حذف (للأرشيف)
            </button>
          )}
        </div>
      )}

      {open === 'postpone' && (
        <div className="mt-2 rounded-xl border border-blue-200 bg-blue-50/50 p-3">
          <h4 className="mb-2 text-sm font-bold text-blue-900">📅 تأجيل الموعد</h4>
          <label className="mb-1 block text-xs font-medium text-slate-600">
            الموعد الجديد (النهاية تنحسب ساعة بعده تلقائياً)
          </label>
          <input
            type="datetime-local"
            value={when}
            onChange={(e) => setWhen(e.target.value)}
            className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-blue-500"
          />
          <label className="mb-1 mt-2 block text-xs font-medium text-slate-600">سبب التأجيل *</label>
          <input
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder="مثال: الزبون مسافر، الزبون طلب يوم آخر..."
            className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-blue-500"
          />
          <Actions
            busy={busy}
            err={err}
            onCancel={() => setOpen(null)}
            onSave={() => run(() => api.postponeBooking(booking.id, when, reason))}
            disabled={!when || !reason.trim()}
            label="أجّل الموعد"
          />
        </div>
      )}

      {open === 'waiting' && (
        <div className="mt-2 rounded-xl border border-slate-300 bg-slate-50 p-3">
          <h4 className="mb-1 text-sm font-bold text-slate-800">📞 الزبون ما رد</h4>
          <p className="mb-2 text-xs text-slate-500">
            الحجز راح ينزاح من طابور الشغل ويضل محفوظ. لمن الزبون يرد ترجّعه بضغطة.
          </p>
          <input
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder="ملاحظة (اختيارية): اتصلنا مرتين، هاتفه مغلق..."
            className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-slate-500"
          />
          <Actions
            busy={busy}
            err={err}
            onCancel={() => setOpen(null)}
            onSave={() => run(() => api.markBookingWaiting(booking.id, reason))}
            label="حطه بالانتظار"
          />
        </div>
      )}

      {open === 'archive' && (
        <div className="mt-2 rounded-xl border border-red-200 bg-red-50/60 p-3">
          <h4 className="mb-1 text-sm font-bold text-red-800">🗑️ حذف الحجز</h4>
          <p className="mb-2 text-xs text-slate-600">
            الحجز راح يختفي من الحجوزات ومن تنسيق الحجوزات، بس <b>ما ينمحي</b> — يضل
            بالأرشيف بكل تفاصيله وتكدر ترجّعه.
          </p>
          <input
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder="سبب الحذف * — مثال: الزبون ألغى الحجز"
            className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-red-500"
          />
          <Actions
            busy={busy}
            err={err}
            onCancel={() => setOpen(null)}
            onSave={() => run(() => api.archiveBooking(booking.id, reason))}
            disabled={!reason.trim()}
            label="احذفه للأرشيف"
            danger
          />
        </div>
      )}
    </div>
  )
}

function Actions({
  busy, err, onCancel, onSave, label, disabled, danger,
}: {
  busy: boolean
  err: string | null
  onCancel: () => void
  onSave: () => void
  label: string
  disabled?: boolean
  danger?: boolean
}) {
  return (
    <>
      <div className="mt-3 flex items-center gap-2">
        <button
          onClick={onSave}
          disabled={busy || disabled}
          className={`rounded-lg px-4 py-2 text-sm font-bold text-white disabled:opacity-50 ${
            danger ? 'bg-red-600' : 'bg-brand-700'
          }`}
        >
          {busy ? 'جاري...' : label}
        </button>
        <button onClick={onCancel} className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-bold text-slate-600">
          إلغاء
        </button>
      </div>
      {err && <div className="mt-2 whitespace-pre-line text-xs font-bold text-red-600">{err}</div>}
    </>
  )
}
