# I18n Translations to Add

Add these translations to your locale files:

## For `zh.json`:
```json
{
  "checkin": {
    "title": "每日签到",
    "subtitle": "每天签到领取额度奖励",
    "consecutive_days": "连续签到",
    "days": "天",
    "today_status": "今日状态",
    "checked": "已签到",
    "not_checked": "未签到",
    "today_reward": "今日奖励",
    "quota": "额度",
    "check_in_now": "立即签到",
    "already_checked_in": "今日已签到",
    "success": "签到成功",
    "failed": "签到失败",
    "awarded": "获得奖励",
    "reward_rules": "奖励规则",
    "consecutive": "连续",
    "view_history": "查看签到历史",
    "history_title": "签到历史",
    "no_history": "暂无签到记录"
  },
  "lottery": {
    "title": "幸运抽奖",
    "subtitle": "试试手气，赢取丰厚奖励",
    "draw_now": "立即抽奖",
    "drawing": "抽奖中...",
    "congratulations": "恭喜中奖",
    "confirm": "确认",
    "failed": "抽奖失败",
    "quota": "额度",
    "plan": "套餐",
    "rules": "抽奖说明",
    "rule_1": "每日签到后可参与抽奖",
    "rule_2": "奖品包括额度奖励和套餐升级",
    "rule_3": "中奖概率根据配置动态调整",
    "rule_4": "所有奖品实时发放",
    "view_history": "查看抽奖记录",
    "history_title": "抽奖记录",
    "no_history": "暂无抽奖记录"
  }
}
```

## For `en.json`:
```json
{
  "checkin": {
    "title": "Daily Check-in",
    "subtitle": "Check in daily to receive quota rewards",
    "consecutive_days": "Consecutive Days",
    "days": "Days",
    "today_status": "Today's Status",
    "checked": "Checked In",
    "not_checked": "Not Checked In",
    "today_reward": "Today's Reward",
    "quota": "Quota",
    "check_in_now": "Check In Now",
    "already_checked_in": "Already checked in today",
    "success": "Check-in successful",
    "failed": "Check-in failed",
    "awarded": "Awarded",
    "reward_rules": "Reward Rules",
    "consecutive": "Consecutive",
    "view_history": "View Check-in History",
    "history_title": "Check-in History",
    "no_history": "No check-in records"
  },
  "lottery": {
    "title": "Lucky Draw",
    "subtitle": "Try your luck and win great rewards",
    "draw_now": "Draw Now",
    "drawing": "Drawing...",
    "congratulations": "Congratulations",
    "confirm": "Confirm",
    "failed": "Draw failed",
    "quota": "Quota",
    "plan": "Plan",
    "rules": "Rules",
    "rule_1": "Participate in the lottery after daily check-in",
    "rule_2": "Prizes include quota rewards and plan upgrades",
    "rule_3": "Winning probability is dynamically adjusted",
    "rule_4": "All prizes are issued in real-time",
    "view_history": "View Lottery Records",
    "history_title": "Lottery Records",
    "no_history": "No lottery records"
  }
}
```

## Instructions:
1. Open `web/src/i18n/locales/zh.json`
2. Add the Chinese translations to the `translation` object
3. Open `web/src/i18n/locales/en.json`
4. Add the English translations to the `translation` object
5. Optionally translate to French (`fr.json`) and Russian (`ru.json`)
