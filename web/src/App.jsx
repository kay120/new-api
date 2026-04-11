/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { lazy, Suspense, useContext, useMemo } from 'react';
import { Route, Routes, useLocation, useParams, Navigate } from 'react-router-dom';
import Loading from './components/common/ui/Loading';
import User from './pages/User';
import { AuthRedirect, PrivateRoute, AdminRoute } from './helpers';
import RegisterForm from './components/auth/RegisterForm';
import LoginForm from './components/auth/LoginForm';
import NotFound from './pages/NotFound';
import Forbidden from './pages/Forbidden';
import Setting from './pages/Setting';
import { StatusContext } from './context/Status';

import PasswordResetForm from './components/auth/PasswordResetForm';
import PasswordResetConfirm from './components/auth/PasswordResetConfirm';
import Channel from './pages/Channel';
import Token from './pages/Token';
import Redemption from './pages/Redemption';
import TopUp from './pages/TopUp';
import Log from './pages/Log';
import Chat from './pages/Chat';
import Chat2Link from './pages/Chat2Link';
import Midjourney from './pages/Midjourney';
import Pricing from './pages/Pricing';
import Task from './pages/Task';
import ModelPage from './pages/Model';
import ModelDeploymentPage from './pages/ModelDeployment';
import Playground from './pages/Playground';
import Report from './pages/Report';
import Subscription from './pages/Subscription';
import OAuth2Callback from './components/auth/OAuth2Callback';
import PersonalSetting from './components/settings/PersonalSetting';
import Setup from './pages/Setup';
import SetupCheck from './components/layout/SetupCheck';

// 设计项目组件
import DesignApp from './components/design/DesignApp';
import { Dashboard as DesignDashboard } from './design/components/Dashboard';
import { UserDashboard } from './design/components/UserDashboard';
import { ApiKeys } from './design/components/ApiKeys';
import { Models } from './design/components/Models';
import { Monitoring } from './design/components/Monitoring';
import { RateLimit } from './design/components/RateLimit';
import { Billing } from './design/components/Billing';
import { Settings } from './design/components/Settings';
import { UserManagement } from './design/components/UserManagement';
import { useUser } from './design/contexts/UserContext';

const Home = lazy(() => import('./pages/Home'));
const OldDashboard = lazy(() => import('./pages/Dashboard'));
const About = lazy(() => import('./pages/About'));
const UserAgreement = lazy(() => import('./pages/UserAgreement'));
const PrivacyPolicy = lazy(() => import('./pages/PrivacyPolicy'));

function DynamicOAuth2Callback() {
  const { provider } = useParams();
  return <OAuth2Callback type={provider} />;
}

// 根据用户角色显示不同的仪表板
function DashboardRouter() {
  const { isAdmin } = useUser();
  return isAdmin() ? <DesignDashboard /> : <UserDashboard />;
}

function App() {
  const location = useLocation();
  const [statusState] = useContext(StatusContext);

  // 获取模型广场权限配置
  const pricingRequireAuth = useMemo(() => {
    const headerNavModulesConfig = statusState?.status?.HeaderNavModules;
    if (headerNavModulesConfig) {
      try {
        const modules = JSON.parse(headerNavModulesConfig);

        // 处理向后兼容性：如果pricing是boolean，默认不需要登录
        if (typeof modules.pricing === 'boolean') {
          return false; // 默认不需要登录鉴权
        }

        // 如果是对象格式，使用requireAuth配置
        return modules.pricing?.requireAuth === true;
      } catch (error) {
        console.error('解析顶栏模块配置失败:', error);
        return false; // 默认不需要登录
      }
    }
    return false; // 默认不需要登录
  }, [statusState?.status?.HeaderNavModules]);

  return (
    <SetupCheck>
      <Routes>
        <Route path='/' element={<Navigate to='/console' replace />} />
        <Route
          path='/setup'
          element={
            <Suspense fallback={<Loading />} key={location.pathname}>
              <Setup />
            </Suspense>
          }
        />
        <Route path='/forbidden' element={<Forbidden />} />

        {/* ===== 设计项目布局路由 ===== */}
        <Route path='/console' element={<PrivateRoute><DesignApp /></PrivateRoute>}>
          <Route index element={<DashboardRouter />} />
          <Route path='users' element={<UserManagement />} />
          <Route path='api-keys' element={<ApiKeys />} />
          <Route path='models' element={<Models />} />
          <Route path='monitoring' element={<Monitoring />} />
          <Route path='rate-limit' element={<RateLimit />} />
          <Route path='billing' element={<Billing />} />
          <Route path='setting' element={<Settings />} />
        </Route>

        {/* ===== 保留的旧路由（不在设计项目布局中） ===== */}
        <Route
          path='/console/deployment'
          element={<AdminRoute><ModelDeploymentPage /></AdminRoute>}
        />
        <Route
          path='/console/subscription'
          element={<AdminRoute><Subscription /></AdminRoute>}
        />
        <Route
          path='/console/channel'
          element={<AdminRoute><Channel /></AdminRoute>}
        />
        <Route
          path='/console/token'
          element={<PrivateRoute><Token /></PrivateRoute>}
        />
        <Route
          path='/console/playground'
          element={<PrivateRoute><Playground /></PrivateRoute>}
        />
        <Route
          path='/console/redemption'
          element={<AdminRoute><Redemption /></AdminRoute>}
        />
        <Route
          path='/console/personal'
          element={<PrivateRoute><PersonalSetting /></PrivateRoute>}
        />
        <Route
          path='/console/topup'
          element={<PrivateRoute><TopUp /></PrivateRoute>}
        />
        <Route
          path='/console/log'
          element={<PrivateRoute><Log /></PrivateRoute>}
        />
        <Route
          path='/console/report'
          element={<PrivateRoute><Report /></PrivateRoute>}
        />
        <Route
          path='/console/midjourney'
          element={<PrivateRoute><Midjourney /></PrivateRoute>}
        />
        <Route
          path='/console/task'
          element={<PrivateRoute><Task /></PrivateRoute>}
        />
        <Route
          path='/console/chat/:id?'
          element={<Suspense fallback={<Loading />}><Chat /></Suspense>}
        />
        <Route
          path='/user/reset'
          element={<Suspense fallback={<Loading />}><PasswordResetConfirm /></Suspense>}
        />

        {/* ===== 认证路由 ===== */}
        <Route
          path='/login'
          element={<Suspense fallback={<Loading />}><AuthRedirect><LoginForm /></AuthRedirect></Suspense>}
        />
        <Route
          path='/register'
          element={<Suspense fallback={<Loading />}><AuthRedirect><RegisterForm /></AuthRedirect></Suspense>}
        />
        <Route
          path='/reset'
          element={<Suspense fallback={<Loading />}><PasswordResetForm /></Suspense>}
        />
        <Route
          path='/oauth/github'
          element={<Suspense fallback={<Loading />}><OAuth2Callback type='github' /></Suspense>}
        />
        <Route
          path='/oauth/discord'
          element={<Suspense fallback={<Loading />}><OAuth2Callback type='discord' /></Suspense>}
        />
        <Route
          path='/oauth/oidc'
          element={<Suspense fallback={<Loading />}><OAuth2Callback type='oidc' /></Suspense>}
        />
        <Route
          path='/oauth/linuxdo'
          element={<Suspense fallback={<Loading />}><OAuth2Callback type='linuxdo' /></Suspense>}
        />
        <Route
          path='/oauth/:provider'
          element={<Suspense fallback={<Loading />}><DynamicOAuth2Callback /></Suspense>}
        />

        {/* ===== 公开页面 ===== */}
        <Route
          path='/pricing'
          element={
            pricingRequireAuth ? (
              <PrivateRoute><Suspense fallback={<Loading />}><Pricing /></Suspense></PrivateRoute>
            ) : (
              <Suspense fallback={<Loading />}><Pricing /></Suspense>
            )
          }
        />
        <Route path='/about' element={<Suspense fallback={<Loading />}><About /></Suspense>} />
        <Route path='/user-agreement' element={<Suspense fallback={<Loading />}><UserAgreement /></Suspense>} />
        <Route path='/privacy-policy' element={<Suspense fallback={<Loading />}><PrivacyPolicy /></Suspense>} />
        <Route path='/chat2link' element={<PrivateRoute><Suspense fallback={<Loading />}><Chat2Link /></Suspense></PrivateRoute>} />
        <Route path='*' element={<NotFound />} />
      </Routes>
    </SetupCheck>
  );
}

export default App;
