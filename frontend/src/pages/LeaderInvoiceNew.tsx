import { useEffect, useMemo, useState } from 'react'
import { onEnter } from '../utils/enterKey'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useSession } from '../session'
import {
  api,
  type CreateMaterialLineRequest,
  type EstimateExecutionCostResponse,
  type ExecutionCostItem,
  type LeaderInvoice,
  type SystemPriceCatalog,
  type FreeWorkReason,
  type Booking,
  type DirectedProject,
} from '../api'
import EntityIdentity from '../components/EntityIdentity'

// صفحة إنشاء فاتورة ليدر — تحل محل شيت جوجل "تكاليف المشروع" + "انشاء الفواتير":
// اختيار حتى 3 منظومات، بنود تنفيذ لكل منظومة (عدد/ارتفاع/تسليك/طول كيبل/برمجة)،
// مواد بالكود أو يدوية، خصم، وحساب المجموع الصافي وكود المحاسبة عند الحفظ
// (الحساب الفعلي والملزم يتم بالسيرفر دائماً — هذي معاينة تقريبية فقط للمستخدم).

interface DraftItem extends ExecutionCostItem {
  key: string
}

interface DraftMaterial extends CreateMaterialLineRequest {
  key: string
}

export default function LeaderInvoiceNew({ initialMode }: { initialMode?: 'estimate' | 'booking' } = {}) {
  // تفصيل الحساب (المعادلات والحدود الدنيا) للمالك ومدير النظام فقط —
  // الليدر يشوف المبلغ وبس. هذي أسعار داخلية ما تنعرض لكل من يحسب كلفة.
  // (المالك ينطبّع دوره لـADMIN بالجلسة، فالشرط يغطي الاثنين.)
  const { employee } = useSession()
  const canSeeBreakdown = employee?.role === 'ADMIN'

  const [params] = useSearchParams()
  const navigate = useNavigate()
  // الحجز ممكن ييجي من الرابط (لما ينضغط من شاشة الحجز) أو ينختار من قائمة
  // الحجوزات المكتملة تحت — الليدر ما يحتاج يعرف رابط ولا معرّف.
  const [selectedBookingId, setSelectedBookingId] = useState(params.get('bookingId') || '')
  const bookingId = selectedBookingId || undefined
  const [completedBookings, setCompletedBookings] = useState<Booking[]>([])
  // المشاريع الموجّهة لهذا الموظف — هذي هي المصدر الأساسي للفاتورة: الليدر
  // يسوي فاتورة للشغل الموجّه له، مو من قائمة عامة بكل الحجوزات.
  const [myProjects, setMyProjects] = useState<DirectedProject[]>([])
  // ═══ استفسار ولا مربوط بحجز؟ ═══
  //
  // نفس محرك الحساب بالضبط، بس فرقين:
  //   استفسار  → رقم تقريبي بس، ما ينحفظ ولا ينربط بشي. للزبون الي
  //              يسأل عن السعر قبل ما يحجز.
  //   بحجز     → نفس الحساب مربوط بحجز، وينترحّل فاتورة للمحاسب.
  //
  // ⚠️ كان الفرق **رابطين مختلفين بالقائمة** بعنوانين مختلفين لنفس
  // الشاشة — فالليدر يحتار أي وحدة يفتح، وأحياناً يفتح الغلط ويكتشف
  // بعد ما يعبّي كلشي.
  //
  // صار قرار جوّا الشاشة: يفتح مدخل واحد ويختار من فوگ.
  //
  // الافتراضي «استفسار» لأنه الأكثر استعمالاً وما يحفظ شي — والخيار
  // الي ما يحفظ أسلم كافتراضي من الي يحفظ.
  // الوضع يجي إما من مكان النداء (لما تنفتح جوّا «فواتيري» تبدي
  // مربوطة بحجز مباشرة) أو من الرابط.
  const [mode, setMode] = useState<'estimate' | 'booking'>(
    initialMode ?? (params.get('mode') === 'booking' ? 'booking' : 'estimate'),
  )
  const estimateOnly = mode === 'estimate'

  const [catalog, setCatalog] = useState<SystemPriceCatalog[]>([])
  const [systems, setSystems] = useState<string[]>([])
  const [items, setItems] = useState<DraftItem[]>([])
  const [materials, setMaterials] = useState<DraftMaterial[]>([])
  const [customerName, setCustomerName] = useState('')
  const [customerPhone, setCustomerPhone] = useState('')
  const [customerAddress, setCustomerAddress] = useState('')
  const [discountValue, setDiscountValue] = useState(0)

  // ── الشغل المجاني ──
  // أكو حجوزات تنشتغل مجاناً (ضمان، إعادة عمل، تعويض شكوى). قبل هذا
  // الليدر إما يسوي فاتورة بصفر بلا سبب — فمحد يعرف ليش بعد شهر —
  // أو ما يسوي فاتورة أصلاً، فالشغل ينضاع من السجل كله.
  const [isFree, setIsFree] = useState(false)
  const [freeReasonId, setFreeReasonId] = useState('')
  const [freeReasonNote, setFreeReasonNote] = useState('')
  const [freeReasons, setFreeReasons] = useState<FreeWorkReason[]>([])
  useEffect(() => { api.getFreeWorkReasons().then(setFreeReasons).catch(() => {}) }, [])
  const selectedReason = freeReasons.find((r) => r.id === freeReasonId)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [result, setResult] = useState<LeaderInvoice | null>(null)
  const [estimateResult, setEstimateResult] = useState<EstimateExecutionCostResponse | null>(null)

  useEffect(() => {
    api.getSystemPriceCatalog().then(setCatalog)
  }, [])

  // الحجوزات المكتملة الي يقدر الليدر يسويلها فاتورة — ما تنجلب بوضع الاستفسار
  useEffect(() => {
    if (estimateOnly) return
    api.getBookings({ status: 'COMPLETED' })
      .then(setCompletedBookings)
      .catch(() => setCompletedBookings([]))
    api.getProjectsDirectedToMe()
      .then((r) => setMyProjects(r.projects))
      .catch(() => setMyProjects([]))
  }, [estimateOnly])

  const allSystemNames = useMemo(
    () => Array.from(new Set(catalog.map((c) => c.systemName))).sort(),
    [catalog],
  )

  const installItemsFor = (systemName: string) =>
    catalog.filter((c) => c.systemName === systemName && c.category === 'install')
  const wiringItemsFor = (systemName: string) =>
    catalog.filter((c) => c.systemName === systemName && c.category === 'wiring')
  const programmingItemsFor = (systemName: string) =>
    catalog.filter((c) => c.systemName === systemName && c.category === 'programming')

  const toggleSystem = (name: string) => {
    setSystems((prev) => {
      if (prev.includes(name)) {
        setItems((its) => its.filter((it) => it.systemName !== name))
        return prev.filter((s) => s !== name)
      }
      if (prev.length >= 3) return prev
      return [...prev, name]
    })
  }

  const addItem = (systemName: string) => {
    setItems((prev) => [
      ...prev,
      {
        key: `${Date.now()}-${Math.random()}`,
        systemName,
        itemName: '',
        count: 1,
        heightMeters: 4,
        wiringItemName: '',
        wiringHeightMeters: 0,
        cableLengthMeters: 0,
        programmingItem: '',
        notes: '',
      },
    ])
  }

  const updateItem = (key: string, patch: Partial<DraftItem>) => {
    setItems((prev) => prev.map((it) => (it.key === key ? { ...it, ...patch } : it)))
  }

  const removeItem = (key: string) => setItems((prev) => prev.filter((it) => it.key !== key))

  const addMaterial = () => {
    setMaterials((prev) => [
      ...prev,
      { key: `${Date.now()}-${Math.random()}`, materialCode: '', quantity: 1 },
    ])
  }

  const updateMaterial = (key: string, patch: Partial<DraftMaterial>) => {
    setMaterials((prev) => prev.map((m) => (m.key === key ? { ...m, ...patch } : m)))
  }

  const removeMaterial = (key: string) => setMaterials((prev) => prev.filter((m) => m.key !== key))

  // ═══ الملخّص المباشر ═══
  //
  // «ملخص الكلفة التقريبية» بالتصميم — يتحدّث وأنت تعبّي.
  //
  // 🔴 القرار المهم: نجيبه **من السيرفر** مو نحسبه بالمتصفح.
  //
  // الحساب بالمتصفح أسرع، بس يخلق **رقمين مختلفين بنفس الشاشة**:
  // واحد بالملخّص وواحد بالنتيجة النهائية. والليدر ينطي الزبون
  // الرقم الي شافه أول — وإذا فرق، الشركة تخسر أو الزبون ينزعج.
  // المعادلة بالسيرفر بيها حدود دنيا لكل منظومة وأوزان ارتفاع
  // ومضاعفات تسليك، ونسخها بالواجهة يعني نسختين تفترقن أول تعديل.
  //
  // ⚠️ وننادي بعد **٦٠٠ مللي من آخر تعديل** مو بكل ضغطة زر: بلا
  // هالتأخير، كتابة «١٥» بخانة الطول تدز نداءين، والليدر الي يعبّي
  // عشر خانات يدز عشرين نداء على انترنت موبايل.
  const [livePreview, setLivePreview] = useState<EstimateExecutionCostResponse | null>(null)
  const [previewing, setPreviewing] = useState(false)
  useEffect(() => {
    // ⚠️ كلشي جوّا المؤقّت — حتى التصفير. تعديل الحالة **مباشرة** بجسم
    // الـeffect يولّد رندر متسلسل (وهذا الي يمنعه react-hooks).
    const t = setTimeout(() => {
      if (items.length === 0) { setLivePreview(null); setPreviewing(false); return }
      setPreviewing(true)
      const clean = items.map(({ key, ...rest }) => { void key; return rest })
      api.estimateLeaderInvoiceCost(clean)
        .then(setLivePreview)
        // الفشل ما ينعرض: هذا ملخّص مساعد، والرقم الرسمي يجي بالضغط
        // على «احسب». رسالة خطأ على كل تعديل ناقص تزعج بلا فايدة.
        .catch(() => setLivePreview(null))
        .finally(() => setPreviewing(false))
    }, 600)
    return () => clearTimeout(t)
  }, [items])

  // ═══ خانة «الأجهزة» بالملخّص ═══
  //
  // بالتصميم أربع خانات، وأول وحدة «الأجهزة». وهاي **مو** من محرك
  // التنفيذ — محرك التنفيذ يحسب الشغل (تركيب/تسليك/برمجة) مو ثمن
  // الجهاز نفسه. ثمن الأجهزة يجي من بنود المواد.
  //
  // ولأنه وضع الاستفسار ما بيه مواد أصلاً، الخانة تظل تحچي الصدق:
  // تنعرض «غير مشمولة» بدل رقم مخترع. رقم مخترع هنا يعني الليدر
  // ينطي الزبون سعر يشمل أجهزة ما انحسبت.
  const materialsTotal = materials.reduce(
    (n, m) => n + (Number(m.unitPrice) || 0) * (Number(m.quantity) || 0),
    0,
  )

  // ═══ تجميع الملخّص ═══
  //
  // 🔴 `?? []` مو زيادة احتياط — هاي **كانت تكسر الشاشة**:
  // Go يحوّل الـslice الفارغة لـ`null` مو `[]` بالـJSON. فأول ما
  // تضيف بند وما تختار عنصر تركيب بعد، السيرفر يرجّع
  // `breakdown: null`، و`null.reduce(...)` يرمي خطأ **بالرندر** —
  // يعني ما تطلع رسالة خطأ صغيرة، تطلع «صار خطأ غير متوقع» وتضيع
  // الشاشة كلها والبنود الي عبّاها الليدر.
  //
  // (نفس السبب موجود بالكود القديم بـ`systemMinimums?.length` — نفس
  // الفخ انلدغ بيه مرة قبل.)
  const breakdown = livePreview?.breakdown ?? []
  const summary = livePreview && {
    devicesTotal: materialsTotal,
    install: breakdown.reduce((n, b) => n + b.installTotal, 0),
    wiring: breakdown.reduce((n, b) => n + b.wiringTotal, 0),
    programming: breakdown.reduce((n, b) => n + b.programmingTotal, 0),
    total: (livePreview.executionCost || 0) + materialsTotal,
    devices: livePreview.totalDeviceCount || 0,
  }

  // ═══ حفظ كمسودة ═══
  //
  // بالتصميم زر «حفظ كمسودة». وما ينحفظ بالسيرفر عن قصد: الاستفسار
  // أصلاً ما ينحفظ (هذا تعريفه)، والفاتورة الناقصة إذا انحفظت
  // بالسيرفر تطلع بقوائم المحاسب كفاتورة حقيقية ناقصة.
  //
  // فالمسودة تظل بجهاز الليدر: يعبّي نص الحساب، يطلع لحجز، ويرجع
  // يكمّل بلا ما يعيد كلشي من الصفر.
  const DRAFT_KEY = 'leaderInvoiceDraft'
  const [draftSaved, setDraftSaved] = useState(false)
  const saveDraft = () => {
    try {
      localStorage.setItem(DRAFT_KEY, JSON.stringify({ mode, systems, items, materials }))
      setDraftSaved(true)
      setTimeout(() => setDraftSaved(false), 2500)
    } catch { /* مساحة التخزين ممتلئة — المسودة ميزة مساعدة، ما توقف الشغل */ }
  }
  const [hasDraft, setHasDraft] = useState(() => !!localStorage.getItem(DRAFT_KEY))
  const restoreDraft = () => {
    try {
      const raw = localStorage.getItem(DRAFT_KEY)
      if (!raw) return
      const d = JSON.parse(raw)
      if (d.mode) setMode(d.mode)
      if (Array.isArray(d.systems)) setSystems(d.systems)
      if (Array.isArray(d.items)) setItems(d.items)
      if (Array.isArray(d.materials)) setMaterials(d.materials)
      setHasDraft(false)
    } catch { setHasDraft(false) }
  }

  // نص المشاركة — نفس النص بالنتيجة، حتى الزبون يوصله رقم واحد
  const shareText = (total: number, deviceCount: number) => [
    'الكلفة التقريبية — شركة الأماني',
    `المبلغ: ${total.toLocaleString()} د.ع`,
    `عدد الأجهزة: ${deviceCount}`,
    '',
    'ⓘ هذا تقدير تقريبي وليس فاتورة نهائية.',
  ].join('\n')

  const shareResult = async (txt: string) => {
    try {
      if (navigator.share) await navigator.share({ text: txt })
      else { await navigator.clipboard.writeText(txt); alert('انتنسخ النص — تكدر تلصقه بالواتساب') }
    } catch { /* الموظف ألغى المشاركة — مو خطأ */ }
  }

  const handleSave = async () => {
    setError(null)
    if (systems.length === 0) {
      setError('اختر منظومة واحدة على الأقل')
      return
    }
    if (items.length === 0) {
      setError('أضف بند تنفيذ واحد على الأقل')
      return
    }
    if (isFree && !freeReasonId) {
      setError('اختار سبب المجانية')
      return
    }
    setSaving(true)
    try {
      const cleanItems = items.map(({ key, ...rest }) => { void key; return rest })
      if (estimateOnly) {
        const est = await api.estimateLeaderInvoiceCost(cleanItems)
        setEstimateResult(est)
        return
      }
      const payload = {
        bookingId,
        customerName: customerName || undefined,
        customerPhone: customerPhone || undefined,
        customerAddress: customerAddress || undefined,
        systems,
        items: cleanItems,
        materials: materials
          .filter((m) => m.quantity > 0 && (m.materialCode || m.name))
          .map(({ key, ...rest }) => { void key; return rest }),
        discountValue,
        isFree,
        freeReasonId: isFree ? freeReasonId : undefined,
        freeReasonNote: isFree ? freeReasonNote.trim() : undefined,
      }
      const invoice = await api.createLeaderInvoice(payload)
      setResult(invoice)
    } catch (e) {
      setError(e instanceof Error ? e.message : (estimateOnly ? 'تعذر حساب الكلفة' : 'تعذر حفظ الفاتورة'))
    } finally {
      setSaving(false)
    }
  }

  // ⚠️ النتيجة كانت **تستبدل الشاشة كلها**: الليدر يضغط «احسب»
  // فيختفي كلشي عبّاه ويطلع رقم لحاله. وإذا أراد يعدّل بند واحد لازم
  // يرجع ويشوف شنو كان مكتوب — أو يعيد الإدخال من الصفر.
  //
  // هسه النتيجة تطلع **فوگ الصفحة** والبيانات تبقى تحتها، فيگدر
  // يعدّل ويعيد الحساب بلا ما يضيع شي.
  if (estimateResult) {
    return (
      <div className="mx-auto max-w-3xl">
        <div className="mb-4 flex flex-wrap items-center justify-between gap-2">
          <h2 className="text-2xl font-bold text-brand-900">الكلفة التقريبية</h2>
          <button
            onClick={() => setEstimateResult(null)}
            className="rounded-lg border border-slate-300 bg-white px-4 py-2 text-xs font-bold text-slate-600 transition hover:bg-slate-50"
          >
            ← رجوع وتعديل البنود
          </button>
        </div>
        <div className="mt-4 rounded-xl border border-white bg-white p-6 text-center shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <p className="text-sm text-slate-400">تكلفة التنفيذ التقريبية</p>
          <p className="mt-1 text-3xl font-bold text-brand-800">{estimateResult.executionCost.toLocaleString()} د.ع</p>
          <p className="mt-2 text-xs text-slate-400">إجمالي عدد الأجهزة: {estimateResult.totalDeviceCount}</p>
          <p className="mt-3 text-xs text-amber-600">هذا رقم تقريبي بس للاستفسار — ما ينحفظ ولا يرتبط بأي حجز.</p>

          {/* ═══ مشاركة مع الزبون ═══
              الليدر يطلعله الرقم وهو واقف عند الزبون، والزبون يريد
              يشوفه أو ياخذه وياه. قبل، لازم يصوّر الشاشة بتلفونه
              ويدزها — أو يقرا الرقم بصوته والزبون يكتبه ويغلط.

              ⚠️ نستخدم مشاركة النظام إذا متوفرة (تلفون)، وإلا ننسخ
              للحافظة (كمبيوتر) — بدل ما نعتمد على وحدة وتفشل بالثانية. */}
          <button
            onClick={() => shareResult(shareText(estimateResult.executionCost, estimateResult.totalDeviceCount))}
            className="mt-4 w-full rounded-xl bg-emerald-700 px-4 py-3 text-sm font-bold text-white shadow-md transition hover:bg-emerald-800"
          >
            📤 مشاركة النتيجة مع الزبون
          </button>
        </div>

        {/* تفصيل الحساب — للمالك ومدير النظام فقط */}
        {canSeeBreakdown && estimateResult.breakdown?.length > 0 && (
          <div className="mt-4 rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <h3 className="mb-3 font-bold text-brand-800">شلون انحسب الرقم؟</h3>
            <div className="space-y-3">
              {estimateResult.breakdown.map((b, i) => (
                <div key={i} className="rounded-xl border border-slate-100 bg-slate-50/60 p-4 text-sm">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <span className="font-bold text-brand-900">{b.itemName}</span>
                    <span className="font-bold text-brand-700">{b.lineTotal.toLocaleString()} د.ع</span>
                  </div>
                  <div className="mt-2 space-y-1 text-xs text-slate-600">
                    <div>
                      <span className="text-slate-400">التركيب: </span>
                      {b.unitInstallPrice.toLocaleString()} × {b.count} قطعة
                      {b.heightMultiplier !== 1 && <> × {b.heightMultiplier} (ارتفاع {b.heightMeters}م)</>}
                      {' = '}<span className="font-bold">{b.installTotal.toLocaleString()}</span>
                      {b.wiringItemName && (
                        <span className="text-amber-600"> — تصفّرت لأنه البند بي تسليك (شرط الاكسل)</span>
                      )}
                    </div>
                    {b.wiringTotal > 0 && (
                      <div>
                        <span className="text-slate-400">التسليك ({b.wiringItemName}): </span>
                        نأخذ الأكبر لكل جهاز ثم × العدد ({b.count}) — حسب العدد الكلي{' '}
                        {b.wiringByDeviceCount.toLocaleString()} (بدون مضاعف نوع الكيبل) / حسب الطول{' '}
                        {b.wiringByCableLength.toLocaleString()}
                        {b.cableLengthMeters > 0 && (
                          <> ({b.wiringPricePerMeter.toLocaleString()}/م × {b.cableLengthMeters}م × {b.wiringMultiplier}
                          {b.wiringHeightWeight > 1 && <> × {b.wiringHeightWeight} (ارتفاع {b.wiringHeightMeters}م)</>}
                          {' '}× {b.count})</>
                        )}
                        {' → '}<span className="font-bold">{b.wiringTotal.toLocaleString()}</span>
                        <span className="text-slate-400"> (حسب {b.wiringBasis})</span>
                      </div>
                    )}
                    {b.programmingTotal > 0 && (
                      <div>
                        <span className="text-slate-400">البرمجة ({b.programmingItem}): </span>
                        <span className="font-bold">{b.programmingTotal.toLocaleString()}</span>
                        <span className="text-slate-400"> — سعر ثابت ما يتضاعف بالعدد</span>
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
            {/* الحدود الدنيا — الشيت يطبّقها لكل منظومة على حدة قبل الجمع */}
            {estimateResult.systemMinimums?.length > 0 && (
              <div className="mt-4 rounded-xl border border-brand-100 bg-brand-50/50 p-4">
                <h4 className="mb-2 text-sm font-bold text-brand-800">الحدود الدنيا (لكل منظومة)</h4>
                <div className="space-y-2 text-xs">
                  {estimateResult.systemMinimums.map((m) => (
                    <div key={m.systemName} className="rounded-lg bg-white p-3">
                      <div className="font-bold text-brand-900">{m.systemName}</div>
                      <div className="mt-1 text-slate-600">
                        التركيب والتسليك: محسوب {m.installWiringCalculated.toLocaleString()} · الحد الأدنى{' '}
                        {m.deviceCount} جهاز × {m.installMinimumPerDevice.toLocaleString()} ={' '}
                        {m.installMinimumTotal.toLocaleString()} →{' '}
                        <span className="font-bold">{m.installApplied.toLocaleString()}</span>
                        {m.installFloorUsed && <span className="text-amber-600"> (انطبق الحد الأدنى)</span>}
                      </div>
                      {(m.programmingCount > 0 || m.programmingCalculated > 0) && (
                        <div className="mt-1 text-slate-600">
                          البرمجة: محسوب {m.programmingCalculated.toLocaleString()} · الحد الأدنى لـ
                          {m.programmingCount} خدمة = {m.programmingMinimum.toLocaleString()} →{' '}
                          <span className="font-bold">{m.programmingApplied.toLocaleString()}</span>
                          {m.programmingFloorUsed && <span className="text-amber-600"> (انطبق الحد الأدنى)</span>}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
                <p className="mt-2 text-[11px] text-slate-500">
                  الحد الأدنى للتركيب: ١-٤ أجهزة ١٤٠٠٠/جهاز · ٥-٨ ١٢٥٠٠ · ٩-١٦ ١١٥٠٠ · ١٧ فأكثر ١٠٠٠٠.
                  الحد الأدنى للبرمجة: خدمة ١٣٥٠٠ · خدمتين ٢٤٥٠٠ · ٣ خدمات ٣٢٥٠٠ · ٤ فأكثر ٣٥٠٠٠.
                </p>
              </div>
            )}
            <p className="mt-3 text-xs text-slate-400">
              المجموع النهائي يتقرّب لأعلى لأقرب ١٠٠٠ دينار.
            </p>
          </div>
        )}
        <button
          onClick={() => { setEstimateResult(null) }}
          className="mt-4 w-full rounded-lg bg-slate-100 px-4 py-2 text-sm font-bold text-slate-700 hover:bg-slate-200"
        >
          حساب استفسار ثاني
        </button>
      </div>
    )
  }

  if (result) {
    return (
      <div className="mx-auto max-w-2xl">
        <h2 className="text-2xl font-bold text-brand-900">تم إنشاء الفاتورة</h2>
        <div className="mt-4 rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <p className="text-sm text-slate-500">كود المحاسبة</p>
          <p className="font-mono text-lg font-bold text-brand-800">{result.accountingCode}</p>
          <div className="mt-4 grid grid-cols-2 gap-3 text-sm">
            <div>
              <span className="text-slate-400">تكاليف التنفيذ: </span>
              {result.executionCost.toLocaleString()}
            </div>
            <div>
              <span className="text-slate-400">مجموع المواد: </span>
              {result.materialsTotal.toLocaleString()}
            </div>
            <div>
              <span className="text-slate-400">الخصم: </span>
              {result.discountValue.toLocaleString()}
            </div>
            <div className="font-bold text-brand-800">
              <span className="text-slate-400">المجموع الصافي: </span>
              {result.netTotal.toLocaleString()}
            </div>
          </div>
        </div>
        <button
          onClick={() => navigate('/leader-invoices')}
          className="mt-4 rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-2 text-sm font-medium text-white shadow-md"
        >
          عرض كل الفواتير
        </button>
      </div>
    )
  }

  return (
    <div className="mx-auto max-w-4xl" dir="rtl">
      <div className="flex items-center gap-3">
        <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-sky-50 text-lg sm:h-11 sm:w-11 sm:text-xl">🧮</span>
        <div className="min-w-0">
          <h2 className="text-xl font-black text-[#0f2040] sm:text-2xl">حساب الكلفة</h2>
          <p className="text-[11px] text-slate-500 sm:text-xs">نفس محرك الحساب — تختار إذا استفسار لو مربوط بحجز</p>
        </div>
      </div>

      {/* ═══ استفسار ولا حجز؟ ═══
          الاختيار من فوگ قبل أي شي، لأنه يغيّر شنو ينحفظ. لو انحط
          بالآخر، الليدر يعبّي كلشي وبعدين يكتشف إنه بالوضع الغلط. */}
      <div className="mt-4 grid grid-cols-1 gap-2.5 sm:grid-cols-2">
        {([
          { k: 'estimate', icon: '💬', title: 'استفسار زبون', desc: 'رقم تقريبي بس — ما ينحفظ ولا ينربط بحجز' },
          { k: 'booking',  icon: '🔖', title: 'مربوط بحجز',   desc: 'نفس الحساب، وينترحّل فاتورة للمحاسب' },
        ] as const).map((o) => (
          <button
            key={o.k}
            onClick={() => setMode(o.k)}
            className={`rounded-2xl border-2 p-3.5 text-right transition ${
              mode === o.k
                ? 'border-[#2c5aad] bg-sky-50 shadow-md'
                : 'border-slate-200 bg-white hover:border-slate-300'
            }`}
          >
            <div className="flex items-start justify-between gap-2">
              <div className="min-w-0">
                <p className={`text-sm font-extrabold ${mode === o.k ? 'text-[#0f2040]' : 'text-slate-700'}`}>
                  {o.icon} {o.title}
                </p>
                <p className="mt-0.5 text-[10px] leading-relaxed text-slate-500 sm:text-[11px]">{o.desc}</p>
              </div>
              <span className={`mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-full border-2 ${
                mode === o.k ? 'border-[#2c5aad] bg-[#2c5aad]' : 'border-slate-300'
              }`}>
                {mode === o.k && <span className="text-[10px] font-black text-white">✓</span>}
              </span>
            </div>
          </button>
        ))}
      </div>

      <p className="mt-3 rounded-lg bg-slate-50 px-3 py-2 text-[11px] leading-relaxed text-slate-500">
        {estimateOnly
          ? 'ⓘ استفسار: اختر نوع العمل وأضف العناصر وتطلعلك الكلفة — نفس محرك الفاتورة بالضبط، بس بلا حفظ ولا ربط.'
          : 'ⓘ مربوط بحجز: اختر الحجز أول، وبيانات الزبون تنملي لحالها، والفاتورة تترحّل للمحاسب.'}
      </p>

      {hasDraft && items.length === 0 && (
        <div className="mt-3 flex flex-wrap items-center justify-between gap-2 rounded-xl border border-amber-200 bg-amber-50 px-3 py-2">
          <p className="text-[11px] font-bold text-amber-900">📝 عندك مسودة محفوظة من آخر مرة</p>
          <div className="flex gap-2">
            <button onClick={restoreDraft} className="rounded-lg bg-amber-600 px-3 py-1 text-[11px] font-bold text-white hover:bg-amber-700">
              كمّلها
            </button>
            <button
              onClick={() => { localStorage.removeItem(DRAFT_KEY); setHasDraft(false) }}
              className="rounded-lg border border-amber-300 px-3 py-1 text-[11px] font-bold text-amber-800 hover:bg-amber-100"
            >
              احذفها
            </button>
          </div>
        </div>
      )}

      {/* ═══ خطوات الحساب ═══
          الشاشة طويلة وبيها خانات كثيرة. بلا شريط خطوات، الليدر ما
          يعرف وين واصل ولا شكد باقي عليه — فيوقف بالنص أو ينسى بند.
          الشريط يگله بالضبط بأي مرحلة هو، وتتقدّم لحالها من الحالة. */}
      <ol className="mt-4 flex items-center gap-1 overflow-x-auto rounded-2xl border border-slate-200 bg-white p-3 sm:gap-2">
        {[
          { n: 1, label: 'اختيار نوع العمل', done: systems.length > 0 },
          { n: 2, label: 'إضافة العناصر', done: items.length > 0 },
          { n: 3, label: 'مراجعة الكلفة', done: !!estimateResult },
          { n: 4, label: estimateOnly ? 'مشاركة النتيجة' : 'حفظ الفاتورة', done: false },
        ].map((st, i, arr) => {
          // الخطوة الحالية = أول وحدة ما خلصت
          const currentIdx = arr.findIndex((x) => !x.done)
          const isCurrent = i === (currentIdx === -1 ? arr.length - 1 : currentIdx)
          return (
            <li key={st.n} className="flex min-w-0 flex-1 items-center gap-1.5 sm:gap-2">
              <span className={`flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-[11px] font-black sm:h-7 sm:w-7 sm:text-xs ${
                st.done ? 'bg-emerald-500 text-white'
                : isCurrent ? 'bg-[#2c5aad] text-white'
                : 'bg-slate-100 text-slate-400'
              }`}>
                {st.done ? '✓' : st.n}
              </span>
              <span className={`truncate text-[10px] font-bold sm:text-[11px] ${
                isCurrent ? 'text-[#0f2040]' : st.done ? 'text-emerald-700' : 'text-slate-400'
              }`}>
                {st.label}
              </span>
              {i < arr.length - 1 && <span className="hidden h-px flex-1 bg-slate-200 sm:block" />}
            </li>
          )
        })}
      </ol>

      {/* ═══ عمودين: الشغل يمين والملخّص يسار ═══
          بالتصميم الملخّص جنب الحقول مو تحتهن — لأن الليدر يحتاج يشوف
          الرقم وهو يعبّي. وبالموبايل ينقلب عمود واحد والملخّص فوگ
          (order-first) حتى يبقى بأول الشاشة بلا ما ينزّل ويطلع. */}
      <div className="mt-5 grid gap-4 lg:grid-cols-[1fr_300px] lg:items-start">
        <div className="min-w-0 space-y-4">

      {/* اختيار الحجز المكتمل — الليدر يلكه أسماء حجوزاته المكتملة ويسويلها
          فاتورة، ومعلومات الزبون تنملي تلقائياً منه. */}
      {!estimateOnly && (
        <div className="rounded-xl border border-brand-100 bg-brand-50/50 p-4">
          <label className="mb-1 block text-sm font-bold text-brand-800">
            اختر الشغل الي راح تسويله فاتورة
          </label>
          <select
            value={selectedBookingId}
            onChange={(e) => {
              const id = e.target.value
              setSelectedBookingId(id)
              const b = completedBookings.find((x) => x.id === id)
              if (b) {
                setCustomerName(b.customer?.name || '')
                setCustomerPhone(b.customer?.phone || '')
                setCustomerAddress(b.customer?.location || '')
                return
              }
              const pr = myProjects.find((x) => x.bookingId === id || `project:${x.id}` === id)
              if (pr) {
                setCustomerName(pr.name || '')
                setCustomerPhone(pr.phone || '')
                setCustomerAddress(pr.location || '')
              }
            }}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
          >
            <option value="">— بدون ربط (فاتورة مستقلة) —</option>
            {myProjects.length > 0 && (
              <optgroup label="📤 المشاريع الموجّهة لي">
                {myProjects.map((p) => (
                  <option key={p.id} value={p.bookingId || `project:${p.id}`}>
                    {p.code} — {p.name}{p.stage ? ` (${p.stage})` : ''}
                  </option>
                ))}
              </optgroup>
            )}
            {completedBookings.length > 0 && (
              <optgroup label="حجوزات مكتملة">
                {completedBookings.map((b) => (
                  <option key={b.id} value={b.id}>
                    {b.code ? `${b.code} — ` : ''}{b.customer?.name || 'بدون اسم'}
                    {b.scheduledAt ? ` (${new Date(b.scheduledAt).toLocaleDateString('ar-IQ')})` : ''}
                  </option>
                ))}
              </optgroup>
            )}
          </select>
          {/* بعد الاختيار: الهوية الكاملة للحجز المربوط. الفاتورة الغلط
              تنكتب لما الليدر يختار حجز ويطلع الزبون واحد ثاني. */}
          {(() => {
            const linked = completedBookings.find((x) => x.id === selectedBookingId)
            return linked ? <EntityIdentity booking={linked} variant="full" className="mt-2" /> : null
          })()}
          <p className="mt-1 text-xs text-slate-500">
            {myProjects.length > 0
              ? `${myProjects.length} مشروع موجّه لك — لما تختار واحد تنملي معلومات الزبون تلقائياً.`
              : 'ما اكو مشروع موجّه لك حالياً — تكدر تسوي فاتورة مستقلة وتكتب معلومات الزبون يدوياً.'}
          </p>
        </div>
      )}

      {!estimateOnly && (
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
          <input
            placeholder="اسم الزبون"
            value={customerName}
            onChange={(e) => setCustomerName(e.target.value)}
            className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
          />
          <input
            placeholder="هاتف الزبون"
            value={customerPhone}
            onChange={(e) => setCustomerPhone(e.target.value)}
            className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
          />
          <input
            placeholder="العنوان"
            value={customerAddress}
            onChange={(e) => setCustomerAddress(e.target.value)}
            className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
          />
        </div>
      )}

      {/* ═══ (١) اختيار نوع العمل ═══ */}
      <div className="rounded-2xl border border-slate-200 bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.05)]">
      <h3 className="mb-1 flex flex-wrap items-center gap-2 font-bold text-brand-800">
        <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-[#2c5aad] text-[11px] font-black text-white">1</span>
        اختيار نوع العمل <span className="text-xs font-normal text-slate-400">(حتى ٣)</span>
      </h3>
      <p className="mb-3 text-[11px] text-slate-400">اختر فئة أو أكثر لإضافة العناصر المتعلقة بها.</p>
      <div className="flex flex-wrap gap-2">
        {allSystemNames.map((name) => (
          <button
            key={name}
            onClick={() => toggleSystem(name)}
            className={`rounded-xl border-2 px-3 py-1.5 text-xs font-bold transition ${
              systems.includes(name)
                ? 'border-[#2c5aad] bg-sky-50 text-[#0f2040]'
                : 'border-slate-200 bg-white text-slate-600 hover:border-slate-300'
            }`}
          >
            {systemIcon(name)} {name}
          </button>
        ))}
        {/* ═══ شبكات ═══
            «بجانب منظومة الصوت أكو خيار اسمه شبكات».
            ⚠️ ما انضافت كمنظومة عادية عن قصد: تسعير الشبكات **شرائح**
            (تسليك: ١٢٬٠٠٠ لحد ٢٠ متر وبعدها ١٬٤٠٠ للمتر؛ تنظيم الراك:
            سعر البورت يتغيّر بحجم السويتج). جدول المنظومات هنا يعرف سعر
            واحد ثابت للوحدة بس — يعني لو حطيناها هنا راح تطلع أرقام
            **غلط**. فالخيار يوديك للحاسبة الي تعرف الشرائح، ومنها تطلع
            الفاتورة بنفس الطريقة. */}
        <button
          onClick={() => navigate('/network-cost')}
          className="rounded-xl border-2 border-dashed border-brand-400 bg-white px-3 py-1.5 text-xs font-bold text-brand-700 hover:bg-brand-50"
        >
          🌐 شبكات ↗
        </button>
        {/* ═══ كاميرات المراقبة ═══
            نفس سبب الشبكات: شيت مستقل بمعادلة غير معادلة تكاليف
            المشروع. لو انحطت كمنظومة عادية هنا راح تطلع أرقام غلط. */}
        <button
          onClick={() => navigate('/camera-cost')}
          className="rounded-xl border-2 border-dashed border-brand-400 bg-white px-3 py-1.5 text-xs font-bold text-brand-700 hover:bg-brand-50"
        >
          📷 كاميرات المراقبة ↗
        </button>
      </div>
      <p className="mt-2 text-[11px] text-slate-400">
        الشبكات وكاميرات المراقبة سعرهن بالشرائح — تنحسب كل وحدة بحاسبتها الخاصة وتطلع منها الفاتورة أو استفسار الزبون.
      </p>
      </div>

      {/* ═══ (٢) العناصر المختارة ═══
          بالتصميم البنود **بطاقات** جنب بعض مو صفوف طويلة — وهذا مو
          شكل بس: الصف الطويل بالموبايل يصير عمود من ٨ خانات بلا عنوان
          يجمعهن، فالليدر يضيع أي خانة تخص أي بند. البطاقة بيها راس
          باسم البند، فتضل مفهومة حتى بشاشة ٥ إنچ. */}
      {systems.length > 0 && (
      <div className="rounded-2xl border border-slate-200 bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.05)]">
        <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
          <h3 className="flex items-center gap-2 font-bold text-brand-800">
            <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-[#2c5aad] text-[11px] font-black text-white">2</span>
            العناصر المختارة
            {items.length > 0 && <span className="text-xs font-normal text-slate-400">({items.length})</span>}
          </h3>
          {/* منظومة وحدة؟ زر واحد. أكثر؟ زر لكل منظومة — حتى الليدر
              ما يضيف بند وبعدين يكتشف إنه راح للمنظومة الغلط. */}
          <div className="flex flex-wrap gap-1.5">
            {systems.map((s) => (
              <button
                key={s}
                onClick={() => addItem(s)}
                className="rounded-lg border border-slate-200 bg-slate-50 px-2.5 py-1 text-[11px] font-bold text-brand-700 hover:bg-slate-100"
              >
                + إضافة بند{systems.length > 1 ? ` — ${s}` : ''}
              </button>
            ))}
          </div>
        </div>

        {items.length === 0 && (
          <p className="rounded-xl border border-dashed border-slate-200 bg-slate-50/60 px-3 py-6 text-center text-xs text-slate-400">
            ما أضفت ولا بند بعد — اضغط «إضافة بند» وابدأ.
          </p>
        )}

        <div className="grid grid-cols-1 gap-3 xl:grid-cols-2">
          {items.map((it) => {
            const systemName = it.systemName
            return (
              <div key={it.key} className="rounded-2xl border border-slate-200 bg-slate-50/40 p-3">
                <div className="mb-2.5 flex items-center justify-between gap-2 border-b border-slate-200 pb-2">
                  <p className="min-w-0 truncate text-xs font-extrabold text-[#0f2040]">
                    {systemIcon(systemName)} {it.itemName || `بند جديد — ${systemName}`}
                  </p>
                  <button
                    onClick={() => removeItem(it.key)}
                    aria-label="حذف البند"
                    className="shrink-0 rounded-lg px-1.5 py-0.5 text-sm text-red-500 hover:bg-red-50"
                  >
                    🗑
                  </button>
                </div>
              <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
                <div className="sm:col-span-1">
                <label className="mb-1 block text-[11px] font-bold text-slate-500">عنصر التركيب — شنو الشغلة الي راح تنعمل</label>
                <select
                  value={it.itemName}
                  onChange={(e) => updateItem(it.key, { itemName: e.target.value })}
                  className="w-full rounded-lg border border-slate-300 px-2 py-1 text-sm"
                >
                  <option value="">اختر عنصر التركيب</option>
                  {installItemsFor(systemName).map((c) => (
                    <option key={c.id} value={c.itemName}>
                      {c.itemName} ({c.value.toLocaleString()})
                    </option>
                  ))}
                </select>
                </div>

                <div>
                <label className="mb-1 block text-[11px] font-bold text-slate-500">العدد — كل قطعة تزيد السعر</label>
                <input
                  type="number"
                  min={0}
                  placeholder="العدد"
                  value={it.count}
                  onChange={(e) => updateItem(it.key, { count: Number(e.target.value) })}
                  className="w-full rounded-lg border border-slate-300 px-2 py-1 text-sm"
                />
                </div>

                <div>
                <label className="mb-1 block text-[11px] font-bold text-slate-500">الارتفاع (م) — فوق ٤م يزيد السعر</label>
                <input
                  type="number"
                  min={0}
                  placeholder="الارتفاع (م)"
                  value={it.heightMeters}
                  onChange={(e) => updateItem(it.key, { heightMeters: Number(e.target.value) })}
                  className="w-full rounded-lg border border-slate-300 px-2 py-1 text-sm"
                />
                <p className="mt-0.5 text-[10px] text-slate-400">
                  ≤٤م بدون زيادة · ٥م ×١.١٥ · ٦م ×١.٣ · ٧م ×١.٥ · ٨م ×١.٧ · فوق ٨م ×٢
                </p>
                </div>

                <div>
                <label className="mb-1 block text-[11px] font-bold text-slate-500">نوع التسليك — نوع الكيبل المستخدم</label>
                <select
                  value={it.wiringItemName || ''}
                  onChange={(e) => updateItem(it.key, { wiringItemName: e.target.value })}
                  className="w-full rounded-lg border border-slate-300 px-2 py-1 text-sm"
                >
                  <option value="">بدون تسليك</option>
                  {wiringItemsFor(systemName).map((c) => (
                    <option key={c.id} value={c.itemName}>
                      {c.itemName} (×{c.value})
                    </option>
                  ))}
                </select>
                </div>

                {it.wiringItemName && (
                  <div>
                  <label className="mb-1 block text-[11px] font-bold text-slate-500">ارتفاع التسليك (م) — قاعدة غير ارتفاع التركيب</label>
                  <input
                    type="number"
                    min={0}
                    placeholder="ارتفاع التسليك (م)"
                    value={it.wiringHeightMeters ?? 0}
                    onChange={(e) => updateItem(it.key, { wiringHeightMeters: Number(e.target.value) })}
                    className="w-full rounded-lg border border-slate-300 px-2 py-1 text-sm"
                  />
                  <p className="mt-0.5 text-[10px] text-amber-600">
                    أقل من ٥م بدون زيادة · ٥م فما فوق ×٢ مباشرة (مو متدرج مثل التركيب)
                  </p>
                  </div>
                )}

                {it.wiringItemName && (
                  <div className="sm:col-span-3 rounded-lg bg-amber-50 px-3 py-2 text-[11px] text-amber-700">
                    ⚠️ لأنك اخترت نوع تسليك لهذا البند، أجور التركيب <b>ما تنحسب</b> — تنعتمد قيمة التسليك بس
                    (المبلغ المعتمد = الأكبر بين التسليك حسب العدد الكلي والتسليك حسب الطول). هذا شرط الاكسل.
                  </div>
                )}

                <div>
                <label className="mb-1 block text-[11px] font-bold text-slate-500">طول الكيبل (م) — كل متر يزيد السعر</label>
                <input
                  type="number"
                  min={0}
                  placeholder="طول الكيبل (م)"
                  value={it.cableLengthMeters}
                  onChange={(e) => updateItem(it.key, { cableLengthMeters: Number(e.target.value) })}
                  className="w-full rounded-lg border border-slate-300 px-2 py-1 text-sm"
                />
                <p className="mt-0.5 text-[10px] text-slate-400">
                  ١-٤م: ١٠٠٠/م · ٥-١٠م: ٨٠٠/م · ١١-٢٩م: ينزل لـ٦٩٠/م · ≥٣٠م: ٧١٠/م
                </p>
                </div>

                <div>
                <label className="mb-1 block text-[11px] font-bold text-slate-500">البرمجة — خدمة إضافية بسعر ثابت</label>
                <select
                  value={it.programmingItem || ''}
                  onChange={(e) => updateItem(it.key, { programmingItem: e.target.value })}
                  className="w-full rounded-lg border border-slate-300 px-2 py-1 text-sm"
                >
                  <option value="">بدون برمجة</option>
                  {programmingItemsFor(systemName).map((c) => (
                    <option key={c.id} value={c.itemName}>
                      {c.itemName} ({c.value.toLocaleString()})
                    </option>
                  ))}
                </select>
                </div>

                {/* ملاحظة البند — ما تدخل بالحساب أبداً. الفني يوصل الموقع
                    ويعرف وين بالضبط، والمحاسب يعرف ليش السطر سعره هيج. */}
                <div className="sm:col-span-3">
                <label className="mb-1 block text-[11px] font-bold text-slate-500">ملاحظات قصيرة (اختيارية)</label>
                <input
                  value={it.notes || ''}
                  onChange={(e) => updateItem(it.key, { notes: e.target.value })}
                  placeholder="مثال: تركيب على الجدار قرب المدخل الرئيسي"
                  className="w-full rounded-lg border border-slate-300 px-2 py-1 text-sm"
                />
                </div>
              </div>
            </div>
            )
          })}
        </div>
      </div>
      )}

      {!estimateOnly && (
      <div className="rounded-2xl border border-slate-200 bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.05)]">
        <div className="flex items-center justify-between">
          <h4 className="font-bold text-brand-800">المواد</h4>
          <button
            onClick={addMaterial}
            className="rounded-lg bg-brand-50 px-3 py-1 text-xs font-bold text-brand-700 hover:bg-brand-100"
          >
            + إضافة مادة
          </button>
        </div>
        {materials.map((m) => (
          <div key={m.key} className="mt-3 grid grid-cols-1 gap-2 border-t border-slate-100 pt-3 sm:grid-cols-5">
            <input
              placeholder="كود المادة (أو اتركه فارغاً)"
              value={m.materialCode || ''}
              onChange={(e) => updateMaterial(m.key, { materialCode: e.target.value })}
              className="rounded-lg border border-slate-300 px-2 py-1 text-sm"
            />
            <input
              placeholder="اسم المادة (لو بدون كود)"
              value={m.name || ''}
              onChange={(e) => updateMaterial(m.key, { name: e.target.value })}
              className="rounded-lg border border-slate-300 px-2 py-1 text-sm"
            />
            <input
              type="number"
              min={0}
              placeholder="الكمية"
              value={m.quantity}
              onChange={(e) => updateMaterial(m.key, { quantity: Number(e.target.value) })}
              className="rounded-lg border border-slate-300 px-2 py-1 text-sm"
            />
            {/* ⚠️ Enter بآخر خانة = صف مادة جديد: الليدر يدخل عشر
                مواد، ولو لازم يمسك الماوس بين كل وحدة والثانية يصير
                إدخال الفاتورة أطول من الشغل نفسه. */}
            <input
              type="number"
              min={0}
              placeholder="سعر الوحدة (لو مادة يدوية)"
              value={m.unitPrice ?? ''}
              onChange={(e) => updateMaterial(m.key, { unitPrice: Number(e.target.value) })}
              {...onEnter(addMaterial)}
              className="rounded-lg border border-slate-300 px-2 py-1 text-sm"
            />
            <button
              onClick={() => removeMaterial(m.key)}
              className="rounded-lg border border-red-200 px-2 py-1 text-xs font-bold text-red-600 hover:bg-red-50"
            >
              حذف
            </button>
          </div>
        ))}
        {materials.length === 0 && <p className="mt-2 text-sm text-slate-400">لا توجد مواد مضافة بعد.</p>}
      </div>
      )}

      {!estimateOnly && (
        <>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
            <input
              type="number"
              min={0}
              placeholder="قيمة الخصم"
              value={discountValue}
              onChange={(e) => setDiscountValue(Number(e.target.value))}
              disabled={isFree}
              className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 disabled:bg-slate-100"
            />
          </div>

          {/* ── الشغل المجاني ── */}
          <div className={`rounded-xl border-2 p-4 ${isFree ? 'border-emerald-400 bg-emerald-50' : 'border-slate-200 bg-white'}`}>
            <label className="flex cursor-pointer items-center gap-2">
              <input
                type="checkbox"
                checked={isFree}
                onChange={(e) => { setIsFree(e.target.checked); if (!e.target.checked) { setFreeReasonId(''); setFreeReasonNote('') } }}
                className="h-4 w-4"
              />
              <span className="text-sm font-bold text-[#0f2040]">🎁 هذا الشغل مجاني — الزبون ما يدفع شي</span>
            </label>

            {isFree && (
              <div className="mt-3 space-y-2">
                <div>
                  <label className="block text-xs font-bold text-slate-600">سبب المجانية <span className="text-red-500">*</span></label>
                  <select
                    value={freeReasonId}
                    onChange={(e) => setFreeReasonId(e.target.value)}
                    className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
                  >
                    <option value="">— اختار السبب —</option>
                    {freeReasons.map((r) => (
                      <option key={r.id} value={r.id}>{r.label}</option>
                    ))}
                  </select>
                </div>
                {selectedReason?.needsNote && (
                  <div>
                    <label className="block text-xs font-bold text-slate-600">
                      وضّح أكثر <span className="text-red-500">*</span>
                    </label>
                    <input
                      value={freeReasonNote}
                      onChange={(e) => setFreeReasonNote(e.target.value)}
                      placeholder="مثال: رجعنا لأن الكاميرا الثالثة ما اشتغلت بالتركيب الأول"
                      className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
                    />
                  </div>
                )}
                {/* الفرق المهم: الكلفة تبقى مسجّلة، الزبون بس ما يدفع */}
                <p className="rounded-lg bg-white/70 px-3 py-2 text-[11px] text-emerald-900">
                  💡 الفاتورة تنحفظ بكلفتها الحقيقية (شغل الكادر والمواد) والصافي على الزبون صفر —
                  حتى نعرف بعدين شكد كلّفنا الشغل المجاني فعلاً، بدل ما يبين مجاني على الشركة بعد.
                </p>
              </div>
            )}
          </div>
        </>
      )}

        </div>

        {/* ═══ ملخّص الكلفة التقريبية ═══
            لاصق بالكمبيوتر (يضل ظاهر وأنت تنزل بالبنود)، وفوگ بالموبايل. */}
        <aside className="order-first lg:sticky lg:top-4 lg:order-last">
          <div className="rounded-2xl border-2 border-sky-200 bg-white p-4 shadow-lg">
            <div className="flex items-center justify-between gap-2">
              <p className="text-xs font-extrabold text-[#0f2040]">📋 ملخص الكلفة التقريبية</p>
              {previewing && <span className="text-[10px] text-slate-400">يحسب...</span>}
            </div>

            {summary ? (
              <>
                <div className="mt-3 space-y-1.5">
                  {/* الأجهزة أول سطر مثل التصميم — وبالاستفسار تحچي الصدق:
                      ماكو بنود مواد، فماكو ثمن أجهزة ينحسب. */}
                  <SumRow
                    label="الأجهزة"
                    value={summary.devicesTotal}
                    hint={summary.devicesTotal === 0 ? 'غير مشمولة' : undefined}
                  />
                  <SumRow label="مواد التسليك" value={summary.wiring} />
                  <SumRow label="البرمجة" value={summary.programming} />
                  <SumRow label="أجور التركيب" value={summary.install} />
                </div>
                <div className="mt-3 border-t border-slate-100 pt-3">
                  <p className="text-[10px] text-slate-500">الإجمالي التقريبي</p>
                  <p className="text-2xl font-black text-[#2c5aad]">
                    {summary.total.toLocaleString()} <span className="text-sm">د.ع</span>
                  </p>
                  <p className="mt-0.5 text-[10px] text-slate-400">عدد الأجهزة: {summary.devices}</p>
                </div>
                <p className="mt-2.5 rounded-lg bg-amber-50 px-2.5 py-1.5 text-[10px] leading-relaxed text-amber-800">
                  ⓘ هذه الكلفة تقديرية وليست فاتورة نهائية.
                </p>
              </>
            ) : (
              <p className="mt-2 text-[11px] text-slate-400">
                {previewing ? 'يحسب الكلفة...' : 'كمّل بنود التنفيذ ويطلعلك الملخّص'}
              </p>
            )}
          </div>
        </aside>
      </div>

      {error && <p className="mt-3 text-sm font-bold text-red-600">{error}</p>}

      {/* ═══ شريط الأزرار ═══
          بالتصميم ثلاثة أزرار بالأسفل. «مشاركة» و«مسودة» ينطفون لما
          ماكو شي ينشارك — زر يضغطه الليدر وما يصير شي أسوأ من زر مقفل. */}
      <div className="sticky bottom-0 z-20 mt-5 flex flex-col-reverse gap-2 border-t border-slate-200 bg-white/95 py-3 backdrop-blur sm:flex-row sm:items-center sm:justify-center">
        <button
          onClick={() => summary && shareResult(shareText(summary.total, summary.devices))}
          disabled={!summary}
          className="rounded-xl border border-emerald-300 px-4 py-2.5 text-sm font-bold text-emerald-700 transition hover:bg-emerald-50 disabled:opacity-40"
        >
          📤 مشاركة مع الزبون
        </button>
        <button
          onClick={saveDraft}
          disabled={items.length === 0}
          className="rounded-xl border border-slate-300 px-4 py-2.5 text-sm font-bold text-slate-700 transition hover:bg-slate-50 disabled:opacity-40"
        >
          {draftSaved ? '✅ انحفظت المسودة' : '💾 حفظ كمسودة'}
        </button>
        <button
          onClick={handleSave}
          disabled={saving}
          className="rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-2.5 text-sm font-bold text-white shadow-md disabled:opacity-50"
        >
          {saving
            ? 'جاري الحساب...'
            : estimateOnly
              ? '🧮 احسب الكلفة التقريبية'
              : '🧾 حفظ الفاتورة'}
        </button>
      </div>
      {!estimateOnly && (
        <p className="mt-1 text-center text-[10px] text-slate-400">
          المجموع النهائي وكود المحاسبة يحسبهن السيرفر عند الحفظ.
        </p>
      )}
    </div>
  )
}

/* ───── أيقونة المنظومة ─────
   شكل بس — بس يفرّق: الليدر يلمح الأيقونة أسرع ما يقرا الاسم، خصوصاً
   لما تكون عشر بطاقات بالشاشة. أي منظومة جديدة تطلع بأيقونة عامة
   بدل ما تنكسر. */
function systemIcon(name: string): string {
  const n = name.toLowerCase()
  if (name.includes('بصمة') || n.includes('finger')) return '🔒'
  if (name.includes('قفل') || name.includes('أقفال') || n.includes('lock')) return '🔓'
  if (name.includes('كامير') || n.includes('cam') || n.includes('ip')) return '📹'
  if (name.includes('حريق') || n.includes('fire') || name.includes('إنذار')) return '🔥'
  if (name.includes('صوت') || n.includes('audio') || n.includes('sound')) return '🎙'
  if (name.includes('شبك') || n.includes('network')) return '🌐'
  if (name.includes('شمس') || n.includes('solar')) return '☀️'
  return '🔧'
}

/* ───── خانة بملخّص الكلفة ───── */

function SumRow({ label, value, hint }: { label: string; value: number; hint?: string }) {
  return (
    <div className="flex items-center justify-between gap-2 rounded-lg bg-slate-50/70 px-2.5 py-1.5">
      <span className="text-[11px] text-slate-600">{label}</span>
      {hint ? (
        <span className="text-[10px] text-slate-400">{hint}</span>
      ) : (
        <span className="text-xs font-black text-slate-800">{value.toLocaleString()} <span className="text-[9px] font-normal text-slate-400">د.ع</span></span>
      )}
    </div>
  )
}
