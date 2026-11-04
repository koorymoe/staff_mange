import { useEffect, useRef, useState } from 'react'
import { apiRequest } from '../api'
import { orgDeskWrite, type DeskItem } from '../orgDesk'
import { useSession } from '../session'
import '../pages/OrgPage.css'
import '../pages/OrgDeskPage.css'

type MyItem = DeskItem & {
  divisionName: string
  unitName: string | null
  period: string
  planTitle: string
  planStatus: string
}
export default function MyOrgItems() {
  const { employee } = useSession()
  return employee ? <MyItems key={employee.id} viewerId={employee.id} /> : null
}
function MyItems({ viewerId }: { viewerId: string }) {
  const [items, setItems] = useState<MyItem[] | null>(null),
    [error, setError] = useState(''),
    [reload, setReload] = useState(0)
  useEffect(() => {
    let alive = true
    apiRequest<MyItem[]>('/org/my-items')
      .then((rows) => {
        if (alive) {
          setItems(rows ?? [])
          setError('')
        }
      })
      .catch((e: Error) => {
        if (alive) setError(e.message)
      })
    return () => {
      alive = false
    }
  }, [reload])
  const refresh = () => {
    setItems(null)
    setReload((n) => n + 1)
  }
  return (
    <section className="org-ui org-my-card" dir="rtl" aria-label="بنودي">
      <h2>📌 بنودي</h2>
      <p>بنود الخطط الي إنت مسؤول عنها، بدون ما تحتاج تفتح مكتب الإدارة.</p>
      {error ? (
        <p role="alert" className="org-error">
          {error} <button onClick={refresh}>أعد جلب بنودي</button>
        </p>
      ) : !items ? (
        <p role="status">دا أجيب بنودك…</p>
      ) : items.length === 0 ? (
        <p className="desk-meta">ماكو بنود مكلف بيها حالياً.</p>
      ) : (
        <div className="desk-cards">
          {items.map((item) => (
            <OwnItem
              key={item.id}
              item={item}
              allowed={item.ownerId === viewerId}
              saved={refresh}
            />
          ))}
        </div>
      )}
    </section>
  )
}
function OwnItem({
  item,
  allowed,
  saved,
}: {
  item: MyItem
  allowed: boolean
  saved: () => void
}) {
  const [manual, setManual] = useState(item.manualProgress),
    [error, setError] = useState(''),
    [busy, setBusy] = useState(false)
  const lock = useRef(false),
    computed = item.byMatrix || !!item.metric
  const update = async (body: { manualProgress?: number; done?: boolean }) => {
    if (!allowed || lock.current || (computed && body.manualProgress != null))
      return
    lock.current = true
    setBusy(true)
    setError('')
    try {
      await orgDeskWrite(
        `plan-items/${encodeURIComponent(item.id)}`,
        body,
        'PATCH',
      )
      saved()
    } catch (e) {
      setError((e as Error).message)
    } finally {
      lock.current = false
      setBusy(false)
    }
  }
  return (
    <article className="org-panel desk-plan-card">
      <h3>{item.title}</h3>
      <span className={`desk-plan-status plan-${item.planStatus}`}>
        {(
          {
            DRAFT: 'مسودة',
            SUBMITTED: 'مرفوعة',
            APPROVED: 'معتمدة',
            RETURNED: 'مرجعة للتعديل',
          } as Record<string, string>
        )[item.planStatus] ?? 'حالة الخطة مو متوفرة'}
      </span>
      <p className="desk-meta">
        {item.divisionName}
        {item.unitName && ` · ${item.unitName}`} · {item.planTitle} ·{' '}
        <bdi>{item.period}</bdi>
      </p>
      <div className="desk-item-meta">
        <span>
          الموعد: <bdi>{item.dueDate || 'مو محدد'}</bdi>
        </span>
        {item.overdue && <span className="desk-overdue">متأخر</span>}
      </div>
      <p>الإنجاز: {item.progress}٪</p>
      <progress
        aria-label={`إنجاز ${item.title}`}
        max={100}
        value={Math.min(100, Math.max(0, item.progress))}
      />
      {computed ? (
        <p className="desk-meta">
          بند مربوط بمقياس؛ ما نعدّل نسبته يدوياً. النسبة المعروضة هي الي رجعها
          الخادم.
        </p>
      ) : (
        allowed &&
        !item.doneAt && (
          <div className="desk-manual">
            <label>
              نسبة إنجاز بندي: {manual}٪
              <input
                type="range"
                min={0}
                max={100}
                step={1}
                disabled={busy}
                value={manual}
                onChange={(e) => setManual(Number(e.target.value))}
              />
            </label>
            <button
              disabled={busy || manual === item.manualProgress}
              onClick={() => void update({ manualProgress: manual })}
            >
              احفظ نسبة بندي
            </button>
          </div>
        )
      )}
      {item.doneAt ? (
        <p className="desk-resolution">✓ خلص</p>
      ) : (
        allowed && (
          <button disabled={busy} onClick={() => void update({ done: true })}>
            خلص بندي ✓
          </button>
        )
      )}
      {error && (
        <p role="alert" className="org-error">
          {error}
        </p>
      )}
    </article>
  )
}
