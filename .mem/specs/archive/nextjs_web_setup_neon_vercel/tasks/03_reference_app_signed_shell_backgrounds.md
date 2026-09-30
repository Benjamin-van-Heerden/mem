---
title: 'Reference app: signed-in shell, backgrounds, home and login'
status: completed
created_at: "2026-09-30T13:26:00+02:00"
updated_at: "2026-09-30T13:40:20+02:00"
completed_at: "2026-09-30T13:40:20+02:00"
---

Port shannon's navigation (~/Documents/DNA/shannon/src/frontend/components/navigation/) to Cache Components: routes.ts with permission-filtered items, app-sidebar (64px icon rail with tooltips, logo), app-mobile-menu (left sheet), app-header (page title, email, sign-out) with user-dependent parts in Suspense; dashboard placeholder. Add .bg-dot-grid and .bg-hatch utilities on tokens in globals.css, selectable on the shell. Basic (site) home page and login page using public/logo.svg (placeholder logo) and tokens. Check in the browser at desktop and phone widths, both backgrounds, light and dark.

## Completion Notes

ref2: src/components/navigation/{routes.ts (items with optional permission, visibleItems, isActive, pageTitle), app-sidebar.tsx (64px icon rail, logo, tooltips, aria-current), app-mobile-menu.tsx (left sheet with email and sign-out), page-title.tsx, sign-out-button.tsx}; (app)/layout.tsx renders the shell statically and streams the user's parts (permission-filtered items, email) in Suspense with the same components as fallbacks; server passes only permission names (lucide icons cannot cross to client components); permissionsFor in permissions.ts. globals.css: @utility bg-dot-grid and bg-hatch on tokens. Dashboard: greeting, three placeholder metric cards, the notes example. (site) home (logo, headline, lead, sign-in) and login page (logo, card, 'accounts are created by an administrator'); placeholder public/logo.svg. Build passes (/, /login static; /dashboard, /admin/users partial prerender). Checked in the browser: desktop rail with both items for the super admin, tooltip, header with email and sign-out; phone: menu sheet with items, email, active state and sign-out; users table fixed to fit phones (role badge under the name below sm); dark mode with bg-hatch and light mode with bg-dot-grid render.
