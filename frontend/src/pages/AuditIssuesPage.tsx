import { useEffect, useState } from 'react'
import { api, auditCloseLabels, type AuditCloseAction, type AuditIssue } from '../api'
import { useSaveGuard } from '../useSaveGuard'
import SaveError from '../components/SaveError'
import { useSession } from '../session'

/**
 * بلاغات أخطاء التدقيق.
 *
 * المحاسب يأشّرها وهو يدقق، والنظام يوجّهها:
 *   - «المبلغ غير مطابق» → الرقابة والجودة
 *   - «خطأ بالسعر»        → الرقابة والإداري
 *
 * كل واحد يشوف الي يخصه بس — الفلترة تصير بالسيرفر حسب دوره.
 */

const fmt = (n: number | null) => (n == null ? '—' : n.toLocaleString('en-IQ') + ' د.ع')

/** ⚠️ `embedded`: نفس الشاشة بالضبط بلا ترويستها — تنضمّ بمكتب
 *  المراقب. **ما ننسخ المحتوى**: نسختان تفترقان بأول تصحيح، فالمراقب
 *  يشوف صفاً بشاشة ومحلولاً بالثانية ويفقد الثقة بالاثنتين. */
interface EmbeddedProps { embedded?: boolean }

export default function AuditIssuesPage({ embedded }: EmbeddedProps = {}) {
  // ═══ المحاسب مو مراقب ═══
  // نفس الشاشة، بس معناها يختلف: عند المحاسب **صادر** — هو الي أشّر
  // الأخطاء ويتابع شنو صار بيها. وعند المراقب **وارد للتدقيق** — يروح
  // يتأكد من الليدر ليش عنده أخطاء. السيرفر يفلتر (المحاسب يشوف بلاغاته
  // بس)، وهنا نغيّر العنوان والأعمدة حتى الواجهة تگول نفس الكلام.
  const { employee } = useSession()
  const asAccountant = employee?.role === 'FINANCE'
  const [items, setItems] = useState<AuditIssue[]>([])
  const [loading, setLoading] = useState(true)
  const [busy, setBusy] = useState<string | null>(null)

  const load = () => {
    api.getAuditIssues()
      .then(setItems)
      .catch(() => setItems([]))
      .finally(() => setLoading(false))
  }
  useEffect(load, [])

  // ⚠️ البلاغ ما ينغلق إلا بإجراء وسبب — الخادم يرفض بدونهما.
  // قبلها چان زراً واحداً يسكّره بلا أثر: ماكو منو أغلقه ولا ليش
  // ولا إشعار للمحاسب ولا أثر على الليدر. «وين يروح البلاغ؟»
  // چان جوابه: ما يروح لأي مكان.
  const [closing, setClosing] = useState<{ issue: AuditIssue; action: AuditCloseAction } | null>(null)
  const [reason, setReason] = useState('')
  const [points, setPoints] = useState(1)
  const guard = useSaveGuard()

  const resolve = async () => {
    if (!closing) return
    setBusy(closing.issue.id)
    const ok = await guard.run('إغلاق البلاغ', () =>
      api.resolveAuditIssue(closing.issue.id, {
        action: closing.action, reason, points,
      }))
    setBusy(null)
    if (ok) {
      setClosing(null)
      setReason('')
      setPoints(1)
      load()
    }
  }

  const open = items.filter((i) => i.status === 'OPEN')

  return (
    <div dir="rtl" className="space-y-6">
      <SaveError message={guard.error} onClose={guard.clear} />

      {closing && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
          onClick={() => setClosing(null)}>
          <div onClick={(e) => e.stopPropagation()}
            className="w-full max-w-md rounded-2xl border p-5"
            style={{ backgroundColor: 'var(--sf-card)', borderColor: 'var(--bd-line)' }}>
            <h3 className="text-base font-bold" style={{ color: 'var(--t-title)' }}>
              {auditCloseLabels[closing.action]} — بلاغ {closing.issue.bookingCode}
            </h3>
            <p className="mt-1 text-xs" style={{ color: 'var(--t-muted)' }}>
              {closing.action === 'PENALIZE'
                ? `المخالفة تنسجّل على ${closing.issue.leaderName || 'ليدر الحجز'}، وتنعرض بلوحة الإعلانات، وتنحط بسجله الانضباطي.`
                : 'يعني راجعت وتأكدت إن ماكو خطأ فعلاً — والسبب ينحفظ بالبلاغ.'}
            </p>

            {closing.action === 'PENALIZE' && (
              <label className="mt-3 block text-xs font-bold" style={{ color: 'var(--t-body)' }}>
                نقاط الخصم
                <select value={points} onChange={(e) => setPoints(Number(e.target.value))}
                  className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm">
                  {[1, 2, 3, 4, 5].map((n) => <option key={n} value={n}>{n}</option>)}
                </select>
              </label>
            )}

            {/* ⚠️ السبب إجباري بالحالتين — بدونه «تأكدت ماكو خطأ»
                تصير باباً خلفياً للإغلاق الروتيني. */}
            <label className="mt-3 block text-xs font-bold" style={{ color: 'var(--t-body)' }}>
              السبب (إجباري)
              <textarea value={reason} onChange={(e) => setReason(e.target.value)} rows={3}
                placeholder={closing.action === 'PENALIZE'
                  ? 'شنو التقصير بالضبط؟'
                  : 'شنو راجعت وشلون تأكدت؟'}
                className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
            </label>

            <div className="mt-4 flex justify-end gap-2">
              <button onClick={() => setClosing(null)}
                className="rounded-lg border px-4 py-2 text-sm font-bold"
                style={{ borderColor: 'var(--bd-line)', color: 'var(--t-body)' }}>
                إلغاء
              </button>
              <button
                disabled={!reason.trim() || busy === closing.issue.id}
                onClick={resolve}
                className={`rounded-lg px-4 py-2 text-sm font-bold text-white disabled:opacity-50 ${
                  closing.action === 'PENALIZE' ? 'bg-red-600' : 'bg-emerald-700'}`}>
                {busy === closing.issue.id ? 'جاري...' : 'أكّد وأغلق البلاغ'}
              </button>
            </div>
          </div>
        </div>
      )}
      {!embedded && (
        <div
          className="relative overflow-hidden rounded-2xl p-6 shadow-md"
          style={{ background: 'linear-gradient(135deg, #1a3a5c 0%, #24507e 55%, #2f6ba8 100%)' }}
        >
          <span
            aria-hidden
            className="pointer-events-none absolute -left-16 -top-24 h-64 w-64 rounded-full opacity-20"
            style={{ background: 'radial-gradient(circle, #c8a45a 0%, transparent 70%)' }}
          />
          <div className="relative flex flex-wrap items-end justify-between gap-4">
            <div>
              <h1 className="text-2xl font-black text-white">
                {asAccountant ? '💸 أخطاء الفواتير الي أشّرتها' : '💸 بلاغات أخطاء التدقيق'}
              </h1>
              <p className="mt-1 max-w-xl text-sm text-blue-100">
                {asAccountant
                  ? 'هاي البلاغات الي أرسلتها إنت وأنت تدقق — تابع شنو صار بيها.'
                  : 'أشّرها المحاسب وهو يدقق الحجوزات — تدقّق بيها على الليدر.'}
              </p>
            </div>
            <div className="rounded-xl bg-amber-400/20 px-4 py-2 text-center ring-1 ring-amber-200/40 backdrop-blur">
              <p className="text-2xl font-black leading-none text-amber-100">{open.length}</p>
              <p className="mt-1 text-[11px] text-amber-50">لسه مفتوح</p>
            </div>
          </div>
        </div>
      )}

      {loading && <p className="text-center text-slate-400">جاري التحميل...</p>}

      <div className="space-y-3">
        {items.map((i) => (
          <div
            key={i.id}
            /* شريط بلون البلاغ: المفتوح يشدّ العين والمنحسم يهدى — الشاشة
               تطول بالوقت وكلها تبين نفس الشي بلا هذا. */
            className={`relative overflow-hidden rounded-xl border bg-white p-4 pr-5 shadow-sm transition hover:shadow-md ${
              i.status !== 'OPEN' ? 'border-slate-200 opacity-75'
                : i.kind === 'MISMATCH' ? 'border-amber-200' : 'border-red-200'
            }`}
          >
            <span
              aria-hidden
              className={`absolute inset-y-0 right-0 w-1.5 ${
                i.status !== 'OPEN' ? 'bg-slate-300' : i.kind === 'MISMATCH' ? 'bg-amber-400' : 'bg-red-400'
              }`}
            />
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p className="font-bold text-slate-800">
                  <span className="rounded-lg bg-slate-100 px-2 py-0.5 font-mono text-sm text-slate-700">{i.bookingCode}</span>
                  <span className="mr-2 text-sm font-normal text-slate-500">{i.customerName}</span>
                </p>
                <p className="mt-1 text-sm">
                  <span className={`rounded-full px-3 py-1 text-xs font-bold ${
                    i.kind === 'MISMATCH' ? 'bg-amber-100 text-amber-800' : 'bg-red-100 text-red-800'
                  }`}>
                    {i.kindLabel}
                  </span>
                  <span className="mr-2 text-xs text-slate-400">← {i.routedTo}</span>
                </p>
                <p className="mt-2 text-sm text-slate-600">
                  المبلغ بالفاتورة: <b>{fmt(i.expectedAmount)}</b> · المسجّل بالنظام: <b>{fmt(i.actualAmount)}</b>
                </p>
                {i.note && <p className="mt-1 text-sm text-slate-600">📝 {i.note}</p>}
                <p className="mt-1 text-xs text-slate-400">
                  سجّله: {i.raisedByName} · {new Date(i.createdAt).toLocaleDateString('ar-IQ')}
                </p>
                {/* الليدر يظهر للمراقب بس — هو الي يروح يسأله. المحاسب
                    يعرف ليدره أصلاً من الفاتورة الي دققها. */}
                {!asAccountant && i.leaderName && (
                  <p className="mt-1 text-sm font-bold text-slate-700">👷 الليدر: {i.leaderName}</p>
                )}
              </div>
              <span className={`rounded-full px-3 py-1 text-xs font-bold ${
                i.status === 'OPEN' ? 'bg-amber-100 text-amber-800' : 'bg-slate-200 text-slate-600'
              }`}>
                {i.status === 'OPEN' ? 'مفتوح' : 'انحسم'}
              </span>
            </div>

            {i.status === 'OPEN' && (
              <div className="mt-3 flex flex-wrap gap-2">
                {/* ⚠️ زران مو واحد: الإغلاق قرار — إما تعاقب، أو
                    تصرّح إنك تأكدت وماكو خطأ. والاثنان ينحفظان. */}
                <button disabled={busy === i.id}
                  onClick={() => { setClosing({ issue: i, action: 'PENALIZE' }); setReason('') }}
                  className="rounded-lg bg-red-600 px-4 py-2 text-sm font-bold text-white hover:bg-red-700 disabled:opacity-50">
                  ⚠️ {auditCloseLabels.PENALIZE}
                </button>
                <button disabled={busy === i.id}
                  onClick={() => { setClosing({ issue: i, action: 'NO_FAULT' }); setReason('') }}
                  className="rounded-lg border border-emerald-300 bg-emerald-50 px-4 py-2 text-sm font-bold text-emerald-800 hover:bg-emerald-100 disabled:opacity-50">
                  ✔ {auditCloseLabels.NO_FAULT}
                </button>
              </div>
            )}

            {/* البلاغ المغلق يبيّن منو أغلقه وشنو سوّى — قبلها
                چان يختفي بلا أثر. */}
            {i.status === 'RESOLVED' && i.actionKind && (
              <p className="mt-2 rounded-lg bg-slate-50 px-3 py-2 text-xs text-slate-600">
                {auditCloseLabels[i.actionKind]}
                {i.resolvedByName && <> · أغلقه <b>{i.resolvedByName}</b></>}
                {i.resolveReason && <> — {i.resolveReason}</>}
              </p>
            )}
          </div>
        ))}
        {!loading && items.length === 0 && (
          <div className="rounded-2xl border border-dashed border-slate-200 bg-white p-10 text-center shadow-sm">
            <p className="text-4xl">✅</p>
            <p className="mt-3 font-bold text-slate-600">ماكو بلاغات</p>
            <p className="mx-auto mt-1 max-w-sm text-sm text-slate-400">
              {asAccountant
                ? 'ما أشّرت أي خطأ لحد هسه — البلاغ ينسجّل لمن تختار «غير مطابق» أو «خطأ بالسعر» بالتدقيق اليومي.'
                : 'ماكو خطأ أشّره المحاسب — لمن يأشّر واحد يوصلك هنا باسم الليدر صاحب الفاتورة.'}
            </p>
          </div>
        )}
      </div>
    </div>
  )
}
