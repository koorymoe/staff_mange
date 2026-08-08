import { useEffect, useState } from 'react'
import { api, type EmployeeLetter } from '../api'
import { useSession } from '../session'
import vstripUrl from '../assets/print/quotation-vstrip.png'
import bannerUrl from '../assets/print/quotation-banner.png'

// ═══ الطلبات — كتاب رسمي من الموظف للإدارة ═══
//
// الموظف الي يريد شي (سلفة، نقل، شكوى، اقتراح) كان يفتح الوورد
// بالموبايل أو يكتب بورقة ويوصلها بيده. وبعدها: الورقة تضيع، أو
// المدير ينساها، أو الموظف ما يعرف إذا وصلت أصلاً.
//
// هنا الطلب ينكتب بصيغة الكتاب الرسمي، يوصل الإدارة بإشعار فوري،
// ينطبع بورقة الشركة (نفس إطار عرض السعر)، ويبقى بالسجل بجوابه.
export default function LettersPage() {
  const { employee } = useSession()
  const isAdmin = employee?.role === 'ADMIN'

  const [tab, setTab] = useState<'mine' | 'inbox'>('mine')
  const [mine, setMine] = useState<EmployeeLetter[]>([])
  const [inbox, setInbox] = useState<EmployeeLetter[]>([])
  const [addressees, setAddressees] = useState<string[]>([])
  const [loading, setLoading] = useState(true)

  // نموذج الطلب
  const [showForm, setShowForm] = useState(false)
  const [addressedTo, setAddressedTo] = useState('')
  const [subject, setSubject] = useState('')
  const [body, setBody] = useState('')
  const [busy, setBusy] = useState(false)
  const [msg, setMsg] = useState<{ ok: boolean; text: string } | null>(null)

  // جواب الإدارة
  const [noteDraft, setNoteDraft] = useState<Record<string, string>>({})

  const load = () => {
    const jobs: Promise<unknown>[] = [
      api.getMyLetters().then(setMine),
      api.getLetterAddressees().then((a) => { setAddressees(a); if (!addressedTo) setAddressedTo(a[0] || '') }),
    ]
    if (isAdmin) jobs.push(api.getLetters().then(setInbox))
    Promise.all(jobs).catch(() => {}).finally(() => setLoading(false))
  }
  useEffect(load, [isAdmin])

  const submit = async () => {
    if (subject.trim().length < 3) return setMsg({ ok: false, text: 'اكتب موضوع الطلب' })
    if (body.trim().length < 10) return setMsg({ ok: false, text: 'اكتب تفاصيل الطلب' })
    setBusy(true); setMsg(null)
    try {
      await api.createLetter({ addressedTo, subject: subject.trim(), body: body.trim() })
      setMsg({ ok: true, text: 'انرسل الطلب — الإدارة انبلغت' })
      setSubject(''); setBody(''); setShowForm(false)
      load()
    } catch (e) {
      setMsg({ ok: false, text: e instanceof Error ? e.message : 'تعذر إرسال الطلب' })
    } finally { setBusy(false) }
  }

  const decide = async (id: string, approve: boolean) => {
    const note = (noteDraft[id] || '').trim()
    if (!approve && note.length < 5) {
      alert('اكتب سبب الرفض — الموظف لازم يعرف ليش')
      return
    }
    try {
      await api.decideLetter(id, { approve, note })
      setNoteDraft({ ...noteDraft, [id]: '' })
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر تسجيل القرار')
    }
  }

  // ── الطباعة على ورقة الشركة ──
  // نفس إطار عرض السعر بالضبط (الشريط الجانبي والبانر السفلي)، حتى
  // الكتاب يطلع بنفس هوية الشركة مو ورقة بيضاء.
  const print = (l: EmployeeLetter) => {
    const esc = (v: string) =>
      v.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
    const vstrip = new URL(vstripUrl, window.location.origin).href
    const banner = new URL(bannerUrl, window.location.origin).href
    const date = new Date(l.createdAt).toLocaleDateString('ar-IQ', { year: 'numeric', month: 'long', day: 'numeric' })

    const html = `<!DOCTYPE html><html dir="rtl" lang="ar"><head><meta charset="UTF-8">
<link href="https://fonts.googleapis.com/css2?family=Cairo:wght@400;600;700&family=Amiri:wght@400;700&display=swap" rel="stylesheet">
<style>
@page { size: A4; margin: 0; }
* { margin: 0; padding: 0; box-sizing: border-box; }
body { background: #fff; font-family: 'Cairo', sans-serif; direction: rtl; -webkit-print-color-adjust: exact; print-color-adjust: exact; }
.page { width: 210mm; min-height: 297mm; position: relative; overflow: hidden; background: #fff; }
.vstrip { position: absolute; top: 0; right: 0; width: 26mm; height: 297mm; z-index: 10; }
.vstrip img { width: 100%; height: 100%; object-fit: cover; display: block; }
.fbanner { position: absolute; bottom: 0; left: 0; right: 0; height: 19mm; z-index: 10; }
.fbanner img { width: 100%; height: 100%; object-fit: cover; display: block; }
.content { position: relative; z-index: 5; padding: 14mm 32mm 24mm 14mm; min-height: 297mm; }
.header { display: flex; justify-content: space-between; align-items: flex-start; padding-bottom: 10px; border-bottom: 1.5px solid #47528f; }
.co-ar { font-size: 16px; font-weight: 700; color: #47528f; line-height: 1.4; }
.co-en { font-size: 11px; font-weight: 600; color: #c97a3a; margin-top: 2px; }
.meta { margin-top: 14px; display: flex; justify-content: space-between; font-size: 13px; color: #47528f; }
.to { margin-top: 26px; font-family: 'Amiri', serif; font-size: 19px; font-weight: 700; color: #47528f; }
.subject { margin-top: 16px; font-size: 16px; font-weight: 700; color: #47528f; }
.subject span { font-weight: 400; }
.body { margin-top: 18px; font-family: 'Amiri', serif; font-size: 16px; line-height: 2.1; color: #1f2937; text-align: justify; white-space: pre-wrap; }
.closing { margin-top: 26px; font-family: 'Amiri', serif; font-size: 16px; color: #47528f; }
.sign { margin-top: 34px; text-align: left; font-family: 'Amiri', serif; font-size: 15px; color: #47528f; line-height: 2; }
.verdict { margin-top: 30px; border-top: 1px dashed #a0a8c8; padding-top: 12px; font-size: 14px; color: #47528f; }
.verdict b { font-size: 15px; }
</style></head><body>
<div class="page">
  <div class="vstrip"><img src="${vstrip}" alt=""></div>
  <div class="fbanner"><img src="${banner}" alt=""></div>
  <div class="content">
    <div class="header">
      <div>
        <div class="co-ar">شركة الأماني للتجارة العامة والاستثمارات العقارية والوكالات التجارية محدودة المسؤولية</div>
        <div class="co-en">Al-Amani for General Trading, Real Estate &amp; Commercial Agencies LLC</div>
      </div>
    </div>
    <div class="meta"><span>التاريخ: ${date}</span></div>
    <div class="to">${esc(l.addressedTo)}</div>
    <div class="subject">م/ <span>${esc(l.subject)}</span></div>
    <div class="body">${esc(l.body)}</div>
    <div class="closing">... ولكم جزيل الشكر والتقدير</div>
    <div class="sign">
      مقدّم الطلب: <b>${esc(l.employee?.name || '')}</b><br>
      ${l.employeeJobTitle ? 'المسمى الوظيفي: ' + esc(l.employeeJobTitle) + '<br>' : ''}
      التوقيع: ........................
    </div>
    ${l.status !== 'PENDING' ? `<div class="verdict">
      <b>قرار الإدارة: ${l.status === 'APPROVED' ? 'موافقة' : 'رفض'}</b>
      ${l.decisionNote ? '<br>' + esc(l.decisionNote) : ''}
      ${l.decidedBy ? '<br>الموقّع: ' + esc(l.decidedBy.name) : ''}
    </div>` : ''}
  </div>
</div>
</body></html>`

    const win = window.open('', '_blank')
    if (!win) return
    win.document.write(html)
    win.document.close()
    // ننتظر الصور حتى ما تطلع الورقة بلا إطار
    const imgs = Array.from(win.document.images)
    void Promise.race([
      Promise.all(imgs.map((im) => im.complete ? Promise.resolve() : new Promise<void>((res) => { im.onload = () => res(); im.onerror = () => res() }))),
      new Promise((res) => setTimeout(res, 4000)),
    ]).then(() => setTimeout(() => win.print(), 150))
  }

  const statusBadge = (s: string) => {
    const map: Record<string, { t: string; c: string }> = {
      PENDING: { t: '⏳ بانتظار الجواب', c: 'bg-amber-50 text-amber-800' },
      APPROVED: { t: '✅ موافقة', c: 'bg-emerald-50 text-emerald-800' },
      REJECTED: { t: '❌ مرفوض', c: 'bg-red-50 text-red-700' },
    }
    const x = map[s] || { t: s, c: 'bg-slate-100 text-slate-600' }
    return <span className={`rounded-full px-2.5 py-0.5 text-xs font-bold ${x.c}`}>{x.t}</span>
  }

  const list = tab === 'mine' ? mine : inbox

  return (
    <div className="space-y-4">
      <div className="rounded-2xl bg-white p-5 shadow-[0_2px_12px_rgba(0,0,0,0.06)]">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h1 className="text-xl font-extrabold text-[#0f2040]">📄 الطلبات</h1>
            <p className="mt-1 text-sm text-slate-500">
              كتاب رسمي للإدارة — تكتبه هنا، يوصلهم فوراً، وتكدر تطبعه بورقة الشركة.
            </p>
          </div>
          <button
            onClick={() => setShowForm((v) => !v)}
            className="rounded-lg bg-[#0f2040] px-4 py-2 text-sm font-bold text-white"
          >
            {showForm ? 'إلغاء' : '✍️ اكتب طلب جديد'}
          </button>
        </div>

        {showForm && (
          <div className="mt-4 space-y-3 rounded-xl border border-slate-200 p-4">
            <div>
              <label className="block text-xs font-bold text-slate-600">إلى</label>
              <select
                value={addressedTo}
                onChange={(e) => setAddressedTo(e.target.value)}
                className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
              >
                {addressees.map((a) => <option key={a} value={a}>{a}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-xs font-bold text-slate-600">م/ الموضوع</label>
              <input
                value={subject}
                onChange={(e) => setSubject(e.target.value)}
                placeholder="مثال: طلب سلفة"
                className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-bold text-slate-600">نص الطلب</label>
              <textarea
                value={body}
                onChange={(e) => setBody(e.target.value)}
                rows={8}
                placeholder="تحية طيبة وبعد...&#10;&#10;اكتب طلبك هنا بالتفصيل."
                className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm leading-7"
              />
            </div>
            <button
              onClick={submit}
              disabled={busy}
              className="rounded-lg bg-emerald-600 px-5 py-2 text-sm font-bold text-white disabled:opacity-50"
            >
              {busy ? 'جاري الإرسال...' : '📨 أرسل الطلب'}
            </button>
            {msg && <p className={`text-sm font-bold ${msg.ok ? 'text-emerald-700' : 'text-red-600'}`}>{msg.text}</p>}
          </div>
        )}
      </div>

      {isAdmin && (
        <div className="flex gap-2">
          <button
            onClick={() => setTab('mine')}
            className={`rounded-lg px-4 py-2 text-sm font-bold ${tab === 'mine' ? 'bg-[#0f2040] text-white' : 'bg-white text-slate-600'}`}
          >
            طلباتي ({mine.length})
          </button>
          <button
            onClick={() => setTab('inbox')}
            className={`rounded-lg px-4 py-2 text-sm font-bold ${tab === 'inbox' ? 'bg-[#0f2040] text-white' : 'bg-white text-slate-600'}`}
          >
            طلبات الموظفين ({inbox.filter((l) => l.status === 'PENDING').length} معلّق)
          </button>
        </div>
      )}

      {loading && <p className="text-slate-400">جاري التحميل...</p>}
      {!loading && list.length === 0 && (
        <p className="rounded-2xl bg-white p-5 text-center text-sm text-slate-400">ماكو طلبات بعد.</p>
      )}

      {list.map((l) => (
        <div key={l.id} className="rounded-2xl bg-white p-5 shadow-[0_2px_12px_rgba(0,0,0,0.06)]">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <div>
              <span className="text-base font-extrabold text-[#0f2040]">م/ {l.subject}</span>
              {tab === 'inbox' && <span className="mr-2 text-sm text-slate-600">— {l.employee?.name}</span>}
            </div>
            {statusBadge(l.status)}
          </div>
          <p className="mt-1 text-xs text-slate-400">
            {l.addressedTo} · {new Date(l.createdAt).toLocaleDateString('ar-IQ', { dateStyle: 'medium' })}
          </p>

          <p className="mt-3 whitespace-pre-wrap rounded-xl bg-slate-50 p-3 text-sm leading-7 text-slate-700">{l.body}</p>

          {l.status !== 'PENDING' && (
            <div className={`mt-3 rounded-lg px-3 py-2 text-sm ${l.status === 'APPROVED' ? 'bg-emerald-50 text-emerald-900' : 'bg-red-50 text-red-800'}`}>
              <b>قرار الإدارة:</b> {l.decisionNote || (l.status === 'APPROVED' ? 'موافقة' : 'رفض')}
              {l.decidedBy && <span className="text-slate-500"> — {l.decidedBy.name}</span>}
            </div>
          )}

          <div className="mt-3 flex flex-wrap items-center gap-2">
            <button
              onClick={() => print(l)}
              className="rounded-lg border border-slate-300 px-4 py-1.5 text-sm font-bold text-slate-700"
            >
              🖨️ اطبعه بورقة الشركة
            </button>
          </div>

          {isAdmin && tab === 'inbox' && l.status === 'PENDING' && (
            <div className="mt-3 rounded-xl border border-slate-200 p-3">
              <input
                value={noteDraft[l.id] || ''}
                onChange={(e) => setNoteDraft({ ...noteDraft, [l.id]: e.target.value })}
                placeholder="جوابك للموظف (إلزامي بالرفض)"
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
              />
              <div className="mt-2 flex gap-2">
                <button onClick={() => decide(l.id, true)} className="rounded-lg bg-emerald-100 px-4 py-2 text-sm font-bold text-emerald-800">
                  ✅ موافقة
                </button>
                <button onClick={() => decide(l.id, false)} className="rounded-lg bg-red-100 px-4 py-2 text-sm font-bold text-red-800">
                  ❌ رفض
                </button>
              </div>
            </div>
          )}
        </div>
      ))}
    </div>
  )
}
