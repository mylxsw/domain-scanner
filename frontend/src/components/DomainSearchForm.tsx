import { useState } from 'react'
import { Search, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useStartProbe } from '@/hooks/use-probe'
import type { TldMode } from '@/types'

const TLD_MODES: { value: TldMode; label: string; description: string }[] = [
  { value: 'all', label: '全部 TLD', description: '检测所有支持的域名后缀（约30+个）' },
  { value: 'mainstream-only', label: '热门后缀', description: '仅检测主流后缀如 .com .net .org .io 等' },
  { value: 'api-registerable-only', label: '低价后缀', description: '仅检测价格较低的域名后缀' },
]

export default function DomainSearchForm() {
  const [word, setWord] = useState('')
  const [tldMode, setTldMode] = useState<TldMode>('all')
  const startProbe = useStartProbe()

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!word.trim()) return

    startProbe.mutate({
      word: word.trim().toLowerCase(),
      tld_mode: tldMode,
    })
  }

  const isLoading = startProbe.isPending

  return (
    <Card className="w-full max-w-2xl mx-auto">
      <CardHeader>
        <CardTitle className="text-2xl">域名可用性探测</CardTitle>
        <CardDescription>
          输入关键词，选择 TLD 模式，快速检测域名是否可注册
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="space-y-2">
            <Label htmlFor="word">域名关键词</Label>
            <div className="relative">
              <Input
                id="word"
                placeholder="例如: mycompany, blog, shop..."
                value={word}
                onChange={(e) => setWord(e.target.value)}
                disabled={isLoading}
                className="pr-4"
              />
            </div>
            <p className="text-sm text-muted-foreground">
              输入您想要注册的域名前缀，支持字母、数字和连字符
            </p>
          </div>

          <div className="space-y-2">
            <Label>TLD 检测模式</Label>
            <Select value={tldMode} onValueChange={(v) => setTldMode(v as TldMode)} disabled={isLoading}>
              <SelectTrigger>
                <SelectValue placeholder="选择检测模式" />
              </SelectTrigger>
              <SelectContent>
                {TLD_MODES.map((mode) => (
                  <SelectItem key={mode.value} value={mode.value}>
                    <div className="flex flex-col items-start">
                      <span>{mode.label}</span>
                      <span className="text-xs text-muted-foreground">{mode.description}</span>
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="flex items-center justify-between pt-4">
            <div className="text-sm text-muted-foreground">
              检测模式: <span className="font-medium text-foreground">
                {TLD_MODES.find(m => m.value === tldMode)?.label}
              </span>
            </div>
            <Button
              type="submit"
              disabled={!word.trim() || isLoading}
              size="lg"
            >
              {isLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  启动探测中...
                </>
              ) : (
                <>
                  <Search className="mr-2 h-4 w-4" />
                  开始探测
                </>
              )}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}
