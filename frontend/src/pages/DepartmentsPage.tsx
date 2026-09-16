import { useEffect, useMemo, useState } from 'react'
import { api, type Department } from '../api'
import PageHeader from '../components/PageHeader'
import StatTile from '../components/StatTile'
import EmptyState from '../components/EmptyState'
import SearchBar from '../components/SearchBar'
import SaveError from '../components/SaveError'
import { useSaveGuard } from '../useSaveGuard'
import { matches } from '../utils/search'

/**
 * ═══ الأقسام ومسؤوليها ═══
 *
 * «نكدر نضيف الاقسام يدوياً ونضيف اسماء المسؤولين يدوياً، مع قابلية
 * اضافة اكثر من شخص يكدر يطلب حجز لنفس القسم، بالاضافة الى ارقام
 * المسؤولين، وامكانية التعديل فقط لمالك ومدير النظام حالياً».
 *
 * ⚠️ **القسم ينعطّل ما ينحذف**: الحجوزات القديمة تشير له، وحذفه
 * يخليها بلا قسم — نفس مبدأ الأرشفة بكل النظام. المعطّل ما يطلع
 * بمنتقي الحجز، ويبقى ظاهر هنا حتى يُرجَّع.
 *
 * ⚠️ **حسابات دخول للمسؤولين مرحلة ثانية** — هنا سجل أسماء وأرقام.
 */
export default function DepartmentsPage() {
  // ⚠️ `null` = لسه ما وصلت، `[]` = وصلت وفاضية. بلا التفريق، الشاشة
  // تگول «ماكو أقسام» وهي لسه تحمّل.
  const [rows, setRows] = useState<Department[] | null>(null)
  const [query, setQuery] = useState('')
  const [newDept, setNewDept] = useState('')
  const [openId, setOpenId] = useState<string | null>(null)
  const [headName, setHeadName] = useState('')
  const [headPhone, setHeadPhone] = useState('')
  const guard = useSaveGuard()

  const load = () => {
    api.getDepartments(true).then(setRows).catch(() => setRows([]))
  }
  useEffect(load, [])

  const loading = rows === null
  const list = useMemo(() => rows ?? [], [rows])
  const filtered = useMemo(() => {
    if (!query.trim()) return list
    return list.filter((d) => matches([d.name, ...(d.heads ?? []).map((h) => h.name)], query))
  }, [list, query])

  // ⚠️ عدد المسؤولين المسجَّلين — بلا بيانة «—» لا «٠»: قسم بلا
  // مسؤولين شي، و«ما انسجّل أحد بعد» شي ثاني.
  const activeDepts = list.filter((d) => d.active).length
  const totalHeads = list.reduce((n, d) => n + (d.heads ?? []).filter((h) => h.active).length, 0)

  const addDept = async () => {
    const name = newDept.trim()
    if (!name) return
    const ok = await guard.run('إضافة القسم', () => api.createDepartment(name))
    if (ok) { setNewDept(''); load() }
  }

  const toggleDept = async (d: Department) => {
    const ok = await guard.run(d.active ? 'تعطيل القسم' : 'إرجاع القسم',
      () => api.updateDepartment(d.id, { active: !d.active }))
    if (ok) load()
  }

  const addHead = async (departmentId: string) => {
    const name = headName.trim()
    if (!name) return
    const ok = await guard.run('إضافة المسؤول',
      () => api.createDepartmentHead({ departmentId, name, phone: headPhone.trim() || undefined }))
    if (ok) { setHeadName(''); setHeadPhone(''); load() }
  }

  const toggleHead = async (id: string, active: boolean) => {
    const ok = await guard.run(active ? 'تعطيل المسؤول' : 'إرجاع المسؤول',
      () => api.updateDepartmentHead(id, { active: !active }))
    if (ok) load()
  }

  return (
    <div dir="rtl" className="space-y-4">
      <SaveError message={guard.error} onClose={guard.clear} />
      <PageHeader
        title="🏢 الأقسام ومسؤوليها"
        subtitle="سجل أقسام الشركة والأشخاص الي يكدرون يطلبون حجزاً داخلياً لكل قسم — القسم ينتخب من هنا بشاشة الحجز، ما ينكتب بالإيد."
      />

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
        <StatTile label="أقسام فعّالة" icon="🏢" tone="info" value={activeDepts} hint={`من ${list.length} مسجّل`} />
        <StatTile label="مسؤولون مسجّلون" icon="👤" tone="default"
          value={totalHeads > 0 ? totalHeads : '—'}
          hint={totalHeads > 0 ? 'يكدرون يطلبون حجزاً' : 'ما انسجّل ولا مسؤول بعد'} />
        <StatTile label="بالبحث" icon="🔍" tone="default" value={filtered.length} />
      </div>

      <SearchBar value={query} onChange={setQuery} placeholder="ابحث باسم القسم أو المسؤول...">
        <input
          value={newDept}
          onChange={(e) => setNewDept(e.target.value)}
          placeholder="اسم قسم جديد"
          className="rounded-lg border px-3 py-1.5 text-xs"
          style={{ borderColor: 'var(--bd-line)', backgroundColor: 'var(--sf-card)', color: 'var(--t-body)' }}
        />
        <button onClick={addDept} disabled={guard.busy || !newDept.trim()}
          className="rounded-lg bg-brand-600 px-3 py-1.5 text-xs font-bold text-white disabled:opacity-50">
          ＋ أضف قسماً
        </button>
      </SearchBar>

      {loading && <p style={{ color: 'var(--t-faint)' }}>جاري التحميل...</p>}
      {!loading && filtered.length === 0 && (
        <EmptyState icon="🏢" title="ماكو أقسام"
          reason={list.length === 0 ? 'السجل فاضي — أضف قسماً من فوق.' : 'ماكو قسم يطابق البحث.'} />
      )}

      <div className="space-y-2">
        {filtered.map((d) => {
          const heads = d.heads ?? []
          const open = openId === d.id
          return (
            <div key={d.id} className="rounded-xl border p-3"
              style={{ backgroundColor: 'var(--sf-card)', borderColor: 'var(--bd-line)', opacity: d.active ? 1 : 0.6 }}>
              <div className="flex flex-wrap items-center justify-between gap-2">
                <button onClick={() => setOpenId(open ? null : d.id)} className="text-right">
                  <span className="font-bold" style={{ color: 'var(--t-title)' }}>{d.name}</span>
                  <span className="mr-2 text-xs" style={{ color: 'var(--t-faint)' }}>
                    {heads.length > 0 ? `${heads.length} مسؤول` : 'بلا مسؤولين'}
                  </span>
                  {!d.active && <span className="mr-2 rounded-full bg-slate-200 px-2 py-0.5 text-[10.5px] font-bold text-slate-600">معطّل</span>}
                </button>
                <button onClick={() => toggleDept(d)} disabled={guard.busy}
                  className={`rounded-lg px-3 py-1 text-xs font-bold ${d.active ? 'text-red-600 hover:bg-red-50' : 'text-emerald-700 hover:bg-emerald-50'}`}>
                  {d.active ? '⏸ عطّله' : '↩︎ رجّعه'}
                </button>
              </div>

              {open && (
                <div className="mt-3 space-y-2 border-t pt-3" style={{ borderColor: 'var(--bd-line)' }}>
                  {heads.length === 0 && (
                    <p className="text-xs" style={{ color: 'var(--t-faint)' }}>
                      ماكو مسؤول مسجّل — طالب الحجز راح ينكتب اسمه ورقمه بالإيد لحد ما تسجّل واحد هنا.
                    </p>
                  )}
                  {heads.map((h) => (
                    <div key={h.id} className="flex flex-wrap items-center justify-between gap-2 rounded-lg px-2 py-1.5"
                      style={{ backgroundColor: 'var(--sf-sunken)' }}>
                      <span className="text-sm" style={{ color: 'var(--t-body)' }}>
                        👤 {h.name}
                        {h.phone && <span className="mr-2 font-mono text-xs" dir="ltr">{h.phone}</span>}
                        {!h.active && <span className="mr-2 text-[10.5px] text-slate-500">(معطّل)</span>}
                      </span>
                      <button onClick={() => toggleHead(h.id, h.active)} disabled={guard.busy}
                        className={`rounded-lg px-2 py-0.5 text-[11px] font-bold ${h.active ? 'text-red-600 hover:bg-red-50' : 'text-emerald-700 hover:bg-emerald-50'}`}>
                        {h.active ? '✖ عطّله' : '↩︎ رجّعه'}
                      </button>
                    </div>
                  ))}
                  <div className="flex flex-wrap gap-2">
                    <input value={headName} onChange={(e) => setHeadName(e.target.value)}
                      placeholder="اسم المسؤول"
                      className="flex-1 rounded-lg border px-3 py-1.5 text-xs"
                      style={{ borderColor: 'var(--bd-line)', backgroundColor: 'var(--sf-card)', color: 'var(--t-body)' }} />
                    <input value={headPhone} onChange={(e) => setHeadPhone(e.target.value.replace(/\D/g, '').slice(0, 11))}
                      placeholder="رقمه (اختياري)" dir="ltr" inputMode="numeric"
                      className="w-40 rounded-lg border px-3 py-1.5 text-xs"
                      style={{ borderColor: 'var(--bd-line)', backgroundColor: 'var(--sf-card)', color: 'var(--t-body)' }} />
                    <button onClick={() => addHead(d.id)} disabled={guard.busy || !headName.trim()}
                      className="rounded-lg bg-brand-600 px-3 py-1.5 text-xs font-bold text-white disabled:opacity-50">
                      ＋ أضف مسؤولاً
                    </button>
                  </div>
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
