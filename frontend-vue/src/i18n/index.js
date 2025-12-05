import { createI18n } from 'vue-i18n'
import en from './en'
import zh from './zh'
import jp from './jp'
import ru from './ru'
import es from './es'
import ms from './ms'
import ta from './ta'

// 支持的语言列表
const supportedLocales = ['en', 'zh', 'jp', 'ru', 'es', 'ms', 'ta']

// 检测浏览器语言并匹配支持的语言
function detectLocale() {
  // 优先从 localStorage 读取用户之前的选择
  const savedLocale = localStorage.getItem('locale')
  if (savedLocale && supportedLocales.includes(savedLocale)) {
    return savedLocale
  }

  // 获取浏览器语言
  const browserLang = navigator.language || navigator.userLanguage || 'zh'

  // 语言代码映射（处理各种变体）
  const langMap = {
    'zh': 'zh', 'zh-CN': 'zh', 'zh-TW': 'zh', 'zh-HK': 'zh',
    'en': 'en', 'en-US': 'en', 'en-GB': 'en', 'en-AU': 'en',
    'ja': 'jp', 'ja-JP': 'jp',
    'ru': 'ru', 'ru-RU': 'ru',
    'es': 'es', 'es-ES': 'es', 'es-MX': 'es',
    'ms': 'ms', 'ms-MY': 'ms',
    'ta': 'ta', 'ta-IN': 'ta'
  }

  // 精确匹配
  if (langMap[browserLang]) {
    return langMap[browserLang]
  }

  // 前缀匹配（如 'en-AU' -> 'en'）
  const langPrefix = browserLang.split('-')[0]
  if (supportedLocales.includes(langPrefix)) {
    return langPrefix
  }
  if (langMap[langPrefix]) {
    return langMap[langPrefix]
  }

  // 默认英文
  return 'en'
}

const i18n = createI18n({
  locale: detectLocale(),
  fallbackLocale: 'en',
  messages: {
    en,
    zh,
    jp,
    ru,
    es,
    ms,
    ta
  }
})

export default i18n
