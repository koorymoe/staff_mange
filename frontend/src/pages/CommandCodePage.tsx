import { useCallback, useEffect, useState } from 'react'
import { api, type Employee } from '../api'
import { useSession } from '../session'

// ═══ رمز مركز القيادة ═══
//
// «اليوزر نفسه، من ادخل باسورد ثاني مختلف يفتحلي النظام الأكبر».
//
// نفس اسم المستخدم يفتح عالمين: الرمز العادي يفتح نظام الشركة، والرمز
// الثاني يفتح مركز القيادة. ماكو حساب ثاني ولا زر «بدّل النظام» — الي
// يشوف شاشة الدخول ما يعرف أصلاً إن الطبقة الثانية موجودة.
//
// ⚠️ ليش الرمزين لازم يختلفون؟
// لو انطابقوا، الطبقة الثانية تنفتح بالخطأ بأول تسجيل دخول عادي — وتضيع
// الفكرة كلها. السيرفر يرفض التطابق، والواجهة تفحص محلياً حتى ما يستنى
// المستخدم رحلة كاملة حتى يشوف الرفض.
//
// ⚠️ الشاشة للمالك وحده، والسيرفر يرجّع **404** لغيره — مو 403. وجود
// الطبقة العليا نفسه سر: الـ403 يعترف إنها موجودة.
export default function CommandCodePage() {
  const { employee } = useSession()
  const [code, setCode] = useState('')
  const [confirmCode, setConfirmCode] = useState('')
  const [busy, setBusy] = useState(false)
  const [msg, setMsg] = useState<string | null>(null)
  const [err, setErr] = useState<string | null>(null)
  // قرار (ع) 10-09: يختار أي حساب موجود وينطيه الرمز الثاني
  const [employees, setEmployees] = useState<Employee[]>([])
  const [target, setTarget] = useState('')
  const [accounts, setAccounts] = useState<{ id: string; name: string; username: string | null; setAt: string | null }[]>([])
  const loadAccounts = useCallback(() => { api.getCommandAccounts().then(setAccounts).catch(() => {}) }, [])
  useEffect(() => {
    if (employee?.actualRole !== 'OWNER') return
    api.getEmployees().then((r) => setEmployees(r.filter((e) => e.status === 'ACTIVE' && e.username))).catch(() => {})
    loadAccounts()
  }, [employee?.actualRole, loadAccounts])

  const isOwner = employee?.actualRole === 'OWNER'
  if (!isOwner) {
    return (
      <div dir="rtl" className="p-8 text-center text-slate-400">
        الصفحة غير موجودة.
      </div>
    )
  }

  const save = async () => {
    setErr(null)
    setMsg(null)
    if (code.length < 8) {
      setErr('الرمز لازم يكون ٨ خانات على الأقل')
      return
    }
    if (code !== confirmCode) {
      setErr('الرمزين مو نفسهم')
      return
    }
    setBusy(true)
    try {
      const who = target || employee?.id || ''
      await api.setCommandPassword(code, who)
      const name = employees.find((e) => e.id === who)?.username || employee?.username
      setMsg(`انحفظ. اليوزر «${name}» بهذا الرمز يفتح مركز القيادة، وبرمزه العادي نظام الشركة.`)
      loadAccounts()
      setCode('')
      setConfirmCode('')
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر الحفظ')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div dir="rtl" className="mx-auto max-w-xl p-1">
      <h2 className="text-xl font-extrabold text-[#0f2040]">🔐 رمز مركز القيادة</h2>
      <p className="mt-1 text-sm text-slate-500">
        نفس اسم المستخدم مالك. الرمز العادي يفتح نظام الشركة، وهذا الرمز يفتح مركز القيادة.
      </p>

      <div className="mt-5 rounded-2xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        <label className="mb-1 block text-xs font-bold text-slate-500">اسم المستخدم</label>
        <select
          value={target || employee?.id || ''}
          onChange={(e) => setTarget(e.target.value)}
          className="mb-4 w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm"
        >
          {employee && <option value={employee.id}>{employee.username} — حسابي</option>}
          {employees.filter((e) => e.id !== employee?.id).map((e) => (
            <option key={e.id} value={e.id}>{e.username} — {e.name}</option>
          ))}
        </select>

        <label className="mb-1 block text-xs font-bold text-slate-500">الرمز الثاني</label>
        <input
          type="password"
          value={code}
          onChange={(e) => setCode(e.target.value)}
          autoComplete="new-password"
          placeholder="٨ خانات على الأقل، ولازم يختلف عن رمزك العادي"
          className="mb-3 w-full rounded-xl border border-slate-200 bg-slate-50/50 px-3 py-2.5 text-sm outline-none focus:border-[#2c5aad] focus:bg-white"
        />

        <label className="mb-1 block text-xs font-bold text-slate-500">تأكيد الرمز</label>
        <input
          type="password"
          value={confirmCode}
          onChange={(e) => setConfirmCode(e.target.value)}
          autoComplete="new-password"
          className="mb-4 w-full rounded-xl border border-slate-200 bg-slate-50/50 px-3 py-2.5 text-sm outline-none focus:border-[#2c5aad] focus:bg-white"
        />

        {err && <p className="mb-3 rounded-lg bg-red-50 px-3 py-2 text-sm font-bold text-red-700">{err}</p>}
        {msg && <p className="mb-3 rounded-lg bg-emerald-50 px-3 py-2 text-sm font-bold text-emerald-700">{msg}</p>}

        <button
          onClick={save}
          disabled={busy}
          className="w-full rounded-xl bg-gradient-to-l from-[#8b0f3a] to-[#e0245e] px-4 py-2.5 text-sm font-bold text-white shadow-lg transition-all hover:shadow-xl disabled:opacity-50"
        >
          {busy ? 'جاري الحفظ...' : 'احفظ الرمز'}
        </button>

        {accounts.length > 0 && (
          <div className="mt-4 border-t border-slate-100 pt-3">
            <p className="mb-2 text-xs font-bold text-slate-600">الحسابات الي تفتح مركز القيادة</p>
            {accounts.map((a) => (
              <div key={a.id} className="flex items-center justify-between gap-2 border-b border-slate-50 py-1.5 text-sm">
                <span><b>{a.username}</b> · {a.name}{a.setAt && <span className="ms-2 text-[11px] text-slate-400">{new Date(a.setAt).toLocaleDateString('ar-IQ')}</span>}</span>
                <button
                  onClick={() => {
                    if (!window.confirm(`تسكّر مركز القيادة عن «${a.username}»؟`)) return
                    api.clearCommandPassword(a.id).then(loadAccounts).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر'))
                  }}
                  className="text-xs font-bold text-red-600 hover:underline"
                >شيل</button>
              </div>
            ))}
          </div>
        )}

        <p className="mt-4 border-t border-slate-100 pt-3 text-xs leading-6 text-slate-500">
          • ماكو زر «بدّل النظام» — الفرق بالرمز بس، والي يشوف شاشة الدخول ما يعرف إن الطبقة الثانية موجودة.
          <br />
          • توكن مركز القيادة ما يشتغل على مسارات نظام الشركة، والعكس صحيح.
          <br />
          • <b>ما أكدر أختار الرمز عنك ولا أخزنه.</b> لو ضيّعته، غيّره من هنا بحسابك.
        </p>
      </div>
    </div>
  )
}
