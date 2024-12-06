<script setup>
import SakuraAnimation from './components/SakuraAnimation.vue'
</script>

<template>
  <v-app>
    <v-main>
      <router-view />
    </v-main>
  </v-app>
</template>

<style>
.app-container {
  background: linear-gradient(140deg, #fdfbfb 0%, #ebedee 100%);
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px;
}
</style>
