import { useEffect, useState } from 'react'
import { api, type ProductRequest, type ProductProcurement, type PayerKind, type RevolvingFund } from '../api'
import { useSession } from '../session'

const money = (n: number) => n.toLocaleString('en-IQ')

type SupplierOption = { id: string; companyName: string; ownerName: string }

const emptyFulfill = {
  fundId: '', supplierId: '', spentAmount: '', reason: '',
  receiptImage: '', payerKind: 'COMPANY' as PayerKind, customerNote: '',
}

export default function ProductRequestsPage() {
  const { permissions, employee } = useSession()
  const canAdd = employee?.role === 'ADMIN' || employee?.role === 'PROCUREMENT_ADMIN' || permissions.includes('unit_technicians')
  const isAdmin = employee?.role === 'ADMIN'
  // أبو الحسابات نفسه أبو الكميات — صلاحية الدوار هي الي تحدده
  const canFulfill = isAdmin || employee?.role === 'OWNER' || permissions.includes('revolving_fund')

  const [items, setItems] = useState<ProductRequest[]>([])
  const [procurements, setProcurements] = useState<ProductProcurement[]>([])
  const [funds, setFunds] = useState<RevolvingFund[]>([])
  const [suppliers, setSuppliers] = useState<SupplierOption[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({ productName: '', specs: '', source: '', model: '', category: '', price: '' })
  const [saving, setSaving] = useState(false)
  const [busyId, setBusyId] = useState<string | null>(null)

  // تجهيز طلب: أي طلب مفتوح عليه الفورم حالياً
  const [fulfillFor, setFulfillFor] = useState<ProductRequest | null>(null)
  const [ff, setFf] = useState(emptyFulfill)

  const load = () => {
    api.getProductRequests().then((rows) => setItems(rows ?? [])).finally(() => setLoading(false))
    if (canFulfill) {
      api.getProductProcurements().then((rows) => setProcurements(rows ?? [])).catch(() => setProcurements([]))
    }
  }
  useEffect(load, [])

  // الدوارات والموردين ننزّلهم بس للي يجهّز — الباقي ما يحتاجهم
  useEffect(() => {
    if (!canFulfill) return
    api.getFunds().then((rows) => setFunds(rows ?? [])).catch(() => setFunds([]))
    api.getSupplierOptions().then((rows) => setSuppliers(rows ?? [])).catch(() => setSuppliers([]))
  }, [canFulfill])

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!form.productName.trim()) return
    setSaving(true)
    try {
      await api.createProductRequest({
        productName: form.productName.trim(),
        specs: form.specs.trim() || undefined,
        source: form.source.trim() || undefined,
        model: form.model.trim() || undefined,
        category: form.category.trim() || undefined,
        price: form.price ? Number(form.price) : undefined,
      })
      setForm({ productName: '', specs: '', source: '', model: '', category: '', price: '' })
      setShowForm(false)
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر إضافة الطلب')
    } finally {
      setSaving(false)
    }
  }

  const resolve = async (id: string, approve: boolean) => {
    setBusyId(id)
    try {
      if (approve) await api.approveProductRequest(id)
      else await api.rejectProductRequest(id)
      load()
    } finally {
      setBusyId(null)
    }
  }

  const openFulfill = (p: ProductRequest) => {
    setFulfillFor(p)
    setFf({ ...emptyFulfill, fundId: funds[0]?.id ?? '' })
  }

  const readReceipt = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    const reader = new FileReader()
    reader.onload = () => setFf((v) => ({ ...v, receiptImage: reader.result as string }))
    reader.readAsDataURL(file)
  }

  const submitFulfill = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!fulfillFor) return
    setSaving(true)
    try {
      await api.fulfillProductRequest(fulfillFor.id, {
        fundId: ff.fundId,
        supplierId: ff.supplierId,
        spentAmount: Number(ff.spentAmount),
        reason: ff.reason.trim(),
        receiptImage: ff.receiptImage,
        payerKind: ff.payerKind,
        customerNote: ff.customerNote.trim() || undefined,
      })
      setFulfillFor(null)
      setFf(emptyFulfill)
      load()
      api.getFunds().then((rows) => setFunds(rows ?? [])).catch(() => {})
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر تجهيز الطلب')
    } finally {
      setSaving(false)
    }
  }

  const settle = async (id: string) => {
    if (!confirm('تأكيد رجوع المبلغ للدوار؟')) return
    setBusyId(id)
    try {
      await api.settleProductProcurement(id)
      load()
      api.getFunds().then((rows) => setFunds(rows ?? [])).catch(() => {})
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر تسوية الطلب')
    } finally {
      setBusyId(null)
    }
  }

  const statusLabel: Record<string, string> = { PENDING: '⏳ بالانتظار', APPROVED: '✔ موافق عليه', REJECTED: '✘ مرفوض' }
  const statusColor: Record<string, string> = { PENDING: 'text-amber-600', APPROVED: 'text-emerald-600', REJECTED: 'text-red-600' }

  const inputCls = 'rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500'
  const pendingMoney = procurements.filter((p) => p.status === 'PENDING').reduce((s, p) => s + p.spentAmount, 0)

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">طلبات المنتجات</h2>
      <p className="mt-1 text-slate-500">
        التقني يطلب منتج، وأبو الحسابات (نفسه أبو الكميات) يجهّزه من الدوار — ويبقى معلّق لحد ما المبلغ يرجع للدوار.
      </p>

      {canAdd && (
        <div className="mt-4 mb-4">
          <button onClick={() => setShowForm((v) => !v)} className="rounded-lg bg-brand-500 px-4 py-2 text-sm font-bold text-white hover:bg-brand-600">
            {showForm ? '× إغلاق' : '+ طلب منتج جديد'}
          </button>
          {showForm && (
            <form onSubmit={handleCreate} className="mt-3 grid grid-cols-1 gap-3 rounded-xl border border-white bg-white p-5 shadow-sm sm:grid-cols-3">
              <input required value={form.productName} onChange={(e) => setForm({ ...form, productName: e.target.value })} placeholder="اسم المنتج" className={`${inputCls} sm:col-span-3`} />
              <input value={form.specs} onChange={(e) => setForm({ ...form, specs: e.target.value })} placeholder="المواصفات" className={inputCls} />
              <input value={form.source} onChange={(e) => setForm({ ...form, source: e.target.value })} placeholder="المصدر" className={inputCls} />
              <input value={form.model} onChange={(e) => setForm({ ...form, model: e.target.value })} placeholder="الموديل" className={inputCls} />
              <input value={form.category} onChange={(e) => setForm({ ...form, category: e.target.value })} placeholder="التصنيف/الدورة" className={inputCls} />
              <input value={form.price} type="number" onChange={(e) => setForm({ ...form, price: e.target.value })} placeholder="السعر" className={inputCls} />
              <button type="submit" disabled={saving} className="rounded-lg bg-brand-500 px-4 py-2 text-sm font-bold text-white hover:bg-brand-600 disabled:opacity-50 sm:col-span-3">
                {saving ? 'جاري الحفظ...' : 'إرسال الطلب'}
              </button>
            </form>
          )}
        </div>
      )}

      {loading && <p className="text-slate-400">جاري التحميل...</p>}
      {!loading && items.length === 0 && <p className="text-slate-400">لا توجد طلبات منتجات بعد.</p>}

      <div className="overflow-x-auto rounded-xl border border-white bg-white shadow-sm">
        <table className="w-full border-collapse text-sm">
          <thead>
            <tr className="bg-brand-900 text-white">
              <th className="p-2 text-right">المنتج</th>
              <th className="p-2 text-right">المواصفات</th>
              <th className="p-2 text-right">المصدر</th>
              <th className="p-2 text-right">الموديل</th>
              <th className="p-2 text-right">التصنيف</th>
              <th className="p-2 text-right">السعر</th>
              <th className="p-2 text-right">الحالة</th>
              <th className="p-2 text-right">طلبه</th>
              {(isAdmin || canFulfill) && <th className="p-2 text-right">إجراء</th>}
            </tr>
          </thead>
          <tbody>
            {items.map((p) => (
              <tr key={p.id} className="border-b border-slate-100">
                <td className="p-2 font-bold">{p.productName}</td>
                <td className="p-2 text-slate-500">{p.specs || '—'}</td>
                <td className="p-2 text-slate-500">{p.source || '—'}</td>
                <td className="p-2 text-slate-500">{p.model || '—'}</td>
                <td className="p-2 text-slate-500">{p.category || '—'}</td>
                <td className="p-2 text-slate-500">{p.price != null ? money(p.price) : '—'}</td>
                <td className={`p-2 font-bold ${statusColor[p.status]}`}>{statusLabel[p.status]}</td>
                <td className="p-2 text-slate-400">{p.requestedBy?.name || '—'}</td>
                {(isAdmin || canFulfill) && (
                  <td className="p-2">
                    {p.status === 'PENDING' && (
                      <div className="flex flex-wrap gap-1">
                        {canFulfill && (
                          <button onClick={() => openFulfill(p)} className="rounded bg-brand-50 px-2 py-1 text-xs font-bold text-brand-700 hover:bg-brand-100">
                            🛒 جهّز الطلب
                          </button>
                        )}
                        {isAdmin && (
                          <>
                            <button onClick={() => resolve(p.id, true)} disabled={busyId === p.id} className="rounded bg-emerald-50 px-2 py-1 text-xs font-bold text-emerald-700 hover:bg-emerald-100">✔</button>
                            <button onClick={() => resolve(p.id, false)} disabled={busyId === p.id} className="rounded bg-red-50 px-2 py-1 text-xs font-bold text-red-700 hover:bg-red-100">✘</button>
                          </>
                        )}
                      </div>
                    )}
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* فورم التجهيز — يظهر بس للي عنده صلاحية الدوار */}
      {fulfillFor && (
        <div className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/40 p-4">
          <form onSubmit={submitFulfill} className="mt-10 w-full max-w-2xl rounded-xl bg-white p-6 shadow-xl">
            <div className="flex items-start justify-between">
              <div>
                <h3 className="text-lg font-bold text-brand-900">تجهيز الطلب: {fulfillFor.productName}</h3>
                <p className="mt-1 text-sm text-slate-500">
                  المبلغ ينخصم من الدوار الحين، والطلب يبقى معلّق لحد ما المحاسب يرجّع المبلغ.
                </p>
              </div>
              <button type="button" onClick={() => setFulfillFor(null)} className="text-2xl leading-none text-slate-400">×</button>
            </div>

            <div className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
              <label className="flex flex-col gap-1 text-sm text-slate-600">
                الدوار الي انصرف منه *
                <select required value={ff.fundId} onChange={(e) => setFf({ ...ff, fundId: e.target.value })} className={inputCls}>
                  <option value="">— اختر الدوار —</option>
                  {funds.map((f) => (
                    <option key={f.id} value={f.id}>{f.name} — الرصيد {money(f.balance)}</option>
                  ))}
                </select>
              </label>

              <label className="flex flex-col gap-1 text-sm text-slate-600">
                المورد الي اشتريت منه *
                <select required value={ff.supplierId} onChange={(e) => setFf({ ...ff, supplierId: e.target.value })} className={inputCls}>
                  <option value="">— اختر من الموردين المضافين —</option>
                  {suppliers.map((s) => (
                    <option key={s.id} value={s.id}>{s.companyName} — {s.ownerName}</option>
                  ))}
                </select>
              </label>

              <label className="flex flex-col gap-1 text-sm text-slate-600">
                المبلغ المصروف *
                <input required type="number" min="1" value={ff.spentAmount} onChange={(e) => setFf({ ...ff, spentAmount: e.target.value })} className={inputCls} />
              </label>

              <label className="flex flex-col gap-1 text-sm text-slate-600">
                صورة الوصل *
                <input required={!ff.receiptImage} type="file" accept="image/*" onChange={readReceipt} className={inputCls} />
              </label>

              <label className="flex flex-col gap-1 text-sm text-slate-600 sm:col-span-2">
                سبب الصرف *
                <textarea required rows={3} value={ff.reason} onChange={(e) => setFf({ ...ff, reason: e.target.value })} className={inputCls} placeholder="ليش انصرف هذا المبلغ؟" />
              </label>

              {/* هاي الي تحدد منو يعوّض الدوار */}
              <label className="flex flex-col gap-1 text-sm text-slate-600">
                هذا المنتج لمنو؟ *
                <select value={ff.payerKind} onChange={(e) => setFf({ ...ff, payerKind: e.target.value as PayerKind })} className={inputCls}>
                  <option value="COMPANY">للشركة — الشركة تعوّض الدوار</option>
                  <option value="CUSTOMER">للزبون — الزبون يعوّض الدوار</option>
                </select>
              </label>

              {ff.payerKind === 'CUSTOMER' && (
                <label className="flex flex-col gap-1 text-sm text-slate-600">
                  الزبون / الحجز
                  <input value={ff.customerNote} onChange={(e) => setFf({ ...ff, customerNote: e.target.value })} className={inputCls} placeholder="اسم الزبون أو رقم الحجز" />
                </label>
              )}
            </div>

            {ff.receiptImage && (
              <img src={ff.receiptImage} alt="الوصل" className="mt-3 h-24 w-24 rounded-lg border border-slate-200 object-cover" />
            )}

            <div className="mt-5 flex gap-2">
              <button type="submit" disabled={saving} className="flex-1 rounded-lg bg-brand-500 px-4 py-2 text-sm font-bold text-white hover:bg-brand-600 disabled:opacity-50">
                {saving ? 'جاري التجهيز...' : 'جهّز واصرف من الدوار'}
              </button>
              <button type="button" onClick={() => setFulfillFor(null)} className="rounded-lg border border-slate-300 px-4 py-2 text-sm">إلغاء</button>
            </div>
          </form>
        </div>
      )}

      {/* التجهيزات: المعلّق يعني الدوار ناقص */}
      {canFulfill && (
        <div className="mt-8">
          <div className="flex flex-wrap items-baseline justify-between gap-2">
            <h3 className="text-lg font-bold text-brand-900">تجهيزات المنتجات من الدوار</h3>
            <span className="rounded-lg bg-amber-50 px-3 py-1 text-sm font-bold text-amber-700">
              معلّق من الدوار: {money(pendingMoney)} د.ع
            </span>
          </div>
          <div className="mt-3 overflow-x-auto rounded-xl border border-white bg-white shadow-sm">
            <table className="w-full border-collapse text-sm">
              <thead>
                <tr className="bg-brand-900 text-white">
                  <th className="p-2 text-right">المنتج</th>
                  <th className="p-2 text-right">المورد</th>
                  <th className="p-2 text-right">الدوار</th>
                  <th className="p-2 text-right">المبلغ المصروف</th>
                  <th className="p-2 text-right">السبب</th>
                  <th className="p-2 text-right">لمنو</th>
                  <th className="p-2 text-right">الوصل</th>
                  <th className="p-2 text-right">جهّزه</th>
                  <th className="p-2 text-right">الحالة</th>
                  <th className="p-2 text-right">إجراء</th>
                </tr>
              </thead>
              <tbody>
                {procurements.map((p) => (
                  <tr key={p.id} className="border-b border-slate-100">
                    <td className="p-2 font-bold">{p.productName}</td>
                    <td className="p-2 text-slate-500">{p.supplierName}</td>
                    <td className="p-2 text-slate-500">{p.fundName}</td>
                    <td className="p-2 font-bold text-slate-700">{money(p.spentAmount)}</td>
                    <td className="p-2 text-slate-500">{p.reason}</td>
                    <td className="p-2 text-slate-500">
                      {p.payerLabel}
                      {p.customerNote && <span className="block text-xs text-slate-400">{p.customerNote}</span>}
                    </td>
                    <td className="p-2">
                      {p.receiptImage
                        ? <a href={p.receiptImage} target="_blank" rel="noreferrer" className="text-brand-600 underline">عرض</a>
                        : '—'}
                    </td>
                    <td className="p-2 text-slate-400">{p.purchasedByName}</td>
                    <td className={`p-2 font-bold ${p.status === 'PENDING' ? 'text-amber-600' : 'text-emerald-600'}`}>
                      {p.statusLabel}
                      {p.settledByName && <span className="block text-xs font-normal text-slate-400">رجّعه: {p.settledByName}</span>}
                    </td>
                    <td className="p-2">
                      {p.status === 'PENDING' && (
                        <button onClick={() => settle(p.id)} disabled={busyId === p.id} className="rounded bg-emerald-50 px-2 py-1 text-xs font-bold text-emerald-700 hover:bg-emerald-100 disabled:opacity-50">
                          💵 رجّعت المبلغ للدوار
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
                {procurements.length === 0 && (
                  <tr><td colSpan={10} className="p-6 text-center text-slate-400">ماكو تجهيزات بعد</td></tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}
