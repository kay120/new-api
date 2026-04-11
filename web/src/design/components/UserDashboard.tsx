import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { LineChart, Line, AreaChart, Area, BarChart, Bar, PieChart, Pie, Cell, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from "recharts";
import { Activity, Key, DollarSign, Zap, Cpu, TrendingUp } from "lucide-react";
import { Badge } from "./ui/badge";

// 用户视角的模拟数据
const userRequestData = [
  { date: "4/1", requests: 120 },
  { date: "4/2", requests: 180 },
  { date: "4/3", requests: 150 },
  { date: "4/4", requests: 220 },
  { date: "4/5", requests: 190 },
  { date: "4/6", requests: 210 },
  { date: "4/7", requests: 250 },
];

const userCostData = [
  { date: "4/1", cost: 12.5 },
  { date: "4/2", cost: 18.3 },
  { date: "4/3", cost: 15.7 },
  { date: "4/4", cost: 22.1 },
  { date: "4/5", cost: 19.4 },
  { date: "4/6", cost: 21.8 },
  { date: "4/7", cost: 25.2 },
];

// 模型使用占比数据
const modelUsageData = [
  { name: "GPT-4 Turbo", value: 450, percentage: 42.5, color: "#3b82f6" },
  { name: "GPT-3.5 Turbo", value: 320, percentage: 30.2, color: "#10b981" },
  { name: "Claude-3 Opus", value: 180, percentage: 17.0, color: "#f59e0b" },
  { name: "文心一言 4.0", value: 110, percentage: 10.3, color: "#8b5cf6" },
];

// 每个密钥的模型使用情况
const keyModelUsageData = [
  {
    keyName: "生产环境密钥",
    models: [
      { name: "GPT-4 Turbo", requests: 350, color: "#3b82f6" },
      { name: "GPT-3.5 Turbo", requests: 180, color: "#10b981" },
      { name: "Claude-3 Opus", requests: 120, color: "#f59e0b" },
    ],
    totalRequests: 650,
  },
  {
    keyName: "测试环境密钥",
    models: [
      { name: "GPT-3.5 Turbo", requests: 140, color: "#10b981" },
      { name: "Claude-3 Opus", requests: 60, color: "#f59e0b" },
      { name: "文心一言 4.0", requests: 110, color: "#8b5cf6" },
    ],
    totalRequests: 310,
  },
];

// 个人模型使用7天趋势数据
const userModelTrendData = [
  { 
    date: "4/1", 
    "GPT-4 Turbo": 55, 
    "GPT-3.5 Turbo": 38, 
    "Claude-3 Opus": 18, 
    "文心一言 4.0": 9 
  },
  { 
    date: "4/2", 
    "GPT-4 Turbo": 70, 
    "GPT-3.5 Turbo": 52, 
    "Claude-3 Opus": 28, 
    "文心一言 4.0": 15 
  },
  { 
    date: "4/3", 
    "GPT-4 Turbo": 58, 
    "GPT-3.5 Turbo": 45, 
    "Claude-3 Opus": 22, 
    "文心一言 4.0": 12 
  },
  { 
    date: "4/4", 
    "GPT-4 Turbo": 82, 
    "GPT-3.5 Turbo": 60, 
    "Claude-3 Opus": 35, 
    "文心一言 4.0": 18 
  },
  { 
    date: "4/5", 
    "GPT-4 Turbo": 68, 
    "GPT-3.5 Turbo": 48, 
    "Claude-3 Opus": 26, 
    "文心一言 4.0": 14 
  },
  { 
    date: "4/6", 
    "GPT-4 Turbo": 75, 
    "GPT-3.5 Turbo": 55, 
    "Claude-3 Opus": 30, 
    "文心一言 4.0": 16 
  },
  { 
    date: "4/7", 
    "GPT-4 Turbo": 92, 
    "GPT-3.5 Turbo": 62, 
    "Claude-3 Opus": 41, 
    "文心一言 4.0": 21 
  },
];

export function UserDashboard() {
  return (
    <div className="p-6 space-y-6">
      {/* 标题 */}
      <div className="bg-gradient-to-r from-blue-50 to-purple-50 -m-6 p-6 mb-0 border-b border-blue-100">
        <div className="flex items-center gap-3">
          <div className="w-12 h-12 bg-gradient-to-br from-blue-500 to-purple-600 rounded-xl flex items-center justify-center">
            <Activity className="w-6 h-6 text-white" />
          </div>
          <div>
            <h1 className="text-2xl font-semibold text-gray-900">我的控制台</h1>
            <p className="text-sm text-gray-500 mt-1">查看您的使用情况和统计数据</p>
          </div>
        </div>
      </div>

      <div className="pt-6">
        {/* 用户统计卡片 */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-gray-600">我的API密钥</CardTitle>
              <Key className="w-4 h-4 text-blue-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-semibold text-gray-900">2</div>
              <p className="text-sm text-gray-500 mt-1">1个活跃密钥</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-gray-600">今日请求</CardTitle>
              <Activity className="w-4 h-4 text-green-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-semibold text-gray-900">250</div>
              <p className="text-sm text-green-500 mt-1">+19% vs 昨日</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-gray-600">本月费用</CardTitle>
              <DollarSign className="w-4 h-4 text-orange-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-semibold text-gray-900">¥135.0</div>
              <p className="text-sm text-gray-500 mt-1">剩余额度 ¥865</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-gray-600">成功率</CardTitle>
              <Zap className="w-4 h-4 text-purple-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-semibold text-gray-900">98.5%</div>
              <p className="text-sm text-gray-500 mt-1">服务稳定</p>
            </CardContent>
          </Card>
        </div>

        {/* 图表区域 */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mt-6">
          {/* 7天请求趋势 */}
          <Card>
            <CardHeader>
              <CardTitle>请求趋势（7天）</CardTitle>
            </CardHeader>
            <CardContent>
              <ResponsiveContainer width="100%" height={250}>
                <AreaChart data={userRequestData}>
                  <defs>
                    <linearGradient id="userRequests" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3}/>
                      <stop offset="95%" stopColor="#3b82f6" stopOpacity={0}/>
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                  <XAxis dataKey="date" stroke="#9ca3af" fontSize={12} />
                  <YAxis stroke="#9ca3af" fontSize={12} />
                  <Tooltip />
                  <Area 
                    type="monotone" 
                    dataKey="requests" 
                    stroke="#3b82f6" 
                    fill="url(#userRequests)"
                    strokeWidth={2}
                  />
                </AreaChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>

          {/* 费用趋势 */}
          <Card>
            <CardHeader>
              <CardTitle>费用趋势（7天）</CardTitle>
            </CardHeader>
            <CardContent>
              <ResponsiveContainer width="100%" height={250}>
                <LineChart data={userCostData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                  <XAxis dataKey="date" stroke="#9ca3af" fontSize={12} />
                  <YAxis stroke="#9ca3af" fontSize={12} />
                  <Tooltip />
                  <Line 
                    type="monotone" 
                    dataKey="cost" 
                    stroke="#f59e0b" 
                    strokeWidth={2}
                    dot={{ fill: '#f59e0b', r: 4 }}
                  />
                </LineChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </div>

        {/* 模型使用分析 */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* 模型使用占比 - 饼图 */}
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="flex items-center gap-2">
                  <Cpu className="w-5 h-5 text-blue-600" />
                  模型使用占比
                </CardTitle>
                <Badge variant="outline" className="text-xs">
                  总请求：{modelUsageData.reduce((sum, m) => sum + m.value, 0)}
                </Badge>
              </div>
            </CardHeader>
            <CardContent>
              <div className="flex flex-col lg:flex-row items-center gap-6">
                <div className="flex-shrink-0">
                  <ResponsiveContainer width={200} height={200}>
                    <PieChart>
                      <Pie
                        data={modelUsageData}
                        cx={100}
                        cy={100}
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
                </div>
                <div className="flex-1 w-full space-y-3">
                  {modelUsageData.map((model) => (
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

          {/* 模型使用趋势 */}
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
                  data={userModelTrendData}
                  margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
                >
                  <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                  <XAxis dataKey="date" stroke="#9ca3af" fontSize={12} />
                  <YAxis stroke="#9ca3af" fontSize={12} />
                  <Tooltip />
                  <Legend />
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
                    stroke="#f59e0b" 
                    strokeWidth={2}
                    dot={{ fill: '#f59e0b', r: 3 }}
                  />
                  <Line 
                    type="monotone" 
                    dataKey="文心一言 4.0" 
                    stroke="#8b5cf6" 
                    strokeWidth={2}
                    dot={{ fill: '#8b5cf6', r: 3 }}
                  />
                </LineChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </div>

        {/* 每个密钥的模型使用情况 */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Key className="w-5 h-5 text-purple-600" />
              API密钥模型使用详情
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            {keyModelUsageData.map((keyData) => (
              <div key={keyData.keyName} className="space-y-3">
                <div className="flex items-center justify-between">
                  <div>
                    <h4 className="font-medium text-gray-900">{keyData.keyName}</h4>
                    <p className="text-sm text-gray-500">
                      总请求：{keyData.totalRequests.toLocaleString()} 次
                    </p>
                  </div>
                </div>
                
                <div className="space-y-2">
                  {keyData.models.map((model) => {
                    const percentage = ((model.requests / keyData.totalRequests) * 100).toFixed(1);
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
                {keyModelUsageData.indexOf(keyData) < keyModelUsageData.length - 1 && (
                  <div className="border-t border-gray-100 pt-3" />
                )}
              </div>
            ))}
          </CardContent>
        </Card>

        {/* 快速操作 */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Card className="hover:shadow-md transition-shadow cursor-pointer">
            <CardContent className="pt-6">
              <div className="flex items-center gap-3">
                <div className="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center">
                  <Key className="w-6 h-6 text-blue-600" />
                </div>
                <div>
                  <h3 className="font-medium text-gray-900">管理API密钥</h3>
                  <p className="text-sm text-gray-500">创建和管理您的密钥</p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="hover:shadow-md transition-shadow cursor-pointer">
            <CardContent className="pt-6">
              <div className="flex items-center gap-3">
                <div className="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center">
                  <Activity className="w-6 h-6 text-green-600" />
                </div>
                <div>
                  <h3 className="font-medium text-gray-900">查看请求日志</h3>
                  <p className="text-sm text-gray-500">监控API调用情况</p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="hover:shadow-md transition-shadow cursor-pointer">
            <CardContent className="pt-6">
              <div className="flex items-center gap-3">
                <div className="w-12 h-12 bg-orange-100 rounded-lg flex items-center justify-center">
                  <DollarSign className="w-6 h-6 text-orange-600" />
                </div>
                <div>
                  <h3 className="font-medium text-gray-900">费用详情</h3>
                  <p className="text-sm text-gray-500">查看使用费用明细</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* 使用提示 */}
        <Card className="bg-gradient-to-r from-blue-50 to-purple-50 border-blue-200">
          <CardContent className="pt-6">
            <h3 className="font-medium text-gray-900 mb-2">💡 使用提示</h3>
            <ul className="space-y-1 text-sm text-gray-600">
              <li>• 请妥善保管您的API密钥，避免泄露</li>
              <li>• 建议为不同的应用创建不同的密钥，便于管理</li>
              <li>• 定期查看使用情况，避免超出预算</li>
              <li>• 如有异常请求，请立即联系管理员</li>
            </ul>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}