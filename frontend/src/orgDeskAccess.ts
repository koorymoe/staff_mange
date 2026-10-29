import { useEffect, useState } from 'react'
import type { Employee } from './api'
import { orgApi, type OrgPerson } from './org'

export const orgAdmin = (e: Employee | null | undefined) =>
  e?.actualRole === 'OWNER' || e?.role === 'ADMIN'
export const orgHeads = (items: OrgPerson[] | null | undefined) =>
  (items ?? []).filter(
    (p) => p.position === 'DIRECTOR' || p.position === 'DEPUTY',
  )
export const orgSeesAll = (items: OrgPerson[] | null | undefined) =>
  orgHeads(items).some((p) => p.divisionId === 'admin' && !p.unitId)
export const orgDeskAllowed = (
  e: Employee | null | undefined,
  items: OrgPerson[] | null | undefined,
  id?: string,
) =>
  !!e &&
  (orgAdmin(e) ||
    orgHeads(items).some(
      (p) =>
        !id || p.divisionId === id || (p.divisionId === 'admin' && !p.unitId),
    ))
// Navigation metadata only, kept in sync with model.OrgSeed. Management heads
// cannot call /org/chart. Never probe desks to discover their names.
export const orgDeskSeeds = [
  { id: 'admin', name: 'الإدارة' },
  { id: 'ops', name: 'العمليات' },
  { id: 'bizdev', name: 'تطوير الأعمال' },
  { id: 'eng', name: 'الهندسية' },
  { id: 'quality', name: 'الجودة' },
  { id: 'supply', name: 'سلاسل التوريد' },
  { id: 'contracts', name: 'العقود والشؤون التجارية والإعلامية' },
  { id: 'it', name: 'التقنية IT' },
]
export function useOrgMemberships(employee: Employee | null | undefined) {
  const id = employee?.id
  const admin = orgAdmin(employee)
  const [state, setState] = useState<{
    id: string
    items: OrgPerson[]
    error: string
  } | null>(null)
  const [reload, setReload] = useState(0)
  useEffect(() => {
    const refresh = () => setReload((n) => n + 1)
    window.addEventListener('focus', refresh)
    window.addEventListener('org-memberships-changed', refresh)
    return () => {
      window.removeEventListener('focus', refresh)
      window.removeEventListener('org-memberships-changed', refresh)
    }
  }, [])
  useEffect(() => {
    if (!id || admin) return
    let alive = true
    orgApi
      .me()
      .then((items) => {
        if (alive) setState({ id, items: items ?? [], error: '' })
      })
      .catch((e: Error) => {
        if (alive) setState({ id, items: [], error: e.message })
      })
    return () => {
      alive = false
    }
  }, [id, admin, reload])
  const current = state?.id === id ? state : null
  return {
    items: admin ? [] : (current?.items ?? null),
    error: admin ? '' : (current?.error ?? ''),
    retry: () => setReload((n) => n + 1),
  }
}
