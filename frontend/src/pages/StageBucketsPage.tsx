import { useCallback, useEffect, useState } from 'react'
import { api, STAGE_BUCKETS, type Booking, type StageBucket } from '../api'
import EntityIdentity from '../components/EntityIdentity'

// ═══ سلال المراحل ═══
//
// «حجوزات مؤجلة وزبون ما رد وزبون ألغى — هاي تصير بيها حالتين: قبل
// التثبيت وبعد التثبيت».
//
// وهو محق: الزبون الي ألغى **قبل** ما نثبتله موعد ما كلّفنا شي، والي
// ألغى **بعد** ما وعدناه وحضّرنا كادر وحجزنا يومه كلّفنا فعلاً. دمج
// الاثنين بسلّة وحدة يخبّي فرقاً بالمسؤولية وبالخسارة، ويخلي «شنو
// صار بهذا الشهر؟» ما إلها جواب دقيق.
//
// ⚠️ الفرز يصير بالسيرفر: الواجهة چان لازم تجيب كل حجوزات الشركة
// وتفرزهن بالمتصفح، وهذا يثقل كل ما تتراكم البيانات.
export default function StageBucketsPage() {
  const [bucket, setBucket] = useState<StageBucket>('POSTPONED_AFTER_CONFIRM')
  const [rows, setRows] = useState<Booking[]>([])
  const [counts, setCounts] = useState<Record<string, number>>({})
  const [loading, setLoading] = useState(true)

  const load = useCallback(() => {
    setLoading(true)
    Promise.all([api.getBookingsByStageBucket(bucket), api.getStageBucketCounts()])
      .then(([list, c]) => { setRows(list); setCounts(c) })
      .catch(() => setRows([]))
      .finally(() => setLoading(false))
  }, [bucket])

  useEffect(load, [load])

  const meta = STAGE_BUCKETS.find((b) => b.key === bucket)

  return (
    <div dir="rtl">
      <h2 className="text-xl font-extrabold text-[#0f2040]">🗂️ حجوزات ما وصلت للتنفيذ</h2>
      <p className="mt-1 text-sm text-slate-500">
        المؤجّل والي ما رد والملغى — كل واحد مفروز: صار <b>قبل</b> التثبيت لو <b>بعده</b>.
      </p>

      <div className="mt-4 flex flex-wrap gap-2">
        {STAGE_BUCKETS.map((b) => (
          <button
            key={b.key}
            onClick={() => setBucket(b.key)}
            className={`rounded-xl px-3.5 py-2 text-sm font-bold transition-colors ${
              bucket === b.key ? 'bg-[#0f2040] text-white' : 'bg-white text-slate-600 hover:bg-slate-50'
            }`}
          >
            {b.icon} {b.label}
            {counts[b.key] > 0 && (
              <span className={`mr-1.5 rounded-full px-1.5 text-xs ${bucket === b.key ? 'bg-white/20' : 'bg-slate-100'}`}>
                {counts[b.key]}
              </span>
            )}
          </button>
        ))}
      </div>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {!loading && rows.length === 0 && (
        <p className="mt-6 rounded-xl border border-white bg-white p-8 text-center text-slate-400">
          ماكو حجوزات بـ«{meta?.label}» ✓
        </p>
      )}

      <div className="mt-4 space-y-3">
        {rows.map((b) => (
          <div key={b.id} className="rounded-2xl bg-white p-4 shadow-[0_2px_12px_rgba(0,0,0,0.06)]">
            <EntityIdentity booking={b} variant="full" />

            <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-slate-500">
              {/* منو أدخل الحجز — أول سؤال ينسأل لما تكون المعلومة ناقصة */}
              {b.createdByName && <span>📝 أدخله: <b className="text-slate-700">{b.createdByName}</b></span>}
              {b.confirmedByName && <span>✅ ثبّته: <b className="text-slate-700">{b.confirmedByName}</b></span>}
              {b.postponeCount > 0 && <span>⏳ انأجّل {b.postponeCount} مرة</span>}
              {b.contactAttempts > 0 && <span>📞 {b.contactAttempts} محاولة اتصال</span>}
            </div>

            {/* السبب — هو الي يفرّق بين «الزبون بدّل رأيه» و«غلط منّا» */}
            {b.cancelReason && (
              <p className="mt-2 rounded-lg bg-red-50 px-3 py-2 text-xs text-red-800">
                سبب الإلغاء: <b>{b.cancelReason}</b>
                {b.cancelledByName && <> — ألغاه {b.cancelledByName}</>}
                {b.cancelledAt && <> · {new Date(b.cancelledAt).toLocaleString('en-GB')}</>}
              </p>
            )}
            {b.postponeReason && !b.cancelReason && (
              <p className="mt-2 rounded-lg bg-amber-50 px-3 py-2 text-xs text-amber-900">
                سبب التأجيل: <b>{b.postponeReason}</b>
              </p>
            )}
            {b.waitingNote && !b.cancelReason && (
              <p className="mt-2 rounded-lg bg-slate-50 px-3 py-2 text-xs text-slate-700">
                ملاحظة الانتظار: <b>{b.waitingNote}</b>
              </p>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
