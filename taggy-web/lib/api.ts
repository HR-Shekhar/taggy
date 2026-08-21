const API_BASE =
  process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") ?? "http://localhost:8080";

export type ApiResult<T = unknown> = {
  ok: boolean;
  status: number;
  data: T;
  path: string;
  method: string;
};

export type TokenPair = {
  access_token: string;
  refresh_token: string;
  username: string;
  is_admin?: boolean;
  subscription?: string;
};

const ACCESS_KEY = "access_token";
const REFRESH_KEY = "refresh_token";
const USERNAME_KEY = "username";
const IS_ADMIN_KEY = "is_admin";
const SUBSCRIPTION_KEY = "subscription";
const AVATAR_KEY = "profile_picture_url";

export function getTokens() {
  if (typeof window === "undefined") {
    return { access: null as string | null, refresh: null as string | null };
  }
  return {
    access: localStorage.getItem(ACCESS_KEY),
    refresh: localStorage.getItem(REFRESH_KEY),
  };
}

export function setTokens(access: string, refresh: string) {
  localStorage.setItem(ACCESS_KEY, access);
  localStorage.setItem(REFRESH_KEY, refresh);
}

export function clearTokens() {
  localStorage.removeItem(ACCESS_KEY);
  localStorage.removeItem(REFRESH_KEY);
}

export function getUsername() {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(USERNAME_KEY);
}

export function setUsername(username: string) {
  localStorage.setItem(USERNAME_KEY, username);
}

export function clearUsername() {
  localStorage.removeItem(USERNAME_KEY);
}

export function getIsAdmin() {
  if (typeof window === "undefined") return false;
  return localStorage.getItem(IS_ADMIN_KEY) === "1";
}

export function setIsAdmin(isAdmin: boolean) {
  localStorage.setItem(IS_ADMIN_KEY, isAdmin ? "1" : "0");
}

export function clearIsAdmin() {
  localStorage.removeItem(IS_ADMIN_KEY);
}

export function getSubscription() {
  if (typeof window === "undefined") return "FREE";
  return localStorage.getItem(SUBSCRIPTION_KEY) || "FREE";
}

export function setSubscription(subscription: string) {
  localStorage.setItem(SUBSCRIPTION_KEY, subscription || "FREE");
}

export function clearSubscription() {
  localStorage.removeItem(SUBSCRIPTION_KEY);
}

export function getAvatarUrl() {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(AVATAR_KEY);
}

export function setAvatarUrl(url: string | null | undefined) {
  if (!url) {
    localStorage.removeItem(AVATAR_KEY);
    return;
  }
  localStorage.setItem(AVATAR_KEY, url);
}

export function clearAvatarUrl() {
  localStorage.removeItem(AVATAR_KEY);
}

export function clearSession() {
  clearTokens();
  clearUsername();
  clearIsAdmin();
  clearSubscription();
  clearAvatarUrl();
}

export function userBasePath(username?: string | null) {
  const u = username ?? getUsername();
  return u ? `/users/${u}` : null;
}

export function googleStartUrl() {
  return `${API_BASE}/auth/google/start`;
}

async function parseBody<T>(response: Response): Promise<T> {
  const text = await response.text();
  if (!text) return undefined as T;
  try {
    return JSON.parse(text) as T;
  } catch {
    return text as T;
  }
}

export async function rawRequest<T = unknown>(
  method: string,
  path: string,
  body?: unknown,
  useAuth = true,
  accessOverride?: string | null
): Promise<ApiResult<T>> {
  const headers: Record<string, string> = {
    Accept: "application/json",
  };
  if (body !== undefined && !(body instanceof FormData)) {
    headers["Content-Type"] = "application/json";
  }
  const access = accessOverride ?? getTokens().access;
  if (useAuth && access) {
    headers.Authorization = `Bearer ${access}`;
  }

  try {
    const response = await fetch(`${API_BASE}${path}`, {
      method,
      headers,
      body:
        body === undefined
          ? undefined
          : body instanceof FormData
            ? body
            : JSON.stringify(body),
    });

    const data = await parseBody<T>(response);
    return {
      ok: response.ok,
      status: response.status,
      data,
      path,
      method,
    };
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Network request failed";
    return {
      ok: false,
      status: 0,
      data: { message } as T,
      path,
      method,
    };
  }
}

let refreshPromise: Promise<boolean> | null = null;

async function tryRefresh(): Promise<boolean> {
  if (refreshPromise) return refreshPromise;
  refreshPromise = (async () => {
    const { refresh } = getTokens();
    if (!refresh) return false;
    const result = await rawRequest<TokenPair>(
      "POST",
      "/auth/refresh",
      { refresh_token: refresh },
      false
    );
    if (!result.ok || !result.data?.access_token) {
      clearSession();
      return false;
    }
    setTokens(result.data.access_token, result.data.refresh_token);
    if (result.data.username) setUsername(result.data.username);
    if (typeof result.data.is_admin === "boolean") {
      setIsAdmin(result.data.is_admin);
    }
    if (typeof result.data.subscription === "string") {
      setSubscription(result.data.subscription);
    }
    return true;
  })().finally(() => {
    refreshPromise = null;
  });
  return refreshPromise;
}

/** Authenticated request with one refresh retry on 401. */
export async function api<T = unknown>(
  method: string,
  path: string,
  body?: unknown,
  useAuth = true
): Promise<ApiResult<T>> {
  const first = await rawRequest<T>(method, path, body, useAuth);
  if (first.status !== 401 || !useAuth) return first;
  const refreshed = await tryRefresh();
  if (!refreshed) return first;
  return rawRequest<T>(method, path, body, true);
}

export function apiErrorMessage(result: ApiResult): string {
  if (result.status === 0) {
    const data = result.data as { message?: string } | string | undefined;
    if (typeof data === "string" && data) return friendlyErrorText(data);
    if (data && typeof data === "object" && data.message) {
      return friendlyErrorText(data.message);
    }
    return "Cannot reach the API. Is the backend running on port 8080?";
  }
  const data = result.data as
    | { message?: string; error?: string }
    | string
    | undefined;
  if (typeof data === "string" && data) return friendlyErrorText(data);
  if (data && typeof data === "object") {
    if (data.message) return friendlyErrorText(data.message);
    if (data.error) return friendlyErrorText(data.error);
  }
  return `Request failed (${result.status})`;
}

/** Sentence-case API error text for display. */
export function friendlyErrorText(raw: string): string {
  const text = raw.trim();
  if (!text) return "Something went wrong";
  return text.charAt(0).toUpperCase() + text.slice(1);
}

export function isFreeSkillLimitError(result: ApiResult): boolean {
  if (result.status !== 409) return false;
  const msg = apiErrorMessage(result).toLowerCase();
  return msg.includes("free users") && msg.includes("one active skill");
}

/** Navigate to Google OAuth. Top-level navigation avoids CORS. */
export function startGoogleOAuth() {
  window.location.assign(googleStartUrl());
}

// --- typed helpers ---

export async function register(input: {
  email: string;
  username: string;
  password: string;
  name?: string;
}) {
  return api<{ username: string; email: string; dev_otp?: string }>(
    "POST",
    "/auth/register",
    input,
    false
  );
}

export async function verifyEmail(email: string, otp: string) {
  return api("POST", "/auth/verify-email", { email, otp }, false);
}

export async function resendVerification(email: string) {
  return api<{ dev_otp?: string } | undefined>(
    "POST",
    "/auth/resend-verification",
    { email },
    false
  );
}

export async function login(email: string, password: string) {
  return api<TokenPair>("POST", "/auth/login", { email, password }, false);
}

export async function logout() {
  const { refresh } = getTokens();
  if (refresh) {
    await api("POST", "/auth/logout", { refresh_token: refresh }, false);
  }
  clearSession();
}

export async function completeGoogleRegistration(input: {
  registration_token: string;
  username: string;
  name?: string;
}) {
  return api<TokenPair>("POST", "/auth/google/complete", input, false);
}

export async function getProfile(username: string) {
  return api<Record<string, unknown>>("GET", `/users/${username}`);
}

export async function updateProfile(
  username: string,
  body: Record<string, unknown>
) {
  return api("PATCH", `/users/${username}`, body);
}

export async function uploadProfilePhoto(username: string, file: File) {
  const form = new FormData();
  form.append("file", file);
  return api<Record<string, unknown>>("POST", `/users/${username}/avatar`, form);
}

export async function listSkills() {
  return api<Array<{ id: number; name: string; slug: string; description?: string }>>(
    "GET",
    "/skills"
  );
}

export async function getSkill(slug: string) {
  return api<{
    skill: { id?: number; name: string; slug: string; description?: string };
    community?: { slug: string; name: string; description?: string };
  }>("GET", `/skills/${slug}`);
}

export async function joinSkill(slug: string) {
  return api("POST", `/skills/${slug}/join`);
}

export type RoadmapVersionSummary = {
  version_number: number;
  status: string;
  generated_by: string;
  is_current: boolean;
  milestone_count: number;
  published_at?: string | null;
  created_at: string;
};

export type RoadmapOverview = {
  skill_slug: string;
  skill_name: string;
  current_version: RoadmapVersionSummary | null;
  versions: RoadmapVersionSummary[];
};

export type RoadmapMilestone = {
  slug: string;
  title: string;
  description?: string | null;
  estimated_hours?: number | null;
  order_index: number;
  difficulty?: string | null;
  chapter?: string | null;
  kind?: string;
};

export type RoadmapVersionDetail = {
  skill_slug: string;
  skill_name: string;
  version_number: number;
  status: string;
  generated_by: string;
  is_current: boolean;
  published_at?: string | null;
  created_at: string;
  milestones: RoadmapMilestone[];
};

export async function getSkillRoadmap(slug: string) {
  return api<RoadmapOverview>("GET", `/skills/${slug}/roadmap`);
}

export async function listRoadmapVersions(slug: string) {
  return api<RoadmapVersionSummary[]>("GET", `/skills/${slug}/roadmap/versions`);
}

export async function getRoadmapVersion(slug: string, versionNumber: number) {
  return api<RoadmapVersionDetail>(
    "GET",
    `/skills/${slug}/roadmap/versions/${versionNumber}`
  );
}

export type MySkill = {
  skill_slug: string;
  skill_name: string;
  status: string;
  started_at: string;
  roadmap_version_number: number;
  roadmap_version_status: string;
  milestone_count: number;
  completed_count: number;
  completion_percent: number;
};

export async function listMySkills(username: string) {
  return api<MySkill[]>("GET", `/users/${username}/skills`);
}

export async function switchRoadmapVersion(
  username: string,
  skillSlug: string,
  versionNumber: number
) {
  return api<MySkill>("PUT", `/users/${username}/skills/${skillSlug}/roadmap-version`, {
    version_number: versionNumber,
  });
}

export async function listMilestones(username: string, skillSlug: string) {
  return api<
    Array<{
      slug: string;
      title: string;
      description?: string;
      order_index: number;
      status: string;
      completed_at?: string | null;
      estimated_hours?: number | null;
      difficulty?: string | null;
      chapter?: string | null;
      kind?: string;
    }>
  >("GET", `/users/${username}/skills/${skillSlug}/milestones`);
}

export async function updateMilestone(
  username: string,
  skillSlug: string,
  milestoneSlug: string,
  action: "COMPLETE" | "POSTPONE",
  postponedUntil?: string
) {
  return api(
    "PATCH",
    `/users/${username}/skills/${skillSlug}/milestones/${milestoneSlug}`,
    {
      action,
      ...(postponedUntil ? { postponed_until: postponedUntil } : {}),
    }
  );
}

export async function logStudySession(
  username: string,
  body: { skill_slug: string; duration_minutes: number; notes?: string }
) {
  return api("POST", `/users/${username}/study-sessions`, body);
}

export async function listStudySessions(username: string) {
  return api("GET", `/users/${username}/study-sessions`);
}

export async function getStreak(username: string) {
  return api<{ current_streak: number; longest_streak: number }>(
    "GET",
    `/users/${username}/streak`
  );
}

export type ProgressSummary = {
  total_minutes: number;
  weekly_minutes: number;
  monthly_minutes: number;
  current_streak: number;
  longest_streak: number;
};

export async function getProgressSummary(username: string) {
  return api<ProgressSummary>("GET", `/users/${username}/progress/summary`);
}

export type MyPod = {
  slug?: string;
  pod_slug?: string;
  name?: string;
  pod_name?: string;
  skill_slug?: string;
  status?: string;
  role?: string;
};

export async function listMyPods(username: string) {
  return api<MyPod[] | { pods?: MyPod[] }>("GET", `/users/${username}/pods`);
}

export async function listPodsBySkill(skillSlug: string) {
  return api("GET", `/skills/${skillSlug}/pods`);
}

export async function createPod(
  skillSlug: string,
  body: { name: string; slug: string; description?: string }
) {
  return api("POST", `/skills/${skillSlug}/pods`, body);
}

export async function getPod(podSlug: string) {
  return api("GET", `/pods/${podSlug}`);
}

export async function joinPod(podSlug: string) {
  return api("POST", `/pods/${podSlug}/join`);
}

export async function leavePod(podSlug: string) {
  return api("POST", `/pods/${podSlug}/leave`);
}

export async function acceptMember(podSlug: string, username: string) {
  return api("POST", `/pods/${podSlug}/members/${username}/accept`);
}

export async function rejectMember(podSlug: string, username: string) {
  return api("POST", `/pods/${podSlug}/members/${username}/reject`);
}

export async function removeMember(podSlug: string, username: string) {
  return api("POST", `/pods/${podSlug}/members/${username}/remove`);
}

export async function setMemberRole(
  podSlug: string,
  username: string,
  role: string
) {
  return api("POST", `/pods/${podSlug}/members/${username}/role`, { role });
}

export async function listPodMessages(podSlug: string, limit = 50) {
  return api("GET", `/pods/${podSlug}/messages?limit=${limit}`);
}

export async function sendPodMessage(
  podSlug: string,
  content: string,
  replyToMessageId?: number | null
) {
  return api("POST", `/pods/${podSlug}/messages`, {
    content,
    ...(replyToMessageId ? { reply_to_message_id: replyToMessageId } : {}),
  });
}

export async function listChannels(skillSlug: string) {
  return api("GET", `/skills/${skillSlug}/community/channels`);
}

export async function listChannelMessages(
  skillSlug: string,
  channelSlug: string,
  limit = 50
) {
  return api(
    "GET",
    `/skills/${skillSlug}/community/channels/${channelSlug}/messages?limit=${limit}`
  );
}

export async function sendChannelMessage(
  skillSlug: string,
  channelSlug: string,
  content: string,
  replyToMessageId?: number | null
) {
  return api(
    "POST",
    `/skills/${skillSlug}/community/channels/${channelSlug}/messages`,
    {
      content,
      ...(replyToMessageId ? { reply_to_message_id: replyToMessageId } : {}),
    }
  );
}

export async function editMessage(messageId: number | string, content: string) {
  return api("PATCH", `/messages/${messageId}`, { content });
}

export async function deleteMessage(messageId: number | string) {
  return api("DELETE", `/messages/${messageId}`);
}

export async function listPodAudioRooms(podSlug: string) {
  return api("GET", `/pods/${podSlug}/audio-rooms?status=ACTIVE`);
}

export async function createPodAudioRoom(podSlug: string, title: string) {
  return api("POST", `/pods/${podSlug}/audio-rooms`, { title });
}

export async function listChannelAudioRooms(
  skillSlug: string,
  channelSlug: string
) {
  return api(
    "GET",
    `/skills/${skillSlug}/community/channels/${channelSlug}/audio-rooms?status=ACTIVE`
  );
}

export async function createChannelAudioRoom(
  skillSlug: string,
  channelSlug: string,
  title: string
) {
  return api(
    "POST",
    `/skills/${skillSlug}/community/channels/${channelSlug}/audio-rooms`,
    { title }
  );
}

export async function getAudioRoom(roomId: string | number) {
  return api<{
    room: {
      id: string;
      entity_id?: number;
      title: string;
      status: string;
      host_username?: string;
      pod_slug?: string | null;
      livekit_room_name?: string;
      max_participants?: number;
    };
    participants: Array<{
      username: string;
      name?: string;
      role?: string;
    }>;
  }>("GET", `/audio-rooms/${roomId}`);
}

export async function joinAudioRoom(roomId: string | number) {
  return api<{
    token: string;
    livekit_url: string;
    livekit_room_name?: string;
    role?: string;
  }>("POST", `/audio-rooms/${roomId}/join`);
}

export async function leaveAudioRoom(roomId: string | number) {
  return api("POST", `/audio-rooms/${roomId}/leave`);
}

export async function endAudioRoom(roomId: string | number) {
  return api("POST", `/audio-rooms/${roomId}/end`);
}

export async function listNotifications(
  username: string,
  unreadOnly = false
) {
  const q = unreadOnly ? "?unread_only=true" : "";
  return api("GET", `/users/${username}/notifications${q}`);
}

export async function markNotificationRead(username: string, id: number) {
  return api("PATCH", `/users/${username}/notifications/${id}/read`);
}

export async function markAllNotificationsRead(username: string) {
  return api("POST", `/users/${username}/notifications/read-all`);
}

export async function clearReadNotifications(username: string) {
  return api("POST", `/users/${username}/notifications/clear-read`);
}

export async function search(q: string, types?: string, limit = 20) {
  const params = new URLSearchParams({ q, limit: String(limit) });
  if (types) params.set("types", types);
  return api("GET", `/search?${params.toString()}`);
}

export async function createReport(body: {
  target_type: string;
  target_id: number;
  reason: string;
}) {
  return api("POST", "/reports", body);
}

export async function listReports(username: string, limit = 50) {
  const q = limit !== 50 ? `?limit=${limit}` : "";
  return api("GET", `/users/${username}/reports${q}`);
}

export type MilestoneDraft = {
  title: string;
  description: string;
  estimated_hours: number;
  order_index: number;
  difficulty: string;
  slug: string;
  chapter?: string;
  kind?: string;
};

export type SimilarSkill = {
  id: number;
  name: string;
  slug: string;
  description?: string | null;
  score: number;
};

export type SkillCreationRequest = {
  id: string;
  name: string;
  slug_candidate: string;
  description?: string | null;
  status: string;
  similar_skills: SimilarSkill[];
  draft_milestones: MilestoneDraft[];
  admin_note?: string | null;
  created_skill_id?: number | null;
  created_at: string;
  updated_at: string;
};

export type RoadmapEditRequest = {
  id: string;
  skill_slug: string;
  skill_name: string;
  rationale?: string | null;
  status: string;
  base_version_number: number;
  draft_milestones: MilestoneDraft[];
  admin_note?: string | null;
  created_version_id?: number | null;
  created_at: string;
  updated_at: string;
};

export async function listSimilarSkills(q: string) {
  const params = new URLSearchParams({ q });
  return api<{ query: string; similar: SimilarSkill[] }>(
    "GET",
    `/skills/similar?${params.toString()}`
  );
}

export async function createSkillRequest(body: {
  name: string;
  description?: string;
  force?: boolean;
}) {
  return api<{
    requires_confirm?: boolean;
    already_exists?: boolean;
    similar?: SimilarSkill[];
    message?: string;
    request?: SkillCreationRequest;
  }>("POST", "/skills/requests", body);
}

export async function listMySkillRequests(username: string) {
  return api<SkillCreationRequest[]>("GET", `/users/${username}/skill-requests`);
}

export async function cancelSkillRequest(username: string, id: string) {
  return api<SkillCreationRequest>(
    "POST",
    `/users/${username}/skill-requests/${id}/cancel`
  );
}

export async function createRoadmapEditRequest(
  skillSlug: string,
  rationale?: string
) {
  return api<RoadmapEditRequest>(
    "POST",
    `/skills/${skillSlug}/roadmap-edit-requests`,
    { rationale: rationale || undefined }
  );
}

export async function listMyRoadmapEditRequests(username: string) {
  return api<RoadmapEditRequest[]>(
    "GET",
    `/users/${username}/roadmap-edit-requests`
  );
}

export async function cancelRoadmapEditRequest(username: string, id: string) {
  return api<RoadmapEditRequest>(
    "POST",
    `/users/${username}/roadmap-edit-requests/${id}/cancel`
  );
}

export async function adminMe() {
  return api<{ public_id: string; username: string; is_admin: boolean }>(
    "GET",
    "/admin/me"
  );
}

export async function adminListSkillRequests() {
  return api<SkillCreationRequest[]>("GET", "/admin/skill-requests");
}

export async function adminApproveSkillRequest(id: string) {
  return api<SkillCreationRequest>("POST", `/admin/skill-requests/${id}/approve`);
}

export async function adminRejectSkillRequest(id: string, adminNote?: string) {
  return api<SkillCreationRequest>("POST", `/admin/skill-requests/${id}/reject`, {
    admin_note: adminNote || undefined,
  });
}

export async function adminListRoadmapEditRequests() {
  return api<RoadmapEditRequest[]>("GET", "/admin/roadmap-edit-requests");
}

export async function adminApproveRoadmapEditRequest(id: string) {
  return api<RoadmapEditRequest>(
    "POST",
    `/admin/roadmap-edit-requests/${id}/approve`
  );
}

export async function adminRejectRoadmapEditRequest(
  id: string,
  adminNote?: string
) {
  return api<RoadmapEditRequest>(
    "POST",
    `/admin/roadmap-edit-requests/${id}/reject`,
    { admin_note: adminNote || undefined }
  );
}

export type PodQuizQuestion = {
  order_index: number;
  difficulty: number;
  prompt: string;
  options: string[];
  topic_title: string;
  answered: boolean;
  is_correct?: boolean | null;
  timed_out?: boolean | null;
  correct_indices?: number[];
  seconds_left?: number | null;
};

export type PodQuiz = {
  id: string;
  status: string;
  topic_count: number;
  correct_count: number;
  score: number;
  topics: string[];
  questions: PodQuizQuestion[];
  created_at: string;
  completed_at?: string | null;
};

export type PodLeaderboardEntry = {
  username: string;
  public_id: string;
  name: string;
  best_score: number;
  topic_count: number;
  rank: number;
};

export type CommunityLeaderboardEntry = {
  pod_slug: string;
  pod_name: string;
  total_score: number;
  member_count: number;
  rank: number;
};

export async function startPodQuiz(podSlug: string) {
  return api<PodQuiz>("POST", `/pods/${podSlug}/quizzes`);
}

export async function getPodQuiz(podSlug: string, quizId: string) {
  return api<PodQuiz>("GET", `/pods/${podSlug}/quizzes/${quizId}`);
}

export async function listMyPodQuizzes(podSlug: string) {
  return api<PodQuiz[]>("GET", `/pods/${podSlug}/quizzes/mine`);
}

export async function startQuizQuestion(
  podSlug: string,
  quizId: string,
  order: number
) {
  return api<{ order_index: number; seconds_left: number }>(
    "POST",
    `/pods/${podSlug}/quizzes/${quizId}/questions/${order}/start`
  );
}

export async function answerQuizQuestion(
  podSlug: string,
  quizId: string,
  order: number,
  selectedIndices: number[]
) {
  return api<{
    is_correct: boolean;
    timed_out: boolean;
    correct_indices: number[];
  }>("POST", `/pods/${podSlug}/quizzes/${quizId}/questions/${order}/answer`, {
    selected_indices: selectedIndices,
  });
}

export async function completePodQuiz(podSlug: string, quizId: string) {
  return api<PodQuiz>("POST", `/pods/${podSlug}/quizzes/${quizId}/complete`);
}

export async function getPodLeaderboard(podSlug: string) {
  return api<PodLeaderboardEntry[]>("GET", `/pods/${podSlug}/leaderboard`);
}

export async function getCommunityLeaderboard(skillSlug: string) {
  return api<CommunityLeaderboardEntry[]>(
    "GET",
    `/skills/${skillSlug}/community/leaderboard`
  );
}
