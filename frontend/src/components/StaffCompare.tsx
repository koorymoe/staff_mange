import { type ReactNode } from 'react'
import { type StaffScore } from '../api'
import PctWhy from '../ui/PctWhy'

// ═══ مقارنة الموظفين — قرار (ع) 10-10 ═══
// «ماتركس يسوي مقارنة حسب التقييم والإحصائيات والاعتمادية… كأنو نفرز الأسوأ
// من الأفضل». كل الأرقام من لوحة التقييم نفسها (ماتركس ٦٠٪ + البشر ٤٠٪،
// والاعتمادية رقم منفصل) — المقارنة ما تخترع رقم جديد، ترتّب وتقارن بس.

const SRC: Record<string, string> = { BOOKING: 'الحجوزات', DAY: 'الحضور', TASK: 'المهام', PROJECT: 'المشاريع', PROCUREMENT: 'المخازن', MONITOR: 'التدقيق', DUTY: 'واجب التقييم', REPORT: 'تقرير الإنجاز' }
const pct = (v: number | null | undefined) => (v == null ? '—' : `${Math.round(v)}%`)
const tone = (v: number | null | undefined) => (v == null ? '#94a3b8' : v >= 80 ? '#059669' : v >= 60 ? '#d97706' : '#dc2626')
const srcPct = (s: StaffScore, k: string) => {
  const v = s.bySource?.[k]
  return v && v[1] > 0 ? (v[0] / v[1]) * 100 : null
}

// «كل نسبة لازم الها سبب» — شرح كل رقم بلوحة المقارنة من نفس بيانات الموظف.
type WhyKind = 'final' | 'matrix' | 'human' | 'rel' | { src: string } | { part: string }
const r1 = (v: number) => Math.round(v * 10) / 10
const SCORE_SRC = <>نقاط ماتركس المسجلة للموظف بهذا الشهر (كل نقطة إلها «مكسوب» و«أقصى»)، وتقييمات البشر الي انطت إله بنفس الشهر من المقيّمين المسموحين لمجموعته.</>
function staffWhy(s: StaffScore, kind: WhyKind, v: number | null): { title: string; formula: ReactNode; meaning: ReactNode; source: ReactNode } {
  if (kind === 'final') return {
    title: `التقييم النهائي — ${s.name}`,
    formula: s.matrixPct != null && s.humanPct != null
      ? <>٦٠٪ × ماتركس {r1(s.matrixPct)}٪ + ٤٠٪ × البشر {r1(s.humanPct)}٪ = <b>{v != null ? r1(v) : '—'}٪</b></>
      : s.matrixPct != null ? <>ماكو تقييم بشري بعد، فالنهائي = نسبة ماتركس لحالها ({r1(s.matrixPct)}٪).</>
      : s.humanPct != null ? <>ماكو نقاط ماتركس، فالنهائي = تقييم البشر لحاله ({r1(s.humanPct)}٪).</> : <>ماكو بيانات.</>,
    meaning: <>الرقم الي ينرتب عليه الموظف. ٨٠٪ وفوق زين، ٦٠–٨٠ وسط، تحت ٦٠ ضعيف. يرتفع لمن يكسب نقاط ماتركس أكثر وياخذ تقييمات بشر أعلى.</>,
    source: SCORE_SRC,
  }
  if (kind === 'matrix') return {
    title: `ماتركس — ${s.name}`,
    formula: s.max > 0 ? <>{s.earned} نقطة مكسوبة ÷ {s.max} نقطة أقصى × ١٠٠ = <b>{v != null ? r1(v) : '—'}٪</b></> : <>ماكو نقاط ماتركس بهذا الشهر.</>,
    meaning: <>شكد التزم بالي مطلوب منه بوقته (حضور، مهام، حجوزات، مراحل…). كل بند يفوته أو يتأخر بيه ينقص من المكسوب. وزنه ٦٠٪ من النهائي.</>,
    source: <>نقاط ماتركس المسجلة تلقائياً للموظف بهذا الشهر.</>,
  }
  if (kind === 'human') return {
    title: `تقييم البشر — ${s.name}`,
    formula: s.humanAvg != null ? <>معدل {s.humanCount} تقييم = {r1(s.humanAvg)} من ٥ × ٢٠ = <b>{v != null ? r1(v) : '—'}٪</b></> : <>ماكو تقييم بشري بعد.</>,
    meaning: <>رأي المسؤولين والمقيّمين بيه (كل تقييم من ١ لـ٥ نجوم). وزنه ٤٠٪ من النهائي.</>,
    source: <>تقييمات البشر للموظف بهذا الشهر — بس من المقيّمين المسموحين لمجموعته.</>,
  }
  if (kind === 'rel') return {
    title: `الاعتمادية — ${s.name}`,
    formula: s.reliabilityParts.length ? <>معدل {s.reliabilityParts.length} بند: {s.reliabilityParts.map((p) => `${p.label} ${Math.round(p.pct)}٪`).join(' + ')} ÷ {s.reliabilityParts.length} = <b>{v != null ? r1(v) : '—'}٪</b></> : <>ماكو بيانات كافية.</>,
    meaning: <>رقم منفصل ما يدخل بالنهائي: يگول شكد تكدر تعتمد عليه (حضور، جرد، ورق، مهام، شكاوى، إجازات، مشاكل، عدّة، KPI). كل شكوى تنزّل ٢٥، وكل يوم إجازة فوق اليومين بالشهر ينزّل ١٠.</>,
    source: <>نقاط ماتركس (الحضور، الجرد، الورق، المهام) ووقائع الشهر: الشكاوى، الإجازات، المشاكل الوظيفية، ضياع العدّة، نقاط الـKPI.</>,
  }
  if ('src' in kind) {
    const b = s.bySource?.[kind.src]
    return {
      title: `نقاط ${SRC[kind.src]} — ${s.name}`,
      formula: b && b[1] > 0 ? <>{b[0]} مكسوبة ÷ {b[1]} أقصى × ١٠٠ = <b>{v != null ? r1(v) : '—'}٪</b></> : <>ماكو نقاط من هذا المصدر.</>,
      meaning: <>جزء ماتركس الي جاي من {SRC[kind.src]} بس — يبين وين قوي ووين ضعيف.</>,
      source: <>نقاط ماتركس المسجلة للموظف بهذا الشهر من مصدر «{SRC[kind.src]}».</>,
    }
  }
  const p = s.reliabilityParts.find((x) => x.key === kind.part)
  return {
    title: `${p?.label ?? 'بند الاعتمادية'} — ${s.name}`,
    formula: p ? <>{p.detail} ← <b>{Math.round(p.pct)}٪</b></> : <>ماكو بيانات.</>,
    meaning: <>بند من بنود الاعتمادية. بنود الحضور/الجرد/الورق/المهام = المكسوب ÷ الأقصى؛ والباقي يبدي من ١٠٠ وينقص لكل واقعة (شكوى/مشكلة/ضياع عدّة ٢٥، نقطة KPI ٢٠، يوم إجازة فوق اليومين ١٠).</>,
    source: <>نقاط ماتركس ووقائع الموظف بهذا الشهر.</>,
  }
}
function StaffPct({ s, kind, v, color }: { s: StaffScore; kind: WhyKind; v: number | null | undefined; color?: string }) {
  if (v == null) return <>—</>
  const w = staffWhy(s, kind, v)
  return <span style={{ color }}><PctWhy title={w.title} value={Math.round(v)} formula={w.formula} meaning={w.meaning} source={w.source} /></span>
}

function Bar({ v, s, kind }: { v: number | null; s: StaffScore; kind: WhyKind }) {
  return (
    <div className="flex items-center gap-2">
      <div className="h-2 flex-1 overflow-hidden rounded-full bg-slate-100">
        <div className="h-full rounded-full" style={{ width: `${Math.max(0, Math.min(100, v ?? 0))}%`, background: tone(v) }} />
      </div>
      <b className="w-10 text-left text-xs" style={{ color: tone(v) }}><StaffPct s={s} kind={kind} v={v} /></b>
    </div>
  )
}

// رقم ماتركس لحاله (ماكو تقييم بشري) ناقص — ينعلَّم واضح حتى ما ينقرا حكماً كاملاً.
export function PendingHumanBadge() {
  return <span title="الرقم من ماتركس بس — ما قيّمه أحد من البشر بعد"
    className="mr-1 whitespace-nowrap rounded-full border border-amber-300 bg-amber-100 px-2 py-0.5 text-[11px] font-bold text-amber-900">⏳ بانتظار تقييم بشري</span>
}

// ── ترتيب المجموعة من الأفضل للأسوأ، وماتركس يكتب الخلاصة ──
export function GroupRanking({ staff }: { staff: StaffScore[] }) {
  const rated = staff.filter((s) => s.final != null).sort((a, b) => (b.final ?? 0) - (a.final ?? 0))
  const unrated = staff.filter((s) => s.final == null)
  if (rated.length === 0) return <p className="p-6 text-center text-sm text-slate-400">ماكو تقييمات بهالمجموعة بعد.</p>
  // الأفضل والأضعف والمعدّل من الي عندهم تقييم بشري بس — رقم ماتركس لحاله ما يحكم.
  const judged = rated.filter((s) => !s.noHuman)
  const pending = rated.length - judged.length
  const best = judged[0]
  const worst = judged[judged.length - 1]
  const avg = judged.length ? judged.reduce((a, s) => a + (s.final ?? 0), 0) / judged.length : 0
  const weakest = worst?.topLosses?.[0]
  return (
    <div className="space-y-3">
      <div className="rounded-2xl border border-violet-200 bg-violet-50/60 p-3 text-sm leading-relaxed text-violet-950">
        <b>🤖 ماتركس يقارن:</b>{' '}
        {best ? <>
          الأفضل <b>{best.name}</b> (<StaffPct s={best} kind="final" v={best.final} />)
          {judged.length > 1 && <> والأضعف <b>{worst.name}</b> (<StaffPct s={worst} kind="final" v={worst.final} />) — الفرق {Math.round((best.final ?? 0) - (worst.final ?? 0))} نقطة.</>}
          {' '}معدّل المجموعة <PctWhy title="معدّل المجموعة" value={Math.round(avg)} formula={<>مجموع التقييم النهائي لـ{judged.length} موظف عندهم تقييم بشري ({r1(judged.reduce((a, s) => a + (s.final ?? 0), 0))}) ÷ {judged.length} = <b>{r1(avg)}٪</b></>} meaning={<>المعدل الي نقارن بيه: الي فوقه «فوق المعدّل» والي تحته «تحت المعدّل». الموظفين الي بس عندهم رقم ماتركس ما دخلوا.</>} source={SCORE_SRC}><ul className="text-xs">{judged.map((s) => <li key={s.id}>{s.name}: {Math.round(s.final ?? 0)}٪</li>)}</ul></PctWhy>، و{judged.filter((s) => (s.final ?? 0) < avg).length} تحت المعدّل.
          {weakest && judged.length > 1 && <> أكثر شي نزّل {worst.name}: «{weakest.title}» ({weakest.count} مرة) — {weakest.advice}</>}
        </> : <>ماكو أحد بهالمجموعة قيّمه إنسان بعد، فما أحكم منو الأفضل والأضعف.</>}
        {pending > 0 && <> {pending} موظف بانتظار تقييم بشري — ما دخلوا بالمقارنة.</>}
      </div>
      <div className="overflow-x-auto rounded-2xl border border-slate-200 bg-white">
        <table className="w-full min-w-[640px] text-right text-sm">
          <thead className="bg-slate-50 text-xs text-slate-500">
            <tr>
              <th className="p-2">#</th><th className="p-2">الموظف</th>
              <th className="p-2 w-[18%]">النهائي</th><th className="p-2 w-[15%]">ماتركس</th>
              <th className="p-2 w-[15%]">البشر</th><th className="p-2 w-[15%]">الاعتمادية</th><th className="p-2">الحكم</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {rated.map((s, i) => {
              const verdict = s.noHuman ? null
                : s.id === best?.id ? ['🥇 الأفضل', 'bg-emerald-100 text-emerald-800']
                : s.id === worst?.id && judged.length > 2 ? ['⚠️ الأضعف', 'bg-red-100 text-red-700']
                : (s.final ?? 0) >= avg ? ['فوق المعدّل', 'bg-sky-50 text-sky-700'] : ['تحت المعدّل', 'bg-amber-50 text-amber-700']
              return (
                <tr key={s.id}>
                  <td className="p-2 text-xs text-slate-400">{i + 1}</td>
                  <td className="p-2 font-bold">{s.name}{s.noHuman && <PendingHumanBadge />}</td>
                  <td className="p-2"><Bar v={s.final} s={s} kind="final" /></td>
                  <td className="p-2"><Bar v={s.matrixPct} s={s} kind="matrix" /></td>
                  <td className="p-2"><Bar v={s.humanPct} s={s} kind="human" /></td>
                  <td className="p-2"><Bar v={s.reliability} s={s} kind="rel" /></td>
                  <td className="p-2">{verdict ? <span className={`whitespace-nowrap rounded-full px-2 py-0.5 text-xs font-bold ${verdict[1]}`}>{verdict[0]}</span> : <PendingHumanBadge />}</td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
      {unrated.length > 0 && <p className="text-xs text-slate-400">بلا تقييم بعد: {unrated.map((s) => s.name).join('، ')}</p>}
    </div>
  )
}

// ── موظف قبال موظف ──
export function PairCompare({ a, b, peers, onPick, onClose }: {
  a: StaffScore; b: StaffScore | null; peers: StaffScore[]; onPick: (id: string) => void; onClose: () => void
}) {
  type Row = { label: string; kind: WhyKind; va: number | null; vb: number | null; text?: [string, string] }
  const rows: Row[] = b ? [
    { label: 'التقييم النهائي', kind: 'final', va: a.final, vb: b.final },
    { label: 'ماتركس (٦٠٪)', kind: 'matrix', va: a.matrixPct, vb: b.matrixPct },
    { label: 'تقييم البشر (٤٠٪)', kind: 'human', va: a.humanPct, vb: b.humanPct },
    { label: 'الاعتمادية', kind: 'rel', va: a.reliability, vb: b.reliability },
    ...Object.keys(SRC).filter((k) => a.bySource?.[k] || b.bySource?.[k])
      .map((k): Row => ({ label: `نقاط ${SRC[k]}`, kind: { src: k }, va: srcPct(a, k), vb: srcPct(b, k) })),
    ...a.reliabilityParts.map((p) => {
      const q = b.reliabilityParts.find((x) => x.key === p.key)
      return { label: `🛡️ ${p.label}`, kind: { part: p.key }, va: p.pct, vb: q?.pct ?? null, text: [p.detail, q?.detail ?? '—'] as [string, string] }
    }),
  ] : []
  const wins = rows.reduce((acc, r) => {
    if (r.va == null || r.vb == null || Math.abs(r.va - r.vb) < 1) return acc
    return r.va > r.vb ? { ...acc, a: acc.a + 1 } : { ...acc, b: acc.b + 1 }
  }, { a: 0, b: 0 })
  const gaps = rows.filter((r) => r.va != null && r.vb != null).map((r) => ({ ...r, gap: (r.va ?? 0) - (r.vb ?? 0) }))
    .sort((x, y) => Math.abs(y.gap) - Math.abs(x.gap))
  const top = gaps[0]
  const winner = b && a.final != null && b.final != null ? (a.final >= b.final ? a : b) : null

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between gap-2">
        <p className="text-base font-extrabold">⚖️ مقارنة {a.name}</p>
        <button type="button" onClick={onClose} className="text-xs font-bold text-slate-500 hover:text-slate-800">✕ سكّر</button>
      </div>
      <select value={b?.id ?? ''} onChange={(e) => onPick(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm">
        <option value="">— اختار موظف من نفس الدور —</option>
        {peers.filter((p) => p.id !== a.id).map((p) => <option key={p.id} value={p.id}>{p.name} ({pct(p.final)})</option>)}
      </select>
      {b && (
        <>
          <div className="rounded-xl border border-violet-200 bg-violet-50/60 p-3 text-sm leading-relaxed text-violet-950">
            <b>🤖 ماتركس:</b>{' '}
            {winner ? <>الأفضل <b>{winner.name}</b> بفرق {Math.abs(Math.round((a.final ?? 0) - (b.final ?? 0)))} نقطة بالنهائي. </> : 'واحد منهم بلا تقييم نهائي بعد. '}
            {a.name} متفوق بـ{wins.a} بند، و{b.name} بـ{wins.b}.
            {top && Math.abs(top.gap) >= 1 && <> أكبر فرق بـ«{top.label}»: {top.gap > 0 ? a.name : b.name} أعلى بـ{Math.abs(Math.round(top.gap))} نقطة.</>}
          </div>
          <div className="overflow-hidden rounded-xl border border-slate-200">
            <table className="w-full text-right text-sm">
              <thead className="bg-slate-50 text-xs text-slate-500">
                <tr><th className="p-2">البند</th><th className="p-2">{a.name}</th><th className="p-2">{b.name}</th></tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {rows.map((r) => {
                  const aw = r.va != null && r.vb != null && r.va - r.vb >= 1
                  const bw = r.va != null && r.vb != null && r.vb - r.va >= 1
                  return (
                    <tr key={r.label}>
                      <td className="p-2 text-xs text-slate-600">{r.label}</td>
                      <td className={`p-2 ${aw ? 'bg-emerald-50' : ''}`}>
                        <b style={{ color: tone(r.va) }}><StaffPct s={a} kind={r.kind} v={r.va} /></b>{aw && ' ✓'}
                        {r.text && <span className="block text-[10px] text-slate-400">{r.text[0]}</span>}
                      </td>
                      <td className={`p-2 ${bw ? 'bg-emerald-50' : ''}`}>
                        <b style={{ color: tone(r.vb) }}><StaffPct s={b} kind={r.kind} v={r.vb} /></b>{bw && ' ✓'}
                        {r.text && <span className="block text-[10px] text-slate-400">{r.text[1]}</span>}
                      </td>
                    </tr>
                  )
                })}
                <tr>
                  <td className="p-2 text-xs text-slate-600">عدد تقييمات البشر</td>
                  <td className="p-2 text-xs">{a.humanCount}</td><td className="p-2 text-xs">{b.humanCount}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <div className="grid gap-2 sm:grid-cols-2">
            {[a, b].map((s) => (
              <div key={s.id} className="rounded-xl bg-slate-50 p-2">
                <p className="mb-1 text-xs font-extrabold text-slate-700">شنو نزّل {s.name}</p>
                {s.topLosses?.length ? s.topLosses.slice(0, 3).map((l) => (
                  <p key={l.rule} className="text-[11px] text-slate-600">• {l.title} ({l.count} مرة)</p>
                )) : <p className="text-[11px] text-slate-400">ماكو خسارات تذكر.</p>}
              </div>
            ))}
          </div>
        </>
      )}
    </div>
  )
}
