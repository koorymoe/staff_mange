import { useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react'
import { apiRequest } from '../../api'
import { LineIcon } from '../../components/LineIcon'
import bannerDay from '../../assets/solar-banner-day.webp'
import bannerNight from '../../assets/solar-banner-night.webp'
import './SolarIntake.css'

// ═══ (ك 11-25) زينا: مكالمة الطاقة الشمسية بلا تقني ═══
// «أي واحد يقعد يخابر الزبون، يسأله أسئلة، والنظام ينطيه تقريباً المنظومة من هلقد لهلقد
//  والسعر من هلقد لهلقد حسب البراند — مو سعر مباشر. وإذا وافق نطلعله كادر كشف.»
// التصميم طبق صورتها (نهاري وليلي). الأرقام: مدى الأسعار من المنظومات المسجّلة فعلاً
// (/api/solar-inquiries/price-ranges)، واستهلاك الأجهزة من جدول التقنيين تحت.

// ⚠️ جدول استهلاك الأجهزة (كيلو واط لكل جهاز: [أقل، أكثر]) — ينطوه التقنيين، ممنوع أرقام من عندنا.
// لحد ما يتعبّى، التقدير يطلع «بانتظار جدول الأجهزة» والكادر يحدد بالكشف.
const APPLIANCE_KW: Record<string, [number, number]> = {}

const APPLIANCES = ['ثلاجة', 'مجمدة', 'مبردة', 'مكيف', 'غسالة', 'سخان ماء', 'مضخة ماء', 'تلفزيون', 'إنارة']
const HOUSE_TYPES = ['بيت مستقل', 'شقة', 'محل']
const USAGE: Record<string, string> = { DAY_NIGHT: 'ليلي ونهاري', NIGHT: 'ليلي بس', DAY: 'نهاري بس' }
const TIER: Record<string, { t: string; icon: string; cls: string }> = {
  ECONOMY: { t: 'اقتصادي', icon: 'leaf', cls: 'si-eco' },
  MID: { t: 'وسط', icon: 'star', cls: 'si-mid' },
  HIGH: { t: 'غالي', icon: 'gem', cls: 'si-high' },
}
type Appliance = { type: string; count: string; condition: '' | 'OLD' | 'NEW' }
type Tier = { tier: string; min: number; max: number; count: number }
type Sec = 'customer' | 'house' | 'usage' | 'devices' | 'result' | 'visit'

const money = (n: number) => `${Math.round(n).toLocaleString('en-US')} د.ع`
const kw = (n: number) => (Number.isInteger(n) ? String(n) : n.toFixed(1))

export default function SolarIntake({ onSaved }: { onSaved: (msg: string) => void }) {
  // أضيق من 1000 بكسل: ترتيب عمودي
  const pageRef = useRef<HTMLDivElement>(null)
  useLayoutEffect(() => {
    const el = pageRef.current
    const host = el?.parentElement
    if (!el || !host) return
    const fit = () => el.classList.toggle('si-narrow', host.clientWidth < 1000)
    fit()
    const ro = new ResizeObserver(fit)
    ro.observe(host)
    return () => ro.disconnect()
  }, [])
  const [open, setOpen] = useState<Record<Sec, boolean>>({ customer: true, house: true, usage: false, devices: false, result: false, visit: false })
  const [f, setF] = useState({ name: '', phone: '', houseArea: '', roofArea: '', houseType: HOUSE_TYPES[0], region: '', obstacles: false, obstaclesNote: '', usage: '', visit: '' })
  const [devs, setDevs] = useState<Appliance[]>([])
  const [other, setOther] = useState('')
  const addOther = () => {
    const t = other.trim()
    if (!t) return
    setDevs((ds) => (ds.some((d) => d.type === t) ? ds : [...ds, { type: t, count: '1', condition: '' }]))
    setOther('')
  }
  const [tierState, setTierState] = useState<{ key: string; tiers: Tier[] }>({ key: '', tiers: [] })
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const [ok, setOk] = useState('')
  // أقل موعد للكشف = هسه (بتوقيت الجهاز) — ينحسب مرة وحدة
  const [minVisit] = useState(() => { const d = new Date(); d.setMinutes(d.getMinutes() - d.getTimezoneOffset()); return d.toISOString().slice(0, 16) })

  const toggle = (s: Sec) => setOpen((o) => ({ ...o, [s]: !o[s] }))
  const set = (k: keyof typeof f, v: string | boolean) => setF((x) => ({ ...x, [k]: v }))

  const appliances = devs.filter((d) => d.type && Number(d.count) > 0)
  const devSummary = appliances.map((d) => `${d.type} ×${d.count}${d.condition ? ` (${d.condition === 'OLD' ? 'قديمة' : 'جديدة'})` : ''}`).join('، ')

  // التقدير: من جدول التقنيين بس — جهاز ماله رقم = ما نقدّر
  const estimate = useMemo(() => {
    const list = devs.filter((d) => d.type && Number(d.count) > 0)
    if (list.length === 0 || list.some((d) => !APPLIANCE_KW[d.type])) return null
    let lo = 0, hi = 0
    for (const d of list) { lo += APPLIANCE_KW[d.type][0] * Number(d.count); hi += APPLIANCE_KW[d.type][1] * Number(d.count) }
    return { min: Math.max(1, Math.floor(lo)), max: Math.max(1, Math.ceil(hi)) }
  }, [devs])
  const estKey = estimate ? `${estimate.min}-${estimate.max}` : ''

  useEffect(() => {
    if (!estimate) return
    let alive = true
    const key = `${estimate.min}-${estimate.max}`
    apiRequest<{ tiers: Tier[] }>(`/solar-inquiries/price-ranges?minKw=${estimate.min}&maxKw=${estimate.max}`)
      .then((r) => { if (alive) setTierState({ key, tiers: r.tiers }) })
      .catch(() => { if (alive) setTierState({ key, tiers: [] }) })
    return () => { alive = false }
  }, [estimate])
  const tiers = tierState.key === estKey && estKey ? tierState.tiers : null

  const step1 = f.name.trim() !== '' && f.phone.replace(/\D/g, '').length >= 10 && f.houseArea !== '' && f.roofArea !== ''
  const done = [step1, f.usage !== '', appliances.length > 0, estimate !== null, f.visit !== '']
  const firstOpen = done.findIndex((d) => !d)
  const current = firstOpen === -1 ? 5 : firstOpen + 1
  const pct = Math.max(20, done.filter(Boolean).length * 20)
  const stepSec: Sec[] = ['house', 'usage', 'devices', 'result', 'visit']
  const sizeText = estimate ? `من ${kw(estimate.min)} لـ ${kw(estimate.max)} كيلو واط` : 'بانتظار جدول الأجهزة'
  const phrase = estimate
    ? `"المنظومة الي تناسبك من ${kw(estimate.min)} لـ ${kw(estimate.max)} كيلو واط، والسعر يختلف حسب الماركة والمواصفات."`
    : '"راح نطلعلك كادر يكشف على البيت ويتأكد من المعلومات ويحدد المنظومة الي تناسبك، والسعر يختلف حسب الماركة والمواصفات."'

  const payload = (visit: boolean) => ({
    customerName: f.name, phone: f.phone, area: f.region,
    note: devSummary ? `الأجهزة: ${devSummary}` : '',
    siteVisitAt: visit && f.visit ? new Date(f.visit).toISOString() : null,
    intake: {
      houseArea: f.houseArea ? Number(f.houseArea) : undefined,
      roofArea: f.roofArea ? Number(f.roofArea) : undefined,
      houseType: f.houseType, roofObstacles: f.obstacles, obstaclesNote: f.obstaclesNote,
      usage: f.usage || undefined,
      appliances: appliances.map((d) => ({ type: d.type, count: Number(d.count), condition: d.condition })),
    },
  })
  const save = async (visit: boolean) => {
    setErr('')
    if (!f.name.trim() || f.phone.replace(/\D/g, '').length < 10) { setErr('اكتب اسم الزبون ورقم هاتف صحيح'); setOpen((o) => ({ ...o, customer: true })); return }
    if (visit && !f.visit) { setErr('اختار موعد الكشف أول'); setOpen((o) => ({ ...o, visit: true })); return }
    setBusy(true)
    try {
      await apiRequest('/solar-inquiries', { method: 'POST', body: JSON.stringify(payload(visit)) })
      const t = visit ? 'انحجز الكشف بالموقع وانرسل للتقنيين ✅' : 'انسجل الاستفسار ✅'
      setOk(t); onSaved(t)
      setF({ name: '', phone: '', houseArea: '', roofArea: '', houseType: HOUSE_TYPES[0], region: '', obstacles: false, obstaclesNote: '', usage: '', visit: '' })
      setDevs([])
    } catch (e) { setErr((e as Error).message) } finally { setBusy(false) }
  }

  const print = () => {
    const w = window.open('', '_blank', 'width=800,height=900')
    if (!w) return
    const row = (k: string, v: string) => `<tr><th>${k}</th><td>${v.replace(/</g, '&lt;') || '—'}</td></tr>`
    w.document.write(`<html dir="rtl"><head><title>وصف طلب الزبون</title><style>body{font-family:Tajawal,Tahoma,sans-serif;padding:24px}h1{font-size:22px}table{width:100%;border-collapse:collapse}th,td{border:1px solid #999;padding:8px;text-align:right;font-size:14px}th{background:#eef2f7;width:35%}.n{margin-top:24px;border:1px solid #999;min-height:120px;padding:8px}</style></head><body>
      <h1>☀️ وصف طلب الزبون — كشف موقعي</h1><table>
      ${row('اسم الزبون', f.name)}${row('رقم الهاتف', f.phone)}${row('المنطقة', f.region)}${row('نوع البيت', f.houseType)}
      ${row('مساحة البيت تقريباً (م²)', f.houseArea)}${row('مساحة السطح الفارغة (م²)', f.roofArea)}
      ${row('عوائق بالسطح', f.obstacles ? `إي — ${f.obstaclesNote}` : 'ماكو')}${row('شلون يريدها', USAGE[f.usage] ?? '')}
      ${row('الأجهزة', devSummary)}${row('التقدير التقريبي', estimate ? sizeText : '')}
      ${row('موعد الكشف', f.visit ? new Date(f.visit).toLocaleString('en-GB', { timeZone: 'Asia/Baghdad' }) : 'ما محدد')}
      </table><div class="n"><b>ملاحظات الكادر بعد الكشف (شنو تطابق وشنو اختلف، والمنظومة النهائية):</b></div></body></html>`)
    w.document.close(); w.focus(); w.print()
  }

  const head = (s: Sec, icon: string, title: string, value?: string) => (
    <button type="button" className={`si-sec-head ${open[s] ? 'is-open' : ''}`} onClick={() => toggle(s)} aria-expanded={open[s]}>
      <LineIcon n={icon} className="si-sec-icon" />
      <span className="si-sec-title">{title}</span>
      {!open[s] && value !== undefined && <span className="si-sec-val">{value}</span>}
      <LineIcon n="down" className="si-chev" />
    </button>
  )

  return (
    <div ref={pageRef} className="si-page" dir="rtl">
      <header className="si-hero">
        <div className="si-hero-text">
          <h1><LineIcon n="phone" className="si-hero-icon" />استفسارات الطاقة الشمسية</h1>
          <p>سجّل معلومات الزبون وشوف المنظومة الي تناسبه</p>
        </div>
        <img className="si-banner si-banner-day" src={bannerDay} alt="" />
        <img className="si-banner si-banner-night" src={bannerNight} alt="" />
      </header>

      <div className="si-grid">
        <main className="si-main">
          <div className="si-card si-steps">
            <div className="si-progress">
              <span className="si-progress-label">التقدم</span>
              <div className="si-progress-bar"><b>{pct}%</b><span><i style={{ width: `${pct}%` }} /></span></div>
            </div>
            <ol>
              {['البيت', 'الاستعمال', 'الأجهزة', 'التقدير', 'الكشف'].map((t, i) => (
                <li key={t} className={i + 1 === current ? 'is-on' : done[i] ? 'is-done' : ''}>
                  <button type="button" onClick={() => setOpen((o) => ({ ...o, [stepSec[i]]: true, ...(i === 0 ? { customer: true } : {}) }))}>
                    <span className="si-dot">{i + 1}</span><span className="si-step-t">{t}</span>
                  </button>
                </li>
              ))}
            </ol>
          </div>

          <section className="si-sec">
            {head('customer', 'user', 'معلومات الزبون', f.name || '—')}
            {open.customer && (
              <div className="si-sec-body si-cols-2">
                <label><span>اسم الزبون <em>*</em></span>
                  <span className="si-in"><LineIcon n="user" /><input value={f.name} onChange={(e) => set('name', e.target.value)} placeholder="اكتب اسم الزبون" /></span>
                </label>
                <label><span>رقم الهاتف <em>*</em></span>
                  <span className="si-in"><LineIcon n="phone" /><input dir="ltr" value={f.phone} onChange={(e) => set('phone', e.target.value)} placeholder="07xx xxx xxxx" /></span>
                </label>
              </div>
            )}
          </section>

          <section className="si-sec">
            {head('house', 'house', 'البيت والسطح', f.houseArea ? `البيت ${f.houseArea} م² · السطح ${f.roofArea || '—'} م²` : '—')}
            {open.house && (
              <div className="si-sec-body">
                <div className="si-cols-4">
                  <label>مساحة البيت تقريباً (م²)<span className="si-in"><input inputMode="decimal" value={f.houseArea} onChange={(e) => set('houseArea', e.target.value.replace(/[^0-9.]/g, ''))} placeholder="اكتب المساحة" /></span></label>
                  <label>مساحة السطح الفارغة (م²)<span className="si-in"><input inputMode="decimal" value={f.roofArea} onChange={(e) => set('roofArea', e.target.value.replace(/[^0-9.]/g, ''))} placeholder="اكتب المساحة" /></span></label>
                  <label>نوع البيت<span className="si-in si-sel"><select value={f.houseType} onChange={(e) => set('houseType', e.target.value)}>{HOUSE_TYPES.map((h) => <option key={h}>{h}</option>)}</select><LineIcon n="down" /></span></label>
                  <label>المنطقة<span className="si-in"><input value={f.region} onChange={(e) => set('region', e.target.value)} placeholder="اكتب المنطقة" /><LineIcon n="pin" /></span></label>
                </div>
                <div className="si-obst">
                  <div><span className="si-lbl">اكو عوائق بالسطح؟</span>
                    <div className="si-yn">
                      <button type="button" className={!f.obstacles ? 'is-on' : ''} onClick={() => set('obstacles', false)}>{!f.obstacles && <LineIcon n="check" />}لا ماكو</button>
                      <button type="button" className={f.obstacles ? 'is-on' : ''} onClick={() => set('obstacles', true)}>{f.obstacles && <LineIcon n="check" />}إي اكو</button>
                    </div>
                  </div>
                  <label>شنو هي العوائق؟<span className="si-in"><input disabled={!f.obstacles} value={f.obstaclesNote} onChange={(e) => set('obstaclesNote', e.target.value)} placeholder="مثلاً: خزان ماي، ستلايت، مظلة…" /></span></label>
                </div>
              </div>
            )}
          </section>

          <section className="si-sec si-tone-blue">
            {head('usage', 'bolt', 'شلون يريدها', USAGE[f.usage] ?? 'اختار')}
            {open.usage && (
              <div className="si-sec-body si-yn">
                {Object.entries(USAGE).map(([k, t]) => (
                  <button key={k} type="button" className={f.usage === k ? 'is-on' : ''} onClick={() => set('usage', k)}>{f.usage === k && <LineIcon n="check" />}{t}</button>
                ))}
              </div>
            )}
          </section>

          <section className="si-sec si-tone-grey">
            {head('devices', 'layers', 'الأجهزة الي يشغّلها', devSummary || 'ماكو أجهزة بعد')}
            {open.devices && (
              <div className="si-sec-body si-devs">
                <span className="si-lbl">اختار كل الأجهزة الي يشغّلها</span>
                <div className="si-chips">
                  {APPLIANCES.map((a) => {
                    const on = devs.some((d) => d.type === a)
                    return (
                      <button key={a} type="button" className={on ? 'is-on' : ''} aria-pressed={on}
                        onClick={() => setDevs((ds) => (ds.some((d) => d.type === a) ? ds.filter((d) => d.type !== a) : [...ds, { type: a, count: '1', condition: '' }]))}>
                        {on && <LineIcon n="check" />}{a}
                      </button>
                    )
                  })}
                </div>
                <div className="si-other">
                  <span className="si-in"><input value={other} onChange={(e) => setOther(e.target.value)} onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); addOther() } }} placeholder="أخرى: اكتب اسم الجهاز إذا مو موجود" /></span>
                  <button type="button" className="si-add-btn" disabled={!other.trim()} onClick={addOther}><LineIcon n="plus" />ضيف</button>
                </div>
                {devs.map((d) => (
                  <div key={d.type} className="si-dev">
                    <b className="si-dev-name">{d.type}</b>
                    <span className="si-in"><input type="number" min="1" placeholder="العدد" aria-label={`عدد ${d.type}`} value={d.count} onChange={(e) => setDevs((ds) => ds.map((x) => (x.type === d.type ? { ...x, count: e.target.value } : x)))} /></span>
                    <div className="si-yn">
                      {(['OLD', 'NEW'] as const).map((c) => (
                        <button key={c} type="button" className={d.condition === c ? 'is-on' : ''} onClick={() => setDevs((ds) => ds.map((x) => (x.type === d.type ? { ...x, condition: c } : x)))}>{c === 'OLD' ? 'قديمة' : 'جديدة'}</button>
                      ))}
                    </div>
                    <button type="button" className="si-x" aria-label={`شيل ${d.type}`} onClick={() => setDevs((ds) => ds.filter((x) => x.type !== d.type))}><LineIcon n="trash" /></button>
                  </div>
                ))}
              </div>
            )}
          </section>

          <section className="si-sec si-tone-green">
            {head('result', 'gear', 'التقدير', sizeText)}
            {open.result && (
              <div className="si-sec-body">
                <p className="si-note">{estimate ? sizeText : 'التقدير يحتاج جدول استهلاك الأجهزة من التقنيين — لحد ما يوصل، الكادر يحدد المنظومة بالكشف.'}</p>
              </div>
            )}
          </section>

          <section className="si-sec si-tone-purple">
            {head('visit', 'cal', 'موعد الكشف', `${f.visit ? new Date(f.visit).toLocaleString('en-GB', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' }) : 'ما محدد'}`)}
            {open.visit && (
              <div className="si-sec-body si-cols-2">
                <label>وكت الكشف<span className="si-in"><input type="datetime-local" min={minVisit} value={f.visit} onChange={(e) => set('visit', e.target.value)} /></span></label>
              </div>
            )}
          </section>
          {err && <p className="si-err" role="alert">{err}</p>}
        </main>

        <aside className="si-card si-side">
          <div className="si-side-head"><LineIcon n="bars" className="si-side-hicon" />ملخص التقدير</div>
          <div className="si-side-body">
            <h3><LineIcon n="layers" />حجم المنظومة الي تناسبه</h3>
            <div className="si-size">{sizeText}</div>
            <h3><LineIcon n="tag" />الأسعار تقريباً (حسب الماركة)</h3>
            <div className="si-tiers">
              {(['ECONOMY', 'MID', 'HIGH'] as const).map((k) => {
                const t = tiers?.find((x) => x.tier === k)
                return (
                  <div key={k} className="si-tier">
                    <span className={`si-pill ${TIER[k].cls}`}><LineIcon n={TIER[k].icon} />{TIER[k].t}</span>
                    <span className="si-range">{t ? `من ${money(t.min)} لـ ${money(t.max)}` : '—'}</span>
                  </div>
                )
              })}
            </div>
            <div className="si-phrase-box">
              <h3><LineIcon n="chat" />شنو تگول للزبون</h3>
              <p className="si-phrase">{phrase}</p>
            </div>
            <button type="button" className="si-btn si-btn-green" disabled={busy} onClick={() => void save(true)}><LineIcon n="cal" />احجز كشف بالموقع</button>
            <button type="button" className="si-btn si-btn-yellow" onClick={print}><LineIcon n="printer" />اطبع وصف الطلب</button>
            <button type="button" className="si-btn si-btn-out" disabled={busy} onClick={() => void save(false)}><LineIcon n="chat" />بس يستفسر</button>
            {err && <p className="si-err" role="alert">{err}</p>}
            {ok && <p className="si-ok" role="status">{ok}</p>}
          </div>
        </aside>
      </div>
    </div>
  )
}
