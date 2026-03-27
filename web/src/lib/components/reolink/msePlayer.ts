/**
 * MSE (Media Source Extensions) player for go2rtc WebSocket streaming.
 * Handles WebSocket connection, MediaSource, SourceBuffer and buffering.
 */

export interface MsePlayerCallbacks {
  onActive: (active: boolean) => void
  onFallback: () => void
}

export class MsePlayer {
  private ws: WebSocket | null = null
  private mediaSource: MediaSource | null = null
  private sourceBuffer: SourceBuffer | null = null
  private bufferQueue: ArrayBuffer[] = []
  private objectUrl: string | null = null
  private destroyed = false
  private wsUrl: string
  private videoEl: HTMLVideoElement | undefined
  private callbacks: MsePlayerCallbacks

  constructor(wsUrl: string, callbacks: MsePlayerCallbacks) {
    this.wsUrl = wsUrl
    this.callbacks = callbacks
  }

  /** Begin MSE streaming on the given video element */
  start(videoEl: HTMLVideoElement | undefined) {
    if (this.destroyed || !videoEl) return
    this.videoEl = videoEl
    this.mediaSource = new MediaSource()
    this.objectUrl = URL.createObjectURL(this.mediaSource)
    this.callbacks.onActive(true)

    this.mediaSource.addEventListener(
      'sourceopen',
      () => this.connectWebSocket(),
      { once: true },
    )
    videoEl.src = this.objectUrl
  }

  /** Stop streaming, clean up all resources */
  stop() {
    this.callbacks.onActive(false)
    this.bufferQueue = []
    this.closeWebSocket()
    this.cleanupMediaSource()
    this.revokeObjectUrl()
    if (this.videoEl) this.videoEl.src = ''
    this.videoEl = undefined
  }

  /** Mark this player as permanently destroyed */
  destroy() {
    this.destroyed = true
    this.stop()
  }

  isActive(): boolean {
    return this.objectUrl !== null
  }

  // --- WebSocket ---

  private connectWebSocket() {
    if (this.destroyed) return

    this.ws = new WebSocket(this.wsUrl)
    this.ws.binaryType = 'arraybuffer'
    let codecReceived = false

    this.ws.onopen = () => {
      this.ws!.send(JSON.stringify({ type: 'mse', value: '' }))
    }

    this.ws.onmessage = (event) => {
      if (!codecReceived && typeof event.data === 'string') {
        codecReceived = true
        this.handleCodecMessage(event.data)
        return
      }
      if (event.data instanceof ArrayBuffer) {
        this.appendToBuffer(event.data)
      }
    }

    this.ws.onclose = () => {
      if (!this.destroyed && this.objectUrl) {
        setTimeout(() => {
          if (!this.destroyed && this.objectUrl) this.connectWebSocket()
        }, 2000)
      }
    }

    this.ws.onerror = () => {
      this.ws?.close()
    }
  }

  private handleCodecMessage(data: string) {
    try {
      const msg = JSON.parse(data)
      const mimeCodec = msg.value || 'video/mp4; codecs="avc1.640029"'

      if (!MediaSource.isTypeSupported(mimeCodec)) {
        console.warn('[MSE] Codec not supported:', mimeCodec)
        this.stop()
        this.callbacks.onFallback()
        return
      }

      this.sourceBuffer = this.mediaSource!.addSourceBuffer(mimeCodec)
      this.sourceBuffer.mode = 'segments'
      this.sourceBuffer.addEventListener('updateend', () => this.flushQueue())
    } catch (e) {
      console.error('[MSE] Failed to create SourceBuffer:', e)
      this.stop()
      this.callbacks.onFallback()
    }
  }

  // --- Buffer management ---

  private appendToBuffer(data: ArrayBuffer) {
    if (!this.sourceBuffer) return
    if (this.sourceBuffer.updating) {
      this.bufferQueue.push(data)
      return
    }
    try {
      this.sourceBuffer.appendBuffer(data)
    } catch {
      this.bufferQueue = []
    }
  }

  private flushQueue() {
    if (!this.sourceBuffer || this.sourceBuffer.updating || this.bufferQueue.length === 0) return

    if (this.shouldSeekToLiveEdge()) return

    const chunk = this.bufferQueue.shift()
    if (chunk) {
      try {
        this.sourceBuffer.appendBuffer(chunk)
      } catch {
        this.bufferQueue = []
      }
    }
  }

  /** Keep video near live edge, trim old buffer to avoid memory growth */
  private shouldSeekToLiveEdge(): boolean {
    if (!this.videoEl || !this.sourceBuffer || this.videoEl.buffered.length === 0) return false

    const bufferedEnd = this.videoEl.buffered.end(this.videoEl.buffered.length - 1)
    if (bufferedEnd - (this.videoEl.currentTime || 0) > 4) {
      this.videoEl.currentTime = bufferedEnd - 0.5
    }
    if (this.videoEl.buffered.start(0) < bufferedEnd - 10) {
      try {
        this.sourceBuffer.remove(0, bufferedEnd - 5)
        return true // wait for updateend after remove
      } catch { /* ignored */ }
    }
    return false
  }

  // --- Cleanup helpers ---

  private closeWebSocket() {
    if (this.ws) {
      this.ws.onclose = null
      this.ws.onerror = null
      this.ws.close()
      this.ws = null
    }
  }

  private cleanupMediaSource() {
    if (this.sourceBuffer && this.mediaSource && this.mediaSource.readyState === 'open') {
      try { this.mediaSource.removeSourceBuffer(this.sourceBuffer) } catch { /* ignored */ }
    }
    this.sourceBuffer = null
    if (this.mediaSource && this.mediaSource.readyState === 'open') {
      try { this.mediaSource.endOfStream() } catch { /* ignored */ }
    }
    this.mediaSource = null
  }

  private revokeObjectUrl() {
    if (this.objectUrl) {
      URL.revokeObjectURL(this.objectUrl)
      this.objectUrl = null
    }
  }
}

/** Build an absolute WebSocket URL from a relative or absolute go2rtc path */
export function buildWsUrl(go2rtcUrl: string): string {
  if (go2rtcUrl.startsWith('/')) {
    const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    return `${proto}//${window.location.host}${go2rtcUrl}`
  }
  return go2rtcUrl.replace(/^http/, 'ws')
}
