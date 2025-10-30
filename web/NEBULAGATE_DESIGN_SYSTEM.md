# NebulaGate Design System

## Overview

This document outlines the NebulaGate design system implemented across the application. The design system provides a consistent, premium enterprise aesthetic with careful attention to spacing, typography, colors, and interactive states.

## Design Tokens

### Spacing Scale

```css
--nebula-spacing-xs: 4px;
--nebula-spacing-sm: 8px;
--nebula-spacing-md: 16px;
--nebula-spacing-lg: 24px;
--nebula-spacing-xl: 32px;
--nebula-spacing-2xl: 48px;
```

**Tailwind utilities**: `nebula-xs`, `nebula-sm`, `nebula-md`, `nebula-lg`, `nebula-xl`, `nebula-2xl`

### Border Radius

```css
--nebula-radius-sm: 12px;
--nebula-radius-md: 16px;
--nebula-radius-lg: 20px;
--nebula-radius-xl: 24px;
```

**Tailwind utilities**: `rounded-nebula-radius-sm`, `rounded-nebula-radius-md`, etc.

### Typography Scale

```css
--nebula-text-xs: 12px;
--nebula-text-sm: 14px;
--nebula-text-base: 16px;
--nebula-text-lg: 18px;
--nebula-text-xl: 20px;
--nebula-text-2xl: 24px;
--nebula-text-3xl: 30px;
```

**Classes**: `nebula-heading-1`, `nebula-heading-2`, `nebula-heading-3`, `nebula-body`, `nebula-body-small`, `nebula-caption`

### Elevation & Shadows

```css
--nebula-shadow-sm: 0 2px 8px rgba(15, 23, 42, 0.06);
--nebula-shadow-md: 0 8px 24px rgba(15, 23, 42, 0.08);
--nebula-shadow-lg: 0 16px 40px rgba(15, 23, 42, 0.12);
--nebula-shadow-xl: 0 24px 60px rgba(15, 23, 42, 0.16);
```

Dark mode equivalents: `--nebula-shadow-dark-sm`, etc.

**Tailwind utilities**: `shadow-nebula-sm`, `shadow-nebula-md`, etc.

## Components

### Cards

#### Standard Card
```html
<div class="nebula-card">
  <div class="nebula-card-body">
    <!-- Content -->
  </div>
</div>
```

**Variants**:
- `.nebula-card-compact` - Reduced padding
- `.nebula-card-premium` - Enhanced gradient background
- `.admin-card` - For admin console tables and forms

**Features**:
- Smooth hover states with elevation change
- Consistent border treatments
- Light/dark mode support
- Gradient backgrounds

### Buttons

**Variants**:
- `.nebula-btn-primary` - Primary action with gradient
- `.nebula-btn-secondary` - Secondary action with border
- `.nebula-btn-tertiary` - Borderless tertiary action
- `.nebula-btn-danger` - Destructive action
- `.nebula-btn-success` - Success action

**Sizes**:
- `.nebula-btn-sm` - 32px height
- `.nebula-btn-md` - 40px height (default)
- `.nebula-btn-lg` - 48px height

**Icon Buttons**:
- `.nebula-btn-icon` - 40x40px
- `.nebula-btn-icon-sm` - 32x32px
- `.nebula-btn-icon-lg` - 48x48px

**States**:
- Hover: Slight elevation with transform
- Active: Returns to original position
- Disabled: 50% opacity, no interactions
- Focus: Accessible focus ring

### Forms

Wrap forms with `.nebula-form` class to apply consistent styling:

```html
<form class="nebula-form">
  <div class="nebula-input-group">
    <label class="nebula-input-label">Label</label>
    <Input />
    <span class="nebula-input-hint">Helper text</span>
  </div>
</form>
```

**Features**:
- Consistent border radius and padding
- Hover and focus states with color transitions
- Validation states (error, warning)
- Accessible focus rings (3px with appropriate color)
- Disabled state styling

**Validation States**:
- `.semi-input-wrapper-error` - Red border with subtle background
- `.semi-input-wrapper-warning` - Orange border with subtle background

### Tables

**Classes**:
- `.nebula-table` - Enhanced table wrapper
- `.nebula-table-wrapper` - Adds border and border-radius
- `.nebula-table-zebra` - Alternating row colors
- `.nebula-table-compact` - Reduced padding

**Features**:
- Gradient header backgrounds
- Enhanced hover states
- Zebra striping option
- Consistent cell padding
- Empty and loading states

**Example**:
```html
<div class="nebula-table-wrapper">
  <Table className="nebula-table nebula-table-zebra" />
</div>
```

### Tabs

```html
<Tabs className="nebula-tabs">
  <!-- Tab items -->
</Tabs>
```

For admin console:
```html
<Tabs className="nebula-admin-tabs">
  <!-- Tab items -->
</Tabs>
```

**Features**:
- Pill-style active indicators
- Smooth transitions
- Elevated active state
- Gradient backgrounds

### Modals

```html
<Modal className="nebula-modal">
  <!-- Modal content -->
</Modal>
```

**Features**:
- Large border radius (20px)
- Enhanced shadows
- Proper header/footer borders
- Consistent padding

### Drawers

```html
<SideSheet className="nebula-drawer">
  <!-- Drawer content -->
</SideSheet>
```

**Features**:
- Rounded corners on the edge
- Enhanced shadows
- Proper header borders
- Scrollable body with custom scrollbar

### Toast Notifications

```html
<Toast className="nebula-toast nebula-toast-success" />
```

**Variants**:
- `.nebula-toast-success` - Green accent
- `.nebula-toast-error` - Red accent
- `.nebula-toast-warning` - Orange accent
- `.nebula-toast-info` - Blue accent

**Features**:
- Status-specific border colors
- Colored shadows for emphasis
- Consistent typography
- Rounded corners

### Badges & Tags

```html
<div class="nebula-badge nebula-badge-primary">Badge</div>
```

**Variants**:
- `.nebula-badge-primary` - Blue
- `.nebula-badge-success` - Green
- `.nebula-badge-warning` - Orange
- `.nebula-badge-danger` - Red

## Icons

### Icon Sizing

```css
--nebula-icon-xs: 14px;
--nebula-icon-sm: 16px;
--nebula-icon-md: 20px;
--nebula-icon-lg: 24px;
--nebula-icon-xl: 32px;
```

**Classes**: `.nebula-icon-xs`, `.nebula-icon-sm`, `.nebula-icon-md`, `.nebula-icon-lg`, `.nebula-icon-xl`

### Icon Colors

- `.nebula-icon-primary` - Primary brand color
- `.nebula-icon-success` - Green
- `.nebula-icon-warning` - Orange
- `.nebula-icon-danger` - Red
- `.nebula-icon-muted` - Muted gray

### Interactive Icons

```html
<span class="nebula-icon nebula-icon-md nebula-icon-clickable">
  <IconName />
</span>
```

**Features**:
- Hover background
- Scale transform on hover
- Color change on interaction

### Lucide Icons

All Lucide icons automatically use consistent stroke width:
- Standard: `var(--nebula-icon-stroke)` (2px)
- Light: `var(--nebula-icon-stroke-light)` (1.5px)

Apply `.lucide-light` class for lighter stroke weight.

## Layouts

### Console Container

```html
<div class="nebula-console-container">
  <!-- Console content -->
</div>
```

**Features**:
- Responsive padding (increases on larger screens)
- Minimum height adjustment
- Automatic card styling within

### Grid System

```html
<div class="nebula-grid nebula-grid-cols-1 nebula-grid-cols-md-2 nebula-grid-cols-lg-3">
  <!-- Grid items -->
</div>
```

**Stats Grid**:
```html
<div class="nebula-stats-grid">
  <div class="nebula-stat-card">
    <div class="nebula-stat-label">Label</div>
    <div class="nebula-stat-value">1,234</div>
    <div class="nebula-stat-trend nebula-stat-trend-up">+12%</div>
  </div>
</div>
```

### Stack & Inline

```html
<div class="nebula-stack"><!-- Vertical stack --></div>
<div class="nebula-inline"><!-- Horizontal inline --></div>
```

## Header & Navigation

### Header

```html
<header class="nebula-header">
  <div class="nebula-header-container">
    <!-- Header content -->
  </div>
</header>
```

**Components**:
- `.nebula-header-logo` - Logo with brand name
- `.nebula-header-nav` - Navigation links
- `.nebula-header-actions` - Action buttons
- `.nebula-header-action-btn` - Individual action button

### Sidebar

```html
<aside class="nebula-sidebar">
  <div class="nebula-sidebar-content">
    <div class="nebula-sidebar-section">
      <div class="nebula-sidebar-section-label">Section</div>
      <a class="nebula-sidebar-item nebula-sidebar-item-active">
        <span class="nebula-sidebar-item-icon">
          <Icon />
        </span>
        <span>Item</span>
      </a>
    </div>
  </div>
</aside>
```

## Compact Mode

Add `.nebula-compact` class to any container to reduce spacing:

```html
<div class="nebula-console-container nebula-compact">
  <!-- Reduced padding and spacing -->
</div>
```

## Color Usage

### Primary Colors
- Use for primary actions, links, and selected states
- Gradient: `linear-gradient(135deg, #2563eb 0%, #14b8a6 100%)`

### Status Colors
- **Success**: Green tones for positive actions and states
- **Warning**: Orange tones for cautions
- **Danger**: Red tones for errors and destructive actions
- **Info**: Blue tones for informational content

### Backgrounds
- Card backgrounds use subtle blue tints: `rgba(var(--semi-blue-0), 0.01-0.04)`
- Hover states: `rgba(var(--semi-blue-0), 0.08)`
- Active states: `rgba(var(--semi-blue-0), 0.12)`

## Accessibility

### Focus States
All interactive elements have accessible focus states:
- 3px outline with appropriate color
- High contrast for visibility
- Respects prefers-reduced-motion

### Color Contrast
- Text colors meet WCAG AA standards
- Interactive elements have sufficient contrast in all states
- Dark mode maintains equivalent contrast ratios

### Keyboard Navigation
- All interactive elements are keyboard accessible
- Logical tab order
- Visual focus indicators

## Responsive Breakpoints

```css
/* Mobile first approach */
@media (min-width: 768px) { /* Tablet */ }
@media (min-width: 1024px) { /* Desktop */ }
@media (min-width: 1280px) { /* Large Desktop */ }
```

### Responsive Utilities
- Grid columns adjust automatically
- Padding scales up on larger screens
- Navigation collapses on mobile
- Tables become cards on mobile (via CardTable component)

## Dark Mode

All components automatically support dark mode via `html.dark` selector:
- Darker backgrounds with blue tints
- Adjusted border colors
- Enhanced shadows for depth
- Maintained color contrast

## Best Practices

### Cards
- Use `.nebula-card` for user-facing content
- Use `.admin-card` for admin console tables
- Add `.nebula-card-premium` for highlighted content

### Forms
- Always wrap forms with `.nebula-form`
- Use consistent spacing with `.nebula-input-group`
- Provide validation feedback
- Add helper text where appropriate

### Tables
- Wrap tables with `.nebula-table-wrapper`
- Apply `.nebula-table-zebra` for alternating rows
- Use `.nebula-table-compact` when space is limited
- Provide empty and loading states

### Icons
- Maintain consistent sizing within contexts
- Use semantic colors (success, warning, danger)
- Add `.nebula-icon-clickable` for interactive icons
- Ensure adequate touch targets (minimum 44x44px)

### Typography
- Use semantic heading classes
- Maintain consistent line heights
- Limit line length for readability (max 65-75 characters)
- Use appropriate font weights for hierarchy

## Migration Guide

### From Legacy to NebulaGate

1. **Cards**: Add `.nebula-card` or `.admin-card` class
2. **Forms**: Wrap with `.nebula-form` class
3. **Tables**: Add `.nebula-table` and wrap with `.nebula-table-wrapper`
4. **Buttons**: Classes apply automatically to Semi-UI buttons
5. **Icons**: Add size classes (`.nebula-icon-md`, etc.)

### Gradual Adoption
- Components work alongside existing styles
- Apply classes incrementally
- Test in both light and dark modes
- Verify responsive behavior

## Future Enhancements

- Animation library for consistent micro-interactions
- Additional icon variants
- Extended color palette for data visualization
- Component variants for specific use cases
- Accessibility enhancements (improved screen reader support)
