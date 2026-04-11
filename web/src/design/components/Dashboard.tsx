import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { AreaChart, Area, BarChart, Bar, LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from "recharts";
import { TrendingUp, TrendingDown, Activity, Cpu, Clock, DollarSign, Users, Key, Trophy, Crown } from "lucide-react";
import { Badge } from "./ui/badge";
import { Avatar, AvatarFallback } from "./ui/avatar";

// 模拟数据
const requestData = [
  { time: "00:00", requests: 1200 },
  { time: "04:00", requests: 800 },
  { time: "08:00", requests: 2400 },
  { time: "12:00", requests: 3800 },
  { time: "16:00", requests: 3200 },
  { time: "20:00", requests: 2800 },
  { time: "23:59", requests: 1600 },
];

const modelUsageData = [
  { name: "GPT-4", value: 45, color: "#3b82f6" },
  { name: "GPT-3.5", value: 30, color: "#8b5cf6" },
  { name: "Claude-3", value: 15, color: "#ec4899" },
  { name: "文心一言", value: 10, color: "#f59e0b" },
];

const responseTimeData = [
  { time: "周一", avg: 245 },
  { time: "周二", avg: 198 },
  { time: "周三", avg: 267 },
  { time: "周四", avg: 223 },
  { time: "周五", avg: 189 },
  { time: "周六", avg: 156 },
  { time: "周日", avg: 134 },
];

const costData = [
  { month: "1月", cost: 12500 },
  { month: "2月", cost: 15200 },
  { month: "3月", cost: 18900 },
  { month: "4月", cost: 16700 },
];

// 全局模型使用详细数据
const globalModelUsageData = [
  { name: "GPT-4 Turbo", value: 11050, percentage: 45.0, color: "#3b82f6" },
  { name: "GPT-3.5 Turbo", value: 7380, percentage: 30.0, color: "#10b981" },
  { name: "Claude-3 Opus", value: 3690, percentage: 15.0, color: "#ec4899" },
  { name: "文心一言 4.0", value: 2460, percentage: 10.0, color: "#f59e0b" },
];

// 用户请求排行榜数据
const userRequestRanking = [
  { 
    id: 1, 
    name: "张伟", 
    email: "zhangwei@company.com", 
    requests: 8920, 
    trend: "+15.2%",
    apiKeys: 3,
    topModel: "GPT-4 Turbo"
  },
  { 
    id: 2, 
    name: "李娜", 
    email: "lina@company.com", 
    requests: 6540, 
    trend: "+8.7%",
    apiKeys: 2,
    topModel: "GPT-3.5 Turbo"
  },
  { 
    id: 3, 
    name: "王芳", 
    email: "wangfang@company.com", 
    requests: 5320, 
    trend: "+12.1%",
    apiKeys: 4,
    topModel: "Claude-3 Opus"
  },
  { 
    id: 4, 
    name: "刘强", 
    email: "liuqiang@company.com", 
    requests: 4180, 
    trend: "+5.3%",
    apiKeys: 2,
    topModel: "GPT-4 Turbo"
  },
  { 
    id: 5, 
    name: "陈静", 
    email: "chenjing@company.com", 
    requests: 3760, 
    trend: "+9.8%",
    apiKeys: 1,
    topModel: "文心一言 4.0"
  },
];

// 用户费用排行榜数据
const userCostRanking = [
  { 
    id: 1, 
    name: "张伟", 
    email: "zhangwei@company.com", 
    cost: 4560, 
    trend: "+12.5%",
    percentage: 27.3
  },
  { 
    id: 2, 
    name: "王芳", 
    email: "wangfang@company.com", 
    cost: 3890, 
    trend: "+18.2%",
    percentage: 23.3
  },
  { 
    id: 3, 
    name: "李娜", 
    email: "lina@company.com", 
    cost: 2940, 
    trend: "+6.7%",
    percentage: 17.6
  },
  { 
    id: 4, 
    name: "刘强", 
    email: "liuqiang@company.com", 
    cost: 2650, 
    trend: "+4.1%",
    percentage: 15.9
  },
  { 
    id: 5, 
    name: "陈静", 
    email: "chenjing@company.com", 
    cost: 2660, 
    trend: "+15.3%",
    percentage: 15.9
  },
];

// 用户模型使用详细数据
const userModelBreakdown = [
  {
    userId: 1,
    name: "张伟",
    models: [
      { name: "GPT-4 Turbo", requests: 5890, color: "#3b82f6" },
      { name: "GPT-3.5 Turbo", requests: 2130, color: "#10b981" },
      { name: "Claude-3 Opus", requests: 900, color: "#ec4899" },
    ],
    totalRequests: 8920,
  },
  {
    userId: 2,
    name: "李娜",
    models: [
      { name: "GPT-3.5 Turbo", requests: 4320, color: "#10b981" },
      { name: "GPT-4 Turbo", requests: 1880, color: "#3b82f6" },
      { name: "文心一言 4.0", requests: 340, color: "#f59e0b" },
    ],
    totalRequests: 6540,
  },
  {
    userId: 3,
    name: "王芳",
    models: [
      { name: "Claude-3 Opus", requests: 2890, color: "#ec4899" },
      { name: "GPT-4 Turbo", requests: 1560, color: "#3b82f6" },
      { name: "GPT-3.5 Turbo", requests: 870, color: "#10b981" },
    ],
    totalRequests: 5320,
  },
];

// 模型使用7天趋势数据
const modelTrendData = [
  { 
    date: "4/1", 
    "GPT-4 Turbo": 1420, 
    "GPT-3.5 Turbo": 980, 
    "Claude-3 Opus": 450, 
    "文心一言 4.0": 320 
  },
  { 
    date: "4/2", 
    "GPT-4 Turbo": 1580, 
    "GPT-3.5 Turbo": 1120, 
    "Claude-3 Opus": 520, 
    "文心一言 4.0": 340 
  },
  { 
    date: "4/3", 
    "GPT-4 Turbo": 1520, 
    "GPT-3.5 Turbo": 1050, 
    "Claude-3 Opus": 490, 
    "文心一言 4.0": 310 
  },
  { 
    date: "4/4", 
    "GPT-4 Turbo": 1680, 
    "GPT-3.5 Turbo": 1180, 
    "Claude-3 Opus": 580, 
    "文心一言 4.0": 360 
  },
  { 
    date: "4/5", 
    "GPT-4 Turbo": 1550, 
    "GPT-3.5 Turbo": 1090, 
    "Claude-3 Opus": 510, 
    "文心一言 4.0": 330 
  },
  { 
    date: "4/6", 
    "GPT-4 Turbo": 1620, 
    "GPT-3.5 Turbo": 1140, 
    "Claude-3 Opus": 540, 
    "文心一言 4.0": 350 
  },
  { 
    date: "4/7", 
    "GPT-4 Turbo": 1680, 
    "GPT-3.5 Turbo": 1200, 
    "Claude-3 Opus": 600, 
    "文心一言 4.0": 370 
  },
];

export function Dashboard() {
  return (
    <div className="p-6 space-y-6">
      {/* 标题 */}
      <div>
        <h1 className="text-2xl font-semibold text-gray-900">控制台</h1>
        <p className="text-sm text-gray-500 mt-1">实时监控大模型网关运行状态</p>
      </div>

      {/* 关键指标卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-gray-600">今日请求</CardTitle>
            <Activity className="w-4 h-4 text-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-semibold text-gray-900">24,567</div>
            <div className="flex items-center gap-1 mt-1 text-sm">
              <TrendingUp className="w-4 h-4 text-green-500" />
              <span className="text-green-500">+12.5%</span>
              <span className="text-gray-500">vs 昨日</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-gray-600">成功率</CardTitle>
            <Cpu className="w-4 h-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-semibold text-gray-900">99.8%</div>
            <div className="flex items-center gap-1 mt-1 text-sm">
              <TrendingUp className="w-4 h-4 text-green-500" />
              <span className="text-green-500">+0.2%</span>
              <span className="text-gray-500">vs 昨日</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-gray-600">平均响应</CardTitle>
            <Clock className="w-4 h-4 text-purple-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-semibold text-gray-900">245ms</div>
            <div className="flex items-center gap-1 mt-1 text-sm">
              <TrendingDown className="w-4 h-4 text-green-500" />
              <span className="text-green-500">-18ms</span>
              <span className="text-gray-500">vs 昨日</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-gray-600">本月费用</CardTitle>
            <DollarSign className="w-4 h-4 text-orange-500" />
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
      </div>

      {/* 图表区域 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* 请求趋势 */}
        <Card>
          <CardHeader>
            <CardTitle>24小时请求趋势</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={250}>
              <AreaChart data={requestData}>
                <defs>
                  <linearGradient id="colorRequests" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3}/>
                    <stop offset="95%" stopColor="#3b82f6" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis dataKey="time" stroke="#9ca3af" fontSize={12} />
                <YAxis stroke="#9ca3af" fontSize={12} />
                <Tooltip />
                <Area 
                  type="monotone" 
                  dataKey="requests" 
                  stroke="#3b82f6" 
                  fill="url(#colorRequests)"
                  strokeWidth={2}
                />
              </AreaChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        {/* 模型使用分布 */}
        <Card>
          <CardHeader>
            <CardTitle>模型使用分布</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <ResponsiveContainer width="50%" height={250}>
                <PieChart>
                  <Pie
                    data={modelUsageData}
                    cx="50%"
                    cy="50%"
                    innerRadius={60}
                    outerRadius={90}
                    paddingAngle={2}
                    dataKey="value"
                  >
                    {modelUsageData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
              <div className="space-y-3">
                {modelUsageData.map((item) => (
                  <div key={item.name} className="flex items-center gap-3">
                    <div 
                      className="w-3 h-3 rounded-full" 
                      style={{ backgroundColor: item.color }}
                    />
                    <span className="text-sm text-gray-600 min-w-20">{item.name}</span>
                    <span className="text-sm font-medium text-gray-900">{item.value}%</span>
                  </div>
                ))}
              </div>
            </div>
          </CardContent>
        </Card>

        {/* 响应时间趋势 */}
        <Card>
          <CardHeader>
            <CardTitle>平均响应时间（7天）</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={250}>
              <LineChart data={responseTimeData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis dataKey="time" stroke="#9ca3af" fontSize={12} />
                <YAxis stroke="#9ca3af" fontSize={12} />
                <Tooltip />
                <Line 
                  type="monotone" 
                  dataKey="avg" 
                  stroke="#8b5cf6" 
                  strokeWidth={2}
                  dot={{ fill: '#8b5cf6', r: 4 }}
                />
              </LineChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        {/* 成本趋势 */}
        <Card>
          <CardHeader>
            <CardTitle>月度成本趋势</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={250}>
              <BarChart data={costData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis dataKey="month" stroke="#9ca3af" fontSize={12} />
                <YAxis stroke="#9ca3af" fontSize={12} />
                <Tooltip />
                <Bar dataKey="cost" fill="#f59e0b" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      </div>

      {/* 全局模型使用分析 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* 全局模型使用占比详情 */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="flex items-center gap-2">
                <Cpu className="w-5 h-5 text-blue-600" />
                全局模型使用统计
              </CardTitle>
              <Badge variant="outline" className="text-xs">
                总请求：{globalModelUsageData.reduce((sum, m) => sum + m.value, 0).toLocaleString()}
              </Badge>
            </div>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col lg:flex-row items-center gap-6">
              <div className="flex-shrink-0">
                <ResponsiveContainer width={200} height={200}>
                  <PieChart>
                    <Pie
                      data={globalModelUsageData}
                      cx={100}
                      cy={100}
                      innerRadius={60}
                      outerRadius={90}
                      paddingAngle={2}
                      dataKey="value"
                    >
                      {globalModelUsageData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={entry.color} />
                      ))}
                    </Pie>
                    <Tooltip />
                  </PieChart>
                </ResponsiveContainer>
              </div>
              <div className="flex-1 w-full space-y-3">
                {globalModelUsageData.map((model) => (
                  <div key={model.name} className="space-y-1">
                    <div className="flex items-center justify-between text-sm">
                      <div className="flex items-center gap-2">
                        <div 
                          className="w-3 h-3 rounded-full" 
                          style={{ backgroundColor: model.color }}
                        />
                        <span className="text-gray-700">{model.name}</span>
                      </div>
                      <span className="font-medium text-gray-900">
                        {model.percentage}%
                      </span>
                    </div>
                    <div className="w-full bg-gray-100 rounded-full h-2">
                      <div 
                        className="h-2 rounded-full transition-all"
                        style={{ 
                          width: `${model.percentage}%`,
                          backgroundColor: model.color 
                        }}
                      />
                    </div>
                    <div className="text-xs text-gray-500">
                      {model.value.toLocaleString()} 次请求
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </CardContent>
        </Card>

        {/* 模型使用趋势（7天） */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <TrendingUp className="w-5 h-5 text-green-600" />
              模型使用趋势（7天）
            </CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={250}>
              <LineChart 
                data={modelTrendData}
                margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
              >
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis dataKey="date" stroke="#9ca3af" fontSize={12} />
                <YAxis stroke="#9ca3af" fontSize={12} />
                <Tooltip />
                <Line 
                  type="monotone" 
                  dataKey="GPT-4 Turbo" 
                  stroke="#3b82f6" 
                  strokeWidth={2}
                  dot={{ fill: '#3b82f6', r: 3 }}
                />
                <Line 
                  type="monotone" 
                  dataKey="GPT-3.5 Turbo" 
                  stroke="#10b981" 
                  strokeWidth={2}
                  dot={{ fill: '#10b981', r: 3 }}
                />
                <Line 
                  type="monotone" 
                  dataKey="Claude-3 Opus" 
                  stroke="#ec4899" 
                  strokeWidth={2}
                  dot={{ fill: '#ec4899', r: 3 }}
                />
                <Line 
                  type="monotone" 
                  dataKey="文心一言 4.0" 
                  stroke="#f59e0b" 
                  strokeWidth={2}
                  dot={{ fill: '#f59e0b', r: 3 }}
                />
              </LineChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      </div>

      {/* 用户排行榜 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* 用户请求排行榜 */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Trophy className="w-5 h-5 text-yellow-600" />
              用户请求排行榜
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {userRequestRanking.map((user, index) => (
                <div 
                  key={user.id} 
                  className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
                >
                  <div className="flex-shrink-0 w-8 h-8 flex items-center justify-center">
                    {index === 0 ? (
                      <Crown className="w-6 h-6 text-yellow-500" />
                    ) : (
                      <div className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-100 to-purple-100 flex items-center justify-center">
                        <span className="text-sm font-semibold text-gray-700">#{index + 1}</span>
                      </div>
                    )}
                  </div>
                  <Avatar className="w-10 h-10">
                    <AvatarFallback className="bg-gradient-to-br from-blue-500 to-purple-600 text-white">
                      {user.name.charAt(0)}
                    </AvatarFallback>
                  </Avatar>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <p className="font-medium text-gray-900 truncate">{user.name}</p>
                      {index < 3 && (
                        <Badge variant="outline" className="text-xs border-yellow-300 text-yellow-700 bg-yellow-50">
                          Top {index + 1}
                        </Badge>
                      )}
                    </div>
                    <div className="flex items-center gap-3 mt-1">
                      <p className="text-xs text-gray-500 truncate">{user.email}</p>
                      <div className="flex items-center gap-1 text-xs text-gray-500">
                        <Key className="w-3 h-3" />
                        {user.apiKeys}
                      </div>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="text-lg font-semibold text-gray-900">
                      {user.requests.toLocaleString()}
                    </p>
                    <p className="text-xs text-green-600">{user.trend}</p>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* 用户费用排行榜 */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <DollarSign className="w-5 h-5 text-orange-600" />
              用户费用排行榜
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {userCostRanking.map((user, index) => (
                <div 
                  key={user.id} 
                  className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
                >
                  <div className="flex-shrink-0 w-8 h-8 flex items-center justify-center">
                    {index === 0 ? (
                      <Crown className="w-6 h-6 text-orange-500" />
                    ) : (
                      <div className="w-8 h-8 rounded-full bg-gradient-to-br from-orange-100 to-red-100 flex items-center justify-center">
                        <span className="text-sm font-semibold text-gray-700">#{index + 1}</span>
                      </div>
                    )}
                  </div>
                  <Avatar className="w-10 h-10">
                    <AvatarFallback className="bg-gradient-to-br from-orange-500 to-red-600 text-white">
                      {user.name.charAt(0)}
                    </AvatarFallback>
                  </Avatar>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <p className="font-medium text-gray-900 truncate">{user.name}</p>
                      {index < 3 && (
                        <Badge variant="outline" className="text-xs border-orange-300 text-orange-700 bg-orange-50">
                          Top {index + 1}
                        </Badge>
                      )}
                    </div>
                    <p className="text-xs text-gray-500 truncate mt-1">{user.email}</p>
                  </div>
                  <div className="text-right">
                    <p className="text-lg font-semibold text-gray-900">
                      ¥{user.cost.toLocaleString()}
                    </p>
                    <p className="text-xs text-gray-500">{user.percentage}%</p>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* 用户模型使用详细分析 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Users className="w-5 h-5 text-purple-600" />
            Top用户模型使用详情
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          {userModelBreakdown.map((userData, index) => (
            <div key={userData.userId} className="space-y-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <Avatar className="w-10 h-10">
                    <AvatarFallback className="bg-gradient-to-br from-purple-500 to-pink-600 text-white">
                      {userData.name.charAt(0)}
                    </AvatarFallback>
                  </Avatar>
                  <div>
                    <h4 className="font-medium text-gray-900 flex items-center gap-2">
                      {userData.name}
                      {index === 0 && <Crown className="w-4 h-4 text-yellow-500" />}
                    </h4>
                    <p className="text-sm text-gray-500">
                      总请求：{userData.totalRequests.toLocaleString()} 次
                    </p>
                  </div>
                </div>
              </div>
              
              <div className="space-y-2">
                {userData.models.map((model) => {
                  const percentage = ((model.requests / userData.totalRequests) * 100).toFixed(1);
                  return (
                    <div key={model.name} className="space-y-1">
                      <div className="flex items-center justify-between text-sm">
                        <div className="flex items-center gap-2">
                          <div 
                            className="w-3 h-3 rounded-full" 
                            style={{ backgroundColor: model.color }}
                          />
                          <span className="text-gray-700">{model.name}</span>
                        </div>
                        <div className="flex items-center gap-3">
                          <span className="text-gray-500">
                            {model.requests.toLocaleString()} 次
                          </span>
                          <span className="font-medium text-gray-900 w-12 text-right">
                            {percentage}%
                          </span>
                        </div>
                      </div>
                      <div className="w-full bg-gray-100 rounded-full h-1.5">
                        <div 
                          className="h-1.5 rounded-full transition-all"
                          style={{ 
                            width: `${percentage}%`,
                            backgroundColor: model.color 
                          }}
                        />
                      </div>
                    </div>
                  );
                })}
              </div>

              {/* 分隔线 */}
              {index < userModelBreakdown.length - 1 && (
                <div className="border-t border-gray-100 pt-3" />
              )}
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  );
}