import { useEffect, useState } from 'react'
import { api } from '../api'
import MatrixNote from './MatrixNote'

// ═══ «لهجة ماتركس» — العبارات العراقية الي تعلّمها من فويسات الموظفين ═══
// الصوت تحوّل لنص بسيرفرنا (ما طلع)، والعبارات تنضاف لتوجيهات ماتركس حتى
// يحچي أقرب للهجة الموظفين.
type Dialect = Awaited<ReturnType<typeof api.getMatrixDialect>>

export default function MatrixDialect() {
  const [d, setD] = useState<Dialect | null>(null)
  useEffect(() => { void api.getMatrixDialect().then(setD).catch(() => {}) }, [])
  if (!d) return null
  return (
    <details dir="rtl" className="rounded-2xl border border-slate-200 bg-white p-3 text-slate-800">
      <summary className="cursor-pointer text-sm font-extrabold text-[#0f2040]">📚 لهجة ماتركس <span className="text-[11px] font-normal text-slate-500">{d.phrases.length} عبارة من {d.analyzed} فويس{d.pending ? ` · ${d.pending} ينتظر` : ''}</span></summary>
      <MatrixNote className="my-2">
        {!d.enabled ? 'تحويل الصوت لنص بعده ما شغّال بالسيرفر — الفويسات محفوظة وتنحلّل أول ما يشتغل.'
          : d.phrases.length === 0 ? 'بعد ما سمعت عبارات كافية. كل ما يسجّلون فويسات أكثر، أتعلّم لهجتهم أكثر.'
            : 'هاي العبارات سمعتها من فويسات الموظفين، وصرت أستعملها بتوجيهاتي.'}
      </MatrixNote>
      {d.phrases.length > 0 && (
        <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
          {d.phrases.map((p) => (
            <div key={p.phrase} className="rounded-xl border border-slate-200 bg-slate-50 p-2 text-xs">
              <b className="text-sm">«{p.phrase}»</b> <span className="text-slate-400">×{p.uses}</span>
              <p className="text-slate-600">{p.meaning}</p>
            </div>
          ))}
        </div>
      )}
    </details>
  )
}
