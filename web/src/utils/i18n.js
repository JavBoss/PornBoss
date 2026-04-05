const primaryLanguage = (() => {
  if (typeof navigator !== 'undefined') {
    if (Array.isArray(navigator.languages) && navigator.languages.length > 0) {
      return navigator.languages[0]
    }
    if (navigator.language) {
      return navigator.language
    }
  }
  if (typeof document !== 'undefined') {
    const lang = document.documentElement?.lang
    if (lang) {
      return lang
    }
  }
  return ''
})()

const prefersChinese = /^zh\b/i.test(String(primaryLanguage || '').trim())

export function isChineseLocale() {
  return prefersChinese
}

export function zh(cn, en) {
  return prefersChinese ? cn : en
}
