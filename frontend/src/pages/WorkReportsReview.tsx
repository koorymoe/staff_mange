import { useEffect, useState } from 'react'
import { api, type WorkReport } from '../api'
import { formatCustomerCode } from '../utils/identity'
import EntityIdentity from '../components/EntityIdentity'
import { ExecutionSheetView } from '../components/ExecutionSheet'
import WorkWindowLine from '../components/WorkWindowLine'
import { asSheet } from '../utils/executionSheet'
import { useSession } from '../session'
import BookingCodeChip from '../components/BookingCodeChip'
import { loadFailed } from '../netErrors'

export default function WorkReportsReview() {
  const [reports, setReports] = useState<WorkReport[]>([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState<'ALL' | 'COMPLETED' | 'STOPPED'>('ALL')
  // (ع) 10-16: «تم التدقيق» حتى ما تبقى التقارير معلّكة — الافتراضي الي تنتظر
  const [review, setReview] = useState<'PENDING' | 'DONE' | 'ALL'>('PENDING')
  const { employee, permissions } = useSession()
  const canReview = employee?.role === 'ADMIN' || employee?.role === 'MONITOR'
    || (employee?.role !== 'QUALITY_ENGINEER' && (permissions.includes('monitoring') || permissions.includes('auditing')))
  const [notes, setNotes] = useState<Record<string, string>>({})
  const markReviewed = async (id: string) => {
    try {
      await api.reviewWorkReport(id, notes[id]?.trim() || undefined)
      setReports((rs) => rs.map((x) => (x.id === id ? { ...x, reviewedAt: new Date().toISOString(), reviewNote: notes[id]?.trim() || null, reviewedByName: employee?.name ?? null } : x)))
    } catch (e) { alert(e instanceof Error ? e.message : 'تعذر') }
  }

  useEffect(() => {
    api.getWorkReports().then(setReports).finally(() => setLoading(false)).catch(loadFailed)
  }, [])

  const pendingCount = reports.filter((r) => !r.reviewedAt).length
  const filtered = (filter === 'ALL' ? reports : reports.filter((r) => r.workStatus === filter))
    .filter((r) => review === 'ALL' || (review === 'PENDING' ? !r.reviewedAt : !!r.reviewedAt))

  return (
    <div dir="rtl" className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-brand-900">تقارير العمل — مراجعة</h2>
        <p className="mt-1 text-slate-500">كل التقارير اللي يرسلها الفنيون عن الحجوزات اللي اشتغلوا عليها</p>
      </div>

      <div className="flex flex-wrap gap-2">
        {([
          { key: 'PENDING', label: `⏳ تنتظر التدقيق (${pendingCount})` },
          { key: 'DONE', label: '✅ متدقّقة' },
          { key: 'ALL', label: 'كل التقارير' },
        ] as const).map((t) => (
          <button key={t.key} onClick={() => setReview(t.key)}
            className={`rounded-lg px-4 py-2 text-sm font-bold ${review === t.key ? 'bg-[#0f2040] text-white' : 'border border-slate-200 bg-white text-slate-600'}`}>
            {t.label}
          </button>
        ))}
      </div>
      <div className="flex gap-2">
        {([
          { key: 'ALL', label: 'الكل' },
          { key: 'COMPLETED', label: 'مكتملة' },
          { key: 'STOPPED', label: 'متوقفة' },
        ] as const).map((t) => (
          <button
            key={t.key}
            onClick={() => setFilter(t.key)}
            className={`rounded-lg px-4 py-2 text-sm font-bold transition-colors ${filter === t.key ? 'bg-brand-600 text-white' : 'bg-white text-slate-600 border border-slate-200 hover:bg-slate-50'}`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {loading && <p className="text-slate-400">جاري التحميل...</p>}

      {/* (ع) 10-16: «اريد فاصلة توضحلي بين تقرير وتقرير» — كل تقرير بإطار ملوّن
          وشريط فوكه برقمه، وفراغ واضح بين التقارير */}
      <div className="space-y-8">
        {filtered.map((r, idx) => (
          <div key={r.id} className="overflow-hidden rounded-2xl border-2 border-[#2c5aad] bg-white shadow-md">
            <div className="flex items-center justify-between bg-[#0f2040] px-5 py-2 text-sm font-extrabold text-white">
              <span>📄 تقرير {idx + 1} من {filtered.length}</span>
              <span className="font-mono">{r.booking?.code}</span>
            </div>
            <div className="p-5">
              {r.bookingId && <div className="mb-3"><WorkWindowLine bookingId={r.bookingId} /></div>}
              {r.reviewedAt ? (
                <p className="mb-3 rounded-lg bg-emerald-50 px-3 py-1.5 text-xs font-bold text-emerald-800">
                  ✅ تم التدقيق{r.reviewedByName ? ` — ${r.reviewedByName}` : ''} · {new Date(r.reviewedAt).toLocaleString('ar-IQ', { timeZone: 'Asia/Baghdad' })}{r.reviewNote ? ` · ${r.reviewNote}` : ''}
                </p>
              ) : canReview && (
                <div className="mb-3 flex flex-wrap gap-2 rounded-lg border border-dashed border-emerald-300 bg-emerald-50/50 p-2">
                  <input value={notes[r.id] ?? ''} onChange={(e) => setNotes({ ...notes, [r.id]: e.target.value })} placeholder="ملاحظة التدقيق (اختياري)"
                    className="min-w-0 flex-1 rounded-lg border border-slate-300 bg-white px-2 py-1 text-sm" />
                  <button type="button" onClick={() => markReviewed(r.id)} className="rounded-lg bg-emerald-600 px-4 py-1 text-sm font-bold text-white">✅ تم التدقيق</button>
                </div>
              )}
            <div className="flex items-center justify-between">
              <div>
                <span className="font-mono font-bold text-brand-700"><BookingCodeChip code={r.booking?.code} /></span>
                <span className="mr-2 text-sm text-slate-500">{r.booking?.customerName}</span>
                {/* رقم هاتف الزبون — حتى مهندس الجودة يتصل مباشرة من التقرير */}
                {r.booking?.customerPhone && (
                  <a href={`tel:${r.booking.customerPhone}`} dir="ltr"
                    className="mr-2 inline-block rounded-lg bg-emerald-50 px-2.5 py-1 text-sm font-medium text-emerald-700 hover:bg-emerald-100">
                    📞 {r.booking.customerPhone}
                  </a>
                )}
              </div>
              <span className={`rounded-full px-2 py-1 text-xs font-bold ${r.workStatus === 'COMPLETED' ? 'bg-emerald-100 text-emerald-700' : 'bg-amber-100 text-amber-700'}`}>
                {r.workStatus === 'COMPLETED' ? 'تم الإنجاز' : 'توقف العمل'}
              </span>
            </div>
            {/* هوية كاملة: التقرير كان يقول كود الحجز والزبون بس */}
            <EntityIdentity
              fields={{
                bookingCode: r.booking?.code,
                customerCode: formatCustomerCode(r.booking),
                customerName: r.booking?.customerName,
                customerPhone: r.booking?.customerPhone || undefined,
                address: r.booking?.address || undefined,
                leaderName: r.booking?.leaderName || undefined,
              }}
              variant="full"
              className="mt-2"
            />
            <p className="mt-1 text-xs text-slate-400">
              الفني: {r.employee?.name || '-'} · {new Date(r.createdAt).toLocaleString('ar-IQ')}
            </p>

            {/* (ع 10-16) التقرير بنموذج المشرف الأصلي — نفس ترتيب ومدخلات الليدر */}
            {asSheet(r.sheet) ? (
              <div className="mt-3"><ExecutionSheetView sheet={asSheet(r.sheet)!} /></div>
            ) : r.workStatus === 'COMPLETED' ? (
              <div className="mt-3 space-y-2 text-sm">
                <p><span className="text-slate-500">الأحداث: </span>{r.events || '-'}</p>
                {r.extraRequests && <p><span className="text-slate-500">طلبات الزبون: </span>{r.extraRequests}</p>}
                <div className="flex gap-4 text-xs">
                  <span className={r.cleanedSite ? 'text-emerald-600' : 'text-red-500'}>{r.cleanedSite ? '✓' : '✗'} تنظيف المكان</span>
                  <span className={r.gaveInfo ? 'text-emerald-600' : 'text-red-500'}>{r.gaveInfo ? '✓' : '✗'} إعطاء المعلومات</span>
                  <span className={r.tookPhotos ? 'text-emerald-600' : 'text-red-500'}>{r.tookPhotos ? '✓' : '✗'} تصوير العمل</span>
                </div>
                {r.notes && <p><span className="text-slate-500">ملاحظات: </span>{r.notes}</p>}
              </div>
            ) : (
              <div className="mt-3 space-y-2 text-sm">
                <p><span className="text-slate-500">سبب التوقف: </span>{r.stopReason || '-'}</p>
                {r.notes && <p><span className="text-slate-500">ملاحظات: </span>{r.notes}</p>}
              </div>
            )}
            </div>
          </div>
        ))}
        {!loading && filtered.length === 0 && (
          <div className="rounded-xl border border-white bg-white p-10 text-center text-slate-400 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            لا توجد تقارير
          </div>
        )}
      </div>
    </div>
  )
}
