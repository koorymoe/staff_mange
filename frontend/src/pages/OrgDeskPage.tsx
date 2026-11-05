import { useParams } from 'react-router-dom'
import { useSession } from '../session'
import { orgAdmin, orgDeskAllowed, orgDeskSeeds, orgHeads, orgSeesAll, useOrgMemberships } from '../orgDeskAccess'
import './OrgPage.css'
import './OrgDeskPage.css'
import { DeskDirectory } from './orgdesk/DeskContent'

export default function OrgDeskPage() {
  const { employee } = useSession()
  const { divisionId } = useParams()
  const membership = useOrgMemberships(employee)
  if (!employee) return <p>سجّل دخولك حتى تفتح المكتب.</p>
  if (!orgAdmin(employee) && membership.error)
    return (
      <p role="alert">
        {membership.error}{' '}
        <button onClick={membership.retry}>أعد المحاولة</button>
      </p>
    )
  if (!orgAdmin(employee) && membership.items === null)
    return <p role="status">دا أتحقق من منصبك…</p>
  if (!orgDeskAllowed(employee, membership.items, divisionId))
    return <p role="alert">ما عندك منصب إداري يفتح هالمكتب.</p>
  const heads = orgHeads(membership.items)
  const choices = orgSeesAll(heads)
    ? orgDeskSeeds
    : [
        ...new Map(
          heads.map((p) => [
            p.divisionId,
            { id: p.divisionId, name: p.divisionName || p.divisionId },
          ]),
        ).values(),
      ]
  return (
    <DeskDirectory
      key={employee.id}
      admin={orgAdmin(employee)}
      choices={choices}
      divisionId={divisionId}
      managementHead={orgSeesAll(heads)}
    />
  )
}
