import TabbedHub from '../components/TabbedHub'
import { lazyRetry } from '../utils/staleReload'

// ٧و (دازاي 11-05): الموردين — ٣ شاشات صارت وحدة بتبويبات. القائمتين (الشركة
// والتقنيين) يبقن منفصلات بالبيانات مثل قرار (ع) 10-06. كل تبويب مساره القديم.
const SuppliersPage = lazyRetry(() => import('./SuppliersPage'))
const TechSuppliersPage = lazyRetry(() => import('./TechSuppliersPage'))
const SupplierRelationsPage = lazyRetry(() => import('./SupplierRelationsPage'))

export default function SuppliersHub() {
  return (
    <TabbedHub tabs={[
      { path: '/suppliers', label: '🏭 موردين الشركة', hint: 'موردين المشتريات', render: () => <SuppliersPage /> },
      { path: '/tech-suppliers', label: '🏪 موردين التقنيين', hint: 'القائمة الي يختار منها المدير لكل تقني', render: () => <TechSuppliersPage /> },
      { path: '/supplier-relations', label: '🤝 العلاقات', hint: 'شكد صارلنا ما تواصلنا ويا كل مورد', render: () => <SupplierRelationsPage /> },
    ]} />
  )
}
