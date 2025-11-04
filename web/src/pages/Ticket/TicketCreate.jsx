import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  Form,
  Button,
  Card,
  Toast,
  Select,
  TextArea,
  Input,
} from '@douyinfe/semi-ui';
import { API } from '../../helpers';
import './Ticket.css';

const TicketCreate = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);

  const priorityOptions = [
    { label: t('低'), value: 'low' },
    { label: t('中'), value: 'medium' },
    { label: t('高'), value: 'high' },
    { label: t('紧急'), value: 'urgent' },
  ];

  const categoryOptions = [
    { label: t('技术支持'), value: 'technical' },
    { label: t('账号问题'), value: 'account' },
    { label: t('功能建议'), value: 'feature' },
    { label: t('其他'), value: 'other' },
  ];

  const handleSubmit = async (values) => {
    // Client-side validation
    if (!values.title || values.title.trim().length === 0) {
      Toast.error(t('标题不能为空'));
      return;
    }

    if (values.title.trim().length > 200) {
      Toast.error(t('标题长度不能超过200字符'));
      return;
    }

    if (!values.content || values.content.trim().length === 0) {
      Toast.error(t('内容不能为空'));
      return;
    }

    if (values.content.trim().length > 5000) {
      Toast.error(t('内容长度不能超过5000字符'));
      return;
    }

    setLoading(true);
    try {
      const res = await API.post('/api/ticket/create', {
        title: values.title.trim(),
        content: values.content.trim(),
        priority: values.priority,
        category: values.category,
      });

      const { success, message } = res.data;

      if (success) {
        Toast.success(t('工单创建成功'));
        navigate('/ticket');
      } else {
        Toast.error(message || t('创建失败'));
      }
    } catch (error) {
      Toast.error(t('创建失败') + ': ' + error.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className='ticket-container'>
      <Card
        title={t('创建工单')}
        headerExtraContent={
          <Button onClick={() => navigate('/ticket')}>{t('返回列表')}</Button>
        }
      >
        <Form
          onSubmit={handleSubmit}
          labelPosition='left'
          labelWidth={100}
          style={{ maxWidth: 800 }}
        >
          <Form.Input
            field='title'
            label={t('标题')}
            placeholder={t('请输入工单标题')}
            rules={[
              { required: true, message: t('标题不能为空') },
              { max: 200, message: t('标题长度不能超过200字符') },
            ]}
            maxLength={200}
            showClear
          />

          <Form.Select
            field='category'
            label={t('分类')}
            placeholder={t('请选择分类')}
            rules={[{ required: true, message: t('请选择分类') }]}
            initValue='other'
          >
            {categoryOptions.map((option) => (
              <Select.Option key={option.value} value={option.value}>
                {option.label}
              </Select.Option>
            ))}
          </Form.Select>

          <Form.Select
            field='priority'
            label={t('优先级')}
            placeholder={t('请选择优先级')}
            rules={[{ required: true, message: t('请选择优先级') }]}
            initValue='medium'
          >
            {priorityOptions.map((option) => (
              <Select.Option key={option.value} value={option.value}>
                {option.label}
              </Select.Option>
            ))}
          </Form.Select>

          <Form.TextArea
            field='content'
            label={t('内容')}
            placeholder={t('请详细描述您的问题或建议')}
            rules={[
              { required: true, message: t('内容不能为空') },
              { max: 5000, message: t('内容长度不能超过5000字符') },
            ]}
            maxLength={5000}
            rows={10}
            showClear
          />

          <div style={{ marginTop: 20, display: 'flex', gap: 10 }}>
            <Button
              theme='solid'
              type='primary'
              htmlType='submit'
              loading={loading}
            >
              {t('提交')}
            </Button>
            <Button onClick={() => navigate('/ticket')}>{t('取消')}</Button>
          </div>
        </Form>
      </Card>
    </div>
  );
};

export default TicketCreate;
