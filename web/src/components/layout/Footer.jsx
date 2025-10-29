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

import React, { useEffect, useState, useMemo, useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { Typography } from '@douyinfe/semi-ui';
import { getFooterHTML, getLogo, getSystemName } from '../../helpers';
import { StatusContext } from '../../context/Status';
import { useBillingFeatures } from '../../hooks/billing/useBillingFeatures';

const FooterBar = () => {
  const { t } = useTranslation();
  const [footer, setFooter] = useState(getFooterHTML());
  const systemName = getSystemName();
  const logo = getLogo();
  const [statusState] = useContext(StatusContext);
  const isDemoSiteMode = statusState?.status?.demo_site_enabled || false;
  const { config } = useBillingFeatures();
  const publicLogsEnabled = config?.public_logs?.enabled;

  const loadFooter = () => {
    let footer_html = localStorage.getItem('footer_html');
    if (footer_html) {
      setFooter(footer_html);
    }
  };

  const currentYear = new Date().getFullYear();

  const customFooter = useMemo(
    () => (
      <footer className='nebula-footer'>
        <div className='nebula-footer-content'>
          {isDemoSiteMode && (
            <div className='flex flex-col md:flex-row gap-8 mb-10'>
              <div className='flex-shrink-0'>
                <img
                  src={logo}
                  alt={systemName}
                  className='w-14 h-14 rounded-xl shadow-md object-contain'
                />
              </div>

              <div className='nebula-footer-grid flex-1'>
                <div>
                  <h3 className='nebula-footer-section-title'>
                    {t('关于我们')}
                  </h3>
                  <div className='nebula-footer-links'>
                    <a
                      href='https://docs.newapi.pro/wiki/project-introduction/'
                      target='_blank'
                      rel='noopener noreferrer'
                      className='nebula-footer-link'
                    >
                      {t('关于项目')}
                    </a>
                    <a
                      href='https://docs.newapi.pro/support/community-interaction/'
                      target='_blank'
                      rel='noopener noreferrer'
                      className='nebula-footer-link'
                    >
                      {t('联系我们')}
                    </a>
                    <a
                      href='https://docs.newapi.pro/wiki/features-introduction/'
                      target='_blank'
                      rel='noopener noreferrer'
                      className='nebula-footer-link'
                    >
                      {t('功能特性')}
                    </a>
                  </div>
                </div>

                <div>
                  <h3 className='nebula-footer-section-title'>
                    {t('文档')}
                  </h3>
                  <div className='nebula-footer-links'>
                    <a
                      href='https://docs.newapi.pro/getting-started/'
                      target='_blank'
                      rel='noopener noreferrer'
                      className='nebula-footer-link'
                    >
                      {t('快速开始')}
                    </a>
                    <a
                      href='https://docs.newapi.pro/installation/'
                      target='_blank'
                      rel='noopener noreferrer'
                      className='nebula-footer-link'
                    >
                      {t('安装指南')}
                    </a>
                    <a
                      href='https://docs.newapi.pro/api/'
                      target='_blank'
                      rel='noopener noreferrer'
                      className='nebula-footer-link'
                    >
                      {t('API 文档')}
                    </a>
                  </div>
                </div>

                <div>
                  <h3 className='nebula-footer-section-title'>
                    {t('相关项目')}
                  </h3>
                  <div className='nebula-footer-links'>
                    <a
                      href='https://github.com/songquanpeng/one-api'
                      target='_blank'
                      rel='noopener noreferrer'
                      className='nebula-footer-link'
                    >
                      One API
                    </a>
                    <a
                      href='https://github.com/novicezk/midjourney-proxy'
                      target='_blank'
                      rel='noopener noreferrer'
                      className='nebula-footer-link'
                    >
                      Midjourney-Proxy
                    </a>
                    <a
                      href='https://github.com/Calcium-Ion/neko-api-key-tool'
                      target='_blank'
                      rel='noopener noreferrer'
                      className='nebula-footer-link'
                    >
                      neko-api-key-tool
                    </a>
                  </div>
                </div>

                <div>
                  <h3 className='nebula-footer-section-title'>
                    {t('友情链接')}
                  </h3>
                  <div className='nebula-footer-links'>
                    <a
                      href='https://github.com/Calcium-Ion/new-api-horizon'
                      target='_blank'
                      rel='noopener noreferrer'
                      className='nebula-footer-link'
                    >
                      new-api-horizon
                    </a>
                    <a
                      href='https://github.com/coaidev/coai'
                      target='_blank'
                      rel='noopener noreferrer'
                      className='nebula-footer-link'
                    >
                      CoAI
                    </a>
                    <a
                      href='https://www.gpt-load.com/'
                      target='_blank'
                      rel='noopener noreferrer'
                      className='nebula-footer-link'
                    >
                      GPT-Load
                    </a>
                  </div>
                </div>
              </div>
            </div>
          )}

          <div className='nebula-footer-bottom'>
            <div className='nebula-footer-copyright'>
              © {currentYear} {systemName}. {t('版权所有')}
              {publicLogsEnabled && (
                <>
                  {' • '}
                  <a href='/public/logs' className='nebula-footer-link'>
                    {t('公开日志')}
                  </a>
                </>
              )}
            </div>

            <div className='nebula-footer-credits'>
              {t('设计与开发由')}{' '}
              <a
                href='https://github.com/QuantumNous/new-api'
                target='_blank'
                rel='noopener noreferrer'
              >
                NebulaGate
              </a>
            </div>
          </div>
        </div>
      </footer>
    ),
    [logo, systemName, t, currentYear, isDemoSiteMode, publicLogsEnabled],
  );

  useEffect(() => {
    loadFooter();
  }, []);

  return (
    <div className='w-full'>
      {footer ? (
        <div className='relative'>
          <div
            className='custom-footer'
            dangerouslySetInnerHTML={{ __html: footer }}
          ></div>
          <div className='absolute bottom-2 right-4 text-xs !text-semi-color-text-2 opacity-70'>
            <span>{t('设计与开发由')} </span>
            <a
              href='https://github.com/QuantumNous/new-api'
              target='_blank'
              rel='noopener noreferrer'
              className='!text-semi-color-primary font-medium'
            >
              NebulaGate
            </a>
          </div>
        </div>
      ) : (
        customFooter
      )}
    </div>
  );
};

export default FooterBar;
