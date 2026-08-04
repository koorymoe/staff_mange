import { useState, useEffect } from 'react'
import { useAutoRefresh } from '../useAutoRefresh'
import PrivacyPolicyGate from './PrivacyPolicyGate'
import { NavLink, Outlet, useLocation } from 'react-router-dom'
import { api, type Employee, type EmployeeRole } from '../api'
import { SessionContext, roleLabels, hasGpsSkill } from '../session'
import Login from '../pages/Login'
import TrainingPage from '../pages/TrainingPage'
import AssistantWidget from './AssistantWidget'
import ManagerAssistantChat from './ManagerAssistantChat'
import SettingsPanel from './SettingsPanel'

interface NavItem {
  to: string
  label: string
  icon: React.ReactNode
  end?: boolean
  roles?: EmployeeRole[]
  permission?: string
  anyPermission?: string[]
  leaderOnly?: boolean
  gpsSkillOnly?: boolean
  // يظهر لكل الكوادر عدا الفني العادي (الفني الليدر يشوفه). القيد
  // بالعكس حتى ما ننسى دور جديد لما ينضاف.
  notForPlainTechnician?: boolean
  // صلاحية ظهور الوحدة كاملة: منحها للموظف يفتحله الوحدة وكل صفحاتها،
  // بغض النظر عن صلاحياته التفصيلية — هذا معنى "أنطيه الصلاحية ويشوف".
  unitPermission?: string
  children?: NavItem[]
  divider?: boolean
}

const I = ({ d }: { d: string }) => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d={d} /></svg>
)

const navItems: NavItem[] = [
  { to: '/', label: 'الرئيسية', end: true, icon: <I d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z M9 22V12h6v10" /> },
  { to: '/attendance', label: 'الحضور', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg> },
  // حساب تكلفة التنصيب للتنفيذ — فقرة رئيسية تحت الرئيسية مباشرة،
  // بكل الحسابات (إداري، ليدر، إدارة) عدا الفني العادي.
  {
    to: '/leader-invoices/new?mode=estimate', label: 'حساب تكلفة التنصيب للتنفيذ',
    icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><rect x="4" y="2" width="16" height="20" rx="2"/><path d="M8 6h8M8 10h3M13 10h3M8 14h3M13 14h3M8 18h8"/></svg>,
    notForPlainTechnician: true,
  },
  // الإجازات: مع نظام الحضور — أي موظف يقدّم طلبه من هنا، والمخوّل يشوف صندوق الموافقات
  { to: '/leaves', label: 'الإجازات', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="4" width="18" height="18" rx="2"/><path d="M16 2v4M8 2v4M3 10h18"/><path d="m9 16 2 2 4-4"/></svg> },
  // تصنيفي: صفحة شخصية عامة لكل الأدوار — لازم تبقى بمستوى مستقل بره "الإدارة"،
  // لأنه الفني/الليدر ما عندهم وصول لأي شي ثاني بالإدارة، فتضل قائمة فاضية
  // بالنسبة الهم لو حطيناها جوه.
  { to: '/my-ranking', label: 'تصنيفي', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>, roles: ['ADMIN', 'SALES', 'HR_COORDINATOR', 'TECHNICIAN', 'MONITOR', 'FINANCE', 'GPS_ADMIN', 'QUALITY_ENGINEER', 'PROCUREMENT_ADMIN'] },

  // ── الإدارة ──
  {
    // مجموعة الإدارة بدون قيد أدوار — ظهورها يعتمد على أبنائها (كل ابن مقيّد بدوره/صلاحيته)،
    // حتى أي موظف ينمنح صلاحية إدارية (مثل إدارة المشاريع) توصله من دون تغيير دوره.
    to: '/admin-group', label: 'الإدارة', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9"/></svg>,
    children: [
      {
        // بدون قيد أدوار هنا أيضاً — نفس مبدأ مجموعة "الإدارة" الأعلى: قيد على المجموعة
        // الوسيطة يمنع ظهورها بالكامل حتى لو ابن معيّن مسموح لدور/صلاحية موظف غير
        // مذكور بهذي القائمة (مثال حقيقي: PROCUREMENT_ADMIN مع صلاحية "جرد الأدوات").
        to: '/mgmt-employees', label: 'إدارة الموظفين', icon: <></>,
        children: [
          { to: '/employees', label: 'إدارة الكوادر', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'], permission: 'staff_management' },
          {
            to: '/mgmt-permissions', label: 'إدارة الصلاحيات', icon: <></>,
            children: [
              { to: '/permissions', label: 'الصلاحيات', icon: <></>, roles: ['ADMIN'] },
              // صلاحية التقني (محتوى) معزولة عن باقي الصلاحيات — تطلع لأي موظف عنده
              // هذي الصلاحية المخصصة بغض النظر عن دوره، بدل ما تكون بس تحت ADMIN.
              { to: '/training-management', label: 'صلاحية التقني (محتوى)', icon: <></>, permission: 'content_technician' },
            ],
          },
          { to: '/employee-stats', label: 'إحصائيات الموظفين الشهرية', icon: <></>, roles: ['ADMIN'] },
          { to: '/kpi', label: 'تقييم الأداء', icon: <></>, roles: ['ADMIN', 'MONITOR'], permission: 'kpi_management' },
          { to: '/inventory', label: 'جرد الأدوات', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR', 'PROCUREMENT_ADMIN'], permission: 'inventory' },
          { to: '/stats', label: 'إحصائيات الموظفين', icon: <></>, roles: ['ADMIN', 'MONITOR'], permission: 'staff_management' },
          { to: '/complaints', label: 'الشكاوى', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'], permission: 'complaints' },
          { to: '/quality-follow-ups', label: 'متابعة الجودة', icon: <></>, roles: ['ADMIN', 'MONITOR', 'QUALITY_ENGINEER'], permission: 'quality_control' },
          // طلبات الكادر الواردة من إدارة المشاريع — إداري الكوادر يلبيها
          { to: '/staff-requests', label: 'طلبات الكادر', icon: <></>, roles: ['HR_COORDINATOR'] },
          // تقييم الأداء (منفصل عن KPI) — إداري الكوادر يقيّم تيم ليدرات الفرق
          { to: '/performance-review', label: 'تقييم الأداء', icon: <></>, roles: ['HR_COORDINATOR'] },
        ],
      },
      // إدارة الإحصائيات — عنصر مستقل مباشر تحت "الإدارة"، مو داخل إدارة
      // الموظفين، حصراً لمدير النظام.
      { to: '/stats-management', label: 'إدارة الإحصائيات', icon: <></>, roles: ['ADMIN'] },
      {
        // بدون قيد أدوار على المجموعة الوسيطة — ظهورها يتقرر من أبنائها فقط.
        // قيد الأدوار هنا كان يخفي "إدارة العمل" كاملة عن أي دور مو بالقائمة
        // (مثل إداري الكميات) حتى لو منحناه صلاحيات حجز/تنسيق/عملاء صراحةً.
        to: '/mgmt-work', label: 'إدارة العمل', icon: <></>,
        children: [
          { to: '/sales', label: 'حجز جديد', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'QUALITY_ENGINEER'], permission: 'sales_booking' },
          { to: '/customers', label: 'العملاء', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'], permission: 'manage_customers' },
          { to: '/bookings', label: 'الحجوزات', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR', 'FINANCE'], permission: 'view_bookings' },
          { to: '/coordinator', label: 'تنسيق الحجوزات', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'], permission: 'coordinator' },
          { to: '/services', label: 'الخدمات', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'], permission: 'manage_services' },
          { to: '/missions', label: 'تتبع المهام', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'], permission: 'mission_tracking' },
        ],
      },
      {
        // الجي بي اس صارت خدمة بتحكم صلاحية "gps_system" — مو دور وظيفي منفصل،
        // فأي موظف عنده هذي الصلاحية (مسؤول خدمة الجي بي اس أو المراقب) يشوفها.
        to: '/mgmt-services', label: 'إدارة الخدمات', icon: <></>,
        children: [
          { to: '/gps', label: 'نظام GPS', icon: <></>, permission: 'gps_system' },
          { to: '/gps/requests', label: 'طلبات GPS المعلقة', icon: <></>, permission: 'gps_system' },
          // متابعة التجديد تخص مسؤول الجي بي اس ومهندس الجودة سوا
          { to: '/gps/follow-up', label: '🔄 متابعة تجديد الاشتراكات', icon: <></>, anyPermission: ['gps_system', 'quality_control'] },
          { to: '/gps/renewals-review', label: 'طلبات تجديد GPS', icon: <></>, permission: 'gps_system' },
          { to: '/gps/maintenance-review', label: 'طلبات صيانة GPS', icon: <></>, permission: 'gps_system' },
          { to: '/service-managers', label: 'مسؤولو الخدمات', icon: <></>, roles: ['ADMIN'] },
        ],
      },
      {
        // إدارة المشاريع صارت صلاحية: أي موظف عنده project_management يشوفها بغض النظر عن دوره
        to: '/mgmt-projects', label: 'إدارة المشاريع', icon: <></>,
        children: [
          { to: '/projects', label: 'المشاريع', icon: <></>, anyPermission: ['project_management', 'project_create_only'] },
          { to: '/project-work-types', label: 'إعدادات: أنواع الأعمال', icon: <></>, permission: 'project_management' },
          { to: '/project-statistics', label: '📊 إحصائيات المشاريع', icon: <></>, permission: 'project_management' },
      { to: '/checklists', label: 'الكشوفات', icon: <></>, permission: 'project_management' },
          { to: '/staff-requests', label: 'طلبات الكادر', icon: <></>, permission: 'project_management' },
          { to: '/quotations', label: 'عروض الأسعار', icon: <></>, anyPermission: ['quotation_create', 'quotation_edit_own', 'quotation_manage_all', 'quotation_system'] },
          { to: '/products', label: 'المنتجات', icon: <></>, anyPermission: ['quotation_manage_all', 'quotation_system'] },
        ],
      },
      {
        to: '/mgmt-finance', label: 'إدارة الحسابات', icon: <></>,
        children: [
          { to: '/finance', label: 'تدقيق الحسابات', icon: <></>, roles: ['ADMIN', 'FINANCE', 'MONITOR'], permission: 'finance' },
          // فواتير الليدر تترحّل للمحاسب بتفاصيلها حتى يدققها ويعتمدها
          { to: '/revolving-fund', label: '💵 الدوار', icon: <></>, permission: 'revolving_fund' },
          { to: '/gps-install-costs', label: '🔧 حساب تكاليف الشد', icon: <></>, roles: ['ADMIN', 'FINANCE'] },
          { to: '/leader-invoices', label: '🧾 فواتير الليدر', icon: <></>, anyPermission: ['finance', 'leader_basket'] },
          { to: '/expenses', label: 'إدارة المصاريف', icon: <></>, roles: ['ADMIN', 'FINANCE'] },
          { to: '/my-expenses', label: 'المصاريف', icon: <></>, roles: ['ADMIN'], permission: 'expenses' },
        ],
      },
      {
        to: '/mgmt-procurement', label: 'إدارة المشتريات', icon: <></>,
        children: [
          { to: '/procurement', label: 'طلبات المواد', icon: <></>, roles: ['ADMIN', 'MONITOR', 'PROJECT_MANAGER', 'TECHNICIAN', 'PROCUREMENT_ADMIN'], permission: 'procurement' },
          { to: '/suppliers', label: 'الموردون', icon: <></>, anyPermission: ['suppliers_management'] },
        ],
      },
      {
        // إدارة المركبات ما تظل معزولة بره قائمة الإدارة — صارت مجموعة فرعية هنا
        to: '/mgmt-vehicles', label: 'إدارة المركبات', icon: <></>,
        children: [
          { to: '/vehicles', label: 'إدارة المركبات', icon: <></>, permission: 'vehicle_management' },
        ],
      },
    ],
  },

  { to: '/monitor', label: 'لوحة المراقبة', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>, roles: ['ADMIN', 'MONITOR'], permission: 'monitoring' },
  // "خريطة المواقع" انشالت من القائمة — الفني هسه يشوف طريق مهمته مباشرة
  // من صفحة "مهامي" (بوب-أب داخل نفس الصفحة، بدون تحويل لصفحة ثانية).
  // مدير المشاريع مدير مو فني: ما عنده مهام تنستلم ولا تقييم ولا تصنيف ولا تقارير عمل
  {
    // مجموعة "العمل" للفني/الليدر — مهامه اليومية ومصاريفه وتقاريره وفواتيره
    to: '/tech-work-group', label: 'العمل', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>,
    children: [
      { to: '/my-tasks', label: 'مهامي', icon: <></>, roles: ['TECHNICIAN'] },
      { to: '/my-expenses', label: 'مصاريفي', icon: <></>, roles: ['TECHNICIAN', 'PROJECT_MANAGER'] },
      { to: '/work-reports', label: 'تقارير العمل', icon: <></>, roles: ['TECHNICIAN'] },
      { to: '/leader-invoices', label: 'فواتير الليدر', icon: <></>, roles: ['TECHNICIAN'], leaderOnly: true },
      { to: '/leader-invoices/new?mode=estimate', label: 'حساب كلفة (استفسار زبون)', icon: <></>, roles: ['TECHNICIAN'], leaderOnly: true },
      // استمارة الكاميرات — شيت مستقل بالاكسل بمعادلة مختلفة عن تكاليف المشروع
      { to: '/camera-cost', label: 'حساب كلفة كاميرات المراقبة', icon: <></>, roles: ['TECHNICIAN'], leaderOnly: true },
    ],
  },
  {
    // مجموعة "الجرد" — جرد الأدوات الشخصية وجرد الفريق (تيم ليدر بس يشوف الثانية)
    to: '/tech-inventory-group', label: 'الجرد', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/></svg>,
    children: [
      { to: '/my-inventory', label: 'جرد أدواتي', icon: <></>, roles: ['TECHNICIAN'] },
      { to: '/team-inventory', label: 'جرد الفريق', icon: <></>, roles: ['TECHNICIAN'], leaderOnly: true },
    ],
  },
  // تيم ليدر بس يقيّم فنيي فريقه (منفصل عن KPI)
  { to: '/performance-review', label: 'تقييم فريقي', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>, roles: ['TECHNICIAN'], leaderOnly: true },
  // صيانة الأجهزة العامة: حصراً للتيم ليدر (شيت "صيانة الاجهزة")
  { to: '/device-maintenance', label: 'صيانة الأجهزة', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><rect x="2" y="7" width="20" height="14" rx="2"/><path d="M16 3h-8a2 2 0 0 0-2 2v2h12V5a2 2 0 0 0-2-2z"/></svg>, roles: ['TECHNICIAN'], leaderOnly: true },
  { to: '/gps/employee', label: 'لوحتي GPS', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/></svg>, roles: ['TECHNICIAN'], gpsSkillOnly: true },

  // ═══ فاصل: تحته "الوحدات" — كل وحدة إدارية تجمع محتوياتها تحت باب واحد.
  // الجودة/مراجعة تقارير العمل صارت بس داخل "وحدة الجودة والسلامة المهنية"،
  // وتدقيق تنسيق الحجوزات صار بس داخل "وحدة الرقابة" — ما ضلوا عناصر
  // منفصلة هنا حتى ما تتكرر بالقائمة.
  // ملاحظة: وحدات "الإعلام والعلاقات العامة" و"التصميم" و"التقنيات (IT)" ما
  // ضفناها لأنه ما عندها صفحات مبنية بالنظام بعد — تحتاج طلب منفصل لبنائها.
  // المشاريع الموجّهة لي: أي موظف ينوجّهله مشروع يشوفه هنا بكل مراحله — بدون
  // ما ننطيه صلاحية إدارة المشاريع العامة. الصفحة تطلع فاضية لو ماكو شي.
  { to: '/my-projects', label: 'المشاريع الموجّهة لي', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><path d="M14 2v6h6"/><path d="M9 15l2 2 4-4"/></svg> },

  { to: '/units-divider', label: '── الوحدات ──', icon: <></>, divider: true },

  {
    // وحدة الخدمة: استقبال وتنسيق طلبات الزبائن وتنفيذها
    to: '/unit-service', label: 'وحدة الخدمة', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>,
    unitPermission: 'unit_service',
    children: [
      { to: '/sales', label: 'حجز جديد', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'QUALITY_ENGINEER'], permission: 'sales_booking' },
      { to: '/customers', label: 'العملاء', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'], permission: 'manage_customers' },
      { to: '/bookings', label: 'الحجوزات', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR', 'FINANCE'], permission: 'view_bookings' },
      { to: '/coordinator', label: 'تنسيق الحجوزات', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'], permission: 'coordinator' },
      { to: '/services', label: 'الخدمات', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'], permission: 'manage_services' },
      { to: '/missions', label: 'تتبع المهام', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'], permission: 'mission_tracking' },
    ],
  },
  {
    // وحدة التقنيين: خمس فقرات منفصلة بالقائمة الجانبية (كل ضغطة تفتح صفحتها
    // هي بس، مو كل الفقرات مع بعض) — إدارة المعارض / المنتجات / الخدمات /
    // مفردات التدريب (ترحّلت من وحدة التدريب المستقلة السابقة) / معرض الأعمال.
    // permission: 'unit_technicians' على المجموعة نفسها (مو بس على الأبناء) —
    // بوابة إضافية فوق بوابات الأبناء، حتى الوحدة كاملة تختفي عن أي حد ما
    // يملك هذي الصلاحية تحديداً، بغض النظر عن أي صلاحيات ثانية عنده (كان
    // content_technician لحاله كافي يفتحها بالغلط لأي حد يملكها من مكان ثاني).
    to: '/unit-technicians', label: 'وحدة التقنيين', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/></svg>,
    unitPermission: 'unit_technicians',
    permission: 'unit_technicians',
    children: [
      { to: '/exhibitions', label: 'إدارة المعارض', icon: <></>, permission: 'unit_technicians' },
      { to: '/product-requests', label: 'إدارة المنتجات', icon: <></>, permission: 'unit_technicians' },
      { to: '/service-studies', label: 'إدارة الخدمات', icon: <></>, permission: 'unit_technicians' },
      { to: '/training-management', label: 'مفردات التدريب', icon: <></>, permission: 'content_technician' },
      { to: '/tech-showcase', label: 'معرض الأعمال', icon: <></>, permission: 'content_technician' },
    ],
  },
  // وحدتان فارغتان مؤقتاً (بانتظار تحديد آلية العمل والصلاحيات المطلوبة لكل
  // وحدة) — تطلعان لمدير النظام بس، وتفتحان صفحة "قريباً" بدل ما تختفيان بالكامل.
  {
    to: '/unit-design-group', label: 'وحدة التصميم', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 2a4.5 4.5 0 0 0 0 9 4.5 4.5 0 0 1 0 9"/></svg>, roles: ['ADMIN'],
    unitPermission: 'unit_design',
    children: [
      { to: '/design-forms/quick-add', label: 'إضافة سؤال', icon: <></>, roles: ['ADMIN'] },
      { to: '/design-forms', label: 'فورمة التصميم', icon: <></>, roles: ['ADMIN'] },
    ],
  },
  {
    // وحدة الإعلام والعلاقات العامة: علاقات الشركة مع زبائنها — الشخصيات
    // المهمة (VIP) الي يأشّرها الموظفون، وسياسة الخصوصية المعلنة.
    to: '/unit-pr', label: 'وحدة الإعلام والعلاقات العامة', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M11 5L6 9H2v6h4l5 4V5z"/><path d="M15.5 8.5a5 5 0 0 1 0 7"/></svg>,
    unitPermission: 'unit_pr',
    children: [
      { to: '/vip-customers', label: '⭐ الشخصيات المهمة', icon: <></>, roles: ['ADMIN'] },
    ],
  },

  {
    to: '/unit-quality', label: 'وحدة الجودة والسلامة المهنية', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z"/></svg>,
    unitPermission: 'unit_quality',
    children: [
      { to: '/quality', label: 'الجودة', icon: <></>, permission: 'quality_control' },
      { to: '/work-reports-review', label: 'مراجعة تقارير العمل', icon: <></>, anyPermission: ['monitoring', 'quality_control'] },
      { to: '/quality-follow-ups', label: 'متابعة الجودة', icon: <></>, roles: ['ADMIN', 'MONITOR', 'QUALITY_ENGINEER'], permission: 'quality_control' },
    ],
  },
  {
    // وحدة الرقابة: كل أدوات المراقبة والتدقيق العام بمكان واحد
    to: '/unit-monitoring', label: 'وحدة الرقابة', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>,
    unitPermission: 'unit_monitoring',
    children: [
      { to: '/monitor', label: 'لوحة المراقبة', icon: <></>, roles: ['ADMIN', 'MONITOR'], permission: 'monitoring' },
      { to: '/crew-bookings-audit', label: 'تدقيق تنسيق الحجوزات', icon: <></>, permission: 'crew_management' },
      { to: '/complaints', label: 'الشكاوى', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'], permission: 'complaints' },
      // إدارة سياسة الخصوصية (إضافة/تعديل النقاط) — صلاحية مستقلة تماماً عن
      // قراءتها. القراءة متاحة لكل موظف من الرابط فوق "تسجيل الخروج".
      { to: '/privacy-policy', label: '🔒 إدارة سياسة الخصوصية', icon: <></>, permission: 'privacy_policy_manage' },
    ],
  },
  {
    to: '/unit-procurement', label: 'وحدة المشتريات والمخازن', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M3 3h18v18H3z"/></svg>,
    unitPermission: 'unit_procurement',
    children: [
      { to: '/procurement', label: 'طلبات المواد', icon: <></>, roles: ['ADMIN', 'MONITOR', 'PROJECT_MANAGER', 'TECHNICIAN', 'PROCUREMENT_ADMIN'], permission: 'procurement' },
      { to: '/suppliers', label: 'الموردون', icon: <></>, anyPermission: ['suppliers_management'] },
      { to: '/inventory', label: 'جرد الأدوات', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR', 'PROCUREMENT_ADMIN'], permission: 'inventory' },
      // إدارة المركبات تظهر هنا كمان (مو بس بمجموعتها) — المخازن مسؤولة عنها
      { to: '/vehicles', label: 'إدارة المركبات', icon: <></>, permission: 'vehicle_management' },
      // لوحة تحكم الأسطول محلها هنا — المخازن هيه المسؤولة عن الأسطول
      { to: '/fleet-dashboard', label: 'لوحة تحكم الأسطول', icon: <></>, permission: 'vehicle_management' },
    ],
  },
  {
    to: '/unit-finance', label: 'وحدة الحسابات', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M12 1v22M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>,
    unitPermission: 'unit_finance',
    children: [
      { to: '/finance', label: 'تدقيق الحسابات', icon: <></>, roles: ['ADMIN', 'FINANCE', 'MONITOR'], permission: 'finance' },
      { to: '/revolving-fund', label: '💵 الدوار', icon: <></>, permission: 'revolving_fund' },
          { to: '/gps-install-costs', label: '🔧 حساب تكاليف الشد', icon: <></>, roles: ['ADMIN', 'FINANCE'] },
          { to: '/leader-invoices', label: '🧾 فواتير الليدر', icon: <></>, anyPermission: ['finance', 'leader_basket'] },
      { to: '/expenses', label: 'إدارة المصاريف', icon: <></>, roles: ['ADMIN', 'FINANCE'] },
      { to: '/my-expenses', label: 'المصاريف', icon: <></>, roles: ['ADMIN'], permission: 'expenses' },
    ],
  },
  {
    to: '/unit-hr', label: 'وحدة الكوادر التنفيذية', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="8.5" cy="7" r="4"/></svg>,
    unitPermission: 'unit_hr',
    children: [
      { to: '/employees', label: 'إدارة الكوادر', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'], permission: 'staff_management' },
      { to: '/kpi', label: 'تقييم الأداء', icon: <></>, roles: ['ADMIN', 'MONITOR'], permission: 'kpi_management' },
      { to: '/staff-requests', label: 'طلبات الكادر', icon: <></>, roles: ['HR_COORDINATOR'] },
      { to: '/performance-review', label: 'تقييم الأداء', icon: <></>, roles: ['HR_COORDINATOR'] },
    ],
  },
  {
    to: '/unit-projects', label: 'وحدة إدارة المشاريع', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="3" width="18" height="18" rx="2"/></svg>,
    unitPermission: 'unit_projects',
    children: [
      { to: '/projects', label: 'المشاريع', icon: <></>, anyPermission: ['project_management', 'project_create_only'] },
          { to: '/project-work-types', label: 'إعدادات: أنواع الأعمال', icon: <></>, permission: 'project_management' },
      { to: '/project-statistics', label: '📊 إحصائيات المشاريع', icon: <></>, permission: 'project_management' },
      { to: '/checklists', label: 'الكشوفات', icon: <></>, permission: 'project_management' },
      { to: '/quotations', label: 'عروض الأسعار', icon: <></>, anyPermission: ['quotation_create', 'quotation_edit_own', 'quotation_manage_all', 'quotation_system'] },
      { to: '/products', label: 'المنتجات', icon: <></>, anyPermission: ['quotation_manage_all', 'quotation_system'] },
    ],
  },

  // ── اختصارات سريعة (أهم إجراءات المبيعات) ──
  { to: '/sales', label: 'حجز جديد', icon: <I d="M12 5v14M5 12h14" />, roles: ['SALES'] },
  { to: '/complaints', label: 'حجز شكوى', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M12 9v4M12 17h.01M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/></svg>, roles: ['SALES'] },

  // ── مجموعة GPS (تلم كل طلبات الـ GPS الخاصة بالمبيعات تحت باب وحد) ──
  {
    to: '/gps-group', label: 'GPS', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/></svg>,
    roles: ['SALES'],
    children: [
      { to: '/gps/purchase', label: 'طلب GPS جديد', icon: <></>, roles: ['SALES'] },
      { to: '/gps/delivery', label: 'تسليم أجهزة GPS', icon: <></>, roles: ['SALES'] },
      { to: '/gps/renewal', label: 'طلب تجديد GPS', icon: <></>, roles: ['SALES'] },
      { to: '/gps/maintenance-request', label: 'طلب صيانة GPS', icon: <></>, roles: ['SALES'] },
    ],
  },
]

const loadStoredEmployee = (): Employee | null => {
  const raw = localStorage.getItem('currentEmployee')
  if (!raw) return null
  try { return JSON.parse(raw) as Employee } catch { return null }
}

function hasActiveChild(item: NavItem, pathname: string): boolean {
  if (!item.children) return pathname === item.to || pathname.startsWith(item.to + '/')
  return item.children.some(c => hasActiveChild(c, pathname))
}

const roleColors: Record<string, string> = {
  ADMIN: 'from-amber-500 to-orange-600',
  SALES: 'from-emerald-500 to-teal-600',
  HR_COORDINATOR: 'from-violet-500 to-purple-600',
  TECHNICIAN: 'from-sky-500 to-blue-600',
  PROJECT_MANAGER: 'from-rose-500 to-pink-600',
  MONITOR: 'from-cyan-500 to-teal-600',
  FINANCE: 'from-lime-500 to-green-600',
  GPS_ADMIN: 'from-indigo-500 to-blue-600',
  QUALITY_ENGINEER: 'from-fuchsia-500 to-purple-600',
  ENGINEER: 'from-teal-500 to-cyan-700',
}

export default function Layout() {
  const [employee, setEmployeeState] = useState<Employee | null>(loadStoredEmployee)
  const [employeePermissions, setEmployeePermissions] = useState<string[]>([])
  const [gpsServiceId, setGpsServiceId] = useState<string | null>(null)
  const [expandedGroups, setExpandedGroups] = useState<Record<string, boolean>>({})
  const [collapsed, setCollapsed] = useState(false)
  const [mobileOpen, setMobileOpen] = useState(false)
  const [notifications, setNotifications] = useState<import('../api').Notification[]>([])
  const [unreadCount, setUnreadCount] = useState(0)
  const [notifOpen, setNotifOpen] = useState(false)
  const [settingsOpen, setSettingsOpen] = useState(false)
  const location = useLocation()

  // Closing the mobile nav on route change is a one-line UI reset tied to router
  // navigation, not data fetching; safe to keep as a synchronous effect.
  // eslint-disable-next-line react-hooks/set-state-in-effect
  useEffect(() => { setMobileOpen(false) }, [location.pathname])

  // تحديث تلقائي كل نص ساعة + معالجة رجوع الموظف للتبويب بعد غياب طويل
  useAutoRefresh(!!employee)

  const toggleSidebar = () => {
    if (typeof window !== 'undefined' && window.innerWidth < 1024) setMobileOpen((o) => !o)
    else setCollapsed((c) => !c)
  }

  const setEmployee = (emp: Employee | null) => {
    // المالك (OWNER) لازم يشوف ويسوي كل شي مدير النظام (ADMIN) يسويه، وأكثر —
    // نطبّع role إلى 'ADMIN' حتى كل شرط role === 'ADMIN' بالواجهة يشتغل له
    // تلقائياً بدون تعديل كل مكان، ونحفظ الدور الحقيقي بـ actualRole للعرض
    // وللتحقق الحصري بصفحة المراقبة الخلفية.
    const normalized = emp && emp.role === 'OWNER' ? { ...emp, actualRole: 'OWNER' as const, role: 'ADMIN' as const } : emp
    setEmployeeState(normalized)
    if (normalized) {
      localStorage.setItem('currentEmployee', JSON.stringify(normalized))
    } else {
      localStorage.removeItem('currentEmployee')
      localStorage.removeItem('authToken')
    }
  }

  // نتحقق من هوية الموظف الحقيقية من السيرفر مرة وحدة عند فتح النظام —
  // هذا يصحح تلقائياً أي بيانات جلسة قديمة/معدَّلة (مثلاً بأدوات المطورين
  // بالمتصفح) بقيت محفوظة بذاكرة المتصفح المحلية من قبل، ويسجل خروج
  // الحساب تلقائياً إذا صار موقوف (SUSPENDED) بينما الجلسة القديمة لسه مفتوحة
  useEffect(() => {
    if (!employee) return
    api.getMe()
      .then((fresh) => setEmployee(fresh))
      .catch(() => setEmployee(null))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    // Clearing notifications when the user logs out is a guard-clause reset of local
    // state tied to the `employee` dependency, not a fetch; safe to keep as-is.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    if (!employee) { setNotifications([]); setUnreadCount(0); return }
    const load = () => api.getNotifications().then((r) => { setNotifications(r.notifications); setUnreadCount(r.unreadCount) }).catch(() => {})
    load()
    const interval = setInterval(load, 30000)
    return () => clearInterval(interval)
  }, [employee])

  const handleNotifClick = async (n: import('../api').Notification) => {
    if (!n.read) {
      try {
        await api.markNotificationRead(n.id)
        setNotifications((prev) => prev.map((x) => (x.id === n.id ? { ...x, read: true } : x)))
        setUnreadCount((c) => Math.max(0, c - 1))
      } catch { /* ignore */ }
    }
  }

  const handleMarkAllRead = async () => {
    try {
      await api.markAllNotificationsRead()
      setNotifications((prev) => prev.map((x) => ({ ...x, read: true })))
      setUnreadCount(0)
    } catch { /* ignore */ }
  }

  useEffect(() => {
    // Guard-clause reset when logged out; not part of the fetch itself.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    if (!employee) { setEmployeePermissions([]); return }
    api.getEmployeePermissions(employee.id)
      .then((perms) => setEmployeePermissions(perms.map((p) => p.name)))
      .catch(() => setEmployeePermissions([]))
  }, [employee])

  useEffect(() => {
    // Guard-clause reset when logged out; not part of the fetch itself.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    if (!employee) { setGpsServiceId(null); return }
    api.getServices()
      .then((services) => setGpsServiceId(services.find((s) => s.name === 'GPS')?.id || null))
      .catch(() => setGpsServiceId(null))
  }, [employee])

  useEffect(() => {
    const autoExpand = (items: NavItem[]) => {
      items.forEach((item) => {
        if (item.children && hasActiveChild(item, location.pathname)) {
          setExpandedGroups((prev) => ({ ...prev, [item.label]: true }))
          autoExpand(item.children)
        }
      })
    }
    autoExpand(navItems)
  }, [location.pathname])

  if (!employee) {
    return (
      <SessionContext.Provider value={{ employee, setEmployee, permissions: employeePermissions, gpsServiceId }}>
        <Login />
      </SessionContext.Provider>
    )
  }

  // موظف قيد التدريب: يشوف صفحة التدريب فقط، بدون أي وصول لباقي النظام
  if (employee.isTrainee) {
    return (
      <SessionContext.Provider value={{ employee, setEmployee, permissions: employeePermissions, gpsServiceId }}>
        <div dir="rtl" className="min-h-screen bg-[#f0f4f9]">
          <header className="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-slate-200/60 bg-white/80 backdrop-blur-xl px-8">
            <span className="text-lg font-extrabold text-[#0f2040] tracking-tight">نظام شركة الأماني — التدريب</span>
            <button onClick={() => setEmployee(null)}
              className="rounded-lg px-3 py-1.5 text-sm font-bold text-red-500 hover:bg-red-50">
              تسجيل الخروج
            </button>
          </header>
          <main className="p-3 sm:p-5 lg:p-8">
            <TrainingPage />
          </main>
        </div>
      </SessionContext.Provider>
    )
  }

  const role = employee?.role
  const hasMonitor = role === 'MONITOR' || employeePermissions.includes('monitoring')
  const hasAudit = employeePermissions.includes('auditing')
  // unitGranted: صحيح لما يكون الموظف عنده صلاحية الوحدة الي هذا العنصر
  // داخلها — وقتها كل شي جوّا الوحدة يظهر له بدون فحص صلاحيات تفصيلية.
  const isVisible = (item: NavItem, unitGranted = false): boolean => {
    if (item.divider) return true
    // الوحدات: بوابة صارمة. الوحدة ما تطلع أبداً إلا لمن عنده صلاحية الوحدة
    // نفسها (أو مدير النظام). قبل، الوحدة جانت تطلع لمجرد إنه ابن واحد جوّاها
    // مسموح بصلاحية عامة — فصار الموظف يشوف وحدات مو إلها علاقة بشغله ويتكرر
    // نفس المحتوى مرتين (مرة بـ"الإدارة" ومرة بالوحدة).
    // القاعدة المطلوبة: "الإدارة" = كل شي الموظف عنده صلاحيته مهما كانت وحدته،
    // و"الوحدات" = بس الوحدة الي انمنحت له صراحةً.
    if (item.unitPermission && role !== 'ADMIN' && !employeePermissions.includes(item.unitPermission)) {
      return false
    }
    const granted =
      unitGranted ||
      (!!item.unitPermission && (role === 'ADMIN' || employeePermissions.includes(item.unitPermission)))
    if (!granted) {
      // الصلاحية الممنوحة فعلياً تكفي بحالها. قبل، العنصر كان يشترط الدور
      // *و* الصلاحية سوه — فلو منحت إداري الكميات صلاحية "عرض الحجوزات"
      // تضل مخفية عنه لأن دوره مو بقائمة الأدوار المسموحة. هذا خالف معنى
      // منح الصلاحية أصلاً، وكان يخلي صلاحيات كثيرة "ما تنطبق".
      const hasOwnPermission =
        (!!item.permission && employeePermissions.includes(item.permission)) ||
        (!!item.anyPermission && item.anyPermission.some((p) => employeePermissions.includes(p)))

      if (!hasOwnPermission) {
        if (item.roles && role && !item.roles.includes(role)) {
          if (!((hasMonitor || hasAudit) && item.roles.includes('MONITOR'))) return false
        }
        if (item.permission && role !== 'ADMIN' && !employeePermissions.includes(item.permission)) return false
        if (item.anyPermission && role !== 'ADMIN' && !item.anyPermission.some((p) => employeePermissions.includes(p))) return false
      }
      if (item.leaderOnly && !employee?.isLeader && role !== 'ADMIN') return false
      if (item.gpsSkillOnly && role !== 'ADMIN' && !hasGpsSkill(employee, gpsServiceId)) return false
      // الفني العادي بس ينمنع — الليدر يشوفه (نفس قيد السيرفر بالضبط)
      if (item.notForPlainTechnician && role === 'TECHNICIAN' && !employee?.isLeader) return false
    }
    if (item.children) return item.children.some((c) => isVisible(c, granted))
    return true
  }

  // نشيل أي فاصل ("── الوحدات ──") ما يتبعه ولا عنصر ظاهر — مثلاً فني عادي
  // ما عنده صلاحية توصله لأي وحدة، فيصير الفاصل معلّق بدون شي تحته.
  const visibleItems = navItems.filter((it) => isVisible(it)).filter((item, idx, arr) => {
    if (!item.divider) return true
    const next = arr[idx + 1]
    return !!next && !next.divider
  })

  const toggle = (label: string) => setExpandedGroups((p) => ({ ...p, [label]: !p[label] }))

  const gradientClass = roleColors[employee.role] || 'from-blue-500 to-indigo-600'

  // unitGranted ينتقل للأولاد: لما الموظف عنده صلاحية الوحدة، كل صفحاتها
  // تنعرض له بدون فحص صلاحياتها التفصيلية.
  const renderNavItem = (item: NavItem, depth: number = 0, unitGranted = false): React.ReactNode => {
    if (!isVisible(item, unitGranted)) return null
    const granted =
      unitGranted ||
      (!!item.unitPermission && (role === 'ADMIN' || employeePermissions.includes(item.unitPermission)))

    if (item.divider) {
      return (
        <div key={item.label} className="my-2 flex items-center gap-2 px-1">
          <span className="h-px flex-1 bg-white/20" />
          {!collapsed && <span className="text-[11px] font-bold text-white/50">{item.label}</span>}
          <span className="h-px flex-1 bg-white/20" />
        </div>
      )
    }

    if (item.children) {
      const kids = item.children.filter((c) => isVisible(c, granted))
      if (!kids.length) return null
      const open = expandedGroups[item.label]
      const active = hasActiveChild(item, location.pathname)

      if (depth === 0) {
        return (
          <div key={item.label}>
            <button
              onClick={() => toggle(item.label)}
              className={`group flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-semibold transition-all duration-200 ${
                active
                  ? 'bg-white/[0.12] text-white shadow-[inset_0_0_0_1px_rgba(255,255,255,0.1)]'
                  : 'text-blue-200/70 hover:bg-white/[0.06] hover:text-white'
              }`}
            >
              {!collapsed && (
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"
                  className={`transition-transform duration-300 ${open ? 'rotate-180' : ''}`} style={{ flexShrink: 0 }}>
                  <polyline points="6 9 12 15 18 9"/>
                </svg>
              )}
              {!collapsed && <span className="flex-1 text-right">{item.label}</span>}
              <span style={{ flexShrink: 0 }} className="opacity-80 group-hover:opacity-100 transition-opacity">{item.icon}</span>
            </button>
            {open && !collapsed && (
              <div className="mt-1 mr-4 flex flex-col gap-0.5 border-r-2 border-white/[0.08] pr-2 animate-in">
                {kids.map(child => renderNavItem(child, 1, granted))}
              </div>
            )}
          </div>
        )
      }

      return (
        <div key={item.label}>
          <button
            onClick={() => toggle(item.label)}
            className={`group flex w-full items-center gap-2 rounded-lg px-3 py-2 text-[13px] font-bold transition-all duration-200 ${
              active ? 'bg-white/[0.08] text-white' : 'text-slate-400 hover:bg-white/[0.04] hover:text-slate-200'
            }`}
          >
            <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round"
              className={`transition-transform duration-300 ${open ? 'rotate-180' : ''}`} style={{ flexShrink: 0 }}>
              <polyline points="6 9 12 15 18 9"/>
            </svg>
            <span className="flex-1 text-right">{item.label}</span>
            {active && <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 animate-pulse" style={{ flexShrink: 0 }}/>}
          </button>
          {open && (
            <div className="mr-4 flex flex-col gap-0.5 border-r border-white/[0.06] pr-2">
              {kids.map(child => renderNavItem(child, 2, granted))}
            </div>
          )}
        </div>
      )
    }

    // Leaf
    if (depth === 0) {
      return (
        <NavLink key={item.to} to={item.to} end={item.end}
          className={({ isActive }) =>
            `group relative flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-semibold transition-all duration-200 ${
              isActive
                ? 'glossy-btn bg-gradient-to-l from-[#2c5aad]/90 to-[#1e3f7a] text-white shadow-lg shadow-blue-900/30'
                : 'text-blue-200/70 hover:bg-white/[0.06] hover:text-white'
            }`
          }>
          {({ isActive }) => (
            <>
              {isActive && <span className="absolute right-0 top-1/2 -translate-y-1/2 h-6 w-1 rounded-l-full bg-white/80" />}
              {!collapsed && <span className="flex-1 text-right">{item.label}</span>}
              <span style={{ flexShrink: 0 }} className="opacity-80 group-hover:opacity-100 transition-opacity">{item.icon}</span>
            </>
          )}
        </NavLink>
      )
    }

    // Leaf مباشرة تحت مجموعة رئيسية (زي "الإدارة") — نفس شكل عناوين المجموعات
    // الشقيقة (بولد وأبيض) حتى ما توهم إنها ابن تابع لمجموعة ثانية فوقها.
    if (depth === 1) {
      return (
        <NavLink key={item.to} to={item.to} end={item.end}
          className={({ isActive }) =>
            `group flex items-center gap-2 rounded-lg px-3 py-2 text-[13px] font-bold transition-all duration-200 ${
              isActive ? 'bg-white/[0.08] text-white' : 'text-slate-300 hover:bg-white/[0.04] hover:text-white'
            }`
          }>
          {({ isActive }) => (
            <>
              <span className="flex-1 text-right">{item.label}</span>
              {isActive && <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 animate-pulse" style={{ flexShrink: 0 }}/>}
            </>
          )}
        </NavLink>
      )
    }

    return (
      <NavLink key={item.to} to={item.to} end={item.end}
        className={({ isActive }) =>
          `group relative rounded-lg px-4 py-1.5 text-right text-[12.5px] font-medium transition-all duration-200 ${
            isActive
              ? 'bg-white/[0.1] text-white font-semibold'
              : 'text-gray-500 hover:bg-white/[0.04] hover:text-gray-300'
          }`
        }>
        {({ isActive }) => (
          <>
            {isActive && <span className="absolute right-0 top-1/2 -translate-y-1/2 h-4 w-0.5 rounded-l-full bg-[#2c5aad]" />}
            {item.label}
          </>
        )}
      </NavLink>
    )
  }

  return (
    <SessionContext.Provider value={{ employee, setEmployee, permissions: employeePermissions, gpsServiceId }}>
      {/* سياسة الخصوصية: تنعرض أول دخول ولما تنضاف نقاط جديدة */}
      <PrivacyPolicyGate />
      <div dir="ltr" className="app-shell flex bg-[#f0f4f9]">

        {/* ===== Main Area ===== */}
        <div dir="rtl" className="flex min-w-0 flex-1 flex-col">
          {/* Top Header — Glass effect */}
          <header className="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-slate-200/60 bg-white/80 backdrop-blur-xl px-3 sm:px-5 lg:px-8">
            <div className="flex items-center gap-2 sm:gap-3">
              <button
                onClick={toggleSidebar}
                className="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600"
              >
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="15" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/>
                </svg>
              </button>
              <div className="hidden h-8 w-px bg-slate-200 sm:block"/>
              <span className="truncate text-sm font-extrabold text-[#0f2040] tracking-tight sm:text-lg">نظام شركة الأماني</span>
            </div>
            <div className="flex items-center gap-2 sm:gap-3">
              <div className="relative">
                <button
                  onClick={() => setSettingsOpen((o) => !o)}
                  className="flex h-10 w-10 items-center justify-center rounded-full text-slate-400 transition-all hover:bg-slate-100 hover:text-slate-600"
                  title="الإعدادات"
                >
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <circle cx="12" cy="12" r="3"/>
                    <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.6a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09A1.65 1.65 0 0 0 15 4.6a1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>
                  </svg>
                </button>
                {settingsOpen && <SettingsPanel onClose={() => setSettingsOpen(false)} />}
              </div>
              <div className="relative">
                <button
                  onClick={() => setNotifOpen((o) => !o)}
                  className="relative flex h-10 w-10 items-center justify-center rounded-xl text-slate-400 transition-all hover:bg-slate-100 hover:text-slate-600"
                >
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9M13.73 21a2 2 0 0 1-3.46 0"/>
                  </svg>
                  {unreadCount > 0 && (
                    <span className="absolute -top-0.5 -left-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-red-500 px-1 text-[10px] font-bold text-white">
                      {unreadCount > 9 ? '9+' : unreadCount}
                    </span>
                  )}
                </button>
                {notifOpen && (
                  <>
                    <div className="fixed inset-0 z-40" onClick={() => setNotifOpen(false)} />
                    <div className="absolute left-0 top-12 z-50 w-80 max-w-[90vw] overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-xl">
                      <div className="flex items-center justify-between border-b border-slate-100 px-4 py-3">
                        <span className="text-sm font-bold text-slate-800">الإشعارات</span>
                        {unreadCount > 0 && (
                          <button onClick={handleMarkAllRead} className="text-xs font-medium text-brand-600 hover:underline">تحديد الكل كمقروء</button>
                        )}
                      </div>
                      <div className="max-h-96 overflow-y-auto">
                        {notifications.length === 0 && (
                          <p className="px-4 py-6 text-center text-sm text-slate-400">ماكو إشعارات</p>
                        )}
                        {notifications.map((n) => (
                          <button
                            key={n.id}
                            onClick={() => handleNotifClick(n)}
                            className={`block w-full border-b border-slate-50 px-4 py-3 text-right text-sm transition-colors hover:bg-slate-50 ${n.read ? 'text-slate-500' : 'bg-brand-50/50 font-medium text-slate-800'}`}
                          >
                            <p>{n.message}</p>
                            <p className="mt-1 text-[11px] text-slate-400">{new Date(n.createdAt).toLocaleString('ar-IQ')}</p>
                          </button>
                        ))}
                      </div>
                    </div>
                  </>
                )}
              </div>
              <div className="flex items-center gap-2 rounded-xl bg-slate-50/80 px-2 py-1.5 transition-colors hover:bg-slate-100 sm:gap-3 sm:px-3">
                <div className="hidden text-left sm:block">
                  <p className="text-sm font-bold text-slate-800">{employee.name}</p>
                  <p className="text-[11px] text-slate-400">{roleLabels[employee.actualRole || employee.role]}</p>
                </div>
                <div className={`relative flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-xl bg-gradient-to-br ${gradientClass} text-sm font-bold text-white shadow-md sm:h-10 sm:w-10`}>
                  {employee.attendanceIcon || employee.name.charAt(0)}
                  <span className="absolute -bottom-0.5 -left-0.5 h-3 w-3 rounded-full border-2 border-white bg-emerald-500"/>
                </div>
              </div>
            </div>
          </header>

          {/* Content */}
          <main className="min-w-0 flex-1 overflow-x-hidden overflow-y-auto p-3 sm:p-5 lg:p-8">
            <Outlet />
          </main>
        </div>

        {/* Mobile backdrop */}
        {mobileOpen && (
          <div onClick={() => setMobileOpen(false)} className="fixed inset-0 z-40 bg-black/40 lg:hidden" />
        )}

        {/* ===== Right Sidebar — Premium ===== */}
        <aside
          dir="rtl"
          className={`app-sidebar glossy-dark fixed inset-y-0 right-0 z-50 flex flex-col bg-[#0f2040] transition-transform duration-300 ease-in-out lg:sticky lg:top-0 lg:z-auto lg:h-auto lg:translate-x-0 ${
            mobileOpen ? 'translate-x-0' : 'translate-x-full'
          }`}
          style={{ width: collapsed ? 72 : 270, minWidth: collapsed ? 72 : 270 }}
        >
          {/* Logo */}
          <div className={`flex items-center ${collapsed ? 'justify-center' : 'justify-center gap-3'} px-4 py-5`}>
            <div className="relative flex h-11 w-11 items-center justify-center rounded-xl bg-gradient-to-br from-[#2c5aad] to-[#1a3a6e] shadow-lg shadow-blue-900/40">
              <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
                <polyline points="3.27 6.96 12 12.01 20.73 6.96"/>
                <line x1="12" y1="22.08" x2="12" y2="12"/>
              </svg>
              <span className="absolute -top-1 -left-1 h-3 w-3 rounded-full border-2 border-[#0f2040] bg-emerald-400"/>
            </div>
            {!collapsed && (
              <div>
                <p className="text-sm font-extrabold text-white tracking-tight">الأماني</p>
                <p className="text-[10px] text-blue-300/50 font-medium">Management System</p>
              </div>
            )}
          </div>

          {/* User card */}
          {!collapsed ? (
            <div className="mx-3 mb-4 overflow-hidden rounded-2xl bg-gradient-to-br from-white/[0.08] to-white/[0.03] p-4 shadow-[inset_0_0_0_1px_rgba(255,255,255,0.06)]">
              <div className="flex items-center gap-3">
                <div className="flex-1 text-right">
                  <p className="text-sm font-bold text-white">{employee.name}</p>
                  <p className="text-[11px] text-blue-300/60">{roleLabels[employee.actualRole || employee.role]}</p>
                </div>
                <div className={`relative flex h-11 w-11 items-center justify-center rounded-xl bg-gradient-to-br ${gradientClass} text-sm font-bold text-white shadow-lg`}>
                  {employee.attendanceIcon || employee.name.charAt(0)}
                  <span className="absolute -bottom-0.5 -left-0.5 h-3 w-3 rounded-full border-2 border-[#0f2040] bg-emerald-400"/>
                </div>
              </div>
              <div className="mt-3 flex items-center gap-2">
                <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 animate-pulse"/>
                <span className="text-[10px] text-emerald-400/80 font-medium">متصل الآن</span>
                <span className={`mr-auto rounded-full bg-gradient-to-l ${gradientClass} px-2.5 py-0.5 text-[10px] font-bold text-white shadow-sm`}>
                  {roleLabels[employee.actualRole || employee.role]}
                </span>
              </div>
            </div>
          ) : (
            <div className="mx-auto mb-3">
              <div className={`relative flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br ${gradientClass} text-sm font-bold text-white shadow-lg`}>
                {employee.attendanceIcon || employee.name.charAt(0)}
                <span className="absolute -bottom-0.5 -left-0.5 h-2.5 w-2.5 rounded-full border-2 border-[#0f2040] bg-emerald-400"/>
              </div>
            </div>
          )}

          {/* Divider */}
          <div className="mx-4 mb-2 h-px bg-gradient-to-l from-transparent via-white/10 to-transparent"/>

          {/* Nav */}
          <nav className="flex flex-1 flex-col gap-0.5 overflow-y-auto px-3 pb-3 scrollbar-thin">
            {visibleItems.map(item => renderNavItem(item, 0))}
            {employee.actualRole === 'OWNER' && (
              <NavLink
                to="/owner-security"
                className={({ isActive }) =>
                  `mt-1 flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors ${
                    isActive ? 'bg-white/10 text-white' : 'text-white/60 hover:bg-white/5 hover:text-white'
                  }`
                }
              >
                👁️ لوحة المراقبة الخلفية
              </NavLink>
            )}
            {employee.actualRole === 'OWNER' && (
              <NavLink
                to="/assistant-conversations"
                className={({ isActive }) =>
                  `mt-1 flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors ${
                    isActive ? 'bg-white/10 text-white' : 'text-white/60 hover:bg-white/5 hover:text-white'
                  }`
                }
              >
                💬 محادثات المساعد الذكي
              </NavLink>
            )}
          </nav>

          {/* سياسة الخصوصية — قراءة متاحة لكل موظف بدون أي صلاحية، ثابتة فوق
              زر تسجيل الخروج. إضافة/تعديل النقاط شي ثاني تماماً: صلاحية
              "privacy_policy_manage" وتظهر داخل وحدة الرقابة. */}
          <div className="mx-3 mt-1">
            <NavLink
              to="/privacy-policy"
              className={({ isActive }) =>
                `flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors ${
                  isActive ? 'bg-white/10 text-white' : 'text-white/60 hover:bg-white/5 hover:text-white'
                }`
              }
            >
              {collapsed ? '🔒' : '🔒 سياسة الخصوصية'}
            </NavLink>
          </div>

          {/* Logout */}
          <div className="mx-3 mb-3 mt-1">
            <div className="h-px bg-gradient-to-l from-transparent via-white/10 to-transparent mb-3"/>
            <button onClick={() => setEmployee(null)}
              className="group flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium text-red-300/50 transition-all duration-200 hover:bg-red-500/10 hover:text-red-300">
              {!collapsed && <span className="flex-1 text-right">تسجيل الخروج</span>}
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" className="opacity-60 group-hover:opacity-100 transition-opacity">
                <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4M16 17l5-5-5-5M21 12H9"/>
              </svg>
            </button>
          </div>
        </aside>
      </div>
      {(employee?.role === 'ADMIN' || employee?.role === 'MONITOR' || employee?.actualRole === 'OWNER')
        ? <ManagerAssistantChat />
        : <AssistantWidget />}
    </SessionContext.Provider>
  )
}
