import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, type Permission } from '../api'
import { useSession } from '../session'

interface SubsystemCard {
  key: string
  icon: string
  title: string
  description: string
  path: string
  permissionName: string
  borderColor: string
}

const cards: SubsystemCard[] = [
  {
    key: 'staff',
    icon: '\u{1F465}',
    title: 'إدارة الكوادر',
    description: 'إدارة بيانات الموظفين والكوادر البشرية',
    path: '/employees',
    permissionName: 'staff_management',
    borderColor: 'border-t-blue-600',
  },
  {
    key: 'gps',
    icon: '\u{1F4E1}',
    title: 'نظام GPS',
    description: 'تتبع المواقع وإدارة الأجهزة',
    path: '/gps',
    permissionName: 'gps',
    borderColor: 'border-t-teal-600',
  },
  {
    key: 'quotations',
    icon: '\u{1F4B0}',
    title: 'عرض السعر',
    description: 'إعداد وإدارة عروض الأسعار',
    path: '/quotations',
    permissionName: 'quotations',
    borderColor: 'border-t-amber-600',
  },
  {
    key: 'complaints',
    icon: '\u{1F4E2}',
    title: 'الشكاوى',
    description: 'متابعة شكاوى العملاء والموظفين',
    path: '/complaints',
    permissionName: 'complaints',
    borderColor: 'border-t-red-500',
  },
  {
    key: 'inventory',
    icon: '\u{1F527}',
    title: 'جرد الأدوات',
    description: 'إدارة المخزون والأدوات',
    path: '/inventory',
    permissionName: 'inventory',
    borderColor: 'border-t-orange-600',
  },
  {
    key: 'kpi',
    icon: '⭐',
    title: 'تقييم الأداء',
    description: 'مؤشرات الأداء والتقييمات',
    path: '/kpi',
    permissionName: 'kpi',
    borderColor: 'border-t-yellow-500',
  },
  {
    key: 'finance',
    icon: '\u{1F4B5}',
    title: 'تدقيق الحسابات',
    description: 'التقارير المالية وتدقيق الحسابات',
    path: '/finance',
    permissionName: 'finance',
    borderColor: 'border-t-emerald-600',
  },
]

export default function Dashboard() {
  const { employee } = useSession()
  const navigate = useNavigate()
  const [allowedCards, setAllowedCards] = useState<SubsystemCard[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!employee) return

    if (employee.role === 'ADMIN') {
      setAllowedCards(cards)
      setLoading(false)
      return
    }

    Promise.all([
      api.getEmployeePermissions(employee.id),
      api.getPermissions(),
    ])
      .then(([empPerms, allPerms]) => {
        const empPermNames = new Set(empPerms.map((p) => p.name))
        // Also include permission names from allPerms that match
        const filtered = cards.filter((card) => empPermNames.has(card.permissionName))
        setAllowedCards(filtered)
      })
      .catch(() => {
        setAllowedCards([])
      })
      .finally(() => setLoading(false))
  }, [employee])

  if (!employee) return null

  return (
    <div dir="rtl">
      <div className="mb-8">
        <h2 className="text-3xl font-extrabold text-brand-900">
          مرحباً، {employee.name}
        </h2>
        <p className="mt-2 text-lg text-slate-500">
          اختر أحد الأنظمة للبدء
        </p>
      </div>

      {loading ? (
        <p className="text-slate-400">جاري تحميل الصلاحيات...</p>
      ) : allowedCards.length === 0 ? (
        <div className="rounded-xl bg-amber-50 p-6 text-center">
          <p className="text-lg font-medium text-amber-700">
            لا توجد لديك صلاحيات للوصول إلى أي نظام حالياً.
          </p>
          <p className="mt-2 text-sm text-amber-600">
            يرجى التواصل مع مدير النظام لتفعيل الصلاحيات.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {allowedCards.map((card) => (
            <button
              key={card.key}
              onClick={() => navigate(card.path)}
              className={`group rounded-xl border-t-4 ${card.borderColor} border border-white bg-white p-8 text-right shadow-[0_4px_20px_rgba(15,32,64,0.06)] transition-all hover:-translate-y-1 hover:shadow-lg hover:shadow-brand-900/10`}
            >
              <div className="text-5xl">{card.icon}</div>
              <h3 className="mt-4 text-xl font-bold text-brand-900 group-hover:text-brand-700">
                {card.title}
              </h3>
              <p className="mt-2 text-sm text-slate-500">{card.description}</p>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
