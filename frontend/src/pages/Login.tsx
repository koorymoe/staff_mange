import { useState } from 'react'
import { api } from '../api'
import { useSession } from '../session'

export default function Login() {
  const { setEmployee } = useSession()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      const employee = await api.login(username, password)
      setEmployee(employee)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-b from-brand-500 via-brand-700 to-brand-900 p-4">
      <form
        onSubmit={handleSubmit}
        className="w-full max-w-sm rounded-xl border border-white bg-white p-8 shadow-[0_4px_20px_rgba(15,32,64,0.2)]"
      >
        <h1 className="text-center text-xl font-extrabold text-brand-900">شركة الأماني</h1>
        <p className="mt-1 text-center text-sm text-slate-500">نظام الإدارة المتكامل</p>

        <div className="mt-6">
          <label className="mb-1 block text-sm font-medium text-slate-600">اسم المستخدم</label>
          <input
            required
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>

        <div className="mt-4">
          <label className="mb-1 block text-sm font-medium text-slate-600">كلمة المرور</label>
          <input
            required
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>

        {error && <p className="mt-4 rounded-lg bg-red-50 p-3 text-sm text-red-600">{error}</p>}

        <button
          type="submit"
          disabled={submitting}
          className="mt-6 w-full rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-2 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30 disabled:opacity-50"
        >
          {submitting ? 'جاري الدخول...' : 'تسجيل الدخول'}
        </button>
      </form>
    </div>
  )
}
