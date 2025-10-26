import React, { useMemo } from 'react';
import { Modal, Table, Space, Select, Tag, Typography, Card } from '@douyinfe/semi-ui';
import { timestamp2string } from '../../../helpers';

const WINDOW_OPTIONS = [
  { label: '1小时', value: '1h' },
  { label: '24小时', value: '24h' },
  { label: '7天', value: '7d' },
  { label: '30天', value: '30d' },
  { label: '365天', value: '365d' },
  { label: '全部', value: 'all_time' },
];

const IpUsageModal = ({
  visible,
  title,
  loading,
  data,
  windowValue,
  onWindowChange,
  onClose,
  t,
}) => {
  const columns = useMemo(
    () => [
      {
        title: t('IP 地址'),
        dataIndex: 'ip',
        render: (text) => <Typography.Text code>{text}</Typography.Text>,
      },
      {
        title: t('请求次数'),
        dataIndex: 'request_count',
        render: (value) => Number(value || 0).toLocaleString(),
      },
      {
        title: t('首次出现'),
        dataIndex: 'first_seen_at',
        render: (value) => (value ? timestamp2string(value) : '-'),
      },
      {
        title: t('最近出现'),
        dataIndex: 'last_seen_at',
        render: (value) => (value ? timestamp2string(value) : '-'),
      },
    ],
    [t],
  );

  const summaryItems = useMemo(
    () => [
      {
        title: t('唯一 IP 数'),
        value: data?.unique_count || 0,
      },
      {
        title: t('总请求数'),
        value: data?.total_requests || 0,
      },
    ],
    [data, t],
  );

  return (
    <Modal
      title={title}
      visible={visible}
      onCancel={onClose}
      footer={null}
      closeOnEsc
      centered
      width={720}
    >
      <Space direction='vertical' className='w-full' spacing='medium'>
        <Space align='center' wrap className='justify-between w-full'>
          <Typography.Text type='tertiary'>
            {data?.subject_label ? `${t('目标')}: ${data.subject_label}` : ''}
          </Typography.Text>
          <Select
            value={windowValue}
            onChange={onWindowChange}
            optionList={WINDOW_OPTIONS.map((option) => ({
              label: t(option.label),
              value: option.value,
            }))}
            style={{ width: 160 }}
          />
        </Space>

        <Space wrap spacing='medium'>
          {summaryItems.map((item) => (
            <Card
              key={item.title}
              bodyStyle={{ padding: '12px 16px' }}
              style={{
                minWidth: 160,
                borderRadius: 12,
                border: '1px solid var(--semi-color-border)',
              }}
            >
              <Space direction='vertical' spacing={4}>
                <Typography.Text type='tertiary'>{item.title}</Typography.Text>
                <Typography.Title heading={4} style={{ margin: 0 }}>
                  {Number(item.value || 0).toLocaleString()}
                </Typography.Title>
              </Space>
            </Card>
          ))}
          <Tag color='white'>
            {t('当前窗口')}: {t(WINDOW_OPTIONS.find((opt) => opt.value === windowValue)?.label || '')}
          </Tag>
        </Space>

        <Table
          loading={loading}
          columns={columns}
          dataSource={data?.items || []}
          pagination={{ pageSize: 10 }}
          scroll={{ y: 360 }}
          empty={<Typography.Text type='tertiary'>{t('暂无数据')}</Typography.Text>}
        />
      </Space>
    </Modal>
  );
};

export default IpUsageModal;
