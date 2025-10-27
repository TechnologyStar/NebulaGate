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

import React, { useEffect, useMemo, useState } from 'react';
import { Button, Card, Col, Empty, Row, Skeleton, Tag, Typography } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import { API, timestamp2string } from '../../helpers';
import { useBillingFeatures } from '../../hooks/billing/useBillingFeatures';
import QuotaProgress from '../billing/QuotaProgress';

const PlanUsagePanel = () => {
  const { t } = useTranslation();
  const { config, loading: configLoading, error, refresh } = useBillingFeatures();
  const [assignments, setAssignments] = useState([]);
  const [loading, setLoading] = useState(false);
  const billingEnabled = config?.billing?.enabled;

  useEffect(() => {
    if (!billingEnabled) {
      return;
    }
    setLoading(true);
    API.get('/api/plan/self', { skipErrorHandler: true })
      .then((res) => {
        if (res?.data?.success && Array.isArray(res.data.data)) {
          setAssignments(res.data.data);
        } else {
          setAssignments([]);
        }
      })
      .catch(() => {
        setAssignments([]);
      })
      .finally(() => setLoading(false));
  }, [billingEnabled]);

  const cards = useMemo(() => {
    if (!billingEnabled) {
      return null;
    }
    if (loading) {
      return (
        <Row gutter={16} className='mt-4'>
          {[0, 1].map((key) => (
            <Col span={12} key={key}>
              <Skeleton loading>
                <div className='h-[140px] rounded-lg bg-[var(--semi-color-fill-0)]' />
              </Skeleton>
            </Col>
          ))}
        </Row>
      );
    }
    if (!assignments || assignments.length === 0) {
      return (
        <div className='mt-4'>
          <Empty description={t('暂无激活的订阅计划')} />
        </div>
      );
    }
    return (
      <Row gutter={16} className='mt-4'>
        {assignments.map((item) => {
          const planName = item?.plan?.name || t('未命名计划');
          const billingMode = item?.assignment?.billing_mode || 'plan';
          const usage = item?.usage || {};
          const daily = usage?.daily || {};
          const monthly = usage?.monthly || {};
          const nextReset = usage?.next_reset || item?.assignment?.cycle_end;
          return (
            <Col span={12} key={item?.assignment?.id || planName}>
              <Card bordered>
                <div className='flex items-center justify-between mb-2'>
                  <Typography.Title heading={5}>{planName}</Typography.Title>
                  <Tag color='blue'>{billingMode.toUpperCase()}</Tag>
                </div>
                <Typography.Text type='tertiary'>
                  {nextReset ? t('下次重置时间：{{time}}', { time: timestamp2string(nextReset) }) : t('无重置时间信息')}
                </Typography.Text>
                <div className='mt-3 grid gap-3'>
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
                </div>
              </Card>
            </Col>
          );
        })}
      </Row>
    );
  }, [assignments, billingEnabled, loading, t]);

  if (configLoading) {
    return <Card bordered loading></Card>;
  }

  if (error) {
    return (
      <Card bordered>
        <Typography.Title heading={4}>{t('订阅用量')}</Typography.Title>
        <Typography.Text type='tertiary'>
          {t('暂时无法加载订阅用量信息，请稍后重试。')}
        </Typography.Text>
        <div className='mt-4 flex justify-center'>
          <Button theme='solid' type='primary' onClick={() => refresh()}>
            {t('重试')}
          </Button>
        </div>
      </Card>
    );
  }

  if (!billingEnabled) {
    return null;
  }

  return (
    <Card bordered className='!rounded-2xl'>
      <Typography.Title heading={4}>{t('订阅用量')}</Typography.Title>
      <Typography.Text type='tertiary'>
        {t('概览当前订阅计划的每日与每月剩余额度。')}
      </Typography.Text>
      {cards}
    </Card>
  );
};

export default PlanUsagePanel;
