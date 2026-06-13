import { useEffect, useState } from 'react'
import { api, type Stats } from '../api'
import { useSession } from '../session'

const levels = [
  { level: 1, label: 'متدرب', min: 0 },
  { level: 2, label: 'فني مبتدئ', min: 3 },
  { level: 3, label: 'فني', min: 6 },
  { level: 4, label: 'فني متمرس', min: 10 },
  { level: 5, label: 'فني خبير', min: 15 },
]

const BOOKINGS_PER_RANK = 10

export default function MyRanking() {
  const { employee } = useSession()
  const [stats, setStats] = useState<Stats | null>(null)

  useEffect(() => {
    api.getStats().then(setStats).catch(() => setStats(null))
  }, [])

  const skillCount = employee?.skills.filter((s) => s.canPerform).length || 0
  const currentLevel = [...levels].reverse().find((l) => skillCount >= l.min) || levels[0]
  const nextLevel = levels.find((l) => l.min > skillCount)

  const myTechStat = stats?.technicianStats.find((s) => s.employeeId === employee?.id)
  const completedCount = myTechStat?.completed || 0

  const rank = Math.floor(completedCount / BOOKINGS_PER_RANK) + 1
  const remainingForNextRank = BOOKINGS_PER_RANK - (completedCount % BOOKINGS_PER_RANK)

  const sortedTechs = stats
    ? [...stats.technicianStats].sort((a, b) => b.completed - a.completed)
    : []
  const myPosition = sortedTechs.findIndex((s) => s.employeeId === employee?.id)

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">تصنيفي</h2>
      <p className="mt-1 text-slate-500">مستواك ورانكك بالنظام، ومركزك بين باقي الفنيين.</p>

      <div className="mt-6 grid grid-cols-1 gap-6 sm:grid-cols-2">
        <div className="rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 p-6 text-white shadow-lg shadow-brand-900/20">
          <p className="text-sm text-brand-100">المستوى (حسب المهارات المعتمدة)</p>
          <p className="mt-1 text-3xl font-extrabold">
            {currentLevel.level} - {currentLevel.label}
          </p>
          <p className="mt-2 text-sm text-brand-100">
            {skillCount} مهارة معتمدة
            {nextLevel
              ? ` - يحتاج ${nextLevel.min - skillCount} مهارة إضافية للترقي إلى "${nextLevel.label}"`
              : ' - وصلت لأعلى مستوى!'}
          </p>
        </div>

        <div className="rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <p className="text-sm text-slate-500">الرانك</p>
          <p className="mt-1 text-3xl font-extrabold text-emerald-700">{rank}</p>
          <p className="mt-2 text-sm text-slate-500">
            {completedCount} حجز منجز - يحتاج {remainingForNextRank} حجز إضافي للصعود للرانك التالي
          </p>
          {myPosition >= 0 && (
            <p className="mt-2 text-sm text-slate-500">
              ترتيبك بين الفنيين: <span className="font-bold text-brand-700">#{myPosition + 1}</span> من{' '}
              {sortedTechs.length}
            </p>
          )}
        </div>
      </div>
    </div>
  )
}
