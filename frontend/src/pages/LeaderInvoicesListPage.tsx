import { useEffect, useState } from 'react'
import { onEnter } from '../utils/enterKey'
import { api, AUDIT_VERDICTS, type LeaderInvoice, type LeaderInvoiceAdjustment } from '../api'
import EntityIdentity from '../components/EntityIdentity'
import StatTile from '../components/StatTile'
import Pager from '../components/Pager'
import { formatCustomerCode } from '../utils/identity'
import { useSession } from '../session'
import { matches } from '../utils/search'
import { esc, printIdentityCss, printIdentityHtml } from '../utils/printIdentity'
import LocateHint from '../components/LocateHint'

// ═══ طباعة فاتورة الليدر ═══
//
// هاي أكثر ورقة تطلع للزبون، وكانت **ماكو بيها طباعة أصلاً** — المحاسب
// يصوّر الشاشة أو ينسخ الأرقام بالإيد.
//
// الرأس نفسه رأس بقية المطبوعات (`printIdentityHtml`) حتى الورقة تعرّف
// نفسها: كود الحجز وكود الزبون واسم الليدر المسؤول.
//
// ⚠️ كل نص زبون أو مادة يمرّ بـ`esc()`. اسم بيه `<` يكسر الصفحة، واسم
// بيه وسم فيه onerror ينفّذ بجلسة الي يفتح المعاينة.
//
// ⚠️ سجل التعديلات ينطبع هو هم: الورقة الي تعرض الرقم الجديد بلا ما
// تگول إنه انتعدّل تخلي القارئ يظن إنه الأصلي.
function printInvoice(inv: LeaderInvoice, adjustments: LeaderInvoiceAdjustment[]) {
  const money = (n: number) => `${Number(n || 0).toLocaleString('en-US')} د.ع`
  const identity = printIdentityHtml({
    bookingCode: inv.bookingCode || undefined,
    customerCode: formatCustomerCode(inv.booking?.customer) || undefined,
    customerName: inv.customerName || inv.booking?.customer?.name || undefined,
    customerPhone: inv.customerPhone || inv.booking?.customer?.phone || undefined,
    address: inv.customerAddress || inv.booking?.address || undefined,
    serviceName: inv.booking?.service?.name || undefined,
    leaderName: inv.employeeName || undefined,
  })

  const itemRows = inv.items
    .map(
      (it) => `<tr><td>${esc(it.itemName)} <small>${esc(it.systemName)}</small></td>
        <td>${esc(it.count)}</td>
        <td><small>${esc(
          [it.heightMeters ? `ارتفاع ${it.heightMeters}م` : '', it.cableLengthMeters ? `كيبل ${it.cableLengthMeters}م` : '']
            .filter(Boolean)
            .join(' · '),
        )}</small></td></tr>`,
    )
    .join('')

  const materialRows = inv.materials
    .map(
      (m) => `<tr><td>${esc(m.name)}</td><td>${esc(m.quantity)}</td><td>${esc(money(m.lineTotal))}</td></tr>`,
    )
    .join('')

  const adjRows = adjustments
    .map((a) => {
      const changed = (
        [
          ['تكاليف التنفيذ', a.oldExecutionCost, a.newExecutionCost],
          ['مجموع المواد', a.oldMaterialsTotal, a.newMaterialsTotal],
          ['الخصم', a.oldDiscountValue, a.newDiscountValue],
          ['المجموع الصافي', a.oldNetTotal, a.newNetTotal],
        ] as [string, number, number][]
      ).filter(([, o, n]) => o !== n)
      if (changed.length === 0) return ''
      const lines = changed
        .map(([label, o, n]) => `${esc(label)}: <s>${esc(money(o))}</s> ← <b>${esc(money(n))}</b>`)
        .join('<br>')
      return `<div class="adj">${lines}<br>السبب: ${esc(a.reason)}
        — ${esc(a.adjustedByName || 'غير معروف')}، ${esc(new Date(a.createdAt).toLocaleString('ar-IQ'))}</div>`
    })
    .join('')

  const html = `<!doctype html><html dir="rtl" lang="ar"><head><meta charset="utf-8">
<title>فاتورة ${esc(inv.accountingCode)}</title><style>
  body{font-family:'Segoe UI',Tahoma,sans-serif;padding:22px;color:#0f2040}
  h1{font-size:19px;margin:0} .code{font-family:monospace;color:#1d4ed8;font-size:13px}
  table{width:100%;border-collapse:collapse;margin-top:6px;font-size:12.5px}
  th,td{border-bottom:1px solid var(--bd-line);padding:5px;text-align:right}
  th{color:#64748b;font-weight:600} small{color:#94a3b8}
  h2{font-size:13px;margin:16px 0 2px;color:#334155}
  .tot{margin-top:14px;border-top:2px solid #0f2040;padding-top:8px;font-size:13.5px}
  .tot div{display:flex;justify-content:space-between;padding:2px 0}
  .tot .net{font-size:16px;font-weight:800;border-top:1px solid #cbd5e1;margin-top:5px;padding-top:5px}
  .adj{border:1px solid #fcd34d;background:#fffbeb;border-radius:7px;padding:7px 10px;
       font-size:11.5px;margin-top:5px}
  ${printIdentityCss}
  @media print{body{padding:0}}
</style></head><body>
  <h1>فاتورة ليدر</h1>
  <p class="code">${esc(inv.accountingCode)}${inv.externalInvoiceNumber ? ` · محاسبياً: ${esc(inv.externalInvoiceNumber)}` : ''}
    · ${esc(new Date(inv.createdAt).toLocaleString('ar-IQ'))}</p>
  ${identity}
  <h2>بنود التنفيذ — ${esc(inv.systems.join('، ') || '—')} · ${esc(inv.totalDeviceCount)} جهاز</h2>
  <table><thead><tr><th>البند</th><th>العدد</th><th>تفاصيل</th></tr></thead>
    <tbody>${itemRows || '<tr><td colspan="3"><small>ماكو بنود</small></td></tr>'}</tbody></table>
  ${materialRows ? `<h2>المواد</h2><table><thead><tr><th>المادة</th><th>الكمية</th><th>المجموع</th></tr></thead><tbody>${materialRows}</tbody></table>` : ''}
  <div class="tot">
    <div><span>تكاليف التنفيذ</span><b>${esc(money(inv.executionCost))}</b></div>
    <div><span>مجموع المواد</span><b>${esc(money(inv.materialsTotal))}</b></div>
    <div><span>الخصم</span><b>${esc(money(inv.discountValue))}</b></div>
    <div class="net"><span>المجموع الصافي</span><span>${esc(money(inv.netTotal))}</span></div>
  </div>
  ${adjRows ? `<h2>تعديلات المحاسب</h2>${adjRows}` : ''}
</body></html>`

  const w = window.open('', '_blank')
  // مانع النوافذ ممكن يرجّع null — بلا هذا الفحص الطباعة تطيح بصمت
  if (!w) {
    alert('المتصفح منع فتح نافذة الطباعة — سمح النوافذ المنبثقة لهذا الموقع')
    return
  }
  w.document.write(html)
  w.document.close()
  w.focus()
  w.print()
}

// قائمة بسيطة لعرض فواتير الليدر السابقة (كل الفواتير أو حسب الموظف).
// الفاتورة تضل SUBMITTED (ظاهرة عند الليدر) لين مدير/محاسب يعتمدها لـAPPROVED —
// الليدر ما يقدر يعتمد فاتورته بنفسه (زر "اعتماد" ما يظهر إلا لـADMIN/FINANCE).
export default function LeaderInvoicesListPage() {
  const { employee } = useSession()
  const canApprove = employee?.role === 'ADMIN' || employee?.role === 'FINANCE'
  /** ⚠️ المراقب يبتّ بطلب المحاسب — والمحاسب ما يبتّ بطلبه بنفسه،
   *  وإلا الإرسال للمراقب يصير شكلياً. (والخادم يمنعه فعلياً.) */
  const isMonitor = employee?.role === 'ADMIN' || employee?.role === 'MONITOR'
  /** ⚠️ `actualRole` مو `role`: المالك ينوصل كـADMIN بالواجهة،
   *  ودوره الحقيقي بـ`actualRole`. */
  const isOwner = employee?.actualRole === 'OWNER'
  const [invoices, setInvoices] = useState<LeaderInvoice[]>([])
  const [loading, setLoading] = useState(true)
  const [busyId, setBusyId] = useState<string | null>(null)
  // ═══ رقم الفاتورة المحاسبية ═══
  // المحاسب يصدّر فواتيره بنظام ثاني برّا نظامنا. بدون ربط الرقم وقت
  // الاعتماد ينقطع الخيط: عدنا فاتورة معتمدة وعنده فاتورة صادرة وماكو
  // شي يربطهن — فأي مراجعة لاحقة تصير يدوية. الرقم إجباري، ومؤرشف
  // حتى يلكاها بيه لمن يحتاجها.
  const [approveFor, setApproveFor] = useState<LeaderInvoice | null>(null)
  // ═══ التدقيق ═══ الحكم يتحط **قبل** الاعتماد، والسحب لما ينصار بالغلط
  const [auditFor, setAuditFor] = useState<LeaderInvoice | null>(null)
  const [auditVerdict, setAuditVerdict] = useState<'MATCHED' | 'MISMATCH' | 'PRICE_ERROR'>('MATCHED')
  const [auditNote, setAuditNote] = useState('')
  const [auditAmount, setAuditAmount] = useState('')
  const [auditErr, setAuditErr] = useState<string | null>(null)
  const [invoiceNo, setInvoiceNo] = useState('')
  const [approveErr, setApproveErr] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  // ⚠️ أي مرشّح يتغيّر لازم يرجّع الصفحة لواحد — وإلا يبقى المستخدم
  // بصفحة ٧ ونتيجة الترشيح ثلاث فواتير، فيشوف فراغاً ويظن ماكو نتائج.
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState(10)
  const [fLeader, setFLeader] = useState('')
  const [fCustomer, setFCustomer] = useState('')
  const [fSystem, setFSystem] = useState('')
  const [fMonth, setFMonth] = useState('')
  const resetPage = () => setPage(1)
  // ═══ ترتيب شاشة المحاسب ═══
  // الشاشة جانت جدول واحد طويل فيه كل شي مخلوط: المعتمد وغير المعتمد،
  // والمربوط برقم وغير المربوط. المحاسب يدوّر بعينه. هسه تبويبات
  // تحصر شغله بالي محتاجه، وبطاقات فوق تكله بالأرقام وين الشغل.
  // ⚠️⚠️ **مرحلتان مو وحدة.** قبل هيچ چان تبويب واحد «بانتظار
  // الاعتماد» = كل فاتورة مو معتمدة — يعني الفاتورة **قبل** التدقيق
  // و**بعده** بنفس المكان بالضبط، والضغط على أي حكم ما ينقلها ولا
  // ملّيمتر. والمحاسب ما يگدر يفرّق بين شغل باقي عليه تدقيق وشغل
  // باقي عليه قرار.
  const [tab, setTab] = useState<'AUDIT' | 'PENDING' | 'MONITOR' | 'NO_NUMBER' | 'APPROVED' | 'ALL'>('AUDIT')
  // إرسال للمراقب · بتّ المراقب · إرجاع المالك
  const [monitorFor, setMonitorFor] = useState<LeaderInvoice | null>(null)
  const [monitorNote, setMonitorNote] = useState('')
  const [monitorErr, setMonitorErr] = useState<string | null>(null)
  const [decideFor, setDecideFor] = useState<LeaderInvoice | null>(null)
  const [decideVerdict, setDecideVerdict] = useState<'OK' | 'FLAGGED'>('OK')
  const [decideNote, setDecideNote] = useState('')
  const [decideErr, setDecideErr] = useState<string | null>(null)
  // ربط رقم لفاتورة معتمدة قبل ما يصير الرقم إجبارياً
  const [linkFor, setLinkFor] = useState<LeaderInvoice | null>(null)
  const [linkNo, setLinkNo] = useState('')
  const [linkErr, setLinkErr] = useState<string | null>(null)
  // تعديل المحاسب على المبالغ
  const [adjustFor, setAdjustFor] = useState<LeaderInvoice | null>(null)
  const [adj, setAdj] = useState({ executionCost: '', materialsTotal: '', discountValue: '', reason: '' })
  const [adjErr, setAdjErr] = useState<string | null>(null)

  const linkNumber = async () => {
    if (!linkFor || !linkNo.trim()) return
    setBusyId(linkFor.id); setLinkErr(null)
    try {
      await api.setInvoiceExternalNumber(linkFor.id, linkNo.trim())
      setLinkFor(null); setLinkNo(''); load()
    } catch (e) {
      setLinkErr(e instanceof Error ? e.message : 'تعذر ربط الرقم')
    } finally { setBusyId(null) }
  }

  const openAdjust = (inv: LeaderInvoice) => {
    setAdjustFor(inv)
    setAdjErr(null)
    setAdj({
      executionCost: String(inv.executionCost),
      materialsTotal: String(inv.materialsTotal),
      discountValue: String(inv.discountValue),
      reason: '',
    })
  }
  const adjNet = Math.max(0, (Number(adj.executionCost) || 0) + (Number(adj.materialsTotal) || 0) - (Number(adj.discountValue) || 0))
  const saveAdjust = async () => {
    if (!adjustFor || !adj.reason.trim()) return
    setBusyId(adjustFor.id); setAdjErr(null)
    try {
      await api.adjustLeaderInvoice(adjustFor.id, {
        executionCost: Number(adj.executionCost) || 0,
        materialsTotal: Number(adj.materialsTotal) || 0,
        discountValue: Number(adj.discountValue) || 0,
        reason: adj.reason.trim(),
      })
      setAdjustFor(null); load()
    } catch (e) {
      setAdjErr(e instanceof Error ? e.message : 'تعذر حفظ التعديل')
    } finally { setBusyId(null) }
  }

  // الفلترة: تبويب + بحث
  const matchesSearch = (inv: LeaderInvoice) => {
    return matches([inv.externalInvoiceNumber, inv.accountingCode, inv.customerName, inv.employeeName], search)
  }
  /** ⚠️ نفس اشتقاق الخادم بالضبط — المرحلة تنحسب من الحالة والحكم
   *  سوا، مو من عمود ثالث ينحرف. */
  const audited = (inv: LeaderInvoice) => !!inv.auditVerdict && inv.auditVerdict.trim() !== ''
  /** ⚠️ نفس اشتقاق الخادم: طُلبت ولا انبتّ بيها = عند المراقب الآن. */
  const atMonitor = (inv: LeaderInvoice) => !!inv.monitorRequestedAt && !inv.monitorDecidedAt
  const inTab = (inv: LeaderInvoice) => {
    if (tab === 'ALL') return true
    if (tab === 'AUDIT') return inv.status !== 'APPROVED' && !audited(inv)
    if (tab === 'MONITOR') return inv.status !== 'APPROVED' && atMonitor(inv)
    if (tab === 'PENDING') return inv.status !== 'APPROVED' && audited(inv) && !atMonitor(inv)
    if (tab === 'APPROVED') return inv.status === 'APPROVED'
    return inv.status === 'APPROVED' && !inv.externalInvoiceNumber
  }
  // ⚠️ المرشّحات تتراكم مع البحث مو تستبدله.
  const matchesFilters = (i: LeaderInvoice) =>
    (!fLeader || i.employeeName === fLeader) &&
    (!fCustomer || i.customerName === fCustomer) &&
    (!fSystem || (i.systems || []).includes(fSystem)) &&
    (!fMonth || i.createdAt.slice(0, 7) === fMonth)
  const shown = invoices.filter((i) => inTab(i) && matchesSearch(i) && matchesFilters(i))
  const counts = {
    AUDIT: invoices.filter((i) => i.status !== 'APPROVED' && !audited(i)).length,
    MONITOR: invoices.filter((i) => i.status !== 'APPROVED' && atMonitor(i)).length,
    PENDING: invoices.filter((i) => i.status !== 'APPROVED' && audited(i) && !atMonitor(i)).length,
    NO_NUMBER: invoices.filter((i) => i.status === 'APPROVED' && !i.externalInvoiceNumber).length,
    APPROVED: invoices.filter((i) => i.status === 'APPROVED').length,
    ALL: invoices.length,
  }
  const sumShown = shown.reduce((t, i) => t + i.netTotal, 0)
  const pageStart = (page - 1) * perPage
  const paged = shown.slice(pageStart, pageStart + perPage)

  // خيارات القوائم تنشتق من المحمّل — بلا مسار ولا معامل خادم.
  const leaderOptions = Array.from(new Set(invoices.map((i) => i.employeeName).filter(Boolean) as string[])).sort()
  const customerOptions = Array.from(new Set(invoices.map((i) => i.customerName).filter(Boolean) as string[])).sort()
  const systemOptions = Array.from(new Set(invoices.flatMap((i) => i.systems || []))).sort()
  /** المحاسب يرسلها للمراقب — الطريق الثالث للشك. */
  const sendToMonitor = async () => {
    if (!monitorFor) return
    setBusyId(monitorFor.id); setMonitorErr(null)
    try {
      await api.requestInvoiceMonitorReview(monitorFor.id, monitorNote.trim())
      setMonitorFor(null); setMonitorNote(''); load()
    } catch (e) { setMonitorErr(e instanceof Error ? e.message : 'تعذر الإرسال للمراقب') }
    finally { setBusyId(null) }
  }

  /** المراقب يبتّ — وترجع للمحاسب بحكمها. */
  const decideMonitor = async () => {
    if (!decideFor) return
    setBusyId(decideFor.id); setDecideErr(null)
    try {
      await api.decideInvoiceMonitorReview(decideFor.id, decideVerdict, decideNote.trim())
      setDecideFor(null); setDecideNote(''); setDecideVerdict('OK'); load()
    } catch (e) { setDecideErr(e instanceof Error ? e.message : 'تعذر حفظ الحكم') }
    finally { setBusyId(null) }
  }

  /** ⚠️ المالك وحده — والسبب إجباري بالخادم. */
  const returnToAccountant = async (inv: LeaderInvoice) => {
    const reason = window.prompt('سبب الإرجاع للمحاسب (المحاسب راح يقراه):')
    if (reason === null) return
    setBusyId(inv.id)
    try { await api.returnInvoiceToAccountant(inv.id, reason.trim()); load() }
    catch (e) { window.alert(e instanceof Error ? e.message : 'تعذر الإرجاع') }
    finally { setBusyId(null) }
  }

  const saveAudit = async () => {
    if (!auditFor) return
    setAuditErr(null)
    setBusyId(auditFor.id)
    try {
      await api.setInvoiceAuditVerdict(auditFor.id, {
        verdict: auditVerdict,
        note: auditNote.trim(),
        // ⚠️ Number('') = صفر مو فاضي — بلا الفحص نرسل صفراً كأنه
        // «ما دخل ولا دينار» وهي معلومة غلط تماماً.
        auditedAmount: auditAmount.trim() ? Number(auditAmount) : null,
      })
      setAuditFor(null); setAuditNote(''); setAuditAmount('')
      load()
    } catch (e) {
      setAuditErr(e instanceof Error ? e.message : 'تعذر حفظ الحكم')
    } finally { setBusyId(null) }
  }

  const revoke = async (inv: LeaderInvoice) => {
    const reason = prompt(`ليش تسحب اعتماد الفاتورة ${inv.accountingCode}؟\n(السبب إجباري — المراقب والمدير راح يقرونه)`)
    if (!reason) return
    setBusyId(inv.id)
    try {
      await api.revokeInvoiceApproval(inv.id, reason)
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر سحب الاعتماد')
    } finally { setBusyId(null) }
  }

  const [details, setDetails] = useState<LeaderInvoice | null>(null)
  // سجل تعديلات المحاسب — «شنو كان وشنو صار ومنو غيّره»
  const [adjustments, setAdjustments] = useState<LeaderInvoiceAdjustment[]>([])

  const load = () => { api.getLeaderInvoices().then(setInvoices).finally(() => setLoading(false)) }
  useEffect(load, [])

  const handleApprove = async () => {
    if (!approveFor || !invoiceNo.trim()) return
    setBusyId(approveFor.id)
    setApproveErr(null)
    try {
      await api.approveLeaderInvoice(approveFor.id, invoiceNo.trim())
      setApproveFor(null)
      setInvoiceNo('')
      load()
    } catch (e) {
      setApproveErr(e instanceof Error ? e.message : 'تعذر اعتماد الفاتورة')
    } finally {
      setBusyId(null)
    }
  }

  return (
    <div>
      {/* الرأس بنفس عائلة شاشات المحاسب (التدقيق اليومي/البلاغات)، ويگول
          صراحةً إنه هذا **طابور الاعتماد** — سؤال صاحب العمل «وين ألگه
          الفواتير الي بانتظار الاعتماد؟» جوابه هاي الشاشة، بس الاسم
          «فواتير الليدر» ما جان يدل عليها. */}
      <div
        className="relative overflow-hidden rounded-2xl p-6 shadow-md"
        style={{ background: 'linear-gradient(135deg, #1a3a5c 0%, #24507e 55%, #2f6ba8 100%)' }}
      >
        <span
          aria-hidden
          className="pointer-events-none absolute -left-16 -top-24 h-64 w-64 rounded-full opacity-20"
          style={{ background: 'radial-gradient(circle, #c8a45a 0%, transparent 70%)' }}
        />
        <div className="relative flex flex-wrap items-end justify-between gap-4">
          <div>
            <h2 className="text-2xl font-black text-white">🧾 فواتير الليدر</h2>
            {/* ⚠️ النص چان يقول «الفاتورة الي تأشّرت مطابق توصل هنا»
                وهذا **ما چان منفَّذ** — الشاشة تقول شي والكود يسوي
                شي ثاني. صار يوصف المسار الحقيقي بالضبط. */}
            <p className="mt-1 max-w-2xl text-sm text-blue-100">
              الفاتورة تجي أول لـ<b className="text-white">بانتظار التدقيق</b> — تأشّر عليها
              مطابق أو غير مطابق أو خطأ بالسعر، فتنتقل لـ<b className="text-white">بانتظار الاعتماد</b>.
              وهناك تقرر: تعتمدها بـ<b className="text-white">رقم الفاتورة</b> من نظامك الثاني، أو تتركها.
            </p>
          </div>
        </div>
      </div>

      {/* ⚠️ البطاقات الخمس تظهر **كلها دائماً**.
          قبلها چانت ثلاث، كل وحدة تنخفي لمن يصير عدّها صفراً،
          و«عند المراقب» محصورة بالمالك. وبطاقة تختفي لمن تصفّر
          تخلّي المستخدم يظن الشاشة تغيّرت — وما يعرف إذا صفر لو
          الميزة اختفت. و«صفر عند المراقب» معلومة مثل غيرها.
          وكل بطاقة **مرشّح**: الرقم يودّي لشغله. */}
      <div className="mt-4 grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
        {([
          { k: 'ALL' as const, label: 'الكل', icon: '📚', tone: 'default' as const, c: counts.ALL },
          { k: 'APPROVED' as const, label: 'معتمدة', icon: '✔', tone: 'success' as const, c: counts.APPROVED },
          { k: 'MONITOR' as const, label: 'عند المراقب', icon: '👁️', tone: 'violet' as const, c: counts.MONITOR },
          { k: 'PENDING' as const, label: 'بانتظار الاعتماد', icon: '⏳', tone: 'warning' as const, c: counts.PENDING },
          { k: 'AUDIT' as const, label: 'بانتظار التدقيق', icon: '🔍', tone: 'info' as const, c: counts.AUDIT },
        ]).map((t) => (
          <StatTile key={t.k} label={t.label} value={t.c} icon={t.icon} tone={t.tone}
            hint={tab === t.k ? 'معروضة الآن' : 'اضغط للعرض'}
            onClick={() => { setTab(t.k); resetPage() }} />
        ))}
      </div>

      {/* التبويبات: كل تبويب شغلة وحدة يشتغلها المحاسب */}
      <div className="mt-4 flex flex-wrap gap-2">
        {([
          { k: 'AUDIT' as const, t: '🔍 بانتظار التدقيق', c: counts.AUDIT },
          { k: 'PENDING' as const, t: '⏳ بانتظار الاعتماد', c: counts.PENDING },
          { k: 'MONITOR' as const, t: '👁️ عند المراقب', c: counts.MONITOR },
          { k: 'NO_NUMBER' as const, t: '🔗 معتمدة بلا رقم فاتورة', c: counts.NO_NUMBER },
          { k: 'APPROVED' as const, t: '✔ معتمدة', c: counts.APPROVED },
          { k: 'ALL' as const, t: 'الكل', c: counts.ALL },
        ]).map((o) => (
          <button
            key={o.k}
            onClick={() => { setTab(o.k); resetPage() }}
            className={`rounded-xl border px-4 py-2 text-sm font-bold transition-all ${
              tab === o.k
                ? 'border-brand-500 bg-brand-50 text-brand-800'
                : 'border-slate-200 bg-white text-slate-500 hover:border-slate-300'
            }`}
          >
            {o.t} <span className="text-xs">({o.c})</span>
          </button>
        ))}
      </div>

      {/* المرشّحات الأربعة — خياراتها من المحمّل، بلا مسار خادم.
          ⚠️ وكلها ترجّع الصفحة لواحد. */}
      <div className="mt-3 flex flex-wrap gap-2">
        {([
          { v: fCustomer, set: setFCustomer, all: 'كل الزبائن', opts: customerOptions },
          { v: fLeader, set: setFLeader, all: 'كل الليدرز', opts: leaderOptions },
          { v: fSystem, set: setFSystem, all: 'كل المنظومات', opts: systemOptions },
        ]).map((f) => (
          <select key={f.all} value={f.v}
            onChange={(e) => { f.set(e.target.value); resetPage() }}
            className="rounded-xl border px-3 py-2 text-sm"
            style={{ backgroundColor: 'var(--sf-card)', borderColor: 'var(--bd-line)', color: 'var(--t-body)' }}>
            <option value="">{f.all}</option>
            {f.opts.map((o) => <option key={o} value={o}>{o}</option>)}
          </select>
        ))}
        <input type="month" value={fMonth}
          onChange={(e) => { setFMonth(e.target.value); resetPage() }}
          className="rounded-xl border px-3 py-2 text-sm"
          style={{ backgroundColor: 'var(--sf-card)', borderColor: 'var(--bd-line)', color: 'var(--t-body)' }} />
        {(fCustomer || fLeader || fSystem || fMonth) && (
          <button onClick={() => { setFCustomer(''); setFLeader(''); setFSystem(''); setFMonth(''); resetPage() }}
            className="rounded-xl border px-4 py-2 text-sm font-bold"
            style={{ borderColor: 'var(--bd-line)', color: 'var(--t-body)' }}>
            امسح التصفية ✕
          </button>
        )}
      </div>

      <div className="mt-3 rounded-xl border border-white bg-white px-4 py-3 text-sm shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        العرض: <b className="text-brand-800">{shown.length}</b> فاتورة
        {/* «من إجمالي» چان ناقصاً — بدونه ما تعرف إذا الرقم كل شي
            لو نتيجة ترشيح. */}
        <span className="text-slate-400"> من إجمالي {counts.ALL} فاتورة</span> ·
        مجموعها: <b className="text-brand-800">{sumShown.toLocaleString()} د.ع</b>
        {counts.NO_NUMBER > 0 && tab !== 'NO_NUMBER' && (
          <span className="mr-3 text-amber-700">
            ⚠ أكو {counts.NO_NUMBER} فاتورة معتمدة بلا رقم فاتورة محاسبية
          </span>
        )}
      </div>

      <LocateHint query={search} localCount={shown.length} currentRoute="/leader-invoices" />

      {/* البحث برقم فاتورة المحاسب — هذا سبب أرشفة الرقم: يلكاها بيه */}
      <input
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        placeholder="🔍 بحث برقم الفاتورة المحاسبية، كود المحاسبة، الزبون، أو الليدر..."
        className="mt-4 w-full rounded-xl border border-slate-300 px-4 py-2.5 text-sm outline-none focus:border-brand-500"
      />

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}

      {!loading && (
        <div className="mt-6 overflow-x-auto rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <table className="w-full min-w-[820px] text-sm">
            <thead>
              <tr className="border-b border-slate-100 text-right text-slate-400">
                <th className="px-4 py-3">كود المحاسبة</th>
                <th className="px-4 py-3">رقم الفاتورة المحاسبية</th>
                <th className="px-4 py-3">الليدر</th>
                <th className="px-4 py-3">الزبون</th>
                <th className="px-4 py-3">المنظومات</th>
                <th className="px-4 py-3">تكاليف التنفيذ</th>
                <th className="px-4 py-3">مجموع المواد</th>
                <th className="px-4 py-3">المجموع الصافي</th>
                <th className="px-4 py-3">الحالة</th>
                <th className="px-4 py-3">التاريخ</th>
                <th className="px-4 py-3">إجراءات</th>
              </tr>
            </thead>
            <tbody>
              {paged.map((inv) => (
                <tr key={inv.id} className="border-b border-slate-50 last:border-0">
                  <td className="px-4 py-3 font-mono text-brand-700">{inv.accountingCode}</td>
                  <td className="px-4 py-3">
                    {inv.externalInvoiceNumber
                      ? <span className="rounded-lg bg-emerald-50 px-2 py-1 font-mono text-xs font-bold text-emerald-800">{inv.externalInvoiceNumber}</span>
                      : <span className="text-xs text-slate-400">—</span>}
                  </td>
                  <td className="px-4 py-3">
                    <span className="font-bold text-slate-700">{inv.employeeName || '—'}</span>
                    {inv.bookingCode && <div className="text-xs text-slate-400">حجز {inv.bookingCode}</div>}
                  </td>
                  <td className="px-4 py-3">
                    {inv.customerName || inv.booking?.customer?.name || '—'}
                    {formatCustomerCode(inv.booking?.customer) && (
                      <div className="font-mono text-xs text-slate-400">{formatCustomerCode(inv.booking?.customer)}</div>
                    )}
                  </td>
                  <td className="px-4 py-3">{inv.systems.join('، ')}</td>
                  <td className="px-4 py-3">{inv.executionCost.toLocaleString()}</td>
                  <td className="px-4 py-3">{inv.materialsTotal.toLocaleString()}</td>
                  <td className="px-4 py-3 font-bold text-brand-800">{inv.netTotal.toLocaleString()}</td>
                  <td className="px-4 py-3">
                    <span className={`rounded-full px-3 py-1 text-xs font-bold ${
                      inv.status === 'APPROVED' ? 'bg-emerald-50 text-emerald-700' : 'bg-amber-50 text-amber-700'
                    }`}>
                      {inv.status === 'APPROVED' ? '✔ معتمدة'
                        : atMonitor(inv) ? '👁️ عند المراقب'
                          : audited(inv) ? 'بانتظار الاعتماد' : '🔍 بانتظار التدقيق'}
                    </span>
                    {/* حكم التدقيق — «يطلعن وينن بفواتير الليدر» */}
                    {inv.auditVerdict && (
                      <span className={`mt-1 block rounded-full px-2 py-0.5 text-center text-[10.5px] font-bold ${
                        AUDIT_VERDICTS.find((v) => v.key === inv.auditVerdict)?.cls || ''
                      }`}>
                        {AUDIT_VERDICTS.find((v) => v.key === inv.auditVerdict)?.label}
                      </span>
                    )}
                    {inv.externalInvoiceNumber && (
                      <span className="mt-1 block text-center font-mono text-[10.5px] text-slate-500">
                        {inv.externalInvoiceNumber}
                      </span>
                    )}
                    {/* ═══ حكم المراقب ═══
                        ⚠️ يظهر **بعد ما ترجع** للمحاسب: هذا الي طلب
                        الرأي لأجله، وإخفاؤه يخلّي الإرسال للمراقب بلا
                        فايدة — يرسلها ويرجعها وما يعرف شنو قال. */}
                    {inv.monitorDecidedAt && (
                      <span className={`mt-1 block rounded-lg px-2 py-0.5 text-center text-[10.5px] font-bold ${
                        inv.monitorVerdict === 'FLAGGED' ? 'bg-amber-100 text-amber-800' : 'bg-indigo-100 text-indigo-800'}`}>
                        👁️ المراقب: {inv.monitorVerdict === 'FLAGGED' ? 'ملاحظة' : 'سليمة'}
                        {inv.monitorNote ? ` — ${inv.monitorNote}` : ''}
                      </span>
                    )}
                    {/* ⚠️ رجّعها المالك: المحاسب لازم يعرف **ليش** رجعت
                        — وإلا يعيد نفس الشغل بنفس الطريقة وترجع مرة ثانية. */}
                    {(inv.returnedCount ?? 0) > 0 && (
                      <span className="mt-1 block rounded-lg bg-amber-50 px-2 py-0.5 text-center text-[10.5px] font-bold text-amber-800">
                        ↩️ رجّعها المالك {inv.returnedCount} مرة{inv.returnReason ? ` — ${inv.returnReason}` : ''}
                      </span>
                    )}
                    {/* انسحب اعتمادها قبل — إشارة تحتاج انتباه */}
                    {(inv.revokedCount ?? 0) > 0 && (
                      <span className="mt-1 block text-center text-[10.5px] font-bold text-red-600">
                        ↩ انسحب اعتمادها {inv.revokedCount} مرة
                      </span>
                    )}
                  </td>
                  <td className="px-4 py-3 text-slate-400">
                    {new Date(inv.createdAt).toLocaleDateString('ar-IQ')}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex gap-2">
                      <button
                        onClick={() => {
                          setDetails(inv)
                          // نجيب السجل عند الفتح بس — ما نثقّل القائمة
                          setAdjustments([])
                          api.getInvoiceAdjustments(inv.id).then(setAdjustments).catch(() => setAdjustments([]))
                        }}
                        className="rounded-lg bg-slate-100 px-3 py-1.5 text-xs font-bold text-slate-700 hover:bg-slate-200"
                      >
                        📋 التفاصيل
                      </button>
                      {/* ⚠️ زر التدقيق يبقى ظاهراً **حتى بعد الحكم**:
                          المحاسب يگدر يبدّل حكمه قبل الاعتماد. حجبه
                          بعد أول ضغطة يخلّي الحكم الغلط بلا طريق رجوع
                          إلا سحب اعتماد ما صار. */}
                      {canApprove && inv.status !== 'APPROVED' && (
                        <button
                          onClick={() => {
                            setAuditFor(inv)
                            setAuditVerdict(inv.auditVerdict || 'MATCHED')
                            setAuditNote(inv.auditNote || '')
                            setAuditAmount(inv.auditedAmount != null ? String(inv.auditedAmount) : '')
                            setAuditErr(null)
                          }}
                          className="rounded-lg border border-brand-300 px-3 py-1.5 text-xs font-bold text-brand-700 hover:bg-brand-50"
                        >
                          🔍 تدقيق
                        </button>
                      )}
                      {/* ↩ سحب الاعتماد — «لازم تخليلي خيار أكدر أرجعله
                          الفواتير الما معتمدة» */}
                      {canApprove && inv.status === 'APPROVED' && (
                        <button
                          onClick={() => revoke(inv)}
                          disabled={busyId === inv.id}
                          className="rounded-lg border border-red-300 px-3 py-1.5 text-xs font-bold text-red-700 hover:bg-red-50 disabled:opacity-50"
                        >
                          ↩ اسحب الاعتماد
                        </button>
                      )}
                      {/* ═══ الاعتماد — بعد الحكم بس ═══
                          ⚠️ الإخفاء **راحة للمحاسب مو حماية**: المنع
                          الحقيقي بالخادم (نداء مباشر يتخطّى أي إخفاء).
                          والي نكسبه هنا إن الشاشة ما تعرض زراً يفشل —
                          زر يفشل يخلّي المحاسب يظن النظام خربان. */}
                      {canApprove && inv.status !== 'APPROVED' && audited(inv) && (
                        <button
                          onClick={() => { setApproveFor(inv); setInvoiceNo(''); setApproveErr(null) }}
                          disabled={busyId === inv.id}
                          className="rounded-lg bg-brand-600 px-3 py-1.5 text-xs font-bold text-white hover:bg-brand-700 disabled:opacity-50"
                        >
                          {busyId === inv.id ? 'جاري الاعتماد...' : 'اعتماد'}
                        </button>
                      )}
                      {/* ⚠️ **الطريق الثالث** جنب الاعتماد: المحاسب
                          الي عنده شك ما يبقى محصوراً بين «اعتمدها»
                          و«اتركها معلّقة». والمعلّقة بلا سبب مكتوب
                          تبقى معلّقة شهوراً وما أحد يعرف إنها تنتظر
                          رأياً — لأن الانتظار ما انسجّل بمكان. */}
                      {canApprove && inv.status !== 'APPROVED' && audited(inv) && !atMonitor(inv) && (
                        <button
                          onClick={() => { setMonitorFor(inv); setMonitorNote(''); setMonitorErr(null) }}
                          disabled={busyId === inv.id}
                          className="rounded-lg border border-indigo-300 bg-indigo-50 px-3 py-1.5 text-xs font-bold text-indigo-700 hover:bg-indigo-100 disabled:opacity-50"
                        >
                          👁️ أرسلها للمراقب
                        </button>
                      )}
                      {/* المراقب يبتّ — وهو وحده يشوف هالزر */}
                      {isMonitor && atMonitor(inv) && (
                        <button
                          onClick={() => { setDecideFor(inv); setDecideVerdict('OK'); setDecideNote(''); setDecideErr(null) }}
                          disabled={busyId === inv.id}
                          className="rounded-lg bg-indigo-600 px-3 py-1.5 text-xs font-bold text-white hover:bg-indigo-700 disabled:opacity-50"
                        >
                          👁️ راجعها
                        </button>
                      )}
                      {/* ⚠️ المالك وحده: يرجّعها للمحاسب حتى يرتّبها من
                          جديد — للفواتير الي مرّت قبل فصل الطوابير. */}
                      {isOwner && (
                        <button
                          onClick={() => returnToAccountant(inv)}
                          disabled={busyId === inv.id}
                          className="rounded-lg border border-amber-300 bg-amber-50 px-3 py-1.5 text-xs font-bold text-amber-800 hover:bg-amber-100 disabled:opacity-50"
                        >
                          ↩️ أرجعها للمحاسب
                        </button>
                      )}
                      {/* ⚠️ وبدل ما يختفي الزر بلا تفسير: سطر يقول
                          **ليش** والخطوة الجاية. زر يختفي بصمت يخلّي
                          المحاسب يدوّر عليه. */}
                      {canApprove && inv.status !== 'APPROVED' && !audited(inv) && (
                        <span className="rounded-lg bg-sky-50 px-3 py-1.5 text-xs font-bold text-sky-700 ring-1 ring-sky-200">
                          دقّقها أول
                        </span>
                      )}
                      {/* ⚠️ المحاسب يشوف **وين** هي مو زراً معطّلاً:
                          «عند المراقب» تقول له ينتظر منو، والزر
                          المعطّل يقول «ممنوع» بلا ما يقول ليش. */}
                      {canApprove && !isMonitor && atMonitor(inv) && (
                        <span className="rounded-lg bg-indigo-50 px-3 py-1.5 text-xs font-bold text-indigo-700 ring-1 ring-indigo-200">
                          👁️ عند المراقب
                        </span>
                      )}
                      {/* الفواتير الي انعتمدت قبل ما يصير الرقم إجبارياً —
                          المحاسب يربطها بأرقامها بأثر رجعي */}
                      {canApprove && inv.status === 'APPROVED' && !inv.externalInvoiceNumber && (
                        <button
                          onClick={() => { setLinkFor(inv); setLinkNo(''); setLinkErr(null) }}
                          disabled={busyId === inv.id}
                          className="rounded-lg bg-amber-500 px-3 py-1.5 text-xs font-bold text-white hover:bg-amber-600 disabled:opacity-50"
                        >
                          🔗 اربط رقم فاتورة
                        </button>
                      )}
                      {/* تقدير الإداري يطلع غلط أحياناً والفاتورة هي الصح */}
                      {canApprove && (
                        <button
                          onClick={() => openAdjust(inv)}
                          disabled={busyId === inv.id}
                          className="rounded-lg border border-slate-300 px-3 py-1.5 text-xs font-bold text-slate-600 hover:bg-slate-50 disabled:opacity-50"
                        >
                          ✏️ تعديل المبالغ
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
              {invoices.length === 0 && (
                <tr>
                  <td colSpan={10} className="px-4 py-6 text-center text-slate-400">
                    لا توجد فواتير بعد.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

      )}

      {/* ⚠️ الصفحة چانت ترسم **كل** فاتورة بلا حد — والاستعلام
          بالخادم هم بلا LIMIT. بـ١٢١ فاتورة تمشي، ما تبقى تمشي.
          الترقيم بالواجهة لأن العدّادات الخمس تنحسب من القائمة
          الكاملة، ولو صار بالخادم تنكسر كلها. */}
      {shown.length > 0 && (
        <div className="mt-3 rounded-xl border px-4 py-3"
          style={{ backgroundColor: 'var(--sf-card)', borderColor: 'var(--bd-line)' }}>
          <Pager page={page} perPage={perPage} total={shown.length} unit="فاتورة"
            onPage={setPage} onPerPage={setPerPage} />
        </div>
      )}

      {/* تفاصيل الفاتورة — كل الي يحتاجه المحاسب بمكان واحد */}
      {/* ═══ ربط رقم فاتورة بفاتورة معتمدة أصلاً ═══ */}
      {linkFor && (
        <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 p-4">
          <div className="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl">
            <h3 className="text-lg font-bold text-[#0f2040]">🔗 ربط رقم الفاتورة المحاسبية</h3>
            <p className="mt-1 text-xs text-slate-500">
              هاي فاتورة انعتمدت قبل ما يصير الرقم إجبارياً — اربطها برقمها حتى يكتمل الأرشيف.
              <br />كود المحاسبة: <span className="font-mono font-bold text-brand-700">{linkFor.accountingCode}</span>
              {' · '}المجموع: <b>{linkFor.netTotal.toLocaleString()} د.ع</b>
            </p>
            <input
              value={linkNo}
              onChange={(e) => setLinkNo(e.target.value)}
              autoFocus
              placeholder="رقم الفاتورة الصادر من نظام المحاسبة"
              className="mt-3 w-full rounded-xl border border-slate-300 px-3 py-2.5 text-sm outline-none focus:border-brand-500"
            />
            {linkErr && <p className="mt-2 rounded-lg bg-red-50 px-3 py-2 text-xs font-bold text-red-700">{linkErr}</p>}
            <div className="mt-4 flex gap-2">
              <button
                onClick={linkNumber}
                disabled={!linkNo.trim() || busyId === linkFor.id}
                className="flex-1 rounded-xl bg-gradient-to-l from-amber-500 to-amber-700 px-4 py-2.5 text-sm font-bold text-white disabled:opacity-50"
              >
                {busyId === linkFor.id ? 'جاري الربط...' : 'اربط الرقم'}
              </button>
              <button onClick={() => setLinkFor(null)} className="rounded-xl border border-slate-300 px-4 py-2.5 text-sm font-bold text-slate-600">
                رجوع
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ═══ تعديل المحاسب على المبالغ ═══ */}
      {adjustFor && (
        <div className="fixed inset-0 z-[60] flex items-center justify-center overflow-y-auto bg-black/40 p-4">
          <div className="w-full max-w-lg rounded-2xl bg-white p-6 shadow-2xl">
            <h3 className="text-lg font-bold text-[#0f2040]">✏️ تعديل مبالغ الفاتورة</h3>
            <p className="mt-1 text-xs text-slate-500">
              كود المحاسبة: <span className="font-mono font-bold text-brand-700">{adjustFor.accountingCode}</span>
              {adjustFor.booking && (
                <> · المستلم بالحجز: <b>{(adjustFor.booking.amountCollected ?? 0).toLocaleString()} د.ع</b></>
              )}
            </p>

            <div className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-3">
              {([
                { k: 'executionCost' as const, l: 'تكاليف التنفيذ' },
                { k: 'materialsTotal' as const, l: 'مجموع المواد' },
                { k: 'discountValue' as const, l: 'الخصم' },
              ]).map((f) => (
                <div key={f.k}>
                  <label className="mb-1 block text-xs font-medium text-slate-600">{f.l}</label>
                  <input
                    type="number"
                    value={adj[f.k]}
                    onChange={(e) => setAdj((p) => ({ ...p, [f.k]: e.target.value }))}
                    className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                  />
                </div>
              ))}
            </div>

            <div className="mt-3 rounded-xl bg-slate-50 px-4 py-3 text-sm">
              المجموع الصافي الجديد: <b className="text-brand-800">{adjNet.toLocaleString()} د.ع</b>
              <span className="text-xs text-slate-400"> (كان {adjustFor.netTotal.toLocaleString()})</span>
              {adjustFor.booking && (
                <div className={`mt-1 text-xs font-bold ${
                  adjNet === (adjustFor.booking.amountCollected ?? 0) ? 'text-emerald-600' : 'text-amber-700'
                }`}>
                  {adjNet === (adjustFor.booking.amountCollected ?? 0)
                    ? '✔ مطابق للمبلغ الواصل'
                    : `الفرق عن المبلغ الواصل: ${Math.abs(adjNet - (adjustFor.booking.amountCollected ?? 0)).toLocaleString()} د.ع`}
                </div>
              )}
            </div>

            <label className="mt-3 block text-sm font-medium text-slate-600">
              سبب التعديل <span className="text-red-500">*</span>
            </label>
            <textarea
              value={adj.reason}
              onChange={(e) => setAdj((p) => ({ ...p, reason: e.target.value }))}
              rows={2}
              placeholder="مثال: تقدير الإداري جان غلط، والمعتمد فاتورة الليدر"
              className="mt-1 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
            />
            <p className="mt-1 text-[11px] text-slate-500">السبب ينحفظ مع الفاتورة — التعديل على مبلغ ما يمر بلا تفسير.</p>

            {adjErr && <p className="mt-2 rounded-lg bg-red-50 px-3 py-2 text-xs font-bold text-red-700">{adjErr}</p>}

            <div className="mt-4 flex gap-2">
              <button
                onClick={saveAdjust}
                disabled={!adj.reason.trim() || busyId === adjustFor.id}
                className="flex-1 rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-2.5 text-sm font-bold text-white disabled:opacity-50"
              >
                {busyId === adjustFor.id ? 'جاري الحفظ...' : 'حفظ التعديل'}
              </button>
              <button onClick={() => setAdjustFor(null)} className="rounded-xl border border-slate-300 px-4 py-2.5 text-sm font-bold text-slate-600">
                رجوع
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ═══ الاعتماد يمر برقم الفاتورة المحاسبية ═══ */}
      {/* ═══ نافذة التدقيق ═══
          «أول شي بالتدقيق: مطابق / غير مطابق / خطأ بالسعر».
          مطابق = سعر الفاتورة نفسه المبلغ الداخل.
          غير مطابق = يختلف.
          خطأ بالسعر = الموظف جاب أعلى أو أوطى من الفاتورة. */}
      {auditFor && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={() => setAuditFor(null)}>
          <div dir="rtl" className="w-full max-w-md rounded-2xl bg-white p-5" onClick={(e) => e.stopPropagation()}>
            <h3 className="font-bold text-[#0f2040]">🔍 تدقيق الفاتورة {auditFor.accountingCode}</h3>
            <p className="mt-1 text-xs text-slate-500">
              مبلغ الفاتورة: <b>{auditFor.netTotal.toLocaleString()} د.ع</b>
            </p>

            <div className="mt-3 flex flex-wrap gap-2">
              {AUDIT_VERDICTS.map((v) => (
                <button
                  key={v.key}
                  onClick={() => setAuditVerdict(v.key)}
                  className={`rounded-xl px-3 py-2 text-sm font-bold ${
                    auditVerdict === v.key ? v.cls + ' ring-2 ring-offset-1 ring-slate-400' : 'bg-slate-100 text-slate-600'
                  }`}
                >
                  {v.label}
                </button>
              ))}
            </div>

            <label className="mb-1 mt-3 block text-xs font-bold text-slate-500">المبلغ الي دخل فعلاً (اختياري)</label>
            <input
              type="number"
              value={auditAmount}
              onChange={(e) => setAuditAmount(e.target.value)}
              {...onEnter(saveAudit, { disabled: busyId !== null })}
              placeholder="اتركه فاضي إذا نفس مبلغ الفاتورة"
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
            />

            <label className="mb-1 mt-3 block text-xs font-bold text-slate-500">
              الملاحظة {auditVerdict !== 'MATCHED' && <span className="text-red-600">* إجبارية</span>}
            </label>
            <textarea
              value={auditNote}
              onChange={(e) => setAuditNote(e.target.value)}
              rows={3}
              placeholder="مثال: الموظف أخذ ١٢٠ ألف والفاتورة ١٠٠ ألف"
              className="w-full resize-y rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
            />
            <p className="mt-1 text-[11px] text-slate-400">👁️ المراقب والمدير راح يقرون هذي الملاحظة.</p>

            {auditErr && <p className="mt-2 rounded-lg bg-red-50 px-3 py-2 text-xs font-bold text-red-700">{auditErr}</p>}

            <div className="mt-4 flex gap-2">
              <button
                onClick={saveAudit}
                disabled={busyId !== null}
                className="flex-1 rounded-xl bg-brand-600 px-4 py-2.5 text-sm font-bold text-white disabled:opacity-50"
              >
                احفظ الحكم
              </button>
              <button onClick={() => setAuditFor(null)} className="rounded-xl border border-slate-300 px-4 py-2.5 text-sm font-bold text-slate-600">
                رجوع
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ═══ إرسال للمراقب ═══ */}
      {monitorFor && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={() => setMonitorFor(null)}>
          <div className="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-lg font-bold text-[#0f2040]">👁️ إرسال للمراقب</h3>
            <p className="mt-1 text-sm text-slate-600">
              الفاتورة راح تنتقل لطابور المراقب حتى يراجعها ويدققها، وترجعلك بحكمه.
              حكمك ({AUDIT_VERDICTS.find((v) => v.key === monitorFor.auditVerdict)?.label ?? '—'}) يبقى محفوظاً.
            </p>
            <label className="mt-4 block text-sm font-medium text-slate-600">شنو تريد المراقب ينتبهله؟</label>
            <textarea
              value={monitorNote}
              onChange={(e) => setMonitorNote(e.target.value)}
              rows={3}
              autoFocus
              placeholder="مثال: المبلغ يحتاج تأكيد من الزبون"
              className="mt-1 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm"
            />
            {monitorErr && <p className="mt-2 rounded-lg bg-red-50 px-3 py-2 text-xs font-bold text-red-700">{monitorErr}</p>}
            <div className="mt-4 flex gap-2">
              <button onClick={sendToMonitor} disabled={busyId === monitorFor.id}
                className="flex-1 rounded-xl bg-indigo-600 px-4 py-2.5 text-sm font-bold text-white disabled:opacity-50">
                👁️ أرسلها للمراقب
              </button>
              <button onClick={() => setMonitorFor(null)}
                className="rounded-xl border border-slate-300 px-4 py-2.5 text-sm font-bold text-slate-500">إلغاء</button>
            </div>
          </div>
        </div>
      )}

      {/* ═══ المراقب يبتّ ═══ */}
      {decideFor && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={() => setDecideFor(null)}>
          <div className="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-lg font-bold text-[#0f2040]">👁️ مراجعة المراقب</h3>
            <p className="mt-1 text-sm text-slate-600">
              فاتورة <b>{decideFor.accountingCode}</b> — {(decideFor.netTotal ?? 0).toLocaleString()} د.ع
            </p>
            {decideFor.monitorRequestNote && (
              <p className="mt-2 rounded-xl bg-slate-50 px-3 py-2 text-xs leading-relaxed text-slate-600">
                طلب المحاسب: {decideFor.monitorRequestNote}
              </p>
            )}
            <div className="mt-4 flex gap-2">
              {([['OK', '✅ سليمة'], ['FLAGGED', '⚠️ عندي ملاحظة']] as const).map(([k, lbl]) => (
                <button key={k} onClick={() => setDecideVerdict(k)}
                  className={`flex-1 rounded-xl px-3 py-2.5 text-sm font-bold transition ${
                    decideVerdict === k ? 'bg-indigo-600 text-white' : 'border border-slate-300 text-slate-600 hover:bg-slate-50'}`}>
                  {lbl}
                </button>
              ))}
            </div>
            <label className="mt-4 block text-sm font-medium text-slate-600">
              الملاحظة {decideVerdict === 'FLAGGED' && <span className="text-red-500">*</span>}
            </label>
            <textarea
              value={decideNote}
              onChange={(e) => setDecideNote(e.target.value)}
              rows={3}
              placeholder={decideVerdict === 'FLAGGED' ? 'المحاسب لازم يعرف شنو يصلّح' : 'اختيارية'}
              className="mt-1 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm"
            />
            {decideErr && <p className="mt-2 rounded-lg bg-red-50 px-3 py-2 text-xs font-bold text-red-700">{decideErr}</p>}
            <p className="mt-2 text-[11px] text-slate-500">
              ⚠️ الفاتورة ترجع للمحاسب بحكمك — القرار المالي يبقى بيده.
            </p>
            <div className="mt-3 flex gap-2">
              <button onClick={decideMonitor} disabled={busyId === decideFor.id}
                className="flex-1 rounded-xl bg-indigo-600 px-4 py-2.5 text-sm font-bold text-white disabled:opacity-50">
                سجّل حكمي
              </button>
              <button onClick={() => setDecideFor(null)}
                className="rounded-xl border border-slate-300 px-4 py-2.5 text-sm font-bold text-slate-500">إلغاء</button>
            </div>
          </div>
        </div>
      )}

      {approveFor && (
        <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 p-4">
          <div className="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl">
            <h3 className="text-lg font-bold text-[#0f2040]">اعتماد الفاتورة</h3>
            <p className="mt-1 text-xs text-slate-500">
              كود المحاسبة: <span className="font-mono font-bold text-brand-700">{approveFor.accountingCode}</span>
              {' · '}المجموع: <b>{approveFor.netTotal.toLocaleString()} د.ع</b>
            </p>

            <label className="mt-4 block text-sm font-medium text-slate-600">
              رقم الفاتورة المحاسبية <span className="text-red-500">*</span>
            </label>
            {/* ⚠️ Enter = «اعتماد»: المحاسب يدخل عشرات الأرقام باليوم،
                وكل رقم يعني يرفع إيده عن الكيبورد ويمسك الماوس ويدوّر
                الزر. الشرط نفس شرط تعطيل الزر — ما ينفّذ برقم فاضي. */}
            <input
              value={invoiceNo}
              onChange={(e) => setInvoiceNo(e.target.value)}
              {...onEnter(handleApprove, { disabled: !invoiceNo.trim() || busyId === approveFor.id })}
              autoFocus
              placeholder="الرقم الصادر من نظام المحاسبة"
              className="mt-1 w-full rounded-xl border border-slate-300 px-3 py-2.5 text-sm outline-none focus:border-brand-500"
            />
            <p className="mt-1.5 text-[11px] leading-relaxed text-slate-500">
              الرقم ينحفظ ويتأرشف مع الفاتورة، وتكدر تدوّر بيه بعدين من خانة البحث فوق.
              وما ينعاد على فاتورة ثانية.
            </p>

            {approveErr && (
              <p className="mt-2 rounded-lg bg-red-50 px-3 py-2 text-xs font-bold text-red-700">{approveErr}</p>
            )}

            <div className="mt-4 flex gap-2">
              <button
                onClick={handleApprove}
                disabled={!invoiceNo.trim() || busyId === approveFor.id}
                className="flex-1 rounded-xl bg-gradient-to-l from-emerald-500 to-emerald-700 px-4 py-2.5 text-sm font-bold text-white disabled:opacity-50"
              >
                {busyId === approveFor.id ? 'جاري الاعتماد...' : '✔ اعتماد الفاتورة'}
              </button>
              <button
                onClick={() => setApproveFor(null)}
                className="rounded-xl border border-slate-300 px-4 py-2.5 text-sm font-bold text-slate-600"
              >
                رجوع
              </button>
            </div>
          </div>
        </div>
      )}

      {details && (
        <div className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/40 p-4" onClick={() => setDetails(null)}>
          <div dir="rtl" className="my-8 w-full max-w-3xl rounded-2xl bg-white p-6 shadow-xl" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-start justify-between">
              <div>
                <h3 className="text-lg font-bold text-brand-900">تفاصيل الفاتورة</h3>
                <p className="font-mono text-sm text-brand-700">{details.accountingCode}</p>
              </div>
              <div className="flex items-center gap-1">
                <button
                  onClick={() => printInvoice(details, adjustments)}
                  className="rounded-lg bg-brand-600 px-3 py-1.5 text-xs font-bold text-white hover:bg-brand-700"
                >
                  🖨️ طباعة
                </button>
                <button onClick={() => setDetails(null)} className="rounded-lg px-3 py-1 text-slate-400 hover:bg-slate-100">✕</button>
              </div>
            </div>

            <section className="mt-4 rounded-xl bg-slate-50 p-4">
              <h4 className="mb-2 text-sm font-bold text-slate-700">منو رفعها</h4>
              <div className="grid grid-cols-2 gap-2 text-sm">
                <p>الليدر: <b>{details.employeeName || '—'}</b></p>
                <p>الدور: {details.employeeRole || '—'}</p>
                <p dir="ltr" className="text-right">الهاتف: {details.employeePhone || '—'}</p>
                <p>التاريخ: {new Date(details.createdAt).toLocaleString('ar-IQ')}</p>
                <p>الحالة: {details.status === 'APPROVED' ? '✔ معتمدة'
                  : atMonitor(details) ? '👁️ عند المراقب'
                    : audited(details) ? 'بانتظار الاعتماد' : '🔍 بانتظار التدقيق'}</p>
                {details.approvedByName && <p>اعتمدها: <b>{details.approvedByName}</b></p>}
                {details.adjustedReason && (
                  <p className="col-span-2 rounded-lg bg-amber-50 px-2 py-1 text-amber-800">
                    ✏️ عدّل المحاسب المبالغ: {details.adjustedReason}
                  </p>
                )}
                {details.externalInvoiceNumber && (
                  <p className="col-span-2">
                    رقم الفاتورة المحاسبية: <b className="font-mono text-emerald-700">{details.externalInvoiceNumber}</b>
                  </p>
                )}
              </div>
            </section>

            {/* رأس الهوية الموحّد — نفس الي بكل الشاشات */}
            <EntityIdentity
              booking={details.booking}
              fields={{
                customerName: details.customerName || undefined,
                customerPhone: details.customerPhone || undefined,
                address: details.customerAddress || undefined,
                bookingCode: details.bookingCode || undefined,
                leaderName: details.employeeName || undefined,
              }}
              variant="full"
              className="mt-3"
            />

            {/* ═══ شنو غيّر المحاسب ═══
                قبل، الفاتورة تعرض «عدّل المحاسب المبالغ» وسببه بس —
                بلا ما تقول شنو كان الرقم. وهسه صار محفوظ. */}
            {adjustments.length > 0 && (
              <section className="mt-3 rounded-xl border border-amber-200 bg-amber-50 p-4">
                <h4 className="mb-2 text-sm font-bold text-amber-900">✏️ تعديلات المحاسب على المبالغ</h4>
                <div className="space-y-2">
                  {adjustments.map((a) => {
                    const rows: [string, number, number][] = [
                      ['التنفيذ', a.oldExecutionCost, a.newExecutionCost],
                      ['المواد', a.oldMaterialsTotal, a.newMaterialsTotal],
                      ['الخصم', a.oldDiscountValue, a.newDiscountValue],
                      ['الصافي', a.oldNetTotal, a.newNetTotal],
                    ]
                    return (
                      <div key={a.id} className="rounded-lg bg-white px-3 py-2 text-xs">
                        <p className="font-bold text-slate-700">
                          {a.adjustedByName || 'محاسب'} • {new Date(a.createdAt).toLocaleString('en-GB')}
                        </p>
                        {/* المبلغ الي ما انتغيّر ما ينعرض — ضجيج */}
                        {rows.filter(([, o, n]) => o !== n).map(([label, o, n]) => (
                          <p key={label} className="mt-0.5">
                            {label}: <span className="text-slate-500 line-through">{o.toLocaleString()}</span>
                            {' → '}
                            <b className="text-amber-900">{n.toLocaleString()}</b>
                          </p>
                        ))}
                        <p className="mt-1 text-slate-600">السبب: {a.reason}</p>
                      </div>
                    )
                  })}
                </div>
              </section>
            )}

            <section className="mt-3 rounded-xl bg-blue-50 p-4">
              <h4 className="mb-2 text-sm font-bold text-blue-900">الزبون والحجز</h4>
              <div className="grid grid-cols-2 gap-2 text-sm">
                <p>الزبون: <b>{details.customerName || details.booking?.customer?.name || '—'}</b></p>
                <p dir="ltr" className="text-right">الهاتف: {details.customerPhone || details.booking?.customer?.phone || '—'}</p>
                <p className="col-span-2">العنوان: {details.customerAddress || details.booking?.address || '—'}</p>
                {details.bookingCode && <p>رمز الحجز: <b>{details.bookingCode}</b></p>}
                {details.booking?.service && <p>الخدمة: {details.booking.service.name}</p>}
                {details.booking && (
                  <>
                    <p>حالة الحجز: {details.booking.status}</p>
                    <p>المستلم بالحجز: <b>{(details.booking.amountCollected ?? 0).toLocaleString()} د.ع</b></p>
                    <p>تقدير الإداري: <b>{(details.booking.quotedPrice ?? 0).toLocaleString()} د.ع</b></p>
                    <p>التدقيق: {details.booking.amountVerified ? '✔ مدقق' : 'بانتظار التدقيق'}</p>
                  </>
                )}
              </div>
              {details.booking?.assignments && details.booking.assignments.length > 0 && (
                <p className="mt-2 text-sm">
                  الكادر المنفّذ: {details.booking.assignments.map((a) => a.employee?.name).filter(Boolean).join('، ')}
                </p>
              )}
            </section>

            <section className="mt-3 rounded-xl border border-slate-200 p-4">
              <h4 className="mb-2 text-sm font-bold text-slate-700">بنود التنفيذ</h4>
              <p className="mb-2 text-xs text-slate-500">المنظومات: {details.systems.join('، ') || '—'} · عدد الأجهزة: {details.totalDeviceCount}</p>
              <table className="w-full text-right text-sm">
                <thead><tr className="text-slate-400">
                  <th className="py-1">البند</th><th className="py-1">العدد</th><th className="py-1">تفاصيل</th>
                </tr></thead>
                <tbody>
                  {details.items.map((it, i) => (
                    <tr key={i} className="border-t border-slate-100">
                      <td className="py-1">{it.itemName} <span className="text-xs text-slate-400">{it.systemName}</span></td>
                      <td className="py-1">{it.count}</td>
                      <td className="py-1 text-xs text-slate-500">
                        {it.heightMeters ? `ارتفاع ${it.heightMeters}م` : ''}
                        {it.cableLengthMeters ? ` · كيبل ${it.cableLengthMeters}م` : ''}
                      </td>
                    </tr>
                  ))}
                  {details.items.length === 0 && <tr><td colSpan={3} className="py-2 text-slate-400">ماكو بنود</td></tr>}
                </tbody>
              </table>
            </section>

            {details.materials.length > 0 && (
              <section className="mt-3 rounded-xl border border-slate-200 p-4">
                <h4 className="mb-2 text-sm font-bold text-slate-700">المواد</h4>
                <table className="w-full text-right text-sm">
                  <thead><tr className="text-slate-400">
                    <th className="py-1">المادة</th><th className="py-1">الكمية</th><th className="py-1">المجموع</th>
                  </tr></thead>
                  <tbody>
                    {details.materials.map((m, i) => (
                      <tr key={i} className="border-t border-slate-100">
                        <td className="py-1">{m.name}</td>
                        <td className="py-1">{m.quantity}</td>
                        <td className="py-1">{m.lineTotal.toLocaleString()}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </section>
            )}

            <section className="mt-3 rounded-xl bg-brand-50 p-4 text-sm">
              <div className="flex justify-between"><span>تكاليف التنفيذ</span><b>{details.executionCost.toLocaleString()} د.ع</b></div>
              <div className="flex justify-between"><span>مجموع المواد</span><b>{details.materialsTotal.toLocaleString()} د.ع</b></div>
              <div className="flex justify-between"><span>الخصم</span><b>{details.discountValue.toLocaleString()} د.ع</b></div>
              <div className="mt-2 flex justify-between border-t border-brand-200 pt-2 text-base">
                <span className="font-bold">المجموع الصافي</span>
                <b className="text-brand-800">{details.netTotal.toLocaleString()} د.ع</b>
              </div>
            </section>
          </div>
        </div>
      )}
    </div>
  )
}
