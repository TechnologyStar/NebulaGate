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

import React, { useEffect, useRef, useState } from 'react';
import { Button, Card, Col, Form, Row, Typography, Spin } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import { API, showError, showSuccess } from '../../helpers';
import { useBillingFeatures } from '../../hooks/billing/useBillingFeatures';

const BillingGovernanceSetting = () => {
  const { t } = useTranslation();
  const { config, refresh: refreshConfig, loading } = useBillingFeatures();
  const [saving, setSaving] = useState(false);
  const formRef = useRef(null);

  useEffect(() => {
    if (!loading && config && formRef.current) {
      formRef.current.setValues({
        billingEnabled: config?.billing?.enabled ?? false,
        defaultMode: config?.billing?.defaultMode ?? 'balance',
        governanceEnabled: config?.governance?.enabled ?? false,
        publicLogsEnabled: config?.public_logs?.enabled ?? false,
        publicLogsRetention: config?.public_logs?.retention_days ?? 7,
      });
    }
  }, [config, loading]);

  const handleSubmit = async (values) => {
    setSaving(true);
    try {
      const payload = {
        billing: {
          enabled: values.billingEnabled,
          defaultMode: values.defaultMode,
        },
        governance: {
          enabled: values.governanceEnabled,
        },
        public_logs: {
          enabled: values.publicLogsEnabled,
          retention_days: values.publicLogsRetention,
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
      showError(error?.response?.data?.message || error?.message || t('保存失败'));
    } finally {
      setSaving(false);
    }
  };

  return (
    <Spin spinning={loading || saving}>
      <Form
        getFormApi={(api) => {
          formRef.current = api;
        }}
        labelPosition='left'
        labelWidth={220}
        onSubmit={handleSubmit}
        className='space-y-6'
      >
        <Card title={t('计费配置')} bordered>
          <Row gutter={16} align='middle'>
            <Col span={12}>
              <Form.Switch field='billingEnabled' label={t('启用计费功能')} />
            </Col>
            <Col span={12}>
              <Form.Select
                field='defaultMode'
                label={t('默认计费模式')}
                optionList={[
                  { label: t('余额'), value: 'balance' },
                  { label: t('计划'), value: 'plan' },
                  { label: t('自动'), value: 'auto' },
                ]}
              />
            </Col>
          </Row>
        </Card>

        <Card title={t('治理配置')} bordered>
          <Form.Switch field='governanceEnabled' label={t('启用治理标记')} />
        </Card>

        <Card title={t('公开日志')} bordered>
          <Row gutter={16} align='middle'>
            <Col span={12}>
              <Form.Switch field='publicLogsEnabled' label={t('启用公开日志')} />
            </Col>
            <Col span={12}>
              <Form.InputNumber
                field='publicLogsRetention'
                label={t('日志保留天数')}
                min={1}
                max={30}
                step={1}
              />
            </Col>
          </Row>
          <Typography.Text type='tertiary'>
            {t('公开日志将展示匿名化的请求信息，并在设定的保留时间后自动清理。')}
          </Typography.Text>
        </Card>

        <Button theme='solid' type='primary' htmlType='submit'>
          {t('保存设置')}
        </Button>
      </Form>
    </Spin>
  );
};

export default BillingGovernanceSetting;
