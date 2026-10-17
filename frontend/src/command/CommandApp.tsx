import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import './theme.css'

// ═══ مركز القيادة — الهيكل ═══
//
// طبقة عليا منفصلة عن نظام إدارة الشركة. الفكرة مأخوذة من PPSK
// بالشبكات: **نفس اليوزر**، والباسورد هو الي يحدد أي نظام ينفتح.
//
// بهذي المرحلة: الشكل والهيكل بس — ماكو أي بيانات ولا أوامر تنتقل
// بين الطبقتين. صاحب العمل طلبها هيج بالضبط: «كلشي ماريد منك حاليا
// ترسل للنظام الكبير لا معلومات ولا شي، فقط تهيكله».
//
// ⚠️ ما يشارك النظام الحالي ولا مكوّن ولا لون. الطبقتين لازم يبقن
// منفصلات بصرياً حتى المستخدم يعرف وين هو — الخلط خطير بطبقة
// شغلها إنها تنطي أوامر.
//
// ⚠️ الشاشات فارغة **قصداً ومعلنة**، مو ناقصة بالخفية. كل لوحة
// تكتب بأسفلها شنو تنتظر — حتى ما تنقري «معطّلة» ولا «مكسورة».

type Panel = { title: string; sub: string; pending: string }
type Tab = { key: string; label: string; tiles?: string[]; panels: Panel[] }
type Section = { key: string; label: string; icon: string; tabs: Tab[] }

const SECTIONS: Section[] = [
  {
    key: 'overview',
    label: 'نظرة عامة',
    icon: '◆',
    tabs: [
      {
        key: 'now',
        label: 'الوضع الحالي',
        tiles: ['حجوزات اليوم', 'كوادر بالميدان', 'أوامر معلّقة', 'تنبيهات'],
        panels: [
          { title: 'خريطة الحركة', sub: 'وين الكوادر هسه', pending: 'ينتظر: ربط تتبّع المواقع' },
          { title: 'آخر الأحداث', sub: 'كلشي صار بآخر ساعة', pending: 'ينتظر: مجرى أحداث موحّد' },
        ],
      },
      {
        key: 'pulse',
        label: 'النبض',
        tiles: ['معدّل الإنجاز', 'متوسط التأخير', 'إنتاجية الكادر'],
        panels: [
          { title: 'الاتجاه الأسبوعي', sub: 'صعود ولا نزول', pending: 'ينتظر: تجميع تاريخي' },
          { title: 'نقاط الاختناق', sub: 'وين يعلگ الشغل', pending: 'ينتظر: قياسات التأخير' },
        ],
      },
      {
        key: 'alerts',
        label: 'الإنذارات',
        panels: [
          { title: 'إنذارات حمراء', sub: 'تحتاج قرار هسه', pending: 'ينتظر: قواعد الإنذار' },
          { title: 'انتباهات', sub: 'مو مستعجلة بس مهمة', pending: 'ينتظر: قواعد الإنذار' },
        ],
      },
    ],
  },
  {
    key: 'orders',
    label: 'الأوامر',
    icon: '▶',
    tabs: [
      {
        key: 'issue',
        label: 'إصدار أمر',
        panels: [
          { title: 'أمر جديد', sub: 'لمنو، شنو، وشوقت', pending: 'ينتظر: قناة الأوامر' },
          { title: 'قوالب جاهزة', sub: 'أوامر تتكرر', pending: 'ينتظر: قناة الأوامر' },
        ],
      },
      {
        key: 'live',
        label: 'قيد التنفيذ',
        tiles: ['منفَّذة اليوم', 'قيد التنفيذ', 'متأخرة'],
        panels: [{ title: 'الأوامر الشغّالة', sub: 'وين وصل كل أمر', pending: 'ينتظر: قناة الأوامر' }],
      },
      {
        key: 'log',
        label: 'السجل',
        panels: [{ title: 'كل أمر انصدر', sub: 'منو أصدره ومنو نفّذه ومتى', pending: 'ينتظر: سجل الأوامر' }],
      },
    ],
  },
  {
    key: 'intel',
    label: 'المعلومات',
    icon: '◈',
    tabs: [
      {
        key: 'ops',
        label: 'التشغيل',
        panels: [
          { title: 'الحجوزات', sub: 'نظرة عليا بلا تفاصيل', pending: 'ينتظر: قرارك بشنو ينتقل' },
          { title: 'الكوادر', sub: 'التوزيع والحمل', pending: 'ينتظر: قرارك بشنو ينتقل' },
        ],
      },
      {
        key: 'money',
        label: 'المال',
        tiles: ['واردات', 'مصاريف', 'صافي'],
        panels: [{ title: 'الحركة المالية', sub: 'أرقام مجمّعة', pending: 'ينتظر: قرارك بشنو ينتقل' }],
      },
      {
        key: 'customers',
        label: 'الزبائن',
        panels: [{ title: 'قاعدة الزبائن', sub: 'مجمّعة مو تفصيلية', pending: 'ينتظر: قرارك بشنو ينتقل' }],
      },
    ],
  },
  {
    key: 'units',
    label: 'الأقسام',
    icon: '⬢',
    tabs: [
      {
        key: 'field',
        label: 'الميدان',
        panels: [
          { title: 'الشد والتنفيذ', sub: '', pending: 'ينتظر: ربط القسم' },
          { title: 'الجي بي اس', sub: '', pending: 'ينتظر: ربط القسم' },
          { title: 'الشبكات', sub: '', pending: 'ينتظر: ربط القسم' },
          { title: 'الطاقة الشمسية', sub: '', pending: 'ينتظر: ربط القسم' },
        ],
      },
      {
        key: 'support',
        label: 'الإسناد',
        panels: [
          { title: 'المخزن والكميات', sub: '', pending: 'ينتظر: ربط القسم' },
          { title: 'الأسطول', sub: '', pending: 'ينتظر: ربط القسم' },
          { title: 'الحسابات', sub: '', pending: 'ينتظر: ربط القسم' },
        ],
      },
    ],
  },
  {
    key: 'reports',
    label: 'التقارير',
    icon: '▤',
    tabs: [
      { key: 'daily', label: 'يومي', panels: [{ title: 'تقرير اليوم', sub: '', pending: 'ينتظر: مصدر البيانات' }] },
      { key: 'monthly', label: 'شهري', panels: [{ title: 'تقرير الشهر', sub: '', pending: 'ينتظر: مصدر البيانات' }] },
      { key: 'custom', label: 'مخصص', panels: [{ title: 'ابنِ تقريرك', sub: 'تختار الحقول والمدة', pending: 'ينتظر: مصدر البيانات' }] },
    ],
  },
  {
    key: 'security',
    label: 'الأمان',
    icon: '⛨',
    tabs: [
      {
        key: 'sessions',
        label: 'الجلسات',
        panels: [{ title: 'منو داخل هسه', sub: 'الجهاز والوقت', pending: 'ينتظر: سجل الجلسات' }],
      },
      {
        key: 'audit',
        label: 'سجل التدقيق',
        panels: [{ title: 'كل حركة حسّاسة', sub: 'منو سواها ومتى', pending: 'ينتظر: سجل التدقيق' }],
      },
    ],
  },
]

/** لوحة فارغة مبنية بالكامل — البيانات بس ناقصة. */
function PanelCard({ p }: { p: Panel }) {
  return (
    <div className="cmd-panel">
      <h3>{p.title}</h3>
      {p.sub && <p className="sub">{p.sub}</p>}
      <div><div className="cmd-ghost" /><div className="cmd-ghost" /><div className="cmd-ghost" /></div>
      <p className="cmd-pending">{p.pending}</p>
    </div>
  )
}

export default function CommandApp({ onExit }: { onExit: () => void }) {
  const [active, setActive] = useState(SECTIONS[0].key)
  const section = SECTIONS.find((s) => s.key === active) || SECTIONS[0]
  // ⚠️ التبويب المختار **مشتقّ** مو محفوظ بحالة تنعدّل بـeffect:
  // لمن يتبدّل القسم، المفتاح المحفوظ ما يطابق ولا تبويب بالقسم
  // الجديد فينرجع لأول تبويب لحاله. تصفيره بـeffect چان يخلّي رسمة
  // زايدة يظهر بيها القسم الجديد بتبويب القسم القديم للحظة.
  const [pickedTab, setPickedTab] = useState<string | null>(null)

  const railRef = useRef<HTMLDivElement>(null)
  const navRef = useRef<HTMLDivElement>(null)
  const tabsRef = useRef<HTMLDivElement>(null)
  const [indicator, setIndicator] = useState({ y: 0, h: 0, ready: false })
  const [rail, setRail] = useState({ y: 0, h: 0, ready: false })
  const [beam, setBeam] = useState({ x: 0, w: 0, ready: false })

  const current = section.tabs.find((t) => t.key === pickedTab) || section.tabs[0]
  const tab = current.key

  // المؤشرات كلها **تنقاس** من العناصر نفسها مو تنحسب بأرقام ثابتة:
  // الكلمات العربية أطوالها تختلف، وأي رقم ثابت يخلي الشعاع أقصر أو
  // أطول من الكلمة الي تحتها.
  useLayoutEffect(() => {
    const btn = navRef.current?.querySelector<HTMLElement>(`[data-key="${active}"]`)
    if (btn) setIndicator({ y: btn.offsetTop, h: btn.offsetHeight, ready: true })
    const railBtn = railRef.current?.querySelector<HTMLElement>(`[data-key="${active}"]`)
    if (railBtn) setRail({ y: railBtn.offsetTop, h: railBtn.offsetHeight, ready: true })
  }, [active])

  useLayoutEffect(() => {
    const strip = tabsRef.current
    const el = strip?.querySelector<HTMLElement>(`[data-tab="${tab}"]`)
    if (!strip || !el) return
    // RTL: المسافة من الحافة اليمنى. offsetLeft يقيس من اليسار دائماً
    // بأي اتجاه، فنقلبها بأنفسنا — وإلا الشعاع ينقلب للجهة الغلط.
    const right = strip.scrollWidth - (el.offsetLeft + el.offsetWidth)
    setBeam({ x: -right, w: el.offsetWidth, ready: true })
  }, [tab, active])

  // Esc يرجّعك — مخرج واضح دائماً
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onExit() }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onExit])

  return (
    <div className="cmd-root" dir="rtl">
      <div
        style={{
          position: 'relative', zIndex: 1, display: 'flex', gap: 18,
          padding: 22, height: '100%', boxSizing: 'border-box',
        }}
      >
        {/* ═══ شريط الأيقونات الضيّق ═══ */}
        <aside className="cmd-glass cmd-rail">
          <div className="cmd-dots" style={{ marginBottom: 20, justifyContent: 'center' }}>
            <span className="cmd-dot r" /><span className="cmd-dot y" /><span className="cmd-dot g" />
          </div>
          <div className="cmd-nav" ref={railRef} style={{ flex: 1 }}>
            <div
              className="cmd-indicator"
              style={{ transform: `translateY(${rail.y}px)`, height: rail.h, opacity: rail.ready ? 1 : 0 }}
            />
            {SECTIONS.map((s) => (
              <button
                key={s.key}
                data-key={s.key}
                title={s.label}
                className={`cmd-navitem cmd-railitem${active === s.key ? ' active' : ''}`}
                onClick={() => setActive(s.key)}
              >
                {s.icon}
              </button>
            ))}
          </div>
        </aside>

        {/* ═══ القائمة المفتوحة ═══ */}
        <aside className="cmd-glass" style={{ width: 244, padding: 18, display: 'flex', flexDirection: 'column' }}>
          <p className="cmd-label" style={{ margin: '0 0 4px' }}>الأماني</p>
          <p style={{ margin: 0, fontSize: 21, fontWeight: 800, letterSpacing: -0.3 }}>مركز القيادة</p>
          <div className="cmd-divider" />

          <div className="cmd-nav" ref={navRef} style={{ flex: 1 }}>
            <div
              className="cmd-indicator"
              style={{ transform: `translateY(${indicator.y}px)`, height: indicator.h, opacity: indicator.ready ? 1 : 0 }}
            />
            {SECTIONS.map((s) => (
              <button
                key={s.key}
                data-key={s.key}
                className={`cmd-navitem${active === s.key ? ' active' : ''}`}
                onClick={() => setActive(s.key)}
              >
                <span style={{ opacity: 0.85 }}>{s.icon}</span>
                {s.label}
              </button>
            ))}
          </div>

          <button className="cmd-navitem" onClick={onExit} style={{ marginTop: 8 }}>
            <span style={{ opacity: 0.85 }}>⤶</span> رجوع لنظام الشركة
          </button>
        </aside>

        {/* ═══ المحتوى ═══ */}
        <main className="cmd-glass" style={{ flex: 1, padding: '26px 30px', overflowY: 'auto' }}>
          <h1 className="cmd-title" style={{ margin: 0, fontSize: 30 }}>{section.label}</h1>

          {/* شريط التبويبات بالشعاع المنزلق — من صورة «Modern Navbar» */}
          <div className="cmd-tabs" ref={tabsRef} style={{ marginTop: 16 }}>
            <div
              className="cmd-beam"
              style={{ transform: `translateX(${beam.x}px)`, width: beam.w, opacity: beam.ready ? 1 : 0 }}
            />
            {section.tabs.map((t) => (
              <button
                key={t.key}
                data-tab={t.key}
                className={`cmd-tab${tab === t.key ? ' active' : ''}`}
                onClick={() => setPickedTab(t.key)}
              >
                {t.label}
              </button>
            ))}
          </div>

          {/* المفتاح يشمل القسم والتبويب: بدونه React يعيد استعمال نفس
              العقدة عند تبديل التبويب فما تنعاد حركة الدخول. */}
          <div className="cmd-page" key={`${active}/${tab}`}>
            {current.tiles && (
              <div className="cmd-tiles">
                {current.tiles.map((k) => (
                  <div className="cmd-tile" key={k}>
                    <div className="k">{k}</div>
                    {/* شرطة مو صفر: الرقم **مو موجود**، والصفر معلومة غلط */}
                    <div className="v">—</div>
                    <div className="h">بانتظار المصدر</div>
                  </div>
                ))}
              </div>
            )}

            <div className="cmd-grid2">
              {current.panels.map((p) => <PanelCard p={p} key={p.title} />)}
            </div>

            <p style={{ marginTop: 20, fontSize: 11.5, color: 'rgba(255,255,255,0.32)' }}>
              الهيكل جاهز والشاشات فارغة قصداً — ماكو أي بيانات تنتقل من نظام الشركة لحد الان، مثل ما اتفقنا.
            </p>
          </div>
        </main>
      </div>
    </div>
  )
}
