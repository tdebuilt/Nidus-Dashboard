import { writable } from 'svelte/store'
import { api } from '../api/client'
import { registerTheme, unregisterTheme } from '../themes'
import { parseThemeJSON } from '../themes'

interface CustomThemeRecord {
  id: number
  name: string
  theme_json: string
  created_at: string
  updated_at: string
}

function createCustomThemesStore() {
  const { subscribe, set, update } = writable<CustomThemeRecord[]>([])

  return {
    subscribe,
    async load() {
      try {
        const records = await api.get<CustomThemeRecord[]>('/api/themes')
        set(records ?? [])
        for (const record of records ?? []) {
          try {
            const parsed = JSON.parse(record.theme_json)
            const theme = parseThemeJSON(parsed)
            if (typeof theme !== 'string') {
              registerTheme(theme)
            }
          } catch {
            // skip invalid themes
          }
        }
      } catch {
        set([])
      }
    },

    async create(name: string, themeJSON: string): Promise<CustomThemeRecord | null> {
      try {
        const record = await api.post<CustomThemeRecord>('/api/themes', { name, theme_json: themeJSON })
        const parsed = parseThemeJSON(JSON.parse(themeJSON))
        if (typeof parsed !== 'string') {
          registerTheme(parsed)
        }
        update(list => [...list, record])
        return record
      } catch {
        return null
      }
    },

    async update(id: number, name: string, themeJSON: string): Promise<boolean> {
      try {
        await api.put(`/api/themes/${id}`, { name, theme_json: themeJSON })
        const parsed = parseThemeJSON(JSON.parse(themeJSON))
        if (typeof parsed !== 'string') {
          registerTheme(parsed)
        }
        update(list => list.map(t => t.id === id ? { ...t, name, theme_json: themeJSON } : t))
        return true
      } catch {
        return false
      }
    },

    async remove(id: number, themeId: string): Promise<boolean> {
      try {
        await api.delete(`/api/themes/${id}`)
        unregisterTheme(themeId)
        update(list => list.filter(t => t.id !== id))
        return true
      } catch {
        return false
      }
    },
  }
}

export const customThemes = createCustomThemesStore()
