# Phase 33 — Relaxed Suggestion Entry

Builds on phase 12 (song suggestions).

## Goal

Make both the link and message fields optional (at least one required) instead of requiring a link. Update placeholder copy so users understand they can submit a text-based suggestion without a link.

## Scope

### In scope
- Backend: remove link-required validation, enforce "at least one field non-empty"
- Frontend: remove link-required validation, update `canSubmit` logic, update placeholder text
- Placeholder copy changes

### Out of scope
- Changing the form layout or field count
- Any database schema changes

## Backend changes

### `server/handlers/suggestions.go` — `SubmitSuggestion()`

Current validation (lines ~44-56):
1. Link is required (empty → 400)
2. Link must be valid URL
3. Link max 2048
4. Message max 200

New validation:
1. At least one of `link` or `message` must be non-empty → 400 `"link or message is required"`
2. **If** link is non-empty, it must be a valid URL (http/https, valid host) → 400 `"link must be a valid url"`
3. **If** link is non-empty, max 2048 chars → 400 `"link is too long"`
4. **If** message is non-empty, max 200 chars → 400 `"message is too long"`

## Frontend changes

### `client/src/components/SuggestionForm.tsx`

**`canSubmit` logic:** Change from requiring a valid link to requiring at least one field non-empty. If link is non-empty, it must still pass `isValidUrl()`.

```
canSubmit = !submitting && (
  (link.trim() !== "" && isValidUrl(link)) ||
  message.trim() !== ""
)
```

Show the URL validation error only when the link field is non-empty and invalid.

**Placeholder text:**
| Field   | Current                                  | New                                          |
|---------|------------------------------------------|----------------------------------------------|
| Link    | `"spotify, apple music, or youtube link"` | `"spotify, youtube, or any link (optional)"` |
| Message | `"leave a message (optional)"`           | `"song name, artist, or a message"`          |

## Definition of done

- [ ] Can submit with only a link (valid URL) — works as before
- [ ] Can submit with only a message — no link validation triggered
- [ ] Can submit with both a link and message
- [ ] Cannot submit when both fields are empty
- [ ] Invalid URL in link field still shows validation error and blocks submit
- [ ] Backend returns 400 with `"link or message is required"` when both empty
- [ ] Placeholder text updated to new copy
