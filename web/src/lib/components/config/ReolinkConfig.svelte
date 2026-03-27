<script lang="ts">
  import { Trash2, Pencil, GripVertical } from 'lucide-svelte'
  import { t } from '../../i18n'
  import ResponsiveColumnsConfig from './ResponsiveColumnsConfig.svelte'
  import CameraForm from './CameraForm.svelte'
  import CameraDiscoverySection from './CameraDiscoverySection.svelte'

  interface CameraEntry {
    name: string
    ip: string
    username: string
    password: string
    channel: number
    source: string
    entity_id: string
  }

  interface Props {
    value?: string
    onchange?: (value: string) => void
  }

  const { value = '{}', onchange }: Props = $props()

  let cameras = $state<CameraEntry[]>([])
  let columns = $state(2)
  let columnsTablet = $state(0)
  let columnsMobile = $state(0)
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
      if (typeof parsed.columnsTablet === 'number') columnsTablet = parsed.columnsTablet
      if (typeof parsed.columnsMobile === 'number') columnsMobile = parsed.columnsMobile
    } catch { /* ignored */ }
  })

  function emitChange() {
    const config: Record<string, unknown> = { cameras }
    if (columns !== 2) config.columns = columns
    if (columnsTablet > 0) config.columnsTablet = columnsTablet
    if (columnsMobile > 0) config.columnsMobile = columnsMobile
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

  function addDiscoveredCamera(cam: { ip: string; name: string; model: string }) {
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

</script>

<div class="space-y-3">
  <!-- Columns -->
  <ResponsiveColumnsConfig
    {columns}
    columnsTablet={columnsTablet || Math.min(columns, 2)}
    columnsMobile={columnsMobile || 1}
    onchange={(d, t, m) => { columns = d; columnsTablet = t; columnsMobile = m; emitChange() }}
  />

  <!-- Camera list -->
  <div>
    <CameraDiscoverySection onAddCamera={addDiscoveredCamera} />

    <!-- Configured cameras -->
    {#if cameras.length === 0}
      <p class="text-center text-xs text-[var(--color-text-muted)]">{$t('reolink.noCameras')}</p>
    {:else}
      <div class="space-y-1">
        {#each cameras as cam, i (i)}
          {#if editingIndex === i}
            <CameraForm
              bind:name={editName}
              bind:ip={editIP}
              bind:username={editUsername}
              bind:password={editPassword}
              isEdit={true}
              onSave={saveEdit}
              onCancel={cancelEdit}
            />
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
  <CameraForm
    bind:name={newName}
    bind:ip={newIP}
    bind:username={newUsername}
    bind:password={newPassword}
    isEdit={false}
    onSave={addCamera}
    onCancel={() => {}}
  />
</div>
