import { useEffect, useMemo, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import {
  api,
  type CreateMaterialLineRequest,
  type ExecutionCostItem,
  type LeaderInvoice,
  type SystemPriceCatalog,
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
  const bookingId = params.get('bookingId') || undefined

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

  useEffect(() => {
    api.getSystemPriceCatalog().then(setCatalog)
  }, [])

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
      const payload = {
        bookingId,
        customerName: customerName || undefined,
        customerPhone: customerPhone || undefined,
        customerAddress: customerAddress || undefined,
        systems,
        items: items.map(({ key, ...rest }) => { void key; return rest }),
        materials: materials
          .filter((m) => m.quantity > 0 && (m.materialCode || m.name))
          .map(({ key, ...rest }) => { void key; return rest }),
        discountValue,
      }
      const invoice = await api.createLeaderInvoice(payload)
      setResult(invoice)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'تعذر حفظ الفاتورة')
    } finally {
      setSaving(false)
    }
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
      <h2 className="text-2xl font-bold text-brand-900">فاتورة ليدر جديدة</h2>
      <p className="mt-1 text-slate-500">
        تحل محل شيت "تكاليف المشروع" — اختر المنظومات، بنود التنفيذ، والمواد، والحساب النهائي يتم بالسيرفر.
      </p>

      <div className="mt-6 grid grid-cols-1 gap-3 sm:grid-cols-3">
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
                <select
                  value={it.itemName}
                  onChange={(e) => updateItem(it.key, { itemName: e.target.value })}
                  className="rounded-lg border border-slate-300 px-2 py-1 text-sm sm:col-span-2"
                >
                  <option value="">اختر عنصر التركيب</option>
                  {installItemsFor(systemName).map((c) => (
                    <option key={c.id} value={c.itemName}>
                      {c.itemName} ({c.value.toLocaleString()})
                    </option>
                  ))}
                </select>
                <input
                  type="number"
                  min={0}
                  placeholder="العدد"
                  value={it.count}
                  onChange={(e) => updateItem(it.key, { count: Number(e.target.value) })}
                  className="rounded-lg border border-slate-300 px-2 py-1 text-sm"
                />
                <input
                  type="number"
                  min={0}
                  placeholder="الارتفاع (م)"
                  value={it.heightMeters}
                  onChange={(e) => updateItem(it.key, { heightMeters: Number(e.target.value) })}
                  className="rounded-lg border border-slate-300 px-2 py-1 text-sm"
                />
                <select
                  value={it.wiringItemName || ''}
                  onChange={(e) => updateItem(it.key, { wiringItemName: e.target.value })}
                  className="rounded-lg border border-slate-300 px-2 py-1 text-sm"
                >
                  <option value="">بدون تسليك</option>
                  {wiringItemsFor(systemName).map((c) => (
                    <option key={c.id} value={c.itemName}>
                      {c.itemName} (×{c.value})
                    </option>
                  ))}
                </select>
                <input
                  type="number"
                  min={0}
                  placeholder="طول الكيبل (م)"
                  value={it.cableLengthMeters}
                  onChange={(e) => updateItem(it.key, { cableLengthMeters: Number(e.target.value) })}
                  className="rounded-lg border border-slate-300 px-2 py-1 text-sm"
                />
                <select
                  value={it.programmingItem || ''}
                  onChange={(e) => updateItem(it.key, { programmingItem: e.target.value })}
                  className="rounded-lg border border-slate-300 px-2 py-1 text-sm sm:col-span-2"
                >
                  <option value="">بدون برمجة</option>
                  {programmingItemsFor(systemName).map((c) => (
                    <option key={c.id} value={c.itemName}>
                      {c.itemName} ({c.value.toLocaleString()})
                    </option>
                  ))}
                </select>
                <button
                  onClick={() => removeItem(it.key)}
                  className="rounded-lg border border-red-200 px-2 py-1 text-xs font-bold text-red-600 hover:bg-red-50"
                >
                  حذف البند
                </button>
              </div>
            ))}
        </div>
      ))}

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

      {error && <p className="mt-3 text-sm font-bold text-red-600">{error}</p>}

      <button
        onClick={handleSave}
        disabled={saving}
        className="mt-6 w-full rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-3 text-sm font-bold text-white shadow-md disabled:opacity-50"
      >
        {saving ? 'جاري الحفظ...' : 'حفظ الفاتورة (يحسب السيرفر التكاليف والمجموع النهائي)'}
      </button>
    </div>
  )
}
