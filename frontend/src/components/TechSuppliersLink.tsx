import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api'

// رابط «موردين التقنيين» لمسؤول الخدمة الي دوره فني (القائمة مالته مقفلة على شغله).
export default function TechSuppliersLink() {
  const [can, setCan] = useState(false)
  useEffect(() => { void api.techSuppliersCan().then((r) => setCan(r.can)).catch(() => {}) }, [])
  if (!can) return null
  return (
    <Link to="/tech-suppliers" dir="rtl" className="flex items-center justify-between rounded-2xl border border-sky-200 bg-sky-50 p-3 text-sm font-bold text-sky-900">
      <span>🏪 موردين التقنيين — موردينك وإضافة مورّد جديد</span><span>←</span>
    </Link>
  )
}
