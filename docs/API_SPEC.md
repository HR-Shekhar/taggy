# Taggy API Specification (v0)

Base URL: `http://localhost:<PORT>`

All JSON requests use `Content-Type: application/json`.

## Error format

```json
{
  "message": "human readable error",
  "fields": {
    "Email": "must be a valid email"
  }
}
```

`fields` is only present for validation errors (422).

| Status | Meaning |
|--------|---------|
| 400 | Bad request (malformed JSON) |
| 401 | Unauthorized |
| 404 | Not found |
| 409 | Conflict (duplicate email/username) |
| 422 | Validation failed |
| 429 | Too many requests (per-IP rate limit on auth/OTP/join/report) |
| 500 | Internal server error |

Sensitive endpoints use in-memory per-IP token buckets (single process; Redis optional later):

| Endpoint | Limit |
|----------|-------|
| `POST /auth/register` | 5 / min |
| `POST /auth/login` | 20 / min |
| `POST /auth/verify-email`, `POST /auth/resend-verification` | 5 / min |
| `POST /skills/{slug}/join`, `POST /pods/{podSlug}/join` | 30 / min |
| `POST /reports` | 10 / min |

---

## Health

### GET /health

**Auth:** none

**Response 200:**
```json
{ "status": "ok" }
```

---

## Auth

### POST /auth/register

**Auth:** none

**Body:**
```json
{
  "email": "user@example.com",
  "username": "taggy_user",
  "name": "Jane Doe",
  "password": "Password1!"
}
```

**Password rules:** min 8 characters, at least one uppercase, one lowercase, one number, and one special character.

**Response 201:**
```json
{
  "public_id": "uuid",
  "email": "user@example.com",
  "username": "taggy_user",
  "name": "Jane Doe",
  "email_verified": false,
  "dev_otp": "123456"
}
```

`dev_otp` is included only when the app uses the **DevLogger** mailer (development **and** `RESEND_API_KEY` unset). When Resend is configured, OTP is emailed and `dev_otp` is omitted. Login is blocked until verified.

---

### POST /auth/verify-email

**Auth:** none

**Body:**
```json
{
  "email": "user@example.com",
  "otp": "123456"
}
```

**Response 200:** user object with `email_verified: true`

**Errors:** 401 invalid OTP · 400 expired OTP · 409 already verified

---

### POST /auth/resend-verification

**Auth:** none

**Body:** `{ "email": "user@example.com" }`

**Response 204:** new OTP issued (invalidates previous codes). With DevLogger mailer, **200** with `{ "dev_otp": "123456" }` instead.

---

### POST /auth/login

**Auth:** none

**Body:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response 200:**
```json
{
  "access_token": "jwt...",
  "refresh_token": "opaque...",
  "username": "taggy_user"
}
```

**Response 403:** email not verified (local accounts)

---

### POST /auth/refresh

**Auth:** none

**Body:**
```json
{
  "refresh_token": "opaque..."
}
```

**Response 200:** same shape as login (includes `username`)

---

### GET /auth/google/start

Starts Google OAuth. Requires `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` in env.

**Auth:** none

**Response 200:** `{ "url": "https://accounts.google.com/..." }`

**Errors:** 400 `oauth not configured`

---

### GET /auth/google/callback

Google redirects here with `?code=...&state=...`. Redirect URI must match `GOOGLE_REDIRECT_URL`.

**Auth:** none

**If the Google account already maps to a Taggy user (or email can be linked):**
- Browser **302** → `{FRONTEND_URL}/auth/callback#access_token=...&refresh_token=...&username=...`
- API (`Accept: application/json`) → **200** token JSON

**If this is a brand-new Google user (no Taggy account yet):**
- **No DB row is created yet**
- Browser **302** → `{FRONTEND_URL}/auth/complete-google#registration_token=...&email=...&name=...&picture=...`
- API → **200** `{ "registration_token", "email", "name", "picture" }`
- Token TTL: 30 minutes. If the user never completes, nothing is stored (full rollback).

On failure, redirects to `{FRONTEND_URL}/auth/callback?error=...`

---

### POST /auth/google/complete

Finishes Google signup after the user chooses a username (and optional name).

**Auth:** none

**Body:**
```json
{
  "registration_token": "...",
  "username": "bright_bit",
  "name": "Bright Bit"
}
```

**Response 201:** same token response as login

**Errors:**
- 400 `invalid or expired google registration token`
- 400 `username is required` / `username format is invalid`
- 409 `username already in use` / `email already in use`

**OAuth flow (teaching):**
1. Client calls `/auth/google/start` and redirects to Google
2. User approves; Google hits backend callback
3. Existing user → session tokens; new user → pending registration token only
4. New user submits username via `/auth/google/complete` → user + Google identity created in one transaction → session tokens
5. Abandoning step 4 = no account (token expires)

---

### POST /auth/logout

**Auth:** none

**Body:**
```json
{
  "refresh_token": "opaque..."
}
```

**Response 204:** empty body

---

### POST /auth/logout-all

**Auth:** Bearer access token

**Response 204:** empty body

---

## Users (Profile)

GitHub-style: always use `/users/{username}`. There is no `/users/me`.

### GET /users/{username}

**Auth:** optional. No token → public fields only. Valid token for the profile owner → public + private fields.

**Example:** `GET /users/taggy_tester`

**Response 200 (public — anyone, or another user’s profile):**
```json
{
  "username": "taggy_tester",
  "name": "Jane Doe",
  "bio": "Learning backend",
  "profile_picture_url": "https://example.com/photo.jpg"
}
```

**Response 200 (owner — same URL with your access token):**
```json
{
  "username": "taggy_tester",
  "name": "Jane Doe",
  "bio": "Learning backend",
  "profile_picture_url": "https://example.com/photo.jpg",
  "public_id": "uuid",
  "email": "user@example.com",
  "email_verified": false,
  "subscription": "FREE"
}
```

`bio` and `profile_picture_url` are `null` when not set.

**Response 404:** username not found or account deleted.

---

### POST /users/{username}/avatar

**Auth:** Bearer access token required. You can only upload **your own** photo.

Multipart form field: `file` (JPG, PNG, WEBP, or GIF, max 500 KB). Stored on Cloudinary; the resulting HTTPS URL is saved to `users.profile_picture_url`.

**Response 200:** same as owner GET profile.

**Errors:** 400 invalid file · 403 not owner · 503 Cloudinary not configured · 502 upload failed

---

### PATCH /users/{username}

**Auth:** Bearer access token required. You can only update **your own** username (JWT subject must match the profile).

**Body (all fields optional):**
```json
{
  "name": "New Name",
  "bio": "Updated bio",
  "profile_picture_url": "https://example.com/new.jpg",
  "username": "new_username"
}
```

**Response 200:** full profile (same shape as GET when viewing yourself).

**Response 403:** `{username}` is not you.

**Business rules:**
- `username` must match `^[a-zA-Z0-9._]{3,30}$`
- Username cannot be taken by another user
- Old usernames are reserved forever (cannot be reused)
- Email cannot be changed via this endpoint

---

### Login / refresh

`POST /auth/login` and `POST /auth/refresh` include `username` in the response so clients know which profile path to use.

```json
{
  "access_token": "jwt...",
  "refresh_token": "opaque...",
  "username": "taggy_tester"
}
```

---

## Skills & Milestones (Phase 2)

All routes require `Authorization: Bearer <access_token>`.

Public identifiers use **slugs** (not numeric IDs) in URLs and JSON responses.

### GET /skills

List all active skills.

**Response 200:**
```json
[
  {
    "name": "Web Development",
    "slug": "web-development",
    "description": "..."
  }
]
```

---

### GET /skills/{slug}

Get skill detail and its official community.

**Response 200:**
```json
{
  "skill": { "name": "...", "slug": "web-development", "description": "..." },
  "community": { "slug": "web-development", "name": "...", "description": "..." }
}
```

Community `slug` matches the skill slug (1:1 official community per skill).

---

## Roadmaps (Phase 2)

Each skill has one roadmap container with one or more versions (`DRAFT` / `ACTIVE` / `ARCHIVED`).  
`roadmaps.current_version_id` points at the official catalog version. Enrollment pins the learner to a specific `roadmap_version_id`.

### GET /skills/{slug}/roadmap

Roadmap overview: current version + all versions.

**Response 200:**
```json
{
  "skill_slug": "web-development",
  "skill_name": "Web Development",
  "current_version": {
    "version_number": 1,
    "status": "ACTIVE",
    "generated_by": "ADMIN",
    "is_current": true,
    "milestone_count": 5,
    "published_at": "2026-08-06T...",
    "created_at": "2026-08-06T..."
  },
  "versions": [ /* same shape */ ]
}
```

---

### GET /skills/{slug}/roadmap/versions

List versions (newest first).

---

### GET /skills/{slug}/roadmap/versions/{versionNumber}

Version detail including ordered milestones (catalog preview; no progress).

**Response 200:**
```json
{
  "skill_slug": "web-development",
  "skill_name": "Web Development",
  "version_number": 2,
  "status": "ARCHIVED",
  "generated_by": "ADMIN",
  "is_current": false,
  "milestones": [
    {
      "slug": "html-css-basics",
      "title": "HTML & CSS Basics",
      "description": "...",
      "estimated_hours": 40,
      "order_index": 1,
      "difficulty": "BEGINNER"
    }
  ]
}
```

---

### PUT /users/{username}/skills/{skillSlug}/roadmap-version

Switch the authenticated user's enrollment to another **published** version (`ACTIVE` or `ARCHIVED`). Drafts cannot be selected. Progress is remapped by matching milestone `slug` (completed slugs stay completed).

**Request:**
```json
{ "version_number": 2 }
```

**Business rules:**
- Must be enrolled in the skill
- Target version must exist and not be `DRAFT`
- Same version → 409 conflict

---

### POST /skills/{slug}/join

Enroll in a skill. Creates `userskill` + milestone progress rows. Joins community implicitly.

**Response 201:**
```json
{
  "user_skill": {
    "skill_slug": "web-development",
    "skill_name": "Web Development",
    "status": "ACTIVE",
    "started_at": "2026-08-06T..."
  },
  "community": { "slug": "web-development", "name": "...", "description": "..." }
}
```

**Business rules:**
- Free users: max **1 ACTIVE** skill total
- Cannot join same skill twice
- Skill must be active and have an ACTIVE roadmap version

---

### GET /users/{username}/skills

List skill enrollments. Path username must be **your own** (JWT subject must match).

**Example:** `GET /users/taggy_tester/skills`

**Response 200:**
```json
[
  {
    "skill_slug": "web-development",
    "skill_name": "Web Development",
    "status": "ACTIVE",
    "started_at": "2026-08-01T10:00:00Z"
  }
]
```

---

### GET /users/{username}/skills/{skillSlug}/milestones

Milestones with progress for an enrolled skill.

**Example:** `GET /users/taggy_tester/skills/web-development/milestones`

**Response 200:**
```json
[
  {
    "slug": "html-css-basics",
    "title": "HTML & CSS Basics",
    "description": "Learn semantic HTML and modern CSS layout",
    "estimated_hours": 8,
    "order_index": 1,
    "difficulty": "BEGINNER",
    "status": "NOT_STARTED",
    "completed_at": null,
    "postponed_until": null
  }
]
```

---

### PATCH /users/{username}/skills/{skillSlug}/milestones/{milestoneSlug}

Update milestone progress.

**Example:** `PATCH /users/taggy_tester/skills/web-development/milestones/html-css-basics`

**Complete:**
```json
{ "action": "COMPLETE" }
```

**Postpone:**
```json
{
  "action": "POSTPONE",
  "postponed_until": "2026-08-15T00:00:00Z"
}
```

**Business rules:**
- Must complete milestones in order (`order_index`)
- Cannot skip milestones permanently
- Cannot complete an already-completed milestone

---

## Progress (Study Hours & Streaks)

All endpoints require JWT. Path `{username}` must be **your own**.

### POST /users/{username}/study-sessions

Log a study session and update the user's streak atomically.

**Body:**
```json
{
  "skill_slug": "web-development",
  "duration_minutes": 45,
  "notes": "Worked on flexbox layouts",
  "studied_at": "2026-08-06T10:30:00Z"
}
```

`studied_at` is optional; defaults to server time. `notes` is optional.

**Response 201:**
```json
{
  "session": {
    "skill_slug": "web-development",
    "duration_minutes": 45,
    "notes": "Worked on flexbox layouts",
    "studied_at": "2026-08-06T10:30:00Z",
    "created_at": "2026-08-06T10:30:01Z"
  },
  "streak": {
    "current_streak": 1,
    "longest_streak": 1,
    "last_activity_date": "2026-08-06",
    "freeze_count": 0
  }
}
```

**Errors:**
- 403 — not enrolled in the skill
- 422 — validation (e.g. `duration_minutes` < 1)

**Streak rules (UTC calendar days):**
- First session → streak = 1
- Another session same day → streak unchanged
- Session on consecutive day → streak + 1
- Gap > 1 day → streak resets to 1

---

### GET /users/{username}/study-sessions

List study session history (most recent first).

**Query params:**
- `skill_slug` (optional) — filter to one enrolled skill

**Response 200:**
```json
[
  {
    "skill_slug": "web-development",
    "duration_minutes": 45,
    "notes": "Worked on flexbox layouts",
    "studied_at": "2026-08-06T10:30:00Z",
    "created_at": "2026-08-06T10:30:01Z"
  }
]
```

---

### GET /users/{username}/streak

Current streak cache for the user.

**Response 200:**
```json
{
  "current_streak": 3,
  "longest_streak": 5,
  "last_activity_date": "2026-08-06",
  "freeze_count": 0
}
```

Returns zeros when no sessions have been logged yet.

---

### GET /users/{username}/progress/summary

Dashboard aggregate: total / weekly / monthly study minutes plus streak.

**Response 200:**
```json
{
  "total_minutes": 320,
  "weekly_minutes": 120,
  "monthly_minutes": 280,
  "current_streak": 3,
  "longest_streak": 5
}
```

---

## Pods (Phase 4)

All routes require `Authorization: Bearer <access_token>`.

Pods are skill-specific. Public URLs use **pod slug**. Rules:

- Max 7 ACCEPTED members (`max_members`)
- One ACCEPTED pod per user
- Join by request (PENDING → owner accept/reject)
- Roles: `OWNER` | `ADMIN` | `MEMBER` (on ACCEPTED memberships)
- Owner can promote members to `ADMIN`, demote to `MEMBER`, or transfer `OWNER`
- Owner leave: if alone → pod is deleted; else ownership transfers to a random admin, or if none a random accepted member
- Owner may `DELETE` the pod only when empty (sole accepted member)
- Must be enrolled in the skill to create or join

### POST /skills/{skillSlug}/pods

Create a pod. Caller becomes owner + ACCEPTED member (`role: OWNER`).

**Body:**
```json
{
  "slug": "web-dev-grinders",
  "name": "Web Dev Grinders",
  "description": "Daily accountability"
}
```

`slug` is chosen by the user (3–60 chars: `a-z`, `0-9`, hyphens). If it already exists → **409** `pod slug already in use` (no auto-suffix).

**Response 201:** pod object (`slug`, `name`, `skill_slug`, `owner_username`, `max_members`, `accepted_count`, …)

**Errors:** 400 invalid slug/name · 403 not enrolled · 409 already in an active pod / slug taken · 404 skill not found

---

### GET /skills/{skillSlug}/pods

List pods for a skill (includes `accepted_count`).

---

### GET /pods/{podSlug}

Pod detail plus:
- `members` — ACCEPTED members (each with `role`)
- `join_requests` — PENDING join requests (owner can accept/reject via member endpoints)

---

### GET /users/{username}/pods

List the authenticated user's memberships (PENDING/ACCEPTED/… + `role`). JWT must own `{username}`.

---

### POST /pods/{podSlug}/join

Request to join → status `PENDING` (or re-open after REJECTED/LEFT/REMOVED).

**Errors:** 403 not enrolled · 409 already pending / already member / already in active pod · 404 pod not found

---

### POST /pods/{podSlug}/members/{username}/accept

Owner accepts a PENDING request (transaction: capacity + one-active-pod checks). Accepted members get `role: MEMBER`.

**Errors:** 403 not owner · 409 pod full / already in active pod / not pending

---

### POST /pods/{podSlug}/members/{username}/reject

Owner rejects a PENDING request.

---

### POST /pods/{podSlug}/members/{username}/role

Owner sets an ACCEPTED member's role.

**Body:**
```json
{
  "role": "ADMIN"
}
```

Allowed values: `ADMIN`, `MEMBER`, `OWNER`.

- `ADMIN` / `MEMBER`: promote or demote (not self)
- `OWNER`: transfer ownership; previous owner becomes `ADMIN`

**Errors:** 400 invalid role · 403 not owner / cannot change own role · 404 membership not found

---

### POST /pods/{podSlug}/leave

Leave pod (PENDING or ACCEPTED).

- Non-owner: status → `LEFT`
- Owner alone (empty pod): pod is **deleted**
- Owner with others: ownership transfers to a random `ADMIN`, or if none a random accepted member (promoted to `OWNER`), then former owner → `LEFT`

**Response 204**

---

### DELETE /pods/{podSlug}

Owner deletes the pod **only when empty** (sole ACCEPTED member).

**Errors:** 403 not owner · 409 pod is not empty

**Response 204**

---

### POST /pods/{podSlug}/members/{username}/remove

Owner removes an ACCEPTED member (not self).

**Response 204**

---

## Community Chat (Phase 5)

All routes require `Authorization: Bearer <access_token>` (except the WebSocket token query).

REST text chat for skill communities and pods, with live fan-out over WebSocket.

**Rules:**

- Community channel access: enrolled in the skill (`userskill`)
- Pod chat access: ACCEPTED pod membership
- Channel URLs use `channel.slug` (seeded: `general`, `resources`, `projects` for web-development)
- Message content: 1–4000 chars
- Optional reply: `reply_to_message_id` must reference a message in the same channel/pod
- Edit/delete: author only
- List pagination: `?before=<message_id>&limit=<n>` (default limit 50, max 100); results returned oldest → newest

### Live chat WebSocket

`GET /ws/chat?room={room}&token={access_token}`

Room keys:
- `pod:{podSlug}`
- `channel:{skillSlug}:{channelSlug}`

Server events (JSON text frames):
- `{ "type": "message.created", "message": { ... } }`
- `{ "type": "message.updated", "message": { ... } }`
- `{ "type": "message.deleted", "message_id": 12 }`

Writes still go through REST; the socket is for realtime delivery.

---

### GET /skills/{skillSlug}/community

Community detail for an enrolled skill.

**Response 200:**
```json
{
  "slug": "web-development",
  "name": "Web Development Community",
  "description": "..."
}
```

**Errors:** 403 not enrolled · 404 community not found

---

### GET /skills/{skillSlug}/community/channels

List channels in the skill community.

**Response 200:**
```json
[
  { "slug": "general", "name": "General", "description": "..." }
]
```

---

### GET /skills/{skillSlug}/community/channels/{channelSlug}/messages

List channel messages (`before`, `limit` query params).

**Response 200:** array of
```json
{
  "id": 12,
  "author_username": "taggy_tester",
  "author_name": "Taggy",
  "content": "hello",
  "edited_at": null,
  "created_at": "2026-08-10T...",
  "reply_to_message_id": 10,
  "reply_to_content": "earlier note",
  "reply_to_author_username": "alice"
}
```

---

### POST /skills/{skillSlug}/community/channels/{channelSlug}/messages

Send a channel message.

**Body:**
```json
{
  "content": "hello learners",
  "reply_to_message_id": 10
}
```

`reply_to_message_id` is optional.

**Response 201:** message object

---

### GET /pods/{podSlug}/messages

List pod chat messages (ACCEPTED members only). Same pagination as channel list.

---

### POST /pods/{podSlug}/messages

Send a pod chat message.

**Body:** `{ "content": "..." }`

**Response 201:** message object

**Errors:** 403 not an accepted member · 404 pod not found

---

### PATCH /messages/{id}

Edit own message.

**Body:** `{ "content": "updated text" }`

**Response 200:** message object

**Errors:** 403 not author · 404 not found

---

### DELETE /messages/{id}

Delete own message (hard delete).

**Response 204**

---

## Audio Rooms (Phase 6)

All routes require `Authorization: Bearer <access_token>`.

Taggy stores room metadata; **LiveKit** handles media. Join mints a LiveKit access token.

**Rules:**

- Room types: `POD` | `COMMUNITY_CHANNEL`
- Create starts as `ACTIVE` immediately (scheduling out of scope)
- At most one `ACTIVE` room per pod or per community channel (409 otherwise)
- Creator = `HOST`; other joiners = `SPEAKER`
- Pod rooms: ACCEPTED membership · Channel rooms: enrolled in skill
- Join requires `LIVEKIT_URL`, `LIVEKIT_API_KEY`, `LIVEKIT_API_SECRET` (503 if missing)
- API room id = `public_id` (UUID)
- **Empty auto-delete:** ACTIVE rooms with no participants (`left_at` set for all) for **more than 3 minutes** are hard-deleted by a background sweeper (frees the one-ACTIVE slot). Sweeper also calls LiveKit `DeleteRoom` best-effort before DB delete.
- **Host end / shutdown:** Ending a room (or process graceful shutdown) marks Taggy rooms `ENDED`, sets participant `left_at`, and best-effort calls LiveKit `DeleteRoom` so media sessions are forcibly disconnected.
- **Room object fields:** `id`, `room_type`, `title`, `description`, `status`, `host_username`, `livekit_room_name`, `pod_slug` / `skill_slug` / `channel_slug` (context-dependent), `max_participants`, `actual_start_time`, `ended_at`, `created_at`

### POST /pods/{podSlug}/audio-rooms

**Body:**
```json
{
  "title": "Evening study sync",
  "description": "optional",
  "max_participants": 7
}
```

**Response 201:** room object (see fields above)

---

### GET /pods/{podSlug}/audio-rooms

Defaults to `ACTIVE` rooms only. Optional `?status=` override.

**Response 200:** array of room objects

---

### POST /skills/{skillSlug}/community/channels/{channelSlug}/audio-rooms

Same body as pod create. Channel must exist; caller enrolled.

---

### GET /skills/{skillSlug}/community/channels/{channelSlug}/audio-rooms

Defaults to `ACTIVE` rooms only. Optional `?status=` override.

---

### GET /audio-rooms/{roomId}

**Response 200:**
```json
{
  "room": {
    "id": "<uuid>",
    "room_type": "POD",
    "title": "Evening study sync",
    "description": null,
    "status": "ACTIVE",
    "host_username": "taggy_tester",
    "livekit_room_name": "taggy-pod-<uuid>",
    "pod_slug": "daily-grind",
    "max_participants": 7,
    "actual_start_time": "2026-08-10T01:00:00Z",
    "ended_at": null,
    "created_at": "2026-08-10T01:00:00Z"
  },
  "participants": [
    {
      "username": "taggy_tester",
      "name": "Taggy Tester",
      "role": "HOST",
      "joined_at": "2026-08-10T01:00:00Z"
    }
  ]
}
```

---

### POST /audio-rooms/{roomId}/join

Upserts participant and returns LiveKit credentials.

**Response 200:**
```json
{
  "room_id": "<uuid>",
  "livekit_url": "wss://...",
  "livekit_room_name": "taggy-pod-<uuid>",
  "token": "<jwt>",
  "role": "SPEAKER"
}
```

**Errors:** 409 not active / full · 403 no access · 503 livekit not configured

---

### POST /audio-rooms/{roomId}/leave

Marks participant left (sets `left_at` + duration).

**Response 204**

---

### POST /audio-rooms/{roomId}/end

Host only. Deletes the room record (participants cascade) and best-effort LiveKit `DeleteRoom`.

**Response 204**

---

## Notifications

All routes require `Authorization: Bearer <access_token>`. Path `{username}` must be the JWT owner.

In-app notifications are created as side effects (best-effort) when:

- Someone requests to join your pod → `POD_JOIN_REQUEST` (owner)
- Join accepted / rejected → `POD_JOIN_ACCEPTED` / `POD_JOIN_REJECTED` (requester)
- Member removed → `POD_MEMBER_REMOVED`
- Milestone completed → `MILESTONE_COMPLETED`
- Milestone postponed (new deadline) → `MILESTONE_DUE`
- Roadmap version switched → `ROADMAP_UPDATED`
- Skill join welcome → `COMMUNITY_ANNOUNCEMENT`

### GET /users/{username}/notifications

Query: `unread_only=true|false`, `limit` (default 50, max 100).

**Response 200:**
```json
{
  "unread_count": 2,
  "notifications": [
    {
      "id": 1,
      "type": "POD_JOIN_REQUEST",
      "entity_type": "pod",
      "entity_id": 12,
      "title": "Pod join request",
      "body": "alice requested to join web-dev-grinders",
      "is_read": false,
      "read_at": null,
      "created_at": "2026-08-10T..."
    }
  ]
}
```

---

### PATCH /users/{username}/notifications/{id}/read

**Response 200:** notification object · **409** already read · **404** not found

---

### POST /users/{username}/notifications/read-all

**Response 200:** `{ "updated": 3 }`

---

### POST /users/{username}/notifications/clear-read

Deletes all **read** notifications for the user (used on login to keep the inbox clean). Unread notifications are kept.

**Response 200:** `{ "deleted": 5 }`

---

## Reporting

Authenticated users can file a report against polymorphic targets. Duplicate OPEN reports for the same reporter + target are rejected. Admins resolve reports offline (no resolve API in V1).

**Target types:** `USER` · `PROPOSAL` · `POD` · `MESSAGE` · `AUDIO_ROOM` · `COMMUNITY_CHANNEL`

`target_id` is the numeric (BIGINT) entity id. Pods and community channels expose `id` in their list/detail responses. Audio rooms expose public UUID as `id` and the reportable BIGINT as `entity_id`.

### POST /reports

**Auth:** JWT required

**Request:**
```json
{
  "target_type": "POD",
  "target_id": 12,
  "reason": "Spam invites / harassment in chat"
}
```

**Response 201:**
```json
{
  "id": 1,
  "target_type": "POD",
  "target_id": 12,
  "reason": "Spam invites / harassment in chat",
  "status": "OPEN",
  "created_at": "2026-08-10T01:00:00Z"
}
```

**Errors:** **400** invalid target/reason or cannot report yourself · **409** open report already exists

---

### GET /users/{username}/reports

**Auth:** JWT + scoped username

**Query:** `limit` (default 50, max 100)

**Response 200:** array of report objects (own reports only). Each object matches create response and may include `resolved_at` when resolved.

---

## Search

### GET /search

**Auth:** JWT required

**Query:**
- `q` (required; alias `query`) — 2–80 chars after trim; `%` / `_` stripped
- `types` (optional; singular alias `type`) — comma-separated: `skills`, `users`, `communities` (default: all)
- `limit` (optional) — per-type limit, default 20, max 50

**Response 200:**
```json
{
  "query": "web",
  "skills": [
    { "id": 1, "name": "Web Development", "slug": "web-development", "description": "..." }
  ],
  "users": [
    { "public_id": "...", "username": "web_learner", "name": "Ada", "profile_picture_url": null, "bio": null }
  ],
  "communities": [
    { "id": 1, "name": "Web Development", "description": null, "skill_slug": "web-development", "skill_name": "Web Development" }
  ]
}
```

Omitted type keys mean that type was not requested. Empty arrays mean no matches.

**Errors:** **400** invalid query or type

---

## Admin + AI catalog requests (Phase 7)

Platform admins (`users.role = ADMIN`). Bootstrapped on API start from `ADMIN_USERNAMES` (comma-separated). Roadmap drafts via any OpenAI-compatible chat API — prefer NVIDIA NIM (`NVIDIA_API_KEY`, `AI_MODEL=nvidia/nemotron-3-super-120b-a12b`, `AI_BASE_URL=https://integrate.api.nvidia.com/v1`). OpenRouter remains a fallback if only `OPENROUTER_API_KEY` is set.

Login / refresh / profile (owner) include `is_admin`.

### GET /skills/similar?q=

**Auth:** JWT

Returns `{ "query", "similar": [{ id, name, slug, description?, score }] }` (trgm similarity ≥ 0.3, top 5).

### POST /skills/requests

**Auth:** JWT · rate-limited

**Body:** `{ "name", "description?", "force?" }`

- If similar skills exist and `force` is false → **200** `{ "requires_confirm": true, "similar": [...], "message" }` (no request created).
- Else generates AI milestones and stores PENDING → **201** `{ "request": { ...draft_milestones... } }`.

**Errors:** **400** invalid name/description · **409** duplicate pending · **503** AI unavailable · **502** AI failed

### GET /users/{username}/skill-requests

**Auth:** JWT scoped

### POST /users/{username}/skill-requests/{id}/cancel

**Auth:** JWT scoped · pending only

### POST /skills/{skillSlug}/roadmap-edit-requests

**Auth:** JWT · must be enrolled · rate-limited

**Body:** `{ "rationale?" }` → **201** request with draft milestones

**Errors:** **403** not enrolled · **409** duplicate pending · **503**/**502** AI

### GET /users/{username}/roadmap-edit-requests

### POST /users/{username}/roadmap-edit-requests/{id}/cancel

### GET /admin/me

**Auth:** JWT + admin

### GET /admin/skill-requests

Lists PENDING skill creation requests (includes draft milestones).

### POST /admin/skill-requests/{id}/approve

Creates skill + roadmap v1 ACTIVE (`generated_by=AI`) + milestones + community + default channels; notifies requester.

### POST /admin/skill-requests/{id}/reject

**Body:** `{ "admin_note?" }`

### GET /admin/roadmap-edit-requests

### POST /admin/roadmap-edit-requests/{id}/approve

Archives previous ACTIVE version, publishes next version from draft; notifies requester + enrolled users (`ROADMAP_UPDATED`).

### POST /admin/roadmap-edit-requests/{id}/reject

**Body:** `{ "admin_note?" }`

---

## Ops notes (non-HTTP)

### Graceful shutdown audio cleanup

On SIGINT/SIGTERM, after Echo stops accepting traffic and before DB close:

1. Mark all `ACTIVE` audio rooms `ENDED` and set `left_at` on open participants
2. Best-effort LiveKit `DeleteRoom` for each previously-active LiveKit room name

Failures in LiveKit cleanup are logged and do not block process exit.
