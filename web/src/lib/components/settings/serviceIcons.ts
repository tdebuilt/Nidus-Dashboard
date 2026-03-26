/** Maps each service type to a Lucide icon name. */
export const serviceIconMap: Record<string, string> = {
  portainer: 'Container',
  proxmox: 'Server',
  homeassistant: 'House',
  adguard: 'ShieldCheck',
  pihole: 'ShieldBan',
  jdownloader: 'Download',
  transmission: 'ArrowDownUp',
  uptimekuma: 'Activity',
  plex: 'Play',
  jellyfin: 'Film',
  sonarr: 'Tv',
  radarr: 'Clapperboard',
  lidarr: 'Music',
  prowlarr: 'Search',
  reolink: 'Camera',
}

/** Returns the Lucide icon name for a service type, with fallback. */
export function getServiceIcon(type: string): string {
  return serviceIconMap[type] || 'Server'
}

/** Maps each service type to a display group. */
export const serviceGroupMap: Record<string, string> = {
  portainer: 'infra',
  proxmox: 'infra',
  homeassistant: 'home',
  adguard: 'network',
  pihole: 'network',
  jdownloader: 'downloads',
  transmission: 'downloads',
  uptimekuma: 'monitoring',
  plex: 'media',
  jellyfin: 'media',
  sonarr: 'media',
  radarr: 'media',
  lidarr: 'media',
  prowlarr: 'media',
  reolink: 'home',
}

/** Display order for service groups. */
export const groupOrder = ['infra', 'media', 'home', 'network', 'downloads', 'monitoring']

/** Returns the group for a service type, with fallback. */
export function getServiceGroup(type: string): string {
  return serviceGroupMap[type] || 'other'
}
