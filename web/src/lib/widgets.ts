import { writable } from 'svelte/store';
import { api } from './api';

export interface Widget {
  id: number;
  category_id: number;
  type: string;
  title: string;
  config: string;
  collapsed: boolean;
  pos_x: number;
  pos_y: number;
  width: number;
  height: number;
  created_at: string;
  updated_at: string;
}

export interface WidgetLayout {
  id: number;
  pos_x: number;
  pos_y: number;
  width: number;
  height: number;
}

export const widgets = writable<Widget[]>([]);

export const WIDGET_TYPES = [
  { type: 'portainer-stack', label: 'Stack Docker', description: 'Groupe de containers Docker' },
  { type: 'portainer-container', label: 'Container Docker', description: 'Container Docker individuel' },
  { type: 'proxmox-vm', label: 'VM/LXC Proxmox', description: 'Machine virtuelle ou container' },
  { type: 'homeassistant-entity', label: 'Entité Home Assistant', description: 'Capteur, switch, lumière...' },
  { type: 'adguard-stats', label: 'AdGuard Stats', description: 'Statistiques DNS et blocage' },
  { type: 'jdownloader-queue', label: 'JDownloader Queue', description: 'File de téléchargement' },
  { type: 'transmission-torrents', label: 'Transmission', description: 'Torrents actifs' },
  { type: 'app-link', label: 'Lien applicatif', description: 'Raccourci vers une URL ou app' },
];

export async function fetchWidgets(categoryId: number): Promise<void> {
  try {
    const data = await api.get<Widget[]>(`/api/categories/${categoryId}/widgets`);
    widgets.set(data);
  } catch {
    widgets.set([]);
  }
}

export async function createWidget(
  categoryId: number,
  type: string,
  title: string,
  config: string,
  width = 2,
  height = 2,
): Promise<Widget> {
  const existing = await api.get<Widget[]>(`/api/categories/${categoryId}/widgets`);
  const maxY = existing.reduce((max, w) => Math.max(max, w.pos_y + w.height), 0);
  const created = await api.post<Widget>(`/api/categories/${categoryId}/widgets`, {
    type,
    title,
    config,
    pos_x: 0,
    pos_y: maxY,
    width,
    height,
  });
  await fetchWidgets(categoryId);
  return created;
}

export async function updateWidget(id: number, type: string, title: string, config: string): Promise<void> {
  await api.put(`/api/widgets/${id}`, { type, title, config });
}

export async function deleteWidget(id: number, categoryId: number): Promise<void> {
  await api.delete(`/api/widgets/${id}`);
  await fetchWidgets(categoryId);
}

export async function saveLayout(layouts: WidgetLayout[]): Promise<void> {
  await api.put('/api/widgets/layout', { widgets: layouts });
}
