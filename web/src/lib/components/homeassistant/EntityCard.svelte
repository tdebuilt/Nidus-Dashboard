<script lang="ts">
  import LightCard from './LightCard.svelte'
  import SwitchCard from './SwitchCard.svelte'
  import SensorCard from './SensorCard.svelte'
  import ClimateCard from './ClimateCard.svelte'
  import CameraCard from './CameraCard.svelte'
  import ButtonCard from './ButtonCard.svelte'
  import CoverCard from './CoverCard.svelte'
  import LockCard from './LockCard.svelte'

  interface EntityInfo {
    entity_id: string
    domain: string
    name: string
    state: string
    attributes: Record<string, unknown>
    icon?: string
    unit_of_measurement?: string
    last_changed: string
  }

  interface Props {
    entity: EntityInfo
    onAction?: () => void
    cameraWidth?: number
    cameraHeight?: number
    onCameraResize?: (width: number) => void
  }

  const { entity, onAction, cameraWidth, cameraHeight, onCameraResize }: Props = $props()
</script>

{#if entity.domain === 'light'}
  <LightCard {entity} {onAction} />
{:else if entity.domain === 'switch' || entity.domain === 'input_boolean'}
  <SwitchCard {entity} {onAction} />
{:else if entity.domain === 'sensor'}
  <SensorCard {entity} />
{:else if entity.domain === 'climate'}
  <ClimateCard {entity} {onAction} />
{:else if entity.domain === 'camera'}
  <CameraCard {entity} width={cameraWidth} height={cameraHeight} onResize={onCameraResize} />
{:else if entity.domain === 'cover'}
  <CoverCard {entity} {onAction} />
{:else if entity.domain === 'button' || entity.domain === 'input_button' || entity.domain === 'script' || entity.domain === 'scene'}
  <ButtonCard {entity} {onAction} />
{:else if entity.domain === 'lock'}
  <LockCard {entity} {onAction} />
{:else}
  <SensorCard {entity} />
{/if}
