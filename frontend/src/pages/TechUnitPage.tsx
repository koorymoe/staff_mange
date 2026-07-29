import { useEffect, useRef, useState } from 'react'
import { api, type Exhibition, type Employee, type ProductRequest, type ServiceStudy } from '../api'
import { useSession } from '../session'
import TrainingManagement from './TrainingManagement'
import TechShowcasePage from './TechShowcasePage'

function fileToBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result as string)
    reader.onerror = reject
    reader.readAsDataURL(file)
  })
}

function splitList(v: string) {
  return v.split(',').map((s) => s.trim()).filter(Boolean)
}

// ————————————————————————————————————————————————— إدارة المعارض
function ExhibitionsTab({ canAdd, isAdmin }: { canAdd: boolean; isAdmin: boolean }) {
  const [items, setItems] = useState<Exhibition[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({ title: '', location: '', startDate: '', endDate: '', companies: '', productsToShow: '' })
  const [saving, setSaving] = useState(false)
  const [busyId, setBusyId] = useState<string | null>(null)
  const [findingsDraft, setFindingsDraft] = useState<Record<string, string>>({})
  const fileRefs = useRef<Record<string, HTMLInputElement | null>>({})

  const load = () => {
    api.getExhibitions().then(setItems).finally(() => setLoading(false))
  }
  useEffect(load, [])
  useEffect(() => { if (isAdmin) api.getEmployees().then(setEmployees) }, [isAdmin])

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!form.title.trim()) return
    setSaving(true)
    try {
      await api.createExhibition({
        title: form.title.trim(), location: form.location.trim(),
        startDate: form.startDate, endDate: form.endDate,
        companies: splitList(form.companies), productsToShow: splitList(form.productsToShow),
      })
      setForm({ title: '', location: '', startDate: '', endDate: '', companies: '', productsToShow: '' })
      setShowForm(false)
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر إضافة المعرض')
    } finally {
      setSaving(false)
    }
  }

  const handleNominate = async (id: string, current: string[]) => {
    const options = employees.map((e) => `${e.name} :: ${e.id}`).join('\n')
    const picked = prompt(`اكتب أسماء الموظفين المرشّحين مفصولين بفاصلة (اختر من):\n${options}`, current.join(', '))
    if (picked === null) return
    const names = splitList(picked)
    const ids = employees.filter((e) => names.includes(e.name)).map((e) => e.id)
    setBusyId(id)
    try {
      await api.nominateExhibition(id, ids)
      load()
    } finally {
      setBusyId(null)
    }
  }

  const handleUploadPhotos = async (id: string, files: FileList | null) => {
    if (!files || files.length === 0) return
    setBusyId(id)
    try {
      const base64s = await Promise.all(Array.from(files).map(fileToBase64))
      await api.addExhibitionPhotos(id, base64s)
      load()
    } finally {
      setBusyId(null)
    }
  }

  const handleSaveFindings = async (id: string) => {
    setBusyId(id)
    try {
      await api.setExhibitionFindings(id, findingsDraft[id] ?? '')
      load()
    } finally {
      setBusyId(null)
    }
  }

  const handleGenerateReport = async (id: string) => {
    setBusyId(id)
    try {
      await api.generateExhibitionReport(id)
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر توليد التقرير')
    } finally {
      setBusyId(null)
    }
  }

  const handleArchive = async (id: string) => {
    setBusyId(id)
    try {
      await api.archiveExhibition(id)
      load()
    } finally {
      setBusyId(null)
    }
  }

  return (
    <div>
      {canAdd && (
        <div className="mb-4">
          <button onClick={() => setShowForm((v) => !v)} className="rounded-lg bg-brand-500 px-4 py-2 text-sm font-bold text-white hover:bg-brand-600">
            {showForm ? '× إغلاق' : '+ إضافة معرض جديد'}
          </button>
          {showForm && (
            <form onSubmit={handleCreate} className="mt-3 grid grid-cols-1 gap-3 rounded-xl border border-white bg-white p-5 shadow-sm sm:grid-cols-2">
              <input required value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} placeholder="عنوان المعرض" className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500" />
              <input required value={form.location} onChange={(e) => setForm({ ...form, location: e.target.value })} placeholder="مكان المعرض" className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500" />
              <input required type="date" value={form.startDate} onChange={(e) => setForm({ ...form, startDate: e.target.value })} className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500" />
              <input required type="date" value={form.endDate} onChange={(e) => setForm({ ...form, endDate: e.target.value })} className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500" />
              <input value={form.companies} onChange={(e) => setForm({ ...form, companies: e.target.value })} placeholder="الشركات الموجودة بالمعرض (مفصولة بفاصلة)" className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 sm:col-span-2" />
              <input value={form.productsToShow} onChange={(e) => setForm({ ...form, productsToShow: e.target.value })} placeholder="المنتجات المراد عرضها (مفصولة بفاصلة)" className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 sm:col-span-2" />
              <button type="submit" disabled={saving} className="rounded-lg bg-brand-500 px-4 py-2 text-sm font-bold text-white hover:bg-brand-600 disabled:opacity-50 sm:col-span-2">
                {saving ? 'جاري الحفظ...' : 'حفظ المعرض'}
              </button>
            </form>
          )}
        </div>
      )}

      {loading && <p className="text-slate-400">جاري التحميل...</p>}
      {!loading && items.length === 0 && <p className="text-slate-400">لا توجد معارض بعد.</p>}

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {items.map((ex) => (
          <div key={ex.id} className={`rounded-xl border p-4 shadow-sm ${ex.archived ? 'border-slate-200 bg-slate-50' : 'border-white bg-white'}`}>
            <div className="flex items-start justify-between">
              <div>
                <p className="font-bold text-brand-900">{ex.title} {ex.archived && <span className="text-xs font-normal text-slate-400">(مؤرشف)</span>}</p>
                <p className="text-sm text-slate-500">📍 {ex.location} — {ex.startDate} → {ex.endDate}</p>
              </div>
              {isAdmin && !ex.archived && (
                <button onClick={() => handleArchive(ex.id)} disabled={busyId === ex.id} className="rounded-lg bg-slate-100 px-2 py-1 text-xs font-bold text-slate-600 hover:bg-slate-200">
                  🗄 أرشفة
                </button>
              )}
            </div>

            {ex.companies.length > 0 && <p className="mt-2 text-xs text-slate-600"><b>الشركات:</b> {ex.companies.join('، ')}</p>}
            {ex.productsToShow.length > 0 && <p className="mt-1 text-xs text-slate-600"><b>المنتجات المعروضة:</b> {ex.productsToShow.join('، ')}</p>}

            <div className="mt-2 flex items-center gap-2 text-xs">
              <b className="text-slate-600">الترشيح:</b>
              <span className="text-slate-500">{ex.nominatedEmployees.length > 0 ? ex.nominatedEmployees.map((e) => e.name).join('، ') : 'لا يوجد بعد'}</span>
              {isAdmin && (
                <button onClick={() => handleNominate(ex.id, ex.nominatedEmployees.map((e) => e.name))} disabled={busyId === ex.id} className="rounded bg-brand-50 px-2 py-0.5 font-bold text-brand-700 hover:bg-brand-100">
                  تعديل
                </button>
              )}
            </div>

            {ex.businessCardPhotos.length > 0 && (
              <div className="mt-2 flex flex-wrap gap-2">
                {ex.businessCardPhotos.map((url, i) => (
                  <img key={i} src={url} className="h-16 w-16 rounded-lg border object-cover" />
                ))}
              </div>
            )}

            {canAdd && (
              <div className="mt-3 flex flex-wrap items-center gap-2">
                <button onClick={() => fileRefs.current[ex.id]?.click()} disabled={busyId === ex.id} className="rounded-lg bg-brand-50 px-3 py-1.5 text-xs font-bold text-brand-700 hover:bg-brand-100 disabled:opacity-50">
                  📇 رفع صور كارتات
                </button>
                <input ref={(el) => { fileRefs.current[ex.id] = el }} type="file" accept="image/*" multiple className="hidden" onChange={(e) => handleUploadPhotos(ex.id, e.target.files)} />
              </div>
            )}

            {canAdd && (
              <div className="mt-3">
                <label className="text-xs font-bold text-slate-600">أهم ما اكتُشف بالمعرض</label>
                <textarea
                  defaultValue={ex.keyFindings ?? ''}
                  onChange={(e) => setFindingsDraft((d) => ({ ...d, [ex.id]: e.target.value }))}
                  rows={2}
                  className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                />
                <div className="mt-1 flex gap-2">
                  <button onClick={() => handleSaveFindings(ex.id)} disabled={busyId === ex.id} className="rounded-lg bg-slate-100 px-3 py-1.5 text-xs font-bold text-slate-700 hover:bg-slate-200 disabled:opacity-50">
                    حفظ الملاحظات
                  </button>
                  <button onClick={() => handleGenerateReport(ex.id)} disabled={busyId === ex.id} className="rounded-lg bg-brand-500 px-3 py-1.5 text-xs font-bold text-white hover:bg-brand-600 disabled:opacity-50">
                    {busyId === ex.id ? 'جاري التوليد...' : '🤖 توليد تقرير الزيارة'}
                  </button>
                </div>
              </div>
            )}

            {ex.visitReport && (
              <div className="mt-3 whitespace-pre-wrap rounded-lg bg-brand-50 p-3 text-xs leading-relaxed text-brand-900">
                <b>تقرير الزيارة:</b>
                <p className="mt-1">{ex.visitReport}</p>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}

// ————————————————————————————————————————————————— إدارة المنتجات
function ProductsTab({ canAdd, isAdmin }: { canAdd: boolean; isAdmin: boolean }) {
  const [items, setItems] = useState<ProductRequest[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({ productName: '', specs: '', source: '', model: '', category: '', price: '' })
  const [saving, setSaving] = useState(false)
  const [busyId, setBusyId] = useState<string | null>(null)

  const load = () => { api.getProductRequests().then(setItems).finally(() => setLoading(false)) }
  useEffect(load, [])

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

  const statusLabel: Record<string, string> = { PENDING: '⏳ بالانتظار', APPROVED: '✔ موافق عليه', REJECTED: '✘ مرفوض' }
  const statusColor: Record<string, string> = { PENDING: 'text-amber-600', APPROVED: 'text-emerald-600', REJECTED: 'text-red-600' }

  return (
    <div>
      {canAdd && (
        <div className="mb-4">
          <button onClick={() => setShowForm((v) => !v)} className="rounded-lg bg-brand-500 px-4 py-2 text-sm font-bold text-white hover:bg-brand-600">
            {showForm ? '× إغلاق' : '+ طلب منتج جديد'}
          </button>
          {showForm && (
            <form onSubmit={handleCreate} className="mt-3 grid grid-cols-1 gap-3 rounded-xl border border-white bg-white p-5 shadow-sm sm:grid-cols-3">
              <input required value={form.productName} onChange={(e) => setForm({ ...form, productName: e.target.value })} placeholder="اسم المنتج" className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 sm:col-span-3" />
              <input value={form.specs} onChange={(e) => setForm({ ...form, specs: e.target.value })} placeholder="المواصفات" className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500" />
              <input value={form.source} onChange={(e) => setForm({ ...form, source: e.target.value })} placeholder="المصدر" className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500" />
              <input value={form.model} onChange={(e) => setForm({ ...form, model: e.target.value })} placeholder="الموديل" className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500" />
              <input value={form.category} onChange={(e) => setForm({ ...form, category: e.target.value })} placeholder="التصنيف/الدورة" className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500" />
              <input value={form.price} type="number" onChange={(e) => setForm({ ...form, price: e.target.value })} placeholder="السعر" className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500" />
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
              {isAdmin && <th className="p-2 text-right">إجراء</th>}
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
                <td className="p-2 text-slate-500">{p.price != null ? p.price.toLocaleString('en-IQ') : '—'}</td>
                <td className={`p-2 font-bold ${statusColor[p.status]}`}>{statusLabel[p.status]}</td>
                <td className="p-2 text-slate-400">{p.requestedBy?.name || '—'}</td>
                {isAdmin && (
                  <td className="p-2">
                    {p.status === 'PENDING' && (
                      <div className="flex gap-1">
                        <button onClick={() => resolve(p.id, true)} disabled={busyId === p.id} className="rounded bg-emerald-50 px-2 py-1 text-xs font-bold text-emerald-700 hover:bg-emerald-100">✔</button>
                        <button onClick={() => resolve(p.id, false)} disabled={busyId === p.id} className="rounded bg-red-50 px-2 py-1 text-xs font-bold text-red-700 hover:bg-red-100">✘</button>
                      </div>
                    )}
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

// ————————————————————————————————————————————————— إدارة الخدمات
function ServicesTab({ canAdd, isAdmin, employeeId }: { canAdd: boolean; isAdmin: boolean; employeeId?: string }) {
  const [items, setItems] = useState<ServiceStudy[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [loading, setLoading] = useState(true)
  const [newName, setNewName] = useState('')
  const [saving, setSaving] = useState(false)
  const [busyId, setBusyId] = useState<string | null>(null)
  const [reportDraft, setReportDraft] = useState<Record<string, string>>({})

  const load = () => { api.getServiceStudies().then(setItems).finally(() => setLoading(false)) }
  useEffect(load, [])
  useEffect(() => { if (isAdmin) api.getEmployees().then(setEmployees) }, [isAdmin])

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newName.trim()) return
    setSaving(true)
    try {
      await api.createServiceStudy(newName.trim())
      setNewName('')
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر إضافة الخدمة')
    } finally {
      setSaving(false)
    }
  }

  const handleAssign = async (id: string, current: string[]) => {
    const options = employees.map((e) => `${e.name} :: ${e.id}`).join('\n')
    const picked = prompt(`اكتب أسماء التقنيين الموكَّلين مفصولين بفاصلة (اختر من):\n${options}`, current.join(', '))
    if (picked === null) return
    const names = splitList(picked)
    const ids = employees.filter((e) => names.includes(e.name)).map((e) => e.id)
    setBusyId(id)
    try {
      await api.assignServiceStudy(id, ids)
      load()
    } finally {
      setBusyId(null)
    }
  }

  const handleAddReport = async (id: string) => {
    const content = reportDraft[id]?.trim()
    if (!content) return
    setBusyId(id)
    try {
      await api.addServiceStudyReport(id, content)
      setReportDraft((d) => ({ ...d, [id]: '' }))
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر رفع التقرير')
    } finally {
      setBusyId(null)
    }
  }

  const handleArchive = async (id: string) => {
    setBusyId(id)
    try {
      await api.archiveServiceStudy(id)
      load()
    } finally {
      setBusyId(null)
    }
  }

  return (
    <div>
      {canAdd && (
        <form onSubmit={handleCreate} className="mb-4 flex gap-2 rounded-xl border border-white bg-white p-4 shadow-sm">
          <input value={newName} onChange={(e) => setNewName(e.target.value)} placeholder="اسم الخدمة المراد فتحها" className="flex-1 rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500" />
          <button type="submit" disabled={saving} className="rounded-lg bg-brand-500 px-4 py-2 text-sm font-bold text-white hover:bg-brand-600 disabled:opacity-50">
            {saving ? 'جاري...' : '+ إضافة'}
          </button>
        </form>
      )}

      {loading && <p className="text-slate-400">جاري التحميل...</p>}
      {!loading && items.length === 0 && <p className="text-slate-400">لا توجد خدمات مقترحة بعد.</p>}

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {items.map((s) => {
          const isAssignedToMe = !!employeeId && s.assignedEmployees.some((e) => e.id === employeeId)
          return (
            <div key={s.id} className={`rounded-xl border p-4 shadow-sm ${s.archived ? 'border-slate-200 bg-slate-50' : 'border-white bg-white'}`}>
              <div className="flex items-start justify-between">
                <p className="font-bold text-brand-900">{s.name} {s.archived && <span className="text-xs font-normal text-slate-400">(مؤرشفة)</span>}</p>
                {isAdmin && !s.archived && (
                  <button onClick={() => handleArchive(s.id)} disabled={busyId === s.id} className="rounded-lg bg-slate-100 px-2 py-1 text-xs font-bold text-slate-600 hover:bg-slate-200">🗄 أرشفة</button>
                )}
              </div>
              <div className="mt-2 flex items-center gap-2 text-xs">
                <b className="text-slate-600">الموكَّلون:</b>
                <span className="text-slate-500">{s.assignedEmployees.length > 0 ? s.assignedEmployees.map((e) => e.name).join('، ') : 'لا يوجد بعد'}</span>
                {isAdmin && (
                  <button onClick={() => handleAssign(s.id, s.assignedEmployees.map((e) => e.name))} disabled={busyId === s.id} className="rounded bg-brand-50 px-2 py-0.5 font-bold text-brand-700 hover:bg-brand-100">تعديل</button>
                )}
              </div>

              {s.reports.length > 0 && (
                <div className="mt-3 flex flex-col gap-2">
                  {s.reports.map((r) => (
                    <div key={r.id} className="rounded-lg bg-slate-50 p-2 text-xs">
                      <b className="text-slate-700">{r.employee?.name}</b> — <span className="text-slate-400">{new Date(r.createdAt).toLocaleDateString('ar-IQ')}</span>
                      <p className="mt-1 whitespace-pre-wrap text-slate-600">{r.content}</p>
                    </div>
                  ))}
                </div>
              )}

              {isAssignedToMe && !s.archived && (
                <div className="mt-3">
                  <textarea
                    value={reportDraft[s.id] ?? ''}
                    onChange={(e) => setReportDraft((d) => ({ ...d, [s.id]: e.target.value }))}
                    placeholder="اكتب تقرير/دراسة عن هذه الخدمة..."
                    rows={2}
                    className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                  />
                  <button onClick={() => handleAddReport(s.id)} disabled={busyId === s.id} className="mt-1 rounded-lg bg-brand-500 px-3 py-1.5 text-xs font-bold text-white hover:bg-brand-600 disabled:opacity-50">
                    رفع التقرير
                  </button>
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}

export default function TechUnitPage() {
  const { permissions, employee } = useSession()
  const canAdd = employee?.role === 'ADMIN' || permissions.includes('content_technician')
  const isAdmin = employee?.role === 'ADMIN'
  const [tab, setTab] = useState<'exhibitions' | 'products' | 'services' | 'training' | 'showcase'>('exhibitions')

  const tabs: { key: typeof tab; label: string }[] = [
    { key: 'exhibitions', label: '🏢 إدارة المعارض' },
    { key: 'products', label: '📦 إدارة المنتجات' },
    { key: 'services', label: '🛠️ إدارة الخدمات' },
    { key: 'training', label: '🎓 مفردات التدريب' },
    { key: 'showcase', label: '🖼️ معرض الأعمال' },
  ]

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">وحدة التقنيين</h2>
      <p className="mt-1 text-slate-500">إدارة المعارض، المنتجات، والخدمات — بالإضافة لمواد التدريب ومعرض الأعمال.</p>

      <div className="mt-4 flex gap-2 overflow-x-auto">
        {tabs.map((t) => (
          <button
            key={t.key}
            onClick={() => setTab(t.key)}
            className={`whitespace-nowrap rounded-lg px-4 py-2 text-sm font-bold ${tab === t.key ? 'bg-brand-500 text-white' : 'bg-white text-brand-700 shadow-sm hover:bg-brand-50'}`}
          >
            {t.label}
          </button>
        ))}
      </div>

      <div className="mt-4">
        {tab === 'exhibitions' && <ExhibitionsTab canAdd={canAdd} isAdmin={!!isAdmin} />}
        {tab === 'products' && <ProductsTab canAdd={canAdd} isAdmin={!!isAdmin} />}
        {tab === 'services' && <ServicesTab canAdd={canAdd} isAdmin={!!isAdmin} employeeId={employee?.id} />}
        {tab === 'training' && <TrainingManagement />}
        {tab === 'showcase' && <TechShowcasePage />}
      </div>
    </div>
  )
}
