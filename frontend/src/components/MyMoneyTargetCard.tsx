import { useEffect, useState } from 'react'
import { apiRequest } from '../api'

// ═══ 🎯 هدفي من الفلوس — شاشة الموظف (ل) 11-24 دازاي ═══
// «كل موظف يطلعله بشاشته: هلكد دخّلت وهلكد باقي عليك، ويشوف راتبه حسب الشغل الي اشتغله».
// يطلع بس للي المالك حطّه «يدخّل فلوس» بتكاليف الشركة.

type T = { target: number; daily: number; hourly: number; brought: number; remaining: number; forecast: number; pct: number; salary: number; baseSalary: number; expectedNow: number; expectedSalary: number; dailyNeeded: number; errors: number; fixes: number; rank: number }
const iqd = (v: number) => Math.round(v).toLocaleString('en-US')

export default function MyMoneyTargetCard() {
  const [t, setT] = useState<T | null>(null)
  useEffect(() => { apiRequest<{ target: T | null }>('/me/money-target').then((r) => setT(r.target)).catch(() => {}) }, [])
  if (!t) return null
  const p = Math.min(100, Math.max(0, t.pct))
  const color = t.pct >= 100 ? 'bg-emerald-500' : t.pct >= 60 ? 'bg-amber-500' : 'bg-rose-500'
  return (
    <section className="space-y-3 rounded-2xl bg-white p-4 shadow-sm" data-testid="my-money-target">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-extrabold text-[#0f2040]">🎯 هدفي هالشهر</h2>
        <span className="text-xs text-slate-500">ترتيبك #{t.rank}</span>
      </div>
      <div>
        <div className="flex justify-between text-sm"><span>دخّلت <b dir="ltr">{iqd(t.brought)}</b></span><span>من <b dir="ltr">{iqd(t.target)}</b> ({t.pct}٪)</span></div>
        <div className="mt-1 h-3 overflow-hidden rounded-full bg-slate-100"><div className={`h-full ${color}`} style={{ width: `${p}%` }} /></div>
      </div>
      <div className="grid grid-cols-2 gap-2 text-sm sm:grid-cols-4">
        <div className="rounded-xl bg-slate-50 p-2"><p className="text-xs text-slate-500">باقي عليك</p><b dir="ltr">{iqd(t.remaining)}</b></div>
        <div className="rounded-xl bg-slate-50 p-2"><p className="text-xs text-slate-500">لازم باليوم</p><b dir="ltr">{iqd(t.dailyNeeded || t.daily)}</b></div>
        <div className="rounded-xl bg-slate-50 p-2"><p className="text-xs text-slate-500">على هالمنوال توصل</p><b dir="ltr" className={t.forecast >= t.target ? 'text-emerald-700' : 'text-rose-700'}>{iqd(t.forecast)}</b></div>
        <div className="rounded-xl bg-emerald-50 p-2 ring-1 ring-emerald-200"><p className="text-xs text-emerald-800">راتبك المتوقع</p><b dir="ltr" className="text-emerald-800">{iqd(t.expectedSalary)}</b><p className="text-[10px] text-slate-500">من {iqd(t.salary)} · الأساسي {iqd(t.baseSalary)}</p></div>
      </div>
      {(t.errors > 0 || t.fixes > 0) && <p className="text-xs text-slate-600">🔧 أخطاء فنية عليك: <b>{t.errors}</b> · صلّحت لغيرك: <b>{t.fixes}</b></p>}
    </section>
  )
}
