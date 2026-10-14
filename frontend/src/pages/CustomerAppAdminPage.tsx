import { useCallback, useEffect, useState } from 'react'
import { apiRequest, api, type Service } from '../api'

// ═══ «📱 تطبيق الزبائن» — قرار (ع) 10-14 ═══
// المالك يدير الي يشوفه الزبون بالتطبيق: العروض وصور الأعمال (يدوي، أو
// ماتركس يقترح مسودة والمالك يوافق قبل النشر)، ويشوف حسابات الزبائن
// ومحادثاتهم ويّا ماتركس.

interface Promo { id: string; kind: 'offer' | 'work' | 'service'; title: string; body: string | null; hasImage: boolean; serviceId: string | null; active: boolean; ord: number; source: string; createdAt: string }
interface CAccount { id: string; username: string; name: string; phone: string; email: string; verified: boolean; messages: number; bookings: number; lastChatAt: string | null; createdAt: string }
interface CMsg { id: string; role: 'user' | 'matrix'; content: string; bookingCode?: string; createdAt: string }

const KIND: Record<Promo['kind'], string> = { offer: '🎁 عرض', work: '🏗️ من أعمالنا', service: '🛠️ خدمة مميزة' }
const empty = { kind: 'offer' as Promo['kind'], title: '', body: '', serviceId: '', active: true, ord: 0 }

export default function CustomerAppAdminPage() {
  const [tab, setTab] = useState<'content' | 'accounts'>('content')
  const [promos, setPromos] = useState<Promo[] | null>(null)
  const [services, setServices] = useState<Service[]>([])
  const [edit, setEdit] = useState<(typeof empty & { id?: string }) | null>(null)
  const [file, setFile] = useState<File | null>(null)
  const [busy, setBusy] = useState(false)
  const [msg, setMsg] = useState('')
  const [accounts, setAccounts] = useState<CAccount[] | null>(null)
  const [chatOf, setChatOf] = useState<CAccount | null>(null)
  const [chat, setChat] = useState<CMsg[] | null>(null)

  const load = useCallback(() => {
    apiRequest<Promo[]>('/customer-app/promos').then(setPromos).catch((e) => { setMsg(e.message); setPromos([]) })
  }, [])
  useEffect(() => { load(); api.getServices().then(setServices).catch(() => {}) }, [load])
  useEffect(() => { if (tab === 'accounts' && !accounts) apiRequest<CAccount[]>('/customer-app/accounts').then(setAccounts).catch(() => setAccounts([])) }, [tab, accounts])
  useEffect(() => { if (chatOf) apiRequest<CMsg[]>(`/customer-app/accounts/${chatOf.id}/chat`).then(setChat).catch(() => setChat([])) }, [chatOf])

  const save = async () => {
    if (!edit) return
    setBusy(true); setMsg('')
    try {
      const body = JSON.stringify({ ...edit, serviceId: edit.serviceId || null, body: edit.body || null })
      let id = edit.id
      if (id) await apiRequest(`/customer-app/promos/${id}`, { method: 'PUT', body })
      else id = (await apiRequest<{ id: string }>('/customer-app/promos', { method: 'POST', body })).id
      if (file && id) {
        await api.uploadCustomerPromoImage(id, file)
      }
      setEdit(null); setFile(null); load()
    } catch (e) { setMsg(e instanceof Error ? e.message : 'خطأ') } finally { setBusy(false) }
  }
  const toggle = async (p: Promo) => {
    await apiRequest(`/customer-app/promos/${p.id}`, { method: 'PUT', body: JSON.stringify({ ...p, active: !p.active }) }).catch((e) => setMsg(e.message))
    load()
  }
  const remove = async (p: Promo) => {
    if (!confirm(`تحذف «${p.title}» من التطبيق؟`)) return
    await apiRequest(`/customer-app/promos/${p.id}`, { method: 'DELETE' }).catch((e) => setMsg(e.message))
    load()
  }
  const suggest = async () => {
    setBusy(true); setMsg('')
    try { await apiRequest('/customer-app/promos/suggest', { method: 'POST' }); load(); setMsg('ماتركس كتب مسودة عرض — راجعها وانشرها إذا تعجبك ✓') }
    catch (e) { setMsg(e instanceof Error ? e.message : 'خطأ') } finally { setBusy(false) }
  }

  const live = (promos ?? []).filter((p) => p.active).length
  const drafts = (promos ?? []).filter((p) => !p.active).length
  const inp = 'w-full rounded-xl border border-slate-300 bg-white px-3 py-2 text-sm'

  return (
    <div dir="rtl" className="w-full space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-sky-100 bg-gradient-to-l from-sky-50 to-white p-5 shadow-sm">
        <div>
          <h1 className="text-2xl font-black text-[#0f2040]">📱 تطبيق الزبائن</h1>
          <p className="text-sm text-slate-500">الي يشوفه الزبون بالتطبيق: العروض وصور أعمالنا — وماتركس هو الي يرد عليه ويسجّل حجوزاته.</p>
        </div>
        <div className="flex gap-2 text-center text-xs">
          <span className="rounded-xl bg-emerald-50 px-3 py-2 text-emerald-700"><b className="block text-lg">{live}</b>منشور</span>
          <span className="rounded-xl bg-amber-50 px-3 py-2 text-amber-700"><b className="block text-lg">{drafts}</b>مسودة</span>
        </div>
      </div>

      <div className="inline-flex gap-1.5 rounded-2xl border border-slate-200 bg-white p-1.5 shadow-sm">
        {([['content', '🎁 العروض والأعمال'], ['accounts', '👥 الزبائن ومحادثاتهم']] as const).map(([k, l]) => (
          <button key={k} onClick={() => setTab(k)} className={`rounded-xl px-4 py-2 text-xs font-extrabold ${tab === k ? 'bg-[#0f2040] text-white' : 'text-slate-600'}`}>{l}</button>
        ))}
      </div>
      {msg && <p className="rounded-xl bg-sky-50 px-3 py-2 text-sm text-sky-800">{msg}</p>}

      {tab === 'content' && (
        <>
          <div className="flex flex-wrap gap-2">
            <button onClick={() => { setEdit({ ...empty }); setFile(null) }} className="rounded-xl bg-emerald-700 px-4 py-2 text-sm font-bold text-white">+ عرض / عمل جديد</button>
            <button disabled={busy} onClick={suggest} className="rounded-xl border border-violet-300 bg-violet-50 px-4 py-2 text-sm font-bold text-violet-700 disabled:opacity-50">👁️ ماتركس يقترح عرض</button>
          </div>
          <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
            {promos === null && <div className="h-40 animate-pulse rounded-2xl bg-slate-100" />}
            {promos?.length === 0 && <p className="rounded-2xl bg-white p-6 text-center text-slate-400 shadow-sm sm:col-span-2 xl:col-span-3">ماكو محتوى بعد — ضيف أول عرض، أو خلي ماتركس يقترح.</p>}
            {promos?.map((p) => (
              <article key={p.id} className={`overflow-hidden rounded-2xl border bg-white shadow-sm ${p.active ? 'border-emerald-200' : 'border-amber-200'}`}>
                {p.hasImage ? <img src={api.customerPromoImageUrl(p.id)} alt="" className="h-36 w-full object-cover" /> : <div className="grid h-24 place-items-center bg-slate-50 text-3xl">{p.kind === 'work' ? '🏗️' : '🎁'}</div>}
                <div className="p-3">
                  <div className="flex flex-wrap items-center gap-1.5 text-[11px]">
                    <span className="rounded-md bg-slate-100 px-2 py-0.5">{KIND[p.kind]}</span>
                    <span className={`rounded-md px-2 py-0.5 font-bold ${p.active ? 'bg-emerald-50 text-emerald-700' : 'bg-amber-50 text-amber-700'}`}>{p.active ? 'منشور' : 'مسودة'}</span>
                    {p.source === 'matrix' && <span className="rounded-md bg-violet-50 px-2 py-0.5 text-violet-700">👁️ اقتراح ماتركس</span>}
                  </div>
                  <p className="mt-1.5 font-extrabold text-[#0f2040]">{p.title}</p>
                  {p.body && <p className="mt-0.5 line-clamp-3 text-xs text-slate-500">{p.body}</p>}
                  <div className="mt-2 flex gap-1.5">
                    <button onClick={() => toggle(p)} className={`rounded-lg px-3 py-1 text-xs font-bold ${p.active ? 'border border-slate-300 text-slate-600' : 'bg-emerald-700 text-white'}`}>{p.active ? 'إخفاء' : 'نشر'}</button>
                    <button onClick={() => { setEdit({ id: p.id, kind: p.kind, title: p.title, body: p.body ?? '', serviceId: p.serviceId ?? '', active: p.active, ord: p.ord }); setFile(null) }} className="rounded-lg border border-slate-300 px-3 py-1 text-xs text-[#2c5aad]">تعديل</button>
                    <button onClick={() => remove(p)} className="rounded-lg border border-rose-200 px-3 py-1 text-xs text-rose-600">حذف</button>
                  </div>
                </div>
              </article>
            ))}
          </div>
        </>
      )}

      {tab === 'accounts' && (
        <div className="overflow-x-auto rounded-2xl border border-slate-200 bg-white shadow-sm">
          <table className="w-full min-w-[760px] text-sm">
            <thead className="bg-slate-50 text-xs text-slate-500"><tr>{['الزبون', 'اسم المستخدم', 'الهاتف', 'الإيميل', 'رسائل', 'حجوزات من التطبيق', 'سجّل', ''].map((h, i) => <th key={i} className="p-3 text-right">{h}</th>)}</tr></thead>
            <tbody>
              {accounts === null && <tr><td colSpan={8} className="p-6 text-center text-slate-400">جاري التحميل…</td></tr>}
              {accounts?.length === 0 && <tr><td colSpan={8} className="p-6 text-center text-slate-400">ماكو زبون سجّل بالتطبيق بعد.</td></tr>}
              {accounts?.map((a) => (
                <tr key={a.id} className="border-t border-slate-100">
                  <td className="p-3 font-bold text-[#0f2040]">{a.name}{!a.verified && <span className="mr-1 text-[10px] text-amber-600">(ما أكّد إيميله)</span>}</td>
                  <td className="p-3 text-xs" dir="ltr">@{a.username}</td>
                  <td className="p-3 text-xs" dir="ltr">{a.phone}</td>
                  <td className="p-3 text-xs" dir="ltr">{a.email}</td>
                  <td className="p-3 tabular-nums">{a.messages}</td>
                  <td className="p-3 tabular-nums">{a.bookings}</td>
                  <td className="p-3 text-xs text-slate-500">{new Date(a.createdAt).toLocaleDateString('ar-IQ')}</td>
                  <td className="p-3">{a.messages > 0 && <button onClick={() => { setChat(null); setChatOf(a) }} className="rounded-lg border border-slate-300 px-2.5 py-1 text-xs font-bold text-[#2c5aad]">💬 المحادثة</button>}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {edit && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={() => setEdit(null)}>
          <div dir="rtl" className="w-full max-w-lg space-y-3 rounded-2xl bg-white p-5 shadow-2xl" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-lg font-extrabold text-[#0f2040]">{edit.id ? 'تعديل' : 'إضافة'} للتطبيق</h3>
            <label className="block text-xs text-slate-500">النوع
              <select className={`${inp} mt-1`} value={edit.kind} onChange={(e) => setEdit({ ...edit, kind: e.target.value as Promo['kind'] })}>
                <option value="offer">🎁 عرض (بالبانر المتحرك)</option><option value="work">🏗️ من أعمالنا (صورة عمل منجز)</option><option value="service">🛠️ خدمة مميزة</option>
              </select>
            </label>
            <label className="block text-xs text-slate-500">العنوان<input className={`${inp} mt-1`} value={edit.title} onChange={(e) => setEdit({ ...edit, title: e.target.value })} /></label>
            <label className="block text-xs text-slate-500">الوصف<textarea rows={3} className={`${inp} mt-1`} value={edit.body} onChange={(e) => setEdit({ ...edit, body: e.target.value })} /></label>
            <label className="block text-xs text-slate-500">الخدمة المرتبطة (اختياري)
              <select className={`${inp} mt-1`} value={edit.serviceId} onChange={(e) => setEdit({ ...edit, serviceId: e.target.value })}>
                <option value="">—</option>{services.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
              </select>
            </label>
            <label className="block text-xs text-slate-500">الصورة {edit.id && '(اتركها إذا ما تريد تغيّرها)'}<input type="file" accept="image/jpeg,image/png,image/webp" className="mt-1 block text-sm" onChange={(e) => setFile(e.target.files?.[0] ?? null)} /></label>
            <div className="flex items-center gap-4">
              <label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={edit.active} onChange={(e) => setEdit({ ...edit, active: e.target.checked })} /> منشور بالتطبيق</label>
              <label className="flex items-center gap-2 text-xs text-slate-500">الترتيب<input type="number" className="w-16 rounded-lg border border-slate-300 px-2 py-1" value={edit.ord} onChange={(e) => setEdit({ ...edit, ord: Number(e.target.value) || 0 })} /></label>
            </div>
            <div className="flex gap-2 pt-1">
              <button disabled={busy || !edit.title.trim()} onClick={save} className="flex-1 rounded-xl bg-[#0f2040] py-2.5 font-bold text-white disabled:opacity-50">{busy ? '…' : 'حفظ'}</button>
              <button onClick={() => setEdit(null)} className="rounded-xl border border-slate-300 px-4">إلغاء</button>
            </div>
          </div>
        </div>
      )}

      {chatOf && (
        <div className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/40 p-4" onClick={() => setChatOf(null)}>
          <div dir="rtl" className="w-full max-w-xl rounded-2xl bg-slate-50 p-4 shadow-2xl" onClick={(e) => e.stopPropagation()}>
            <div className="mb-3 flex items-center justify-between"><h3 className="font-extrabold text-[#0f2040]">💬 {chatOf.name} ويّا ماتركس</h3><button onClick={() => setChatOf(null)} className="text-2xl text-slate-400">×</button></div>
            <div className="max-h-[70vh] space-y-2 overflow-y-auto">
              {chat === null && <p className="text-center text-slate-400">…</p>}
              {chat?.map((m) => (
                <div key={m.id} className={`flex ${m.role === 'user' ? 'justify-start' : 'justify-end'}`}>
                  <div className={`max-w-[85%] rounded-2xl px-3 py-2 text-sm whitespace-pre-wrap ${m.role === 'user' ? 'bg-[#0f2040] text-white' : 'bg-white text-slate-800 shadow-sm'}`}>
                    {m.content}{m.bookingCode && <span className="mt-1 block text-xs font-bold text-emerald-600">✅ حجز {m.bookingCode}</span>}
                    <span className="mt-0.5 block text-[10px] opacity-60">{new Date(m.createdAt).toLocaleString('ar-IQ')}</span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
