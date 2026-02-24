import { useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, RefreshCw, Download } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Separator } from '@/components/ui/separator'
import { Badge } from '@/components/ui/badge'
import ProbeProgress from '@/components/ProbeProgress'
import ResultsTable from '@/components/ResultsTable'
import PriceChart from '@/components/PriceChart'
import { useProbeTask, useProbeResults, useUpdateProbeResults } from '@/hooks/use-probe'
import { useSSE } from '@/hooks/use-sse'
import type { ProbeItem, ProgressEvent, SummaryEvent } from '@/types'

export default function ResultsPage() {
  const { probeId } = useParams<{ probeId: string }>()
  const [activeTab, setActiveTab] = useState('table')

  const { data: task, isLoading: isTaskLoading, error: taskError } = useProbeTask(probeId)
  const { data: resultsData, isLoading: isResultsLoading } = useProbeResults(probeId)
  const { addResult, updateTask } = useUpdateProbeResults(probeId)

  const results = resultsData?.results || []

  const { isConnected } = useSSE(
    probeId ? `/api/probe/${probeId}/stream` : null,
    {
      onMessage: (data: unknown) => {
        const event = data as ProgressEvent | SummaryEvent | { type: string }

        if ((event as ProgressEvent).type === 'progress') {
          const progressEvent = event as ProgressEvent
          updateTask({
            completed: progressEvent.index,
            total: progressEvent.total,
          })
          // Add to results if not already present
          addResult(progressEvent.data)
        } else if ((event as SummaryEvent).type === 'summary') {
          const summaryEvent = event as SummaryEvent
          updateTask({
            status: 'completed',
            total: summaryEvent.total,
            elapsed_seconds: summaryEvent.elapsed_seconds,
          })
        }
      },
      onError: (error) => {
        console.error('SSE error:', error)
      },
    }
  )

  const handleExport = () => {
    if (!results.length) return

    const csv = [
      ['Domain', 'TLD', 'Available', 'Is Premium', 'Currency', 'Base Price', 'Premium Price', 'ICANN Fee', 'EAP Fee', 'Total Price', 'Error'].join(','),
      ...results.map(r => [
        r.domain,
        r.tld,
        r.available === null ? 'Unknown' : (r.available ? 'Yes' : 'No'),
        r.is_premium === null ? '' : (r.is_premium ? 'Yes' : 'No'),
        r.currency || '',
        r.base_register_price || '',
        r.premium_register_price || '',
        r.icann_fee || '',
        r.eap_fee || '',
        r.total_price || '',
        r.error || '',
      ].join(','))
    ].join('\n')

    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement('a')
    link.href = URL.createObjectURL(blob)
    link.download = `probe-results-${probeId}.csv`
    link.click()
  }

  if (isTaskLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center space-y-4">
          <RefreshCw className="h-8 w-8 animate-spin mx-auto text-muted-foreground" />
          <p className="text-muted-foreground">加载中...</p>
        </div>
      </div>
    )
  }

  if (taskError || !task) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center space-y-4">
          <p className="text-red-500">加载失败，请检查探测 ID 是否正确</p>
          <Button asChild variant="outline">
            <Link to="/">
              <ArrowLeft className="mr-2 h-4 w-4" />
              返回首页
            </Link>
          </Button>
        </div>
      </div>
    )
  }

  const availableCount = results.filter(r => r.available === true).length
  const unavailableCount = results.filter(r => r.available === false && !r.error).length
  const errorCount = results.filter(r => r.error).length

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'pending':
        return <Badge variant="secondary">待处理</Badge>
      case 'running':
        return <Badge variant="default" className="bg-blue-500">进行中</Badge>
      case 'completed':
        return <Badge variant="default" className="bg-green-500">已完成</Badge>
      case 'failed':
        return <Badge variant="destructive">失败</Badge>
      default:
        return <Badge variant="outline">{status}</Badge>
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Button variant="ghost" size="sm" asChild className="-ml-2">
              <Link to="/">
                <ArrowLeft className="h-4 w-4 mr-1" />
                返回
              </Link>
            </Button>
            <h1 className="text-2xl font-bold">探测结果</h1>
            {getStatusBadge(task.status)}
          </div>
          <p className="text-sm text-muted-foreground">
            关键词: <span className="font-medium text-foreground">{task.word}</span>
            {' · '}
            TLD 模式: <span className="font-medium text-foreground">{task.tld_mode}</span>
            {' · '}
            {isConnected ? (
              <span className="text-green-600">实时连接中</span>
            ) : (
              <span className="text-muted-foreground">未连接</span>
            )}
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleExport}
            disabled={results.length === 0}
          >
            <Download className="mr-2 h-4 w-4" />
            导出 CSV
          </Button>
        </div>
      </div>

      <ProbeProgress task={task} />

      <div className="grid gap-4 sm:grid-cols-4">
        <div className="bg-card border rounded-lg p-4">
          <div className="text-sm text-muted-foreground">总检测数</div>
          <div className="text-2xl font-bold">{task.total || results.length}</div>
        </div>
        <div className="bg-card border rounded-lg p-4">
          <div className="text-sm text-muted-foreground">可注册</div>
          <div className="text-2xl font-bold text-green-600">{availableCount}</div>
        </div>
        <div className="bg-card border rounded-lg p-4">
          <div className="text-sm text-muted-foreground">已注册</div>
          <div className="text-2xl font-bold text-red-600">{unavailableCount}</div>
        </div>
        <div className="bg-card border rounded-lg p-4">
          <div className="text-sm text-muted-foreground">错误</div>
          <div className="text-2xl font-bold text-orange-600">{errorCount}</div>
        </div>
      </div>

      <Separator />

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="table">结果列表</TabsTrigger>
          <TabsTrigger value="charts">数据分析</TabsTrigger>
        </TabsList>

        <TabsContent value="table" className="mt-6">
          <ResultsTable results={results} />
        </TabsContent>

        <TabsContent value="charts" className="mt-6">
          <PriceChart results={results} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
