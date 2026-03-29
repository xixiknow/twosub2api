<template>
  <AppLayout>
    <div class="mx-auto max-w-3xl space-y-6">
      <!-- Stats Cards -->
      <div class="grid grid-cols-3 gap-4">
        <div class="card rounded-xl bg-emerald-50 p-5 text-center dark:bg-emerald-900/20">
          <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('wallet.referral.totalEarnings') }}</p>
          <p class="mt-1 text-2xl font-bold text-emerald-600 dark:text-emerald-400">
            ${{ referralInfo?.total_earnings?.toFixed(2) || '0.00' }}
          </p>
        </div>
        <div class="card rounded-xl bg-blue-50 p-5 text-center dark:bg-blue-900/20">
          <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('wallet.referral.totalReferred') }}</p>
          <p class="mt-1 text-2xl font-bold text-blue-600 dark:text-blue-400">
            {{ referralInfo?.total_referred || 0 }} {{ t('wallet.referral.people') }}
          </p>
        </div>
        <div class="card rounded-xl bg-purple-50 p-5 text-center dark:bg-purple-900/20">
          <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('wallet.referral.commissionRate') }}</p>
          <p class="mt-1 text-2xl font-bold text-purple-600 dark:text-purple-400">
            {{ referralInfo ? (referralInfo.commission_rate * 100).toFixed(0) : '0' }}%
          </p>
        </div>
      </div>

      <!-- Referral Code & Link -->
      <div class="card">
        <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('wallet.referral.title') }}
          </h2>
        </div>
        <div class="space-y-4 p-6">
          <div>
            <label class="input-label mb-1">{{ t('wallet.referral.yourCode') }}</label>
            <div class="flex items-center gap-2">
              <div class="flex-1 rounded-lg border border-gray-200 bg-gray-50 px-4 py-2.5 font-mono text-sm text-gray-900 dark:border-dark-600 dark:bg-dark-800 dark:text-white">
                {{ referralInfo?.referral_code || '-' }}
              </div>
              <button type="button" class="btn btn-primary flex-shrink-0 px-4 py-2.5" @click="copyToClipboard(referralInfo?.referral_code || '', 'code')">
                {{ copied === 'code' ? t('wallet.referral.copied') : t('wallet.referral.copyCode') }}
              </button>
            </div>
          </div>
          <div>
            <label class="input-label mb-1">{{ t('wallet.referral.referralLink') }}</label>
            <div class="flex items-center gap-2">
              <div class="flex-1 truncate rounded-lg border border-gray-200 bg-gray-50 px-4 py-2.5 font-mono text-sm text-gray-900 dark:border-dark-600 dark:bg-dark-800 dark:text-white">
                {{ referralLink }}
              </div>
              <button type="button" class="btn btn-primary flex-shrink-0 px-4 py-2.5" @click="copyToClipboard(referralLink, 'link')">
                {{ copied === 'link' ? t('wallet.referral.copied') : t('wallet.referral.copyLink') }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Referred Users -->
      <div class="card">
        <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('referralPage.referredUsers') }}
          </h2>
        </div>
        <div class="p-6">
          <div v-if="usersLoading" class="flex items-center justify-center py-8">
            <svg class="h-6 w-6 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
          </div>
          <div v-else-if="referredUsers.length === 0" class="py-8 text-center text-sm text-gray-500 dark:text-dark-400">
            {{ t('referralPage.noUsers') }}
          </div>
          <table v-else class="w-full text-sm">
            <thead>
              <tr class="border-b border-gray-100 text-left text-xs font-medium uppercase text-gray-500 dark:border-dark-700 dark:text-dark-400">
                <th class="pb-3 pr-4">{{ t('referralPage.email') }}</th>
                <th class="pb-3 pr-4">{{ t('referralPage.registerTime') }}</th>
                <th class="pb-3 text-right">{{ t('referralPage.generatedCommission') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="u in referredUsers" :key="u.email" class="border-b border-gray-50 dark:border-dark-800">
                <td class="py-3 pr-4 font-mono text-gray-900 dark:text-white">{{ u.email }}</td>
                <td class="py-3 pr-4 text-gray-500 dark:text-dark-400">{{ formatDateTime(u.created_at) }}</td>
                <td class="py-3 text-right font-semibold text-emerald-600 dark:text-emerald-400">${{ u.total_commission.toFixed(2) }}</td>
              </tr>
            </tbody>
          </table>
          <!-- Users Pagination -->
          <div v-if="usersPages > 1" class="mt-4 flex items-center justify-center gap-2">
            <button class="btn btn-secondary px-3 py-1 text-xs" :disabled="usersPage <= 1" @click="loadUsers(usersPage - 1)">{{ t('referralPage.prev') }}</button>
            <span class="text-sm text-gray-500 dark:text-dark-400">{{ usersPage }} / {{ usersPages }}</span>
            <button class="btn btn-secondary px-3 py-1 text-xs" :disabled="usersPage >= usersPages" @click="loadUsers(usersPage + 1)">{{ t('referralPage.next') }}</button>
          </div>
        </div>
      </div>

      <!-- Commission Records -->
      <div class="card">
        <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('wallet.referral.commissionRecords') }}
          </h2>
        </div>
        <div class="p-6">
          <div v-if="commissionsLoading" class="flex items-center justify-center py-8">
            <svg class="h-6 w-6 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
          </div>
          <div v-else-if="commissions.length === 0" class="py-8 text-center text-sm text-gray-500 dark:text-dark-400">
            {{ t('wallet.referral.noRecords') }}
          </div>
          <table v-else class="w-full text-sm">
            <thead>
              <tr class="border-b border-gray-100 text-left text-xs font-medium uppercase text-gray-500 dark:border-dark-700 dark:text-dark-400">
                <th class="pb-3 pr-4">{{ t('wallet.referral.orderAmount') }}</th>
                <th class="pb-3 pr-4">{{ t('wallet.referral.commissionRate') }}</th>
                <th class="pb-3 pr-4">{{ t('wallet.referral.commissionAmount') }}</th>
                <th class="pb-3 text-right">{{ t('wallet.referral.time') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="c in commissions" :key="c.id" class="border-b border-gray-50 dark:border-dark-800">
                <td class="py-3 pr-4 text-gray-900 dark:text-white">${{ c.order_amount.toFixed(2) }}</td>
                <td class="py-3 pr-4 text-gray-500 dark:text-dark-400">{{ (c.commission_rate * 100).toFixed(0) }}%</td>
                <td class="py-3 pr-4 font-semibold text-emerald-600 dark:text-emerald-400">${{ c.commission_amount.toFixed(2) }}</td>
                <td class="py-3 text-right text-gray-500 dark:text-dark-400">{{ formatDateTime(c.created_at) }}</td>
              </tr>
            </tbody>
          </table>
          <!-- Commissions Pagination -->
          <div v-if="commissionsPages > 1" class="mt-4 flex items-center justify-center gap-2">
            <button class="btn btn-secondary px-3 py-1 text-xs" :disabled="commissionsPage <= 1" @click="loadCommissions(commissionsPage - 1)">{{ t('referralPage.prev') }}</button>
            <span class="text-sm text-gray-500 dark:text-dark-400">{{ commissionsPage }} / {{ commissionsPages }}</span>
            <button class="btn btn-secondary px-3 py-1 text-xs" :disabled="commissionsPage >= commissionsPages" @click="loadCommissions(commissionsPage + 1)">{{ t('referralPage.next') }}</button>
          </div>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="infoLoading && !referralInfo" class="card">
        <div class="flex items-center justify-center p-8">
          <svg class="h-6 w-6 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { referralAPI, type ReferralInfo, type CommissionRecord } from '@/api'
import type { ReferredUser } from '@/api/referral'
import AppLayout from '@/components/layout/AppLayout.vue'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const appStore = useAppStore()

const referralInfo = ref<ReferralInfo | null>(null)
const infoLoading = ref(false)
const copied = ref<'code' | 'link' | null>(null)

// Referred users
const referredUsers = ref<ReferredUser[]>([])
const usersLoading = ref(false)
const usersPage = ref(1)
const usersPages = ref(0)

// Commissions
const commissions = ref<CommissionRecord[]>([])
const commissionsLoading = ref(false)
const commissionsPage = ref(1)
const commissionsPages = ref(0)

const referralLink = computed(() => {
  if (!referralInfo.value) return ''
  return `${window.location.origin}/register?ref=${referralInfo.value.referral_code}`
})

async function copyToClipboard(text: string, type: 'code' | 'link') {
  try {
    await navigator.clipboard.writeText(text)
    copied.value = type
    setTimeout(() => { copied.value = null }, 2000)
  } catch {
    console.error('Failed to copy')
  }
}

async function fetchInfo() {
  infoLoading.value = true
  try {
    referralInfo.value = await referralAPI.getReferralInfo()
  } catch {
    console.error('Failed to fetch referral info')
  } finally {
    infoLoading.value = false
  }
}

async function loadUsers(page = 1) {
  usersLoading.value = true
  try {
    const res = await referralAPI.getReferredUsers(page)
    referredUsers.value = res.items
    usersPage.value = res.page
    usersPages.value = res.pages
  } catch {
    console.error('Failed to fetch referred users')
  } finally {
    usersLoading.value = false
  }
}

async function loadCommissions(page = 1) {
  commissionsLoading.value = true
  try {
    const res = await referralAPI.getCommissions(page)
    commissions.value = res.items
    commissionsPage.value = res.page
    commissionsPages.value = res.pages
  } catch {
    console.error('Failed to fetch commissions')
  } finally {
    commissionsLoading.value = false
  }
}

onMounted(() => {
  fetchInfo()
  loadUsers()
  loadCommissions()
  if (!appStore.cachedPublicSettings) {
    appStore.fetchPublicSettings()
  }
})
</script>
