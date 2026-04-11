import { useMemo } from 'react';
import {
  IconHistogram,
  IconSend,
  IconCoinMoneyStroked,
  IconTextStroked,
  IconPulse,
  IconStopwatchStroked,
  IconTypograph,
} from '@douyinfe/semi-icons';
import { renderQuota } from '../../helpers';

export const useDashboardStats = (
  userState,
  consumeQuota,
  consumeTokens,
  times,
  trendData,
  performanceMetrics,
  t,
) => {
  const groupedStatsData = useMemo(
    () => [
      {
        title: t('历史消耗'),
        value: renderQuota(userState?.user?.used_quota),
        icon: 'IconHistogram',
        iconColor: '#6366f1',
        bgColor: '#eef2ff',
      },
      {
        title: t('请求次数'),
        value: userState.user?.request_count?.toLocaleString?.() ?? 0,
        icon: 'IconSend',
        iconColor: '#10b981',
        bgColor: '#ecfdf5',
      },
      {
        title: t('统计次数'),
        value: times?.toLocaleString?.() ?? 0,
        icon: 'IconPulse',
        iconColor: '#06b6d4',
        bgColor: '#ecfeff',
      },
      {
        title: t('统计额度'),
        value: renderQuota(consumeQuota),
        icon: 'IconCoinMoneyStroked',
        iconColor: '#f59e0b',
        bgColor: '#fffbeb',
      },
      {
        title: t('统计Tokens'),
        value: isNaN(consumeTokens) ? '0' : consumeTokens.toLocaleString(),
        icon: 'IconTextStroked',
        iconColor: '#ec4899',
        bgColor: '#fdf2f8',
      },
      {
        title: t('平均RPM'),
        value: performanceMetrics.avgRPM,
        icon: 'IconStopwatchStroked',
        iconColor: '#6366f1',
        bgColor: '#eef2ff',
      },
      {
        title: t('平均TPM'),
        value: performanceMetrics.avgTPM,
        icon: 'IconTypograph',
        iconColor: '#f97316',
        bgColor: '#fff7ed',
      },
    ],
    [
      userState?.user?.used_quota,
      userState?.user?.request_count,
      times,
      consumeQuota,
      consumeTokens,
      trendData,
      performanceMetrics,
      t,
    ],
  );

  return {
    groupedStatsData,
  };
};
