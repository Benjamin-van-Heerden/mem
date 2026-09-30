---
title: 'Reference app: signed-in shell, backgrounds, home and login'
status: todo
created_at: "2026-09-30T13:26:00+02:00"
updated_at: "2026-09-30T13:26:00+02:00"
---

Port shannon's navigation (~/Documents/DNA/shannon/src/frontend/components/navigation/) to Cache Components: routes.ts with permission-filtered items, app-sidebar (64px icon rail with tooltips, logo), app-mobile-menu (left sheet), app-header (page title, email, sign-out) with user-dependent parts in Suspense; dashboard placeholder. Add .bg-dot-grid and .bg-hatch utilities on tokens in globals.css, selectable on the shell. Basic (site) home page and login page using public/logo.svg (placeholder logo) and tokens. Check in the browser at desktop and phone widths, both backgrounds, light and dark.
