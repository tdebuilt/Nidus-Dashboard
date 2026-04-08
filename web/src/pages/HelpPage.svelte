<script lang="ts">
  import {
    HelpCircle, ChevronDown, ChevronRight,
    Container, Server, Home, Shield, Download, Link,
    HeartPulse, MonitorPlay, CloudSun, CalendarDays, Rss,
    ShieldCheck, Film, Camera, TrendingUp, Activity,
    LayoutDashboard, Bell, Webhook, Users, SquareTerminal
  } from 'lucide-svelte'
  import { t } from '../lib/i18n'
  import { navigate } from '../lib/stores/router'

  type WidgetSection = {
    key: string
    icon: typeof Container
    hasDualAuth?: boolean
    hasConfig?: boolean
    configExamples?: Array<{ label?: string; code: string }>
    setupStepKeys?: string[]
    featureKeys: string[]
  }

  const sections: WidgetSection[] = [
    {
      key: 'docker',
      icon: Container,
      hasDualAuth: true,
      setupStepKeys: ['step1', 'step2', 'step3', 'step4'],
      featureKeys: ['stacks', 'actions', 'updates', 'health']
    },
    {
      key: 'proxmox',
      icon: Server,
      hasDualAuth: true,
      setupStepKeys: ['step1', 'step2', 'step3', 'step4', 'step5'],
      featureKeys: ['list', 'resources', 'actions', 'uptime']
    },
    {
      key: 'homeassistant',
      icon: Home,
      setupStepKeys: ['step1', 'step2', 'step3', 'step4'],
      featureKeys: ['domains', 'toggle', 'climate', 'camera', 'sensors']
    },
    {
      key: 'adguard',
      icon: Shield,
      featureKeys: ['stats', 'response', 'toggle', 'filters']
    },
    {
      key: 'pihole',
      icon: ShieldCheck,
      featureKeys: ['stats', 'toggle', 'topDomains']
    },
    {
      key: 'jdownloader',
      icon: Download,
      setupStepKeys: ['step1', 'step2', 'step3', 'step4'],
      featureKeys: ['queue', 'progress', 'add', 'control']
    },
    {
      key: 'transmission',
      icon: Activity,
      featureKeys: ['torrents', 'speed', 'control', 'cleanup']
    },
    {
      key: 'qbittorrent',
      icon: Download,
      featureKeys: ['torrents', 'speed', 'control', 'search', 'cleanup']
    },
    {
      key: 'mediaserver',
      icon: MonitorPlay,
      setupStepKeys: ['plex1', 'plex2', 'plex3', 'jellyfin1', 'jellyfin2'],
      featureKeys: ['sessions', 'poster', 'progress', 'user']
    },
    {
      key: 'uptimekuma',
      icon: HeartPulse,
      featureKeys: ['monitors', 'status', 'uptime']
    },
    {
      key: 'arr',
      icon: Film,
      featureKeys: ['overview', 'queue', 'calendar', 'status']
    },
    {
      key: 'reolink',
      icon: Camera,
      featureKeys: ['snapshot', 'live', 'fullscreen', 'multiCam']
    },
    {
      key: 'grafana',
      icon: LayoutDashboard,
      hasConfig: true,
      setupStepKeys: ['step1', 'step2', 'step3', 'step4', 'step5'],
      featureKeys: ['embed', 'browse', 'multiPanel', 'theme']
    },
    {
      key: 'weather',
      icon: CloudSun,
      hasConfig: true,
      featureKeys: ['current', 'forecast', 'details']
    },
    {
      key: 'calendar',
      icon: CalendarDays,
      hasConfig: true,
      featureKeys: ['events', 'ical', 'multiCal']
    },
    {
      key: 'rss',
      icon: Rss,
      hasConfig: true,
      featureKeys: ['feeds', 'multiSource', 'preview']
    },
    {
      key: 'finance',
      icon: TrendingUp,
      hasConfig: true,
      featureKeys: ['quotes', 'chart', 'portfolio']
    },
    {
      key: 'terminal',
      icon: SquareTerminal,
      hasConfig: true,
      featureKeys: ['ssh', 'resize', 'reconnect']
    },
    {
      key: 'system',
      icon: Server,
      featureKeys: ['cpu', 'memory', 'disk', 'network']
    },
    {
      key: 'applinks',
      icon: Link,
      hasConfig: true,
      featureKeys: ['links', 'health', 'icons']
    },
    {
      key: 'notifications',
      icon: Bell,
      featureKeys: ['gotify', 'rules', 'events', 'test']
    },
    {
      key: 'webhooks_help',
      icon: Webhook,
      featureKeys: ['incoming', 'hmac', 'actions', 'notify']
    },
    {
      key: 'users',
      icon: Users,
      featureKeys: ['roles', 'invite', 'manage']
    }
  ]

  let openSections = $state<Record<string, boolean>>({})

  function toggle(key: string) {
    openSections = { ...openSections, [key]: !openSections[key] }
  }
</script>

<div class="mx-auto max-w-3xl">
  <!-- Title -->
  <div class="mb-6 flex items-center gap-3">
    <HelpCircle size={28} class="text-[var(--color-primary)]" />
    <h1 class="text-2xl font-bold text-[var(--color-text)]">{$t('help.title')}</h1>
  </div>

  <!-- Getting started -->
  <div class="mb-8 rounded-lg border border-[var(--color-border)] bg-[var(--color-card-bg)] p-4">
    <h2 class="mb-2 text-lg font-semibold text-[var(--color-text)]">{$t('help.gettingStarted')}</h2>
    <p class="text-sm text-[var(--color-text-secondary)]">
      {$t('help.gettingStartedText')}
      <button
        onclick={() => navigate('/settings')}
        class="font-medium text-[var(--color-primary)] underline hover:opacity-80"
      >{$t('help.gettingStartedLink')}</button>{$t('help.gettingStartedAfter')}
    </p>
  </div>

  <!-- Widget sections -->
  <div class="space-y-3">
    {#each sections as section (section.key)}
      {@const isOpen = openSections[section.key] ?? false}
      <div class="rounded-lg border border-[var(--color-border)] bg-[var(--color-card-bg)] overflow-hidden">
        <!-- Header -->
        <button
          onclick={() => toggle(section.key)}
          class="flex w-full items-center gap-3 px-4 py-3 text-start transition-colors hover:bg-[var(--color-sidebar-hover)]"
          data-testid="help-section-{section.key}"
        >
          {#if isOpen}
            <ChevronDown size={18} class="text-[var(--color-text-muted)] shrink-0" />
          {:else}
            <ChevronRight size={18} class="text-[var(--color-text-muted)] shrink-0" />
          {/if}
          <section.icon size={20} class="text-[var(--color-primary)] shrink-0" />
          <span class="font-semibold text-[var(--color-text)]">{$t(`help.${section.key}.title`)}</span>
        </button>

        <!-- Content -->
        {#if isOpen}
          <div class="border-t border-[var(--color-border)] px-4 py-4 space-y-4">
            <!-- Prerequisite -->
            <div>
              <span class="inline-block rounded-full bg-[var(--color-primary)]/10 px-2.5 py-0.5 text-xs font-medium text-[var(--color-primary)]">
                {$t('help.prerequisite')}
              </span>
              <p class="mt-1.5 text-sm text-[var(--color-text-secondary)]">
                {$t(`help.${section.key}.prerequisite`)}
              </p>
              {#if section.hasDualAuth}
                <p class="mt-2 text-xs text-[var(--color-text-muted)]">{$t('help.authModes')}</p>
                <ul class="mt-1 space-y-0.5 text-sm text-[var(--color-text-secondary)]">
                  <li class="flex items-start gap-2">
                    <span class="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full bg-[var(--color-primary)]"></span>
                    <span><span class="font-medium">{$t('help.authTokenMode')}</span> — {$t(`help.${section.key}.authToken`)}</span>
                  </li>
                  <li class="flex items-start gap-2">
                    <span class="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full bg-[var(--color-text-muted)]"></span>
                    <span><span class="font-medium">{$t('help.authUserPassMode')}</span> — {$t(`help.${section.key}.authUserPass`)}</span>
                  </li>
                </ul>
              {/if}
              {#if section.setupStepKeys}
                <p class="mt-3 text-xs font-medium text-[var(--color-text-muted)]">{$t('help.setupSteps')}</p>
                <ol class="mt-1 space-y-1">
                  {#each section.setupStepKeys as stepKey, i (stepKey)}
                    <li class="flex items-start gap-2 text-sm text-[var(--color-text-secondary)]">
                      <span class="mt-0.5 flex h-4 w-4 shrink-0 items-center justify-center rounded-full bg-[var(--color-primary)]/15 text-[10px] font-bold text-[var(--color-primary)]">{i + 1}</span>
                      {$t(`help.${section.key}.setupSteps.${stepKey}`)}
                    </li>
                  {/each}
                </ol>
              {/if}
            </div>

            <!-- Config -->
            <div>
              <span class="inline-block rounded-full bg-[var(--color-text-muted)]/10 px-2.5 py-0.5 text-xs font-medium text-[var(--color-text-muted)]">
                {$t('help.config')}
              </span>
              {#if section.hasConfig}
                <p class="mt-1.5 text-sm text-[var(--color-text-secondary)]">
                  {$t(`help.${section.key}.configDesc`)}
                </p>
              {:else}
                <p class="mt-1.5 text-sm text-[var(--color-text-secondary)]">
                  {$t('help.noConfigNeeded')}
                </p>
              {/if}
            </div>

            <!-- Features -->
            <div>
              <span class="inline-block rounded-full bg-green-500/10 px-2.5 py-0.5 text-xs font-medium text-green-500">
                {$t('help.features')}
              </span>
              <ul class="mt-1.5 space-y-1 text-sm text-[var(--color-text-secondary)]">
                {#each section.featureKeys as fk (fk)}
                  <li class="flex items-start gap-2">
                    <span class="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full bg-[var(--color-primary)]"></span>
                    {$t(`help.${section.key}.features.${fk}`)}
                  </li>
                {/each}
              </ul>
            </div>
          </div>
        {/if}
      </div>
    {/each}
  </div>
</div>
