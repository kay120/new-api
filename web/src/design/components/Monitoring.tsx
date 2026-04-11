import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Input } from "./ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./ui/select";
import { Badge } from "./ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "./ui/table";
import { Search, Filter, CheckCircle2, XCircle, AlertCircle, Clock } from "lucide-react";

interface RequestLog {
  id: string;
  timestamp: string;
  model: string;
  method: string;
  endpoint: string;
  status: "success" | "error" | "timeout";
  statusCode: number;
  latency: number;
  tokens: number;
  cost: string;
  userId: string;
}

const mockLogs: RequestLog[] = [
  {
    id: "1",
    timestamp: "2026-04-07 14:23:45",
    model: "GPT-4",
    method: "POST",
    endpoint: "/v1/chat/completions",
    status: "success",
    statusCode: 200,
    latency: 1245,
    tokens: 1580,
    cost: "¥0.237",
    userId: "user_a1b2c3",
  },
  {
    id: "2",
    timestamp: "2026-04-07 14:23:42",
    model: "GPT-3.5",
    method: "POST",
    endpoint: "/v1/chat/completions",
    status: "success",
    statusCode: 200,
    latency: 680,
    tokens: 856,
    cost: "¥0.009",
    userId: "user_d4e5f6",
  },
  {
    id: "3",
    timestamp: "2026-04-07 14:23:38",
    model: "Claude-3",
    method: "POST",
    endpoint: "/v1/chat/completions",
    status: "error",
    statusCode: 429,
    latency: 120,
    tokens: 0,
    cost: "¥0.000",
    userId: "user_g7h8i9",
  },
  {
    id: "4",
    timestamp: "2026-04-07 14:23:35",
    model: "GPT-4",
    method: "POST",
    endpoint: "/v1/chat/completions",
    status: "success",
    statusCode: 200,
    latency: 1567,
    tokens: 2341,
    cost: "¥0.351",
    userId: "user_j0k1l2",
  },
  {
    id: "5",
    timestamp: "2026-04-07 14:23:30",
    model: "通义千问",
    method: "POST",
    endpoint: "/v1/chat/completions",
    status: "timeout",
    statusCode: 504,
    latency: 30000,
    tokens: 0,
    cost: "¥0.000",
    userId: "user_m3n4o5",
  },
  {
    id: "6",
    timestamp: "2026-04-07 14:23:28",
    model: "GPT-3.5",
    method: "POST",
    endpoint: "/v1/chat/completions",
    status: "success",
    statusCode: 200,
    latency: 720,
    tokens: 1024,
    cost: "¥0.010",
    userId: "user_p6q7r8",
  },
  {
    id: "7",
    timestamp: "2026-04-07 14:23:25",
    model: "Claude-3",
    method: "POST",
    endpoint: "/v1/chat/completions",
    status: "success",
    statusCode: 200,
    latency: 980,
    tokens: 1456,
    cost: "¥0.175",
    userId: "user_s9t0u1",
  },
  {
    id: "8",
    timestamp: "2026-04-07 14:23:20",
    model: "GPT-4",
    method: "POST",
    endpoint: "/v1/embeddings",
    status: "success",
    statusCode: 200,
    latency: 450,
    tokens: 512,
    cost: "¥0.077",
    userId: "user_v2w3x4",
  },
];

export function Monitoring() {
  const [logs] = useState<RequestLog[]>(mockLogs);
  const [searchTerm, setSearchTerm] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");

  const getStatusIcon = (status: RequestLog["status"]) => {
    switch (status) {
      case "success":
        return <CheckCircle2 className="w-4 h-4 text-green-600" />;
      case "error":
        return <XCircle className="w-4 h-4 text-red-600" />;
      case "timeout":
        return <Clock className="w-4 h-4 text-orange-600" />;
    }
  };

  const getStatusBadge = (status: RequestLog["status"]) => {
    switch (status) {
      case "success":
        return <Badge className="bg-green-100 text-green-700">成功</Badge>;
      case "error":
        return <Badge className="bg-red-100 text-red-700">错误</Badge>;
      case "timeout":
        return <Badge className="bg-orange-100 text-orange-700">超时</Badge>;
    }
  };

  const getLatencyColor = (latency: number) => {
    if (latency < 1000) return "text-green-600";
    if (latency < 3000) return "text-orange-600";
    return "text-red-600";
  };

  const filteredLogs = logs.filter(log => {
    const matchesSearch = 
      log.model.toLowerCase().includes(searchTerm.toLowerCase()) ||
      log.userId.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesStatus = statusFilter === "all" || log.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  const successCount = logs.filter(l => l.status === "success").length;
  const errorCount = logs.filter(l => l.status === "error").length;
  const timeoutCount = logs.filter(l => l.status === "timeout").length;
  const avgLatency = Math.round(
    logs.filter(l => l.status === "success").reduce((sum, l) => sum + l.latency, 0) / successCount
  );

  return (
    <div className="p-6 space-y-6">
      {/* 标题 */}
      <div>
        <h1 className="text-2xl font-semibold text-gray-900">请求监控</h1>
        <p className="text-sm text-gray-500 mt-1">实时查看API请求日志和状态</p>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center">
                <CheckCircle2 className="w-5 h-5 text-green-600" />
              </div>
              <div>
                <div className="text-sm text-gray-600">成功请求</div>
                <div className="text-xl font-semibold text-gray-900">{successCount}</div>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-red-100 rounded-lg flex items-center justify-center">
                <XCircle className="w-5 h-5 text-red-600" />
              </div>
              <div>
                <div className="text-sm text-gray-600">失败请求</div>
                <div className="text-xl font-semibold text-gray-900">{errorCount}</div>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-orange-100 rounded-lg flex items-center justify-center">
                <Clock className="w-5 h-5 text-orange-600" />
              </div>
              <div>
                <div className="text-sm text-gray-600">超时请求</div>
                <div className="text-xl font-semibold text-gray-900">{timeoutCount}</div>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center">
                <AlertCircle className="w-5 h-5 text-blue-600" />
              </div>
              <div>
                <div className="text-sm text-gray-600">平均延迟</div>
                <div className="text-xl font-semibold text-gray-900">{avgLatency}ms</div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* 过滤和搜索 */}
      <Card>
        <CardHeader>
          <div className="flex flex-col sm:flex-row gap-4">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
              <Input
                placeholder="搜索模型或用户ID..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-9"
              />
            </div>
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-full sm:w-[180px]">
                <Filter className="w-4 h-4 mr-2" />
                <SelectValue placeholder="状态筛选" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部状态</SelectItem>
                <SelectItem value="success">成功</SelectItem>
                <SelectItem value="error">错误</SelectItem>
                <SelectItem value="timeout">超时</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>时间</TableHead>
                  <TableHead>模型</TableHead>
                  <TableHead>接口</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>状态码</TableHead>
                  <TableHead>延迟</TableHead>
                  <TableHead>Token</TableHead>
                  <TableHead>成本</TableHead>
                  <TableHead>用户ID</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredLogs.map((log) => (
                  <TableRow key={log.id}>
                    <TableCell className="text-sm text-gray-600">
                      {log.timestamp}
                    </TableCell>
                    <TableCell>
                      <span className="font-medium text-gray-900">{log.model}</span>
                    </TableCell>
                    <TableCell>
                      <code className="text-xs bg-gray-100 px-2 py-1 rounded">
                        {log.method} {log.endpoint}
                      </code>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        {getStatusIcon(log.status)}
                        {getStatusBadge(log.status)}
                      </div>
                    </TableCell>
                    <TableCell>
                      <span className={`font-medium ${
                        log.statusCode === 200 ? "text-green-600" : "text-red-600"
                      }`}>
                        {log.statusCode}
                      </span>
                    </TableCell>
                    <TableCell>
                      <span className={`font-medium ${getLatencyColor(log.latency)}`}>
                        {log.latency}ms
                      </span>
                    </TableCell>
                    <TableCell className="text-gray-900">
                      {log.tokens.toLocaleString()}
                    </TableCell>
                    <TableCell className="font-medium text-gray-900">
                      {log.cost}
                    </TableCell>
                    <TableCell>
                      <code className="text-xs text-gray-600">{log.userId}</code>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
