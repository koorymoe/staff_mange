import { lazy, Suspense, useState } from 'react'

// ═══ «مهامي وإنجازاتي» — ترتيب (ع) 10-04 ═══
// «مهامي الإضافية» و«إنجازاتي اليوم» شاشة وحدة بتبويبين. نفس الشاشتين بالضبط
// (مو نسخة) — الرابطان القديمان يبقون شغّالين.
const MyExtraTasksPage = lazy(() => import('./MyExtraTasksPage'))
const MyAchievementsPage = lazy(() => import('./MyAchievementsPage'))

const TABS = [
  { id: 'tasks', label: '📋 مهامي الإضافية' },
  { id: 'achievements', label: '📝 إنجازاتي اليوم' },
] as const

export default function MyWorkPage() {
  const [tab, setTab] = useState<(typeof TABS)[number]['id']>('tasks')
  return (
    <div dir="rtl" className="space-y-4">
      <div className="flex flex-wrap gap-2">
        {TABS.map((t) => (
          <button key={t.id} onClick={() => setTab(t.id)}
            className={`rounded-xl px-4 py-2 text-sm font-bold transition ${tab === t.id ? 'bg-brand-700 text-white' : 'bg-white text-slate-600 shadow-sm'}`}>
            {t.label}
          </button>
        ))}
      </div>
      <Suspense fallback={<p className="text-slate-400">جاري التحميل…</p>}>
        {tab === 'tasks' ? <MyExtraTasksPage /> : <MyAchievementsPage />}
      </Suspense>
    </div>
  )
}
