import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Label } from "./ui/label";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "./ui/table";
import { Badge } from "./ui/badge";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "./ui/dialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./ui/select";
import { Plus, Edit, Trash2, Users, Shield } from "lucide-react";
import { toast } from "sonner";

interface UserData {
  id: string;
  name: string;
  email: string;
  role: "admin" | "user";
  status: "active" | "disabled";
  apiKeys: number;
  requests: number;
  createdAt: string;
  lastLogin: string;
}

const initialUsers: UserData[] = [
  {
    id: "1",
    name: "管理员",
    email: "admin@example.com",
    role: "admin",
    status: "active",
    apiKeys: 3,
    requests: 125678,
    createdAt: "2025-01-15",
    lastLogin: "2026-04-07 14:23",
  },
  {
    id: "2",
    name: "张三",
    email: "user@example.com",
    role: "user",
    status: "active",
    apiKeys: 2,
    requests: 45231,
    createdAt: "2025-03-20",
    lastLogin: "2026-04-07 10:15",
  },
  {
    id: "3",
    name: "李四",
    email: "lisi@example.com",
    role: "user",
    status: "active",
    apiKeys: 1,
    requests: 12456,
    createdAt: "2025-05-10",
    lastLogin: "2026-04-06 16:42",
  },
  {
    id: "4",
    name: "王五",
    email: "wangwu@example.com",
    role: "user",
    status: "disabled",
    apiKeys: 1,
    requests: 8934,
    createdAt: "2025-08-05",
    lastLogin: "2026-03-28 09:12",
  },
];

export function UserManagement() {
  const [users, setUsers] = useState<UserData[]>(initialUsers);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [formData, setFormData] = useState({
    name: "",
    email: "",
    role: "user" as "admin" | "user",
    password: "",
  });

  const handleDelete = (id: string) => {
    setUsers(users.filter(u => u.id !== id));
    toast.success("用户已删除");
  };

  const handleCreate = () => {
    if (!formData.name || !formData.email || !formData.password) {
      toast.error("请填写完整信息");
      return;
    }

    const newUser: UserData = {
      id: Date.now().toString(),
      name: formData.name,
      email: formData.email,
      role: formData.role,
      status: "active",
      apiKeys: 0,
      requests: 0,
      createdAt: new Date().toLocaleDateString('zh-CN'),
      lastLogin: "-",
    };

    setUsers([...users, newUser]);
    setFormData({ name: "", email: "", role: "user", password: "" });
    setDialogOpen(false);
    toast.success("用户创建成功");
  };

  const totalUsers = users.length;
  const activeUsers = users.filter(u => u.status === "active").length;
  const adminUsers = users.filter(u => u.role === "admin").length;
  const totalRequests = users.reduce((sum, u) => sum + u.requests, 0);

  return (
    <div className="p-6 space-y-6">
      {/* 标题 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">用户管理</h1>
          <p className="text-sm text-gray-500 mt-1">管理平台用户和权限</p>
        </div>
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger asChild>
            <Button className="gap-2">
              <Plus className="w-4 h-4" />
              添加用户
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>创建新用户</DialogTitle>
            </DialogHeader>
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="userName">用户名</Label>
                <Input
                  id="userName"
                  placeholder="例如：张三"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="userEmail">邮箱</Label>
                <Input
                  id="userEmail"
                  type="email"
                  placeholder="user@example.com"
                  value={formData.email}
                  onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="userPassword">密码</Label>
                <Input
                  id="userPassword"
                  type="password"
                  placeholder="设置初始密码"
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="userRole">角色</Label>
                <Select
                  value={formData.role}
                  onValueChange={(value: "admin" | "user") =>
                    setFormData({ ...formData, role: value })
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="user">普通用户</SelectItem>
                    <SelectItem value="admin">管理员</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => setDialogOpen(false)}>
                取消
              </Button>
              <Button onClick={handleCreate}>创建</Button>
            </div>
          </DialogContent>
        </Dialog>
      </div>

      {/* 统计概览 */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center">
                <Users className="w-5 h-5 text-blue-600" />
              </div>
              <div>
                <div className="text-sm text-gray-600">总用户数</div>
                <div className="text-2xl font-semibold text-gray-900">{totalUsers}</div>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center">
                <Users className="w-5 h-5 text-green-600" />
              </div>
              <div>
                <div className="text-sm text-gray-600">活跃用户</div>
                <div className="text-2xl font-semibold text-green-600">{activeUsers}</div>
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
                <div className="text-sm text-gray-600">管理员</div>
                <div className="text-2xl font-semibold text-purple-600">{adminUsers}</div>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-orange-100 rounded-lg flex items-center justify-center">
                <Users className="w-5 h-5 text-orange-600" />
              </div>
              <div>
                <div className="text-sm text-gray-600">总请求量</div>
                <div className="text-2xl font-semibold text-gray-900">
                  {totalRequests.toLocaleString()}
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* 用户列表 */}
      <Card>
        <CardHeader>
          <CardTitle>用户列表</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>用户名</TableHead>
                <TableHead>邮箱</TableHead>
                <TableHead>角色</TableHead>
                <TableHead>状态</TableHead>
                <TableHead className="text-right">API密钥</TableHead>
                <TableHead className="text-right">请求次数</TableHead>
                <TableHead>创建时间</TableHead>
                <TableHead>最后登录</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {users.map((user) => (
                <TableRow key={user.id}>
                  <TableCell className="font-medium">{user.name}</TableCell>
                  <TableCell className="text-gray-600">{user.email}</TableCell>
                  <TableCell>
                    {user.role === "admin" ? (
                      <Badge className="bg-purple-100 text-purple-700">
                        <Shield className="w-3 h-3 mr-1" />
                        管理员
                      </Badge>
                    ) : (
                      <Badge variant="outline">普通用户</Badge>
                    )}
                  </TableCell>
                  <TableCell>
                    {user.status === "active" ? (
                      <Badge className="bg-green-100 text-green-700">活跃</Badge>
                    ) : (
                      <Badge variant="secondary">已禁用</Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-right font-medium">
                    {user.apiKeys}
                  </TableCell>
                  <TableCell className="text-right font-medium">
                    {user.requests.toLocaleString()}
                  </TableCell>
                  <TableCell className="text-gray-600">{user.createdAt}</TableCell>
                  <TableCell className="text-gray-600">{user.lastLogin}</TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-2">
                      <Button variant="ghost" size="sm">
                        <Edit className="w-4 h-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleDelete(user.id)}
                        className="text-red-600 hover:text-red-700 hover:bg-red-50"
                      >
                        <Trash2 className="w-4 h-4" />
                      </Button>
                    </div>
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
