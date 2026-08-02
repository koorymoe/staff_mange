import { useEffect, useState } from 'react'
import { api, type WorkReport } from '../api'

export default function WorkReportsReview() {
  const [reports, setReports] = useState<WorkReport[]>([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState<'ALL' | 'COMPLETED' | 'STOPPED'>('ALL')

  useEffect(() => {
    api.getWorkReports().then(setReports).finally(() => setLoading(false))
  }, [])

  const filtered = filter === 'ALL' ? reports : reports.filter((r) => r.workStatus === filter)

  return (
    <div dir="rtl" className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-brand-900">تقارير العمل — مراجعة</h2>
        <p className="mt-1 text-slate-500">كل التقارير اللي يرسلها الفنيون عن الحجوزات اللي اشتغلوا عليها</p>
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

      <div className="space-y-3">
        {filtered.map((r) => (
          <div key={r.id} className="rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <div className="flex items-center justify-between">
              <div>
                <span className="font-mono font-bold text-brand-700">{r.booking?.code}</span>
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
            <p className="mt-1 text-xs text-slate-400">
              الفني: {r.employee?.name || '-'} · {new Date(r.createdAt).toLocaleString('ar-IQ')}
            </p>

            {r.workStatus === 'COMPLETED' ? (
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
