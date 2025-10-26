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
import { Button, Card, Empty, Typography } from '@douyinfe/semi-ui';
import { AlertTriangle } from 'lucide-react';

class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
    };
  }

  static getDerivedStateFromError(error) {
    return { hasError: true };
  }

  componentDidCatch(error, errorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo);
    this.setState({
      error,
      errorInfo,
    });
  }

  handleReset = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
    });
    if (this.props.onReset) {
      this.props.onReset();
    }
  };

  handleReload = () => {
    window.location.reload();
  };

  render() {
    if (this.state.hasError) {
      const { error, errorInfo } = this.state;
      const { fallback } = this.props;

      if (fallback) {
        return fallback(error, this.handleReset);
      }

      return (
        <div className='flex items-center justify-center min-h-[400px] p-4'>
          <Card
            style={{ maxWidth: 600, width: '100%' }}
            bodyStyle={{ padding: '24px' }}
          >
            <Empty
              image={
                <div className='flex justify-center'>
                  <AlertTriangle size={64} color='#f53f3f' />
                </div>
              }
              title={
                <Typography.Title heading={4} style={{ marginTop: 16 }}>
                  页面加载失败
                </Typography.Title>
              }
              description={
                <div className='space-y-4'>
                  <Typography.Text type='tertiary'>
                    抱歉，页面遇到了一些问题。您可以尝试刷新页面或返回重试。
                  </Typography.Text>
                  {error && (
                    <div className='mt-4 p-3 bg-gray-50 rounded'>
                      <Typography.Text
                        type='danger'
                        size='small'
                        style={{
                          fontFamily: 'monospace',
                          wordBreak: 'break-word',
                        }}
                      >
                        {error.toString()}
                      </Typography.Text>
                    </div>
                  )}
                  {process.env.NODE_ENV === 'development' && errorInfo && (
                    <details className='mt-4 p-3 bg-gray-50 rounded'>
                      <summary className='cursor-pointer text-gray-600 text-sm mb-2'>
                        查看详细错误信息
                      </summary>
                      <pre
                        className='text-xs overflow-auto'
                        style={{
                          maxHeight: 300,
                          fontFamily: 'monospace',
                        }}
                      >
                        {errorInfo.componentStack}
                      </pre>
                    </details>
                  )}
                </div>
              }
            />
            <div className='flex gap-2 justify-center mt-6'>
              <Button onClick={this.handleReset} type='tertiary'>
                返回重试
              </Button>
              <Button onClick={this.handleReload} theme='solid' type='primary'>
                刷新页面
              </Button>
            </div>
          </Card>
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
