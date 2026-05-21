## Summary
- **Backend**: PATCH /api/users/{id} extended with profile fields (avatarPreset, avatarColor, isRestricted, maxRating); new POST /api/users/{id}/pin endpoint to set/clear profile PINs; governance triple (router + authz + route-policy) complete
- **Profile token plumbing**: `profile-token-store.ts` persists the X-Profile-Token in localStorage; `api/client.ts` injects it automatically on every request so the server enforces the rating ceiling
- **profileStore**: Svelte 5 rune store tracking the active profile card and picker visibility
- **WhoIsWatching**: Full-screen Netflix-style profile picker with two-step PIN flows (exit restricted profile -> entry PIN for adult profile) and animated avatar tiles
- **PinPad**: Reusable 4-digit modal with keyboard entry, animated dots, backspace, and error shake
- **Header**: Active profile avatar shown in the account button; "Switch Profile" opens the picker
- **Settings / Users**: Expand-in-place profile editor with avatar preset picker (10 SVG animals), rating ceiling dropdown, Kids toggle, and separate PIN management panel
- **10 animal SVG avatars** added to `static/avatars/`: cat, dog, fox, bear, rabbit, owl, penguin, fish, lion, panda

## Test plan
- [ ] `go test ./...` passes in server/
- [ ] `npx vitest run` passes (28 tests)
- [ ] `npx svelte-check` 0 errors
- [ ] On first load after login, Who's Watching picker appears full-screen
- [ ] Selecting a profile without a PIN immediately enters the app
- [ ] Selecting a restricted (kids) profile shows no PIN pad (enters directly)
- [ ] Setting a kids profile exit PIN: switching away prompts for it
- [ ] Setting an adult entry PIN: clicking that profile prompts for it
- [ ] Wrong PIN shows error and resets dots; correct PIN proceeds
- [ ] Active profile avatar shown in header button; dropdown says "Switch Profile"
- [ ] Settings > Users: Edit profile expands inline; saving avatar preset, rating, kids toggle updates list
- [ ] Setting a PIN shows "PIN set" badge on the user row
- [ ] Rating ceiling: logged in as PG-13 profile, R-rated movies not visible in browse/search

Closes #83

Generated with [Claude Code](https://claude.com/claude-code)
