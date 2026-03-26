import { writable } from 'svelte/store';
import { api } from './api';

export interface Category {
  id: number;
  name: string;
  slug: string;
  icon: string;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

export const categories = writable<Category[]>([]);
export const activeCategory = writable<number | null>(null);

export async function fetchCategories(): Promise<void> {
  try {
    const data = await api.get<Category[]>('/api/categories');
    categories.set(data);
    if (data.length > 0) {
      activeCategory.update((current) => current ?? data[0].id);
    }
  } catch {
    // Silently fail
  }
}

export async function createCategory(name: string, icon: string): Promise<Category> {
  const created = await api.post<Category>('/api/categories', { name, icon });
  await fetchCategories();
  activeCategory.set(created.id);
  return created;
}

export async function updateCategory(id: number, name: string, icon: string): Promise<void> {
  await api.put(`/api/categories/${id}`, { name, icon });
  await fetchCategories();
}

export async function deleteCategory(id: number): Promise<void> {
  await api.delete(`/api/categories/${id}`);
  await fetchCategories();
  activeCategory.update((current) => current === id ? null : current);
}

export async function reorderCategories(ids: number[]): Promise<void> {
  await api.put('/api/categories/reorder', { ids });
  await fetchCategories();
}
