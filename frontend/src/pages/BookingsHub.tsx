import { useState } from 'react'
import { useSession } from '../session'
import BookingsList from './BookingsList'
import Coordinator from './Coordinator'
import SalesBooking from './SalesBooking'
import PartialBookings from './PartialBookings'
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

// ═══ الخيارات = مراحل الحجز بالترتيب ═══
//
// «أني أريد الموضوع يكون مرحلة مرحلة: بانتظار التثبيت، بعدها يتحوّل
// (تم التثبيت — بانتظار التنسيق)، بعدها يطلعلي مكلّف… على هل أساس
// امشيلي».
//
// فالصف الأعلى صار **طريق** مو مجموعة شاشات: كل خيار محطة، والحجز
// يمشي من وحدة للي بعدها ولا يوقف بثنتين. والرقم قدّام كل محطة
// يخلّي الطريق مقروء بنظرة — الإداري يعرف وين واقف الشغل بلا ما
// يحفظ الترتيب.
//
// ⚠️ «تنسيق الحجوزات» محطة **شغل** مو حالة: هي المكان الي تنسّق بيه
// الحجز الي انثبّت. محلها بين «بانتظار التنسيق» و«مكلّف» لأن هذا
// بالضبط وين تصير بالواقع.
const TABS = [
  // ١ — الإدخال
  { key: 'new' as const, label: 'حجز جديد', icon: '＋' },
  // ٢ — شاشة الشغل: المنسّق يفتحها أول ما يدخل، فمحلها بالمقدمة
  { key: 'coord' as const, label: 'تنسيق الحجوزات', icon: '🧩' },
  // ٣ — انسجّل وما أحد حچى وية زبونه بعد
  { key: 'pending' as const, label: 'بانتظار التثبيت — بحاجة لتنسيق', icon: '⏳' },
  // ٤ — انثبّت وانرحّل، وينتظر كادراً
  { key: 'confirmed' as const, label: 'تم التثبيت — بحاجة لكادر', icon: '✅' },
  // ٥ — انكلّف عليه كادر، وينتظر يوم التنفيذ
  { key: 'assigned' as const, label: 'مكلّف — بانتظار التنفيذ', icon: '👥' },
  // ٦ — الاتجاه الأول بعد التكليف
  // «بعد ما يوصل الحجز مرحلة التكليف راح ياخذ اتجاهين: الاتجاه الأول
  // الي هو تم الإنجاز… وراها بنود نتفرّع».
  // التفرّعات (كامل · بلا فاتورة · بلا تقرير · بلا الاثنين) تنفتح
  // جوّا الشاشة مو كخيارات بالصف الأعلى.
  { key: 'done' as const, label: 'تم الإنجاز', icon: '🏁' },
  // ٧ — طلع الكادر وما خلّص: يحتاج موعد إكمال
  { key: 'partial' as const, label: 'تحتاج إكمال', icon: '🔄' },
  // ٨ — والاتجاه الثاني: الي ما وصل
  // «اكو حجوزات ما توصل — الزبون يلغي أو ما يرد، لازم تترتب».
  // محبوس عند إدارة المشاريع — يرجع لحاله أول ما يوصل التنفيذ
  { key: 'projects' as const, label: 'عند إدارة المشاريع', icon: '🏗️' },
  { key: 'stuck' as const, label: 'ما وصلت للتنفيذ', icon: '🚫' },
  // ٩ — مخرج ثاني: انطلب حذفه وينتظر قرار المراقب
  // «الحجوزات الي ينحذفن أريدهن يترحّلن بعد، ما أريد يضلن بمكان واحد».
  { key: 'deleting' as const, label: 'بانتظار قرار الحذف', icon: '🗑️' },
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
    // «تحتاج إكمال» شغل تنسيق: منو يحدد موعد الإكمال
    || (t.key === 'partial' && canCoord)
    // متابعة طلبات الحذف شغل تنسيق/إشراف
    || (t.key === 'deleting' && canCoord)
    // متابعة المحبوس عند المشاريع شغل تنسيق
    || (t.key === 'projects' && canCoord)
    || (t.key !== 'new' && t.key !== 'coord' && t.key !== 'stuck' && t.key !== 'partial' && t.key !== 'deleting' && t.key !== 'projects'),
  )

  return (
    <div dir="rtl">
      {/* ═══ الخيارات ═══ */}
      {shown.length > 1 && (
        // ═══ الشريط يبقى بمكانه ═══
        // «من أنزل لجوّه أشوف الحجوزات البعيدة ما أريدهن يصعدن لفوگ،
        // أريدهن يضلن ثابتات».
        // ⚠️ `sticky` مو `fixed`: يبقى جوّا تدفّق الصفحة فما يغطي
        // المحتوى ولا يحتاج حساب ارتفاع يدوي. وz-30 يخلّيه فوگ صفوف
        // الجدول ولاصق تحت رأس الصفحة.
        <div className="sticky top-0 z-30 mb-4 grid grid-cols-2 gap-1.5 rounded-2xl border border-slate-200 bg-white/95 p-1.5 shadow-[0_2px_12px_rgba(15,32,64,0.08)] backdrop-blur sm:inline-flex sm:gap-2">
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
      {tab === 'partial' && <PartialBookings />}
      {tab === 'done' && <BookingsList key="done" bucket="done" />}
      {tab === 'projects' && <BookingsList key="projects" bucket="at_projects" />}
      {tab === 'deleting' && <BookingsList key="deleting" bucket="delete_pending" />}
      {tab === 'stuck' && <StageBucketsPage />}
    </div>
  )
}
