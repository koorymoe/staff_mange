import { useEffect, useState } from 'react'
import { apiRequest } from './api'
import { orgRegulations, type OrgClause } from './orgDesk'
export type OutputTarget = {
  kind: 'employee' | 'unit' | 'division'
  id: string
}
export type OutputRow = {
  clause?: number
  code: string
  title: string
  score: number | null
  matrixPct: number | null
  tasksOnTime: number
  tasksLate: number
  tasksOverdue: number
  kpiDeductions: number
  complaints: number
}
export type OutputCard = {
  label: string
  period: string
  people: number
  rows: OutputRow[]
  weakest: string[]
}
export const getOutputs = (target: OutputTarget, period: string) =>
  apiRequest<OutputCard>(
    `/org/outputs?${new URLSearchParams({ [`${target.kind}Id`]: target.id, period })}`,
  )
export function useOutputRegulations(enabled: boolean) {
  const [clauses, setClauses] = useState<OrgClause[] | null>(null),
    [error, setError] = useState(''),
    [reload, setReload] = useState(0)
  useEffect(() => {
    if (!enabled) return
    let alive = true
    orgRegulations()
      .then((rows) => {
        if (alive) {
          setClauses(rows ?? [])
          setError('')
        }
      })
      .catch((e: Error) => {
        if (alive) setError(e.message)
      })
    return () => {
      alive = false
    }
  }, [enabled, reload])
  return {
    clauses: enabled ? clauses : null,
    error: enabled ? error : '',
    retry: () => setReload((n) => n + 1),
  }
}
