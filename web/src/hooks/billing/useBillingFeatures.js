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

import { useCallback, useEffect, useMemo, useState } from 'react';
import { API } from '../../helpers';

const DEFAULT_CONFIG = {
  billing: { enabled: false, defaultMode: 'balance' },
  governance: { enabled: false },
  public_logs: { enabled: false, retention_days: 7 },
};

const FEATURE_SECTIONS = Object.freeze(['billing', 'governance', 'public_logs']);

const SECTION_ALIASES = {
  billing: 'billing',
  finance: 'billing',
  fea_mance: 'billing',
  governance: 'governance',
  public_logs: 'public_logs',
  publiclog: 'public_logs',
  publiclogs: 'public_logs',
  'public-log': 'public_logs',
};

const normaliseSectionKey = (value) => {
  if (!value && value !== 0) {
    return null;
  }
  const safeValue = String(value)
    .trim()
    .toLowerCase()
    .replace(/[\s-]+/g, '_');
  if (!safeValue) {
    return null;
  }
  return SECTION_ALIASES[safeValue] || safeValue;
};

const serialiseSections = (sections) => {
  const raw = Array.isArray(sections)
    ? sections
    : String(sections ?? '')
        .split(',')
        .map((item) => item.trim());

  const deduped = Array.from(
    new Set(raw.map(normaliseSectionKey).filter(Boolean)),
  );

  const params = new URLSearchParams();
  deduped.forEach((section) => params.append('sections', section));
  return params.toString();
};

const normaliseConfig = (payload = {}) => {
  try {
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
        retention_days: Number(payload?.public_logs?.retention_days) || 7,
      },
    };
  } catch (error) {
    console.error('Failed to normalize config:', error);
    return DEFAULT_CONFIG;
  }
};

export const useBillingFeatures = () => {
  const [config, setConfig] = useState(DEFAULT_CONFIG);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const requestSections = useMemo(() => FEATURE_SECTIONS, []);

  const fetchConfig = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await API.get('/api/option/features', {
        params: { sections: requestSections },
        paramsSerializer: (params) => serialiseSections(params.sections),
        skipErrorHandler: true,
      });
      if (res?.data?.success && res.data.data) {
        setConfig(normaliseConfig(res.data.data));
        setError(null);
      } else {
        throw new Error(res?.data?.message || 'Failed to load configuration');
      }
    } catch (err) {
      console.error('Failed to fetch billing features config:', err);
      setConfig(DEFAULT_CONFIG);
      let errorMessage = 'Failed to load configuration';
      
      if (err?.response?.status === 401) {
        errorMessage = '未授权访问，请重新登录';
        try {
          localStorage.removeItem('user');
        } catch (storageError) {
          console.warn('Failed to clear user session:', storageError);
        }
        window.location.href = '/login?expired=true';
        return;
      } else if (err?.response?.status === 403) {
        errorMessage = '权限不足，仅管理员可访问此页面';
        window.location.href = '/forbidden';
        return;
      } else if (err?.response?.status === 404) {
        errorMessage = 'API接口不存在，请检查后端版本';
      } else if (err?.response?.status >= 500) {
        errorMessage = '服务器错误，请稍后再试';
      } else if (err?.message) {
        errorMessage = err.message;
      } else if (err?.response?.data?.message) {
        errorMessage = err.response.data.message;
      }
      
      const normalisedError = err instanceof Error ? err : new Error(errorMessage);
      normalisedError.message = errorMessage;
      setError(normalisedError);
    } finally {
      setLoading(false);
    }
  }, [requestSections]);

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
