import { writable } from 'svelte/store'
import type { ThemeDefinition } from '../themes'

export type ThemeEditorState = {
  open: boolean
  editingTheme: ThemeDefinition | null
  editingDbId: number | null
}

export const themeEditorState = writable<ThemeEditorState>({
  open: false,
  editingTheme: null,
  editingDbId: null,
})

export function openThemeEditor(theme?: ThemeDefinition, dbId?: number) {
  themeEditorState.set({
    open: true,
    editingTheme: theme ?? null,
    editingDbId: dbId ?? null,
  })
}

export function closeThemeEditor() {
  themeEditorState.set({ open: false, editingTheme: null, editingDbId: null })
}
