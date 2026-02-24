import DomainSearchForm from '@/components/DomainSearchForm'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Globe, Zap, BarChart3, Shield } from 'lucide-react'

export default function HomePage() {
  return (
    <div className="space-y-8">
      <div className="text-center space-y-4 py-8">
        <h1 className="text-4xl font-bold tracking-tight">
          域名可用性探测工具
        </h1>
        <p className="text-lg text-muted-foreground max-w-2xl mx-auto">
          快速检测多个 TLD 后缀的域名注册情况，实时获取价格和可用性信息
        </p>
      </div>

      <DomainSearchForm />

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              支持 TLD
            </CardTitle>
            <Globe className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">30+</div>
            <p className="text-xs text-muted-foreground">
              覆盖主流域名后缀
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              实时检测
            </CardTitle>
            <Zap className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">SSE</div>
            <p className="text-xs text-muted-foreground">
              Server-Sent Events 推送
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              数据分析
            </CardTitle>
            <BarChart3 className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">图表</div>
            <p className="text-xs text-muted-foreground">
              价格分布可视化
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              可靠准确
            </CardTitle>
            <Shield className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">99%</div>
            <p className="text-xs text-muted-foreground">
              检测结果准确率
            </p>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>使用说明</CardTitle>
            <CardDescription>
              如何高效使用域名探测工具
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">
            <p>1. 输入您想要注册的域名关键词（不含后缀）</p>
            <p>2. 选择需要检测的 TLD 后缀，支持批量选择</p>
            <p>3. 点击"开始探测"按钮启动检测任务</p>
            <p>4. 实时查看检测结果和价格信息</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>支持的 TLD 分类</CardTitle>
            <CardDescription>
              按用途和价格分类的域名后缀
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">
            <p><strong>热门后缀：</strong>.com .net .org .io .co .app .dev .ai</p>
            <p><strong>低价后缀：</strong>.xyz .top .club .site .online .store</p>
            <p><strong>国家/地区：</strong>.cn .uk .de .jp .eu .us</p>
            <p><strong>其他后缀：</strong>.info .biz .me .tv .cc .ws</p>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
