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
  Form,
  Modal,
  Radio,
  Select,
  Space,
  Table,
  Tag,
  Typography,
  InputNumber,
} from '@douyinfe/semi-ui';
import { Plus, Gift, RefreshCcw } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { formatDateTimeString, showError, showSuccess } from '../../helpers';
import {
  fetchPlans,
  createPlan,
  updatePlan,
  deletePlan,
  generateVoucher,
  listVouchers,
} from '../../services/api';

const { Title, Text } = Typography;

const CYCLE_OPTIONS = [
  { value: 'daily', labelKey: '按日' },
  { value: 'monthly', labelKey: '按月' },
  { value: 'custom', labelKey: '自定义' },
];

const QUOTA_METRIC_OPTIONS = [
  { value: 'requests', labelKey: '请求次数' },
  { value: 'tokens', labelKey: 'Token 数量' },
];

const ROLLOVER_OPTIONS = [
  { value: 'none', labelKey: '不允许结转' },
  { value: 'carry_all', labelKey: '全部结转' },
  { value: 'cap', labelKey: '结转封顶' },
];

const GRANT_TYPE_OPTIONS = [
  { value: 'credit', labelKey: '余额卡券' },
  { value: 'plan', labelKey: '套餐卡券' },
];

const mapPlanToFormValues = (plan) => {
  if (!plan) {
    return null;
  }
  let rolloverPolicy = 'none';
  if (plan.allow_carry_over) {
    rolloverPolicy = plan.carry_limit_percent > 0 ? 'cap' : 'carry_all';
  }
  return {
    id: plan.id,
    name: plan.name,
    description: plan.description,
    cycle: plan.cycle_type,
    cycle_length_days: plan.cycle_duration_days || undefined,
    quota: plan.quota_amount,
    quota_metric: plan.quota_metric,
    rollover_policy: rolloverPolicy,
  };
};

const buildPlanPayload = (values) => {
  const payload = {
    name: values.name?.trim(),
    description: values.description?.trim(),
    cycle: values.cycle,
    quota: values.quota,
    quota_metric: values.quota_metric,
    rollover_policy: values.rollover_policy,
  };

  if (values.cycle === 'custom') {
    payload.cycle_length_days = values.cycle_length_days;
  }

  if (values.cycle !== 'custom') {
    payload.cycle_length_days = undefined;
  }

  return payload;
};

const PlanFormModal = ({
  visible,
  onCancel,
  onSubmit,
  initialValues,
  loading,
}) => {
  const { t } = useTranslation();
  const [formApi, setFormApi] = useState(null);

  useEffect(() => {
    if (!formApi) {
      return;
    }
    if (visible) {
      if (initialValues) {
        formApi.setValues({
          name: initialValues.name,
          description: initialValues.description,
          cycle: initialValues.cycle || 'monthly',
          cycle_length_days: initialValues.cycle_length_days,
          quota: initialValues.quota,
          quota_metric: initialValues.quota_metric || 'requests',
          rollover_policy: initialValues.rollover_policy || 'none',
        });
      } else {
        formApi.reset();
        formApi.setValues({
          cycle: 'monthly',
          quota_metric: 'requests',
          rollover_policy: 'none',
        });
      }
    } else {
      formApi.reset();
    }
  }, [formApi, initialValues, visible]);

  const handleSubmit = async (values) => {
    await onSubmit(values);
  };

  return (
    <Modal
      visible={visible}
      title={initialValues ? t('编辑套餐') : t('创建套餐')}
      onCancel={onCancel}
      onOk={() => formApi?.submitForm?.()}
      confirmLoading={loading}
      okText={t('保存套餐')}
      cancelText={t('取消')}
      maskClosable={false}
    >
      <Form
        layout='horizontal'
        labelPosition='left'
        labelAlign='left'
        getFormApi={setFormApi}
        onSubmit={handleSubmit}
      >
        <Form.Input
          field='name'
          label={t('套餐名称')}
          rules={[{ required: true, message: t('套餐名称') }]}
        />
        <Form.TextArea
          field='description'
          label={t('套餐描述')}
          autosize={{ minRows: 3, maxRows: 6 }}
        />
        <Form.Select
          field='cycle'
          label={t('结算周期')}
          rules={[{ required: true, message: t('结算周期') }]}
          onChange={(value) => {
            if (value !== 'custom') {
              formApi?.setValue('cycle_length_days', undefined);
            }
          }}
        >
          {CYCLE_OPTIONS.map((option) => (
            <Select.Option key={option.value} value={option.value}>
              {t(option.labelKey)}
            </Select.Option>
          ))}
        </Form.Select>
        <Form.InputNumber
          field='cycle_length_days'
          label={t('自定义周期天数')}
          min={1}
          rules={[
            {
              validator: (rule, value) => {
                const cycle = formApi?.getValue('cycle');
                if (cycle === 'custom') {
                  if (!value || value < 1) {
                    return Promise.reject(t('周期天数必须大于等于1'));
                  }
                }
                return Promise.resolve();
              },
            },
          ]}
        />
        <Form.InputNumber
          field='quota'
          label={t('额度数量')}
          min={0}
          rules={[
            { required: true, message: t('额度数量') },
            {
              validator: (rule, value) => {
                if (value === null || value === undefined) {
                  return Promise.reject(t('额度数量'));
                }
                if (value < 0) {
                  return Promise.reject(t('额度必须大于等于0'));
                }
                return Promise.resolve();
              },
            },
          ]}
        />
        <Form.Select
          field='quota_metric'
          label={t('额度单位')}
          rules={[{ required: true, message: t('额度单位') }]}
        >
          {QUOTA_METRIC_OPTIONS.map((option) => (
            <Select.Option key={option.value} value={option.value}>
              {t(option.labelKey)}
            </Select.Option>
          ))}
        </Form.Select>
        <Form.Select
          field='rollover_policy'
          label={t('结转策略')}
          rules={[{ required: true, message: t('结转策略') }]}
        >
          {ROLLOVER_OPTIONS.map((option) => (
            <Select.Option key={option.value} value={option.value}>
              {t(option.labelKey)}
            </Select.Option>
          ))}
        </Form.Select>
      </Form>
    </Modal>
  );
};

const VoucherModal = ({ visible, onCancel, onSubmit, loading, plans }) => {
  const { t } = useTranslation();
  const [formApi, setFormApi] = useState(null);

  useEffect(() => {
    if (!formApi) {
      return;
    }
    if (visible) {
      formApi.reset();
      formApi.setValues({
        grant_type: 'credit',
        count: 1,
      });
    }
  }, [formApi, visible]);

  const handleSubmit = async (values) => {
    await onSubmit(values);
  };

  const selectedGrantType = formApi?.getValue('grant_type') || 'credit';

  return (
    <Modal
      visible={visible}
      title={t('生成卡券')}
      onCancel={onCancel}
      onOk={() => formApi?.submitForm?.()}
      confirmLoading={loading}
      okText={t('生成卡券')}
      cancelText={t('取消')}
      maskClosable={false}
    >
      <Form
        layout='horizontal'
        labelPosition='left'
        labelAlign='left'
        getFormApi={setFormApi}
        onSubmit={handleSubmit}
      >
        <Form.RadioGroup field='grant_type' label={t('卡券类型')} type='button'>
          {GRANT_TYPE_OPTIONS.map((option) => (
            <Radio value={option.value} key={option.value}>
              {t(option.labelKey)}
            </Radio>
          ))}
        </Form.RadioGroup>
        <Form.Input field='prefix' label={t('前缀')} placeholder='PROMO' />
        <Form.InputNumber
          field='count'
          label={t('发放数量')}
          min={1}
          rules={[{ required: true, message: t('发放数量') }]}
        />
        {selectedGrantType === 'credit' ? (
          <Form.InputNumber
            field='credit_amount'
            label={t('余额额度')}
            min={1}
            rules={[{ required: true, message: t('余额额度') }]}
          />
        ) : (
          <Form.Select
            field='plan_id'
            label={t('选择套餐')}
            rules={[{ required: true, message: t('选择套餐') }]}
          >
            {plans.map((plan) => (
              <Select.Option key={plan.id} value={plan.id}>
                {plan.name}
              </Select.Option>
            ))}
          </Form.Select>
        )}
        <Form.InputNumber
          field='expire_days'
          label={t('有效期天数')}
          min={1}
          rules={[{ required: true, message: t('有效期天数') }]}
        />
        <Form.TextArea
          field='note'
          label={t('卡券备注')}
          autosize={{ minRows: 2, maxRows: 4 }}
        />
      </Form>
    </Modal>
  );
};

const PlansPage = () => {
  const { t } = useTranslation();
  const [plans, setPlans] = useState([]);
  const [planLoading, setPlanLoading] = useState(false);
  const [voucherLoading, setVoucherLoading] = useState(false);
  const [planModalVisible, setPlanModalVisible] = useState(false);
  const [voucherModalVisible, setVoucherModalVisible] = useState(false);
  const [planModalLoading, setPlanModalLoading] = useState(false);
  const [voucherModalLoading, setVoucherModalLoading] = useState(false);
  const [editingPlan, setEditingPlan] = useState(null);
  const [voucherBatches, setVoucherBatches] = useState([]);

  const loadPlans = useCallback(async () => {
    setPlanLoading(true);
    try {
      const result = await fetchPlans();
      if (result.success) {
        setPlans(Array.isArray(result.data) ? result.data : []);
      } else {
        showError(result.message || t('加载套餐失败'));
      }
    } catch (error) {
      showError(error?.response?.data?.message || t('加载套餐失败'));
    } finally {
      setPlanLoading(false);
    }
  }, [t]);

  const loadVoucherBatches = useCallback(async () => {
    setVoucherLoading(true);
    try {
      const result = await listVouchers();
      if (result.success) {
        setVoucherBatches(Array.isArray(result.data) ? result.data : []);
      } else {
        showError(result.message || t('加载卡券批次失败'));
      }
    } catch (error) {
      showError(error?.response?.data?.message || t('加载卡券批次失败'));
    } finally {
      setVoucherLoading(false);
    }
  }, [t]);

  useEffect(() => {
    loadPlans();
    loadVoucherBatches();
  }, [loadPlans, loadVoucherBatches]);

  const handleCreatePlan = useCallback(() => {
    setEditingPlan(null);
    setPlanModalVisible(true);
  }, []);

  const handleEditPlan = useCallback((plan) => {
    setEditingPlan(mapPlanToFormValues(plan));
    setPlanModalVisible(true);
  }, []);

  const handleDeletePlan = useCallback(
    (plan) => {
      Modal.confirm({
        title: t('删除套餐'),
        content: t('确定要删除此套餐吗？'),
        okText: t('删除'),
        cancelText: t('取消'),
        onOk: async () => {
          try {
            const result = await deletePlan(plan.id);
            if (result.success) {
              showSuccess(t('套餐删除成功'));
              await loadPlans();
            } else {
              showError(result.message || t('套餐删除失败'));
            }
          } catch (error) {
            showError(error?.response?.data?.message || t('套餐删除失败'));
          }
        },
      });
    },
    [loadPlans, t],
  );

  const handlePlanSubmit = useCallback(
    async (values) => {
      setPlanModalLoading(true);
      try {
        const payload = buildPlanPayload(values);
        let result;
        if (editingPlan?.id) {
          result = await updatePlan(editingPlan.id, payload);
        } else {
          result = await createPlan(payload);
        }

        if (result.success) {
          showSuccess(editingPlan?.id ? t('套餐更新成功') : t('套餐创建成功'));
          setPlanModalVisible(false);
          await loadPlans();
        } else {
          showError(result.message || t('套餐保存失败'));
        }
      } catch (error) {
        showError(error?.response?.data?.message || t('套餐保存失败'));
      } finally {
        setPlanModalLoading(false);
      }
    },
    [editingPlan, loadPlans, t],
  );

  const handleVoucherSubmit = useCallback(
    async (values) => {
      setVoucherModalLoading(true);
      try {
        const payload = {
          grant_type: values.grant_type,
          count: values.count,
          prefix: values.prefix?.trim(),
          expire_days: values.expire_days,
          note: values.note?.trim(),
        };

        if (values.grant_type === 'credit') {
          payload.credit_amount = values.credit_amount;
        } else {
          payload.plan_id = values.plan_id;
        }

        const result = await generateVoucher(payload);
        if (result.success) {
          showSuccess(t('生成卡券成功'));
          setVoucherModalVisible(false);
          await loadVoucherBatches();
        } else {
          showError(result.message || t('生成卡券失败'));
        }
      } catch (error) {
        showError(error?.response?.data?.message || t('生成卡券失败'));
      } finally {
        setVoucherModalLoading(false);
      }
    },
    [loadVoucherBatches, t],
  );

  const planColumns = useMemo(() => {
    const cycleLabel = (plan) => {
      switch (plan.cycle_type) {
        case 'daily':
          return t('按日');
        case 'custom':
          return `${t('自定义')} (${plan.cycle_duration_days || 0})`;
        case 'monthly':
        default:
          return t('按月');
      }
    };

    const quotaLabel = (plan) => {
      switch (plan.quota_metric) {
        case 'tokens':
          return t('Token 数量');
        case 'requests':
        default:
          return t('请求次数');
      }
    };

    return [
      {
        title: t('套餐名称'),
        dataIndex: 'name',
        render: (text, record) => (
          <div className='flex flex-col gap-1'>
            <Text strong>{text}</Text>
            {record.code ? <Text type='tertiary'>{record.code}</Text> : null}
          </div>
        ),
      },
      {
        title: t('套餐描述'),
        dataIndex: 'description',
        render: (text) => text || '-',
      },
      {
        title: t('结算周期'),
        dataIndex: 'cycle_type',
        render: (_, record) => cycleLabel(record),
      },
      {
        title: t('额度数量'),
        dataIndex: 'quota_amount',
        render: (_, record) => (
          <div className='flex flex-col'>
            <Text>{record.quota_amount}</Text>
            <Text type='tertiary'>{quotaLabel(record)}</Text>
          </div>
        ),
      },
      {
        title: t('结转策略'),
        dataIndex: 'allow_carry_over',
        render: (_, record) => {
          if (!record.allow_carry_over) {
            return <Tag type='solid'>{t('不允许结转')}</Tag>;
          }
          if (record.carry_limit_percent > 0) {
            return <Tag color='amber'>{t('结转封顶')}</Tag>;
          }
          return <Tag color='green'>{t('全部结转')}</Tag>;
        },
      },
      {
        title: t('启用'),
        dataIndex: 'is_active',
        render: (value) => (
          <Tag color={value ? 'green' : 'grey'}>
            {value ? t('启用') : t('未启用')}
          </Tag>
        ),
      },
      {
        title: t('更新时间'),
        dataIndex: 'updated_at',
        render: (value) =>
          value ? formatDateTimeString(new Date(value)) : '-',
      },
      {
        title: t('操作'),
        dataIndex: 'actions',
        width: 160,
        render: (_, record) => (
          <Space>
            <Button
              size='small'
              theme='borderless'
              type='primary'
              onClick={() => handleEditPlan(record)}
            >
              {t('编辑套餐')}
            </Button>
            <Button
              size='small'
              theme='borderless'
              type='danger'
              onClick={() => handleDeletePlan(record)}
            >
              {t('删除')}
            </Button>
          </Space>
        ),
      },
    ];
  }, [handleDeletePlan, handleEditPlan, t]);

  const voucherColumns = useMemo(
    () => [
      {
        title: t('批次'),
        dataIndex: 'batch_label',
        render: (text, record) => (
          <div className='flex flex-col gap-1'>
            <Text strong>{text}</Text>
            <Text type='tertiary'>{record.code_prefix}</Text>
          </div>
        ),
      },
      {
        title: t('卡券类型'),
        dataIndex: 'grant_type',
        render: (value, record) => {
          if (value === 'plan') {
            return (
              <Space direction='vertical' align='start'>
                <Tag color='blue'>{t('套餐卡券')}</Tag>
                {record.plan_grant_id ? (
                  <Text type='tertiary'>ID: {record.plan_grant_id}</Text>
                ) : null}
              </Space>
            );
          }
          return <Tag color='purple'>{t('余额卡券')}</Tag>;
        },
      },
      {
        title: t('余额额度'),
        dataIndex: 'credit_amount',
        render: (value) => (value ? value : '-'),
      },
      {
        title: t('发放数量'),
        dataIndex: 'max_redemptions',
        render: (value) => value || '-',
      },
      {
        title: t('创建时间'),
        dataIndex: 'created_at',
        render: (value) =>
          value ? formatDateTimeString(new Date(value)) : '-',
      },
      {
        title: t('备注'),
        dataIndex: 'notes',
        render: (value) => value || '-',
      },
    ],
    [t],
  );

  return (
    <div className='mt-[60px] px-2 space-y-6'>
      <Card
        title={
          <Space align='center'>
            <Title heading={5}>{t('套餐与卡券')}</Title>
          </Space>
        }
        extra={
          <Space>
            <Button
              icon={<RefreshCcw size={16} />}
              onClick={loadPlans}
              loading={planLoading}
            >
              {t('刷新')}
            </Button>
            <Button
              type='primary'
              icon={<Plus size={16} />}
              onClick={handleCreatePlan}
            >
              {t('创建套餐')}
            </Button>
          </Space>
        }
      >
        <Table
          loading={planLoading}
          dataSource={plans}
          columns={planColumns}
          rowKey='id'
          pagination={false}
          empty={<Text type='tertiary'>-</Text>}
        />
      </Card>

      <Card
        title={
          <Space align='center'>
            <Gift size={18} />
            <Title heading={5} style={{ margin: 0 }}>
              {t('卡券批次')}
            </Title>
          </Space>
        }
        extra={
          <Space>
            <Button
              icon={<RefreshCcw size={16} />}
              onClick={loadVoucherBatches}
              loading={voucherLoading}
            >
              {t('刷新')}
            </Button>
            <Button
              icon={<Gift size={16} />}
              onClick={() => setVoucherModalVisible(true)}
            >
              {t('生成卡券')}
            </Button>
          </Space>
        }
      >
        <Table
          loading={voucherLoading}
          dataSource={voucherBatches}
          columns={voucherColumns}
          rowKey='id'
          pagination={false}
          empty={<Text type='tertiary'>-</Text>}
        />
      </Card>

      <PlanFormModal
        visible={planModalVisible}
        onCancel={() => {
          setPlanModalVisible(false);
          setEditingPlan(null);
        }}
        onSubmit={handlePlanSubmit}
        initialValues={editingPlan}
        loading={planModalLoading}
      />

      <VoucherModal
        visible={voucherModalVisible}
        onCancel={() => setVoucherModalVisible(false)}
        onSubmit={handleVoucherSubmit}
        loading={voucherModalLoading}
        plans={plans}
      />
    </div>
  );
};

export default PlansPage;
