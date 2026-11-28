import { Suspense, lazy, useEffect, useState } from 'react'
import { apiRequest } from '../api'
import { useSession } from '../session'
import { useOrgMemberships, greetingTitle } from '../orgDeskAccess'
import { roleLabel } from '../roleLabels'
import { card } from '../ui/classes'
import MyScoreCard from '../components/MyScoreCard'

// ═══ 👤 ملفي الشخصي (دازاي 11-28) ═══
// «من يضغط الموظف على اسمه يدخل ع الملف الشخصي مالته ويكدر يشوف كلشي يخصه»:
// معلوماتي · تقييمي ومقارنتي بزملائي · دوامي وحضوري · خصوماتي ومنو خصمها.
const MyRanking = lazy(() => import('./MyRanking'))

type Info = {
  name: string; username: string; phone: string | null; role: string; jobTitle: string | null; certificate: string | null
  salary: number | null; shift: string | null; shiftStart: string | null; shiftEnd: string | null; hireDate: string | null
  contractEnd: string | null; experienceYears: number | null; hasDrivingLicense: boolean; isLeader: boolean; createdAt: string
}
type Att = { day: string; checkIn: string | null; checkOut: string | null; minutes: number | null }
type Ded = { source: 'KPI' | 'POINTS' | 'LEDGER'; id: string; amount: number; points: number; reason: string; byName: string | null; bookingCode: string | null; at: string; returned: boolean }
type Profile = { info: Info; attendance: Att[]; deductions: Ded[] }
type Tab = 'info' | 'rating' | 'time' | 'deductions'

const money = (n: number) => `${Math.round(n).toLocaleString('en-US')} د.ع`
const time = (s: string | null) => (s ? new Date(s).toLocaleTimeString('ar-IQ', { hour: '2-digit', minute: '2-digit', timeZone: 'Asia/Baghdad' }) : '—')
const dur = (m: number | null) => (m == null ? 'بعده ما طلع' : `${Math.floor(m / 60)} ساعة و${m % 60} دقيقة`)
const SHIFT: Record<string, string> = { MORNING: 'صباحي', EVENING: 'مسائي' }
const SRC: Record<Ded['source'], string> = { KPI: '💸 خصم', POINTS: '📉 نقاط انضباط', LEDGER: '🧾 تسوية عهدة' }

function Row({ k, v }: { k: string; v: string | null | undefined }) {
  return (
    <div className="flex justify-between gap-3 border-b border-slate-100 py-2 text-sm last:border-0">
      <span className="text-slate-500">{k}</span>
      <b className="text-left text-slate-800">{v || '—'}</b>
    </div>
  )
}

export default function ProfilePage() {
  const { employee } = useSession()
  const { items: orgMine } = useOrgMemberships(employee)
  const [p, setP] = useState<Profile | null>(null)
  const [err, setErr] = useState('')
  const [tab, setTab] = useState<Tab>('info')
  useEffect(() => {
    apiRequest<Profile>('/me/profile').then(setP).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر التحميل'))
  }, [])
  if (err) return <p className="rounded-lg bg-red-50 p-4 text-red-600">{err}</p>
  if (!p || !employee) return <p className="py-10 text-center text-slate-400">جاري التحميل…</p>
  const i = p.info
  const active = p.deductions.filter((d) => !d.returned)
  const tabs: { k: Tab; t: string }[] = [
    { k: 'info', t: '👤 معلوماتي' },
    { k: 'rating', t: '⭐ تقييمي' },
    { k: 'time', t: '🕘 دوامي' },
    { k: 'deductions', t: `💸 خصوماتي${active.length ? ` (${active.length})` : ''}` },
  ]
  return (
    <div className="space-y-4" dir="rtl">
      <div className="rounded-2xl bg-gradient-to-l from-[#0f2040] to-[#1f4488] p-6 text-white shadow">
        <h1 className="text-2xl font-extrabold">{i.name}</h1>
        <p className="mt-1 text-sm text-blue-100">{greetingTitle(employee, orgMine)}{i.jobTitle ? ` · ${i.jobTitle}` : ''}</p>
      </div>
      <div role="tablist" className="flex flex-wrap gap-2">
        {tabs.map((x) => (
          <button key={x.k} type="button" role="tab" aria-selected={tab === x.k} onClick={() => setTab(x.k)}
            className={`rounded-xl px-4 py-2 text-sm font-bold ${tab === x.k ? 'bg-[#0f2040] text-white' : 'bg-white text-slate-700 ring-1 ring-slate-200'}`}>
            {x.t}
          </button>
        ))}
      </div>

      {tab === 'info' && (
        <div className="grid gap-4 md:grid-cols-2">
          <section className={card}>
            <h2 className="mb-2 font-extrabold text-[#0f2040]">بياناتي</h2>
            <Row k="الاسم" v={i.name} />
            <Row k="اسم الدخول" v={i.username} />
            <Row k="رقم الهاتف" v={i.phone} />
            <Row k="الدور" v={roleLabel(i.role)} />
            <Row k="المنصب" v={greetingTitle(employee, orgMine)} />
            <Row k="الشهادة" v={i.certificate} />
            <Row k="سنين الخبرة" v={i.experienceYears != null ? String(i.experienceYears) : null} />
            <Row k="إجازة سوق" v={i.hasDrivingLicense ? 'إي' : 'لا'} />
          </section>
          <section className={card}>
            <h2 className="mb-2 font-extrabold text-[#0f2040]">عملي</h2>
            <Row k="الراتب (يشوفه إنت بس)" v={i.salary ? money(i.salary) : null} />
            <Row k="تاريخ التعيين" v={i.hireDate ?? i.createdAt} />
            <Row k="نهاية العقد" v={i.contractEnd} />
            <Row k="الشفت" v={i.shift ? SHIFT[i.shift] ?? i.shift : null} />
            <Row k="وقت الدوام" v={i.shiftStart && i.shiftEnd ? `${i.shiftStart} ← ${i.shiftEnd}` : null} />
            <p className="mt-3 text-xs text-slate-500">إذا معلومة غلط، كلّم الإدارة حتى يصححوها.</p>
          </section>
        </div>
      )}

      {tab === 'rating' && (
        <div className="space-y-4">
          <MyScoreCard />
          <Suspense fallback={<p className="py-6 text-center text-slate-400">جاري التحميل…</p>}><MyRanking /></Suspense>
        </div>
      )}

      {tab === 'time' && (
        <section className={card}>
          <h2 className="mb-1 font-extrabold text-[#0f2040]">🕘 دوامي وحضوري</h2>
          <p className="mb-3 text-sm text-slate-500">
            دوامك: {i.shiftStart && i.shiftEnd ? `${i.shiftStart} ← ${i.shiftEnd}` : 'ما محدد'}{i.shift ? ` · ${SHIFT[i.shift] ?? i.shift}` : ''}. هذا سجل حضورك وانصرافك لآخر ٤٥ يوم.
          </p>
          {p.attendance.length === 0 ? <p className="text-sm text-slate-400">ماكو تسجيلات حضور.</p> : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead className="text-xs text-slate-500"><tr><th className="text-right">اليوم</th><th>الحضور</th><th>الانصراف</th><th>المدة</th></tr></thead>
                <tbody>
                  {p.attendance.map((a, k) => (
                    <tr key={k} className="border-t border-slate-100">
                      <td className="py-1.5" dir="ltr">{a.day}</td><td className="text-center">{time(a.checkIn)}</td>
                      <td className="text-center">{time(a.checkOut)}</td><td className="text-center">{dur(a.minutes)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      )}

      {tab === 'deductions' && (
        <section className={card}>
          <h2 className="mb-1 font-extrabold text-[#0f2040]">💸 خصوماتي وغراماتي</h2>
          <p className="mb-3 text-sm text-slate-500">كل خصم عليك: شكد، ليش، منو خصمه، ومتى. الي «انرجع» ما ينحسب عليك.</p>
          {p.deductions.length === 0 ? <p className="text-sm text-emerald-600">✅ ماكو عليك أي خصم.</p> : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead className="text-xs text-slate-500"><tr><th className="text-right">النوع</th><th className="text-right">السبب</th><th>المبلغ</th><th>النقاط</th><th>منو خصمه</th><th>متى</th><th>الحالة</th></tr></thead>
                <tbody>
                  {p.deductions.map((d) => (
                    <tr key={d.source + d.id} className={`border-t border-slate-100 ${d.returned ? 'opacity-50' : ''}`}>
                      <td className="py-1.5">{SRC[d.source]}</td>
                      <td>{d.reason}{d.bookingCode ? <span className="mr-1 text-xs text-slate-400">({d.bookingCode})</span> : null}</td>
                      <td className="text-center">{d.amount ? money(d.amount) : '—'}</td>
                      <td className="text-center">{d.points || '—'}</td>
                      <td className="text-center">{d.byName || 'النظام'}</td>
                      <td className="text-center" dir="ltr">{new Date(d.at).toLocaleDateString('en-GB')}</td>
                      <td className="text-center">{d.returned ? '↩️ انرجع' : 'فعّال'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      )}
    </div>
  )
}
