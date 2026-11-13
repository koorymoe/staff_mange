import { splitSentences } from './matrixNoteText'
import { Children, isValidElement, useState, type ReactNode } from 'react'

// تعليق أو توجيه من ماتركس — نفس اللون بكل مكان بالنظام (طلب (ع) 10-04).
//
// (دازاي 11-13) «الخط ضعيف وصغير، والشرح مطوّل — اريد ملخص ومرتب، مو بس حچي»:
//   - خط أكبر وأغمق (١٥px وزن ٥٠٠) — ما ينصغّر حتى لو المكان مرّر text-xs.
//   - إذا الرسالة نص طويل: أول جملة ملخّص بالخط العريض، والباقي نقاط مرتبة تحت
//     «التفاصيل» تنفتح بضغطة. الرسالة القصيرة تطلع مثل ما هي.

const SIZE_CLASS = /(^|\s)!?(text-(xs|sm|base|\[\d+px\])|leading-\S+)(?=\s|$)/g

function plainText(children: ReactNode): string | null {
  let out = ''
  let ok = true
  Children.forEach(children, (c) => {
    if (typeof c === 'string' || typeof c === 'number') out += String(c)
    else if (c == null || typeof c === 'boolean') return
    else if (isValidElement(c)) ok = false
  })
  return ok ? out.trim() : null
}

export default function MatrixNote({ children, className = '', tag = true }: { children: ReactNode; className?: string; tag?: boolean }) {
  const [open, setOpen] = useState(false)
  const cls = className.replace(SIZE_CLASS, ' ').trim()
  const text = plainText(children)
  const parts = text && text.length > 140 ? splitSentences(text) : null
  if (!parts || parts.length < 2) {
    return (
      <div className={`matrix-note ${cls}`}>
        {tag && <>🤖 <b className="mx-tag">ماتركس:</b> </>}
        {children}
      </div>
    )
  }
  const [head, ...rest] = parts
  return (
    <div className={`matrix-note ${cls}`}>
      <p className="mx-head">{tag && <>🤖 <b className="mx-tag">ماتركس:</b> </>}{head}</p>
      {open && (
        <ul className="mx-points">
          {rest.map((x, i) => <li key={i}>{x}</li>)}
        </ul>
      )}
      <button type="button" className="mx-more" onClick={() => setOpen(!open)} aria-expanded={open}>
        {open ? '▲ أخفي التفاصيل' : `▼ التفاصيل (${rest.length})`}
      </button>
    </div>
  )
}
