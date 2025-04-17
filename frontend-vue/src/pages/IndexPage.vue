<!-- src/pages/QuestionPage.vue -->
<template>
  <div class="page-root">
    <!-- 樱花背景层 -->
    <div class="sakura-background">
      <div
        v-for="sakura in sakuras"
        :key="sakura.id"
        class="sakura"
        :style="sakura.style"
      >
        🌸
      </div>
    </div>

    <!-- 主内容区域 -->
    <div class="page-wrapper">
      <div class="content-container">
        <div class="container animate__animated animate__fadeIn">
          <h1 class="text-center animate__animated animate__bounceIn">
             {{ $t('nav.Questionshere') }}
          </h1>
          <LanguageSwitcher />

          <!-- 如果配置加载出现错误 -->
          <div v-if="configError" class="status error animate__animated animate__fadeInUp">
             {{ $t('message.configError') }}
          </div>
          <!-- 如果配置尚未加载完成 (可选，可以在此处放一个加载中状态) -->
          <div v-else-if="!config" class="status loading animate__animated animate__fadeInUp">
             {{ $t('message.loading') }}
          </div>

          <!-- 问题输入框和提交按钮，只有在配置加载成功后才显示 -->
          <div v-else class="question-form">
            <v-textarea
              v-model="question"
              :placeholder="$t('message.placeholder')"
              variant="outlined"
              :disabled="isSubmitting"
              rows="5"
              class="animate__animated animate__fadeInUp question-textarea"
              hide-details
              @keydown.ctrl.enter="submitQuestion"
            ></v-textarea>

            <v-btn
              :loading="isSubmitting"
              :disabled="isSubmitting"
              color="primary"
              size="large"
              class="submit-btn animate__animated animate__fadeInUp"
              @click="submitQuestion"
            >
              {{ $t('button.send') }}
            </v-btn>
          </div>

          <!-- Status Message -->
          <transition name="fade">
            <div
              v-if="statusMessage"
              class="status animate__animated animate__fadeInUp"
              :class="[statusType]"
            >
              {{ statusMessage }}
            </div>
          </transition>
        </div>

        <!-- Mascot -->
        <div 
          ref="mascot"
          class="mascot"
        >
          <v-img
            src="/api/placeholder/150/150"
            width="150"
            height="150"
          ></v-img>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watchEffect } from 'vue'; // Added watchEffect
// import jsyaml from 'js-yaml'; // No longer needed
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import LanguageSwitcher from '../components/LanguageSwitcher.vue';
import { getConfig } from '../config/index.js'; // Import the correct config loader

const question = ref('');
const statusMessage = ref('');
const statusType = ref('');
const isSubmitting = ref(false);
const mascot = ref(null);
const sakuras = ref([]);
let sakuraId = 0;
let sakuraInterval = null;

// --- Configuration State ---
const loadedConfig = ref(null); // Store the loaded config
const configError = ref(null); // Store potential config loading errors
const sessionId = ref(null); // Session ID from query params
// --- End Configuration State ---

const route = useRoute();
const { t } = useI18n();

// Helper function to get API endpoint, ensuring config is loaded
function getApiEndpoint() {
  if (!loadedConfig.value || !loadedConfig.value.api || !loadedConfig.value.api.endpoint) {
    console.error('API Configuration not loaded yet.');
    // Set error state or throw, depending on desired behavior
    configError.value = t('message.configError'); // Use i18n key
    throw new Error('API configuration is not available.');
  }
  // Clear config error if we successfully get the endpoint
  configError.value = null;
  return loadedConfig.value.api.endpoint;
}

// --- Status Display Logic ---
function showStatus(messageKey, type) { // Pass i18n key directly
  statusMessage.value = t(messageKey); // Translate the key
  statusType.value = type;

  if (type !== 'loading') {
    setTimeout(() => {
      statusMessage.value = '';
      statusType.value = '';
    }, 3000);
  }
}
// --- End Status Display Logic ---

// --- Submit Question Logic ---
async function submitQuestion() { // Make async to handle potential errors from getApiEndpoint
  const content = question.value.trim();
  if (!content || isSubmitting.value) return; // Prevent empty or double submission

  // Check if config is loaded and session ID exists
  if (!loadedConfig.value) {
      showStatus('message.configError', 'error');
      return;
  }
   if (!sessionId.value) {
      showStatus('message.noSessionIdError', 'error'); // Add a specific i18n key for this
      return;
  }

  isSubmitting.value = true;
  const currentQuestion = question.value; // Store current question before clearing
  question.value = ''; // Clear input immediately
  showStatus('message.submitting', 'loading');

  const data = {
    sessionId: sessionId.value,
    content: content
  };

  try {
    const apiEndpoint = getApiEndpoint(); // Get endpoint (might throw if config not ready)

    // Use fetch API
    const response = await fetch(`${apiEndpoint}/question`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(data)
    });

    if (response.ok) { // Check if response status is 2xx
      showStatus('message.success', 'success');
    } else {
      // Handle non-2xx responses (e.g., 400, 500)
      console.error('Submission failed with status:', response.status);
      showStatus('message.submitError', 'error');
      question.value = currentQuestion; // Restore input on error
    }
  } catch (error) {
    // Handle fetch errors (network issues, config loading errors from getApiEndpoint)
    console.error('Submission fetch error:', error);
    showStatus('message.submitError', 'error');
    question.value = currentQuestion; // Restore input on error

    // Fallback with sendBeacon (optional, consider if necessary)
    // Note: sendBeacon doesn't provide response feedback.
    // try {
    //   const apiEndpoint = getApiEndpoint(); // Get endpoint again for beacon
    //   navigator.sendBeacon(
    //     `${apiEndpoint}/question`,
    //     JSON.stringify(data)
    //   );
    //   // Assume success for beacon, as we don't get feedback
    //   // showStatus('message.success', 'success'); // Or maybe a different message?
    // } catch (beaconError) {
    //   console.error("Beacon fallback failed:", beaconError);
    // }
  } finally {
    isSubmitting.value = false; // Ensure loading state is reset
  }
}
// --- End Submit Question Logic ---


// --- Sakura Effect ---
function createSakura() {
  const id = sakuraId++;
  const duration = Math.random() * 3 + 2; // 2-5秒的随机时长

  const style = {
    left: `${Math.random() * window.innerWidth}px`,
    top: '-20px',
    fontSize: `${Math.random() * 20 + 10}px`,
    opacity: Math.random() * 0.6 + 0.4,
    animation: `falling ${duration}s linear forwards`
  };

  sakuras.value.push({ id, style });

  setTimeout(() => {
    sakuras.value = sakuras.value.filter(s => s.id !== id);
  }, duration * 1000);
}
// --- End Sakura Effect ---


// --- Lifecycle Hooks ---
onMounted(async () => {
  // Get session ID from route query params
  sessionId.value = route.query.sessionId;
  if (!sessionId.value) {
      console.error("Missing sessionId in URL query parameters.");
      configError.value = t('message.noSessionIdError'); // Show error if no session ID
  }

  // Load configuration
  try {
    loadedConfig.value = await getConfig();
    console.log('IndexPage config loaded:', loadedConfig.value);
    configError.value = null; // Clear error on success
  } catch (error) {
    console.error("Failed to load configuration in IndexPage:", error);
    configError.value = t('message.configError'); // Show config error
  }

  // Start Sakura effect only if config loaded successfully (or regardless?)
  sakuraInterval = setInterval(createSakura, 500);
});

onBeforeUnmount(() => {
  if (sakuraInterval) {
    clearInterval(sakuraInterval);
  }
});
// --- End Lifecycle Hooks ---
</script>

<style scoped>
.page-root {
  min-height: 100vh;
  width: 100%;
  position: relative;
  background: linear-gradient(140deg, #fdfbfb 0%, #ebedee 100%);
  overflow: hidden;
}

.sakura-background {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100vh;
  pointer-events: none;
  z-index: 1;
  overflow: hidden;
}

.sakura {
  position: absolute;
  z-index: 1;
  pointer-events: none;
  user-select: none;
}

.page-wrapper {
  position: relative;
  z-index: 2;
  min-height: 100vh;
}

.content-container {
  position: relative;
  padding: 20px;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.container {
  max-width: 800px;
  width: 95%;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 20px;
  padding: 30px;
  box-shadow: 0 10px 30px rgba(255, 105, 180, 0.15);
  backdrop-filter: blur(5px);
  margin: 20px auto;
  position: relative;
  z-index: 2;
}

h1 {
  color: #FF69B4;
  font-size: 2.5em;
  margin-bottom: 30px;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.1);
  text-align: center;
}

.question-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.question-textarea {
  background: rgba(255, 255, 255, 0.9) !important;
}

.question-textarea :deep(.v-field__outline) {
  border-color: #FFB7C5 !important;
}

.question-textarea :deep(.v-field--focused .v-field__outline) {
  border-color: #FF69B4 !important;
}

.submit-btn {
  align-self: stretch;
  height: 50px !important;
  border-radius: 25px !important;
  background: #FF69B4 !important;
  transition: all 0.3s ease !important;
}

.submit-btn:hover {
  background: #FF1493 !important;
  transform: translateY(-2px);
  box-shadow: 0 5px 15px rgba(255, 105, 180, 0.3) !important;
}

.status {
  margin-top: 20px;
  padding: 15px;
  border-radius: 15px;
  text-align: center;
  font-weight: bold;
}

.success {
  background: rgba(212, 237, 218, 0.9);
  color: #155724;
}

.error {
  background: rgba(248, 215, 218, 0.9);
  color: #721c24;
}

.loading {
  background: rgba(255, 255, 255, 0.9);
  color: #4A4A4A;
  display: flex;
  justify-content: center;
  align-items: center;
}

.loading::after {
  content: '';
  display: inline-block;
  width: 20px;
  height: 20px;
  margin-left: 10px;
  border: 2px solid #FF69B4;
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.mascot {
  position: fixed;
  bottom: 20px;
  right: 20px;
  z-index: 1000;
  transition: transform 0.3s ease;
}

@keyframes falling {
  0% {
    transform: translate(0, -20px) rotate(0deg);
    opacity: 1;
  }
  100% {
    transform: translate(0, 100vh) rotate(360deg);
    opacity: 0.6;
  }
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* 响应式设计 */
@media (max-width: 768px) {
  .container {
    padding: 20px;
    width: 100%;
  }

  h1 {
    font-size: 2em;
  }

  .mascot {
    width: 100px !important;
    height: 100px !important;
    bottom: 10px;
    right: 10px;
  }
}

/* 过渡动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* 添加语言切换按钮样式 */
.language-switcher {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 1000;
  display: flex;
  gap: 10px;
}

.lang-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.9);
  color: #FF69B4;
  cursor: pointer;
  transition: all 0.3s ease;
  font-weight: bold;
  backdrop-filter: blur(5px);
  box-shadow: 0 2px 10px rgba(255, 105, 180, 0.2);
}

.lang-btn:hover {
  background: #FF69B4;
  color: white;
  transform: translateY(-2px);
  box-shadow: 0 4px 15px rgba(255, 105, 180, 0.3);
}

/* 添加响应式样式 */
@media (max-width: 768px) {
  .language-switcher {
    top: 10px;
    right: 10px;
  }
  
  .lang-btn {
    padding: 6px 12px;
    font-size: 14px;
  }
}
</style>
