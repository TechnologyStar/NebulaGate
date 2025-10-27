# Menu Configuration Guide

## Adding Check-in and Lottery to Navigation Menu

You need to update your navigation menu component to include links to the new check-in and lottery pages.

### 1. Locate Navigation Component

Find your sidebar/navigation component. It's likely in one of these locations:
- `web/src/components/layout/Sidebar.jsx`
- `web/src/components/layout/Navigation.jsx`
- `web/src/components/common/ui/Layout.jsx`

### 2. Add Menu Items

Add these menu items to your navigation:

```jsx
// For user-accessible features
{
  icon: <IconGift />,
  text: t('menu.checkin', '每日签到'),
  path: '/console/checkin'
},
{
  icon: <IconTickCircle />,
  text: t('menu.lottery', '幸运抽奖'),
  path: '/console/lottery'
}
```

### 3. Import Required Icons

```jsx
import { IconGift, IconTickCircle } from '@douyinfe/semi-icons';
```

### 4. Add to i18n

Add to `zh.json`:
```json
{
  "menu": {
    "checkin": "每日签到",
    "lottery": "幸运抽奖"
  }
}
```

Add to `en.json`:
```json
{
  "menu": {
    "checkin": "Daily Check-in",
    "lottery": "Lucky Draw"
  }
}
```

### 5. Position in Menu

Recommended position: After "Top Up" / "充值" menu item

### Example Complete Menu Structure:

```jsx
const userMenuItems = [
  { path: '/console', icon: <IconHome />, text: t('menu.dashboard') },
  { path: '/console/token', icon: <IconKey />, text: t('menu.token') },
  { path: '/console/topup', icon: <IconCreditCard />, text: t('menu.topup') },
  { path: '/console/checkin', icon: <IconGift />, text: t('menu.checkin') },      // NEW
  { path: '/console/lottery', icon: <IconTickCircle />, text: t('menu.lottery') },  // NEW
  { path: '/console/log', icon: <IconList />, text: t('menu.log') },
  // ... other items
];
```

### 6. Admin Menu Items (for Lottery Management)

For admin users, you may want to add a lottery configuration menu:

```jsx
const adminMenuItems = [
  // ... other admin items
  {
    path: '/console/lottery-config',
    icon: <IconSetting />,
    text: t('menu.lottery_config', '抽奖配置')
  }
];
```

Note: You'll need to create an admin page for lottery configuration management if you want this feature.
