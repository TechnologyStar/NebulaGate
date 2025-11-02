import React, { useState, useEffect } from 'react';
import { Table, Button, Modal, Form, InputNumber, Input, Select, Toast } from '@douyinfe/semi-ui';
import { getAllPackages, createPackage, updatePackage, deletePackage } from '../../services/packageService';

const PackageManagement = () => {
  const [packages, setPackages] = useState([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingPackage, setEditingPackage] = useState(null);
  const [formApi, setFormApi] = useState(null);

  const fetchPackages = async () => {
    setLoading(true);
    try {
      const res = await getAllPackages(0, 100);
      if (res.success) {
        setPackages(res.data?.items || []);
      }
    } catch (error) {
      Toast.error('获取套餐列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPackages();
  }, []);

  const handleCreate = () => {
    setEditingPackage(null);
    setModalVisible(true);
  };

  const handleEdit = (pkg) => {
    setEditingPackage(pkg);
    setModalVisible(true);
  };

  const handleSubmit = async (values) => {
    try {
      const data = {
        name: values.name,
        description: values.description || '',
        token_quota: values.token_quota,
        model_scope: values.model_scope || '[]',
        validity_days: values.validity_days,
        price: values.price || 0,
        status: values.status || 1,
      };

      if (editingPackage) {
        await updatePackage(editingPackage.id, data);
        Toast.success('更新成功');
      } else {
        await createPackage(data);
        Toast.success('创建成功');
      }
      setModalVisible(false);
      fetchPackages();
    } catch (error) {
      Toast.error(editingPackage ? '更新失败' : '创建失败');
    }
  };

  const handleDelete = async (id) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除这个套餐吗？',
      onOk: async () => {
        try {
          await deletePackage(id);
          Toast.success('删除成功');
          fetchPackages();
        } catch (error) {
          Toast.error('删除失败');
        }
      },
    });
  };

  const columns = [
    { title: 'ID', dataIndex: 'id' },
    { title: '套餐名称', dataIndex: 'name' },
    { title: 'Token额度', dataIndex: 'token_quota' },
    { title: '有效期(天)', dataIndex: 'validity_days' },
    { title: '价格', dataIndex: 'price' },
    {
      title: '状态',
      dataIndex: 'status',
      render: (status) => (status === 1 ? '启用' : '禁用'),
    },
    {
      title: '操作',
      render: (_, record) => (
        <>
          <Button size="small" onClick={() => handleEdit(record)}>编辑</Button>
          <Button size="small" type="danger" onClick={() => handleDelete(record.id)} style={{ marginLeft: 8 }}>删除</Button>
        </>
      ),
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Button onClick={handleCreate}>创建套餐</Button>
      </div>
      <Table columns={columns} dataSource={packages} loading={loading} rowKey="id" />
      
      <Modal
        title={editingPackage ? '编辑套餐' : '创建套餐'}
        visible={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => formApi?.submitForm()}
      >
        <Form
          getFormApi={setFormApi}
          onSubmit={handleSubmit}
          initValues={editingPackage || {}}
        >
          <Form.Input field="name" label="套餐名称" rules={[{ required: true }]} />
          <Form.TextArea field="description" label="描述" />
          <Form.InputNumber field="token_quota" label="Token额度" rules={[{ required: true }]} />
          <Form.TextArea field="model_scope" label="模型范围(JSON)" placeholder='["gpt-4", "gpt-3.5-turbo"]' />
          <Form.InputNumber field="validity_days" label="有效期(天)" rules={[{ required: true }]} />
          <Form.InputNumber field="price" label="价格" />
          <Form.Select field="status" label="状态">
            <Select.Option value={1}>启用</Select.Option>
            <Select.Option value={2}>禁用</Select.Option>
          </Form.Select>
        </Form>
      </Modal>
    </div>
  );
};

export default PackageManagement;
