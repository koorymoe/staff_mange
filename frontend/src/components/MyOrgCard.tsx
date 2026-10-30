import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import MyOrgItems from './MyOrgItems'
import { MyOutputs } from './OrgOutputs'
import { orgApi, orgPositions, type OrgPerson } from '../org'
import { useSession } from '../session'
import '../pages/OrgPage.css'

export default function MyOrgCard() {
  const { employee } = useSession()
  return employee ? (
    <>
      <MyPlace key={employee.id} />
      <MyOrgItems />
      <MyOutputs />
    </>
  ) : null
}
function MyPlace() {
  const [items, setItems] = useState<OrgPerson[] | null>(null)
  const [error, setError] = useState('')
  const [reload, setReload] = useState(0)
  useEffect(() => {
    let alive = true
    orgApi
      .me()
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
  return (
    <section
      className="org-ui org-my-card"
      dir="rtl"
      aria-label="مكاني بالهيكل"
    >
      <h2>🏛️ مكاني</h2>
      {error ? (
        <p role="alert">
          {error}{' '}
          <button onClick={() => setReload((n) => n + 1)}>أعد المحاولة</button>
        </p>
      ) : !items ? (
        <p role="status">دا أجيب مكانك…</p>
      ) : items.length === 0 ? (
        <p>بعدك ما انوضعت بالهيكل</p>
      ) : (
        <ul>
          {items.map((p) => (
            <li key={p.id}>
              <strong>{p.divisionName || 'اسم الإدارة مو متوفر'}</strong>
              {p.unitId && (
                <span> · {p.unitName || 'اسم الوحدة مو متوفر'}</span>
              )}
              <p>
                {orgPositions[p.position] ?? p.position}
                {p.acting && ' · بالوكالة'}
              </p>
              {(p.position === 'DIRECTOR' || p.position === 'DEPUTY') && (
                <Link to={`/org/desk/${encodeURIComponent(p.divisionId)}`}>
                  افتح مكتب الإدارة ←
                </Link>
              )}
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
