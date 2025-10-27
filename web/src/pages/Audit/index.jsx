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
import { Tabs, TabPane } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import { useNavigate, useLocation } from 'react-router-dom';
import { BarChart3, Image as ImageIcon, CheckSquare } from 'lucide-react';
import UsageLogsTable from '../../components/table/usage-logs';
import MjLogsTable from '../../components/table/mj-logs';
import TaskLogsTable from '../../components/table/task-logs';

const Audit = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const [tabActiveKey, setTabActiveKey] = useState('usage');

  const onChangeTab = (key) => {
    setTabActiveKey(key);
    navigate(`?tab=${key}`);
  };

  useEffect(() => {
    const searchParams = new URLSearchParams(window.location.search);
    const tab = searchParams.get('tab');
    if (tab) {
      setTabActiveKey(tab);
    } else {
      onChangeTab('usage');
    }
  }, [location.search]);

  const enableDrawing = localStorage.getItem('enable_drawing') === 'true';
  const enableTask = localStorage.getItem('enable_task') === 'true';

  return (
    <div className='mt-[60px] px-2'>
      <Tabs
        type='card'
        activeKey={tabActiveKey}
        onChange={(key) => onChangeTab(key)}
      >
        <TabPane
          itemKey='usage'
          tab={
            <span style={{ display: 'flex', alignItems: 'center', gap: '5px' }}>
              <BarChart3 size={18} />
              {t('使用日志')}
            </span>
          }
        >
          {tabActiveKey === 'usage' && <UsageLogsTable />}
        </TabPane>
        {enableDrawing && (
          <TabPane
            itemKey='drawing'
            tab={
              <span
                style={{ display: 'flex', alignItems: 'center', gap: '5px' }}
              >
                <ImageIcon size={18} />
                {t('绘图日志')}
              </span>
            }
          >
            {tabActiveKey === 'drawing' && <MjLogsTable />}
          </TabPane>
        )}
        {enableTask && (
          <TabPane
            itemKey='task'
            tab={
              <span
                style={{ display: 'flex', alignItems: 'center', gap: '5px' }}
              >
                <CheckSquare size={18} />
                {t('任务日志')}
              </span>
            }
          >
            {tabActiveKey === 'task' && <TaskLogsTable />}
          </TabPane>
        )}
      </Tabs>
    </div>
  );
};

export default Audit;
