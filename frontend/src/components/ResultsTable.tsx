import { useState, useMemo } from 'react'
import { ArrowUpDown, ArrowUp, ArrowDown, CheckCircle2, XCircle, AlertCircle } from 'lucide-react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import type { ProbeItem } from '@/types'

type SortField = 'domain' | 'tld' | 'available' | 'price'
type SortDirection = 'asc' | 'desc'

interface ResultsTableProps {
  results: ProbeItem[]
}

export default function ResultsTable({ results }: ResultsTableProps) {
  const [sortField, setSortField] = useState<SortField>('available')
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc')
  const [filterText, setFilterText] = useState('')
  const [availabilityFilter, setAvailabilityFilter] = useState<'all' | 'available' | 'unavailable'>('all')

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortField(field)
      setSortDirection('asc')
    }
  }

  const getSortIcon = (field: SortField) => {
    if (sortField !== field) {
      return <ArrowUpDown className="h-4 w-4" />
    }
    return sortDirection === 'asc'
      ? <ArrowUp className="h-4 w-4" />
      : <ArrowDown className="h-4 w-4" />
  }

  const formatPrice = (price: number | null, currency: string | null) => {
    if (price === null) return '-'
    const symbol = currency === 'USD' ? '$' : (currency || '')
    return `${symbol}${price.toFixed(2)}`
  }

  const filteredAndSortedResults = useMemo(() => {
    let filtered = results

    if (filterText) {
      const lowerFilter = filterText.toLowerCase()
      filtered = filtered.filter(r =>
        r.domain.toLowerCase().includes(lowerFilter) ||
        r.tld.toLowerCase().includes(lowerFilter)
      )
    }

    if (availabilityFilter !== 'all') {
      filtered = filtered.filter(r =>
        availabilityFilter === 'available' ? r.available === true : r.available === false
      )
    }

    return [...filtered].sort((a, b) => {
      let comparison = 0

      switch (sortField) {
        case 'domain':
          comparison = a.domain.localeCompare(b.domain)
          break
        case 'tld':
          comparison = a.tld.localeCompare(b.tld)
          break
        case 'available':
          const aAvail = a.available === true ? 2 : (a.available === false ? 0 : 1)
          const bAvail = b.available === true ? 2 : (b.available === false ? 0 : 1)
          comparison = bAvail - aAvail
          break
        case 'price':
          const priceA = a.total_price ?? Infinity
          const priceB = b.total_price ?? Infinity
          comparison = priceA - priceB
          break
      }

      return sortDirection === 'asc' ? comparison : -comparison
    })
  }, [results, sortField, sortDirection, filterText, availabilityFilter])

  const availableCount = results.filter(r => r.available === true).length
  const unavailableCount = results.filter(r => r.available === false && !r.error).length
  const errorCount = results.filter(r => r.error).length

  return (
    <div className="space-y-4">
      <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
        <div className="flex gap-2 flex-wrap">
          <Button
            variant={availabilityFilter === 'all' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setAvailabilityFilter('all')}
          >
            全部 ({results.length})
          </Button>
          <Button
            variant={availabilityFilter === 'available' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setAvailabilityFilter('available')}
            className="gap-1"
          >
            <CheckCircle2 className="h-4 w-4 text-green-500" />
            可注册 ({availableCount})
          </Button>
          <Button
            variant={availabilityFilter === 'unavailable' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setAvailabilityFilter('unavailable')}
            className="gap-1"
          >
            <XCircle className="h-4 w-4 text-red-500" />
            已注册 ({unavailableCount})
          </Button>
        </div>

        <Input
          placeholder="搜索域名或后缀..."
          value={filterText}
          onChange={(e) => setFilterText(e.target.value)}
          className="w-full sm:w-64"
        />
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => handleSort('domain')}
                  className="gap-1"
                >
                  域名
                  {getSortIcon('domain')}
                </Button>
              </TableHead>
              <TableHead>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => handleSort('tld')}
                  className="gap-1"
                >
                  后缀
                  {getSortIcon('tld')}
                </Button>
              </TableHead>
              <TableHead>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => handleSort('available')}
                  className="gap-1"
                >
                  状态
                  {getSortIcon('available')}
                </Button>
              </TableHead>
              <TableHead>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => handleSort('price')}
                  className="gap-1"
                >
                  总价(1年)
                  {getSortIcon('price')}
                </Button>
              </TableHead>
              <TableHead className="hidden sm:table-cell">详情</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredAndSortedResults.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} className="text-center py-8 text-muted-foreground">
                  暂无数据
                </TableCell>
              </TableRow>
            ) : (
              filteredAndSortedResults.map((result) => (
                <TableRow key={result.domain}>
                  <TableCell className="font-medium">
                    {result.domain}
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline">.{result.tld}</Badge>
                  </TableCell>
                  <TableCell>
                    {result.error ? (
                      <div className="flex items-center gap-1 text-yellow-600" title={result.error}>
                        <AlertCircle className="h-4 w-4" />
                        <span className="text-sm">检查失败</span>
                      </div>
                    ) : result.available === true ? (
                      <div className="flex items-center gap-1 text-green-600">
                        <CheckCircle2 className="h-4 w-4" />
                        <span className="text-sm">可注册</span>
                      </div>
                    ) : result.available === false ? (
                      <div className="flex items-center gap-1 text-red-600">
                        <XCircle className="h-4 w-4" />
                        <span className="text-sm">已注册</span>
                      </div>
                    ) : (
                      <span className="text-sm text-muted-foreground">未知</span>
                    )}
                  </TableCell>
                  <TableCell>
                    {result.total_price !== null ? (
                      <span className="font-medium">
                        {formatPrice(result.total_price, result.currency)}
                      </span>
                    ) : (
                      <span className="text-muted-foreground">-</span>
                    )}
                  </TableCell>
                  <TableCell className="hidden sm:table-cell text-muted-foreground text-sm">
                    {result.is_premium === true && (
                      <Badge variant="secondary" className="mr-1">Premium</Badge>
                    )}
                    {result.base_register_price !== null && result.total_price !== null && (
                      <span className="text-xs">
                        基础: {formatPrice(result.base_register_price, result.currency)}
                      </span>
                    )}
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}
