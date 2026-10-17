import { useCallback, useEffect, useState } from 'react'
import { api, type ServicePriceStats, type ServicePriceSuggestion } from '../api'
import { useSession } from '../session'

// ═══ أسعار يقترحها النظام ═══
//
// (ع): «النظام يدرس ويحلل هاي الأرقام حتى يطلع أرقام ويسعّر الخدمات
// الي احنه ماذاكريهن… يخلي إلها سعر أفيرج حسب معدل الأخذ من الزبائن
// السابقين». وقراره: **النظام يقترح، والمحاسب يراجع، وبعدين المالك
// يعتمد** · وبعد **٥ عيّنات** · و**ما يقترح مرة ثانية أبداً**.
//
// ⚠️ **الرقم ما ينعرض لحاله أبداً**: كل معدّل معاه عدد عيّناته وأقل
// وأعلى مبلغ. «١٧٢ ألف» من عيّنتين مو نفسه من عشرين، ومعدّل بين
// ١٥٠ و٢٠٠ ألف غير معدّل بين ٥٠ و٤٠٠ ألف — الثاني يعني الخدمة
// أسعارها متفرقة والمعدّل ما يمثّلها.
//
// ⚠️ **والزر يطلع حسب الحالة مو حسب الدور وحده**: «اعتماد» ما يظهر
// لاقتراح لسه ما راجعه المحاسب، لأن الخادم يرفضه برسالة «لازم
// المحاسب يراجعه قبل الاعتماد» — وزر يطلع ويرجع خطأ أسوأ من زر مو
// موجود.

const fmt = (n: number) => Math.round(n).toLocaleString('en-US')

const STATUS: Record<string, { label: string; cls: string }> = {
  PROPOSED: { label: '🧠 اقتراح جديد — ينتظر المحاسب', cls: 'bg-amber-50 text-amber-800 border-amber-200' },
  REVIEWED: { label: '✅ راجعه المحاسب — ينتظر المالك', cls: 'bg-sky-50 text-sky-800 border-sky-200' },
  APPROVED: { label: '✔ معتمد', cls: 'bg-emerald-50 text-emerald-800 border-emerald-200' },
  REJECTED: { label: '✖ مرفوض', cls: 'bg-red-50 text-red-700 border-red-200' },
}

export default function ServicePriceSuggestions() {
  const { employee } = useSession()
  // ⚠️ مطابق لحارس الخادم: المراجعة لـADMIN/FINANCE (`requireFinance`)،
  // والاعتماد لـADMIN/OWNER (`requireAdmin`). و`role` يتطبّع لـADMIN
  // للمالك بالجلسة، فهذا الشرط يغطي الاثنين.
  const canReview = employee?.role === 'ADMIN' || employee?.role === 'FINANCE'
  const canDecide = employee?.role === 'ADMIN'

  const [rows, setRows] = useState<ServicePriceSuggestion[] | null>(null)
  // 🔴 المبتوت لازم يبقى مرئياً: بلاه، المالك يعتمد سعراً فيختفي
  // الصف بلا أي أثر — «وين راح؟». ونفس العلّة انصلّحت قبل بطابور
  // المراقب (الي يدقّقها كانت تختفي منه تماماً).
  const [decided, setDecided] = useState<ServicePriceSuggestion[]>([])
  const [samples, setSamples] = useState<ServicePriceStats[]>([])
  const [failed, setFailed] = useState(false)
  const [busy, setBusy] = useState<string | null>(null)
  const [msg, setMsg] = useState<string | null>(null)

  const load = useCallback(() => {
    Promise.all([
      api.getServicePriceSuggestions(),
      api.getServicePriceSamples(),
      api.getServicePriceSuggestions('APPROVED'),
      api.getServicePriceSuggestions('REJECTED'),
    ])
      .then(([g, s, ap, rj]) => {
        setRows(g); setSamples(s)
        setDecided([...ap, ...rj].sort((a, b) => (b.decidedAt || '').localeCompare(a.decidedAt || '')))
        setFailed(false)
      })
      .catch(() => setFailed(true))
  }, [])
  useEffect(() => { load() }, [load])

  const act = async (id: string, fn: () => Promise<unknown>, done: string) => {
    setBusy(id); setMsg(null)
    try { await fn(); setMsg(done); load() } catch (e) {
      setMsg(e instanceof Error ? e.message : 'تعذر تنفيذ الطلب')
    } finally { setBusy(null) }
  }

  if (failed) return <p className="text-sm text-red-600">تعذّر جلب اقتراحات الأسعار</p>
  if (rows === null) return <p className="text-sm text-slate-400">جاري التحميل…</p>

  // العيّنات الي لسه ما وصلت الحد — يشوف شنو يتجمّع
  const growing = samples.filter((s) => s.remaining > 0)

  return (
    <div dir="rtl" className="space-y-4">
      <p className="rounded-lg border border-slate-200 bg-slate-50/70 px-3 py-2 text-[12px] leading-relaxed text-slate-600">
        ⓘ النظام يجمع أسعار الفواتير اليدوية لكل خدمة، وبعد <b>٥ عيّنات</b> يطلّع
        <b> معدّلاً مقترحاً</b> — مو سعراً. المحاسب يراجعه، والمالك يعتمده.
        <br />والاقتراح <b>مرة وحدة لكل خدمة</b>: بعد ما تبتّ بيه ما يرجع يزعجك،
        والعيّنات تستمر تتجمّع وتبان تحت.
      </p>

      {msg && (
        <p className={`text-sm font-bold ${msg.includes('✓') ? 'text-emerald-600' : 'text-red-600'}`}>{msg}</p>
      )}

      {rows.length === 0 ? (
        <p className="text-sm text-slate-500">ماكو اقتراحات — لسه ماكو خدمة وصلت ٥ عيّنات.</p>
      ) : (
        <div className="space-y-2">
          {rows.map((g) => {
            const st = STATUS[g.status] ?? { label: g.status, cls: 'bg-slate-50 text-slate-600 border-slate-200' }
            // ⚠️ التشتّت يبان: فرق كبير بين الأقل والأعلى يعني المعدّل
            // ما يمثّل الخدمة، والمحاسب لازم يشوفها قبل ما يصادق.
            const spread = g.maxAmount > 0 ? (g.maxAmount - g.minAmount) / g.maxAmount : 0
            return (
              <div key={g.id} className="rounded-xl border border-slate-200 bg-white p-3">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <p className="font-extrabold text-[#0f2040]">
                    🛠️ {g.serviceName || '—'}
                  </p>
                  <span className={`rounded-md border px-2 py-0.5 text-[11px] font-bold ${st.cls}`}>{st.label}</span>
                </div>

                <p className="mt-1.5 text-sm font-bold text-[#0f2040]">
                  المعدّل المقترح: {fmt(g.avgAmount)} د.ع
                  <span className="ms-2 text-[11px] font-bold text-slate-500">
                    من {g.sampleCount} عيّنة · أقل {fmt(g.minAmount)} · أعلى {fmt(g.maxAmount)}
                  </span>
                </p>
                {spread >= 0.5 && (
                  <p className="mt-1 text-[11px] font-bold text-amber-800">
                    ⚠️ الأسعار متفرقة هواي (الأعلى أكثر من ضعف الأقل) — المعدّل ممكن ما يمثّل الخدمة
                  </p>
                )}

                {g.reviewedName && (
                  <p className="mt-1 text-[11px] text-slate-500">
                    راجعه <b className="text-[#0f2040]">{g.reviewedName}</b>
                    {g.reviewNote ? ` — ${g.reviewNote}` : ''}
                  </p>
                )}
                {g.decidedName && (
                  <p className="mt-0.5 text-[11px] text-slate-500">
                    بتّ بيه <b className="text-[#0f2040]">{g.decidedName}</b>
                  </p>
                )}

                <div className="mt-2 flex flex-wrap gap-2">
                  {g.status === 'PROPOSED' && canReview && (
                    <button
                      type="button" disabled={busy === g.id}
                      onClick={() => {
                        const note = window.prompt('ملاحظتك على المعدّل (اختيارية)', 'المعدّل معقول')
                        if (note === null) return
                        void act(g.id, () => api.reviewServicePrice(g.id, note.trim()), 'انتأشرت مراجعة ✓')
                      }}
                      className="rounded-lg bg-sky-600 px-3 py-1.5 text-xs font-bold text-white hover:bg-sky-700 disabled:opacity-50"
                    >✅ راجعتها</button>
                  )}
                  {g.status === 'REVIEWED' && canDecide && (
                    <button
                      type="button" disabled={busy === g.id}
                      onClick={() => {
                        if (!window.confirm(`اعتماد ${fmt(g.avgAmount)} د.ع سعراً مقترحاً لـ«${g.serviceName || ''}»؟`)) return
                        void act(g.id, () => api.decideServicePrice(g.id, true), 'انعتمد ✓')
                      }}
                      className="rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-bold text-white hover:bg-emerald-700 disabled:opacity-50"
                    >✔ اعتماد</button>
                  )}
                  {(g.status === 'PROPOSED' || g.status === 'REVIEWED') && canDecide && (
                    <button
                      type="button" disabled={busy === g.id}
                      onClick={() => {
                        if (!window.confirm('رفض هذا الاقتراح؟ ما يرجع يطلع لهاي الخدمة.')) return
                        void act(g.id, () => api.decideServicePrice(g.id, false), 'انرفض ✓')
                      }}
                      className="rounded-lg border border-red-300 bg-red-50 px-3 py-1.5 text-xs font-bold text-red-700 hover:bg-red-100 disabled:opacity-50"
                    >✖ رفض</button>
                  )}
                  {g.status === 'PROPOSED' && !canReview && (
                    <span className="text-[11px] font-bold text-slate-400">ينتظر مراجعة المحاسب</span>
                  )}
                  {g.status === 'REVIEWED' && !canDecide && (
                    <span className="text-[11px] font-bold text-slate-400">ينتظر اعتماد المالك</span>
                  )}
                </div>
              </div>
            )
          })}
        </div>
      )}

      {/* الي انبتّ بيه — ما يختفي */}
      {decided.length > 0 && (
        <div className="rounded-xl border border-slate-200 bg-white p-3">
          <p className="text-xs font-extrabold text-slate-600">📘 أسعار انبتّ بيها</p>
          <div className="mt-1.5 space-y-1">
            {decided.map((g) => (
              <p key={g.id} className="text-[11px] text-slate-600">
                <b className="text-[#0f2040]">{g.serviceName || '—'}</b> —{' '}
                {g.status === 'APPROVED'
                  ? <span className="font-bold text-emerald-700">✔ معتمد {fmt(g.avgAmount)} د.ع</span>
                  : <span className="font-bold text-red-700">✖ مرفوض ({fmt(g.avgAmount)} د.ع)</span>}
                {' '}· من {g.sampleCount} عيّنة
                {g.decidedName ? ` · بتّ بيه ${g.decidedName}` : ''}
                {g.decidedAt ? ` · ${new Date(g.decidedAt).toLocaleDateString('ar-IQ')}` : ''}
              </p>
            ))}
          </div>
          <p className="mt-1.5 text-[10px] text-slate-400">
            ⓘ هذولا ما يرجعون يطلعون اقتراحاً — العيّنات تستمر تتجمّع وتبان تحت.
          </p>
        </div>
      )}

      {/* شنو يتجمّع — حتى يعرف إن الشغل ماشي مو واقف */}
      {growing.length > 0 && (
        <div className="rounded-xl border border-slate-200 bg-slate-50/60 p-3">
          <p className="text-xs font-extrabold text-slate-600">📈 خدمات لسه تجمّع عيّناتها</p>
          <div className="mt-1.5 space-y-1">
            {growing.map((s) => (
              <p key={s.serviceId} className="text-[11px] text-slate-600">
                <b className="text-[#0f2040]">{s.serviceName}</b> — {s.sampleCount} عيّنة
                (معدّلها الحالي {fmt(s.avgAmount)} د.ع) · باقي <b>{s.remaining}</b> حتى يطلع الاقتراح
              </p>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
