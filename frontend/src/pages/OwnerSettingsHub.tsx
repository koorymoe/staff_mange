import { Suspense } from 'react'
import { Link } from 'react-router-dom'
import { useSession } from '../session'
import TabbedHub from '../components/TabbedHub'
import RemotePeriodCard from '../components/RemotePeriodCard'
import { card } from '../ui/classes'
import { lazyRetry } from '../utils/staleReload'

// ═══ (دازاي 11-26) «⚙️ الإعدادات»: كل إعداد يدوي يأثر على الفلوس بمكان واحد ═══
// «اريد اعدادات… كلهن بنفس المكان بس مسميات مختلفه حته اعرف كل اعداد وين يروح».
// كل تبويب شاشته القديمة نفسها بمساره القديم، فالصلاحيات ما تتغيّر.
const OwnerMoneySplitPage = lazyRetry(() => import('./OwnerMoneySplitPage'))
const OwnerCostsPage = lazyRetry(() => import('./OwnerCostsPage'))
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
    to: '/owner-settings-remote',
    title: '🏠 فترة ساعات البيت',
    enter: 'يوم بداية الشهر (مثلاً 26)، ويوم وساعة رفع الساعات.',
    affects: 'ساعات العمل من البيت: من يا يوم ليا يوم تنحسب، وشوكت تنسكّر الفترة وتروح لسجل الأشهر.',
    when: 'مرة وحدة.',
  },
]

function Guide() {
  return (
    <div className="space-y-3">
      <div className={card}>
        <h1 className="text-xl font-extrabold text-[#0f2040]">⚙️ الإعدادات — دليل سريع</h1>
        <p className="mt-1 text-sm text-slate-500">كل شي يندخل بيدك ويأثر على الفلوس هنا. اختار التبويب من فوگ، وهاي شنو يسوي كل واحد:</p>
      </div>
      <div className="grid gap-3 md:grid-cols-2">
        {ITEMS.map((x) => (
          <Link key={x.to} to={x.to} className={`${card} block transition hover:ring-2 hover:ring-[#0f2040]/30`}>
            <h2 className="text-lg font-extrabold text-[#0f2040]">{x.title}</h2>
            <dl className="mt-2 space-y-1.5 text-sm">
              <div><dt className="inline font-bold text-slate-700">شنو تدخّل: </dt><dd className="inline text-slate-600">{x.enter}</dd></div>
              <div><dt className="inline font-bold text-slate-700">يأثر على: </dt><dd className="inline text-slate-600">{x.affects}</dd></div>
              <div><dt className="inline font-bold text-slate-700">شوكت: </dt><dd className="inline text-slate-600">{x.when}</dd></div>
            </dl>
          </Link>
        ))}
      </div>
    </div>
  )
}

export default function OwnerSettingsHub() {
  const { employee } = useSession()
  // الأدمن يوصل لخيارات الفاتورة بس — باقي التبويبات للمالك
  if (employee?.actualRole !== 'OWNER') {
    return <Suspense fallback={null}><ServiceInvoiceOptionsPage /></Suspense>
  }
  return (
    <TabbedHub tabs={[
      { path: '/owner-settings', label: '📖 الدليل', hint: 'شنو كل إعداد ووين يروح', render: () => <Guide /> },
      ...ITEMS.map((x) => ({
        path: x.to,
        label: x.title,
        hint: x.affects.length > 48 ? x.affects.slice(0, 46) + '…' : x.affects,
        render: () =>
          x.to === '/owner-money-split' ? <OwnerMoneySplitPage />
          : x.to === '/owner-costs' ? <OwnerCostsPage />
          : x.to === '/service-invoice-options' ? <ServiceInvoiceOptionsPage />
          : <RemotePeriodCard canEdit />,
      })),
    ]} />
  )
}
