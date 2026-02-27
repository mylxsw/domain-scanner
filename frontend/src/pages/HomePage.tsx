import DomainSearchForm from "@/components/DomainSearchForm";
import { Link } from "react-router-dom";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useProbeHistory } from "@/hooks/use-probe";

export default function HomePage() {
  const { data: historyData, isLoading: historyLoading } = useProbeHistory(
    20,
    0,
  );
  const history = historyData?.items || [];

  const formatTime = (value?: string) => {
    if (!value) return "-";
    const d = new Date(value);
    if (Number.isNaN(d.getTime())) return value;
    return d.toLocaleString("zh-CN", { hour12: false });
  };

  const statusBadge = (status: string) => {
    switch (status) {
      case "completed":
        return <Badge className="bg-green-500">已完成</Badge>;
      case "running":
        return <Badge className="bg-blue-500">进行中</Badge>;
      case "failed":
        return <Badge variant="destructive">失败</Badge>;
      default:
        return <Badge variant="secondary">{status}</Badge>;
    }
  };

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

      <Card>
        <CardHeader>
          <CardTitle>历史查询记录</CardTitle>
          <CardDescription>可点击任务跳转查看历史探测结果</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>关键词</TableHead>
                  <TableHead>模式</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>进度</TableHead>
                  <TableHead>创建时间</TableHead>
                  <TableHead>操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {historyLoading ? (
                  <TableRow>
                    <TableCell
                      colSpan={6}
                      className="text-center text-muted-foreground"
                    >
                      加载中...
                    </TableCell>
                  </TableRow>
                ) : history.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={6}
                      className="text-center text-muted-foreground"
                    >
                      暂无历史记录
                    </TableCell>
                  </TableRow>
                ) : (
                  history.map((item) => (
                    <TableRow key={item.id}>
                      <TableCell className="font-medium">{item.word}</TableCell>
                      <TableCell>{item.tld_mode}</TableCell>
                      <TableCell>{statusBadge(item.status)}</TableCell>
                      <TableCell>
                        {item.completed}/{item.total}
                      </TableCell>
                      <TableCell>{formatTime(item.created_at)}</TableCell>
                      <TableCell>
                        <Link
                          to={`/results/${item.id}`}
                          className="text-primary hover:underline"
                        >
                          查看
                        </Link>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
