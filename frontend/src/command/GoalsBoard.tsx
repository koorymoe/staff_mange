import { useEffect, useState } from 'react'
import { api, type CommandGoal } from '../api'

// ═══ مركز القيادة: الأهداف والخطة (قرار (ع) 10-10) ═══
// «ماريد البيانات الموجودة بالنظام الأول تتكرر» — هنا شي جديد: أهداف المالك
// للفصل والسنة، ولكل هدف مراحل بموعد ومسؤول. التقدّم = المراحل الي تمّت.

const STATUS: Record<CommandGoal['status'], string> = { ACTIVE: 'شغّال', DONE: 'تحقّق ✓', PAUSED: 'متوقف', DROPPED: 'ملغي' }
const CATS = ['النمو', 'المال', 'الخدمات', 'الكوادر', 'التوسّع', 'عام']
const today = () => new Date(Date.now() + 3 * 3600_000).toISOString().slice(0, 10)
const field: React.CSSProperties = {
  background: 'rgba(255,255,255,0.06)', border: '1px solid rgba(255,255,255,0.14)', borderRadius: 10,
  color: '#fff', padding: '8px 10px', fontSize: 13, fontFamily: 'inherit', outline: 'none',
}
const btn = (bg: string): React.CSSProperties => ({
  background: bg, color: '#fff', border: 'none', borderRadius: 10, padding: '8px 14px',
  fontSize: 12.5, fontWeight: 800, cursor: 'pointer', fontFamily: 'inherit',
})

function daysLeft(d: string | null): { text: string; late: boolean } | null {
  if (!d) return null
  const n = Math.round((new Date(d).getTime() - new Date(today()).getTime()) / 86400000)
  return n < 0 ? { text: `متأخر ${-n} يوم`, late: true } : { text: n === 0 ? 'اليوم' : `باقي ${n} يوم`, late: false }
}

export default function GoalsBoard() {
  const [goals, setGoals] = useState<CommandGoal[] | null>(null)
  const [err, setErr] = useState<string | null>(null)
  const [adding, setAdding] = useState(false)
  const [form, setForm] = useState({ title: '', category: 'النمو', ownerLabel: '', targetDate: '', description: '' })
  const [stepDraft, setStepDraft] = useState<Record<string, { title: string; dueDate: string }>>({})

  const run = (p: Promise<CommandGoal[]>) => p.then((g) => { setGoals(g); setErr(null) }).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر'))
  useEffect(() => { run(api.commandGoals()) }, [])

  const add = () => {
    if (!form.title.trim()) return
    run(api.commandAddGoal(form)).then(() => { setAdding(false); setForm({ title: '', category: 'النمو', ownerLabel: '', targetDate: '', description: '' }) })
  }

  const active = goals?.filter((g) => g.status === 'ACTIVE') ?? []
  const steps = active.flatMap((g) => g.steps)
  const lateSteps = steps.filter((s) => !s.doneAt && s.dueDate && s.dueDate < today()).length

  return (
    <div>
      <div className="cmd-tiles">
        <div className="cmd-tile"><div className="k">أهداف شغّالة</div><div className="v">{goals ? active.length : '—'}</div><div className="h">من {goals?.length ?? 0}</div></div>
        <div className="cmd-tile"><div className="k">مراحل تمّت</div><div className="v">{goals ? steps.filter((s) => s.doneAt).length : '—'}</div><div className="h">من {steps.length}</div></div>
        <div className="cmd-tile"><div className="k">مراحل متأخرة</div><div className="v" style={{ color: lateSteps ? '#ff8a8a' : undefined }}>{goals ? lateSteps : '—'}</div><div className="h">موعدها فات</div></div>
      </div>

      {err && <p style={{ color: '#ff8a8a', fontSize: 13, fontWeight: 700 }}>{err}</p>}

      <div style={{ margin: '16px 0' }}>
        {!adding ? <button style={btn('#3b6fd8')} onClick={() => setAdding(true)}>＋ هدف جديد</button> : (
          <div className="cmd-panel" style={{ display: 'grid', gap: 8 }}>
            <input style={field} placeholder="الهدف (مثلاً: فتح خدمة الطاقة بأربيل)" value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} />
            <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
              <select style={field} value={form.category} onChange={(e) => setForm({ ...form, category: e.target.value })}>
                {CATS.map((c) => <option key={c} style={{ color: '#000' }}>{c}</option>)}
              </select>
              <input style={{ ...field, flex: 1 }} placeholder="المسؤول عنه" value={form.ownerLabel} onChange={(e) => setForm({ ...form, ownerLabel: e.target.value })} />
              <input style={field} type="date" value={form.targetDate} onChange={(e) => setForm({ ...form, targetDate: e.target.value })} />
            </div>
            <textarea style={{ ...field, minHeight: 60 }} placeholder="ليش هذا الهدف؟ وشلون نعرف تحقق؟" value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
            <div style={{ display: 'flex', gap: 8 }}>
              <button style={btn('#2e9d63')} onClick={add}>حفظ</button>
              <button style={btn('rgba(255,255,255,0.12)')} onClick={() => setAdding(false)}>إلغاء</button>
            </div>
          </div>
        )}
      </div>

      {goals && goals.length === 0 && <p className="cmd-pending">ماكو أهداف بعد — ابدي بأول هدف للفصل.</p>}

      <div className="cmd-grid2">
        {goals?.map((g) => {
          const done = g.steps.filter((s) => s.doneAt).length
          const pct = g.steps.length ? Math.round((done / g.steps.length) * 100) : g.status === 'DONE' ? 100 : 0
          const dl = g.status === 'ACTIVE' ? daysLeft(g.targetDate) : null
          const d = stepDraft[g.id] ?? { title: '', dueDate: '' }
          return (
            <div className="cmd-panel" key={g.id} style={{ opacity: g.status === 'ACTIVE' ? 1 : 0.6 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', gap: 8, alignItems: 'flex-start' }}>
                <div>
                  <span className="cmd-label">{g.category}{g.ownerLabel ? ` · ${g.ownerLabel}` : ''}</span>
                  <h3 style={{ marginTop: 4 }}>{g.title}</h3>
                </div>
                <select style={{ ...field, padding: '4px 6px', fontSize: 11.5 }} value={g.status}
                  onChange={(e) => run(api.commandUpdateGoal(g.id, { ...g, status: e.target.value as CommandGoal['status'] }))}>
                  {Object.entries(STATUS).map(([k, v]) => <option key={k} value={k} style={{ color: '#000' }}>{v}</option>)}
                </select>
              </div>
              {g.description && <p className="sub">{g.description}</p>}
              <div style={{ display: 'flex', alignItems: 'center', gap: 10, margin: '8px 0' }}>
                <div style={{ flex: 1, height: 8, borderRadius: 8, background: 'rgba(255,255,255,0.1)', overflow: 'hidden' }}>
                  <div style={{ width: `${pct}%`, height: '100%', background: 'linear-gradient(90deg,#3b6fd8,#5fd3a0)' }} />
                </div>
                <b style={{ fontSize: 13 }}>{pct}%</b>
              </div>
              <p style={{ margin: '0 0 8px', fontSize: 11.5, color: dl?.late ? '#ff8a8a' : 'rgba(255,255,255,0.55)' }}>
                {g.targetDate ? `الموعد ${g.targetDate}` : 'بلا موعد'}{dl ? ` — ${dl.text}` : ''}
              </p>
              <div style={{ display: 'grid', gap: 4 }}>
                {g.steps.map((s) => {
                  const late = !s.doneAt && s.dueDate && s.dueDate < today()
                  return (
                    <div key={s.id} style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 13 }}>
                      <input type="checkbox" checked={!!s.doneAt} onChange={() => run(api.commandToggleStep(s.id))} />
                      <span style={{ flex: 1, textDecoration: s.doneAt ? 'line-through' : 'none', opacity: s.doneAt ? 0.55 : 1 }}>{s.title}</span>
                      {s.dueDate && <span style={{ fontSize: 11, color: late ? '#ff8a8a' : 'rgba(255,255,255,0.45)' }}>{s.dueDate}</span>}
                      <button title="حذف" onClick={() => { if (window.confirm('تحذف هاي المرحلة؟')) run(api.commandDeleteStep(s.id)) }}
                        style={{ background: 'none', border: 'none', color: 'rgba(255,255,255,0.35)', cursor: 'pointer' }}>✕</button>
                    </div>
                  )
                })}
              </div>
              <div style={{ display: 'flex', gap: 6, marginTop: 8 }}>
                <input style={{ ...field, flex: 1, padding: '6px 8px' }} placeholder="＋ مرحلة" value={d.title}
                  onChange={(e) => setStepDraft({ ...stepDraft, [g.id]: { ...d, title: e.target.value } })} />
                <input style={{ ...field, padding: '6px 8px' }} type="date" value={d.dueDate}
                  onChange={(e) => setStepDraft({ ...stepDraft, [g.id]: { ...d, dueDate: e.target.value } })} />
                <button style={btn('#3b6fd8')} disabled={!d.title.trim()}
                  onClick={() => run(api.commandAddStep(g.id, d.title, d.dueDate)).then(() => setStepDraft({ ...stepDraft, [g.id]: { title: '', dueDate: '' } }))}>أضف</button>
              </div>
              <button onClick={() => { if (window.confirm(`تحذف الهدف «${g.title}» ومراحله؟`)) run(api.commandDeleteGoal(g.id)) }}
                style={{ marginTop: 10, background: 'none', border: 'none', color: 'rgba(255,138,138,0.7)', fontSize: 11.5, cursor: 'pointer', fontFamily: 'inherit' }}>حذف الهدف</button>
            </div>
          )
        })}
      </div>
    </div>
  )
}
