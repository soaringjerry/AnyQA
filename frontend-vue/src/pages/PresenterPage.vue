<template>
  <div class="container">
    <div class="header">
      <h1>演讲者控制台（点击问题查看详情页）</h1>
    </div>
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
          <button class="btn btn-secondary" @click.stop="markAsFinished(q.id)">标记完成</button>
          <button class="btn btn-danger" @click.stop="deleteQuestion(q.id)">删除</button>
        </div>
      </div>
    </div>

    <div v-if="showModal" class="modal" @click.self="hideModal">
      <div class="modal-content">
        <div class="modal-header">
          <h2>问题详情</h2>
        </div>
        <div class="modal-body">
          <div class="question-section">
            <div class="section-title">问题内容</div>
            <div 
              id="modalQuestionContent" 
              class="markdown-content" 
              v-html="currentQuestionMarkdown"
            ></div>
          </div>
          <div class="question-section">
            <div class="section-title">AI 建议回复</div>
            <div 
              id="modalAiSuggestion" 
              class="markdown-content" 
              v-html="currentAISuggestionMarkdown"
            ></div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-primary" @click="showQuestionOnDisplay">显示问题</button>
          <button class="btn btn-secondary" @click="hideModal">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import config from '../config/index.js' // 假设此文件已加载YAML配置
import { marked } from 'marked'

const questions = ref([])  
const showModal = ref(false)
const currentQuestionId = ref(null)
const currentQuestionContent = ref('')
const currentQuestionAiSuggestion = ref('')

let intervalId = null

function openModal(q) {
  currentQuestionId.value = q.id
  currentQuestionContent.value = q.content || ''
  currentQuestionAiSuggestion.value = q.ai_suggestion || '暂无AI建议'
  showModal.value = true
}

function hideModal() {
  showModal.value = false
}

// 将markdown内容转化为安全的HTML
const currentQuestionMarkdown = computed(() => marked.parse(currentQuestionContent.value || ''))
const currentAISuggestionMarkdown = computed(() => marked.parse(currentQuestionAiSuggestion.value || '暂无AI建议'))

async function loadQuestions() {
  try {
    const response = await fetch(`${config.api.endpoint}/questions/${config.session.id}`)
    if (!response.ok) throw new Error('加载问题列表失败')
    const data = await response.json()
    questions.value = data
  } catch (error) {
    console.error('加载问题失败:', error)
  }
}

async function deleteQuestion(id) {
  if (!confirm('确定要删除这个问题吗？')) return
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

  loadQuestions()
  intervalId = setInterval(loadQuestions, 5000)
})

onUnmounted(() => {
  if (intervalId) {
    clearInterval(intervalId)
    intervalId = null
  }
})
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

.btn:hover {
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
}

.markdown-content h1,
.markdown-content h2,
.markdown-content h3 {
    margin: 1em 0 0.5em;
}

.markdown-content p {
    margin-bottom: 1em;
}

.markdown-content code {
    background: #f8f9fa;
    padding: 2px 4px;
    border-radius: 4px;
}

.markdown-content pre {
    background: #f8f9fa;
    padding: 15px;
    border-radius: 4px;
    overflow-x: auto;
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
</style>
