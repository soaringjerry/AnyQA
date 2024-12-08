<template>
  <v-container class="pa-4" max-width="600px">
    <v-card elevation="2" class="pa-4">
      <v-card-title class="text-h5 pb-2 justify-center">
        创建新的会话
      </v-card-title>
      <v-card-subtitle class="text-center pb-4">
        在这里快速生成一个全新会话并获取对应的访问链接和二维码。
      </v-card-subtitle>
      
      <v-divider class="mb-4"></v-divider>

      <v-card-text class="text-center">
        <div v-if="!sessionId">
          <p class="mb-6">点击下方按钮生成一个新的 Session ID，并获得相应的访问链接。</p>
          <div class="d-flex justify-center">
            <v-btn 
              color="primary" 
              @click="createSession" 
              elevation="1" 
              rounded 
              class="px-6"
            >
              开始
            </v-btn>
          </div>
        </div>
        <div v-else>
          <v-alert 
            type="success" 
            class="mb-4" 
            border="start" 
            colored-border
            :icon="false"
          >
            会话已成功生成！
          </v-alert>
          <p class="text-body-1 mb-2">已生成的 Session ID: <strong>{{ sessionId }}</strong></p>
          <p class="mb-4">请使用以下链接访问：</p>

          <ul class="text-start mb-4">
            <li class="mb-2">提问页面：<a :href="indexUrl" target="_blank">{{ indexUrl }}</a></li>
            <li class="mb-2">管理后台：<a :href="presenterUrl" target="_blank">{{ presenterUrl }}</a></li>
            <li class="mb-2">展示大屏：<a :href="displayUrl" target="_blank">{{ displayUrl }}</a></li>
          </ul>

          <v-divider class="my-4"></v-divider>

          <p class="mb-2">扫描下方二维码直接打开提问页面 (Index)：</p>
          <div class="d-flex justify-center">
            <canvas ref="qrCanvas" width="200" height="200"></canvas>
          </div>
        </div>
      </v-card-text>

      <v-card-actions v-if="sessionId" class="justify-center mt-4">
        <v-btn color="secondary" @click="createSession" rounded prepend-icon="mdi-refresh">
          重新生成
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-container>
</template>

<script setup>
import { ref, watchEffect, computed, nextTick } from 'vue'
import QRCode from 'qrcode'

// 假设域名为 yourdomain.com
const baseDomain = window.location.origin

const sessionId = ref('')
const qrCanvas = ref(null)

// 简易生成一个随机字符串作为sessionId
const createSession = () => {
  sessionId.value = Math.random().toString(36).substring(2, 8)
}

const indexUrl = computed(() => `${baseDomain}/#/?sessionId=${sessionId.value}`)
const presenterUrl = computed(() => `${baseDomain}/#/presenter?sessionId=${sessionId.value}`)
const displayUrl = computed(() => `${baseDomain}/#/display?sessionId=${sessionId.value}`)

watchEffect(async () => {
  if (sessionId.value) {
    await nextTick()
    if (qrCanvas.value) {
      try {
        await QRCode.toCanvas(qrCanvas.value, indexUrl.value, { 
          width: 200,
          color: {
            dark: '#000',
            light: '#fff'
          }
        })
      } catch (err) {
        console.error('二维码生成出错:', err)
      }
    }
  }
})
</script>

<style scoped>
/* 根据需要定制样式 */
</style>
