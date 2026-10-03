import TabsShell from '../components/TabsShell'
import ProductsPage from './ProductsPage'
import ProductRequestsPage from './ProductRequestsPage'
import ExhibitionsPage from './ExhibitionsPage'

// «هاي بوحدة التقنيين، ذن يكونن بواجهة وحدة — ما أريدهن يكونن هيج
// بالقائمة الجانبية».
//
// الثلاثة سلسلة وحدة: المنتج ينقترح (طلب) ← ينضاف للكتالوگ ← وينعرض
// بالمعرض. ثلاث بنود بالقائمة تخلّي التقني يقفز بينهن ويضيع خيط الشغل.
export default function TechProductsHub() {
  return (
    <TabsShell
      title="📦 المنتجات والمعارض"
      subtitle="كتالوگ المنتجات وطلباتها والمعارض"
      tabs={[
        { key: 'products', label: 'المنتجات', icon: '📦', render: () => <ProductsPage embedded /> },
        { key: 'requests', label: 'طلبات المنتجات', icon: '📝', render: () => <ProductRequestsPage embedded /> },
        { key: 'exhibitions', label: 'إدارة المعارض', icon: '🏛️', render: () => <ExhibitionsPage embedded /> },
      ]}
    />
  )
}
