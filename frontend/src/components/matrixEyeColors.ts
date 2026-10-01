// ألوان وتسميات عيون ماتركس — مشتركة بين عين الهيدر وعيون المدير.

export type EyeGroup = 'ADMINS' | 'MONITORS' | 'COORDINATORS' | 'FINANCE' | 'FIELD' | 'QUALITY' | 'IT' | 'STAFF'
export type EyeMood = 'CALM' | 'PLEASED' | 'ALERT' | 'ANGRY'

export const GROUP_COLOR: Record<EyeGroup, string> = {
  ADMINS: '#f5b301',
  MONITORS: '#a855f7',
  COORDINATORS: '#22d3ee',
  FINANCE: '#10b981',
  FIELD: '#fb923c',
  QUALITY: '#38bdf8',
  IT: '#3b82f6',
  STAFF: '#818cf8',
}

export const GROUP_LABEL: Record<EyeGroup, string> = {
  ADMINS: 'المدراء',
  MONITORS: 'المراقبين والمدققين',
  COORDINATORS: 'الإداريين والتنسيق',
  FINANCE: 'المحاسبين',
  FIELD: 'الليدرية والفنيين',
  QUALITY: 'الجودة',
  IT: 'تقنية المعلومات',
  STAFF: 'بقية الموظفين',
}

export const MOOD_COLOR: Record<Exclude<EyeMood, 'CALM'>, string> = {
  PLEASED: '#4ade80',
  ALERT: '#facc15',
  ANGRY: '#ef4444',
}

export function eyeColor(group: EyeGroup, mood: EyeMood): string {
  return mood === 'CALM' ? GROUP_COLOR[group] ?? GROUP_COLOR.STAFF : MOOD_COLOR[mood]
}

