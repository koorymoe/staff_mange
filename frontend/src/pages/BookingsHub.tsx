import { useState } from 'react'
import { useSession } from '../session'
import BookingsList from './BookingsList'
import Coordinator from './Coordinator'
import SalesBooking from './SalesBooking'
import StageBucketsPage from './StageBucketsPage'

// ═══ «الحجوزات» — ثلاث شاشات بمدخل واحد ═══
//
// «راح ندمج الحجوزات وحجز جديد وتنسيق الحجوزات بشاشة وحدة… من ندخل
// عليها يطلعلنا خيار أول الي هو حجوزات اليوم، وبصفها خيار تنسيق
// الحجوزات، وبصفها خيار حجز جديد».
//
// الثلاثة نفس الشغلة بثلاث لحظات: تسجّل الحجز، تنسّقه، وتتابعه.
// ولما كانوا ثلاث بنود بالقائمة، الإداري يسجّل حجز ويطلع للقائمة
// يدوّر على «تنسيق» حتى يكمّله — وأحياناً ينسى، فالحجز يبقى مسجّل
// بلا تنسيق.
//
// ⚠️ كل تبويب يعيد استعمال شاشته الأصلية كما هي — ما ننسخ منطق
// الحجوزات ولا التنسيق بمكان ثاني، وإلا صارت نسختين تفترقن أول
// تعديل.

// الترتيب مقصود: من أول ما يوصل الحجز لآخر مرحلة يمر بيها الإداري.
// تسجّل → تنسّق → تتابع الي ما انثبّت → المثبّت → المكلّف.
const TABS = [
  { key: 'new' as const, label: 'حجز جديد', icon: '＋' },
  { key: 'coord' as const, label: 'تنسيق الحجوزات', icon: '🧩' },
  { key: 'pending' as const, label: 'بانتظار التثبيت', icon: '⏳' },
  { key: 'confirmed' as const, label: 'حجوزات مثبّتة', icon: '✅' },
  { key: 'assigned' as const, label: 'حجوزات مكلّفة', icon: '👥' },
  // ═══ الي ما توصل للتنفيذ ═══
  // «اكو حجوزات ما توصل — الزبون يلغي أو ما يرد، لازم تترتب».
  // كانت شاشة مستقلة بالقائمة الجانبية وانشالت من هناك؛ محلها
  // المنطقي هنا: هي حجوزات، وآخر محطة برحلتهن.
  { key: 'stuck' as const, label: 'ما وصلت للتنفيذ', icon: '🚫' },
]

type TabKey = (typeof TABS)[number]['key']

export default function BookingsHub() {
  const { employee, permissions } = useSession()
  const perms = permissions ?? []
  const isAdmin = employee?.role === 'ADMIN' || employee?.role === 'OWNER'

  // ⚠️ التبويب ما يطلع إلا لمن عنده صلاحيته — تبويب يفتح شاشة تگله
  // «ممنوع» أسوأ من تبويب ما موجود.
  const canCoord = isAdmin || perms.includes('coordinator')
  const canCreate = isAdmin || perms.includes('sales_booking')

  // نبدي بـ«بانتظار التثبيت» لأنها الي تحتاج تصرّف: «حجز جديد»
  // شاشة إدخال، وفتحها افتراضياً يعني الإداري يفتح النظام ويلگه
  // نموذج فاضي بدل شغله الي ينتظره.
  const [tab, setTab] = useState<TabKey>('pending')
  const shown = TABS.filter((t) =>
    (t.key === 'new' && canCreate)
    || (t.key === 'coord' && canCoord)
    // «ما وصلت للتنفيذ» شغل تنسيق: منو يتابع الزبون الي ما رد
    || (t.key === 'stuck' && canCoord)
    || (t.key !== 'new' && t.key !== 'coord' && t.key !== 'stuck'),
  )

  return (
    <div dir="rtl">
      {/* ═══ الخيارات ═══ */}
      {shown.length > 1 && (
        <div className="mb-4 grid grid-cols-2 gap-1.5 sm:grid-cols-3 rounded-2xl border border-slate-200 bg-white p-1.5 shadow-[0_2px_12px_rgba(15,32,64,0.05)] sm:inline-flex sm:gap-2">
          {shown.map((t) => (
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
      )}

      {/* ⚠️ نبني الشاشة المختارة بس (مو نخفي الباقي بـCSS): الثلاثة
          تجيب بيانات من السيرفر، وبناؤهن كلهن يعني ثلاثة أضعاف
          النداءات بكل فتحة وشاشة ثقيلة على الموبايل. */}
      {tab === 'new' && canCreate && <SalesBooking />}
      {tab === 'coord' && canCoord && <Coordinator />}
      {/* كل سلّة تفتح نفس القائمة بفلاترها (بحث · يوم · شهر) —
          والسلّة تقرر شنو يطلع. `key` تجبر React يبني القائمة من
          جديد عند التبديل، وإلا تضل فلاتر السلّة السابقة شغّالة. */}
      {tab === 'pending' && <BookingsList key="pending" bucket="pending" />}
      {tab === 'confirmed' && <BookingsList key="confirmed" bucket="confirmed" />}
      {tab === 'assigned' && <BookingsList key="assigned" bucket="assigned" />}
      {tab === 'stuck' && <StageBucketsPage />}
    </div>
  )
}
