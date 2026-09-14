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
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useSession } from '../session'
import PageHeader from '../components/PageHeader'
import type { AvatarHandle, AvatarStats, ClipName } from '../components/entityAvatarEngine'
import type { Tracker } from '../components/faceTracking'
import { api, entityModelUrl, ensureFileToken, type EntityAvatarModel } from '../api'
import { forgetActiveModel } from '../components/EntityAvatar'

/**
 * ⚠️ **مفتاح الإطفاء.** النظام ماكو بيه منظومة رايات ميزات (فحصتها:
 * ماكو `featureFlag` ولا `FEATURE_` بكل المستودع)، فالإطفاء يصير هنا:
 * `false` ← الشاشة تعرض «غير موجودة» للكل **بما فيهم المالك**،
 * والعارض الثقيل **ما ينحمّل إطلاقاً** (الاستيراد مؤجَّل تحت الشرط).
 */
const LAB_ENABLED = true

/**
 * المجسّمات المدمجة بالتطبيق — موجودة دايماً بلا رفع.
 *
 * ⚠️ **وأي مجسّم غيرها يُرفَع من الشاشة نفسها**، ما ينسخ بالإيد
 * للسيرفر. الطريقة القديمة (ملف ثابت بـ`public/`) كانت تطلب من مالك
 * النظام SSH ونسخاً يدوياً وبناءَ حاوية بترتيب معيّن — وهاي شغلة
 * مبرمج، وفعلاً ما انفهمت لمن طلبتها. والرفع من الشاشة يشيلها كلياً.
 */
const BUILT_IN = [
  { id: 'builtin:amani', label: 'الحالي (أماني v4)', file: 'amani-tech-v4.glb' },
] as const

/** مجسّم بالقائمة: مدمج أو مرفوع — الشاشة تتعامل معهم بنفس الشكل. */
type ModelChoice = { id: string; label: string; url: string; uploaded: boolean; sizeBytes?: number; isActive: boolean }

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
  const [modelId, setModelId] = useState<string>(BUILT_IN[0].id)
  const [uploaded, setUploaded] = useState<EntityAvatarModel[] | null>(null)
  const [uploadBusy, setUploadBusy] = useState(false)
  const fileInputRef = useRef<HTMLInputElement | null>(null)
  // ═══ الكاميرا: حالة منفصلة تماماً، ومطفَيّة بالافتراضي ═══
  // 🔒 **ما تشتغل إلا بضغط المالك**، والإطفاء يوقف المسارات فعلياً
  // (ضوء الكاميرا ينطفي) مو يخفي المعاينة بس.
  const trackerRef = useRef<Tracker | null>(null)
  const previewRef = useRef<HTMLDivElement | null>(null)
  const [camOn, setCamOn] = useState(false)
  const [camBusy, setCamBusy] = useState(false)
  const [track, setTrack] = useState<{ fps: number; ms: number; head: string; hands: number; expr: number } | null>(null)
  const [fingerReport, setFingerReport] = useState<Record<string, number> | null>(null)
  const [exprNames, setExprNames] = useState<string[] | null>(null)

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

  // ═══ المجسّمات: المدمجة + المرفوعة ═══
  // ⚠️ **الوسم يُجاب قبل القائمة**: رابط الملف المُخدَّم يحمل وسماً
  // موقَّعاً بنفسه (العارض يجيب الملف بلا ترويسة `Authorization`)،
  // وبلا الوسم يرجع ٤٠١ والمجسّم ما يُحمَّل. فالترتيب مضمون:
  // وسم ← قائمة ← روابط.
  const loadUploaded = useCallback(
    () => ensureFileToken()
      .then(() => api.getEntityModels())
      .then(setUploaded)
      .catch((e: unknown) => {
        say(`تعذر جلب المجسّمات المرفوعة: ${e instanceof Error ? e.message : String(e)}`, true)
        setUploaded([])
      }),
    [say],
  )
  useEffect(() => { loadUploaded() }, [loadUploaded])

  // ⚠️ `useMemo` إلزامي: القائمة تدخل باعتماديات مراقب التركيب، ومصفوفة
  // جديدة كل رسم تعني **إعادة تحميل المجسّم كل رسم** — وهذا مو بطء
  // بس، بل وميض بالمشهد وتحميل ٥ م.ب متكرر.
  const choices: ModelChoice[] = useMemo(() => [
    ...BUILT_IN.map((m) => ({
      id: m.id, label: m.label, uploaded: false,
      url: `${import.meta.env.BASE_URL}${m.file}`,
      // المدمج «شخصية النظام» لمّا ماكو ولا مرفوع نشط — لأن هذا
      // بالضبط شنو يشوفه الموظف وقتها.
      isActive: !(uploaded ?? []).some((u) => u.isActive),
    })),
    ...(uploaded ?? []).map((m) => ({
      id: m.id, label: m.label, uploaded: true, sizeBytes: m.sizeBytes,
      url: entityModelUrl(m.fileKey),
      isActive: m.isActive,
    })),
  ], [uploaded])

  useEffect(() => {
    let disposed = false
    const canvas = canvasRef.current
    if (!canvas) return
    // ⚠️ الاستيراد مؤجَّل: العارض ١.٢ م.ب، وما ينحمّل إلا لمن
    // المالك يفتح هاي الشاشة فعلاً.
    const chosen = choices.find((m) => m.id === modelId) ?? choices[0]
    if (!chosen) return
    const url = chosen.url
    import('../components/entityAvatarEngine')
      .then((mod) => mod.mountAvatar(canvas, url, 'WALK_ALT', framing, { influencers: 8, motion: true }))
      .then((h) => {
        if (disposed) { h.dispose(); return }
        handleRef.current = h
        // ⚠️ **منفذ قياس مقصود** على شاشة المالك وحدها: يخلّي القياس
        // من الكونسول ممكناً بلا ما نزرع أزراراً لكل فحص. ماكو منه
        // خطر — الشاشة نفسها ما تنفتح إلا للمالك، ومطفَيّة بسطر.
        ;(window as unknown as { __entityLab?: AvatarHandle }).__entityLab = h
        // قياس التأطير: وين يطلع الرأس والحوض **بالبكسل** — حتى
        // «الشخصية مزيّحة» تصير رقماً مو انطباعاً.
        // ⚠️ **«—» مو NaN**: القياس يصير قبل أول رسمة، ومصفوفة العرض
        // لسه ما انحسبت فالإسقاط يرجّع NaN. وطبع «NaN×NaN» بسجل
        // قياسات **أسوأ من ما نطبع شي** — يخلي المالك يحسب إن العارض
        // مكسور وهو سليم. فنفحص العدد ونعرض «—» مع السبب.
        const hp = h.headPixel()
        const okPixel = hp && Number.isFinite(hp.x) && Number.isFinite(hp.y)
        say(`تأطير: الرأس عند ${
          okPixel ? `${Math.round(hp.x)}×${Math.round(hp.y)}` : '— (قبل أول رسمة)'
        } من ${h.canvasSize().w}×${h.canvasSize().h}`)
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
  }, [say, framing, modelId, choices])


  const pickFile = () => fileInputRef.current?.click()
  const onFilePicked = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    e.target.value = ''            // حتى اختيار نفس الملف مرة ثانية يشتغل
    if (!file) return
    setUploadBusy(true)
    try {
      const label = file.name.replace(/\.glb$/i, '')
      const row = await api.uploadEntityModel(file, label)
      say(`رُفع «${row.label}» — ${(row.sizeBytes / 1048576).toFixed(2)} م.ب`)
      await loadUploaded()
      setModelId(row.id)           // نبدّل إله فوراً حتى يشوفه
    } catch (err) {
      say(err instanceof Error ? err.message : String(err), true)
    } finally {
      setUploadBusy(false)
    }
  }

  /**
   * يخلي هذا المجسّم **شخصية النظام** لكل موظف.
   *
   * 🔴 هذا الزر هو الي كان ناقصاً: المجسّم يُرفَع ويُقارَن هنا، بس
   * الودجة الي يشوفها الموظف كان اسم ملفها **ثابتاً بالكود** — فما
   * وصل ولا مجسّم جديد لولا موظف.
   */
  const activateModel = async (m: ModelChoice) => {
    try {
      await api.activateEntityModel(m.uploaded ? m.id : 'builtin')
      // ⚠️ ننسّي المخزون بالودجة: الرابط محفوظ بوعد على مستوى الوحدة،
      // فبلا هذا يبقى (ع) يشوف القديمة حتى بعد التفعيل ويحسبه ما اشتغل.
      forgetActiveModel()
      say(`«${m.label}» صارت شخصية النظام — تبيّن لكل الموظفين بعد إعادة تحميل الصفحة`)
      await loadUploaded()
    } catch (err) {
      say(err instanceof Error ? err.message : String(err), true)
    }
  }

  const archiveModel = async (m: ModelChoice) => {
    try {
      await api.archiveEntityModel(m.id)
      say(`أُرشف «${m.label}»`)
      if (modelId === m.id) setModelId(BUILT_IN[0].id)
      await loadUploaded()
    } catch (err) {
      say(err instanceof Error ? err.message : String(err), true)
    }
  }

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

  // ═══ تشغيل/إطفاء الكاميرا ═══
  const stopCam = useCallback(() => {
    trackerRef.current?.stop()
    trackerRef.current = null
    // نرجّع الرأس والأصابع للمقطع، وإلا تبقى على آخر قراءة مجمّدة.
    handleRef.current?.motion?.setHeadPose(null)
    setCamOn(false)
    setTrack(null)
    say('الكاميرا انطفت — المسارات موقوفة والضوء ينطفي')
  }, [say])

  const startCam = async () => {
    const h = handleRef.current
    if (!h?.motion) { say('منظومة الحركة ما انبنت — الكاميرا بلا فايدة', true); return }
    setCamBusy(true)
    try {
      // ⚠️ الاستيراد مؤجَّل: المكتبة ٨٦٠ ك.ب والـWASM ١١.٧ م.ب،
      // فما تنحمّل إلا لمن المالك يضغط الزر فعلاً.
      const mod = await import('../components/faceTracking')
      let last = 0
      const t = await mod.startTracking({
        hands: true,
        onError: (m) => say(`تتبّع: ${m}`, true),
        onReading: (r) => {
          const rig = handleRef.current?.motion
          if (!rig) return
          if (r.head) rig.setHeadPose(r.head)
          for (const hand of r.hands) {
            rig.setFingerCurl(hand.right ? 'right' : 'left', hand.curls)
          }
          // ⚠️ تحديث الحالة **مخنوق لعشر مرات بالثانية**: لو حدّثنا
          // كل إطار، إعادة رسم رياكت تصير هي عنق الزجاجة وتطلّع
          // fps واطي — فيصير القياس عن رياكت مو عن التتبّع.
          const now = performance.now()
          if (now - last < 100) return
          last = now
          setTrack({
            fps: t.fps(),
            ms: t.inferenceMs(),
            head: r.head
              ? `${((r.head.yaw * 180) / Math.PI).toFixed(0)}° / ${((r.head.pitch * 180) / Math.PI).toFixed(0)}°`
              : '—',
            hands: r.hands.length,
            expr: Object.keys(r.expressions).length,
          })
        },
      })
      trackerRef.current = t
      // معاينة للمالك حتى يشوف شنو تشوفه الكاميرا — مو مخزّنة.
      t.video.className = 'w-full rounded-xl'
      previewRef.current?.replaceChildren(t.video)
      setCamOn(true)
      say('الكاميرا اشتغلت — النموذج محمّل من نطاقنا، ولا بايت يطلع من الجهاز')
    } catch (e) {
      say(e instanceof Error ? e.message : String(e), true)
    } finally {
      setCamBusy(false)
    }
  }

  // 🔒 إطفاء إلزامي عند مغادرة الشاشة — بلاه الكاميرا تبقى شاغلة.
  useEffect(() => () => { trackerRef.current?.stop(); trackerRef.current = null }, [])

  /** يقيس أي أصابع إلها عظام فعلاً — الأداة الي تحكم على أي مجسّم. */
  const measureRig = () => {
    const rig = handleRef.current?.motion
    if (!rig) return
    const rep = rig.fingerBoneReport()
    setFingerReport(rep)
    const names = rig.expressionNames()
    setExprNames(names)
    const zero = Object.entries(rep).filter(([, n]) => n === 0).map(([k]) => k)
    say(`عظام الأصابع: ${Object.entries(rep).map(([k, n]) => `${k}=${n}`).join(' · ')}`)
    if (zero.length) say(`أصابع بلا عظام (${zero.length}): ${zero.join(', ')} — الإمساك مستحيل عليها`, true)
    say(names.length === 0
      ? 'التعابير: صفر — المجسّم بلا morph targets، فالرمشة والابتسامة مستحيلتان'
      : `التعابير المتوفرة (${names.length}): ${names.slice(0, 8).join(', ')}`,
      names.length === 0)
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
            <div className="mb-2 flex items-center justify-between gap-2">
              <h3 className="text-sm font-bold text-slate-700">المجسّم</h3>
              <button
                onClick={pickFile}
                disabled={uploadBusy}
                className="rounded-xl bg-emerald-600 px-3 py-2 text-xs font-bold text-white hover:bg-emerald-700 disabled:opacity-60"
              >
                {uploadBusy ? 'يرفع…' : '⬆ ارفع مجسّم'}
              </button>
              {/* الامتداد بالمُنتقي راحة بس — الخادم يفحص **محتوى**
                  الملف مو اسمه، لأن الاسم يتغيّر بثانية. */}
              <input
                ref={fileInputRef}
                type="file"
                accept=".glb,model/gltf-binary"
                onChange={onFilePicked}
                className="hidden"
              />
            </div>
            <p className="mb-3 text-[11px] leading-5 text-slate-500">
              تختار الملف من حاسبتك وبس — <b>ما تحتاج تنسخ شي للسيرفر</b>.
              الحد ١٠ ميغا، والصيغة <b>GLB</b>.
            </p>
            <div className="flex flex-wrap gap-2">
              {choices.map((m) => (
                <span key={m.id} className="inline-flex items-center">
                  <button
                    onClick={() => setModelId(m.id)}
                    className={`rounded-xl px-3 py-2 text-xs font-bold ${
                      modelId === m.id
                        ? 'bg-sky-600 text-white'
                        : 'bg-slate-100 text-slate-700 hover:bg-slate-200'
                    }`}
                  >
                    {m.label}
                    {m.sizeBytes ? ` · ${(m.sizeBytes / 1048576).toFixed(1)}م` : ''}
                  </button>
                  {m.isActive ? (
                    <span
                      className="ms-1 rounded-lg bg-emerald-50 px-2 py-2 text-[11px] font-bold text-emerald-700"
                      title="هاي الشخصية الي يشوفها كل الموظفين"
                    >
                      شخصية النظام
                    </span>
                  ) : (
                    <button
                      onClick={() => void activateModel(m)}
                      title="خليها شخصية النظام لكل الموظفين"
                      className="ms-1 rounded-lg bg-emerald-600 px-2 py-2 text-[11px] font-bold text-white hover:bg-emerald-700"
                    >
                      خليها شخصية النظام
                    </button>
                  )}
                  {m.uploaded && (
                    <button
                      onClick={() => void archiveModel(m)}
                      title="أرشفة"
                      className="ms-1 rounded-lg bg-slate-100 px-2 py-2 text-xs text-slate-500 hover:bg-red-50 hover:text-red-600"
                    >
                      ✕
                    </button>
                  )}
                </span>
              ))}
              {uploaded === null && <span className="text-xs text-slate-500">…</span>}
            </div>
          </div>

          <div className="rounded-2xl bg-white p-4 shadow">
            <h3 className="mb-2 text-sm font-bold text-slate-700">القياسات</h3>
            <dl className="space-y-1 text-sm">
              <Row k="زمن التحميل" v={stats ? `${stats.loadMs} م.ث` : '—'} />
              <Row k="المفاصل" v={stats ? String(stats.joints) : '—'} />
              <Row k="الرؤوس" v={stats ? stats.vertices.toLocaleString('en-US') : '—'} />
              <Row k="تأثيرات العظام" v={String(influencers)} />
              <Row k="الذاكرة" v={heapMB === null ? '—' : `${heapMB.toFixed(1)} م.ب`} />
              <Row k="المقاطع" v={stats ? String(stats.clips.length) : '—'} />
              <Row k="التعابير" v={exprNames === null ? '—' : String(exprNames.length)} />
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

          {/* ═══ الكاميرا ═══ */}
          <div className="rounded-2xl bg-white p-4 shadow">
            <h3 className="mb-1 text-sm font-bold text-slate-700">
              الكاميرا — الرأس والأصابع
            </h3>
            <p className="mb-3 text-[11px] leading-5 text-slate-500">
              النموذج والـWASM محمّلان <b>من نطاقنا</b> مو من گوگل، والاستنتاج
              كلّه <b>داخل جهازك</b>. ولا صورة تُرفع ولا إطار يُخزَّن.
            </p>
            <div className="grid grid-cols-2 gap-2">
              <Btn onClick={camOn ? stopCam : startCam}>
                {camBusy ? 'لحظة…' : camOn ? 'أطفِ الكاميرا' : 'شغّل الكاميرا'}
              </Btn>
              <Btn onClick={measureRig}>قِس الأصابع والتعابير</Btn>
            </div>
            <div ref={previewRef} className={camOn ? 'mt-3' : 'hidden'} />
            {camOn && (
              <dl className="mt-3 space-y-1 text-sm">
                <Row k="إطارات/ثانية" v={track ? String(track.fps) : '—'} />
                <Row k="زمن الاستنتاج" v={track ? `${track.ms} م.ث` : '—'} />
                <Row k="زاوية الرأس (لف/ميل)" v={track?.head ?? '—'} />
                <Row k="إيدين مكتشفة" v={track ? String(track.hands) : '—'} />
                <Row k="قنوات التعابير المقروءة" v={track ? String(track.expr) : '—'} />
              </dl>
            )}
            {exprNames !== null && exprNames.length === 0 && (
              <p className="mt-3 rounded-lg bg-amber-50 p-2 text-[11px] font-semibold leading-5 text-amber-800">
                ⚠️ التعابير تُقرأ من وجهك بس <b>ما تنطبّق</b>: هذا المجسّم فيه
                صفر تعابير. الرمشة والابتسامة تحتاج مجسّماً جديداً.
              </p>
            )}
            {fingerReport && (
              <div className="mt-3">
                <h4 className="mb-1 text-xs font-bold text-slate-600">عظام الأصابع (مقاسة)</h4>
                <div className="grid grid-cols-2 gap-x-3 gap-y-1 text-[11px]">
                  {Object.entries(fingerReport).map(([k, n]) => (
                    <div key={k} className="flex justify-between">
                      <span className="text-slate-500">{k}</span>
                      <span className={n === 0 ? 'font-bold text-red-600' : 'font-bold text-emerald-700'}>
                        {n === 0 ? 'ماكو' : `${n} عظام`}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            )}
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
