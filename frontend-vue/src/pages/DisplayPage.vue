<!-- QuestionDisplay.vue -->
<template>
  <div class="page-root">
    <div class="cyber-line"></div>
    <div class="container">
      <header class="header">
        <h1>{{ $t('display.title') }}</h1>
        <LanguageSwitcher />
      </header>
      
      <div 
        class="current-question" 
        :class="{ loading: !currentQuestion }"
      >
        {{ currentQuestion ? currentQuestion.content : $t('display.waiting') }}
      </div>

      <div class="question-list">
        <div 
          v-for="(q, index) in pendingQuestions" 
          :key="index"
          class="question-card animate__animated animate__fadeInUp"
        >
          <div class="question-content">{{ q.content }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import { getConfig } from '../config/index.js'; // 导入具名函数
import LanguageSwitcher from '../components/LanguageSwitcher.vue';

const { t } = useI18n();

const loadedConfig = ref(null); // 用于存储加载的配置
const questions = ref([]);
const sessionId = ref('');

// 获取 URL 中的 sessionId
watchEffect(() => {
  const urlParams = new URLSearchParams(window.location.hash.split('?')[1]);
  sessionId.value = urlParams.get('sessionId');
});

const currentQuestion = computed(() => {
  if (!questions.value) return null;
  return questions.value.find(q => q.status === 'showing');
});

const pendingQuestions = computed(() => {
  if (!questions.value) return [];
  return questions.value.filter(q => q.status === 'pending');
});

let ws = null;
let intervalId = null;

// Helper function to get API endpoint, ensuring config is loaded
function getApiEndpoint() {
  if (!loadedConfig.value || !loadedConfig.value.api || !loadedConfig.value.api.endpoint) {
    console.error('API Configuration not loaded yet.');
    throw new Error('API configuration is not available.');
  }
  return loadedConfig.value.api.endpoint;
}

// Helper function to get WebSocket endpoint, ensuring config is loaded
function getWsEndpoint() {
  if (!loadedConfig.value || !loadedConfig.value.ws || !loadedConfig.value.ws.endpoint) {
    console.error('WebSocket Configuration not loaded yet.');
    throw new Error('WebSocket configuration is not available.');
  }
  return loadedConfig.value.ws.endpoint;
}


async function loadQuestions() {
  try {
    if (!sessionId.value || !loadedConfig.value) { // 确保配置已加载
      console.warn('No sessionId provided or config not loaded');
      questions.value = [];
      return;
    }
    const apiEndpoint = getApiEndpoint(); // 使用辅助函数获取端点
    const response = await fetch(`${apiEndpoint}/questions/${sessionId.value}`);
    if (!response.ok) {
      console.error('Failed to load questions:', response.status);
      questions.value = [];
      return;
    }
    const data = await response.json();
    questions.value = data || [];
  } catch (error) {
    console.error('Loading questions failed:', error);
    questions.value = [];
  }
}

function initializeWebSocket() {
  if (ws) { // Close existing connection if any
      ws.close();
  }
  try {
    const wsEndpoint = getWsEndpoint(); // 使用辅助函数获取端点
    console.log(`Initializing WebSocket connection to: ${wsEndpoint}`);
    ws = new WebSocket(wsEndpoint);
    ws.onopen = () => {
        console.log('WebSocket connection established.');
        // Optionally send session ID or other info upon connection
        // ws.send(JSON.stringify({ type: 'register', sessionId: sessionId.value }));
    };
    ws.onmessage = (event) => {
        console.log('WebSocket message received:', event.data);
        // Assuming the message indicates an update is needed
        loadQuestions();
    };
    ws.onerror = (error) => {
        console.error('WebSocket error:', error);
    };
    ws.onclose = (event) => {
        console.log('WebSocket connection closed:', event.code, event.reason);
        ws = null; // Reset ws variable
        // Optionally attempt to reconnect after a delay
        // setTimeout(initializeWebSocket, 5000);
    };
  } catch (error) {
      console.error("Failed to initialize WebSocket:", error);
  }
}

async function initializeApp() {
  if (!loadedConfig.value) {
      console.error("Cannot initialize app, config not loaded.");
      return;
  }
  // Initial load
  await loadQuestions(); // Await initial load

  // Initialize WebSocket connection only after config is loaded
  initializeWebSocket();

  // Periodic refresh (consider if WS replaces the need for polling)
  if (intervalId) clearInterval(intervalId); // Clear previous interval if any
  intervalId = setInterval(loadQuestions, 5000); // Re-establish interval
}

onMounted(async () => { // Make onMounted async
  try {
    loadedConfig.value = await getConfig(); // Load config first
    console.log('DisplayPage config loaded:', loadedConfig.value);
    await initializeApp(); // Initialize app after config is loaded
  } catch (error) {
    console.error("Failed to load configuration in DisplayPage:", error);
    // Handle config load error (e.g., show error message)
  }

  // Add these meta tags to ensure proper viewport behavior
  const meta = document.createElement('meta');
  meta.name = 'viewport'
  meta.content = 'width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no'
  document.head.appendChild(meta)
  
  // Prevent scrolling on mobile devices
  document.body.style.overflow = 'hidden'
  document.documentElement.style.overflow = 'hidden'
})

onUnmounted(() => {
  if (ws) {
    ws.close()
    ws = null
  }
  if (intervalId) {
    clearInterval(intervalId)
    intervalId = null
  }
  
  // Clean up added styles
  document.body.style.overflow = ''
  document.documentElement.style.overflow = ''
})
</script>

<style>
/* Reset CSS */
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

:root {
  --primary-bg: #1a1c2e;
  --secondary-bg: #2a2d4a;
  --accent-color: #6e7dff;
  --text-primary: #ffffff;
  --text-secondary: #b3b9ff;
  --card-bg: rgba(42, 45, 74, 0.8);
  --border-color: rgba(110, 125, 255, 0.2);
  --glow-color: rgba(110, 125, 255, 0.5);
}

html, body {
  height: 100%;
  width: 100%;
  margin: 0;
  padding: 0;
  overflow: hidden;
  position: fixed;
  background: var(--primary-bg);
}

#app {
  height: 100%;
  width: 100%;
}
</style>

<style scoped>
.page-root {
  height: 100vh;
  width: 100vw;
  background: var(--primary-bg);
  position: fixed;
  top: 0;
  left: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.container {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 2rem;
  gap: 2rem;
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
  position: relative;
  background-image: 
    radial-gradient(circle at 10% 20%, rgba(110, 125, 255, 0.1) 0%, transparent 50%),
    radial-gradient(circle at 90% 80%, rgba(110, 125, 255, 0.1) 0%, transparent 50%);
  overflow: hidden;
}

.header {
  text-align: center;
  padding: 1rem;
  background: var(--secondary-bg);
  border-radius: 1rem;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
  animation: slideInDown 0.5s ease-out;
  flex-shrink: 0;
  position: relative;
  display: flex;
  justify-content: center;
  align-items: center;
}

.header h1 {
  font-size: 1.5rem;
  color: var(--accent-color);
  text-transform: uppercase;
  letter-spacing: 2px;
  margin: 0;
}

.current-question {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: clamp(2rem, 5vw, 4rem);
  text-align: center;
  padding: 2rem;
  background: var(--card-bg);
  border-radius: 1.5rem;
  border: 1px solid var(--border-color);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
  margin: 1rem 0;
  animation: fadeIn 0.5s ease-out;
  transition: all 0.3s ease;
  position: relative;
  color: var(--text-primary);
  min-height: 200px;
}

.current-question:not(.loading) {
  box-shadow: 0 0 20px var(--glow-color), 0 0 40px var(--glow-color);
}

.current-question.loading::after {
  content: '';
  animation: dots 1.5s steps(4, end) infinite;
}

.question-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 0.8rem;
  padding: 1rem;
  background: rgba(26, 28, 46, 0.9);
  border-radius: 1rem;
  height: 30vh;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: var(--accent-color) var(--primary-bg);
  animation: fadeInUp 0.5s ease-out;
  flex-shrink: 0;
}

.question-list::-webkit-scrollbar {
  width: 8px;
}

.question-list::-webkit-scrollbar-track {
  background: var(--primary-bg);
}

.question-list::-webkit-scrollbar-thumb {
  background: var(--accent-color);
  border-radius: 4px;
}

.question-card {
  background: var(--card-bg);
  padding: 0.8rem;
  border-radius: 0.8rem;
  border: 1px solid var(--border-color);
  transition: transform 0.3s ease, box-shadow 0.3s ease;
  font-size: 1rem;
  line-height: 1.4;
  color: var(--text-primary);
  max-height: 100px;
  overflow: hidden;
  display: flex;
  align-items: center;
}

.question-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 20px var(--glow-color);
}

.question-content {
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  word-break: break-word;
  margin: 0;
}

.cyber-line {
  position: fixed;
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--accent-color), transparent);
  width: 100%;
  animation: scan 8s linear infinite;
  opacity: 0.5;
  z-index: 100;
}

@media (max-width: 768px) {
  .container {
    padding: 1rem;
  }

  .current-question {
    font-size: clamp(1.5rem, 4vw, 2.5rem);
    padding: 1rem;
    min-height: 150px;
  }

  .question-list {
    grid-template-columns: 1fr;
    height: 40vh;
    gap: 0.6rem;
    padding: 0.8rem;
  }

  .question-card {
    padding: 0.6rem;
    max-height: 80px;
    font-size: 0.9rem;
  }

  .question-content {
    -webkit-line-clamp: 2;
  }
}

@keyframes dots {
  0%, 20% { content: ''; }
  40% { content: '.'; }
  60% { content: '..'; }
  80% { content: '...'; }
  100% { content: ''; }
}

@keyframes scan {
  0% { top: -10%; }
  100% { top: 110%; }
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes fadeInUp {
  from { 
    opacity: 0;
    transform: translateY(20px);
  }
  to { 
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes slideInDown {
  from {
    transform: translateY(-100%);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

.language-switcher {
  position: absolute;
  right: 1rem;
  top: 50%;
  transform: translateY(-50%);
}

.lang-menu-btn {
  color: var(--text-primary);
  background: var(--card-bg);
  border: 1px solid var(--border-color);
}
</style>