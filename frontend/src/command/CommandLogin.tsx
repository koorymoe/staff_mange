import { useState } from 'react'
import { api } from '../api'
import { useSession } from '../session'

// ═══ واجهة دخول مركز القيادة — الي ورا الستارة ═══
//
// ⚠️ هاي الواجهة **ما تقرر شي**. الي يقرر لأي نظام تدخل هو الرمز الي
// تكتبه مو الشاشة الي تكتبه بيها:
//
//   يوزر ع + رمز ع  ←  من أي شاشة  ←  نظام الشركة
//   يوزر ع + رمز م  ←  من أي شاشة  ←  مركز القيادة
//
// لهذا تنادي نفس `api.login` بالضبط مثل الشاشة الزرقاء — ماكو مسار
// ثاني ولا علم «أني الشاشة الحمراء» ينرسل للسيرفر. لو خلّينا الشاشة
// تفرض الطبقة، چان صار عندنا طريقين للدخول ولازم يتحرسون الاثنين —
// وأي واحد ينكسر يفتح النظام كله. طريق واحد أسهل تأميناً.
//
// السيرفر يرجّع realm، وLayout يقرر شنو يعرض. الشاشة ديكور بس.
export default function CommandLogin() {
  const { setEmployee } = useSession()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      setEmployee(await api.login(username, password))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form onSubmit={submit} className="cmdlogin-card" dir="rtl">
      <div className="cmdlogin-mark">◆</div>
      <h1 className="cmdlogin-title">مركز القيادة</h1>
      <p className="cmdlogin-sub">الأماني</p>

      <label className="cmdlogin-label">اسم المستخدم</label>
      <input
        required
        dir="ltr"
        autoCapitalize="off"
        autoCorrect="off"
        spellCheck={false}
        value={username}
        onChange={(e) => setUsername(e.target.value)}
        className="cmdlogin-input"
      />

      <label className="cmdlogin-label">الرمز</label>
      <input
        required
        type="password"
        dir="ltr"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        className="cmdlogin-input"
      />

      {error && <p className="cmdlogin-error">{error}</p>}

      <button type="submit" disabled={submitting} className="cmdlogin-btn">
        {submitting ? '...' : 'دخول'}
      </button>
    </form>
  )
}
