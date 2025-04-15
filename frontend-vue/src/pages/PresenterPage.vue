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
          <input type="file" ref="fileInputRef" accept=".pdf,.docx,.txt,.md,.html,.htm,.csv,.json,.xlsx,.xls" required>
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
         <button class="btn btn-primary" @click="handleUpdatePrompts" :disabled="savingPrompts || loadingPrompts">
            {{ savingPrompts ? $t('presenter.savingPrompts') : $t('presenter.savePromptsButton') }}
        </button>
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
import { ref, onMounted, onUnmounted, computed, watch } from 'vue' // 导入 watch
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
const currentQuestionKbSuggestion = ref('')
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
const loadingPrompts = ref(false)
const savingPrompts = ref(false)

let intervalId = null

function openModal(q) {
  currentQuestionId.value = q.id
  currentQuestionContent.value = q.content || ''
  currentQuestionAiSuggestion.value = q.ai_suggestion || t('presenter.noAiSuggestion')
  currentQuestionKbSuggestion.value = q.kb_suggestion || ''
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
      return;
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

// 封装加载初始数据的函数
function loadInitialData() {
    if (sessionId.value) {
        console.log("Session ID found, loading initial data:", sessionId.value);
        loadQuestions();
        loadUploadedDocuments();
        loadSessionPrompts();
        if (!intervalId) { // 避免重复设置 interval
            intervalId = setInterval(loadQuestions, 5000);
        }
    } else {
        console.warn("Session ID not available yet.");
    }
}

onMounted(() => {
  marked.setOptions({
    breaks: true,
    gfm: true,
    headerIds: false
  })
  loadInitialData(); // 尝试加载初始数据
})

onUnmounted(() => {
  if (intervalId) {
    clearInterval(intervalId)
    intervalId = null
  }
})

// 监听 sessionId 的变化，以便在路由参数可用时加载数据
watch(sessionId, (newSessionId, oldSessionId) => {
    console.log("Session ID changed:", oldSessionId, "->", newSessionId);
    if (newSessionId) {
        loadInitialData();
    } else {
        // Session ID 丢失，可能需要清理状态
        questions.value = [];
        uploadedDocuments.value = [];
        genericPrompt.value = '';
        kbPrompt.value = '';
        if (intervalId) {
            clearInterval(intervalId);
            intervalId = null;
        }
    }
});


// --- 文档管理方法 ---

// 格式化日期时间
function formatDateTime(dateTimeString) {
  if (!dateTimeString) return '';
  try {
    const date = new Date(dateTimeString);
    if (isNaN(date.getTime())) {
        return dateTimeString;
    }
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
      fileInputRef.value.value = '';
      loadUploadedDocuments();
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
    uploadedDocuments.value = [];
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
      loadUploadedDocuments();
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

// --- 提示词管理方法 ---

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
    genericPrompt.value = data.genericPrompt || '';
    kbPrompt.value = data.kbPrompt || '';
    promptStatus.value = '';
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
    promptStatus.value = t('presenter.noSessionIdError'); // 复用错误消息
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

/* 使 body 和 html 占满高度，为 flex 容器提供基础 */
/* 注意：这些样式通常在全局 CSS (如 src/style.css 或 App.vue) 中设置更合适 */
/* 这里为了演示效果暂时放在 scoped style 中，但可能需要调整 */
:global(html, body) {
  height: auto; /* 改为 auto，允许根据内容伸展 */
  margin: 0;
  padding: 0;
  overflow-y: auto; /* 明确设置为 auto，允许滚动 */
  background-color: #ffffff; /* 纯白色背景 */
}
:global(#app) { /* 假设 Vue 应用挂载在 #app */
  height: auto; /* 改为 auto，允许根据内容伸展 */
  overflow-y: auto; /* 明确设置为 auto，允许滚动 */
  background-color: #ffffff; /* 纯白色背景 */
}


.container {
    max-width: 1200px;
    margin: 0 auto;
    display: flex; /* 使用 Flexbox 布局 */
    flex-direction: column; /* 垂直排列 */
    min-height: 100%; /* 最小高度为100%，但可以根据内容伸展 */
    height: auto; /* 明确设置为 auto，允许根据内容伸展 */
    padding: 30px; /* 增加内边距 */
    padding-bottom: 50px; /* 增加底部内边距 */
    box-sizing: border-box; /* 确保 padding 不会增加总高度 */
    overflow-y: visible; /* 允许内容溢出并显示滚动条 */
    background-color: #ffffff; /* 纯白色背景 */
}

.header {
    flex-shrink: 0; /* 防止 header 被压缩 */
    position: relative;
    display: flex;
    justify-content: space-between;
    align-items: center;
    background: #ffffff;
    padding: 24px;
    border-radius: 16px;
    margin-bottom: 30px;
    box-shadow: 0 10px 30px rgba(0,0,0,0.03);
    border: 1px solid rgba(0,0,0,0.02);
}

.question-card {
    background: #ffffff;
    padding: 24px;
    margin: 20px 0;
    border-radius: 16px;
    box-shadow: 0 10px 30px rgba(0,0,0,0.03);
    transition: transform 0.3s, box-shadow 0.3s;
    cursor: pointer;
    border: 1px solid rgba(0,0,0,0.02);
}

.question-card:hover {
    transform: translateY(-3px);
    box-shadow: 0 15px 40px rgba(0,0,0,0.06);
    border-color: rgba(0,0,0,0.03);
}

.content {
    font-size: 1.15em;
    margin-bottom: 18px;
    color: #333333;
    line-height: 1.5;
}

.meta {
    display: flex;
    align-items: center;
    gap: 15px;
    margin-bottom: 15px;
}

.status-badge {
    padding: 6px 12px;
    border-radius: 30px;
    font-size: 0.85em;
    font-weight: 600;
    background: #f0f5ff;
    color: #4a6cf7;
}

.ai-suggestion {
    background: #f8f9fa;
    padding: 15px;
    border-radius: 6px;
    border-left: 4px solid #6c757d;
    margin-top: 10px;
}
.btn {
    padding: 10px 20px;
    border: none;
    border-radius: 30px;
    cursor: pointer;
    font-weight: 600;
    transition: all 0.3s ease;
    font-size: 0.95em;
}
.btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}



.btn-primary {
    background: linear-gradient(135deg, #4a6cf7, #3b5fe2);
    color: white;
    box-shadow: 0 5px 15px rgba(74, 108, 247, 0.2);
}

.btn-secondary {
    background: linear-gradient(135deg, #6c757d, #5a6268);
    color: white;
    box-shadow: 0 5px 15px rgba(108, 117, 125, 0.2);
}

.btn-danger {
    background: linear-gradient(135deg, #dc3545, #c82333);
    color: white;
    box-shadow: 0 5px 15px rgba(220, 53, 69, 0.2);
}
.btn:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 8px 20px rgba(0, 0, 0, 0.1);
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
  gap: 30px; /* 增加左右间距 */
  margin-bottom: 30px;
  flex-wrap: wrap; /* 在小屏幕上换行 */
  flex-shrink: 0; /* 防止此区域被压缩 */
}

.document-upload-section,
.document-list-section {
  flex: 1; /* 让两部分平分宽度 */
  background: #ffffff;
  padding: 24px;
  border-radius: 16px;
  box-shadow: 0 10px 30px rgba(0,0,0,0.03);
  min-width: 300px; /* 保证最小宽度 */
  border: 1px solid rgba(0,0,0,0.02);
  margin: 0 10px; /* 添加左右间距 */
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
  background: #ffffff;
  padding: 24px;
  border-radius: 16px;
  margin-bottom: 30px;
  box-shadow: 0 10px 30px rgba(0,0,0,0.03);
  flex-shrink: 0; /* 防止此区域被压缩 */
  border: 1px solid rgba(0,0,0,0.02);
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
    min-height: 1.2em; /* 避免状态消失时布局跳动 */
}

/* 问题列表样式调整 */
#questionList {
    flex-grow: 1; /* 让问题列表填充剩余空间 */
    overflow-y: visible; /* 改为 visible，让内容自然溢出 */
    /* margin-top: 20px; */ /* 移除或调整，因为上面有提示词区域 */
    padding-right: 15px; /* 为滚动条留出空间 */
    min-height: 200px; /* 设置一个合理的最小高度 */
    background: #ffffff; /* 给问题列表区域加个背景 */
    padding: 20px 25px; /* 增加内边距 */
    padding-bottom: 30px; /* 增加底部内边距 */
    margin-bottom: 30px; /* 增加底部外边距 */
    border-radius: 16px;
    box-shadow: 0 10px 30px rgba(0,0,0,0.03);
    border: 1px solid rgba(0,0,0,0.02);
}


</style>
