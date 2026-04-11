import React from 'react';
import { Tabs, TabPane } from '@douyinfe/semi-ui';
import { VChart } from '@visactor/react-vchart';
import { BarChart3, TrendingUp, PieChart, Trophy, Users, Activity } from 'lucide-react';

const TAB_CONFIG = [
  { key: '1', label: '消耗分布', icon: BarChart3 },
  { key: '2', label: '调用趋势', icon: TrendingUp },
  { key: '3', label: '次数分布', icon: PieChart },
  { key: '4', label: '次数排行', icon: Trophy },
  { key: '5', label: '用户排行', icon: Users, admin: true },
  { key: '6', label: '用户趋势', icon: Activity, admin: true },
];

const specMap = {
  '1': 'spec_line',
  '2': 'spec_model_line',
  '3': 'spec_pie',
  '4': 'spec_rank_bar',
  '5': 'spec_user_rank',
  '6': 'spec_user_trend',
};

const ChartCard = ({ title, spec, CHART_CONFIG }) => (
  <div className='bg-white rounded-lg border border-gray-100 overflow-hidden'>
    <div className='px-4 py-3 border-b border-gray-50'>
      <h3 className='text-sm font-medium text-gray-700'>{title}</h3>
    </div>
    <div className='h-80 p-2'>
      <VChart spec={spec} option={CHART_CONFIG} />
    </div>
  </div>
);

const ChartsPanel = ({
  activeChartTab,
  setActiveChartTab,
  spec_line,
  spec_model_line,
  spec_pie,
  spec_rank_bar,
  spec_user_rank,
  spec_user_trend,
  isAdminUser,
  CHART_CONFIG,
  t,
}) => {
  const specs = { spec_line, spec_model_line, spec_pie, spec_rank_bar, spec_user_rank, spec_user_trend };

  const visibleTabs = TAB_CONFIG.filter(
    (tab) => !tab.admin || isAdminUser
  );

  const activeTabConfig = TAB_CONFIG.find((tab) => tab.key === activeChartTab);
  const activeSpec = specs[specMap[activeChartTab]];

  return (
    <div className='space-y-3'>
      {/* Tab 导航 */}
      <div className='flex items-center gap-1 bg-gray-50 rounded-lg p-1'>
        {visibleTabs.map((tab) => {
          const Icon = tab.icon;
          const isActive = activeChartTab === tab.key;
          return (
            <button
              key={tab.key}
              onClick={() => setActiveChartTab(tab.key)}
              className={`flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-medium transition-colors ${
                isActive
                  ? 'bg-white text-gray-900 shadow-sm'
                  : 'text-gray-500 hover:text-gray-700'
              }`}
            >
              <Icon size={13} />
              {t(tab.label)}
            </button>
          );
        })}
      </div>

      {/* 图表 */}
      <ChartCard
        title={t(activeTabConfig?.label || '')}
        spec={activeSpec}
        CHART_CONFIG={CHART_CONFIG}
      />
    </div>
  );
};

export default ChartsPanel;
