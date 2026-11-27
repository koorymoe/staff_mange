import { Suspense } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { useSession } from '../session'
import { isPathAllowed } from '../components/navTree'
import TabbedHub from '../components/TabbedHub'
import RemotePeriodCard from '../components/RemotePeriodCard'
import { card } from '../ui/classes'
import { lazyRetry } from '../utils/staleReload'

// ═══ (دازاي 11-26) «⚙️ الإعدادات»: كل إعداد يدوي يأثر على الفلوس بمكان واحد ═══
// «اريد اعدادات… كلهن بنفس المكان بس مسميات مختلفه حته اعرف كل اعداد وين يروح».
// كل تبويب شاشته القديمة نفسها بمساره القديم، فالصلاحيات ما تتغيّر.
const OwnerMoneySplitPage = lazyRetry(() => import('./OwnerMoneySplitPage'))
const OwnerCostsPage = lazyRetry(() => import('./OwnerCostsPage'))
const SolarSettingsPage = lazyRetry(() => import('./SolarSettingsPage'))
const ScoreWeightsPage = lazyRetry(() => import('./ScoreWeightsPage'))
const MoneyAuditPage = lazyRetry(() => import('./MoneyAuditPage'))
const ServiceInvoiceOptionsPage = lazyRetry(() => import('./ServiceInvoiceOptionsPage'))

type Item = { to: string; title: string; enter: string; affects: string; when: string }
const ITEMS: Item[] = [
  {
    to: '/owner-money-split',
    title: '💰 توزيع الفلوس',
    enter: 'نسب تقسيم فلوس الحجز: الليدر، الفنيين، الفني بلا ليدر، والشركة.',
    affects: '«الفلوس الي دخّلها» كل موظف بالإحصائيات والترتيب، وأهداف الموظفين.',
    when: 'مرة وحدة، وتغيّرها بس إذا تغيّرت سياسة التقسيم.',
  },
  {
    to: '/owner-costs',
    title: '💼 التكاليف والأهداف',
    enter: 'الموظفين بالحساب ورواتبهم والأساسي، ومصاريف الشهر، وفترات التوقف (عطل)، وربح الشركة فوق التكاليف.',
    affects: 'هدف كل موظف بالشهر (شكد لازم يدخّل)، والراتب المتوقع، وتقرير الترتيب.',
    when: 'الموظفين والرواتب مرة وحدة؛ المصاريف والتوقف كل شهر.',
  },
  {
    to: '/service-invoice-options',
    title: '🧾 خيارات فاتورة الخدمات',
    enter: 'لكل خدمة: البنود وسعر كل بند والحد الأدنى للعمل.',
    affects: 'تجريبي بس — ما مربوط بالفواتير الحقيقية لحد ما نجربه ونعتمده.',
    when: 'لما تضيف خدمة أو تتغيّر الأسعار.',
  },
  {
    to: '/owner-settings-solar',
    title: '☀️ الطاقة الشمسية',
    enter: 'استهلاك كل جهاز (كيلو واط)، وأسعار المنظومات لكل حجم: اقتصادي ووسط وغالي.',
    affects: 'شاشة استفسارات الطاقة الشمسية: حجم المنظومة والأسعار الي يكولها الموظف للزبون.',
    when: 'مرة وحدة، وكل ما تتغيّر الأسعار.',
  },
  {
    to: '/owner-settings-score',
    title: '⚖️ بنود التقييم',
    enter: 'لكل مجموعة موظفين: بنود تقييمها ووزن كل بند، وتكدر تطفّي بند.',
    affects: 'رقم تقييم كل موظف بهالمجموعة من لحظة الحفظ.',
    when: 'لما تريد تغيّر شنو أهم بتقييم دور معيّن.',
  },
  {
    to: '/owner-settings-remote',
    title: '🏠 فترة ساعات البيت',
    enter: 'يوم بداية الشهر (مثلاً 26)، ويوم وساعة رفع الساعات.',
    affects: 'ساعات العمل من البيت: من يا يوم ليا يوم تنحسب، وشوكت تنسكّر الفترة وتروح لسجل الأشهر.',
    when: 'مرة وحدة.',
  },
]

// (دازاي 11-26) «كلشي ممكن يصير يدوي سويله كارت» — كل شاشة إدخال يدوي بالنظام هنا،
// مقسّمة حسب النوع. الكارت يفتح شاشته نفسها بصلاحياتها؛ الي ما يگدر يفتحها ما يشوف كارتها.
type Card = { to: string; title: string; enter: string; affects: string }
const GROUPS: { title: string; cards: Card[] }[] = [
  { title: '💰 الفلوس والأسعار', cards: [
    { to: '/owner-money-split', title: '💰 توزيع الفلوس', enter: 'نسب تقسيم فلوس الحجز.', affects: 'الفلوس الي دخّلها كل موظف، والترتيب.' },
    { to: '/owner-costs', title: '💼 الرواتب والتكاليف والأهداف', enter: 'الرواتب والأساسي، المصاريف، التوقف، ربح الشركة.', affects: 'هدف كل موظف وراتبه المتوقع.' },
    { to: '/service-invoice-options', title: '🧾 خيارات فاتورة الخدمات', enter: 'بنود كل خدمة وسعرها والحد الأدنى.', affects: 'تجريبي — ما مربوط بالفواتير لسه.' },
    { to: '/network-prices', title: '🌐 أسعار الشبكات', enter: 'سعر كل مادة وشغلة بالشبكات.', affects: 'حساب كلفة الشبكات وعروض الأسعار.' },
    { to: '/decor-rates', title: '🪚 أسعار الديكور', enter: 'سعر المتر لكل نوع ديكور.', affects: 'حساب كلفة الديكور.' },
    { to: '/gps-install-costs', title: '🔧 تكاليف شد الـGPS', enter: 'كلفة الشد لكل نوع.', affects: 'حساب ربح أجهزة الـGPS.' },
    { to: '/products', title: '📦 المنتجات', enter: 'المنتجات وأسعارها.', affects: 'الفواتير والمخزن وعروض الأسعار.' },
    { to: '/expenses', title: '🧾 المصاريف', enter: 'مصاريف الشركة اليومية.', affects: 'الحسابات والتدقيق.' },
    { to: '/revolving-fund', title: '💵 الدوار', enter: 'تمويل الدوار وحركاته.', affects: 'رصيد الدوار عند الموظفين.' },
    { to: '/vouchers', title: '🧾 سندات الصرف', enter: 'سند صرف لكل دفعة.', affects: 'سجل الصرف والحسابات.' },
  ] },
  { title: '🚗 السيارات', cards: [
    { to: '/vehicles', title: '🚙 المركبات', enter: 'السيارات ومعلوماتها وأوراقها ومشاكلها وقطعها.', affects: 'الأسطول والتنبيهات والصيانة.' },
    { to: '/vehicle-drivers', title: '🪪 السواق وبطاقات البنزين', enter: 'منو يسوق شنو، وبطاقات البنزين.', affects: 'متابعة البنزين والسواق.' },
    { to: '/car-wash', title: '🧽 غسل السيارات', enter: 'تسجيل الغسل.', affects: 'تقييم نظافة السيارات.' },
  ] },
  { title: '🛠️ الخدمات والشغل', cards: [
    { to: '/services', title: '🛠️ الخدمات', enter: 'الخدمات الي تقدمها الشركة.', affects: 'الحجوزات والفواتير والإحصائيات.' },
    { to: '/service-managers', title: '👔 مسؤولو الخدمات', enter: 'مسؤول كل خدمة.', affects: 'منو يستلم حجوزات وتنبيهات الخدمة.' },
    { to: '/project-work-types', title: '🏗️ أنواع الأعمال', enter: 'أنواع أعمال المشاريع.', affects: 'المشاريع وحساب كلفتها.' },
    { to: '/stage-buckets', title: '⏱️ مراحل المشاريع', enter: 'مراحل المشروع ومددها.', affects: 'متابعة تأخير المشاريع.' },
    { to: '/maintenance-contracts', title: '🔄 عقود الصيانة', enter: 'عقود الصيانة ومواعيدها.', affects: 'حجوزات الصيانة التلقائية.' },
    { to: '/store-tools', title: '🧰 أدوات المخزن', enter: 'أدوات المخزن.', affects: 'العهدة والجرد.' },
    { to: '/suppliers', title: '🏪 الموردون', enter: 'الموردين ومعلوماتهم.', affects: 'المشتريات والأسعار.' },
  ] },
  { title: '👥 الموظفين والدوام', cards: [
    { to: '/employees', title: '👥 الموظفين', enter: 'الموظفين وأدوارهم.', affects: 'كل النظام.' },
    { to: '/departments', title: '🏢 الأقسام', enter: 'الأقسام.', affects: 'الهيكل والتقارير.' },
    { to: '/org', title: '🏛️ الهيكل التنظيمي', enter: 'الإدارات والوحدات ومدرائها.', affects: 'منو يشوف شنو، والتقييم.' },
    { to: '/work-schedule', title: '🕘 جدول الدوام', enter: 'أوقات دوام كل موظف.', affects: 'الحضور والتأخير.' },
    { to: '/owner-settings-score', title: '⚖️ بنود التقييم وأوزانها', enter: 'وزن كل بند لكل مجموعة، أو تطفّيه.', affects: 'رقم تقييم الموظفين.' },
    { to: '/kpi', title: '⭐ معايير التقييم', enter: 'معايير التقييم ونقاطها.', affects: 'تقييم الموظفين.' },
    { to: '/owner-settings-solar', title: '☀️ الطاقة الشمسية', enter: 'استهلاك الأجهزة وأسعار المنظومات لكل حجم.', affects: 'حجم المنظومة والأسعار باستفسارات الطاقة.' },
    { to: '/owner-settings-remote', title: '🏠 فترة ساعات البيت', enter: 'يوم بداية الشهر وموعد الرفع.', affects: 'حساب ساعات البيت.' },
    { to: '/core-matrix', title: '🏴‍☠️ الصلاحيات', enter: 'صلاحيات كل دور وموظف.', affects: 'منو يفتح شنو.' },
  ] },
  { title: '📋 ثانية', cards: [
    { to: '/vip-customers', title: '⭐ الشخصيات المهمة', enter: 'الزبائن المهمين.', affects: 'تنبيهات وأولوية حجوزاتهم.' },
    { to: '/voice-terms', title: '👂 قاموس الأذن', enter: 'كلمات ومصطلحات الشركة.', affects: 'دقة تفريغ الصوت.' },
    { to: '/privacy-policy', title: '🔒 سياسة الخصوصية', enter: 'بنود السياسة.', affects: 'موافقة الموظفين.' },
  ] },
]

function Guide() {
  const { employee, permissions, gpsServiceId } = useSession()
  const owner = employee?.actualRole === 'OWNER'
  const ctx = { employee, permissions, gpsServiceId }
  const ownerOnly = new Set(['/owner-settings-score', '/owner-money-split', '/owner-costs', '/owner-settings-remote', '/owner-money-audit'])
  const ok = (to: string) => owner || (!ownerOnly.has(to) && isPathAllowed(to, ctx) !== false)
  return (
    <div className="space-y-4">
      <div className={card}>
        <h1 className="text-xl font-extrabold text-[#0f2040]">⚙️ الإعدادات — كل شي يندخل باليد</h1>
        <p className="mt-1 text-sm text-slate-500">كل كارت يكلك شنو تدخّل بيه وعلى شنو يأثر. اضغط عليه يفتح شاشته.</p>
      </div>
      {GROUPS.map((g) => {
        const cards = g.cards.filter((c) => ok(c.to))
        if (!cards.length) return null
        return (
          <section key={g.title}>
            <h2 className="mb-2 text-lg font-extrabold text-[#0f2040]">{g.title}</h2>
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {cards.map((x) => (
                <Link key={x.to} to={x.to} className={`${card} block transition hover:ring-2 hover:ring-[#0f2040]/30`}>
                  <h3 className="font-extrabold text-[#0f2040]">{x.title}</h3>
                  <p className="mt-1 text-sm"><b className="text-slate-700">تدخّل: </b><span className="text-slate-600">{x.enter}</span></p>
                  <p className="text-sm"><b className="text-slate-700">يأثر على: </b><span className="text-slate-600">{x.affects}</span></p>
                </Link>
              ))}
            </div>
          </section>
        )
      })}
    </div>
  )
}

export default function OwnerSettingsHub() {
  const { employee } = useSession()
  const location = useLocation()
  // الأدمن يوصل لخيارات الفاتورة بس — باقي التبويبات للمالك
  if (employee?.actualRole !== 'OWNER') {
    if (location.pathname.startsWith('/service-invoice-options')) return <Suspense fallback={null}><ServiceInvoiceOptionsPage /></Suspense>
    // الأدمن يعدّل إعدادات الطاقة الشمسية هم (الخادم يسمح للمدير والمالك)
    if (location.pathname.startsWith('/owner-settings-solar') && employee?.actualRole === 'ADMIN') return <Suspense fallback={null}><SolarSettingsPage /></Suspense>
    return <Guide />
  }
  return (
    <TabbedHub tabs={[
      { path: '/owner-settings', label: '🗂️ كل الإعدادات', hint: 'كل شي يندخل باليد ووين يأثر', render: () => <Guide /> },
      { path: '/owner-money-audit', label: '🔎 فحص الأرقام', hint: 'الفواتير والليدر والصافي والتقييم', render: () => <MoneyAuditPage /> },
      ...ITEMS.map((x) => ({
        path: x.to,
        label: x.title,
        hint: x.affects.length > 48 ? x.affects.slice(0, 46) + '…' : x.affects,
        render: () =>
          x.to === '/owner-money-split' ? <OwnerMoneySplitPage />
          : x.to === '/owner-costs' ? <OwnerCostsPage />
          : x.to === '/service-invoice-options' ? <ServiceInvoiceOptionsPage />
          : x.to === '/owner-settings-solar' ? <SolarSettingsPage />
          : x.to === '/owner-settings-score' ? <ScoreWeightsPage />
          : <RemotePeriodCard canEdit />,
      })),
    ]} />
  )
}
