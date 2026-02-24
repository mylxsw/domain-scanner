import { useMemo } from 'react'
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Cell,
} from 'recharts'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { ProbeItem } from '@/types'

interface PriceChartProps {
  results: ProbeItem[]
}

interface ChartData {
  tld: string
  count: number
  minPrice: number
  avgPrice: number
  maxPrice: number
}

export default function PriceChart({ results }: PriceChartProps) {
  const chartData = useMemo(() => {
    const tldStats = new Map<string, {
      prices: number[]
    }>()

    results.forEach(result => {
      if (result.available !== true || !result.total_price) return

      const stats = tldStats.get(result.tld) || { prices: [] }
      stats.prices.push(result.total_price)
      tldStats.set(result.tld, stats)
    })

    const data: ChartData[] = Array.from(tldStats.entries())
      .map(([tld, stats]) => ({
        tld: `.${tld}`,
        count: stats.prices.length,
        minPrice: Math.min(...stats.prices),
        avgPrice: stats.prices.reduce((a, b) => a + b, 0) / stats.prices.length,
        maxPrice: Math.max(...stats.prices),
      }))
      .sort((a, b) => b.count - a.count)
      .slice(0, 10)

    return data
  }, [results])

  const priceDistribution = useMemo(() => {
    const ranges = [
      { label: '$0-10', min: 0, max: 10, count: 0 },
      { label: '$10-20', min: 10, max: 20, count: 0 },
      { label: '$20-50', min: 20, max: 50, count: 0 },
      { label: '$50-100', min: 50, max: 100, count: 0 },
      { label: '$100+', min: 100, max: Infinity, count: 0 },
    ]

    results.forEach(result => {
      if (result.available !== true || !result.total_price) return
      const range = ranges.find(r => result.total_price! >= r.min && result.total_price! < r.max)
      if (range) range.count++
    })

    return ranges
  }, [results])

  const availableDomains = results.filter(r => r.available === true && r.total_price !== null)
  const avgPrice = availableDomains.length > 0
    ? availableDomains.reduce((sum, r) => sum + (r.total_price || 0), 0) / availableDomains.length
    : 0

  const colors = ['#22c55e', '#3b82f6', '#f59e0b', '#ef4444', '#8b5cf6']

  return (
    <div className="grid gap-4 md:grid-cols-2">
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">可注册域名 - TLD 分布</CardTitle>
        </CardHeader>
        <CardContent>
          {chartData.length > 0 ? (
            <ResponsiveContainer width="100%" height={250}>
              <BarChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis
                  dataKey="tld"
                  tick={{ fontSize: 12 }}
                  interval={0}
                />
                <YAxis tick={{ fontSize: 12 }} />
                <Tooltip
                  formatter={(value: number) => [`${value} 个域名`, '数量']}
                  labelFormatter={(label) => `TLD: ${label}`}
                />
                <Bar dataKey="count" fill="#3b82f6" radius={[4, 4, 0, 0]}>
                  {chartData.map((_, index) => (
                    <Cell key={`cell-${index}`} fill={colors[index % colors.length]} />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          ) : (
            <div className="h-[250px] flex items-center justify-center text-muted-foreground">
              暂无可注册的域名数据
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">价格分布</CardTitle>
        </CardHeader>
        <CardContent>
          {priceDistribution.some(p => p.count > 0) ? (
            <ResponsiveContainer width="100%" height={250}>
              <BarChart data={priceDistribution}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis
                  dataKey="label"
                  tick={{ fontSize: 12 }}
                />
                <YAxis tick={{ fontSize: 12 }} />
                <Tooltip
                  formatter={(value: number) => [`${value} 个域名`, '数量']}
                  labelFormatter={(label) => `价格区间: ${label}`}
                />
                <Bar dataKey="count" fill="#22c55e" radius={[4, 4, 0, 0]}>
                  {priceDistribution.map((_, index) => (
                    <Cell key={`cell-${index}`} fill={colors[index % colors.length]} />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          ) : (
            <div className="h-[250px] flex items-center justify-center text-muted-foreground">
              暂无可注册的域名数据
            </div>
          )}
        </CardContent>
      </Card>

      <Card className="md:col-span-2">
        <CardHeader>
          <CardTitle className="text-lg">价格统计</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
            <div className="space-y-1">
              <div className="text-sm text-muted-foreground">可注册域名</div>
              <div className="text-2xl font-bold text-green-600">
                {availableDomains.length}
              </div>
            </div>
            <div className="space-y-1">
              <div className="text-sm text-muted-foreground">平均价格</div>
              <div className="text-2xl font-bold">
                ${avgPrice.toFixed(2)}
              </div>
            </div>
            <div className="space-y-1">
              <div className="text-sm text-muted-foreground">最低价格</div>
              <div className="text-2xl font-bold text-green-600">
                ${availableDomains.length > 0
                  ? Math.min(...availableDomains.map(r => r.total_price || 0)).toFixed(2)
                  : '0.00'}
              </div>
            </div>
            <div className="space-y-1">
              <div className="text-sm text-muted-foreground">最高价格</div>
              <div className="text-2xl font-bold text-red-600">
                ${availableDomains.length > 0
                  ? Math.max(...availableDomains.map(r => r.total_price || 0)).toFixed(2)
                  : '0.00'}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
