import { useEffect, useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { api, type DesignForm, type DesignFormQuestion, type DesignFormQuestionType } from '../api'

const PRIMARY = '#47528f'

const TYPE_LABELS: Record<DesignFormQuestionType, string> = {
  TEXT: 'نص قصير',
  TEXTAREA: 'نص طويل',
  NUMBER: 'رقم',
  DATE: 'تاريخ',
  SELECT: 'اختيار واحد من قائمة',
  CHECKBOX: 'اختيار متعدد (خيارات)',
  FILE: 'إرفاق ملف',
}

const NEEDS_OPTIONS: DesignFormQuestionType[] = ['SELECT', 'CHECKBOX']

function emptyForm() {
  return { label: '', type: 'TEXT' as DesignFormQuestionType, options: '', required: false }
}

export default function DesignFormBuilderPage() {
  const { formId } = useParams<{ formId: string }>()
  const navigate = useNavigate()
  const [forms, setForms] = useState<DesignForm[]>([])
  const [questions, setQuestions] = useState<DesignFormQuestion[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [form, setForm] = useState(emptyForm())
  const [saving, setSaving] = useState(false)
  const [busyId, setBusyId] = useState<string | null>(null)

  // ⚠️ الجلب بمكان واحد داخل الـeffect، والحفظ والحذف يطلبونه برفع
  // العدّاد. و`alive` يمنع سباق الطلبات: تبديل الاستمارة بسرعة چان
  // يخلّي أسئلة استمارة قديمة تطلع تحت اسم استمارة ثانية.
  const [reload, setReload] = useState(0)
  const refresh = () => setReload((n) => n + 1)

  useEffect(() => {
    if (!formId) return
    let alive = true
    void (async () => {
      try {
        const [allForms, qs] = await Promise.all([
          api.getDesignForms(),
          api.getDesignFormQuestions(formId),
        ])
        if (alive) { setForms(allForms); setQuestions(qs) }
      } finally {
        if (alive) setLoading(false)
      }
    })()
    return () => { alive = false }
  }, [formId, reload])

  const currentForm = forms.find((f) => f.id === formId)

  const startCreate = () => { setEditingId(null); setForm(emptyForm()); setShowForm(true) }
  const startEdit = (q: DesignFormQuestion) => {
    setEditingId(q.id)
    setForm({ label: q.label, type: q.type, options: q.options.join(', '), required: q.required })
    setShowForm(true)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!form.label.trim() || !formId) return
    setSaving(true)
    const options = NEEDS_OPTIONS.includes(form.type)
      ? form.options.split(',').map((s) => s.trim()).filter(Boolean)
      : []
    try {
      if (editingId) {
        await api.updateDesignFormQuestion(editingId, { label: form.label.trim(), type: form.type, options, required: form.required })
      } else {
        await api.createDesignFormQuestion(formId, { label: form.label.trim(), type: form.type, options, required: form.required })
      }
      setShowForm(false)
      setForm(emptyForm())
      setEditingId(null)
      refresh()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر حفظ السؤال')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('حذف هذا السؤال نهائياً؟')) return
    setBusyId(id)
    try {
      await api.deleteDesignFormQuestion(id)
      refresh()
    } finally {
      setBusyId(null)
    }
  }

  const move = async (index: number, direction: -1 | 1) => {
    const target = index + direction
    if (target < 0 || target >= questions.length) return
    const reordered = [...questions]
    ;[reordered[index], reordered[target]] = [reordered[target], reordered[index]]
    setQuestions(reordered)
    try {
      await api.reorderDesignFormQuestions(reordered.map((q) => q.id))
    } catch {
      refresh()
    }
  }

  if (!formId) return null

  return (
    <div dir="rtl">
      <Link to="/design-forms" className="text-sm font-bold" style={{ color: PRIMARY }}>← رجوع لكل الفورمات</Link>
      <h2 className="mt-2 text-2xl font-bold" style={{ color: PRIMARY }}>
        {currentForm ? `أسئلة فورمة: ${currentForm.name}` : 'جاري التحميل...'}
      </h2>
      <p className="mt-1 text-slate-500">
        أضف الأسئلة الي تريدها بنفسك (نص، رقم، تاريخ، اختيارات...) ورتّبها — هذي الأسئلة خاصة بهذي الفورمة بس، وما تنعرض بأي فورمة ثانية.
      </p>

      <div className="mt-4">
        <button onClick={startCreate} className="rounded-lg px-4 py-2 text-sm font-bold text-white" style={{ background: PRIMARY }}>
          + إضافة سؤال جديد
        </button>
        <button onClick={() => navigate(`/design-forms/${formId}/submissions`)} className="mr-2 rounded-lg bg-slate-100 px-4 py-2 text-sm font-bold text-slate-700 hover:bg-slate-200">
          عرض الأجوبة المستلمة
        </button>
      </div>

      {showForm && (
        <form onSubmit={handleSubmit} className="mt-4 grid grid-cols-1 gap-3 rounded-xl border border-white bg-white p-5 shadow-sm sm:grid-cols-2">
          <input
            required value={form.label} onChange={(e) => setForm({ ...form, label: e.target.value })}
            placeholder="نص السؤال (مثال: اسم الزبون)"
            className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 sm:col-span-2"
          />
          <select
            value={form.type} onChange={(e) => setForm({ ...form, type: e.target.value as DesignFormQuestionType })}
            className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
          >
            {(Object.keys(TYPE_LABELS) as DesignFormQuestionType[]).map((t) => (
              <option key={t} value={t}>{TYPE_LABELS[t]}</option>
            ))}
          </select>
          <label className="flex items-center gap-2 rounded-lg border border-slate-200 px-3 py-2 text-sm">
            <input type="checkbox" checked={form.required} onChange={(e) => setForm({ ...form, required: e.target.checked })} />
            سؤال إلزامي
          </label>
          {NEEDS_OPTIONS.includes(form.type) && (
            <input
              value={form.options} onChange={(e) => setForm({ ...form, options: e.target.value })}
              placeholder="الخيارات مفصولة بفاصلة (مثال: سكني، تجاري، تعليمي)"
              className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 sm:col-span-2"
            />
          )}
          <div className="flex gap-2 sm:col-span-2">
            <button type="submit" disabled={saving} className="rounded-lg px-4 py-2 text-sm font-bold text-white disabled:opacity-50" style={{ background: PRIMARY }}>
              {saving ? 'جاري الحفظ...' : editingId ? 'حفظ التعديل' : 'إضافة السؤال'}
            </button>
            <button type="button" onClick={() => { setShowForm(false); setEditingId(null) }} className="rounded-lg bg-slate-100 px-4 py-2 text-sm font-bold text-slate-600 hover:bg-slate-200">
              إلغاء
            </button>
          </div>
        </form>
      )}

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {!loading && questions.length === 0 && (
        <div className="mt-6 rounded-xl border border-white bg-white p-8 text-center shadow-sm">
          <p className="text-slate-400">ما أضفت أي أسئلة بهذي الفورمة بعد — اضغط "+ إضافة سؤال جديد" وابدأ ببناء الاستمارة.</p>
        </div>
      )}

      <div className="mt-6 flex flex-col gap-3">
        {questions.map((q, i) => (
          <div key={q.id} className="flex items-start justify-between gap-3 rounded-xl border border-white bg-white p-4 shadow-sm">
            <div className="flex items-start gap-3">
              <span className="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-brand-50 text-xs font-bold text-brand-700">{i + 1}</span>
              <div>
                <p className="font-bold" style={{ color: PRIMARY }}>
                  {q.label} {q.required && <span className="text-red-500">*</span>}
                </p>
                <p className="text-xs text-slate-500">{TYPE_LABELS[q.type]}{q.options.length > 0 && ` — ${q.options.join('، ')}`}</p>
              </div>
            </div>
            <div className="flex shrink-0 items-center gap-1">
              <button onClick={() => move(i, -1)} disabled={i === 0} className="rounded-lg bg-slate-50 px-2 py-1 text-xs font-bold text-slate-600 hover:bg-slate-100 disabled:opacity-30">▲</button>
              <button onClick={() => move(i, 1)} disabled={i === questions.length - 1} className="rounded-lg bg-slate-50 px-2 py-1 text-xs font-bold text-slate-600 hover:bg-slate-100 disabled:opacity-30">▼</button>
              <button onClick={() => startEdit(q)} className="rounded-lg bg-brand-50 px-2 py-1 text-xs font-bold text-brand-700 hover:bg-brand-100">تعديل</button>
              <button onClick={() => handleDelete(q.id)} disabled={busyId === q.id} className="rounded-lg bg-red-50 px-2 py-1 text-xs font-bold text-red-700 hover:bg-red-100 disabled:opacity-50">حذف</button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
