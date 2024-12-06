// src/composables/useConfig.js
import { ref, onMounted } from 'vue'
import yaml from 'js-yaml'

export function useConfig() {
  const config = ref(null)
  const configError = ref(null)

  const loadConfig = async () => {
    try {
      const response = await fetch('./config/config.yaml')
      if (!response.ok) {
        throw new Error(`Failed to load config: ${response.status}`)
      }
      
      const text = await response.text()
      const parsedConfig = yaml.load(text)

      if (!parsedConfig?.api?.endpoint) {
        throw new Error('Missing api.endpoint configuration')
      }
      if (!parsedConfig?.session?.id) {
        throw new Error('Missing session.id configuration')
      }

      config.value = parsedConfig
      configError.value = null
    } catch (error) {
      console.error('Config loading error:', error)
      configError.value = error.message
    }
  }

  onMounted(() => {
    loadConfig()
  })

  return {
    config,
    configError
  }
}