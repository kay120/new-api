import React from 'react';
import { Skeleton } from '@douyinfe/semi-ui';
import {
  IconHistogram,
  IconSend,
  IconCoinMoneyStroked,
  IconTextStroked,
  IconPulse,
  IconStopwatchStroked,
  IconTypograph,
} from '@douyinfe/semi-icons';

const iconMap = {
  IconHistogram,
  IconSend,
  IconCoinMoneyStroked,
  IconTextStroked,
  IconPulse,
  IconStopwatchStroked,
  IconTypograph,
};

const MetricCard = ({ item, loading }) => (
  <div className='bg-white rounded-lg border border-gray-100 px-4 py-3.5 hover:border-gray-200 transition-colors'>
    <div className='flex items-center gap-2.5 mb-2'>
      <div
        className='w-7 h-7 rounded-md flex items-center justify-center'
        style={{ backgroundColor: item.bgColor }}
      >
        {React.createElement(iconMap[item.icon] || IconSend, {
          style: { color: item.iconColor, fontSize: 14 },
        })}
      </div>
      <span className='text-xs text-gray-400'>{item.title}</span>
    </div>
    <div className='text-xl font-bold text-gray-900'>
      <Skeleton
        loading={loading}
        active
        placeholder={<Skeleton.Paragraph active rows={1} style={{ width: 80, height: 22, marginTop: 2 }} />}
      >
        {item.value}
      </Skeleton>
    </div>
  </div>
);

const StatsCards = ({ groupedStatsData, loading }) => {
  return (
    <div className='grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-7 gap-3'>
      {groupedStatsData.map((item, idx) => (
        <MetricCard key={idx} item={item} loading={loading} />
      ))}
    </div>
  );
};

export default StatsCards;
