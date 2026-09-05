import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, type Quotation } from '../api'
import { useSession } from '../session'

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

const fmt = (n: number) => n.toLocaleString('en-IQ')

export default function QuotationsPage() {
  const navigate = useNavigate()
  const { employee } = useSession()
  const isAdmin = employee?.role === 'ADMIN'
  const [quotations, setQuotations] = useState<Quotation[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')

  const load = (searchValue: string) => {
    api.getQuotations(searchValue)
      .then(setQuotations)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setLoading(true)
    const t = setTimeout(() => load(search), 300)
    return () => clearTimeout(t)
  }, [search])

  const handleDelete = async (id: string) => {
    if (!confirm('هل أنت متأكد من حذف عرض السعر؟')) return
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
          <h1 style={{ margin: 0, fontSize: '24px' }}>عروض الأسعار</h1>
          <span style={{ color: GOLD_TEXT, fontSize: '14px' }}>إدارة ومتابعة عروض الأسعار</span>
        </div>
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

      <div style={{ marginBottom: '20px' }}>
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="ابحث باسم الزبون أو المشروع أو رقم العرض..."
          style={{
            width: '100%', maxWidth: '420px', padding: '10px 16px',
            border: `2px solid ${PRIMARY}`, borderRadius: '10px', fontSize: '14px', outline: 'none',
          }}
        />
      </div>

      {loading && <p style={{ color: 'var(--t-faint)', textAlign: 'center', padding: '40px' }}>جاري التحميل...</p>}
      {error && (
        <div style={{ background: 'var(--sf-danger)', border: '1px solid #fecaca', borderRadius: '8px', padding: '16px', color: 'var(--t-danger)' }}>
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
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ background: PRIMARY, color: 'white' }}>
                <th style={{ padding: '12px 16px', fontSize: '13px', textAlign: 'right' }}>رقم العرض</th>
                <th style={{ padding: '12px 16px', fontSize: '13px', textAlign: 'right' }}>اسم الزبون</th>
                <th style={{ padding: '12px 16px', fontSize: '13px', textAlign: 'right' }}>المشروع</th>
                <th style={{ padding: '12px 16px', fontSize: '13px', textAlign: 'center' }}>المجموع</th>
                <th style={{ padding: '12px 16px', fontSize: '13px', textAlign: 'center' }}>الخصم</th>
                <th style={{ padding: '12px 16px', fontSize: '13px', textAlign: 'center' }}>الصافي</th>
                <th style={{ padding: '12px 16px', fontSize: '13px', textAlign: 'center' }}>الحالة</th>
                <th style={{ padding: '12px 16px', fontSize: '13px', textAlign: 'center' }}>التاريخ</th>
                <th style={{ padding: '12px 16px', fontSize: '13px', textAlign: 'center' }}>إجراءات</th>
              </tr>
            </thead>
            <tbody>
              {quotations.map((q) => {
                const st = statusConfig[q.status]
                return (
                  <tr key={q.id} style={{ borderBottom: '1px solid var(--bd-line)' }}
                    onMouseEnter={(e) => (e.currentTarget.style.background = '#f8f9ff')}
                    onMouseLeave={(e) => (e.currentTarget.style.background = 'white')}
                  >
                    <td style={{ padding: '12px 16px', fontWeight: 'bold', color: PRIMARY_TEXT }}>{q.quotationNumber}</td>
                    <td style={{ padding: '12px 16px' }}>{q.customerName}</td>
                    <td style={{ padding: '12px 16px', color: 'var(--t-muted)' }}>{q.projectName || '-'}</td>
                    <td style={{ padding: '12px 16px', textAlign: 'center' }}>{fmt(q.grandTotal)} د.ع</td>
                    <td style={{ padding: '12px 16px', textAlign: 'center', color: 'var(--t-danger)' }}>{fmt(q.discountValue)} د.ع</td>
                    <td style={{ padding: '12px 16px', textAlign: 'center', fontWeight: 'bold', color: PRIMARY_TEXT }}>{fmt(q.netTotal)} د.ع</td>
                    <td style={{ padding: '12px 16px', textAlign: 'center' }}>
                      <span style={{
                        background: st.bg, color: st.color,
                        padding: '4px 12px', borderRadius: '12px',
                        fontSize: '12px', fontWeight: 'bold',
                      }}>
                        {st.label}
                      </span>
                    </td>
                    <td style={{ padding: '12px 16px', textAlign: 'center', fontSize: '13px', color: 'var(--t-muted)' }}>
                      {new Date(q.createdAt).toLocaleDateString('ar-IQ')}
                    </td>
                    <td style={{ padding: '12px 16px', textAlign: 'center' }}>
                      <div style={{ display: 'flex', gap: '8px', justifyContent: 'center' }}>
                        <button
                          onClick={() => navigate(`/quotations/${q.id}/edit?preview=1`)}
                          style={{
                            background: 'var(--sf-violet)', color: 'var(--t-violet)', border: 'none',
                            padding: '6px 14px', borderRadius: '6px', cursor: 'pointer',
                            fontSize: '12px', fontWeight: 'bold',
                          }}
                        >
                          👁️ معاينة
                        </button>
                        <button
                          onClick={() => navigate(`/quotations/${q.id}/edit`)}
                          style={{
                            background: 'var(--sf-info)', color: 'var(--t-info)', border: 'none',
                            padding: '6px 14px', borderRadius: '6px', cursor: 'pointer',
                            fontSize: '12px', fontWeight: 'bold',
                          }}
                        >
                          تعديل
                        </button>
                        {isAdmin && (
                          <button
                            onClick={() => handleDelete(q.id)}
                            style={{
                              background: 'var(--sf-danger)', color: 'var(--t-danger)', border: 'none',
                              padding: '6px 14px', borderRadius: '6px', cursor: 'pointer',
                              fontSize: '12px', fontWeight: 'bold',
                            }}
                          >
                            حذف
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                )
              })}
              {quotations.length === 0 && (
                <tr>
                  <td colSpan={9} style={{ padding: '40px', textAlign: 'center', color: 'var(--t-faint)' }}>
                    {search ? 'لا توجد نتائج مطابقة' : 'لا توجد عروض أسعار بعد'}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
