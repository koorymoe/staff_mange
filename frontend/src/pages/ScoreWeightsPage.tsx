import { useEffect, useState } from 'react'
import { apiRequest } from '../api'
import { card } from '../ui/classes'
import { notify } from '../ui/notify'

// ═══ ⚖️ بنود التقييم وأوزانها (دازاي 11-27، المرحلة الثانية) ═══
// لكل مجموعة: بنودها بأوزانها. تغيّر الوزن، أو تطفّي البند (يصير صفر) ووزنه يتوزع على الباقي.
// الأوزان نسبية: النظام يقسمها على مجموعها، والباقي «رأي ماتركس» ثابت.

type Item = { key: string; label: string; default: number; weight: number }
type Group = { group: string; label: string; items: Item[] }

export default function ScoreWeightsPage() {
  const [groups, setGroups] = useState<Group[] | null>(null)
  const [open, setOpen] = useState<string>('')
  const [busy, setBusy] = useState(false)
  useEffect(() => {
    apiRequest<{ groups: Group[] }>('/owner/score-weights')
      .then((x) => { setGroups(x.groups); setOpen(x.groups[0]?.group ?? '') })
      .catch((e) => notify(e instanceof Error ? e.message : 'تعذر التحميل'))
  }, [])
  if (!groups) return <p className="py-8 text-center text-slate-400">جاري التحميل…</p>
  const setW = (g: string, k: string, v: number) =>
    setGroups(groups.map((x) => (x.group !== g ? x : { ...x, items: x.items.map((i) => (i.key === k ? { ...i, weight: v } : i)) })))
  const save = async () => {
    setBusy(true)
    try {
      const weights: Record<string, Record<string, number>> = {}
      for (const g of groups) weights[g.group] = Object.fromEntries(g.items.map((i) => [i.key, i.weight]))
      const x = await apiRequest<{ groups: Group[] }>('/owner/score-weights', { method: 'PUT', body: JSON.stringify({ weights }) })
      setGroups(x.groups)
      notify('انحفظت أوزان التقييم ✅ — التقييم ينحسب بيها من هسه')
    } catch (e) {
      notify(e instanceof Error ? e.message : 'تعذر الحفظ')
    } finally {
      setBusy(false)
    }
  }
  return (
    <div className="space-y-3" dir="rtl">
      <div className={card}>
        <h2 className="text-lg font-extrabold text-[#0f2040]">⚖️ بنود التقييم وأوزانها</h2>
        <p className="mt-1 text-sm text-slate-500">
          كل مجموعة موظفين إلها بنود. الرقم = شكد البند مهم مقارنة بالباقي (النظام يحوّلها لنسبة). خلّي البند <b>صفر</b> حتى تطفّيه ووزنه يتوزع على الباقي.
          البند الي ماله بيانات لموظف معيّن (مثلاً ماعنده مشاريع) وزنه يتوزع هم. وفوگها «رأي ماتركس» ثابت.
        </p>
      </div>
      {groups.map((g) => {
        const sum = g.items.reduce((a, i) => a + i.weight, 0)
        const changed = g.items.some((i) => i.weight !== i.default)
        const isOpen = open === g.group
        return (
          <section key={g.group} className={card}>
            <button type="button" className="flex w-full items-center justify-between text-right" onClick={() => setOpen(isOpen ? '' : g.group)}>
              <span className="font-extrabold text-[#0f2040]">{g.label} {changed && <span className="mr-2 rounded-full bg-amber-100 px-2 text-xs text-amber-800">معدّل</span>}</span>
              <span className="text-xs text-slate-500">{g.items.filter((i) => i.weight > 0).length} بند شغّال {isOpen ? '▲' : '▼'}</span>
            </button>
            {isOpen && (
              <table className="mt-3 w-full text-sm">
                <thead className="text-xs text-slate-500"><tr><th className="text-right">البند</th><th>الأصلي</th><th>الوزن</th><th>النسبة</th></tr></thead>
                <tbody>
                  {g.items.map((i) => (
                    <tr key={i.key} className={`border-t ${i.weight === 0 ? 'opacity-50' : ''}`}>
                      <td className="py-1.5">{i.label}</td>
                      <td className="text-center text-slate-400">{i.default}</td>
                      <td className="text-center">
                        <input type="number" min={0} max={100} value={i.weight}
                          onChange={(e) => setW(g.group, i.key, Math.min(100, Math.max(0, Math.round(Number(e.target.value) || 0))))}
                          className="w-20 rounded-lg border border-slate-300 px-2 py-1 text-center" />
                      </td>
                      <td className="text-center text-slate-600">{sum > 0 && i.weight > 0 ? `${Math.round((i.weight / sum) * 100)}%` : 'مطفي'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
            {isOpen && changed && (
              <button type="button" className="mt-2 text-xs text-slate-500 underline"
                onClick={() => setGroups(groups.map((x) => (x.group !== g.group ? x : { ...x, items: x.items.map((i) => ({ ...i, weight: i.default })) })))}>
                ↺ رجّع الأصلي لهالمجموعة
              </button>
            )}
          </section>
        )
      })}
      <div className="flex justify-end">
        <button type="button" disabled={busy} onClick={save} className="rounded-lg bg-[#0f2040] px-6 py-2 font-bold text-white disabled:opacity-50">حفظ</button>
      </div>
    </div>
  )
}
