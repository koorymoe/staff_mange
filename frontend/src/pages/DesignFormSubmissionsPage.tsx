import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { api, type DesignForm, type DesignFormQuestion, type DesignFormSubmission } from '../api'

// ⚠️ نسخة **النص** تنقلب بالوضع الليلي، والأصل يبقى للأسطح:
// نفس اللون يخدم عنواناً غامقاً على أبيض، ورأس جدول كحلي عليه نص أبيض.
// قلب الاثنين سوا يكسر واحداً منهما — نفس فخّ --color-white.
const PRIMARY_TEXT = 'var(--design-ink)'

export default function DesignFormSubmissionsPage() {
  const { formId } = useParams<{ formId: string }>()
  const [forms, setForms] = useState<DesignForm[]>([])
  const [questions, setQuestions] = useState<DesignFormQuestion[]>([])
  const [submissions, setSubmissions] = useState<DesignFormSubmission[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!formId) return
    Promise.all([api.getDesignForms(), api.getDesignFormQuestions(formId), api.getDesignFormSubmissions(formId)])
      .then(([allForms, qs, subs]) => { setForms(allForms); setQuestions(qs); setSubmissions(subs) })
      .finally(() => setLoading(false))
  }, [formId])

  if (!formId) return null
  const currentForm = forms.find((f) => f.id === formId)

  return (
    <div dir="rtl">
      <Link to="/design-forms" className="text-sm font-bold" style={{ color: PRIMARY_TEXT }}>← رجوع لكل الفورمات</Link>
      <h2 className="mt-2 text-2xl font-bold" style={{ color: PRIMARY_TEXT }}>
        الأجوبة المستلمة{currentForm ? ` — ${currentForm.name}` : ''}
      </h2>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {!loading && submissions.length === 0 && (
        <div className="mt-6 rounded-xl border border-white bg-white p-8 text-center shadow-sm">
          <p className="text-slate-400">ما وصل أي جواب على هذي الفورمة لحد الآن.</p>
        </div>
      )}

      {!loading && submissions.length > 0 && (
        <div className="mt-6 flex flex-col gap-4">
          {submissions.map((sub, si) => (
            <div key={sub.id} className="rounded-xl border border-white bg-white p-5 shadow-sm">
              <div className="mb-3 flex items-center justify-between">
                <span className="font-bold" style={{ color: PRIMARY_TEXT }}>جواب #{si + 1}</span>
                <span className="text-xs text-slate-400">{new Date(sub.submittedAt).toLocaleString('ar-IQ')}</span>
              </div>
              <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
                {questions.map((q) => {
                  const raw = sub.answers?.[q.id]
                  const value = Array.isArray(raw) ? raw.join('، ') : (raw ?? '-')
                  return (
                    <div key={q.id} className="rounded-lg bg-slate-50 px-3 py-2">
                      <p className="text-xs font-bold text-slate-500">{q.label}</p>
                      <p className="mt-0.5 text-sm text-slate-800">{String(value)}</p>
                    </div>
                  )
                })}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
