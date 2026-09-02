import { Fragment, useEffect, useMemo, useState } from 'react'
import { api, COORDINATION_ALERT_THRESHOLD, type Booking, type CoordinationAlert, type CoordinationAlertSummary } from '../api'
import { useSaveGuard } from '../useSaveGuard'
import SaveError from '../components/SaveError'
import EmptyState from '../components/EmptyState'

// صفحة تدقيق للمراقب (صلاحية crew_management) — تعرض الحجوزات الموجّهة/المسندة
// من موظف مبيعات (أو غيره) لكن لسه ما ثبّتها الإداري (حالة PENDING). الهدف:
// يقدر المراقب يقارن ويدقق هل الإداري تواصل فعلاً مع الزبون وأقفل الاتفاق قبل
// التثبيت (حقل confirmationContactedAt) أو لسه ما تواصل، من مصدر مستقل عن
// المسار العام /api/bookings.
//
// ⚠️ العدّاد N/10 وسجله والإعلان عند العاشر — كلها جديدة هذي الجولة.
// المراقب يضغط «تسجيل تقصير» فيزيد العدّاد واحداً، وبالعاشر ينشر
// الخادم إعلاناً يسمّي الإداري والحجز (مرة وحدة بالضبط).
/** ⚠️ `embedded`: نفس الشاشة بالضبط بلا ترويستها — تنضمّ بمكتب
 *  المراقب. **ما ننسخ المحتوى**: نسختان تفترقان بأول تصحيح، فالمراقب
 *  يشوف صفاً بشاشة ومحلولاً بالثانية ويفقد الثقة بالاثنتين. */
interface EmbeddedProps { embedded?: boolean }

function relTime(iso: string | null) {
  if (!iso) return '—'
  return new Date(iso).toLocaleString('ar-IQ', { dateStyle: 'short', timeStyle: 'short' })
}

export default function MonitorCrewBookingsPage({ embedded }: EmbeddedProps = {}) {
  const [bookings, setBookings] = useState<Booking[]>([])
  const [summaries, setSummaries] = useState<Record<string, CoordinationAlertSummary>>({})
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const guard = useSaveGuard()

  const [reasonFor, setReasonFor] = useState<string | null>(null)
  const [reason, setReason] = useState('')
  const [resolveFor, setResolveFor] = useState<string | null>(null)
  const [resolveNote, setResolveNote] = useState('')
  const [logFor, setLogFor] = useState<string | null>(null)
  const [log, setLog] = useState<CoordinationAlert[] | null>(null)

  const load = () => {
    Promise.all([api.getPendingAudit(), api.getCoordinationAlertSummaries()])
      .then(([b, s]) => {
        setBookings(b)
        setSummaries(Object.fromEntries(s.map((x) => [x.bookingId, x])))
      })
      .catch((e) => setError(e instanceof Error ? e.message : 'حدث خطأ'))
      .finally(() => setLoading(false))
  }
  useEffect(load, [])

  const summaryFor = useMemo(() => (id: string): CoordinationAlertSummary =>
    summaries[id] ?? { bookingId: id, openCount: 0, totalCount: 0, lastAlertAt: null }, [summaries])

  const submitAlert = async (b: Booking) => {
    if (reasonFor !== b.id) { setReasonFor(b.id); setReason(''); return }
    if (!reason.trim()) return
    const ok = await guard.run('تسجيل التقصير', () => api.addCoordinationAlert(b.id, reason.trim()))
    if (ok) { setReasonFor(null); setReason(''); load() }
  }

  const submitResolve = async (b: Booking) => {
    if (resolveFor !== b.id) { setResolveFor(b.id); setResolveNote(''); return }
    const ok = await guard.run('تعليم المعالجة', () => api.resolveCoordinationAlerts(b.id, resolveNote.trim()))
    if (ok !== undefined) { setResolveFor(null); setResolveNote(''); load() }
  }

  const showLog = async (b: Booking) => {
    if (logFor === b.id) { setLogFor(null); setLog(null); return }
    setLogFor(b.id)
    const rows = await guard.run('جلب السجل', () => api.getCoordinationAlertsForBooking(b.id))
    setLog(rows ?? [])
  }

  return (
    <div>
      {!embedded && (
        <>
          <h2 className="text-2xl font-bold text-brand-900">تدقيق تنسيق الحجوزات</h2>
          <p className="mt-1 text-slate-500">
            الحجوزات الموجّهة من المبيعات (أو غيرهم) وما زالت بانتظار تثبيت الإداري — لمقارنة
            هل الإداري تواصل فعلاً مع الزبون وأقفل الاتفاق قبل التثبيت.
          </p>
        </>
      )}

      <SaveError message={guard.error} onClose={guard.clear} />

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">تعذر الاتصال بالخادم: {error}</p>}

      {!loading && !error && (
        <div className="mt-6 flex flex-col gap-3">
          {bookings.length === 0 && (
            <EmptyState icon="📋" title="ماكو حجوزات بانتظار التثبيت حالياً"
              reason="الحجوزات توصل هنا لمن يوجّهها المبيعات أو غيرهم وتبقى PENDING." />
          )}
          {bookings.map((b) => {
            const s = summaryFor(b.id)
            const isOverThreshold = s.openCount >= COORDINATION_ALERT_THRESHOLD
            const isLogOpen = logFor === b.id
            return (
              <Fragment key={b.id}>
                <div
                  className={`rounded-xl border bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)] ${
                    isOverThreshold ? 'border-red-300' : 'border-amber-200'}`}
                >
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <span className="font-mono text-sm font-semibold text-brand-600">{b.code}</span>
                    <div className="flex items-center gap-2">
                      <span className="rounded-full bg-amber-100 px-3 py-1 text-xs font-bold text-amber-700">
                        بانتظار التثبيت
                      </span>
                      {!b.confirmationContactedAt && (
                        <span className="rounded-full bg-red-100 px-3 py-1 text-xs font-bold text-red-700">
                          بانتظار تواصل الإداري
                        </span>
                      )}
                      <span className={`rounded-full px-3 py-1 text-xs font-bold tabular-nums ${
                        isOverThreshold ? 'bg-red-600 text-white' : 'bg-slate-100 text-slate-600'}`}>
                        {s.openCount}/{COORDINATION_ALERT_THRESHOLD}
                      </span>
                    </div>
                  </div>

                  <div className="mt-2 grid grid-cols-1 gap-1 text-sm text-slate-500 sm:grid-cols-2 lg:grid-cols-4">
                    <p><span className="text-slate-400">الزبون: </span>{b.customer?.name || 'زبون غير معروف'}</p>
                    <p><span className="text-slate-400">الهاتف: </span>{b.customer?.phone || '-'}</p>
                    <p><span className="text-slate-400">وجّهه: </span>{b.transferEmployee?.name || 'غير محدد'}</p>
                    <p><span className="text-slate-400">الإداري المسؤول: </span>{b.confirmedByEmployee?.name || 'غير محدد'}</p>
                    <p><span className="text-slate-400">موعد الحجز: </span>{b.scheduledAt ? relTime(b.scheduledAt) : 'غير محدد'}</p>
                    <p><span className="text-slate-400">آخر تنبيه: </span>{relTime(s.lastAlertAt)}</p>
                    <p><span className="text-slate-400">إجمالي التنبيهات: </span>{s.totalCount}</p>
                  </div>

                  <div className="mt-3">
                    {b.confirmationContactedAt ? (
                      <div className="rounded-lg bg-emerald-50 px-3 py-2 text-xs font-bold text-emerald-700">
                        ✅ الإداري تواصل مع الزبون وأقفل الاتفاق ({b.confirmationContactedBy?.name || 'غير معروف'}) —{' '}
                        {relTime(b.confirmationContactedAt)}
                      </div>
                    ) : (
                      <div className="rounded-lg bg-red-50 px-3 py-2 text-xs font-bold text-red-700">
                        ⏳ الإداري لسه ما تواصل مع الزبون (ما ضغط "تم")
                      </div>
                    )}
                  </div>

                  {reasonFor === b.id && (
                    <input
                      value={reason} onChange={(e) => setReason(e.target.value)} placeholder="شنو التقصير؟"
                      className="mt-3 w-full rounded-lg border border-red-300 px-3 py-2 text-right text-sm outline-none focus:border-red-500"
                    />
                  )}
                  {resolveFor === b.id && (
                    <input
                      value={resolveNote} onChange={(e) => setResolveNote(e.target.value)} placeholder="ملاحظة المعالجة (اختياري)"
                      className="mt-3 w-full rounded-lg border border-emerald-300 px-3 py-2 text-right text-sm outline-none focus:border-emerald-500"
                    />
                  )}

                  <div className="mt-3 flex flex-wrap gap-2">
                    <button disabled={guard.busy} onClick={() => submitAlert(b)}
                      className="rounded-lg bg-red-600 px-4 py-2 text-xs font-bold text-white hover:bg-red-700 disabled:opacity-50">
                      {reasonFor === b.id ? 'تأكيد التسجيل' : '⚠️ تسجيل تقصير'}
                    </button>
                    <button disabled={guard.busy || s.openCount === 0} onClick={() => submitResolve(b)}
                      className="rounded-lg border border-emerald-500 px-4 py-2 text-xs font-bold text-emerald-700 hover:bg-emerald-50 disabled:opacity-50">
                      {resolveFor === b.id ? 'تأكيد المعالجة' : '✅ تمت المعالجة'}
                    </button>
                    <button onClick={() => showLog(b)}
                      className="rounded-lg border border-slate-300 px-4 py-2 text-xs font-bold text-slate-600 hover:bg-slate-50">
                      {isLogOpen ? '▲ اخفي السجل' : '📄 عرض السجل'}
                    </button>
                  </div>
                </div>

                {isLogOpen && (
                  <div className="rounded-xl border border-slate-200 bg-slate-50 p-3">
                    {log === null ? (
                      <p className="text-xs text-slate-400">جاري التحميل...</p>
                    ) : log.length === 0 ? (
                      <p className="text-xs text-slate-400">ماكو تنبيهات مسجّلة على هذا الحجز.</p>
                    ) : (
                      <ol className="space-y-1.5">
                        {log.map((a) => (
                          <li key={a.id} className="flex flex-wrap items-baseline gap-2 text-xs text-slate-600">
                            <span className="text-slate-400">{relTime(a.createdAt)}</span>
                            {a.resolvedAt ? (
                              <b className="text-emerald-700">معالج — {a.resolvedByName}{a.resolveNote ? ` (${a.resolveNote})` : ''}</b>
                            ) : (
                              <b className="text-red-700">تقصير — {a.byName}{a.reason ? `: ${a.reason}` : ''}</b>
                            )}
                          </li>
                        ))}
                      </ol>
                    )}
                  </div>
                )}
              </Fragment>
            )
          })}
        </div>
      )}
    </div>
  )
}
