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

type SetFn = (v: CustomThemeRecord[]) => void
type UpdateFn = (fn: (list: CustomThemeRecord[]) => CustomThemeRecord[]) => void

async function loadThemes(set: SetFn) {
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
      } catch (e) {
        console.error('customThemes: failed to parse theme', record.name, e)
      }
    }
  } catch (e) {
    console.error('customThemes: failed to load themes', e)
    set([])
  }
}

async function createThemeRecord(
  name: string,
  themeJSON: string,
  update: UpdateFn,
): Promise<CustomThemeRecord | null> {
  try {
    const record = await api.post<CustomThemeRecord>('/api/themes', { name, theme_json: themeJSON })
    const parsed = parseThemeJSON(JSON.parse(themeJSON))
    if (typeof parsed !== 'string') {
      registerTheme(parsed)
    }
    update(list => [...list, record])
    return record
  } catch (e) {
    console.error('customThemes: failed to create theme', e)
    return null
  }
}

async function updateThemeRecord(
  id: number,
  name: string,
  themeJSON: string,
  update: UpdateFn,
): Promise<boolean> {
  try {
    await api.put(`/api/themes/${id}`, { name, theme_json: themeJSON })
    const parsed = parseThemeJSON(JSON.parse(themeJSON))
    if (typeof parsed !== 'string') {
      registerTheme(parsed)
    }
    update(list => list.map(t => t.id === id ? { ...t, name, theme_json: themeJSON } : t))
    return true
  } catch (e) {
    console.error('customThemes: failed to update theme', id, e)
    return false
  }
}

async function removeThemeRecord(
  id: number,
  themeId: string,
  update: UpdateFn,
): Promise<boolean> {
  try {
    await api.delete(`/api/themes/${id}`)
    unregisterTheme(themeId)
    update(list => list.filter(t => t.id !== id))
    return true
  } catch (e) {
    console.error('customThemes: failed to remove theme', id, e)
    return false
  }
}

function createCustomThemesStore() {
  const { subscribe, set, update } = writable<CustomThemeRecord[]>([])

  return {
    subscribe,
    load: () => loadThemes(set),
    create: (name: string, themeJSON: string) => createThemeRecord(name, themeJSON, update),
    update: (id: number, name: string, themeJSON: string) => updateThemeRecord(id, name, themeJSON, update),
    remove: (id: number, themeId: string) => removeThemeRecord(id, themeId, update),
  }
}

export const customThemes = createCustomThemesStore()
