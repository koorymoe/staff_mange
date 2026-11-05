import TabbedHub from '../components/TabbedHub'
import { lazyRetry } from '../utils/staleReload'

// ٧و (دازاي 11-05): «تقييم الموظفين» و«مؤشرات أداء الموظف» — نفس الناس ونفس البيانات،
// صارن شاشة وحدة بتبويبين. كل تبويب مساره القديم، فالحراس والروابط نفسها.
const StaffScorePage = lazyRetry(() => import('./StaffScorePage'))
const EmployeeIndicatorsPage = lazyRetry(() => import('./EmployeeIndicatorsPage'))

export default function StaffScoreHub() {
  return (
    <TabbedHub tabs={[
      { path: '/staff-score', label: '🏅 الترتيب', hint: 'ماتركس ٦٠٪ + التقييم البشري ٤٠٪', render: () => <StaffScorePage /> },
      { path: '/employee-indicators', label: '📈 المؤشرات', hint: 'مؤشرات أداء كل موظف', render: () => <EmployeeIndicatorsPage /> },
    ]} />
  )
}
