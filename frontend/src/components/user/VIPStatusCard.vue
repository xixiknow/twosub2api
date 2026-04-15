<template>
  <div v-if="user?.current_vip" class="card overflow-hidden">
    <div
      class="border-b border-amber-100 bg-gradient-to-r from-amber-50 via-orange-50 to-amber-100 px-6 py-4 dark:border-amber-900/40 dark:from-amber-950/40 dark:via-orange-950/30 dark:to-amber-900/30"
    >
      <div class="flex items-center justify-between gap-4">
        <div class="flex items-center gap-3">
          <div
            class="flex h-11 w-11 items-center justify-center rounded-2xl bg-amber-500/15 text-amber-600 dark:bg-amber-400/10 dark:text-amber-300"
          >
            <Icon name="sparkles" size="lg" />
          </div>
          <div>
            <div class="text-sm font-semibold text-amber-900 dark:text-amber-100">
              {{ t('profile.vip.title') }}
            </div>
            <div class="text-xs text-amber-700 dark:text-amber-300">
              {{ user.current_vip.enabled ? t('profile.vip.enabledHint') : t('profile.vip.disabledHint') }}
            </div>
          </div>
        </div>
        <div
          class="rounded-full border border-amber-200 bg-white/80 px-3 py-1 text-sm font-semibold text-amber-700 dark:border-amber-800 dark:bg-amber-950/40 dark:text-amber-200"
        >
          {{ user.current_vip.level_name }}
        </div>
      </div>
    </div>

    <div class="space-y-5 px-6 py-5">
      <div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
        <div class="rounded-2xl bg-gray-50 px-4 py-3 dark:bg-dark-800/70">
          <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('profile.vip.currentLevel') }}</div>
          <div class="mt-1 text-base font-semibold text-gray-900 dark:text-white">
            {{ user.current_vip.level_name }}
          </div>
        </div>
        <div class="rounded-2xl bg-gray-50 px-4 py-3 dark:bg-dark-800/70">
          <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('profile.vip.multiplier') }}</div>
          <div class="mt-1 text-base font-semibold text-gray-900 dark:text-white">
            {{ user.current_vip.base_multiplier.toFixed(2) }}x
          </div>
        </div>
        <div class="rounded-2xl bg-gray-50 px-4 py-3 dark:bg-dark-800/70">
          <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('profile.vip.progress') }}</div>
          <div class="mt-1 text-base font-semibold text-gray-900 dark:text-white">
            {{ Math.round(user.current_vip.progress_percent || 0) }}%
          </div>
        </div>
      </div>

      <div class="space-y-2">
        <div class="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
          <span>{{ t('profile.vip.upgradeProgress') }}</span>
          <span>{{ Math.round(user.current_vip.progress_percent || 0) }}%</span>
        </div>
        <div class="h-2 overflow-hidden rounded-full bg-amber-100 dark:bg-amber-950/50">
          <div
            class="h-full rounded-full bg-gradient-to-r from-amber-400 to-orange-500 transition-all"
            :style="{ width: `${Math.max(0, Math.min(100, user.current_vip.progress_percent || 0))}%` }"
          />
        </div>
      </div>

      <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <div class="rounded-2xl border border-gray-100 px-4 py-3 dark:border-dark-700">
          <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('profile.vip.rechargeTotal') }}</div>
          <div class="mt-1 text-base font-semibold text-gray-900 dark:text-white">
            ¥{{ (user.current_vip.recharge_total || 0).toFixed(2) }}
          </div>
        </div>
        <div class="rounded-2xl border border-gray-100 px-4 py-3 dark:border-dark-700">
          <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('profile.vip.spendTotal') }}</div>
          <div class="mt-1 text-base font-semibold text-gray-900 dark:text-white">
            ${{ (user.current_vip.spend_total || 0).toFixed(2) }}
          </div>
        </div>
      </div>

      <div
        v-if="user.next_vip"
        class="rounded-2xl border border-dashed border-amber-300 bg-amber-50/60 px-4 py-4 dark:border-amber-800 dark:bg-amber-950/20"
      >
        <div class="text-sm font-semibold text-amber-900 dark:text-amber-100">
          {{ t('profile.vip.nextLevel') }}: {{ user.next_vip.level_name }}
        </div>
        <div class="mt-1 text-sm text-amber-800 dark:text-amber-200">
          {{ user.next_vip.unlock_condition_label }}
        </div>
        <div class="mt-3 grid grid-cols-1 gap-3 sm:grid-cols-2">
          <div class="rounded-xl bg-white/80 px-3 py-2 dark:bg-dark-900/40">
            <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('profile.vip.remainingRecharge') }}</div>
            <div class="mt-1 text-sm font-medium text-gray-900 dark:text-white">
              ¥{{ user.next_vip.remaining_recharge.toFixed(2) }}
            </div>
          </div>
          <div class="rounded-xl bg-white/80 px-3 py-2 dark:bg-dark-900/40">
            <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('profile.vip.remainingSpend') }}</div>
            <div class="mt-1 text-sm font-medium text-gray-900 dark:text-white">
              ${{ user.next_vip.remaining_spend.toFixed(2) }}
            </div>
          </div>
        </div>
      </div>

      <div
        v-else
        class="rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700 dark:border-emerald-900/40 dark:bg-emerald-950/20 dark:text-emerald-300"
      >
        {{ t('profile.vip.maxLevelReached') }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { User } from '@/types'

defineProps<{
  user: User | null
}>()

const { t } = useI18n()
</script>
