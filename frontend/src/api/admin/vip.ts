import usersAPI from './users'
import settingsAPI from './settings'
import type { AdminUser, PaginatedResponse, VIPRule } from '@/types'

export interface VIPUsersFilters {
  search?: string
  status?: 'active' | 'disabled'
}

export interface VIPSettingsPayload {
  vip_enabled: boolean
  vip_rules: string
}

async function listUsers(
  page = 1,
  pageSize = 20,
  filters?: VIPUsersFilters
): Promise<PaginatedResponse<AdminUser>> {
  return usersAPI.list(page, pageSize, {
    search: filters?.search,
    status: filters?.status
  })
}

async function getSettings(): Promise<{ vip_enabled: boolean; vip_rules: string }> {
  const settings = await settingsAPI.getSettings()
  return {
    vip_enabled: settings.vip_enabled,
    vip_rules: settings.vip_rules
  }
}

async function updateSettings(payload: VIPSettingsPayload): Promise<{ vip_enabled: boolean; vip_rules: string }> {
  const settings = await settingsAPI.updateSettings(payload)
  return {
    vip_enabled: settings.vip_enabled,
    vip_rules: settings.vip_rules
  }
}

export function parseVIPRules(raw: string): VIPRule[] {
  if (!raw.trim()) return []
  try {
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? (parsed as VIPRule[]) : []
  } catch {
    return []
  }
}

const vipAPI = {
  listUsers,
  getSettings,
  updateSettings,
  parseVIPRules
}

export default vipAPI
