import TabsShell from '../components/TabsShell'
import ServiceStudiesPage from './ServiceStudiesPage'
import TrainingManagement from './TrainingManagement'
import TechShowcasePage from './TechShowcasePage'

// «وهذن بواجهة ثانية».
//
// الثلاثة محتوى تقني يُكتب ويُنشر: دراسة الخدمة، ومفردة التدريب،
// وشغل ينعرض بالمعرض — نفس صاحب الشغل ونفس الجلسة.
export default function TechContentHub() {
  return (
    <TabsShell
      title="📚 المحتوى التقني"
      subtitle="دراسات الخدمات ومفردات التدريب ومعرض الأعمال"
      tabs={[
        { key: 'studies', label: 'دراسات الخدمات', icon: '🧪', render: () => <ServiceStudiesPage embedded /> },
        { key: 'training', label: 'مفردات التدريب', icon: '🎓', render: () => <TrainingManagement embedded /> },
        { key: 'showcase', label: 'معرض الأعمال', icon: '🖼️', render: () => <TechShowcasePage embedded /> },
      ]}
    />
  )
}
