import { lazy } from 'react'
import TabHub from '../components/TabHub'
import { useSession } from '../session'

const LeaderInvoiceNew = lazy(() => import('./LeaderInvoiceNew'))
const GpsInstallCostsPage = lazy(() => import('./GpsInstallCostsPage'))
const NetworkCostPage = lazy(() => import('./NetworkCostPage'))
const CameraCostPage = lazy(() => import('./CameraCostPage'))

// «حساب الكلفة» — الحاسبات الأربع بشاشة وحدة (قرار (ع) 10-08).
export default function CostCalcPage() {
  const { employee, permissions } = useSession()
  const top = employee?.role === 'ADMIN' || employee?.actualRole === 'OWNER'
  const has = (...p: string[]) => top || p.some((x) => permissions.includes(x))
  return (
    <TabHub title="حساب الكلفة" icon="🧮" subtitle="اختار شنو تريد تحسب"
      tabs={[
        { id: 'cost', label: 'حساب الكلفة', icon: '🧮', show: has('execution_cost'), render: () => <LeaderInvoiceNew /> },
        { id: 'install', label: 'تكاليف الشد', icon: '🔧', show: top || employee?.role === 'FINANCE' || has('gps_install_costs'), render: () => <GpsInstallCostsPage /> },
        { id: 'network', label: 'كلفة الشبكات', icon: '🌐', show: has('execution_cost'), render: () => <NetworkCostPage /> },
        { id: 'camera', label: 'كلفة الكاميرات', icon: '📷', show: has('execution_cost'), render: () => <CameraCostPage /> },
      ]} />
  )
}
