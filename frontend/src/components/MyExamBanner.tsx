import { useEffect, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { api } from '../api'
import { useSession } from '../session'

// تذكير «عندك اختبار الشهر» — يطلع لحد ما يسلّمه (ع 10-13)
export default function MyExamBanner() {
  const { employee } = useSession()
  const loc = useLocation()
  const [title, setTitle] = useState<string | null>(null)
  useEffect(() => {
    if (!employee) return
    api.tcMyExam().then((d) => setTitle(d.exam && !d.attempt?.finishedAt ? d.exam.title : null)).catch(() => {})
  }, [employee, loc.pathname])
  if (!title || loc.pathname === '/my-exam') return null
  return (
    <Link to="/my-exam" dir="rtl" className="mb-3 flex items-center justify-between rounded-xl bg-gradient-to-l from-[#0f2040] to-[#2c5aad] px-4 py-2.5 text-sm font-bold text-white shadow">
      <span>📝 عندك {title} — سوّيه من هنا</span><span>←</span>
    </Link>
  )
}
