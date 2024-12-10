<template>
  <v-container class="pa-4" max-width="600px">
    <v-card elevation="2" class="pa-4">
      <v-card-title class="text-h5 pb-2 justify-center position-relative">
        {{ $t('title.createSession') }}
        <div class="language-switcher">
          <v-menu>
            <template v-slot:activator="{ props }">
              <v-btn
                v-bind="props"
                color="primary"
                class="lang-menu-btn"
                icon
              >
                <v-icon>mdi-translate</v-icon>
              </v-btn>
            </template>
            <v-list>
              <v-list-item @click="changeLang('zh')">
                <v-list-item-title>{{ $t('button.chinese') }}</v-list-item-title>
              </v-list-item>
              <v-list-item @click="changeLang('en')">
                <v-list-item-title>{{ $t('button.english') }}</v-list-item-title>
              </v-list-item>
              <v-list-item @click="changeLang('jp')">
                <v-list-item-title>{{ $t('button.japanese') }}</v-list-item-title>
              </v-list-item>
            </v-list>
          </v-menu>
        </div>
      </v-card-title>
      
      <v-card-subtitle class="text-center pb-4">
        {{ $t('title.sessionSetup') }}
      </v-card-subtitle>
      
      <v-divider class="mb-4"></v-divider>

      <v-card-text class="text-center">
        <div v-if="!sessionId">
          <p class="mb-6">{{ $t('message.clickToGenerate') }}</p>
          <div class="d-flex justify-center">
            <v-btn 
              color="primary" 
              @click="createSession" 
              elevation="1" 
              rounded 
              class="px-6"
            >
              {{ $t('button.start') }}
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
            {{ $t('message.sessionCreated') }}
          </v-alert>
          <p class="text-body-1 mb-2">{{ $t('message.sessionIdGenerated') }} <strong>{{ sessionId }}</strong></p>
          <p class="mb-4">{{ $t('message.useFollowingLinks') }}</p>

          <ul class="text-start mb-4">
            <li class="mb-2">{{ $t('message.questionPage') }}<a :href="indexUrl" target="_blank">{{ indexUrl }}</a></li>
            <li class="mb-2">{{ $t('message.adminPanel') }}<a :href="presenterUrl" target="_blank">{{ presenterUrl }}</a></li>
            <li class="mb-2">{{ $t('message.displayScreen') }}<a :href="displayUrl" target="_blank">{{ displayUrl }}</a></li>
          </ul>

          <v-divider class="my-4"></v-divider>

          <p class="mb-2">{{ $t('message.scanQrCode') }}</p>
          <div class="d-flex justify-center">
            <canvas ref="qrCanvas" width="200" height="200"></canvas>
          </div>
        </div>
      </v-card-text>

      <v-card-actions v-if="sessionId" class="justify-center mt-4">
        <v-btn color="secondary" @click="createSession" rounded prepend-icon="mdi-refresh">
          {{ $t('button.regenerate') }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-container>
</template>

<script setup>
import { ref, watchEffect, computed, nextTick } from 'vue'
import QRCode from 'qrcode'
import { useI18n } from 'vue-i18n'

const { locale } = useI18n()

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

const changeLang = (lang) => {
  locale.value = lang
}

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

.language-switcher {
  position: absolute;
  right: 0;
  top: 50%;
  transform: translateY(-50%);
}

.position-relative {
  position: relative;
}
</style>
