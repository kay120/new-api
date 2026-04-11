import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Table,
  Tabs,
  TabPane,
  DatePicker,
  Button,
  Toast,
  Typography,
  Space,
  Select,
  Empty,
} from '@douyinfe/semi-ui';
import { VChart } from '@visactor/react-vchart';
import { isAdmin } from '../../helpers';
import { API } from '../../helpers';

const { Text } = Typography;

const Report = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState('overview');
  const [loading, setLoading] = useState(false);
  const [dateRange, setDateRange] = useState([
    Date.now() / 1000 - 7 * 86400,
    Date.now() / 1000,
  ]);
  const [groupFilter, setGroupFilter] = useState('');
  const [overviewData, setOverviewData] = useState(null);
  const [groupData, setGroupData] = useState([]);
  const [modelData, setModelData] = useState([]);
  const [userData, setUserData] = useState([]);

  const loadReportData = async () => {
    setLoading(true);
    const params = `start_timestamp=${Math.floor(dateRange[0] / 1000)}&end_timestamp=${Math.floor(dateRange[1] / 1000)}${groupFilter ? `&group=${groupFilter}` : ''}`;

    try {
      const [overviewRes, groupRes, modelRes, userRes] = await Promise.all([
        API.get(`/api/report/overview?${params}`),
        API.get(`/api/report/by-group?${params}`),
        API.get(`/api/report/by-model?${params}`),
        API.get(`/api/report/by-user?${params}`),
      ]);

      if (overviewRes.data.success) setOverviewData(overviewRes.data.data);
      if (groupRes.data.success) setGroupData(groupRes.data.data);
      if (modelRes.data.success) setModelData(modelRes.data.data);
      if (userRes.data.success) setUserData(userRes.data.data);
    } catch (err) {
      Toast.error('加载报表数据失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadReportData();
  }, []);

  const handleExport = async (type) => {
    const params = `type=${type}&start_timestamp=${Math.floor(dateRange[0] / 1000)}&end_timestamp=${Math.floor(dateRange[1] / 1000)}${groupFilter ? `&group=${groupFilter}` : ''}`;
    try {
      const response = await fetch(`/api/report/export?${params}`, {
        headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
      });
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `report_${type}.csv`;
      a.click();
      window.URL.revokeObjectURL(url);
      Toast.success('导出成功');
    } catch (err) {
      Toast.error('导出失败');
    }
  };

  const quotaToAmount = (quota) => {
    return (quota / 500000).toFixed(4);
  };

  const groupColumns = [
    { title: '组', dataIndex: 'group', width: 150 },
    { title: '消费额度', dataIndex: 'quota', render: (v) => `$${quotaToAmount(v)}`, width: 120 },
    { title: '请求次数', dataIndex: 'request_count', width: 100 },
    { title: '输入Tokens', dataIndex: 'prompt_tokens', width: 120 },
    { title: '输出Tokens', dataIndex: 'completion_tokens', width: 120 },
    { title: '使用人数', dataIndex: 'user_count', width: 100 },
  ];

  const modelColumns = [
    { title: '模型', dataIndex: 'model_name', width: 200 },
    { title: '消费额度', dataIndex: 'quota', render: (v) => `$${quotaToAmount(v)}`, width: 120 },
    { title: '请求次数', dataIndex: 'request_count', width: 100 },
    { title: '输入Tokens', dataIndex: 'prompt_tokens', width: 120 },
    { title: '输出Tokens', dataIndex: 'completion_tokens', width: 120 },
    { title: '使用人数', dataIndex: 'user_count', width: 100 },
  ];

  const userColumns = [
    { title: '用户名', dataIndex: 'username', width: 150 },
    { title: '组', dataIndex: 'group', width: 100 },
    { title: '消费额度', dataIndex: 'quota', render: (v) => `$${quotaToAmount(v)}`, width: 120 },
    { title: '请求次数', dataIndex: 'request_count', width: 100 },
    { title: '使用模型数', dataIndex: 'model_count', width: 120 },
  ];

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <Typography.Title heading={4}>{t('用量报表')}</Typography.Title>
        <Space>
          {isAdmin() && (
            <Select
              placeholder="按组筛选"
              style={{ width: 150 }}
              value={groupFilter || undefined}
              onChange={setGroupFilter}
              allowClear
            >
              {groupData.map((g) => (
                <Select.Option key={g.group} value={g.group}>
                  {g.group}
                </Select.Option>
              ))}
            </Select>
          )}
          <DatePicker
            type="dateTimeRange"
            value={dateRange}
            onChange={setDateRange}
            style={{ width: 320 }}
          />
          <Button type="primary" onClick={loadReportData} loading={loading}>
            查询
          </Button>
        </Space>
      </div>

      <Tabs
        type="line"
        activeKey={activeTab}
        onChange={setActiveTab}
        tabBarExtraContent={
          activeTab !== 'overview' && (
            <Space>
              <Button size="small" onClick={() => handleExport(activeTab === 'group' ? 'group' : activeTab === 'model' ? 'model' : 'user')}>
                导出 CSV
              </Button>
            </Space>
          )
        }
      >
        <TabPane tab="概览" itemKey="overview">
          {overviewData ? (
            <div>
              <div className="grid grid-cols-4 gap-4 mb-6">
                <div className="bg-white p-4 rounded-lg shadow">
                  <Text type="tertiary">总消费</Text>
                  <Typography.Title heading={3}>${quotaToAmount(overviewData.total_quota)}</Typography.Title>
                </div>
                <div className="bg-white p-4 rounded-lg shadow">
                  <Text type="tertiary">总请求</Text>
                  <Typography.Title heading={3}>{overviewData.total_requests.toLocaleString()}</Typography.Title>
                </div>
                <div className="bg-white p-4 rounded-lg shadow">
                  <Text type="tertiary">活跃用户</Text>
                  <Typography.Title heading={3}>{overviewData.total_users}</Typography.Title>
                </div>
                <div className="bg-white p-4 rounded-lg shadow">
                  <Text type="tertiary">使用模型</Text>
                  <Typography.Title heading={3}>{overviewData.total_models}</Typography.Title>
                </div>
              </div>

              {overviewData.daily_stats.length > 0 && (
                <div className="mb-6">
                  <Typography.Title heading={5}>每日消费趋势</Typography.Title>
                  <VChart
                    spec={{
                      type: 'line',
                      width: 800,
                      height: 300,
                      data: {
                        values: overviewData.daily_stats.map((d) => ({
                          date: d.date,
                          quota: parseFloat(quotaToAmount(d.quota)),
                        })),
                      },
                      xField: 'date',
                      yField: 'quota',
                      tooltip: { visible: true },
                    }}
                  />
                </div>
              )}

              {overviewData.top_models.length > 0 && (
                <div className="mb-6">
                  <Typography.Title heading={5}>Top 模型消费</Typography.Title>
                  <VChart
                    spec={{
                      type: 'bar',
                      width: 800,
                      height: 300,
                      data: {
                        values: overviewData.top_models.map((m) => ({
                          model: m.model_name,
                          quota: parseFloat(quotaToAmount(m.quota)),
                        })),
                      },
                      xField: 'model',
                      yField: 'quota',
                      tooltip: { visible: true },
                    }}
                  />
                </div>
              )}
            </div>
          ) : (
            <Empty description="暂无数据" />
          )}
        </TabPane>

        <TabPane tab="按组分析" itemKey="group">
          <Table
            columns={groupColumns}
            dataSource={groupData}
            loading={loading}
            pagination={{ pageSize: 20 }}
          />
        </TabPane>

        <TabPane tab="按模型分析" itemKey="model">
          <Table
            columns={modelColumns}
            dataSource={modelData}
            loading={loading}
            pagination={{ pageSize: 20 }}
          />
        </TabPane>

        <TabPane tab="用户排行" itemKey="user">
          <Table
            columns={userColumns}
            dataSource={userData}
            loading={loading}
            pagination={{ pageSize: 20 }}
          />
        </TabPane>
      </Tabs>
    </div>
  );
};

export default Report;
