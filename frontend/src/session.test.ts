import { describe, expect, it } from 'vitest'
import type { Employee } from './api'
import { can, hasRole, isAdminOrOwner } from './session'

const emp = (x: Partial<Employee>) => ({ id: 'e', name: 'x', role: 'TECHNICIAN', ...x }) as Employee

describe('session', () => {
  it('isAdminOrOwner: الأدمن، والمالك بالدور أو بالحقيقي', () => {
    expect(isAdminOrOwner(emp({ role: 'ADMIN' }))).toBe(true)
    expect(isAdminOrOwner(emp({ role: 'OWNER' }))).toBe(true)
    expect(isAdminOrOwner(emp({ role: 'ADMIN', actualRole: 'OWNER' } as Partial<Employee>))).toBe(true)
    expect(isAdminOrOwner(emp({}))).toBe(false)
    expect(isAdminOrOwner(null)).toBe(false)
  })
  it('hasRole: الدور الأصلي أو من المنصب', () => {
    expect(hasRole(emp({ role: 'FINANCE' }), 'FINANCE')).toBe(true)
    expect(hasRole(emp({ orgRoles: ['MONITOR'] }), 'MONITOR')).toBe(true)
    expect(hasRole(emp({}), 'MONITOR')).toBe(false)
  })
  it('can: من جدول السيرفر، وundefined للجلسة القديمة', () => {
    expect(can(emp({ can: { funds: true, credits: false } }), 'funds')).toBe(true)
    expect(can(emp({ can: { funds: true, credits: false } }), 'credits')).toBe(false)
    expect(can(emp({}), 'funds')).toBeUndefined()
  })
})
