# NebulaGate Design Tokens

## Overview

This document describes the NebulaGate design system tokens — a Nordic-inspired aesthetic combining muted blues, cool grays, earthen neutrals, and aurora-inspired accent hues.

## Color Palette

### Brand Colors

#### Primary (Muted Nordic Blue)
- `--nebula-brand-primary`: #5B7C99 (light) / #7B9FBD (dark)
- `--nebula-brand-primary-hover`: #4A6A85 (light) / #91B3D1 (dark)
- `--nebula-brand-primary-active`: #3B5770 (light) / #A8C7E3 (dark)

#### Secondary
- `--nebula-brand-secondary`: #2563eb (light) / #60a5fa (dark)
- `--nebula-brand-secondary-hover`: #1d4ed8 (light) / #93c5fd (dark)

### Aurora Accent Colors

Inspired by the Northern Lights:

- `--nebula-aurora-teal`: #14B8A6 (light) / #5EEAD4 (dark)
- `--nebula-aurora-purple`: #8B5CF6 (light) / #A78BFA (dark)
- `--nebula-aurora-cyan`: #06B6D4 (light) / #67E8F9 (dark)
- `--nebula-aurora-emerald`: #10B981 (light) / #6EE7B7 (dark)

### Gray Scale (Cool Scandinavian Grays)

- `--nebula-gray-50`: #F8FAFC
- `--nebula-gray-100`: #F1F5F9
- `--nebula-gray-200`: #E2E8F0
- `--nebula-gray-300`: #CBD5E1
- `--nebula-gray-400`: #94A3B8
- `--nebula-gray-500`: #64748B
- `--nebula-gray-600`: #475569
- `--nebula-gray-700`: #334155
- `--nebula-gray-800`: #1E293B
- `--nebula-gray-900`: #0F172A

## Semantic Tokens

### Background Colors

- `--nebula-bg-base`: Base application background
- `--nebula-bg-surface`: Primary surface (cards, panels)
- `--nebula-bg-elevated`: Elevated surfaces (modals, popovers)
- `--nebula-bg-overlay`: Semi-transparent overlay
- `--nebula-bg-subtle`: Subtle background variation

### Surface Colors

- `--nebula-surface-primary`: Primary surface color
- `--nebula-surface-secondary`: Secondary surface color
- `--nebula-surface-tertiary`: Tertiary surface color
- `--nebula-surface-hover`: Interactive hover state
- `--nebula-surface-active`: Interactive active state

### Border Colors

- `--nebula-border-subtle`: Subtle borders (dividers, separators)
- `--nebula-border-default`: Default border color
- `--nebula-border-strong`: Strong emphasis borders
- `--nebula-border-accent`: Accent borders (highlights)
- `--nebula-border-focus`: Focus ring color

### Text Colors

- `--nebula-text-primary`: Primary text (headings, important content)
- `--nebula-text-secondary`: Secondary text (body copy)
- `--nebula-text-tertiary`: Tertiary text (metadata, labels)
- `--nebula-text-quaternary`: Quaternary text (placeholder, disabled)
- `--nebula-text-disabled`: Disabled state text
- `--nebula-text-inverse`: Inverse text (light text on dark backgrounds)
- `--nebula-text-brand`: Brand-colored text
- `--nebula-text-link`: Link color
- `--nebula-text-link-hover`: Link hover color

### Status Colors

#### Success
- `--nebula-success-[50-700]`: Success color scale
- `--nebula-success-text`: Success text color
- `--nebula-success-bg`: Success background
- `--nebula-success-border`: Success border

#### Warning
- `--nebula-warning-[50-700]`: Warning color scale
- `--nebula-warning-text`: Warning text color
- `--nebula-warning-bg`: Warning background
- `--nebula-warning-border`: Warning border

#### Error
- `--nebula-error-[50-700]`: Error color scale
- `--nebula-error-text`: Error text color
- `--nebula-error-bg`: Error background
- `--nebula-error-border`: Error border

#### Info
- `--nebula-info-[50-700]`: Info color scale
- `--nebula-info-text`: Info text color
- `--nebula-info-bg`: Info background
- `--nebula-info-border`: Info border

### Interactive States

- `--nebula-interactive-default`: Default interactive element color
- `--nebula-interactive-hover`: Hover state
- `--nebula-interactive-active`: Active/pressed state
- `--nebula-interactive-disabled`: Disabled state
- `--nebula-interactive-focus-ring`: Focus ring color with opacity

## Spacing Scale

Generous whitespace following the 8pt grid system:

- `--nebula-spacing-xs`: 4px
- `--nebula-spacing-sm`: 8px
- `--nebula-spacing-md`: 16px
- `--nebula-spacing-lg`: 24px
- `--nebula-spacing-xl`: 32px
- `--nebula-spacing-2xl`: 48px

### Tailwind Usage
```jsx
<div className="p-nebula-md gap-nebula-lg">
  <div className="mt-nebula-xl">...</div>
</div>
```

## Border Radius

Refined geometry with smooth corners:

- `--nebula-radius-sm`: 12px (buttons, tags)
- `--nebula-radius-md`: 16px (cards, inputs)
- `--nebula-radius-lg`: 20px (large cards, modals)
- `--nebula-radius-xl`: 24px (hero sections, special containers)

### Tailwind Usage
```jsx
<div className="rounded-nebula-radius-md">...</div>
```

## Typography

### Font Families

- `--nebula-font-sans`: Inter Variable (primary font stack)
- `--nebula-font-mono`: JetBrains Mono (code, monospace)

### Font Sizes

- `--nebula-text-xs`: 12px
- `--nebula-text-sm`: 14px
- `--nebula-text-base`: 16px
- `--nebula-text-lg`: 18px
- `--nebula-text-xl`: 20px
- `--nebula-text-2xl`: 24px
- `--nebula-text-3xl`: 30px

### Tailwind Usage
```jsx
<h1 className="text-nebula-2xl font-bold">Heading</h1>
<p className="text-nebula-sm text-nebula-text-secondary">Body text</p>
```

## Elevation/Shadows

Light mode shadows (subtle, Nordic-inspired):

- `--nebula-shadow-sm`: 0 2px 8px rgba(15, 23, 42, 0.06)
- `--nebula-shadow-md`: 0 8px 24px rgba(15, 23, 42, 0.08)
- `--nebula-shadow-lg`: 0 16px 40px rgba(15, 23, 42, 0.12)
- `--nebula-shadow-xl`: 0 24px 60px rgba(15, 23, 42, 0.16)

Dark mode shadows (deeper, more pronounced):

- `--nebula-shadow-dark-sm`: 0 2px 8px rgba(0, 0, 0, 0.3)
- `--nebula-shadow-dark-md`: 0 8px 24px rgba(0, 0, 0, 0.4)
- `--nebula-shadow-dark-lg`: 0 16px 40px rgba(0, 0, 0, 0.5)
- `--nebula-shadow-dark-xl`: 0 24px 60px rgba(0, 0, 0, 0.6)

### Tailwind Usage
```jsx
<div className="shadow-nebula-md hover:shadow-nebula-lg">Card</div>
```

## Dark Mode

All tokens automatically adjust for dark mode when `html.dark` or `[theme-mode='dark']` is applied to the root element.

### Key Differences in Dark Mode:
- Primary colors become lighter for better contrast
- Backgrounds transition to deep slate tones
- Text hierarchy maintains readability with adjusted contrast
- Aurora accents become more vibrant
- Shadows become deeper and more pronounced

## Semi UI Integration

NebulaGate tokens are mapped to Semi UI variables for seamless component integration:

```css
/* Example mappings */
--semi-color-primary → --nebula-brand-primary
--semi-color-bg-0 → --nebula-bg-base
--semi-color-text-0 → --nebula-text-primary
--semi-color-success → --nebula-success-600
```

This ensures all Semi UI components automatically inherit the NebulaGate aesthetic.

## Usage Examples

### Card Component
```jsx
<div className="nebula-card p-nebula-lg">
  <h3 className="text-nebula-xl text-nebula-text-primary font-semibold">
    Card Title
  </h3>
  <p className="text-nebula-sm text-nebula-text-secondary mt-nebula-sm">
    Card description text with proper spacing and colors.
  </p>
</div>
```

### Button with Brand Colors
```jsx
<button className="bg-nebula-brand-primary hover:bg-nebula-brand-primary-hover 
  text-white px-nebula-md py-nebula-sm rounded-nebula-radius-sm">
  Primary Action
</button>
```

### Status Badge
```jsx
<span className="bg-nebula-success-bg text-nebula-success-text 
  border border-nebula-success-border px-nebula-sm py-nebula-xs 
  rounded-nebula-radius-sm text-nebula-xs font-medium">
  Success
</span>
```

## Migration Guide

When migrating existing components to use NebulaGate tokens:

1. **Replace hardcoded colors** with semantic tokens
2. **Use spacing tokens** instead of arbitrary values
3. **Apply consistent border radius** from the scale
4. **Leverage text hierarchy** tokens for typography
5. **Test in both light and dark modes**

## Best Practices

1. **Prefer semantic tokens** over raw color values
2. **Use spacing scale** consistently for predictable layouts
3. **Apply appropriate text hierarchy** for readability
4. **Test dark mode** appearance for all UI elements
5. **Use Interactive states** for hover/active feedback
6. **Maintain generous whitespace** for Nordic aesthetic
7. **Apply elevation shadows** thoughtfully to establish hierarchy

## Resources

- Font: [Inter Variable](https://rsms.me/inter/) (installed via @fontsource-variable/inter)
- Inspiration: Nordic design principles, Aurora Borealis color palettes
- Typography: Modern, clean sans-serif with excellent i18n support

---

**Last Updated:** 2025-01-30
**Design System Version:** 1.0.0
