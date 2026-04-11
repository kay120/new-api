import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { 
  LayoutDashboard, 
  Key, 
  Cpu, 
  Activity, 
  Gauge, 
  CreditCard,
  Settings as SettingsIcon,
  Menu,
  X,
  LogOut,
  Users
} from "lucide-react";
import { useState } from "react";
import { useUser } from "../contexts/UserContext";
import { Button } from "./ui/button";

const adminNavItems = [
  { path: "/", label: "控制台", icon: LayoutDashboard },
  { path: "/users", label: "用户管理", icon: Users },
  { path: "/api-keys", label: "API密钥", icon: Key },
  { path: "/models", label: "模型管理", icon: Cpu },
  { path: "/monitoring", label: "请求监控", icon: Activity },
  { path: "/rate-limit", label: "限流配置", icon: Gauge },
  { path: "/billing", label: "计费统计", icon: CreditCard },
  { path: "/settings", label: "系统设置", icon: SettingsIcon },
];

const userNavItems = [
  { path: "/", label: "我的控制台", icon: LayoutDashboard },
  { path: "/api-keys", label: "我的密钥", icon: Key },
  { path: "/monitoring", label: "请求日志", icon: Activity },
  { path: "/billing", label: "费用明细", icon: CreditCard },
];

export function Layout() {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const { user, logout, isAdmin } = useUser();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  const navItems = isAdmin() ? adminNavItems : userNavItems;

  return (
    <div className="flex h-screen bg-gray-50">
      {/* 侧边栏 */}
      <aside className={`fixed inset-y-0 left-0 z-50 w-64 bg-white border-r border-gray-200 transform transition-transform duration-200 ease-in-out lg:translate-x-0 lg:static lg:inset-0 ${sidebarOpen ? 'translate-x-0' : '-translate-x-full'}`}>
        <div className="flex items-center justify-between h-16 px-6 border-b border-gray-200">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
              <Cpu className="w-5 h-5 text-white" />
            </div>
            <span className="font-semibold text-gray-900">AI网关平台</span>
          </div>
          <button 
            onClick={() => setSidebarOpen(false)}
            className="lg:hidden"
          >
            <X className="w-5 h-5 text-gray-500" />
          </button>
        </div>
        
        <nav className="p-4 space-y-1">
          {navItems.map((item) => (
            <NavLink
              key={item.path}
              to={item.path}
              end={item.path === "/"}
              onClick={() => setSidebarOpen(false)}
              className={({ isActive }) =>
                `flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-colors ${
                  isActive
                    ? "bg-blue-50 text-blue-600 font-medium"
                    : "text-gray-700 hover:bg-gray-50"
                }`
              }
            >
              <item.icon className="w-5 h-5" />
              {item.label}
            </NavLink>
          ))}
          <button
            onClick={handleLogout}
            className="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-colors text-gray-700 hover:bg-gray-50 w-full text-left"
          >
            <LogOut className="w-5 h-5" />
            登出
          </button>
        </nav>
      </aside>

      {/* 遮罩层 */}
      {sidebarOpen && (
        <div 
          className="fixed inset-0 bg-black/50 z-40 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* 主内容区 */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* 顶部栏 */}
        <header className="h-16 bg-white border-b border-gray-200 flex items-center justify-between px-6">
          <button
            onClick={() => setSidebarOpen(true)}
            className="lg:hidden"
          >
            <Menu className="w-6 h-6 text-gray-700" />
          </button>
          
          <div className="flex items-center gap-4 ml-auto">
            <div className="text-sm flex items-center gap-2">
              <span className="text-gray-500">欢迎，</span>
              <span className="font-medium text-gray-900">{user?.name}</span>
              {isAdmin() && (
                <span className="px-2 py-0.5 bg-purple-100 text-purple-700 text-xs rounded">
                  管理员
                </span>
              )}
            </div>
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center">
                <span className="text-sm font-medium text-white">
                  {user?.name.charAt(0)}
                </span>
              </div>
            </div>
          </div>
        </header>

        {/* 页面内容 */}
        <main className="flex-1 overflow-auto">
          <Outlet />
        </main>
      </div>
    </div>
  );
}