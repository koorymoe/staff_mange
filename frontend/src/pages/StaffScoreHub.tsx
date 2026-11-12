import TabbedHub from '../components/TabbedHub'
import { lazyRetry } from '../utils/staleReload'

// ═══ 🏆 تصنيف المدراء والموظفين — دازاي 11-12 ═══
// «صار عدي هواي تقييمات (تصنيف المدراء، الإحصائيات، تقييمات ماتركس) — مشتت للانتباه.
//  خليهن كلهن بمكان واحد سميه تصنيف المدراء والموظفين… تقييم ماتركس يضل بمكانه وبنفس الوقت
//  يندمج… وما اريده يظهر بالإحصائيات». كل تبويب شاشته القديمة ومساره القديم (حراسه نفسها)،
// والتبويب يطلع بس للي يگدر يفتحه.
const DirectorsRankingPage = lazyRetry(() => import('./DirectorsRankingPage'))
const StaffScorePage = lazyRetry(() => import('./StaffScorePage'))
const EmployeeIndicatorsPage = lazyRetry(() => import('./EmployeeIndicatorsPage'))
const EmployeeMonthlyStatsPage = lazyRetry(() => import('./EmployeeMonthlyStatsPage'))
const StatsPage = lazyRetry(() => import('./StatsPage'))

export default function StaffScoreHub() {
  return (
    <div className="space-y-3">
      <h1 className="text-2xl font-extrabold text-[#0f2040]">🏆 تصنيف المدراء والموظفين</h1>
      <TabbedHub tabs={[
        { path: '/org/directors-ranking', label: '🏛️ المدراء', hint: 'مدراء الأقسام ونوابهم ورؤساء الوحدات', render: () => <DirectorsRankingPage /> },
        { path: '/staff-score', label: '🏅 الموظفين — ماتركس', hint: 'كل دور حسب شغله: ماتركس ٦٠٪ + التقييم البشري ٤٠٪', render: () => <StaffScorePage /> },
        { path: '/employee-indicators', label: '📈 المؤشرات', hint: 'مؤشرات أداء كل موظف', render: () => <EmployeeIndicatorsPage /> },
        { path: '/employee-stats', label: '📆 أرقام الشهر', hint: 'أرقام كل موظف بالشهر حسب دوره', render: () => <EmployeeMonthlyStatsPage /> },
        { path: '/staff-numbers', label: '👥 أرقام الأقسام', hint: 'المبيعات والتنسيق والفنيين', render: () => <StatsPage /> },
      ]} />
    </div>
  )
}
