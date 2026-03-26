import { writable, get } from 'svelte/store'
import { api } from '../api/client'

const DEFAULT_INTERVAL = 30000

function createPollingStore() {
  const { subscribe, set } = writable<number>(DEFAULT_INTERVAL)

  async function load() {
    try {
      const prefs = await api.get<{ refresh_interval: number }>('/api/preferences')
      const interval = prefs?.refresh_interval ?? 30
      set(Math.max(5, interval) * 1000)
    } catch {
      set(DEFAULT_INTERVAL)
    }
  }

  async function update(seconds: number) {
    set(seconds * 1000)
    try {
      await api.put('/api/preferences', { refresh_interval: seconds })
    } catch {
      // Keep local value even if save fails
    }
  }

  return { subscribe, load, update, get: () => get({ subscribe }) }
}

export const pollingInterval = createPollingStore()
