<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="card">
        <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
          <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
            <div>
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">VIP 规则</h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                VIP 升级按用户累计实付人民币计算，模型倍率支持精确模型名或末尾 `*` 通配。
              </p>
            </div>
            <div class="flex items-center gap-3">
              <span class="text-sm font-medium text-gray-700 dark:text-gray-300">启用 VIP</span>
              <Toggle v-model="vipEnabled" />
              <button class="btn btn-primary" :disabled="saving" @click="saveSettings">
                {{ saving ? '保存中...' : '保存 VIP 设置' }}
              </button>
            </div>
          </div>
        </div>

        <div class="space-y-4 p-6">
          <div
            v-for="(rule, index) in vipRules"
            :key="`${rule.level_code || 'vip'}-${index}`"
            class="rounded-2xl border border-gray-200 bg-gray-50/70 p-4 dark:border-dark-700 dark:bg-dark-800/50"
          >
            <div class="mb-4 flex items-center justify-between gap-3">
              <div class="text-sm font-semibold text-gray-900 dark:text-white">
                VIP 等级 {{ index + 1 }}
              </div>
              <button class="btn btn-secondary px-3 py-1.5 text-xs" @click="removeVIPRule(index)">
                删除等级
              </button>
            </div>

            <div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
              <div>
                <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                  等级编码
                </label>
                <input v-model="rule.level_code" class="input" placeholder="vip1" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                  等级名称
                </label>
                <input v-model="rule.level_name" class="input" placeholder="VIP1" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                  规则键
                </label>
                <input v-model="rule.rule_key" class="input" placeholder="vip1" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                  累计实付门槛
                </label>
                <input v-model.number="rule.required_recharge" type="number" min="0" step="0.01" class="input" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                  累计消费门槛
                </label>
                <input v-model.number="rule.required_spend" type="number" min="0" step="0.01" class="input" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                  默认倍率
                </label>
                <input v-model.number="rule.multiplier" type="number" min="0" step="0.01" class="input" />
              </div>
            </div>

            <div class="mt-4 rounded-xl border border-dashed border-gray-300 p-4 dark:border-dark-600">
              <div class="mb-3 flex items-center justify-between gap-3">
                <div>
                  <div class="text-sm font-medium text-gray-900 dark:text-white">模型倍率关联</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">
                    例如 `gpt-5*`、`claude-sonnet-4-5`
                  </div>
                </div>
                <button class="btn btn-secondary px-3 py-1.5 text-xs" @click="addModelRule(rule)">
                  添加模型规则
                </button>
              </div>

              <div v-if="rule.model_rules.length === 0" class="text-sm text-gray-500 dark:text-gray-400">
                未配置模型专属倍率，默认使用上方倍率。
              </div>

              <div v-else class="space-y-3">
                <div
                  v-for="(modelRule, modelIndex) in rule.model_rules"
                  :key="`${modelRule.model_pattern}-${modelIndex}`"
                  class="grid grid-cols-1 gap-3 lg:grid-cols-[1fr_180px_auto]"
                >
                  <input
                    v-model="modelRule.model_pattern"
                    class="input"
                    placeholder="gpt-5*"
                  />
                  <input
                    v-model.number="modelRule.multiplier"
                    type="number"
                    min="0"
                    step="0.01"
                    class="input"
                    placeholder="0.95"
                  />
                  <button class="btn btn-secondary px-3 py-2 text-xs" @click="removeModelRule(rule, modelIndex)">
                    删除
                  </button>
                </div>
              </div>
            </div>
          </div>

          <button class="btn btn-secondary" @click="addVIPRule">新增 VIP 等级</button>
        </div>
      </div>

      <div class="card">
        <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
          <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
            <div>
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">VIP 用户列表</h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                查看当前用户 VIP 等级、升级进度和下一等级门槛。
              </p>
            </div>
            <div class="flex w-full flex-col gap-3 sm:flex-row lg:w-auto">
              <input
                v-model.trim="filters.search"
                class="input sm:w-64"
                placeholder="搜索邮箱或用户名"
                @keyup.enter="applyFilters"
              />
              <select v-model="filters.status" class="input sm:w-40" @change="applyFilters">
                <option value="">全部状态</option>
                <option value="active">Active</option>
                <option value="disabled">Disabled</option>
              </select>
              <button class="btn btn-secondary" :disabled="loadingUsers" @click="loadUsers">
                刷新
              </button>
            </div>
          </div>
        </div>

        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
            <thead class="bg-gray-50 dark:bg-dark-800/80">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">用户</th>
                <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">状态</th>
                <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">当前 VIP</th>
                <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">默认倍率</th>
                <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">累计实付</th>
                <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">累计消费</th>
                <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">下一等级</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 bg-white dark:divide-dark-700 dark:bg-dark-900/30">
              <tr v-if="loadingUsers">
                <td colspan="7" class="px-6 py-10 text-center text-sm text-gray-500 dark:text-gray-400">
                  加载中...
                </td>
              </tr>
              <tr v-else-if="vipUsers.length === 0">
                <td colspan="7" class="px-6 py-10 text-center text-sm text-gray-500 dark:text-gray-400">
                  暂无用户数据
                </td>
              </tr>
              <tr v-for="user in vipUsers" :key="user.id">
                <td class="px-6 py-4">
                  <div class="text-sm font-medium text-gray-900 dark:text-white">{{ user.email }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">{{ user.username || '-' }}</div>
                </td>
                <td class="px-6 py-4 text-sm text-gray-700 dark:text-gray-300">{{ user.status }}</td>
                <td class="px-6 py-4">
                  <div class="text-sm font-semibold text-amber-700 dark:text-amber-300">
                    {{ user.current_vip?.level_name || 'VIP0' }}
                  </div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">
                    进度 {{ Math.round(user.current_vip?.progress_percent || 0) }}%
                  </div>
                </td>
                <td class="px-6 py-4 text-sm text-gray-700 dark:text-gray-300">
                  {{ formatMultiplier(user.current_vip?.base_multiplier) }}
                </td>
                <td class="px-6 py-4 text-sm text-gray-700 dark:text-gray-300">
                  ¥{{ (user.current_vip?.recharge_total || 0).toFixed(2) }}
                </td>
                <td class="px-6 py-4 text-sm text-gray-700 dark:text-gray-300">
                  ${{ (user.current_vip?.spend_total || 0).toFixed(2) }}
                </td>
                <td class="px-6 py-4">
                  <div v-if="user.next_vip" class="text-sm text-gray-700 dark:text-gray-300">
                    {{ user.next_vip.level_name }}
                  </div>
                  <div v-if="user.next_vip" class="text-xs text-gray-500 dark:text-gray-400">
                    {{ user.next_vip.unlock_condition_label }}
                  </div>
                  <div v-else class="text-sm text-emerald-600 dark:text-emerald-400">已达最高等级</div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="border-t border-gray-100 px-6 py-4 dark:border-dark-700">
          <Pagination
            v-if="pagination.total > 0"
            :page="pagination.page"
            :total="pagination.total"
            :page-size="pagination.page_size"
            @update:page="handlePageChange"
            @update:pageSize="handlePageSizeChange"
          />
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Toggle from '@/components/common/Toggle.vue'
import Pagination from '@/components/common/Pagination.vue'
import { useAppStore } from '@/stores'
import { adminAPI } from '@/api'
import type { AdminUser, VIPModelMultiplierRule, VIPRule } from '@/types'

interface EditableVIPRule extends Omit<VIPRule, 'model_multipliers'> {
  model_rules: VIPModelMultiplierRule[]
}

const appStore = useAppStore()

const loadingUsers = ref(false)
const saving = ref(false)
const vipEnabled = ref(false)
const vipRules = ref<EditableVIPRule[]>([])
const vipUsers = ref<AdminUser[]>([])

const filters = reactive({
  search: '',
  status: ''
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  pages: 0
})

function createEmptyVIPRule(index = vipRules.value.length + 1): EditableVIPRule {
  return {
    level_code: `vip${index}`,
    level_name: `VIP${index}`,
    required_recharge: 0,
    required_spend: 0,
    multiplier: 1,
    rule_key: `vip${index}`,
    model_rules: []
  }
}

function toEditableRules(rules: VIPRule[]): EditableVIPRule[] {
  return rules.map((rule, index) => ({
    level_code: rule.level_code || `vip${index + 1}`,
    level_name: rule.level_name || `VIP${index + 1}`,
    required_recharge: Number(rule.required_recharge || 0),
    required_spend: Number(rule.required_spend || 0),
    multiplier: Number(rule.multiplier || 1),
    rule_key: rule.rule_key || rule.level_code || `vip${index + 1}`,
    model_rules: Object.entries(rule.model_multipliers || {}).map(([model_pattern, multiplier]) => ({
      model_pattern,
      multiplier
    }))
  }))
}

function serializeRules(): VIPRule[] {
  return vipRules.value.map((rule, index) => ({
    level_code: (rule.level_code || `vip${index + 1}`).trim(),
    level_name: (rule.level_name || `VIP${index + 1}`).trim(),
    required_recharge: Number(rule.required_recharge || 0),
    required_spend: Number(rule.required_spend || 0),
    multiplier: Number(rule.multiplier || 1),
    rule_key: (rule.rule_key || rule.level_code || `vip${index + 1}`).trim(),
    model_multipliers: rule.model_rules.reduce<Record<string, number>>((acc, item) => {
      const pattern = item.model_pattern.trim()
      const multiplier = Number(item.multiplier || 0)
      if (pattern && multiplier > 0) {
        acc[pattern] = multiplier
      }
      return acc
    }, {})
  }))
}

function addVIPRule() {
  vipRules.value.push(createEmptyVIPRule())
}

function removeVIPRule(index: number) {
  vipRules.value.splice(index, 1)
}

function addModelRule(rule: EditableVIPRule) {
  rule.model_rules.push({
    model_pattern: '',
    multiplier: 1
  })
}

function removeModelRule(rule: EditableVIPRule, index: number) {
  rule.model_rules.splice(index, 1)
}

function formatMultiplier(value?: number) {
  return `${Number(value || 1).toFixed(2)}x`
}

async function loadSettings() {
  const settings = await adminAPI.vip.getSettings()
  vipEnabled.value = settings.vip_enabled
  vipRules.value = toEditableRules(adminAPI.vip.parseVIPRules(settings.vip_rules))
}

async function loadUsers() {
  loadingUsers.value = true
  try {
    const res = await adminAPI.vip.listUsers(pagination.page, pagination.page_size, {
      search: filters.search || undefined,
      status: filters.status ? (filters.status as 'active' | 'disabled') : undefined
    })
    vipUsers.value = res.items
    pagination.total = res.total
    pagination.pages = res.pages
    pagination.page = res.page
    pagination.page_size = res.page_size
  } catch (error: any) {
    appStore.showError(`加载 VIP 用户失败: ${error?.message || 'unknown error'}`)
  } finally {
    loadingUsers.value = false
  }
}

async function saveSettings() {
  saving.value = true
  try {
    const serializedRules = serializeRules()
    await adminAPI.vip.updateSettings({
      vip_enabled: vipEnabled.value,
      vip_rules: JSON.stringify(serializedRules)
    })
    appStore.showSuccess('VIP 设置已保存')
    await loadSettings()
    await loadUsers()
  } catch (error: any) {
    appStore.showError(`保存 VIP 设置失败: ${error?.message || 'unknown error'}`)
  } finally {
    saving.value = false
  }
}

function applyFilters() {
  pagination.page = 1
  loadUsers()
}

function handlePageChange(page: number) {
  pagination.page = page
  loadUsers()
}

function handlePageSizeChange(pageSize: number) {
  pagination.page_size = pageSize
  pagination.page = 1
  loadUsers()
}

onMounted(async () => {
  try {
    await loadSettings()
    if (vipRules.value.length === 0) {
      vipRules.value = [createEmptyVIPRule()]
    }
    await loadUsers()
  } catch (error: any) {
    appStore.showError(`初始化 VIP 页面失败: ${error?.message || 'unknown error'}`)
  }
})
</script>
