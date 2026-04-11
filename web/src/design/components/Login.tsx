import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useUser } from "../contexts/UserContext";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Input } from "./ui/input";
import { Label } from "./ui/label";
import { Button } from "./ui/button";
import { Cpu, Lock, Mail } from "lucide-react";
import { toast } from "sonner";

export function Login() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const { login } = useUser();
  const navigate = useNavigate();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!email || !password) {
      toast.error("请输入邮箱和密码");
      return;
    }

    const success = login(email, password);
    if (success) {
      toast.success("登录成功");
      navigate("/");
    } else {
      toast.error("邮箱或密码错误");
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-purple-50 flex items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-4">
          <div className="flex justify-center">
            <div className="w-16 h-16 bg-gradient-to-br from-blue-500 to-purple-600 rounded-2xl flex items-center justify-center">
              <Cpu className="w-10 h-10 text-white" />
            </div>
          </div>
          <CardTitle className="text-center text-2xl">
            AI网关平台
          </CardTitle>
          <p className="text-center text-sm text-gray-500">
            登录到您的账户
          </p>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="email">邮箱</Label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
                <Input
                  id="email"
                  type="email"
                  placeholder="your@email.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="pl-9"
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="password">密码</Label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
                <Input
                  id="password"
                  type="password"
                  placeholder="••••••••"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="pl-9"
                />
              </div>
            </div>

            <Button type="submit" className="w-full">
              登录
            </Button>
          </form>

          <div className="mt-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
            <p className="text-sm font-medium text-blue-900 mb-2">测试账户：</p>
            <div className="space-y-1 text-xs text-blue-700">
              <p><strong>管理员：</strong> admin@example.com / admin123</p>
              <p className="text-[10px] text-blue-600 ml-4">→ 可查看所有用户数据和系统管理功能</p>
              <p className="mt-2"><strong>普通用户：</strong> user@example.com / user123</p>
              <p className="text-[10px] text-blue-600 ml-4">→ 查看个人数据、模型使用占比和密钥详情</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}