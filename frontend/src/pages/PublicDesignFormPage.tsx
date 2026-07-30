import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { api, type DesignFormQuestion } from '../api'

const PRIMARY = '#47528f'
const GOLD = '#c97a3a'
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
    background: '#f4f5fa',
    display: 'flex',
    justifyContent: 'center',
    padding: '32px 16px',
  }

  if (loading) {
    return <div style={wrapStyle}><p style={{ color: PRIMARY, fontWeight: 700 }}>جاري التحميل...</p></div>
  }

  if (error && !done) {
    return (
      <div style={wrapStyle}>
        <div style={{ background: '#fff', borderRadius: '16px', padding: '32px', maxWidth: '480px', textAlign: 'center', boxShadow: '0 4px 24px rgba(71,82,143,0.12)' }}>
          <p style={{ color: '#c0392b', fontWeight: 700, fontSize: '16px' }}>{error}</p>
        </div>
      </div>
    )
  }

  return (
    <div style={wrapStyle}>
      <div style={{ width: '100%', maxWidth: '720px' }}>
        <div style={{
          background: '#fff', borderRadius: '18px', overflow: 'hidden',
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
            <div style={{ fontSize: '12px', fontWeight: 600, color: GOLD, marginTop: '4px' }}>
              Al-Amani for General Trading, Real Estate &amp; Commercial Agencies LLC
            </div>
            <div style={{ marginTop: '14px', fontSize: '20px', fontWeight: 700 }}>{name}</div>
            <div style={{ fontSize: '13px', color: '#d7dbf0', marginTop: '2px' }}>وحدة التصميم — يرجى تعبئة كل الحقول المطلوبة</div>
          </div>

          {done ? (
            <div style={{ padding: '48px 30px', textAlign: 'center' }}>
              <div style={{ fontSize: '40px' }}>✓</div>
              <p style={{ color: PRIMARY, fontWeight: 700, fontSize: '18px', marginTop: '10px' }}>تم إرسال الفورمة بنجاح</p>
              <p style={{ color: '#7c85ad', fontSize: '14px', marginTop: '6px' }}>شكراً لك، فريق التصميم بشركة الأماني راح يتواصل معك قريباً.</p>
            </div>
          ) : (
            <form onSubmit={handleSubmit} style={{ padding: '28px 30px', display: 'flex', flexDirection: 'column', gap: '18px' }}>
              {questions.length === 0 && (
                <p style={{ color: '#9aa0c2', textAlign: 'center' }}>ما اكو أسئلة بهذي الفورمة بعد.</p>
              )}
              {questions.map((q) => (
                <div key={q.id}>
                  <label style={{ display: 'block', fontSize: '14px', fontWeight: 700, color: PRIMARY, marginBottom: '6px' }}>
                    {q.label} {q.required && <span style={{ color: '#c0392b' }}>*</span>}
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
                    <div style={{ display: 'flex', flexWrap: 'wrap', gap: '10px' }}>
                      {q.options.map((o) => (
                        <label key={o} style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '13px', background: PEACH, borderRadius: '8px', padding: '6px 10px', color: PRIMARY }}>
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
                    <p style={{ fontSize: '13px', color: '#9aa0c2' }}>إرفاق الملفات غير متاح حالياً بهذا الرابط — يرجى ذكر التفاصيل نصياً أو التواصل المباشر.</p>
                  )}
                </div>
              ))}

              {error && <p style={{ color: '#c0392b', fontSize: '13px' }}>{error}</p>}

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
  color: PRIMARY,
  boxSizing: 'border-box',
}
