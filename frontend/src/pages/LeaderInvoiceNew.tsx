import { useEffect, useMemo, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import {
  api,
  type CreateMaterialLineRequest,
  type EstimateExecutionCostResponse,
  type ExecutionCostItem,
  type LeaderInvoice,
  type SystemPriceCatalog,
  type Booking,
} from '../api'

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

export default function LeaderInvoiceNew() {
  const [params] = useSearchParams()
  const navigate = useNavigate()
  // الحجز ممكن ييجي من الرابط (لما ينضغط من شاشة الحجز) أو ينختار من قائمة
  // الحجوزات المكتملة تحت — الليدر ما يحتاج يعرف رابط ولا معرّف.
  const [selectedBookingId, setSelectedBookingId] = useState(params.get('bookingId') || '')
  const bookingId = selectedBookingId || undefined
  const [completedBookings, setCompletedBookings] = useState<Booking[]>([])
  // وضع "حساب كلفة" السريع: بدون ربط بحجز ولا زبون ولا حفظ — بس رقم تقريبي
  // للليدر لما زبون يستفسر، بنفس محرك الحساب بالضبط.
  const estimateOnly = params.get('mode') === 'estimate'

  const [catalog, setCatalog] = useState<SystemPriceCatalog[]>([])
  const [systems, setSystems] = useState<string[]>([])
  const [items, setItems] = useState<DraftItem[]>([])
  const [materials, setMaterials] = useState<DraftMaterial[]>([])
  const [customerName, setCustomerName] = useState('')
  const [customerPhone, setCustomerPhone] = useState('')
  const [customerAddress, setCustomerAddress] = useState('')
  const [discountValue, setDiscountValue] = useState(0)
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
      }
      const invoice = await api.createLeaderInvoice(payload)
      setResult(invoice)
    } catch (e) {
      setError(e instanceof Error ? e.message : (estimateOnly ? 'تعذر حساب الكلفة' : 'تعذر حفظ الفاتورة'))
    } finally {
      setSaving(false)
    }
  }

  if (estimateResult) {
    return (
      <div className="mx-auto max-w-3xl">
        <h2 className="text-2xl font-bold text-brand-900">الكلفة التقريبية</h2>
        <div className="mt-4 rounded-xl border border-white bg-white p-6 text-center shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <p className="text-sm text-slate-400">تكلفة التنفيذ التقريبية</p>
          <p className="mt-1 text-3xl font-bold text-brand-800">{estimateResult.executionCost.toLocaleString()} د.ع</p>
          <p className="mt-2 text-xs text-slate-400">إجمالي عدد الأجهزة: {estimateResult.totalDeviceCount}</p>
          <p className="mt-3 text-xs text-amber-600">هذا رقم تقريبي بس للاستفسار — ما ينحفظ ولا يرتبط بأي حجز.</p>
        </div>

        {/* تفصيل الحساب — الليدر يشوف كل رقم من وين طلع بدل ما يثق برقم أعمى */}
        {estimateResult.breakdown?.length > 0 && (
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
                        نأخذ الأكبر — حسب العدد الكلي {b.wiringByDeviceCount.toLocaleString()} / حسب الطول {b.wiringByCableLength.toLocaleString()}
                        {b.cableLengthMeters > 0 && (
                          <> ({b.wiringPricePerMeter.toLocaleString()}/م × {b.cableLengthMeters}م × {b.wiringMultiplier}
                          {b.wiringHeightWeight > 1 && <> × {b.wiringHeightWeight} (ارتفاع {b.wiringHeightMeters}م)</>})</>
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
    <div className="mx-auto max-w-4xl">
      <h2 className="text-2xl font-bold text-brand-900">{estimateOnly ? 'حساب كلفة (استفسار زبون)' : 'فاتورة ليدر جديدة'}</h2>
      <p className="mt-1 text-slate-500">
        {estimateOnly
          ? 'لما زبون يستفسر عن سعر تقريبي — اختر المنظومات وبنود التنفيذ، والحساب نفس محرك الفاتورة بالضبط، بس بدون حفظ ولا ربط بحجز.'
          : 'تحل محل شيت "تكاليف المشروع" — اختر المنظومات، بنود التنفيذ، والمواد، والحساب النهائي يتم بالسيرفر.'}
      </p>

      {/* اختيار الحجز المكتمل — الليدر يلكه أسماء حجوزاته المكتملة ويسويلها
          فاتورة، ومعلومات الزبون تنملي تلقائياً منه. */}
      {!estimateOnly && (
        <div className="mt-6 rounded-xl border border-brand-100 bg-brand-50/50 p-4">
          <label className="mb-1 block text-sm font-bold text-brand-800">
            اربط الفاتورة بحجز مكتمل (اختياري)
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
              }
            }}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
          >
            <option value="">— بدون ربط بحجز (فاتورة مستقلة) —</option>
            {completedBookings.map((b) => (
              <option key={b.id} value={b.id}>
                {b.code ? `${b.code} — ` : ''}{b.customer?.name || 'بدون اسم'}
                {b.scheduledAt ? ` (${new Date(b.scheduledAt).toLocaleDateString('ar-IQ')})` : ''}
              </option>
            ))}
          </select>
          <p className="mt-1 text-xs text-slate-500">
            {completedBookings.length > 0
              ? `${completedBookings.length} حجز مكتمل — لما تختار حجز تنملي معلومات الزبون تلقائياً.`
              : 'ما اكو حجوزات مكتملة حالياً — تكدر تسوي فاتورة مستقلة وتكتب معلومات الزبون يدوياً.'}
          </p>
        </div>
      )}

      {!estimateOnly && (
        <div className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-3">
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

      <h3 className="mt-6 mb-2 font-bold text-brand-800">المنظومات (حتى 3)</h3>
      <div className="flex flex-wrap gap-2">
        {allSystemNames.map((name) => (
          <button
            key={name}
            onClick={() => toggleSystem(name)}
            className={`rounded-full px-3 py-1 text-xs font-bold ${
              systems.includes(name)
                ? 'bg-brand-700 text-white'
                : 'bg-brand-50 text-brand-700 hover:bg-brand-100'
            }`}
          >
            {name}
          </button>
        ))}
      </div>

      {systems.map((systemName) => (
        <div key={systemName} className="mt-4 rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <div className="flex items-center justify-between">
            <h4 className="font-bold text-brand-800">{systemName}</h4>
            <button
              onClick={() => addItem(systemName)}
              className="rounded-lg bg-brand-50 px-3 py-1 text-xs font-bold text-brand-700 hover:bg-brand-100"
            >
              + إضافة بند
            </button>
          </div>
          {items
            .filter((it) => it.systemName === systemName)
            .map((it) => (
              <div key={it.key} className="mt-3 grid grid-cols-1 gap-2 border-t border-slate-100 pt-3 sm:grid-cols-6">
                <div className="sm:col-span-2">
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
                  <div className="sm:col-span-2 rounded-lg bg-amber-50 px-3 py-2 text-[11px] text-amber-700">
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

                <div className="sm:col-span-2">
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

                <div className="flex items-end">
                <button
                  onClick={() => removeItem(it.key)}
                  className="w-full rounded-lg border border-red-200 px-2 py-1 text-xs font-bold text-red-600 hover:bg-red-50"
                >
                  حذف البند
                </button>
                </div>
              </div>
            ))}
        </div>
      ))}

      {!estimateOnly && (
      <div className="mt-6 rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
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
            <input
              type="number"
              min={0}
              placeholder="سعر الوحدة (لو مادة يدوية)"
              value={m.unitPrice ?? ''}
              onChange={(e) => updateMaterial(m.key, { unitPrice: Number(e.target.value) })}
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
        <div className="mt-6 grid grid-cols-1 gap-3 sm:grid-cols-3">
          <input
            type="number"
            min={0}
            placeholder="قيمة الخصم"
            value={discountValue}
            onChange={(e) => setDiscountValue(Number(e.target.value))}
            className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
          />
        </div>
      )}

      {error && <p className="mt-3 text-sm font-bold text-red-600">{error}</p>}

      <button
        onClick={handleSave}
        disabled={saving}
        className="mt-6 w-full rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-3 text-sm font-bold text-white shadow-md disabled:opacity-50"
      >
        {saving
          ? 'جاري الحساب...'
          : estimateOnly
            ? 'احسب الكلفة التقريبية'
            : 'حفظ الفاتورة (يحسب السيرفر التكاليف والمجموع النهائي)'}
      </button>
    </div>
  )
}
