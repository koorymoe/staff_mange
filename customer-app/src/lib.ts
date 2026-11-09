// ثوابت ودوال مشتركة (تنسيق الأرقام والتواريخ، المحافظات)

export const SHORT: Record<string, string> = {
  PENDING: 'فعّالة', CONFIRMED: 'مجدولة', WAITING: 'بالانتظار', IN_PROGRESS: 'قيد التنفيذ', PARTIAL: 'قيد التنفيذ', COMPLETED: 'مكتملة', CANCELLED: 'ملغاة',
}

export const iqd = (n: number) => `${Math.round(n).toLocaleString('en-US')} د.ع`

export const fmtDate = (s?: string | null) => (s ? new Date(s).toLocaleDateString('en-CA').replaceAll('-', '/') : '')

export const fmtTime = (s?: string | null) => (s ? new Date(s).toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' }) : '')

export const CITIES = ['بغداد', 'كربلاء', 'النجف', 'بابل', 'البصرة', 'واسط', 'الديوانية', 'ذي قار', 'ميسان', 'المثنى', 'ديالى', 'الأنبار', 'صلاح الدين', 'كركوك', 'نينوى', 'أربيل', 'السليمانية', 'دهوك']
