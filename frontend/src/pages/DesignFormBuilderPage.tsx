import { useEffect, useState } from 'react'
import { api, type DesignFormQuestion, type DesignFormQuestionType } from '../api'

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
  const [questions, setQuestions] = useState<DesignFormQuestion[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [form, setForm] = useState(emptyForm())
  const [saving, setSaving] = useState(false)
  const [busyId, setBusyId] = useState<string | null>(null)

  const load = () => { api.getDesignFormQuestions().then(setQuestions).finally(() => setLoading(false)) }
  useEffect(load, [])

  const startCreate = () => { setEditingId(null); setForm(emptyForm()); setShowForm(true) }
  const startEdit = (q: DesignFormQuestion) => {
    setEditingId(q.id)
    setForm({ label: q.label, type: q.type, options: q.options.join(', '), required: q.required })
    setShowForm(true)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!form.label.trim()) return
    setSaving(true)
    const options = NEEDS_OPTIONS.includes(form.type)
      ? form.options.split(',').map((s) => s.trim()).filter(Boolean)
      : []
    try {
      if (editingId) {
        await api.updateDesignFormQuestion(editingId, { label: form.label.trim(), type: form.type, options, required: form.required })
      } else {
        await api.createDesignFormQuestion({ label: form.label.trim(), type: form.type, options, required: form.required })
      }
      setShowForm(false)
      setForm(emptyForm())
      setEditingId(null)
      load()
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
      load()
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
      load()
    }
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">وحدة التصميم — بناء استمارة طلب التصميم</h2>
      <p className="mt-1 text-slate-500">
        أضف الأسئلة الي تريدها بنفسك (نص، رقم، تاريخ، اختيارات...) ورتّبها — النظام يبني لك الاستمارة تلقائياً من الأسئلة الي تضيفها.
      </p>

      <div className="mt-4">
        <button onClick={startCreate} className="rounded-lg bg-brand-500 px-4 py-2 text-sm font-bold text-white hover:bg-brand-600">
          + إضافة سؤال جديد
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
            <button type="submit" disabled={saving} className="rounded-lg bg-brand-500 px-4 py-2 text-sm font-bold text-white hover:bg-brand-600 disabled:opacity-50">
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
          <p className="text-slate-400">ما أضفت أي أسئلة بعد — اضغط "+ إضافة سؤال جديد" وابدأ ببناء الاستمارة.</p>
        </div>
      )}

      <div className="mt-6 flex flex-col gap-3">
        {questions.map((q, i) => (
          <div key={q.id} className="flex items-start justify-between gap-3 rounded-xl border border-white bg-white p-4 shadow-sm">
            <div className="flex items-start gap-3">
              <span className="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-brand-50 text-xs font-bold text-brand-700">{i + 1}</span>
              <div>
                <p className="font-bold text-brand-900">
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
