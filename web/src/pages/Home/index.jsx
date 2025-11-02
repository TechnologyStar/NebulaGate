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

import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react';
import {
  Button,
  Typography,
  Input,
  ScrollList,
  ScrollItem,
} from '@douyinfe/semi-ui';
import { API, showError, copy, showSuccess } from '../../helpers';
import { useIsMobile } from '../../hooks/common/useIsMobile';
import { API_ENDPOINTS } from '../../constants/common.constant';
import { StatusContext } from '../../context/Status';
import { useActualTheme } from '../../context/Theme';
import { marked } from 'marked';
import { useTranslation } from 'react-i18next';
import {
  IconGithubLogo,
  IconPlay,
  IconFile,
  IconCopy,
} from '@douyinfe/semi-icons';
import { Link } from 'react-router-dom';
import NoticeModal from '../../components/layout/NoticeModal';
import { Activity, BarChart2, Clock3, Layers } from 'lucide-react';
import {
  Moonshot,
  OpenAI,
  XAI,
  Zhipu,
  Volcengine,
  Cohere,
  Claude,
  Gemini,
  Suno,
  Minimax,
  Wenxin,
  Spark,
  Qingyan,
  DeepSeek,
  Qwen,
  Midjourney,
  Grok,
  AzureAI,
  Hunyuan,
  Xinference,
} from '@lobehub/icons';

const { Text } = Typography;

const Home = () => {
  const { t, i18n } = useTranslation();
  const [statusState] = useContext(StatusContext);
  const actualTheme = useActualTheme();
  const [homePageContentLoaded, setHomePageContentLoaded] = useState(false);
  const [homePageContent, setHomePageContent] = useState('');
  const [noticeVisible, setNoticeVisible] = useState(false);
  const isMobile = useIsMobile();
  const isDemoSiteMode = statusState?.status?.demo_site_enabled || false;
  const docsLink = statusState?.status?.docs_link || '';
  const serverAddress =
    statusState?.status?.server_address || `${window.location.origin}`;
  const endpointItems = API_ENDPOINTS.map((e) => ({ value: e }));
  const [endpointIndex, setEndpointIndex] = useState(0);
  const isChinese = i18n.language.startsWith('zh');
  const [runtimeStats, setRuntimeStats] = useState({
    activeConnections: 0,
    uptimeSeconds: 0,
    version: statusState?.status?.version || '',
    apiInfoCount: Array.isArray(statusState?.status?.api_info)
      ? statusState.status.api_info.length
      : 0,
  });
  const [leaderboardData, setLeaderboardData] = useState([]);
  const [liveStatsLoading, setLiveStatsLoading] = useState(false);
  const [liveUpdatedAt, setLiveUpdatedAt] = useState(null);
  const [liveError, setLiveError] = useState(null);

  const displayHomePageContent = async () => {
    setHomePageContent(localStorage.getItem('home_page_content') || '');
    const res = await API.get('/api/home_page_content');
    const { success, message, data } = res.data;
    if (success) {
      let content = data;
      if (!data.startsWith('https://')) {
        content = marked.parse(data);
      }
      setHomePageContent(content);
      localStorage.setItem('home_page_content', content);

      // 如果内容是 URL，则发送主题模式
      if (data.startsWith('https://')) {
        const iframe = document.querySelector('iframe');
        if (iframe) {
          iframe.onload = () => {
            iframe.contentWindow.postMessage({ themeMode: actualTheme }, '*');
            iframe.contentWindow.postMessage({ lang: i18n.language }, '*');
          };
        }
      }
    } else {
      showError(message);
      setHomePageContent('加载首页内容失败...');
    }
    setHomePageContentLoaded(true);
  };

  const handleCopyBaseURL = async () => {
    const ok = await copy(serverAddress);
    if (ok) {
      showSuccess(t('已复制到剪切板'));
    }
  };

  useEffect(() => {
    const checkNoticeAndShow = async () => {
      const lastCloseDate = localStorage.getItem('notice_close_date');
      const today = new Date().toDateString();
      if (lastCloseDate !== today) {
        try {
          const res = await API.get('/api/notice');
          const { success, data } = res.data;
          if (success && data && data.trim() !== '') {
            setNoticeVisible(true);
          }
        } catch (error) {
          console.error('获取公告失败:', error);
        }
      }
    };

    checkNoticeAndShow();
  }, []);

  useEffect(() => {
    displayHomePageContent().then();
  }, []);

  useEffect(() => {
    const timer = setInterval(() => {
      setEndpointIndex((prev) => (prev + 1) % endpointItems.length);
    }, 3000);
    return () => clearInterval(timer);
  }, [endpointItems.length]);

  const defaultNumberFormatter = useMemo(
    () =>
      new Intl.NumberFormat(i18n.language, {
        maximumFractionDigits: 0,
      }),
    [i18n.language],
  );

  const compactNumberFormatter = useMemo(
    () =>
      new Intl.NumberFormat(i18n.language, {
        notation: 'compact',
        maximumFractionDigits: 1,
      }),
    [i18n.language],
  );

  const formatNumber = useCallback(
    (value) => {
      if (typeof value !== 'number' || Number.isNaN(value)) {
        return defaultNumberFormatter.format(0);
      }
      if (value < 1000) {
        return defaultNumberFormatter.format(value);
      }
      return compactNumberFormatter.format(value);
    },
    [compactNumberFormatter, defaultNumberFormatter],
  );

  const formatUptime = useCallback((seconds) => {
    if (!seconds || seconds < 0) return '—';
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    if (days > 0) {
      return `${days}d ${hours}h`;
    }
    if (hours > 0) {
      return `${hours}h ${minutes}m`;
    }
    if (minutes > 0) {
      return `${minutes}m`;
    }
    return '<1m';
  }, []);

  const fetchLiveData = useCallback(async () => {
    setLiveStatsLoading(true);
    setLiveError(null);

    try {
      const statusRes = await API.get('/api/status', {
        params: { _: Date.now() },
        skipErrorHandler: true,
      });
      const { success, data, message } = statusRes?.data || {};
      if (success && data) {
        const runtime = data.runtime_stats || {};
        setRuntimeStats({
          activeConnections: runtime.active_connections ?? 0,
          uptimeSeconds: runtime.uptime_seconds ?? 0,
          version: data.version || '',
          apiInfoCount: Array.isArray(data.api_info) ? data.api_info.length : 0,
        });
      } else if (message) {
        setLiveError(message);
      }
    } catch (error) {
      const message =
        error?.response?.data?.message ||
        error?.message ||
        t('实时数据暂不可用');
      setLiveError(message);
    }

    try {
      const leaderboardRes = await API.get('/api/public/leaderboard', {
        params: { window: '24h', limit: 6 },
        skipErrorHandler: true,
      });
      const { success, data, message } = leaderboardRes?.data || {};
      if (success && Array.isArray(data)) {
        setLeaderboardData(data);
      } else {
        setLeaderboardData([]);
        if (message) {
          setLiveError((prev) => prev ?? message);
        }
      }
    } catch (error) {
      const message =
        error?.response?.data?.message ||
        error?.message ||
        t('排行榜功能未启用');
      setLeaderboardData([]);
      setLiveError((prev) => prev ?? message);
    } finally {
      setLiveUpdatedAt(new Date());
      setLiveStatsLoading(false);
    }
  }, [t]);

  useEffect(() => {
    fetchLiveData();
    const interval = setInterval(fetchLiveData, 15000);
    return () => clearInterval(interval);
  }, [fetchLiveData]);

  const statCards = useMemo(() => {
    const totalRequests = leaderboardData.reduce(
      (sum, entry) => sum + (entry.request_count || 0),
      0,
    );
    const uniqueModels = leaderboardData.length;
    return [
      {
        key: 'active',
        title: t('活跃连接'),
        value: formatNumber(runtimeStats.activeConnections || 0),
        meta: t('当前与网关交互的会话'),
        Icon: Activity,
      },
      {
        key: 'uptime',
        title: t('系统运行时间'),
        value: formatUptime(runtimeStats.uptimeSeconds || 0),
        meta: t('自最近启动以来'),
        Icon: Clock3,
      },
      {
        key: 'requests',
        title: t('24小时请求量'),
        value: formatNumber(totalRequests),
        meta: t('公开排行榜统计'),
        Icon: BarChart2,
      },
      {
        key: 'models',
        title: t('活跃模型数'),
        value: formatNumber(uniqueModels),
        meta: t('统计窗口内的模型覆盖'),
        Icon: Layers,
      },
    ];
  }, [formatNumber, formatUptime, leaderboardData, runtimeStats, t]);

  return (
    <div className='w-full overflow-x-hidden'>
      <NoticeModal
        visible={noticeVisible}
        onClose={() => setNoticeVisible(false)}
        isMobile={isMobile}
      />
      {homePageContentLoaded && homePageContent === '' ? (
        <div className='w-full overflow-x-hidden'>
          {/* Hero Banner Section */}
          <div className='home-hero-shell w-full border-b border-semi-color-border min-h-[600px] md:min-h-[700px] lg:min-h-[800px] relative overflow-x-hidden bg-gradient-to-br from-slate-50 via-blue-50 to-teal-50 dark:from-slate-900 dark:via-blue-950 dark:to-teal-950'>
            {/* Background blur orbs */}
            <div className='blur-ball blur-ball-indigo' />
            <div className='blur-ball blur-ball-teal' />
            <div className='flex items-center justify-center h-full px-4 py-24 md:py-32 lg:py-40 mt-10'>
              {/* Center content area */}
              <div className='flex flex-col items-center justify-center text-center max-w-5xl mx-auto'>
                <div className='flex flex-col items-center justify-center mb-8 md:mb-10'>
                  {/* Brand badge */}
                  <div className='mb-6 md:mb-8'>
                    <span className='inline-flex items-center px-4 py-2 rounded-full bg-white/80 dark:bg-slate-800/80 backdrop-blur-sm border border-blue-200/50 dark:border-blue-700/50 text-sm font-medium text-blue-700 dark:text-blue-300 shadow-sm'>
                      <svg className='w-4 h-4 mr-2' fill='currentColor' viewBox='0 0 20 20'>
                        <path fillRule='evenodd' d='M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z' clipRule='evenodd' />
                      </svg>
                      {t('企业级AI模型网关')}
                    </span>
                  </div>

                    <h1
                      className={`home-hero-title text-5xl md:text-6xl lg:text-7xl xl:text-8xl font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-blue-600 via-teal-600 to-indigo-600 dark:from-blue-400 dark:via-teal-400 dark:to-indigo-400 leading-tight mb-6 ${isChinese ? 'tracking-wide md:tracking-wider' : ''}`}
                    >
                    {i18n.language === 'en' ? (
                      <>
                        The Unified
                        <br />
                        <span className='shine-text'>AI Gateway</span>
                      </>
                    ) : (
                      <>
                        统一的
                        <br />
                        <span className='shine-text'>AI 模型网关</span>
                      </>
                    )}
                  </h1>
                  
                  <p className='home-hero-subtitle text-lg md:text-xl lg:text-2xl text-semi-color-text-1 dark:text-slate-300 mt-4 md:mt-6 max-w-2xl font-light leading-relaxed'>
                    {t('企业级稳定性，透明化定价，一键接入多家AI模型')}
                  </p>
                  
                  <p className='text-base md:text-lg text-semi-color-text-2 mt-3 max-w-xl'>
                    {t('更好的价格，更好的稳定性，只需要将模型基址替换为：')}
                  </p>
                  {/* BASE URL 与端点选择 */}
                  <div className='flex flex-col md:flex-row items-center justify-center gap-4 w-full mt-4 md:mt-6 max-w-md'>
                    <Input
                      readonly
                      value={serverAddress}
                      className='flex-1 !rounded-full'
                      size={isMobile ? 'default' : 'large'}
                      suffix={
                        <div className='flex items-center gap-2'>
                          <ScrollList
                            bodyHeight={32}
                            style={{ border: 'unset', boxShadow: 'unset' }}
                          >
                            <ScrollItem
                              mode='wheel'
                              cycled={true}
                              list={endpointItems}
                              selectedIndex={endpointIndex}
                              onSelect={({ index }) => setEndpointIndex(index)}
                            />
                          </ScrollList>
                          <Button
                            type='primary'
                            onClick={handleCopyBaseURL}
                            icon={<IconCopy />}
                            className='!rounded-full'
                          />
                        </div>
                      }
                    />
                  </div>
                </div>

                {/* Action buttons */}
                <div className='flex flex-col sm:flex-row gap-4 justify-center items-center mt-8'>
                  <Link to='/console'>
                    <Button
                      theme='solid'
                      type='primary'
                      size={isMobile ? 'large' : 'large'}
                      className='!rounded-3xl px-10 py-3 shadow-lg hover:shadow-xl transition-all duration-300 bg-gradient-to-r from-blue-600 to-teal-600 hover:from-blue-700 hover:to-teal-700'
                      icon={<IconPlay />}
                    >
                      {t('获取密钥')}
                    </Button>
                  </Link>
                  {isDemoSiteMode && statusState?.status?.version ? (
                    <Button
                      size={isMobile ? 'large' : 'large'}
                      className='flex items-center !rounded-3xl px-8 py-3 border-2 border-blue-200 dark:border-blue-800 hover:border-blue-300 dark:hover:border-blue-700 bg-white/50 dark:bg-slate-800/50 backdrop-blur-sm hover:bg-white dark:hover:bg-slate-800 transition-all duration-300'
                      icon={<IconGithubLogo />}
                      onClick={() =>
                        window.open(
                          'https://github.com/QuantumNous/new-api',
                          '_blank',
                        )
                      }
                    >
                      {statusState.status.version}
                    </Button>
                  ) : (
                    docsLink && (
                      <Button
                        size={isMobile ? 'large' : 'large'}
                        className='flex items-center !rounded-3xl px-8 py-3 border-2 border-blue-200 dark:border-blue-800 hover:border-blue-300 dark:hover:border-blue-700 bg-white/50 dark:bg-slate-800/50 backdrop-blur-sm hover:bg-white dark:hover:bg-slate-800 transition-all duration-300'
                        icon={<IconFile />}
                        onClick={() => window.open(docsLink, '_blank')}
                      >
                        {t('文档')}
                      </Button>
                    )
                  )}
                </div>

                <div className='w-full max-w-5xl mx-auto'>
                  <div className='home-live-stats'>
                    {statCards.map((card) => {
                      const IconComponent = card.Icon;
                      return (
                        <div key={card.key} className='home-live-card'>
                          <div className='flex items-center justify-between mb-3'>
                            <h3 className='flex items-center gap-2 uppercase tracking-[0.28em] text-[10px] text-slate-500 dark:text-slate-400'>
                              <IconComponent size={16} />
                              {card.title}
                            </h3>
                            {liveStatsLoading && (
                              <span className='text-[10px] font-medium text-slate-400 dark:text-slate-500 uppercase tracking-[0.3em]'>
                                {t('刷新中')}
                              </span>
                            )}
                          </div>
                          <div className='home-live-card-value'>{card.value}</div>
                          <div className='home-live-card-meta'>{card.meta}</div>
                        </div>
                      );
                    })}
                  </div>
                  <div className='mt-3 text-xs text-slate-500 dark:text-slate-400 flex flex-wrap items-center gap-3'>
                    {(runtimeStats.version || statusState?.status?.version) && (
                      <span>
                        {t('当前版本 {{version}}', {
                          version: runtimeStats.version || statusState?.status?.version,
                        })}
                      </span>
                    )}
                    {runtimeStats.apiInfoCount > 0 && (
                      <span>{t('已开放 API：{{count}}', { count: runtimeStats.apiInfoCount })}</span>
                    )}
                    <span className='ml-auto'>
                      {liveStatsLoading
                        ? t('正在刷新实时数据…')
                        : liveUpdatedAt
                        ? t('更新于 {{time}}', {
                            time: liveUpdatedAt.toLocaleTimeString(),
                          })
                        : liveError || ''}
                    </span>
                  </div>
                  {liveError && (
                    <div className='mt-3 text-sm text-amber-600 dark:text-amber-300'>
                      {liveError}
                    </div>
                  )}
                  {leaderboardData.length > 0 && (
                    <div className='mt-8'>
                      <div className='flex items-center justify-between mb-4'>
                        <h3 className='text-lg font-semibold text-slate-800 dark:text-slate-100 flex items-center gap-2'>
                          <BarChart2 size={18} />
                          {t('模型请求排行榜')}
                        </h3>
                        <span className='text-xs text-slate-500 dark:text-slate-400'>
                          {t('统计窗口：{{window}}', { window: t('24小时') })}
                        </span>
                      </div>
                      <div className='grid gap-3 md:grid-cols-2'>
                        {leaderboardData.slice(0, 4).map((item, index) => (
                          <div
                            key={`${item.model}-${index}`}
                            className='flex items-center justify-between rounded-2xl border border-slate-200/60 dark:border-slate-700/60 bg-white/70 dark:bg-slate-800/70 px-5 py-4 backdrop-blur-lg shadow-sm hover:shadow-lg transition-all'
                          >
                            <div className='flex items-center gap-3'>
                              <span className='text-sm font-semibold text-slate-400 dark:text-slate-500'>#{index + 1}</span>
                              <div className='flex flex-col'>
                                <span className='text-base font-semibold text-slate-800 dark:text-slate-100'>
                                  {item.model}
                                </span>
                                <span className='text-xs text-slate-500 dark:text-slate-400'>
                                  {t('唯一调用令牌：{{count}}', {
                                    count: formatNumber(item.unique_tokens || 0),
                                  })}
                                </span>
                              </div>
                            </div>
                            <div className='text-right'>
                              <div className='text-lg font-bold text-slate-900 dark:text-slate-50'>
                                {formatNumber(item.request_count || 0)}
                              </div>
                              <div className='text-xs text-slate-500 dark:text-slate-400'>
                                {t('请求量')}
                              </div>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>

                {/* 框架兼容性图标 */}
                <div className='mt-12 md:mt-16 lg:mt-20 w-full'>
                  <div className='flex items-center mb-6 md:mb-8 justify-center'>
                    <Text
                      type='tertiary'
                      className='text-lg md:text-xl lg:text-2xl font-light'
                    >
                      {t('支持众多的大模型供应商')}
                    </Text>
                  </div>
                  <div className='flex flex-wrap items-center justify-center gap-3 sm:gap-4 md:gap-6 lg:gap-8 max-w-5xl mx-auto px-4'>
                    <div className='w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12 flex items-center justify-center'>
                      <Moonshot size={40} />
                    </div>
                    <div className='w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12 flex items-center justify-center'>
                      <OpenAI size={40} />
                    </div>
                    <div className='w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12 flex items-center justify-center'>
                      <XAI size={40} />
                    </div>
                    <div className='w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12 flex items-center justify-center'>
                      <Zhipu.Color size={40} />
                    </div>
                    <div className='w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12 flex items-center justify-center'>
                      <Volcengine.Color size={40} />
                    </div>
                    <div className='w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12 flex items-center justify-center'>
                      <Cohere.Color size={40} />
                    </div>
                    <div className='w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12 flex items-center justify-center'>
                      <Claude.Color size={40} />
                    </div>
                    <div className='w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12 flex items-center justify-center'>
                      <Gemini.Color size={40} />
                    </div>
                    <div className='w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12 flex items-center justify-center'>
                      <Suno size={40} />
                    </div>
                    <div className='w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12 flex items-center justify-center'>
                      <Minimax.Color size={40} />
                    </div>
                    <div className='w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12 flex items-center justify-center'>
                      <Wenxin.Color size={40} />
                    </div>
                    <div className='w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12 flex items-center justify-center'>
                      <Spark.Color size={40} />
                    </div>
                    <div className='w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12 flex items-center justify-center'>
                      <Qingyan.Color size={40} />
                    </div>
                    <div className='w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12 flex items-center justify-center'>
                      <DeepSeek.Color size={40} />
                    </div>
                    <div className='w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12 flex items-center justify-center'>
                      <Qwen.Color size={40} />
                    </div>
                    <div className='w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12 flex items-center justify-center'>
                      <Midjourney size={40} />
                    </div>
                    <div className='w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12 flex items-center justify-center'>
                      <Grok size={40} />
                    </div>
                    <div className='w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12 flex items-center justify-center'>
                      <AzureAI.Color size={40} />
                    </div>
                    <div className='w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12 flex items-center justify-center'>
                      <Hunyuan.Color size={40} />
                    </div>
                    <div className='w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12 flex items-center justify-center'>
                      <Xinference.Color size={40} />
                    </div>
                    <div className='w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12 flex items-center justify-center'>
                      <Typography.Text className='!text-lg sm:!text-xl md:!text-2xl lg:!text-3xl font-bold'>
                        30+
                      </Typography.Text>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      ) : (
        <div className='overflow-x-hidden w-full'>
          {homePageContent.startsWith('https://') ? (
            <iframe
              src={homePageContent}
              className='w-full h-screen border-none'
            />
          ) : (
            <div
              className='mt-[60px]'
              dangerouslySetInnerHTML={{ __html: homePageContent }}
            />
          )}
        </div>
      )}
    </div>
  );
};

export default Home;
