import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { AreaChart, Area, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from "recharts";
import { DollarSign, TrendingUp, TrendingDown, Calendar } from "lucide-react";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "./ui/table";
import { Badge } from "./ui/badge";
import { useState } from "react";

const monthlyData = [
  { month: "10月", cost: 8500, requests: 1234500 },
  { month: "11月", cost: 11200, requests: 1567800 },
  { month: "12月", cost: 13400, requests: 1892300 },
  { month: "1月", cost: 12500, requests: 1756200 },
  { month: "2月", cost: 15200, requests: 2123400 },
  { month: "3月", cost: 18900, requests: 2654800 },
];

const modelCostData = [
  { model: "GPT-4", cost: 7200 },
  { model: "GPT-3.5", cost: 2800 },
  { model: "Claude-3", cost: 4300 },
  { model: "通义千问", cost: 1600 },
  { model: "其他", cost: 800 },
];

const dailyData = [
  { date: "4/1", cost: 520 },
  { date: "4/2", cost: 680 },
  { date: "4/3", cost: 590 },
  { date: "4/4", cost: 720 },
  { date: "4/5", cost: 650 },
  { date: "4/6", cost: 710 },
  { date: "4/7", cost: 830 },
];

interface Transaction {
  id: string;
  date: string;
  model: string;
  requests: number;
  tokens: number;
  cost: string;
  type: "api" | "embedding" | "fine-tune";
}

const recentTransactions: Transaction[] = [
  {
    id: "1",
    date: "2026-04-07",
    model: "GPT-4 Turbo",
    requests: 3456,
    tokens: 1234567,
    cost: "¥523.18",
    type: "api",
  },
  {
    id: "2",
    date: "2026-04-06",
    model: "GPT-3.5 Turbo",
    requests: 8923,
    tokens: 2345678,
    cost: "¥234.57",
    type: "api",
  },
  {
    id: "3",
    date: "2026-04-06",
    model: "Claude-3 Opus",
    requests: 1245,
    tokens: 987654,
    cost: "¥345.67",
    type: "api",
  },
  {
    id: "4",
    date: "2026-04-05",
    model: "GPT-4 Turbo",
    requests: 234,
    tokens: 156789,
    cost: "¥78.92",
    type: "embedding",
  },
  {
    id: "5",
    date: "2026-04-05",
    model: "通义千问 Max",
    requests: 5678,
    tokens: 1876543,
    cost: "¥112.58",
    type: "api",
  },
];

export function Billing() {
  const [period, setPeriod] = useState("month");

  const getTypeText = (type: Transaction["type"]) => {
    switch (type) {
      case "api": return "API调用";
      case "embedding": return "向量化";
      case "fine-tune": return "微调";
    }
  };

  return (
    <div className="p-6 space-y-6">
      {/* 标题 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">计费统计</h1>
          <p className="text-sm text-gray-500 mt-1">查看使用量和费用明细</p>
        </div>
        <Select value={period} onValueChange={setPeriod}>
          <SelectTrigger className="w-[180px]">
            <Calendar className="w-4 h-4 mr-2" />
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="week">本周</SelectItem>
            <SelectItem value="month">本月</SelectItem>
            <SelectItem value="quarter">本季度</SelectItem>
            <SelectItem value="year">本年</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* 费用概览 */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-gray-600">本月费用</CardTitle>
            <DollarSign className="w-4 h-4 text-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-semibold text-gray-900">¥16,700</div>
            <div className="flex items-center gap-1 mt-1 text-sm">
              <TrendingDown className="w-4 h-4 text-green-500" />
              <span className="text-green-500">-11.6%</span>
              <span className="text-gray-500">vs 上月</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-gray-600">今日费用</CardTitle>
            <DollarSign className="w-4 h-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-semibold text-gray-900">¥830</div>
            <div className="flex items-center gap-1 mt-1 text-sm">
              <TrendingUp className="w-4 h-4 text-green-500" />
              <span className="text-green-500">+16.9%</span>
              <span className="text-gray-500">vs 昨日</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-gray-600">总Token数</CardTitle>
            <DollarSign className="w-4 h-4 text-purple-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-semibold text-gray-900">124.5M</div>
            <div className="flex items-center gap-1 mt-1 text-sm">
              <TrendingUp className="w-4 h-4 text-green-500" />
              <span className="text-green-500">+23.5%</span>
              <span className="text-gray-500">vs 上月</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-gray-600">平均单价</CardTitle>
            <DollarSign className="w-4 h-4 text-orange-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-semibold text-gray-900">¥0.134</div>
            <div className="text-sm text-gray-500 mt-1">每千Token</div>
          </CardContent>
        </Card>
      </div>

      {/* 图表区域 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* 月度费用趋势 */}
        <Card>
          <CardHeader>
            <CardTitle>费用趋势（6个月）</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={280}>
              <AreaChart data={monthlyData}>
                <defs>
                  <linearGradient id="colorCost" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3}/>
                    <stop offset="95%" stopColor="#3b82f6" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis dataKey="month" stroke="#9ca3af" fontSize={12} />
                <YAxis stroke="#9ca3af" fontSize={12} />
                <Tooltip />
                <Area 
                  type="monotone" 
                  dataKey="cost" 
                  stroke="#3b82f6" 
                  fill="url(#colorCost)"
                  strokeWidth={2}
                />
              </AreaChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        {/* 模型费用分布 */}
        <Card>
          <CardHeader>
            <CardTitle>模型费用分布</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={280}>
              <BarChart data={modelCostData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis dataKey="model" stroke="#9ca3af" fontSize={12} />
                <YAxis stroke="#9ca3af" fontSize={12} />
                <Tooltip />
                <Bar dataKey="cost" fill="#8b5cf6" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        {/* 最近7天费用 */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>每日费用明细</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={250}>
              <AreaChart data={dailyData}>
                <defs>
                  <linearGradient id="colorDaily" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#f59e0b" stopOpacity={0.3}/>
                    <stop offset="95%" stopColor="#f59e0b" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis dataKey="date" stroke="#9ca3af" fontSize={12} />
                <YAxis stroke="#9ca3af" fontSize={12} />
                <Tooltip />
                <Area 
                  type="monotone" 
                  dataKey="cost" 
                  stroke="#f59e0b" 
                  fill="url(#colorDaily)"
                  strokeWidth={2}
                />
              </AreaChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      </div>

      {/* 最近交易记录 */}
      <Card>
        <CardHeader>
          <CardTitle>最近交易记录</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>日期</TableHead>
                <TableHead>模型</TableHead>
                <TableHead>类型</TableHead>
                <TableHead className="text-right">请求次数</TableHead>
                <TableHead className="text-right">Token数</TableHead>
                <TableHead className="text-right">费用</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {recentTransactions.map((transaction) => (
                <TableRow key={transaction.id}>
                  <TableCell className="text-gray-600">{transaction.date}</TableCell>
                  <TableCell className="font-medium text-gray-900">
                    {transaction.model}
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline">{getTypeText(transaction.type)}</Badge>
                  </TableCell>
                  <TableCell className="text-right text-gray-900">
                    {transaction.requests.toLocaleString()}
                  </TableCell>
                  <TableCell className="text-right text-gray-900">
                    {transaction.tokens.toLocaleString()}
                  </TableCell>
                  <TableCell className="text-right font-medium text-gray-900">
                    {transaction.cost}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
