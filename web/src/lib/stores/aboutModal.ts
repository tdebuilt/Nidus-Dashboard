import { writable } from 'svelte/store'

export const aboutModalOpen = writable(false)

export function openAboutModal() {
  aboutModalOpen.set(true)
}

export function closeAboutModal() {
  aboutModalOpen.set(false)
}
