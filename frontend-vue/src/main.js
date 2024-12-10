// main.js
import { createApp } from 'vue'
import App from './App.vue'

// 引入路由
import router from './router/index.js'
import i18n from './i18n'

// 引入Vuetify的核心和样式
import 'vuetify/styles' 
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
// ... 其他导入
import '@mdi/font/css/materialdesignicons.css'

// 创建Vuetify实例
const vuetify = createVuetify({
  components,
  directives,
})

// 挂载Vue实例，并使用路由和Vuetify
createApp(App)
  .use(router)    // 这里添加路由的使用
  .use(vuetify)
  .use(i18n)
  .mount('#app')
