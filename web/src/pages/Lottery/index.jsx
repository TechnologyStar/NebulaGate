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

import React, { useState, useEffect } from 'react';
import { Button, Card, Typography, Toast, Modal, List, Space, Tag } from '@douyinfe/semi-ui';
import { API } from '../../helpers';
import { renderQuota } from '../../helpers/render';
import { useTranslation } from 'react-i18next';
import { IconGift, IconHistory, IconSpin } from '@douyinfe/semi-icons';

const { Title, Text } = Typography;

const Lottery = () => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [spinning, setSpinning] = useState(false);
  const [prizeModal, setPrizeModal] = useState(false);
  const [prize, setPrize] = useState(null);
  const [history, setHistory] = useState([]);
  const [historyVisible, setHistoryVisible] = useState(false);

  const loadHistory = async () => {
    try {
      const res = await API.get('/api/user/lottery/records');
      if (res.data.success) {
        setHistory(res.data.data.items || []);
      }
    } catch (error) {
      console.error(error);
    }
  };

  const handleDraw = async () => {
    setLoading(true);
    setSpinning(true);
    try {
      const res = await API.post('/api/user/lottery/draw');
      if (res.data.success) {
        setTimeout(() => {
          setPrize(res.data.data);
          setPrizeModal(true);
          setSpinning(false);
          loadHistory();
        }, 2000); // Animation delay
      } else {
        Toast.error(res.data.message || t('lottery.failed', '抽奖失败'));
        setSpinning(false);
      }
    } catch (error) {
      Toast.error(error.message || t('lottery.failed', '抽奖失败'));
      setSpinning(false);
    } finally {
      setLoading(false);
    }
  };

  const showHistory = () => {
    loadHistory();
    setHistoryVisible(true);
  };

  const getPrizeDisplay = (record) => {
    if (record.prize_type === 'quota') {
      return `${renderQuota(record.prize_value)} ${t('lottery.quota', '额度')}`;
    } else if (record.prize_type === 'plan') {
      return `${t('lottery.plan', '套餐')}: ${record.prize_name}`;
    }
    return record.prize_name;
  };

  return (
    <div className='nebula-console-container max-w-4xl mx-auto'>
      <Card>
        <div className='text-center py-8'>
          <div className='relative inline-block'>
            {spinning ? (
              <IconSpin
                spin
                size='extra-large'
                style={{ fontSize: 128, color: '#F7BA1E' }}
              />
            ) : (
              <IconGift size='extra-large' style={{ fontSize: 128, color: '#F7BA1E' }} />
            )}
          </div>
          <Title heading={2} style={{ marginTop: 16 }}>
            {t('lottery.title', '幸运抽奖')}
          </Title>
          <Text type='tertiary'>{t('lottery.subtitle', '试试手气，赢取丰厚奖励')}</Text>
        </div>

        <div className='text-center my-8'>
          <Button
            theme='solid'
            size='large'
            loading={loading}
            disabled={spinning}
            onClick={handleDraw}
            style={{
              background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
              padding: '16px 64px',
              fontSize: 18
            }}
          >
            {spinning ? t('lottery.drawing', '抽奖中...') : t('lottery.draw_now', '立即抽奖')}
          </Button>
        </div>

        <div className='my-8'>
          <Title heading={4}>{t('lottery.rules', '抽奖说明')}</Title>
          <List
            dataSource={[
              t('lottery.rule_1', '每日签到后可参与抽奖'),
              t('lottery.rule_2', '奖品包括额度奖励和套餐升级'),
              t('lottery.rule_3', '中奖概率根据配置动态调整'),
              t('lottery.rule_4', '所有奖品实时发放')
            ]}
            renderItem={(item) => (
              <List.Item>
                <Text>• {item}</Text>
              </List.Item>
            )}
          />
        </div>

        <div className='text-center'>
          <Button icon={<IconHistory />} onClick={showHistory}>
            {t('lottery.view_history', '查看抽奖记录')}
          </Button>
        </div>
      </Card>

      {/* Prize Modal */}
      <Modal
        title={<span style={{ fontSize: 24 }}>🎉 {t('lottery.congratulations', '恭喜中奖')} 🎉</span>}
        visible={prizeModal}
        onCancel={() => setPrizeModal(false)}
        footer={
          <Button type='primary' onClick={() => setPrizeModal(false)}>
            {t('lottery.confirm', '确认')}
          </Button>
        }
        centered
      >
        {prize && (
          <div className='text-center py-8'>
            <Title heading={3}>{prize.prize_name}</Title>
            <Text size='large' type='success'>
              {getPrizeDisplay(prize)}
            </Text>
          </div>
        )}
      </Modal>

      {/* History Modal */}
      <Modal
        title={t('lottery.history_title', '抽奖记录')}
        visible={historyVisible}
        onCancel={() => setHistoryVisible(false)}
        footer={null}
        width={700}
      >
        <List
          dataSource={history}
          renderItem={(record) => (
            <List.Item>
              <Space>
                <IconGift style={{ color: '#F7BA1E' }} />
                <div>
                  <div>
                    <Text strong>{record.prize_name}</Text>
                  </div>
                  <div>
                    <Text type='tertiary' size='small'>
                      {new Date(record.created_at).toLocaleString()}
                    </Text>
                  </div>
                </div>
                <Tag color='green'>{getPrizeDisplay(record)}</Tag>
              </Space>
            </List.Item>
          )}
          emptyContent={<Text type='tertiary'>{t('lottery.no_history', '暂无抽奖记录')}</Text>}
        />
      </Modal>
    </div>
  );
};

export default Lottery;
