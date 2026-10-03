import { useEffect, useState } from 'react'
import { api, type DuplicateCandidate } from '../api'
import EntityIdentity from '../components/EntityIdentity'

// ═══ تدقيق التكرار (ماتركس) — حجوزات وزبائن مكررون بالغلط ═══
//
// ⚠️ هذا مسار مستقل عن صندوق المراقب (MonitorReview) بالتصميم —
// الأخير مبني على «معرّف واحد ← هوية واحدة»، وزوج التكرار علاقة بين
// سجلين مو هوية مفردة. شوف الخطة (دفعة ماتركس الكبيرة) للتفصيل.
//
// ولا حذف أو دمج تلقائي هنا — الفحص يكتشف بس، والقرار والتصرف
// الفعلي (أرشفة، دمج زبون) يبقى بيد صاحب النظام عبر الأدوات الموجودة.

type Tab = 'BOOKING' | 'CUSTOMER'

export default function DuplicateReviewPage() {
  const [tab, setTab] = useState<Tab>('BOOKING')
  const [rows, setRows] = useState<DuplicateCandidate[]>([])
  const [loading, setLoading] = useState(true)
  const [err, setErr] = useState<string | null>(null)
  const [busyId, setBusyId] = useState<string | null>(null)

  useEffect(() => {
    let alive = true
    queueMicrotask(() => { if (alive) setLoading(true) })
    void (async () => {
      try {
        const data = await api.getDuplicateCandidates(tab, 'PENDING')
        if (alive) setRows(data)
      } catch (e) {
        if (alive) setErr(e instanceof Error ? e.message : 'تعذر الجلب')
      } finally {
        if (alive) setLoading(false)
      }
    })()
    return () => { alive = false }
  }, [tab])

  const dismiss = async (id: string) => {
    setBusyId(id)
    try {
      await api.dismissDuplicateCandidate(id)
      setRows((prev) => prev.filter((r) => r.id !== id))
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر الحفظ')
    } finally {
      setBusyId(null)
    }
  }

  return (
    <div dir="rtl" className="space-y-5">
      <div>
        <h2 className="text-2xl font-bold text-brand-900">🔍 تدقيق التكرار</h2>
        <p className="mt-1 text-sm text-slate-500">
          النظام يفحص دورياً ويرشّح أزواجاً مشتبه بيها — حجزين لنفس الزبون بنفس العنوان بفارق يوم، أو زبونين بنفس
          رقم الهاتف بأسماء مختلفة. <b>القرار والتصرّف يبقى بيدك</b> — الزر هنا يأشّر «مو تكرار» بس.
        </p>
      </div>

      <div className="flex gap-2">
        {(['BOOKING', 'CUSTOMER'] as const).map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`rounded-xl px-4 py-2 text-sm font-bold transition ${
              tab === t ? 'bg-brand-700 text-white' : 'bg-white text-slate-600 shadow-sm'
            }`}
          >
            {t === 'BOOKING' ? '📋 حجوزات مكررة' : '👤 زبائن مكررون'}
          </button>
        ))}
      </div>

      {err && <p className="rounded-lg bg-red-50 p-4 text-red-600">{err}</p>}
      {loading && <p className="text-slate-400">جاري التحميل...</p>}
      {!loading && rows.length === 0 && (
        <p className="rounded-xl border border-white bg-white p-8 text-center text-slate-400">
          ماكو أزواج مشتبه بيها معلّقة حالياً بهذا التبويب.
        </p>
      )}

      <div className="space-y-3">
        {rows.map((r) => (
          <div key={r.id} className="rounded-2xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <p className="text-sm font-bold text-[#0f2040]">{r.matchReason}</p>
              <span className="text-[11px] text-slate-400">{new Date(r.detectedAt).toLocaleString('en-GB')}</span>
            </div>

            <div className="mt-3 space-y-2">
              {r.kind === 'BOOKING' ? (
                <>
                  <EntityIdentity
                    variant="full"
                    fields={{
                      bookingCode: r.bookingA?.code,
                      customerCode: r.bookingA ? `CUST-${String(r.bookingA.customerCode).padStart(5, '0')}` : undefined,
                      customerName: r.bookingA?.customerName,
                      customerPhone: r.bookingA?.customerPhone,
                      address: r.bookingA?.address || undefined,
                      serviceName: r.bookingA?.serviceName || undefined,
                      scheduledAt: r.bookingA?.scheduledAt,
                    }}
                  />
                  <EntityIdentity
                    variant="full"
                    fields={{
                      bookingCode: r.bookingB?.code,
                      customerCode: r.bookingB ? `CUST-${String(r.bookingB.customerCode).padStart(5, '0')}` : undefined,
                      customerName: r.bookingB?.customerName,
                      customerPhone: r.bookingB?.customerPhone,
                      address: r.bookingB?.address || undefined,
                      serviceName: r.bookingB?.serviceName || undefined,
                      scheduledAt: r.bookingB?.scheduledAt,
                    }}
                  />
                </>
              ) : (
                <>
                  <EntityIdentity
                    variant="full"
                    fields={{
                      customerCode: r.customerA ? `CUST-${String(r.customerA.customerCode).padStart(5, '0')}` : undefined,
                      customerName: r.customerA?.name,
                      customerPhone: r.customerA?.phone,
                    }}
                  />
                  <EntityIdentity
                    variant="full"
                    fields={{
                      customerCode: r.customerB ? `CUST-${String(r.customerB.customerCode).padStart(5, '0')}` : undefined,
                      customerName: r.customerB?.name,
                      customerPhone: r.customerB?.phone,
                    }}
                  />
                </>
              )}
            </div>

            <div className="mt-3 flex justify-end">
              <button
                onClick={() => dismiss(r.id)}
                disabled={busyId === r.id}
                className="rounded-lg bg-slate-100 px-3 py-1.5 text-xs font-bold text-slate-700 hover:bg-slate-200 disabled:opacity-50"
              >
                {busyId === r.id ? 'جاري الحفظ...' : '🔗 مو تكرار'}
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
