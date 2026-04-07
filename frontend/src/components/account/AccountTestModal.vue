<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.testAccountConnection')"
    width="normal"
    @close="handleClose"
  >
    <div class="space-y-4">
      <!-- Account Info Card with animated gradient border -->
      <div
        v-if="account"
        class="relative overflow-hidden rounded-xl p-[1px]"
        :class="statusBorderClass"
      >
        <div class="relative flex items-center justify-between rounded-[11px] bg-white p-3 dark:bg-dark-800">
          <div class="flex items-center gap-3">
            <div
              class="relative flex h-11 w-11 items-center justify-center rounded-xl"
              :class="statusIconBgClass"
            >
              <Icon
                :name="statusIconName"
                size="md"
                class="text-white"
                :class="{ 'animate-spin': status === 'connecting' }"
                :stroke-width="2"
              />
              <!-- Pulse ring when connecting -->
              <span
                v-if="status === 'connecting'"
                class="absolute inset-0 animate-ping rounded-xl opacity-30"
                :class="statusIconBgClass"
              />
            </div>
            <div>
              <div class="font-semibold text-gray-900 dark:text-gray-100">{{ account.name }}</div>
              <div class="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400">
                <span
                  class="rounded-md bg-gray-100 px-1.5 py-0.5 text-[10px] font-bold uppercase tracking-wider dark:bg-dark-600"
                >
                  {{ account.type }}
                </span>
                <span class="text-gray-400 dark:text-gray-500">/</span>
                <span>{{ account.platform }}</span>
              </div>
            </div>
          </div>
          <!-- Status Badge with animation -->
          <div class="flex items-center gap-2">
            <span
              :class="[
                'flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-semibold transition-all duration-500',
                statusBadgeClass
              ]"
            >
              <span
                class="h-1.5 w-1.5 rounded-full"
                :class="statusDotClass"
              />
              {{ statusLabel }}
            </span>
          </div>
        </div>
      </div>

      <!-- Model Selection -->
      <div  class="space-y-1.5">
        <label class="text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('admin.accounts.selectTestModel') }}
        </label>
        <Select
          v-model="selectedModelId"
          :options="availableModels"
          :disabled="loadingModels || status === 'connecting'"
          value-key="id"
          label-key="display_name"
          :placeholder="loadingModels ? t('common.loading') + '...' : t('admin.accounts.selectTestModel')"
        />
      </div>

      <!-- Gemini Image Prompt -->
      <div v-if="supportsGeminiImageTest" class="space-y-1.5">
        <TextArea
          v-model="testPrompt"
          :label="t('admin.accounts.geminiImagePromptLabel')"
          :placeholder="t('admin.accounts.geminiImagePromptPlaceholder')"
          :hint="t('admin.accounts.geminiImageTestHint')"
          :disabled="status === 'connecting'"
          rows="3"
        />
      </div>

      <!-- Enhanced Terminal -->
      <div class="group relative">
        <!-- Terminal Header Bar -->
        <div class="flex items-center justify-between rounded-t-xl border border-b-0 border-gray-200 bg-gray-100 px-4 py-2 dark:border-dark-700/50 dark:bg-dark-800">
          <div class="flex items-center gap-2">
            <div class="flex gap-1.5">
              <span class="h-3 w-3 rounded-full" :class="status === 'error' ? 'bg-red-500' : 'bg-red-400/60'" />
              <span class="h-3 w-3 rounded-full" :class="status === 'connecting' ? 'bg-yellow-500 animate-pulse' : 'bg-yellow-400/60'" />
              <span class="h-3 w-3 rounded-full" :class="status === 'success' ? 'bg-green-500' : 'bg-green-400/60'" />
            </div>
            <span class="ml-2 text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ selectedModelId || 'terminal' }}
            </span>
          </div>
          <!-- Copy Button in header -->
          <button
            v-if="outputLines.length > 0"
            @click="copyOutput"
            class="rounded-md p-1 text-gray-400 transition-all hover:bg-gray-200 hover:text-gray-600 dark:text-gray-500 dark:hover:bg-dark-700 dark:hover:text-gray-300"
            :title="t('admin.accounts.copyOutput')"
          >
            <Icon name="copy" size="sm" :stroke-width="2" />
          </button>
        </div>

        <!-- Terminal Body -->
        <div
          ref="terminalRef"
          class="max-h-[280px] min-h-[140px] overflow-y-auto rounded-b-xl border border-t-0 border-gray-200 bg-gray-50 p-4 font-mono text-[13px] leading-relaxed dark:border-dark-700/50 dark:bg-dark-900/60"
        >
          <!-- Idle State with animated cursor -->
          <div v-if="status === 'idle'" class="flex items-center gap-2 text-gray-400 dark:text-gray-500">
            <span class="text-primary-500 dark:text-emerald-500">$</span>
            <span>{{ t('admin.accounts.readyToTest') }}</span>
            <span class="inline-block h-4 w-2 animate-pulse bg-primary-500/50 dark:bg-emerald-500/70" />
          </div>

          <!-- Connecting State -->
          <div v-else-if="status === 'connecting' && outputLines.length === 0" class="space-y-2">
            <div class="flex items-center gap-2 text-amber-600 dark:text-yellow-400">
              <svg class="h-4 w-4 animate-spin" viewBox="0 0 24 24" fill="none">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
              </svg>
              <span>{{ t('admin.accounts.connectingToApi') }}</span>
            </div>
          </div>

          <!-- Output Lines with transition -->
          <TransitionGroup name="terminal-line" tag="div">
            <div
              v-for="(line, index) in outputLines"
              :key="`line-${index}`"
              class="terminal-line-item"
              :class="line.class"
            >
              <span v-if="line.prefix" class="mr-2 select-none" :class="line.prefixClass">{{ line.prefix }}</span>
              <span>{{ line.text }}</span>
            </div>
          </TransitionGroup>

          <!-- Streaming Content with glow effect -->
          <div v-if="streamingContent" class="relative mt-1">
            <span class="text-emerald-600 dark:text-emerald-400">{{ streamingContent }}</span>
            <span class="inline-block h-4 w-[2px] animate-pulse bg-emerald-600 dark:bg-emerald-400" />
          </div>

          <!-- Result Status with animated icons -->
          <div
            v-if="status === 'success'"
            class="mt-3 flex items-center gap-2 border-t border-gray-200 pt-3 dark:border-gray-700/50"
          >
            <div class="flex h-6 w-6 items-center justify-center rounded-full bg-emerald-100 dark:bg-emerald-500/20">
              <Icon name="check" size="sm" class="text-emerald-600 dark:text-emerald-400" :stroke-width="2.5" />
            </div>
            <span class="font-medium text-emerald-600 dark:text-emerald-400">{{ t('admin.accounts.testCompleted') }}</span>
            <span v-if="elapsedTime" class="ml-auto text-xs text-gray-400 dark:text-gray-600">{{ elapsedTime }}</span>
          </div>
          <div
            v-else-if="status === 'error'"
            class="mt-3 flex items-center gap-2 border-t border-gray-200 pt-3 dark:border-gray-700/50"
          >
            <div class="flex h-6 w-6 items-center justify-center rounded-full bg-red-100 dark:bg-red-500/20">
              <Icon name="x" size="sm" class="text-red-600 dark:text-red-400" :stroke-width="2.5" />
            </div>
            <span class="font-medium text-red-600 dark:text-red-400">{{ errorMessage }}</span>
          </div>
        </div>
      </div>

      <!-- Generated Images Gallery -->
      <div v-if="generatedImages.length > 0" class="space-y-2">
        <div class="flex items-center gap-2 text-xs font-medium text-gray-600 dark:text-gray-300">
          <Icon name="sparkles" size="sm" class="text-purple-600 dark:text-purple-400" :stroke-width="2" />
          {{ t('admin.accounts.geminiImagePreview') }}
        </div>
        <div class="grid gap-3 sm:grid-cols-2">
          <a
            v-for="(image, index) in generatedImages"
            :key="`${image.url}-${index}`"
            :href="image.url"
            target="_blank"
            rel="noopener noreferrer"
            class="group/img relative overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm transition-all duration-300 hover:scale-[1.02] hover:border-purple-300 hover:shadow-lg hover:shadow-purple-500/10 dark:border-dark-500 dark:bg-dark-700"
          >
            <img :src="image.url" :alt="`gemini-test-image-${index + 1}`" class="h-48 w-full object-cover transition-transform duration-300 group-hover/img:scale-105" />
            <div class="absolute inset-0 bg-gradient-to-t from-black/40 to-transparent opacity-0 transition-opacity group-hover/img:opacity-100" />
            <div class="border-t border-gray-100 px-3 py-2 text-xs text-gray-500 dark:border-dark-500 dark:text-gray-300">
              {{ image.mimeType || 'image/*' }}
            </div>
          </a>
        </div>
      </div>

      <!-- Test Info Bar -->
      <div class="flex items-center justify-between rounded-lg bg-gray-50 px-3 py-2 dark:bg-dark-700">
        <div class="flex items-center gap-3 text-xs text-gray-500 dark:text-gray-400">
          <span class="flex items-center gap-1.5">
            <Icon name="cpu" size="sm" :stroke-width="2" />
            {{ t('admin.accounts.testModel') }}
          </span>
        </div>
        <span class="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400">
          <Icon name="bolt" size="sm" :stroke-width="2" />
          {{
            supportsGeminiImageTest
                ? t('admin.accounts.geminiImageTestMode')
                : t('admin.accounts.testPrompt')
          }}
        </span>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button
          @click="handleClose"
          class="rounded-lg bg-gray-100 px-4 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-200 dark:bg-dark-600 dark:text-gray-300 dark:hover:bg-dark-500"
          :disabled="status === 'connecting'"
        >
          {{ t('common.close') }}
        </button>
        <button
          @click="startTest"
          :disabled="status === 'connecting' || !selectedModelId"
          :class="[
            'flex items-center gap-2 rounded-lg px-5 py-2 text-sm font-medium shadow-sm transition-all duration-300',
            status === 'connecting' || !selectedModelId
              ? 'cursor-not-allowed bg-gray-400 text-white shadow-none'
              : status === 'success'
                ? 'bg-emerald-500 text-white shadow-emerald-500/25 hover:bg-emerald-600 hover:shadow-emerald-500/40'
                : status === 'error'
                  ? 'bg-orange-500 text-white shadow-orange-500/25 hover:bg-orange-600 hover:shadow-orange-500/40'
                  : 'bg-gradient-to-r from-primary-500 to-primary-600 text-white shadow-primary-500/25 hover:from-primary-600 hover:to-primary-700 hover:shadow-primary-500/40'
          ]"
        >
          <Icon
            v-if="status === 'connecting'"
            name="refresh"
            size="sm"
            class="animate-spin"
            :stroke-width="2"
          />
          <Icon v-else-if="status === 'idle'" name="play" size="sm" :stroke-width="2" />
          <Icon v-else name="refresh" size="sm" :stroke-width="2" />
          <span>
            {{
              status === 'connecting'
                ? t('admin.accounts.testing')
                : status === 'idle'
                  ? t('admin.accounts.startTest')
                  : t('admin.accounts.retry')
            }}
          </span>
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch, nextTick, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import TextArea from '@/components/common/TextArea.vue'
import { Icon } from '@/components/icons'
import { useClipboard } from '@/composables/useClipboard'
import { adminAPI } from '@/api/admin'
import type { Account, ClaudeModel } from '@/types'

const { t } = useI18n()
const { copyToClipboard } = useClipboard()

interface OutputLine {
  text: string
  class: string
  prefix?: string
  prefixClass?: string
}

interface PreviewImage {
  url: string
  mimeType?: string
}

const props = defineProps<{
  show: boolean
  account: Account | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const terminalRef = ref<HTMLElement | null>(null)
const status = ref<'idle' | 'connecting' | 'success' | 'error'>('idle')
const outputLines = ref<OutputLine[]>([])
const streamingContent = ref('')
const errorMessage = ref('')
const availableModels = ref<ClaudeModel[]>([])
const selectedModelId = ref('')
const testPrompt = ref('')
const loadingModels = ref(false)
let eventSource: EventSource | null = null
const generatedImages = ref<PreviewImage[]>([])
const prioritizedGeminiModels = ['gemini-3.1-flash-image', 'gemini-2.5-flash-image', 'gemini-2.5-flash', 'gemini-2.5-pro', 'gemini-3-flash-preview', 'gemini-3-pro-preview', 'gemini-2.0-flash']
const supportsGeminiImageTest = computed(() => {
  const modelID = selectedModelId.value.toLowerCase()
  if (!modelID.startsWith('gemini-') || !modelID.includes('-image')) return false
  return props.account?.platform === 'gemini' || (props.account?.platform === 'antigravity' && props.account?.type === 'apikey')
})

// Timer
const startTime = ref<number>(0)
const elapsedTime = ref('')
let timerInterval: ReturnType<typeof setInterval> | null = null

const startTimer = () => {
  startTime.value = Date.now()
  timerInterval = setInterval(() => {
    const ms = Date.now() - startTime.value
    const seconds = (ms / 1000).toFixed(1)
    elapsedTime.value = `${seconds}s`
  }, 100)
}

const stopTimer = () => {
  if (timerInterval) {
    clearInterval(timerInterval)
    timerInterval = null
  }
}

onUnmounted(() => {
  stopTimer()
  closeEventSource()
})

// Status-driven computed classes
const statusBorderClass = computed(() => {
  switch (status.value) {
    case 'connecting': return 'bg-gradient-to-r from-yellow-400 via-orange-400 to-yellow-400 animate-gradient-x'
    case 'success': return 'bg-gradient-to-r from-emerald-400 to-teal-400'
    case 'error': return 'bg-gradient-to-r from-red-400 to-rose-400'
    default: return 'bg-gradient-to-r from-gray-200 to-gray-300 dark:from-dark-500 dark:to-dark-400'
  }
})

const statusIconBgClass = computed(() => {
  switch (status.value) {
    case 'connecting': return 'bg-gradient-to-br from-yellow-500 to-orange-500'
    case 'success': return 'bg-gradient-to-br from-emerald-500 to-teal-500'
    case 'error': return 'bg-gradient-to-br from-red-500 to-rose-500'
    default: return 'bg-gradient-to-br from-primary-500 to-primary-600'
  }
})

const statusIconName = computed(() => {
  switch (status.value) {
    case 'connecting': return 'refresh'
    case 'success': return 'check'
    case 'error': return 'x'
    default: return 'play'
  }
})

const statusBadgeClass = computed(() => {
  switch (status.value) {
    case 'connecting': return 'bg-yellow-100 text-yellow-700 dark:bg-yellow-500/20 dark:text-yellow-400'
    case 'success': return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-400'
    case 'error': return 'bg-red-100 text-red-700 dark:bg-red-500/20 dark:text-red-400'
    default: return 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400'
  }
})

const statusDotClass = computed(() => {
  switch (status.value) {
    case 'connecting': return 'bg-yellow-500 animate-pulse'
    case 'success': return 'bg-emerald-500'
    case 'error': return 'bg-red-500'
    default: return 'bg-gray-400'
  }
})

const statusLabel = computed(() => {
  switch (status.value) {
    case 'connecting': return t('admin.accounts.testing')
    case 'success': return t('admin.accounts.testCompleted')
    case 'error': return 'Error'
    default: return account.value?.status || 'idle'
  }
})

const account = computed(() => props.account)

const sortTestModels = (models: ClaudeModel[]) => {
  const priorityMap = new Map(prioritizedGeminiModels.map((id, index) => [id, index]))
  return [...models].sort((a, b) => {
    const aPriority = priorityMap.get(a.id) ?? Number.MAX_SAFE_INTEGER
    const bPriority = priorityMap.get(b.id) ?? Number.MAX_SAFE_INTEGER
    if (aPriority !== bPriority) return aPriority - bPriority
    return 0
  })
}

watch(
  () => props.show,
  async (newVal) => {
    if (newVal && props.account) {
      testPrompt.value = ''
      resetState()
      await loadAvailableModels()
    } else {
      closeEventSource()
      stopTimer()
    }
  }
)

watch(selectedModelId, () => {
  if (supportsGeminiImageTest.value && !testPrompt.value.trim()) {
    testPrompt.value = t('admin.accounts.geminiImagePromptDefault')
  }
})

const loadAvailableModels = async () => {
  if (!props.account) return
  loadingModels.value = true
  selectedModelId.value = ''
  try {
    const models = await adminAPI.accounts.getAvailableModels(props.account.id)
    availableModels.value = props.account.platform === 'gemini' || props.account.platform === 'antigravity'
      ? sortTestModels(models)
      : models
    if (availableModels.value.length > 0) {
      if (props.account.platform === 'gemini') {
        selectedModelId.value = availableModels.value[0].id
      } else {
        const sonnetModel = availableModels.value.find((m) => m.id.includes('sonnet'))
        selectedModelId.value = sonnetModel?.id || availableModels.value[0].id
      }
    }
  } catch (error) {
    console.error('Failed to load available models:', error)
    availableModels.value = []
    selectedModelId.value = ''
  } finally {
    loadingModels.value = false
  }
}

const resetState = () => {
  status.value = 'idle'
  outputLines.value = []
  streamingContent.value = ''
  errorMessage.value = ''
  generatedImages.value = []
  elapsedTime.value = ''
  stopTimer()
}

const handleClose = () => {
  if (status.value === 'connecting') return
  closeEventSource()
  stopTimer()
  emit('close')
}

const closeEventSource = () => {
  if (eventSource) {
    eventSource.close()
    eventSource = null
  }
}

const addLine = (text: string, className: string = 'text-gray-300', prefix?: string, prefixClass?: string) => {
  outputLines.value.push({ text, class: className, prefix, prefixClass })
  scrollToBottom()
}

const scrollToBottom = async () => {
  await nextTick()
  if (terminalRef.value) {
    terminalRef.value.scrollTop = terminalRef.value.scrollHeight
  }
}

const startTest = async () => {
  if (!props.account || !selectedModelId.value) return
  resetState()
  status.value = 'connecting'
  startTimer()

  addLine(props.account.name, 'text-blue-600 dark:text-blue-400', '>', 'text-blue-500')
  addLine(`${props.account.type} / ${props.account.platform}`, 'text-gray-500 dark:text-gray-500', ' ', 'text-gray-400 dark:text-gray-600')
  addLine('', 'text-gray-300')

  closeEventSource()

  try {
    const url = `/api/v1/admin/accounts/${props.account.id}/test`
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${localStorage.getItem('auth_token')}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(
        {
              model_id: selectedModelId.value,
              prompt: supportsGeminiImageTest.value ? testPrompt.value.trim() : ''
            }
      )
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const reader = response.body?.getReader()
    if (!reader) {
      throw new Error('No response body')
    }

    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (line.startsWith('data: ')) {
          const jsonStr = line.slice(6).trim()
          if (jsonStr) {
            try {
              const event = JSON.parse(jsonStr)
              handleEvent(event)
            } catch (e) {
              console.error('Failed to parse SSE event:', e)
            }
          }
        }
      }
    }
  } catch (error: any) {
    status.value = 'error'
    stopTimer()
    errorMessage.value = error.message || 'Unknown error'
    addLine(errorMessage.value, 'text-red-600 dark:text-red-400', '!', 'text-red-500')
  }
}

const handleEvent = (event: {
  type: string
  text?: string
  model?: string
  success?: boolean
  error?: string
  image_url?: string
  mime_type?: string
}) => {
  switch (event.type) {
    case 'test_start':
      addLine(t('admin.accounts.connectedToApi'), 'text-emerald-600 dark:text-emerald-400', '+', 'text-emerald-500')
      if (event.model) {
        addLine(event.model, 'text-cyan-600 dark:text-cyan-400', '#', 'text-cyan-500')
      }
      addLine(
        supportsGeminiImageTest.value
            ? t('admin.accounts.sendingGeminiImageRequest')
            : t('admin.accounts.sendingTestMessage'),
        'text-gray-500 dark:text-gray-500',
        '~',
        'text-gray-400 dark:text-gray-600'
      )
      addLine('', 'text-gray-300')
      addLine(t('admin.accounts.response'), 'text-amber-600 dark:text-yellow-400', '>', 'text-amber-500 dark:text-yellow-500')
      break

    case 'content':
      if (event.text) {
        streamingContent.value += event.text
        scrollToBottom()
      }
      break

    case 'image':
      if (event.image_url) {
        generatedImages.value.push({
          url: event.image_url,
          mimeType: event.mime_type
        })
        addLine(t('admin.accounts.geminiImageReceived', { count: generatedImages.value.length }), 'text-purple-600 dark:text-purple-400', '*', 'text-purple-700 dark:text-purple-500')
      }
      break

    case 'test_complete':
      if (streamingContent.value) {
        addLine(streamingContent.value, 'text-emerald-600 dark:text-emerald-300')
        streamingContent.value = ''
      }
      stopTimer()
      if (event.success) {
        status.value = 'success'
      } else {
        status.value = 'error'
        errorMessage.value = event.error || 'Test failed'
      }
      break

    case 'error':
      status.value = 'error'
      stopTimer()
      errorMessage.value = event.error || 'Unknown error'
      if (streamingContent.value) {
        addLine(streamingContent.value, 'text-emerald-600 dark:text-emerald-300')
        streamingContent.value = ''
      }
      break
  }
}

const copyOutput = () => {
  const text = outputLines.value.map((l) => l.text).join('\n')
  copyToClipboard(text, t('admin.accounts.outputCopied'))
}
</script>

<style scoped>
@keyframes gradient-x {
  0%, 100% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
}

.animate-gradient-x {
  background-size: 200% 200%;
  animation: gradient-x 2s ease infinite;
}

.terminal-line-enter-active {
  transition: all 0.3s ease-out;
}

.terminal-line-enter-from {
  opacity: 0;
  transform: translateY(8px);
}

.terminal-line-item {
  min-height: 1.25rem;
}
</style>
