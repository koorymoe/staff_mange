import { useEffect, useState } from 'react'
import { api, type DuplicateCandidate } from '../api'
import EntityIdentity from '../components/EntityIdentity'

// ═══ تدقيق التكرار (ماتركس) — حجوزات وزبائن مكررون بالغلط ═══
//
// ⚠️ هذا مسار مستقل عن صندوق المراقب (MonitorReview) بالتصميم —
// الأخير مبني على «معرّف واحد ← هوية واحدة»، وزوج التكرار علاقة بين
// سجلين مو هوية مفردة. شوف الخطة (دفعة ماتركس الكبيرة) للتفصيل.
//
// ولا حذف أو دمج تلقائي — الفحص يكتشف بس، والموظف (صلاحية duplicate_review)
// يقرر: «مو تكرار»، أو يطلب حذف الحجز المكرر (ينتظر موافقة المراقب/المدير)،
// أو يدمج الزبونين (كل شي ينتقل للأصلي بمعاملة وحدة، والمالك يوصله إشعار).

type Tab = 'BOOKING' | 'CUSTOMER' | 'DONE'
const custCode = (n?: number) => (n != null ? `CUST-${String(n).padStart(5, '0')}` : undefined)

export default function DuplicateReviewPage() {
  const [tab, setTab] = useState<Tab>('BOOKING')
  const [rows, setRows] = useState<DuplicateCandidate[]>([])
  const [loading, setLoading] = useState(true)
  const [err, setErr] = useState<string | null>(null)
  const [busyId, setBusyId] = useState<string | null>(null)
  const [msg, setMsg] = useState<string | null>(null)
  const [merge, setMerge] = useState<{ row: DuplicateCandidate; keepId: string; moves: { label: string; count: number }[] | null } | null>(null)

  useEffect(() => {
    let alive = true
    queueMicrotask(() => { if (alive) setLoading(true) })
    void (async () => {
      try {
        const data = tab === 'DONE'
          ? [...await api.getDuplicateCandidates(undefined, 'RESOLVED'), ...await api.getDuplicateCandidates(undefined, 'DISMISSED')]
              .sort((a, b) => (b.reviewedAt ?? '').localeCompare(a.reviewedAt ?? ''))
          : await api.getDuplicateCandidates(tab, 'PENDING')
        if (alive) setRows(data)
      } catch (e) {
        if (alive) setErr(e instanceof Error ? e.message : 'تعذر الجلب')
      } finally {
        if (alive) setLoading(false)
      }
    })()
    return () => { alive = false }
  }, [tab])

  const run = async (id: string, f: () => Promise<string>) => {
    setBusyId(id); setErr(null)
    try {
      setMsg(await f())
      setRows((prev) => prev.filter((r) => r.id !== id))
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر الحفظ')
    } finally {
      setBusyId(null)
    }
  }
  const dismiss = (id: string) => run(id, async () => { await api.dismissDuplicateCandidate(id); return '✓ انأشّر «مو تكرار».' })
  const requestDelete = (r: DuplicateCandidate, bookingId: string, code?: string) => {
    if (!confirm(`تطلب حذف الحجز ${code ?? ''} لأنه مكرر؟\nالطلب يروح للمراقب أو المدير يوافق عليه — ما ينحذف هسه.`)) return
    void run(r.id, async () => { const x = await api.requestDuplicateBookingDelete(r.id, bookingId); return `✓ انطلب حذف ${x.bookingCode} — ينتظر موافقة المراقب أو المدير.` })
  }
  const openMerge = async (row: DuplicateCandidate, keepId: string) => {
    setMerge({ row, keepId, moves: null })
    try {
      const p = await api.duplicateMergePreview(row.id, keepId)
      setMerge({ row, keepId, moves: p.moves })
    } catch (e) {
      setMerge(null); setErr(e instanceof Error ? e.message : 'تعذر الحساب')
    }
  }
  const doMerge = () => {
    if (!merge) return
    const { row, keepId } = merge
    setMerge(null)
    void run(row.id, async () => (await api.mergeDuplicateCustomers(row.id, keepId)).note)
  }

  const actBtn = 'rounded-lg px-3 py-1.5 text-xs font-bold disabled:opacity-50'
  return (
    <div dir="rtl" className="space-y-5">
      <div>
        <h2 className="text-2xl font-bold text-brand-900">🔍 تدقيق التكرار</h2>
        <p className="mt-1 text-sm text-slate-500">
          النظام يفحص دورياً ويرشّح أزواجاً مشتبه بيها — حجزين لنفس الزبون بنفس العنوان بفارق يوم، أو زبونين بنفس
          رقم الهاتف بأسماء مختلفة. <b>القرار بيدك:</b> «مو تكرار»، أو اطلب حذف الحجز المكرر (ينتظر موافقة المراقب أو المدير)،
          أو ادمج الزبونين بالأصلي.
        </p>
      </div>

      <div className="flex flex-wrap gap-2">
        {(['BOOKING', 'CUSTOMER', 'DONE'] as const).map((t) => (
          <button
            key={t}
            onClick={() => { setTab(t); setMsg(null) }}
            className={`rounded-xl px-4 py-2 text-sm font-bold transition ${
              tab === t ? 'bg-brand-700 text-white' : 'bg-white text-slate-600 shadow-sm'
            }`}
          >
            {t === 'BOOKING' ? '📋 حجوزات مكررة' : t === 'CUSTOMER' ? '👤 زبائن مكررون' : '✅ المحلولة'}
          </button>
        ))}
      </div>

      {msg && <p className="rounded-lg bg-emerald-50 p-3 text-sm font-bold text-emerald-800">{msg}</p>}
      {err && <p className="rounded-lg bg-red-50 p-4 text-red-600">{err}</p>}
      {loading && <p className="text-slate-400">جاري التحميل...</p>}
      {!loading && rows.length === 0 && (
        <p className="rounded-xl border border-white bg-white p-8 text-center text-slate-400">
          {tab === 'DONE' ? 'بعد ما انحل شي.' : 'ماكو أزواج مشتبه بيها معلّقة حالياً بهذا التبويب.'}
        </p>
      )}

      <div className="space-y-3">
        {rows.map((r) => (
          <div key={r.id} className="rounded-2xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <p className="text-sm font-bold text-[#0f2040]">{r.kind === 'BOOKING' ? '📋 ' : '👤 '}{r.matchReason}</p>
              <span className="text-[11px] text-slate-400">{new Date(r.detectedAt).toLocaleString('en-GB')}</span>
            </div>

            {tab === 'DONE' ? (
              <div className="mt-2 rounded-lg bg-slate-50 p-3 text-sm">
                <b className={r.resolution === 'MERGED' ? 'text-sky-700' : r.resolution === 'DELETE_REQUESTED' ? 'text-red-700' : 'text-slate-600'}>
                  {r.resolution === 'MERGED' ? '🔀 اندمجوا' : r.resolution === 'DELETE_REQUESTED' ? '🗑️ انطلب حذف المكرر' : '🔗 مو تكرار'}
                </b>
                {r.resolutionNote && <p className="mt-1 text-slate-700">{r.resolutionNote}</p>}
                <p className="mt-1 text-[11px] text-slate-400">{r.reviewedByName ? `${r.reviewedByName} · ` : ''}{r.reviewedAt ? new Date(r.reviewedAt).toLocaleString('en-GB') : ''}</p>
              </div>
            ) : (
              <div className="mt-3 grid gap-2 md:grid-cols-2">
                {r.kind === 'BOOKING'
                  ? [r.bookingA, r.bookingB].map((b, i) => b && (
                    <div key={i} className="space-y-2">
                      <EntityIdentity
                        variant="full"
                        fields={{
                          bookingCode: b.code, customerCode: custCode(b.customerCode), customerName: b.customerName,
                          customerPhone: b.customerPhone, address: b.address || undefined, serviceName: b.serviceName || undefined, scheduledAt: b.scheduledAt,
                        }}
                      />
                      <button onClick={() => requestDelete(r, b.id, b.code)} disabled={busyId === r.id}
                        className={`${actBtn} w-full bg-red-50 text-red-700 hover:bg-red-100`}>🗑️ {b.code} هو المكرر — اطلب حذفه</button>
                    </div>
                  ))
                  : [r.customerA, r.customerB].map((c, i) => c && (
                    <div key={i} className="space-y-2">
                      <EntityIdentity variant="full" fields={{ customerCode: custCode(c.customerCode), customerName: c.name, customerPhone: c.phone }} />
                      <button onClick={() => void openMerge(r, c.id)} disabled={busyId === r.id}
                        className={`${actBtn} w-full bg-sky-50 text-sky-800 hover:bg-sky-100`}>🔀 خلّي {custCode(c.customerCode)} الأصلي وادمج الثاني بيه</button>
                    </div>
                  ))}
              </div>
            )}

            {tab !== 'DONE' && (
              <div className="mt-3 flex items-center justify-between gap-2">
                <p className="text-[11px] text-slate-400">{r.kind === 'BOOKING' ? 'إذا شغلتين مختلفتين لنفس الزبون:' : 'إذا شخصين مختلفين:'}</p>
                <button onClick={() => void dismiss(r.id)} disabled={busyId === r.id}
                  className={`${actBtn} bg-slate-100 text-slate-700 hover:bg-slate-200`}>
                  {busyId === r.id ? 'جاري الحفظ...' : '🔗 مو تكرار'}
                </button>
              </div>
            )}
          </div>
        ))}
      </div>

      {merge && (
        <div className="fixed inset-0 z-50 grid place-items-center bg-black/50 p-4" onClick={() => setMerge(null)}>
          <div dir="rtl" className="w-full max-w-md rounded-2xl bg-white p-5" onClick={(e) => e.stopPropagation()}>
            {(() => {
              const keep = merge.row.customerA?.id === merge.keepId ? merge.row.customerA : merge.row.customerB
              const drop = keep === merge.row.customerA ? merge.row.customerB : merge.row.customerA
              return (
                <>
                  <h3 className="mb-2 text-lg font-extrabold text-[#0f2040]">🔀 دمج الزبونين</h3>
                  <p className="text-sm">الأصلي يبقى: <b>{custCode(keep?.customerCode)} {keep?.name}</b> ({keep?.phone})</p>
                  <p className="text-sm">ينشال: <b className="text-red-700">{custCode(drop?.customerCode)} {drop?.name}</b> ({drop?.phone})</p>
                  <div className="mt-3 rounded-lg bg-slate-50 p-3 text-sm">
                    <b>ينتقل للأصلي:</b>
                    {merge.moves == null ? <p className="text-slate-400">يحسب…</p>
                      : merge.moves.length === 0 ? <p className="text-slate-500">ماكو شي مربوط بالمكرر.</p>
                      : <ul className="mt-1 list-inside list-disc">{merge.moves.map((m) => <li key={m.label}>{m.count} {m.label}</li>)}</ul>}
                  </div>
                  <p className="mt-2 text-[11px] text-slate-500">الموقع والخريطة: الأصلي ياخذها من المكرر بس إذا هو فاضي. المالك يوصله إشعار بالدمج، ويبقى أثره بتبويب «المحلولة».</p>
                  <div className="mt-4 flex gap-2">
                    <button disabled={merge.moves == null} onClick={doMerge} className="flex-1 rounded-lg bg-sky-700 py-2 font-bold text-white disabled:opacity-50">تأكيد الدمج</button>
                    <button onClick={() => setMerge(null)} className="rounded-lg border px-4 py-2">إلغاء</button>
                  </div>
                </>
              )
            })()}
          </div>
        </div>
      )}
    </div>
  )
}
