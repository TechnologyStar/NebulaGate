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
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { I18nextProvider } from 'react-i18next';
import SecurityCenter from '../pages/Security';
import { API } from '../helpers';

// Mock API calls
jest.mock('../helpers', () => ({
  API: {
    get: jest.fn(),
    post: jest.fn(),
    put: jest.fn(),
    delete: jest.fn(),
  },
}));

// Mock i18n
const i18n = {
  t: (key) => key,
  changeLanguage: () => {},
};

describe('SecurityCenter', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  const renderWithProviders = (component) => {
    return render(
      <I18nextProvider i18n={i18n}>
        <BrowserRouter>
          {component}
        </BrowserRouter>
      </I18nextProvider>
    );
  };

  test('renders security center with tabs', async () => {
    // Mock dashboard API call
    API.get.mockResolvedValue({
      data: {
        success: true,
        data: {
          total_count: 100,
          unique_users: 50,
          today_count: 10,
          active_devices: 25,
          unique_ips: 30,
          active_anomalies: 5,
          anomaly_trend: [],
          device_clusters: [],
          ip_analytics: [],
          top_keywords: [],
        },
      },
    });

    renderWithProviders(<SecurityCenter />);

    await waitFor(() => {
      expect(screen.getByText('security.title')).toBeInTheDocument();
      expect(screen.getByText('security.dashboard')).toBeInTheDocument();
      expect(screen.getByText('security.deviceClusters')).toBeInTheDocument();
      expect(screen.getByText('security.ipAnalytics')).toBeInTheDocument();
      expect(screen.getByText('security.anomalyManagement')).toBeInTheDocument();
      expect(screen.getByText('security.violations')).toBeInTheDocument();
      expect(screen.getByText('security.users')).toBeInTheDocument();
      expect(screen.getByText('security.settings')).toBeInTheDocument();
    });
  });

  test('loads dashboard stats on mount', async () => {
    const mockData = {
      total_count: 100,
      unique_users: 50,
      today_count: 10,
      active_devices: 25,
      unique_ips: 30,
      active_anomalies: 5,
    };

    API.get.mockResolvedValue({
      data: { success: true, data: mockData },
    });

    renderWithProviders(<SecurityCenter />);

    await waitFor(() => {
      expect(API.get).toHaveBeenCalledWith('/api/security/dashboard');
      expect(screen.getByText('100')).toBeInTheDocument(); // total violations
      expect(screen.getByText('50')).toBeInTheDocument(); // unique users
      expect(screen.getByText('10')).toBeInTheDocument(); // today violations
      expect(screen.getByText('25')).toBeInTheDocument(); // active devices
      expect(screen.getByText('30')).toBeInTheDocument(); // unique ips
      expect(screen.getByText('5')).toBeInTheDocument(); // active anomalies
    });
  });

  test('handles tab navigation', async () => {
    API.get.mockResolvedValue({
      data: { success: true, data: { devices: [], total: 0 } },
    });

    renderWithProviders(<SecurityCenter />);

    // Click on devices tab
    const devicesTab = screen.getByText('security.deviceClusters');
    fireEvent.click(devicesTab);

    await waitFor(() => {
      expect(API.get).toHaveBeenCalledWith('/api/security/devices', {
        params: { page: 1, page_size: 10 },
      });
    });
  });

  test('displays device management interface', async () => {
    const mockDevices = [
      {
        device_id: 'device1',
        device_type: 'mobile',
        user_agent: 'Mozilla/5.0...',
        user_count: 5,
        request_count: 100,
        anomaly_score: 0.3,
        is_blocked: false,
        is_nat: false,
      },
    ];

    API.get.mockResolvedValue({
      data: { success: true, data: { devices: mockDevices, total: 1 } },
    });

    renderWithProviders(<SecurityCenter />);

    // Navigate to devices tab
    const devicesTab = screen.getByText('security.deviceClusters');
    fireEvent.click(devicesTab);

    await waitFor(() => {
      expect(screen.getByText('device1')).toBeInTheDocument();
      expect(screen.getByText('mobile')).toBeInTheDocument();
      expect(screen.getByText('5')).toBeInTheDocument(); // user count
      expect(screen.getByText('100')).toBeInTheDocument(); // request count
    });
  });

  test('handles anomaly actions', async () => {
    const mockAnomalies = [
      {
        id: 1,
        anomaly_type: 'suspicious',
        user_id: 'user1',
        target_identifier: 'device1',
        severity: 'medium',
        description: 'Unusual activity detected',
        status: 'pending',
        detected_at: '2024-01-01T12:00:00Z',
      },
    ];

    API.get.mockResolvedValue({
      data: { success: true, data: { anomalies: mockAnomalies, total: 1 } },
    });

    API.post.mockResolvedValue({
      data: { success: true },
    });

    renderWithProviders(<SecurityCenter />);

    // Navigate to anomalies tab
    const anomaliesTab = screen.getByText('security.anomalyManagement');
    fireEvent.click(anomalies);

    await waitFor(() => {
      expect(screen.getByText('suspicious')).toBeInTheDocument();
      expect(screen.getByText('user1')).toBeInTheDocument();
      expect(screen.getByText('device1')).toBeInTheDocument();
      expect(screen.getByText('Unusual activity detected')).toBeInTheDocument();
    });

    // Test ban action
    const banButton = screen.getByText('security.ban');
    fireEvent.click(banButton);

    await waitFor(() => {
      expect(API.post).toHaveBeenCalledWith('/api/security/anomalies/1/action', {
        action: 'ban',
      });
    });
  });

  test('handles settings form submission', async () => {
    API.get.mockResolvedValue({
      data: { success: true, data: {} },
    });

    API.put.mockResolvedValue({
      data: { success: true },
    });

    renderWithProviders(<SecurityCenter />);

    // Navigate to settings tab
    const settingsTab = screen.getByText('security.settings');
    fireEvent.click(settingsTab);

    await waitFor(() => {
      expect(screen.getByText('security.enforcementSettings')).toBeInTheDocument();
      expect(screen.getByText('security.detectionSettings')).toBeInTheDocument();
      expect(screen.getByText('security.notificationSettings')).toBeInTheDocument();
    });

    // Fill form and submit
    const modelInput = screen.getByPlaceholderText('gpt-3.5-turbo');
    fireEvent.change(modelInput, { target: { value: 'gpt-4' } });

    const thresholdInput = screen.getByPlaceholderText('10');
    fireEvent.change(thresholdInput, { target: { value: 20 } });

    const saveButton = screen.getByText('common.save');
    fireEvent.click(saveButton);

    await waitFor(() => {
      expect(API.put).toHaveBeenCalledWith('/api/security/settings', {
        violation_redirect_model: 'gpt-4',
        auto_ban_threshold: 20,
      });
    });
  });

  test('displays error messages on API failure', async () => {
    // Mock console.error to avoid test output noise
    const originalError = console.error;
    console.error = jest.fn();

    API.get.mockRejectedValue(new Error('Network error'));

    renderWithProviders(<SecurityCenter />);

    await waitFor(() => {
      // Should handle API errors gracefully
      expect(console.error).toHaveBeenCalled();
    });

    console.error = originalError;
  });

  test('handles device details modal', async () => {
    const mockDevice = {
      device_id: 'device1',
      device_type: 'mobile',
      user_agent: 'Mozilla/5.0...',
      user_count: 5,
      request_count: 100,
      anomaly_score: 0.3,
      is_blocked: false,
      is_nat: false,
      associated_users: [
        { user_id: 'user1', username: 'testuser', last_seen: '2024-01-01T12:00:00Z', request_count: 50 },
      ],
    };

    API.get.mockResolvedValue({
      data: { success: true, data: { devices: [mockDevice], total: 1 } },
    });

    renderWithProviders(<SecurityCenter />);

    // Navigate to devices tab
    const devicesTab = screen.getByText('security.deviceClusters');
    fireEvent.click(devicesTab);

    await waitFor(() => {
      expect(screen.getByText('security.viewDetails')).toBeInTheDocument();
    });

    // Click view details
    const viewDetailsButton = screen.getByText('security.viewDetails');
    fireEvent.click(viewDetailsButton);

    await waitFor(() => {
      expect(screen.getByText('security.deviceDetails')).toBeInTheDocument();
      expect(screen.getByText('device1')).toBeInTheDocument();
      expect(screen.getByText('mobile')).toBeInTheDocument();
      expect(screen.getByText('testuser')).toBeInTheDocument();
    });
  });
});