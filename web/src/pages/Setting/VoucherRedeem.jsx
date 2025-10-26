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
  Divider,
  Empty,
  Form,
  Skeleton,
  Space,
  Table,
  Tag,
  Typography,
} from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import { API, showError, showSuccess, timestamp2string } from '../../helpers';
import QuotaProgress from '../../components/billing/QuotaProgress';
import { useBillingFeatures } from '../../hooks/billing/useBillingFeatures';

const VoucherRedeem = () => {
  const { t } = useTranslation();
  const { config, loading: configLoading, error, refresh } = useBillingFeatures();

  const [redeemLoading, setRedeemLoading] = useState(false);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [planLoading, setPlanLoading] = useState(false);
  const [voucherCode, setVoucherCode] = useState('');
  const [history, setHistory] = useState([]);
  const [planAssignments, setPlanAssignments] = useState([]);

  const billingEnabled = config?.billing?.enabled;

  const loadHistory = useCallback(async () => {
    if (!billingEnabled) {
      setHistory([]);
      return;
    }
    setHistoryLoading(true);
    try {
      const res = await API.get('/api/voucher/redemptions/self', {
        skipErrorHandler: true,
      });
      if (res?.data?.success) {
        setHistory(Array.isArray(res.data.data) ? res.data.data : []);
      } else {
        setHistory([]);
      }
    } catch (error) {
      setHistory([]);
    } finally {
      setHistoryLoading(false);
    }
  }, [billingEnabled]);

  const loadPlan = useCallback(async () => {
    if (!billingEnabled) {
      setPlanAssignments([]);
      return;
    }
    setPlanLoading(true);
    try {
      const res = await API.get('/api/plan/self', { skipErrorHandler: true });
      if (res?.data?.success && Array.isArray(res.data.data)) {
        setPlanAssignments(res.data.data);
      } else {
        setPlanAssignments([]);
      }
    } catch (error) {
      setPlanAssignments([]);
    } finally {
      setPlanLoading(false);
    }
  }, [billingEnabled]);

  useEffect(() => {
    if (!billingEnabled) {
      return;
    }
    loadHistory();
    loadPlan();
  }, [billingEnabled, loadHistory, loadPlan]);

  const handleRedeem = async () => {
    if (!voucherCode || !voucherCode.trim()) {
      showError(t('请输入兑换码'));
      return;
    }
    setRedeemLoading(true);
    try {
      const res = await API.post(
        '/api/voucher/redeem',
        { code: voucherCode.trim() },
        { skipErrorHandler: true },
      );
      if (res?.data?.success) {
        showSuccess(res.data.message || t('兑换成功'));
        setVoucherCode('');
        await Promise.all([loadHistory(), loadPlan()]);
      } else {
        showError(res?.data?.message || t('兑换失败'));
      }
    } catch (error) {
      showError(error?.response?.data?.message || error?.message || t('兑换失败'));
    } finally {
      setRedeemLoading(false);
    }
  };

  const historyColumns = useMemo(
    () => [
      {
        title: t('兑换码'),
        dataIndex: 'code',
        render: (text) => <Typography.Text code>{text}</Typography.Text>,
      },
      {
        title: t('类型'),
        dataIndex: 'grant_type',
        render: (text) => {
          if (text === 'plan') {
            return <Tag color='purple'>{t('计划')}</Tag>;
          }
          if (text === 'credit') {
            return <Tag color='green'>{t('余额')}</Tag>;
          }
          return <Tag>{text}</Tag>;
        },
      },
      {
        title: t('额度变更'),
        dataIndex: 'credit_amount',
        render: (amount, record) => {
          if (amount) {
            return (
              <Typography.Text type={amount > 0 ? 'success' : 'danger'}>
                {amount > 0 ? '+' : ''}
                {amount}
              </Typography.Text>
            );
          }
          if (record?.plan_name) {
            return <Typography.Text>{record.plan_name}</Typography.Text>;
          }
          return '-';
        },
      },
      {
        title: t('兑换时间'),
        dataIndex: 'redeemed_at',
        render: (text) => (text ? timestamp2string(text) : '-'),
      },
      {
        title: t('备注'),
        dataIndex: 'message',
        render: (text) => text || '-',
      },
    ],
    [t],
  );

  const planCards = useMemo(() => {
    if (!billingEnabled) {
      return null;
    }
    if (planLoading) {
      return (
        <div className='grid gap-4 md:grid-cols-2'>
          <Skeleton loading>
            <div className='h-[120px] rounded-lg bg-[var(--semi-color-fill-0)]' />
          </Skeleton>
          <Skeleton loading>
            <div className='h-[120px] rounded-lg bg-[var(--semi-color-fill-0)]' />
          </Skeleton>
        </div>
      );
    }
    if (!planAssignments || planAssignments.length === 0) {
      return <Empty description={t('暂无激活的订阅计划')} />;
    }

    return (
      <div className='grid gap-4 md:grid-cols-2'>
        {planAssignments.map((item) => {
          const planName = item?.plan?.name || t('未命名计划');
          const billingMode = item?.assignment?.billing_mode || 'plan';
          const usage = item?.usage || {};
          const daily = usage?.daily || {};
          const monthly = usage?.monthly || {};
          const nextReset = usage?.next_reset || item?.assignment?.cycle_end;
          return (
            <Card key={item?.assignment?.id || planName} bordered>
              <Space vertical spacing='tight' className='w-full'>
                <div className='flex items-center justify-between'>
                  <Typography.Title heading={5}>{planName}</Typography.Title>
                  <Tag color='blue'>{billingMode.toUpperCase()}</Tag>
                </div>
                <Typography.Text type='tertiary'>
                  {nextReset ? t('下次重置时间：{{time}}', { time: timestamp2string(nextReset) }) : t('无重置时间信息')}
                </Typography.Text>
                <QuotaProgress
                  title={t('今日额度')}
                  used={daily?.used ?? 0}
                  total={daily?.limit ?? 0}
                  suffix={daily?.unit || ''}
                  description={t('今日剩余额度')}
                />
                <QuotaProgress
                  title={t('本月额度')}
                  used={monthly?.used ?? 0}
                  total={monthly?.limit ?? 0}
                  suffix={monthly?.unit || ''}
                  description={t('本月剩余额度')}
                />
              </Space>
            </Card>
          );
        })}
      </div>
    );
  }, [billingEnabled, planAssignments, planLoading, t]);

  if (configLoading) {
    return (
      <div className='mt-[60px] px-2 pb-6'>
        <Skeleton loading>
          <div className='h-[120px] rounded-lg bg-[var(--semi-color-fill-0)]' />
        </Skeleton>
      </div>
    );
  }

  if (error) {
    return (
      <div className='mt-[60px] px-2 pb-6'>
        <Card bordered>
          <Space direction='vertical' align='center' className='w-full'>
            <Empty description={t('计费配置加载失败，无法使用兑换码功能')} />
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
        <Empty description={t('计费功能未启用，无法使用兑换码功能。')} />
      </div>
    );
  }

  return (
    <div className='mt-[60px] px-2 pb-6 space-y-4'>
      <Card bordered>
        <Space vertical spacing='medium' className='w-full'>
          <Typography.Title heading={4}>{t('兑换卡券')}</Typography.Title>
          <Form labelPosition='top'>
            <Form.Input
              field='voucherCode'
              label={t('兑换码')}
              value={voucherCode}
              onChange={setVoucherCode}
              placeholder={t('请输入兑换码')}
            />
            <Button
              type='primary'
              loading={redeemLoading}
              onClick={handleRedeem}
              disabled={!voucherCode || redeemLoading}
            >
              {t('立即兑换')}
            </Button>
          </Form>
        </Space>
      </Card>

      <Card bordered>
        <Typography.Title heading={5}>{t('我的订阅计划')}</Typography.Title>
        <Divider margin='12px 0' />
        {planCards}
      </Card>

      <Card bordered>
        <Space vertical spacing='medium' className='w-full'>
          <Typography.Title heading={5}>{t('兑换记录')}</Typography.Title>
          <Table
            loading={historyLoading}
            columns={historyColumns}
            dataSource={history}
            pagination={{ pageSize: 10 }}
            empty={<Empty description={t('暂无兑换记录')} />}
          />
        </Space>
      </Card>
    </div>
  );
};

export default VoucherRedeem;
