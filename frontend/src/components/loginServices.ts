export type ServiceKind =
  | 'audio'
  | 'fire'
  | 'gps'
  | 'technical'
  | 'camera'
  | 'network'
  | 'solar'
  | 'home'
  | 'lock'

export interface ServiceItem {
  title: string
  desc: string
  kind: ServiceKind
  color: string
}

export const RIGHT_SERVICES: ServiceItem[] = [
  { title: 'أنظمة الصوتيات', desc: 'صوت نقي وتجربة احترافية', kind: 'audio', color: '#d946ef' },
  { title: 'أنظمة الحريق والسلامة', desc: 'سلامة موثوقة وحماية شاملة', kind: 'fire', color: '#ff4438' },
  { title: 'أنظمة تتبع المركبات GPS', desc: 'تتبع لحظي وأمان كامل', kind: 'gps', color: '#55e86a' },
  { title: 'الخدمات الفنية والحلول الذكية', desc: 'حلول مبتكرة لدعم أعمالك', kind: 'technical', color: '#ff9f32' },
]

export const LEFT_SERVICES: ServiceItem[] = [
  { title: 'أنظمة المراقبة والكاميرات', desc: 'مراقبة ذكية على مدار الساعة', kind: 'camera', color: '#36d9ff' },
  { title: 'الشبكات والبنية التحتية', desc: 'اتصالات مستقرة وآمنة', kind: 'network', color: '#27d7ff' },
  { title: 'الطاقة الشمسية', desc: 'طاقة نظيفة ومستدامة', kind: 'solar', color: '#ffd64b' },
  { title: 'المنزل الذكي', desc: 'تحكم ذكي لحياة أكثر راحة', kind: 'home', color: '#9e6cff' },
  { title: 'الأقفال الذكية', desc: 'أمان متطور وتحكم ذكي', kind: 'lock', color: '#c56cff' },
]
