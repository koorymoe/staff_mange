import { Router } from 'express'
import { prisma } from '../prisma'

const router = Router()

// GET /api/stats - overview statistics for the MONITOR/ADMIN dashboard
router.get('/', async (_req, res) => {
  const [totalCustomers, totalBookings, confirmedBookings, completedBookings, revenue] =
    await Promise.all([
      prisma.customer.count(),
      prisma.booking.count(),
      prisma.booking.count({ where: { status: 'CONFIRMED' } }),
      prisma.booking.count({ where: { status: 'COMPLETED' } }),
      prisma.booking.aggregate({
        where: { amountCollected: { not: null } },
        _sum: { amountCollected: true },
      }),
    ])

  // Per sales employee: how many bookings they transferred / got confirmed
  const employees = await prisma.employee.findMany({
    include: {
      transferredBookings: {
        select: { id: true, status: true },
      },
      assignments: {
        select: { booking: { select: { id: true, status: true } } },
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

  const technicianStats = employees
    .filter((e) => e.role === 'TECHNICIAN' || e.role === 'PROJECT_MANAGER')
    .map((e) => ({
      employeeId: e.id,
      name: e.name,
      totalAssigned: e.assignments.length,
      completed: e.assignments.filter((a) => a.booking.status === 'COMPLETED').length,
    }))

  res.json({
    totalCustomers,
    totalBookings,
    confirmedBookings,
    completedBookings,
    totalRevenue: revenue._sum.amountCollected || 0,
    salesStats,
    technicianStats,
  })
})

export default router
