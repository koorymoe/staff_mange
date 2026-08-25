// ═══ لوح الرسم ═══
//
// «شي منظّم ومرتّب وقوي».
//
// ⚠️ كل شي بفضاء SVG واحد بإحداثيات اللوح، والتحويل بين إحداثيات
// الشاشة وإحداثيات اللوح يصير **بمحل واحد** (`toBoard`). أي حساب
// إحداثيات متناثر بالملف يعني قطعة تنط تحت الماوس لمن تتغيّر
// درجة التكبير — وهذا أكثر بگ متكرّر بمحرّرات الرسم.
//
// ⚠️ الربط **بالضغط على منفذ ثم منفذ** مو بالسحب: السحب يحتاج تتبّع
// مؤشر مستمر ويفشل بأول تمرير على عنصر ثاني، والضغط المزدوج يشتغل
// باللمس هم.

import { useCallback, useEffect, useRef, useState } from 'react'
import { linkParamsFor } from './cables'
import { PART_BY_ID } from './catalog'
import { Symbol } from './symbols'
import type { LabDoc, LabNode, SimResult } from './types'

interface Props {
  doc: LabDoc
  setDoc: (updater: (d: LabDoc) => LabDoc) => void
  result: SimResult | null
  selected: string | null
  setSelected: (id: string | null) => void
  /** القطعة المختارة من الكتالوگ حتى تنحط بالضغط على اللوح. */
  pendingPart: string | null
  onPlaced: () => void
  /** يتغيّر ← اللوح يضبط عرضه ليبيّن كل القطع. */
  fitSignal?: number
}

const GRID = 20
const snap = (v: number) => Math.round(v / GRID) * GRID

export default function Canvas({ doc, setDoc, result, selected, setSelected, pendingPart, onPlaced, fitSignal }: Props) {
  const svgRef = useRef<SVGSVGElement>(null)
  const [view, setView] = useState({ x: 0, y: 0, k: 1 })
  const [drag, setDrag] = useState<{ id: string; dx: number; dy: number } | null>(null)
  const [pan, setPan] = useState<{ sx: number; sy: number; vx: number; vy: number } | null>(null)
  const [armed, setArmed] = useState<{ node: string; port: string } | null>(null)
  const [hoverPort, setHoverPort] = useState<string | null>(null)
  const [reject, setReject] = useState<string | null>(null)

  /** ⚠️ التحويل الوحيد بين الشاشة واللوح — لا تحسبه بمحل ثاني. */
  const toBoard = useCallback((clientX: number, clientY: number) => {
    const r = svgRef.current?.getBoundingClientRect()
    if (!r) return { x: 0, y: 0 }
    return { x: (clientX - r.left - view.x) / view.k, y: (clientY - r.top - view.y) / view.k }
  }, [view])

  const nodeById = useCallback((id: string) => doc.nodes.find((n) => n.id === id), [doc.nodes])

  /** موقع منفذ بإحداثيات اللوح. */
  const portPos = useCallback((nodeId: string, portId: string) => {
    const n = nodeById(nodeId)
    if (!n) return null
    const part = PART_BY_ID[n.partId]
    const p = part?.ports.find((x) => x.id === portId)
    if (!part || !p) return null
    return { x: n.x + p.x * part.w, y: n.y + p.y * part.h }
  }, [nodeById])

  // ═══ ضبط العرض ═══
  //
  // ⚠️ هذا **مو تحسيناً شكلياً**: مخطط ينفتح بقطع برّا الحافة يعني
  // متدرّب ما يشوف نص التحدي ولا يگدر يوصلها. انكشف بتحدي «اعزل
  // شبكة الضيوف» — حاسبة الضيف چانت تحت حافة اللوح المضمّن، والفحص
  // الآلي گدر «يضغطها» بس الإنسان ما يشوفها أصلاً.
  const fit = useCallback(() => {
    const r = svgRef.current?.getBoundingClientRect()
    if (!r || doc.nodes.length === 0) return
    let x0 = Infinity, y0 = Infinity, x1 = -Infinity, y1 = -Infinity
    for (const n of doc.nodes) {
      const part = PART_BY_ID[n.partId]
      if (!part) continue
      x0 = Math.min(x0, n.x); y0 = Math.min(y0, n.y)
      x1 = Math.max(x1, n.x + part.w); y1 = Math.max(y1, n.y + part.h + 20)
    }
    if (!Number.isFinite(x0)) return
    const pad = 40
    const k = Math.min(1.4, Math.max(0.35, Math.min((r.width - pad * 2) / (x1 - x0), (r.height - pad * 2) / (y1 - y0))))
    setView({ k, x: (r.width - (x1 - x0) * k) / 2 - x0 * k, y: (r.height - (y1 - y0) * k) / 2 - y0 * k })
  }, [doc.nodes])

  // ضبط أول مرة يوصل بيها مخطط فيه قطع — وبعدين بس لمن ينطلب.
  const fittedRef = useRef(false)
  useEffect(() => {
    if (fittedRef.current || doc.nodes.length === 0) return
    fittedRef.current = true
    fit()
  }, [doc.nodes.length, fit])
  useEffect(() => {
    if (fitSignal === undefined) return
    fit()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [fitSignal])

  // ═══ الحذف بالمفتاح ═══
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') { setArmed(null); setReject(null); return }
      if ((e.key === 'Delete' || e.key === 'Backspace') && selected) {
        const t = e.target as HTMLElement
        if (t.tagName === 'INPUT' || t.tagName === 'TEXTAREA' || t.isContentEditable) return
        e.preventDefault()
        setDoc((d) => ({
          ...d,
          nodes: d.nodes.filter((n) => n.id !== selected),
          links: d.links.filter((l) => l.from.node !== selected && l.to.node !== selected && l.id !== selected),
        }))
        setSelected(null)
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [selected, setDoc, setSelected])

  // ═══ الوضع والسحب ═══
  const onCanvasClick = (e: React.MouseEvent) => {
    if (!pendingPart) { setSelected(null); setArmed(null); return }
    const part = PART_BY_ID[pendingPart]
    if (!part) return
    const b = toBoard(e.clientX, e.clientY)
    const id = `n${Date.now().toString(36)}${Math.floor(Math.random() * 1e4).toString(36)}`
    // ⚠️ ترقيم تلقائي: بدونه كل حاسبة تنحط تاخذ نفس الاسم ونفس
    // العنوان، فالمحاكي يشتكي «تعارض عناوين» من أول قطعتين — والفني
    // يظن إن المحاكي خربان مو إن التسمية افتراضية. Packet Tracer
    // يرقّم تلقائياً لنفس السبب.
    const sameKind = doc.nodes.filter((n) => n.partId === part.id).length
    const params = Object.fromEntries(part.params.map((pd) => [pd.id, pd.default]))
    if (sameKind > 0) {
      for (const pd of part.params) {
        if (pd.kind === 'text' && /^(name|hostname)$/.test(pd.id)) {
          params[pd.id] = String(pd.default).replace(/\d+$/, '') + (sameKind + 1)
        }
        // العنوان يزيد بآخر خانة — أبسط شي يمنع التعارض ويبقى مفهوماً.
        if (pd.id === 'ip' && typeof pd.default === 'string') {
          const m = /^(\d+\.\d+\.\d+\.)(\d+)$/.exec(pd.default)
          if (m) params[pd.id] = m[1] + Math.min(254, Number(m[2]) + sameKind)
        }
      }
    }
    const node: LabNode = {
      id, partId: part.id,
      x: snap(b.x - part.w / 2), y: snap(b.y - part.h / 2),
      rot: 0,
      params,
    }
    setDoc((d) => ({ ...d, nodes: [...d.nodes, node] }))
    setSelected(id)
    onPlaced()
  }

  const onNodeDown = (e: React.MouseEvent, n: LabNode) => {
    e.stopPropagation()
    setSelected(n.id)
    const b = toBoard(e.clientX, e.clientY)
    setDrag({ id: n.id, dx: b.x - n.x, dy: b.y - n.y })
  }

  const onMove = (e: React.MouseEvent) => {
    if (drag) {
      const b = toBoard(e.clientX, e.clientY)
      setDoc((d) => ({
        ...d,
        nodes: d.nodes.map((n) => (n.id === drag.id ? { ...n, x: snap(b.x - drag.dx), y: snap(b.y - drag.dy) } : n)),
      }))
      return
    }
    if (pan) setView((v) => ({ ...v, x: pan.vx + (e.clientX - pan.sx), y: pan.vy + (e.clientY - pan.sy) }))
  }

  const onUp = () => { setDrag(null); setPan(null) }

  const onBgDown = (e: React.MouseEvent) => {
    // الزر الأوسط أو Space+سحب = تحريك اللوح؛ نستعمل الأوسط واليمين.
    if (e.button === 1 || e.button === 2) {
      e.preventDefault()
      setPan({ sx: e.clientX, sy: e.clientY, vx: view.x, vy: view.y })
    }
  }

  const onWheel = (e: React.WheelEvent) => {
    const r = svgRef.current?.getBoundingClientRect()
    if (!r) return
    const mx = e.clientX - r.left, my = e.clientY - r.top
    const k2 = Math.min(2.5, Math.max(0.35, view.k * (e.deltaY < 0 ? 1.12 : 1 / 1.12)))
    // ⚠️ التكبير حوالين المؤشر مو حوالين الأصل: بدونه اللوح ينزلق
    // بعيداً كل مرة يكبّر بيها المستخدم.
    setView({ k: k2, x: mx - ((mx - view.x) / view.k) * k2, y: my - ((my - view.y) / view.k) * k2 })
  }

  // ═══ الربط ═══
  const clickPort = (e: React.MouseEvent, nodeId: string, portId: string) => {
    e.stopPropagation()
    setReject(null)
    if (!armed) { setArmed({ node: nodeId, port: portId }); return }
    if (armed.node === nodeId && armed.port === portId) { setArmed(null); return }

    const a = PART_BY_ID[nodeById(armed.node)?.partId ?? '']?.ports.find((p) => p.id === armed.port)
    const b = PART_BY_ID[nodeById(nodeId)?.partId ?? '']?.ports.find((p) => p.id === portId)
    if (!a || !b) { setArmed(null); return }

    // ⚠️ توافق نوع المنفذ: كيبل شبكة ما ينسنّب بطرف تغذية مستمرة.
    // الرفض هنا **استحالة فيزيائية** مو غلط تدريبي — ما ينحسب على أحد.
    // ⚠️ قفص SFP يقبل الربط مع منفذ نحاس **باللوح**، والفحص الحقيقي
    // يصير بالمحرّك (ترانسيفر نحاسي SFP-T موجود فعلاً بالسوق). المنع
    // هنا يخفي درساً بدل ما يعلّمه.
    const bothNet = (k: string) => k === 'eth' || k === 'sfp'
    if (a.kind !== b.kind && !(bothNet(a.kind) && bothNet(b.kind))) {
      setReject(`ما ينربط: «${a.label}» نوعه ${KIND_AR[a.kind]} و«${b.label}» نوعه ${KIND_AR[b.kind]}.`)
      setArmed(null)
      return
    }
    const dup = doc.links.some((l) =>
      (l.from.node === armed.node && l.from.port === armed.port && l.to.node === nodeId && l.to.port === portId) ||
      (l.to.node === armed.node && l.to.port === armed.port && l.from.node === nodeId && l.from.port === portId))
    if (dup) { setReject('موصولين أصلاً.'); setArmed(null); return }

    setDoc((d) => ({
      ...d,
      links: [...d.links, {
        id: `l${Date.now().toString(36)}${Math.floor(Math.random() * 1e4).toString(36)}`,
        from: { node: armed.node, port: armed.port },
        to: { node: nodeId, port: portId },
        // ⚠️ الخصائص حسب **نوع المنفذ**: كيبل شبكة إله نوع وترانسيفر،
        // وخط سماعات إله مقطع وطول، ووصلة تغذية مستمرة ماكو إلها شي.
        params: (() => {
          const defs = linkParamsFor(a.kind)
          return defs ? Object.fromEntries(defs.map((lp) => [lp.id, lp.default])) : undefined
        })(),
      }],
    }))
    setArmed(null)
  }

  const linkColor = (id: string) => {
    const st = result?.linkState[id]
    if (st === 'ok') return '#22c55e'
    if (st === 'bad') return '#ef4444'
    return '#64748b'
  }

  return (
    <div className="relative flex-1 overflow-hidden rounded-xl bg-[#0b1017] ring-1 ring-slate-700">
      <svg
        ref={svgRef}
        aria-label="لوح المحاكاة"
        className="h-full w-full"
        style={{ cursor: pendingPart ? 'copy' : drag ? 'grabbing' : 'default' }}
        onClick={onCanvasClick}
        onMouseDown={onBgDown}
        onMouseMove={onMove}
        onMouseUp={onUp}
        onMouseLeave={onUp}
        onWheel={onWheel}
        onContextMenu={(e) => e.preventDefault()}
      >
        <defs>
          <pattern id="grid" width={GRID} height={GRID} patternUnits="userSpaceOnUse">
            <circle cx={1} cy={1} r={0.9} fill="#1e293b" />
          </pattern>
        </defs>
        <rect width="100%" height="100%" fill="url(#grid)" />

        <g transform={`translate(${view.x},${view.y}) scale(${view.k})`}>
          {/* ═══ الوصلات ═══ تنرسم **أول** حتى تبقى تحت القطع */}
          {doc.links.map((l) => {
            const a = portPos(l.from.node, l.from.port)
            const b = portPos(l.to.node, l.to.port)
            if (!a || !b) return null
            // مسار متعامد — أوضح من الخط المائل بمخطط فيه قطع كثيرة
            const midY = (a.y + b.y) / 2
            const d = Math.abs(a.x - b.x) < 8
              ? `M ${a.x} ${a.y} L ${b.x} ${b.y}`
              : `M ${a.x} ${a.y} L ${a.x} ${midY} L ${b.x} ${midY} L ${b.x} ${b.y}`
            return (
              <g key={l.id} onClick={(e) => { e.stopPropagation(); setSelected(l.id) }} className="cursor-pointer">
                <path d={d} fill="none" stroke="transparent" strokeWidth={12} />
                <path d={d} fill="none" stroke={linkColor(l.id)} strokeWidth={selected === l.id ? 3.5 : 2}
                  strokeLinejoin="round" strokeDasharray={result?.linkState[l.id] === 'off' ? '5 4' : undefined} />
              </g>
            )
          })}

          {/* ═══ القطع ═══ */}
          {doc.nodes.map((n) => {
            const part = PART_BY_ID[n.partId]
            if (!part) return null
            const isSel = selected === n.id
            const readings = result?.nodeReadings[n.id] ?? []
            const live = readings.some((r) => r.tone === 'ok')
            const accent = isSel ? '#38bdf8' : live ? '#4ade80' : '#475569'
            return (
              <g key={n.id} transform={`translate(${n.x},${n.y}) rotate(${n.rot},${part.w / 2},${part.h / 2})`}>
                {/* ⚠️ `onClick` لازم يوگّف الانتشار هم مو بس `onMouseDown`:
                    `stopPropagation` بالـmousedown ما تمنع حدث الـclick
                    من الوصول لخلفية الـSVG، وهناك `onCanvasClick` تلغي
                    التحديد. النتيجة: تضغط القطعة، تنحدّد بالـmousedown،
                    وتنلغي بالـclick بعدها بميلي ثانية — فتحس إنها ما
                    تنحدّد أبداً. */}
                <g
                  onMouseDown={(e) => onNodeDown(e, n)}
                  onClick={(e) => e.stopPropagation()}
                  className="cursor-grab"
                >
                  <Symbol symbol={part.symbol} w={part.w} h={part.h} accent={accent} live={live} params={n.params} />
                </g>

                {/* الاسم تحت القطعة */}
                <text x={part.w / 2} y={part.h + 14} fontSize={11} fill="#cbd5e1" textAnchor="middle">
                  {String(n.params.name ?? n.params.hostname ?? n.label ?? part.name)}
                </text>

                {/* القراءات — تطلع بعد التشغيل */}
                {readings.map((r, i) => (
                  <text key={i} x={part.w / 2} y={-6 - i * 13} fontSize={11} textAnchor="middle" fontWeight="bold"
                    fill={r.tone === 'bad' ? '#f87171' : r.tone === 'warn' ? '#fbbf24' : '#4ade80'}>
                    {r.text}
                  </text>
                ))}

                {/* المنافذ */}
                {part.ports.map((p) => {
                  const pid = `${n.id}:${p.id}`
                  const isArmed = armed?.node === n.id && armed.port === p.id
                  const hov = hoverPort === pid
                  return (
                    <g key={p.id}>
                      <circle
                        cx={p.x * part.w} cy={p.y * part.h} r={isArmed || hov ? 7 : 5}
                        fill={isArmed ? '#38bdf8' : p.polarity === 'pos' ? '#ef4444' : p.polarity === 'neg' ? '#334155' : '#0ea5e9'}
                        stroke="#e2e8f0" strokeWidth={1.2}
                        className="cursor-crosshair"
                        onClick={(e) => clickPort(e, n.id, p.id)}
                        onMouseEnter={() => setHoverPort(pid)}
                        onMouseLeave={() => setHoverPort(null)}
                      />
                      {(hov || isArmed) && (
                        <text x={p.x * part.w} y={p.y * part.h - 11} fontSize={10.5} fill="#e2e8f0" textAnchor="middle">
                          {p.label}
                        </text>
                      )}
                    </g>
                  )
                })}
              </g>
            )
          })}
        </g>
      </svg>

      {/* ═══ شريط الحالة ═══ */}
      {armed && (
        <div className="pointer-events-none absolute bottom-3 right-3 rounded-lg bg-sky-500/90 px-3 py-1.5 text-xs font-bold text-white">
          مسلّح — اضغط المنفذ الثاني (Esc يلغي)
        </div>
      )}
      {reject && (
        <div className="absolute left-1/2 top-3 -translate-x-1/2 rounded-lg bg-slate-900/95 px-4 py-2 text-xs text-amber-200 ring-1 ring-amber-500/40">
          🔌 {reject}
        </div>
      )}
      {pendingPart && (
        <div className="pointer-events-none absolute bottom-3 left-3 rounded-lg bg-emerald-600/90 px-3 py-1.5 text-xs font-bold text-white">
          اضغط باللوح حتى تحط: {PART_BY_ID[pendingPart]?.name}
        </div>
      )}
      <div className="pointer-events-none absolute left-3 top-3 rounded-md bg-black/40 px-2 py-1 font-mono text-[10px] text-slate-400">
        {Math.round(view.k * 100)}٪ · {doc.nodes.length} قطعة · {doc.links.length} وصلة
      </div>
    </div>
  )
}

const KIND_AR: Record<string, string> = {
  dc: 'تيار مستمر',
  ac: 'تيار متناوب',
  eth: 'شبكة',
  signal: 'إشارة',
}
