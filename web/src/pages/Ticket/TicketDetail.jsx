import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  Card,
  Button,
  Tag,
  Divider,
  Form,
  TextArea,
  Toast,
  Space,
  Descriptions,
  Modal,
} from '@douyinfe/semi-ui';
import { API } from '../../helpers';
import { isAdmin } from '../../helpers';
import './Ticket.css';

const TicketDetail = () => {
  const { t } = useTranslation();
  const { id } = useParams();
  const navigate = useNavigate();
  const [ticket, setTicket] = useState(null);
  const [loading, setLoading] = useState(false);
  const [replyLoading, setReplyLoading] = useState(false);
  const [reply, setReply] = useState('');

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

  const loadTicket = async () => {
    setLoading(true);
    try {
      const res = await API.get(`/api/ticket/${id}`);
      const { success, data, message } = res.data;

      if (success) {
        setTicket(data);
      } else {
        Toast.error(message || t('加载失败'));
        navigate('/ticket');
      }
    } catch (error) {
      Toast.error(t('加载失败') + ': ' + error.message);
      navigate('/ticket');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadTicket();
  }, [id]);

  const handleReply = async () => {
    if (!reply || reply.trim().length === 0) {
      Toast.error(t('回复内容不能为空'));
      return;
    }

    setReplyLoading(true);
    try {
      const res = await API.put(`/api/ticket/${id}/reply`, {
        reply: reply.trim(),
      });
      const { success, message } = res.data;

      if (success) {
        Toast.success(t('回复成功'));
        setReply('');
        loadTicket();
      } else {
        Toast.error(message || t('回复失败'));
      }
    } catch (error) {
      Toast.error(t('回复失败') + ': ' + error.message);
    } finally {
      setReplyLoading(false);
    }
  };

  const handleStatusChange = async (newStatus) => {
    try {
      const res = await API.put(`/api/ticket/${id}/status`, {
        status: newStatus,
      });
      const { success, message } = res.data;

      if (success) {
        Toast.success(t('状态更新成功'));
        loadTicket();
      } else {
        Toast.error(message || t('状态更新失败'));
      }
    } catch (error) {
      Toast.error(t('状态更新失败') + ': ' + error.message);
    }
  };

  const handleDelete = () => {
    Modal.confirm({
      title: t('确认删除'),
      content: t('确定要删除此工单吗？'),
      onOk: async () => {
        try {
          const res = await API.delete(`/api/ticket/${id}`);
          const { success, message } = res.data;

          if (success) {
            Toast.success(t('删除成功'));
            navigate('/ticket');
          } else {
            Toast.error(message || t('删除失败'));
          }
        } catch (error) {
          Toast.error(t('删除失败') + ': ' + error.message);
        }
      },
    });
  };

  if (loading || !ticket) {
    return (
      <div className='ticket-container'>
        <Card loading={true} />
      </div>
    );
  }

  const statusOptions = [
    { label: t('待处理'), value: 'pending' },
    { label: t('处理中'), value: 'processing' },
    { label: t('已解决'), value: 'resolved' },
    { label: t('已关闭'), value: 'closed' },
  ];

  const data = [
    {
      key: t('工单编号'),
      value: ticket.id,
    },
    isAdmin() && {
      key: t('创建者'),
      value: ticket.username,
    },
    {
      key: t('状态'),
      value: (
        <Tag color={getStatusColor(ticket.status)}>
          {t(
            ticket.status === 'pending'
              ? '待处理'
              : ticket.status === 'processing'
              ? '处理中'
              : ticket.status === 'resolved'
              ? '已解决'
              : '已关闭'
          )}
        </Tag>
      ),
    },
    {
      key: t('优先级'),
      value: (
        <Tag color={getPriorityColor(ticket.priority)}>
          {t(
            ticket.priority === 'urgent'
              ? '紧急'
              : ticket.priority === 'high'
              ? '高'
              : ticket.priority === 'medium'
              ? '中'
              : '低'
          )}
        </Tag>
      ),
    },
    {
      key: t('分类'),
      value: t(
        ticket.category === 'technical'
          ? '技术支持'
          : ticket.category === 'account'
          ? '账号问题'
          : ticket.category === 'feature'
          ? '功能建议'
          : '其他'
      ),
    },
    {
      key: t('创建时间'),
      value: new Date(ticket.created_at).toLocaleString(),
    },
    ticket.replied_at && {
      key: t('回复时间'),
      value: new Date(ticket.replied_at).toLocaleString(),
    },
  ].filter(Boolean);

  return (
    <div className='ticket-container'>
      <Card
        title={t('工单详情')}
        headerExtraContent={
          <Space>
            {isAdmin() && ticket.status !== 'closed' && (
              <>
                {statusOptions.map((option) => (
                  <Button
                    key={option.value}
                    size='small'
                    type={ticket.status === option.value ? 'primary' : 'secondary'}
                    onClick={() => handleStatusChange(option.value)}
                    disabled={ticket.status === option.value}
                  >
                    {option.label}
                  </Button>
                ))}
              </>
            )}
            <Button type='danger' onClick={handleDelete}>
              {t('删除')}
            </Button>
            <Button onClick={() => navigate('/ticket')}>{t('返回')}</Button>
          </Space>
        }
      >
        <Descriptions data={data} row />

        <Divider margin='24px' />

        <div className='ticket-content-section'>
          <h3>{t('标题')}</h3>
          <div className='ticket-content-box'>{ticket.title}</div>
        </div>

        <div className='ticket-content-section'>
          <h3>{t('内容')}</h3>
          <div className='ticket-content-box' style={{ whiteSpace: 'pre-wrap' }}>
            {ticket.content}
          </div>
        </div>

        {ticket.admin_reply && (
          <>
            <Divider margin='24px' />
            <div className='ticket-content-section ticket-reply-section'>
              <h3>{t('管理员回复')}</h3>
              <div className='ticket-content-box' style={{ whiteSpace: 'pre-wrap' }}>
                {ticket.admin_reply}
              </div>
            </div>
          </>
        )}

        {isAdmin() && ticket.status !== 'closed' && (
          <>
            <Divider margin='24px' />
            <div className='ticket-reply-form'>
              <h3>{t('添加回复')}</h3>
              <TextArea
                placeholder={t('请输入回复内容')}
                rows={6}
                value={reply}
                onChange={setReply}
                maxLength={5000}
                showClear
              />
              <Button
                theme='solid'
                type='primary'
                onClick={handleReply}
                loading={replyLoading}
                style={{ marginTop: 10 }}
              >
                {t('提交回复')}
              </Button>
            </div>
          </>
        )}
      </Card>
    </div>
  );
};

export default TicketDetail;
