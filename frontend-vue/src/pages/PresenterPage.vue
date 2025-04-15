<template>
  <div class="container">
    <div class="header">
      <h1>{{ $t('presenter.title') }}</h1>
      <LanguageSwitcher />
    </div>

    <!-- 文档管理区域 -->
    <div class="document-management-section">
      <!-- 文档上传 -->
      <div class="document-upload-section">
        <h2>{{ $t('presenter.uploadTitle') }}</h2>
        <form @submit.prevent="handleDocumentUpload">
          <input type="file" ref="fileInputRef" accept=".pdf,.docx,.txt" required>
          <button type="submit" class="btn btn-primary">{{ $t('presenter.uploadButton') }}</button>
        </form>
        <div id="uploadStatus" :class="uploadStatusClass">{{ uploadStatus }}</div>
      </div>

      <!-- 已上传文档列表 -->
      <div class="document-list-section">
        <h2>{{ $t('presenter.uploadedDocsTitle') }}</h2>
        <div id="documentList">
          <p v-if="loadingDocuments">{{ $t('presenter.loadingDocs') }}</p>
          <ul v-else-if="uploadedDocuments.length > 0">
            <li v-for="doc in uploadedDocuments" :key="doc.id">
              <span class="doc-title" :title="`${doc.title}.${doc.fileType}`">{{ doc.title }}.{{ doc.fileType }}</span>
              <span class="doc-meta">{{ $t('presenter.uploadedAt') }}: {{ formatDateTime(doc.uploadTime) }}</span>
              <button class="btn btn-danger btn-delete-doc" @click="handleDeleteDocument(doc.id, `${doc.title}.${doc.fileType}`)">
                {{ $t('presenter.delete') }}
              </button>
            </li>
          </ul>
          <p v-else>{{ $t('presenter.noDocs') }}</p>
        </div>
      </div>
    </div>

    <!-- 提示词编辑区域 -->
    <div class="prompt-editing-section">
      <h2>{{ $t('presenter.promptSettingsTitle') }}</h2>
      <div class="prompt-editor">
        <label for="genericPrompt">{{ $t('presenter.genericPromptLabel') }}</label>
        <textarea id="genericPrompt" v-model="genericPrompt" rows="5" :placeholder="$t('presenter.genericPromptPlaceholder')"></textarea>
      </div>
      <div class="prompt-editor">
        <label for="kbPrompt">{{ $t('presenter.kbPromptLabel') }}</label>
        <textarea id="kbPrompt" v-model="kbPrompt" rows="8" :placeholder="$t('presenter.kbPromptPlaceholder')"></textarea>
        <small>{{ $t('presenter.kbPromptHint') }}</small>
      </div>
      <div class="prompt-actions">
        <button class="btn btn-primary" @click="handleUpdatePrompts" :disabled="loadingPrompts">{{ $t('presenter.savePromptsButton') }}</button>
        <span id="promptStatus" :class="promptStatusClass">{{ promptStatus }}</span>
      </div>
    </div>


    <!-- 问题列表 -->
    <div id="questionList">
      <div
        class="question-card"
        v-for="q in questions"
        :key="q.id"
        @click="openModal(q)"
      >
        <div class="content">{{ q.content }}</div>
        <div class="meta">
          <span class="status-badge">{{ q.status }}</span>
          <button class="btn btn-secondary" @click.stop="markAsFinished(q.id)">
            {{ $t('presenter.markFinished') }}
          </button>
          <button class="btn btn-danger" @click.stop="deleteQuestion(q.id)">
            {{ $t('presenter.delete') }}
          </button>
        </div>
      </div>
    </div>

    <!-- 问题详情弹窗 -->
    <div v-if="showModal" class="modal" @click.self="hideModal">
      <div class="modal-content">
        <div class="modal-header">
          <h2>{{ $t('presenter.questionDetail') }}</h2>
        </div>
        <div class="modal-body">
          <div class="question-section">
            <div class="section-title">{{ $t('presenter.questionContent') }}</div>
            <div
              id="modalQuestionContent"
              class="markdown-content"
              v-html="currentQuestionMarkdown"
            ></div>
          </div>
          <div class="question-section">
            <div class="section-title">{{ $t('presenter.aiSuggestion') }}</div>
            <div
              id="modalAiSuggestion"
              class="markdown-content"
              v-html="currentAISuggestionMarkdown"
            ></div>
          </div>
          <!-- 显示知识库建议 -->
          <div class="question-section" v-if="currentQuestionKbSuggestion"> <!-- 仅当有知识库建议时显示 -->
            <div class="section-title">{{ $t('presenter.kbSuggestion') }}</div>
            <div
              id="modalKbSuggestion"
              class="markdown-content"
              v-html="currentKbSuggestionMarkdown"
            ></div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-primary" @click="showQuestionOnDisplay">
            {{ $t('presenter.show') }}
          </button>
          <button class="btn btn-secondary" @click="hideModal">
            {{ $t('presenter.close') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import config from '../config/index.js'
import { marked } from 'marked'
import { useRoute } from 'vue-router'
import LanguageSwitcher from '../components/LanguageSwitcher.vue'

const { t } = useI18n()

const questions = ref([])
const showModal = ref(false)
const currentQuestionId = ref(null)
const currentQuestionContent = ref('')
const currentQuestionAiSuggestion = ref('')
const currentQuestionKbSuggestion = ref('') // 新增：用于存储知识库建议
const route = useRoute()
const sessionId = computed(() => route.query.sessionId)

// 文档管理相关的 ref
const fileInputRef = ref(null)
const uploadStatus = ref('')
const uploadStatusClass = ref('')
const uploadedDocuments = ref([])
const loadingDocuments = ref(false)

// 提示词编辑相关的 ref
const genericPrompt = ref('')
const kbPrompt = ref('')
const promptStatus = ref('')
const promptStatusClass = ref('')
const loadingPrompts = ref(false) // 用于加载状态
const savingPrompts = ref(false) // 用于保存状态

let intervalId = null

function openModal(q) {
  currentQuestionId.value = q.id
  currentQuestionContent.value = q.content || ''
  currentQuestionAiSuggestion.value = q.ai_suggestion || t('presenter.noAiSuggestion')
  currentQuestionKbSuggestion.value = q.kb_suggestion || '' // 获取知识库建议
  showModal.value = true
}

function hideModal() {
  showModal.value = false
}

// 将markdown内容转化为安全的HTML
const currentQuestionMarkdown = computed(() => marked.parse(currentQuestionContent.value || ''))
const currentAISuggestionMarkdown = computed(() => marked.parse(currentQuestionAiSuggestion.value || t('presenter.noAiSuggestion')))
const currentKbSuggestionMarkdown = computed(() => marked.parse(currentQuestionKbSuggestion.value || t('presenter.noKbSuggestion')))

async function loadQuestions() {
  try {
    if (!sessionId.value) {
      // console.warn('Session ID not available yet for loading questions.');
      return; // 等待 Session ID
    }
    const response = await fetch(`${config.api.endpoint}/questions/${sessionId.value}`)
    if (!response.ok) throw new Error('加载问题列表失败')
    const data = await response.json()
    questions.value = data
  } catch (error) {
    console.error('加载问题失败:', error)
  }
}

async function deleteQuestion(id) {
  if (!confirm(t('presenter.confirmDelete'))) return
  try {
    await fetch(`${config.api.endpoint}/question/${id}`, {
      method: 'DELETE'
    })
    loadQuestions()
  } catch (error) {
    console.error('删除失败:', error)
  }
}

async function markAsFinished(id) {
  try {
    await fetch(`${config.api.endpoint}/question/status`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id, status: 'finished' })
    })
    loadQuestions()
  } catch (error) {
    console.error('更新失败:', error)
  }
}

async function showQuestionOnDisplay() {
  if (!currentQuestionId.value) return
  try {
    const response = await fetch(`${config.api.endpoint}/question/status`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        id: currentQuestionId.value,
        status: 'showing'
      })
    })
    if (response.ok) {
      hideModal()
      loadQuestions()
    }
  } catch (error) {
    console.error('设置显示问题失败:', error)
  }
}

onMounted(() => {
  marked.setOptions({
    breaks: true,
    gfm: true,
    headerIds: false
  })

  // 确保 sessionId 有效后再加载数据
  if (sessionId.value) {
      loadQuestions()
      loadUploadedDocuments()
      loadSessionPrompts()
      intervalId = setInterval(loadQuestions, 5000)
  } else {
      console.warn("Session ID not found on mount, waiting for route query.")
      // 可以考虑使用 watch 来监听 sessionId 的变化
  }
})

onUnmounted(() => {
  if (intervalId) {
    clearInterval(intervalId)
    intervalId = null
  }
})

// --- 文档管理方法 ---

// 格式化日期时间
function formatDateTime(dateTimeString) {
  if (!dateTimeString) return '';
  try {
    // 尝试解析可能为 UTC 的时间字符串
    const date = new Date(dateTimeString);
    if (isNaN(date.getTime())) { // 无效日期
        return dateTimeString; // 返回原始字符串
    }
    // 如果需要，可以指定时区，例如 'zh-CN', { timeZone: 'Asia/Shanghai' }
    return date.toLocaleString();
  } catch (e) {
    console.error("Error formatting date:", e);
    return dateTimeString;
  }
}


// 处理文档上传
async function handleDocumentUpload() {
  const file = fileInputRef.value?.files[0];
  if (!file) {
    uploadStatus.value = t('presenter.selectFileError');
    uploadStatusClass.value = 'status-error';
    return;
  }
  if (!sessionId.value) {
    uploadStatus.value = t('presenter.noSessionIdError');
    uploadStatusClass.value = 'status-error';
    return;
  }

  uploadStatus.value = t('presenter.uploading');
  uploadStatusClass.value = '';

  const formData = new FormData();
  formData.append('file', file);
  formData.append('sessionId', sessionId.value);

  try {
    const response = await fetch(`${config.api.endpoint}/documents`, {
      method: 'POST',
      body: formData,
    });
    const result = await response.json();

    if (response.ok && result.status === 'success') {
      uploadStatus.value = t('presenter.uploadSuccess', { title: result.document.title });
      uploadStatusClass.value = 'status-success';
      fileInputRef.value.value = ''; // 清空文件选择
      loadUploadedDocuments(); // 刷新列表
    } else {
      throw new Error(result.error || t('presenter.uploadFailed'));
    }
  } catch (error) {
    console.error('上传错误:', error);
    uploadStatus.value = t('presenter.uploadError', { message: error.message });
    uploadStatusClass.value = 'status-error';
  }
}

// 加载已上传文档列表
async function loadUploadedDocuments() {
  if (!sessionId.value) {
    console.error("无法加载文档列表：缺少 sessionId");
    uploadedDocuments.value = [];
    return;
  }
  loadingDocuments.value = true;
  try {
    const response = await fetch(`${config.api.endpoint}/documents/${sessionId.value}`);
    if (!response.ok) {
      throw new Error(`获取文档列表失败: ${response.statusText}`);
    }
    const data = await response.json();
    uploadedDocuments.value = data || [];
  } catch (error) {
    console.error('加载文档列表错误:', error);
    uploadedDocuments.value = []; // 出错时清空列表
    uploadStatus.value = t('presenter.loadDocsError', { message: error.message });
    uploadStatusClass.value = 'status-error';
  } finally {
    loadingDocuments.value = false;
  }
}

// 处理删除文档
async function handleDeleteDocument(docId, docTitle) {
  if (!confirm(t('presenter.confirmDeleteDoc', { title: docTitle }))) return;

  uploadStatus.value = t('presenter.deletingDoc', { title: docTitle });
  uploadStatusClass.value = '';

  try {
    const response = await fetch(`${config.api.endpoint}/document/${docId}`, {
      method: 'DELETE',
    });
    const result = await response.json();

    if (response.ok && result.status === 'success') {
      uploadStatus.value = t('presenter.deleteDocSuccess', { title: docTitle });
      uploadStatusClass.value = 'status-success';
      loadUploadedDocuments(); // 刷新列表
    } else {
      throw new Error(result.error || t('presenter.deleteDocFailed'));
    }
  } catch (error) {
    console.error('删除文档错误:', error);
    uploadStatus.value = t('presenter.deleteDocError', { title: docTitle, message: error.message });
    uploadStatusClass.value = 'status-error';
  }
}
// --- 文档管理方法结束 ---

// --- 新增：提示词管理方法 ---

// 加载会话提示词
async function loadSessionPrompts() {
  if (!sessionId.value) {
    console.error("无法加载提示词：缺少 sessionId");
    return;
  }
  loadingPrompts.value = true;
  promptStatus.value = t('presenter.loadingPrompts');
  promptStatusClass.value = '';
  try {
    const response = await fetch(`${config.api.endpoint}/prompts/${sessionId.value}`);
    if (!response.ok) {
      throw new Error(`获取提示词失败: ${response.statusText}`);
    }
    const data = await response.json();
    // 后端返回的是指针，可能为 null，需要处理
    genericPrompt.value = data.genericPrompt || ''; // 如果为 null 则设为空字符串
    kbPrompt.value = data.kbPrompt || '';       // 如果为 null 则设为空字符串
    promptStatus.value = ''; // 加载成功，清除状态
  } catch (error) {
    console.error('加载提示词错误:', error);
    promptStatus.value = t('presenter.loadPromptsError', { message: error.message });
    promptStatusClass.value = 'status-error';
  } finally {
    loadingPrompts.value = false;
  }
}

// 处理更新提示词
async function handleUpdatePrompts() {
  if (!sessionId.value) {
    promptStatus.value = t('presenter.noSessionIdError');
    promptStatusClass.value = 'status-error';
    return;
  }
  savingPrompts.value = true;
  promptStatus.value = t('presenter.savingPrompts');
  promptStatusClass.value = '';

  try {
    const response = await fetch(`${config.api.endpoint}/prompts/${sessionId.value}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      // 发送 null 如果文本框为空，让后端知道是清空操作
      body: JSON.stringify({
          genericPrompt: genericPrompt.value.trim() === '' ? null : genericPrompt.value,
          kbPrompt: kbPrompt.value.trim() === '' ? null : kbPrompt.value
      })
    });
    const result = await response.json();

    if (response.ok && result.status === 'success') {
      promptStatus.value = t('presenter.savePromptsSuccess');
      promptStatusClass.value = 'status-success';
    } else {
      throw new Error(result.error || t('presenter.savePromptsFailed'));
    }
  } catch (error) {
    console.error('更新提示词错误:', error);
    promptStatus.value = t('presenter.savePromptsError', { message: error.message });
    promptStatusClass.value = 'status-error';
  } finally {
    savingPrompts.value = false;
  }
}
// --- 提示词管理方法结束 ---

</script>

<style scoped>
* {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    line-height: 1.6;
    color: #333;
    background: #f5f5f5;
    padding: 20px;
}

.container {
    max-width: 1200px;
    margin: 0 auto;
}

.header {
    position: relative;
    display: flex;
    justify-content: space-between;
    align-items: center;
    background: #fff;
    padding: 20px;
    border-radius: 8px;
    margin-bottom: 20px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.question-card {
    background: #fff;
    padding: 20px;
    margin: 15px 0;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    transition: transform 0.2s, box-shadow 0.2s;
    cursor: pointer;
}

.question-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(0,0,0,0.15);
}

.content {
    font-size: 1.1em;
    margin-bottom: 15px;
}

.meta {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 15px;
}

.status-badge {
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 0.9em;
    font-weight: 500;
    background: #e9ecef;
}

.ai-suggestion {
    background: #f8f9fa;
    padding: 15px;
    border-radius: 6px;
    border-left: 4px solid #6c757d;
    margin-top: 10px;
}

.btn {
    padding: 8px 16px;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-weight: 500;
    transition: background-color 0.2s;
}
.btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}


.btn-primary {
    background: #007bff;
    color: white;
}

.btn-secondary {
    background: #6c757d;
    color: white;
}

.btn-danger {
    background: #dc3545;
    color: white;
}

.btn:hover:not(:disabled) {
    opacity: 0.9;
}


.modal {
    display: block;
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0,0,0,0.7);
    z-index: 1000;
}

.modal-content {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    background: white;
    padding: 30px;
    border-radius: 8px;
    width: 90%;
    max-width: 800px;
    max-height: 90vh;
    overflow-y: auto;
}

.modal-header {
    margin-bottom: 20px;
    padding-bottom: 15px;
    border-bottom: 1px solid #dee2e6;
}

.modal-body {
    margin-bottom: 20px;
}

.modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    padding-top: 15px;
    border-top: 1px solid #dee2e6;
}

.markdown-content {
    line-height: 1.8;
    background: #f8f9fa; /* 给markdown内容加个背景 */
    padding: 10px 15px;
    border-radius: 4px;
    border: 1px solid #eee;
}

.markdown-content h1,
.markdown-content h2,
.markdown-content h3 {
    margin: 1em 0 0.5em;
    border-bottom: 1px solid #eee;
    padding-bottom: 0.3em;
}

.markdown-content p {
    margin-bottom: 1em;
}

.markdown-content code {
    background: #e9ecef; /* 调整代码块背景 */
    padding: 2px 4px;
    border-radius: 4px;
    font-family: Consolas, Monaco, 'Andale Mono', 'Ubuntu Mono', monospace;
}

.markdown-content pre {
    background: #e9ecef; /* 调整代码块背景 */
    padding: 15px;
    border-radius: 4px;
    overflow-x: auto;
    border: 1px solid #ddd;
}
.markdown-content pre code {
    background: none;
    padding: 0;
    border: none;
}


.question-section {
    margin-bottom: 25px;
}

.section-title {
    font-size: 1.1em;
    font-weight: 600;
    margin-bottom: 10px;
    color: #495057;
}

.language-switcher {
    margin-left: 20px;
}

/* 文档管理区域样式 */
.document-management-section {
  display: flex;
  gap: 20px; /* 左右间距 */
  margin-bottom: 20px;
  flex-wrap: wrap; /* 在小屏幕上换行 */
}

.document-upload-section,
.document-list-section {
  flex: 1; /* 让两部分平分宽度 */
  background: #fff;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  min-width: 300px; /* 保证最小宽度 */
}

.document-upload-section h2,
.document-list-section h2 {
  margin-top: 0; /* 移除默认的上边距 */
  margin-bottom: 15px;
  font-size: 1.2em;
  color: #333;
  border-bottom: 1px solid #eee;
  padding-bottom: 10px;
}

#uploadForm {
  display: flex;
  align-items: center;
  flex-wrap: wrap; /* 允许换行 */
  gap: 10px;
}

#uploadForm input[type="file"] {
  flex-grow: 1; /* 让文件输入框填充空间 */
  border: 1px solid #ccc;
  padding: 5px;
  border-radius: 4px;
  min-width: 200px; /* 保证最小宽度 */
}

#uploadStatus {
  margin-top: 10px;
  font-weight: 500;
  min-height: 1.2em; /* 避免状态消失时布局跳动 */
  width: 100%; /* 占满整行 */
}

.status-success {
  color: #28a745;
}

.status-error {
  color: #dc3545;
}

/* 文档列表样式 */
#documentList ul {
  list-style: none;
  padding: 0;
  max-height: 250px; /* 调整最大高度 */
  overflow-y: auto;
}

#documentList li {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 5px; /* 调整内边距 */
  border-bottom: 1px solid #eee;
  gap: 10px; /* 元素间距 */
}

#documentList li:last-child {
  border-bottom: none;
}

#documentList .doc-title {
  flex-grow: 1;
  margin-right: 10px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.95em;
}

#documentList .doc-meta {
  font-size: 0.85em;
  color: #6c757d;
  margin-right: 10px;
  min-width: 140px; /* 调整宽度以适应日期时间 */
  text-align: right;
  white-space: nowrap;
}

#documentList .btn-delete-doc {
  padding: 4px 8px; /* 调整按钮大小 */
  font-size: 0.85em;
  flex-shrink: 0; /* 防止按钮被压缩 */
}

/* 提示词编辑区域样式 */
.prompt-editing-section {
  background: #fff;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 20px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}
.prompt-editing-section h2 {
  margin-top: 0;
  margin-bottom: 15px;
  font-size: 1.2em;
  color: #333;
  border-bottom: 1px solid #eee;
  padding-bottom: 10px;
}
.prompt-editor {
  margin-bottom: 15px;
}
.prompt-editor label {
  display: block;
  margin-bottom: 5px;
  font-weight: 600;
  color: #495057;
}
.prompt-editor textarea {
  width: 100%;
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-family: inherit;
  font-size: 0.95em;
  line-height: 1.5;
  resize: vertical; /* 允许垂直调整大小 */
}
.prompt-editor small {
    display: block;
    margin-top: 5px;
    font-size: 0.85em;
    color: #6c757d;
}
.prompt-actions {
    display: flex;
    align-items: center;
    gap: 15px;
    margin-top: 15px;
}
#promptStatus {
    font-weight: 500;
}

</style>
