## Summary
- Settings > Account: "Save profile", "Update password", and "Sign out" were all completely dead — no `onclick`, no `bind:value`
- Added `updateUser` (display name) and `logout` imports (both were missing)
- Added `$state` variables for all form inputs and status flags
- `$effect` seeds the display name field from `currentUser.displayName` once the session loads
- `handleAccountSaveProfile()` — calls `updateUser(id, { displayName })`, shows saving state + success/error feedback
- `handleAccountUpdatePassword()` — validates all three fields (min 8 chars, confirm match), calls `updateUserPassword`, clears fields on success
- `handleSignOut()` — calls `logout()`, clears `xuva-auth-token` and `xuva-profile-token` from localStorage, redirects to `/`
- All buttons show in-flight text and are `disabled` while requests are pending

## Test plan
- [ ] Open Settings → Account
- [ ] Change display name → click Save profile → success message appears, name persists on reload
- [ ] Submit Save profile with network error → error message shown
- [ ] Enter current password + mismatched new/confirm → error "Passwords do not match"
- [ ] Enter short new password → error "must be at least 8 characters"
- [ ] Enter valid current + new + confirm → click Update password → success, fields clear
- [ ] Click Sign out → redirected to `/` (login/picker screen)

Closes #146

🤖 Generated with [Claude Code](https://claude.com/claude-code)
