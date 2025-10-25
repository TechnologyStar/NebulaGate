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

import { useCallback, useEffect, useState } from 'react';
import { API } from '../../helpers';

const DEFAULT_CONFIG = {
  billing: { enabled: false, defaultMode: 'balance' },
  governance: { enabled: false },
  public_logs: { enabled: false },
};

const normaliseConfig = (payload = {}) => {
  return {
    billing: {
      enabled: Boolean(payload?.billing?.enabled),
      defaultMode: payload?.billing?.defaultMode || 'balance',
    },
    governance: {
      enabled: Boolean(payload?.governance?.enabled),
    },
    public_logs: {
      enabled: Boolean(payload?.public_logs?.enabled),
      retention_days: payload?.public_logs?.retention_days,
    },
  };
};

export const useBillingFeatures = () => {
  const [config, setConfig] = useState(DEFAULT_CONFIG);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchConfig = useCallback(async () => {
    setLoading(true);
    try {
      const res = await API.get('/api/option/features', {
        params: { sections: 'billing,governance,public_logs' },
        skipErrorHandler: true,
      });
      if (res?.data?.success && res.data.data) {
        setConfig(normaliseConfig(res.data.data));
        setError(null);
      } else {
        throw new Error(res?.data?.message || 'Failed to load configuration');
      }
    } catch (err) {
      setConfig(DEFAULT_CONFIG);
      setError(err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchConfig();
  }, [fetchConfig]);

  return {
    config,
    loading,
    error,
    refresh: fetchConfig,
  };
};
