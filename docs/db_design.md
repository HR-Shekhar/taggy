# DATABASE_DESIGN.md (Part 1)

## User

### Purpose

Represents a Taggy user.

### Columns

| Column              | Type         | Constraints            |
| ------------------- | ------------ | ---------------------- |
| id                  | BIGSERIAL    | PK                     |
| email               | VARCHAR(255) | UNIQUE NOT NULL        |
| username            | VARCHAR(30)  | UNIQUE NOT NULL        |
| name                | VARCHAR(255) | NOT NULL               |
| profile_picture_url | TEXT         | NULL                   |
| bio                 | TEXT         | NULL                   |
| premium_status      | BOOLEAN      | NOT NULL DEFAULT FALSE |
| email_verified      | BOOLEAN      | NOT NULL DEFAULT FALSE |
| is_deleted          | BOOLEAN      | NOT NULL DEFAULT FALSE |
| deleted_at          | TIMESTAMPTZ  | NULL                   |
| created_at          | TIMESTAMPTZ  | NOT NULL DEFAULT NOW() |
| updated_at          | TIMESTAMPTZ  | NOT NULL DEFAULT NOW() |

### Constraints

* Username length: 3–30 characters.
* Username regex:

  * `^[a-zA-Z0-9._]{3,30}$`

### Indexes

* UNIQUE(email)
* UNIQUE(username)

### Delete Behavior

* Soft delete.

---

## UserIdentity

### Purpose

Represents authentication providers linked to a user.

### Columns

| Column           | Type         | Constraints            |
| ---------------- | ------------ | ---------------------- |
| id               | BIGSERIAL    | PK                     |
| user_id          | BIGINT       | FK → User(id)          |
| provider         | VARCHAR(50)  | NOT NULL               |
| provider_user_id | VARCHAR(255) | NULL                   |
| password_hash    | TEXT         | NULL                   |
| created_at       | TIMESTAMPTZ  | NOT NULL DEFAULT NOW() |
| updated_at       | TIMESTAMPTZ  | NOT NULL DEFAULT NOW() |

### Provider Values

* LOCAL
* GOOGLE

### Constraints

```text
LOCAL → password_hash required.

GOOGLE → provider_user_id required.
```

* UNIQUE(user_id, provider)

### Delete Behavior

* ON DELETE CASCADE

---

## EmailVerification

### Purpose

Temporary storage for email signups before verification.

### Columns

| Column             | Type         | Constraints            |
| ------------------ | ------------ | ---------------------- |
| id                 | BIGSERIAL    | PK                     |
| email              | VARCHAR(255) | NOT NULL               |
| username           | VARCHAR(30)  | NOT NULL               |
| name               | VARCHAR(255) | NOT NULL               |
| password_hash      | TEXT         | NOT NULL               |
| verification_token | TEXT         | UNIQUE NOT NULL        |
| expires_at         | TIMESTAMPTZ  | NOT NULL               |
| created_at         | TIMESTAMPTZ  | NOT NULL DEFAULT NOW() |

### Constraints

* Verification tokens expire.
* Email and username are NOT reserved until successful verification.

### Delete Behavior

* Hard delete after verification or expiration.

---

## UsernameHistory

### Purpose

Tracks previous usernames and prevents impersonation.

### Columns

| Column     | Type        | Constraints            |
| ---------- | ----------- | ---------------------- |
| id         | BIGSERIAL   | PK                     |
| user_id    | BIGINT      | FK → User(id)          |
| username   | VARCHAR(30) | UNIQUE NOT NULL        |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() |

### Constraints

* Once used, a username can never be reused.

### Delete Behavior

* ON DELETE CASCADE

---

## Skill

### Purpose

Represents a learnable skill.

### Columns

| Column      | Type         | Constraints            |
| ----------- | ------------ | ---------------------- |
| id          | BIGSERIAL    | PK                     |
| name        | VARCHAR(255) | NOT NULL               |
| slug        | VARCHAR(255) | UNIQUE NOT NULL        |
| description | TEXT         | NULL                   |
| is_active   | BOOLEAN      | NOT NULL DEFAULT TRUE  |
| created_at  | TIMESTAMPTZ  | NOT NULL DEFAULT NOW() |
| updated_at  | TIMESTAMPTZ  | NOT NULL DEFAULT NOW() |

### Constraints

* Slug must be unique.

### Delete Behavior

* Soft delete via `is_active`.

---

## Roadmap

### Purpose

Container for roadmap versions.

### Columns

| Column             | Type        | Constraints                  |
| ------------------ | ----------- | ---------------------------- |
| id                 | BIGSERIAL   | PK                           |
| skill_id           | BIGINT      | UNIQUE FK → Skill(id)        |
| current_version_id | BIGINT      | NULL FK → RoadmapVersion(id) |
| created_at         | TIMESTAMPTZ | NOT NULL DEFAULT NOW()       |
| updated_at         | TIMESTAMPTZ | NOT NULL DEFAULT NOW()       |

### Constraints

* One Skill has one Roadmap.

### Delete Behavior

* RESTRICT

---

## RoadmapVersion

### Purpose

Represents a version of a roadmap.

### Columns

| Column         | Type        | Constraints            |
| -------------- | ----------- | ---------------------- |
| id             | BIGSERIAL   | PK                     |
| roadmap_id     | BIGINT      | FK → Roadmap(id)       |
| version_number | INTEGER     | NOT NULL               |
| status         | VARCHAR(20) | NOT NULL               |
| generated_by   | VARCHAR(20) | NOT NULL               |
| published_at   | TIMESTAMPTZ | NULL                   |
| created_at     | TIMESTAMPTZ | NOT NULL DEFAULT NOW() |

### Status Values

* DRAFT
* ACTIVE
* ARCHIVED

### Generated By Values

* AI
* ADMIN

### Constraints

* UNIQUE(roadmap_id, version_number)

### Delete Behavior

* Archive only.

---

## Milestone

### Purpose

Represents a step in a roadmap.

### Columns

| Column             | Type         | Constraints             |
| ------------------ | ------------ | ----------------------- |
| id                 | BIGSERIAL    | PK                      |
| roadmap_version_id | BIGINT       | FK → RoadmapVersion(id) |
| title              | VARCHAR(255) | NOT NULL                |
| description        | TEXT         | NULL                    |
| estimated_hours    | INTEGER      | NULL                    |
| order_index        | INTEGER      | NOT NULL                |
| difficulty         | VARCHAR(20)  | NULL                    |
| created_at         | TIMESTAMPTZ  | NOT NULL DEFAULT NOW()  |
| updated_at         | TIMESTAMPTZ  | NOT NULL DEFAULT NOW()  |

### Difficulty Values

* BEGINNER
* INTERMEDIATE
* ADVANCED

### Constraints

* CHECK(estimated_hours > 0)
* UNIQUE(roadmap_version_id, order_index)

### Delete Behavior

* RESTRICT

---

## UserSkill

### Purpose

Represents a user's enrollment in a skill.

### Columns

| Column             | Type        | Constraints             |
| ------------------ | ----------- | ----------------------- |
| id                 | BIGSERIAL   | PK                      |
| user_id            | BIGINT      | FK → User(id)           |
| skill_id           | BIGINT      | FK → Skill(id)          |
| roadmap_version_id | BIGINT      | FK → RoadmapVersion(id) |
| status             | VARCHAR(20) | NOT NULL                |
| started_at         | TIMESTAMPTZ | NOT NULL DEFAULT NOW()  |
| completed_at       | TIMESTAMPTZ | NULL                    |

### Status Values

* ACTIVE
* COMPLETED
* PAUSED

### Constraints

* Only one ACTIVE UserSkill per user-skill pair.

### Delete Behavior

* CASCADE from User.
* RESTRICT from Skill.

---

## UserMilestoneProgress

### Purpose

Tracks milestone progress.

### Columns

| Column          | Type        | Constraints        |
| --------------- | ----------- | ------------------ |
| id              | BIGSERIAL   | PK                 |
| user_skill_id   | BIGINT      | FK → UserSkill(id) |
| milestone_id    | BIGINT      | FK → Milestone(id) |
| status          | VARCHAR(20) | NOT NULL           |
| completed_at    | TIMESTAMPTZ | NULL               |
| postponed_until | TIMESTAMPTZ | NULL               |

### Status Values

* NOT_STARTED
* IN_PROGRESS
* COMPLETED
* POSTPONED

### Constraints

* UNIQUE(user_skill_id, milestone_id)

### Delete Behavior

* ON DELETE CASCADE

---

## StudySession

### Purpose

Source of truth for study activity.

### Columns

| Column           | Type        | Constraints            |
| ---------------- | ----------- | ---------------------- |
| id               | BIGSERIAL   | PK                     |
| user_id          | BIGINT      | FK → User(id)          |
| skill_id         | BIGINT      | FK → Skill(id)         |
| duration_minutes | INTEGER     | NOT NULL               |
| notes            | TEXT        | NULL                   |
| studied_at       | TIMESTAMPTZ | NOT NULL               |
| created_at       | TIMESTAMPTZ | NOT NULL DEFAULT NOW() |

### Constraints

* CHECK(duration_minutes > 0)

### Delete Behavior

* ON DELETE CASCADE from User.

---

## Streak

### Purpose

Cached aggregate of study consistency.

### Columns

| Column             | Type        | Constraints            |
| ------------------ | ----------- | ---------------------- |
| id                 | BIGSERIAL   | PK                     |
| user_id            | BIGINT      | UNIQUE FK → User(id)   |
| current_streak     | INTEGER     | NOT NULL DEFAULT 0     |
| longest_streak     | INTEGER     | NOT NULL DEFAULT 0     |
| last_activity_date | DATE        | NULL                   |
| freeze_count       | INTEGER     | NOT NULL DEFAULT 0     |
| updated_at         | TIMESTAMPTZ | NOT NULL DEFAULT NOW() |

### Constraints

* CHECK(current_streak >= 0)
* CHECK(longest_streak >= 0)
* CHECK(freeze_count >= 0)

### Delete Behavior

* ON DELETE CASCADE

# DATABASE_DESIGN.md (Part 2)

## Community

### Purpose

Represents the official community for a skill.

Each skill has exactly one official community.

---

### Columns

| Column      | Type         | Constraints                    |
| ----------- | ------------ | ------------------------------ |
| id          | BIGSERIAL    | PK                             |
| skill_id    | BIGINT       | UNIQUE NOT NULL FK → Skill(id) |
| name        | VARCHAR(255) | NOT NULL                       |
| description | TEXT         | NULL                           |
| created_at  | TIMESTAMPTZ  | NOT NULL DEFAULT NOW()         |
| updated_at  | TIMESTAMPTZ  | NOT NULL DEFAULT NOW()         |

---

### Constraints

```text
One Skill can have only one Community.
```

---

### Relationships

```text
Community (1)
    ↓
CommunityChannel (Many)
```

---

### Delete Behavior

```text
RESTRICT
```

Communities should never disappear accidentally.

---

## CommunityChannel

### Purpose

Represents channels inside a community.

Examples:

```text
General
Resources
Projects
Placements
Mock Interviews
```

---

### Columns

| Column       | Type         | Constraints                 |
| ------------ | ------------ | --------------------------- |
| id           | BIGSERIAL    | PK                          |
| community_id | BIGINT       | NOT NULL FK → Community(id) |
| name         | VARCHAR(255) | NOT NULL                    |
| description  | TEXT         | NULL                        |
| created_at   | TIMESTAMPTZ  | NOT NULL DEFAULT NOW()      |
| updated_at   | TIMESTAMPTZ  | NOT NULL DEFAULT NOW()      |

---

### Constraints

```text
UNIQUE(community_id, name)
```

Channel names only need to be unique within a community.

---

### Relationships

```text
CommunityChannel (1)
    ↓
AudioRoom (Many)
```

---

### Delete Behavior

```text
RESTRICT
```

Avoid accidental deletion of channels with discussions/history.

---

## Pod

### Purpose

Small accountability groups for learners.

Pods are skill-specific.

---

### Columns

| Column      | Type         | Constraints             |
| ----------- | ------------ | ----------------------- |
| id          | BIGSERIAL    | PK                      |
| public_id   | UUID         | UNIQUE NOT NULL         |
| name        | VARCHAR(255) | NOT NULL                |
| description | TEXT         | NULL                    |
| owner_id    | BIGINT       | NOT NULL FK → User(id)  |
| skill_id    | BIGINT       | NOT NULL FK → Skill(id) |
| max_members | INTEGER      | NOT NULL DEFAULT 7      |
| created_at  | TIMESTAMPTZ  | NOT NULL DEFAULT NOW()  |
| updated_at  | TIMESTAMPTZ  | NOT NULL DEFAULT NOW()  |

---

### Constraints

```text
CHECK (
    max_members > 0
    AND max_members <= 7
)
```

---

### Relationships

```text
User (1)
    ↓
Pod (Many)

Skill (1)
    ↓
Pod (Many)

Pod (1)
    ↓
PodMembership (Many)

Pod (1)
    ↓
AudioRoom (Many)
```

---

### Delete Behavior

```text
RESTRICT
```

Pods should generally not disappear.

If deletion is supported later:

```text
Use soft delete.
```

---

## PodMembership

### Purpose

Tracks pod membership.

---

### Columns

| Column     | Type        | Constraints            |
| ---------- | ----------- | ---------------------- |
| id         | BIGSERIAL   | PK                     |
| pod_id     | BIGINT      | NOT NULL FK → Pod(id)  |
| user_id    | BIGINT      | NOT NULL FK → User(id) |
| status     | VARCHAR(20) | NOT NULL               |
| joined_at  | TIMESTAMPTZ | NULL                   |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() |

---

### Status Values

```text
PENDING
ACCEPTED
REJECTED
LEFT
REMOVED
```

---

### Constraints

```text
UNIQUE(pod_id, user_id)
```

---

### Relationships

```text
Pod (1)
    ↓
PodMembership (Many)

User (1)
    ↓
PodMembership (Many)
```

---

### Delete Behavior

```text
ON DELETE CASCADE
```

If a pod truly disappears, memberships disappear too.

---

## AudioRoom

### Purpose

Metadata for audio rooms.

LiveKit handles the actual audio.

Taggy stores only room information.

---

### Columns

| Column               | Type         | Constraints                    |
| -------------------- | ------------ | ------------------------------ |
| id                   | BIGSERIAL    | PK                             |
| public_id            | UUID         | UNIQUE NOT NULL                |
| room_type            | VARCHAR(50)  | NOT NULL                       |
| pod_id               | BIGINT       | NULL FK → Pod(id)              |
| community_channel_id | BIGINT       | NULL FK → CommunityChannel(id) |
| host_id              | BIGINT       | NOT NULL FK → User(id)         |
| title                | VARCHAR(255) | NOT NULL                       |
| description          | TEXT         | NULL                           |
| livekit_room_name    | VARCHAR(255) | UNIQUE NOT NULL                |
| status               | VARCHAR(50)  | NOT NULL                       |
| scheduled_start_time | TIMESTAMPTZ  | NULL                           |
| actual_start_time    | TIMESTAMPTZ  | NULL                           |
| ended_at             | TIMESTAMPTZ  | NULL                           |
| max_participants     | INTEGER      | NULL                           |
| created_at           | TIMESTAMPTZ  | NOT NULL DEFAULT NOW()         |
| updated_at           | TIMESTAMPTZ  | NOT NULL DEFAULT NOW()         |

---

### Room Types

```text
POD
COMMUNITY_CHANNEL
```

---

### Status Values

```text
SCHEDULED
ACTIVE
ENDED
CANCELLED
```

---

### Constraints

```text
CHECK (
    (
        room_type = 'POD'
        AND pod_id IS NOT NULL
        AND community_channel_id IS NULL
    )
    OR
    (
        room_type = 'COMMUNITY_CHANNEL'
        AND community_channel_id IS NOT NULL
        AND pod_id IS NULL
    )
)
```

---

### Relationships

```text
User (1)
    ↓
AudioRoom (Many)

Pod (1)
    ↓
AudioRoom (Many)

CommunityChannel (1)
    ↓
AudioRoom (Many)

AudioRoom (1)
    ↓
AudioRoomParticipant (Many)
```

---

### Delete Behavior

```text
RESTRICT
```

Audio rooms represent historical events.

---

## AudioRoomParticipant

### Purpose

Tracks participation in audio rooms.

---

### Columns

| Column           | Type        | Constraints                 |
| ---------------- | ----------- | --------------------------- |
| id               | BIGSERIAL   | PK                          |
| audio_room_id    | BIGINT      | NOT NULL FK → AudioRoom(id) |
| user_id          | BIGINT      | NOT NULL FK → User(id)      |
| joined_at        | TIMESTAMPTZ | NOT NULL                    |
| left_at          | TIMESTAMPTZ | NULL                        |
| duration_seconds | INTEGER     | NULL                        |
| role             | VARCHAR(20) | NOT NULL                    |

---

### Role Values

```text
HOST
SPEAKER
LISTENER
```

---

### Constraints

```text
CHECK (
    duration_seconds IS NULL
    OR duration_seconds >= 0
)
```

```text
UNIQUE(audio_room_id, user_id)
```

---

### Relationships

```text
AudioRoom (1)
    ↓
AudioRoomParticipant (Many)

User (1)
    ↓
AudioRoomParticipant (Many)
```

---

### Delete Behavior

```text
ON DELETE CASCADE
```

If an audio room is removed, participation records disappear too.

# DATABASE_DESIGN.md (Part 3)

## MilestoneProposal

### Purpose

Represents community proposals to improve a roadmap.

These are future features and will not be implemented in V1.

---

### Columns

| Column             | Type         | Constraints                      |
| ------------------ | ------------ | -------------------------------- |
| id                 | BIGSERIAL    | PK                               |
| roadmap_version_id | BIGINT       | NOT NULL FK → RoadmapVersion(id) |
| proposer_id        | BIGINT       | NOT NULL FK → User(id)           |
| proposal_type      | VARCHAR(20)  | NOT NULL                         |
| title              | VARCHAR(255) | NOT NULL                         |
| description        | TEXT         | NULL                             |
| status             | VARCHAR(20)  | NOT NULL DEFAULT 'PENDING'       |
| reviewed_at        | TIMESTAMPTZ  | NULL                             |
| reviewed_by        | BIGINT       | NULL FK → User(id)               |
| created_at         | TIMESTAMPTZ  | NOT NULL DEFAULT NOW()           |
| updated_at         | TIMESTAMPTZ  | NOT NULL DEFAULT NOW()           |

---

### Proposal Types

```text id="p1h6c5"
ADD
REMOVE
EDIT
REORDER
MERGE
SPLIT
```

---

### Status Values

```text id="sxql5v"
PENDING
APPROVED
REJECTED
IMPLEMENTED
```

---

### Relationships

```text id="xur5a4"
RoadmapVersion (1)
    ↓
MilestoneProposal (Many)

User (1)
    ↓
MilestoneProposal (Many)

User (1)
    ↓
MilestoneProposal (Many)
(as reviewer)
```

---

### Constraints

```text id="ecpk1v"
reviewed_at IS NOT NULL
ONLY IF reviewed_by IS NOT NULL
```

```text id="clt4vn"
APPROVED or REJECTED proposals
must have reviewed_by.
```

---

### Delete Behavior

```
text id="ubknzc"
roadmap_version_id → RESTRICT

proposer_id → RESTRICT

reviewed_by → SET NULL
```

Historical proposals should remain.

---

### Business Rules

```text id="97t5ij"
Only eligible users may create proposals.

Eligibility:
✓ Member of skill community
AND
(
    10+ study hours logged
    OR
    3+ completed milestones
)
```

---

## ProposalVote

### Purpose

Tracks votes on milestone proposals.

Future feature.

---

### Columns

| Column      | Type        | Constraints                         |
| ----------- | ----------- | ----------------------------------- |
| id          | BIGSERIAL   | PK                                  |
| proposal_id | BIGINT      | NOT NULL FK → MilestoneProposal(id) |
| user_id     | BIGINT      | NOT NULL FK → User(id)              |
| vote_type   | VARCHAR(20) | NOT NULL                            |
| created_at  | TIMESTAMPTZ | NOT NULL DEFAULT NOW()              |

---

### Vote Types

```text id="4tr3yq"
UPVOTE
DOWNVOTE
```

---

### Relationships

```text id="lymopv"
MilestoneProposal (1)
    ↓
ProposalVote (Many)

User (1)
    ↓
ProposalVote (Many)
```

---

### Constraints

```text id="gnd5nk"
UNIQUE(proposal_id, user_id)
```

One user can vote once per proposal.

---

### Delete Behavior

```text id="2n6s1o"
proposal_id → CASCADE

user_id → RESTRICT
```

If a proposal is removed, its votes disappear.

---

### Business Rules

```text id="c2hz3e"
Only eligible users may vote.

Eligibility:
✓ Member of skill community
AND
(
    10+ study hours logged
    OR
    3+ completed milestones
)
```

---

## Report

### Purpose

Supports moderation.

Allows users to report inappropriate behavior or content.

---

### Columns

| Column      | Type        | Constraints             |
| ----------- | ----------- | ----------------------- |
| id          | BIGSERIAL   | PK                      |
| reporter_id | BIGINT      | NOT NULL FK → User(id)  |
| target_type | VARCHAR(50) | NOT NULL                |
| target_id   | BIGINT      | NOT NULL                |
| reason      | TEXT        | NOT NULL                |
| status      | VARCHAR(20) | NOT NULL DEFAULT 'OPEN' |
| resolved_at | TIMESTAMPTZ | NULL                    |
| resolved_by | BIGINT      | NULL FK → User(id)      |
| created_at  | TIMESTAMPTZ | NOT NULL DEFAULT NOW()  |
| updated_at  | TIMESTAMPTZ | NOT NULL DEFAULT NOW()  |

---

### Target Types

```text id="9v13sj"
USER
PROPOSAL
AUDIO_ROOM
COMMUNITY_CHANNEL
```

---

### Status Values

```text id="4evf1v"
OPEN
UNDER_REVIEW
RESOLVED
DISMISSED
```

---

### Relationships

```text id="8a2tpg"
User (1)
    ↓
Report (Many)
(as reporter)

User (1)
    ↓
Report (Many)
(as resolver)
```

---

### Constraints

```text id="m15lke"
resolved_at IS NOT NULL
ONLY IF resolved_by IS NOT NULL
```

---

### Delete Behavior

```text id="1cbvyd"
reporter_id → RESTRICT

resolved_by → SET NULL
```

Reports should remain for moderation history.

---

### Business Rules

```text id="s7x8kz"
A user should not be able to submit
duplicate OPEN reports against the same target.
```

Application-level validation recommended.

---


1. Proposal statuses:
   PENDING → APPROVED/REJECTED → IMPLEMENTED.

2. Proposal and vote eligibility:
   Community member
   AND
   (10+ study hours OR 3+ completed milestones).

3. Proposal votes are preserved unless the proposal itself is removed.

4. Reports use a polymorphic target:
   target_type + target_id.

5. Duplicate OPEN reports should be prevented at the application level.



