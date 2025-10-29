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
import { API, showError } from '../../helpers';
import { marked } from 'marked';
import { Empty } from '@douyinfe/semi-ui';
import {
  IllustrationConstruction,
  IllustrationConstructionDark,
} from '@douyinfe/semi-illustrations';
import { useTranslation } from 'react-i18next';

const About = () => {
  const { t } = useTranslation();
  const [about, setAbout] = useState('');
  const [aboutLoaded, setAboutLoaded] = useState(false);
  const currentYear = new Date().getFullYear();

  const displayAbout = async () => {
    setAbout(localStorage.getItem('about') || '');
    const res = await API.get('/api/about');
    const { success, message, data } = res.data;
    if (success) {
      let aboutContent = data;
      if (!data.startsWith('https://')) {
        aboutContent = marked.parse(data);
      }
      setAbout(aboutContent);
      localStorage.setItem('about', aboutContent);
    } else {
      showError(message);
      setAbout(t('加载关于内容失败...'));
    }
    setAboutLoaded(true);
  };

  useEffect(() => {
    displayAbout().then();
  }, []);

  const emptyStyle = {
    padding: '24px',
  };

  const customDescription = (
    <div style={{ textAlign: 'center' }} className='space-y-4'>
      <p className='text-lg text-semi-color-text-1'>{t('可在设置页面设置关于内容，支持 HTML & Markdown')}</p>
      <div className='my-6'>
        <p className='text-semi-color-text-2 mb-2'>{t('NebulaGate 项目仓库地址：')}</p>
        <a
          href='https://github.com/QuantumNous/new-api'
          target='_blank'
          rel='noopener noreferrer'
          className='!text-semi-color-primary hover:underline font-medium'
        >
          https://github.com/QuantumNous/new-api
        </a>
      </div>
      <div className='bg-gradient-to-r from-blue-50 to-teal-50 dark:from-blue-950 dark:to-teal-950 rounded-xl p-6 my-6'>
        <p className='text-semi-color-text-1 leading-relaxed'>
          <a
            href='https://github.com/QuantumNous/new-api'
            target='_blank'
            rel='noopener noreferrer'
            className='!text-semi-color-primary hover:underline font-semibold'
          >
            NebulaGate
          </a>{' '}
          {t('© {{currentYear}}', { currentYear })}{' '}
          <a
            href='https://github.com/QuantumNous'
            target='_blank'
            rel='noopener noreferrer'
            className='!text-semi-color-primary hover:underline font-medium'
          >
            QuantumNous
          </a>
        </p>
        <p className='text-sm text-semi-color-text-2 mt-3'>
          {t('| 基于')}{' '}
          <a
            href='https://github.com/songquanpeng/one-api/releases/tag/v0.5.4'
            target='_blank'
            rel='noopener noreferrer'
            className='!text-semi-color-primary hover:underline'
          >
            One API v0.5.4
          </a>{' '}
          © 2023{' '}
          <a
            href='https://github.com/songquanpeng'
            target='_blank'
            rel='noopener noreferrer'
            className='!text-semi-color-primary hover:underline'
          >
            JustSong
          </a>
        </p>
      </div>
      <p className='text-sm text-semi-color-text-2 leading-relaxed'>
        {t('本项目根据')}
        <a
          href='https://github.com/songquanpeng/one-api/blob/v0.5.4/LICENSE'
          target='_blank'
          rel='noopener noreferrer'
          className='!text-semi-color-primary hover:underline mx-1'
        >
          {t('MIT许可证')}
        </a>
        {t('授权，需在遵守')}
        <a
          href='https://www.gnu.org/licenses/agpl-3.0.html'
          target='_blank'
          rel='noopener noreferrer'
          className='!text-semi-color-primary hover:underline mx-1'
        >
          {t('AGPL v3.0协议')}
        </a>
        {t('的前提下使用。')}
      </p>
    </div>
  );

  return (
    <div className='relative overflow-hidden min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-teal-50 dark:from-slate-900 dark:via-slate-950 dark:to-cyan-950 pt-[80px] pb-20 px-4'>
      <div className='blur-ball blur-ball-indigo' style={{ top: '-120px', right: '-80px' }} />
      <div className='blur-ball blur-ball-teal' style={{ bottom: '-120px', left: '-100px' }} />
      <div className='relative max-w-5xl mx-auto'>
        {aboutLoaded && about === '' ? (
          <div className='flex justify-center items-center min-h-[60vh]'>
            <Empty
              image={
                <IllustrationConstruction style={{ width: 150, height: 150 }} />
              }
              darkModeImage={
                <IllustrationConstructionDark
                  style={{ width: 150, height: 150 }}
                />
              }
              description={t('管理员暂时未设置任何关于内容')}
              style={emptyStyle}
              className='bg-white/70 dark:bg-slate-900/70 backdrop-blur-sm rounded-3xl shadow-xl border border-blue-100/40 dark:border-blue-900/40 px-6 py-8'
            >
              {customDescription}
            </Empty>
          </div>
        ) : (
          <div className='bg-white/80 dark:bg-slate-900/70 backdrop-blur-md rounded-3xl shadow-2xl border border-blue-100/40 dark:border-blue-900/40 p-8 sm:p-10 lg:p-12'>
            {about.startsWith('https://') ? (
              <div className='aspect-[3/4] sm:aspect-[16/9] w-full rounded-2xl overflow-hidden border border-blue-100/40 dark:border-blue-900/40 shadow-inner'>
                <iframe
                  src={about}
                  title='NebulaGate About'
                  className='w-full h-full border-0'
                />
              </div>
            ) : (
              <div
                className='prose prose-lg max-w-none text-slate-700 dark:text-slate-200'
                dangerouslySetInnerHTML={{ __html: about }}
              ></div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

export default About;
