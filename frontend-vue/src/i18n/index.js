import { createI18n } from 'vue-i18n'  // Vue3
// 对于Vue2，请从 'vue-i18n' 中使用 new VueI18n()
import en from './en'
import zh from './zh'

const i18n = createI18n({
  locale: 'zh', // 设置默认语言
  fallbackLocale: 'en', // 当当前语言无对应翻译时使用的备用语言
  messages: {
    en,
    zh
  }
});

export default i18n;
