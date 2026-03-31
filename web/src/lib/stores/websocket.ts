import { writable, get } from 'svelte/store'

export type WSStatus = 'connecting' | 'connected' | 'disconnected'

type MessageHandler = (payload: unknown) => void

interface WSState {
  status: WSStatus
  reconnectAttempts: number
}

const MAX_RECONNECT_DELAY = 30000
const INITIAL_RECONNECT_DELAY = 1000

// eslint-disable-next-line max-lines-per-function
function createWebSocketStore() {
  const { subscribe, set, update } = writable<WSState>({
    status: 'disconnected',
    reconnectAttempts: 0,
  })

  let socket: WebSocket | null = null
  const handlers: Map<string, Set<MessageHandler>> = new Map()
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let intentionalClose = false

  function getWSUrl(): string {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    return `${protocol}//${window.location.host}/api/ws`
  }

  function attachSocketHandlers(ws: WebSocket) {
    ws.onopen = () => {
      set({ status: 'connected', reconnectAttempts: 0 })
    }

    ws.onclose = () => {
      update((s) => ({ ...s, status: 'disconnected' }))
      if (!intentionalClose) {
        scheduleReconnect()
      }
    }

    ws.onerror = () => {
      ws.close()
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data) as { type: string; payload?: unknown }
        const typeHandlers = handlers.get(msg.type)
        if (typeHandlers) {
          typeHandlers.forEach((handler) => handler(msg.payload))
        }

        const wildcardHandlers = handlers.get('*')
        if (wildcardHandlers) {
          wildcardHandlers.forEach((handler) => handler(msg))
        }
      } catch {
        // Ignore malformed messages
      }
    }
  }

  function connect() {
    if (socket?.readyState === WebSocket.OPEN || socket?.readyState === WebSocket.CONNECTING) {
      return
    }

    intentionalClose = false
    update((s) => ({ ...s, status: 'connecting' }))

    try {
      socket = new WebSocket(getWSUrl())
    } catch {
      scheduleReconnect()
      return
    }

    attachSocketHandlers(socket)
  }

  function disconnect() {
    intentionalClose = true
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (socket) {
      socket.close()
      socket = null
    }
    set({ status: 'disconnected', reconnectAttempts: 0 })
  }

  function scheduleReconnect() {
    if (intentionalClose) return

    const state = get({ subscribe })
    const delay = Math.min(
      INITIAL_RECONNECT_DELAY * Math.pow(2, state.reconnectAttempts),
      MAX_RECONNECT_DELAY,
    )

    update((s) => ({ ...s, reconnectAttempts: s.reconnectAttempts + 1 }))

    reconnectTimer = setTimeout(() => {
      reconnectTimer = null
      connect()
    }, delay)
  }

  function on(type: string, handler: MessageHandler): () => void {
    if (!handlers.has(type)) {
      handlers.set(type, new Set())
    }
    handlers.get(type)!.add(handler)

    return () => {
      const typeHandlers = handlers.get(type)
      if (typeHandlers) {
        typeHandlers.delete(handler)
        if (typeHandlers.size === 0) {
          handlers.delete(type)
        }
      }
    }
  }

  return {
    subscribe,
    connect,
    disconnect,
    on,
  }
}

export const ws = createWebSocketStore()
