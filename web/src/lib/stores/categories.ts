import { writable } from 'svelte/store'
import { api } from '../api/client'

export interface Category {
  id: number
  name: string
  slug: string
  icon: string
  sort_order: number
}

function createCategoriesStore() {
  const { subscribe, set } = writable<Category[]>([])

  async function load() {
    try {
      const data = await api.get<Category[]>('/api/categories')
      set(data ?? [])
    } catch (e) {
      console.error('categories: failed to load', e)
      set([])
    }
  }

  async function reload() {
    await load()
  }

  return { subscribe, load, reload }
}

export const categories = createCategoriesStore()
