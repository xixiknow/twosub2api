<template>
  <div v-if="visible" class="relative" ref="containerRef">
    <button
      class="flex items-center gap-2 rounded-xl bg-amber-50 px-3 py-1.5 transition-colors hover:bg-amber-100 dark:bg-amber-900/20 dark:hover:bg-amber-900/30"
      :title="user?.current_vip?.level_name || 'VIP0'"
      @click="toggleTooltip"
    >
      <Icon name="sparkles" size="sm" class="text-amber-600 dark:text-amber-300" />
      <span class="text-xs font-semibold text-amber-700 dark:text-amber-200">
        {{ user?.current_vip?.level_name || 'VIP0' }}
      </span>
    </button>

    <transition name="dropdown">
      <div
        v-if="tooltipOpen"
        class="absolute right-0 z-50 mt-2 w-[320px] overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xl dark:border-dark-700 dark:bg-dark-800"
      >
        <div class="border-b border-gray-100 p-4 dark:border-dark-700">
          <div class="flex items-center justify-between gap-3">
            <div>
              <div class="text-sm font-semibold text-gray-900 dark:text-white">
                {{ user?.current_vip?.level_name || 'VIP0' }}
              </div>
              <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                {{ user?.current_vip?.enabled ? 'VIP 计费已启用' : '仅展示等级进度，未启用折扣' }}
              </div>
            </div>
            <div class="rounded-full bg-amber-50 px-2.5 py-1 text-xs font-semibold text-amber-700 dark:bg-amber-900/20 dark:text-amber-200">
              {{ formatMultiplier(user?.current_vip?.base_multiplier) }}
            </div>
          </div>
        </div>

        <div class="space-y-3 p-4">
          <div class="grid grid-cols-2 gap-3">
            <div class="rounded-xl bg-gray-50 px-3 py-2 dark:bg-dark-900/40">
              <div class="text-[11px] text-gray-500 dark:text-gray-400">累计实付</div>
              <div class="mt-1 text-sm font-medium text-gray-900 dark:text-white">
                ¥{{ (user?.current_vip?.recharge_total || 0).toFixed(2) }}
              </div>
            </div>
            <div class="rounded-xl bg-gray-50 px-3 py-2 dark:bg-dark-900/40">
              <div class="text-[11px] text-gray-500 dark:text-gray-400">累计消费</div>
              <div class="mt-1 text-sm font-medium text-gray-900 dark:text-white">
                ${{ (user?.current_vip?.spend_total || 0).toFixed(2) }}
              </div>
            </div>
          </div>

          <div class="space-y-1.5">
            <div class="flex items-center justify-between text-[11px] text-gray-500 dark:text-gray-400">
              <span>升级进度</span>
              <span>{{ Math.round(user?.current_vip?.progress_percent || 0) }}%</span>
            </div>
            <div class="h-2 overflow-hidden rounded-full bg-amber-100 dark:bg-amber-950/40">
              <div
                class="h-full rounded-full bg-gradient-to-r from-amber-400 to-orange-500 transition-all"
                :style="{ width: `${progressWidth}%` }"
              />
            </div>
          </div>

          <div v-if="user?.next_vip" class="rounded-xl border border-dashed border-amber-300 bg-amber-50/70 p-3 dark:border-amber-800 dark:bg-amber-950/20">
            <div class="text-sm font-medium text-amber-900 dark:text-amber-100">
              下一等级 {{ user.next_vip.level_name }}
            </div>
            <div class="mt-1 text-xs text-amber-800 dark:text-amber-200">
              {{ user.next_vip.unlock_condition_label }}
            </div>
            <div class="mt-3 grid grid-cols-2 gap-3 text-xs text-gray-600 dark:text-gray-300">
              <div>
                <div class="text-gray-500 dark:text-gray-400">还差实付</div>
                <div class="mt-1 font-medium">¥{{ user.next_vip.remaining_recharge.toFixed(2) }}</div>
              </div>
              <div>
                <div class="text-gray-500 dark:text-gray-400">还差消费</div>
                <div class="mt-1 font-medium">${{ user.next_vip.remaining_spend.toFixed(2) }}</div>
              </div>
            </div>
          </div>

          <div v-else class="rounded-xl bg-emerald-50 px-3 py-2 text-sm text-emerald-700 dark:bg-emerald-950/20 dark:text-emerald-300">
            已达到最高 VIP 等级
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useAuthStore } from '@/stores'
import Icon from '@/components/icons/Icon.vue'

const authStore = useAuthStore()

const user = computed(() => authStore.user)
const visible = computed(() => !!user.value?.current_vip?.enabled)
const tooltipOpen = ref(false)
const containerRef = ref<HTMLElement | null>(null)

const progressWidth = computed(() => {
  const progress = Number(user.value?.current_vip?.progress_percent || 0)
  return Math.max(0, Math.min(100, progress))
})

function formatMultiplier(value?: number) {
  return `${Number(value || 1).toFixed(2)}x`
}

function toggleTooltip() {
  tooltipOpen.value = !tooltipOpen.value
}

function closeTooltip() {
  tooltipOpen.value = false
}

function handleClickOutside(event: MouseEvent) {
  if (containerRef.value && !containerRef.value.contains(event.target as Node)) {
    closeTooltip()
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.2s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: scale(0.95) translateY(-4px);
}
</style>
