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

import React from 'react';
import { Progress, Skeleton, Tooltip, Typography } from '@douyinfe/semi-ui';

const formatPercent = (value, total) => {
  if (!total || total <= 0) {
    return 0;
  }
  const pct = Math.min(Math.max((value / total) * 100, 0), 100);
  if (Number.isNaN(pct)) {
    return 0;
  }
  return pct;
};

const getStatusColor = (percent) => {
  if (percent >= 75) {
    return 'var(--semi-color-success)';
  }
  if (percent >= 40) {
    return 'var(--semi-color-warning)';
  }
  return 'var(--semi-color-danger)';
};

const QuotaProgress = ({
  title,
  used = 0,
  total = 0,
  description,
  loading = false,
  suffix,
  precision = 0,
  tooltip,
}) => {
  if (loading) {
    return <Skeleton placeholder className='w-full h-[80px]' />;
  }

  const percent = formatPercent(total - used, total);
  const remaining = Math.max(total - used, 0);

  const content = (
    <div className='rounded-lg border border-[var(--semi-color-border)] p-4 bg-[var(--semi-color-fill-0)]'>
      <div className='flex items-center justify-between mb-2'>
        <Typography.Text strong>{title}</Typography.Text>
        <Typography.Text type='tertiary'>
          {suffix ? `${remaining.toFixed(precision)} ${suffix}` : remaining.toFixed(precision)}
        </Typography.Text>
      </div>
      <Progress
        percent={percent}
        showInfo
        format={(p) => `${p.toFixed(0)}%`}
        stroke={getStatusColor(percent)}
      />
      {description ? (
        <Typography.Text type='tertiary' size='small'>
          {description}
        </Typography.Text>
      ) : null}
    </div>
  );

  if (tooltip) {
    return <Tooltip content={tooltip}>{content}</Tooltip>;
  }

  return content;
};

export default QuotaProgress;
