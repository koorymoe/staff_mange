import { useEffect, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { useSession } from '../session'
import { inTrainingWindow } from '../utils/trainingWindow'

// (دازاي 11-18) التدريب ما يحبس الموظف عن شغله: تنبيه بوقت التدريب الي جدوله المدير
// (أو دائماً إذا ما انحدد وقت)، والموظف يشتغل عادي ويدخل التدريب من الزر.
export default function TrainingBanner({ hasMaterials }: { hasMaterials: boolean | null }) {
  const { employee } = useSession()
  const location = useLocation()
  const [now, setNow] = useState(() => new Date())
  useEffect(() => {
    const t = setInterval(() => setNow(new Date()), 60_000)
    return () => clearInterval(t)
  }, [])
  if (!employee?.isTrainee || hasMaterials !== true || location.pathname === '/training') return null
  const from = employee.trainingFrom, to = employee.trainingTo
  if (!inTrainingWindow(from, to, now)) return null
  return (
    <div className="mb-3 flex flex-wrap items-center justify-between gap-2 rounded-2xl bg-gradient-to-l from-[#1e3a8a] to-[#2563eb] px-4 py-3 text-white shadow">
      <span className="text-sm font-bold">📚 {from && to ? `هسه وقت تدريبك (${from}–${to})` : 'عندك مواد تدريب'} — كمّلها، وشغلك يبقى مفتوح.</span>
      <Link to="/training" className="rounded-lg bg-white px-3 py-1.5 text-sm font-extrabold text-[#1e3a8a] hover:bg-blue-50">ادخل التدريب</Link>
    </div>
  )
}
