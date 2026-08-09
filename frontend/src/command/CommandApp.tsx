import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import './theme.css'

// ═══ مركز القيادة — الهيكل ═══
//
// طبقة عليا منفصلة عن نظام إدارة الشركة. الفكرة مأخوذة من PPSK
// بالشبكات: **نفس اليوزر**، والباسورد هو الي يحدد أي نظام ينفتح.
//
// بهذي المرحلة: الشكل والدخول بس — ماكو أي بيانات ولا أوامر تنتقل
// بين الطبقتين. صاحب العمل طلبها هيج بالضبط: «كلشي ماريد منك حاليا
// ترسل للنظام الكبير لا معلومات ولا شي، فقط تهيكله».
//
// ⚠️ ما يشارك النظام الحالي ولا مكوّن ولا لون. الطبقتين لازم يبقن
// منفصلات بصرياً حتى المستخدم يعرف وين هو — الخلط خطير بطبقة
// شغلها إنها تنطي أوامر.

const SECTIONS = [
  { key: 'overview', label: 'نظرة عامة', icon: '◆' },
  { key: 'orders', label: 'الأوامر', icon: '▶' },
  { key: 'intel', label: 'المعلومات', icon: '◈' },
  { key: 'units', label: 'الأقسام', icon: '⬢' },
  { key: 'reports', label: 'التقارير', icon: '▤' },
]

/** الشاشة الفارغة — تقول صراحةً إنها هيكل، مو تتظاهر إنها معطّلة. */
function Placeholder({ title }: { title: string }) {
  return (
    <div className="cmd-page">
      <h1 className="cmd-title">{title}</h1>
      <div className="cmd-divider" />
      <div className="cmd-glass cmd-card" style={{ marginTop: 18 }}>
        <p style={{ color: 'var(--cmd-muted)', lineHeight: 2, margin: 0 }}>
          الهيكل جاهز. هاي الشاشة فارغة قصداً — ماكو أي بيانات تنتقل من نظام
          الشركة لحد الان، مثل ما اتفقنا.
        </p>
      </div>
    </div>
  )
}

export default function CommandApp({ onExit }: { onExit: () => void }) {
  const [active, setActive] = useState(SECTIONS[0].key)
  const navRef = useRef<HTMLDivElement>(null)
  const railRef = useRef<HTMLDivElement>(null)
  const [indicator, setIndicator] = useState({ y: 0, h: 0, ready: false })
  const [rail, setRail] = useState({ y: 0, h: 0, ready: false })

  // المؤشر المنزلق: عنصر واحد ينتقل بالـtransform لمحل الزر الفعّال.
  // نقيس المحل بدل ما نحسبه بارتفاع ثابت — لأن أي تغيير بالحشو أو حجم
  // الخط چان خلّى المؤشر ينزاح عن مكانه.
  useLayoutEffect(() => {
    const nav = navRef.current
    if (!nav) return
    const btn = nav.querySelector<HTMLElement>(`[data-key="${active}"]`)
    if (btn) setIndicator({ y: btn.offsetTop, h: btn.offsetHeight, ready: true })
    const railBtn = railRef.current?.querySelector<HTMLElement>(`[data-key="${active}"]`)
    if (railBtn) setRail({ y: railBtn.offsetTop, h: railBtn.offsetHeight, ready: true })
  }, [active])

  // Esc يرجّعك — مخرج واضح دائماً
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onExit() }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onExit])

  const current = SECTIONS.find((s) => s.key === active) || SECTIONS[0]

  return (
    <div className="cmd-root" dir="rtl">
      <div
        style={{
          position: 'relative', zIndex: 1, display: 'flex', gap: 18,
          padding: 22, height: '100%', boxSizing: 'border-box',
        }}
      >
        {/* ═══ شريط الأيقونات الضيّق ═══
            الترتيب مأخوذ من الصورة الثانية: شريط أيقونات لحاله جنب
            القائمة المفتوحة، لوحتين منفصلات طايرات مو لوحة وحدة. */}
        <aside className="cmd-glass cmd-rail">
          <div className="cmd-dots" style={{ marginBottom: 20, justifyContent: 'center' }}>
            <span className="cmd-dot r" /><span className="cmd-dot y" /><span className="cmd-dot g" />
          </div>
          <div className="cmd-nav" ref={railRef} style={{ flex: 1 }}>
            {/* نفس المؤشر المنزلق مالت الصورة الأولى، بس عمودي */}
            <div
              className="cmd-indicator"
              style={{
                transform: `translateY(${rail.y}px)`,
                height: rail.h,
                opacity: rail.ready ? 1 : 0,
              }}
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
        <aside className="cmd-glass" style={{ width: 262, padding: 18, display: 'flex', flexDirection: 'column' }}>
          <p className="cmd-label" style={{ margin: '0 0 4px' }}>الأماني</p>
          <p style={{ margin: 0, fontSize: 21, fontWeight: 800, letterSpacing: -0.3 }}>مركز القيادة</p>
          <div className="cmd-divider" />

          <div className="cmd-nav" ref={navRef} style={{ flex: 1 }}>
            <div
              className="cmd-indicator"
              style={{
                transform: `translateY(${indicator.y}px)`,
                height: indicator.h,
                opacity: indicator.ready ? 1 : 0,
              }}
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

        {/* المحتوى */}
        <main className="cmd-glass" style={{ flex: 1, padding: 30, overflowY: 'auto' }}>
          <Placeholder title={current.label} />
        </main>
      </div>
    </div>
  )
}
