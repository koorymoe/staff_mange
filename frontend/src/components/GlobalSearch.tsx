import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api'
import { useSession } from '../session'
import { isNavVisible, isPathAllowed, navItems, type NavItem } from './navTree'

// ═══ البحث بكل النظام — طلب (ع) 10-04 ═══
// زر 🔍 ثابت جنب العين (أو Ctrl+K). يدوّر على:
//   • الشاشات والتبويبات — ويطلّع **الطريق** إلها («العمل ← إدارة الموظفين ← الصلاحيات»).
//   • الحجوزات والزبائن والموظفين — من الخادم، ويودّيك عليهم.
// يطلّع بس الي يگدر يفتحه (نفس حارس القائمة)، فما يودّيك لـ«ما عندك صلاحية».

interface Hit { key: string; icon: string; title: string; path: string; to: string; group: string; words?: string }

// تبويبات داخل الشاشات المدموجة — ما إلها بند بالقائمة، فنضيفها يدوياً بطريقها.
const TAB_HITS: { title: string; path: string; to: string; words?: string }[] = [
  { title: 'المتابعة', path: 'مكتب المدير ← المتابعة', to: '/?board=follow' },
  { title: 'الإجراءات', path: 'مكتب المدير ← الإجراءات', to: '/?board=actions' },
  { title: 'مركز قيادة ماتركس', path: 'مكتب المدير ← ماتركس', to: '/?board=matrix', words: 'ماتركس عين' },
  { title: 'العملاء', path: 'العمل ← الحجوزات ← تبويب العملاء', to: '/bookings', words: 'زبائن زبون' },
  { title: 'أرشيف الحجوزات', path: 'العمل ← الحجوزات ← تبويب الأرشيف', to: '/bookings' },
  { title: 'فرص البيع', path: 'العمل ← الحجوزات ← تبويب فرص البيع', to: '/bookings' },
  { title: 'تتبع المهام', path: 'مكتب المراقب ← تتبع المهام', to: '/monitor-desk', words: 'مهام ميدان' },
  { title: 'لوحة المراقبة', path: 'مكتب المراقب ← لوحة المراقبة', to: '/monitor-desk' },
  { title: 'أحكام ماتركس', path: 'مكتب المراقب ← صندوق المراقب ← أحكام ماتركس', to: '/monitor-inbox', words: 'ماتركس' },
  { title: 'تدقيق التكرار', path: 'صندوق المراقب ← تدقيق التكرار', to: '/duplicate-review', words: 'تكرار مكرر ماتركس' },
  { title: 'مهامي الإضافية وإنجازاتي', path: 'مهامي وإنجازاتي', to: '/my-work', words: 'مهام انجاز' },
  { title: 'تقرير ماتركس عن موظف', path: 'مكتب المدير ← ماتركس ← عيون ماتركس ← اسم الموظف', to: '/?board=matrix', words: 'ماتركس تقرير' },
]

// تطبيع عربي بسيط: أ/إ/آ → ا، ة → ه، ى → ي — حتى «الإجراءات» تطلع لـ«اجراءات».
const norm = (s: string) => s.replace(/[ً-ْ]/g, '').replace(/[أإآ]/g, 'ا').replace(/ة/g, 'ه').replace(/ى/g, 'ي')
  .replace(/[^\p{L}\p{N}\s]/gu, ' ').toLowerCase().trim()

export default function GlobalSearch() {
  const { employee, permissions, gpsServiceId } = useSession()
  const navigate = useNavigate()
  const [open, setOpen] = useState(false)
  const [q, setQ] = useState('')
  const [remote, setRemote] = useState<Hit[]>([])
  const [busy, setBusy] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'k') { e.preventDefault(); setOpen(true) }
      if (e.key === 'Escape') setOpen(false)
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [])
  useEffect(() => { if (open) window.setTimeout(() => inputRef.current?.focus(), 30) }, [open])

  const ctx = useMemo(() => ({ employee, permissions, gpsServiceId }), [employee, permissions, gpsServiceId])

  // كل الشاشات الي يگدر يفتحها، ويا طريقها بالقائمة.
  const screens = useMemo(() => {
    const out: Hit[] = []
    const seen = new Set<string>()
    const walk = (items: NavItem[], trail: string[], granted: boolean) => {
      for (const it of items) {
        if (it.divider) continue
        if (!isNavVisible(it, ctx, granted)) continue
        const g = granted || (!!it.unitPermission && (employee?.role === 'ADMIN' || permissions.includes(it.unitPermission)))
        const label = it.label.replace(/^[^\p{L}\p{N}]+/u, '').trim()
        if (it.children) { walk(it.children, [...trail, label], g); continue }
        if (seen.has(it.to)) continue
        seen.add(it.to)
        out.push({ key: 'nav' + it.to, icon: '🧭', title: label, path: [...trail, label].join(' ← '), to: it.to, group: 'الشاشات' })
      }
    }
    walk(navItems, [], false)
    for (const t of TAB_HITS) {
      if (isPathAllowed(t.to, ctx) === false) continue
      // تبويبات «مكتب المدير» للمالك والمدير بس.
      if (t.path.startsWith('مكتب المدير') && employee?.role !== 'ADMIN') continue
      out.push({ key: 'tab' + t.title, icon: '📑', title: t.title, path: t.path, to: t.to, group: 'الشاشات', words: t.words })
    }
    return out
  }, [ctx, employee, permissions])

  const local = useMemo(() => {
    const n = norm(q)
    if (!n) return []
    const words = n.split(/\s+/)
    return screens.filter((h) => { const t = norm(`${h.title} ${h.path} ${h.words ?? ''}`); return words.every((w) => t.includes(w)) }).slice(0, 8)
  }, [q, screens])

  // الحجوزات والزبائن والموظفين — من الخادم، بعد ما يوقف عن الكتابة.
  useEffect(() => {
    const term = q.trim()
    if (term.length < 2) return
    let alive = true
    const t = window.setTimeout(async () => {
      setBusy(true)
      const hits: Hit[] = []
      const safe = async <T,>(p: Promise<T>) => { try { return await p } catch { return null } }
      const canBookings = isPathAllowed('/bookings', ctx) !== false
      const canCustomers = isPathAllowed('/customers', ctx) !== false
      const canEmployees = employee?.role === 'ADMIN' || isPathAllowed('/employees', ctx) === true
      const [bk, cs, em] = await Promise.all([
        canBookings ? safe(api.getBookingsPaged({ bucket: 'all', search: term, page: 1, pageSize: 5 })) : null,
        canCustomers ? safe(api.getCustomers({ search: term, limit: 5 })) : null,
        canEmployees ? safe(api.getEmployees()) : null,
      ])
      for (const b of bk?.items ?? []) {
        hits.push({ key: 'b' + b.id, icon: '📋', title: `${b.code} — ${b.customer?.name ?? ''}`, path: 'العمل ← الحجوزات', to: `/bookings?focus=${b.id}`, group: 'الحجوزات' })
      }
      for (const c of (cs ?? []).slice(0, 5)) {
        hits.push({ key: 'c' + c.id, icon: '👤', title: `${c.name} · ${c.phone}`, path: 'العمل ← الحجوزات ← تبويب العملاء', to: `/customers?q=${encodeURIComponent(c.phone || c.name)}`, group: 'الزبائن' })
      }
      const nt = norm(term)
      for (const e of (em ?? []).filter((x) => norm(x.name).includes(nt)).slice(0, 5)) {
        hits.push({ key: 'e' + e.id, icon: '🧑‍💼', title: e.name, path: 'تقرير ماتركس عن الموظف', to: `/matrix/employee/${e.id}`, group: 'الموظفين' })
      }
      if (alive) { setRemote(hits); setBusy(false) }
    }, 350)
    return () => { alive = false; window.clearTimeout(t) }
  }, [q, ctx, employee])

  // كلمة قصيرة: نتائج الخادم القديمة ما تنعرض (بدل ما نصفّرها داخل الـeffect).
  const all = [...local, ...(q.trim().length < 2 ? [] : remote)]
  const groups = [...new Set(all.map((h) => h.group))]
  const go = (h: Hit) => { setOpen(false); setQ(''); navigate(h.to) }

  return (
    <>
      <button type="button" onClick={() => setOpen(true)} title="دوّر بكل النظام (Ctrl+K)" aria-label="بحث"
        className="grid h-9 w-9 place-items-center rounded-xl text-lg text-slate-600 hover:bg-slate-100">🔍</button>
      {open && (
        <div className="fixed inset-0 z-[80] flex items-start justify-center bg-black/40 p-4 pt-[10vh]" onClick={() => setOpen(false)}>
          <div dir="rtl" className="w-full max-w-xl overflow-hidden rounded-2xl bg-white shadow-2xl" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center gap-2 border-b border-slate-100 p-3">
              <span className="text-lg">🔍</span>
              <input ref={inputRef} value={q} onChange={(e) => setQ(e.target.value)}
                onKeyDown={(e) => { if (e.key === 'Enter' && all[0]) go(all[0]) }}
                placeholder="دوّر عن شاشة، حجز، زبون، موظف… مثلاً: الصلاحيات، ماتركس، B695"
                className="flex-1 bg-transparent text-base outline-none" />
              {busy && <span className="text-xs text-slate-400">يدوّر…</span>}
            </div>
            <div className="max-h-[60vh] overflow-y-auto p-2">
              {!q.trim() ? (
                <p className="p-4 text-center text-sm text-slate-400">اكتب شنو تدوّر — النتيجة تطلع ويا طريقها، واضغط عليها حتى توصلها.</p>
              ) : all.length === 0 && !busy ? (
                <p className="p-4 text-center text-sm text-slate-400">ما لگيت شي بهالاسم.</p>
              ) : groups.map((g) => (
                <div key={g} className="mb-2">
                  <p className="px-2 py-1 text-[11px] font-bold text-slate-400">{g}</p>
                  {all.filter((h) => h.group === g).map((h) => (
                    <button key={h.key} type="button" onClick={() => go(h)}
                      className="flex w-full items-start gap-3 rounded-xl px-3 py-2 text-right hover:bg-sky-50">
                      <span className="text-lg">{h.icon}</span>
                      <span className="min-w-0">
                        <span className="block truncate font-bold text-slate-800">{h.title}</span>
                        <span className="block truncate text-[11px] text-slate-500">📍 {h.path}</span>
                      </span>
                    </button>
                  ))}
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </>
  )
}
