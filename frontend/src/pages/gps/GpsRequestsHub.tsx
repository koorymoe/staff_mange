import { useState } from 'react'
import PageHeader from '../../components/PageHeader'
import GpsRequestsReview from './GpsRequestsReview'
import GpsRenewalsReview from './GpsRenewalsReview'
import GpsMaintenanceReview from './GpsMaintenanceReview'

// ═══ طلبات GPS — ثلاث شاشات بمدخل واحد ═══
//
// «المفروض ما يكونن هيج، المفروض يتوزعن ع الواجهات: طلبات الصيانة
// وطلبات جديد وطلبات معلقة يكونن بواجهة وحدهن، والشرائح والمتابعة
// ونظام الجي بي اس يكونن بواجهة وحدهن».
//
// الثلاثة **نفس الشغلة**: طلب يوصل من المبيعات ومسؤول الجي بي اس
// يبتّ بيه. ولمن كانوا ثلاث بنود بالقائمة، المسؤول يفتح بنداً ويشتغل
// عليه وينسى الثاني — لأن ماكو مكان واحد يگله وين متكدّس الشغل.
//
// ⚠️ كل تبويب يعيد استعمال شاشته الأصلية كما هي بـ`embedded` — ما
// ننسخ منطق البتّ بمكان ثاني، وإلا صارت نسختان تفترقن أول تعديل.
// (نفس نمط `BookingsHub` بالضبط.)
const TABS = [
  { key: 'pending' as const, label: 'طلبات معلقة', icon: '📋' },
  { key: 'renewals' as const, label: 'طلبات تجديد', icon: '🔄' },
  { key: 'maintenance' as const, label: 'طلبات صيانة', icon: '🔧' },
]

export default function GpsRequestsHub() {
  const [tab, setTab] = useState<(typeof TABS)[number]['key']>('pending')

  return (
    <div dir="rtl">
      <div className="mb-6">
        <PageHeader title="📥 طلبات GPS" subtitle="الطلبات الواصلة من المبيعات — تنبتّ من هنا" />
      </div>

      <div className="sticky top-0 z-30 mb-4 grid grid-cols-3 gap-1.5 rounded-2xl border border-slate-200 bg-white/95 p-1.5 shadow-[0_2px_12px_rgba(15,32,64,0.08)] backdrop-blur sm:inline-flex sm:gap-2">
        {TABS.map((t) => (
          <button
            key={t.key}
            onClick={() => setTab(t.key)}
            className={`rounded-xl px-3 py-2 text-[11px] font-extrabold transition sm:px-5 sm:text-xs ${
              tab === t.key
                ? 'bg-gradient-to-l from-brand-500 to-brand-800 text-white shadow-md'
                : 'text-slate-600 hover:bg-slate-50'
            }`}
          >
            {t.icon} {t.label}
          </button>
        ))}
      </div>

      {/* ⚠️ نبني المختارة بس (مو نخفي الباقي بـCSS): الثلاثة يجبن
          بيانات من السيرفر، وبناؤهن كلهن ثلاثة أضعاف النداءات. */}
      {tab === 'pending' && <GpsRequestsReview embedded />}
      {tab === 'renewals' && <GpsRenewalsReview embedded />}
      {tab === 'maintenance' && <GpsMaintenanceReview embedded />}
    </div>
  )
}
