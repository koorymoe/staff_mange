import { useCallback, useEffect, useMemo, useState } from 'react'
import { api, type Booking, type StageBucket } from '../api'

// ═══ سلال المراحل ═══
//
// «حجوزات مؤجلة وزبون ما رد وزبون ألغى — هاي تصير بيها حالتين: قبل
// التثبيت وبعد التثبيت».
//
// وهو محق: الزبون الي ألغى **قبل** ما نثبتله موعد ما كلّفنا شي، والي
// ألغى **بعد** ما وعدناه وحضّرنا كادر وحجزنا يومه كلّفنا فعلاً. دمج
// الاثنين بسلّة وحدة يخبّي فرقاً بالمسؤولية وبالخسارة.
//
// ⚠️ إلا التأجيل: «مؤجّل قبل التثبيت ينلغي، ما عدي هيج شي — يعني شلون
// أأجّل موعد وأني أصلاً ما محددله موعد؟». والكود يوافقه: التأجيل ينقل
// الحجز من موعد قديم لموعد جديد، فالحجز بلا موعد ما يوصله أبداً.
// فصارت «مؤجّلة» سلّة وحدة بلا انقسام.
//
// ⚠️ الفرز يصير بالسيرفر: الواجهة چان لازم تجيب كل حجوزات الشركة
// وتفرزهن بالمتصفح، وهذا يثقل كل ما تتراكم البيانات.

const PAGE_SIZE = 10

/** شنو تعني كل سلّة، وشنو عنوان آخر عمودين بيها. */
const BUCKET_META: Record<StageBucket, { note: string; reasonHead: string; whenHead: string }> = {
  POSTPONED_AFTER_CONFIRM: {
    note: 'هذي الحجوزات انأجّلت بعد ما انثبّتت — ينتظرن موعداً جديداً من الإداري.',
    reasonHead: 'سبب التأجيل',
    whenHead: 'تاريخ التأجيل',
  },
  NO_ANSWER_BEFORE_CONFIRM: {
    note: 'اتصلنا بالزبون قبل التثبيت وما رد — الحجز ما كلّفنا شي لحد الآن.',
    reasonHead: 'ملاحظة الانتظار',
    whenHead: 'بالانتظار من',
  },
  NO_ANSWER_AFTER_CONFIRM: {
    note: 'حجز مثبّت والزبون ما يرد — هذا الي يكلّف: يومه محجوز وكادره منتظر.',
    reasonHead: 'ملاحظة الانتظار',
    whenHead: 'بالانتظار من',
  },
  CANCELLED_BEFORE_CONFIRM: {
    note: 'انلغى قبل التثبيت — ما حضّرنا إله كادر ولا حجزنا يوم.',
    reasonHead: 'سبب الإلغاء',
    whenHead: 'تاريخ الإلغاء',
  },
  CANCELLED_AFTER_CONFIRM: {
    note: 'انلغى بعد ما وعدنا الزبون وحضّرنا كادر وحجزنا يومه — هذي خسارة فعلية.',
    reasonHead: 'سبب الإلغاء',
    whenHead: 'تاريخ الإلغاء',
  },
}

// ═══ خيار واحد بالأعلى، والتفصيل جوّاه ═══
//
// «هذن ينلغن، يصير بدالهن (زبون ما رد) ومن ندخل عليه يطلعلنا: ما رد
// قبل التثبيت لو بعد التثبيت؟».
//
// الفرق بين قبل وبعد التثبيت يبقى محفوظ — بس ما ياخذ خيارين بالصف
// الأعلى. الصف الأعلى يجاوب «شنو صار بالحجز»، والتفصيل «بأي لحظة».
interface Group {
  key: string
  icon: string
  label: string
  /** أكثر من وحدة = تطلع خيارات فرعية جوّا الشاشة */
  buckets: { key: StageBucket; label: string }[]
}

const GROUPS: Group[] = [
  { key: 'POSTPONED', icon: '⏳', label: 'حجوزات مؤجّلة', buckets: [{ key: 'POSTPONED_AFTER_CONFIRM', label: 'مؤجّلة' }] },
  {
    key: 'NO_ANSWER', icon: '📵', label: 'زبون ما رد',
    buckets: [
      { key: 'NO_ANSWER_BEFORE_CONFIRM', label: 'قبل التثبيت' },
      { key: 'NO_ANSWER_AFTER_CONFIRM', label: 'بعد التثبيت' },
    ],
  },
  {
    key: 'CANCELLED', icon: '✖️', label: 'حجوزات ملغية',
    buckets: [
      { key: 'CANCELLED_BEFORE_CONFIRM', label: 'قبل التثبيت' },
      { key: 'CANCELLED_AFTER_CONFIRM', label: 'بعد التثبيت' },
    ],
  },
]

/** لون ثابت للزبون من اسمه — نفس الاسم يعطي نفس اللون كل مرة. */
const AVATAR_COLORS = [
  'bg-violet-500', 'bg-sky-500', 'bg-emerald-500', 'bg-amber-500',
  'bg-rose-500', 'bg-indigo-500', 'bg-teal-500', 'bg-orange-500',
]
function colorOf(seed: string) {
  let h = 0
  for (let i = 0; i < seed.length; i++) h = (h * 31 + seed.charCodeAt(i)) >>> 0
  return AVATAR_COLORS[h % AVATAR_COLORS.length]
}

function serviceNames(b: Booking) {
  if (b.services && b.services.length > 0) return b.services.map((s) => s.name).join(' + ')
  return b.service?.name || 'بدون خدمة محددة'
}

/** التاريخ واليوم منفصلين — الإداري يقرا اليوم أسرع من الرقم. */
function dateParts(iso?: string | null) {
  if (!iso) return null
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return null
  const pad = (n: number) => String(n).padStart(2, '0')
  return {
    date: `${d.getFullYear()}/${pad(d.getMonth() + 1)}/${pad(d.getDate())}`,
    weekday: d.toLocaleDateString('ar-IQ', { weekday: 'long' }),
    time: d.toLocaleTimeString('ar-IQ', { hour: '2-digit', minute: '2-digit', hour12: true }).replace(/[صم].*$/, '').trim(),
    period: d.getHours() < 12 ? 'صباحاً' : d.getHours() < 17 ? 'عصراً' : 'مساءً',
  }
}

export default function StageBucketsPage() {
  const [bucket, setBucket] = useState<StageBucket>('POSTPONED_AFTER_CONFIRM')
  const [rows, setRows] = useState<Booking[]>([])
  const [counts, setCounts] = useState<Record<string, number>>({})
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)

  const load = useCallback(() => {
    // ⚠️ التحميل بـtimeout مو بجسم الأثر: `setState` مباشرة جوّا الأثر
    // تسبّب رندرات متتالية (قاعدة react-hooks/set-state-in-effect).
    const t = setTimeout(() => {
    setLoading(true)
    setPage(1)
    Promise.all([api.getBookingsByStageBucket(bucket), api.getStageBucketCounts()])
      .then(([list, c]) => { setRows(list); setCounts(c) })
      .catch(() => setRows([]))
      .finally(() => setLoading(false))
    }, 0)
    return () => clearTimeout(t)
  }, [bucket])

  useEffect(load, [load])

  const group = GROUPS.find((g) => g.buckets.some((b) => b.key === bucket)) ?? GROUPS[0]
  const subLabel = group.buckets.find((b) => b.key === bucket)?.label ?? ''
  const info = BUCKET_META[bucket]
  const pages = Math.max(1, Math.ceil(rows.length / PAGE_SIZE))
  const shown = useMemo(() => rows.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE), [rows, page])

  /** السبب والوقت يفرقون حسب السلّة — نفس العمود، مصدر مختلف. */
  const reasonOf = (b: Booking) =>
    bucket.startsWith('CANCELLED') ? b.cancelReason
      : bucket.startsWith('NO_ANSWER') ? b.waitingNote
        : b.postponeReason
  const whenOf = (b: Booking) =>
    bucket.startsWith('CANCELLED') ? b.cancelledAt
      : bucket.startsWith('NO_ANSWER') ? b.waitingSince
        : b.lastPostponedAt

  return (
    <div dir="rtl">
      <h2 className="text-xl font-extrabold text-[#0f2040]">🗂️ حجوزات ما وصلت للتنفيذ</h2>
      <p className="mt-1 text-sm text-slate-500">
        المؤجّل والي ما رد والملغى — الإلغاء وعدم الرد مفروزين: صار <b>قبل</b> التثبيت لو <b>بعده</b>.
      </p>

      {/* الخيارات — تلتف بالموبايل بدل ما تطلع شريط ينقص من الجهة */}
      <div className="mt-4 flex flex-wrap gap-2">
        {GROUPS.map((g) => {
          const total = g.buckets.reduce((n, b) => n + (counts[b.key] || 0), 0)
          const active = g.key === group.key
          return (
            <button
              key={g.key}
              // الضغط على المجموعة يفتح **أول** سلّة بيها — ما نحتفظ
              // بآخر خيار فرعي: الإداري يضغط «ما رد» فيتوقع يشوف
              // البداية، مو الي كان فاتحه قبل نص ساعة.
              onClick={() => setBucket(g.buckets[0].key)}
              className={`rounded-xl px-3 py-2 text-[11px] font-bold transition-colors sm:px-3.5 sm:text-sm ${
                active ? 'bg-[#0f2040] text-white shadow-md' : 'bg-white text-slate-600 hover:bg-slate-50'
              }`}
            >
              {g.icon} {g.label}
              {total > 0 && (
                <span className={`mr-1.5 rounded-full px-1.5 text-[10px] ${active ? 'bg-white/20' : 'bg-slate-100'}`}>
                  {total}
                </span>
              )}
            </button>
          )
        })}
      </div>

      <div className="mt-4 overflow-hidden rounded-2xl bg-white shadow-[0_2px_16px_rgba(15,32,64,0.07)]">
        <div className="border-b border-slate-100 px-5 py-4">
          <h3 className="text-base font-extrabold text-[#0f2040]">
            {group.icon} {group.label}
            {group.buckets.length > 1 && <span className="text-slate-400"> — {subLabel}</span>}
          </h3>
          <p className="mt-0.5 text-xs text-slate-400">{info.note}</p>

          {/* ═══ التفصيل: بأي لحظة صار؟ ═══
              ما ينعرض إلا لمن يكون إله معنى — «مؤجّلة» ماكو بيها
              قبل/بعد أصلاً، فخيار وحيد بلا بديل ضجيج. */}
          {group.buckets.length > 1 && (
            <div className="mt-3 inline-flex rounded-xl bg-slate-100 p-1">
              {group.buckets.map((sb) => (
                <button
                  key={sb.key}
                  onClick={() => setBucket(sb.key)}
                  className={`rounded-lg px-3 py-1.5 text-[11px] font-bold transition-colors ${
                    bucket === sb.key ? 'bg-white text-[#0f2040] shadow-sm' : 'text-slate-500 hover:text-slate-700'
                  }`}
                >
                  {sb.label}
                  {counts[sb.key] > 0 && <span className="mr-1 text-slate-400">({counts[sb.key]})</span>}
                </button>
              ))}
            </div>
          )}
        </div>

        {loading && <p className="px-5 py-10 text-center text-slate-400">جاري التحميل...</p>}

        {!loading && rows.length === 0 && (
          <p className="px-5 py-12 text-center text-slate-400">ماكو حجوزات بـ«{group.label}{group.buckets.length > 1 ? ` — ${subLabel}` : ''}» ✓</p>
        )}

        {!loading && rows.length > 0 && (
          <>
            {/* ⚠️ التمرير جوّا الجدول مو بالصفحة: سبع أعمدة على شاشة
                موبايل تدفع الصفحة كلها للجنب وتخرّب القائمة الجانبية. */}
            <div className="overflow-x-auto">
              <table className="w-full min-w-[860px] text-right">
                <thead>
                  <tr className="text-[11px] font-bold text-slate-400">
                    <th className="px-4 py-3 font-bold">#</th>
                    <th className="px-4 py-3 font-bold">الزبون</th>
                    <th className="px-4 py-3 font-bold">الخدمة</th>
                    <th className="px-4 py-3 font-bold">التاريخ</th>
                    <th className="px-4 py-3 font-bold">الوقت</th>
                    <th className="px-4 py-3 font-bold">{info.reasonHead}</th>
                    <th className="px-4 py-3 font-bold">{info.whenHead}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-50">
                  {shown.map((b, i) => {
                    const name = b.customer?.name || 'زبون غير معروف'
                    const when = dateParts(b.scheduledAt)
                    const evt = dateParts(whenOf(b))
                    return (
                      <tr key={b.id} className="transition-colors hover:bg-slate-50/70">
                        <td className="px-4 py-3 text-xs font-bold text-slate-400">{(page - 1) * PAGE_SIZE + i + 1}</td>
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-2.5">
                            <span className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-sm font-black text-white ${colorOf(name)}`}>
                              {name.trim().charAt(0)}
                            </span>
                            <div className="min-w-0">
                              <p className="truncate text-[13px] font-bold text-[#0f2040]">{name}</p>
                              <p className="text-[11px] text-slate-400" dir="ltr">{b.customer?.phone || '—'}</p>
                            </div>
                          </div>
                        </td>
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-2">
                            <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-brand-50 text-sm">🧰</span>
                            <span className="text-[12px] font-semibold leading-tight text-slate-600">{serviceNames(b)}</span>
                          </div>
                        </td>
                        <td className="px-4 py-3">
                          {when ? (
                            <div className="flex items-start gap-1.5">
                              <span className="text-xs">📅</span>
                              <div>
                                <p className="text-[12px] font-bold text-slate-700">{when.date}</p>
                                <p className="text-[11px] text-slate-400">{when.weekday}</p>
                              </div>
                            </div>
                          ) : (
                            // الحجز المؤجل بلا موعد جديد — هذي حالته الطبيعية
                            <span className="text-[11px] font-bold text-amber-600">بلا موعد</span>
                          )}
                        </td>
                        <td className="px-4 py-3">
                          {when ? (
                            <div className="flex items-start gap-1.5">
                              <span className="text-xs">🕐</span>
                              <div>
                                <p className="text-[12px] font-bold text-slate-700">{when.time}</p>
                                <p className="text-[11px] text-slate-400">{when.period}</p>
                              </div>
                            </div>
                          ) : <span className="text-slate-300">—</span>}
                        </td>
                        <td className="px-4 py-3">
                          {reasonOf(b) ? (
                            <span className="flex items-center gap-1.5 text-[12px] text-slate-600">
                              <span className={`h-1.5 w-1.5 shrink-0 rounded-full ${
                                bucket.startsWith('CANCELLED') ? 'bg-red-500' : bucket.startsWith('NO_ANSWER') ? 'bg-slate-400' : 'bg-amber-500'
                              }`} />
                              {reasonOf(b)}
                            </span>
                          ) : (
                            // ⚠️ «بلا سبب» تنعرض صريحة مو شرطة: السبب الناقص
                            // معلومة بحد ذاتها — منو أجّل بلا ما يعلّل؟
                            <span className="text-[11px] text-slate-300">ما انكتب سبب</span>
                          )}
                        </td>
                        <td className="px-4 py-3">
                          {evt ? (
                            <div className="flex items-start gap-1.5">
                              <span className="text-xs">📅</span>
                              <div>
                                <p className="text-[12px] font-bold text-slate-700">{evt.date}</p>
                                <p className="text-[11px] text-slate-400">{evt.time} {evt.period}</p>
                              </div>
                            </div>
                          ) : <span className="text-slate-300">—</span>}
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>

            <div className="flex flex-wrap items-center justify-between gap-3 border-t border-slate-100 px-4 py-3">
              <span className="rounded-lg border border-slate-200 px-3 py-1.5 text-[11px] font-bold text-slate-500">
                عرض {shown.length} من {rows.length} حجز
              </span>
              {pages > 1 && (
                <div className="flex items-center gap-1">
                  <button
                    onClick={() => setPage((p) => Math.max(1, p - 1))}
                    disabled={page === 1}
                    className="rounded-lg px-2 py-1 text-slate-400 hover:bg-slate-100 disabled:opacity-30"
                  >‹</button>
                  {Array.from({ length: pages }, (_, i) => i + 1)
                    // ⚠️ ما نطبع خمسين رقم: أول وآخر والي حوالين الصفحة الحالية
                    .filter((n) => n === 1 || n === pages || Math.abs(n - page) <= 1)
                    .map((n, idx, arr) => (
                      <span key={n} className="flex items-center gap-1">
                        {idx > 0 && n - arr[idx - 1] > 1 && <span className="px-1 text-slate-300">…</span>}
                        <button
                          onClick={() => setPage(n)}
                          className={`h-7 w-7 rounded-lg text-[11px] font-bold transition-colors ${
                            n === page ? 'bg-[#0f2040] text-white' : 'text-slate-500 hover:bg-slate-100'
                          }`}
                        >{n}</button>
                      </span>
                    ))}
                  <button
                    onClick={() => setPage((p) => Math.min(pages, p + 1))}
                    disabled={page === pages}
                    className="rounded-lg px-2 py-1 text-slate-400 hover:bg-slate-100 disabled:opacity-30"
                  >›</button>
                </div>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  )
}
