import React, { useState, useEffect } from 'react';
import { Table, Button, Modal, Form, Select, InputNumber, Toast } from '@douyinfe/semi-ui';
import { getAllRedemptionCodes, generateRedemptionCodes, revokeRedemptionCode, exportRedemptionCodes, getAllPackages } from '../../services/packageService';

const RedemptionCodeManagement = () => {
  const [codes, setCodes] = useState([]);
  const [loading, setLoading] = useState(false);
  const [packages, setPackages] = useState([]);
  const [filters, setFilters] = useState({ package_id: '', status: '' });
  const [modalVisible, setModalVisible] = useState(false);
  const [formApi, setFormApi] = useState(null);

  const fetchCodes = async () => {
    setLoading(true);
    try {
      const res = await getAllRedemptionCodes(0, 100, filters.package_id, filters.status);
      if (res.success) {
        setCodes(res.data?.items || []);
      }
    } catch (error) {
      Toast.error('获取兑换码列表失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchPackagesForSelect = async () => {
    try {
      const res = await getAllPackages(0, 100);
      if (res.success) {
        setPackages(res.data?.items || []);
      }
    } catch (error) {
      Toast.error('获取套餐列表失败');
    }
  };

  useEffect(() => {
    fetchPackagesForSelect();
    fetchCodes();
  }, []);

  useEffect(() => {
    fetchCodes();
  }, [filters]);

  const handleGenerate = () => {
    setModalVisible(true);
  };

  const handleRevoke = async (id) => {
    Modal.confirm({
      title: '确认作废',
      content: '确定要作废这个兑换码吗？',
      onOk: async () => {
        try {
          await revokeRedemptionCode(id);
          Toast.success('兑换码已作废');
          fetchCodes();
        } catch (error) {
          Toast.error('作废兑换码失败');
        }
      },
    });
  };

  const handleExport = () => {
    exportRedemptionCodes(filters.package_id, filters.status);
  };

  const handleSubmit = async (values) => {
    try {
      await generateRedemptionCodes(values.package_id, values.quantity);
      Toast.success('生成兑换码成功');
      setModalVisible(false);
      fetchCodes();
    } catch (error) {
      Toast.error('生成兑换码失败');
    }
  };

  const columns = [
    { title: '兑换码', dataIndex: 'code' },
    {
      title: '套餐',
      dataIndex: 'package',
      render: (pkg) => pkg?.name || '未知套餐',
    },
    {
      title: '状态',
      dataIndex: 'status',
      render: (status) => {
        switch (status) {
          case 1:
            return '未使用';
          case 2:
            return '已使用';
          case 3:
            return '已作废';
          default:
            return '未知';
        }
      },
    },
    { title: '使用者', dataIndex: 'used_by_user_id' },
    {
      title: '操作',
      render: (_, record) => (
        record.status === 1 ? <Button size="small" type="danger" onClick={() => handleRevoke(record.id)}>作废</Button> : null
      ),
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', gap: 12 }}>
        <Select
          placeholder='选择套餐'
          value={filters.package_id}
          onChange={(value) => setFilters({ ...filters, package_id: value })}
          style={{ width: 200 }}
          showClear
        >
          {packages.map((pkg) => (
            <Select.Option value={pkg.id} key={pkg.id}>{pkg.name}</Select.Option>
          ))}
        </Select>
        <Select
          placeholder='状态'
          value={filters.status}
          onChange={(value) => setFilters({ ...filters, status: value })}
          style={{ width: 140 }}
          showClear
        >
          <Select.Option value={1}>未使用</Select.Option>
          <Select.Option value={2}>已使用</Select.Option>
          <Select.Option value={3}>已作废</Select.Option>
        </Select>
        <Button onClick={handleGenerate}>批量生成</Button>
        <Button onClick={handleExport}>导出兑换码</Button>
      </div>
      <Table columns={columns} dataSource={codes} loading={loading} rowKey='id' />

      <Modal
        title='批量生成兑换码'
        visible={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => formApi?.submitForm()}
      >
        <Form
          getFormApi={setFormApi}
          onSubmit={handleSubmit}
        >
          <Form.Select field='package_id' label='选择套餐' rules={[{ required: true }]}>
            {packages.map((pkg) => (
              <Select.Option value={pkg.id} key={pkg.id}>{pkg.name}</Select.Option>
            ))}
          </Form.Select>
          <Form.InputNumber field='quantity' label='生成数量' rules={[{ required: true, min: 1, max: 500 }]} />
        </Form>
      </Modal>
    </div>
  );
};

export default RedemptionCodeManagement;
