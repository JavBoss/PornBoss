import { zh } from '@/utils/i18n'

export function getJavDisplayTitle(item, javMetadataLanguage = 'zh') {
  const code = item?.code?.trim()
  const title = javMetadataLanguage === 'en' ? item?.title_en || item?.title : item?.title
  return title || code || zh('未知标题', 'Untitled')
}
