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

import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import {
  Alert,
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  Modal,
  Select,
  Space,
  Table,
  Tag,
  Typography,
} from '@douyinfe/semi-ui';
import {
  IconPlus,
  IconGift,
  IconRefresh,
  IconDelete,
  IconEdit,
} from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';
import {
  createPlan,
  deletePlan,
  fetchPlans,
  generateVoucherBatch,
  listVoucherBatches,
  showError,
  showSuccess,
  updatePlan,
} from '../../helpers';

const { Title, Text } = Typography;

const cycleDefaults = {
  daily: 1,
  monthly: 30,
};

const PlanManagementPage = () => {
  const { t } = useTranslation();

  const [plans, setPlans] = useState([]);
  const [plansLoading, setPlansLoading] = useState(false);
  const [planModalVisible, setPlanModalVisible] = useState(false);
  const [editingPlan, setEditingPlan] = useState(null);
  const [planSubmitting, setPlanSubmitting] = useState(false);

  const [voucherModalVisible, setVoucherModalVisible] = useState(false);
  const [voucherResult, setVoucherResult] = useState(null);
  const [voucherSubmitting, setVoucherSubmitting] = useState(false);
  const [voucherBatches, setVoucherBatches] = useState([]);
  const [voucherLoading, setVoucherLoading] = useState(false);

  const loadPlans = useCallback(async () => {
    setPlansLoading(true);
    try {
      const res = await fetchPlans();
      if (res?.success) {
        setPlans(Array.isArray(res.data) ? res.data : []);
      } else {
        showError(res?.message || t('加载失败'));
      }
    } catch (error) {
      showError(error?.message || t('加载失败'));
    } finally {
      setPlansLoading(false);
    }
  }, [t]);

  const loadVoucherBatches = useCallback(async () => {
    setVoucherLoading(true);
    try {
      const res = await listVoucherBatches();
      if (res?.success) {
        setVoucherBatches(Array.isArray(res.data) ? res.data : []);
      } else {
        showError(res?.message || t('加载失败'));
      }
    } catch (error) {
      showError(error?.message || t('加载失败'));
    } finally {
      setVoucherLoading(false);
    }
  }, [t]);

  useEffect(() => {
    loadPlans();
    loadVoucherBatches();
  }, [loadPlans, loadVoucherBatches]);

  const openCreatePlan = () => {
    setEditingPlan(null);
    setPlanModalVisible(true);
  };

  const openEditPlan = (plan) => {
    setEditingPlan(plan);
    setPlanModalVisible(true);
  };

  const closePlanModal = useCallback(() => {
    setPlanModalVisible(false);
    setEditingPlan(null);
  }, []);

  const handlePlanSubmit = useCallback(
    async (payload) => {
      setPlanSubmitting(true);
      try {
        const res = editingPlan
          ? await updatePlan(editingPlan.id, payload)
          : await createPlan(payload);

        if (res?.success) {
          showSuccess(
            editingPlan ? t('计划更新成功') : t('计划创建成功'),
          );
          await loadPlans();
          closePlanModal();
          return true;
        }
        showError(res?.message || t('操作失败'));
      } catch (error) {
        showError(error?.message || t('操作失败'));
      } finally {
        setPlanSubmitting(false);
      }
      return false;
    },
    [editingPlan, loadPlans, closePlanModal, t],
  );

  const handleDeletePlan = useCallback(
    (plan) => {
      Modal.confirm({
        title: t('确认删除'),
        content: t('此操作不可撤销'),
        okText: t('删除'),
        okType: 'danger',
        cancelText: t('取消'),
        onOk: async () => {
          setPlansLoading(true);
          try {
            const res = await deletePlan(plan.id);
            if (res?.success) {
              showSuccess(t('计划删除成功'));
              await loadPlans();
            } else {
              showError(res?.message || t('操作失败'));
            }
          } catch (error) {
            showError(error?.message || t('操作失败'));
          } finally {
            setPlansLoading(false);
          }
        },
      });
    },
    [loadPlans, t],
  );

  const openVoucherModal = () => {
    setVoucherModalVisible(true);
  };

  const closeVoucherModal = useCallback(() => {
    setVoucherModalVisible(false);
    setVoucherResult(null);
  }, []);

  const handleVoucherSubmit = useCallback(
    async (payload) => {
      setVoucherSubmitting(true);
      try {
        const res = await generateVoucherBatch(payload);
        if (res?.success) {
          setVoucherResult(res.data || null);
          showSuccess(t('代金券生成成功'));
          await loadVoucherBatches();
          return true;
        }
        showError(res?.message || t('操作失败'));
      } catch (error) {
        showError(error?.message || t('操作失败'));
      } finally {
        setVoucherSubmitting(false);
      }
      return false;
    },
    [loadVoucherBatches, t],
  );

  const planColumns = useMemo(
    () => [
      {
        title: t('计划名称'),
        dataIndex: 'name',
        render: (value) => value || '-',
      },
      {
        title: t('描述'),
        dataIndex: 'description',
        render: (value) => value || '-',
      },
      {
        title: t('结算周期'),
        dataIndex: 'cycle',
        render: (value, record) => {
          const labels = {
            daily: t('每日'),
            monthly: t('每月'),
            custom: t('自定义'),
          };
          if (value === 'custom') {
            return `${labels[value]} · ${record?.cycle_length_days || '-'} ${t('天')}`;
          }
          return labels[value] || value || '-';
        },
      },
      {
        title: t('配额'),
        dataIndex: 'quota',
        render: (value) =>
          value === null || value === undefined ? '-' : value.toLocaleString(),
      },
      {
        title: t('额度类型'),
        dataIndex: 'quota_metric',
        render: (value) => {
          const labels = {
            requests: t('请求数'),
            tokens: t('令牌数'),
          };
          return labels[value] || '-';
        },
      },
      {
        title: t('结转策略'),
        dataIndex: 'rollover_policy',
        render: (value) => {
          const labels = {
            none: t('不结转'),
            carry_all: t('全部结转'),
            cap: t('限额结转'),
          };
          return labels[value] || '-';
        },
      },
      {
        title: t('价格'),
        dataIndex: 'price',
        render: (value) => {
          if (value === null || value === undefined) return '-';
          const numberValue = Number(value);
          if (Number.isNaN(numberValue)) {
            return '-';
          }
          return numberValue.toFixed(2);
        },
      },
      {
        title: t('状态'),
        dataIndex: 'status',
        render: (value) => {
          const labels = {
            draft: t('草稿'),
            active: t('启用'),
            archived: t('已归档'),
          };
          const colors = {
            draft: 'amber',
            active: 'green',
            archived: 'grey',
          };
          return (
            <Tag color={colors[value] || 'blue'} size='small'>
              {labels[value] || value || '-'}
            </Tag>
          );
        },
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
              icon={<IconEdit size='small' />}
              onClick={() => openEditPlan(record)}
            >
              {t('编辑')}
            </Button>
            <Button
              size='small'
              theme='borderless'
              type='danger'
              icon={<IconDelete size='small' />}
              onClick={() => handleDeletePlan(record)}
            >
              {t('删除')}
            </Button>
          </Space>
        ),
      },
    ],
    [t, handleDeletePlan],
  );

  const formatDateTime = useCallback(
    (value) => {
      if (!value) return '-';
      const date = new Date(value);
      if (Number.isNaN(date.getTime())) {
        return '-';
      }
      return date.toLocaleString();
    },
    [],
  );

  const voucherColumns = useMemo(
    () => [
      {
        title: t('批次标签'),
        dataIndex: 'batch_label',
        render: (value) => value || '-',
      },
      {
        title: t('发放类型'),
        dataIndex: 'grant_type',
        render: (value) => {
          const labels = {
            credit: t('余额发放'),
            plan: t('计划发放'),
          };
          return labels[value] || value || '-';
        },
      },
      {
        title: t('代金券面额'),
        dataIndex: 'credit_amount',
        render: (value, record) =>
          record?.grant_type === 'credit'
            ? value?.toLocaleString?.() ?? value ?? '-'
            : '-',
      },
      {
        title: t('关联计划'),
        dataIndex: 'plan_grant_id',
        render: (_, record) => {
          if (record?.grant_type !== 'plan') return '-';
          if (record?.metadata && record.metadata.plan_name) {
            return record.metadata.plan_name;
          }
          if (record?.plan_grant_id) {
            return record.plan_grant_id;
          }
          return '-';
        },
      },
      {
        title: t('最大兑换次数'),
        dataIndex: 'max_redemptions',
        render: (value) => (value ? value : '-'),
      },
      {
        title: t('单用户上限'),
        dataIndex: 'max_per_subject',
        render: (value) => (value ? value : '-'),
      },
      {
        title: t('有效期'),
        dataIndex: 'valid_until',
        render: (_, record) => {
          if (!record?.valid_from && !record?.valid_until) {
            return '-';
          }
          return `${formatDateTime(record?.valid_from)} → ${formatDateTime(record?.valid_until)}`;
        },
      },
      {
        title: t('创建人'),
        dataIndex: 'created_by',
        render: (value) => value || '-',
      },
      {
        title: t('备注'),
        dataIndex: 'notes',
        ellipsis: true,
        render: (value) => value || '-',
      },
    ],
    [formatDateTime, t],
  );

  return (
    <div className='mt-[60px] px-2 flex flex-col gap-4'>
      <Card
        title={
          <Space>
            <Title heading={5} className='m-0'>
              {t('计划管理')}
            </Title>
          </Space>
        }
        extra={
          <Space>
            <Button
              icon={<IconRefresh />}
              theme='light'
              onClick={loadPlans}
              loading={plansLoading}
            >
              {t('刷新')}
            </Button>
            <Button
              icon={<IconPlus />}
              theme='solid'
              type='primary'
              onClick={openCreatePlan}
            >
              {t('新建计划')}
            </Button>
          </Space>
        }
      >
        <Table
          rowKey='id'
          loading={plansLoading}
          dataSource={plans}
          columns={planColumns}
          pagination={false}
          empty={t('暂无数据')}
        />
      </Card>

      <Card
        title={
          <Space>
            <Title heading={5} className='m-0'>
              {t('代金券批次')}
            </Title>
          </Space>
        }
        extra={
          <Space>
            <Button
              icon={<IconRefresh />}
              theme='light'
              onClick={loadVoucherBatches}
              loading={voucherLoading}
            >
              {t('刷新')}
            </Button>
            <Button
              icon={<IconGift />}
              theme='solid'
              type='primary'
              onClick={openVoucherModal}
            >
              {t('生成代金券')}
            </Button>
          </Space>
        }
      >
        <Table
          rowKey='id'
          loading={voucherLoading}
          dataSource={voucherBatches}
          columns={voucherColumns}
          pagination={false}
          empty={t('暂无数据')}
        />
      </Card>

      <PlanFormModal
        visible={planModalVisible}
        onCancel={closePlanModal}
        onSubmit={handlePlanSubmit}
        confirmLoading={planSubmitting}
        initialValues={editingPlan}
        t={t}
      />

      <VoucherGeneratorModal
        visible={voucherModalVisible}
        onCancel={closeVoucherModal}
        onSubmit={handleVoucherSubmit}
        confirmLoading={voucherSubmitting}
        plans={plans}
        result={voucherResult}
        t={t}
      />
    </div>
  );
};

const PlanFormModal = ({
  visible,
  onCancel,
  onSubmit,
  confirmLoading,
  initialValues,
  t,
}) => {
  const formApiRef = useRef(null);
  const isEdit = Boolean(initialValues?.id);

  const defaultValues = useMemo(
    () => ({
      name: '',
      description: '',
      cycle: 'monthly',
      cycle_length_days: cycleDefaults.monthly,
      quota: 0,
      quota_metric: 'requests',
      rollover_policy: 'none',
      price: null,
      status: 'active',
    }),
    [],
  );

  useEffect(() => {
    if (!formApiRef.current) return;

    if (visible) {
      const nextValues = {
        ...defaultValues,
        ...initialValues,
      };
      if (!nextValues.cycle) {
        nextValues.cycle = 'monthly';
      }
      if (nextValues.cycle !== 'custom') {
        nextValues.cycle_length_days =
          cycleDefaults[nextValues.cycle] || nextValues.cycle_length_days || 30;
      }
      formApiRef.current.setValues(nextValues);
    } else {
      formApiRef.current.reset();
    }
  }, [visible, initialValues, defaultValues]);

  const handleCycleChange = useCallback((value) => {
    if (!formApiRef.current) return;
    if (value !== 'custom') {
      formApiRef.current.setValue(
        'cycle_length_days',
        cycleDefaults[value] || cycleDefaults.monthly,
      );
    }
  }, []);

  const handleSubmit = useCallback(
    async (values) => {
      const payload = { ...values };

      payload.name = (payload.name || '').trim();
      payload.description = payload.description?.trim();
      payload.quota_metric = payload.quota_metric || undefined;
      payload.rollover_policy = payload.rollover_policy || undefined;
      payload.status = payload.status || 'active';

      if (payload.cycle !== 'custom') {
        payload.cycle_length_days = cycleDefaults[payload.cycle] || 0;
      } else {
        payload.cycle_length_days = Number(payload.cycle_length_days) || 0;
      }

      if (payload.cycle_length_days <= 0) {
        showError(t('周期天数不能为空'));
        return;
      }

      if (payload.description === '') {
        delete payload.description;
      }

      if (payload.quota === '' || payload.quota === undefined) {
        delete payload.quota;
      } else {
        payload.quota = Number(payload.quota);
      }

      if (payload.price === '' || payload.price === null) {
        delete payload.price;
      } else {
        payload.price = Number(payload.price);
      }

      const success = await onSubmit(payload);
      if (success) {
        formApiRef.current?.reset();
      }
    },
    [onSubmit, t],
  );

  return (
    <Modal
      title={isEdit ? t('编辑计划') : t('新建计划')}
      visible={visible}
      onCancel={onCancel}
      footer={null}
      maskClosable={false}
    >
      <Form
        getFormApi={(api) => (formApiRef.current = api)}
        initValues={defaultValues}
        onSubmit={handleSubmit}
      >
        {({ values }) => (
          <Space vertical style={{ width: '100%' }}>
            <Form.Input
              field='name'
              label={t('计划名称')}
              placeholder={t('请输入名称')}
              rules={[{ required: true, message: t('请输入名称') }]}
              showClear
            />
            <Form.TextArea
              field='description'
              label={t('描述')}
              placeholder={t('请输入备注信息')}
              maxCount={512}
              showClear
            />
            <Form.Select
              field='cycle'
              label={t('结算周期')}
              optionList={[
                { label: t('每日'), value: 'daily' },
                { label: t('每月'), value: 'monthly' },
                { label: t('自定义'), value: 'custom' },
              ]}
              rules={[{ required: true, message: t('请选择') }]}
              onChange={handleCycleChange}
            />
            <Form.InputNumber
              field='cycle_length_days'
              label={t('周期天数')}
              min={1}
              step={1}
              disabled={values.cycle !== 'custom'}
              rules={[
                {
                  required: values.cycle === 'custom',
                  message: t('请输入周期天数'),
                },
              ]}
              style={{ width: '100%' }}
            />
            <Form.InputNumber
              field='quota'
              label={t('配额')}
              min={0}
              step={100}
              precision={0}
              rules={[
                {
                  required: true,
                  message: t('请输入额度'),
                },
              ]}
              style={{ width: '100%' }}
            />
            <Form.Select
              field='quota_metric'
              label={t('额度类型')}
              optionList={[
                { label: t('请求数'), value: 'requests' },
                { label: t('令牌数'), value: 'tokens' },
              ]}
              placeholder={t('请选择')}
            />
            <Form.Select
              field='rollover_policy'
              label={t('结转策略')}
              optionList={[
                { label: t('不结转'), value: 'none' },
                { label: t('全部结转'), value: 'carry_all' },
                { label: t('限额结转'), value: 'cap' },
              ]}
              placeholder={t('请选择')}
            />
            <Form.InputNumber
              field='price'
              label={t('价格')}
              min={0}
              step={0.01}
              precision={2}
              style={{ width: '100%' }}
            />
            <Form.Select
              field='status'
              label={t('状态')}
              optionList={[
                { label: t('草稿'), value: 'draft' },
                { label: t('启用'), value: 'active' },
                { label: t('已归档'), value: 'archived' },
              ]}
              rules={[{ required: true, message: t('请选择') }]}
            />

            <Space style={{ justifyContent: 'flex-end' }}>
              <Button onClick={onCancel}>{t('取消')}</Button>
              <Button
                theme='solid'
                type='primary'
                loading={confirmLoading}
                onClick={() => formApiRef.current?.submitForm()}
              >
                {t('提交')}
              </Button>
            </Space>
          </Space>
        )}
      </Form>
    </Modal>
  );
};

const VoucherGeneratorModal = ({
  visible,
  onCancel,
  onSubmit,
  confirmLoading,
  plans,
  result,
  t,
}) => {
  const formApiRef = useRef(null);

  const defaultValues = useMemo(
    () => ({
      count: 10,
      prefix: '',
      grant_type: 'credit',
      credit_amount: 1000,
      plan_id: undefined,
      expire_days: 30,
      note: '',
    }),
    [],
  );

  useEffect(() => {
    if (!formApiRef.current) return;
    if (visible) {
      formApiRef.current.setValues(defaultValues);
    } else {
      formApiRef.current.reset();
    }
  }, [visible, defaultValues]);

  const planOptions = useMemo(
    () =>
      Array.isArray(plans)
        ? plans.map((plan) => ({ label: plan.name, value: plan.id }))
        : [],
    [plans],
  );

  const handleSubmit = useCallback(
    async (values) => {
      const payload = {
        count: Number(values.count) || 0,
        grant_type: values.grant_type,
      };

      if (values.prefix) {
        payload.prefix = values.prefix.trim();
      }

      if (values.expire_days) {
        payload.expire_days = Number(values.expire_days);
      }

      if (values.note) {
        payload.note = values.note.trim();
      }

      if (values.grant_type === 'credit') {
        payload.credit_amount = Number(values.credit_amount) || 0;
      } else if (values.grant_type === 'plan') {
        payload.plan_id = values.plan_id;
      }

      const success = await onSubmit(payload);
      if (success) {
        formApiRef.current?.setValue('count', values.count);
      }
    },
    [onSubmit],
  );

  return (
    <Modal
      title={t('生成代金券')}
      visible={visible}
      onCancel={onCancel}
      footer={null}
      maskClosable={false}
    >
      <Form
        getFormApi={(api) => (formApiRef.current = api)}
        initValues={defaultValues}
        onSubmit={handleSubmit}
      >
        {({ values }) => (
          <Space vertical style={{ width: '100%' }}>
            <Form.InputNumber
              field='count'
              label={t('生成数量')}
              min={1}
              max={1000}
              rules={[{ required: true, message: t('请输入数量') }]}
              style={{ width: '100%' }}
            />
            <Form.Input
              field='prefix'
              label={t('批次前缀')}
              placeholder={t('可选，支持字母数字')}
              maxLength={16}
              showClear
            />
            <Form.Select
              field='grant_type'
              label={t('发放类型')}
              optionList={[
                { label: t('余额发放'), value: 'credit' },
                { label: t('计划发放'), value: 'plan' },
              ]}
              rules={[{ required: true, message: t('请选择') }]}
            />
            {values.grant_type === 'credit' && (
              <Form.InputNumber
                field='credit_amount'
                label={t('代金券面额')}
                min={1}
                step={100}
                rules={[{ required: true, message: t('请输入额度') }]}
                style={{ width: '100%' }}
              />
            )}
            {values.grant_type === 'plan' && (
              <Form.Select
                field='plan_id'
                label={t('关联计划')}
                optionList={planOptions}
                placeholder={t('请选择计划')}
                rules={[{ required: true, message: t('请选择计划') }]}
              />
            )}
            <Form.InputNumber
              field='expire_days'
              label={t('有效天数')}
              min={1}
              max={3650}
              placeholder={t('可选，不填表示不限')}
              style={{ width: '100%' }}
            />
            <Form.TextArea
              field='note'
              label={t('备注')}
              placeholder={t('可选，便于区分用途')}
              maxCount={128}
              showClear
            />

            <Space style={{ justifyContent: 'flex-end' }}>
              <Button onClick={onCancel}>{t('取消')}</Button>
              <Button
                theme='solid'
                type='primary'
                loading={confirmLoading}
                onClick={() => formApiRef.current?.submitForm()}
              >
                {t('提交')}
              </Button>
            </Space>

            {result?.codes?.length ? (
              <Alert
                type='success'
                title={t('生成结果')}
                description={
                  <div className='mt-2'>
                    <Text>
                      {t('共生成 {{count}} 张代金券', {
                        count: result.codes.length,
                      })}
                    </Text>
                    <Input.TextArea
                      value={result.codes.join('\n')}
                      rows={6}
                      readOnly
                      className='mt-2'
                    />
                  </div>
                }
              />
            ) : null}
          </Space>
        )}
      </Form>
    </Modal>
  );
};

export default PlanManagementPage;
