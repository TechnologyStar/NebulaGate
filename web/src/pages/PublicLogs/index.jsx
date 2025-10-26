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
import {
  Button,
  Card,
  Empty,
  Form,
  Pagination,
  Space,
  Table,
  Tag,
  Typography,
} from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import { API, timestamp2string } from '../../helpers';
import { useBillingFeatures } from '../../hooks/billing/useBillingFeatures';

const PublicLogsPage = () => {
  const { t } = useTranslation();
  const { config, loading: configLoading, error, refresh } = useBillingFeatures();

  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [pageSize, setPageSize] = useState(20);
  const [filters, setFilters] = useState({ window: '24h', model: '', search: '' });
  const [models, setModels] = useState([]);
  const [exporting, setExporting] = useState(false);

  const publicLogsEnabled = config?.public_logs?.enabled;

  const loadModels = useCallback(async () => {
    try {
      const res = await API.get('/api/public/logs/models', { skipErrorHandler: true });
      if (res?.data?.success) {
        setModels(Array.isArray(res.data.data) ? res.data.data : []);
      }
    } catch (_) {
      setModels([]);
    }
  }, []);

  const loadLogs = useCallback(async () => {
    if (!publicLogsEnabled) {
      return;
    }
    setLoading(true);
    try {
      const res = await API.get('/api/public/logs', {
        params: {
          page,
          page_size: pageSize,
          window: filters.window,
          model: filters.model,
          search: filters.search,
        },
        skipErrorHandler: true,
      });
      if (res?.data?.success) {
        const payload = res.data.data || {};
        setLogs(Array.isArray(payload.items) ? payload.items : []);
        setTotal(payload.total || 0);
      } else {
        setLogs([]);
        setTotal(0);
      }
    } catch (error) {
      setLogs([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  }, [filters, page, pageSize, publicLogsEnabled]);

  useEffect(() => {
    if (publicLogsEnabled) {
      loadModels();
    }
  }, [publicLogsEnabled, loadModels]);

  useEffect(() => {
    if (!publicLogsEnabled) {
      return;
    }
    loadLogs();
  }, [loadLogs, publicLogsEnabled]);

  const columns = useMemo(
    () => [
      {
        title: t('时间'),
        dataIndex: 'created_at',
        render: (text) => (text ? timestamp2string(text) : '-'),
        width: 160,
      },
      {
        title: t('匿名用户'),
        dataIndex: 'subject_label',
        render: (text) => <Typography.Text>{text || t('匿名用户')}</Typography.Text>,
      },
      {
        title: t('模型'),
        dataIndex: 'model',
        render: (text) => <Typography.Text code>{text}</Typography.Text>,
      },
      {
        title: t('Tokens'),
        dataIndex: 'tokens',
      },
      {
        title: t('请求状态'),
        dataIndex: 'status',
        render: (text) => {
          if (text === 'success') {
            return <Tag color='green'>{t('成功')}</Tag>;
          }
          if (text === 'cached') {
            return <Tag color='blue'>{t('缓存')}</Tag>;
          }
          return <Tag color='orange'>{t('失败')}</Tag>;
        },
      },
      {
        title: t('上游通道'),
        dataIndex: 'upstream_alias',
        render: (text) => text || '-',
      },
      {
        title: t('摘要'),
        dataIndex: 'summary',
        render: (text) => text || '-',
      },
    ],
    [t],
  );

  const handleFilterSubmit = (values) => {
    setFilters({
      window: values.window,
      model: values.model,
      search: values.search,
    });
    setPage(1);
  };

  const exportLogs = async () => {
    setExporting(true);
    try {
      const params = new URLSearchParams({
        window: filters.window,
        model: filters.model,
        search: filters.search,
      });
      window.open(`/api/public/logs/export?${params.toString()}`, '_blank');
    } finally {
      setExporting(false);
    }
  };

  if (configLoading) {
    return (
      <div className='mt-[60px] px-2 pb-6'>
        <Card loading></Card>
      </div>
    );
  }

  if (error) {
    return (
      <div className='mt-[60px] px-2 pb-6'>
        <Card bordered>
          <Space vertical align='center' className='w-full'>
            <Empty description={t('计费配置加载失败，暂时无法展示公开日志')} />
            <Button theme='solid' type='primary' onClick={() => refresh()}>
              {t('重试')}
            </Button>
          </Space>
        </Card>
      </div>
    );
  }

  if (!publicLogsEnabled) {
    return (
      <div className='mt-[60px] px-2 pb-6'>
        <Empty description={t('公开日志已禁用')} />
      </div>
    );
  }

  return (
    <div className='mt-[60px] px-2 pb-6 space-y-4'>
      <Card bordered>
        <Form
          layout='horizontal'
          labelPosition='left'
          onSubmit={handleFilterSubmit}
          initValues={filters}
        >
          <Space wrap>
            <Form.Select
              field='window'
              label={t('时间范围')}
              optionList={[
                { label: t('24小时'), value: '24h' },
                { label: t('7天'), value: '7d' },
                { label: t('全部'), value: 'all_time' },
              ]}
              style={{ width: 160 }}
            />
            <Form.Select
              field='model'
              label={t('模型')}
              optionList={[{ label: t('全部'), value: '' }].concat(
                models.map((model) => ({ label: model, value: model })),
              )}
              style={{ width: 200 }}
            />
            <Form.Input
              field='search'
              label={t('关键词')}
              placeholder={t('搜索匿名用户或摘要')}
              style={{ width: 240 }}
            />
            <Button type='primary' htmlType='submit'>
              {t('筛选')}
            </Button>
            <Button loading={exporting} onClick={exportLogs} type='tertiary'>
              {t('导出 CSV')}
            </Button>
          </Space>
        </Form>
      </Card>

      <Card bordered>
        <Table
          loading={loading}
          columns={columns}
          dataSource={logs}
          pagination={false}
          empty={<Empty description={t('暂无数据')} />}
        />
        <div className='flex justify-end mt-4'>
          <Pagination
            currentPage={page}
            pageSize={pageSize}
            total={total}
            showSizeChanger
            pageSizeOpts={[20, 50, 100]}
            onPageChange={(p) => setPage(p)}
            onPageSizeChange={(size) => {
              setPageSize(size);
              setPage(1);
            }}
          />
        </div>
      </Card>
    </div>
  );
};

export default PublicLogsPage;
