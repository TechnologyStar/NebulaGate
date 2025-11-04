import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Table, Button, Select, Tag, Space, Modal, Toast } from '@douyinfe/semi-ui';
import { API } from '../../helpers';
import { isAdmin } from '../../helpers';
import './Ticket.css';

const TicketList = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [tickets, setTickets] = useState([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [statusFilter, setStatusFilter] = useState('');

  const statusOptions = [
    { label: t('全部'), value: '' },
    { label: t('待处理'), value: 'pending' },
    { label: t('处理中'), value: 'processing' },
    { label: t('已解决'), value: 'resolved' },
    { label: t('已关闭'), value: 'closed' },
  ];

  const getStatusColor = (status) => {
    const colorMap = {
      pending: 'amber',
      processing: 'blue',
      resolved: 'green',
      closed: 'grey',
    };
    return colorMap[status] || 'grey';
  };

  const getPriorityColor = (priority) => {
    const colorMap = {
      urgent: 'red',
      high: 'orange',
      medium: 'amber',
      low: 'green',
    };
    return colorMap[priority] || 'grey';
  };

  const loadTickets = async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        page: page.toString(),
        page_size: pageSize.toString(),
      });
      if (statusFilter) {
        params.append('status', statusFilter);
      }

      const res = await API.get(`/api/ticket/list?${params.toString()}`);
      const { success, data, total: totalCount, message } = res.data;

      if (success) {
        setTickets(data || []);
        setTotal(totalCount || 0);
      } else {
        Toast.error(message || t('加载失败'));
      }
    } catch (error) {
      Toast.error(t('加载失败') + ': ' + error.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadTickets();
  }, [page, pageSize, statusFilter]);

  const handleDelete = async (id) => {
    Modal.confirm({
      title: t('确认删除'),
      content: t('确定要删除此工单吗？'),
      onOk: async () => {
        try {
          const res = await API.delete(`/api/ticket/${id}`);
          const { success, message } = res.data;

          if (success) {
            Toast.success(t('删除成功'));
            loadTickets();
          } else {
            Toast.error(message || t('删除失败'));
          }
        } catch (error) {
          Toast.error(t('删除失败') + ': ' + error.message);
        }
      },
    });
  };

  const handleStatusChange = async (id, newStatus) => {
    try {
      const res = await API.put(`/api/ticket/${id}/status`, {
        status: newStatus,
      });
      const { success, message } = res.data;

      if (success) {
        Toast.success(t('状态更新成功'));
        loadTickets();
      } else {
        Toast.error(message || t('状态更新失败'));
      }
    } catch (error) {
      Toast.error(t('状态更新失败') + ': ' + error.message);
    }
  };

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
    },
    {
      title: t('标题'),
      dataIndex: 'title',
      render: (text, record) => (
        <a
          onClick={() => navigate(`/ticket/${record.id}`)}
          style={{ cursor: 'pointer', color: 'var(--semi-color-primary)' }}
        >
          {text}
        </a>
      ),
    },
    isAdmin() && {
      title: t('创建者'),
      dataIndex: 'username',
      width: 120,
    },
    {
      title: t('状态'),
      dataIndex: 'status',
      width: 100,
      render: (status) => (
        <Tag color={getStatusColor(status)}>
          {t(
            status === 'pending'
              ? '待处理'
              : status === 'processing'
              ? '处理中'
              : status === 'resolved'
              ? '已解决'
              : '已关闭'
          )}
        </Tag>
      ),
    },
    {
      title: t('优先级'),
      dataIndex: 'priority',
      width: 100,
      render: (priority) => (
        <Tag color={getPriorityColor(priority)}>
          {t(
            priority === 'urgent'
              ? '紧急'
              : priority === 'high'
              ? '高'
              : priority === 'medium'
              ? '中'
              : '低'
          )}
        </Tag>
      ),
    },
    {
      title: t('分类'),
      dataIndex: 'category',
      width: 120,
      render: (category) =>
        t(
          category === 'technical'
            ? '技术支持'
            : category === 'account'
            ? '账号问题'
            : category === 'feature'
            ? '功能建议'
            : '其他'
        ),
    },
    {
      title: t('创建时间'),
      dataIndex: 'created_at',
      width: 180,
      render: (time) => new Date(time).toLocaleString(),
    },
    {
      title: t('操作'),
      width: 200,
      render: (_, record) => (
        <Space>
          <Button
            size='small'
            onClick={() => navigate(`/ticket/${record.id}`)}
          >
            {t('查看')}
          </Button>
          {isAdmin() && record.status !== 'closed' && (
            <Button
              size='small'
              type='secondary'
              onClick={() => handleStatusChange(record.id, 'closed')}
            >
              {t('关闭')}
            </Button>
          )}
          <Button
            size='small'
            type='danger'
            onClick={() => handleDelete(record.id)}
          >
            {t('删除')}
          </Button>
        </Space>
      ),
    },
  ].filter(Boolean);

  return (
    <div className='ticket-container'>
      <div className='ticket-header'>
        <h2>{t('工单系统')}</h2>
        <Space>
          <Select
            value={statusFilter}
            onChange={setStatusFilter}
            style={{ width: 150 }}
            placeholder={t('筛选状态')}
          >
            {statusOptions.map((option) => (
              <Select.Option key={option.value} value={option.value}>
                {option.label}
              </Select.Option>
            ))}
          </Select>
          <Button
            theme='solid'
            type='primary'
            onClick={() => navigate('/ticket/create')}
          >
            {t('创建工单')}
          </Button>
        </Space>
      </div>

      <Table
        columns={columns}
        dataSource={tickets}
        loading={loading}
        pagination={{
          currentPage: page,
          pageSize: pageSize,
          total: total,
          onPageChange: setPage,
          onPageSizeChange: (size) => {
            setPageSize(size);
            setPage(1);
          },
          showSizeChanger: true,
        }}
        rowKey='id'
      />
    </div>
  );
};

export default TicketList;
