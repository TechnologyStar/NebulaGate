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

export default {
  content: ['./index.html', './src/**/*.{js,jsx,ts,tsx}'],
  darkMode: 'class',
  theme: {
    colors: {
      /* ====== Semi UI Color Mappings (Preserved for Component Compatibility) ====== */
      'semi-color-white': 'var(--semi-color-white)',
      'semi-color-black': 'var(--semi-color-black)',
      'semi-color-primary': 'var(--semi-color-primary)',
      'semi-color-primary-hover': 'var(--semi-color-primary-hover)',
      'semi-color-primary-active': 'var(--semi-color-primary-active)',
      'semi-color-primary-disabled': 'var(--semi-color-primary-disabled)',
      'semi-color-primary-light-default':
        'var(--semi-color-primary-light-default)',
      'semi-color-primary-light-hover': 'var(--semi-color-primary-light-hover)',
      'semi-color-primary-light-active':
        'var(--semi-color-primary-light-active)',
      'semi-color-secondary': 'var(--semi-color-secondary)',
      'semi-color-secondary-hover': 'var(--semi-color-secondary-hover)',
      'semi-color-secondary-active': 'var(--semi-color-secondary-active)',
      'semi-color-secondary-disabled': 'var(--semi-color-secondary-disabled)',
      'semi-color-secondary-light-default':
        'var(--semi-color-secondary-light-default)',
      'semi-color-secondary-light-hover':
        'var(--semi-color-secondary-light-hover)',
      'semi-color-secondary-light-active':
        'var(--semi-color-secondary-light-active)',
      'semi-color-tertiary': 'var(--semi-color-tertiary)',
      'semi-color-tertiary-hover': 'var(--semi-color-tertiary-hover)',
      'semi-color-tertiary-active': 'var(--semi-color-tertiary-active)',
      'semi-color-tertiary-light-default':
        'var(--semi-color-tertiary-light-default)',
      'semi-color-tertiary-light-hover':
        'var(--semi-color-tertiary-light-hover)',
      'semi-color-tertiary-light-active':
        'var(--semi-color-tertiary-light-active)',
      'semi-color-default': 'var(--semi-color-default)',
      'semi-color-default-hover': 'var(--semi-color-default-hover)',
      'semi-color-default-active': 'var(--semi-color-default-active)',
      'semi-color-info': 'var(--semi-color-info)',
      'semi-color-info-hover': 'var(--semi-color-info-hover)',
      'semi-color-info-active': 'var(--semi-color-info-active)',
      'semi-color-info-disabled': 'var(--semi-color-info-disabled)',
      'semi-color-info-light-default': 'var(--semi-color-info-light-default)',
      'semi-color-info-light-hover': 'var(--semi-color-info-light-hover)',
      'semi-color-info-light-active': 'var(--semi-color-info-light-active)',
      'semi-color-success': 'var(--semi-color-success)',
      'semi-color-success-hover': 'var(--semi-color-success-hover)',
      'semi-color-success-active': 'var(--semi-color-success-active)',
      'semi-color-success-disabled': 'var(--semi-color-success-disabled)',
      'semi-color-success-light-default':
        'var(--semi-color-success-light-default)',
      'semi-color-success-light-hover': 'var(--semi-color-success-light-hover)',
      'semi-color-success-light-active':
        'var(--semi-color-success-light-active)',
      'semi-color-danger': 'var(--semi-color-danger)',
      'semi-color-danger-hover': 'var(--semi-color-danger-hover)',
      'semi-color-danger-active': 'var(--semi-color-danger-active)',
      'semi-color-danger-light-default':
        'var(--semi-color-danger-light-default)',
      'semi-color-danger-light-hover': 'var(--semi-color-danger-light-hover)',
      'semi-color-danger-light-active': 'var(--semi-color-danger-light-active)',
      'semi-color-warning': 'var(--semi-color-warning)',
      'semi-color-warning-hover': 'var(--semi-color-warning-hover)',
      'semi-color-warning-active': 'var(--semi-color-warning-active)',
      'semi-color-warning-light-default':
        'var(--semi-color-warning-light-default)',
      'semi-color-warning-light-hover': 'var(--semi-color-warning-light-hover)',
      'semi-color-warning-light-active':
        'var(--semi-color-warning-light-active)',
      'semi-color-focus-border': 'var(--semi-color-focus-border)',
      'semi-color-disabled-text': 'var(--semi-color-disabled-text)',
      'semi-color-disabled-border': 'var(--semi-color-disabled-border)',
      'semi-color-disabled-bg': 'var(--semi-color-disabled-bg)',
      'semi-color-disabled-fill': 'var(--semi-color-disabled-fill)',
      'semi-color-shadow': 'var(--semi-color-shadow)',
      'semi-color-link': 'var(--semi-color-link)',
      'semi-color-link-hover': 'var(--semi-color-link-hover)',
      'semi-color-link-active': 'var(--semi-color-link-active)',
      'semi-color-link-visited': 'var(--semi-color-link-visited)',
      'semi-color-border': 'var(--semi-color-border)',
      'semi-color-nav-bg': 'var(--semi-color-nav-bg)',
      'semi-color-overlay-bg': 'var(--semi-color-overlay-bg)',
      'semi-color-fill-0': 'var(--semi-color-fill-0)',
      'semi-color-fill-1': 'var(--semi-color-fill-1)',
      'semi-color-fill-2': 'var(--semi-color-fill-2)',
      'semi-color-bg-0': 'var(--semi-color-bg-0)',
      'semi-color-bg-1': 'var(--semi-color-bg-1)',
      'semi-color-bg-2': 'var(--semi-color-bg-2)',
      'semi-color-bg-3': 'var(--semi-color-bg-3)',
      'semi-color-bg-4': 'var(--semi-color-bg-4)',
      'semi-color-text-0': 'var(--semi-color-text-0)',
      'semi-color-text-1': 'var(--semi-color-text-1)',
      'semi-color-text-2': 'var(--semi-color-text-2)',
      'semi-color-text-3': 'var(--semi-color-text-3)',
      'semi-color-highlight-bg': 'var(--semi-color-highlight-bg)',
      'semi-color-highlight': 'var(--semi-color-highlight)',
      'semi-color-data-0': 'var(--semi-color-data-0)',
      'semi-color-data-1': 'var(--semi-color-data-1)',
      'semi-color-data-2': 'var(--semi-color-data-2)',
      'semi-color-data-3': 'var(--semi-color-data-3)',
      'semi-color-data-4': 'var(--semi-color-data-4)',
      'semi-color-data-5': 'var(--semi-color-data-5)',
      'semi-color-data-6': 'var(--semi-color-data-6)',
      'semi-color-data-7': 'var(--semi-color-data-7)',
      'semi-color-data-8': 'var(--semi-color-data-8)',
      'semi-color-data-9': 'var(--semi-color-data-9)',
      'semi-color-data-10': 'var(--semi-color-data-10)',
      'semi-color-data-11': 'var(--semi-color-data-11)',
      'semi-color-data-12': 'var(--semi-color-data-12)',
      'semi-color-data-13': 'var(--semi-color-data-13)',
      'semi-color-data-14': 'var(--semi-color-data-14)',
      'semi-color-data-15': 'var(--semi-color-data-15)',
      'semi-color-data-16': 'var(--semi-color-data-16)',
      'semi-color-data-17': 'var(--semi-color-data-17)',
      'semi-color-data-18': 'var(--semi-color-data-18)',
      'semi-color-data-19': 'var(--semi-color-data-19)',

      /* ====== NebulaGate Design Token Mappings ====== */
      
      /* Brand Colors */
      'nebula-brand-primary': 'var(--nebula-brand-primary)',
      'nebula-brand-primary-hover': 'var(--nebula-brand-primary-hover)',
      'nebula-brand-primary-active': 'var(--nebula-brand-primary-active)',
      'nebula-brand-secondary': 'var(--nebula-brand-secondary)',
      
      /* Aurora Accents */
      'nebula-aurora-teal': 'var(--nebula-aurora-teal)',
      'nebula-aurora-purple': 'var(--nebula-aurora-purple)',
      'nebula-aurora-cyan': 'var(--nebula-aurora-cyan)',
      'nebula-aurora-emerald': 'var(--nebula-aurora-emerald)',
      
      /* Gray Scale */
      'nebula-gray': {
        50: 'var(--nebula-gray-50)',
        100: 'var(--nebula-gray-100)',
        200: 'var(--nebula-gray-200)',
        300: 'var(--nebula-gray-300)',
        400: 'var(--nebula-gray-400)',
        500: 'var(--nebula-gray-500)',
        600: 'var(--nebula-gray-600)',
        700: 'var(--nebula-gray-700)',
        800: 'var(--nebula-gray-800)',
        900: 'var(--nebula-gray-900)',
      },
      
      /* Semantic Background Colors */
      'nebula-bg-base': 'var(--nebula-bg-base)',
      'nebula-bg-surface': 'var(--nebula-bg-surface)',
      'nebula-bg-elevated': 'var(--nebula-bg-elevated)',
      'nebula-bg-subtle': 'var(--nebula-bg-subtle)',
      
      /* Surface Colors */
      'nebula-surface-primary': 'var(--nebula-surface-primary)',
      'nebula-surface-secondary': 'var(--nebula-surface-secondary)',
      'nebula-surface-tertiary': 'var(--nebula-surface-tertiary)',
      
      /* Border Colors */
      'nebula-border-subtle': 'var(--nebula-border-subtle)',
      'nebula-border-default': 'var(--nebula-border-default)',
      'nebula-border-strong': 'var(--nebula-border-strong)',
      'nebula-border-accent': 'var(--nebula-border-accent)',
      'nebula-border-focus': 'var(--nebula-border-focus)',
      
      /* Text Colors */
      'nebula-text-primary': 'var(--nebula-text-primary)',
      'nebula-text-secondary': 'var(--nebula-text-secondary)',
      'nebula-text-tertiary': 'var(--nebula-text-tertiary)',
      'nebula-text-quaternary': 'var(--nebula-text-quaternary)',
      'nebula-text-disabled': 'var(--nebula-text-disabled)',
      'nebula-text-inverse': 'var(--nebula-text-inverse)',
      'nebula-text-brand': 'var(--nebula-text-brand)',
      'nebula-text-link': 'var(--nebula-text-link)',
      
      /* Status Colors */
      'nebula-success': {
        50: 'var(--nebula-success-50)',
        100: 'var(--nebula-success-100)',
        500: 'var(--nebula-success-500)',
        600: 'var(--nebula-success-600)',
        700: 'var(--nebula-success-700)',
        text: 'var(--nebula-success-text)',
        bg: 'var(--nebula-success-bg)',
        border: 'var(--nebula-success-border)',
      },
      'nebula-warning': {
        50: 'var(--nebula-warning-50)',
        100: 'var(--nebula-warning-100)',
        500: 'var(--nebula-warning-500)',
        600: 'var(--nebula-warning-600)',
        700: 'var(--nebula-warning-700)',
        text: 'var(--nebula-warning-text)',
        bg: 'var(--nebula-warning-bg)',
        border: 'var(--nebula-warning-border)',
      },
      'nebula-error': {
        50: 'var(--nebula-error-50)',
        100: 'var(--nebula-error-100)',
        500: 'var(--nebula-error-500)',
        600: 'var(--nebula-error-600)',
        700: 'var(--nebula-error-700)',
        text: 'var(--nebula-error-text)',
        bg: 'var(--nebula-error-bg)',
        border: 'var(--nebula-error-border)',
      },
      'nebula-info': {
        50: 'var(--nebula-info-50)',
        100: 'var(--nebula-info-100)',
        500: 'var(--nebula-info-500)',
        600: 'var(--nebula-info-600)',
        700: 'var(--nebula-info-700)',
        text: 'var(--nebula-info-text)',
        bg: 'var(--nebula-info-bg)',
        border: 'var(--nebula-info-border)',
      },
      
      /* Interactive States */
      'nebula-interactive-default': 'var(--nebula-interactive-default)',
      'nebula-interactive-hover': 'var(--nebula-interactive-hover)',
      'nebula-interactive-active': 'var(--nebula-interactive-active)',
      'nebula-interactive-disabled': 'var(--nebula-interactive-disabled)',
    },
    extend: {
      /* ====== Font Family ====== */
      fontFamily: {
        sans: [
          'Inter Variable',
          'Inter',
          '-apple-system',
          'BlinkMacSystemFont',
          'Segoe UI',
          'Helvetica Neue',
          'Arial',
          'Noto Sans',
          'sans-serif',
          'Microsoft YaHei',
          'Apple Color Emoji',
          'Segoe UI Emoji',
          'Segoe UI Symbol',
          'Noto Color Emoji',
        ],
      },
      
      /* ====== Border Radius ====== */
      borderRadius: {
        'semi-border-radius-extra-small':
          'var(--semi-border-radius-extra-small)',
        'semi-border-radius-small': 'var(--semi-border-radius-small)',
        'semi-border-radius-medium': 'var(--semi-border-radius-medium)',
        'semi-border-radius-large': 'var(--semi-border-radius-large)',
        'semi-border-radius-circle': 'var(--semi-border-radius-circle)',
        'semi-border-radius-full': 'var(--semi-border-radius-full)',
        'nebula-radius-sm': 'var(--nebula-radius-sm)',
        'nebula-radius-md': 'var(--nebula-radius-md)',
        'nebula-radius-lg': 'var(--nebula-radius-lg)',
        'nebula-radius-xl': 'var(--nebula-radius-xl)',
      },
      
      /* ====== Spacing ====== */
      spacing: {
        'nebula-xs': 'var(--nebula-spacing-xs)',
        'nebula-sm': 'var(--nebula-spacing-sm)',
        'nebula-md': 'var(--nebula-spacing-md)',
        'nebula-lg': 'var(--nebula-spacing-lg)',
        'nebula-xl': 'var(--nebula-spacing-xl)',
        'nebula-2xl': 'var(--nebula-spacing-2xl)',
      },
      
      /* ====== Typography ====== */
      fontSize: {
        'nebula-xs': 'var(--nebula-text-xs)',
        'nebula-sm': 'var(--nebula-text-sm)',
        'nebula-base': 'var(--nebula-text-base)',
        'nebula-lg': 'var(--nebula-text-lg)',
        'nebula-xl': 'var(--nebula-text-xl)',
        'nebula-2xl': 'var(--nebula-text-2xl)',
        'nebula-3xl': 'var(--nebula-text-3xl)',
      },
      
      /* ====== Shadows/Elevation ====== */
      boxShadow: {
        'nebula-sm': 'var(--nebula-shadow-sm)',
        'nebula-md': 'var(--nebula-shadow-md)',
        'nebula-lg': 'var(--nebula-shadow-lg)',
        'nebula-xl': 'var(--nebula-shadow-xl)',
        'nebula-dark-sm': 'var(--nebula-shadow-dark-sm)',
        'nebula-dark-md': 'var(--nebula-shadow-dark-md)',
        'nebula-dark-lg': 'var(--nebula-shadow-dark-lg)',
        'nebula-dark-xl': 'var(--nebula-shadow-dark-xl)',
      },
    },
  },
  plugins: [],
};
