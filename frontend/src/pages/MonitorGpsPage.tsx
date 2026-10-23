import { useEffect, useState } from 'react'
import { api, type GpsMonitorSnapshot } from '../api'
import PageHeader from '../components/PageHeader'
import StatTile from '../components/StatTile'
import EmptyState from '../components/EmptyState'

/**
 * ═══ نتائج الجي بي اس — للمراقب ═══
 *
 * «ماريدها تضهر بهاي الطريقة للمراقب. أريد تضهر كنتائج: يشوف أسماء
 * الزبائن والمشاكل الي عنده والاشتراكات وشوكت تخلص… ما تخلي وحدة
 * هيج، ماريد أزيد ازدحامها».
 *
 * ⚠️ ست شاشات تشغيلية (موافقة على طلب، حرق شريحة، تجديد اشتراك)
 * انخفت عن المراقب بالقائمة الجانبية، وصارت **تبويب واحد** هنا
 * بثلاث نتائج. الصندوق ما انزاد عليه ولا محطة — قراره الصريح.
 *
 * ⚠️ **قراءة فقط**: المراقب يشوف ويحاسب، ما ينفّذ. نفس مبدأ «يشوف
 * الجودة كاملة بلا ما يتصل بالزبون».
 */
const fmtDate = (v?: string | null) =>
  v ? new Date(v).toLocaleDateString('en-CA') : '—'

export default function MonitorGpsPage({ embedded }: { embedded?: boolean } = {}) {
  // null = لسه ما انجاب · الكائن = انجاب.
  // ⚠️ التفريق مقصود: «جاري الجلب» غير «ماكو ولا شي».
  const [snap, setSnap] = useState<GpsMonitorSnapshot | null>(null)
  const [failed, setFailed] = useState(false)

  useEffect(() => {
    let alive = true
    api.getGpsMonitorSnapshot()
      .then((s) => { if (alive) setSnap(s) })
      .catch(() => { if (alive) setFailed(true) })
    return () => { alive = false }
  }, [])

  if (failed) {
    return <EmptyState icon="📡" title="تعذر جلب نتائج الجي بي اس" reason="جرّب تحدّث الصفحة" />
  }
  if (!snap) {
    return <div className="rounded-2xl bg-white p-8 text-center text-slate-400">جاري الجلب...</div>
  }

  return (
    <div dir="rtl" className="space-y-4">
      {!embedded && (
        <PageHeader
          title="📡 نتائج الجي بي اس"
          subtitle="اشتراكات قربت تنتهي · انتهت وما انجدّدت · مشاكل مفتوحة — عرض بلا تنفيذ"
        />
      )}

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
        <StatTile label={`قربت تنتهي (${snap.expiringWindowDays} يوم)`} icon="⏳" tone="warning"
          value={snap.expiring.length} />
        <StatTile label="انتهت وما انجدّدت" icon="⛔" tone="danger" value={snap.expired.length} />
        <StatTile label="مشاكل مفتوحة" icon="🛠️" tone="info" value={snap.problems.length} />
      </div>

      {/* ═══ قربت تنتهي ═══
          ⚠️ هذا القسم ما چان موجوداً بالنظام إطلاقاً: الاستعلام
          الموجود يجيب **المنتهية فعلاً** بس. المراقب يحتاج يشوفها
          قبل ما تنتهي، وإلا شغله بعد فوات الأوان. */}
      <Section
        title={`⏳ اشتراكات قربت تنتهي — خلال ${snap.expiringWindowDays} يوم`}
        rows={snap.expiring}
        empty="ماكو ولا اشتراك قرب ينتهي بهذي المدة"
        renderDays={(d) => <span className="text-amber-700">باقي {d} يوم</span>}
      />

      <Section
        title="⛔ انتهت وما انجدّدت"
        rows={snap.expired}
        empty="ماكو ولا اشتراك منتهي بلا تجديد"
        renderDays={(d) => <span className="text-red-700">فاتت {Math.abs(d)} يوم</span>}
      />

      <div className="rounded-2xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        <p className="mb-3 font-bold text-slate-800">🛠️ مشاكل مفتوحة</p>
        {snap.problems.length === 0 ? (
          <EmptyState icon="✓" title="ماكو مشاكل مفتوحة" reason="كل طلبات الصيانة منجزة" />
        ) : (
          <div className="space-y-2">
            {snap.problems.map((p) => (
              <div key={p.id} className="rounded-xl border border-slate-200 p-3">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <span className="font-bold text-slate-800">{p.customerName || 'زبون غير معروف'}</span>
                  <span className="flex items-center gap-2 text-xs">
                    <span className="rounded-full bg-slate-100 px-2 py-0.5 font-bold text-slate-600">
                      {p.status === 'PENDING' ? 'معلّق' : 'قيد المعالجة'}
                    </span>
                    {/* عمر المشكلة — المراقب يحاسب على التأخير، لا على وجودها */}
                    <span className={p.ageDays >= 7 ? 'font-bold text-red-700' : 'text-slate-500'}>
                      صار عليها {p.ageDays} يوم
                    </span>
                  </span>
                </div>
                {p.customerPhone && <p className="mt-0.5 text-xs text-slate-500">{p.customerPhone}</p>}
                <p className="mt-1 text-sm text-slate-700">{p.problemDescription || '—'}</p>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

function Section({
  title, rows, empty, renderDays,
}: {
  title: string
  rows: GpsMonitorSnapshot['expiring']
  empty: string
  renderDays: (days: number) => React.ReactNode
}) {
  return (
    <div className="rounded-2xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
      <p className="mb-3 font-bold text-slate-800">{title}</p>
      {rows.length === 0 ? (
        <EmptyState icon="✓" title={empty} />
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-slate-100 text-right text-sm">
            <thead className="bg-slate-50">
              <tr>
                <th className="px-3 py-2 font-semibold text-slate-600">الزبون</th>
                <th className="px-3 py-2 font-semibold text-slate-600">الهاتف</th>
                <th className="px-3 py-2 font-semibold text-slate-600">رقم الجهاز</th>
                <th className="px-3 py-2 font-semibold text-slate-600">ينتهي بتاريخ</th>
                <th className="px-3 py-2 font-semibold text-slate-600">المتبقي</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {rows.map((r) => (
                <tr key={r.deviceRequestId}>
                  <td className="px-3 py-2 font-medium text-slate-700">{r.customerName || '—'}</td>
                  <td className="px-3 py-2 text-slate-600">{r.customerPhone || '—'}</td>
                  <td className="px-3 py-2 font-mono text-xs text-slate-600">{r.gpsNumber || '—'}</td>
                  <td className="px-3 py-2 text-slate-600">{fmtDate(r.subscriptionEnd)}</td>
                  <td className="px-3 py-2 font-bold">{renderDays(r.daysLeft)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
