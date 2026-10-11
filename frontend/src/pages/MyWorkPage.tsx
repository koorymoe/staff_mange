import { lazy, Suspense, useState } from 'react'
import { useSession } from '../session'

// ═══ «مهامي وإنجازاتي» — ترتيب (ع) 10-04 ═══
// «مهامي الإضافية» و«إنجازاتي اليوم» شاشة وحدة بتبويبين. نفس الشاشتين بالضبط
// (مو نسخة) — الرابطان القديمان يبقون شغّالين.
const MyExtraTasksPage = lazy(() => import('./MyExtraTasksPage'))
const MyAchievementsPage = lazy(() => import('./MyAchievementsPage'))
// (ع) 10-11: المالك والمدير يشوفون مهام وإنجازات الموظفين بدل «مهامي» الفارغة
const ExtraTasksPage = lazy(() => import('./ExtraTasksPage'))
const AchievementsPage = lazy(() => import('./AchievementsPage'))

const TABS = [
  { id: 'tasks', label: '📋 مهامي الإضافية' },
  { id: 'achievements', label: '📝 إنجازاتي اليوم' },
] as const

export default function MyWorkPage() {
  const { employee } = useSession()
  const boss = employee?.role === 'ADMIN' || employee?.actualRole === 'OWNER'
  const [tab, setTab] = useState<(typeof TABS)[number]['id']>('tasks')
  return (
    <div dir="rtl" className="space-y-4">
      <div className="flex flex-wrap gap-2">
        {TABS.map((t) => (
          <button key={t.id} onClick={() => setTab(t.id)}
            className={`rounded-xl px-4 py-2 text-sm font-bold transition ${tab === t.id ? 'bg-brand-700 text-white' : 'bg-white text-slate-600 shadow-sm'}`}>
            {boss ? (t.id === 'tasks' ? '📋 مهام الموظفين' : '📝 إنجازات الموظفين') : t.label}
          </button>
        ))}
      </div>
      <Suspense fallback={<p className="text-slate-400">جاري التحميل…</p>}>
        {boss
          ? (tab === 'tasks' ? <ExtraTasksPage embedded /> : <AchievementsPage />)
          : (tab === 'tasks' ? <MyExtraTasksPage /> : <MyAchievementsPage />)}
      </Suspense>
    </div>
  )
}
