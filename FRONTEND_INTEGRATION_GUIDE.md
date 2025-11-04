# Frontend Integration Guide

## Overview
This guide provides instructions for frontend developers to integrate the new security and UX features into the React UI.

## Module 1: Security Center - IP Tracking

### Location
Add to existing Security Center or create new "IP Tracking" page

### UI Components Needed

1. **Token IP Usage Table**
   - Endpoint: `GET /api/log/ip-usage/token/:id`
   - Display in token detail page or security dashboard
   - Columns: IP Address, First Seen, Last Seen, Request Count
   - Add time range filter (dropdown: 24h, 7d, 30d, All)
   - Optional: IP geolocation display (requires external service)

2. **User IP Usage Table**
   - Endpoint: `GET /api/log/ip-usage/user/:id`
   - Display in user detail page
   - Same columns as token IP usage
   - Add export functionality

3. **IP Search & Filter**
   - Search by IP address
   - Filter by date range
   - Sort by request count or last seen

### Example React Component Structure
```jsx
// components/Security/IPUsageTable.jsx
import { Table, DatePicker, Input } from '@douyinfe/semi-ui';

const IPUsageTable = ({ tokenId, userId }) => {
  const [usages, setUsages] = useState([]);
  const [timeRange, setTimeRange] = useState('24h');
  
  useEffect(() => {
    fetchIPUsage();
  }, [tokenId, userId, timeRange]);
  
  // Fetch data from API
  // Display in table
  // Add pagination
}
```

## Module 2: Group Rate Scheduling

### Location
Add to Admin > Groups page or create new "Rate Scheduling" section

### UI Components Needed

1. **Schedule Rule List**
   - Table showing all schedules for a group
   - Columns: Time Range, Multiplier, Status, Actions
   - Enable/disable toggle
   - Edit and delete buttons

2. **Create/Edit Schedule Form**
   - Group selector (dropdown)
   - Time Start picker (HH:MM format)
   - Time End picker (HH:MM format)
   - Rate multiplier input (number, step 0.1)
   - Enabled checkbox
   - Validation: ensure time format is correct

3. **24-Hour Timeline Visualization**
   - Visual representation of rate changes throughout the day
   - Color-coded segments for different multipliers
   - Interactive: click to edit
   - Show current active rate highlighted

4. **Current Rate Display**
   - Real-time display of current effective multiplier
   - Group selector
   - Auto-refresh every minute
   - Visual indicator (badge or chip)

### Example Timeline Component
```jsx
// components/Admin/RateScheduleTimeline.jsx
import { Timeline, Badge } from '@douyinfe/semi-ui';

const RateScheduleTimeline = ({ groupName }) => {
  const [schedules, setSchedules] = useState([]);
  const [currentRate, setCurrentRate] = useState(1.0);
  
  // Create 24-hour timeline with colored segments
  // Fetch current rate every minute
  // Allow click to edit schedule
}
```

### Design Recommendations
- Use gradient colors for different multiplier levels (low = green, medium = yellow, high = red)
- Show time in 24-hour format
- Support cross-midnight ranges with visual indication
- Add preview of next rate change

## Module 3: IP Protection & Rate Limiting

### Location
Create new Admin > IP Protection page with tabs

### UI Components Needed

1. **Blacklist/Whitelist Management**
   - Two tabs or sections
   - Add IP form with:
     - IP address input (support CIDR)
     - Reason textarea
     - Scope selector (Global, User, Key)
     - Expiration date picker (optional)
   - Table showing current entries
   - Remove button per entry
   - CIDR validation in form

2. **Rate Limit Rules**
   - Table showing all rules
   - Add/Edit form:
     - Rule name input
     - IP address input
     - Max requests input (number)
     - Time window input (seconds)
     - Action selector (Reject, Warn, Ban)
     - Enabled toggle
   - Visual indicator of active rules

3. **Ban Management**
   - Active bans table
   - Columns: IP, Reason, Type, Banned At, Expires At, Actions
   - Manual ban form
   - Unban button
   - Auto-unban countdown timer
   - Filter: show/hide expired bans

4. **Statistics Dashboard**
   - Cards showing:
     - Active bans count
     - Blacklist entries count
     - Whitelist entries count
     - Rate limit rules count
   - Recent violations list
   - Charts (optional):
     - Bans over time
     - Most blocked IPs

### Example Components
```jsx
// components/Admin/IPProtection/Blacklist.jsx
const IPBlacklist = () => {
  const [ips, setIps] = useState([]);
  const [addModalVisible, setAddModalVisible] = useState(false);
  
  const handleAddIP = async (values) => {
    await api.post('/api/admin/ip-protection/blacklist', values);
    fetchIPs();
  };
  
  // Table with add/remove functionality
}

// components/Admin/IPProtection/BanManagement.jsx
const BanManagement = () => {
  const [bans, setBans] = useState([]);
  
  const CountdownTimer = ({ expiresAt }) => {
    // Calculate remaining time
    // Update every second
  };
  
  // Table with unban functionality
}
```

### Design Recommendations
- Use danger color (red) for blacklist and bans
- Use success color (green) for whitelist
- Add confirmation dialog for ban/unban actions
- Show toast notifications for all operations
- Highlight expired entries with gray color

## Module 4: Daily Check-in System

### Location
- Add "Check-in" button in top navigation or sidebar
- Create dedicated check-in page
- Show notification badge when not checked in

### UI Components Needed

1. **Check-in Button**
   - Large, prominent button
   - Shows "Check In" or "Checked In ✓" based on status
   - Display consecutive days count
   - Animation on successful check-in
   - Confetti or particle effect (use react-confetti)

2. **Reward Display**
   - Show today's reward amount
   - Display next day's potential reward
   - Progress bar for current tier
   - Highlight bonus days (e.g., day 30+)

3. **Monthly Calendar View**
   - Calendar grid showing current month
   - Mark checked-in days with checkmark
   - Highlight today
   - Show consecutive streak visually
   - Allow navigation to previous months (read-only)

4. **Reward Tiers Table**
   - Display all reward tiers
   - Highlight current progress
   - Show locked/unlocked tiers

5. **Statistics Panel**
   - Total check-ins this month
   - Longest streak
   - Total rewards earned
   - Current consecutive days

6. **Check-in History**
   - Paginated table
   - Columns: Date, Reward, Consecutive Days
   - Export functionality

### Example Components
```jsx
// components/CheckIn/CheckInButton.jsx
import { Button, Toast } from '@douyinfe/semi-ui';
import Confetti from 'react-confetti';

const CheckInButton = () => {
  const [hasCheckedIn, setHasCheckedIn] = useState(false);
  const [showConfetti, setShowConfetti] = useState(false);
  const [consecutiveDays, setConsecutiveDays] = useState(0);
  
  const handleCheckIn = async () => {
    const result = await api.post('/api/user/checkin');
    if (result.success) {
      setShowConfetti(true);
      Toast.success(`签到成功！获得 ${result.data.quota_awarded} 额度`);
      setTimeout(() => setShowConfetti(false), 5000);
    }
  };
  
  return (
    <>
      {showConfetti && <Confetti />}
      <Button 
        size="large" 
        theme="solid" 
        disabled={hasCheckedIn}
        onClick={handleCheckIn}
      >
        {hasCheckedIn ? '已签到 ✓' : '每日签到'}
      </Button>
      <div>连续签到 {consecutiveDays} 天</div>
    </>
  );
};

// components/CheckIn/Calendar.jsx
import { Calendar } from '@douyinfe/semi-ui';

const CheckInCalendar = () => {
  const [records, setRecords] = useState([]);
  
  const dateRender = (date) => {
    const dateStr = date.format('YYYY-MM-DD');
    const hasCheckedIn = records.some(r => r.check_in_date === dateStr);
    
    return hasCheckedIn ? (
      <div className="checked-in">
        <CheckCircleOutlined />
      </div>
    ) : null;
  };
  
  return <Calendar dateFullCellRender={dateRender} />;
};
```

### Design Recommendations
- Use glassmorphism design for cards:
  ```css
  .glass-card {
    background: rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(10px);
    border-radius: 20px;
    border: 1px solid rgba(255, 255, 255, 0.18);
    box-shadow: 0 8px 32px 0 rgba(31, 38, 135, 0.15);
  }
  ```
- Animate check-in success:
  - Scale up button
  - Show +quota floating number
  - Trigger confetti
  - Play success sound (optional)
- Show daily reminder notification if not checked in
- Add streak protection mechanism (allow 1 skip every 7 days)

### Animation Libraries
- `react-confetti` for celebration
- `framer-motion` for smooth animations
- `lottie-react` for Lottie animations
- CSS transitions for hover effects

## Module 5: Playground UI Enhancement

### Location
Existing Playground page

### UI Enhancements Needed

1. **Glassmorphism Design**
   - Apply to all cards and panels
   - Use large border radius (16-24px)
   - Semi-transparent backgrounds with backdrop blur
   - Subtle borders and shadows

2. **Message Bubbles**
   - User messages: right-aligned, blue gradient
   - AI messages: left-aligned, purple/pink gradient
   - Rounded corners
   - Smooth fade-in animation
   - Typing indicator for AI responses

3. **Input Area**
   - Multi-line textarea with auto-height
   - Glassmorphism background
   - Glow effect on focus
   - Character counter
   - Send button with icon
   - Shift+Enter for newline, Enter to send

4. **Sidebar Layout**
   - Left sidebar: Model selector, parameters
   - Main area: Chat window
   - Right sidebar: History, statistics
   - Collapsible sidebars (mobile responsive)

5. **IP Statistics Panel (Admin Only)**
   - Endpoint: `GET /api/playground/ip-stats?hours=24`
   - Show in admin dashboard or settings
   - Table with:
     - IP address
     - User count (highlight if >5)
     - Usernames list (expandable)
     - Last active time
     - Total requests
   - Time range selector
   - Export functionality

### Example Glassmorphism Component
```jsx
// components/Playground/GlassCard.jsx
import styled from 'styled-components';

const GlassCard = styled.div`
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  border-radius: 20px;
  border: 1px solid rgba(255, 255, 255, 0.18);
  box-shadow: 0 8px 32px 0 rgba(31, 38, 135, 0.15);
  padding: 20px;
  transition: all 0.3s ease;
  
  &:hover {
    box-shadow: 0 12px 48px 0 rgba(31, 38, 135, 0.25);
    transform: translateY(-2px);
  }
`;

// components/Playground/MessageBubble.jsx
const MessageBubble = ({ message, isUser }) => {
  return (
    <BubbleWrapper isUser={isUser}>
      <Bubble isUser={isUser}>
        <Markdown>{message.content}</Markdown>
      </Bubble>
      <Timestamp>{message.timestamp}</Timestamp>
    </BubbleWrapper>
  );
};

const BubbleWrapper = styled.div`
  display: flex;
  flex-direction: column;
  align-items: ${props => props.isUser ? 'flex-end' : 'flex-start'};
  margin-bottom: 16px;
  animation: fadeIn 0.3s ease;
  
  @keyframes fadeIn {
    from { opacity: 0; transform: translateY(10px); }
    to { opacity: 1; transform: translateY(0); }
  }
`;

const Bubble = styled.div`
  max-width: 70%;
  padding: 12px 16px;
  border-radius: 18px;
  background: ${props => props.isUser 
    ? 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)'
    : 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)'
  };
  color: white;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
`;

// components/Playground/IPStatsPanel.jsx (Admin Only)
const IPStatsPanel = () => {
  const [stats, setStats] = useState(null);
  const [timeRange, setTimeRange] = useState(24);
  
  useEffect(() => {
    fetchIPStats(timeRange);
  }, [timeRange]);
  
  const fetchIPStats = async (hours) => {
    const result = await api.get(`/api/playground/ip-stats?hours=${hours}`);
    setStats(result.data);
  };
  
  return (
    <GlassCard>
      <h3>IP User Statistics</h3>
      <Select value={timeRange} onChange={setTimeRange}>
        <Select.Option value={24}>Last 24 hours</Select.Option>
        <Select.Option value={168}>Last 7 days</Select.Option>
        <Select.Option value={720}>Last 30 days</Select.Option>
      </Select>
      
      <StatCards>
        <StatCard>
          <h4>Total IPs</h4>
          <p>{stats?.total_ips}</p>
        </StatCard>
        <StatCard warning>
          <h4>Suspicious IPs</h4>
          <p>{stats?.suspicious_ips}</p>
        </StatCard>
      </StatCards>
      
      <Table
        columns={[
          { title: 'IP', dataIndex: 'ip' },
          { title: 'Users', dataIndex: 'user_count', render: (count) => (
            <Badge count={count} type={count > 5 ? 'danger' : 'primary'} />
          )},
          { title: 'Usernames', dataIndex: 'usernames', render: (users) => (
            <Collapsible items={users} />
          )},
          { title: 'Last Active', dataIndex: 'last_active_at' },
          { title: 'Requests', dataIndex: 'total_requests' }
        ]}
        dataSource={stats?.ip_stats}
      />
    </GlassCard>
  );
};
```

### Markdown Rendering
Use `react-markdown` with syntax highlighting:
```jsx
import ReactMarkdown from 'react-markdown';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';

const MarkdownMessage = ({ content }) => (
  <ReactMarkdown
    components={{
      code({node, inline, className, children, ...props}) {
        const match = /language-(\w+)/.exec(className || '');
        return !inline && match ? (
          <SyntaxHighlighter
            style={vscDarkPlus}
            language={match[1]}
            PreTag="div"
            {...props}
          >
            {String(children).replace(/\n$/, '')}
          </SyntaxHighlighter>
        ) : (
          <code className={className} {...props}>
            {children}
          </code>
        );
      }
    }}
  >
    {content}
  </ReactMarkdown>
);
```

## General UI/UX Guidelines

### Color Scheme
- Primary: Purple/Blue gradient (#667eea to #764ba2)
- Secondary: Pink/Red gradient (#f093fb to #f5576c)
- Success: Green (#52c41a)
- Warning: Orange (#faad14)
- Danger: Red (#f5222d)

### Typography
- Use system fonts or Inter/Roboto
- Headings: Bold, large
- Body: Regular, readable size (14-16px)
- Monospace for code and IPs

### Spacing
- Use consistent spacing (4px, 8px, 16px, 24px, 32px)
- Generous padding in cards
- Adequate margin between sections

### Responsive Design
- Mobile: Stack sidebars, collapse tables
- Tablet: Side-by-side layout with collapsible panels
- Desktop: Full three-column layout

### Animations
- Fade in: 300ms ease
- Slide: 200ms ease-out
- Scale: 150ms ease
- Use `framer-motion` for complex animations

### Loading States
- Skeleton screens for tables
- Spinner for buttons
- Shimmer effect for cards
- Progressive loading for lists

### Error Handling
- Toast notifications for errors
- Inline validation messages
- Retry buttons for failed requests
- Graceful degradation

## API Integration Example

```jsx
// utils/api.js
import axios from 'axios';

const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json'
  }
});

// Add auth token to all requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle errors globally
api.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response?.status === 401) {
      // Redirect to login
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export default api;

// Example usage in component
const CheckInPage = () => {
  const handleCheckIn = async () => {
    try {
      const result = await api.post('/user/checkin');
      Toast.success(result.message);
    } catch (error) {
      Toast.error(error.response?.data?.message || 'Check-in failed');
    }
  };
};
```

## Testing Checklist

- [ ] All API endpoints return expected data
- [ ] Loading states display correctly
- [ ] Error states handled gracefully
- [ ] Forms validate input
- [ ] Tables paginate properly
- [ ] Animations perform smoothly (60fps)
- [ ] Responsive on mobile/tablet/desktop
- [ ] Accessibility: keyboard navigation, screen readers
- [ ] Dark mode support (if applicable)
- [ ] i18n translations complete

## Additional Libraries Recommended

- `@douyinfe/semi-ui` - Already in use
- `framer-motion` - Animations
- `react-confetti` - Check-in celebration
- `react-markdown` - Markdown rendering
- `react-syntax-highlighter` - Code highlighting
- `recharts` or `echarts-for-react` - Charts
- `date-fns` - Date formatting
- `lodash` - Utility functions
- `react-virtual` - Virtual scrolling for long lists

## Performance Optimization

- Use React.memo for expensive components
- Implement virtual scrolling for long lists
- Lazy load images and heavy components
- Debounce search inputs
- Use SWR or React Query for caching
- Code split by route
- Optimize bundle size

## Accessibility

- Use semantic HTML
- Add ARIA labels
- Ensure keyboard navigation
- Maintain color contrast ratios
- Provide alternative text for images
- Support screen readers

## Browser Compatibility

- Chrome/Edge: Latest 2 versions
- Firefox: Latest 2 versions
- Safari: Latest 2 versions
- Mobile Safari: iOS 12+
- Chrome Mobile: Android 8+
