import { CheckCircle2, Loader2, XCircle, Globe } from 'lucide-react'
import { Progress } from '@/components/ui/progress'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import type { ProbeTask } from '@/types'

interface ProbeProgressProps {
  task: ProbeTask
  className?: string
  phaseMessage?: string
  currentDomain?: string
}

export default function ProbeProgress({ task, className, phaseMessage, currentDomain }: ProbeProgressProps) {
  const percentage = task.total > 0
    ? Math.round((task.completed / task.total) * 100)
    : 0

  const getStatusIcon = () => {
    switch (task.status) {
      case 'completed':
        return <CheckCircle2 className="h-5 w-5 text-green-500" />
      case 'failed':
        return <XCircle className="h-5 w-5 text-red-500" />
      case 'running':
        return <Loader2 className="h-5 w-5 animate-spin text-blue-500" />
      default:
        return <Globe className="h-5 w-5 text-muted-foreground" />
    }
  }

  const getStatusBadge = () => {
    switch (task.status) {
      case 'completed':
        return <Badge variant="default" className="bg-green-500">已完成</Badge>
      case 'failed':
        return <Badge variant="destructive">失败</Badge>
      case 'running':
        return <Badge variant="default" className="bg-blue-500">探测中</Badge>
      default:
        return <Badge variant="secondary">等待中</Badge>
    }
  }

  return (
    <Card className={className}>
      <CardContent className="pt-6">
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              {getStatusIcon()}
              <div>
                <div className="font-medium">探测进度</div>
                <div className="text-sm text-muted-foreground">
                  {task.status === 'running' && (
                    phaseMessage || (currentDomain ? `正在探测 ${currentDomain}` : '正在探测域名可用性...')
                  )}
                  {task.status === 'completed' && '所有域名检查完成'}
                  {task.status === 'failed' && (task.error || '探测过程中出现错误')}
                  {task.status === 'pending' && (phaseMessage || '等待开始探测')}
                </div>
                {currentDomain && task.status === 'running' && phaseMessage && (
                  <div className="text-xs text-muted-foreground mt-1">
                    当前: <span className="font-mono text-foreground">{currentDomain}</span>
                  </div>
                )}
              </div>
            </div>
            {getStatusBadge()}
          </div>

          <div className="space-y-2">
            <Progress value={percentage} className="h-2" />
            <div className="flex justify-between text-sm text-muted-foreground">
              <span>{task.completed} / {task.total} 完成</span>
              <span>{percentage}%</span>
            </div>
          </div>

          {task.elapsed_seconds !== undefined && task.elapsed_seconds > 0 && (
            <div className="text-xs text-muted-foreground">
              耗时: {task.elapsed_seconds.toFixed(2)} 秒
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
