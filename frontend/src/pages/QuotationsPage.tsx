import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, type Quotation } from '../api'

const PRIMARY = '#1a237e'
const GOLD = '#c8a45a'

const statusConfig: Record<Quotation['status'], { label: string; bg: string; color: string }> = {
  NEW: { label: 'جديد', bg: '#fff3cd', color: '#856404' },
  SENT: { label: 'مرسل', bg: '#cce5ff', color: '#004085' },
  ACCEPTED: { label: 'مقبول', bg: '#d4edda', color: '#155724' },
  REJECTED: { label: 'مرفوض', bg: '#f8d7da', color: '#721c24' },
}

const fmt = (n: number) => n.toLocaleString('en-IQ')

export default function QuotationsPage() {
  const navigate = useNavigate()
  const [quotations, setQuotations] = useState<Quotation[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = () => {
    api.getQuotations()
      .then(setQuotations)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const handleDelete = async (id: string) => {
    if (!confirm('هل أنت متأكد من حذف عرض السعر؟')) return
    try {
      await api.deleteQuotation(id)
      load()
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
          <span style={{ color: GOLD, fontSize: '14px' }}>إدارة ومتابعة عروض الأسعار</span>
        </div>
        <button
          onClick={() => navigate('/quotations/new')}
          style={{
            background: GOLD,
            color: PRIMARY,
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

      {loading && <p style={{ color: '#999', textAlign: 'center', padding: '40px' }}>جاري التحميل...</p>}
      {error && (
        <div style={{ background: '#fef2f2', border: '1px solid #fecaca', borderRadius: '8px', padding: '16px', color: '#dc2626' }}>
          تعذر الاتصال بالخادم: {error}
        </div>
      )}

      {!loading && !error && (
        <div style={{
          background: 'white',
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
                  <tr key={q.id} style={{ borderBottom: '1px solid #eee' }}
                    onMouseEnter={(e) => (e.currentTarget.style.background = '#f8f9ff')}
                    onMouseLeave={(e) => (e.currentTarget.style.background = 'white')}
                  >
                    <td style={{ padding: '12px 16px', fontWeight: 'bold', color: PRIMARY }}>{q.quotationNumber}</td>
                    <td style={{ padding: '12px 16px' }}>{q.customerName}</td>
                    <td style={{ padding: '12px 16px', color: '#666' }}>{q.projectName || '-'}</td>
                    <td style={{ padding: '12px 16px', textAlign: 'center' }}>{fmt(q.grandTotal)} د.ع</td>
                    <td style={{ padding: '12px 16px', textAlign: 'center', color: '#dc2626' }}>{fmt(q.discountValue)} د.ع</td>
                    <td style={{ padding: '12px 16px', textAlign: 'center', fontWeight: 'bold', color: PRIMARY }}>{fmt(q.netTotal)} د.ع</td>
                    <td style={{ padding: '12px 16px', textAlign: 'center' }}>
                      <span style={{
                        background: st.bg, color: st.color,
                        padding: '4px 12px', borderRadius: '12px',
                        fontSize: '12px', fontWeight: 'bold',
                      }}>
                        {st.label}
                      </span>
                    </td>
                    <td style={{ padding: '12px 16px', textAlign: 'center', fontSize: '13px', color: '#666' }}>
                      {new Date(q.createdAt).toLocaleDateString('ar-IQ')}
                    </td>
                    <td style={{ padding: '12px 16px', textAlign: 'center' }}>
                      <div style={{ display: 'flex', gap: '8px', justifyContent: 'center' }}>
                        <button
                          onClick={() => navigate(`/quotations/${q.id}/edit`)}
                          style={{
                            background: '#e3f2fd', color: '#1565c0', border: 'none',
                            padding: '6px 14px', borderRadius: '6px', cursor: 'pointer',
                            fontSize: '12px', fontWeight: 'bold',
                          }}
                        >
                          تعديل
                        </button>
                        <button
                          onClick={() => handleDelete(q.id)}
                          style={{
                            background: '#fee2e2', color: '#dc2626', border: 'none',
                            padding: '6px 14px', borderRadius: '6px', cursor: 'pointer',
                            fontSize: '12px', fontWeight: 'bold',
                          }}
                        >
                          حذف
                        </button>
                      </div>
                    </td>
                  </tr>
                )
              })}
              {quotations.length === 0 && (
                <tr>
                  <td colSpan={9} style={{ padding: '40px', textAlign: 'center', color: '#999' }}>
                    لا توجد عروض أسعار بعد
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
