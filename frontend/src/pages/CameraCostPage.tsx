import { useEffect, useState } from 'react'
import { api, type CameraCostRow, type CameraCostExtras, type CameraCostResponse } from '../api'

// استمارة "حساب تكلفة التنفيذ" — منقولة من شيت مستقل بالاكسل خاص بمنظومة
// كاميرات المراقبة. معادلتها مختلفة تماماً عن "تكاليف المشروع": سعر أساس من
// شريحة طول الكيبل، ثم ثلاث طبقات ضرب (نوع المكان، نوع المنظومة، ارتفاع
// الكاميرا)، ثم أعمال إضافية وخصم.

const emptyRow = (): CameraCostRow => ({ normalCableMeters: 0, vipCableMeters: 0, heightAbove3m: false })

const emptyExtras = (): CameraCostExtras => ({
  screenLarge43Count: 0,
  screenSmall43Count: 0,
  rackCount: 0,
  boardCount: 0,
  vipInternetMeters: 0,
  normalInternetMeters: 0,
  programmingAmount: 0,
  otherAmount: 0,
})

const extraFields: { key: keyof CameraCostExtras; label: string; hint: string }[] = [
  { key: 'screenLarge43Count', label: 'تثبيت شاشة ٤٣ وأكبر', hint: '١٥٬٠٠٠ للوحدة' },
  { key: 'screenSmall43Count', label: 'تثبيت شاشة أصغر من ٤٣', hint: '٧٬٥٠٠ للوحدة' },
  { key: 'rackCount', label: 'تثبيت راك', hint: '١٥٬٠٠٠ للوحدة' },
  { key: 'boardCount', label: 'تثبيت بورد', hint: '٧٬٥٠٠ للوحدة' },
  { key: 'vipInternetMeters', label: 'مد كيبل انترنيت VIP (متر)', hint: '٤٠٠ للمتر' },
  { key: 'normalInternetMeters', label: 'مد كيبل انترنيت عادي (متر)', hint: '٢٠٠ للمتر' },
  { key: 'programmingAmount', label: 'برمجة (مبلغ)', hint: 'مبلغ يُدخل يدوي' },
  { key: 'otherAmount', label: 'غيرها (مبلغ)', hint: 'مبلغ يُدخل يدوي' },
]

export default function CameraCostPage() {
  const [placeTypes, setPlaceTypes] = useState<string[]>([])
  const [systemTypes, setSystemTypes] = useState<string[]>([])
  const [note, setNote] = useState('')
  const [placeType, setPlaceType] = useState('')
  const [systemType, setSystemType] = useState('')
  const [rows, setRows] = useState<CameraCostRow[]>([emptyRow()])
  const [extras, setExtras] = useState<CameraCostExtras>(emptyExtras())
  const [discount, setDiscount] = useState(0)
  const [result, setResult] = useState<CameraCostResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    api.getCameraCostOptions()
      .then((o) => {
        setPlaceTypes(o.placeTypes)
        setSystemTypes(o.systemTypes)
        setNote(o.note)
        setPlaceType(o.placeTypes[0] || '')
        setSystemType(o.systemTypes[0] || '')
      })
      .catch((e) => setError(e instanceof Error ? e.message : 'تعذر جلب الخيارات'))
  }, [])

  const updateRow = (i: number, patch: Partial<CameraCostRow>) =>
    setRows((prev) => prev.map((r, idx) => (idx === i ? { ...r, ...patch } : r)))

  const calculate = async () => {
    setLoading(true)
    setError(null)
    try {
      setResult(await api.calculateCameraCost({ placeType, systemType, rows, extras, discount }))
    } catch (e) {
      setError(e instanceof Error ? e.message : 'تعذر الحساب')
      setResult(null)
    } finally {
      setLoading(false)
    }
  }

  const num = (v: number) => v.toLocaleString('en-US')

  return (
    <div dir="rtl" className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-brand-900">حساب تكلفة التنفيذ — كاميرات المراقبة</h2>
        <p className="mt-1 text-slate-500">
          استمارة خاصة بمنظومة الكاميرات: السعر يعتمد على طول الكيبل لكل كاميرا، وينضرب بنوع المكان
          ونوع المنظومة وارتفاع الكاميرا.
        </p>
      </div>

      {error && <p className="rounded-lg bg-red-50 p-4 text-red-600">{error}</p>}

      {/* رأس الاستمارة */}
      <div className="rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        <div className="grid gap-4 sm:grid-cols-2">
          <div>
            <label className="mb-1 block text-sm font-bold text-slate-600">نوع المكان</label>
            <select
              value={placeType}
              onChange={(e) => setPlaceType(e.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
            >
              {placeTypes.map((p) => <option key={p} value={p}>{p}</option>)}
            </select>
            <p className="mt-1 text-[11px] text-slate-400">
              منزل سكني ×١ · محل تجاري ×٠٫٩٥ · مدرسة أو شركة ×١٫٣ · مصنع أو معمل ×١٫٤
            </p>
          </div>
          <div>
            <label className="mb-1 block text-sm font-bold text-slate-600">نوع المنظومة</label>
            <select
              value={systemType}
              onChange={(e) => setSystemType(e.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
            >
              {systemTypes.map((s) => <option key={s} value={s}>{s}</option>)}
            </select>
            <p className="mt-1 text-[11px] text-slate-400">ANLOGE ×١ · IP ×١٫٢</p>
          </div>
        </div>
      </div>

      {/* صفوف الكاميرات */}
      <div className="rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        <div className="mb-3 flex items-center justify-between">
          <h3 className="font-bold text-brand-800">الكاميرات</h3>
          <button
            onClick={() => setRows((p) => [...p, emptyRow()])}
            className="rounded-lg border border-brand-300 bg-brand-50 px-3 py-1.5 text-sm font-medium text-brand-700"
          >
            + كاميرا
          </button>
        </div>
        <p className="mb-3 text-[11px] text-slate-400">
          شرائح طول الكيبل: صفر ٧٬٥٠٠ · أقل من ١٠ ١٠٬٠٠٠ · أقل من ٢٠ ١٢٬٠٠٠ · أقل من ٣٠ ١٣٬٠٠٠ ·
          أقل من ٤٠ ١٤٬٠٠٠ · أقل من ٥٠ ١٥٬٠٠٠ · ٥٠ فأكثر ١٧٬٠٠٠. كيبل الـVIP ينضرب ×١٫٢.
        </p>
        <div className="space-y-3">
          {rows.map((r, i) => (
            <div key={i} className="grid gap-3 rounded-xl border border-slate-100 bg-slate-50/60 p-4 sm:grid-cols-4">
              <div className="flex items-center text-sm font-bold text-brand-900">كاميرا رقم {i + 1}</div>
              <div>
                <label className="mb-1 block text-[11px] font-bold text-slate-500">طول الكيبل عادي (م)</label>
                <input
                  type="number" min={0} value={r.normalCableMeters}
                  onChange={(e) => updateRow(i, { normalCableMeters: Number(e.target.value) })}
                  className="w-full rounded-lg border border-slate-300 px-2 py-1 text-sm"
                />
              </div>
              <div>
                <label className="mb-1 block text-[11px] font-bold text-slate-500">طول الكيبل VIP (م)</label>
                <input
                  type="number" min={0} value={r.vipCableMeters}
                  onChange={(e) => updateRow(i, { vipCableMeters: Number(e.target.value) })}
                  className="w-full rounded-lg border border-slate-300 px-2 py-1 text-sm"
                />
              </div>
              <div className="flex items-end justify-between gap-2">
                <label className="flex items-center gap-2 text-xs text-slate-600">
                  <input
                    type="checkbox" checked={r.heightAbove3m}
                    onChange={(e) => updateRow(i, { heightAbove3m: e.target.checked })}
                  />
                  أعلى من ٣ متر (×١٫١)
                </label>
                {rows.length > 1 && (
                  <button
                    onClick={() => setRows((p) => p.filter((_, idx) => idx !== i))}
                    className="text-xs text-red-500"
                  >
                    حذف
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* الأعمال الإضافية */}
      <div className="rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        <h3 className="mb-3 font-bold text-brand-800">أعمال إضافية</h3>
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          {extraFields.map((f) => (
            <div key={f.key}>
              <label className="mb-1 block text-[11px] font-bold text-slate-500">{f.label}</label>
              <input
                type="number" min={0} value={extras[f.key]}
                onChange={(e) => setExtras((p) => ({ ...p, [f.key]: Number(e.target.value) }))}
                className="w-full rounded-lg border border-slate-300 px-2 py-1 text-sm"
              />
              <p className="mt-0.5 text-[10px] text-slate-400">{f.hint}</p>
            </div>
          ))}
        </div>
        <div className="mt-4 max-w-xs">
          <label className="mb-1 block text-[11px] font-bold text-slate-500">مقدار الخصم</label>
          <input
            type="number" min={0} value={discount}
            onChange={(e) => setDiscount(Number(e.target.value))}
            className="w-full rounded-lg border border-slate-300 px-2 py-1 text-sm"
          />
        </div>
      </div>

      <button
        onClick={calculate}
        disabled={loading}
        className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-10 py-2.5 font-medium text-white shadow-md disabled:opacity-50"
      >
        {loading ? 'جاري الحساب...' : 'احسب التكلفة'}
      </button>

      {result && (
        <div className="rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <p className="text-sm text-slate-400">المبلغ النهائي</p>
          <p className="mt-1 text-3xl font-bold text-brand-800">{num(result.finalAmount)} د.ع</p>
          <div className="mt-3 flex flex-wrap gap-2 text-xs">
            <span className="rounded-md bg-blue-50 px-2 py-1 text-blue-700">عدد الكاميرات: {result.cameraCount}</span>
            <span className="rounded-md bg-slate-100 px-2 py-1 text-slate-600">الكاميرات: {num(result.camerasTotal)}</span>
            <span className="rounded-md bg-slate-100 px-2 py-1 text-slate-600">أعمال إضافية: {num(result.extrasTotal)}</span>
            {result.discount > 0 && (
              <span className="rounded-md bg-red-50 px-2 py-1 text-red-600">الخصم: −{num(result.discount)}</span>
            )}
          </div>

          <h4 className="mt-5 mb-2 text-sm font-bold text-brand-800">شلون انحسبت كل كاميرا؟</h4>
          <div className="space-y-2">
            {result.rows.map((r) => (
              <div key={r.index} className="rounded-lg border border-slate-100 bg-slate-50/60 p-3 text-xs text-slate-600">
                <span className="font-bold text-brand-900">كاميرا {r.index}: </span>
                أساس {num(r.basePrice)} × {r.placeMultiplier} (المكان) = {num(r.afterPlace)} ×{' '}
                {r.systemMultiplier} (المنظومة) = {num(r.afterSystem)} × {r.heightMultiplier} (الارتفاع) ={' '}
                <span className="font-bold text-brand-700">{num(r.total)}</span>
              </div>
            ))}
          </div>

          {note && <p className="mt-4 rounded-lg bg-amber-50 p-3 text-xs text-amber-700">📌 {note}</p>}
        </div>
      )}
    </div>
  )
}
