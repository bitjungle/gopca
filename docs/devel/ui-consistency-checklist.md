# UI Consistency Checklist

This document defines standards for maintaining consistent user interface elements across GoPCA applications.

## Button Styles

### Primary Buttons
- **Purpose**: Main actions (Save, Submit, Analyze)
- **Style**: Blue background with white text
- **Classes**: `bg-blue-600 hover:bg-blue-700 text-white`
- **Focus**: `focus:ring-2 focus:ring-blue-500 focus:ring-offset-2`
- **Disabled**: `disabled:bg-gray-300 disabled:cursor-not-allowed`

### Secondary Buttons
- **Purpose**: Secondary actions (Cancel, Back)
- **Style**: Gray border with gray text
- **Classes**: `border border-gray-300 hover:bg-gray-50 text-gray-700`
- **Focus**: `focus:ring-2 focus:ring-gray-500 focus:ring-offset-2`
- **Dark mode**: `dark:border-gray-600 dark:hover:bg-gray-700 dark:text-gray-300`

### Danger Buttons
- **Purpose**: Destructive actions (Delete, Remove)
- **Style**: Red background with white text
- **Classes**: `bg-red-600 hover:bg-red-700 text-white`
- **Focus**: `focus:ring-2 focus:ring-red-500 focus:ring-offset-2`
- **Confirmation**: Always show confirmation dialog

### Button Sizes
- **Small**: `px-3 py-1.5 text-sm`
- **Medium** (default): `px-4 py-2 text-base`
- **Large**: `px-6 py-3 text-lg`

## Form Elements

### Input Fields
- **Base styles**: `border rounded-md px-3 py-2`
- **Normal**: `border-gray-300 focus:border-blue-500`
- **Error**: `border-red-500 focus:border-red-600`
- **Disabled**: `bg-gray-100 cursor-not-allowed`
- **Focus ring**: `focus:ring-2 focus:ring-blue-500`

### Labels
- **Position**: Above input field
- **Style**: `text-sm font-medium text-gray-700 dark:text-gray-300`
- **Required indicator**: Red asterisk `<span className="text-red-500">*</span>`
- **Help text**: Below input, `text-xs text-gray-500`

### Select Dropdowns
- **Same styling as input fields**
- **Custom arrow icon for consistency**
- **Keyboard navigation support (arrow keys)**

### Checkboxes & Radio Buttons
- **Size**: `w-4 h-4`
- **Color**: `text-blue-600`
- **Focus**: `focus:ring-2 focus:ring-blue-500`
- **Label spacing**: `ml-2` from control

## Error Messages

### Field-Level Errors
- **Position**: Below input field
- **Style**: `text-sm text-red-600 mt-1`
- **Icon**: Error icon before text
- **ARIA**: `role="alert"`

### Form-Level Errors
- **Position**: Top of form
- **Style**: Red background with border
- **Classes**: `bg-red-50 border border-red-200 rounded-md p-4`
- **Icon**: Exclamation icon
- **Dismissible**: Include close button

### Success Messages
- **Style**: Green background with border
- **Classes**: `bg-green-50 border border-green-200 rounded-md p-4`
- **Icon**: Check icon
- **Auto-dismiss**: After 5 seconds

## Loading States

### Spinner
- **Size options**: Small (16px), Medium (24px), Large (32px)
- **Color**: Match primary brand color
- **Animation**: Consistent rotation speed
- **Accessibility**: Include `aria-label="Loading"`

### Skeleton Screens
- **Background**: `bg-gray-200 dark:bg-gray-700`
- **Animation**: Shimmer effect
- **Match content layout**: Same size as loaded content

### Progress Bars
- **Height**: `h-2`
- **Background**: `bg-gray-200 dark:bg-gray-700`
- **Fill**: `bg-blue-600`
- **Text**: Show percentage if space allows

## Modal/Dialog Styles

### Backdrop
- **Color**: `bg-black bg-opacity-50`
- **Click behavior**: Close on backdrop click (configurable)
- **Animation**: Fade in/out

### Dialog Container
- **Background**: `bg-white dark:bg-gray-800`
- **Border radius**: `rounded-lg`
- **Shadow**: `shadow-xl`
- **Max width**: Configurable, default `max-w-md`

### Dialog Header
- **Title**: `text-lg font-semibold`
- **Close button**: Top right, `text-gray-400 hover:text-gray-600`
- **Border**: Optional bottom border

### Dialog Footer
- **Alignment**: Right-aligned buttons
- **Spacing**: `gap-2` between buttons
- **Order**: Secondary action left, primary action right

## Color Palette

### Light Mode
- **Background**: `white`, `gray-50`
- **Text Primary**: `gray-900`
- **Text Secondary**: `gray-600`
- **Borders**: `gray-300`
- **Primary**: `blue-600`
- **Success**: `green-600`
- **Warning**: `yellow-600`
- **Error**: `red-600`

### Dark Mode
- **Background**: `gray-900`, `gray-800`
- **Text Primary**: `gray-100`
- **Text Secondary**: `gray-400`
- **Borders**: `gray-700`
- **Primary**: `blue-500`
- **Success**: `green-500`
- **Warning**: `yellow-500`
- **Error**: `red-500`

## Typography

### Headings
- **H1**: `text-3xl font-bold` (Page titles)
- **H2**: `text-2xl font-semibold` (Section headers)
- **H3**: `text-xl font-semibold` (Subsections)
- **H4**: `text-lg font-medium` (Card headers)

### Body Text
- **Default**: `text-base`
- **Small**: `text-sm`
- **Extra small**: `text-xs` (Help text, labels)

### Font Families
- **Sans-serif**: System font stack
- **Monospace**: For code, data values

## Spacing

### Padding
- **Extra small**: `p-1` (4px)
- **Small**: `p-2` (8px)
- **Medium**: `p-4` (16px)
- **Large**: `p-6` (24px)
- **Extra large**: `p-8` (32px)

### Margins
- **Between sections**: `my-6`
- **Between form fields**: `my-4`
- **Between paragraphs**: `my-2`

### Grid/Flex Gaps
- **Tight**: `gap-2`
- **Normal**: `gap-4`
- **Wide**: `gap-6`

## Icons

### Source
- **Library**: Heroicons (outline style)
- **Size classes**: `w-4 h-4`, `w-5 h-5`, `w-6 h-6`
- **Stroke width**: `strokeWidth={1.5}`

### Common Icons
- **Save**: Floppy disk or download arrow
- **Delete**: Trash can
- **Edit**: Pencil
- **Settings**: Cog
- **Help**: Question mark circle
- **Info**: Information circle
- **Warning**: Exclamation triangle
- **Error**: X circle
- **Success**: Check circle

## Animation & Transitions

### Duration
- **Fast**: `duration-150` (150ms)
- **Normal**: `duration-300` (300ms)
- **Slow**: `duration-500` (500ms)

### Easing
- **Default**: `ease-in-out`
- **Enter**: `ease-out`
- **Exit**: `ease-in`

### Common Transitions
- **Hover states**: `transition-colors duration-150`
- **Expand/Collapse**: `transition-all duration-300`
- **Modal fade**: `transition-opacity duration-300`

## Responsive Design

### Breakpoints
- **Mobile**: `< 640px` (sm)
- **Tablet**: `640px - 1024px` (md-lg)
- **Desktop**: `> 1024px` (lg+)

### Mobile Considerations
- **Touch targets**: Minimum 44x44px
- **Font size**: Minimum 16px to prevent zoom
- **Spacing**: Increase padding on mobile
- **Navigation**: Hamburger menu on mobile

## Accessibility

### Focus Indicators
- **Style**: `ring-2 ring-offset-2`
- **Color**: Match component theme
- **Visible**: Never remove focus indicators

### ARIA Labels
- **Icons**: Always include `aria-label`
- **Interactive elements**: Descriptive labels
- **Loading states**: Include status messages

### Keyboard Support
- **Tab navigation**: Logical order
- **Enter/Space**: Activate buttons
- **Escape**: Close modals/dropdowns
- **Arrow keys**: Navigate lists/menus

## Testing Checklist

### Visual Consistency
- [ ] All buttons follow style guide
- [ ] Form elements have consistent styling
- [ ] Error messages use standard format
- [ ] Loading states are consistent
- [ ] Colors match palette

### Functionality
- [ ] All interactive elements are keyboard accessible
- [ ] Focus indicators are visible
- [ ] ARIA labels are present
- [ ] Responsive design works on all breakpoints
- [ ] Dark mode styling is complete

### Cross-Browser
- [ ] Chrome/Edge
- [ ] Firefox
- [ ] Safari
- [ ] Mobile browsers

### Performance
- [ ] Animations are smooth
- [ ] No layout shifts
- [ ] Images are optimized
- [ ] Lazy loading where appropriate

## Component Audit Status

| Component | Consistent | Accessible | Responsive | Dark Mode | Notes |
|-----------|------------|------------|------------|-----------|-------|
| Buttons | ✅ | ✅ | ✅ | ✅ | |
| Forms | ✅ | ⚠️ | ✅ | ✅ | Need ARIA improvements |
| Modals | ✅ | ✅ | ✅ | ✅ | |
| Charts | ✅ | ⚠️ | ✅ | ✅ | Need data tables |
| Navigation | ✅ | ⚠️ | ✅ | ✅ | Add skip links |
| Tables | ✅ | ⚠️ | ⚠️ | ✅ | Improve mobile view |
| Alerts | ✅ | ✅ | ✅ | ✅ | |

## Implementation Notes

1. **Use Tailwind CSS classes** for consistency
2. **Create reusable components** for common patterns
3. **Document any deviations** from standards
4. **Test with real users** including those using assistive technology
5. **Maintain this checklist** as the design system evolves

## Resources

- [Tailwind CSS Documentation](https://tailwindcss.com/docs)
- [Heroicons](https://heroicons.com/)
- [WCAG Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)
- [Material Design Guidelines](https://material.io/design)
- [Apple Human Interface Guidelines](https://developer.apple.com/design/human-interface-guidelines/)