import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type MatrixWatch } from '../api'
import MatrixBusiness from './MatrixBusiness'
import MatrixEyeGraphic from './MatrixEyeGraphic'
import { GROUP_COLOR, GROUP_LABEL, eyeColor, type EyeGroup, type EyeMood } from './matrixEyeColors'

// ═══ عيون ماتركس عند المدير ═══
// عين لكل مجموعة أدوار: «أنا مسؤولة عن …». لونها أسوأ حالة بمجموعتها،
// والضغط يطلع تقرير مباشر: كل موظف، عينه، شغله اليوم، وتذكيراته.

type GroupRep = { group: string; employees: MatrixWatch[]; red: number; alert: number }

const MOOD_LABEL: Record<string, string> = { ANGRY: '🔴 غاضبة', ALERT: '🟡 منتبهة', PLEASED: '🟢 راضية', CALM: '⚪ هادئة' }

function groupMood(g: GroupRep): EyeMood {
  if (g.red > 0) return 'ANGRY'
  if (g.alert > 0) return 'ALERT'
  const all = g.employees.length > 0 && g.employees.every((e) => e.mood === 'PLEASED')
  return all ? 'PLEASED' : 'CALM'
}

export default function MatrixRoleEyes() {
  const [data, setData] = useState<GroupRep[] | null>(null)
  const [err, setErr] = useState<string | null>(null)
  const [sel, setSel] = useState<string | null>(null)
  const [why, setWhy] = useState<string | null>(null)

  useEffect(() => {
    let alive = true
    api.getRoleWatch().then((d) => { if (alive) setData(d) }).catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر جلب العيون') })
    return () => { alive = false }
  }, [])

  if (err) return <p className="rounded-lg bg-red-50 p-3 text-red-600">{err}</p>
  if (!data) return <p className="text-slate-400">ماتركس يجمع العيون…</p>
  const current = data.find((g) => g.group === sel)

  return (
    <div className="space-y-4">
      <MatrixBusiness />
      <section className="rounded-2xl bg-[#0b1220] p-4 shadow-sm">
        <h3 className="mb-1 text-base font-extrabold text-white">👁️ عيون ماتركس</h3>
        <p className="mb-4 text-xs text-slate-400">كل عين مسؤولة عن مجموعة. لونها يتبع أسوأ حالة بموظفيها. اضغطها حتى تشوف تقريرها.</p>
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
          {data.map((g) => {
            const grp = g.group as EyeGroup
            const mood = groupMood(g)
            const color = eyeColor(grp, mood)
            return (
              <button key={g.group} type="button" onClick={() => setSel(sel === g.group ? null : g.group)}
                className="matrix-group-eye flex flex-col items-center gap-2 rounded-2xl p-3 text-center"
                style={{ background: sel === g.group ? `color-mix(in srgb, ${color} 18%, #0b1220)` : '#111a2e', boxShadow: `0 0 0 1px color-mix(in srgb, ${color} ${sel === g.group ? 90 : 35}%, transparent)` }}>
                <MatrixEyeGraphic group={grp} mood={mood} width={96} />
                <span className="text-[13px] font-bold" style={{ color }}>أنا مسؤولة عن {GROUP_LABEL[grp] ?? g.group}</span>
                <span className="text-[11px] text-slate-400">
                  {g.employees.length} موظف{g.red > 0 && <> · <b className="text-red-400">{g.red} حمرة</b></>}{g.alert > 0 && <> · <b className="text-amber-300">{g.alert} منتبهة</b></>}
                </span>
              </button>
            )
          })}
        </div>
      </section>

      {current && (
        <section className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm" style={{ borderColor: GROUP_COLOR[current.group as EyeGroup] }}>
          <h3 className="mb-3 text-base font-extrabold text-[#0f2040]">تقرير عين {GROUP_LABEL[current.group as EyeGroup] ?? current.group}</h3>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead><tr className="text-right text-xs text-slate-500"><th className="p-2">الموظف</th><th className="p-2">العين</th><th className="p-2">شغله اليوم</th><th className="p-2">تذكيرات مفتوحة</th></tr></thead>
              <tbody>
                {current.employees.map((e) => (
                  <tr key={e.id} className="border-t border-slate-100 align-top">
                    <td className="p-2 font-bold"><Link to={`/matrix/employee/${e.id}`} className="text-brand-700 hover:underline">{e.name} ←</Link></td>
                    <td className="p-2 whitespace-nowrap">{MOOD_LABEL[e.mood] ?? e.mood}</td>
                    <td className="p-2">
                      {e.workload.length === 0 ? <span className="text-slate-400">—</span> : (
                        <ul className="space-y-0.5 text-xs">
                          {e.workload.map((w) => <li key={w.key}>{w.label}: {w.done > 0 && <span className="text-emerald-700">{w.verb} {w.done} · </span>}<b className={w.left ? 'text-amber-700' : 'text-emerald-700'}>{w.left ? `باقي ${w.left}` : 'خالص'}</b></li>)}
                        </ul>
                      )}
                    </td>
                    <td className="p-2">
                      {e.open === 0 ? <span className="text-slate-400">0</span> : (
                        <>
                          <button type="button" onClick={() => setWhy(why === e.id ? null : e.id ?? null)} className="font-bold text-brand-700 underline">{e.open} · ليش؟</button>
                          {why === e.id && (
                            <ul className="mt-1 space-y-1 text-xs text-slate-600">
                              {e.items.map((it, i) => <li key={i}>{it.escalated ? '⬆️ ' : '• '}{it.summary}</li>)}
                            </ul>
                          )}
                        </>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      )}
    </div>
  )
}
