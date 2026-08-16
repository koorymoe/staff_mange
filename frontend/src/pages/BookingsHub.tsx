import { useState } from 'react'
import { useSession } from '../session'
import BookingsList from './BookingsList'
import Coordinator from './Coordinator'
import SalesBooking from './SalesBooking'

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

const TABS = [
  { key: 'today' as const, label: 'حجوزات اليوم', icon: '📋' },
  { key: 'coord' as const, label: 'تنسيق الحجوزات', icon: '🧩' },
  { key: 'new' as const, label: 'حجز جديد', icon: '＋' },
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

  const [tab, setTab] = useState<TabKey>('today')
  const shown = TABS.filter((t) =>
    t.key === 'today' || (t.key === 'coord' && canCoord) || (t.key === 'new' && canCreate),
  )

  return (
    <div dir="rtl">
      {/* ═══ الخيارات ═══ */}
      {shown.length > 1 && (
        <div className="mb-4 grid grid-cols-3 gap-1.5 rounded-2xl border border-slate-200 bg-white p-1.5 shadow-[0_2px_12px_rgba(15,32,64,0.05)] sm:inline-flex sm:gap-2">
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
      {tab === 'today' && <BookingsList />}
      {tab === 'coord' && canCoord && <Coordinator />}
      {tab === 'new' && canCreate && <SalesBooking />}
    </div>
  )
}
