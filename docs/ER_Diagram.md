1. User
----
id         PK
email      Unique
name       Not null
profile_picture_url
bio
premium_status
email_verified
created_at
updated_at

2. UserIdentity
-------------
id              PK
user_id         FK
provider        NULLABLE
provider_user_id  NULLABLE
password_hash     
created_at
updated_at

3. Skill
-----
id              PK
name            NOT NULL
slug            UNIQUE NOT NULL
description     
is_active
created_at
updated_at

4. Roadmap
--------
id                  PK
skill_id            FK
current_version_id  FK
created_at
updated_at

5. RoadmapVersion
---------------
id                                  PK
roadmap_id                          FK
version_number                      NOT NULL
status (DRAFT, ACTIVE, ARCHIVED)
generated_by (AI / ADMIN)
published_at
created_at

6. Milestone
----------
id                                  PK
roadmap_version_id                  FK
title                               NOT NULL
description                 
estimated_hours
order_index                         NOT NULL UNIQUE
difficulty
created_at
updated_at

7. UserSkill
-----------
id                                      PK
user_id                                 FK
skill_id                                FK
roadmap_version_id                      FK
status (ACTIVE, COMPLETED, PAUSED)
started_at
completed_at

8. UserMilestoneProgress
----------------------
id                                      PK
user_skill_id                           FK
milestone_id                            FK
status (NOT_STARTED, IN_PROGRESS, COMPLETED, POSTPONED)
completed_at
postponed_until

9. StudySession
--------------
id                          PK
user_id                     FK
skill_id                    FK
duration_minutes
notes
studied_at
created_at

10. Streak
--------
id                          PK
user_id                     FK
current_streak              NOT NULL
longest_streak              NOT NULL
last_activity_date
freeze_count                NOT NULL
updated_at

11. Pod
----
id                      PK
name                    NOT NULL
description             
owner_id                FK
skill_id                FK
max_members
created_at
updated_at

12. PodMembership
---------------
id                      PK
pod_id                  FK
user_id                 FK
status (PENDING, ACCEPTED, REJECTED)
joined_at

13. Community
-----------
id                  PK
skill_id            FK
name                NOTT NULL
description
created_at

14. CommunityChannel
-----------------
id                  PK
community_id        FK
name                NOT NULL
description
created_at

15. MilestoneProposal
------------------
id                      PK
roadmap_version_id      FK
proposer_id             FK
proposal_type           
title                   NOT NULL
description
status
created_at
reviewed_at
reviewed_by

Proposal types:
ADD
REMOVE
EDIT
REORDER
MERGE
SPLIT

16. ProposalVote
--------------
id                  PK
proposal_id         FK
user_id             FK
vote_type
created_at

Vote types:
UPVOTE
DOWNVOTE

17. Report
--------
id                  PK
reporter_id         FK
target_type
target_id           FK
reason
status
created_at
resolved_at
resolved_by

18. AudioRoom
-----------
id (UUID)                           PK
room_type(POD, COMMUNITY_CHANNEL)   
pod_id (nullable)                   FK
community_channel_id (nullable)     FK
host_id                             FK
title                               NOT NULL
description
livekit_room_name
status(SCHEDULED, ACTIVE, ENDED, CANCELLED)
scheduled_start_time (nullable)
actual_start_time (nullable)
ended_at (nullable)
max_participants
created_at
updated_at

**Constraints**
IF room_type = POD
THEN pod_id IS NOT NULL

IF room_type = COMMUNITY_CHANNEL
THEN community_channel_id IS NOT NULL

19. AudioRoomParticipant
----------------------
id                              PK
audio_room_id                   FK
user_id                         FK
joined_at
left_at
duration_seconds
role(HOST, SPEAKER, LISTENER)



User
├──< UserIdentity
├──< UserSkill >──────── Skill
├──< StudySession >───── Skill
├─── (1:1) Streak
├──< PodMembership >──── Pod >──── Skill
├──< ProposalVote >───── MilestoneProposal
├──< AudioRoomParticipant >──── AudioRoom
├──< Report
├──< hosts >──────────── AudioRoom
└──< MilestoneProposal


Skill
├──── (1:1) Roadmap
├──── (1:1) Community
├────< Pod
└────< UserSkill


Roadmap
└──< RoadmapVersion
        │
        ├──< Milestone
        │       │
        │       └──< UserMilestoneProgress >── UserSkill
        │
        └──< MilestoneProposal
                    │
                    └──< ProposalVote


Community
└──< CommunityChannel
            │
            └──< AudioRoom


Pod
├──< PodMembership
└──< AudioRoom


AudioRoom
├──> Pod (nullable)
├──> CommunityChannel (nullable)
├──> User (host)
└──< AudioRoomParticipant >── User