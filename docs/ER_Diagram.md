# ERD.md

# Taggy - Physical ER Diagram

Legend

```text
PK  = Primary Key
FK  = Foreign Key

1   = One
N   = Many

1 ───── N
```

---

# 1. High Level Domain Diagram

```text
┌──────────────┐
│    User      │
└──────┬───────┘
       │
       │
       ├────────────── Authentication
       │
       ├────────────── Learning Progress
       │
       ├────────────── Communities & Pods
       │
       ├────────────── Audio Rooms
       │
       └────────────── Governance


┌──────────────┐
│    Skill     │
└──────┬───────┘
       │
       ├────────────── Roadmap
       │
       ├────────────── Community
       │
       ├────────────── Pods
       │
       └────────────── Study Sessions
```

---

# 2. Authentication Domain

```text
┌─────────────────────────┐
│ User                    │
├─────────────────────────┤
│ PK id                   │
│ email                   │
│ username                │
│ name                    │
│ premium_status          │
│ email_verified          │
└──────────┬──────────────┘
           │ 1
           │
           │ N
┌──────────▼──────────────┐
│ UserIdentity            │
├─────────────────────────┤
│ PK id                   │
│ FK user_id             │
│ provider               │
│ provider_user_id       │
│ password_hash          │
└─────────────────────────┘
```

---

```text
┌─────────────────────────┐
│ User                    │
└──────────┬──────────────┘
           │ 1
           │
           │ N
┌──────────▼──────────────┐
│ UsernameHistory         │
├─────────────────────────┤
│ PK id                   │
│ FK user_id             │
│ username               │
└─────────────────────────┘
```

---

```text
┌─────────────────────────┐
│ EmailVerification       │
├─────────────────────────┤
│ PK id                   │
│ email                   │
│ username                │
│ password_hash           │
│ verification_token      │
│ expires_at              │
└─────────────────────────┘
```

Standalone temporary table.

---

# 3. Learning Domain

```text
┌─────────────────────────┐
│ Skill                   │
├─────────────────────────┤
│ PK id                   │
│ slug                    │
│ name                    │
└──────────┬──────────────┘
           │ 1
           │
           │ 1
┌──────────▼──────────────┐
│ Roadmap                 │
├─────────────────────────┤
│ PK id                   │
│ FK skill_id            │
│ current_version_id      │
└──────────┬──────────────┘
           │ 1
           │
           │ N
┌──────────▼──────────────┐
│ RoadmapVersion          │
├─────────────────────────┤
│ PK id                   │
│ FK roadmap_id          │
│ version_number          │
│ status                  │
└──────────┬──────────────┘
           │ 1
           │
           │ N
┌──────────▼──────────────┐
│ Milestone               │
├─────────────────────────┤
│ PK id                   │
│ FK roadmap_version_id  │
│ order_index             │
└─────────────────────────┘
```

---

# 4. Learning Progress Domain

```text
┌─────────────────────────┐
│ User                    │
└──────────┬──────────────┘
           │
           │ N
           │
┌──────────▼──────────────┐
│ UserSkill               │
├─────────────────────────┤
│ PK id                   │
│ FK user_id             │
│ FK skill_id            │
│ FK roadmap_version_id  │
│ status                  │
└──────────┬──────────────┘
           │
           │ 1
           │
           │ N
┌──────────▼──────────────┐
│ UserMilestoneProgress   │
├─────────────────────────┤
│ PK id                   │
│ FK user_skill_id       │
│ FK milestone_id        │
│ status                  │
└─────────────────────────┘
```

---

```text
┌─────────────┐         ┌─────────────┐
│ User        │         │ Skill       │
└──────┬──────┘         └──────┬──────┘
       │ N                     │ N
       │                       │
       ▼                       ▼
┌─────────────────────────────┐
│ StudySession                │
├─────────────────────────────┤
│ PK id                       │
│ FK user_id                 │
│ FK skill_id                │
│ duration_minutes           │
│ studied_at                 │
└─────────────────────────────┘
```

---

```text
┌─────────────┐
│ User        │
└──────┬──────┘
       │ 1
       │
       │ 1
┌──────▼──────┐
│ Streak      │
├─────────────┤
│ PK id       │
│ FK user_id  │
│ current     │
│ longest     │
└─────────────┘
```

---

# 5. Community & Pods Domain

```text
┌─────────────┐
│ Skill       │
└──────┬──────┘
       │ 1
       │
       │ 1
┌──────▼────────────┐
│ Community         │
├───────────────────┤
│ PK id             │
│ FK skill_id       │
└──────┬────────────┘
       │ 1
       │
       │ N
┌──────▼────────────┐
│ CommunityChannel  │
├───────────────────┤
│ PK id             │
│ FK community_id   │
└───────────────────┘
```

---

```text
┌─────────────┐
│ Skill       │
└──────┬──────┘
       │ 1
       │
       │ N
┌──────▼────────────┐
│ Pod               │
├───────────────────┤
│ PK id             │
│ public_id UUID    │
│ FK owner_id       │
│ FK skill_id       │
└──────┬────────────┘
       │
       │ 1
       │
       │ N
┌──────▼────────────┐
│ PodMembership     │
├───────────────────┤
│ PK id             │
│ FK pod_id         │
│ FK user_id        │
│ status            │
└───────────────────┘
```

---

```text
User
  │
  │ 1
  │
  │ N
  ▼
Pod

(owner relationship)
```

---

# 6. Audio Domain

```text
┌──────────────┐
│ User         │
└──────┬───────┘
       │ 1
       │
       │ N
┌──────▼────────────┐
│ AudioRoom         │
├───────────────────┤
│ PK id             │
│ public_id UUID    │
│ FK host_id        │
│ room_type         │
│ pod_id            │
│ channel_id        │
└──────┬────────────┘
       │
       │ 1
       │
       │ N
┌──────▼──────────────────┐
│ AudioRoomParticipant    │
├─────────────────────────┤
│ PK id                   │
│ FK audio_room_id       │
│ FK user_id             │
│ role                    │
└─────────────────────────┘
```

---

```text
Pod (1)
   │
   │ N
   ▼
AudioRoom


OR


CommunityChannel (1)
   │
   │ N
   ▼
AudioRoom
```

---

# 7. Governance Domain

```text
┌────────────────────┐
│ RoadmapVersion     │
└─────────┬──────────┘
          │ 1
          │
          │ N
┌─────────▼──────────┐
│ MilestoneProposal  │
├────────────────────┤
│ PK id              │
│ FK roadmap_version │
│ FK proposer_id     │
└─────────┬──────────┘
          │ 1
          │
          │ N
┌─────────▼──────────┐
│ ProposalVote       │
├────────────────────┤
│ PK id              │
│ FK proposal_id     │
│ FK user_id         │
└────────────────────┘
```

---

```text
User (1)
   │
   │ N
   ▼
MilestoneProposal


User (1)
   │
   │ N
   ▼
ProposalVote
```

---

# 8. Moderation Domain

```text
┌─────────────┐
│ User        │
└──────┬──────┘
       │ 1
       │
       │ N
┌──────▼────────────┐
│ Report            │
├───────────────────┤
│ PK id             │
│ reporter_id       │
│ target_type       │
│ target_id         │
│ resolved_by       │
└───────────────────┘
```
---

# 10. Notification Domain

┌─────────────┐
│ User        │
└──────┬──────┘
       │ 1
       │
       │ N
┌──────▼────────────┐
│ Notification      │
├───────────────────┤
│ PK id             │
│ FK user_id        │
│ type              │
│ is_read           │
└───────────────────┘
---

# 10. Notification Domain

┌─────────────┐
│ User        │
└──────┬──────┘
       │ 1
       │
       │ N
┌──────▼────────────┐
│ Notification      │
├───────────────────┤
│ PK id             │
│ FK user_id        │
│ type              │
│ is_read           │
└───────────────────┘

---

# Updated Core Relationship Summary

User
├── UserIdentity
├── UsernameHistory
├── UserSkill
├── StudySession
├── Streak
├── PodMembership
├── AudioRoomParticipant
├── AudioRoom (host)
├── MilestoneProposal
├── ProposalVote
├── Report
├── Message (author)
└── Notification

Skill
├── Roadmap
├── Community
├── Pod
├── UserSkill
└── StudySession

Roadmap
└── RoadmapVersion
├── Milestone
├── UserSkill
└── MilestoneProposal

Community
└── CommunityChannel
├── AudioRoom
└── Message

Pod
├── PodMembership
├── AudioRoom
└── Message

AudioRoom
└── AudioRoomParticipant

MilestoneProposal
└── ProposalVote
