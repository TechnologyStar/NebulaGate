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

import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Card,
  Tabs,
  TabPane,
  Table,
  Button,
  Form,
  Input,
  DatePicker,
  Select,
  Modal,
  Toast,
  Tag,
  Space,
  Typography,
  Row,
  Col,
  Spin,
} from '@douyinfe/semi-ui';
import {
  IconShieldStroked,
  IconUserStroked,
  IconAlertCircle,
  IconDelete,
  IconSearch,
  IconRefresh,
} from '@douyinfe/semi-icons';
import { API } from '../../helpers';

const { Title, Text } = Typography;

const SecurityCenter = () => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [activeTab, setActiveTab] = useState('dashboard');
  
  // Dashboard stats
  const [stats, setStats] = useState({});
  const [statsLoading, setStatsLoading] = useState(false);
  
  // Violations
  const [violations, setViolations] = useState([]);
  const [violationsTotal, setViolationsTotal] = useState(0);
  const [violationsPage, setViolationsPage] = useState(1);
  const [violationsPageSize] = useState(10);
  const [violationsFilters, setViolationsFilters] = useState({});
  
  // Users
  const [users, setUsers] = useState([]);
  const [usersTotal, setUsersTotal] = useState(0);
  const [usersPage, setUsersPage] = useState(1);
  const [usersPageSize] = useState(10);
  
  // Settings
  const [settings, setSettings] = useState({});
  const [settingsForm] = Form.useForm();
  
  // Modals
  const [redirectModalVisible, setRedirectModalVisible] = useState(false);
  const [selectedUser, setSelectedUser] = useState(null);
  const [redirectModel, setRedirectModel] = useState('');

  useEffect(() => {
    if (activeTab === 'dashboard') {
      loadDashboardStats();
    } else if (activeTab === 'violations') {
      loadViolations();
    } else if (activeTab === 'users') {
      loadUsers();
    } else if (activeTab === 'settings') {
      loadSettings();
    }
  }, [activeTab, violationsPage, usersPage]);

  const loadDashboardStats = async () => {
    setStatsLoading(true);
    try {
      const res = await API.get('/api/security/dashboard');
      if (res.data.success) {
        setStats(res.data.data);
      }
    } catch (error) {
      Toast.error(t('security.loadStatsFailed'));
    } finally {
      setStatsLoading(false);
    }
  };

  const loadViolations = async () => {
    setLoading(true);
    try {
      const params = {
        page: violationsPage,
        page_size: violationsPageSize,
        ...violationsFilters,
      };
      const res = await API.get('/api/security/violations', { params });
      if (res.data.success) {
        setViolations(res.data.data.violations || []);
        setViolationsTotal(res.data.data.total || 0);
      }
    } catch (error) {
      Toast.error(t('security.loadViolationsFailed'));
    } finally {
      setLoading(false);
    }
  };

  const loadUsers = async () => {
    setLoading(true);
    try {
      const params = {
        page: usersPage,
        page_size: usersPageSize,
      };
      const res = await API.get('/api/security/users', { params });
      if (res.data.success) {
        setUsers(res.data.data.users || []);
        setUsersTotal(res.data.data.total || 0);
      }
    } catch (error) {
      Toast.error(t('security.loadUsersFailed'));
    } finally {
      setLoading(false);
    }
  };

  const loadSettings = async () => {
    setLoading(true);
    try {
      const res = await API.get('/api/security/settings');
      if (res.data.success) {
        setSettings(res.data.data);
        settingsForm.setValues(res.data.data);
      }
    } catch (error) {
      Toast.error(t('security.loadSettingsFailed'));
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteViolation = async (id) => {
    Modal.confirm({
      title: t('security.confirmDelete'),
      content: t('security.confirmDeleteViolation'),
      onOk: async () => {
        try {
          const res = await API.delete(`/api/security/violations/${id}`);
          if (res.data.success) {
            Toast.success(t('security.deleteSuccess'));
            loadViolations();
          }
        } catch (error) {
          Toast.error(t('security.deleteFailed'));
        }
      },
    });
  };

  const handleBanUser = async (userId) => {
    Modal.confirm({
      title: t('security.confirmBan'),
      content: t('security.confirmBanUser'),
      onOk: async () => {
        try {
          const res = await API.post(`/api/security/users/${userId}/ban`);
          if (res.data.success) {
            Toast.success(t('security.banSuccess'));
            loadUsers();
          }
        } catch (error) {
          Toast.error(t('security.banFailed'));
        }
      },
    });
  };

  const handleUnbanUser = async (userId) => {
    try {
      const res = await API.post(`/api/security/users/${userId}/unban`);
      if (res.data.success) {
        Toast.success(t('security.unbanSuccess'));
        loadUsers();
      }
    } catch (error) {
      Toast.error(t('security.unbanFailed'));
    }
  };

  const handleSetRedirect = (user) => {
    setSelectedUser(user);
    setRedirectModel(user.redirect_model || '');
    setRedirectModalVisible(true);
  };

  const handleConfirmRedirect = async () => {
    if (!redirectModel) {
      Toast.error(t('security.pleaseSelectModel'));
      return;
    }
    
    try {
      const res = await API.post(`/api/security/users/${selectedUser.user_id}/redirect`, {
        model: redirectModel,
      });
      if (res.data.success) {
        Toast.success(t('security.redirectSetSuccess'));
        setRedirectModalVisible(false);
        loadUsers();
      }
    } catch (error) {
      Toast.error(t('security.redirectSetFailed'));
    }
  };

  const handleClearRedirect = async (userId) => {
    try {
      const res = await API.delete(`/api/security/users/${userId}/redirect`);
      if (res.data.success) {
        Toast.success(t('security.redirectClearSuccess'));
        loadUsers();
      }
    } catch (error) {
      Toast.error(t('security.redirectClearFailed'));
    }
  };

  const handleSaveSettings = async (values) => {
    try {
      const res = await API.put('/api/security/settings', values);
      if (res.data.success) {
        Toast.success(t('security.settingsSaveSuccess'));
      }
    } catch (error) {
      Toast.error(t('security.settingsSaveFailed'));
    }
  };

  const violationColumns = [
    {
      title: t('security.violationTime'),
      dataIndex: 'violated_at',
      render: (text) => new Date(text).toLocaleString(),
    },
    {
      title: t('security.userId'),
      dataIndex: 'user_id',
    },
    {
      title: t('security.contentSnippet'),
      dataIndex: 'content_snippet',
      render: (text) => (
        <Text ellipsis={{ showTooltip: true }} style={{ width: 300 }}>
          {text}
        </Text>
      ),
    },
    {
      title: t('security.matchedKeywords'),
      dataIndex: 'matched_keywords',
      render: (text) => text && <Tag color="red">{text}</Tag>,
    },
    {
      title: t('security.model'),
      dataIndex: 'model',
    },
    {
      title: t('security.severity'),
      dataIndex: 'severity',
      render: (text) => (
        <Tag color={text === 'malicious' ? 'red' : 'orange'}>{text}</Tag>
      ),
    },
    {
      title: t('security.action'),
      dataIndex: 'action_taken',
    },
    {
      title: t('security.operations'),
      render: (_, record) => (
        <Button
          type="danger"
          size="small"
          icon={<IconDelete />}
          onClick={() => handleDeleteViolation(record.id)}
        >
          {t('common.delete')}
        </Button>
      ),
    },
  ];

  const userColumns = [
    {
      title: t('security.userId'),
      dataIndex: 'user_id',
    },
    {
      title: t('security.username'),
      dataIndex: 'username',
    },
    {
      title: t('security.displayName'),
      dataIndex: 'display_name',
    },
    {
      title: t('security.violationCount'),
      dataIndex: 'violation_count',
      sorter: (a, b) => a.violation_count - b.violation_count,
    },
    {
      title: t('security.lastViolation'),
      dataIndex: 'last_violation_at',
      render: (text) => text ? new Date(text).toLocaleString() : '-',
    },
    {
      title: t('security.status'),
      render: (_, record) => (
        <Space>
          {record.is_banned && <Tag color="red">{t('security.banned')}</Tag>}
          {record.redirect_model && (
            <Tag color="orange">{t('security.redirected')}</Tag>
          )}
          {!record.is_banned && !record.redirect_model && (
            <Tag color="green">{t('security.normal')}</Tag>
          )}
        </Space>
      ),
    },
    {
      title: t('security.operations'),
      render: (_, record) => (
        <Space>
          {!record.is_banned ? (
            <Button
              type="danger"
              size="small"
              onClick={() => handleBanUser(record.user_id)}
            >
              {t('security.ban')}
            </Button>
          ) : (
            <Button
              type="primary"
              size="small"
              onClick={() => handleUnbanUser(record.user_id)}
            >
              {t('security.unban')}
            </Button>
          )}
          {!record.redirect_model ? (
            <Button size="small" onClick={() => handleSetRedirect(record)}>
              {t('security.setRedirect')}
            </Button>
          ) : (
            <Button
              size="small"
              type="tertiary"
              onClick={() => handleClearRedirect(record.user_id)}
            >
              {t('security.clearRedirect')}
            </Button>
          )}
        </Space>
      ),
    },
  ];

  const renderDashboard = () => (
    <Spin spinning={statsLoading}>
      <Row gutter={16}>
        <Col span={6}>
          <Card
            title={
              <Space>
                <IconShieldStroked size="large" />
                <Text>{t('security.totalViolations')}</Text>
              </Space>
            }
          >
            <Title heading={2}>{stats.total_count || 0}</Title>
          </Card>
        </Col>
        <Col span={6}>
          <Card
            title={
              <Space>
                <IconUserStroked size="large" />
                <Text>{t('security.violatingUsers')}</Text>
              </Space>
            }
          >
            <Title heading={2}>{stats.unique_users || 0}</Title>
          </Card>
        </Col>
        <Col span={6}>
          <Card
            title={
              <Space>
                <IconAlertCircle size="large" />
                <Text>{t('security.todayViolations')}</Text>
              </Space>
            }
          >
            <Title heading={2}>{stats.today_count || 0}</Title>
          </Card>
        </Col>
      </Row>

      <Card style={{ marginTop: 16 }} title={t('security.topKeywords')}>
        {stats.top_keywords && stats.top_keywords.length > 0 ? (
          <Table
            columns={[
              { title: t('security.keyword'), dataIndex: 'keyword' },
              { title: t('security.count'), dataIndex: 'count' },
            ]}
            dataSource={stats.top_keywords}
            pagination={false}
          />
        ) : (
          <Text>{t('security.noData')}</Text>
        )}
      </Card>
    </Spin>
  );

  const renderViolations = () => (
    <div>
      <Card style={{ marginBottom: 16 }}>
        <Form layout="horizontal">
          <Space>
            <Form.Input
              field="user_id"
              label={t('security.userId')}
              placeholder={t('security.searchByUserId')}
              onChange={(value) =>
                setViolationsFilters({ ...violationsFilters, user_id: value })
              }
            />
            <Form.Input
              field="keyword"
              label={t('security.keyword')}
              placeholder={t('security.searchByKeyword')}
              onChange={(value) =>
                setViolationsFilters({ ...violationsFilters, keyword: value })
              }
            />
            <Button
              icon={<IconSearch />}
              type="primary"
              onClick={loadViolations}
            >
              {t('common.search')}
            </Button>
            <Button icon={<IconRefresh />} onClick={loadViolations}>
              {t('common.refresh')}
            </Button>
          </Space>
        </Form>
      </Card>

      <Card>
        <Table
          columns={violationColumns}
          dataSource={violations}
          loading={loading}
          pagination={{
            currentPage: violationsPage,
            pageSize: violationsPageSize,
            total: violationsTotal,
            onChange: (page) => setViolationsPage(page),
          }}
        />
      </Card>
    </div>
  );

  const renderUsers = () => (
    <div>
      <Button
        icon={<IconRefresh />}
        onClick={loadUsers}
        style={{ marginBottom: 16 }}
      >
        {t('common.refresh')}
      </Button>

      <Card>
        <Table
          columns={userColumns}
          dataSource={users}
          loading={loading}
          pagination={{
            currentPage: usersPage,
            pageSize: usersPageSize,
            total: usersTotal,
            onChange: (page) => setUsersPage(page),
          }}
        />
      </Card>
    </div>
  );

  const renderSettings = () => (
    <Card>
      <Form
        form={settingsForm}
        onSubmit={handleSaveSettings}
        labelPosition="left"
        labelWidth={200}
      >
        <Form.Input
          field="violation_redirect_model"
          label={t('security.violationRedirectModel')}
          placeholder="gpt-3.5-turbo"
        />
        <Form.Switch
          field="auto_ban_enabled"
          label={t('security.autoBanEnabled')}
        />
        <Form.InputNumber
          field="auto_ban_threshold"
          label={t('security.autoBanThreshold')}
          placeholder="10"
        />
        <Button htmlType="submit" type="primary">
          {t('common.save')}
        </Button>
      </Form>
    </Card>
  );

  return (
    <div className="nebula-console-container">
      <Title heading={3}>
        <Space>
          <IconShieldStroked />
          {t('security.title')}
        </Space>
      </Title>

      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        type="line"
        style={{ marginTop: 16 }}
      >
        <TabPane tab={t('security.dashboard')} itemKey="dashboard">
          {renderDashboard()}
        </TabPane>
        <TabPane tab={t('security.violations')} itemKey="violations">
          {renderViolations()}
        </TabPane>
        <TabPane tab={t('security.users')} itemKey="users">
          {renderUsers()}
        </TabPane>
        <TabPane tab={t('security.settings')} itemKey="settings">
          {renderSettings()}
        </TabPane>
      </Tabs>

      <Modal
        title={t('security.setRedirect')}
        visible={redirectModalVisible}
        onOk={handleConfirmRedirect}
        onCancel={() => setRedirectModalVisible(false)}
      >
        <Form>
          <Form.Input
            label={t('security.targetModel')}
            value={redirectModel}
            onChange={setRedirectModel}
            placeholder="gpt-3.5-turbo"
          />
        </Form>
      </Modal>
    </div>
  );
};

export default SecurityCenter;
