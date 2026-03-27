import type { Component } from 'svelte'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type SvelteComponent = Component<any>
type LazyComponent = () => Promise<{ default: SvelteComponent }>

import { Container, Monitor, Activity, Shield, Download, Radio, Link, HeartPulse, MonitorPlay, CloudSun, CalendarDays, Rss, Server, ShieldCheck, Film, Camera, TrendingUp, LayoutDashboard } from 'lucide-svelte'

export interface WidgetDefinition {
  type: string
  label: string
  icon: SvelteComponent
  component: LazyComponent
  configComponent?: LazyComponent
  /** Backend service type(s) that must be configured (null for standalone widgets like applink) */
  serviceType?: string | string[]
  /** Extra props passed to the widget component beyond `config` */
  extraProps?: string[]
}

const registry = new Map<string, WidgetDefinition>()

function register(def: WidgetDefinition) {
  registry.set(def.type, def)
}

// --- Register all built-in widgets ---

register({
  type: 'docker',
  label: 'Docker / Portainer',
  icon: Container,
  component: () => import('./components/docker/DockerWidget.svelte'),
  configComponent: () => import('./components/config/DockerConfig.svelte'),
  serviceType: 'portainer',
})

register({
  type: 'proxmox',
  label: 'Proxmox',
  icon: Monitor,
  component: () => import('./components/proxmox/ProxmoxWidget.svelte'),
  serviceType: 'proxmox',
})

register({
  type: 'homeassistant',
  label: 'Home Assistant',
  icon: Activity,
  component: () => import('./components/homeassistant/HomeAssistantWidget.svelte'),
  configComponent: () => import('./components/config/HomeAssistantConfig.svelte'),
  serviceType: 'homeassistant',
  extraProps: ['widgetId', 'widgetType', 'widgetTitle'],
})

register({
  type: 'adguard',
  label: 'AdGuard',
  icon: Shield,
  component: () => import('./components/adguard/AdGuardWidget.svelte'),
  serviceType: 'adguard',
})

register({
  type: 'jdownloader',
  label: 'JDownloader',
  icon: Download,
  component: () => import('./components/jdownloader/JDownloaderWidget.svelte'),
  serviceType: 'jdownloader',
})

register({
  type: 'transmission',
  label: 'Transmission',
  icon: Radio,
  component: () => import('./components/transmission/TransmissionWidget.svelte'),
  serviceType: 'transmission',
})

register({
  type: 'uptimekuma',
  label: 'Uptime Kuma',
  icon: HeartPulse,
  component: () => import('./components/uptimekuma/UptimeKumaWidget.svelte'),
  serviceType: 'uptimekuma',
})

register({
  type: 'mediaserver',
  label: 'Plex / Jellyfin',
  icon: MonitorPlay,
  component: () => import('./components/mediaserver/MediaServerWidget.svelte'),
  configComponent: () => import('./components/config/MediaServerConfig.svelte'),
  serviceType: ['plex', 'jellyfin'],
})

register({
  type: 'weather',
  label: 'Météo',
  icon: CloudSun,
  component: () => import('./components/weather/WeatherWidget.svelte'),
  configComponent: () => import('./components/config/WeatherConfig.svelte'),
})

register({
  type: 'calendar',
  label: 'Calendrier',
  icon: CalendarDays,
  component: () => import('./components/calendar/CalendarWidget.svelte'),
  configComponent: () => import('./components/config/CalendarConfig.svelte'),
})

register({
  type: 'rss',
  label: 'Flux RSS',
  icon: Rss,
  component: () => import('./components/rss/RSSWidget.svelte'),
  configComponent: () => import('./components/config/RSSConfig.svelte'),
})

register({
  type: 'system',
  label: 'Système',
  icon: Server,
  component: () => import('./components/system/SystemWidget.svelte'),
})

register({
  type: 'applink',
  label: 'Web Links',
  icon: Link,
  component: () => import('./components/applink/AppLinkWidget.svelte'),
  configComponent: () => import('./components/config/AppLinkConfig.svelte'),
})

register({
  type: 'pihole',
  label: 'Pi-hole',
  icon: ShieldCheck,
  component: () => import('./components/pihole/PiholeWidget.svelte'),
  serviceType: 'pihole',
})

register({
  type: 'arr',
  label: 'Sonarr / Radarr',
  icon: Film,
  component: () => import('./components/arr/ArrWidget.svelte'),
  serviceType: ['sonarr', 'radarr', 'lidarr', 'prowlarr'],
})

register({
  type: 'finance',
  label: 'Finance',
  icon: TrendingUp,
  component: () => import('./components/finance/FinanceWidget.svelte'),
  configComponent: () => import('./components/config/FinanceConfig.svelte'),
})

register({
  type: 'reolink',
  label: 'Cameras',
  icon: Camera,
  component: () => import('./components/reolink/ReolinkWidget.svelte'),
  configComponent: () => import('./components/config/ReolinkConfig.svelte'),
  serviceType: 'reolink',
})

register({
  type: 'grafana',
  label: 'Grafana',
  icon: LayoutDashboard,
  component: () => import('./components/grafana/GrafanaWidget.svelte'),
  configComponent: () => import('./components/config/GrafanaConfig.svelte'),
  serviceType: 'grafana',
})

// --- Public API ---

export function getWidget(type: string): WidgetDefinition | undefined {
  return registry.get(type)
}

export function getAllWidgetTypes(): WidgetDefinition[] {
  return Array.from(registry.values())
}

export async function loadWidgetComponent(type: string): Promise<AnyComponent | undefined> {
  const def = registry.get(type)
  if (!def) return undefined
  const mod = await def.component()
  return mod.default
}

export async function loadConfigComponent(type: string): Promise<AnyComponent | undefined> {
  const def = registry.get(type)
  if (!def?.configComponent) return undefined
  const mod = await def.configComponent()
  return mod.default
}

/** Map service type → widget type for filtering available widgets */
export function getServiceToWidgetMap(): Record<string, string> {
  const map: Record<string, string> = {}
  for (const def of registry.values()) {
    if (def.serviceType) {
      const types = Array.isArray(def.serviceType) ? def.serviceType : [def.serviceType]
      for (const st of types) {
        map[st] = def.type
      }
    }
  }
  return map
}
