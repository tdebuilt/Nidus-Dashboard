import { writable } from 'svelte/store'

export const editMode = writable(false)

export function toggleEditMode() {
  editMode.update((v) => !v)
}
