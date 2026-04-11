import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Label } from "./ui/label";
import { Switch } from "./ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./ui/tabs";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./ui/select";
import { Textarea } from "./ui/textarea";
import { Shield, Bell, Key, Settings as SettingsIcon, Globe, Database } from "lucide-react";
import { toast } from "sonner";

export function Settings() {
  const handleSave = () => {
    toast.success("设置已保存");
  };

  return (
    <div className="p-6 space-y-6">
      {/* 标题 */}
      <div>
        <h1 className="text-2xl font-semibold text-gray-900">系统设置</h1>
        <p className="text-sm text-gray-500 mt-1">配置网关平台的各项参数</p>
      </div>

      <Tabs defaultValue="general" className="space-y-6">
        <TabsList className="grid w-full max-w-md grid-cols-4">
          <TabsTrigger value="general">常规</TabsTrigger>
          <TabsTrigger value="security">安全</TabsTrigger>
          <TabsTrigger value="notification">通知</TabsTrigger>
          <TabsTrigger value="advanced">高级</TabsTrigger>
        </TabsList>

        {/* 常规设置 */}
        <TabsContent value="general" className="space-y-4">
          <Card>
            <CardHeader>
              <div className="flex items-center gap-2">
                <SettingsIcon className="w-5 h-5 text-blue-500" />
                <CardTitle>基础配置</CardTitle>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="platformName">平台名称</Label>
                <Input id="platformName" defaultValue="AI网关平台" />
              </div>

              <div className="space-y-2">
                <Label htmlFor="adminEmail">管理员邮箱</Label>
                <Input id="adminEmail" type="email" defaultValue="admin@example.com" />
              </div>

              <div className="space-y-2">
                <Label htmlFor="timezone">时区</Label>
                <Select defaultValue="asia-shanghai">
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="asia-shanghai">Asia/Shanghai (UTC+8)</SelectItem>
                    <SelectItem value="utc">UTC (UTC+0)</SelectItem>
                    <SelectItem value="america-new-york">America/New_York (UTC-5)</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="language">语言</Label>
                <Select defaultValue="zh-CN">
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="zh-CN">简体中文</SelectItem>
                    <SelectItem value="zh-TW">繁體中文</SelectItem>
                    <SelectItem value="en-US">English</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label>自动更新</Label>
                  <p className="text-sm text-gray-500">自动安装安全更新和bug修复</p>
                </div>
                <Switch defaultChecked />
              </div>

              <Button onClick={handleSave}>保存更改</Button>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <div className="flex items-center gap-2">
                <Globe className="w-5 h-5 text-green-500" />
                <CardTitle>域名与访问</CardTitle>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="apiDomain">API域名</Label>
                <Input id="apiDomain" defaultValue="https://api.example.com" />
              </div>

              <div className="space-y-2">
                <Label htmlFor="corsOrigins">CORS允许的源</Label>
                <Textarea
                  id="corsOrigins"
                  defaultValue="https://app.example.com&#10;https://dashboard.example.com"
                  rows={3}
                  placeholder="每行一个域名"
                />
              </div>

              <Button onClick={handleSave}>保存更改</Button>
            </CardContent>
          </Card>
        </TabsContent>

        {/* 安全设置 */}
        <TabsContent value="security" className="space-y-4">
          <Card>
            <CardHeader>
              <div className="flex items-center gap-2">
                <Shield className="w-5 h-5 text-red-500" />
                <CardTitle>访问控制</CardTitle>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label>启用IP白名单</Label>
                  <p className="text-sm text-gray-500">只允许白名单中的IP访问</p>
                </div>
                <Switch />
              </div>

              <div className="space-y-2">
                <Label htmlFor="ipWhitelist">IP白名单</Label>
                <Textarea
                  id="ipWhitelist"
                  placeholder="每行一个IP或IP段，例如：192.168.1.1 或 192.168.1.0/24"
                  rows={4}
                />
              </div>

              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label>强制HTTPS</Label>
                  <p className="text-sm text-gray-500">自动重定向HTTP到HTTPS</p>
                </div>
                <Switch defaultChecked />
              </div>

              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label>API密钥加密存储</Label>
                  <p className="text-sm text-gray-500">使用AES-256加密存储密钥</p>
                </div>
                <Switch defaultChecked />
              </div>

              <Button onClick={handleSave}>保存更改</Button>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <div className="flex items-center gap-2">
                <Key className="w-5 h-5 text-purple-500" />
                <CardTitle>认证设置</CardTitle>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="tokenExpiry">Token过期时间（小时）</Label>
                <Input id="tokenExpiry" type="number" defaultValue="24" />
              </div>

              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label>启用双因素认证</Label>
                  <p className="text-sm text-gray-500">要求管理员使用2FA登录</p>
                </div>
                <Switch />
              </div>

              <div className="space-y-2">
                <Label htmlFor="sessionTimeout">会话超时（分钟）</Label>
                <Input id="sessionTimeout" type="number" defaultValue="30" />
              </div>

              <Button onClick={handleSave}>保存更改</Button>
            </CardContent>
          </Card>
        </TabsContent>

        {/* 通知设置 */}
        <TabsContent value="notification" className="space-y-4">
          <Card>
            <CardHeader>
              <div className="flex items-center gap-2">
                <Bell className="w-5 h-5 text-orange-500" />
                <CardTitle>告警通知</CardTitle>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label>请求失败告警</Label>
                  <p className="text-sm text-gray-500">当失败率超过阈值时发送通知</p>
                </div>
                <Switch defaultChecked />
              </div>

              <div className="space-y-2">
                <Label htmlFor="errorThreshold">失败率阈值（%）</Label>
                <Input id="errorThreshold" type="number" defaultValue="5" />
              </div>

              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label>费用告警</Label>
                  <p className="text-sm text-gray-500">当费用超过预算时发送通知</p>
                </div>
                <Switch defaultChecked />
              </div>

              <div className="space-y-2">
                <Label htmlFor="budgetLimit">月度预算（元）</Label>
                <Input id="budgetLimit" type="number" defaultValue="20000" />
              </div>

              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label>延迟告警</Label>
                  <p className="text-sm text-gray-500">当平均延迟过高时发送通知</p>
                </div>
                <Switch defaultChecked />
              </div>

              <div className="space-y-2">
                <Label htmlFor="latencyThreshold">延迟阈值（毫秒）</Label>
                <Input id="latencyThreshold" type="number" defaultValue="3000" />
              </div>

              <Button onClick={handleSave}>保存更改</Button>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <div className="flex items-center gap-2">
                <Bell className="w-5 h-5 text-blue-500" />
                <CardTitle>通知渠道</CardTitle>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label>邮件通知</Label>
                  <p className="text-sm text-gray-500">通过邮件接收告警</p>
                </div>
                <Switch defaultChecked />
              </div>

              <div className="space-y-2">
                <Label htmlFor="emailRecipients">接收邮箱</Label>
                <Input id="emailRecipients" defaultValue="admin@example.com" />
              </div>

              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label>Webhook通知</Label>
                  <p className="text-sm text-gray-500">发送到自定义Webhook地址</p>
                </div>
                <Switch />
              </div>

              <div className="space-y-2">
                <Label htmlFor="webhookUrl">Webhook URL</Label>
                <Input id="webhookUrl" placeholder="https://hooks.example.com/alerts" />
              </div>

              <Button onClick={handleSave}>保存更改</Button>
            </CardContent>
          </Card>
        </TabsContent>

        {/* 高级设置 */}
        <TabsContent value="advanced" className="space-y-4">
          <Card>
            <CardHeader>
              <div className="flex items-center gap-2">
                <Database className="w-5 h-5 text-indigo-500" />
                <CardTitle>性能优化</CardTitle>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label>启用请求缓存</Label>
                  <p className="text-sm text-gray-500">缓存相同的请求以提高响应速度</p>
                </div>
                <Switch defaultChecked />
              </div>

              <div className="space-y-2">
                <Label htmlFor="cacheTTL">缓存过期时间（秒）</Label>
                <Input id="cacheTTL" type="number" defaultValue="300" />
              </div>

              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label>启用请求压缩</Label>
                  <p className="text-sm text-gray-500">使用gzip压缩响应数据</p>
                </div>
                <Switch defaultChecked />
              </div>

              <div className="space-y-2">
                <Label htmlFor="maxConcurrent">最大并发请求数</Label>
                <Input id="maxConcurrent" type="number" defaultValue="1000" />
              </div>

              <Button onClick={handleSave}>保存更改</Button>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <div className="flex items-center gap-2">
                <Database className="w-5 h-5 text-green-500" />
                <CardTitle>日志配置</CardTitle>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="logLevel">日志级别</Label>
                <Select defaultValue="info">
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="debug">Debug</SelectItem>
                    <SelectItem value="info">Info</SelectItem>
                    <SelectItem value="warn">Warning</SelectItem>
                    <SelectItem value="error">Error</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="logRetention">日志保留天数</Label>
                <Input id="logRetention" type="number" defaultValue="30" />
              </div>

              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label>记录请求body</Label>
                  <p className="text-sm text-gray-500">在日志中包含完整的请求内容</p>
                </div>
                <Switch />
              </div>

              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label>记录响应body</Label>
                  <p className="text-sm text-gray-500">在日志中包含完整的响应内容</p>
                </div>
                <Switch />
              </div>

              <Button onClick={handleSave}>保存更改</Button>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
