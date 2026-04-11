import React, { useContext } from 'react';
import { Outlet, Navigate, useLocation } from 'react-router-dom';
import { DesignUserProvider, useUser } from '../../design/contexts/UserContext';
import { DesignLayout } from './DesignLayout';
import { UserContext } from '../../context/User';
import { Toaster } from '../../design/components/ui/sonner';

function AuthGuard({ children }) {
  const { user } = useUser();
  const location = useLocation();

  if (!user) {
    return <Navigate to='/login' state={{ from: location }} replace />;
  }

  return children;
}

export default function DesignApp() {
  const [userState] = useContext(UserContext);

  // 如果用户未登录，不渲染设计项目布局
  if (!userState?.user) {
    return <Navigate to='/login' replace />;
  }

  return (
    <DesignUserProvider>
      <AuthGuard>
        <DesignLayout>
          <Outlet />
        </DesignLayout>
        <Toaster />
      </AuthGuard>
    </DesignUserProvider>
  );
}
