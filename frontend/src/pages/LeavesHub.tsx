import TabbedHub from '../components/TabbedHub'
import { lazyRetry } from '../utils/staleReload'

// ٧و (دازاي 11-05): «طلبات الإجازات» و«إجازات الموظفين» شاشة وحدة بتبويبين.
// كل تبويب مساره القديم (/leaves و/staff-leaves) فالقائمة والصلاحيات نفسها.
const LeavesPage = lazyRetry(() => import('./LeavesPage'))
const StaffLeavesPage = lazyRetry(() => import('./StaffLeavesPage'))

export default function LeavesHub() {
  return (
    <TabbedHub tabs={[
      { path: '/leaves', label: '🏖️ الطلبات', hint: 'طلباتي، والطلبات الي تنتظر قرارك', render: () => <LeavesPage /> },
      { path: '/staff-leaves', label: '🗓️ سجل إجازات الموظفين', hint: 'منو مجاز اليوم، وكل طلبات الشهر', render: () => <StaffLeavesPage /> },
    ]} />
  )
}
