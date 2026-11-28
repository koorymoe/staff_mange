import { useEffect, useState } from 'react'
import { apiRequest } from '../api'

// ═══ 🎯 هدفي من الفلوس — شاشة الموظف (ل) 11-24 دازاي ═══
// «كل موظف يطلعله بشاشته: هلكد دخّلت وهلكد باقي عليك، ويشوف راتبه حسب الشغل الي اشتغله».
// يطلع بس للي المالك حطّه «يدخّل فلوس» بتكاليف الشركة.

type Team = { earners: number; brought: number; target: number; pct: number; reached: number }
type T = { overhead: number; target: number; daily: number; hourly: number; brought: number; remaining: number; forecast: number; pct: number; salary: number; baseSalary: number; expectedNow: number; expectedSalary: number; dailyNeeded: number; errors: number; fixes: number; rank: number }
const iqd = (v: number) => Math.round(v).toLocaleString('en-US')

export default function MyMoneyTargetCard() {
  const [t, setT] = useState<T | null>(null)
  const [team, setTeam] = useState<Team | null>(null)
  useEffect(() => { apiRequest<{ target: T | null; team?: Team | null }>('/me/money-target').then((r) => { setT(r.target); setTeam(r.team ?? null) }).catch(() => {}) }, [])
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
      {/* (دازاي 11-28) «الفرد للكل والكل للفرد»: هدفك = راتبك + حصتك من رواتب زملائك الإداريين والمصاريف */}
      <div className="rounded-xl bg-slate-50 p-3 text-xs leading-relaxed text-slate-700">
        <p>
          <b>من وين جا هدفك؟</b> راتبك <b dir="ltr">{iqd(t.salary)}</b> + حصتك من رواتب زملائك الي ما يدخّلون فلوس (الإداريين) ومصاريف الشركة <b dir="ltr">{iqd(t.overhead)}</b>
          {' '}= <b dir="ltr">{iqd(t.target)}</b> بالشهر، يعني تقريباً <b dir="ltr">{iqd(t.daily)}</b> باليوم.
        </p>
        <p className="mt-1">
          {t.brought >= t.salary
            ? '✅ غطّيت راتبك — كل دينار تدخّله هسه يغطي رواتب زملائك. هذا معنى «الفرد للكل».'
            : `باقي ${iqd(Math.max(0, t.salary - t.brought))} حتى تغطي راتبك، وبعدها شغلك يغطي زملائك.`}
        </p>
      </div>
      {team && team.earners > 0 && (
        <div className="rounded-xl bg-[#0f2040]/5 p-3 text-xs text-slate-700">
          <div className="flex justify-between"><b>🤝 الفريق كله (الكل للفرد)</b><span>{team.reached} من {team.earners} وصلوا هدفهم</span></div>
          <div className="mt-1 flex justify-between"><span>دخّلنا سوا <b dir="ltr">{iqd(team.brought)}</b></span><span>من <b dir="ltr">{iqd(team.target)}</b> ({team.pct}٪)</span></div>
          <div className="mt-1 h-2 overflow-hidden rounded-full bg-slate-200"><div className="h-full bg-[#2c5aad]" style={{ width: `${Math.min(100, Math.max(0, team.pct))}%` }} /></div>
        </div>
      )}
      {(t.errors > 0 || t.fixes > 0) && <p className="text-xs text-slate-600">🔧 أخطاء فنية عليك: <b>{t.errors}</b> · صلّحت لغيرك: <b>{t.fixes}</b></p>}
    </section>
  )
}
