import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Label } from "./ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./ui/select";
import { Switch } from "./ui/switch";
import { Badge } from "./ui/badge";
import { Slider } from "./ui/slider";
import { Plus, Edit, Trash2, Shield } from "lucide-react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "./ui/dialog";
import { toast } from "sonner";

interface RateLimitRule {
  id: string;
  name: string;
  enabled: boolean;
  type: "global" | "user" | "api-key";
  target: string;
  limit: number;
  window: string;
  action: "throttle" | "block";
  priority: number;
}

const initialRules: RateLimitRule[] = [
  {
    id: "1",
    name: "全局限流",
    enabled: true,
    type: "global",
    target: "*",
    limit: 10000,
    window: "1分钟",
    action: "throttle",
    priority: 1,
  },
  {
    id: "2",
    name: "用户级限流",
    enabled: true,
    type: "user",
    target: "默认用户",
    limit: 100,
    window: "1分钟",
    action: "throttle",
    priority: 2,
  },
  {
    id: "3",
    name: "VIP用户限流",
    enabled: true,
    type: "user",
    target: "VIP用户",
    limit: 500,
    window: "1分钟",
    action: "throttle",
    priority: 3,
  },
  {
    id: "4",
    name: "API密钥限流",
    enabled: true,
    type: "api-key",
    target: "生产环境",
    limit: 1000,
    window: "1小时",
    action: "block",
    priority: 4,
  },
];

export function RateLimit() {
  const [rules, setRules] = useState<RateLimitRule[]>(initialRules);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingRule, setEditingRule] = useState<RateLimitRule | null>(null);
  const [formData, setFormData] = useState({
    name: "",
    type: "user" as RateLimitRule["type"],
    target: "",
    limit: 100,
    window: "1分钟",
    action: "throttle" as RateLimitRule["action"],
  });

  const handleToggle = (id: string) => {
    setRules(rules.map(rule => 
      rule.id === id ? { ...rule, enabled: !rule.enabled } : rule
    ));
    toast.success("规则状态已更新");
  };

  const handleDelete = (id: string) => {
    setRules(rules.filter(rule => rule.id !== id));
    toast.success("规则已删除");
  };

  const handleSave = () => {
    if (!formData.name || !formData.target) {
      toast.error("请填写完整信息");
      return;
    }

    if (editingRule) {
      setRules(rules.map(rule => 
        rule.id === editingRule.id 
          ? { ...rule, ...formData }
          : rule
      ));
      toast.success("规则已更新");
    } else {
      const newRule: RateLimitRule = {
        id: Date.now().toString(),
        ...formData,
        enabled: true,
        priority: rules.length + 1,
      };
      setRules([...rules, newRule]);
      toast.success("规则已创建");
    }

    setDialogOpen(false);
    setEditingRule(null);
    resetForm();
  };

  const resetForm = () => {
    setFormData({
      name: "",
      type: "user",
      target: "",
      limit: 100,
      window: "1分钟",
      action: "throttle",
    });
  };

  const openEditDialog = (rule: RateLimitRule) => {
    setEditingRule(rule);
    setFormData({
      name: rule.name,
      type: rule.type,
      target: rule.target,
      limit: rule.limit,
      window: rule.window,
      action: rule.action,
    });
    setDialogOpen(true);
  };

  const openCreateDialog = () => {
    setEditingRule(null);
    resetForm();
    setDialogOpen(true);
  };

  const getTypeText = (type: RateLimitRule["type"]) => {
    switch (type) {
      case "global": return "全局";
      case "user": return "用户";
      case "api-key": return "API密钥";
    }
  };

  const getActionText = (action: RateLimitRule["action"]) => {
    return action === "throttle" ? "限流" : "阻断";
  };

  const enabledCount = rules.filter(r => r.enabled).length;
  const totalRequests = rules.reduce((sum, r) => sum + r.limit, 0);

  return (
    <div className="p-6 space-y-6">
      {/* 标题 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">限流配置</h1>
          <p className="text-sm text-gray-500 mt-1">管理API访问频率限制规则</p>
        </div>
        <Button onClick={openCreateDialog} className="gap-2">
          <Plus className="w-4 h-4" />
          新建规则
        </Button>
      </div>

      {/* 统计概览 */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center">
                <Shield className="w-5 h-5 text-blue-600" />
              </div>
              <div>
                <div className="text-sm text-gray-600">总规则数</div>
                <div className="text-2xl font-semibold text-gray-900">{rules.length}</div>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center">
                <Shield className="w-5 h-5 text-green-600" />
              </div>
              <div>
                <div className="text-sm text-gray-600">已启用</div>
                <div className="text-2xl font-semibold text-green-600">{enabledCount}</div>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-purple-100 rounded-lg flex items-center justify-center">
                <Shield className="w-5 h-5 text-purple-600" />
              </div>
              <div>
                <div className="text-sm text-gray-600">总限额</div>
                <div className="text-2xl font-semibold text-gray-900">
                  {totalRequests.toLocaleString()}
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* 规则列表 */}
      <div className="space-y-3">
        {rules.map((rule) => (
          <Card key={rule.id} className={!rule.enabled ? "opacity-60" : ""}>
            <CardContent className="pt-6">
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-3 mb-2">
                    <h3 className="font-semibold text-gray-900">{rule.name}</h3>
                    <Badge variant="outline">{getTypeText(rule.type)}</Badge>
                    <Badge 
                      className={rule.action === "throttle" 
                        ? "bg-blue-100 text-blue-700" 
                        : "bg-red-100 text-red-700"
                      }
                    >
                      {getActionText(rule.action)}
                    </Badge>
                  </div>
                  
                  <div className="grid grid-cols-1 sm:grid-cols-4 gap-4 text-sm">
                    <div>
                      <span className="text-gray-500">目标：</span>
                      <span className="font-medium text-gray-900 ml-1">{rule.target}</span>
                    </div>
                    <div>
                      <span className="text-gray-500">限制：</span>
                      <span className="font-medium text-gray-900 ml-1">
                        {rule.limit.toLocaleString()} 次
                      </span>
                    </div>
                    <div>
                      <span className="text-gray-500">时间窗口：</span>
                      <span className="font-medium text-gray-900 ml-1">{rule.window}</span>
                    </div>
                    <div>
                      <span className="text-gray-500">优先级：</span>
                      <span className="font-medium text-gray-900 ml-1">{rule.priority}</span>
                    </div>
                  </div>
                </div>

                <div className="flex items-center gap-2 ml-4">
                  <Switch
                    checked={rule.enabled}
                    onCheckedChange={() => handleToggle(rule.id)}
                  />
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => openEditDialog(rule)}
                  >
                    <Edit className="w-4 h-4" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleDelete(rule.id)}
                    className="text-red-600 hover:text-red-700 hover:bg-red-50"
                  >
                    <Trash2 className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* 创建/编辑对话框 */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>
              {editingRule ? "编辑规则" : "创建新规则"}
            </DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="name">规则名称</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="例如：VIP用户限流"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="type">规则类型</Label>
                <Select 
                  value={formData.type} 
                  onValueChange={(value: RateLimitRule["type"]) => 
                    setFormData({ ...formData, type: value })
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="global">全局</SelectItem>
                    <SelectItem value="user">用户</SelectItem>
                    <SelectItem value="api-key">API密钥</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="action">处理动作</Label>
                <Select 
                  value={formData.action} 
                  onValueChange={(value: RateLimitRule["action"]) => 
                    setFormData({ ...formData, action: value })
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="throttle">限流</SelectItem>
                    <SelectItem value="block">阻断</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="target">目标</Label>
              <Input
                id="target"
                value={formData.target}
                onChange={(e) => setFormData({ ...formData, target: e.target.value })}
                placeholder="例如：VIP用户、生产环境等"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="limit">请求限制：{formData.limit} 次</Label>
              <Slider
                value={[formData.limit]}
                onValueChange={([value]) => setFormData({ ...formData, limit: value })}
                min={10}
                max={10000}
                step={10}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="window">时间窗口</Label>
              <Select 
                value={formData.window} 
                onValueChange={(value) => 
                  setFormData({ ...formData, window: value })
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="1分钟">1分钟</SelectItem>
                  <SelectItem value="5分钟">5分钟</SelectItem>
                  <SelectItem value="1小时">1小时</SelectItem>
                  <SelectItem value="1天">1天</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              取消
            </Button>
            <Button onClick={handleSave}>
              {editingRule ? "保存" : "创建"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
