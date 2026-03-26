<script lang="ts">
  import { Plus, Trash2, Search, Loader2, Pencil, Check, X, GripVertical } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { t, translate } from '../../i18n'
  import { isDocker } from '../../stores/version'

  interface CameraEntry {
    name: string
    ip: string
    username: string
    password: string
    channel: number
    source: string
    entity_id: string
  }

  interface DiscoveredCamera {
    ip: string
    name: string
    model: string
  }

  interface Props {
    value?: string
    onchange?: (value: string) => void
  }

  const { value = '{}', onchange }: Props = $props()

  let cameras = $state<CameraEntry[]>([])
  let columns = $state(2)
  let discovering = $state(false)
  let discovered = $state<DiscoveredCamera[]>([])

  // Add form
  let newName = $state('')
  let newIP = $state('')
  let newUsername = $state('admin')
  let newPassword = $state('')
  let newChannel = $state(0)

  $effect(() => {
    try {
      const parsed = JSON.parse(value)
      if (parsed.cameras && Array.isArray(parsed.cameras)) {
        cameras = parsed.cameras
      }
      if (parsed.columns) columns = parsed.columns
    } catch { /* ignored */ }
  })

  function emitChange() {
    const config: Record<string, unknown> = { cameras }
    if (columns !== 2) config.columns = columns
    onchange?.(JSON.stringify(config))
  }

  function addCamera() {
    if (!newName.trim() || !newIP.trim()) return
    cameras = [...cameras, {
      name: newName.trim(),
      ip: newIP.trim(),
      username: newUsername.trim(),
      password: newPassword,
      channel: newChannel,
      source: 'direct',
      entity_id: '',
    }]
    newName = ''
    newIP = ''
    newPassword = ''
    newChannel = 0
    emitChange()
  }

  function addDiscoveredCamera(cam: DiscoveredCamera) {
    cameras = [...cameras, {
      name: cam.name || cam.model || cam.ip,
      ip: cam.ip,
      username: 'admin',
      password: '',
      channel: 0,
      source: 'direct',
      entity_id: '',
    }]
    emitChange()
  }

  function removeCamera(index: number) {
    cameras = cameras.filter((_, i) => i !== index)
    if (editingIndex === index) editingIndex = -1
    else if (editingIndex > index) editingIndex--
    emitChange()
  }

  // Edit state
  let editingIndex = $state(-1)
  let editName = $state('')
  let editIP = $state('')
  let editUsername = $state('')
  let editPassword = $state('')
  let editChannel = $state(0)

  function startEdit(index: number) {
    const cam = cameras[index]
    editingIndex = index
    editName = cam.name
    editIP = cam.ip
    editUsername = cam.username
    editPassword = cam.password
    editChannel = cam.channel
  }

  function cancelEdit() {
    editingIndex = -1
  }

  function saveEdit() {
    if (editingIndex < 0 || !editName.trim() || !editIP.trim()) return
    cameras[editingIndex] = {
      ...cameras[editingIndex],
      name: editName.trim(),
      ip: editIP.trim(),
      username: editUsername.trim(),
      password: editPassword,
      channel: editChannel,
    }
    cameras = [...cameras]
    editingIndex = -1
    emitChange()
  }

  // Drag & drop reorder
  let dragIndex = $state<number | null>(null)
  let dragOverIndex = $state<number | null>(null)

  function handleDragStart(index: number) {
    dragIndex = index
  }

  function handleDragOver(e: DragEvent, index: number) {
    e.preventDefault()
    dragOverIndex = index
  }

  function handleDrop(index: number) {
    if (dragIndex !== null && dragIndex !== index) {
      const reordered = [...cameras]
      const [moved] = reordered.splice(dragIndex, 1)
      reordered.splice(index, 0, moved)
      cameras = reordered
      editingIndex = -1
      emitChange()
    }
    dragIndex = null
    dragOverIndex = null
  }

  function handleDragEnd() {
    dragIndex = null
    dragOverIndex = null
  }

  async function discover() {
    discovering = true
    discovered = []
    try {
      discovered = await api.get<DiscoveredCamera[]>('/api/reolink/discover')
    } catch {
      toasts.error(translate('reolink.fetchError'))
    } finally {
      discovering = false
    }
  }
</script>

<div class="space-y-3">
  <!-- Columns -->
  <div class="flex items-center gap-2">
    <span class="text-sm text-[var(--color-text-secondary)]">{$t('reolink.columns')}</span>
    {#each [1, 2, 3, 4] as n (n)}
      <button
        onclick={() => { columns = n; emitChange() }}
        class="rounded-lg px-3 py-1 text-sm transition-colors"
        class:bg-[var(--color-primary)]={columns === n}
        class:text-white={columns === n}
        class:bg-[var(--color-bg-tertiary)]={columns !== n}
        class:text-[var(--color-text-secondary)]={columns !== n}
      >{n}</button>
    {/each}
  </div>

  <!-- Camera list -->
  <div>
    <div class="mb-2 flex items-center justify-between">
      <span class="text-sm font-medium text-[var(--color-text)]">{$t('reolink.cameras')}</span>
      {#if !$isDocker}
        <button
          onclick={discover}
          disabled={discovering}
          class="flex items-center gap-1 rounded-lg bg-[var(--color-bg-tertiary)] px-2 py-1 text-xs text-[var(--color-text-secondary)] hover:text-[var(--color-primary)]"
        >
          {#if discovering}
            <Loader2 size={12} class="animate-spin" />
          {:else}
            <Search size={12} />
          {/if}
          {$t('reolink.discover')}
        </button>
      {/if}
    </div>

    <!-- Discovered cameras -->
    {#if discovered.length > 0}
      <div class="mb-2 space-y-1">
        <span class="text-xs text-[var(--color-text-muted)]">{$t('reolink.discovered')}</span>
        {#each discovered as cam (cam.ip)}
          <div class="flex items-center justify-between rounded-lg border border-dashed border-[var(--color-border)] px-3 py-2">
            <div>
              <span class="text-sm text-[var(--color-text)]">{cam.name}</span>
              <span class="ms-2 text-xs text-[var(--color-text-muted)]">{cam.ip}</span>
              {#if cam.model}
                <span class="ms-1 text-xs text-[var(--color-text-muted)]">({cam.model})</span>
              {/if}
            </div>
            <button
              onclick={() => addDiscoveredCamera(cam)}
              class="rounded bg-[var(--color-primary)] px-2 py-1 text-xs text-white hover:bg-[var(--color-primary-hover)]"
            >
              <Plus size={12} />
            </button>
          </div>
        {/each}
      </div>
    {/if}

    <!-- Configured cameras -->
    {#if cameras.length === 0}
      <p class="text-center text-xs text-[var(--color-text-muted)]">{$t('reolink.noCameras')}</p>
    {:else}
      <div class="space-y-1">
        {#each cameras as cam, i (i)}
          {#if editingIndex === i}
            <div class="space-y-2 rounded-lg border border-[var(--color-primary)] p-3">
              <div class="grid grid-cols-2 gap-2">
                <input type="text" bind:value={editName} placeholder={$t('reolink.cameraName')}
                  class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1.5 text-sm text-[var(--color-text)]" />
                <input type="text" bind:value={editIP} placeholder={$t('reolink.cameraIP')}
                  class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1.5 text-sm text-[var(--color-text)]" />
                <input type="text" bind:value={editUsername} placeholder={$t('reolink.username')}
                  class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1.5 text-sm text-[var(--color-text)]" />
                <input type="password" bind:value={editPassword} placeholder={$t('reolink.password')}
                  class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1.5 text-sm text-[var(--color-text)]" />
              </div>
              <div class="flex gap-2">
                <button onclick={saveEdit}
                  class="flex items-center gap-1 rounded-lg bg-[var(--color-primary)] px-2 py-1 text-xs text-white hover:bg-[var(--color-primary-hover)]">
                  <Check size={12} /> {$t('common.save')}
                </button>
                <button onclick={cancelEdit}
                  class="flex items-center gap-1 rounded-lg bg-[var(--color-bg-tertiary)] px-2 py-1 text-xs text-[var(--color-text-secondary)]">
                  <X size={12} /> {$t('common.cancel')}
                </button>
              </div>
            </div>
          {:else}
            <div
              draggable="true"
              role="listitem"
              ondragstart={() => handleDragStart(i)}
              ondragover={(e: DragEvent) => handleDragOver(e, i)}
              ondrop={() => handleDrop(i)}
              ondragend={handleDragEnd}
              class="flex items-center gap-2 rounded-lg border px-3 py-2 transition-colors select-none"
              class:border-[var(--color-primary)]={dragOverIndex === i && dragIndex !== i}
              class:border-[var(--color-border)]={dragOverIndex !== i || dragIndex === i}
              class:opacity-50={dragIndex === i}
            >
              <span class="cursor-grab text-[var(--color-text-muted)]"><GripVertical size={14} /></span>
              <div class="flex-1">
                <span class="text-sm font-medium text-[var(--color-text)]">{cam.name}</span>
                <span class="ms-2 text-xs text-[var(--color-text-muted)]">{cam.ip}</span>
              </div>
              <div class="flex items-center gap-2">
                <button
                  onclick={() => startEdit(i)}
                  class="text-[var(--color-text-muted)] hover:text-[var(--color-primary)]"
                >
                  <Pencil size={14} />
                </button>
                <button
                  onclick={() => removeCamera(i)}
                  class="text-[var(--color-text-muted)] hover:text-[var(--color-danger)]"
                >
                  <Trash2 size={14} />
                </button>
              </div>
            </div>
          {/if}
        {/each}
      </div>
    {/if}
  </div>

  <!-- Add camera form -->
  <div class="space-y-2 rounded-lg border border-[var(--color-border)] p-3">
    <span class="text-xs font-medium text-[var(--color-text-secondary)]">{$t('reolink.addCamera')}</span>
    <div class="grid grid-cols-2 gap-2">
      <input type="text" bind:value={newName} placeholder={$t('reolink.cameraName')}
        class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1.5 text-sm text-[var(--color-text)]" />
      <input type="text" bind:value={newIP} placeholder={$t('reolink.cameraIP')}
        class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1.5 text-sm text-[var(--color-text)]" />
      <input type="text" bind:value={newUsername} placeholder={$t('reolink.username')}
        class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1.5 text-sm text-[var(--color-text)]" />
      <input type="password" bind:value={newPassword} placeholder={$t('reolink.password')}
        class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1.5 text-sm text-[var(--color-text)]" />
    </div>
    <button onclick={addCamera}
      class="flex items-center gap-1 rounded-lg bg-[var(--color-primary)] px-3 py-1.5 text-xs text-white hover:bg-[var(--color-primary-hover)]">
      <Plus size={12} /> {$t('reolink.addCamera')}
    </button>
  </div>
</div>
