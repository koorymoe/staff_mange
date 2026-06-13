import { Router } from 'express'
import { prisma } from '../prisma'

const router = Router()

// GET /api/stats - overview statistics for the MONITOR/ADMIN dashboard
router.get('/', async (_req, res) => {
  const [
    totalCustomers,
    totalBookings,
    pendingBookings,
    confirmedBookings,
    completedBookings,
    cancelledBookings,
    urgentPending,
    revenue,
    unverifiedRevenue,
    recentBookings,
  ] = await Promise.all([
    prisma.customer.count(),
    prisma.booking.count(),
    prisma.booking.count({ where: { status: 'PENDING' } }),
    prisma.booking.count({ where: { status: 'CONFIRMED' } }),
    prisma.booking.count({ where: { status: 'COMPLETED' } }),
    prisma.booking.count({ where: { status: 'CANCELLED' } }),
    prisma.booking.count({ where: { status: 'PENDING', priority: 'URGENT' } }),
    prisma.booking.aggregate({
      where: { amountCollected: { not: null } },
      _sum: { amountCollected: true },
    }),
    prisma.booking.aggregate({
      where: { amountCollected: { not: null }, amountVerified: false },
      _sum: { amountCollected: true },
    }),
    prisma.booking.findMany({
      take: 8,
      orderBy: { createdAt: 'desc' },
      include: { customer: true, service: true },
    }),
  ])

  // Per sales employee: how many bookings they transferred / got confirmed
  const employees = await prisma.employee.findMany({
    include: {
      transferredBookings: {
        select: { id: true, status: true },
      },
      assignments: {
        select: { booking: { select: { id: true, status: true, amountCollected: true } } },
      },
    },
  })

  const salesStats = employees
    .filter((e) => e.role === 'SALES' || e.role === 'ADMIN')
    .map((e) => ({
      employeeId: e.id,
      name: e.name,
      totalTransferred: e.transferredBookings.length,
      confirmed: e.transferredBookings.filter((b) => b.status !== 'PENDING' && b.status !== 'CANCELLED').length,
    }))
    .sort((a, b) => b.confirmed - a.confirmed)

  const technicianStats = employees
    .filter((e) => e.role === 'TECHNICIAN')
    .map((e) => ({
      employeeId: e.id,
      name: e.name,
      onDuty: e.onDuty,
      totalAssigned: e.assignments.length,
      completed: e.assignments.filter((a) => a.booking.status === 'COMPLETED').length,
      revenueHandled: e.assignments
        .filter((a) => a.booking.status === 'COMPLETED')
        .reduce((sum, a) => sum + (a.booking.amountCollected || 0), 0),
    }))
    .sort((a, b) => b.completed - a.completed)

  // Bookings grouped by service
  const bookingsByService = await prisma.booking.groupBy({
    by: ['serviceId'],
    _count: { _all: true },
  })
  const services = await prisma.service.findMany({
    where: { id: { in: bookingsByService.map((b) => b.serviceId).filter((id): id is string => !!id) } },
  })
  const serviceBreakdown = bookingsByService
    .map((b) => ({
      serviceId: b.serviceId,
      name: services.find((s) => s.id === b.serviceId)?.name || 'بدون خدمة',
      count: b._count._all,
    }))
    .sort((a, b) => b.count - a.count)

  // Employee role headcount
  const roleCounts = await prisma.employee.groupBy({
    by: ['role'],
    _count: { _all: true },
  })

  res.json({
    totals: {
      totalCustomers,
      totalBookings,
      pendingBookings,
      confirmedBookings,
      completedBookings,
      cancelledBookings,
      urgentPending,
      totalRevenue: revenue._sum.amountCollected || 0,
      unverifiedRevenue: unverifiedRevenue._sum.amountCollected || 0,
    },
    salesStats,
    technicianStats,
    serviceBreakdown,
    roleCounts: roleCounts.map((r) => ({ role: r.role, count: r._count._all })),
    recentBookings: recentBookings.map((b) => ({
      id: b.id,
      code: b.code,
      status: b.status,
      priority: b.priority,
      customerName: b.customer.name,
      serviceName: b.service?.name || null,
      createdAt: b.createdAt,
    })),
  })
})

export default router
