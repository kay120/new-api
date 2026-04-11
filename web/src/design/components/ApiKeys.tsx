import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Label } from "./ui/label";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "./ui/dialog";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "./ui/table";
import { Badge } from "./ui/badge";
import { Plus, Copy, Eye, EyeOff, Trash2, CheckCircle2, Cpu } from "lucide-react";
import { toast } from "sonner";
import { Checkbox } from "./ui/checkbox";

interface ApiKey {
  id: string;
  name: string;
  key: string;
  status: "active" | "disabled";
  createdAt: string;
  lastUsed: string;
  requests: number;
  models: string[];
}

const availableModels = [
  "GPT-4 Turbo",
  "GPT-3.5 Turbo",
  "Claude-3 Opus",
  "Claude-3 Sonnet",
  "文心一言 4.0",
  "通义千问 Max",
];

const initialKeys: ApiKey[] = [
  {
    id: "1",
    name: "生产环境密钥",
    key: "sk-prod-a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
    status: "active",
    createdAt: "2026-03-01",
    lastUsed: "2026-04-07 14:23",
    requests: 125678,
    models: ["GPT-4 Turbo", "GPT-3.5 Turbo", "Claude-3 Opus"],
  },
  {
    id: "2",
    name: "测试环境密钥",
    key: "sk-test-q9w8e7r6t5y4u3i2o1p0a9s8d7f6g5h4",
    status: "active",
    createdAt: "2026-02-15",
    lastUsed: "2026-04-07 10:15",
    requests: 8934,
    models: ["GPT-3.5 Turbo"],
  },
  {
    id: "3",
    name: "开发环境密钥",
    key: "sk-dev-z1x2c3v4b5n6m7l8k9j0h1g2f3d4s5a6",
    status: "disabled",
    createdAt: "2026-01-20",
    lastUsed: "2026-03-28 16:42",
    requests: 4521,
    models: ["通义千问 Max", "文心一言 4.0"],
  },
];

export function ApiKeys() {
  const [keys, setKeys] = useState<ApiKey[]>(initialKeys);
  const [showKey, setShowKey] = useState<string | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [newKeyName, setNewKeyName] = useState("");
  const [selectedModels, setSelectedModels] = useState<string[]>([]);

  const handleCopy = (key: string) => {
    navigator.clipboard.writeText(key);
    toast.success("API密钥已复制到剪贴板");
  };

  const handleDelete = (id: string) => {
    setKeys(keys.filter(k => k.id !== id));
    toast.success("API密钥已删除");
  };

  const handleCreate = () => {
    if (!newKeyName.trim()) {
      toast.error("请输入密钥名称");
      return;
    }

    if (selectedModels.length === 0) {
      toast.error("请至少选择一个模型");
      return;
    }

    const newKey: ApiKey = {
      id: Date.now().toString(),
      name: newKeyName,
      key: `sk-${Date.now()}-${Math.random().toString(36).substring(2, 15)}`,
      status: "active",
      createdAt: new Date().toLocaleDateString('zh-CN'),
      lastUsed: "-",
      requests: 0,
      models: selectedModels,
    };

    setKeys([newKey, ...keys]);
    setNewKeyName("");
    setSelectedModels([]);
    setDialogOpen(false);
    toast.success("API密钥创建成功");
  };

  const toggleModel = (model: string) => {
    setSelectedModels(prev =>
      prev.includes(model)
        ? prev.filter(m => m !== model)
        : [...prev, model]
    );
  };

  const maskKey = (key: string) => {
    return `${key.slice(0, 12)}${'*'.repeat(20)}${key.slice(-4)}`;
  };

  return (
    <div className="p-6 space-y-6">
      {/* 标题和操作 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">API密钥管理</h1>
          <p className="text-sm text-gray-500 mt-1">创建和管理您的API密钥</p>
        </div>
        
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger asChild>
            <Button className="gap-2">
              <Plus className="w-4 h-4" />
              创建密钥
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>创建新的API密钥</DialogTitle>
            </DialogHeader>
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="keyName">密钥名称</Label>
                <Input
                  id="keyName"
                  placeholder="例如：生产环境密钥"
                  value={newKeyName}
                  onChange={(e) => setNewKeyName(e.target.value)}
                />
              </div>
              <div className="bg-amber-50 border border-amber-200 rounded-lg p-3">
                <p className="text-sm text-amber-800">
                  ⚠️ 密钥创建后请立即保存，系统不会再次显示完整密钥。
                </p>
              </div>
              <div className="space-y-2">
                <Label>可用模型</Label>
                <div className="space-y-2 border border-gray-200 rounded-lg p-3 max-h-60 overflow-y-auto">
                  {availableModels.map(model => (
                    <div key={model} className="flex items-center space-x-2">
                      <Checkbox
                        id={model}
                        checked={selectedModels.includes(model)}
                        onCheckedChange={() => toggleModel(model)}
                      />
                      <Label 
                        htmlFor={model} 
                        className="flex items-center gap-2 cursor-pointer text-sm"
                      >
                        <Cpu className="w-4 h-4 text-gray-500" />
                        {model}
                      </Label>
                    </div>
                  ))}
                </div>
                <p className="text-xs text-gray-500">已选择 {selectedModels.length} 个模型</p>
              </div>
            </div>
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => setDialogOpen(false)}>
                取消
              </Button>
              <Button onClick={handleCreate}>
                创建
              </Button>
            </div>
          </DialogContent>
        </Dialog>
      </div>

      {/* 密钥统计 */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardContent className="pt-6">
            <div className="text-sm text-gray-600">总密钥数</div>
            <div className="text-2xl font-semibold text-gray-900 mt-1">{keys.length}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="text-sm text-gray-600">活跃密钥</div>
            <div className="text-2xl font-semibold text-green-600 mt-1">
              {keys.filter(k => k.status === "active").length}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="text-sm text-gray-600">总请求次数</div>
            <div className="text-2xl font-semibold text-gray-900 mt-1">
              {keys.reduce((sum, k) => sum + k.requests, 0).toLocaleString()}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* 密钥列表 */}
      <Card>
        <CardHeader>
          <CardTitle>密钥列表</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>名称</TableHead>
                <TableHead>密钥</TableHead>
                <TableHead>可用模型</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>创建时间</TableHead>
                <TableHead>最后使用</TableHead>
                <TableHead className="text-right">请求次数</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {keys.map((key) => (
                <TableRow key={key.id}>
                  <TableCell className="font-medium">{key.name}</TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <code className="text-sm bg-gray-100 px-2 py-1 rounded">
                        {showKey === key.id ? key.key : maskKey(key.key)}
                      </code>
                      <button
                        onClick={() => setShowKey(showKey === key.id ? null : key.id)}
                        className="text-gray-500 hover:text-gray-700"
                      >
                        {showKey === key.id ? (
                          <EyeOff className="w-4 h-4" />
                        ) : (
                          <Eye className="w-4 h-4" />
                        )}
                      </button>
                      <button
                        onClick={() => handleCopy(key.key)}
                        className="text-gray-500 hover:text-gray-700"
                      >
                        <Copy className="w-4 h-4" />
                      </button>
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="flex flex-wrap gap-1">
                      {key.models.slice(0, 2).map(model => (
                        <Badge key={model} variant="outline" className="text-xs">
                          {model}
                        </Badge>
                      ))}
                      {key.models.length > 2 && (
                        <Badge variant="outline" className="text-xs">
                          +{key.models.length - 2}
                        </Badge>
                      )}
                    </div>
                  </TableCell>
                  <TableCell>
                    {key.status === "active" ? (
                      <Badge className="bg-green-100 text-green-700 hover:bg-green-100">
                        <CheckCircle2 className="w-3 h-3 mr-1" />
                        活跃
                      </Badge>
                    ) : (
                      <Badge variant="secondary">已禁用</Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-gray-600">{key.createdAt}</TableCell>
                  <TableCell className="text-gray-600">{key.lastUsed}</TableCell>
                  <TableCell className="text-right font-medium">
                    {key.requests.toLocaleString()}
                  </TableCell>
                  <TableCell className="text-right">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleDelete(key.id)}
                      className="text-red-600 hover:text-red-700 hover:bg-red-50"
                    >
                      <Trash2 className="w-4 h-4" />
                    </Button>
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