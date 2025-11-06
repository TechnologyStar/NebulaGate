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

import React, { useEffect, useState, useCallback } from 'react';
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
  Divider,
  Switch,
  InputNumber,
  Popconfirm,
  Timeline,
  Badge,
  Progress,
  TextArea,
} from '@douyinfe/semi-ui';
import {
  IconShieldStroked,
  IconUserStroked,
  IconAlertCircle,
  IconDelete,
  IconSearch,
  IconRefresh,
  IconMonitor,
  IconGlobe,
  IconActivity,
  IconBan,
  IconCheckCircleStroked,
  IconExclamationTriangle,
  IconEyeOpened,
  IconFilter,
} from '@douyinfe/semi-icons';
import { VChart } from '@visactor/react-vchart';
import { API } from '../../helpers';

const { Title, Text } = Typography;

const SecurityCenter = () => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [activeTab, setActiveTab] = useState('dashboard');
  
  // Dashboard stats
  const [stats, setStats] = useState({});
  const [statsLoading, setStatsLoading] = useState(false);
  
  // Charts data
  const [anomalyTrendData, setAnomalyTrendData] = useState([]);
  const [deviceClusterData, setDeviceClusterData] = useState([]);
  const [ipAnalyticsData, setIpAnalyticsData] = useState([]);
  
  // Device clusters
  const [devices, setDevices] = useState([]);
  const [devicesTotal, setDevicesTotal] = useState(0);
  const [devicesPage, setDevicesPage] = useState(1);
  const [devicesPageSize] = useState(10);
  const [devicesFilters, setDevicesFilters] = useState({});
  
  // IP analytics
  const [ips, setIps] = useState([]);
  const [ipsTotal, setIpsTotal] = useState(0);
  const [ipsPage, setIpsPage] = useState(1);
  const [ipsPageSize] = useState(10);
  const [ipsFilters, setIpsFilters] = useState({});
  
  // Anomalies
  const [anomalies, setAnomalies] = useState([]);
  const [anomaliesTotal, setAnomaliesTotal] = useState(0);
  const [anomaliesPage, setAnomaliesPage] = useState(1);
  const [anomaliesPageSize] = useState(10);
  const [anomaliesFilters, setAnomaliesFilters] = useState({});
  
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
  const [deviceModalVisible, setDeviceModalVisible] = useState(false);
  const [selectedDevice, setSelectedDevice] = useState(null);
  const [ipModalVisible, setIpModalVisible] = useState(false);
  const [selectedIp, setSelectedIp] = useState(null);

  useEffect(() => {
    if (activeTab === 'dashboard') {
      loadDashboardStats();
    } else if (activeTab === 'devices') {
      loadDevices();
    } else if (activeTab === 'ips') {
      loadIps();
    } else if (activeTab === 'anomalies') {
      loadAnomalies();
    } else if (activeTab === 'violations') {
      loadViolations();
    } else if (activeTab === 'users') {
      loadUsers();
    } else if (activeTab === 'settings') {
      loadSettings();
    }
  }, [activeTab, violationsPage, usersPage, devicesPage, ipsPage, anomaliesPage]);

  const loadDashboardStats = async () => {
    setStatsLoading(true);
    try {
      const res = await API.get('/api/security/dashboard');
      if (res.data.success) {
        setStats(res.data.data);
        // Set chart data
        setAnomalyTrendData(res.data.data.anomaly_trend || []);
        setDeviceClusterData(res.data.data.device_clusters || []);
        setIpAnalyticsData(res.data.data.ip_analytics || []);
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

  const loadDevices = async () => {
    setLoading(true);
    try {
      const params = {
        page: devicesPage,
        page_size: devicesPageSize,
        ...devicesFilters,
      };
      const res = await API.get('/api/security/devices', { params });
      if (res.data.success) {
        setDevices(res.data.data.devices || []);
        setDevicesTotal(res.data.data.total || 0);
      }
    } catch (error) {
      Toast.error(t('security.loadDevicesFailed'));
    } finally {
      setLoading(false);
    }
  };

  const loadIps = async () => {
    setLoading(true);
    try {
      const params = {
        page: ipsPage,
        page_size: ipsPageSize,
        ...ipsFilters,
      };
      const res = await API.get('/api/security/ips', { params });
      if (res.data.success) {
        setIps(res.data.data.ips || []);
        setIpsTotal(res.data.data.total || 0);
      }
    } catch (error) {
      Toast.error(t('security.loadIpsFailed'));
    } finally {
      setLoading(false);
    }
  };

  const loadAnomalies = async () => {
    setLoading(true);
    try {
      const params = {
        page: anomaliesPage,
        page_size: anomaliesPageSize,
        ...anomaliesFilters,
      };
      const res = await API.get('/api/security/anomalies', { params });
      if (res.data.success) {
        setAnomalies(res.data.data.anomalies || []);
        setAnomaliesTotal(res.data.data.total || 0);
      }
    } catch (error) {
      Toast.error(t('security.loadAnomaliesFailed'));
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

  const handleAnomalyAction = async (anomalyId, action) => {
    try {
      const res = await API.post(`/api/security/anomalies/${anomalyId}/action`, { action });
      if (res.data.success) {
        Toast.success(t(`security.${action}Success`));
        loadAnomalies();
        loadDashboardStats(); // Refresh dashboard stats
      }
    } catch (error) {
      Toast.error(t(`security.${action}Failed`));
    }
  };

  const handleDeviceAction = async (deviceId, action) => {
    try {
      const res = await API.post(`/api/security/devices/${deviceId}/action`, { action });
      if (res.data.success) {
        Toast.success(t(`security.${action}Success`));
        loadDevices();
        loadDashboardStats();
      }
    } catch (error) {
      Toast.error(t(`security.${action}Failed`));
    }
  };

  const handleIpAction = async (ipId, action) => {
    try {
      const res = await API.post(`/api/security/ips/${ipId}/action`, { action });
      if (res.data.success) {
        Toast.success(t(`security.${action}Success`));
        loadIps();
        loadDashboardStats();
      }
    } catch (error) {
      Toast.error(t(`security.${action}Failed`));
    }
  };

  const handleViewDeviceDetails = (device) => {
    setSelectedDevice(device);
    setDeviceModalVisible(true);
  };

  const handleViewIpDetails = (ip) => {
    setSelectedIp(ip);
    setIpModalVisible(true);
  };

  // Chart specifications
  const anomalyTrendSpec = {
    type: 'line',
    data: {
      values: anomalyTrendData
    },
    xField: 'time',
    yField: 'count',
    seriesField: 'type',
    title: {
      visible: true,
      text: t('security.anomalyTrend')
    },
    legends: {
      visible: true,
      selectMode: 'single'
    },
    tooltip: {
      mark: {
        content: [
          {
            key: (datum) => datum['type'],
            value: (datum) => datum['count']
          }
        ]
      }
    },
    color: ['#ff6b6b', '#4ecdc4', '#45b7d1']
  };

  const deviceClusterSpec = {
    type: 'pie',
    data: {
      values: deviceClusterData
    },
    outerRadius: 0.8,
    innerRadius: 0.5,
    valueField: 'count',
    categoryField: 'device_type',
    title: {
      visible: true,
      text: t('security.deviceClusters')
    },
    legends: {
      visible: true,
      orient: 'left'
    },
    label: {
      visible: true
    },
    tooltip: {
      mark: {
        content: [
          {
            key: (datum) => datum['device_type'],
            value: (datum) => datum['count']
          }
        ]
      }
    },
    color: ['#ff6b6b', '#4ecdc4', '#45b7d1', '#96ceb4', '#ffeaa7']
  };

  const ipAnalyticsSpec = {
    type: 'bar',
    data: {
      values: ipAnalyticsData.slice(0, 20) // Top 20 IPs
    },
    xField: 'ip',
    yField: 'request_count',
    title: {
      visible: true,
      text: t('security.topIps')
    },
    tooltip: {
      mark: {
        content: [
          {
            key: (datum) => datum['ip'],
            value: (datum) => datum['request_count']
          }
        ]
      }
    },
    color: '#45b7d1'
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

  const deviceColumns = [
    {
      title: t('security.deviceId'),
      dataIndex: 'device_id',
    },
    {
      title: t('security.deviceType'),
      dataIndex: 'device_type',
      render: (text) => <Tag color="blue">{text}</Tag>,
    },
    {
      title: t('security.userAgent'),
      dataIndex: 'user_agent',
      render: (text) => (
        <Text ellipsis={{ showTooltip: true }} style={{ width: 200 }}>
          {text}
        </Text>
      ),
    },
    {
      title: t('security.associatedUsers'),
      dataIndex: 'user_count',
      sorter: (a, b) => a.user_count - b.user_count,
    },
    {
      title: t('security.requestCount'),
      dataIndex: 'request_count',
      sorter: (a, b) => a.request_count - b.request_count,
    },
    {
      title: t('security.anomalyScore'),
      dataIndex: 'anomaly_score',
      render: (score) => (
        <Progress 
          percent={score * 100} 
          size="small" 
          showInfo={false}
          stroke={score > 0.7 ? '#ff6b6b' : score > 0.4 ? '#ffa726' : '#66bb6a'}
        />
      ),
    },
    {
      title: t('security.status'),
      render: (_, record) => (
        <Space>
          {record.is_blocked && <Tag color="red">{t('security.blocked')}</Tag>}
          {record.is_nat && <Tag color="orange">{t('security.natDetected')}</Tag>}
          {!record.is_blocked && !record.is_nat && (
            <Tag color="green">{t('security.normal')}</Tag>
          )}
        </Space>
      ),
    },
    {
      title: t('security.operations'),
      render: (_, record) => (
        <Space>
          <Button
            size="small"
            icon={<IconEyeOpened />}
            onClick={() => handleViewDeviceDetails(record)}
          >
            {t('security.viewDetails')}
          </Button>
          {!record.is_blocked ? (
            <Popconfirm
              title={t('security.confirmBlockDevice')}
              onConfirm={() => handleDeviceAction(record.device_id, 'block')}
            >
              <Button type="danger" size="small" icon={<IconBan />}>
                {t('security.block')}
              </Button>
            </Popconfirm>
          ) : (
            <Button
              type="primary"
              size="small"
              onClick={() => handleDeviceAction(record.device_id, 'unblock')}
            >
              {t('security.unblock')}
            </Button>
          )}
        </Space>
      ),
    },
  ];

  const ipColumns = [
    {
      title: t('security.ipAddress'),
      dataIndex: 'ip_address',
    },
    {
      title: t('security.country'),
      dataIndex: 'country',
      render: (text) => text ? <Tag color="blue">{text}</Tag> : '-',
    },
    {
      title: t('security.associatedUsers'),
      dataIndex: 'user_count',
      sorter: (a, b) => a.user_count - b.user_count,
    },
    {
      title: t('security.requestCount'),
      dataIndex: 'request_count',
      sorter: (a, b) => a.request_count - b.request_count,
    },
    {
      title: t('security.anomalyScore'),
      dataIndex: 'anomaly_score',
      render: (score) => (
        <Progress 
          percent={score * 100} 
          size="small" 
          showInfo={false}
          stroke={score > 0.7 ? '#ff6b6b' : score > 0.4 ? '#ffa726' : '#66bb6a'}
        />
      ),
    },
    {
      title: t('security.status'),
      render: (_, record) => (
        <Space>
          {record.is_blocked && <Tag color="red">{t('security.blocked')}</Tag>}
          {record.is_nat && <Tag color="orange">{t('security.natDetected')}</Tag>}
          {record.is_proxy && <Tag color="purple">{t('security.proxyDetected')}</Tag>}
          {!record.is_blocked && !record.is_nat && !record.is_proxy && (
            <Tag color="green">{t('security.normal')}</Tag>
          )}
        </Space>
      ),
    },
    {
      title: t('security.operations'),
      render: (_, record) => (
        <Space>
          <Button
            size="small"
            icon={<IconEyeOpened />}
            onClick={() => handleViewIpDetails(record)}
          >
            {t('security.viewDetails')}
          </Button>
          {!record.is_blocked ? (
            <Popconfirm
              title={t('security.confirmBlockIp')}
              onConfirm={() => handleIpAction(record.ip_id, 'block')}
            >
              <Button type="danger" size="small" icon={<IconBan />}>
                {t('security.block')}
              </Button>
            </Popconfirm>
          ) : (
            <Button
              type="primary"
              size="small"
              onClick={() => handleIpAction(record.ip_id, 'unblock')}
            >
              {t('security.unblock')}
            </Button>
          )}
        </Space>
      ),
    },
  ];

  const anomalyColumns = [
    {
      title: t('security.anomalyTime'),
      dataIndex: 'detected_at',
      render: (text) => new Date(text).toLocaleString(),
    },
    {
      title: t('security.anomalyType'),
      dataIndex: 'anomaly_type',
      render: (text) => (
        <Tag color={text === 'malicious' ? 'red' : text === 'suspicious' ? 'orange' : 'blue'}>
          {text}
        </Tag>
      ),
    },
    {
      title: t('security.userId'),
      dataIndex: 'user_id',
    },
    {
      title: t('security.deviceOrIp'),
      dataIndex: 'target_identifier',
    },
    {
      title: t('security.severity'),
      dataIndex: 'severity',
      render: (text) => (
        <Tag color={text === 'high' ? 'red' : text === 'medium' ? 'orange' : 'yellow'}>
          {text}
        </Tag>
      ),
    },
    {
      title: t('security.description'),
      dataIndex: 'description',
      render: (text) => (
        <Text ellipsis={{ showTooltip: true }} style={{ width: 250 }}>
          {text}
        </Text>
      ),
    },
    {
      title: t('security.status'),
      dataIndex: 'status',
      render: (status) => (
        <Tag color={
          status === 'resolved' ? 'green' : 
          status === 'investigating' ? 'blue' : 
          status === 'ignored' ? 'grey' : 'red'
        }>
          {status}
        </Tag>
      ),
    },
    {
      title: t('security.operations'),
      render: (_, record) => (
        <Space>
          {record.status === 'pending' && (
            <>
              <Popconfirm
                title={t('security.confirmBan')}
                onConfirm={() => handleAnomalyAction(record.id, 'ban')}
              >
                <Button type="danger" size="small" icon={<IconBan />}>
                  {t('security.ban')}
                </Button>
              </Popconfirm>
              <Button
                size="small"
                onClick={() => handleAnomalyAction(record.id, 'redirect')}
              >
                {t('security.redirect')}
              </Button>
              <Button
                size="small"
                type="tertiary"
                onClick={() => handleAnomalyAction(record.id, 'ignore')}
              >
                {t('security.ignore')}
              </Button>
            </>
          )}
          {record.status === 'investigating' && (
            <Button
              size="small"
              type="primary"
              onClick={() => handleAnomalyAction(record.id, 'resolve')}
            >
              {t('security.resolve')}
            </Button>
          )}
        </Space>
      ),
    },
  ];

  const renderDashboard = () => (
    <Spin spinning={statsLoading}>
      {/* Enhanced Metrics Cards */}
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={4}>
          <Card
            title={
              <Space>
                <IconShieldStroked size="large" />
                <Text>{t('security.totalViolations')}</Text>
              </Space>
            }
          >
            <Title heading={2}>{stats.total_count || 0}</Title>
            <Text type="secondary" size="small">
              {t('security.last30Days')}
            </Text>
          </Card>
        </Col>
        <Col span={4}>
          <Card
            title={
              <Space>
                <IconUserStroked size="large" />
                <Text>{t('security.violatingUsers')}</Text>
              </Space>
            }
          >
            <Title heading={2}>{stats.unique_users || 0}</Title>
            <Text type="secondary" size="small">
              {t('security.uniqueUsers')}
            </Text>
          </Card>
        </Col>
        <Col span={4}>
          <Card
            title={
              <Space>
                <IconAlertCircle size="large" />
                <Text>{t('security.todayViolations')}</Text>
              </Space>
            }
          >
            <Title heading={2}>{stats.today_count || 0}</Title>
            <Text type="secondary" size="small">
              {t('security.last24Hours')}
            </Text>
          </Card>
        </Col>
        <Col span={4}>
          <Card
            title={
              <Space>
                <IconMonitor size="large" />
                <Text>{t('security.activeDevices')}</Text>
              </Space>
            }
          >
            <Title heading={2}>{stats.active_devices || 0}</Title>
            <Text type="secondary" size="small">
              {t('security.uniqueDevices')}
            </Text>
          </Card>
        </Col>
        <Col span={4}>
          <Card
            title={
              <Space>
                <IconGlobe size="large" />
                <Text>{t('security.uniqueIps')}</Text>
              </Space>
            }
          >
            <Title heading={2}>{stats.unique_ips || 0}</Title>
            <Text type="secondary" size="small">
              {t('security.uniqueIps')}
            </Text>
          </Card>
        </Col>
        <Col span={4}>
          <Card
            title={
              <Space>
                <IconActivity size="large" />
                <Text>{t('security.activeAnomalies')}</Text>
              </Space>
            }
          >
            <Title heading={2}>{stats.active_anomalies || 0}</Title>
            <Text type="secondary" size="small">
              {t('security.pendingInvestigation')}
            </Text>
          </Card>
        </Col>
      </Row>

      {/* Charts Section */}
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={12}>
          <Card title={t('security.anomalyTrend')} style={{ height: 400 }}>
            <VChart spec={anomalyTrendSpec} style={{ height: 320 }} />
          </Card>
        </Col>
        <Col span={12}>
          <Card title={t('security.deviceClusters')} style={{ height: 400 }}>
            <VChart spec={deviceClusterSpec} style={{ height: 320 }} />
          </Card>
        </Col>
      </Row>

      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={24}>
          <Card title={t('security.topIps')} style={{ height: 400 }}>
            <VChart spec={ipAnalyticsSpec} style={{ height: 320 }} />
          </Card>
        </Col>
      </Row>

      {/* Response Actions Summary */}
      <Row gutter={16}>
        <Col span={12}>
          <Card title={t('security.responseActionsSummary')}>
            <Row gutter={16}>
              <Col span={8}>
                <div style={{ textAlign: 'center' }}>
                  <Title heading={3} style={{ color: '#ff6b6b' }}>
                    {stats.actions_banned || 0}
                  </Title>
                  <Text>{t('security.usersBanned')}</Text>
                </div>
              </Col>
              <Col span={8}>
                <div style={{ textAlign: 'center' }}>
                  <Title heading={3} style={{ color: '#ffa726' }}>
                    {stats.actions_redirected || 0}
                  </Title>
                  <Text>{t('security.usersRedirected')}</Text>
                </div>
              </Col>
              <Col span={8}>
                <div style={{ textAlign: 'center' }}>
                  <Title heading={3} style={{ color: '#66bb6a' }}>
                    {stats.actions_resolved || 0}
                  </Title>
                  <Text>{t('security.anomaliesResolved')}</Text>
                </div>
              </Col>
            </Row>
          </Card>
        </Col>
        <Col span={12}>
          <Card title={t('security.topKeywords')}>
            {stats.top_keywords && stats.top_keywords.length > 0 ? (
              <Table
                columns={[
                  { title: t('security.keyword'), dataIndex: 'keyword' },
                  { title: t('security.count'), dataIndex: 'count' },
                ]}
                dataSource={stats.top_keywords}
                pagination={false}
                size="small"
              />
            ) : (
              <Text>{t('security.noData')}</Text>
            )}
          </Card>
        </Col>
      </Row>
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

  const renderDevices = () => (
    <div>
      <Card style={{ marginBottom: 16 }}>
        <Form layout="horizontal">
          <Space>
            <Form.Input
              field="device_type"
              label={t('security.deviceType')}
              placeholder={t('security.searchByDeviceType')}
              onChange={(value) =>
                setDevicesFilters({ ...devicesFilters, device_type: value })
              }
            />
            <Form.Select
              field="status"
              label={t('security.status')}
              placeholder={t('security.filterByStatus')}
              onChange={(value) =>
                setDevicesFilters({ ...devicesFilters, status: value })
              }
              style={{ width: 150 }}
            >
              <Select.Option value="">{t('common.all')}</Select.Option>
              <Select.Option value="normal">{t('security.normal')}</Select.Option>
              <Select.Option value="blocked">{t('security.blocked')}</Select.Option>
              <Select.Option value="nat">{t('security.natDetected')}</Select.Option>
            </Form.Select>
            <Button
              icon={<IconSearch />}
              type="primary"
              onClick={loadDevices}
            >
              {t('common.search')}
            </Button>
            <Button icon={<IconRefresh />} onClick={loadDevices}>
              {t('common.refresh')}
            </Button>
          </Space>
        </Form>
      </Card>

      <Card>
        <Table
          columns={deviceColumns}
          dataSource={devices}
          loading={loading}
          pagination={{
            currentPage: devicesPage,
            pageSize: devicesPageSize,
            total: devicesTotal,
            onChange: (page) => setDevicesPage(page),
          }}
        />
      </Card>
    </div>
  );

  const renderIps = () => (
    <div>
      <Card style={{ marginBottom: 16 }}>
        <Form layout="horizontal">
          <Space>
            <Form.Input
              field="ip_address"
              label={t('security.ipAddress')}
              placeholder={t('security.searchByIp')}
              onChange={(value) =>
                setIpsFilters({ ...ipsFilters, ip_address: value })
              }
            />
            <Form.Select
              field="status"
              label={t('security.status')}
              placeholder={t('security.filterByStatus')}
              onChange={(value) =>
                setIpsFilters({ ...ipsFilters, status: value })
              }
              style={{ width: 150 }}
            >
              <Select.Option value="">{t('common.all')}</Select.Option>
              <Select.Option value="normal">{t('security.normal')}</Select.Option>
              <Select.Option value="blocked">{t('security.blocked')}</Select.Option>
              <Select.Option value="nat">{t('security.natDetected')}</Select.Option>
              <Select.Option value="proxy">{t('security.proxyDetected')}</Select.Option>
            </Form.Select>
            <Button
              icon={<IconSearch />}
              type="primary"
              onClick={loadIps}
            >
              {t('common.search')}
            </Button>
            <Button icon={<IconRefresh />} onClick={loadIps}>
              {t('common.refresh')}
            </Button>
          </Space>
        </Form>
      </Card>

      <Card>
        <Table
          columns={ipColumns}
          dataSource={ips}
          loading={loading}
          pagination={{
            currentPage: ipsPage,
            pageSize: ipsPageSize,
            total: ipsTotal,
            onChange: (page) => setIpsPage(page),
          }}
        />
      </Card>
    </div>
  );

  const renderAnomalies = () => (
    <div>
      <Card style={{ marginBottom: 16 }}>
        <Form layout="horizontal">
          <Space>
            <Form.Select
              field="anomaly_type"
              label={t('security.anomalyType')}
              placeholder={t('security.filterByType')}
              onChange={(value) =>
                setAnomaliesFilters({ ...anomaliesFilters, anomaly_type: value })
              }
              style={{ width: 150 }}
            >
              <Select.Option value="">{t('common.all')}</Select.Option>
              <Select.Option value="malicious">{t('security.malicious')}</Select.Option>
              <Select.Option value="suspicious">{t('security.suspicious')}</Select.Option>
              <Select.Option value="unusual">{t('security.unusual')}</Select.Option>
            </Form.Select>
            <Form.Select
              field="severity"
              label={t('security.severity')}
              placeholder={t('security.filterBySeverity')}
              onChange={(value) =>
                setAnomaliesFilters({ ...anomaliesFilters, severity: value })
              }
              style={{ width: 150 }}
            >
              <Select.Option value="">{t('common.all')}</Select.Option>
              <Select.Option value="high">{t('security.high')}</Select.Option>
              <Select.Option value="medium">{t('security.medium')}</Select.Option>
              <Select.Option value="low">{t('security.low')}</Select.Option>
            </Form.Select>
            <Form.Select
              field="status"
              label={t('security.status')}
              placeholder={t('security.filterByStatus')}
              onChange={(value) =>
                setAnomaliesFilters({ ...anomaliesFilters, status: value })
              }
              style={{ width: 150 }}
            >
              <Select.Option value="">{t('common.all')}</Select.Option>
              <Select.Option value="pending">{t('security.pending')}</Select.Option>
              <Select.Option value="investigating">{t('security.investigating')}</Select.Option>
              <Select.Option value="resolved">{t('security.resolved')}</Select.Option>
              <Select.Option value="ignored">{t('security.ignored')}</Select.Option>
            </Form.Select>
            <Button
              icon={<IconFilter />}
              type="primary"
              onClick={loadAnomalies}
            >
              {t('common.filter')}
            </Button>
            <Button icon={<IconRefresh />} onClick={loadAnomalies}>
              {t('common.refresh')}
            </Button>
          </Space>
        </Form>
      </Card>

      <Card>
        <Table
          columns={anomalyColumns}
          dataSource={anomalies}
          loading={loading}
          pagination={{
            currentPage: anomaliesPage,
            pageSize: anomaliesPageSize,
            total: anomaliesTotal,
            onChange: (page) => setAnomaliesPage(page),
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
        <Divider>{t('security.enforcementSettings')}</Divider>
        
        <Form.Input
          field="violation_redirect_model"
          label={t('security.violationRedirectModel')}
          placeholder="gpt-3.5-turbo"
          rules={[
            { required: true, message: t('security.pleaseEnterModel') },
            { pattern: /^[a-zA-Z0-9\-_]+$/, message: t('security.invalidModelFormat') }
          ]}
        />
        
        <Form.Switch
          field="auto_ban_enabled"
          label={t('security.autoBanEnabled')}
        />
        
        <Form.InputNumber
          field="auto_ban_threshold"
          label={t('security.autoBanThreshold')}
          placeholder="10"
          min={1}
          max={1000}
          rules={[
            { required: true, message: t('security.pleaseEnterThreshold') },
            { type: 'number', min: 1, max: 1000, message: t('security.thresholdRange') }
          ]}
        />

        <Form.InputNumber
          field="auto_ban_duration_hours"
          label={t('security.autoBanDuration')}
          placeholder="24"
          min={1}
          max={8760}
          suffix={t('security.hours')}
          rules={[
            { type: 'number', min: 1, max: 8760, message: t('security.durationRange') }
          ]}
        />

        <Divider>{t('security.detectionSettings')}</Divider>

        <Form.Switch
          field="anomaly_detection_enabled"
          label={t('security.anomalyDetectionEnabled')}
        />

        <Form.InputNumber
          field="anomaly_threshold_score"
          label={t('security.anomalyThresholdScore')}
          placeholder="0.7"
          min={0.1}
          max={1.0}
          step={0.1}
          rules={[
            { type: 'number', min: 0.1, max: 1.0, message: t('security.scoreRange') }
          ]}
        />

        <Form.Switch
          field="device_fingerprinting_enabled"
          label={t('security.deviceFingerprintingEnabled')}
        />

        <Form.Switch
          field="ip_reputation_check_enabled"
          label={t('security.ipReputationCheckEnabled')}
        />

        <Form.InputNumber
          field="max_requests_per_minute"
          label={t('security.maxRequestsPerMinute')}
          placeholder="60"
          min={1}
          max={10000}
          rules={[
            { type: 'number', min: 1, max: 10000, message: t('security.requestsRange') }
          ]}
        />

        <Form.InputNumber
          field="max_requests_per_hour"
          label={t('security.maxRequestsPerHour')}
          placeholder="1000"
          min={1}
          max={100000}
          rules={[
            { type: 'number', min: 1, max: 100000, message: t('security.requestsRangeHour') }
          ]}
        />

        <Divider>{t('security.notificationSettings')}</Divider>

        <Form.Switch
          field="real_time_alerts_enabled"
          label={t('security.realTimeAlertsEnabled')}
        />

        <Form.Select
          field="alert_severity_threshold"
          label={t('security.alertSeverityThreshold')}
          style={{ width: 200 }}
        >
          <Select.Option value="low">{t('security.low')}</Select.Option>
          <Select.Option value="medium">{t('security.medium')}</Select.Option>
          <Select.Option value="high">{t('security.high')}</Select.Option>
        </Form.Select>

        <Form.TextArea
          field="notification_webhook_url"
          label={t('security.notificationWebhookUrl')}
          placeholder="https://hooks.slack.com/..."
          rules={[
            { pattern: /^https?:\/\/.+/, message: t('security.invalidWebhookUrl') }
          ]}
        />

        <div style={{ marginTop: 24 }}>
          <Button htmlType="submit" type="primary">
            {t('common.save')}
          </Button>
          <Button 
            type="tertiary" 
            style={{ marginLeft: 8 }}
            onClick={() => settingsForm.reset()}
          >
            {t('common.reset')}
          </Button>
        </div>
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
        <TabPane tab={t('security.deviceClusters')} itemKey="devices">
          {renderDevices()}
        </TabPane>
        <TabPane tab={t('security.ipAnalytics')} itemKey="ips">
          {renderIps()}
        </TabPane>
        <TabPane tab={t('security.anomalyManagement')} itemKey="anomalies">
          {renderAnomalies()}
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

      {/* Device Details Modal */}
      <Modal
        title={t('security.deviceDetails')}
        visible={deviceModalVisible}
        onCancel={() => setDeviceModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDeviceModalVisible(false)}>
            {t('common.close')}
          </Button>
        ]}
        width={800}
      >
        {selectedDevice && (
          <div>
            <Row gutter={16} style={{ marginBottom: 16 }}>
              <Col span={12}>
                <Text strong>{t('security.deviceId')}:</Text> {selectedDevice.device_id}
              </Col>
              <Col span={12}>
                <Text strong>{t('security.deviceType')}:</Text> 
                <Tag color="blue">{selectedDevice.device_type}</Tag>
              </Col>
            </Row>
            <Row gutter={16} style={{ marginBottom: 16 }}>
              <Col span={12}>
                <Text strong>{t('security.userAgent')}:</Text>
                <Text ellipsis={{ showTooltip: true }} style={{ width: 250 }}>
                  {selectedDevice.user_agent}
                </Text>
              </Col>
              <Col span={12}>
                <Text strong>{t('security.anomalyScore')}:</Text>
                <Progress 
                  percent={selectedDevice.anomaly_score * 100} 
                  size="small"
                  stroke={selectedDevice.anomaly_score > 0.7 ? '#ff6b6b' : selectedDevice.anomaly_score > 0.4 ? '#ffa726' : '#66bb6a'}
                />
              </Col>
            </Row>
            <Divider>{t('security.associatedUsers')}</Divider>
            <Table
              columns={[
                { title: t('security.userId'), dataIndex: 'user_id' },
                { title: t('security.username'), dataIndex: 'username' },
                { title: t('security.lastSeen'), dataIndex: 'last_seen', render: (text) => new Date(text).toLocaleString() },
                { title: t('security.requestCount'), dataIndex: 'request_count' },
              ]}
              dataSource={selectedDevice.associated_users || []}
              pagination={false}
              size="small"
            />
          </div>
        )}
      </Modal>

      {/* IP Details Modal */}
      <Modal
        title={t('security.ipDetails')}
        visible={ipModalVisible}
        onCancel={() => setIpModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setIpModalVisible(false)}>
            {t('common.close')}
          </Button>
        ]}
        width={800}
      >
        {selectedIp && (
          <div>
            <Row gutter={16} style={{ marginBottom: 16 }}>
              <Col span={12}>
                <Text strong>{t('security.ipAddress')}:</Text> {selectedIp.ip_address}
              </Col>
              <Col span={12}>
                <Text strong>{t('security.country')}:</Text> 
                <Tag color="blue">{selectedIp.country || 'Unknown'}</Tag>
              </Col>
            </Row>
            <Row gutter={16} style={{ marginBottom: 16 }}>
              <Col span={12}>
                <Text strong>{t('security.anomalyScore')}:</Text>
                <Progress 
                  percent={selectedIp.anomaly_score * 100} 
                  size="small"
                  stroke={selectedIp.anomaly_score > 0.7 ? '#ff6b6b' : selectedIp.anomaly_score > 0.4 ? '#ffa726' : '#66bb6a'}
                />
              </Col>
              <Col span={12}>
                <Text strong>{t('security.requestCount')}:</Text> {selectedIp.request_count}
              </Col>
            </Row>
            <Divider>{t('security.associatedUsers')}</Divider>
            <Table
              columns={[
                { title: t('security.userId'), dataIndex: 'user_id' },
                { title: t('security.username'), dataIndex: 'username' },
                { title: t('security.lastSeen'), dataIndex: 'last_seen', render: (text) => new Date(text).toLocaleString() },
                { title: t('security.requestCount'), dataIndex: 'request_count' },
              ]}
              dataSource={selectedIp.associated_users || []}
              pagination={false}
              size="small"
            />
          </div>
        )}
      </Modal>

      {/* Redirect Modal */}
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
