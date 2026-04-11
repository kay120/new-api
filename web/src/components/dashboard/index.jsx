import React, { useContext, useEffect, useMemo } from 'react';
import { Card } from '@douyinfe/semi-ui';
import {
  AreaChart, Area, BarChart, Bar, LineChart, Line,
  XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  PieChart, Pie, Cell,
} from 'recharts';
import { TrendingUp, TrendingDown, Activity, Cpu, Clock, DollarSign, Trophy, Crown } from 'lucide-react';
import { getRelativeTime } from '../../helpers';
import { UserContext } from '../../context/User';
import { StatusContext } from '../../context/Status';
import { useDashboardData } from '../../hooks/dashboard/useDashboardData';
import { useDashboardCharts } from '../../hooks/dashboard/useDashboardCharts';
import { renderQuota, modelColorMap } from '../../helpers';
import {
  handleCopyUrl, handleSpeedTest,
  getUptimeStatusColor, getUptimeStatusText, renderMonitorList,
} from '../../helpers/dashboard';
import {
  CHART_CONFIG, CARD_PROPS, FLEX_CENTER_GAP2, ILLUSTRATION_SIZE,
  ANNOUNCEMENT_LEGEND_DATA, UPTIME_STATUS_MAP,
} from '../../constants/dashboard.constants';
import ApiInfoPanel from './ApiInfoPanel';
import AnnouncementsPanel from './AnnouncementsPanel';
import FaqPanel from './FaqPanel';
import UptimePanel from './UptimePanel';
import SearchModal from './modals/SearchModal';

const COLORS = ['#3b82f6', '#10b981', '#ec4899', '#f59e0b', '#8b5cf6', '#06b6d4', '#f97316'];

const Dashboard = () => {
  const [userState, userDispatch] = useContext(UserContext);
  const [statusState, statusDispatch] = useContext(StatusContext);

  const d = useDashboardData(userState, userDispatch, statusState);
  const c = useDashboardCharts(
    d.dataExportDefaultTime, d.setTrendData, d.setConsumeQuota,
    d.setTimes, d.setConsumeTokens, d.setPieData, d.setLineData,
    d.setModelColors, d.t,
  );

  const loadUserData = async () => {
    if (d.isAdminUser) {
      const userData = await d.loadUserQuotaData();
      if (userData?.length > 0) c.updateUserChartData(userData);
    }
  };

  const initChart = async () => {
    const data = await d.loadQuotaData();
    if (data?.length > 0) c.updateChartData(data);
    await loadUserData();
    await d.loadUptimeData();
  };

  const handleRefresh = async () => {
    const data = await d.refresh();
    if (data?.length > 0) c.updateChartData(data);
    await loadUserData();
  };

  useEffect(() => { initChart(); }, []);

  const { trendData, consumeQuota, consumeTokens, times, performanceMetrics, pieData } = d;

  // 请求趋势
  const requestData = useMemo(() =>
    (trendData.times || []).map((v, i) => ({
      time: trendData.labels?.[i] || `${i}`,
      requests: v,
    })), [trendData]);

  // 模型使用分布
  const modelUsageData = useMemo(() =>
    (pieData || []).filter(p => p.type !== 'null').map((p, i) => ({
      name: p.type,
      value: Number(p.value) || p.value,
      color: modelColorMap[p.type] || COLORS[i % COLORS.length],
    })), [pieData]);

  // 成本趋势
  const costData = useMemo(() =>
    (trendData.consumeQuota || []).map((v, i) => ({
      time: trendData.labels?.[i] || `${i}`,
      cost: v,
    })), [trendData]);

  // 模型使用趋势
  const modelTrendData = useMemo(() => {
    const values = c.spec_model_line?.data?.[0]?.values || [];
    const timeMap = {};
    values.forEach(v => {
      if (!timeMap[v.Time]) timeMap[v.Time] = { date: v.Time };
      timeMap[v.Time][v.Model] = v.Count;
    });
    return Object.values(timeMap);
  }, [c.spec_model_line]);

  // 模型排行
  const modelRankData = useMemo(() => {
    const values = c.spec_rank_bar?.data?.[0]?.values || [];
    return values.map(v => ({
      name: v.Model,
      count: v.Count,
      color: modelColorMap[v.Model] || COLORS[0],
    }));
  }, [c.spec_rank_bar]);

  // 用户排行
  const userRankData = useMemo(() => {
    const values = c.spec_user_rank?.data?.[0]?.values || [];
    const totalQuota = values.reduce((s, u) => s + u.rawQuota, 0);
    return values.sort((a, b) => b.rawQuota - a.rawQuota).map((v, i) => ({
      id: i + 1,
      name: v.User,
      cost: v.rawQuota,
      costDisplay: renderQuota(v.rawQuota, 2),
      percentage: totalQuota > 0 ? ((v.rawQuota / totalQuota) * 100).toFixed(1) : '0',
    }));
  }, [c.spec_user_rank]);

  // 指标卡
  const metrics = [
    {
      title: d.t('统计请求'),
      value: times?.toLocaleString?.() ?? '0',
      icon: <Activity className="w-4 h-4 text-blue-500" />,
      trend: '', trendUp: true, trendLabel: '',
    },
    {
      title: d.t('统计额度'),
      value: renderQuota(consumeQuota),
      icon: <DollarSign className="w-4 h-4 text-orange-500" />,
      trend: '', trendUp: true, trendLabel: '',
    },
    {
      title: d.t('统计Tokens'),
      value: isNaN(consumeTokens) ? '0' : consumeTokens.toLocaleString(),
      icon: <Cpu className="w-4 h-4 text-green-500" />,
      trend: '', trendUp: true, trendLabel: '',
    },
    {
      title: d.t('平均RPM'),
      value: performanceMetrics.avgRPM ?? 0,
      icon: <Clock className="w-4 h-4 text-purple-500" />,
      trend: `${performanceMetrics.avgTPM ?? 0} TPM`, trendUp: false, trendLabel: '',
    },
  ];

  // 公告数据
  const apiInfoData = statusState?.status?.api_info || [];
  const announcementData = (statusState?.status?.announcements || []).map((item) => {
    const pubDate = item?.publishDate ? new Date(item.publishDate) : null;
    const absoluteTime = pubDate && !isNaN(pubDate.getTime())
      ? `${pubDate.getFullYear()}-${String(pubDate.getMonth() + 1).padStart(2, '0')}-${String(pubDate.getDate()).padStart(2, '0')} ${String(pubDate.getHours()).padStart(2, '0')}:${String(pubDate.getMinutes()).padStart(2, '0')}`
      : item?.publishDate || '';
    return { ...item, time: absoluteTime, relative: getRelativeTime(item.publishDate) };
  });
  const faqData = statusState?.status?.faq || [];
  const uptimeLegendData = Object.entries(UPTIME_STATUS_MAP).map(([status, info]) => ({
    status: Number(status), color: info.color, label: d.t(info.label),
  }));

  return (
    <div className="p-6 space-y-6">
      {/* 标题 */}
      <div>
        <h1 className="text-2xl font-semibold text-gray-900">{d.t('控制台')}</h1>
        <p className="text-sm text-gray-500 mt-1">{d.t('实时监控大模型网关运行状态')}</p>
      </div>

      {/* 关键指标卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {metrics.map((m, i) => (
          <Card key={i} bodyStyle={{ padding: '16px 20px' }} style={{ border: '1px solid #f0f0f0', borderRadius: 8 }}>
            <div className="flex items-center justify-between pb-1">
              <span className="text-sm font-medium text-gray-600">{m.title}</span>
              {m.icon}
            </div>
            <div className="text-2xl font-semibold text-gray-900">{m.value}</div>
            {m.trend && (
              <div className="flex items-center gap-1 mt-1 text-sm">
                {m.trendUp
                  ? <TrendingUp className="w-4 h-4 text-green-500" />
                  : <TrendingDown className="w-4 h-4 text-green-500" />
                }
                <span className="text-green-500">{m.trend}</span>
                {m.trendLabel && <span className="text-gray-500">{m.trendLabel}</span>}
              </div>
            )}
          </Card>
        ))}
      </div>

      {/* 图表区域 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* 请求趋势 */}
        <Card title={d.t('请求趋势')} style={{ border: '1px solid #f0f0f0', borderRadius: 8 }}>
          <ResponsiveContainer width="100%" height={250}>
            <AreaChart data={requestData}>
              <defs>
                <linearGradient id="colorRequests" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis dataKey="time" stroke="#9ca3af" fontSize={12} />
              <YAxis stroke="#9ca3af" fontSize={12} />
              <Tooltip />
              <Area type="monotone" dataKey="requests" stroke="#3b82f6" fill="url(#colorRequests)" strokeWidth={2} />
            </AreaChart>
          </ResponsiveContainer>
        </Card>

        {/* 模型使用分布 */}
        <Card title={d.t('模型使用分布')} style={{ border: '1px solid #f0f0f0', borderRadius: 8 }}>
          <div className="flex items-center justify-between">
            <ResponsiveContainer width="50%" height={250}>
              <PieChart>
                <Pie data={modelUsageData} cx="50%" cy="50%" innerRadius={60} outerRadius={90} paddingAngle={2} dataKey="value">
                  {modelUsageData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
            <div className="space-y-3">
              {modelUsageData.map((item) => (
                <div key={item.name} className="flex items-center gap-3">
                  <div className="w-3 h-3 rounded-full" style={{ backgroundColor: item.color }} />
                  <span className="text-sm text-gray-600 min-w-20">{item.name}</span>
                  <span className="text-sm font-medium text-gray-900">{typeof item.value === 'number' ? item.value.toLocaleString() : item.value}</span>
                </div>
              ))}
            </div>
          </div>
        </Card>

        {/* 模型调用排行 */}
        <Card title={d.t('模型调用排行')} style={{ border: '1px solid #f0f0f0', borderRadius: 8 }}>
          <ResponsiveContainer width="100%" height={250}>
            <BarChart data={modelRankData} layout="vertical">
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis type="number" stroke="#9ca3af" fontSize={12} />
              <YAxis dataKey="name" type="category" stroke="#9ca3af" fontSize={12} width={100} />
              <Tooltip />
              <Bar dataKey="count" radius={[0, 4, 4, 0]}>
                {modelRankData.map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={entry.color} />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </Card>

        {/* 成本趋势 */}
        <Card title={d.t('额度消耗趋势')} style={{ border: '1px solid #f0f0f0', borderRadius: 8 }}>
          <ResponsiveContainer width="100%" height={250}>
            <BarChart data={costData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis dataKey="time" stroke="#9ca3af" fontSize={12} />
              <YAxis stroke="#9ca3af" fontSize={12} />
              <Tooltip />
              <Bar dataKey="cost" fill="#f59e0b" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </Card>
      </div>

      {/* 全局模型使用分析 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* 全局模型使用占比 */}
        <Card
          title={
            <div className="flex items-center justify-between w-full">
              <span className="flex items-center gap-2">
                <Cpu className="w-5 h-5 text-blue-600" />
                {d.t('全局模型使用统计')}
              </span>
            </div>
          }
          style={{ border: '1px solid #f0f0f0', borderRadius: 8 }}
        >
          <div className="flex flex-col lg:flex-row items-center gap-6">
            <div className="flex-shrink-0">
              <ResponsiveContainer width={200} height={200}>
                <PieChart>
                  <Pie data={modelUsageData} cx={100} cy={100} innerRadius={60} outerRadius={90} paddingAngle={2} dataKey="value">
                    {modelUsageData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
            </div>
            <div className="flex-1 w-full space-y-3">
              {modelUsageData.map((model) => {
                const total = modelUsageData.reduce((s, m) => s + (typeof m.value === 'number' ? m.value : 0), 0);
                const pct = total > 0 && typeof model.value === 'number'
                  ? ((model.value / total) * 100).toFixed(1) : '0';
                return (
                  <div key={model.name} className="space-y-1">
                    <div className="flex items-center justify-between text-sm">
                      <div className="flex items-center gap-2">
                        <div className="w-3 h-3 rounded-full" style={{ backgroundColor: model.color }} />
                        <span className="text-gray-700">{model.name}</span>
                      </div>
                      <span className="font-medium text-gray-900">{pct}%</span>
                    </div>
                    <div className="w-full bg-gray-100 rounded-full h-2">
                      <div className="h-2 rounded-full transition-all" style={{ width: `${pct}%`, backgroundColor: model.color }} />
                    </div>
                    <div className="text-xs text-gray-500">
                      {typeof model.value === 'number' ? model.value.toLocaleString() : model.value} {d.t('次请求')}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </Card>

        {/* 模型使用趋势 */}
        <Card
          title={
            <span className="flex items-center gap-2">
              <TrendingUp className="w-5 h-5 text-green-600" />
              {d.t('模型使用趋势')}
            </span>
          }
          style={{ border: '1px solid #f0f0f0', borderRadius: 8 }}
        >
          <ResponsiveContainer width="100%" height={250}>
            <LineChart data={modelTrendData} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis dataKey="date" stroke="#9ca3af" fontSize={12} />
              <YAxis stroke="#9ca3af" fontSize={12} />
              <Tooltip />
              {modelUsageData.map((model) => (
                <Line key={model.name} type="monotone" dataKey={model.name}
                  stroke={model.color} strokeWidth={2} dot={{ fill: model.color, r: 3 }} />
              ))}
            </LineChart>
          </ResponsiveContainer>
        </Card>
      </div>

      {/* 用户排行 - 仅管理员 */}
      {d.isAdminUser && userRankData.length > 0 && (
        <Card
          title={
            <span className="flex items-center gap-2">
              <Trophy className="w-5 h-5 text-yellow-600" />
              {d.t('用户消耗排行')}
            </span>
          }
          style={{ border: '1px solid #f0f0f0', borderRadius: 8 }}
        >
          <div className="space-y-3">
            {userRankData.map((user, index) => (
              <div key={user.id} className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors">
                <div className="flex-shrink-0 w-8 h-8 flex items-center justify-center">
                  {index === 0 ? (
                    <Crown className="w-6 h-6 text-yellow-500" />
                  ) : (
                    <div className="w-8 h-8 rounded-full bg-gradient-to-br from-orange-100 to-red-100 flex items-center justify-center">
                      <span className="text-sm font-semibold text-gray-700">#{index + 1}</span>
                    </div>
                  )}
                </div>
                <div className="w-10 h-10 rounded-full bg-gradient-to-br from-orange-500 to-red-600 flex items-center justify-center text-white font-medium">
                  {user.name.charAt(0).toUpperCase()}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="font-medium text-gray-900 truncate">{user.name}</p>
                </div>
                <div className="text-right">
                  <p className="text-lg font-semibold text-gray-900">{user.costDisplay}</p>
                  <p className="text-xs text-gray-500">{user.percentage}%</p>
                </div>
              </div>
            ))}
          </div>
        </Card>
      )}

      {/* 搜索弹窗 */}
      <SearchModal
        searchModalVisible={d.searchModalVisible}
        handleSearchConfirm={async () => {
          await d.handleSearchConfirm(c.updateChartData);
          await loadUserData();
        }}
        handleCloseModal={d.handleCloseModal}
        isMobile={d.isMobile}
        isAdminUser={d.isAdminUser}
        inputs={d.inputs}
        dataExportDefaultTime={d.dataExportDefaultTime}
        timeOptions={d.timeOptions}
        handleInputChange={d.handleInputChange}
        t={d.t}
      />

      {/* 系统信息面板 */}
      {d.hasInfoPanels && (
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-5">
          {d.announcementsEnabled && (
            <AnnouncementsPanel announcementData={announcementData}
              announcementLegendData={ANNOUNCEMENT_LEGEND_DATA.map((item) => ({ ...item, label: d.t(item.label) }))}
              CARD_PROPS={CARD_PROPS} ILLUSTRATION_SIZE={ILLUSTRATION_SIZE} t={d.t} />
          )}
          {d.faqEnabled && (
            <FaqPanel faqData={faqData} CARD_PROPS={CARD_PROPS} FLEX_CENTER_GAP2={FLEX_CENTER_GAP2}
              ILLUSTRATION_SIZE={ILLUSTRATION_SIZE} t={d.t} />
          )}
          {d.uptimeEnabled && (
            <UptimePanel uptimeData={d.uptimeData} uptimeLoading={d.uptimeLoading}
              activeUptimeTab={d.activeUptimeTab} setActiveUptimeTab={d.setActiveUptimeTab}
              loadUptimeData={d.loadUptimeData} uptimeLegendData={uptimeLegendData}
              renderMonitorList={(monitors) => renderMonitorList(monitors,
                (status) => getUptimeStatusColor(status, UPTIME_STATUS_MAP),
                (status) => getUptimeStatusText(status, UPTIME_STATUS_MAP, d.t), d.t)}
              CARD_PROPS={CARD_PROPS} ILLUSTRATION_SIZE={ILLUSTRATION_SIZE} t={d.t} />
          )}
        </div>
      )}
    </div>
  );
};

export default Dashboard;
