// main.js
import { createApp } from 'vue';
import App from './App.vue';
import router from './router/index.js'; // 引入路由
import i18n from './i18n';
import { getConfig } from './config/index.js'; // 引入异步配置加载函数

// 引入Vuetify的核心和样式
import 'vuetify/styles';
import { createVuetify } from 'vuetify';
import * as components from 'vuetify/components';
import * as directives from 'vuetify/directives';
import '@mdi/font/css/materialdesignicons.css';

// 创建Vuetify实例
const vuetify = createVuetify({
  components,
  directives,
});

// 异步初始化函数
async function initializeApp() {
  try {
    // 等待配置加载完成
    const config = await getConfig();
    console.log('App initializing with config:', config); // 确认配置已加载

    // 配置加载成功后，创建并挂载Vue实例
    const app = createApp(App);
    app.use(router);    // 使用路由
    app.use(vuetify);
    app.use(i18n);

    // 将配置注入到全局属性，方便组件访问 (可选)
    // app.config.globalProperties.$config = config;

    app.mount('#app');
  } catch (error) {
    console.error('Application initialization failed:', error);
    // 可以在这里显示一个错误消息给用户
    document.getElementById('app').innerHTML = 'Failed to initialize application. Please check configuration or contact support.';
  }
}

// 执行异步初始化
initializeApp();
