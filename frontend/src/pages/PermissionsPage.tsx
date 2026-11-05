import { useSession } from '../session'
import './PermissionsPage.css'
import { AccessPage } from './permissions/AccessPage'

export default function PermissionsPage({
  embedded = false,
}: { embedded?: boolean } = {}) {
  const { employee } = useSession()
  const owner = employee?.actualRole === 'OWNER'
  if (!owner && employee?.role !== 'ADMIN')
    return (
      <p className="p-6" role="alert">
        الصلاحيات والوصول للمالك ومدير النظام فقط.
      </p>
    )
  return <AccessPage owner={owner} embedded={embedded} />
}
