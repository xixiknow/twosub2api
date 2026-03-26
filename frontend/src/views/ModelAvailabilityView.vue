<template>
  <AppLayout>
    <div class="avail-page">
      <!-- Animated circuit background -->
      <div class="circuit-bg"></div>
      <div class="circuit-grid"></div>
      <div class="scanline-overlay"></div>

      <div class="relative z-10 px-2 md:px-6 py-4 sm:py-6 max-w-[1400px] mx-auto">

        <!-- ═══════════ COMMAND CENTER HERO ═══════════ -->
        <div class="hero-panel">
          <!-- Animated corner brackets -->
          <div class="corner-bracket corner-tl"></div>
          <div class="corner-bracket corner-tr"></div>
          <div class="corner-bracket corner-bl"></div>
          <div class="corner-bracket corner-br"></div>

          <!-- Holographic noise layer -->
          <div class="holo-noise"></div>
          <!-- Sweep beam -->
          <div class="sweep-beam"></div>
          <!-- Hex grid -->
          <div class="hex-grid"></div>
          <!-- Floating particles -->
          <div class="particle particle-1"></div>
          <div class="particle particle-2"></div>
          <div class="particle particle-3"></div>
          <div class="particle particle-4"></div>
          <div class="particle particle-5"></div>

          <div class="relative z-10 p-4 sm:p-8">
            <!-- Header row -->
            <div class="flex items-start sm:items-center justify-between gap-3">
              <div class="flex items-center gap-3 sm:gap-4">
                <!-- Radar pulse icon -->
                <div class="radar-wrap">
                  <div class="radar-ring radar-ring-1"></div>
                  <div class="radar-ring radar-ring-2"></div>
                  <div class="radar-ring radar-ring-3"></div>
                  <div class="radar-core">
                    <svg class="w-5 h-5 sm:w-6 sm:h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M9.348 14.652a3.75 3.75 0 010-5.304m5.304 0a3.75 3.75 0 010 5.304m-7.425 2.121a6.75 6.75 0 010-9.546m9.546 0a6.75 6.75 0 010 9.546M5.106 18.894c-3.808-3.807-3.808-9.98 0-13.788m13.788 0c3.808 3.807 3.808 9.98 0 13.788M12 12h.008v.008H12V12zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0z" />
                    </svg>
                  </div>
                </div>
                <div>
                  <div class="flex items-center gap-2">
                    <h1 class="hero-title">
                      {{ t('availability.title') }}
                    </h1>
                    <span class="sys-tag">SYS</span>
                  </div>
                  <p class="hero-sub">
                    <span class="terminal-cursor"></span>
                    {{ t('availability.subtitle') }}
                  </p>
                </div>
              </div>
              <div class="flex items-center gap-2">
                <div v-if="!loading && groups.length > 0" class="hidden sm:flex items-center gap-2 live-badge">
                  <span class="live-dot"></span>
                  <span class="live-text">LIVE</span>
                  <span class="live-time">{{ lastUpdated }}</span>
                </div>
                <button @click="refresh" :disabled="loading" class="refresh-btn">
                  <svg class="h-3.5 w-3.5" :class="{ 'animate-spin': loading }" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182" />
                  </svg>
                  {{ t('availability.refresh') }}
                </button>
              </div>
            </div>

            <!-- Stat cards with neon borders -->
            <div v-if="!loading && groups.length > 0" class="mt-5 sm:mt-8 grid grid-cols-3 gap-2 sm:gap-4">
              <div class="stat-cell stat-cyan">
                <div class="stat-value">{{ groups.length }}</div>
                <div class="stat-label">{{ t('availability.totalGroups') }}</div>
                <div class="stat-bar-bg"><div class="stat-bar-fill stat-bar-cyan" style="width:100%"></div></div>
              </div>
              <div class="stat-cell stat-green">
                <div class="stat-value text-neon-green">{{ availableCount }}</div>
                <div class="stat-label text-neon-green/60">{{ t('availability.online') }}</div>
                <div class="stat-bar-bg"><div class="stat-bar-fill stat-bar-green" :style="{ width: uptimePercent + '%' }"></div></div>
              </div>
              <div class="stat-cell stat-red">
                <div class="stat-value" :class="unavailableCount > 0 ? 'text-neon-red' : 'text-white/20'">{{ unavailableCount }}</div>
                <div class="stat-label" :class="unavailableCount > 0 ? 'text-neon-red/60' : 'text-white/20'">{{ t('availability.offline') }}</div>
                <div class="stat-bar-bg"><div class="stat-bar-fill stat-bar-red" :style="{ width: groups.length ? ((unavailableCount / groups.length) * 100) + '%' : '0%' }"></div></div>
              </div>
            </div>

            <!-- Uptime holographic bar -->
            <div v-if="!loading && groups.length > 0" class="mt-4 sm:mt-6">
              <div class="flex items-center justify-between mb-2">
                <span class="text-[10px] font-mono font-bold uppercase tracking-[0.2em] text-cyan-400/50">{{ t('availability.available') }}</span>
                <span class="text-sm font-mono font-black tabular-nums text-neon-green">{{ uptimePercent }}%</span>
              </div>
              <div class="uptime-track">
                <div class="uptime-fill" :style="{ width: uptimePercent + '%' }"></div>
                <div class="uptime-glow" :style="{ left: uptimePercent + '%' }"></div>
              </div>
            </div>

            <!-- Mobile live indicator -->
            <div v-if="!loading && groups.length > 0" class="mt-3 flex sm:hidden items-center gap-2">
              <span class="live-dot"></span>
              <span class="live-time">{{ lastUpdated }}</span>
            </div>
          </div>
        </div>

        <!-- ═══════════ PLATFORM FILTERS ═══════════ -->
        <div v-if="!loading && groups.length > 0" class="filter-scroll mb-5 sm:mb-6">
          <button @click="selectedPlatform = ''" class="neon-chip" :class="selectedPlatform === '' ? 'neon-chip-active' : 'neon-chip-idle'">
            <span class="chip-glow"></span>
            {{ t('availability.all') }}
            <span class="chip-count">{{ groups.length }}</span>
          </button>
          <button v-for="p in platforms" :key="p.name" @click="selectedPlatform = p.name"
            class="neon-chip" :class="selectedPlatform === p.name ? 'neon-chip-active' : 'neon-chip-idle'">
            <span class="chip-glow"></span>
            <span class="h-1.5 w-1.5 rounded-full flex-shrink-0" :style="{ background: platformColor(p.name), boxShadow: '0 0 6px ' + platformColor(p.name) }"></span>
            {{ platformDisplayName(p.name) }}
            <span class="chip-count">{{ p.count }}</span>
          </button>
        </div>

        <!-- ═══════════ LOADING STATE ═══════════ -->
        <div v-if="loading" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 sm:gap-4">
          <div v-for="i in 6" :key="i" class="skel-card" :style="{ animationDelay: `${i * 80}ms` }">
            <div class="skel-scan"></div>
            <div class="p-4 sm:p-5">
              <div class="flex items-center gap-3 mb-4">
                <div class="h-10 w-10 sm:h-12 sm:w-12 rounded-xl skel-block"></div>
                <div class="flex-1 space-y-2">
                  <div class="h-4 w-28 rounded skel-block"></div>
                  <div class="h-3 w-20 rounded skel-block"></div>
                </div>
              </div>
              <div class="space-y-2">
                <div class="h-8 w-full rounded-lg skel-block"></div>
                <div class="h-3 w-full rounded skel-block"></div>
              </div>
            </div>
          </div>
        </div>

        <!-- ═══════════ ERROR STATE ═══════════ -->
        <div v-else-if="error" class="flex flex-col items-center justify-center py-16 px-4">
          <div class="err-hex">
            <svg class="h-7 w-7 text-neon-red" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
            </svg>
          </div>
          <p class="text-sm font-mono text-red-400/80 mb-4 mt-3">{{ error }}</p>
          <button @click="refresh" class="retry-btn">
            {{ t('availability.retry') }}
          </button>
        </div>

        <!-- ═══════════ EMPTY STATE ═══════════ -->
        <div v-else-if="groups.length === 0" class="flex flex-col items-center justify-center py-16 px-4">
          <div class="empty-hex">
            <svg class="h-7 w-7 text-cyan-500/40" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375" />
            </svg>
          </div>
          <p class="text-sm font-mono text-cyan-400/50">{{ t('availability.noGroups') }}</p>
        </div>

        <!-- ═══════════ GROUP CARDS ═══════════ -->
        <div v-if="!loading && !error && filteredGroups.length > 0"
          class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 sm:gap-4">
          <div v-for="(group, idx) in filteredGroups" :key="group.group_id"
            class="group cyber-card"
            :class="group.available ? 'cyber-card-ok' : 'cyber-card-err'"
            :style="{ animationDelay: `${idx * 40}ms` }">

            <!-- Animated border gradient -->
            <div class="card-border-glow" :class="group.available ? 'border-glow-ok' : 'border-glow-err'"></div>

            <!-- HUD corner marks -->
            <div class="hud-corner hud-tl"></div>
            <div class="hud-corner hud-tr"></div>
            <div class="hud-corner hud-bl"></div>
            <div class="hud-corner hud-br"></div>

            <div class="relative flex-1 p-4 sm:p-5">
              <!-- Header: Icon + Name + Badge -->
              <div class="mb-3 sm:mb-4 flex items-start justify-between">
                <div class="flex min-w-0 flex-1 items-center gap-3">
                  <div class="platform-hex" :style="{ '--p-color': platformColor(group.platform) }">
                    <div class="platform-hex-inner">
                      <div class="scale-75 sm:scale-100">
                        <ModelIcon :model="platformToModel(group.platform)" size="26px" />
                      </div>
                    </div>
                    <div class="platform-hex-ring"></div>
                  </div>
                  <div class="min-w-0 flex-1">
                    <div class="flex items-center justify-between gap-2">
                      <h3 class="flex-1 truncate text-sm font-bold font-mono leading-none tracking-tight text-white sm:text-base">
                        {{ group.group_name }}
                      </h3>
                      <div class="cyber-badge" :class="group.available ? 'cyber-badge-ok' : 'cyber-badge-err'">
                        <span class="cyber-badge-dot" :class="group.available ? 'cbd-ok' : 'cbd-err'"></span>
                        {{ group.available ? t('availability.online') : t('availability.offline') }}
                      </div>
                    </div>
                    <div class="mt-1.5 flex flex-wrap items-center gap-2 text-xs">
                      <span class="platform-tag">
                        {{ platformDisplayName(group.platform) }}
                      </span>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Signal strength bar -->
              <div class="signal-bar" :class="group.available ? 'signal-ok' : 'signal-err'">
                <div class="flex items-center gap-2">
                  <svg class="h-3.5 w-3.5 opacity-70" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M16.247 7.761a6 6 0 0 1 0 8.478"/><path d="M19.075 4.933a10 10 0 0 1 0 14.134"/>
                    <path d="M4.925 19.067a10 10 0 0 1 0-14.134"/><path d="M7.753 16.239a6 6 0 0 1 0-8.478"/>
                    <circle cx="12" cy="12" r="2"/>
                  </svg>
                  <span class="text-[10px] font-mono font-bold uppercase tracking-[0.15em] opacity-70">{{ t('availability.available') }}</span>
                </div>
                <div class="flex items-center gap-1.5">
                  <span class="signal-dot" :class="group.available ? 'sd-ok' : 'sd-err'"></span>
                  <span class="text-xs font-mono font-black uppercase">
                    {{ group.available ? t('availability.online') : t('availability.offline') }}
                  </span>
                </div>
              </div>

              <!-- Waveform timeline -->
              <div class="mt-3 flex gap-[2px] h-4 items-end">
                <div v-for="i in 20" :key="i"
                  class="waveform-bar"
                  :class="group.available ? 'wave-ok' : (Math.random() > 0.3 ? 'wave-err' : 'wave-ok-dim')"
                  :style="{
                    height: group.available ? `${50 + Math.random() * 50}%` : `${20 + Math.random() * 80}%`,
                    animationDelay: `${i * 50}ms`
                  }">
                </div>
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
  anthropic: '#d97706', openai: '#6366f1', gemini: '#4285f4',
  antigravity: '#8b5cf6', sora: '#ec4899', qwen: '#615eff',
  deepseek: '#4d6bfe', glm: '#3859ff', kimi: '#a855f7', iflow: '#10b981'
}
const PLATFORM_MODEL_MAP: Record<string, string> = {
  anthropic: 'claude-3', openai: 'gpt-4', gemini: 'gemini-pro',
  antigravity: 'antigravity', sora: 'sora', qwen: 'qwen-max',
  deepseek: 'deepseek-chat', glm: 'glm-4', kimi: 'moonshot-v1', iflow: 'iflow'
}

function platformDisplayName(p: string): string {
  const m: Record<string, string> = {
    anthropic:'Anthropic', openai:'OpenAI', gemini:'Gemini', antigravity:'Antigravity',
    sora:'Sora', qwen:'Qwen', deepseek:'DeepSeek', glm:'GLM', kimi:'Kimi', iflow:'iFlow'
  }
  return m[p] || p
}
function platformColor(p: string): string { return PLATFORM_COLORS[p] || '#6b7280' }
function platformToModel(p: string): string { return PLATFORM_MODEL_MAP[p] || p }

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

onMounted(() => { refresh() })
</script>

<style scoped>
/* ╔══════════════════════════════════════════╗
   ║  CYBERPUNK COMMAND CENTER — NEON CORE   ║
   ╚══════════════════════════════════════════╝ */

/* ===== NEON COLOR TOKENS ===== */
.avail-page {
  --neon-cyan: #00f0ff;
  --neon-green: #00ff9d;
  --neon-red: #ff3366;
  --neon-purple: #bf5af2;
  --neon-amber: #ffb800;
  --bg-void: #060a14;
  --bg-panel: rgba(6,10,20,0.85);
  --border-dim: rgba(0,240,255,0.08);
  --border-glow: rgba(0,240,255,0.2);
  position: relative;
  min-height: 100vh;
  background: var(--bg-void);
  overflow: hidden;
}
.text-neon-green { color: var(--neon-green); }
.text-neon-red { color: var(--neon-red); }

/* ===== CIRCUIT BOARD BACKGROUND ===== */
.circuit-bg {
  position: fixed;
  inset: 0;
  z-index: 0;
  background:
    radial-gradient(ellipse 80% 60% at 20% 10%, rgba(0,240,255,0.04), transparent),
    radial-gradient(ellipse 60% 50% at 80% 80%, rgba(191,90,242,0.03), transparent),
    radial-gradient(ellipse 40% 40% at 50% 50%, rgba(0,255,157,0.02), transparent);
  pointer-events: none;
}
.circuit-grid {
  position: fixed;
  inset: 0;
  z-index: 0;
  background-image:
    linear-gradient(rgba(0,240,255,0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0,240,255,0.03) 1px, transparent 1px);
  background-size: 60px 60px;
  mask-image: radial-gradient(ellipse 80% 80% at 50% 50%, black 20%, transparent 80%);
  pointer-events: none;
}
.scanline-overlay {
  position: fixed;
  inset: 0;
  z-index: 1;
  background: repeating-linear-gradient(
    0deg,
    transparent,
    transparent 2px,
    rgba(0,240,255,0.008) 2px,
    rgba(0,240,255,0.008) 4px
  );
  pointer-events: none;
}

/* ===== HERO COMMAND PANEL ===== */
.hero-panel {
  position: relative;
  margin-bottom: 1.5rem;
  overflow: hidden;
  border-radius: 4px;
  background: var(--bg-panel);
  border: 1px solid var(--border-dim);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
}
@media (min-width: 640px) { .hero-panel { margin-bottom: 2rem; } }

/* Corner brackets */
.corner-bracket {
  position: absolute;
  width: 20px;
  height: 20px;
  z-index: 20;
}
.corner-bracket::before,
.corner-bracket::after {
  content: '';
  position: absolute;
  background: var(--neon-cyan);
  box-shadow: 0 0 6px var(--neon-cyan);
}
.corner-tl { top: 0; left: 0; }
.corner-tl::before { top: 0; left: 0; width: 20px; height: 1px; }
.corner-tl::after { top: 0; left: 0; width: 1px; height: 20px; }
.corner-tr { top: 0; right: 0; }
.corner-tr::before { top: 0; right: 0; width: 20px; height: 1px; }
.corner-tr::after { top: 0; right: 0; width: 1px; height: 20px; }
.corner-bl { bottom: 0; left: 0; }
.corner-bl::before { bottom: 0; left: 0; width: 20px; height: 1px; }
.corner-bl::after { bottom: 0; left: 0; width: 1px; height: 20px; }
.corner-br { bottom: 0; right: 0; }
.corner-br::before { bottom: 0; right: 0; width: 20px; height: 1px; }
.corner-br::after { bottom: 0; right: 0; width: 1px; height: 20px; }

/* Holographic noise */
.holo-noise {
  position: absolute;
  inset: 0;
  opacity: 0.015;
  background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 256 256' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E");
  background-size: 128px 128px;
  mix-blend-mode: overlay;
  pointer-events: none;
}

/* Hex grid pattern */
.hex-grid {
  position: absolute;
  inset: 0;
  background-image: url("data:image/svg+xml,%3Csvg width='60' height='52' viewBox='0 0 60 52' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M30 0l30 15v22L30 52 0 37V15z' fill='none' stroke='%2300f0ff' stroke-width='0.3' opacity='0.06'/%3E%3C/svg%3E");
  background-size: 60px 52px;
  opacity: 0.5;
  pointer-events: none;
}

/* Sweep beam */
.sweep-beam {
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(0,240,255,0.03), rgba(0,240,255,0.06), rgba(0,240,255,0.03), transparent);
  animation: sweep-lr 6s linear infinite;
  pointer-events: none;
}
@keyframes sweep-lr {
  0% { left: -100%; }
  100% { left: 100%; }
}

/* Floating particles */
.particle {
  position: absolute;
  width: 2px;
  height: 2px;
  border-radius: 50%;
  background: var(--neon-cyan);
  box-shadow: 0 0 4px var(--neon-cyan);
  pointer-events: none;
  animation: float-up 8s linear infinite;
}
.particle-1 { left: 10%; bottom: -5%; animation-delay: 0s; animation-duration: 7s; }
.particle-2 { left: 30%; bottom: -5%; animation-delay: 1.5s; animation-duration: 9s; }
.particle-3 { left: 55%; bottom: -5%; animation-delay: 3s; animation-duration: 6s; background: var(--neon-green); box-shadow: 0 0 4px var(--neon-green); }
.particle-4 { left: 75%; bottom: -5%; animation-delay: 4.5s; animation-duration: 10s; }
.particle-5 { left: 90%; bottom: -5%; animation-delay: 2s; animation-duration: 8s; background: var(--neon-purple); box-shadow: 0 0 4px var(--neon-purple); }
@keyframes float-up {
  0% { transform: translateY(0) scale(1); opacity: 0; }
  10% { opacity: 1; }
  90% { opacity: 0.6; }
  100% { transform: translateY(-400px) scale(0.5); opacity: 0; }
}
@media (max-width: 639px) {
  .particle { display: none; }
  .sweep-beam { display: none; }
  .hex-grid { display: none; }
}

/* ===== RADAR ICON ===== */
.radar-wrap {
  position: relative;
  width: 44px;
  height: 44px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}
@media (min-width: 640px) { .radar-wrap { width: 52px; height: 52px; } }
.radar-core {
  width: 100%;
  height: 100%;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0,240,255,0.08);
  color: var(--neon-cyan);
  border: 1px solid rgba(0,240,255,0.15);
  position: relative;
  z-index: 2;
}
.radar-ring {
  position: absolute;
  border-radius: 4px;
  border: 1px solid var(--neon-cyan);
  opacity: 0;
  z-index: 1;
}
.radar-ring-1 {
  inset: -4px;
  animation: radar-ping 3s ease-out infinite;
}
.radar-ring-2 {
  inset: -8px;
  animation: radar-ping 3s ease-out infinite 1s;
}
.radar-ring-3 {
  inset: -12px;
  animation: radar-ping 3s ease-out infinite 2s;
}
@keyframes radar-ping {
  0% { opacity: 0.5; transform: scale(1); }
  100% { opacity: 0; transform: scale(1.3); }
}

/* Title area */
.hero-title {
  font-family: 'JetBrains Mono', 'Fira Code', 'SF Mono', 'Cascadia Code', monospace;
  font-size: 1.125rem;
  font-weight: 800;
  color: white;
  letter-spacing: -0.02em;
}
@media (min-width: 640px) { .hero-title { font-size: 1.5rem; } }
.sys-tag {
  display: inline-block;
  padding: 1px 6px;
  font-family: monospace;
  font-size: 9px;
  font-weight: 700;
  letter-spacing: 0.1em;
  color: var(--neon-cyan);
  background: rgba(0,240,255,0.08);
  border: 1px solid rgba(0,240,255,0.2);
  border-radius: 2px;
  text-transform: uppercase;
  animation: tag-blink 4s ease-in-out infinite;
}
@keyframes tag-blink {
  0%,90%,100% { opacity: 1; }
  95% { opacity: 0.3; }
}
.hero-sub {
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 0.7rem;
  color: rgba(0,240,255,0.4);
  margin-top: 4px;
  display: flex;
  align-items: center;
  gap: 4px;
}
@media (min-width: 640px) { .hero-sub { font-size: 0.8rem; } }
.terminal-cursor {
  display: inline-block;
  width: 6px;
  height: 12px;
  background: var(--neon-cyan);
  animation: cursor-blink 1s step-end infinite;
  flex-shrink: 0;
}
@keyframes cursor-blink {
  0%,50% { opacity: 1; }
  51%,100% { opacity: 0; }
}

/* Live badge */
.live-badge {
  padding: 4px 10px;
  border-radius: 2px;
  background: rgba(0,255,157,0.06);
  border: 1px solid rgba(0,255,157,0.15);
  backdrop-filter: blur(8px);
}
.live-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--neon-green);
  flex-shrink: 0;
  position: relative;
  box-shadow: 0 0 8px var(--neon-green);
}
.live-dot::after {
  content: '';
  position: absolute;
  inset: -4px;
  border-radius: 50%;
  background: rgba(0,255,157,0.25);
  animation: live-ping 2s ease-out infinite;
}
@keyframes live-ping {
  0% { transform: scale(1); opacity: 1; }
  100% { transform: scale(3); opacity: 0; }
}
.live-text {
  font-family: monospace;
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.15em;
  color: var(--neon-green);
  text-shadow: 0 0 8px rgba(0,255,157,0.5);
}
.live-time {
  font-family: monospace;
  font-size: 10px;
  font-weight: 500;
  color: rgba(0,255,157,0.6);
}

/* Refresh button */
.refresh-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 2px;
  background: rgba(0,240,255,0.06);
  border: 1px solid rgba(0,240,255,0.15);
  color: var(--neon-cyan);
  font-family: monospace;
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.05em;
  cursor: pointer;
  transition: all 0.2s;
  text-transform: uppercase;
}
.refresh-btn:hover {
  background: rgba(0,240,255,0.12);
  border-color: rgba(0,240,255,0.3);
  box-shadow: 0 0 15px rgba(0,240,255,0.15);
}
.refresh-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

/* ===== STAT CELLS ===== */
.stat-cell {
  padding: 12px 14px;
  border-radius: 2px;
  background: rgba(0,240,255,0.03);
  border: 1px solid rgba(0,240,255,0.08);
  text-align: center;
  position: relative;
  overflow: hidden;
  transition: all 0.3s;
}
.stat-cell:hover {
  background: rgba(0,240,255,0.06);
}
@media (min-width: 640px) { .stat-cell { padding: 16px 20px; } }

.stat-cyan { border-color: rgba(0,240,255,0.12); }
.stat-green { border-color: rgba(0,255,157,0.12); }
.stat-red { border-color: rgba(255,51,102,0.12); }
.stat-cyan:hover { border-color: rgba(0,240,255,0.25); box-shadow: 0 0 20px rgba(0,240,255,0.06); }
.stat-green:hover { border-color: rgba(0,255,157,0.25); box-shadow: 0 0 20px rgba(0,255,157,0.06); }
.stat-red:hover { border-color: rgba(255,51,102,0.25); box-shadow: 0 0 20px rgba(255,51,102,0.06); }

.stat-value {
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 1.5rem;
  font-weight: 900;
  color: var(--neon-cyan);
  text-shadow: 0 0 20px rgba(0,240,255,0.3);
  line-height: 1;
}
@media (min-width: 640px) { .stat-value { font-size: 2.25rem; } }
.stat-label {
  font-family: monospace;
  font-size: 9px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.2em;
  color: rgba(0,240,255,0.4);
  margin-top: 6px;
}
@media (min-width: 640px) { .stat-label { font-size: 10px; } }

.stat-bar-bg {
  height: 2px;
  background: rgba(255,255,255,0.04);
  margin-top: 10px;
  border-radius: 1px;
  overflow: hidden;
}
.stat-bar-fill {
  height: 100%;
  border-radius: 1px;
  transition: width 1.2s cubic-bezier(0.4,0,0.2,1);
}
.stat-bar-cyan { background: var(--neon-cyan); box-shadow: 0 0 8px rgba(0,240,255,0.4); }
.stat-bar-green { background: var(--neon-green); box-shadow: 0 0 8px rgba(0,255,157,0.4); }
.stat-bar-red { background: var(--neon-red); box-shadow: 0 0 8px rgba(255,51,102,0.4); }

/* ===== UPTIME BAR ===== */
.uptime-track {
  height: 3px;
  border-radius: 0;
  background: rgba(255,255,255,0.04);
  position: relative;
  overflow: visible;
}
.uptime-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--neon-cyan), var(--neon-green));
  box-shadow: 0 0 12px rgba(0,255,157,0.3);
  transition: width 1.2s cubic-bezier(0.4,0,0.2,1);
  position: relative;
}
.uptime-fill::after {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, transparent 30%, rgba(255,255,255,0.3) 50%, transparent 70%);
  animation: uptime-shimmer 2.5s ease-in-out infinite;
}
@keyframes uptime-shimmer {
  0% { transform: translateX(-150%); }
  100% { transform: translateX(250%); }
}
.uptime-glow {
  position: absolute;
  top: 50%;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--neon-green);
  box-shadow: 0 0 12px var(--neon-green), 0 0 24px rgba(0,255,157,0.3);
  transform: translate(-50%, -50%);
  transition: left 1.2s cubic-bezier(0.4,0,0.2,1);
}

/* ===== FILTER CHIPS ===== */
.filter-scroll {
  display: flex;
  gap: 6px;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
  padding-bottom: 2px;
}
.filter-scroll::-webkit-scrollbar { display: none; }
@media (min-width: 640px) { .filter-scroll { flex-wrap: wrap; overflow-x: visible; } }

.neon-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  border-radius: 2px;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 0.7rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid transparent;
  white-space: nowrap;
  flex-shrink: 0;
  position: relative;
  overflow: hidden;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
@media (min-width: 640px) { .neon-chip { padding: 6px 14px; font-size: 0.75rem; } }

.chip-glow {
  position: absolute;
  inset: 0;
  opacity: 0;
  transition: opacity 0.2s;
}
.neon-chip-active .chip-glow {
  opacity: 1;
  background: radial-gradient(ellipse at center, rgba(0,240,255,0.1), transparent 70%);
}

.neon-chip-active {
  background: rgba(0,240,255,0.1);
  color: var(--neon-cyan);
  border-color: rgba(0,240,255,0.3);
  box-shadow: 0 0 15px rgba(0,240,255,0.1), inset 0 0 15px rgba(0,240,255,0.05);
  text-shadow: 0 0 8px rgba(0,240,255,0.4);
}
.neon-chip-idle {
  background: rgba(255,255,255,0.02);
  color: rgba(255,255,255,0.35);
  border-color: rgba(255,255,255,0.06);
}
.neon-chip-idle:hover {
  background: rgba(0,240,255,0.04);
  color: rgba(0,240,255,0.7);
  border-color: rgba(0,240,255,0.15);
}

.chip-count {
  font-size: 0.6rem;
  padding: 0 5px;
  border-radius: 2px;
  line-height: 1.6;
  background: rgba(255,255,255,0.04);
  font-weight: 700;
}
.neon-chip-active .chip-count {
  background: rgba(0,240,255,0.15);
  color: var(--neon-cyan);
}

/* ===== SKELETON LOADING ===== */
.skel-card {
  border-radius: 2px;
  border: 1px solid rgba(0,240,255,0.06);
  background: rgba(6,10,20,0.6);
  position: relative;
  overflow: hidden;
  animation: card-materialize 0.4s ease-out both;
}
.skel-scan {
  position: absolute;
  top: 0;
  left: -100%;
  width: 50%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(0,240,255,0.04), transparent);
  animation: skel-sweep 2s ease-in-out infinite;
}
@keyframes skel-sweep {
  0% { left: -50%; }
  100% { left: 150%; }
}
.skel-block {
  background: rgba(0,240,255,0.04);
  border: 1px solid rgba(0,240,255,0.04);
  position: relative;
  overflow: hidden;
}
.skel-block::after {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, transparent, rgba(0,240,255,0.06), transparent);
  animation: skel-glint 1.8s ease-in-out infinite;
}
@keyframes skel-glint {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(200%); }
}

/* Error / Empty states */
.err-hex, .empty-hex {
  width: 56px;
  height: 56px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}
.err-hex {
  background: rgba(255,51,102,0.06);
  border: 1px solid rgba(255,51,102,0.15);
  box-shadow: 0 0 20px rgba(255,51,102,0.08);
}
.empty-hex {
  background: rgba(0,240,255,0.04);
  border: 1px solid rgba(0,240,255,0.1);
}
.retry-btn {
  padding: 8px 20px;
  border-radius: 2px;
  background: rgba(0,240,255,0.08);
  border: 1px solid rgba(0,240,255,0.2);
  color: var(--neon-cyan);
  font-family: monospace;
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  text-transform: uppercase;
  letter-spacing: 0.1em;
}
.retry-btn:hover {
  background: rgba(0,240,255,0.15);
  box-shadow: 0 0 20px rgba(0,240,255,0.15);
}

/* ===== CYBER CARDS ===== */
.cyber-card {
  position: relative;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-radius: 2px;
  background: rgba(6,10,20,0.7);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  transition: all 0.35s cubic-bezier(0.4,0,0.2,1);
  animation: card-materialize 0.45s ease-out both;
  border: 1px solid rgba(0,240,255,0.06);
}
@keyframes card-materialize {
  from { opacity: 0; transform: translateY(15px) scale(0.98); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}
.cyber-card:hover {
  transform: translateY(-4px);
  background: rgba(6,10,20,0.85);
}
.cyber-card-ok:hover {
  border-color: rgba(0,255,157,0.25);
  box-shadow:
    0 0 30px rgba(0,255,157,0.06),
    0 20px 50px -15px rgba(0,0,0,0.5),
    inset 0 1px 0 rgba(0,255,157,0.08);
}
.cyber-card-err:hover {
  border-color: rgba(255,51,102,0.25);
  box-shadow:
    0 0 30px rgba(255,51,102,0.06),
    0 20px 50px -15px rgba(0,0,0,0.5),
    inset 0 1px 0 rgba(255,51,102,0.08);
}

/* Animated border gradient */
.card-border-glow {
  position: absolute;
  inset: 0;
  border-radius: 2px;
  opacity: 0;
  transition: opacity 0.35s;
  pointer-events: none;
  z-index: 0;
}
.cyber-card:hover .card-border-glow { opacity: 1; }
.border-glow-ok {
  background: linear-gradient(135deg, rgba(0,255,157,0.04), transparent 50%, rgba(0,240,255,0.03));
}
.border-glow-err {
  background: linear-gradient(135deg, rgba(255,51,102,0.04), transparent 50%, rgba(255,51,102,0.02));
}

/* HUD corner marks */
.hud-corner {
  position: absolute;
  width: 8px;
  height: 8px;
  z-index: 10;
  opacity: 0;
  transition: opacity 0.3s;
}
.cyber-card:hover .hud-corner { opacity: 1; }
.hud-corner::before,
.hud-corner::after {
  content: '';
  position: absolute;
  background: var(--neon-cyan);
}
.hud-tl { top: 3px; left: 3px; }
.hud-tl::before { top: 0; left: 0; width: 8px; height: 1px; }
.hud-tl::after { top: 0; left: 0; width: 1px; height: 8px; }
.hud-tr { top: 3px; right: 3px; }
.hud-tr::before { top: 0; right: 0; width: 8px; height: 1px; }
.hud-tr::after { top: 0; right: 0; width: 1px; height: 8px; }
.hud-bl { bottom: 3px; left: 3px; }
.hud-bl::before { bottom: 0; left: 0; width: 8px; height: 1px; }
.hud-bl::after { bottom: 0; left: 0; width: 1px; height: 8px; }
.hud-br { bottom: 3px; right: 3px; }
.hud-br::before { bottom: 0; right: 0; width: 8px; height: 1px; }
.hud-br::after { bottom: 0; right: 0; width: 1px; height: 8px; }
.cyber-card-ok .hud-corner::before,
.cyber-card-ok .hud-corner::after { background: var(--neon-green); }
.cyber-card-err .hud-corner::before,
.cyber-card-err .hud-corner::after { background: var(--neon-red); }

/* Platform hex icon */
.platform-hex {
  position: relative;
  flex-shrink: 0;
}
.platform-hex-inner {
  width: 40px;
  height: 40px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.06);
  transition: all 0.3s;
  position: relative;
  z-index: 1;
}
@media (min-width: 640px) { .platform-hex-inner { width: 48px; height: 48px; } }
.group:hover .platform-hex-inner {
  background: rgba(255,255,255,0.06);
  border-color: var(--p-color, rgba(255,255,255,0.12));
  box-shadow: 0 0 15px rgba(0,240,255,0.08);
  transform: scale(1.05);
}
.platform-hex-ring {
  position: absolute;
  inset: -3px;
  border-radius: 6px;
  border: 1px solid var(--p-color, rgba(0,240,255,0.15));
  opacity: 0;
  transition: opacity 0.3s;
}
.group:hover .platform-hex-ring {
  opacity: 0.4;
  animation: hex-ring-pulse 2s ease-in-out infinite;
}
@keyframes hex-ring-pulse {
  0%,100% { opacity: 0.2; transform: scale(1); }
  50% { opacity: 0.5; transform: scale(1.06); }
}

/* Platform tag */
.platform-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 1px 8px;
  border-radius: 2px;
  background: rgba(0,240,255,0.04);
  border: 1px solid rgba(0,240,255,0.08);
  font-family: monospace;
  font-size: 10px;
  font-weight: 600;
  color: rgba(0,240,255,0.5);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

/* Cyber badge */
.cyber-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 2px 8px;
  border-radius: 2px;
  font-family: monospace;
  font-size: 9px;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  flex-shrink: 0;
  white-space: nowrap;
}
@media (min-width: 640px) { .cyber-badge { padding: 3px 10px; font-size: 10px; } }
.cyber-badge-ok {
  background: rgba(0,255,157,0.08);
  color: var(--neon-green);
  border: 1px solid rgba(0,255,157,0.15);
  text-shadow: 0 0 8px rgba(0,255,157,0.3);
}
.cyber-badge-err {
  background: rgba(255,51,102,0.08);
  color: var(--neon-red);
  border: 1px solid rgba(255,51,102,0.15);
  text-shadow: 0 0 8px rgba(255,51,102,0.3);
}
.cyber-badge-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
}
.cbd-ok {
  background: var(--neon-green);
  box-shadow: 0 0 6px var(--neon-green), 0 0 12px rgba(0,255,157,0.3);
  animation: cbd-pulse 2s ease-in-out infinite;
}
.cbd-err {
  background: var(--neon-red);
  box-shadow: 0 0 6px var(--neon-red);
  animation: cbd-flash 1.5s ease-in-out infinite;
}
@keyframes cbd-pulse {
  0%,100% { box-shadow: 0 0 4px var(--neon-green); }
  50% { box-shadow: 0 0 10px var(--neon-green), 0 0 20px rgba(0,255,157,0.3); }
}
@keyframes cbd-flash {
  0%,100% { opacity: 1; }
  50% { opacity: 0.5; }
}

/* Signal bar */
.signal-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 2px;
  transition: all 0.2s;
}
.signal-ok {
  background: rgba(0,255,157,0.04);
  color: var(--neon-green);
  border: 1px solid rgba(0,255,157,0.06);
}
.signal-err {
  background: rgba(255,51,102,0.04);
  color: var(--neon-red);
  border: 1px solid rgba(255,51,102,0.06);
}
.group:hover .signal-ok {
  background: rgba(0,255,157,0.08);
  border-color: rgba(0,255,157,0.12);
}
.group:hover .signal-err {
  background: rgba(255,51,102,0.08);
  border-color: rgba(255,51,102,0.12);
}
.signal-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}
.sd-ok {
  background: var(--neon-green);
  box-shadow: 0 0 8px var(--neon-green);
  animation: cbd-pulse 2s ease-in-out infinite;
}
.sd-err {
  background: var(--neon-red);
  box-shadow: 0 0 8px var(--neon-red);
}

/* Waveform bars */
.waveform-bar {
  flex: 1;
  border-radius: 1px;
  transition: all 0.2s;
  animation: wave-grow 0.6s ease-out both;
}
@keyframes wave-grow {
  from { transform: scaleY(0); }
  to { transform: scaleY(1); }
}
.wave-ok {
  background: var(--neon-green);
  opacity: 0.5;
  box-shadow: 0 0 4px rgba(0,255,157,0.2);
}
.wave-ok:hover {
  opacity: 0.9;
  box-shadow: 0 0 8px rgba(0,255,157,0.4);
}
.wave-ok-dim {
  background: rgba(0,255,157,0.15);
}
.wave-err {
  background: var(--neon-red);
  opacity: 0.4;
  box-shadow: 0 0 4px rgba(255,51,102,0.2);
}
.wave-err:hover {
  opacity: 0.8;
  box-shadow: 0 0 8px rgba(255,51,102,0.4);
}

/* ===== LIGHT MODE ADAPTATIONS ===== */
/* In light mode, maintain the dark cyberpunk feel but lighten slightly */
:root:not(.dark) .avail-page {
  --bg-void: #0c1222;
  --bg-panel: rgba(12,18,34,0.9);
}
</style>
