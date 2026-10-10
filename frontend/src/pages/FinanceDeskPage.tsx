import { lazy } from 'react'
import TabHub from '../components/TabHub'
import { useSession } from '../session'

const Finance = lazy(() => import('./Finance'))
const DailyAuditPage = lazy(() => import('./DailyAuditPage'))
const RevolvingFundPage = lazy(() => import('./RevolvingFundPage'))
const ExpensesReview = lazy(() => import('./ExpensesReview'))
const AuditIssuesPage = lazy(() => import('./AuditIssuesPage'))

// «التدقيق والدوار» — شاشة المحاسب الوحدة (قرار (ع) 10-08). نفس شروط بنود القائمة القديمة.
export default function FinanceDeskPage() {
  const { employee, permissions } = useSession()
  const role = employee?.role
  const top = role === 'ADMIN' || employee?.actualRole === 'OWNER'
  const has = (...p: string[]) => top || p.some((x) => permissions.includes(x))
  const audit = top || role === 'FINANCE' || has('finance', 'finance_audit')
  return (
    <TabHub title="التدقيق والدوار" icon="🧾" subtitle="تدقيق الحسابات واليومي والدوار والمصاريف والبلاغات — من مكان واحد"
      tabs={[
        { id: 'audit', label: 'تدقيق الحسابات', icon: '📊', show: audit, render: () => <Finance /> },
        { id: 'daily', label: 'التدقيق اليومي', icon: '📅', show: audit, render: () => <DailyAuditPage /> },
        { id: 'fund', label: 'الدوار', icon: '💵', show: has('revolving_fund'), render: () => <RevolvingFundPage /> },
        { id: 'expenses', label: 'إدارة المصاريف', icon: '🧾', show: top || role === 'FINANCE' || has('expenses_manage'), render: () => <ExpensesReview /> },
        { id: 'issues', label: 'أخطاء التدقيق', icon: '💸', show: top || role === 'FINANCE' || has('audit_issues', 'finance_audit'), render: () => <AuditIssuesPage /> },
      ]} />
  )
}
