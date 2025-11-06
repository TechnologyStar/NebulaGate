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

describe('Security Center E2E Tests', () => {
  beforeEach(() => {
    // Mock API responses
    cy.intercept('GET', '/api/security/dashboard', {
      body: {
        success: true,
        data: {
          total_count: 100,
          unique_users: 50,
          today_count: 10,
          active_devices: 25,
          unique_ips: 30,
          active_anomalies: 5,
          anomaly_trend: [
            { time: '2024-01-01', type: 'malicious', count: 5 },
            { time: '2024-01-02', type: 'suspicious', count: 3 },
          ],
          device_clusters: [
            { device_type: 'mobile', count: 40 },
            { device_type: 'desktop', count: 35 },
            { device_type: 'tablet', count: 25 },
          ],
          ip_analytics: [
            { ip: '192.168.1.1', request_count: 100 },
            { ip: '10.0.0.1', request_count: 80 },
          ],
          top_keywords: [
            { keyword: 'test', count: 15 },
            { keyword: 'admin', count: 10 },
          ],
        },
      },
    }).as('dashboardData');

    cy.intercept('GET', '/api/security/devices', {
      body: {
        success: true,
        data: {
          devices: [
            {
              device_id: 'device1',
              device_type: 'mobile',
              user_agent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)',
              user_count: 5,
              request_count: 100,
              anomaly_score: 0.3,
              is_blocked: false,
              is_nat: false,
            },
            {
              device_id: 'device2',
              device_type: 'desktop',
              user_agent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
              user_count: 3,
              request_count: 50,
              anomaly_score: 0.7,
              is_blocked: true,
              is_nat: true,
            },
          ],
          total: 2,
        },
      },
    }).as('devicesData');

    cy.intercept('GET', '/api/security/anomalies', {
      body: {
        success: true,
        data: {
          anomalies: [
            {
              id: 1,
              anomaly_type: 'suspicious',
              user_id: 'user1',
              target_identifier: 'device1',
              severity: 'medium',
              description: 'Unusual login pattern detected',
              status: 'pending',
              detected_at: '2024-01-01T12:00:00Z',
            },
            {
              id: 2,
              anomaly_type: 'malicious',
              user_id: 'user2',
              target_identifier: '192.168.1.1',
              severity: 'high',
              description: 'Multiple failed login attempts',
              status: 'investigating',
              detected_at: '2024-01-02T14:30:00Z',
            },
          ],
          total: 2,
        },
      },
    }).as('anomaliesData');

    cy.intercept('POST', '/api/security/anomalies/1/action', {
      body: { success: true },
    }).as('banAnomaly');

    cy.intercept('POST', '/api/security/devices/device1/action', {
      body: { success: true },
    }).as('blockDevice');

    // Visit security page
    cy.visit('/security');
    cy.wait('@dashboardData');
  });

  it('should display enhanced dashboard with new metrics', () => {
    // Check enhanced metrics cards
    cy.get('[data-testid="total-violations"]').should('contain', '100');
    cy.get('[data-testid="unique-users"]').should('contain', '50');
    cy.get('[data-testid="today-violations"]').should('contain', '10');
    cy.get('[data-testid="active-devices"]').should('contain', '25');
    cy.get('[data-testid="unique-ips"]').should('contain', '30');
    cy.get('[data-testid="active-anomalies"]').should('contain', '5');

    // Check charts are rendered
    cy.get('[data-testid="anomaly-trend-chart"]').should('be.visible');
    cy.get('[data-testid="device-cluster-chart"]').should('be.visible');
    cy.get('[data-testid="ip-analytics-chart"]').should('be.visible');

    // Check response actions summary
    cy.get('[data-testid="response-actions-summary"]').should('be.visible');
  });

  it('should navigate between tabs', () => {
    // Click on Device Clusters tab
    cy.contains('security.deviceClusters').click();
    cy.wait('@devicesData');
    cy.url().should('include', 'devices');

    // Click on IP Analytics tab
    cy.contains('security.ipAnalytics').click();
    cy.url().should('include', 'ips');

    // Click on Anomaly Management tab
    cy.contains('security.anomalyManagement').click();
    cy.wait('@anomaliesData');
    cy.url().should('include', 'anomalies');

    // Click back to Dashboard
    cy.contains('security.dashboard').click();
    cy.wait('@dashboardData');
    cy.url().should('include', 'dashboard');
  });

  it('should display device clusters with filtering', () => {
    cy.contains('security.deviceClusters').click();
    cy.wait('@devicesData');

    // Check device table
    cy.get('[data-testid="devices-table"]').should('be.visible');
    cy.contains('device1').should('be.visible');
    cy.contains('device2').should('be.visible');
    cy.contains('mobile').should('be.visible');
    cy.contains('desktop').should('be.visible');

    // Test filtering
    cy.get('[data-testid="device-type-filter"]').type('mobile');
    cy.get('[data-testid="apply-filter"]').click();

    // Should only show mobile devices
    cy.contains('device1').should('be.visible');
    cy.contains('device2').should('not.exist');
  });

  it('should handle anomaly management actions', () => {
    cy.contains('security.anomalyManagement').click();
    cy.wait('@anomaliesData');

    // Check anomalies table
    cy.get('[data-testid="anomalies-table"]').should('be.visible');
    cy.contains('suspicious').should('be.visible');
    cy.contains('malicious').should('be.visible');
    cy.contains('user1').should('be.visible');
    cy.contains('Unusual login pattern detected').should('be.visible');

    // Test ban action
    cy.contains('security.ban').first().click();
    cy.get('[data-testid="confirm-modal"]').should('be.visible');
    cy.get('[data-testid="confirm-button"]').click();
    cy.wait('@banAnomaly');

    // Should show success message
    cy.get('[data-testid="success-toast"]').should('be.visible');
  });

  it('should display device details modal', () => {
    cy.contains('security.deviceClusters').click();
    cy.wait('@devicesData');

    // Click view details
    cy.contains('security.viewDetails').first().click();

    // Check modal content
    cy.get('[data-testid="device-details-modal"]').should('be.visible');
    cy.contains('security.deviceDetails').should('be.visible');
    cy.contains('device1').should('be.visible');
    cy.contains('mobile').should('be.visible');
    cy.contains('Mozilla/5.0').should('be.visible');

    // Check associated users table
    cy.get('[data-testid="associated-users-table"]').should('be.visible');

    // Close modal
    cy.get('[data-testid="close-modal"]').click();
    cy.get('[data-testid="device-details-modal"]').should('not.exist');
  });

  it('should handle settings configuration', () => {
    cy.contains('security.settings').click();

    // Check settings sections
    cy.contains('security.enforcementSettings').should('be.visible');
    cy.contains('security.detectionSettings').should('be.visible');
    cy.contains('security.notificationSettings').should('be.visible');

    // Fill enforcement settings
    cy.get('[data-testid="violation-redirect-model"]')
      .clear()
      .type('gpt-4-turbo');

    cy.get('[data-testid="auto-ban-enabled"]').click();

    cy.get('[data-testid="auto-ban-threshold"]')
      .clear()
      .type('15');

    // Fill detection settings
    cy.get('[data-testid="anomaly-detection-enabled"]').click();

    cy.get('[data-testid="anomaly-threshold-score"]')
      .clear()
      .type('0.8');

    // Fill notification settings
    cy.get('[data-testid="real-time-alerts-enabled"]').click();

    cy.get('[data-testid="alert-severity-threshold"]').select('high');

    // Save settings
    cy.get('[data-testid="save-settings"]').click();

    // Should show success message
    cy.get('[data-testid="success-toast"]').should('be.visible');
  });

  it('should validate form inputs', () => {
    cy.contains('security.settings').click();

    // Test invalid model format
    cy.get('[data-testid="violation-redirect-model"]')
      .clear()
      .type('invalid model@#$');

    cy.get('[data-testid="save-settings"]').click();

    // Should show validation error
    cy.get('[data-testid="validation-error"]').should('be.visible');
    cy.contains('security.invalidModelFormat').should('be.visible');

    // Test threshold validation
    cy.get('[data-testid="auto-ban-threshold"]')
      .clear()
      .type('2000'); // Over max limit

    cy.get('[data-testid="save-settings"]').click();

    cy.contains('security.thresholdRange').should('be.visible');
  });

  it('should handle responsive design', () => {
    // Test mobile view
    cy.viewport(375, 667); // iPhone dimensions
    cy.visit('/security');
    cy.wait('@dashboardData');

    // Should stack metrics vertically
    cy.get('[data-testid="metrics-row"]').should('have.class', 'responsive-stack');

    // Charts should be full width
    cy.get('[data-testid="chart-container"]').should('have.css', 'width', '100%');

    // Test tablet view
    cy.viewport(768, 1024); // iPad dimensions
    cy.get('[data-testid="metrics-row"]').should('not.have.class', 'responsive-stack');

    // Test desktop view
    cy.viewport(1200, 800); // Desktop dimensions
    cy.get('[data-testid="tabs-container"]').should('be.visible');
  });

  it('should handle keyboard navigation and accessibility', () => {
    // Test tab navigation
    cy.get('body').tab();
    cy.focused().should('contain', 'security.dashboard');

    // Navigate through tabs using arrow keys
    cy.get('body').type('{rightArrow}');
    cy.focused().should('contain', 'security.deviceClusters');

    // Test modal focus management
    cy.contains('security.deviceClusters').click();
    cy.wait('@devicesData');
    cy.contains('security.viewDetails').first().click();

    // Focus should be trapped in modal
    cy.focused().should('be.within', '[data-testid="device-details-modal"]');

    // Test escape key to close modal
    cy.get('body').type('{esc}');
    cy.get('[data-testid="device-details-modal"]').should('not.exist');
  });
});