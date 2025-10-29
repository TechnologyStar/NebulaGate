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
import { API, showError } from '../../../helpers';
import { Empty, Card, Spin, Typography } from '@douyinfe/semi-ui';
const { Title } = Typography;
import {
  IllustrationConstruction,
  IllustrationConstructionDark,
} from '@douyinfe/semi-illustrations';
import { useTranslation } from 'react-i18next';
import MarkdownRenderer from '../markdown/MarkdownRenderer';

// 检查是否为 URL
const isUrl = (content) => {
  try {
    new URL(content.trim());
    return true;
  } catch {
    return false;
  }
};

// 检查是否为 HTML 内容
const isHtmlContent = (content) => {
  if (!content || typeof content !== 'string') return false;
  
  // 检查是否包含HTML标签
  const htmlTagRegex = /<\/?[a-z][\s\S]*>/i;
  return htmlTagRegex.test(content);
};

// 安全地渲染HTML内容
const sanitizeHtml = (html) => {
  // 创建一个临时元素来解析HTML
  const tempDiv = document.createElement('div');
  tempDiv.innerHTML = html;
  
  // 提取样式
  const styles = Array.from(tempDiv.querySelectorAll('style'))
    .map(style => style.innerHTML)
    .join('\n');
  
  // 提取body内容，如果没有body标签则使用全部内容
  const bodyContent = tempDiv.querySelector('body');
  const content = bodyContent ? bodyContent.innerHTML : html;
  
  return { content, styles };
};

/**
 * 通用文档渲染组件
 * @param {string} apiEndpoint - API 接口地址
 * @param {string} title - 文档标题
 * @param {string} cacheKey - 本地存储缓存键
 * @param {string} emptyMessage - 空内容时的提示消息
 */
const DocumentRenderer = ({ apiEndpoint, title, cacheKey, emptyMessage }) => {
  const { t } = useTranslation();
  const [content, setContent] = useState('');
  const [loading, setLoading] = useState(true);
  const [htmlStyles, setHtmlStyles] = useState('');
  const [processedHtmlContent, setProcessedHtmlContent] = useState('');

  const loadContent = async () => {
    // 先从缓存中获取
    const cachedContent = localStorage.getItem(cacheKey) || '';
    if (cachedContent) {
      setContent(cachedContent);
      processContent(cachedContent);
      setLoading(false);
    }

    try {
      const res = await API.get(apiEndpoint);
      const { success, message, data } = res.data;
      if (success && data) {
        setContent(data);
        processContent(data);
        localStorage.setItem(cacheKey, data);
      } else {
        if (!cachedContent) {
          showError(message || emptyMessage);
          setContent('');
        }
      }
    } catch (error) {
      if (!cachedContent) {
        showError(emptyMessage);
        setContent('');
      }
    } finally {
      setLoading(false);
    }
  };

  const processContent = (rawContent) => {
    if (isHtmlContent(rawContent)) {
      const { content: htmlContent, styles } = sanitizeHtml(rawContent);
      setProcessedHtmlContent(htmlContent);
      setHtmlStyles(styles);
    } else {
      setProcessedHtmlContent('');
      setHtmlStyles('');
    }
  };

  useEffect(() => {
    loadContent();
  }, []);

  // 处理HTML样式注入
  useEffect(() => {
    const styleId = `document-renderer-styles-${cacheKey}`;
    
    if (htmlStyles) {
      let styleEl = document.getElementById(styleId);
      if (!styleEl) {
        styleEl = document.createElement('style');
        styleEl.id = styleId;
        styleEl.type = 'text/css';
        document.head.appendChild(styleEl);
      }
      styleEl.innerHTML = htmlStyles;
    } else {
      const el = document.getElementById(styleId);
      if (el) el.remove();
    }

    return () => {
      const el = document.getElementById(styleId);
      if (el) el.remove();
    };
  }, [htmlStyles, cacheKey]);

  // 显示加载状态
  if (loading) {
    return (
      <div className='relative overflow-hidden min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-teal-50 dark:from-slate-900 dark:via-slate-950 dark:to-cyan-950 flex justify-center items-center'>
        <div className='blur-ball blur-ball-indigo' style={{ top: '-120px', left: '50%' }} />
        <div className='blur-ball blur-ball-teal' style={{ bottom: '-120px', right: '-80px' }} />
        <Spin size='large' />
      </div>
    );
  }

  // 如果没有内容，显示空状态
  if (!content || content.trim() === '') {
    return (
      <div className='relative overflow-hidden min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-teal-50 dark:from-slate-900 dark:via-slate-950 dark:to-cyan-950 flex justify-center items-center pt-20 pb-12 px-4'>
        <div className='blur-ball blur-ball-indigo' style={{ top: '-120px', right: '-80px' }} />
        <div className='blur-ball blur-ball-teal' style={{ bottom: '-120px', left: '-100px' }} />
        <Empty
          title={t('管理员未设置' + title + '内容')}
          image={<IllustrationConstruction style={{ width: 150, height: 150 }} />}
          darkModeImage={<IllustrationConstructionDark style={{ width: 150, height: 150 }} />}
          className='bg-white/70 dark:bg-slate-900/70 backdrop-blur-sm rounded-3xl shadow-xl border border-blue-100/40 dark:border-blue-900/40 p-8'
        />
      </div>
    );
  }

  // 如果是 URL，显示链接卡片
  if (isUrl(content)) {
    return (
      <div className='relative overflow-hidden min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-teal-50 dark:from-slate-900 dark:via-slate-950 dark:to-cyan-950 flex justify-center items-center pt-20 pb-12 px-4'>
        <div className='blur-ball blur-ball-indigo' style={{ top: '-120px', right: '-80px' }} />
        <div className='blur-ball blur-ball-teal' style={{ bottom: '-120px', left: '-100px' }} />
        <Card className='max-w-md w-full bg-white/80 dark:bg-slate-900/70 backdrop-blur-md rounded-3xl shadow-2xl border border-blue-100/40 dark:border-blue-900/40'>
          <div className='text-center p-6'>
            <Title heading={3} className='mb-6 text-transparent bg-clip-text bg-gradient-to-r from-blue-600 to-teal-600 dark:from-blue-400 dark:to-teal-400'>{title}</Title>
            <p className='text-semi-color-text-2 mb-6 leading-relaxed'>
              {t('管理员设置了外部链接，点击下方按钮访问')}
            </p>
            <a
              href={content.trim()}
              target='_blank'
              rel='noopener noreferrer'
              title={content.trim()}
              aria-label={`${t('访问' + title)}: ${content.trim()}`}
              className='inline-flex items-center px-6 py-3 bg-gradient-to-r from-blue-600 to-teal-600 text-white rounded-full hover:from-blue-700 hover:to-teal-700 transition-all duration-300 shadow-lg hover:shadow-xl font-medium'
            >
              <span>{t('访问' + title)}</span>
              <svg className='w-5 h-5 ml-2' fill='none' stroke='currentColor' viewBox='0 0 24 24'>
                <path strokeLinecap='round' strokeLinejoin='round' strokeWidth={2} d='M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14' />
              </svg>
            </a>
          </div>
        </Card>
      </div>
    );
  }

  // 如果是 HTML 内容，直接渲染
  if (isHtmlContent(content)) {
    const { content: htmlContent, styles } = sanitizeHtml(content);
    
    // 设置样式（如果有的话）
    useEffect(() => {
      if (styles && styles !== htmlStyles) {
        setHtmlStyles(styles);
      }
    }, [content, styles, htmlStyles]);
    
    return (
      <div className='relative overflow-hidden min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-teal-50 dark:from-slate-900 dark:via-slate-950 dark:to-cyan-950 pt-[80px] pb-20 px-4'>
        <div className='blur-ball blur-ball-indigo' style={{ top: '-120px', right: '-80px' }} />
        <div className='blur-ball blur-ball-teal' style={{ bottom: '-120px', left: '-100px' }} />
        <div className='relative max-w-5xl mx-auto'>
          <div className='bg-white/80 dark:bg-slate-900/70 backdrop-blur-md rounded-3xl shadow-2xl border border-blue-100/40 dark:border-blue-900/40 p-8 sm:p-10 lg:p-12'>
            <Title heading={2} className='text-center mb-10 text-transparent bg-clip-text bg-gradient-to-r from-blue-600 to-teal-600 dark:from-blue-400 dark:to-teal-400'>{title}</Title>
            <div 
              className='prose prose-lg max-w-none text-slate-700 dark:text-slate-200 prose-headings:text-semi-color-text-0 prose-a:text-blue-600 dark:prose-a:text-blue-400 prose-a:no-underline hover:prose-a:underline'
              dangerouslySetInnerHTML={{ __html: htmlContent }}
            />
          </div>
        </div>
      </div>
    );
  }

  // 其他内容统一使用 Markdown 渲染器
  return (
    <div className='relative overflow-hidden min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-teal-50 dark:from-slate-900 dark:via-slate-950 dark:to-cyan-950 pt-[80px] pb-20 px-4'>
      <div className='blur-ball blur-ball-indigo' style={{ top: '-120px', right: '-80px' }} />
      <div className='blur-ball blur-ball-teal' style={{ bottom: '-120px', left: '-100px' }} />
      <div className='relative max-w-5xl mx-auto'>
        <div className='bg-white/80 dark:bg-slate-900/70 backdrop-blur-md rounded-3xl shadow-2xl border border-blue-100/40 dark:border-blue-900/40 p-8 sm:p-10 lg:p-12'>
          <Title heading={2} className='text-center mb-10 text-transparent bg-clip-text bg-gradient-to-r from-blue-600 to-teal-600 dark:from-blue-400 dark:to-teal-400'>{title}</Title>
          <div className='prose prose-lg max-w-none text-slate-700 dark:text-slate-200 prose-headings:text-semi-color-text-0 prose-a:text-blue-600 dark:prose-a:text-blue-400 prose-a:no-underline hover:prose-a:underline'>
            <MarkdownRenderer content={content} />
          </div>
        </div>
      </div>
    </div>
  );
};

export default DocumentRenderer;