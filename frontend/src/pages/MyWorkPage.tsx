import { Suspense, useState } from 'react'
import { useSession, isAdminOrOwner } from '../session'
import { lazyRetry } from '../utils/staleReload'
import DailyReportBox from '../components/DailyReportBox'
import TodayRatings from '../components/TodayRatings'

// ═══ «مهامي وإنجازاتي» — ترتيب (ع) 10-04 ═══
// «مهامي الإضافية» و«إنجازاتي اليوم» شاشة وحدة بتبويبين. نفس الشاشتين بالضبط
// (مو نسخة) — الرابطان القديمان يبقون شغّالين.
const MyExtraTasksPage = lazyRetry(() => import('./MyExtraTasksPage'))
const MyAchievementsPage = lazyRetry(() => import('./MyAchievementsPage'))
// (ع) 10-11: المالك والمدير يشوفون مهام وإنجازات الموظفين بدل «مهامي» الفارغة
const ExtraTasksPage = lazyRetry(() => import('./ExtraTasksPage'))
const AchievementsPage = lazyRetry(() => import('./AchievementsPage'))

const TABS = [
  { id: 'tasks', label: '📋 مهامي الإضافية' },
  { id: 'achievements', label: '📝 إنجازاتي اليوم' },
] as const

export default function MyWorkPage() {
  const { employee } = useSession()
  const boss = isAdminOrOwner(employee)
  const [tab, setTab] = useState<(typeof TABS)[number]['id']>('tasks')
  return (
    <div dir="rtl" className="space-y-4">
      {!boss && <DailyReportBox />}
      {!boss && <TodayRatings />}
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
