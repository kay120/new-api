import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Button } from "./ui/button";
import { Badge } from "./ui/badge";
import { Switch } from "./ui/switch";
import { Progress } from "./ui/progress";
import { Cpu, Zap, DollarSign, TrendingUp } from "lucide-react";

interface Model {
  id: string;
  name: string;
  provider: string;
  version: string;
  enabled: boolean;
  status: "healthy" | "degraded" | "offline";
  usage: number;
  avgLatency: number;
  cost: string;
  requests24h: number;
}

const initialModels: Model[] = [
  {
    id: "1",
    name: "GPT-4 Turbo",
    provider: "OpenAI",
    version: "gpt-4-turbo-preview",
    enabled: true,
    status: "healthy",
    usage: 85,
    avgLatency: 1250,
    cost: "¥0.15/1K tokens",
    requests24h: 11234,
  },
  {
    id: "2",
    name: "GPT-3.5 Turbo",
    provider: "OpenAI",
    version: "gpt-3.5-turbo",
    enabled: true,
    status: "healthy",
    usage: 62,
    avgLatency: 680,
    cost: "¥0.01/1K tokens",
    requests24h: 7456,
  },
  {
    id: "3",
    name: "Claude 3 Opus",
    provider: "Anthropic",
    version: "claude-3-opus-20240229",
    enabled: true,
    status: "healthy",
    usage: 43,
    avgLatency: 980,
    cost: "¥0.12/1K tokens",
    requests24h: 3721,
  },
  {
    id: "4",
    name: "Claude 3 Sonnet",
    provider: "Anthropic",
    version: "claude-3-sonnet-20240229",
    enabled: true,
    status: "degraded",
    usage: 28,
    avgLatency: 1520,
    cost: "¥0.05/1K tokens",
    requests24h: 1892,
  },
  {
    id: "5",
    name: "文心一言 4.0",
    provider: "百度",
    version: "ernie-bot-4",
    enabled: false,
    status: "offline",
    usage: 0,
    avgLatency: 0,
    cost: "¥0.08/1K tokens",
    requests24h: 0,
  },
  {
    id: "6",
    name: "通义千问 Max",
    provider: "阿里云",
    version: "qwen-max",
    enabled: true,
    status: "healthy",
    usage: 35,
    avgLatency: 750,
    cost: "¥0.06/1K tokens",
    requests24h: 2156,
  },
];

export function Models() {
  const [models, setModels] = useState<Model[]>(initialModels);

  const handleToggle = (id: string) => {
    setModels(models.map(model => 
      model.id === id ? { ...model, enabled: !model.enabled } : model
    ));
  };

  const getStatusColor = (status: Model["status"]) => {
    switch (status) {
      case "healthy":
        return "bg-green-100 text-green-700";
      case "degraded":
        return "bg-yellow-100 text-yellow-700";
      case "offline":
        return "bg-red-100 text-red-700";
    }
  };

  const getStatusText = (status: Model["status"]) => {
    switch (status) {
      case "healthy":
        return "正常";
      case "degraded":
        return "降级";
      case "offline":
        return "离线";
    }
  };

  const totalRequests = models.reduce((sum, m) => sum + m.requests24h, 0);
  const enabledCount = models.filter(m => m.enabled).length;
  const healthyCount = models.filter(m => m.status === "healthy" && m.enabled).length;

  return (
    <div className="p-6 space-y-6">
      {/* 标题 */}
      <div>
        <h1 className="text-2xl font-semibold text-gray-900">模型管理</h1>
        <p className="text-sm text-gray-500 mt-1">管理和监控接入的大语言模型</p>
      </div>

      {/* 统计概览 */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <div className="text-sm text-gray-600">接入模型</div>
                <div className="text-2xl font-semibold text-gray-900 mt-1">
                  {models.length}
                </div>
              </div>
              <Cpu className="w-8 h-8 text-blue-500" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <div className="text-sm text-gray-600">启用中</div>
                <div className="text-2xl font-semibold text-green-600 mt-1">
                  {enabledCount}
                </div>
              </div>
              <Zap className="w-8 h-8 text-green-500" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <div className="text-sm text-gray-600">健康状态</div>
                <div className="text-2xl font-semibold text-gray-900 mt-1">
                  {healthyCount}/{enabledCount}
                </div>
              </div>
              <TrendingUp className="w-8 h-8 text-purple-500" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <div className="text-sm text-gray-600">24h请求</div>
                <div className="text-2xl font-semibold text-gray-900 mt-1">
                  {totalRequests.toLocaleString()}
                </div>
              </div>
              <DollarSign className="w-8 h-8 text-orange-500" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* 模型列表 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {models.map((model) => (
          <Card key={model.id} className={!model.enabled ? "opacity-60" : ""}>
            <CardHeader>
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <CardTitle className="text-lg">{model.name}</CardTitle>
                    <Badge className={getStatusColor(model.status)}>
                      {getStatusText(model.status)}
                    </Badge>
                  </div>
                  <div className="text-sm text-gray-500 mt-1">
                    {model.provider} · {model.version}
                  </div>
                </div>
                <Switch
                  checked={model.enabled}
                  onCheckedChange={() => handleToggle(model.id)}
                />
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* 使用率 */}
              <div>
                <div className="flex items-center justify-between text-sm mb-2">
                  <span className="text-gray-600">使用率</span>
                  <span className="font-medium">{model.usage}%</span>
                </div>
                <Progress value={model.usage} className="h-2" />
              </div>

              {/* 指标 */}
              <div className="grid grid-cols-3 gap-4 pt-2 border-t border-gray-100">
                <div>
                  <div className="text-xs text-gray-500">平均延迟</div>
                  <div className="text-sm font-medium text-gray-900 mt-1">
                    {model.avgLatency > 0 ? `${model.avgLatency}ms` : "-"}
                  </div>
                </div>
                <div>
                  <div className="text-xs text-gray-500">计费标准</div>
                  <div className="text-sm font-medium text-gray-900 mt-1">
                    {model.cost}
                  </div>
                </div>
                <div>
                  <div className="text-xs text-gray-500">24h请求</div>
                  <div className="text-sm font-medium text-gray-900 mt-1">
                    {model.requests24h.toLocaleString()}
                  </div>
                </div>
              </div>

              {/* 操作按钮 */}
              <div className="flex gap-2 pt-2">
                <Button variant="outline" size="sm" className="flex-1">
                  配置
                </Button>
                <Button variant="outline" size="sm" className="flex-1">
                  监控
                </Button>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* 添加模型按钮 */}
      <Card className="border-dashed">
        <CardContent className="flex flex-col items-center justify-center py-12">
          <Cpu className="w-12 h-12 text-gray-400 mb-3" />
          <h3 className="text-lg font-medium text-gray-900 mb-1">添加新模型</h3>
          <p className="text-sm text-gray-500 mb-4">接入更多大语言模型提供商</p>
          <Button>添加模型</Button>
        </CardContent>
      </Card>
    </div>
  );
}
