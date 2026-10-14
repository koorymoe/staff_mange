import { useState } from 'react'
import { api, type Account } from '../api'

export default function AccountPage({ me, onChange, onLogout }: { me: Account; onChange: (a: Account) => void; onLogout: () => void }) {
  const [name, setName] = useState(me.name)
  const [phone, setPhone] = useState(me.phone)
  const [msg, setMsg] = useState('')
  const [busy, setBusy] = useState(false)

  const save = async () => {
    setBusy(true); setMsg('')
    try { onChange(await api.updateMe({ name, phone })); setMsg('انحفظ ✓') } catch (e) { setMsg(e instanceof Error ? e.message : 'خطأ') } finally { setBusy(false) }
  }
  const inp = 'w-full rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 outline-none focus:border-navy focus:bg-white'

  return (
    <div className="rise px-4 pt-6">
      <div className="flex items-center gap-3 rounded-3xl bg-gradient-to-br from-navy to-[#1d4ed8] p-5 text-white">
        <span className="grid h-14 w-14 place-items-center rounded-full bg-white/15 text-xl font-black">{me.name.trim().split(/\s+/).slice(0, 2).map((w) => w[0]).join(' ')}</span>
        <div className="min-w-0">
          <p className="truncate text-lg font-black">{me.name}</p>
          <p className="text-xs text-blue-100" dir="ltr">@{me.username}</p>
        </div>
      </div>

      <section className="mt-4 space-y-3 rounded-3xl bg-white p-5 shadow-sm">
        <label className="block"><span className="mb-1 block text-sm font-bold text-slate-600">الاسم</span><input value={name} onChange={(e) => setName(e.target.value)} className={inp} /></label>
        <label className="block"><span className="mb-1 block text-sm font-bold text-slate-600">رقم الهاتف</span><input dir="ltr" inputMode="tel" value={phone} onChange={(e) => setPhone(e.target.value)} className={inp} /></label>
        <p className="text-xs text-slate-400">الإيميل: <span dir="ltr">{me.email}</span></p>
        {msg && <p className="text-sm font-bold text-emerald-700">{msg}</p>}
        <button disabled={busy} onClick={save} className="w-full rounded-2xl bg-navy py-3 font-black text-white disabled:opacity-60">حفظ</button>
      </section>

      <section className="mt-4 rounded-3xl bg-white p-5 text-sm text-slate-600 shadow-sm">
        <p className="font-black text-ink">📲 نزّل التطبيق على شاشتك</p>
        <p className="mt-1">آيفون: زر المشاركة ← «إضافة إلى الشاشة الرئيسية». أندرويد: القائمة ⋮ ← «تثبيت التطبيق».</p>
      </section>

      <button onClick={onLogout} className="mt-4 w-full rounded-2xl border border-rose-200 bg-white py-3 font-black text-rose-600">تسجيل الخروج</button>
      <p className="mt-6 text-center text-xs text-slate-400">شركة الأماني © {new Date().getFullYear()}</p>
    </div>
  )
}
