import { useEffect, useState, type ReactNode } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api, type Account, type Branch, type SCard } from '../api'
import ServiceArt from '../components/ServiceArt'
import { LogoMark } from '../components/Logo'
import { Empty, Skel, TopBar } from '../components/ui'
import { CITIES } from '../lib'

const ic = (d: ReactNode, color: string) => (
  <span className="grid h-9 w-9 place-items-center rounded-xl" style={{ background: `${color}14`, color }}>
    <svg width="19" height="19" viewBox="0 0 24 24" fill="currentColor">{d}</svg>
  </span>
)

// ═══ حسابي (شاشة 10) ═══
export default function AccountPage({ me, onLogout }: { me: Account; onLogout: () => void }) {
  const initials = me.name.trim().split(/\s+/).slice(0, 2).map((w) => w[0]).join('')
  const rows: [string, string, ReactNode][] = [
    ['/account/favorites', 'المفضلة', ic(<path d="M12 21s-7-4.4-9.3-9A5.4 5.4 0 0 1 12 6a5.4 5.4 0 0 1 9.3 6c-2.3 4.6-9.3 9-9.3 9z" />, '#e11d48')],
    ['/account/notifications', 'الإشعارات', ic(<path d="M12 22a2.5 2.5 0 0 0 2.5-2.5h-5A2.5 2.5 0 0 0 12 22zm7-6V11a7 7 0 0 0-5.5-6.8V3.5a1.5 1.5 0 0 0-3 0v.7A7 7 0 0 0 5 11v5l-2 2v1h18v-1z" />, '#f59e0b')],
    ['/account/branches', 'فروعنا', ic(<path d="M12 2a7 7 0 0 0-7 7c0 5.3 7 13 7 13s7-7.7 7-13a7 7 0 0 0-7-7zm0 9.5A2.5 2.5 0 1 1 12 6.5a2.5 2.5 0 0 1 0 5z" />, '#0d3b8c')],
    ['/matrix', 'مركز المساعدة والدعم', ic(<path d="M12 2a9 9 0 0 0-9 9v6a3 3 0 0 0 3 3h2v-8H5v-1a7 7 0 0 1 14 0v1h-3v8h3a3 3 0 0 0 3-3v-6a9 9 0 0 0-9-9z" />, '#0d3b8c')],
    ['/account/about', 'عن شركة الأماني', ic(<path d="M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20zm1 15h-2v-6h2zm0-8h-2V7h2z" />, '#0d3b8c')],
    ['/account/privacy', 'سياسة الخصوصية', ic(<path d="M12 2L4 5v6c0 5 3.4 9.7 8 11 4.6-1.3 8-6 8-11V5zm-1.2 14.2L7 12.4l1.4-1.4 2.4 2.4 4.8-4.8L17 10z" />, '#0d3b8c')],
  ]
  return (
    <div className="page-in pb-4">
      <TopBar title="حسابي" back={false} />
      <div className="space-y-4 px-4 pt-1">
        <div className="card flex items-center gap-4 p-4">
          <span className="grid h-[68px] w-[68px] shrink-0 place-items-center rounded-full bg-gradient-to-br from-navy to-navy-2 text-[22px] font-black text-white ring-4 ring-blue-50">{initials}</span>
          <div className="min-w-0 flex-1">
            <p className="truncate text-[18px] font-black text-ink">{me.name}</p>
            <p className="text-[13px] text-mute" dir="ltr">{me.phone}</p>
            <Link to="/account/edit" className="mt-1.5 inline-block rounded-lg bg-blue-50 px-3 py-1 text-[12px] font-bold text-navy">تعديل الملف الشخصي</Link>
          </div>
        </div>
        <div className="card divide-y divide-line">
          {rows.map(([to, label, icon]) => (
            <Link key={to} to={to} className="flex items-center gap-3 px-4 py-3.5 active:bg-slate-50">
              {icon}<span className="flex-1 text-[14.5px] font-bold text-ink">{label}</span>
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#9aa5bd" strokeWidth="2.2" strokeLinecap="round"><path d="M15 6l-6 6 6 6" /></svg>
            </Link>
          ))}
          <button onClick={() => { if (confirm('تريد تسجيل الخروج؟')) onLogout() }} className="flex w-full items-center gap-3 px-4 py-3.5 text-right active:bg-rose-50">
            {ic(<path d="M10 17l1.4-1.4L8.8 13H20v-2H8.8l2.6-2.6L10 7l-5 5zM4 5h8V3H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h8v-2H4z" />, '#e11d48')}
            <span className="flex-1 text-[14.5px] font-bold text-rose-600">تسجيل الخروج</span>
          </button>
        </div>
        <p className="text-center text-[11px] text-mute">شركة الأماني © {new Date().getFullYear()}</p>
      </div>
    </div>
  )
}

export function EditProfilePage({ me, onChange }: { me: Account; onChange: (a: Account) => void }) {
  const nav = useNavigate()
  const [f, setF] = useState({ name: me.name, phone: me.phone, city: me.city ?? '' })
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const save = async () => {
    setBusy(true); setErr('')
    try { onChange(await api.updateMe(f)); nav(-1) } catch (e) { setErr(e instanceof Error ? e.message : 'خطأ') } finally { setBusy(false) }
  }
  const field = 'h-[50px] w-full rounded-xl border border-line bg-white px-4 text-[15px] outline-none focus:border-navy focus:ring-4 focus:ring-navy/10'
  return (
    <div className="page-in min-h-full bg-white">
      <TopBar title="تعديل الملف الشخصي" />
      <div className="space-y-4 px-5 pt-2">
        <label className="block"><span className="mb-1.5 block text-[14px] font-extrabold">الاسم</span><input className={field} value={f.name} onChange={(e) => setF({ ...f, name: e.target.value })} /></label>
        <label className="block"><span className="mb-1.5 block text-[14px] font-extrabold">رقم الهاتف</span><input className={field} dir="ltr" inputMode="tel" value={f.phone} onChange={(e) => setF({ ...f, phone: e.target.value })} /></label>
        <label className="block"><span className="mb-1.5 block text-[14px] font-extrabold">المحافظة</span>
          <select className={field} value={f.city} onChange={(e) => setF({ ...f, city: e.target.value })}><option value="">اختر</option>{CITIES.map((c) => <option key={c}>{c}</option>)}</select></label>
        <p className="text-[12px] text-mute">البريد: <span dir="ltr">{me.email}</span> · اسم المستخدم: <span dir="ltr">@{me.username}</span></p>
        {err && <p className="rounded-xl bg-rose-50 p-3 text-[13px] text-rose-600">{err}</p>}
        <button disabled={busy} onClick={save} className="h-[52px] w-full rounded-xl bg-navy-2 font-extrabold text-white disabled:opacity-50">حفظ</button>
      </div>
    </div>
  )
}

export function FavoritesPage() {
  const [rows, setRows] = useState<SCard[] | null>(null)
  useEffect(() => { api.favorites().then(setRows).catch(() => setRows([])) }, [])
  return (
    <div className="page-in">
      <TopBar title="المفضلة" />
      <div className="grid grid-cols-2 gap-3 px-4 pt-1">
        {!rows && [0, 1].map((i) => <Skel key={i} className="h-40" />)}
        {rows?.map((s) => (
          <Link key={s.id} to={`/services/${s.id}`} className="card overflow-hidden">
            <ServiceArt id={s.id} name={s.name} hasImage={s.hasImage} className="h-28 w-full" />
            <p className="truncate px-2 py-2.5 text-center text-[14px] font-extrabold">{s.name}</p>
          </Link>
        ))}
      </div>
      {rows?.length === 0 && <div className="px-4"><Empty icon="🤍" title="ماكو خدمات بالمفضلة"><p className="mt-1 text-[13px] text-mute">اضغط ❤️ بصفحة أي خدمة حتى تنحفظ هنا.</p></Empty></div>}
    </div>
  )
}

export function BranchesPage() {
  const [rows, setRows] = useState<Branch[] | null>(null)
  useEffect(() => { api.branches().then(setRows).catch(() => setRows([])) }, [])
  return (
    <div className="page-in">
      <TopBar title="فروعنا" />
      <div className="space-y-3 px-4 pt-1">
        {!rows && <Skel className="h-24" />}
        {rows?.map((b) => (
          <div key={b.id} className="card p-4">
            <p className="text-[15px] font-extrabold text-ink">📍 {b.name}</p>
            {b.address && <p className="mt-1 text-[13px] text-mute">{b.address}</p>}
            <div className="mt-3 flex gap-2">
              {b.phone && <a href={`tel:${b.phone}`} className="rounded-lg bg-blue-50 px-3 py-1.5 text-[12px] font-bold text-navy" dir="ltr">📞 {b.phone}</a>}
              {b.mapUrl && <a href={b.mapUrl} target="_blank" rel="noreferrer" className="rounded-lg bg-blue-50 px-3 py-1.5 text-[12px] font-bold text-navy">🗺️ الموقع على الخريطة</a>}
            </div>
          </div>
        ))}
        {rows?.length === 0 && <Empty icon="🏢" title="معلومات الفروع قريباً"><p className="mt-1 text-[13px] text-mute">تكدر تسأل ماتركس عن أقرب فرع إلك.</p></Empty>}
      </div>
    </div>
  )
}

export function NotificationsPage() {
  return (
    <div className="page-in">
      <TopBar title="الإشعارات" />
      <div className="px-4"><Empty icon="🔔" title="ماكو إشعارات جديدة"><p className="mt-1 text-[13px] text-mute">تحديثات حجوزاتك والعروض الجديدة تطلع هنا.</p></Empty></div>
    </div>
  )
}

export function AboutPage() {
  return (
    <div className="page-in min-h-full bg-white">
      <TopBar title="عن شركة الأماني" />
      <div className="px-6 pt-4 text-center">
        <div className="flex justify-center"><LogoMark size={80} /></div>
        <p className="mt-3 text-[20px] font-black text-navy">شركة الأماني</p>
        <p className="mt-1 text-[13px] text-mute">للتجارة العامة والاستثمارات العقارية والوكالات التجارية</p>
        <p className="mt-5 text-right text-[14px] leading-8 text-[#4b5672]">
          نقدم حلولاً تقنية وهندسية متكاملة للبيوت والأعمال: كاميرات المراقبة، الشبكات، الطاقة الشمسية، المنازل الذكية، أجهزة التتبع وغيرها —
          بفرق فنية متخصصة، ومتابعة لكل طلب من لحظة الحجز لحد ما يخلص الشغل.
        </p>
      </div>
    </div>
  )
}

export function PrivacyPage() {
  return (
    <div className="page-in min-h-full bg-white">
      <TopBar title="سياسة الخصوصية" />
      <div className="space-y-3 px-6 pt-3 text-[14px] leading-8 text-[#4b5672]">
        <p>نستخدم بياناتك (الاسم، الهاتف، البريد، العنوان) لخدمتك وتنفيذ طلباتك والتواصل وياك فقط.</p>
        <p>ما نبيع بياناتك ولا نشاركها مع أي جهة خارج الشركة.</p>
        <p>محادثاتك ويّا ماتركس تنحفظ حتى نحسّن الخدمة ونتابع طلباتك.</p>
        <p>تكدر تطلب حذف حسابك بأي وقت عن طريق مركز المساعدة.</p>
      </div>
    </div>
  )
}
