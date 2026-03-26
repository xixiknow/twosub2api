<template>
  <AppLayout>
    <div class="avail-page px-2 md:px-6 py-4 sm:py-6 max-w-[1400px] mx-auto">

      <!-- Page Header -->
      <div class="mb-5 sm:mb-8 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <div>
          <h1 class="text-xl sm:text-2xl font-bold tracking-tight text-foreground">
            {{ t('availability.title') }}
          </h1>
          <p class="text-xs sm:text-sm text-muted-fg mt-1">
            {{ t('availability.subtitle') }}
          </p>
        </div>
        <div class="flex items-center gap-2.5">
          <!-- Live badge -->
          <div v-if="!loading && groups.length > 0" class="flex items-center gap-2 rounded-full border border-border/40 bg-background/60 backdrop-blur-md px-3 py-1.5">
            <span class="live-dot"></span>
            <span class="text-[11px] font-medium text-muted-fg">{{ lastUpdated }}</span>
          </div>
          <button @click="refresh" :disabled="loading"
            class="inline-flex items-center gap-1.5 rounded-lg border border-border/40 bg-background/60 backdrop-blur-md px-3 py-1.5 text-xs font-medium text-muted-fg transition-colors hover:bg-muted/50 hover:text-foreground disabled:opacity-50">
            <svg class="h-3.5 w-3.5" :class="{ 'animate-spin': loading }" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182" />
            </svg>
            {{ t('availability.refresh') }}
          </button>
        </div>
      </div>

      <!-- Summary Stats -->
      <div v-if="!loading && groups.length > 0" class="mb-5 sm:mb-6 flex flex-wrap items-center gap-4 sm:gap-6">
        <div class="flex items-center gap-2">
          <span class="text-2xl sm:text-3xl font-bold tabular-nums text-foreground">{{ uptimePercent }}%</span>
          <span class="text-xs font-medium text-emerald-600 dark:text-emerald-400">{{ t('availability.available') }}</span>
        </div>
        <div class="h-5 w-px bg-border/60 hidden sm:block"></div>
        <div class="flex items-center gap-3 text-xs text-muted-fg">
          <span>{{ groups.length }} {{ t('availability.totalGroups') }}</span>
          <span class="text-emerald-600 dark:text-emerald-400">{{ availableCount }} {{ t('availability.online') }}</span>
          <span v-if="unavailableCount > 0" class="text-red-500 dark:text-red-400">{{ unavailableCount }} {{ t('availability.offline') }}</span>
        </div>
      </div>

      <!-- Uptime bar -->
      <div v-if="!loading && groups.length > 0" class="mb-5 sm:mb-6">
        <div class="h-1.5 w-full rounded-full bg-muted/30 overflow-hidden">
          <div class="h-full rounded-full bg-emerald-500 transition-all duration-700" :style="{ width: uptimePercent + '%' }"></div>
        </div>
      </div>

      <!-- Platform Filters -->
      <div v-if="!loading && groups.length > 0" class="filter-scroll mb-5 sm:mb-6">
        <button @click="selectedPlatform = ''" class="filter-chip" :class="selectedPlatform === '' ? 'chip-on' : 'chip-off'">
          {{ t('availability.all') }}
          <span class="chip-num">{{ groups.length }}</span>
        </button>
        <button v-for="p in platforms" :key="p.name" @click="selectedPlatform = p.name"
          class="filter-chip" :class="selectedPlatform === p.name ? 'chip-on' : 'chip-off'">
          <span class="h-1.5 w-1.5 rounded-full flex-shrink-0" :style="{ background: platformColor(p.name) }"></span>
          {{ platformDisplayName(p.name) }}
          <span class="chip-num">{{ p.count }}</span>
        </button>
      </div>

      <!-- Loading Skeleton -->
      <div v-if="loading" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 sm:gap-4">
        <div v-for="i in 6" :key="i" class="skel-card" :style="{ animationDelay: `${i * 60}ms` }">
          <div class="p-4 sm:p-5">
            <div class="flex items-center gap-3 mb-4">
              <div class="h-10 w-10 sm:h-12 sm:w-12 rounded-xl shimmer"></div>
              <div class="flex-1 space-y-2">
                <div class="h-4 w-28 rounded shimmer"></div>
                <div class="h-3 w-20 rounded shimmer"></div>
              </div>
            </div>
            <div class="h-8 w-full rounded-lg shimmer"></div>
          </div>
        </div>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="flex flex-col items-center justify-center py-16 px-4">
        <div class="flex h-14 w-14 items-center justify-center rounded-2xl bg-red-100 dark:bg-red-500/10 mb-4">
          <svg class="h-7 w-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
          </svg>
        </div>
        <p class="text-sm text-muted-fg mb-4">{{ error }}</p>
        <button @click="refresh" class="rounded-lg bg-foreground px-4 py-2 text-xs font-medium text-background transition-opacity hover:opacity-90">
          {{ t('availability.retry') }}
        </button>
      </div>

      <!-- Empty State -->
      <div v-else-if="groups.length === 0" class="flex flex-col items-center justify-center py-16 px-4">
        <div class="flex h-14 w-14 items-center justify-center rounded-2xl bg-muted/50 mb-4">
          <svg class="h-7 w-7 text-muted-fg/50" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375" />
          </svg>
        </div>
        <p class="text-sm text-muted-fg">{{ t('availability.noGroups') }}</p>
      </div>

      <!-- Group Cards -->
      <div v-if="!loading && !error && filteredGroups.length > 0"
        class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 sm:gap-4">
        <div v-for="(group, idx) in filteredGroups" :key="group.group_id"
          class="group card-base"
          :style="{ animationDelay: `${idx * 30}ms` }">
          <!-- Corner crosses (hover only) -->
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1"
            class="pointer-events-none absolute left-2 top-2 h-3.5 w-3.5 text-muted-fg/30 opacity-0 transition-opacity group-hover:opacity-100">
            <line x1="12" y1="0" x2="12" y2="24"/><line x1="0" y1="12" x2="24" y2="12"/>
          </svg>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1"
            class="pointer-events-none absolute right-2 top-2 h-3.5 w-3.5 text-muted-fg/30 opacity-0 transition-opacity group-hover:opacity-100">
            <line x1="12" y1="0" x2="12" y2="24"/><line x1="0" y1="12" x2="24" y2="12"/>
          </svg>

          <div class="flex-1 p-4 sm:p-5">
            <!-- Top: Icon + Name + Status -->
            <div class="mb-3 sm:mb-4 flex items-start justify-between">
              <div class="flex min-w-0 flex-1 items-center gap-3">
                <!-- Platform icon -->
                <div class="relative flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-white/80 to-white/20 shadow-sm ring-1 ring-black/5 transition-transform group-hover:scale-105 dark:from-white/10 dark:to-white/5 dark:ring-white/10 sm:h-12 sm:w-12 sm:rounded-2xl">
                  <div class="scale-75 sm:scale-100">
                    <ModelIcon :model="platformToModel(group.platform)" size="26px" />
                  </div>
                </div>
                <!-- Name + platform label -->
                <div class="min-w-0 flex-1">
                  <div class="flex items-center justify-between gap-2">
                    <h3 class="flex-1 truncate text-sm font-bold leading-none tracking-tight text-foreground sm:text-base">
                      {{ group.group_name }}
                    </h3>
                    <!-- Status badge -->
                    <div class="shrink-0 whitespace-nowrap rounded-lg px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider shadow-sm backdrop-blur-md border border-transparent sm:px-2.5 sm:py-1 sm:text-xs"
                      :class="group.available
                        ? 'bg-green-100 text-green-700 dark:bg-green-500/15 dark:text-green-400'
                        : 'bg-red-100 text-red-600 dark:bg-red-500/15 dark:text-red-400'">
                      {{ group.available ? t('availability.online') : t('availability.offline') }}
                    </div>
                  </div>
                  <div class="mt-1.5 flex flex-wrap items-center gap-2 text-xs text-muted-fg">
                    <span class="inline-flex shrink-0 items-center gap-1 rounded-md bg-muted/50 px-1.5 py-0.5 font-medium text-muted-fg/80">
                      {{ platformDisplayName(group.platform) }}
                    </span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Status indicator bar -->
            <div class="flex items-center justify-between rounded-lg bg-muted/30 px-3 py-2 transition-colors group-hover:bg-muted/50">
              <div class="flex items-center gap-2 text-muted-fg">
                <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M16.247 7.761a6 6 0 0 1 0 8.478"/><path d="M19.075 4.933a10 10 0 0 1 0 14.134"/>
                  <path d="M4.925 19.067a10 10 0 0 1 0-14.134"/><path d="M7.753 16.239a6 6 0 0 1 0-8.478"/>
                  <circle cx="12" cy="12" r="2"/>
                </svg>
                <span class="text-[10px] font-semibold uppercase tracking-wider">{{ t('availability.available') }}</span>
              </div>
              <div class="flex items-center gap-1.5">
                <span class="h-1.5 w-1.5 rounded-full" :class="group.available ? 'bg-emerald-500' : 'bg-red-500'"></span>
                <span class="text-xs font-medium" :class="group.available ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-500 dark:text-red-400'">
                  {{ group.available ? t('availability.online') : t('availability.offline') }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import ModelIcon from '@/components/common/ModelIcon.vue'
import { getAvailability, type GroupAvailabilityItem } from '@/api/groups'

const { t } = useI18n()

const groups = ref<GroupAvailabilityItem[]>([])
const loading = ref(true)
const error = ref('')
const selectedPlatform = ref('')
const lastUpdated = ref('')

const availableCount = computed(() => groups.value.filter(g => g.available).length)
const unavailableCount = computed(() => groups.value.filter(g => !g.available).length)
const uptimePercent = computed(() => {
  if (groups.value.length === 0) return 100
  return Math.round((availableCount.value / groups.value.length) * 100)
})

const platforms = computed(() => {
  const map = new Map<string, number>()
  for (const g of groups.value) {
    map.set(g.platform, (map.get(g.platform) || 0) + 1)
  }
  return Array.from(map.entries())
    .map(([name, count]) => ({ name, count }))
    .sort((a, b) => b.count - a.count)
})

const filteredGroups = computed(() => {
  if (!selectedPlatform.value) return groups.value
  return groups.value.filter(g => g.platform === selectedPlatform.value)
})

const PLATFORM_COLORS: Record<string, string> = {
  anthropic: '#d97706',
  openai: '#6366f1',
  gemini: '#4285f4',
  antigravity: '#8b5cf6',
  sora: '#ec4899',
  qwen: '#615eff',
  deepseek: '#4d6bfe',
  glm: '#3859ff',
  kimi: '#a855f7',
  iflow: '#10b981'
}

// Map platform names to model prefixes so ModelIcon can resolve them
const PLATFORM_MODEL_MAP: Record<string, string> = {
  anthropic: 'claude-3',
  openai: 'gpt-4',
  gemini: 'gemini-pro',
  antigravity: 'antigravity',
  sora: 'sora',
  qwen: 'qwen-max',
  deepseek: 'deepseek-chat',
  glm: 'glm-4',
  kimi: 'moonshot-v1',
  iflow: 'iflow'
}

function platformDisplayName(platform: string): string {
  const names: Record<string, string> = {
    anthropic: 'Anthropic',
    openai: 'OpenAI',
    gemini: 'Gemini',
    antigravity: 'Antigravity',
    sora: 'Sora',
    qwen: 'Qwen',
    deepseek: 'DeepSeek',
    glm: 'GLM',
    kimi: 'Kimi',
    iflow: 'iFlow'
  }
  return names[platform] || platform
}

function platformColor(platform: string): string {
  return PLATFORM_COLORS[platform] || '#6b7280'
}

function platformToModel(platform: string): string {
  return PLATFORM_MODEL_MAP[platform] || platform
}

async function refresh() {
  loading.value = true
  error.value = ''
  try {
    groups.value = await getAvailability()
    lastUpdated.value = new Date().toLocaleTimeString()
  } catch (e: any) {
    error.value = e?.message || 'Failed to load availability data'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  refresh()
})
</script>

<style scoped>
/* === Semantic color tokens === */
.avail-page {
  --foreground: #0f172a;
  --background: #ffffff;
  --muted: #f1f5f9;
  --muted-fg: #64748b;
  --border: #e2e8f0;
}
.dark .avail-page {
  --foreground: #f1f5f9;
  --background: #0f172a;
  --muted: rgba(51, 65, 85, 0.5);
  --muted-fg: #94a3b8;
  --border: rgba(51, 65, 85, 0.6);
}
.text-foreground { color: var(--foreground); }
.text-background { color: var(--background); }
.bg-foreground { background: var(--foreground); }
.text-muted-fg { color: var(--muted-fg); }

/* === Live dot === */
.live-dot {
  width: 6px; height: 6px;
  border-radius: 50%;
  background: #10b981;
  position: relative;
  flex-shrink: 0;
}
.live-dot::after {
  content: '';
  position: absolute;
  inset: -3px;
  border-radius: 50%;
  background: rgba(16, 185, 129, 0.3);
  animation: ping 2s ease-out infinite;
}
@keyframes ping {
  0% { transform: scale(1); opacity: 1; }
  100% { transform: scale(2.5); opacity: 0; }
}

/* === Filter scroll === */
.filter-scroll {
  display: flex;
  gap: 6px;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
  padding-bottom: 2px;
}
.filter-scroll::-webkit-scrollbar { display: none; }
@media (min-width: 640px) {
  .filter-scroll { flex-wrap: wrap; overflow-x: visible; }
}

.filter-chip {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 5px 11px;
  border-radius: 8px;
  font-size: 0.75rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
  border: 1px solid transparent;
  white-space: nowrap;
  flex-shrink: 0;
}
@media (min-width: 640px) {
  .filter-chip { padding: 6px 14px; font-size: 0.8125rem; }
}
.chip-on {
  background: var(--foreground);
  color: var(--background);
  border-color: var(--foreground);
}
.chip-off {
  background: var(--background);
  color: var(--muted-fg);
  border-color: var(--border);
}
.chip-off:hover {
  background: var(--muted);
  color: var(--foreground);
}
.chip-num {
  font-size: 0.65rem;
  padding: 0 5px;
  border-radius: 4px;
  background: rgba(0,0,0,0.06);
  line-height: 1.6;
}
.chip-on .chip-num { background: rgba(255,255,255,0.18); }
.dark .chip-off .chip-num { background: rgba(255,255,255,0.06); }

/* === Card base (reference style) === */
.card-base {
  position: relative;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-radius: 16px;
  border: 1px solid var(--border);
  background: rgba(255,255,255,0.4);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  animation: card-up 0.35s ease-out both;
}
.dark .card-base {
  background: rgba(15, 23, 42, 0.4);
  border-color: rgba(51, 65, 85, 0.4);
}
.card-base:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 30px -8px rgba(0,0,0,0.08);
  border-color: rgba(99, 102, 241, 0.2);
}
.dark .card-base:hover {
  box-shadow: 0 8px 30px -8px rgba(0,0,0,0.3);
  border-color: rgba(99, 102, 241, 0.15);
}
@keyframes card-up {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

/* === Skeleton === */
.skel-card {
  border-radius: 16px;
  border: 1px solid var(--border);
  background: rgba(255,255,255,0.4);
  backdrop-filter: blur(16px);
  animation: card-up 0.35s ease-out both;
}
.dark .skel-card {
  background: rgba(15, 23, 42, 0.4);
  border-color: rgba(51, 65, 85, 0.4);
}
.shimmer {
  background: linear-gradient(90deg, var(--muted) 0%, rgba(255,255,255,0.4) 50%, var(--muted) 100%);
  background-size: 200% 100%;
  animation: shimmer-move 1.5s ease-in-out infinite;
}
.dark .shimmer {
  background: linear-gradient(90deg, rgba(51,65,85,0.4) 0%, rgba(71,85,105,0.3) 50%, rgba(51,65,85,0.4) 100%);
  background-size: 200% 100%;
}
@keyframes shimmer-move {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

/* Muted background utility */
.bg-muted\/30 { background: rgba(241,245,249,0.3); }
.dark .bg-muted\/30 { background: rgba(51,65,85,0.2); }
.bg-muted\/50 { background: rgba(241,245,249,0.5); }
.dark .bg-muted\/50 { background: rgba(51,65,85,0.35); }
</style>
