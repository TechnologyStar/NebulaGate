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
import { Button, Card, Typography, Toast, Space, Modal, List } from '@douyinfe/semi-ui';
import { API } from '../../helpers';
import { renderQuota } from '../../helpers/render';
import { useTranslation } from 'react-i18next';
import { IconCheckCircleStroked, IconGift, IconHistoryStroked } from '@douyinfe/semi-icons';

const { Title, Text } = Typography;

const CheckIn = () => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [status, setStatus] = useState({
    has_checked_in: false,
    consecutive_days: 0,
    today_record: null
  });
  const [history, setHistory] = useState([]);
  const [historyVisible, setHistoryVisible] = useState(false);

  const loadStatus = async () => {
    try {
      const res = await API.get('/api/user/checkin/status');
      if (res.data.success) {
        setStatus(res.data.data);
      }
    } catch (error) {
      console.error(error);
    }
  };

  const loadHistory = async () => {
    try {
      const res = await API.get('/api/user/checkin/history');
      if (res.data.success) {
        setHistory(res.data.data.items || []);
      }
    } catch (error) {
      console.error(error);
    }
  };

  useEffect(() => {
    loadStatus();
  }, []);

  const handleCheckIn = async () => {
    if (status.has_checked_in) {
      Toast.warning(t('checkin.already_checked_in'));
      return;
    }

    setLoading(true);
    try {
      const res = await API.post('/api/user/checkin');
      if (res.data.success) {
        Toast.success(
          `${t('checkin.success')}! ${t('checkin.awarded')}: ${renderQuota(res.data.data.quota_awarded)}`
        );
        await loadStatus();
      } else {
        Toast.error(res.data.message || t('checkin.failed'));
      }
    } catch (error) {
      Toast.error(error.message || t('checkin.failed'));
    } finally {
      setLoading(false);
    }
  };

  const showHistory = () => {
    loadHistory();
    setHistoryVisible(true);
  };

  const getRewardText = (days) => {
    if (days >= 30) return '500,000';
    if (days >= 14) return '300,000';
    if (days >= 7) return '200,000';
    return '100,000';
  };

  return (
    <div className='mt-[60px] px-2 max-w-4xl mx-auto'>
      <Card>
        <div className='text-center py-8'>
          <IconGift size='extra-large' style={{ fontSize: 64, color: '#F7BA1E' }} />
          <Title heading={2} style={{ marginTop: 16 }}>
            {t('checkin.title', '每日签到')}
          </Title>
          <Text type='tertiary'>{t('checkin.subtitle', '每天签到领取额度奖励')}</Text>
        </div>

        <div className='grid grid-cols-1 md:grid-cols-3 gap-4 my-8'>
          <Card shadow>
            <div className='text-center'>
              <Text type='secondary'>{t('checkin.consecutive_days', '连续签到')}</Text>
              <Title heading={3} style={{ margin: '8px 0', color: '#F7BA1E' }}>
                {status.consecutive_days}
              </Title>
              <Text type='tertiary'>{t('checkin.days', '天')}</Text>
            </div>
          </Card>

          <Card shadow>
            <div className='text-center'>
              <Text type='secondary'>{t('checkin.today_status', '今日状态')}</Text>
              <Title heading={3} style={{ margin: '8px 0' }}>
                {status.has_checked_in ? (
                  <span style={{ color: '#52C41A' }}>
                    <IconCheckCircleStroked /> {t('checkin.checked', '已签到')}
                  </span>
                ) : (
                  <span style={{ color: '#FF7D00' }}>{t('checkin.not_checked', '未签到')}</span>
                )}
              </Title>
            </div>
          </Card>

          <Card shadow>
            <div className='text-center'>
              <Text type='secondary'>{t('checkin.today_reward', '今日奖励')}</Text>
              <Title heading={3} style={{ margin: '8px 0', color: '#3F8CFF' }}>
                {getRewardText(status.consecutive_days + 1)}
              </Title>
              <Text type='tertiary'>{t('checkin.quota', '额度')}</Text>
            </div>
          </Card>
        </div>

        <div className='text-center my-8'>
          <Button
            theme='solid'
            size='large'
            loading={loading}
            disabled={status.has_checked_in}
            onClick={handleCheckIn}
            style={{
              background: status.has_checked_in ? undefined : 'linear-gradient(90deg, #F7BA1E 0%, #FF7D00 100%)',
              padding: '12px 48px',
              fontSize: 16
            }}
          >
            {status.has_checked_in ? t('checkin.checked', '已签到') : t('checkin.check_in_now', '立即签到')}
          </Button>
        </div>

        <div className='my-8'>
          <Title heading={4}>{t('checkin.reward_rules', '奖励规则')}</Title>
          <List
            dataSource={[
              { days: '1-6', reward: '100,000' },
              { days: '7-13', reward: '200,000' },
              { days: '14-29', reward: '300,000' },
              { days: '30+', reward: '500,000' }
            ]}
            renderItem={(item) => (
              <List.Item>
                <Space>
                  <Text strong>{t('checkin.consecutive', '连续')} {item.days} {t('checkin.days', '天')}</Text>
                  <Text>→</Text>
                  <Text type='success'>{item.reward} {t('checkin.quota', '额度')}</Text>
                </Space>
              </List.Item>
            )}
          />
        </div>

        <div className='text-center'>
          <Button icon={<IconHistoryStroked />} onClick={showHistory}>
            {t('checkin.view_history', '查看签到历史')}
          </Button>
        </div>
      </Card>

      <Modal
        title={t('checkin.history_title', '签到历史')}
        visible={historyVisible}
        onCancel={() => setHistoryVisible(false)}
        footer={null}
        width={600}
      >
        <List
          dataSource={history}
          renderItem={(record) => (
            <List.Item>
              <Space>
                <IconCheckCircleStroked style={{ color: '#52C41A' }} />
                <Text>{record.check_in_date}</Text>
                <Text type='secondary'>
                  {t('checkin.consecutive', '连续')} {record.consecutive_days} {t('checkin.days', '天')}
                </Text>
                <Text type='success'>+{renderQuota(record.quota_awarded)}</Text>
              </Space>
            </List.Item>
          )}
          emptyContent={<Text type='tertiary'>{t('checkin.no_history', '暂无签到记录')}</Text>}
        />
      </Modal>
    </div>
  );
};

export default CheckIn;
