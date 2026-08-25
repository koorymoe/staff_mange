// ═══ مساحة العمل ═══
//
// «أريد كأنما آني فاتح أقوى تطبيقات المحاكاة — شي منظّم ومرتّب وقوي».
//
// التخطيط المتعارف عليه بكل أدوات المحاكاة (MATLAB/Simulink، Packet
// Tracer، Fusion): **كتالوگ على جنب، لوح بالنص، خصائص على الجنب
// الثاني، وشريط أدوات فوگ ونتائج تحت**. ما نخترع تخطيطاً جديداً —
// الفني الي استعمل أي أداة قبل يعرف وين يدوّر بلا ما نعلّمه.
//
// ⚠️ المحرّك ما يشتغل مع كل حركة: يشتغل لمن تضغط **تشغيل**. المحاكاة
// المستمرة تخلّي اللوح يترمّش وأنت تسحب قطعة، وتخفي متى تغيّرت
// النتيجة فعلاً.

import { lazy, Suspense, useCallback, useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api'
import { CISCO_LIKE } from '../cli/ciscoLike'
import SimGate from '../sim/SimGate'
import Canvas from './Canvas'
import { CABLE_BY_ID, linkParamsFor, MEDIUM_AR } from './cables'
import { DOMAINS, PARTS, PART_BY_ID } from './catalog'
import { electricalEngine } from './engines/electrical'
import { audioEngine } from './engines/audio'
import { fireEngine } from './engines/fire'
import { networkEngine } from './engines/network'
import { solarEngine } from './engines/solar'
import { Symbol } from './symbols'
import type { DomainEngine, DomainId, LabDoc, ParamDef, SimResult } from './types'

// ⚠️ الترمنال `lazy`: أغلب من يفتح اللوح ما يفتح كونسولاً، وتحميل
// محرّك الأوامر مع الصفحة يثقّلها بلا فايدة.
const CliTerminal = lazy(() => import('../cli/CliTerminal'))
// ⚠️ المشهد الثلاثي `lazy` هم: Babylon حزمة ثقيلة، وأغلب الشغل
// يصير بالمنظر التخطيطي. تحميلها مع الاستوديو يثقّل فتحه بلا فايدة.
const Scene3D = lazy(() => import('./Scene3D'))

/** السويچات الي تنهيّأ بالكونسول. */
const CLI_PARTS = new Set(['switch_l2', 'switch_poe', 'switch_l3'])

const ENGINES: Record<DomainId, DomainEngine> = {
  electrical: electricalEngine,
  solar: solarEngine,
  network: networkEngine,
  fire: fireEngine,
  audio: audioEngine,
}

export default function LabWorkbench() {
  return <SimGate><Bench /></SimGate>
}

export interface BenchProps {
  /** مضمّنة داخل تمرين: بلا شريط سلامة ولا عنوان ولا حفظ مخططات. */
  embedded?: boolean
  /** مخطط ابتدائي يبدي بيه التحدي. */
  startDoc?: LabDoc
  /** ينندى بعد كل تشغيل — التمرين يقيّم النتيجة. */
  onResult?: (doc: LabDoc, result: SimResult) => void
}

/** ⚠️ مصدَّرة بلا بوابة عمداً: صفحة التمرين تضمّنها وهي أصلاً جوّا
 *  `SimGate`، وتغليفها مرتين يعني فحص ملكية مكرّراً بكل رندر. */
export function Bench({ embedded, startDoc, onResult }: BenchProps = {}) {
  const [domain, setDomain] = useState<DomainId>(startDoc?.domain ?? 'network')
  const [docs, setDocs] = useState<Record<DomainId, LabDoc>>(() => {
    const empty: Record<DomainId, LabDoc> = {
      network: { domain: 'network', nodes: [], links: [] },
      solar: { domain: 'solar', nodes: [], links: [] },
      electrical: { domain: 'electrical', nodes: [], links: [] },
      fire: { domain: 'fire', nodes: [], links: [] },
      audio: { domain: 'audio', nodes: [], links: [] },
    }
    if (startDoc) empty[startDoc.domain] = startDoc
    return empty
  })
  const [selected, setSelected] = useState<string | null>(null)
  const [pendingPart, setPendingPart] = useState<string | null>(null)
  const [result, setResult] = useState<SimResult | null>(null)
  const [console_, setConsole] = useState<string | null>(null)
  const [fitSignal, setFitSignal] = useState(0)
  const [zoomPct, setZoomPct] = useState(100)
  const [search, setSearch] = useState('')
  /** ⚠️ المنظران يقرآن **نفس** `doc` — مصدر واحد للحقيقة (المخطط ١٩). */
  const [view, setView] = useState<'plan' | '3d'>('plan')
  const [tab, setTab] = useState<'props' | 'ports' | 'state'>('props')

  // ═══ المخططات المحفوظة ═══
  const [projects, setProjects] = useState<{ id: string; name: string; domain: string; updatedAt: string }[]>([])
  const [projectId, setProjectId] = useState<string | null>(null)
  const [projectName, setProjectName] = useState('')
  const [busy, setBusy] = useState<string | null>(null)
  const [reload, setReload] = useState(0)

  useEffect(() => {
    let alive = true
    void (async () => {
      try {
        const rows = await api.listSimProjects()
        if (alive) setProjects(rows)
      } catch { /* القائمة مو حرجة — اللوح يشتغل بدونها */ }
    })()
    return () => { alive = false }
  }, [reload])

  const doc = docs[domain]
  const setDoc = useCallback((updater: (d: LabDoc) => LabDoc) => {
    setDocs((all) => ({ ...all, [domain]: updater(all[domain]) }))
    // ⚠️ النتيجة تنمسح مع أي تعديل: إبقاء قراءات قديمة فوگ مخطط
    // تغيّر يعني الفني يقرأ رقماً ما يخص الي قدّامه.
    setResult(null)
  }, [domain])

  const parts = useMemo(() => PARTS.filter((p) => p.domain === domain), [domain])
  /** ⚠️ البحث بالاسم **وبالموديل**: الفني يدوّر «PoE» أو «١٠٠ فولت»
   *  أكثر ما يدوّر باسم القطعة الكامل. */
  const shownParts = useMemo(() => {
    const q = search.trim().toLowerCase()
    if (!q) return parts
    return parts.filter((p) => (p.name + ' ' + (p.model ?? '')).toLowerCase().includes(q))
  }, [parts, search])
  const selNode = doc.nodes.find((n) => n.id === selected)
  const selPart = selNode ? PART_BY_ID[selNode.partId] : null
  const selLink = doc.links.find((l) => l.id === selected)


  /** ⚠️ حالة الكونسول تنكتب بالعقدة مو بحالة منفصلة: لو انخزنت
   *  برّا، تضيع بأول تبديل مجال وما تنحفظ مع المخطط. */
  const saveCli = useCallback((nodeId: string, state: Record<string, unknown>) => {
    setDocs((all) => ({
      ...all,
      [all[domain].domain]: {
        ...all[domain],
        nodes: all[domain].nodes.map((n) =>
          n.id === nodeId
            ? { ...n, cliState: state, params: { ...n.params, hostname: String(state.hostname ?? n.params.hostname) } }
            : n),
      },
    }))
    setResult(null)
  }, [domain])

  const saveProject = async () => {
    const name = projectName.trim() || `مخطط ${DOMAINS.find((d) => d.id === domain)?.name}`
    setBusy('يحفظ…')
    try {
      const p = await api.saveSimProject({ id: projectId ?? undefined, name, domain, doc })
      setProjectId(p.id)
      setProjectName(p.name)
      setReload((n) => n + 1)
      setBusy('انحفظ ✅')
      window.setTimeout(() => setBusy(null), 1800)
    } catch (e) {
      setBusy(e instanceof Error ? e.message : 'تعذر الحفظ')
    }
  }

  const openProject = async (id: string) => {
    if (!id) return
    setBusy('يفتح…')
    try {
      const p = await api.getSimProject(id)
      const d = p.doc as LabDoc
      setDocs((all) => ({ ...all, [d.domain]: d }))
      setDomain(d.domain)
      setProjectId(p.id)
      setProjectName(p.name)
      setSelected(null)
      setResult(null)
      setFitSignal((n) => n + 1)   // المخطط المفتوح ممكن يكون أوسع من اللوح
      setBusy(null)
    } catch (e) {
      setBusy(e instanceof Error ? e.message : 'تعذر الفتح')
    }
  }
  const clearAll = () => { setDoc(() => ({ domain, nodes: [], links: [] })); setSelected(null) }

  // ⚠️ تبديل القطعة يرجّع التبويب لـ«الخصائص»: لو بقيت بتبويب
  // «الحالة» وضغطت قطعة ما إلها قراءات، تشوف لوحة فاضية وتظن
  // التحديد ما اشتغل.
  //
  // ⚠️ التصفير **أثناء الرندر** مو بتأثير — هذا النمط الموثّق بـReact
  // لتصفير حالة عند تغيّر مدخل. `useEffect` يسبّب رندراً متسلسلاً:
  // رندر بالتبويب القديم ثم رندر ثانٍ بالجديد، والمستخدم يشوف ومضة.
  const [tabFor, setTabFor] = useState<string | null>(null)
  if (tabFor !== selected) {
    setTabFor(selected)
    setTab('props')
  }

  const setParam = (id: string, value: string | number | boolean) => {
    if (!selNode) return
    setDoc((d) => ({ ...d, nodes: d.nodes.map((n) => (n.id === selNode.id ? { ...n, params: { ...n.params, [id]: value } } : n)) }))
  }

  const setLinkParam = (id: string, value: string | number | boolean) => {
    if (!selLink) return
    setDoc((d) => ({
      ...d,
      links: d.links.map((l) => (l.id === selLink.id ? { ...l, params: { ...(l.params ?? {}), [id]: value } } : l)),
    }))
  }

  // ⚠️ إزالة التكرار: المحرّك يفحص كل منفذ لحاله، فغلط واحد بمنفذين
  // (الموجب والسالب) يطلع رسالتين متطابقتين. الفني يقراها مرتين ويظن
  // إن عنده غلطتين.
  const msgs = useMemo(() => {
    const seen = new Set<string>()
    return (result?.messages ?? []).filter((m) => {
      const k = m.kind + '|' + m.text
      if (seen.has(k)) return false
      seen.add(k)
      return true
    })
  }, [result])
  const errors = msgs.filter((m) => m.kind === 'error')
  const warns = msgs.filter((m) => m.kind === 'warn')

  // ═══ سجل الأحداث ═══
  //
  // ⚠️ **سجل** مو قائمة رسائل: كل تشغيل يترك أثراً بوقته، فالفني
  // يشوف «شنو تغيّر بعد آخر تعديل» بدل ما يقارن قائمتين بعينه.
  // ⚠️ والوقت ينختم **وقت التشغيل** مو بالرندر — الرندر يتكرّر
  // فيتغيّر الوقت بلا ما يصير شي.
  const [log, setLog] = useState<{ t: string; kind: string; text: string }[]>([])

  const runAndLog = useCallback(() => {
    const r = ENGINES[domain].run(doc, PART_BY_ID)
    setResult(r)
    onResult?.(doc, r)
    const now = new Date()
    const t = `${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}:${String(now.getSeconds()).padStart(2, '0')}`
    const seen = new Set<string>()
    const fresh = r.messages.filter((m) => {
      const k = m.kind + '|' + m.text
      if (seen.has(k)) return false
      seen.add(k)
      return true
    })
    setLog((prev) => [
      ...fresh.map((m) => ({ t, kind: m.kind, text: m.text })),
      { t, kind: 'run', text: `بدأت المحاكاة بوضع ${DOMAINS.find((d) => d.id === domain)?.name}` },
      ...prev,
    ].slice(0, 80))
  }, [domain, doc, onResult])

  const KIND_CHIP: Record<string, { label: string; cls: string }> = {
    error: { label: 'خطأ', cls: 'bg-red-500/15 text-red-300 ring-red-500/30' },
    warn: { label: 'تحذير', cls: 'bg-amber-500/15 text-amber-300 ring-amber-500/30' },
    info: { label: 'نجاح', cls: 'bg-emerald-500/15 text-emerald-300 ring-emerald-500/30' },
    run: { label: 'معلومات', cls: 'bg-sky-500/15 text-sky-300 ring-sky-500/30' },
  }

  return (
    <div dir="rtl" className="space-y-3">
      {!embedded && (
        <>
          <div className="rounded-xl bg-amber-50 px-4 py-2.5 text-[12.5px] font-bold leading-relaxed text-amber-900 ring-1 ring-amber-200">
            ⚠️ محاكاة تدريب — القيم نمطية عامة مو كتالوگ موديل بعينه. الرجوع لكتالوگ
            الشركة المصنّعة إلزامي قبل أي تنفيذ حقيقي.
          </div>
          <div className="flex flex-wrap items-center justify-between gap-3">
            <Link to="/simulator-lab" className="text-sm text-brand-700 hover:underline">← رجوع للمختبر</Link>
            <h2 className="text-xl font-bold text-brand-900">🧰 مساحة عمل المحاكاة</h2>
          </div>
        </>
      )}

      {/* ⚠️ اللوح يحتاج مساحة — بالموبايل ما يشتغل. */}
      <div className="rounded-2xl bg-white p-8 text-center shadow md:hidden">
        <p className="text-lg font-bold text-slate-700">🖥️ افتحه من الحاسبة</p>
        <p className="mt-2 text-sm text-slate-500">مساحة العمل تحتاج شاشة كبيرة — سحب قطع وربط بشاشة ٦ إنچ ما يشتغل.</p>
      </div>

      {/* ═══ الاستوديو ═══
          ⚠️ سطح **داكن** عمداً: أدوات المحاكاة والتصميم (MATLAB،
          KiCad، Packet Tracer بوضعه الفيزيائي) كلها داكنة، لأن
          المخطط الملوّن يُقرا أحسن على خلفية داكنة والعين ما تتعب
          بجلسة طويلة. وباقي النظام يبقى فاتحاً — هاي **أداة** جوّا
          نظام، مو صفحة إدارية. */}
      <div className="hidden overflow-hidden rounded-2xl bg-[#0b1220] shadow-[0_10px_40px_rgba(2,8,23,0.35)] ring-1 ring-slate-800 md:block">
        {/* شريط الأدوات */}
        <div className="flex flex-wrap items-center gap-2 border-b border-slate-800 bg-[#0e1626] px-3 py-2.5">
          <div className="inline-flex overflow-hidden rounded-lg ring-1 ring-slate-700">
            {DOMAINS.map((d) => (
              <button
                key={d.id}
                onClick={() => { setDomain(d.id); setSelected(null); setPendingPart(null); setResult(null); setTab('props') }}
                className={`px-3 py-1.5 text-xs font-bold transition ${
                  domain === d.id ? 'bg-sky-600 text-white' : 'bg-[#0b1220] text-slate-400 hover:bg-slate-800 hover:text-slate-200'}`}
              >
                {d.icon} {d.name}
              </button>
            ))}
          </div>

          <div className="mr-auto flex items-center gap-2">
            <button onClick={runAndLog}
              className="flex items-center gap-1.5 rounded-lg bg-emerald-600 px-4 py-1.5 text-xs font-bold text-white shadow hover:bg-emerald-500">
              ▶ تشغيل المحاكاة
            </button>
            <button onClick={() => { setResult(null) }} disabled={!result}
              className="rounded-lg bg-slate-800 px-3 py-1.5 text-xs font-bold text-slate-300 disabled:opacity-40">
              ■ إيقاف
            </button>
            <div className="inline-flex overflow-hidden rounded-lg ring-1 ring-slate-700">
              {([['plan', '🗺️ مخطط'], ['3d', '🧊 ثلاثي الأبعاد']] as const).map(([v, label]) => (
                <button key={v} onClick={() => setView(v)}
                  className={`px-3 py-1.5 text-xs font-bold transition ${
                    view === v ? 'bg-slate-700 text-white' : 'bg-[#0b1220] text-slate-400 hover:text-slate-200'}`}>
                  {label}
                </button>
              ))}
            </div>
            {view === 'plan' && (
              <button onClick={() => setFitSignal((n) => n + 1)}
                className="rounded-lg bg-slate-800 px-3 py-1.5 text-xs font-bold text-slate-300 hover:bg-slate-700">
                ⤢ ضبط العرض
              </button>
            )}
            <button onClick={clearAll}
              className="rounded-lg bg-red-500/10 px-3 py-1.5 text-xs font-bold text-red-300 ring-1 ring-red-500/30 hover:bg-red-500/20">
              🗑 تفريغ اللوح
            </button>
          </div>
        </div>

        {/* شريط المخططات */}
        {!embedded && (
          <div className="flex flex-wrap items-center gap-2 border-b border-slate-800 bg-[#0b1220] px-3 py-2">
            <span className="text-[11px] font-bold text-slate-500">المخطط</span>
            <input
              value={projectName} onChange={(e) => setProjectName(e.target.value)}
              placeholder="اسم المخطط"
              className="w-48 rounded-lg border border-slate-700 bg-[#0e1626] px-2.5 py-1.5 text-[12px] text-slate-200 placeholder:text-slate-600"
            />
            <button onClick={saveProject}
              className="rounded-lg bg-sky-600 px-3.5 py-1.5 text-xs font-bold text-white hover:bg-sky-500">
              💾 {projectId ? 'احفظ التعديلات' : 'احفظ جديداً'}
            </button>
            {projectId && (
              <button onClick={() => { setProjectId(null); setProjectName('') }}
                className="rounded-lg bg-slate-800 px-3 py-1.5 text-xs font-bold text-slate-300">نسخة جديدة</button>
            )}
            <select
              value={projectId ?? ''} onChange={(e) => void openProject(e.target.value)}
              className="rounded-lg border border-slate-700 bg-[#0e1626] px-2 py-1.5 text-[12px] text-slate-200"
            >
              <option value="">— افتح مخططاً محفوظاً —</option>
              {projects.map((pr) => (
                <option key={pr.id} value={pr.id}>
                  {pr.name} · {DOMAINS.find((d) => d.id === pr.domain)?.name ?? pr.domain}
                </option>
              ))}
            </select>
            {projectId && (
              <button
                onClick={async () => {
                  if (!projectId) return
                  await api.deleteSimProject(projectId).catch(() => {})
                  setProjectId(null); setProjectName(''); setReload((n) => n + 1)
                }}
                className="rounded-lg bg-red-500/10 px-3 py-1.5 text-xs font-bold text-red-300 ring-1 ring-red-500/30"
              >احذف المخطط</button>
            )}
            {busy && <span className="text-xs font-bold text-slate-400">{busy}</span>}
          </div>
        )}

        {/* ═══ الأعمدة الثلاثة ═══ */}
        <div className="grid gap-0 lg:grid-cols-[16rem_1fr_17rem]">
          {/* ─── مكتبة الأجهزة ─── */}
          <div className="order-2 border-l border-slate-800 bg-[#0e1626] p-3 lg:order-1">
            <div className="mb-2 flex items-center justify-between">
              <span className="text-[10px] text-slate-600">{parts.length} جهاز</span>
              <h3 className="text-[13px] font-bold text-slate-200">مكتبة الأجهزة</h3>
            </div>

            <input
              value={search} onChange={(e) => setSearch(e.target.value)}
              placeholder="🔍 ابحث عن جهاز…"
              className="mb-2.5 w-full rounded-lg border border-slate-700 bg-[#0b1220] px-2.5 py-1.5 text-[12px] text-slate-200 placeholder:text-slate-600"
            />

            <div className="grid max-h-[400px] gap-1.5 overflow-y-auto pl-1">
              {shownParts.length === 0 && (
                <p className="py-6 text-center text-[11px] text-slate-600">ماكو جهاز يطابق البحث.</p>
              )}
              {shownParts.map((pt) => (
                <button
                  key={pt.id}
                  onClick={() => setPendingPart(pendingPart === pt.id ? null : pt.id)}
                  className={`flex items-center gap-2 rounded-lg border p-2 text-right transition ${
                    pendingPart === pt.id
                      ? 'border-emerald-500 bg-emerald-500/10'
                      : 'border-slate-800 bg-[#0b1220] hover:border-sky-600 hover:bg-slate-800/50'}`}
                >
                  <svg viewBox={`0 0 ${pt.w} ${pt.h}`} width={34} height={26} className="shrink-0">
                    <Symbol symbol={pt.symbol} w={pt.w} h={pt.h} accent="#64748b"
                      params={Object.fromEntries(pt.params.map((x) => [x.id, x.default]))} />
                  </svg>
                  <span className="min-w-0 flex-1">
                    <span className="block truncate text-[11.5px] font-bold text-slate-200">{pt.name}</span>
                    {pt.model && <span className="block truncate text-[10px] text-slate-500">{pt.model}</span>}
                  </span>
                </button>
              ))}
            </div>

            <div className="mt-2.5 rounded-lg border border-dashed border-slate-700 px-3 py-2.5 text-center text-[11px] text-slate-500">
              ☝︎ اختر جهازاً ثم اضغط باللوح
            </div>
          </div>

          {/* ─── اللوح ─── */}
          <div className="order-1 flex min-h-[620px] flex-col lg:order-2">
            {/* ⚠️ `flex` على الحاوية إجبارية: `flex-1` جوّا حاوية
                **بلوك** ما تسوي شي، فاللوح ينكمش لارتفاع محتواه
                (والـSVG بـ`h-full` جوّا حاوية بلا ارتفاع = صفر
                تقريباً). طلعت شريطاً نحيفاً بنص شاشة فاضية. */}
            <div className="relative flex min-h-[440px] flex-1 p-2">
              {view === '3d' ? (
                <Suspense fallback={
                  <div className="flex h-full w-full items-center justify-center rounded-xl bg-[#070c14] text-slate-500">
                    جاري تحميل المشهد الثلاثي…
                  </div>
                }>
                  <Scene3D doc={doc} result={result} selected={selected} setSelected={setSelected} />
                </Suspense>
              ) : (
                <Canvas
                  doc={doc} setDoc={setDoc} result={result}
                  selected={selected} setSelected={setSelected}
                  pendingPart={pendingPart} onPlaced={() => setPendingPart(null)}
                  fitSignal={fitSignal}
                  onZoom={setZoomPct}
                />
              )}

              {/* أدوات على اللوح */}
              <div className="pointer-events-none absolute right-5 top-5 flex items-center gap-2">
                <span className={`pointer-events-auto flex items-center gap-1.5 rounded-full px-3 py-1 text-[11px] font-bold ring-1 ${
                  result ? 'bg-emerald-500/15 text-emerald-300 ring-emerald-500/30' : 'bg-slate-800/80 text-slate-400 ring-slate-700'}`}>
                  <span className={`inline-block h-2 w-2 rounded-full ${result ? 'animate-pulse bg-emerald-400' : 'bg-slate-500'}`} />
                  {result ? 'وضع المحاكاة' : 'وضع التحرير'}
                </span>
              </div>

              {view === 'plan' && (
              <div className="pointer-events-auto absolute left-5 top-5 flex items-center gap-1 rounded-lg bg-[#0e1626]/95 p-1 ring-1 ring-slate-700">
                <button onClick={() => setFitSignal((n) => n + 1)} title="ضبط العرض"
                  className="rounded px-2 py-0.5 text-[13px] text-slate-300 hover:bg-slate-700">⤢</button>
                <span className="px-1.5 font-mono text-[11px] tabular-nums text-slate-400">{zoomPct}٪</span>
              </div>
              )}

              {/* خريطة مصغّرة */}
              {view === 'plan' && doc.nodes.length > 0 && (
                <div className="pointer-events-none absolute bottom-5 left-5 h-20 w-28 overflow-hidden rounded-lg bg-[#0e1626]/95 ring-1 ring-slate-700">
                  <Minimap doc={doc} />
                </div>
              )}
            </div>

            {/* ─── سجل الأحداث ─── */}
            <div className="border-t border-slate-800 bg-[#0e1626]">
              <div className="flex items-center justify-between px-3 py-1.5">
                <button onClick={() => setLog([])} className="text-[10px] text-slate-600 hover:text-slate-400">تفريغ</button>
                <h3 className="text-[12px] font-bold text-slate-300">سجل الأحداث</h3>
              </div>
              <div className="max-h-40 overflow-y-auto px-2 pb-2">
                {log.length === 0 && (
                  <p className="px-2 py-3 text-[11px] text-slate-600">اضغط «تشغيل المحاكاة» — النتائج تنسجّل هنا بوقتها.</p>
                )}
                {log.map((e, i) => {
                  const chip = KIND_CHIP[e.kind] ?? KIND_CHIP.run
                  return (
                    <div key={i} className="flex items-start gap-2 border-b border-slate-800/60 px-1.5 py-1.5 last:border-0">
                      <span className="mt-0.5 shrink-0 font-mono text-[10px] tabular-nums text-slate-600">{e.t}</span>
                      <span className={`mt-0.5 shrink-0 rounded px-1.5 py-0.5 text-[9.5px] font-bold ring-1 ${chip.cls}`}>{chip.label}</span>
                      <span className="flex-1 text-[11.5px] leading-relaxed text-slate-300">{e.text}</span>
                    </div>
                  )
                })}
              </div>
            </div>
          </div>

          {/* ─── الخصائص ─── */}
          <div className="order-3 border-r border-slate-800 bg-[#0e1626] p-3">
            {!selNode && !selLink && (
              <>
                <h3 className="mb-2 text-[13px] font-bold text-slate-200">الخصائص</h3>
                <p className="rounded-lg border border-dashed border-slate-700 px-3 py-8 text-center text-[11px] text-slate-500">
                  اضغط قطعة أو كيبلاً باللوح
                </p>
              </>
            )}

            {selNode && selPart && (
              <>
                {/* رأس البطاقة */}
                <div className="mb-2.5 flex items-center gap-2.5 rounded-lg bg-[#0b1220] p-2.5 ring-1 ring-slate-800">
                  <svg viewBox={`0 0 ${selPart.w} ${selPart.h}`} width={44} height={32} className="shrink-0">
                    <Symbol symbol={selPart.symbol} w={selPart.w} h={selPart.h} accent="#38bdf8"
                      params={selNode.params} />
                  </svg>
                  <div className="min-w-0">
                    <p className="truncate text-[13px] font-bold text-slate-100">
                      {String(selNode.params.name ?? selNode.params.hostname ?? selPart.name)}
                    </p>
                    {selPart.model && <p className="truncate text-[10px] text-slate-500">{selPart.model}</p>}
                  </div>
                </div>

                {/* التبويبات */}
                <div className="mb-2.5 flex rounded-lg bg-[#0b1220] p-0.5 ring-1 ring-slate-800">
                  {([['props', 'الخصائص'], ['ports', 'المنافذ'], ['state', 'الحالة']] as const).map(([id, label]) => (
                    <button key={id} onClick={() => setTab(id)}
                      className={`flex-1 rounded px-2 py-1 text-[11px] font-bold transition ${
                        tab === id ? 'bg-sky-600 text-white' : 'text-slate-400 hover:text-slate-200'}`}>
                      {label}
                    </button>
                  ))}
                </div>

                {tab === 'props' && (
                  <div className="space-y-2.5">
                    {selPart.about && <p className="text-[11px] leading-relaxed text-slate-500">{selPart.about}</p>}
                    {selPart.danger && (
                      <p className="rounded-lg bg-red-500/10 px-2.5 py-1.5 text-[10.5px] font-bold leading-relaxed text-red-300 ring-1 ring-red-500/25">
                        ⚠️ {selPart.danger}
                      </p>
                    )}
                    {CLI_PARTS.has(selNode.partId) && (
                      <button
                        onClick={() => setConsole(console_ === selNode.id ? null : selNode.id)}
                        className={`w-full rounded-lg py-2 text-[12px] font-bold transition ${
                          console_ === selNode.id ? 'bg-slate-700 text-white' : 'bg-black text-emerald-300 ring-1 ring-emerald-800 hover:bg-slate-900'}`}
                      >
                        🖥️ {console_ === selNode.id ? 'سكّر الكونسول' : 'افتح الكونسول'}
                      </button>
                    )}
                    {selPart.params.map((pd) => (
                      <Field key={pd.id} def={pd} value={selNode.params[pd.id]} onChange={(v) => setParam(pd.id, v)} dark />
                    ))}
                  </div>
                )}

                {tab === 'ports' && (
                  <div className="space-y-1.5">
                    {selPart.ports.map((pt) => {
                      const linked = doc.links.filter(
                        (l) => (l.from.node === selNode.id && l.from.port === pt.id) ||
                               (l.to.node === selNode.id && l.to.port === pt.id))
                      const peer = linked[0]
                        ? doc.nodes.find((n) => n.id === (linked[0].from.node === selNode.id ? linked[0].to.node : linked[0].from.node))
                        : null
                      return (
                        <div key={pt.id} className="flex items-center gap-2 rounded-lg bg-[#0b1220] px-2.5 py-1.5 ring-1 ring-slate-800">
                          <span className={`inline-block h-2 w-2 shrink-0 rounded-full ${linked.length ? 'bg-emerald-400' : 'bg-slate-600'}`} />
                          <span className="min-w-0 flex-1">
                            <span className="block text-[11.5px] font-bold text-slate-200">{pt.label}</span>
                            <span className="block truncate text-[10px] text-slate-500">
                              {PORT_KIND_AR[pt.kind] ?? pt.kind}
                              {peer ? ` → ${String(peer.params.name ?? peer.params.hostname ?? '')}` : ' · فاضي'}
                            </span>
                          </span>
                        </div>
                      )
                    })}
                  </div>
                )}

                {tab === 'state' && (
                  <div className="space-y-1.5">
                    {!result && <p className="text-[11px] text-slate-500">شغّل المحاكاة حتى تظهر القراءات.</p>}
                    {result && (result.nodeReadings[selNode.id] ?? []).length === 0 && (
                      <p className="text-[11px] text-slate-500">ماكو قراءات لهالقطعة.</p>
                    )}
                    {(result?.nodeReadings[selNode.id] ?? []).map((r, i) => (
                      <div key={i} className="rounded-lg bg-[#0b1220] px-2.5 py-2 ring-1 ring-slate-800">
                        <p className={`font-mono text-[14px] font-bold ${
                          r.tone === 'bad' ? 'text-red-400' : r.tone === 'warn' ? 'text-amber-300' : 'text-emerald-400'}`}>
                          {r.text}
                        </p>
                      </div>
                    ))}
                  </div>
                )}

                <button
                  onClick={() => {
                    setDoc((d) => ({
                      ...d,
                      nodes: d.nodes.filter((n) => n.id !== selNode.id),
                      links: d.links.filter((l) => l.from.node !== selNode.id && l.to.node !== selNode.id),
                    }))
                    setSelected(null)
                  }}
                  className="mt-3 w-full rounded-lg bg-red-500/10 py-1.5 text-[11px] font-bold text-red-300 ring-1 ring-red-500/25 hover:bg-red-500/20"
                >🗑 احذف القطعة</button>
              </>
            )}

            {selLink && selLink.params && (() => {
              const fromKind = PART_BY_ID[doc.nodes.find((n) => n.id === selLink.from.node)?.partId ?? '']
                ?.ports.find((p) => p.id === selLink.from.port)?.kind ?? 'eth'
              const defs = linkParamsFor(fromKind) ?? []
              const net = CABLE_BY_ID[String(selLink.params.cable ?? '')]
              return (
                <div className="space-y-2.5">
                  <div className="rounded-lg bg-[#0b1220] p-2.5 ring-1 ring-slate-800">
                    <p className="text-[13px] font-bold text-slate-100">🔌 {fromKind === 'spk' ? 'خط سماعات' : 'كيبل'}</p>
                    {net && (
                      <p className="mt-1 text-[10px] text-slate-500">
                        {MEDIUM_AR[net.medium]} · الحد {net.maxM} م
                      </p>
                    )}
                  </div>
                  {net?.about && <p className="text-[11px] leading-relaxed text-slate-500">{net.about}</p>}
                  {defs.map((lp) => (
                    <Field key={lp.id} def={lp} value={selLink.params?.[lp.id]} onChange={(v) => setLinkParam(lp.id, v)} dark />
                  ))}
                  <button
                    onClick={() => { setDoc((d) => ({ ...d, links: d.links.filter((l) => l.id !== selLink.id) })); setSelected(null) }}
                    className="w-full rounded-lg bg-red-500/10 py-1.5 text-[11px] font-bold text-red-300 ring-1 ring-red-500/25"
                  >🗑 احذف الكيبل</button>
                </div>
              )
            })()}

            <div className="mt-4 border-t border-slate-800 pt-2.5 text-[10px] leading-relaxed text-slate-600">
              <b className="text-slate-500">الاختصارات</b><br />
              منفذ ثم منفذ = وصلة · Esc يلغي<br />
              Delete يحذف · عجلة الماوس تكبّر<br />
              الزر الأيمن + سحب يحرّك اللوح
            </div>
          </div>
        </div>

        {/* ═══ الكونسول ═══ */}
        {console_ && (() => {
          const nd = doc.nodes.find((n) => n.id === console_)
          if (!nd) return null
          return (
            <div className="border-t border-slate-800 bg-[#0b1220] p-3">
              <div className="mb-2 flex items-center justify-between">
                <button onClick={() => setConsole(null)} className="text-[11px] font-bold text-slate-500 hover:text-slate-300">✕ سكّر</button>
                <p className="text-xs font-bold text-slate-200">
                  🖥️ كونسول {String(nd.cliState?.hostname ?? nd.params.hostname ?? '')}
                </p>
              </div>
              <Suspense fallback={<div className="h-[280px] rounded-xl bg-black" />}>
                <CliTerminal
                  key={nd.id}
                  grammar={CISCO_LIKE}
                  initialState={nd.cliState ?? { hostname: String(nd.params.hostname ?? 'Switch') }}
                  onStateChange={(st) => saveCli(nd.id, st)}
                  heightClass="h-[280px]"
                />
              </Suspense>
              <p className="mt-1.5 text-[10.5px] text-slate-600">
                مثال: <span dir="ltr" className="font-mono">en → conf t → int gi0/2 → switchport access vlan 20 → end</span>
              </p>
            </div>
          )
        })()}
      </div>

      {errors.length > 0 && (
        <p className="hidden text-xs text-slate-400 md:block">{errors.length} مشكلة خطيرة · {warns.length} تحذير</p>
      )}
    </div>
  )
}

/** ═══ خريطة مصغّرة ═══
 *
 *  ⚠️ ما تعرض تفاصيل — تعرض **الشكل العام** بس: وين القطع وكم بعيدة
 *  عن بعض. بمخطط فيه ٤٠ قطعة والفني مكبّر على ركن، هاي الي تكله وين
 *  هو. */
function Minimap({ doc }: { doc: LabDoc }) {
  const boxes = doc.nodes.map((n) => {
    const part = PART_BY_ID[n.partId]
    return { x: n.x, y: n.y, w: part?.w ?? 60, h: part?.h ?? 40 }
  })
  if (boxes.length === 0) return null
  const x0 = Math.min(...boxes.map((b) => b.x)) - 20
  const y0 = Math.min(...boxes.map((b) => b.y)) - 20
  const x1 = Math.max(...boxes.map((b) => b.x + b.w)) + 20
  const y1 = Math.max(...boxes.map((b) => b.y + b.h)) + 20
  return (
    <svg viewBox={`${x0} ${y0} ${x1 - x0} ${y1 - y0}`} className="h-full w-full" preserveAspectRatio="xMidYMid meet">
      {doc.links.map((l) => {
        const a = doc.nodes.find((n) => n.id === l.from.node)
        const b = doc.nodes.find((n) => n.id === l.to.node)
        if (!a || !b) return null
        const pa = PART_BY_ID[a.partId], pb = PART_BY_ID[b.partId]
        return (
          <line key={l.id}
            x1={a.x + (pa?.w ?? 60) / 2} y1={a.y + (pa?.h ?? 40) / 2}
            x2={b.x + (pb?.w ?? 60) / 2} y2={b.y + (pb?.h ?? 40) / 2}
            stroke="#334155" strokeWidth={6} />
        )
      })}
      {boxes.map((b, i) => (
        <rect key={i} x={b.x} y={b.y} width={b.w} height={b.h} rx={6} fill="#38bdf8" opacity={0.75} />
      ))}
    </svg>
  )
}

const PORT_KIND_AR: Record<string, string> = {
  dc: 'تغذية مستمرة', ac: 'تيار متناوب', eth: 'شبكة RJ45',
  sfp: 'قفص SFP', spk: 'خط سماعات', signal: 'إشارة',
}

function Field({ def, value, onChange, dark }: {
  def: ParamDef
  value: string | number | boolean | undefined
  onChange: (v: string | number | boolean) => void
  /** نسخة داكنة لاستوديو المحاكاة. */
  dark?: boolean
}) {
  const inputCls = dark
    ? 'w-full rounded-lg border border-slate-700 bg-[#0b1220] px-2 py-1.5 text-[12px] text-slate-200'
    : 'w-full rounded-lg border border-slate-300 px-2 py-1.5 text-[12px]'
  const label = (
    <span className={`mb-1 flex items-baseline gap-1 text-[11px] font-bold ${dark ? 'text-slate-400' : 'text-slate-600'}`}>
      {def.label}
      {def.unit && <span className="text-slate-400">({def.unit})</span>}
    </span>
  )
  return (
    <label className="block">
      {label}
      {def.kind === 'bool' ? (
        <button
          type="button"
          onClick={() => onChange(!value)}
          className={`w-full rounded-lg px-3 py-1.5 text-[12px] font-bold transition ${
            value ? 'bg-emerald-600 text-white' : dark ? 'bg-slate-800 text-slate-300' : 'bg-slate-100 text-slate-600'}`}
        >
          {value ? 'مغلق ✓' : 'مفتوح'}
        </button>
      ) : def.kind === 'select' ? (
        <select
          value={String(value ?? def.default)}
          onChange={(e) => onChange(e.target.value)}
          className={inputCls}
        >
          {def.options?.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
        </select>
      ) : (
        <input
          type={def.kind === 'number' ? 'number' : 'text'}
          value={String(value ?? def.default)}
          min={def.min} max={def.max}
          dir={def.kind === 'number' ? 'ltr' : undefined}
          onChange={(e) => onChange(def.kind === 'number' ? Number(e.target.value) : e.target.value)}
          className={inputCls}
        />
      )}
      {def.help && <span className={`mt-0.5 block text-[10px] leading-relaxed ${dark ? 'text-slate-500' : 'text-slate-400'}`}>{def.help}</span>}
    </label>
  )
}
