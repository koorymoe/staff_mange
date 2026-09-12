// ═══════════════════════════════════════════════════════════════════
// مختبر الكائن — تجربة معزولة للمالك وحده
// ═══════════════════════════════════════════════════════════════════
//
// (م) طلبها صريحاً قبل أي نشر: «تجربة معزولة داخل النظام» ترجّعله
// نتائج تحميل المقاطع وأخطاء الكونسول وزمن التحميل والذاكرة **مقاسة**.
//
// ⚠️ **ولا شي من هنا يوصل الموظفين**: البوابة `actualRole === 'OWNER'`،
// والشاشة **قابلة للإطفاء بسطر واحد** (`LAB_ENABLED` تحت).
//
// ⚠️ **وكل رقم بهالشاشة مقاس لحظة الضغط، ماكو ولا رقم مكتوب**. شاشة
// تجربة تعرض «٩٨٪ دقة» ثابتة أسوأ من ماكو شاشة — تخلينا نعتمد شي
// ما قِسناه. ولو القياس مو ممكن، يُعرض **«—»** مع سبب.
import { useCallback, useEffect, useRef, useState } from 'react'
import { useSession } from '../session'
import PageHeader from '../components/PageHeader'
import type { AvatarHandle, AvatarStats, ClipName } from '../components/entityAvatarEngine'

/**
 * ⚠️ **مفتاح الإطفاء.** النظام ماكو بيه منظومة رايات ميزات (فحصتها:
 * ماكو `featureFlag` ولا `FEATURE_` بكل المستودع)، فالإطفاء يصير هنا:
 * `false` ← الشاشة تعرض «غير موجودة» للكل **بما فيهم المالك**،
 * والعارض الثقيل **ما ينحمّل إطلاقاً** (الاستيراد مؤجَّل تحت الشرط).
 */
const LAB_ENABLED = true

const GLB = `${import.meta.env.BASE_URL}amani-tech-v4.glb`

type Log = { at: string; text: string; bad?: boolean }

export default function EntityLabPage() {
  const { employee } = useSession()
  // ⚠️ نفس منطق `SimGate`: الواجهة تطبّع المالك إلى ADMIN، فالفحص
  // لازم يكون `actualRole` وإلا المالك نفسه ما يشوفها.
  if (!LAB_ENABLED || employee?.actualRole !== 'OWNER') {
    return (
      <div className="rounded-2xl bg-white p-8 text-center shadow">
        <p className="text-lg font-bold text-slate-700">الصفحة غير موجودة</p>
      </div>
    )
  }
  return <Lab />
}

function Lab() {
  const canvasRef = useRef<HTMLCanvasElement | null>(null)
  const handleRef = useRef<AvatarHandle | null>(null)
  const btnRefs = useRef<(HTMLButtonElement | null)[]>([])
  const [stats, setStats] = useState<AvatarStats | null>(null)
  const [logs, setLogs] = useState<Log[]>([])
  const [failed, setFailed] = useState<string | null>(null)
  const [influencers, setInfluencers] = useState(8)
  const [heapMB, setHeapMB] = useState<number | null>(null)
  // ⚠️ التأطير يُختبر **هنا** لأن نفس المحرّك يخدم ورقة القصة
  // (`bust`) والودجة العائمة (`full`) — وأي تغيير بالمشهد (مثل
  // تحويله لنظام يميني) لازم يتأكد بالعين إنه ما قلب الكاميرا.
  const [framing, setFraming] = useState<'bust' | 'full'>('full')

  const say = useCallback((text: string, bad = false) => {
    setLogs((prev) => [{ at: new Date().toLocaleTimeString('en-GB'), text, bad }, ...prev].slice(0, 60))
  }, [])

  // ═══ التقاط أخطاء الكونسول — (م) طلبها بالاسم ═══
  // ⚠️ نغلّف `console.error` ونرجّعه بالتنظيف. بلا هذا، «صفر أخطاء»
  // تعني «ما دقّقنا» مو «ماكو أخطاء».
  useEffect(() => {
    const original = console.error
    console.error = (...args: unknown[]) => {
      say(`console.error: ${args.map(String).join(' ')}`, true)
      original(...args)
    }
    const onErr = (e: ErrorEvent) => say(`window.error: ${e.message}`, true)
    window.addEventListener('error', onErr)
    return () => {
      console.error = original
      window.removeEventListener('error', onErr)
    }
  }, [say])

  useEffect(() => {
    let disposed = false
    const canvas = canvasRef.current
    if (!canvas) return
    // ⚠️ الاستيراد مؤجَّل: العارض ١.٢ م.ب، وما ينحمّل إلا لمن
    // المالك يفتح هاي الشاشة فعلاً.
    import('../components/entityAvatarEngine')
      .then((mod) => mod.mountAvatar(canvas, GLB, 'WALK_ALT', framing, { influencers: 8, motion: true }))
      .then((h) => {
        if (disposed) { h.dispose(); return }
        handleRef.current = h
        // ⚠️ **منفذ قياس مقصود** على شاشة المالك وحدها: يخلّي القياس
        // من الكونسول ممكناً بلا ما نزرع أزراراً لكل فحص. ماكو منه
        // خطر — الشاشة نفسها ما تنفتح إلا للمالك، ومطفَيّة بسطر.
        ;(window as unknown as { __entityLab?: AvatarHandle }).__entityLab = h
        // قياس التأطير: وين يطلع الرأس والحوض **بالبكسل** — حتى
        // «الشخصية مزيّحة» تصير رقماً مو انطباعاً.
        say(`تأطير: الرأس عند ${h.headPixel() ? `${Math.round(h.headPixel()!.x)}×${Math.round(h.headPixel()!.y)}` : '—'} من ${h.canvasSize().w}×${h.canvasSize().h}`)
        setStats({ ...h.stats })
        say(`تحمّل بـ${h.stats.loadMs} م.ث · ${h.stats.joints} مفصل · ${h.stats.clips.length} مقطع`)
        if (!h.motion) say('منظومة الحركة ما انبنت — ماكو هيكل عظمي بالملف', true)
      })
      .catch((e: unknown) => {
        setFailed(e instanceof Error ? e.message : String(e))
        say(`فشل التحميل: ${String(e)}`, true)
      })
    return () => {
      disposed = true
      handleRef.current?.dispose()
      handleRef.current = null
    }
  }, [say, framing])

  /** الذاكرة مقاسة — ⚠️ `performance.memory` كرومية فقط، ولو ماكو **«—»**. */
  const measureHeap = () => {
    const perf = performance as Performance & { memory?: { usedJSHeapSize: number } }
    if (!perf.memory) { say('الذاكرة ما تنقاس بهذا المتصفح (performance.memory مو موجود)', true); setHeapMB(null); return }
    const mb = perf.memory.usedJSHeapSize / (1024 * 1024)
    setHeapMB(mb)
    say(`الذاكرة المستعملة: ${mb.toFixed(1)} م.ب`)
  }

  const toggleInfluencers = () => {
    const h = handleRef.current
    if (!h) return
    const next = influencers === 8 ? 4 : 8
    h.setInfluencers(next)
    setInfluencers(next)
    say(`تأثيرات العظام ← ${next}${next === 4 ? ' (الافتراضي الي يقصّ صامتاً)' : ' (الكامل مثل الملف)'}`)
  }

  /** ⚠️ قياس الإشارة: مسافة **طرف السبّابة** عن مركز الزر بالسنتيمتر. */
  const pointAt = (i: number) => {
    const h = handleRef.current
    const el = btnRefs.current[i]
    const canvas = canvasRef.current
    if (!h?.motion || !el || !canvas) { say('الإشارة مو متوفرة (منظومة الحركة ما انبنت)', true); return }
    const r = el.getBoundingClientRect()
    const c = canvas.getBoundingClientRect()
    // بكسلات **نسبة للكانفس**؛ الزر برّا الكانفس يطلع رقماً سالباً
    // أو أكبر من العرض، والإسقاط العكسي يتعامل معاه عادي.
    const target = h.motion.pointAtScreen(
      (r.left + r.width / 2 - c.left) * (canvas.width / c.width),
      (r.top + r.height / 2 - c.top) * (canvas.height / c.height),
    )
    if (!target) { say('ماكو كاميرا فعّالة — القياس «—»', true); return }
    // ⚠️ نقيس **بعد** استقرار الـslerp — القياس الفوري يقيس بداية
    // الحركة مو نهايتها ويطلع خطأً وهمياً كبيراً.
    window.setTimeout(() => {
      const deg = h.motion!.pointingErrorDeg(target)
      const reach = h.motion!.armReach()
      if (deg === null) { say('ماكو عظام سبّابة أو كتف — القياس «—»', true); return }
      say(
        `إشارة للزر ${i + 1}: خطأ الاتجاه ${deg.toFixed(1)}° · مدى الذراع ${reach ? (reach * 100).toFixed(0) + ' سم' : '—'}`,
        deg > 8,
      )
    }, 1400)
  }

  /** ⚠️ قياس المشي: هل إزاحة الجذر تتضاعف ولا التحييد نافع. */
  const walkTest = async () => {
    const h = handleRef.current
    if (!h?.motion) return
    h.play('WALK_TO_TARGET' as ClipName)
    const before = h.motion.rootPosition()
    await h.motion.walkBy(1, 0, 0.9)
    const after = h.motion.rootPosition()
    const moved = Math.hypot(after.x - before.x, after.z - before.z)
    say(`مشي لمتر واحد: القطع الفعلي ${moved.toFixed(2)} م — ${moved > 1.3 ? 'يتضاعف ⚠️ التحييد ما نفع' : 'بلا مضاعفة ✅'}`, moved > 1.3)
  }

  return (
    <div className="space-y-4">
      <PageHeader
        title="مختبر الكائن"
        subtitle="تجربة معزولة للمالك — كل رقم هنا مقاس لحظة الضغط"
      />

      <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_22rem]">
        <div className="rounded-2xl bg-slate-900 p-3 shadow">
          <canvas ref={canvasRef} className="h-[26rem] w-full rounded-xl" />
          {failed && (
            <p className="mt-2 rounded-lg bg-red-50 p-3 text-sm font-semibold text-red-700">
              العارض ما اشتغل: {failed}
            </p>
          )}
        </div>

        <div className="space-y-3">
          <div className="rounded-2xl bg-white p-4 shadow">
            <h3 className="mb-2 text-sm font-bold text-slate-700">القياسات</h3>
            <dl className="space-y-1 text-sm">
              <Row k="زمن التحميل" v={stats ? `${stats.loadMs} م.ث` : '—'} />
              <Row k="المفاصل" v={stats ? String(stats.joints) : '—'} />
              <Row k="الرؤوس" v={stats ? stats.vertices.toLocaleString('en-US') : '—'} />
              <Row k="تأثيرات العظام" v={String(influencers)} />
              <Row k="الذاكرة" v={heapMB === null ? '—' : `${heapMB.toFixed(1)} م.ب`} />
              <Row k="المقاطع" v={stats ? String(stats.clips.length) : '—'} />
            </dl>
          </div>

          <div className="grid grid-cols-2 gap-2">
            <Btn onClick={measureHeap}>قِس الذاكرة</Btn>
            <Btn onClick={toggleInfluencers}>{influencers === 8 ? 'جرّب ٤ تأثيرات' : 'رجّع ٨ تأثيرات'}</Btn>
            <Btn onClick={walkTest}>اختبار المشي</Btn>
            <Btn onClick={() => { handleRef.current?.motion?.stopPointing(); handleRef.current?.motion?.lookAt(null) }}>
              وقّف الإشارة
            </Btn>
            <Btn onClick={() => setFraming((f) => (f === 'full' ? 'bust' : 'full'))}>
              {framing === 'full' ? 'تأطير الصدر (القصة)' : 'تأطير كامل (الودجة)'}
            </Btn>
          </div>

          <div className="rounded-2xl bg-white p-4 shadow">
            <h3 className="mb-2 text-sm font-bold text-slate-700">
              أزرار حقيقية — يشاور عليها بالحساب
            </h3>
            <div className="grid grid-cols-3 gap-2">
              {[0, 1, 2].map((i) => (
                <button
                  key={i}
                  ref={(el) => { btnRefs.current[i] = el }}
                  onClick={() => pointAt(i)}
                  className="rounded-xl bg-sky-600 px-3 py-4 text-sm font-bold text-white hover:bg-sky-700"
                >
                  زر {i + 1}
                </button>
              ))}
            </div>
          </div>

          <div className="rounded-2xl bg-white p-4 shadow">
            <h3 className="mb-2 text-sm font-bold text-slate-700">المقاطع</h3>
            <div className="flex flex-wrap gap-1">
              {(stats?.clips ?? []).map((c) => (
                <button
                  key={c}
                  onClick={() => handleRef.current?.play(c as ClipName)}
                  className="rounded-lg bg-slate-100 px-2 py-1 text-xs font-semibold text-slate-700 hover:bg-slate-200"
                >
                  {c}
                </button>
              ))}
              {!stats && <span className="text-xs text-slate-500">—</span>}
            </div>
          </div>
        </div>
      </div>

      <div className="rounded-2xl bg-white p-4 shadow">
        <h3 className="mb-2 text-sm font-bold text-slate-700">
          السجل وأخطاء الكونسول ({logs.filter((l) => l.bad).length} خطأ)
        </h3>
        {logs.length === 0
          ? <p className="text-xs text-slate-500">ماكو شي بعد.</p>
          : (
            <ul className="max-h-56 space-y-1 overflow-auto text-xs">
              {logs.map((l, i) => (
                <li key={i} className={l.bad ? 'font-semibold text-red-700' : 'text-slate-600'}>
                  <span className="text-slate-400">{l.at}</span> — {l.text}
                </li>
              ))}
            </ul>
          )}
      </div>
    </div>
  )
}

function Row({ k, v }: { k: string; v: string }) {
  return (
    <div className="flex justify-between border-b border-slate-100 pb-1">
      <dt className="text-slate-500">{k}</dt>
      <dd className="font-bold text-slate-800">{v}</dd>
    </div>
  )
}

function Btn({ onClick, children }: { onClick: () => void; children: React.ReactNode }) {
  return (
    <button
      onClick={onClick}
      className="rounded-xl bg-slate-800 px-3 py-2 text-xs font-bold text-white hover:bg-slate-900"
    >
      {children}
    </button>
  )
}
