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

import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Button, Card, Empty, Space, Table, Typography } from '@douyinfe/semi-ui';
import { Segmented } from 'antd';
import { useTranslation } from 'react-i18next';
import { API, isAdmin } from '../../helpers';
import { useBillingFeatures } from '../../hooks/billing/useBillingFeatures';

const WINDOW_PRESETS = [
  { labelKey: '1小时', value: '1h' },
  { labelKey: '24小时', value: '24h' },
  { labelKey: '7天', value: '7d' },
  { labelKey: '30天', value: '30d' },
  { labelKey: '365天', value: '365d' },
  { labelKey: '全部', value: 'all_time' },
];

const LeaderboardPage = () => {
  const { t } = useTranslation();
  const { config, loading: configLoading, error: billingError, refresh } = useBillingFeatures();
  const [activeWindow, setActiveWindow] = useState('24h');
  const [loading, setLoading] = useState(false);
  const [entries, setEntries] = useState([]);
  const [leaderboardError, setLeaderboardError] = useState(null);
  const [admin] = useState(() => isAdmin());

  const billingEnabled = config?.billing?.enabled;

  const fetchLeaderboard = useCallback(async () => {
    setLoading(true);
    try {
      const endpoint = admin ? '/api/leaderboard' : '/api/public/leaderboard';
      const res = await API.get(endpoint, {
        params: { window: activeWindow },
        skipErrorHandler: true,
      });
      if (res?.data?.success) {
        setEntries(Array.isArray(res.data.data) ? res.data.data : []);
        setLeaderboardError(null);
      } else {
        setEntries([]);
        setLeaderboardError(res?.data?.message || t('排行榜暂不可用'));
      }
    } catch (err) {
      setEntries([]);
      setLeaderboardError(err?.response?.data?.message || err?.message || t('排行榜暂不可用'));
    } finally {
      setLoading(false);
    }
  }, [activeWindow, admin, t]);

  useEffect(() => {
    if (!billingEnabled || billingError) {
      return;
    }
    fetchLeaderboard();
  }, [fetchLeaderboard, billingEnabled, billingError]);

  const columns = useMemo(() => {
    return [
      {
        title: t('排名'),
        dataIndex: 'rank',
        width: 80,
        render: (text, _, index) => text ?? index + 1,
      },
      {
        title: t('模型'),
        dataIndex: 'model',
        render: (text) => <Typography.Text code>{text}</Typography.Text>,
      },
      {
        title: t('请求数'),
        dataIndex: 'request_count',
        sorter: true,
        render: (value) => Number(value || 0).toLocaleString(),
      },
      {
        title: t('总Tokens'),
        dataIndex: 'token_count',
        sorter: true,
        render: (value) => Number(value || 0).toLocaleString(),
      },
      {
        title: t('唯一用户数'),
        dataIndex: 'unique_users',
        render: (value) => Number(value || 0).toLocaleString(),
      },
      {
        title: t('唯一令牌数'),
        dataIndex: 'unique_tokens',
        render: (value) => Number(value || 0).toLocaleString(),
      },
    ];
  }, [t]);

  const exportUrl = useMemo(() => {
    if (!admin) {
      return null;
    }
    const params = new URLSearchParams({ window: activeWindow });
    return `/api/leaderboard/export?${params.toString()}`;
  }, [admin, activeWindow]);

  if (configLoading) {
    return (
      <div className='mt-[60px] px-2 pb-6'>
        <Card loading></Card>
      </div>
    );
  }

  if (billingError) {
    return (
      <div className='mt-[60px] px-2 pb-6'>
        <Card bordered>
          <Space direction='vertical' align='center' className='w-full'>
            <Empty description={t('计费配置加载失败，暂时无法展示排行榜')} />
            <Button theme='solid' type='primary' onClick={() => refresh()}>
              {t('重试')}
            </Button>
          </Space>
        </Card>
      </div>
    );
  }

  if (!billingEnabled) {
    return (
      <div className='mt-[60px] px-2 pb-6'>
        <Empty description={t('排行榜功能未启用')} />
      </div>
    );
  }

  return (
    <div className='mt-[60px] px-2 pb-6 space-y-4'>
      <Card bordered>
        <Space
          wrap
          align='center'
          spacing='medium'
          className='w-full justify-between flex-col md:flex-row'
        >
          <div>
            <Typography.Title heading={4}>{t('模型请求排行榜')}</Typography.Title>
            <Typography.Text type='tertiary'>
              {t('查看不同时间窗口内各模型的请求次数与消耗情况。')}
            </Typography.Text>
          </div>
          <Space align='center'>
            <Segmented
              size='large'
              value={activeWindow}
              onChange={setActiveWindow}
              options={WINDOW_PRESETS.map((preset) => ({
                label: t(preset.labelKey),
                value: preset.value,
              }))}
            />
            {exportUrl ? (
              <Button
                theme='solid'
                type='tertiary'
                onClick={() => {
                  window.open(exportUrl, '_blank');
                }}
              >
                {t('导出 CSV')}
              </Button>
            ) : null}
          </Space>
        </Space>
      </Card>

      <Card bordered>
        {leaderboardError ? (
          <Empty description={leaderboardError} />
        ) : (
          <Table
            loading={loading}
            columns={columns}
            dataSource={entries}
            pagination={{ pageSize: 15 }}
            empty={<Empty description={t('暂无数据')} />}
            onChange={(pagination, filters, sorter) => {
              if (sorter?.field) {
                const sorted = [...entries].sort((a, b) => {
                  const first = Number(a[sorter.field]) || 0;
                  const second = Number(b[sorter.field]) || 0;
                  return sorter.order === 'descend' ? second - first : first - second;
                });
                setEntries(sorted);
              }
            }}
          />
        )}
      </Card>
    </div>
  );
};

export default LeaderboardPage;
