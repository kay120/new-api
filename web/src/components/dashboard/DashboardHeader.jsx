import React from 'react';
import { Button } from '@douyinfe/semi-ui';
import { RefreshCw } from 'lucide-react';

const DashboardHeader = ({ getGreeting, refresh, loading, t }) => {
  return (
    <div className='flex items-center justify-between'>
      <div>
        <h1 className='text-lg font-bold text-gray-900'>{t('控制台')}</h1>
        <p className='text-xs text-gray-400 mt-0.5'>{getGreeting?.replace(/👋/g, '')}</p>
      </div>
      <Button
        theme='light'
        size='small'
        icon={<RefreshCw size={13} />}
        onClick={refresh}
        loading={loading}
      />
    </div>
  );
};

export default DashboardHeader;
