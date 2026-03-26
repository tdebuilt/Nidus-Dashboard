import { writable } from 'svelte/store'
import { api } from '../api/client'

export type ServiceStatus = 'up' | 'down' | 'pending'

export const serviceStatuses = writable<Record<string, ServiceStatus>>({})

let pollTimer: ReturnType<typeof setInterval> | null = null

async function fetchStatuses(): Promise<void> {
  try {
    const data = await api.get<{ statuses: Record<string, boolean> }>('/api/services/status')
    const result: Record<string, ServiceStatus> = {}
    for (const [type, reachable] of Object.entries(data.statuses)) {
      result[type] = reachable ? 'up' : 'down'
    }
    serviceStatuses.set(result)
  } catch {
    // Keep existing statuses on error
  }
}

export function startServiceStatusPolling(serviceTypes: string[]): void {
  const pending: Record<string, ServiceStatus> = {}
  for (const type of serviceTypes) {
    pending[type] = 'pending'
  }
  serviceStatuses.set(pending)

  fetchStatuses()
  pollTimer = setInterval(fetchStatuses, 60000)
}

export function stopServiceStatusPolling(): void {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}
