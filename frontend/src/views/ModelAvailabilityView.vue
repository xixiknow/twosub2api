<template>
  <AppLayout>
    <div class="availability-page px-2 md:px-6 py-6 max-w-[1400px] mx-auto">

      <!-- Header Bar -->
      <div class="header-bar mb-6 md:mb-8">
        <div class="header-bg"></div>
        <div class="header-noise"></div>
        <div class="header-glow header-glow-1"></div>
        <div class="header-glow header-glow-2"></div>
        <div class="header-glow header-glow-3"></div>
        <div class="relative z-10">
          <!-- Top row: icon + title + actions -->
          <div class="flex items-start sm:items-center justify-between gap-3">
            <div class="flex items-center gap-3 min-w-0">
              <div class="header-icon">
                <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M9.348 14.652a3.75 3.75 0 010-5.304m5.304 0a3.75 3.75 0 010 5.304m-7.425 2.121a6.75 6.75 0 010-9.546m9.546 0a6.75 6.75 0 010 9.546M5.106 18.894c-3.808-3.807-3.808-9.98 0-13.788m13.788 0c3.808 3.807 3.808 9.98 0 13.788M12 12h.008v.008H12V12zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0z" />
                </svg>
              </div>
              <div class="min-w-0">
                <h1 class="text-lg sm:text-xl font-semibold text-gray-900 dark:text-white tracking-tight truncate">
                  {{ t('availability.title') }}
                </h1>
                <p class="text-xs sm:text-sm text-gray-500 dark:text-gray-400 mt-0.5 hidden sm:block">
                  {{ t('availability.subtitle') }}
                </p>
              </div>
            </div>
            <div class="flex items-center gap-2 sm:gap-3 flex-shrink-0">
              <!-- Live indicator -->
              <div v-if="!loading && groups.length > 0" class="hidden sm:flex items-center gap-2 px-3 py-1.5 rounded-full bg-emerald-50 dark:bg-emerald-900/20 border border-emerald-200/60 dark:border-emerald-800/40">
                <span class="live-dot"></span>
                <span class="text-xs font-medium text-emerald-700 dark:text-emerald-400">{{ lastUpdated }}</span>
              </div>
              <button
                @click="refresh"
                :disabled="loading"
                class="refresh-btn"
              >
                <svg class="w-4 h-4" :class="{ 'animate-spin': loading }" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182" />
                </svg>
              </button>
            </div>
          </div>

          <!-- Mobile live indicator -->
          <div v-if="!loading && groups.length > 0" class="flex sm:hidden items-center gap-2 mt-3">
            <span class="live-dot"></span>
            <span class="text-xs font-medium text-emerald-700 dark:text-emerald-400">{{ lastUpdated }}</span>
          </div>
        </div>

        <!-- Stats Row -->
        <div v-if="!loading && groups.length > 0" class="relative z-10 mt-4 sm:mt-6 grid grid-cols-3 gap-2 sm:gap-3">
          <div class="stat-card">
            <div class="stat-icon stat-icon-total">
              <svg class="w-3.5 h-3.5 sm:w-4 sm:h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM3.75 15.75A2.25 2.25 0 016 13.5h2.25a2.25 2.25 0 012.25 2.25V18a2.25 2.25 0 01-2.25 2.25H6A2.25 2.25 0 013.75 18v-2.25zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6zM13.5 15.75a2.25 2.25 0 012.25-2.25H18a2.25 2.25 0 012.25 2.25V18A2.25 2.25 0 0118 20.25h-2.25A2.25 2.25 0 0113.5 18v-2.25z" />
              </svg>
            </div>
            <div class="stat-content">
              <span class="stat-value">{{ groups.length }}</span>
              <span class="stat-label">{{ t('availability.totalGroups') }}</span>
            </div>
          </div>
          <div class="stat-card">
            <div class="stat-icon stat-icon-online">
              <svg class="w-3.5 h-3.5 sm:w-4 sm:h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div class="stat-content">
              <span class="stat-value text-emerald-600 dark:text-emerald-400">{{ availableCount }}</span>
              <span class="stat-label">{{ t('availability.available') }}</span>
            </div>
          </div>
          <div class="stat-card">
            <div class="stat-icon stat-icon-offline">
              <svg class="w-3.5 h-3.5 sm:w-4 sm:h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
              </svg>
            </div>
            <div class="stat-content">
              <span class="stat-value" :class="unavailableCount > 0 ? 'text-red-500 dark:text-red-400' : 'text-gray-400 dark:text-gray-500'">{{ unavailableCount }}</span>
              <span class="stat-label">{{ t('availability.unavailable') }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="space-y-3 sm:space-y-4">
        <div class="flex gap-2 mb-4 sm:mb-6 overflow-x-auto no-scrollbar">
          <div v-for="i in 4" :key="i" class="h-7 sm:h-8 rounded-full animate-pulse flex-shrink-0" :class="i === 1 ? 'w-14 sm:w-16' : 'w-20 sm:w-24'" :style="{ background: 'var(--skeleton-bg)' }"></div>
        </div>
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-2.5 sm:gap-3">
          <div v-for="i in 8" :key="i" class="skeleton-card" :style="{ animationDelay: `${i * 80}ms` }">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 sm:w-9 sm:h-9 rounded-lg animate-pulse flex-shrink-0" style="background: var(--skeleton-bg)"></div>
              <div class="flex-1 space-y-2 min-w-0">
                <div class="h-3 sm:h-3.5 w-20 sm:w-24 rounded animate-pulse" style="background: var(--skeleton-bg)"></div>
                <div class="h-2.5 w-14 sm:w-16 rounded animate-pulse" style="background: var(--skeleton-bg)"></div>
              </div>
              <div class="w-11 sm:w-12 h-5 rounded-full animate-pulse flex-shrink-0" style="background: var(--skeleton-bg)"></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="error-state">
        <div class="error-icon-wrap">
          <svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m0-10.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z" />
          </svg>
        </div>
        <p class="text-sm text-gray-600 dark:text-gray-400 mt-3 mb-4">{{ error }}</p>
        <button @click="refresh" class="retry-btn">
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182" />
          </svg>
          {{ t('availability.retry') }}
        </button>
      </div>

      <!-- Empty State -->
      <div v-else-if="groups.length === 0" class="empty-state">
        <div class="empty-icon-wrap">
          <svg class="w-8 h-8 text-gray-300 dark:text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375m16.5 0v3.75m-16.5-3.75v3.75m16.5 0v3.75C20.25 16.153 16.556 18 12 18s-8.25-1.847-8.25-4.125v-3.75m16.5 0c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125" />
          </svg>
        </div>
        <p class="text-sm text-gray-400 dark:text-gray-500 mt-3">{{ t('availability.noGroups') }}</p>
      </div>

      <!-- Content -->
      <template v-if="!loading && !error && groups.length > 0">
        <!-- Platform Filter -->
        <div class="filter-bar mb-4 sm:mb-5">
          <button
            @click="selectedPlatform = ''"
            class="filter-chip"
            :class="selectedPlatform === '' ? 'filter-chip-active' : 'filter-chip-idle'"
          >
            {{ t('availability.all') }}
            <span class="chip-count">{{ groups.length }}</span>
          </button>
          <button
            v-for="platform in platforms"
            :key="platform.name"
            @click="selectedPlatform = platform.name"
            class="filter-chip"
            :class="selectedPlatform === platform.name ? 'filter-chip-active' : 'filter-chip-idle'"
          >
            <span class="chip-dot" :style="{ background: platformColor(platform.name) }"></span>
            {{ platformDisplayName(platform.name) }}
            <span class="chip-count">{{ platform.count }}</span>
          </button>
        </div>

        <!-- Uptime Summary Bar -->
        <div class="uptime-bar mb-4 sm:mb-5">
          <div class="uptime-track">
            <div
              class="uptime-fill"
              :style="{ width: uptimePercent + '%' }"
            ></div>
          </div>
          <span class="uptime-label">
            {{ uptimePercent }}% {{ t('availability.available') }}
          </span>
        </div>

        <!-- Group Cards Grid -->
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-2.5 sm:gap-3">
          <div
            v-for="(group, index) in filteredGroups"
            :key="group.group_id"
            class="group-card"
            :class="group.available ? 'card-ok' : 'card-down'"
            :style="{ animationDelay: `${index * 40}ms` }"
          >
            <!-- Left color accent -->
            <div class="card-accent" :class="group.available ? 'accent-ok' : 'accent-down'"></div>

            <div class="card-body">
              <div class="flex items-center gap-3">
                <!-- Platform icon circle -->
                <div class="platform-icon" :style="{ background: platformColor(group.platform) + '18', color: platformColor(group.platform) }">
                  <span class="text-xs font-bold">{{ platformDisplayName(group.platform).charAt(0) }}</span>
                </div>

                <!-- Info -->
                <div class="flex-1 min-w-0">
                  <h3 class="card-title">{{ group.group_name }}</h3>
                  <span class="card-platform" :style="{ color: platformColor(group.platform) }">
                    {{ platformDisplayName(group.platform) }}
                  </span>
                </div>

                <!-- Status pill -->
                <div class="status-pill" :class="group.available ? 'pill-ok' : 'pill-down'">
                  <span class="pill-dot" :class="group.available ? 'pill-dot-ok' : 'pill-dot-down'"></span>
                  {{ group.available ? t('availability.online') : t('availability.offline') }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>

    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
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
  gemini: '#3b82f6',
  antigravity: '#8b5cf6',
  sora: '#ec4899',
  qwen: '#f97316',
  deepseek: '#14b8a6',
  glm: '#06b6d4',
  kimi: '#a855f7',
  iflow: '#10b981'
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
/* === CSS Variables === */
.availability-page {
  --skeleton-bg: #e5e7eb;
}
.dark .availability-page {
  --skeleton-bg: rgba(75, 85, 99, 0.4);
}

/* === Header === */
.header-bar {
  position: relative;
  padding: 16px;
  border-radius: 12px;
  overflow: hidden;
}
@media (min-width: 640px) {
  .header-bar {
    padding: 24px;
    border-radius: 16px;
  }
}
.header-bg {
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  border: 1px solid #e2e8f0;
  border-radius: 16px;
}
.dark .header-bg {
  background: linear-gradient(135deg, rgba(30, 41, 59, 0.8) 0%, rgba(15, 23, 42, 0.9) 100%);
  border-color: rgba(51, 65, 85, 0.5);
}
.header-noise {
  position: absolute;
  inset: 0;
  opacity: 0.03;
  background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 256 256' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noise'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noise)'/%3E%3C/svg%3E");
}
.header-glow {
  position: absolute;
  border-radius: 50%;
  filter: blur(50px);
  pointer-events: none;
  display: none;
}
@media (min-width: 640px) {
  .header-glow {
    display: block;
  }
}
.header-glow-1 {
  width: 180px; height: 180px;
  background: rgba(16, 185, 129, 0.12);
  top: -60px; right: 10%;
  animation: drift 12s ease-in-out infinite;
}
.header-glow-2 {
  width: 120px; height: 120px;
  background: rgba(99, 102, 241, 0.08);
  bottom: -30px; left: 15%;
  animation: drift 15s ease-in-out infinite reverse;
}
.header-glow-3 {
  width: 100px; height: 100px;
  background: rgba(236, 72, 153, 0.06);
  top: 10px; left: 50%;
  animation: drift 18s ease-in-out infinite;
}
.dark .header-glow-1 { background: rgba(16, 185, 129, 0.08); }
.dark .header-glow-2 { background: rgba(99, 102, 241, 0.06); }
.dark .header-glow-3 { background: rgba(236, 72, 153, 0.04); }

@keyframes drift {
  0%, 100% { transform: translate(0, 0); }
  33% { transform: translate(10px, -8px); }
  66% { transform: translate(-5px, 5px); }
}

.header-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #10b981, #06b6d4);
  color: white;
  box-shadow: 0 4px 12px -2px rgba(16, 185, 129, 0.3);
  flex-shrink: 0;
}
@media (min-width: 640px) {
  .header-icon {
    width: 42px;
    height: 42px;
    border-radius: 12px;
  }
}

/* === Live Dot === */
.live-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #10b981;
  position: relative;
}
.live-dot::after {
  content: '';
  position: absolute;
  inset: -3px;
  border-radius: 50%;
  background: rgba(16, 185, 129, 0.3);
  animation: pulse-ring 2s ease-out infinite;
}
@keyframes pulse-ring {
  0% { transform: scale(1); opacity: 1; }
  100% { transform: scale(2.2); opacity: 0; }
}

/* === Refresh Button === */
.refresh-btn {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: white;
  border: 1px solid #e2e8f0;
  color: #64748b;
  transition: all 0.2s;
  cursor: pointer;
}
.refresh-btn:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #334155;
}
.refresh-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.dark .refresh-btn {
  background: rgba(51, 65, 85, 0.5);
  border-color: rgba(71, 85, 105, 0.5);
  color: #94a3b8;
}
.dark .refresh-btn:hover {
  background: rgba(71, 85, 105, 0.6);
  color: #e2e8f0;
}

/* === Stat Cards === */
.stat-card {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 10px;
  border-radius: 10px;
  background: white;
  border: 1px solid #f1f5f9;
  box-shadow: 0 1px 2px rgba(0,0,0,0.03);
}
@media (min-width: 640px) {
  .stat-card {
    gap: 10px;
    padding: 12px 14px;
    border-radius: 12px;
  }
}
.dark .stat-card {
  background: rgba(30, 41, 59, 0.6);
  border-color: rgba(51, 65, 85, 0.4);
}
.stat-icon {
  width: 28px;
  height: 28px;
  border-radius: 7px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
@media (min-width: 640px) {
  .stat-icon {
    width: 32px;
    height: 32px;
    border-radius: 8px;
  }
}
.stat-icon-total {
  background: #eff6ff;
  color: #3b82f6;
}
.dark .stat-icon-total {
  background: rgba(59, 130, 246, 0.12);
}
.stat-icon-online {
  background: #ecfdf5;
  color: #10b981;
}
.dark .stat-icon-online {
  background: rgba(16, 185, 129, 0.12);
}
.stat-icon-offline {
  background: #fef2f2;
  color: #ef4444;
}
.dark .stat-icon-offline {
  background: rgba(239, 68, 68, 0.12);
}
.stat-content {
  display: flex;
  flex-direction: column;
}
.stat-value {
  font-size: 1.1rem;
  font-weight: 700;
  line-height: 1.2;
  color: #0f172a;
}
@media (min-width: 640px) {
  .stat-value {
    font-size: 1.25rem;
  }
}
.dark .stat-value {
  color: #f1f5f9;
}
.stat-label {
  font-size: 0.6rem;
  color: #94a3b8;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
@media (min-width: 640px) {
  .stat-label {
    font-size: 0.7rem;
  }
}

/* === Skeleton === */
.skeleton-card {
  padding: 14px 16px;
  border-radius: 12px;
  background: white;
  border: 1px solid #f1f5f9;
  animation: skeleton-in 0.4s ease-out both;
}
.dark .skeleton-card {
  background: rgba(30, 41, 59, 0.4);
  border-color: rgba(51, 65, 85, 0.3);
}
@keyframes skeleton-in {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

/* === Error & Empty States === */
.error-state, .empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 64px 24px;
}
.error-icon-wrap {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fef2f2;
  border: 1px solid #fecaca;
}
.dark .error-icon-wrap {
  background: rgba(239, 68, 68, 0.1);
  border-color: rgba(239, 68, 68, 0.2);
}
.empty-icon-wrap {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
}
.dark .empty-icon-wrap {
  background: rgba(51, 65, 85, 0.3);
  border-color: rgba(71, 85, 105, 0.3);
}
.retry-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 7px 16px;
  border-radius: 8px;
  font-size: 0.8125rem;
  font-weight: 500;
  background: #0f172a;
  color: white;
  transition: all 0.15s;
  cursor: pointer;
  border: none;
}
.retry-btn:hover { background: #1e293b; }
.dark .retry-btn { background: #e2e8f0; color: #0f172a; }
.dark .retry-btn:hover { background: #f1f5f9; }

/* === Filter Bar === */
.filter-bar {
  display: flex;
  flex-wrap: nowrap;
  gap: 6px;
  align-items: center;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
  -ms-overflow-style: none;
  padding-bottom: 2px;
}
.filter-bar::-webkit-scrollbar {
  display: none;
}
@media (min-width: 640px) {
  .filter-bar {
    flex-wrap: wrap;
    overflow-x: visible;
  }
}
.filter-chip {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 5px 10px;
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
  .filter-chip {
    padding: 5px 12px;
    font-size: 0.8125rem;
  }
}
.filter-chip-active {
  background: #0f172a;
  color: white;
  border-color: #0f172a;
}
.dark .filter-chip-active {
  background: #e2e8f0;
  color: #0f172a;
  border-color: #e2e8f0;
}
.filter-chip-idle {
  background: white;
  color: #475569;
  border-color: #e2e8f0;
}
.filter-chip-idle:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
}
.dark .filter-chip-idle {
  background: rgba(30, 41, 59, 0.5);
  color: #cbd5e1;
  border-color: rgba(51, 65, 85, 0.5);
}
.dark .filter-chip-idle:hover {
  background: rgba(51, 65, 85, 0.6);
  border-color: rgba(71, 85, 105, 0.6);
}
.chip-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}
.chip-count {
  font-size: 0.7rem;
  padding: 0 5px;
  border-radius: 4px;
  background: rgba(0,0,0,0.06);
  line-height: 1.5;
}
.filter-chip-active .chip-count {
  background: rgba(255,255,255,0.2);
}
.dark .filter-chip-active .chip-count {
  background: rgba(0,0,0,0.15);
}
.dark .filter-chip-idle .chip-count {
  background: rgba(255,255,255,0.06);
}

/* === Uptime Bar === */
.uptime-bar {
  display: flex;
  align-items: center;
  gap: 12px;
}
.uptime-track {
  flex: 1;
  height: 6px;
  border-radius: 3px;
  background: #f1f5f9;
  overflow: hidden;
}
.dark .uptime-track {
  background: rgba(51, 65, 85, 0.4);
}
.uptime-fill {
  height: 100%;
  border-radius: 3px;
  background: linear-gradient(90deg, #10b981, #34d399);
  transition: width 0.8s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
}
.uptime-fill::after {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,0.3), transparent);
  animation: shimmer 2s ease-in-out infinite;
}
@keyframes shimmer {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}
.uptime-label {
  font-size: 0.7rem;
  font-weight: 600;
  color: #10b981;
  white-space: nowrap;
  min-width: 70px;
  text-align: right;
}
@media (min-width: 640px) {
  .uptime-label {
    font-size: 0.75rem;
    min-width: 80px;
  }
}
.dark .uptime-label {
  color: #34d399;
}

/* === Group Cards === */
.group-card {
  position: relative;
  border-radius: 12px;
  overflow: hidden;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  animation: card-in 0.35s ease-out both;
}
@keyframes card-in {
  from { opacity: 0; transform: translateY(6px); }
  to { opacity: 1; transform: translateY(0); }
}
.group-card:hover {
  transform: translateY(-1px);
}
.card-ok {
  background: white;
  border: 1px solid #e2e8f0;
}
.card-ok:hover {
  border-color: #a7f3d0;
  box-shadow: 0 4px 16px -4px rgba(16, 185, 129, 0.12);
}
.dark .card-ok {
  background: rgba(30, 41, 59, 0.5);
  border-color: rgba(51, 65, 85, 0.4);
}
.dark .card-ok:hover {
  border-color: rgba(16, 185, 129, 0.3);
  box-shadow: 0 4px 16px -4px rgba(16, 185, 129, 0.15);
}
.card-down {
  background: white;
  border: 1px solid #fecdd3;
}
.card-down:hover {
  border-color: #fca5a5;
  box-shadow: 0 4px 16px -4px rgba(239, 68, 68, 0.1);
}
.dark .card-down {
  background: rgba(30, 41, 59, 0.5);
  border-color: rgba(239, 68, 68, 0.2);
}
.dark .card-down:hover {
  border-color: rgba(239, 68, 68, 0.35);
  box-shadow: 0 4px 16px -4px rgba(239, 68, 68, 0.12);
}

.card-accent {
  position: absolute;
  top: 0;
  left: 0;
  width: 3px;
  height: 100%;
}
.accent-ok {
  background: linear-gradient(180deg, #10b981, #06b6d4);
}
.accent-down {
  background: linear-gradient(180deg, #ef4444, #f97316);
}

.card-body {
  padding: 12px 14px 12px 16px;
}
@media (min-width: 640px) {
  .card-body {
    padding: 14px 16px 14px 20px;
  }
}

.platform-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
@media (min-width: 640px) {
  .platform-icon {
    width: 36px;
    height: 36px;
    border-radius: 10px;
  }
}

.card-title {
  font-size: 0.8125rem;
  font-weight: 600;
  color: #1e293b;
  line-height: 1.3;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
@media (min-width: 640px) {
  .card-title {
    font-size: 0.875rem;
  }
}
.dark .card-title {
  color: #f1f5f9;
}

.card-platform {
  font-size: 0.7rem;
  font-weight: 500;
  opacity: 0.8;
}

/* === Status Pill === */
.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 10px;
  border-radius: 6px;
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.01em;
  flex-shrink: 0;
}
.pill-ok {
  background: #ecfdf5;
  color: #059669;
}
.dark .pill-ok {
  background: rgba(16, 185, 129, 0.12);
  color: #34d399;
}
.pill-down {
  background: #fef2f2;
  color: #dc2626;
}
.dark .pill-down {
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
}
.pill-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
}
.pill-dot-ok {
  background: #10b981;
  box-shadow: 0 0 4px rgba(16, 185, 129, 0.5);
}
.pill-dot-down {
  background: #ef4444;
  box-shadow: 0 0 4px rgba(239, 68, 68, 0.5);
}
</style>
