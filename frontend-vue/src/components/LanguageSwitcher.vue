<template>
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
</template>

<script setup>
import { useI18n } from 'vue-i18n'

const { locale } = useI18n()

function changeLang(lang) {
  locale.value = lang
}
</script>

<style scoped>
.language-switcher {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 1000;
  display: flex;
  gap: 10px;
}

@media (max-width: 768px) {
  .language-switcher {
    top: 10px;
    right: 10px;
  }
}
</style>