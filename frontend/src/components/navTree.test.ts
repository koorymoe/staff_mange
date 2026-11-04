import { describe, expect, it } from 'vitest'
import type { Employee } from '../api'
import { isNavVisible, type NavContext, type NavItem } from './navTree'

const ctx = (e: Partial<Employee>, permissions: string[] = []): NavContext =>
  ({ employee: { id: 'e', name: 'x', role: 'SALES', ...e } as Employee, permissions, gpsServiceId: null })
const item = (x: Partial<NavItem>): NavItem => ({ to: '/credits', label: 'الآجل', icon: null, ...x })

describe('isNavVisible', () => {
  const credits = item({ capability: 'credits', roles: ['ADMIN', 'FINANCE', 'MONITOR'], unlockAnyPermission: ['finance'] })
  it('capability: جدول السيرفر يحسم إذا موجود', () => {
    expect(isNavVisible(credits, ctx({ can: { credits: true } }))).toBe(true)
    expect(isNavVisible(credits, ctx({ role: 'FINANCE', can: { credits: false } }))).toBe(false)
  })
  it('capability: بلا جدول يرجع للفحوصات القديمة', () => {
    expect(isNavVisible(credits, ctx({}, ['finance']))).toBe(true)
  })
  it('ownerOnly يمنع غير المالك حتى لو الجدول يگول إي', () => {
    const own = item({ capability: 'credits', ownerOnly: true })
    expect(isNavVisible(own, ctx({ can: { credits: true } }))).toBe(false)
  })
  it('hideForRoles قبل كلشي', () => {
    const hidden = item({ capability: 'credits', hideForRoles: ['SALES'] })
    expect(isNavVisible(hidden, ctx({ can: { credits: true } }))).toBe(false)
  })
})
