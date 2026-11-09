import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, type Quotation } from '../api'
import { useSession, isAdminOrOwner } from '../session'
import QuotationVersions from '../components/QuotationVersions'
import QuotationTermsEditor from '../components/QuotationTermsEditor'
import { numIQ as fmt, todayBaghdad } from '../utils/format'
import { askConfirm } from '../ui/dialogs'

const PRIMARY = '#1a237e'
// ⚠️ نسخة **النص** تنقلب بالوضع الليلي، والأصل يبقى للأسطح:
// نفس اللون يخدم عنواناً غامقاً على أبيض، ورأس جدول كحلي عليه نص أبيض.
// قلب الاثنين سوا يكسر واحداً منهما — نفس فخّ --color-white.
const PRIMARY_TEXT = 'var(--brand-ink)'
const GOLD = '#c8a45a'
// ⚠️ نسخة **النص** تنقلب بالوضع الليلي، والأصل يبقى للأسطح:
// نفس اللون يخدم عنواناً غامقاً على أبيض، ورأس جدول كحلي عليه نص أبيض.
// قلب الاثنين سوا يكسر واحداً منهما — نفس فخّ --color-white.
const GOLD_TEXT = 'var(--gold-ink)'

const statusConfig: Record<Quotation['status'], { label: string; bg: string; color: string }> = {
  NEW: { label: 'جديد', bg: '#fff3cd', color: 'var(--t-warning)' },
  SENT: { label: 'مرسل', bg: '#cce5ff', color: 'var(--t-info)' },
  ACCEPTED: { label: 'مقبول', bg: '#d4edda', color: 'var(--t-success)' },
  REJECTED: { label: 'مرفوض', bg: '#f8d7da', color: 'var(--t-danger)' },
}


// (ع) 11-07: صلاحية العرض — بس للجديد والمرسل (المقبول والمرفوض انحسم)
function ValidityBadge({ q }: { q: Quotation }) {
  if (!q.validUntil || (q.status !== 'NEW' && q.status !== 'SENT')) return null
  const day = q.validUntil.slice(0, 10)
  const left = Math.round((new Date(day).getTime() - new Date(todayBaghdad()).getTime()) / 86400000)
  const [txt, cls] = left < 0 ? ['⌛ منتهي', 'bg-red-100 text-red-700'] : left <= 2 ? [`⏳ ينتهي ${left === 0 ? 'اليوم' : `خلال ${left} يوم`}`, 'bg-amber-100 text-amber-800'] : [`صالح لغاية ${day}`, 'text-t-muted']
  return <span className={`mt-1 block rounded-full px-2 py-0.5 text-[11px] font-bold ${cls}`}>{txt}</span>
}

export default function QuotationsPage() {
  const navigate = useNavigate()
  const { employee, permissions } = useSession()
  const [showTerms, setShowTerms] = useState(false)
  const canTerms = isAdminOrOwner(employee)
    || ['quotation_create', 'quotation_edit_own', 'quotation_manage_all', 'quotation_system'].some((p) => permissions.includes(p))
  const isAdmin = isAdminOrOwner(employee)
  const [quotations, setQuotations] = useState<Quotation[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  /** العرض الي انفتح تاريخه — واحد بس، حتى الجدول ما يتمدّد كله. */
  const [historyId, setHistoryId] = useState<string | null>(null)
  // قرار (ع) 10-06: «صلح عروض الاسعار بتقسيم الشاشه» — ٥٠ عرض بالصفحة، بلا
  // بنود ولا صور (چانت تنزل كل العروض وصور منتجاتها مرة وحدة).
  const [page, setPage] = useState(1)
  const [pages, setPages] = useState(1)
  const [total, setTotal] = useState(0)

  const load = (searchValue: string, p = page) => {
    api.getQuotationsPage(p, searchValue)
      .then((r) => { setQuotations(r.items); setPages(Math.max(1, r.pages)); setTotal(r.total) })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setLoading(true)
    const t = setTimeout(() => load(search, page), 300)
    return () => clearTimeout(t)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [search, page])

  const handleDelete = async (id: string) => {
    if (!(await askConfirm('هل أنت متأكد من حذف عرض السعر؟'))) return
    try {
      await api.deleteQuotation(id)
      load(search)
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    }
  }

  return (
    <div style={{ direction: 'rtl', fontFamily: "'Segoe UI', Tahoma, Arial, sans-serif" }}>
      {/* Header */}
      <div style={{
        background: `linear-gradient(135deg, ${PRIMARY}, #283593)`,
        color: 'white',
        padding: '20px 30px',
        borderRadius: '12px',
        marginBottom: '24px',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
      }}>
        <div>
          <h1 className="m-0 text-[24px]">عروض الأسعار</h1>
          <span style={{ color: GOLD_TEXT, fontSize: '14px' }}>إدارة ومتابعة عروض الأسعار</span>
        </div>
        <div className="flex gap-2 flex-wrap">
        {canTerms && (
          <button onClick={() => setShowTerms(true)}
            style={{ background: 'transparent', color: '#fff', border: '1px solid rgba(255,255,255,.5)', padding: '10px 16px', borderRadius: '8px', cursor: 'pointer', fontWeight: 'bold', fontSize: '14px' }}>
            ⚙️ الشروط الافتراضية
          </button>
        )}
        {/* (دازاي 11-06) عرض المشاريع بالشكل الجديد — بجانب القديم */}
        <button
          onClick={() => navigate('/quotations/project/new')}
          style={{ background: 'transparent', color: '#fff', border: '1px solid rgba(255,255,255,.7)', padding: '10px 16px', borderRadius: '8px', cursor: 'pointer', fontWeight: 'bold', fontSize: '14px' }}
        >
          📘 عرض مشاريع جديد
        </button>
        <button
          onClick={() => navigate('/quotations/new')}
          style={{
            background: GOLD,
            // ⚠️ نص غامق ثابت: الخلفية ذهبية بالوضعين، فالنص ما ينقلب.
            color: '#1a237e',
            border: 'none',
            padding: '10px 24px',
            borderRadius: '8px',
            cursor: 'pointer',
            fontWeight: 'bold',
            fontSize: '14px',
          }}
        >
          + عرض سعر جديد
        </button>
        </div>
        {showTerms && <QuotationTermsEditor onClose={() => setShowTerms(false)} />}
      </div>

      <div className="mb-5">
        <input
          type="text"
          value={search}
          onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          placeholder="ابحث باسم الزبون أو المشروع أو رقم العرض..."
          style={{
            width: '100%', maxWidth: '420px', padding: '10px 16px',
            border: `2px solid ${PRIMARY}`, borderRadius: '10px', fontSize: '14px', outline: 'none',
          }}
        />
      </div>

      {loading && <p className="text-t-faint text-center p-10">جاري التحميل...</p>}
      {error && (
        <div className="bg-sf-danger border border-[#fecaca] rounded-lg p-4 text-t-danger">
          تعذر الاتصال بالخادم: {error}
        </div>
      )}

      {!loading && !error && (
        <div style={{
          background: 'var(--sf-card)',
          border: `2px solid ${PRIMARY}`,
          borderRadius: '12px',
          overflow: 'hidden',
        }}>
          <table className="w-full border-collapse">
            <thead>
              <tr style={{ background: PRIMARY, color: 'white' }}>
                <th className="py-3 px-4 text-[13px] text-right">رقم العرض</th>
                <th className="py-3 px-4 text-[13px] text-right">اسم الزبون</th>
                <th className="py-3 px-4 text-[13px] text-right">المشروع</th>
                <th className="py-3 px-4 text-[13px] text-center">المجموع</th>
                <th className="py-3 px-4 text-[13px] text-center">الخصم</th>
                <th className="py-3 px-4 text-[13px] text-center">الصافي</th>
                <th className="py-3 px-4 text-[13px] text-center">الحالة</th>
                <th className="py-3 px-4 text-[13px] text-center">التاريخ</th>
                <th className="py-3 px-4 text-[13px] text-center">إجراءات</th>
              </tr>
            </thead>
            <tbody>
              {quotations.map((q) => {
                const st = statusConfig[q.status]
                return (
                  <tr key={q.id} className="border-b border-bd-line"
                    onMouseEnter={(e) => (e.currentTarget.style.background = '#f8f9ff')}
                    onMouseLeave={(e) => (e.currentTarget.style.background = 'white')}
                  >
                    <td style={{ padding: '12px 16px', fontWeight: 'bold', color: PRIMARY_TEXT }}>{q.quotationNumber}</td>
                    <td className="py-3 px-4">{q.customerName}</td>
                    <td className="py-3 px-4 text-t-muted">{q.projectName || '-'}</td>
                    <td className="py-3 px-4 text-center">{fmt(q.grandTotal)} د.ع</td>
                    <td className="py-3 px-4 text-center text-t-danger">{fmt(q.discountValue)} د.ع</td>
                    <td style={{ padding: '12px 16px', textAlign: 'center', fontWeight: 'bold', color: PRIMARY_TEXT }}>{fmt(q.netTotal)} د.ع</td>
                    <td className="py-3 px-4 text-center">
                      <span style={{
                        background: st.bg, color: st.color,
                        padding: '4px 12px', borderRadius: '12px',
                        fontSize: '12px', fontWeight: 'bold',
                      }}>
                        {st.label}
                      </span>
                      <ValidityBadge q={q} />
                    </td>
                    <td className="py-3 px-4 text-center text-[13px] text-t-muted">
                      {new Date(q.createdAt).toLocaleDateString('ar-IQ')}
                    </td>
                    <td className="py-3 px-4 text-center">
                      <div className="flex gap-2 justify-center">
                        <button
                          onClick={() => navigate(q.layout === 'PROJECT' ? `/quotations/project/${q.id}` : `/quotations/${q.id}/edit?preview=1`)}
                          className="bg-sf-violet text-t-violet border-0 py-1.5 px-3.5 rounded-md cursor-pointer text-[12px] font-bold"
                        >
                          👁️ معاينة
                        </button>
                        {/* 🔴 باب الأرشيف: التعديل چان يمحي النسخة
                            القديمة تماماً، وهسه تنحفظ — بس أرشيف بلا
                            باب ما إله قيمة. */}
                        <button
                          onClick={() => setHistoryId(historyId === q.id ? null : q.id)}
                          className="bg-sf-sunken text-t-muted border-0 py-1.5 px-3.5 rounded-md cursor-pointer text-[12px] font-bold"
                        >
                          📚 النسخ القديمة
                        </button>
                        <button
                          onClick={() => navigate(q.layout === 'PROJECT' ? `/quotations/project/${q.id}` : `/quotations/${q.id}/edit`)}
                          className="bg-sf-info text-t-info border-0 py-1.5 px-3.5 rounded-md cursor-pointer text-[12px] font-bold"
                        >
                          تعديل
                        </button>
                        {isAdmin && (
                          <button
                            onClick={() => handleDelete(q.id)}
                            className="bg-sf-danger text-t-danger border-0 py-1.5 px-3.5 rounded-md cursor-pointer text-[12px] font-bold"
                          >
                            حذف
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                )
              })}
              {/* صف التاريخ ينفتح تحت العرض نفسه — مو بنافذة، حتى
                  يقرا الأرقام القديمة وهو يشوف الجديدة بنفس الشاشة. */}
              {historyId && quotations.some((q) => q.id === historyId) && (
                <tr>
                  <td colSpan={9} className="py-3 px-4 bg-sf-sunken">
                    <QuotationVersions quotationId={historyId} />
                  </td>
                </tr>
              )}
              {quotations.length === 0 && (
                <tr>
                  <td colSpan={9} className="p-10 text-center text-t-faint">
                    {search ? 'لا توجد نتائج مطابقة' : 'لا توجد عروض أسعار بعد'}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
          {pages > 1 && (
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '10px', padding: '12px', borderTop: `1px solid ${PRIMARY}` }}>
              <button type="button" disabled={page <= 1} onClick={() => setPage(page - 1)}
                style={{ padding: '6px 14px', borderRadius: '8px', border: `1px solid ${PRIMARY}`, background: 'transparent', color: PRIMARY_TEXT, cursor: 'pointer', opacity: page <= 1 ? 0.4 : 1 }}>→ السابقة</button>
              <span style={{ fontSize: '13px', color: PRIMARY_TEXT }}>صفحة {page} من {pages} · {total} عرض</span>
              <button type="button" disabled={page >= pages} onClick={() => setPage(page + 1)}
                style={{ padding: '6px 14px', borderRadius: '8px', border: `1px solid ${PRIMARY}`, background: 'transparent', color: PRIMARY_TEXT, cursor: 'pointer', opacity: page >= pages ? 0.4 : 1 }}>التالية ←</button>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
