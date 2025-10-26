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

import React, { useEffect, useState } from 'react';
import {
  Banner,
  Button,
  Card,
  Col,
  Row,
  Typography,
  Spin,
  Switch,
  Select,
  InputNumber,
} from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import { API, showError, showSuccess } from '../../helpers';
import { useBillingFeatures } from '../../hooks/billing/useBillingFeatures';

const BillingGovernanceSetting = () => {
  const { t } = useTranslation();
  const { config, refresh: refreshConfig, loading, error } = useBillingFeatures();
  const [saving, setSaving] = useState(false);
  const [errorDismissed, setErrorDismissed] = useState(false);

  const [billingEnabled, setBillingEnabled] = useState(false);
  const [defaultMode, setDefaultMode] = useState('balance');
  const [governanceEnabled, setGovernanceEnabled] = useState(false);
  const [publicLogsEnabled, setPublicLogsEnabled] = useState(false);
  const [publicLogsRetention, setPublicLogsRetention] = useState(7);

  useEffect(() => {
    if (!loading && !error && config) {
      try {
        setBillingEnabled(config?.billing?.enabled ?? false);
        setDefaultMode(config?.billing?.defaultMode ?? 'balance');
        setGovernanceEnabled(config?.governance?.enabled ?? false);
        setPublicLogsEnabled(config?.public_logs?.enabled ?? false);
        setPublicLogsRetention(config?.public_logs?.retention_days ?? 7);
      } catch (err) {
        console.error('Failed to load form values:', err);
      }
    }
  }, [config, loading, error]);

  useEffect(() => {
    if (error) {
      setErrorDismissed(false);
    }
  }, [error]);

  const handleSubmit = async () => {
    setSaving(true);
    try {
      const retentionDays = Number(publicLogsRetention) || 7;
      const payload = {
        billing: {
          enabled: Boolean(billingEnabled),
          defaultMode: defaultMode || 'balance',
        },
        governance: {
          enabled: Boolean(governanceEnabled),
        },
        public_logs: {
          enabled: Boolean(publicLogsEnabled),
          retention_days: retentionDays > 0 ? retentionDays : 7,
        },
      };
      const res = await API.put('/api/option/features', payload);
      if (res?.data?.success) {
        showSuccess(t('设置已保存'));
        await refreshConfig();
      } else {
        showError(res?.data?.message || t('保存失败'));
      }
    } catch (error) {
      console.error('Failed to save billing/governance settings:', error);
      showError(error?.response?.data?.message || error?.message || t('保存失败'));
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className='flex justify-center items-center min-h-[400px]'>
        <Spin size='large' />
      </div>
    );
  }

  return (
    <div className='space-y-6'>
      {!errorDismissed && error ? (
        <Banner
          type='danger'
          description={error?.message || t('配置加载失败，请稍后再试')}
          closeIcon
          onClose={() => setErrorDismissed(true)}
          actions={
            <Button
              size='small'
              theme='solid'
              type='tertiary'
              onClick={() => refreshConfig()}
            >
              {t('重试')}
            </Button>
          }
        />
      ) : null}

      <Card title={t('计费配置')} bordered>
        <Row gutter={16} className='mb-4'>
          <Col span={12}>
            <div className='flex items-center justify-between'>
              <Typography.Text>{t('启用计费功能')}</Typography.Text>
              <Switch
                checked={billingEnabled}
                onChange={(checked) => setBillingEnabled(Boolean(checked))}
                disabled={Boolean(error) || saving}
              />
            </div>
          </Col>
          <Col span={12}>
            <div className='flex items-center justify-between'>
              <Typography.Text>{t('默认计费模式')}</Typography.Text>
              <Select
                value={defaultMode}
                onChange={(value) => setDefaultMode(value)}
                disabled={Boolean(error) || saving}
                style={{ width: 200 }}
                optionList={[
                  { label: t('余额'), value: 'balance' },
                  { label: t('计划'), value: 'plan' },
                  { label: t('自动'), value: 'auto' },
                ]}
              />
            </div>
          </Col>
        </Row>
      </Card>

      <Card title={t('治理配置')} bordered>
        <div className='flex items-center justify-between'>
          <Typography.Text>{t('启用治理标记')}</Typography.Text>
          <Switch
            checked={governanceEnabled}
            onChange={(checked) => setGovernanceEnabled(Boolean(checked))}
            disabled={Boolean(error) || saving}
          />
        </div>
      </Card>

      <Card title={t('公开日志')} bordered>
        <Row gutter={16} className='mb-4'>
          <Col span={12}>
            <div className='flex items-center justify-between'>
              <Typography.Text>{t('启用公开日志')}</Typography.Text>
              <Switch
                checked={publicLogsEnabled}
                onChange={(checked) => setPublicLogsEnabled(Boolean(checked))}
                disabled={Boolean(error) || saving}
              />
            </div>
          </Col>
          <Col span={12}>
            <div className='flex items-center justify-between'>
              <Typography.Text>{t('日志保留天数')}</Typography.Text>
              <InputNumber
                value={publicLogsRetention}
                onChange={(value) => setPublicLogsRetention(Number(value) || 7)}
                min={1}
                max={30}
                step={1}
                disabled={Boolean(error) || saving}
                style={{ width: 200 }}
              />
            </div>
          </Col>
        </Row>
        <Typography.Text type='tertiary' size='small'>
          {t('公开日志将展示匿名化的请求信息，并在设定的保留时间后自动清理。')}
        </Typography.Text>
      </Card>

      <Button
        theme='solid'
        type='primary'
        onClick={handleSubmit}
        loading={saving}
        disabled={Boolean(error)}
      >
        {t('保存设置')}
      </Button>
    </div>
  );
};

export default BillingGovernanceSetting;
