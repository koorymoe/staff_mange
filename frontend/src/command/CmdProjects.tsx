import { useCallback, useEffect, useRef, useState } from 'react'
import { apiRequest } from '../api'
import { parseCommand, fmtIQD, STAGE_LABEL, type Parsed, type ProjectFields, type Stage } from './parseCommand'

// ═══ 🏗️ المشاريع والمقاولات — مركز القيادة (دازاي 11-01) ═══
// الإدخال سطر أوامر مو فورم: تكتب جملة، تشوف شنو فهم النظام، Enter يحفظ.
// معزول عن نظام الشركة — يقرا ويكتب «/api/command/projects» بس.

type CmdProject = ProjectFields & {
  id: string; name: string; location: string; owner: string; contractIQD: number
  stage: Stage; progress: number; startDate?: string; updatedAt: string
}
type Line = { who: 'me' | 'sys' | 'err' | 'ok'; text: string }

const STAGES: Stage[] = ['TENDER', 'SIGNED', 'EXECUTION', 'HANDOVER', 'CLOSED']
const HELP = [
  'مشروع <الاسم>، الموقع …، المالك …، العقد 450 مليون، المدة 18 شهر، يبدي 2026-12-01',
  '<اسم المشروع> مرحلة تنفيذ 35%   ·   <اسم المشروع> انجاز 40',
  'عرض   ·   عرض تنفيذ   ·   مساعدة',
  'المراحل: عطاء ← توقيع ← تنفيذ ← تسليم ← مغلق',
]

function useCmdProjects() {
  const [rows, setRows] = useState<CmdProject[] | null>(null)
  const [error, setError] = useState('')
  const load = useCallback(() => {
    apiRequest<CmdProject[]>('/command/projects').then((r) => { setRows(r); setError('') }).catch((e: Error) => setError(e.message))
  }, [])
  useEffect(() => { load() }, [load])
  return { rows, error, load }
}

function describe(f: ProjectFields) {
  const out: [string, string][] = []
  if (f.name) out.push(['المشروع', f.name])
  if (f.location) out.push(['الموقع', f.location])
  if (f.owner) out.push(['المالك', f.owner])
  if (f.contractIQD) out.push(['العقد', `${fmtIQD(f.contractIQD)} دينار`])
  if (f.durationMonths) out.push(['المدة', `${f.durationMonths} شهر`])
  if (f.startDate) out.push(['البدء', f.startDate])
  if (f.stage) out.push(['المرحلة', STAGE_LABEL[f.stage]])
  if (f.progress !== undefined) out.push(['الإنجاز', `${f.progress}%`])
  return out
}

export function CmdTerminal() {
  const { rows, load } = useCmdProjects()
  const [lines, setLines] = useState<Line[]>([{ who: 'sys', text: 'مرحبا. اكتب أمر — أو «مساعدة».' }])
  const [input, setInput] = useState('')
  const [pending, setPending] = useState<{ cmd: string; parsed: Parsed; id?: string } | null>(null)
  const [history, setHistory] = useState<string[]>([])
  const [hIdx, setHIdx] = useState(-1)
  const [busy, setBusy] = useState(false)
  const endRef = useRef<HTMLDivElement>(null)
  const inRef = useRef<HTMLInputElement>(null)
  useEffect(() => { endRef.current?.scrollIntoView({ block: 'end' }) }, [lines, pending])
  const say = (l: Line | Line[]) => setLines((x) => [...x, ...(Array.isArray(l) ? l : [l])].slice(-200))

  const find = (name: string) => {
    const n = name.trim().toLowerCase()
    const list = rows ?? []
    return list.find((p) => p.name.toLowerCase() === n) ?? list.filter((p) => p.name.toLowerCase().includes(n))
  }

  const submit = () => {
    const cmd = input.trim()
    if (!cmd || busy) return
    setHistory((h) => [cmd, ...h].slice(0, 50)); setHIdx(-1); setInput('')
    say({ who: 'me', text: cmd })
    const p = parseCommand(cmd)
    if (p.kind === 'help') return say(HELP.map((t) => ({ who: 'sys' as const, text: t })))
    if (p.kind === 'error') return say({ who: 'err', text: p.message })
    if (p.kind === 'list') {
      const list = (rows ?? []).filter((r) => !p.stage || r.stage === p.stage)
      if (!list.length) return say({ who: 'sys', text: 'ماكو مشاريع.' })
      return say(list.map((r) => ({ who: 'sys' as const, text: `▸ ${r.name} — ${STAGE_LABEL[r.stage]} ${r.progress}% — ${fmtIQD(r.contractIQD)}` })))
    }
    if (p.kind === 'update') {
      const hit = find(p.target)
      if (Array.isArray(hit)) {
        if (hit.length === 1) return setPending({ cmd, parsed: p, id: hit[0].id })
        return say({ who: 'err', text: hit.length ? `أكثر من مشروع يطابق: ${hit.map((h) => h.name).join('، ')}` : `ماكو مشروع اسمه «${p.target}»` })
      }
      return setPending({ cmd, parsed: p, id: hit.id })
    }
    setPending({ cmd, parsed: p })
  }

  const confirm = async () => {
    if (!pending || busy) return
    const { parsed, cmd, id } = pending
    setBusy(true)
    try {
      if (parsed.kind === 'create') {
        const r = await apiRequest<CmdProject>('/command/projects', { method: 'POST', body: JSON.stringify({ ...parsed.fields, command: cmd }) })
        say({ who: 'ok', text: `✓ انضاف «${r.name}»` })
      } else if (parsed.kind === 'update' && id) {
        const r = await apiRequest<CmdProject>(`/command/projects/${id}`, { method: 'PATCH', body: JSON.stringify({ ...parsed.fields, command: cmd }) })
        say({ who: 'ok', text: `✓ «${r.name}» — ${STAGE_LABEL[r.stage]} ${r.progress}%` })
      }
      setPending(null); load()
    } catch (e) {
      say({ who: 'err', text: e instanceof Error ? e.message : 'تعذر' })
    } finally { setBusy(false); inRef.current?.focus() }
  }

  const onKey = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (pending) {
      if (e.key === 'Enter') { e.preventDefault(); void confirm() }
      if (e.key === 'Escape') { e.preventDefault(); e.stopPropagation(); setPending(null); say({ who: 'sys', text: 'انلغى.' }) }
      return
    }
    if (e.key === 'Enter') { e.preventDefault(); submit() }
    else if (e.key === 'ArrowUp' && history.length) {
      e.preventDefault(); const i = Math.min(hIdx + 1, history.length - 1); setHIdx(i); setInput(history[i])
    } else if (e.key === 'ArrowDown') {
      e.preventDefault(); const i = hIdx - 1; setHIdx(Math.max(-1, i)); setInput(i < 0 ? '' : history[i])
    }
  }

  const target = pending?.id ? rows?.find((r) => r.id === pending.id) : undefined
  return (
    <div className="cmd-term" onClick={() => inRef.current?.focus()}>
      <div className="cmd-term-log">
        {lines.map((l, i) => (
          <div key={i} className={`cmd-term-line ${l.who}`}>{l.who === 'me' ? <span className="cmd-term-ps">❯</span> : null}{l.text}</div>
        ))}
        {pending && (pending.parsed.kind === 'create' || pending.parsed.kind === 'update') && (
          <div className="cmd-term-preview">
            <div className="cmd-label">{pending.parsed.kind === 'create' ? 'مشروع جديد — فهمت هيچ:' : `تحديث «${target?.name ?? ''}»:`}</div>
            {describe(pending.parsed.fields).map(([k, v]) => (
              <div key={k} className="cmd-term-kv"><span>{k}</span><b>{v}</b></div>
            ))}
            {pending.parsed.kind === 'create' && pending.parsed.missing.map((m) => (
              <div key={m} className="cmd-term-kv missing"><span>{m}</span><b>ناقص — تگدر تضيفه بعدين</b></div>
            ))}
            <div className="cmd-term-hint">Enter يحفظ · Esc يلغي</div>
          </div>
        )}
        <div ref={endRef} />
      </div>
      <div className="cmd-term-input">
        <span className="cmd-term-ps">❯</span>
        <input ref={inRef} autoFocus value={input} disabled={busy} onChange={(e) => setInput(e.target.value)} onKeyDown={onKey}
          placeholder={pending ? 'Enter يحفظ · Esc يلغي' : 'مشروع برج الكرادة، الموقع …، المالك …، العقد 450 مليون، المدة 18 شهر'} />
      </div>
    </div>
  )
}

export function CmdProjectCards() {
  const { rows, error } = useCmdProjects()
  if (error) return <p className="cmd-term-line err">{error}</p>
  if (!rows) return <p className="cmd-muted">...</p>
  const total = rows.reduce((s, r) => s + Number(r.contractIQD || 0), 0)
  return (
    <>
      <div className="cmd-tiles">
        <div className="cmd-tile"><div className="k">المشاريع</div><div className="v">{rows.length}</div><div className="h">كلها</div></div>
        <div className="cmd-tile"><div className="k">مجموع العقود</div><div className="v">{fmtIQD(total)}</div><div className="h">دينار</div></div>
        {STAGES.slice(0, 3).map((s) => (
          <div className="cmd-tile" key={s}><div className="k">{STAGE_LABEL[s]}</div><div className="v">{rows.filter((r) => r.stage === s).length}</div><div className="h">مشروع</div></div>
        ))}
      </div>
      {rows.length === 0 && <p style={{ color: 'var(--cmd-muted)' }}>ماكو مشاريع بعد — ضيف من «الطرفية».</p>}
      <div className="cmd-grid2">
        {rows.map((r) => (
          <div className="cmd-panel" key={r.id}>
            <h3>{r.name}</h3>
            <p className="sub">{[r.location, r.owner].filter(Boolean).join(' · ') || '—'}</p>
            <div className="cmd-term-kv"><span>العقد</span><b>{r.contractIQD ? `${fmtIQD(Number(r.contractIQD))} دينار` : '—'}</b></div>
            <div className="cmd-term-kv"><span>المدة</span><b>{r.durationMonths ? `${r.durationMonths} شهر` : '—'}{r.startDate ? ` · من ${r.startDate.slice(0, 10)}` : ''}</b></div>
            <div className="cmd-stages">
              {STAGES.map((s) => <span key={s} className={`cmd-stage${s === r.stage ? ' on' : ''}${STAGES.indexOf(s) < STAGES.indexOf(r.stage) ? ' past' : ''}`}>{STAGE_LABEL[s]}</span>)}
            </div>
            <div className="cmd-bar"><i style={{ width: `${r.progress}%` }} /></div>
            <div className="cmd-term-kv"><span>الإنجاز</span><b>{r.progress}%</b></div>
          </div>
        ))}
      </div>
    </>
  )
}
