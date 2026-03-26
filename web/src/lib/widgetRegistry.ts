// eslint-disable-next-line @typescript-eslint/no-explicit-any
type AnyComponent = any
import { Container, Monitor, Activity, Shield, Download, Radio, Link, HeartPulse, MonitorPlay, CloudSun, CalendarDays, Rss, Server, ShieldCheck, Film, Camera, TrendingUp, LayoutDashboard } from 'lucide-svelte'

import DockerWidget from './components/docker/DockerWidget.svelte'
import ProxmoxWidget from './components/proxmox/ProxmoxWidget.svelte'
import HomeAssistantWidget from './components/homeassistant/HomeAssistantWidget.svelte'
import AdGuardWidget from './components/adguard/AdGuardWidget.svelte'
import JDownloaderWidget from './components/jdownloader/JDownloaderWidget.svelte'
import TransmissionWidget from './components/transmission/TransmissionWidget.svelte'
import UptimeKumaWidget from './components/uptimekuma/UptimeKumaWidget.svelte'
import MediaServerWidget from './components/mediaserver/MediaServerWidget.svelte'
import WeatherWidget from './components/weather/WeatherWidget.svelte'
import CalendarWidget from './components/calendar/CalendarWidget.svelte'
import RSSWidget from './components/rss/RSSWidget.svelte'
import SystemWidget from './components/system/SystemWidget.svelte'
import AppLinkWidget from './components/applink/AppLinkWidget.svelte'
import PiholeWidget from './components/pihole/PiholeWidget.svelte'
import ArrWidget from './components/arr/ArrWidget.svelte'
import ReolinkWidget from './components/reolink/ReolinkWidget.svelte'
import FinanceWidget from './components/finance/FinanceWidget.svelte'
import GrafanaWidget from './components/grafana/GrafanaWidget.svelte'

import DockerConfig from './components/config/DockerConfig.svelte'
import HomeAssistantConfig from './components/config/HomeAssistantConfig.svelte'
import MediaServerConfig from './components/config/MediaServerConfig.svelte'
import WeatherConfig from './components/config/WeatherConfig.svelte'
import CalendarConfig from './components/config/CalendarConfig.svelte'
import RSSConfig from './components/config/RSSConfig.svelte'
import AppLinkConfig from './components/config/AppLinkConfig.svelte'
import ReolinkConfig from './components/config/ReolinkConfig.svelte'
import FinanceConfig from './components/config/FinanceConfig.svelte'
import GrafanaConfig from './components/config/GrafanaConfig.svelte'

export interface WidgetDefinition {
  type: string
  label: string
  icon: AnyComponent
  component: AnyComponent
  configComponent?: AnyComponent
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
  component: DockerWidget,
  configComponent: DockerConfig,
  serviceType: 'portainer',
})

register({
  type: 'proxmox',
  label: 'Proxmox',
  icon: Monitor,
  component: ProxmoxWidget,
  serviceType: 'proxmox',
})

register({
  type: 'homeassistant',
  label: 'Home Assistant',
  icon: Activity,
  component: HomeAssistantWidget,
  configComponent: HomeAssistantConfig,
  serviceType: 'homeassistant',
  extraProps: ['widgetId', 'widgetType', 'widgetTitle'],
})

register({
  type: 'adguard',
  label: 'AdGuard',
  icon: Shield,
  component: AdGuardWidget,
  serviceType: 'adguard',
})

register({
  type: 'jdownloader',
  label: 'JDownloader',
  icon: Download,
  component: JDownloaderWidget,
  serviceType: 'jdownloader',
})

register({
  type: 'transmission',
  label: 'Transmission',
  icon: Radio,
  component: TransmissionWidget,
  serviceType: 'transmission',
})

register({
  type: 'uptimekuma',
  label: 'Uptime Kuma',
  icon: HeartPulse,
  component: UptimeKumaWidget,
  serviceType: 'uptimekuma',
})

register({
  type: 'mediaserver',
  label: 'Plex / Jellyfin',
  icon: MonitorPlay,
  component: MediaServerWidget,
  configComponent: MediaServerConfig,
  serviceType: ['plex', 'jellyfin'],
})

register({
  type: 'weather',
  label: 'Météo',
  icon: CloudSun,
  component: WeatherWidget,
  configComponent: WeatherConfig,
})

register({
  type: 'calendar',
  label: 'Calendrier',
  icon: CalendarDays,
  component: CalendarWidget,
  configComponent: CalendarConfig,
})

register({
  type: 'rss',
  label: 'Flux RSS',
  icon: Rss,
  component: RSSWidget,
  configComponent: RSSConfig,
})

register({
  type: 'system',
  label: 'Système',
  icon: Server,
  component: SystemWidget,
})

register({
  type: 'applink',
  label: 'Web Links',
  icon: Link,
  component: AppLinkWidget,
  configComponent: AppLinkConfig,
})

register({
  type: 'pihole',
  label: 'Pi-hole',
  icon: ShieldCheck,
  component: PiholeWidget,
  serviceType: 'pihole',
})

register({
  type: 'arr',
  label: 'Sonarr / Radarr',
  icon: Film,
  component: ArrWidget,
  serviceType: ['sonarr', 'radarr', 'lidarr', 'prowlarr'],
})

register({
  type: 'finance',
  label: 'Finance',
  icon: TrendingUp,
  component: FinanceWidget,
  configComponent: FinanceConfig,
})

register({
  type: 'reolink',
  label: 'Cameras',
  icon: Camera,
  component: ReolinkWidget,
  configComponent: ReolinkConfig,
  serviceType: 'reolink',
})

register({
  type: 'grafana',
  label: 'Grafana',
  icon: LayoutDashboard,
  component: GrafanaWidget,
  configComponent: GrafanaConfig,
  serviceType: 'grafana',
})

// --- Public API ---

export function getWidget(type: string): WidgetDefinition | undefined {
  return registry.get(type)
}

export function getAllWidgetTypes(): WidgetDefinition[] {
  return Array.from(registry.values())
}

export function getWidgetComponent(type: string): AnyComponent | undefined {
  return registry.get(type)?.component
}

export function getConfigComponent(type: string): AnyComponent | undefined {
  return registry.get(type)?.configComponent
}

export function hasConfig(type: string): boolean {
  return registry.get(type)?.configComponent != null
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
