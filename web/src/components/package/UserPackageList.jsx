import React, { useState, useEffect } from 'react';
import { Table, Button, Card, Modal, Form, Input, Toast } from '@douyinfe/semi-ui';
import { getUserPackages, redeemPackageCode } from '../../services/packageService';
import { formatDateTimeString } from '../../helpers';

const UserPackageList = () => {
  const [packages, setPackages] = useState([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [formApi, setFormApi] = useState(null);

  const fetchUserPackages = async () => {
    setLoading(true);
    try {
      const res = await getUserPackages();
      if (res.success) {
        setPackages(res.data || []);
      }
    } catch (error) {
      Toast.error('获取积分包列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUserPackages();
  }, []);

  const handleRedeem = () => {
    setModalVisible(true);
  };

  const handleSubmit = async (values) => {
    try {
      const res = await redeemPackageCode(values.code);
      if (res.success) {
        Toast.success(res.message || '兑换成功');
        setModalVisible(false);
        fetchUserPackages();
      } else {
        Toast.error(res.message || '兑换失败');
      }
    } catch (error) {
      Toast.error('兑换失败');
    }
  };

  const getStatusText = (pkg) => {
    if (pkg.status === 3 || (pkg.expire_at && new Date(pkg.expire_at) < new Date())) {
      return <span style={{ color: 'red' }}>已过期</span>;
    }
    if (pkg.status === 2 || pkg.token_quota <= 0) {
      return <span style={{ color: 'orange' }}>已用完</span>;
    }
    return <span style={{ color: 'green' }}>有效</span>;
  };

  const columns = [
    {
      title: '套餐名称',
      dataIndex: 'package',
      render: (pkg) => pkg?.name || '未知套餐',
    },
    {
      title: '剩余额度',
      dataIndex: 'token_quota',
      render: (quota, record) => `${quota} / ${record.initial_quota}`,
    },
    {
      title: '过期时间',
      dataIndex: 'expire_at',
      render: (time) => formatDateTimeString(time),
    },
    {
      title: '状态',
      render: (_, record) => getStatusText(record),
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Button onClick={handleRedeem} type='primary'>兑换码兑换</Button>
      </div>
      <Table columns={columns} dataSource={packages} loading={loading} rowKey='id' />

      <Modal
        title='兑换套餐'
        visible={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => formApi?.submitForm()}
      >
        <Form
          getFormApi={setFormApi}
          onSubmit={handleSubmit}
        >
          <Form.Input 
            field='code' 
            label='兑换码' 
            rules={[{ required: true }]} 
            placeholder='请输入兑换码'
          />
        </Form>
      </Modal>
    </div>
  );
};

export default UserPackageList;
