import { useEffect, useRef, useState } from 'react'
import { api, type Exhibition, type Employee, type EmployeeRole } from '../api'
import { useSession } from '../session'

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

// ترتيب القائمة المنسدلة: التقنيين والمهندسين أول، بعدين المصممين،
// وبعدهم الباقي — لأن هذول هم المرشّحين الطبيعيين للمعارض.
const NOMINEE_GROUPS: { label: string; roles: EmployeeRole[] }[] = [
  { label: 'التقنيين والمهندسين', roles: ['TECHNICIAN', 'ENGINEER', 'QUALITY_ENGINEER'] },
  { label: 'المصممين', roles: ['DESIGNER'] },
]

function groupNominees(employees: Employee[]) {
  const active = employees.filter((e) => e.status === 'ACTIVE')
  const used = new Set<string>()
  const groups = NOMINEE_GROUPS.map((g) => {
    const members = active
      .filter((e) => g.roles.includes(e.role))
      .sort((a, b) => a.name.localeCompare(b.name, 'ar'))
    members.forEach((m) => used.add(m.id))
    return { label: g.label, members }
  })
  const rest = active.filter((e) => !used.has(e.id)).sort((a, b) => a.name.localeCompare(b.name, 'ar'))
  return [...groups, { label: 'باقي الموظفين', members: rest }].filter((g) => g.members.length > 0)
}

export default function ExhibitionsPage() {
  const { permissions, employee } = useSession()
  const canAdd = employee?.role === 'ADMIN' || permissions.includes('unit_technicians')
  const isAdmin = employee?.role === 'ADMIN'

  const [items, setItems] = useState<Exhibition[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({ title: '', location: '', startDate: '', endDate: '', companies: '', productsToShow: '' })
  const [saving, setSaving] = useState(false)
  const [busyId, setBusyId] = useState<string | null>(null)
  const [findingsDraft, setFindingsDraft] = useState<Record<string, string>>({})
  const [nominateFor, setNominateFor] = useState<string | null>(null)
  const [nominatePicked, setNominatePicked] = useState<string[]>([])
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

  // الترشيح صار قائمة منسدلة مجمّعة بدل ما يكتب الأسماء بيده — الكتابة
  // كانت تخلي أي حرف زايد يضيّع الموظف من الترشيح.
  const openNominate = (ex: Exhibition) => {
    setNominateFor(ex.id)
    setNominatePicked(ex.nominatedEmployees.map((e) => e.id))
  }

  const saveNominate = async () => {
    if (!nominateFor) return
    setBusyId(nominateFor)
    try {
      await api.nominateExhibition(nominateFor, nominatePicked)
      setNominateFor(null)
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
      <h2 className="text-2xl font-bold text-brand-900">إدارة المعارض</h2>
      <p className="mt-1 text-slate-500">معارض تجارية تحضرها الشركة — التاريخ، الشركات، المنتجات، والترشيح.</p>

      {canAdd && (
        <div className="mt-4 mb-4">
          <button onClick={() => setShowForm((v) => !v)} className="rounded-lg bg-brand-500 px-4 py-2 text-sm font-bold text-white hover:bg-brand-600">
            {showForm ? '× إغلاق' : '+ إضافة معرض جديد'}
          </button>
          {showForm && (
            <form onSubmit={handleCreate} className="mt-3 grid grid-cols-1 gap-3 rounded-xl border border-white bg-white p-5 shadow-sm sm:grid-cols-2">
              <input required value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} placeholder="عنوان المعرض" className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500" />
              <input required value={form.location} onChange={(e) => setForm({ ...form, location: e.target.value })} placeholder="مكان المعرض" className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500" />
              {/* التاريخين بدون تسمية ما ينعرف أيهم البداية وأيهم النهاية */}
              <label className="flex flex-col gap-1 text-sm text-slate-600">
                تاريخ بداية المعرض
                <input required type="date" value={form.startDate} onChange={(e) => setForm({ ...form, startDate: e.target.value })} className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500" />
              </label>
              <label className="flex flex-col gap-1 text-sm text-slate-600">
                تاريخ نهاية المعرض
                <input required type="date" value={form.endDate} onChange={(e) => setForm({ ...form, endDate: e.target.value })} className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500" />
              </label>
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
                <button onClick={() => openNominate(ex)} disabled={busyId === ex.id} className="rounded bg-brand-50 px-2 py-0.5 font-bold text-brand-700 hover:bg-brand-100">
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

      {/* قائمة ترشيح الموظفين — مجمّعة بالترتيب: تقنيين ومهندسين، مصممين، الباقي */}
      {nominateFor && (
        <div className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/40 p-4">
          <div className="mt-10 w-full max-w-lg rounded-xl bg-white p-5 shadow-xl">
            <div className="flex items-start justify-between">
              <h3 className="text-lg font-bold text-brand-900">ترشيح موظفين للمعرض</h3>
              <button onClick={() => setNominateFor(null)} className="text-2xl leading-none text-slate-400">×</button>
            </div>
            <p className="mt-1 text-sm text-slate-500">اختر من القائمة — تكدر تختار أكثر من واحد.</p>

            <div className="mt-3 max-h-80 overflow-y-auto rounded-lg border border-slate-200">
              {groupNominees(employees).map((g) => (
                <div key={g.label}>
                  <div className="sticky top-0 bg-slate-100 px-3 py-1.5 text-xs font-bold text-slate-600">{g.label}</div>
                  {g.members.map((m) => (
                    <label key={m.id} className="flex cursor-pointer items-center gap-2 px-3 py-2 text-sm hover:bg-slate-50">
                      <input
                        type="checkbox"
                        checked={nominatePicked.includes(m.id)}
                        onChange={(e) => setNominatePicked((prev) =>
                          e.target.checked ? [...prev, m.id] : prev.filter((id) => id !== m.id))}
                      />
                      <span className="font-medium text-slate-700">{m.name}</span>
                      {m.jobTitle && <span className="text-xs text-slate-400">{m.jobTitle}</span>}
                    </label>
                  ))}
                </div>
              ))}
              {employees.length === 0 && <p className="p-4 text-center text-sm text-slate-400">ماكو موظفين</p>}
            </div>

            <div className="mt-4 flex gap-2">
              <button onClick={saveNominate} disabled={busyId === nominateFor} className="flex-1 rounded-lg bg-brand-500 px-4 py-2 text-sm font-bold text-white hover:bg-brand-600 disabled:opacity-50">
                {busyId === nominateFor ? 'جاري الحفظ...' : `حفظ الترشيح (${nominatePicked.length})`}
              </button>
              <button onClick={() => setNominateFor(null)} className="rounded-lg border border-slate-300 px-4 py-2 text-sm">إلغاء</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
