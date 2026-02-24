import { useEffect, useRef, useState, useCallback } from 'react'

interface SSEOptions {
  onMessage?: (data: unknown) => void
  onError?: (error: Event) => void
  onOpen?: () => void
  onClose?: () => void
}

export function useSSE(url: string | null, options: SSEOptions = {}) {
  const [isConnected, setIsConnected] = useState(false)
  const [lastEvent, setLastEvent] = useState<unknown>(null)
  const eventSourceRef = useRef<EventSource | null>(null)
  const optionsRef = useRef(options)

  useEffect(() => {
    optionsRef.current = options
  }, [options])

  const disconnect = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
      eventSourceRef.current = null
      setIsConnected(false)
      optionsRef.current.onClose?.()
    }
  }, [])

  useEffect(() => {
    if (!url) {
      disconnect()
      return
    }

    const eventSource = new EventSource(url)
    eventSourceRef.current = eventSource

    eventSource.onopen = () => {
      setIsConnected(true)
      optionsRef.current.onOpen?.()
    }

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        setLastEvent(data)
        optionsRef.current.onMessage?.(data)
      } catch (error) {
        console.error('Failed to parse SSE message:', error)
      }
    }

    eventSource.onerror = (error) => {
      console.error('SSE error:', error)
      setIsConnected(false)
      optionsRef.current.onError?.(error)
      eventSource.close()
    }

    return () => {
      disconnect()
    }
  }, [url, disconnect])

  return {
    isConnected,
    lastEvent,
    disconnect,
  }
}
