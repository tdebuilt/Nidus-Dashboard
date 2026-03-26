import { writable } from 'svelte/store'
import { api } from '../api/client'

export const isDocker = writable(false)

function createVersionStore() {
  const { subscribe, set } = writable<string>('')

  async function load() {
    try {
      const data = await api.get<{ version: string; is_docker: boolean }>('/api/version')
      set(data?.version ?? 'dev')
      isDocker.set(data?.is_docker ?? false)
    } catch {
      set('dev')
    }
  }

  return { subscribe, load }
}

export const appVersion = createVersionStore()
