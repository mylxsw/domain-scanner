import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { useToast } from './use-toast'
import type { ProbeTask, ProbeItem, TldMode } from '@/types'

interface CreateProbeRequest {
  word: string
  tld_mode?: TldMode
  rate_per_min?: number
  max_batch?: number
  cache_ttl_hours?: number
}

async function fetchApi<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'API request failed' }))
    throw new Error(error.message || 'API request failed')
  }

  return response.json()
}

export function useStartProbe() {
  const navigate = useNavigate()
  const { toast } = useToast()

  return useMutation({
    mutationFn: async (data: CreateProbeRequest): Promise<ProbeTask> => {
      return fetchApi('/api/probe', {
        method: 'POST',
        body: JSON.stringify(data),
      })
    },
    onSuccess: (data) => {
      navigate(`/results/${data.id}`)
    },
    onError: (error: Error) => {
      toast({
        title: '启动探测失败',
        description: error.message,
        variant: 'destructive',
      })
    },
  })
}

export function useProbeTask(taskId: string | undefined) {
  return useQuery({
    queryKey: ['probe', taskId],
    queryFn: async (): Promise<ProbeTask> => {
      if (!taskId) throw new Error('Task ID is required')
      return fetchApi(`/api/probe/${taskId}`)
    },
    enabled: !!taskId,
    refetchInterval: (query) => {
      const data = query.state.data
      if (data?.status === 'completed' || data?.status === 'failed') {
        return false
      }
      return 2000
    },
  })
}

export function useProbeResults(taskId: string | undefined) {
  return useQuery({
    queryKey: ['probe', taskId, 'results'],
    queryFn: async (): Promise<{ task: ProbeTask; results: ProbeItem[] }> => {
      if (!taskId) throw new Error('Task ID is required')
      return fetchApi(`/api/probe/${taskId}/results`)
    },
    enabled: !!taskId,
  })
}

export function useUpdateProbeResults(taskId: string | undefined) {
  const queryClient = useQueryClient()

  return {
    addResult: (item: ProbeItem) => {
      if (!taskId) return
      queryClient.setQueryData(['probe', taskId, 'results'], (old: any) => {
        if (!old) return { task: null, results: [item] }
        const exists = old.results.find((r: ProbeItem) => r.domain === item.domain)
        if (exists) return old
        return { ...old, results: [...old.results, item] }
      })
    },
    updateTask: (task: Partial<ProbeTask>) => {
      if (!taskId) return
      queryClient.setQueryData(['probe', taskId], (old: ProbeTask | undefined) => {
        if (!old) return old
        return { ...old, ...task }
      })
    },
  }
}
