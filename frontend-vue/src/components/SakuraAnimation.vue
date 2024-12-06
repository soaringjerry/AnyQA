<!-- src/components/SakuraAnimation.vue -->
<template></template>

<script>
export default {
  mounted() {
    this.interval = setInterval(this.createSakura, 300)
  },
  beforeUnmount() {
    clearInterval(this.interval)
  },
  methods: {
    createSakura() {
      const sakura = document.createElement('div')
      sakura.className = 'sakura'
      sakura.style.left = Math.random() * window.innerWidth + 'px'
      sakura.innerHTML = '🌸'
      sakura.style.fontSize = (Math.random() * 20 + 10) + 'px'
      sakura.style.opacity = (Math.random() * 0.6 + 0.4).toString()
      sakura.style.position = 'fixed'
      sakura.style.zIndex = '999'
      sakura.style.pointerEvents = 'none'
      sakura.style.animation = `falling ${Math.random() * 3 + 2}s linear infinite`
      document.body.appendChild(sakura)

      setTimeout(() => {
        sakura.remove()
      }, 5000)
    }
  }
}
</script>

<style>
@keyframes falling {
  0% {
    transform: translateY(-100%) rotate(0deg);
  }
  100% {
    transform: translateY(100vh) rotate(360deg);
  }
}
</style>
