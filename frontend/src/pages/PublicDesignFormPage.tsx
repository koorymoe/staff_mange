import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { api, type DesignFormQuestion } from '../api'

const PRIMARY = '#47528f'
// ⚠️ نسخة **النص** تنقلب بالوضع الليلي، والأصل يبقى للأسطح:
// نفس اللون يخدم عنواناً غامقاً على أبيض، ورأس جدول كحلي عليه نص أبيض.
// قلب الاثنين سوا يكسر واحداً منهما — نفس فخّ --color-white.
const PRIMARY_TEXT = 'var(--design-ink)'
const GOLD = '#c97a3a'
// ⚠️ نسخة **النص** تنقلب بالوضع الليلي، والأصل يبقى للأسطح:
// نفس اللون يخدم عنواناً غامقاً على أبيض، ورأس جدول كحلي عليه نص أبيض.
// قلب الاثنين سوا يكسر واحداً منهما — نفس فخّ --color-white.
const GOLD_TEXT = 'var(--gold-warm-ink)'
const PEACH = '#fbede2'

export default function PublicDesignFormPage() {
  const { token } = useParams<{ token: string }>()
  const [name, setName] = useState('')
  const [questions, setQuestions] = useState<DesignFormQuestion[]>([])
  const [answers, setAnswers] = useState<Record<string, string | string[]>>({})
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [done, setDone] = useState(false)

  useEffect(() => {
    if (!token) return
    api.getPublicDesignForm(token)
      .then((res) => { setName(res.name); setQuestions(res.questions) })
      .catch((e) => setError(e instanceof Error ? e.message : 'تعذر فتح الفورمة'))
      .finally(() => setLoading(false))
  }, [token])

  const setAnswer = (id: string, value: string | string[]) => setAnswers((prev) => ({ ...prev, [id]: value }))

  const toggleCheckbox = (id: string, option: string) => {
    const current = (answers[id] as string[] | undefined) || []
    const next = current.includes(option) ? current.filter((o) => o !== option) : [...current, option]
    setAnswer(id, next)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!token) return
    setSubmitting(true)
    setError(null)
    try {
      await api.submitPublicDesignForm(token, answers)
      setDone(true)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'تعذر إرسال الفورمة')
    } finally {
      setSubmitting(false)
    }
  }

  const wrapStyle: React.CSSProperties = {
    direction: 'rtl',
    fontFamily: "'Cairo', 'Tajawal', sans-serif",
    minHeight: '100vh',
    background: 'var(--sf-sunken)',
    display: 'flex',
    justifyContent: 'center',
    padding: '32px 16px',
  }

  if (loading) {
    return <div style={wrapStyle}><p style={{ color: PRIMARY_TEXT, fontWeight: 700 }}>جاري التحميل...</p></div>
  }

  if (error && !done) {
    return (
      <div style={wrapStyle}>
        <div style={{ background: 'var(--sf-card)', borderRadius: '16px', padding: '32px', maxWidth: '480px', textAlign: 'center', boxShadow: '0 4px 24px rgba(71,82,143,0.12)' }}>
          <p className="text-t-danger font-bold text-[16px]">{error}</p>
        </div>
      </div>
    )
  }

  return (
    <div style={wrapStyle}>
      <div style={{ width: '100%', maxWidth: '720px' }}>
        <div style={{
          background: 'var(--sf-card)', borderRadius: '18px', overflow: 'hidden',
          boxShadow: '0 8px 32px rgba(71,82,143,0.15)', border: `1px solid ${PEACH}`,
        }}>
          <div style={{
            background: `linear-gradient(135deg, ${PRIMARY}, #2f3868)`,
            padding: '28px 30px', color: '#fff', position: 'relative', overflow: 'hidden',
          }}>
            <div style={{ position: 'absolute', top: 0, left: 0, bottom: 0, width: '10px', background: `linear-gradient(${GOLD}, ${PRIMARY})` }} />
            <div style={{ fontSize: '17px', fontWeight: 700, lineHeight: 1.5 }}>
              شركة الأماني للتجارة العامة والاستثمارات العقارية والوكالات التجارية محدودة المسؤولية
            </div>
            <div style={{ fontSize: '12px', fontWeight: 600, color: GOLD_TEXT, marginTop: '4px' }}>
              Al-Amani for General Trading, Real Estate &amp; Commercial Agencies LLC
            </div>
            <div className="mt-3.5 text-[20px] font-bold">{name}</div>
            <div className="text-[13px] text-[#d7dbf0] mt-0.5">وحدة التصميم — يرجى تعبئة كل الحقول المطلوبة</div>
          </div>

          {done ? (
            <div className="py-12 px-[30px] text-center">
              <div className="text-[40px]">✓</div>
              <p style={{ color: PRIMARY_TEXT, fontWeight: 700, fontSize: '18px', marginTop: '10px' }}>تم إرسال الفورمة بنجاح</p>
              <p className="text-[#7c85ad] text-[14px] mt-1.5">شكراً لك، فريق التصميم بشركة الأماني راح يتواصل معك قريباً.</p>
            </div>
          ) : (
            <form onSubmit={handleSubmit} style={{ padding: '28px 30px', display: 'flex', flexDirection: 'column', gap: '18px' }}>
              {questions.length === 0 && (
                <p className="text-[#9aa0c2] text-center">ما اكو أسئلة بهذي الفورمة بعد.</p>
              )}
              {questions.map((q) => (
                <div key={q.id}>
                  <label style={{ display: 'block', fontSize: '14px', fontWeight: 700, color: PRIMARY_TEXT, marginBottom: '6px' }}>
                    {q.label} {q.required && <span className="text-t-danger">*</span>}
                  </label>
                  {q.type === 'TEXTAREA' && (
                    <textarea
                      required={q.required} rows={3}
                      value={(answers[q.id] as string) || ''}
                      onChange={(e) => setAnswer(q.id, e.target.value)}
                      style={inputStyle}
                    />
                  )}
                  {(q.type === 'TEXT' || q.type === 'NUMBER' || q.type === 'DATE') && (
                    <input
                      required={q.required}
                      type={q.type === 'NUMBER' ? 'number' : q.type === 'DATE' ? 'date' : 'text'}
                      value={(answers[q.id] as string) || ''}
                      onChange={(e) => setAnswer(q.id, e.target.value)}
                      style={inputStyle}
                    />
                  )}
                  {q.type === 'SELECT' && (
                    <select
                      required={q.required}
                      value={(answers[q.id] as string) || ''}
                      onChange={(e) => setAnswer(q.id, e.target.value)}
                      style={inputStyle}
                    >
                      <option value="">-- اختر --</option>
                      {q.options.map((o) => <option key={o} value={o}>{o}</option>)}
                    </select>
                  )}
                  {q.type === 'CHECKBOX' && (
                    <div className="flex flex-wrap gap-2.5">
                      {q.options.map((o) => (
                        <label key={o} style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '13px', background: PEACH, borderRadius: '8px', padding: '6px 10px', color: PRIMARY_TEXT }}>
                          <input
                            type="checkbox"
                            checked={((answers[q.id] as string[]) || []).includes(o)}
                            onChange={() => toggleCheckbox(q.id, o)}
                          />
                          {o}
                        </label>
                      ))}
                    </div>
                  )}
                  {q.type === 'FILE' && (
                    <p className="text-[13px] text-[#9aa0c2]">إرفاق الملفات غير متاح حالياً بهذا الرابط — يرجى ذكر التفاصيل نصياً أو التواصل المباشر.</p>
                  )}
                </div>
              ))}

              {error && <p className="text-t-danger text-[13px]">{error}</p>}

              {questions.length > 0 && (
                <button
                  type="submit" disabled={submitting}
                  style={{
                    background: GOLD, color: '#fff', border: 'none', borderRadius: '10px',
                    padding: '12px', fontWeight: 700, fontSize: '15px', cursor: 'pointer', opacity: submitting ? 0.6 : 1,
                  }}
                >
                  {submitting ? 'جاري الإرسال...' : 'إرسال الفورمة'}
                </button>
              )}
            </form>
          )}
        </div>
      </div>
    </div>
  )
}

const inputStyle: React.CSSProperties = {
  width: '100%',
  border: `1.5px solid #d7dbf0`,
  borderRadius: '10px',
  padding: '10px 12px',
  fontSize: '14px',
  outline: 'none',
  fontFamily: "'Cairo', 'Tajawal', sans-serif",
  color: PRIMARY_TEXT,
  boxSizing: 'border-box',
}
