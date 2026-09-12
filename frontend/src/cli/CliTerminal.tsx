// ═══ الترمنال ═══
//
// «واجهات الترمنال الـCLI أريد شي يبدو وكأنه حقيقي».
//
// ⚠️ ماكو `<input>` ظاهر: الجهاز الحقيقي ما بيه خانة نص لها إطار
// وزوايا مدوّرة — بيه سطر يمشي وراه مؤشر يلمع. نستقبل الحروف على
// عنصر مركّز ونرسم السطر بأنفسنا.
//
// ⚠️ الاتجاه **LTR إجباري** جوّا صفحة عربية RTL: أي أمر
// `switchport access vlan 10` ينعرض بـRTL يطلع مقلوباً — والسهم `^`
// تحت الغلط ينزاح لمحل ثاني. `dir="ltr"` هنا مو تفصيلة شكلية.
//
// ⚠️ `?` **ما تنكتب بالسطر**: بالجهاز الحقيقي تضغط `?` فتطلع
// المساعدة فوراً ويُعاد طبع سطرك تحتها مثل ما هو. هذا السلوك بالذات
// هو الي يخلّي الفني يحس إنه على جهاز.

import { useCallback, useEffect, useRef, useState } from 'react'
import type { CliGrammar } from './grammar'
import { complete, contextHelp, execute, newSession, prompt, type CliSession } from './engine'

interface Line {
  text: string
  kind: 'echo' | 'out' | 'err' | 'caret' | 'banner'
}

interface Props {
  grammar: CliGrammar
  initialState?: Record<string, unknown>
  /** ينندى بعد كل أمر — الصفحة تقيّم الحالة مو الحروف المكتوبة.
   *  ⚠️ النمط ينمرّر معها: «ادخل النمط المميّز» خطوة تدريبية حقيقية
   *  وما تنقاس بأي قيمة بالحالة — النمط مو إعداد، هو مكان تقف بيه. */
  onStateChange?: (state: Record<string, unknown>, lastCommand: string, mode: string) => void
  readOnly?: boolean
  heightClass?: string
}

export default function CliTerminal({
  grammar, initialState, onStateChange, readOnly, heightClass = 'h-[420px]',
}: Props) {
  const [session, setSession] = useState<CliSession>(() => newSession(grammar, initialState ?? {}))
  const [lines, setLines] = useState<Line[]>(() =>
    (grammar.banner ?? []).map((text) => ({ text, kind: 'banner' as const })),
  )
  const [input, setInput] = useState('')
  const [histIdx, setHistIdx] = useState<number | null>(null)
  const [focused, setFocused] = useState(false)
  const boxRef = useRef<HTMLDivElement>(null)
  const endRef = useRef<HTMLDivElement>(null)

  const push = useCallback((add: Line[]) => setLines((prev) => [...prev, ...add]), [])

  useEffect(() => { endRef.current?.scrollIntoView({ block: 'end' }) }, [lines])

  const submit = useCallback((line: string) => {
    const p = prompt(grammar, session)
    const res = execute(grammar, session, line)
    const add: Line[] = [{ text: p + line, kind: 'echo' }]
    if (res.caretAt !== undefined) {
      add.push({ text: ' '.repeat(p.length + res.caretAt) + '^', kind: 'caret' })
    }
    for (const o of res.output) {
      if (o === '') { add.push({ text: '', kind: 'out' }); continue }
      add.push({ text: o, kind: o.startsWith('%') ? 'err' : 'out' })
    }
    push(add)
    setSession(res.session)
    setInput('')
    setHistIdx(null)
    onStateChange?.(
      res.session.state as Record<string, unknown>,
      line.trim(),
      res.session.modeStack[res.session.modeStack.length - 1].mode,
    )
  }, [grammar, session, push, onStateChange])

  const onKeyDown = (e: React.KeyboardEvent) => {
    if (readOnly) return
    const k = e.key

    // ⚠️ `?` تُعالج **قبل** أي شي وما تنضاف للسطر — سلوك الجهاز.
    if (k === '?') {
      e.preventDefault()
      const p = prompt(grammar, session)
      push([
        { text: p + input + '?', kind: 'echo' },
        ...contextHelp(grammar, session, input).map((t) => ({ text: t, kind: 'out' as const })),
      ])
      return
    }
    if (k === 'Tab') {
      e.preventDefault()
      setInput((v) => complete(grammar, session, v))
      return
    }
    if (k === 'Enter') { e.preventDefault(); submit(input); return }
    if (k === 'Backspace') { e.preventDefault(); setInput((v) => v.slice(0, -1)); return }
    if (k === 'ArrowUp' || k === 'ArrowDown') {
      e.preventDefault()
      const h = session.history
      if (h.length === 0) return
      const next = k === 'ArrowUp'
        ? Math.min(h.length - 1, (histIdx === null ? -1 : histIdx) + 1)
        : Math.max(-1, (histIdx === null ? -1 : histIdx) - 1)
      setHistIdx(next < 0 ? null : next)
      setInput(next < 0 ? '' : h[h.length - 1 - next])
      return
    }
    // Ctrl-Z يرجّع لـEXEC مثل الجهاز، وCtrl-C يلغي السطر.
    if (e.ctrlKey && (k === 'z' || k === 'Z')) { e.preventDefault(); submit('end'); return }
    if (e.ctrlKey && (k === 'c' || k === 'C')) { e.preventDefault(); push([{ text: prompt(grammar, session) + input, kind: 'echo' }]); setInput(''); return }
    if (k.length === 1 && !e.ctrlKey && !e.metaKey) { e.preventDefault(); setInput((v) => v + k) }
  }

  const color: Record<Line['kind'], string> = {
    echo: 'text-slate-100',
    out: 'text-slate-300',
    err: 'text-amber-300',
    caret: 'text-amber-300',
    banner: 'text-slate-500',
  }

  return (
    <div
      ref={boxRef}
      dir="ltr"
      tabIndex={0}
      role="textbox"
      aria-label="جلسة سطر الأوامر"
      onKeyDown={onKeyDown}
      onFocus={() => setFocused(true)}
      onBlur={() => setFocused(false)}
      className={`${heightClass} overflow-y-auto rounded-xl bg-[#080c10] p-4 font-mono text-[13px] leading-[1.45] outline-none ring-1 transition ${
        focused ? 'ring-emerald-500/60' : 'ring-slate-700'
      }`}
    >
      {lines.map((l, i) => (
        <div key={i} className={`whitespace-pre-wrap break-all ${color[l.kind]}`}>{l.text || ' '}</div>
      ))}

      {!readOnly && (
        <div className="whitespace-pre-wrap break-all text-slate-100">
          {prompt(grammar, session)}{input}
          {/* المؤشر — مربّع يلمع مثل الـconsole */}
          <span className={`ml-px inline-block h-[1.05em] w-[0.55em] translate-y-[0.18em] bg-emerald-400 ${focused ? 'animate-pulse' : 'opacity-30'}`} />
        </div>
      )}

      {!focused && !readOnly && (
        <div className="mt-3 text-[11px] text-slate-500">اضغط داخل الشاشة حتى تكتب · `?` مساعدة · Tab إكمال · ↑ السجل</div>
      )}
      <div ref={endRef} />
    </div>
  )
}
