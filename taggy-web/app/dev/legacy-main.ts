import {
  clearTokens,
  clearUsername,
  getTokens,
  getUsername,
  request,
  setTokens,
  setUsername,
  userBasePath,
  type ApiResult,
} from "./legacy-api";
import { Room, RoomEvent, Track } from "livekit-client";

type SectionId =
  | "auth"
  | "profile"
  | "skills"
  | "milestones"
  | "progress"
  | "pods"
  | "community"
  | "audio"
  | "notifications"
  | "reports"
  | "search";

const sections: { id: SectionId; label: string }[] = [
  { id: "auth", label: "Auth" },
  { id: "profile", label: "Profile" },
  { id: "skills", label: "Skills" },
  { id: "milestones", label: "Milestones" },
  { id: "progress", label: "Progress" },
  { id: "pods", label: "Pods" },
  { id: "community", label: "Community" },
  { id: "audio", label: "Audio" },
  { id: "notifications", label: "Notifications" },
  { id: "reports", label: "Reports" },
  { id: "search", label: "Search" },
];

let activeSection: SectionId = "auth";
let cachedSkills: Array<{ id: number; name: string; slug: string }> = [];
let cachedMySkills: Array<{ skill_slug: string; skill_name: string }> = [];
let livekitRoom: Room | null = null;
let livekitConnectedRoomId: string | null = null;

let nav: HTMLElement;
let main: HTMLElement;
let logEntries: HTMLElement;

function consumeOAuthCallback(): {
  error?: string;
  loggedIn?: boolean;
  pendingGoogle?: {
    registration_token: string;
    email: string;
    name: string;
    picture: string;
  };
} {
  const params = new URLSearchParams(window.location.search);
  const error = params.get("error");
  if (error) {
    window.history.replaceState({}, "", "/");
    return { error };
  }

  const hash = window.location.hash.startsWith("#")
    ? window.location.hash.slice(1)
    : window.location.hash;
  const onCallback =
    window.location.pathname === "/auth/callback" ||
    window.location.pathname === "/auth/complete-google";
  if (!hash && !onCallback) {
    return {};
  }

  const tokenParams = new URLSearchParams(hash);
  const access = tokenParams.get("access_token");
  const refresh = tokenParams.get("refresh_token");
  const username = tokenParams.get("username");
  const registrationToken = tokenParams.get("registration_token");

  if (registrationToken) {
    const pending = {
      registration_token: registrationToken,
      email: tokenParams.get("email") ?? "",
      name: tokenParams.get("name") ?? "",
      picture: tokenParams.get("picture") ?? "",
    };
    sessionStorage.setItem("pending_google", JSON.stringify(pending));
    window.history.replaceState({}, "", "/");
    return { pendingGoogle: pending };
  }

  if (access && refresh) {
    setTokens(access, refresh);
    if (username) setUsername(username);
    sessionStorage.removeItem("pending_google");
    window.history.replaceState({}, "", "/");
    return { loggedIn: true };
  }

  if (onCallback) {
    window.history.replaceState({}, "", "/");
  }
  return {};
}

function getPendingGoogle(): {
  registration_token: string;
  email: string;
  name: string;
  picture: string;
} | null {
  const raw = sessionStorage.getItem("pending_google");
  if (!raw) return null;
  try {
    return JSON.parse(raw) as {
      registration_token: string;
      email: string;
      name: string;
      picture: string;
    };
  } catch {
    return null;
  }
}

function clearPendingGoogle() {
  sessionStorage.removeItem("pending_google");
}

let oauthCallback: ReturnType<typeof consumeOAuthCallback> = {};

function el<K extends keyof HTMLElementTagNameMap>(
  tag: K,
  className?: string,
  text?: string
): HTMLElementTagNameMap[K] {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text) node.textContent = text;
  return node;
}

function field(labelText: string, input: HTMLElement): HTMLElement {
  const label = el("label");
  label.append(el("span", undefined, labelText), input);
  return label;
}

function textInput(name: string, value = "", type = "text"): HTMLInputElement {
  const input = document.createElement("input");
  input.name = name;
  input.type = type;
  input.value = value;
  return input;
}

function passwordInput(name: string, value = ""): HTMLElement {
  const wrap = el("div", "password-field");
  const input = textInput(name, value, "password");
  const toggle = el("button", "password-toggle", "Show");
  toggle.type = "button";
  toggle.setAttribute("aria-label", "Show password");
  toggle.addEventListener("click", () => {
    const visible = input.type === "text";
    input.type = visible ? "password" : "text";
    toggle.textContent = visible ? "Show" : "Hide";
    toggle.setAttribute("aria-label", visible ? "Show password" : "Hide password");
  });
  wrap.append(input, toggle);
  return wrap;
}

function logResult(result: ApiResult) {
  const entry = el("div", "log-entry");
  const header = el(
    "header",
    result.ok ? "ok" : "fail",
    `${result.method} ${result.path} → ${result.status}`
  );
  const pre = el("pre");
  pre.textContent = JSON.stringify(result.data, null, 2);
  entry.append(header, pre);
  logEntries.prepend(entry);
}

async function run<T>(
  method: string,
  path: string,
  body?: unknown,
  useAuth = true
): Promise<ApiResult<T>> {
  const result = await request<T>(method, path, body, useAuth);
  logResult(result);
  return result;
}

function renderNav() {
  nav.innerHTML = "";
  nav.append(el("h1", undefined, "Taggy Tester"));

  for (const s of sections) {
    const btn = el(
      "button",
      s.id === activeSection ? "active" : undefined,
      s.label
    );
    btn.addEventListener("click", () => {
      activeSection = s.id;
      renderNav();
      renderMain();
    });
    nav.append(btn);
  }

  const tokens = getTokens();
  const tokenBar = el("div", "token-bar");
  tokenBar.innerHTML = `<strong>User</strong>: ${getUsername() ?? "none"}<br><strong>Access</strong>: ${tokens.access ? "set" : "none"}<br><strong>Refresh</strong>: ${tokens.refresh ? "set" : "none"}`;
  nav.append(tokenBar);
}

function card(title: string, content: HTMLElement): HTMLElement {
  const c = el("div", "card");
  c.append(el("h3", undefined, title), content);
  return c;
}

function bindForm(
  form: HTMLFormElement,
  handler: (data: Record<string, string>) => Promise<void>
) {
  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    const fd = new FormData(form);
    const data: Record<string, string> = {};
    fd.forEach((v, k) => {
      if (typeof v === "string" && v !== "") data[k] = v;
    });
    await handler(data);
  });
}

function renderAuth(): HTMLElement {
  const wrap = el("div", "section active");
  wrap.append(el("h2", undefined, "Auth"));

  wrap.append(
    el(
      "p",
      "hint",
      "1. Register → copy dev_otp from the response log below. 2. Verify email. 3. Login."
    )
  );

  const verifyEmailInput = textInput("email", "tester@taggy.dev");
  const verifyOtpInput = textInput("otp", "", "text");
  verifyOtpInput.maxLength = 6;
  verifyOtpInput.pattern = "[0-9]{6}";

  const regForm = el("form");
  const regRow = el("div", "row");
  regRow.append(
    field("Email", textInput("email", "tester@taggy.dev")),
    field("Username", textInput("username", "taggy_tester")),
    field("Name", textInput("name", "API Tester")),
    field("Password", passwordInput("password", "Password1!"))
  );
  const regBtn = el("button", "action", "Register");
  regBtn.type = "submit";
  regForm.append(regRow, regBtn);
  bindForm(regForm as HTMLFormElement, async (d) => {
    const res = await run<{
      username: string;
      email: string;
      dev_otp?: string;
    }>("POST", "/auth/register", {
      email: d.email,
      username: d.username,
      name: d.name,
      password: d.password,
    }, false);
    if (res.ok && res.data) {
      if (res.data.username) setUsername(res.data.username);
      if (res.data.email) verifyEmailInput.value = res.data.email;
      if (res.data.dev_otp) verifyOtpInput.value = res.data.dev_otp;
      renderNav();
    } else if (d.username) {
      setUsername(d.username);
      renderNav();
    }
  });

  const verifyForm = el("form");
  const verifyRow = el("div", "row");
  verifyRow.append(
    field("Email", verifyEmailInput),
    field("OTP (6 digits)", verifyOtpInput)
  );
  const verifyBtn = el("button", "action", "Verify email");
  verifyBtn.type = "submit";
  verifyForm.append(verifyRow, verifyBtn);
  bindForm(verifyForm as HTMLFormElement, async (d) => {
    if (!d.otp || d.otp.length !== 6) {
      logResult({
        ok: false,
        status: 0,
        data: { message: "Enter the 6-digit OTP from register or resend response" },
        path: "/auth/verify-email",
        method: "POST",
      });
      return;
    }
    await run("POST", "/auth/verify-email", { email: d.email, otp: d.otp }, false);
  });

  const resendBtn = el("button", "action secondary", "Resend OTP");
  resendBtn.addEventListener("click", async () => {
    const email = verifyEmailInput.value.trim();
    if (!email) return;
    const res = await run<{ dev_otp?: string }>(
      "POST",
      "/auth/resend-verification",
      { email },
      false
    );
    if (res.ok && res.data?.dev_otp) {
      verifyOtpInput.value = res.data.dev_otp;
    }
  });

  const loginForm = el("form");
  const loginRow = el("div", "row");
  loginRow.append(
    field("Email", textInput("email", "tester@taggy.dev")),
    field("Password", passwordInput("password", "Password1!"))
  );
  const loginBtn = el("button", "action", "Login");
  loginBtn.type = "submit";
  loginForm.append(loginRow, loginBtn);
  bindForm(loginForm as HTMLFormElement, async (d) => {
    const res = await run<{ access_token: string; refresh_token: string; username: string }>(
      "POST",
      "/auth/login",
      { email: d.email, password: d.password },
      false
    );
    if (res.ok && res.data) {
      setTokens(res.data.access_token, res.data.refresh_token);
      if (res.data.username) setUsername(res.data.username);
      renderNav();
    }
  });

  const tokenActions = el("div", "row");
  const refreshBtn = el("button", "action secondary", "Refresh token");
  refreshBtn.addEventListener("click", async () => {
    const { refresh } = getTokens();
    if (!refresh) return;
    const res = await run<{ access_token: string; refresh_token: string; username: string }>(
      "POST",
      "/auth/refresh",
      { refresh_token: refresh },
      false
    );
    if (res.ok && res.data) {
      setTokens(res.data.access_token, res.data.refresh_token);
      if (res.data.username) setUsername(res.data.username);
      renderNav();
    }
  });

  const logoutBtn = el("button", "action secondary", "Logout");
  logoutBtn.addEventListener("click", async () => {
    const { refresh } = getTokens();
    if (!refresh) return;
    await run("POST", "/auth/logout", { refresh_token: refresh }, false);
    clearTokens();
    clearUsername();
    renderNav();
  });

  const logoutAllBtn = el("button", "action danger", "Logout all sessions");
  logoutAllBtn.addEventListener("click", async () => {
    await run("POST", "/auth/logout-all");
    clearTokens();
    clearUsername();
    renderNav();
  });

  const healthBtn = el("button", "action secondary", "Health check");
  healthBtn.addEventListener("click", () =>
    run("GET", "/health", undefined, false)
  );

  tokenActions.append(refreshBtn, logoutBtn, logoutAllBtn, healthBtn);

  const googleBtn = el("button", "action", "Continue with Google");
  googleBtn.addEventListener("click", async () => {
    const res = await run<{ url: string }>("GET", "/auth/google/start", undefined, false);
    if (res.ok && res.data?.url) {
      window.location.href = res.data.url;
    }
  });

  const pending = getPendingGoogle() ?? oauthCallback.pendingGoogle ?? null;
  const completeWrap = el("div");
  if (pending) {
    completeWrap.append(
      el(
        "p",
        "hint",
        `Google verified ${pending.email}. Choose a username to finish signup. Closing without this creates no account.`
      )
    );
    const completeForm = el("form");
    const completeRow = el("div", "row");
    completeRow.append(
      field("Username", textInput("username", "")),
      field("Name", textInput("name", pending.name || "")),
      field("Email (from Google)", textInput("email", pending.email, "email"))
    );
    const emailInput = completeRow.querySelector('input[name="email"]') as HTMLInputElement;
    emailInput.readOnly = true;
    const completeBtn = el("button", "action", "Complete Google signup");
    completeBtn.type = "submit";
    const cancelBtn = el("button", "action secondary", "Cancel (discard)");
    cancelBtn.type = "button";
    cancelBtn.addEventListener("click", () => {
      clearPendingGoogle();
      renderNav();
      renderMain();
    });
    completeForm.append(completeRow, completeBtn, cancelBtn);
    bindForm(completeForm as HTMLFormElement, async (d) => {
      if (!d.username) {
        logResult({
          ok: false,
          status: 0,
          data: { message: "username is required" },
          path: "/auth/google/complete",
          method: "POST",
        });
        return;
      }
      const res = await run<{
        access_token: string;
        refresh_token: string;
        username: string;
      }>(
        "POST",
        "/auth/google/complete",
        {
          registration_token: pending.registration_token,
          username: d.username,
          name: d.name || undefined,
        },
        false
      );
      if (res.ok && res.data) {
        setTokens(res.data.access_token, res.data.refresh_token);
        if (res.data.username) setUsername(res.data.username);
        clearPendingGoogle();
        renderNav();
        renderMain();
      }
    });
    completeWrap.append(completeForm);
  } else {
    completeWrap.append(
      el("p", "hint", "New Google users are sent here after OAuth to pick a username. Existing users get tokens immediately.")
    );
  }

  const googleWrap = el("div");
  googleWrap.append(googleBtn, completeWrap);

  wrap.append(
    card("1. Register", regForm),
    card("2. Verify email", verifyForm),
    card("Resend OTP", resendBtn),
    card("3. Login", loginForm),
    card("4. Google", googleWrap),
    card("Session", tokenActions)
  );
  return wrap;
}

function renderProfile(): HTMLElement {
  const wrap = el("div", "section active");
  wrap.append(el("h2", undefined, "Profile"));

  wrap.append(
    el(
      "p",
      "hint",
      "GitHub-style: GET /users/{username}. Send your token when viewing yourself to see email and subscription."
    )
  );

  const viewForm = el("form");
  const viewRow = el("div", "row");
  const viewUsername = textInput("username", getUsername() ?? "taggy_tester");
  viewRow.append(field("Username", viewUsername));
  const viewBtn = el("button", "action", "GET profile");
  viewBtn.type = "submit";
  viewForm.append(viewRow, viewBtn);
  bindForm(viewForm as HTMLFormElement, async (d) => {
    const own = getUsername() === d.username;
    await run("GET", `/users/${d.username}`, undefined, own);
  });

  const patchForm = el("form");
  const row = el("div", "row");
  row.append(
    field("Name (optional)", textInput("name")),
    field("Bio (optional)", textInput("bio")),
    field("New username (optional)", textInput("new_username")),
    field("Profile picture URL (optional)", textInput("profile_picture_url"))
  );
  const patchBtn = el("button", "action", "PATCH profile");
  patchBtn.type = "submit";
  patchForm.append(row, patchBtn);
  bindForm(patchForm as HTMLFormElement, async (d) => {
    const base = userBasePath();
    if (!base) return;
    const body: Record<string, string> = {};
    if (d.name) body.name = d.name;
    if (d.bio) body.bio = d.bio;
    if (d.new_username) body.username = d.new_username;
    if (d.profile_picture_url) body.profile_picture_url = d.profile_picture_url;
    const res = await run("PATCH", base, body);
    if (res.ok && d.new_username) {
      setUsername(d.new_username);
      renderNav();
    }
  });

  wrap.append(card("View profile", viewForm), card("Update profile (self only)", patchForm));
  return wrap;
}

async function loadSkillsList(container: HTMLElement) {
  const res = await run<Array<{ id: number; name: string; slug: string }>>(
    "GET",
    "/skills"
  );
  if (!res.ok) return;
  cachedSkills = res.data ?? [];
  container.innerHTML = "";
  if (cachedSkills.length === 0) {
    container.append(el("p", "hint", "No skills found."));
    return;
  }
  const table = el("table", "list-table");
  const headRow = el("tr");
  headRow.append(
    el("th", undefined, "Name"),
    el("th", undefined, "Slug"),
    el("th", undefined, "Actions")
  );
  const thead = el("thead");
  thead.append(headRow);
  const tbody = el("tbody");
  for (const skill of cachedSkills) {
    const tr = el("tr");
    tr.append(
      el("td", undefined, skill.name),
      el("td", undefined, skill.slug)
    );
    const actions = el("td");
    const detailBtn = el("button", "action secondary", "Detail");
    detailBtn.addEventListener("click", () =>
      run("GET", `/skills/${skill.slug}`)
    );
    const joinBtn = el("button", "action", "Join");
    joinBtn.addEventListener("click", async () => {
      await run("POST", `/skills/${skill.slug}/join`);
      await loadMySkills();
    });
    actions.append(detailBtn, joinBtn);
    tr.append(actions);
    tbody.append(tr);
  }
  table.append(thead, tbody);
  container.append(table);
}

function scopedUserPath(suffix = "") {
  const base = userBasePath();
  return base ? `${base}${suffix}` : null;
}

async function loadMySkills() {
  const base = userBasePath();
  if (!base) return;
  const res = await run<
    Array<{ skill_slug: string; skill_name: string; status: string }>
  >("GET", `${base}/skills`);
  if (res.ok && res.data) {
    cachedMySkills = res.data.map((s) => ({
      skill_slug: s.skill_slug,
      skill_name: s.skill_name,
    }));
  }
}

function renderSkills(): HTMLElement {
  const wrap = el("div", "section active");
  wrap.append(el("h2", undefined, "Skills"));
  wrap.append(
    el(
      "p",
      "hint",
      "Load skills, view detail, or join. Free users: max 1 ACTIVE skill."
    )
  );

  const listContainer = el("div");
  const catalogContent = el("div");
  const refreshBtn = el("button", "action", "Load skills");
  refreshBtn.addEventListener("click", () => loadSkillsList(listContainer));
  catalogContent.append(refreshBtn, listContainer);

  const mySkillsBtn = el("button", "action secondary", "My enrollments");
  mySkillsBtn.addEventListener("click", () => loadMySkills());

  const enrollContent = el("div");
  enrollContent.append(mySkillsBtn);

  wrap.append(card("Catalog", catalogContent), card("Enrollments", enrollContent));
  return wrap;
}

function renderMilestones(): HTMLElement {
  const wrap = el("div", "section active");
  wrap.append(el("h2", undefined, "Milestones"));

  const listContainer = el("div");

  const loadForm = el("form");
  const skillSelect = document.createElement("select");
  skillSelect.name = "skill_slug";
  const skillRow = el("div", "row");
  skillRow.append(field("Skill", skillSelect));
  const loadBtn = el("button", "action", "Load milestones");
  loadBtn.type = "submit";
  loadForm.append(skillRow, loadBtn);

  const refreshSkills = () => {
    skillSelect.innerHTML = "";
    if (cachedMySkills.length === 0) {
      skillSelect.append(el("option", undefined, "Join a skill first"));
      return;
    }
    for (const s of cachedMySkills) {
      const opt = document.createElement("option");
      opt.value = s.skill_slug;
      opt.textContent = `${s.skill_name} (${s.skill_slug})`;
      skillSelect.append(opt);
    }
  };

  refreshSkills();
  loadMySkills().then(() => refreshSkills());

  bindForm(loadForm as HTMLFormElement, async (d) => {
    const res = await run<
      Array<{
        slug: string;
        title: string;
        order_index: number;
        status: string;
      }>
    >("GET", `${scopedUserPath(`/skills/${d.skill_slug}/milestones`) ?? ""}`);
    if (!res.ok) return;

    listContainer.innerHTML = "";
    const milestones = res.data ?? [];
    if (milestones.length === 0) {
      listContainer.append(el("p", "hint", "No milestones."));
      return;
    }

    const table = el("table", "list-table");
    const headRow = el("tr");
    headRow.append(
      el("th", undefined, "Slug"),
      el("th", undefined, "Title"),
      el("th", undefined, "Order"),
      el("th", undefined, "Status"),
      el("th", undefined, "Actions")
    );
    const thead = el("thead");
    thead.append(headRow);
    const tbody = el("tbody");

    for (const m of milestones) {
      const tr = el("tr");
      tr.append(
        el("td", undefined, m.slug),
        el("td", undefined, m.title),
        el("td", undefined, String(m.order_index)),
        el("td", undefined, m.status)
      );
      const actions = el("td");
      const completeBtn = el("button", "action", "Complete");
      completeBtn.addEventListener("click", () =>
        run("PATCH", `${scopedUserPath(`/skills/${d.skill_slug}/milestones/${m.slug}`) ?? ""}`, {
          action: "COMPLETE",
        })
      );
      const postponeBtn = el("button", "action secondary", "Postpone +7d");
      postponeBtn.addEventListener("click", () => {
        const until = new Date();
        until.setDate(until.getDate() + 7);
        run("PATCH", `${scopedUserPath(`/skills/${d.skill_slug}/milestones/${m.slug}`) ?? ""}`, {
          action: "POSTPONE",
          postponed_until: until.toISOString(),
        });
      });
      actions.append(completeBtn, postponeBtn);
      tr.append(actions);
      tbody.append(tr);
    }
    table.append(thead, tbody);
    listContainer.append(table);
  });

  wrap.append(
    card("Milestone progress", loadForm),
    card("Results", listContainer)
  );
  return wrap;
}

function renderProgress(): HTMLElement {
  const wrap = el("div", "section active");
  wrap.append(el("h2", undefined, "Progress"));

  const logForm = el("form");
  const skillSelect = document.createElement("select");
  skillSelect.name = "skill_slug";
  const duration = textInput("duration_minutes", "30", "number");
  const notes = textInput("notes");
  const logRow = el("div", "row");
  logRow.append(
    field("Skill", skillSelect),
    field("Duration (minutes)", duration),
    field("Notes (optional)", notes)
  );
  const logBtn = el("button", "action", "Log study session");
  logBtn.type = "submit";
  logForm.append(logRow, logBtn);

  const refreshSkills = () => {
    skillSelect.innerHTML = "";
    if (cachedMySkills.length === 0) {
      skillSelect.append(el("option", undefined, "Join a skill first"));
      return;
    }
    for (const s of cachedMySkills) {
      const opt = document.createElement("option");
      opt.value = s.skill_slug;
      opt.textContent = `${s.skill_name} (${s.skill_slug})`;
      skillSelect.append(opt);
    }
  };
  refreshSkills();
  loadMySkills().then(() => refreshSkills());

  bindForm(logForm as HTMLFormElement, async (d) => {
    const body: Record<string, unknown> = {
      skill_slug: d.skill_slug,
      duration_minutes: Number(d.duration_minutes),
    };
    if (d.notes) body.notes = d.notes;
    const path = scopedUserPath("/study-sessions");
    if (!path) return;
    await run("POST", path, body);
  });

  const sessionsBtn = el("button", "action secondary", "List all sessions");
  sessionsBtn.addEventListener("click", () => {
    const path = scopedUserPath("/study-sessions");
    if (path) run("GET", path);
  });

  const filterForm = el("form");
  const filterSkill = document.createElement("select");
  filterSkill.name = "skill_slug";
  const populateFilterSkills = () => {
    filterSkill.innerHTML = "";
    if (cachedMySkills.length === 0) {
      filterSkill.append(el("option", undefined, "Join a skill first"));
      return;
    }
    for (const s of cachedMySkills) {
      const opt = document.createElement("option");
      opt.value = s.skill_slug;
      opt.textContent = `${s.skill_name} (${s.skill_slug})`;
      filterSkill.append(opt);
    }
  };
  populateFilterSkills();
  loadMySkills().then(() => populateFilterSkills());
  const filterRow = el("div", "row");
  filterRow.append(field("Skill (optional)", filterSkill));
  const filterBtn = el("button", "action secondary", "List by skill");
  filterBtn.type = "submit";
  filterForm.append(filterRow, filterBtn);
  bindForm(filterForm as HTMLFormElement, async (d) => {
    const path = scopedUserPath(`/study-sessions?skill_slug=${d.skill_slug}`);
    if (path) await run("GET", path);
  });

  const streakBtn = el("button", "action", "GET streak");
  streakBtn.addEventListener("click", () => {
    const path = scopedUserPath("/streak");
    if (path) run("GET", path);
  });

  const summaryBtn = el("button", "action", "GET progress summary");
  summaryBtn.addEventListener("click", () => {
    const path = scopedUserPath("/progress/summary");
    if (path) run("GET", path);
  });

  const readRow = el("div", "row");
  readRow.append(streakBtn, summaryBtn, sessionsBtn);

  wrap.append(
    card("Log study session", logForm),
    card("History", filterForm),
    card("Dashboard", readRow)
  );
  return wrap;
}

function renderPods(): HTMLElement {
  const wrap = el("div", "section active");
  wrap.append(el("h2", undefined, "Pods"));
  wrap.append(
    el(
      "p",
      "hint",
      "Must be enrolled in the skill. One ACCEPTED pod per user. Max 7. Roles: OWNER / ADMIN / MEMBER. Owner leave transfers to an admin (or random member); empty pod delete allowed."
    )
  );

  const skillSelect = document.createElement("select");
  skillSelect.name = "skill_slug";
  const refreshSkills = () => {
    skillSelect.innerHTML = "";
    if (cachedMySkills.length === 0) {
      skillSelect.append(el("option", undefined, "Join a skill first"));
      return;
    }
    for (const s of cachedMySkills) {
      const opt = document.createElement("option");
      opt.value = s.skill_slug;
      opt.textContent = `${s.skill_name} (${s.skill_slug})`;
      skillSelect.append(opt);
    }
  };
  refreshSkills();
  loadMySkills().then(() => refreshSkills());

  const createForm = el("form");
  const createRow = el("div", "row");
  createRow.append(
    field("Skill", skillSelect),
    field("Slug (unique)", textInput("slug", "web-dev-grinders")),
    field("Name", textInput("name", "Web Dev Grinders")),
    field("Description (optional)", textInput("description", "Daily accountability"))
  );
  const createBtn = el("button", "action", "Create pod");
  createBtn.type = "submit";
  createForm.append(createRow, createBtn);
  bindForm(createForm as HTMLFormElement, async (d) => {
    await run("POST", `/skills/${d.skill_slug}/pods`, {
      slug: d.slug,
      name: d.name,
      description: d.description || undefined,
    });
  });

  const listBtn = el("button", "action secondary", "List pods for skill");
  listBtn.addEventListener("click", async () => {
    const slug = skillSelect.value;
    if (!slug) return;
    await run("GET", `/skills/${slug}/pods`);
  });

  const myPodsBtn = el("button", "action secondary", "My pods");
  myPodsBtn.addEventListener("click", () => {
    const path = scopedUserPath("/pods");
    if (path) run("GET", path);
  });

  const detailForm = el("form");
  const detailRow = el("div", "row");
  detailRow.append(field("Pod slug", textInput("pod_slug", "")));
  const detailBtn = el("button", "action", "Get pod");
  detailBtn.type = "submit";
  detailForm.append(detailRow, detailBtn);
  bindForm(detailForm as HTMLFormElement, async (d) => {
    await run("GET", `/pods/${d.pod_slug}`);
  });

  const joinForm = el("form");
  const joinRow = el("div", "row");
  joinRow.append(field("Pod slug", textInput("pod_slug", "")));
  const joinBtn = el("button", "action", "Request join");
  joinBtn.type = "submit";
  joinForm.append(joinRow, joinBtn);
  bindForm(joinForm as HTMLFormElement, async (d) => {
    await run("POST", `/pods/${d.pod_slug}/join`);
  });

  const decideForm = el("form");
  const decideRow = el("div", "row");
  decideRow.append(
    field("Pod slug", textInput("pod_slug", "")),
    field("Member username", textInput("username", ""))
  );
  const acceptBtn = el("button", "action", "Accept");
  acceptBtn.type = "button";
  const rejectBtn = el("button", "action secondary", "Reject");
  rejectBtn.type = "button";
  const removeBtn = el("button", "action danger", "Remove");
  removeBtn.type = "button";
  acceptBtn.addEventListener("click", async () => {
    const fd = new FormData(decideForm as HTMLFormElement);
    const podSlug = String(fd.get("pod_slug") || "");
    const username = String(fd.get("username") || "");
    if (!podSlug || !username) return;
    await run("POST", `/pods/${podSlug}/members/${username}/accept`);
  });
  rejectBtn.addEventListener("click", async () => {
    const fd = new FormData(decideForm as HTMLFormElement);
    const podSlug = String(fd.get("pod_slug") || "");
    const username = String(fd.get("username") || "");
    if (!podSlug || !username) return;
    await run("POST", `/pods/${podSlug}/members/${username}/reject`);
  });
  removeBtn.addEventListener("click", async () => {
    const fd = new FormData(decideForm as HTMLFormElement);
    const podSlug = String(fd.get("pod_slug") || "");
    const username = String(fd.get("username") || "");
    if (!podSlug || !username) return;
    await run("POST", `/pods/${podSlug}/members/${username}/remove`);
  });
  decideForm.append(decideRow, acceptBtn, rejectBtn, removeBtn);

  const roleForm = el("form");
  const roleRow = el("div", "row");
  const roleSelect = document.createElement("select");
  roleSelect.name = "role";
  for (const r of ["ADMIN", "MEMBER", "OWNER"]) {
    const opt = document.createElement("option");
    opt.value = r;
    opt.textContent = r;
    roleSelect.append(opt);
  }
  roleRow.append(
    field("Pod slug", textInput("pod_slug", "")),
    field("Member username", textInput("username", "")),
    field("Role", roleSelect)
  );
  const roleBtn = el("button", "action", "Set role");
  roleBtn.type = "submit";
  roleForm.append(roleRow, roleBtn);
  bindForm(roleForm as HTMLFormElement, async (d) => {
    await run("POST", `/pods/${d.pod_slug}/members/${d.username}/role`, {
      role: d.role,
    });
  });

  const leaveForm = el("form");
  const leaveRow = el("div", "row");
  leaveRow.append(field("Pod slug", textInput("pod_slug", "")));
  const leaveBtn = el("button", "action danger", "Leave pod");
  leaveBtn.type = "submit";
  leaveForm.append(leaveRow, leaveBtn);
  bindForm(leaveForm as HTMLFormElement, async (d) => {
    await run("POST", `/pods/${d.pod_slug}/leave`);
  });

  const deleteForm = el("form");
  const deleteRow = el("div", "row");
  deleteRow.append(field("Pod slug", textInput("pod_slug", "")));
  const deleteBtn = el("button", "action danger", "Delete empty pod");
  deleteBtn.type = "submit";
  deleteForm.append(deleteRow, deleteBtn);
  bindForm(deleteForm as HTMLFormElement, async (d) => {
    await run("DELETE", `/pods/${d.pod_slug}`);
  });

  const browseRow = el("div", "row");
  browseRow.append(listBtn, myPodsBtn);

  wrap.append(
    card("Create", createForm),
    card("Browse", browseRow),
    card("Detail", detailForm),
    card("Join", joinForm),
    card("Owner actions", decideForm),
    card("Set role (owner)", roleForm),
    card("Leave", leaveForm),
    card("Delete (empty only)", deleteForm)
  );
  return wrap;
}

function renderCommunity(): HTMLElement {
  const wrap = el("div", "section active");
  wrap.append(el("h2", undefined, "Community Chat"));
  wrap.append(
    el(
      "p",
      "hint",
      "Must be enrolled in the skill for community channels. ACCEPTED pod membership for pod chat. Edit/delete own messages only."
    )
  );

  const skillSelect = document.createElement("select");
  skillSelect.name = "skill_slug";
  const refreshSkills = () => {
    skillSelect.innerHTML = "";
    if (cachedMySkills.length === 0) {
      skillSelect.append(el("option", undefined, "Join a skill first"));
      return;
    }
    for (const s of cachedMySkills) {
      const opt = document.createElement("option");
      opt.value = s.skill_slug;
      opt.textContent = `${s.skill_name} (${s.skill_slug})`;
      skillSelect.append(opt);
    }
  };
  refreshSkills();
  loadMySkills().then(() => refreshSkills());

  const browseRow = el("div", "row");
  const communityBtn = el("button", "action secondary", "Get community");
  communityBtn.addEventListener("click", async () => {
    const slug = skillSelect.value;
    if (!slug) return;
    await run("GET", `/skills/${slug}/community`);
  });
  const channelsBtn = el("button", "action", "List channels");
  channelsBtn.addEventListener("click", async () => {
    const slug = skillSelect.value;
    if (!slug) return;
    await run("GET", `/skills/${slug}/community/channels`);
  });
  browseRow.append(field("Skill", skillSelect), communityBtn, channelsBtn);

  const listChannelForm = el("form");
  const listChannelRow = el("div", "row");
  listChannelRow.append(
    field("Channel slug", textInput("channel_slug", "general")),
    field("Before id (optional)", textInput("before", "")),
    field("Limit (optional)", textInput("limit", "50"))
  );
  const listChannelBtn = el("button", "action", "List channel messages");
  listChannelBtn.type = "submit";
  listChannelForm.append(listChannelRow, listChannelBtn);
  bindForm(listChannelForm as HTMLFormElement, async (d) => {
    const skill = skillSelect.value;
    if (!skill) return;
    const qs = new URLSearchParams();
    if (d.before) qs.set("before", d.before);
    if (d.limit) qs.set("limit", d.limit);
    const q = qs.toString();
    await run(
      "GET",
      `/skills/${skill}/community/channels/${d.channel_slug}/messages${q ? `?${q}` : ""}`
    );
  });

  const sendChannelForm = el("form");
  const sendChannelRow = el("div", "row");
  sendChannelRow.append(
    field("Channel slug", textInput("channel_slug", "general")),
    field("Content", textInput("content", "hello from community"))
  );
  const sendChannelBtn = el("button", "action", "Send channel message");
  sendChannelBtn.type = "submit";
  sendChannelForm.append(sendChannelRow, sendChannelBtn);
  bindForm(sendChannelForm as HTMLFormElement, async (d) => {
    const skill = skillSelect.value;
    if (!skill) return;
    await run("POST", `/skills/${skill}/community/channels/${d.channel_slug}/messages`, {
      content: d.content,
    });
  });

  const listPodForm = el("form");
  const listPodRow = el("div", "row");
  listPodRow.append(
    field("Pod slug", textInput("pod_slug", "")),
    field("Before id (optional)", textInput("before", "")),
    field("Limit (optional)", textInput("limit", "50"))
  );
  const listPodBtn = el("button", "action", "List pod messages");
  listPodBtn.type = "submit";
  listPodForm.append(listPodRow, listPodBtn);
  bindForm(listPodForm as HTMLFormElement, async (d) => {
    const qs = new URLSearchParams();
    if (d.before) qs.set("before", d.before);
    if (d.limit) qs.set("limit", d.limit);
    const q = qs.toString();
    await run("GET", `/pods/${d.pod_slug}/messages${q ? `?${q}` : ""}`);
  });

  const sendPodForm = el("form");
  const sendPodRow = el("div", "row");
  sendPodRow.append(
    field("Pod slug", textInput("pod_slug", "")),
    field("Content", textInput("content", "hello from pod"))
  );
  const sendPodBtn = el("button", "action", "Send pod message");
  sendPodBtn.type = "submit";
  sendPodForm.append(sendPodRow, sendPodBtn);
  bindForm(sendPodForm as HTMLFormElement, async (d) => {
    await run("POST", `/pods/${d.pod_slug}/messages`, { content: d.content });
  });

  const editForm = el("form");
  const editRow = el("div", "row");
  editRow.append(
    field("Message id", textInput("id", "")),
    field("Content", textInput("content", "edited message"))
  );
  const editBtn = el("button", "action secondary", "Edit message");
  editBtn.type = "submit";
  editForm.append(editRow, editBtn);
  bindForm(editForm as HTMLFormElement, async (d) => {
    await run("PATCH", `/messages/${d.id}`, { content: d.content });
  });

  const deleteForm = el("form");
  const deleteRow = el("div", "row");
  deleteRow.append(field("Message id", textInput("id", "")));
  const deleteBtn = el("button", "action danger", "Delete message");
  deleteBtn.type = "submit";
  deleteForm.append(deleteRow, deleteBtn);
  bindForm(deleteForm as HTMLFormElement, async (d) => {
    await run("DELETE", `/messages/${d.id}`);
  });

  wrap.append(
    card("Browse community", browseRow),
    card("Channel history", listChannelForm),
    card("Send to channel", sendChannelForm),
    card("Pod history", listPodForm),
    card("Send to pod", sendPodForm),
    card("Edit message", editForm),
    card("Delete message", deleteForm)
  );
  return wrap;
}

function renderAudio(): HTMLElement {
  const wrap = el("div", "section active");
  wrap.append(el("h2", undefined, "Audio Rooms"));
  wrap.append(
    el(
      "p",
      "hint",
      "Create/list in Taggy DB, then Connect to LiveKit (mic on). LiveKit Cloud only shows a room after a real WebRTC connect — minting a token alone is not enough."
    )
  );

  const statusEl = el("p", "hint");
  const refreshLivekitStatus = () => {
    statusEl.textContent = livekitRoom
      ? `LiveKit: connected to room ${livekitConnectedRoomId ?? "?"} as ${livekitRoom.localParticipant.identity}`
      : "LiveKit: not connected";
  };
  refreshLivekitStatus();

  const skillSelect = document.createElement("select");
  skillSelect.name = "skill_slug";
  const refreshSkills = () => {
    skillSelect.innerHTML = "";
    if (cachedMySkills.length === 0) {
      skillSelect.append(el("option", undefined, "Join a skill first"));
      return;
    }
    for (const s of cachedMySkills) {
      const opt = document.createElement("option");
      opt.value = s.skill_slug;
      opt.textContent = `${s.skill_name} (${s.skill_slug})`;
      skillSelect.append(opt);
    }
  };
  refreshSkills();
  loadMySkills().then(() => refreshSkills());

  const createPodForm = el("form");
  const createPodRow = el("div", "row");
  createPodRow.append(
    field("Pod slug", textInput("pod_slug", "")),
    field("Title", textInput("title", "Evening study sync")),
    field("Description", textInput("description", "")),
    field("Max participants", textInput("max_participants", "7"))
  );
  const createPodBtn = el("button", "action", "Create pod room");
  createPodBtn.type = "submit";
  createPodForm.append(createPodRow, createPodBtn);
  bindForm(createPodForm as HTMLFormElement, async (d) => {
    const body: Record<string, unknown> = { title: d.title };
    if (d.description) body.description = d.description;
    if (d.max_participants) body.max_participants = Number(d.max_participants);
    await run("POST", `/pods/${d.pod_slug}/audio-rooms`, body);
  });

  const listPodForm = el("form");
  const listPodRow = el("div", "row");
  listPodRow.append(
    field("Pod slug", textInput("pod_slug", "")),
    field("Status filter", textInput("status", "ACTIVE"))
  );
  const listPodBtn = el("button", "action secondary", "List pod rooms");
  listPodBtn.type = "submit";
  listPodForm.append(listPodRow, listPodBtn);
  bindForm(listPodForm as HTMLFormElement, async (d) => {
    const qs = d.status ? `?status=${encodeURIComponent(d.status)}` : "";
    await run("GET", `/pods/${d.pod_slug}/audio-rooms${qs}`);
  });

  const createChannelForm = el("form");
  const createChannelRow = el("div", "row");
  createChannelRow.append(
    field("Skill", skillSelect),
    field("Channel slug", textInput("channel_slug", "general")),
    field("Title", textInput("title", "Community standup")),
    field("Description", textInput("description", ""))
  );
  const createChannelBtn = el("button", "action", "Create channel room");
  createChannelBtn.type = "submit";
  createChannelForm.append(createChannelRow, createChannelBtn);
  bindForm(createChannelForm as HTMLFormElement, async (d) => {
    const skill = skillSelect.value;
    if (!skill) return;
    const body: Record<string, unknown> = { title: d.title };
    if (d.description) body.description = d.description;
    await run(
      "POST",
      `/skills/${skill}/community/channels/${d.channel_slug}/audio-rooms`,
      body
    );
  });

  const listChannelForm = el("form");
  const listChannelRow = el("div", "row");
  listChannelRow.append(
    field("Channel slug", textInput("channel_slug", "general")),
    field("Status filter", textInput("status", "ACTIVE"))
  );
  const listChannelBtn = el("button", "action secondary", "List channel rooms");
  listChannelBtn.type = "submit";
  listChannelForm.append(listChannelRow, listChannelBtn);
  bindForm(listChannelForm as HTMLFormElement, async (d) => {
    const skill = skillSelect.value;
    if (!skill) return;
    const qs = d.status ? `?status=${encodeURIComponent(d.status)}` : "";
    await run(
      "GET",
      `/skills/${skill}/community/channels/${d.channel_slug}/audio-rooms${qs}`
    );
  });

  const detailForm = el("form");
  const detailRow = el("div", "row");
  detailRow.append(field("Room id (uuid)", textInput("room_id", "")));
  const detailBtn = el("button", "action", "Get room");
  detailBtn.type = "submit";
  detailForm.append(detailRow, detailBtn);
  bindForm(detailForm as HTMLFormElement, async (d) => {
    await run("GET", `/audio-rooms/${d.room_id}`);
  });

  const joinForm = el("form");
  const joinRow = el("div", "row");
  joinRow.append(field("Room id (uuid)", textInput("room_id", "")));
  const joinBtn = el("button", "action", "Connect + talk (LiveKit)");
  joinBtn.type = "submit";
  joinForm.append(joinRow, joinBtn);
  bindForm(joinForm as HTMLFormElement, async (d) => {
    const roomId = d.room_id;
    if (!roomId) return;

    type JoinPayload = {
      room_id: string;
      livekit_url: string;
      livekit_room_name: string;
      token: string;
      role: string;
    };

    const result = await run<JoinPayload>("POST", `/audio-rooms/${roomId}/join`);
    if (!result.ok || !result.data?.token || !result.data?.livekit_url) {
      return;
    }

    if (livekitRoom) {
      await livekitRoom.disconnect();
      livekitRoom = null;
      livekitConnectedRoomId = null;
    }

    const room = new Room({
      adaptiveStream: true,
      dynacast: true,
    });

    room.on(RoomEvent.TrackSubscribed, (track) => {
      if (track.kind === Track.Kind.Audio) {
        const audioEl = track.attach();
        audioEl.autoplay = true;
        document.body.appendChild(audioEl);
      }
    });

    room.on(RoomEvent.Disconnected, () => {
      livekitRoom = null;
      livekitConnectedRoomId = null;
      refreshLivekitStatus();
    });

    await room.connect(result.data.livekit_url, result.data.token);
    await room.localParticipant.setMicrophoneEnabled(true);

    livekitRoom = room;
    livekitConnectedRoomId = roomId;
    refreshLivekitStatus();

    logResult({
      ok: true,
      status: 200,
      method: "LIVEKIT",
      path: "/livekit/connect",
      data: {
        message: "Connected to LiveKit with mic enabled",
        room_id: roomId,
        livekit_room_name: result.data.livekit_room_name,
        role: result.data.role,
        identity: room.localParticipant.identity,
      },
    });
  });

  const leaveForm = el("form");
  const leaveRow = el("div", "row");
  leaveRow.append(field("Room id (uuid)", textInput("room_id", "")));
  const leaveBtn = el("button", "action secondary", "Disconnect + leave");
  leaveBtn.type = "submit";
  leaveForm.append(leaveRow, leaveBtn);
  bindForm(leaveForm as HTMLFormElement, async (d) => {
    if (livekitRoom) {
      await livekitRoom.disconnect();
      livekitRoom = null;
      livekitConnectedRoomId = null;
      refreshLivekitStatus();
    }
    if (d.room_id) {
      await run("POST", `/audio-rooms/${d.room_id}/leave`);
    }
  });

  const endForm = el("form");
  const endRow = el("div", "row");
  endRow.append(field("Room id (uuid)", textInput("room_id", "")));
  const endBtn = el("button", "action danger", "End room (host)");
  endBtn.type = "submit";
  endForm.append(endRow, endBtn);
  bindForm(endForm as HTMLFormElement, async (d) => {
    if (livekitRoom && livekitConnectedRoomId === d.room_id) {
      await livekitRoom.disconnect();
      livekitRoom = null;
      livekitConnectedRoomId = null;
      refreshLivekitStatus();
    }
    await run("POST", `/audio-rooms/${d.room_id}/end`);
  });

  wrap.append(
    statusEl,
    card("Create pod room", createPodForm),
    card("List pod rooms", listPodForm),
    card("Create channel room", createChannelForm),
    card("List channel rooms", listChannelForm),
    card("Room detail", detailForm),
    card("Connect (real LiveKit)", joinForm),
    card("Leave", leaveForm),
    card("End", endForm)
  );
  return wrap;
}

function renderNotifications(): HTMLElement {
  const wrap = el("div", "section active");
  wrap.append(el("h2", undefined, "Notifications"));
  wrap.append(
    el(
      "p",
      "hint",
      "In-app only. Created on pod join request / accept / reject / remove. Scoped to your username."
    )
  );

  const listBtn = el("button", "action", "List notifications");
  listBtn.addEventListener("click", () => {
    const path = scopedUserPath("/notifications");
    if (path) run("GET", path);
  });

  const unreadBtn = el("button", "action secondary", "Unread only");
  unreadBtn.addEventListener("click", () => {
    const path = scopedUserPath("/notifications?unread_only=true");
    if (path) run("GET", path);
  });

  const readAllBtn = el("button", "action secondary", "Mark all read");
  readAllBtn.addEventListener("click", () => {
    const path = scopedUserPath("/notifications/read-all");
    if (path) run("POST", path);
  });

  const readForm = el("form");
  const readRow = el("div", "row");
  readRow.append(field("Notification id", textInput("id", "")));
  const readBtn = el("button", "action", "Mark one read");
  readBtn.type = "submit";
  readForm.append(readRow, readBtn);
  bindForm(readForm as HTMLFormElement, async (d) => {
    const path = scopedUserPath(`/notifications/${d.id}/read`);
    if (path) await run("PATCH", path);
  });

  const row = el("div", "row");
  row.append(listBtn, unreadBtn, readAllBtn);

  wrap.append(card("Inbox", row), card("Mark read", readForm));
  return wrap;
}

function renderReports(): HTMLElement {
  const wrap = el("div", "section active");
  wrap.append(el("h2", undefined, "Reports"));
  wrap.append(
    el(
      "p",
      "hint",
      "File a report against USER / POD / MESSAGE / AUDIO_ROOM / COMMUNITY_CHANNEL / PROPOSAL. Duplicate OPEN reports are blocked."
    )
  );

  const targetType = document.createElement("select");
  targetType.name = "target_type";
  for (const t of [
    "USER",
    "POD",
    "MESSAGE",
    "AUDIO_ROOM",
    "COMMUNITY_CHANNEL",
    "PROPOSAL",
  ]) {
    const opt = document.createElement("option");
    opt.value = t;
    opt.textContent = t;
    targetType.append(opt);
  }

  const createForm = el("form");
  const createRow = el("div", "row");
  createRow.append(
    field("Target type", targetType),
    field("Target id", textInput("target_id", "")),
    field("Reason", textInput("reason", "Spam or abuse"))
  );
  const createBtn = el("button", "action", "Create report");
  createBtn.type = "submit";
  createForm.append(createRow, createBtn);
  bindForm(createForm as HTMLFormElement, async (d) => {
    await run("POST", "/reports", {
      target_type: d.target_type,
      target_id: Number(d.target_id),
      reason: d.reason,
    });
  });

  const listBtn = el("button", "action", "List my reports");
  listBtn.addEventListener("click", () => {
    const path = scopedUserPath("/reports");
    if (path) run("GET", path);
  });

  wrap.append(card("Create", createForm), card("Mine", listBtn));
  return wrap;
}

function renderSearch(): HTMLElement {
  const wrap = el("div", "section active");
  wrap.append(el("h2", undefined, "Search"));
  wrap.append(
    el(
      "p",
      "hint",
      "GET /search?q=…&types=skills,users,communities&limit=20"
    )
  );

  const form = el("form");
  const row = el("div", "row");
  row.append(
    field("Query", textInput("q", "web")),
    field("Types (comma)", textInput("types", "skills,users,communities")),
    field("Limit", textInput("limit", "20"))
  );
  const btn = el("button", "action", "Search");
  btn.type = "submit";
  form.append(row, btn);
  bindForm(form as HTMLFormElement, async (d) => {
    const params = new URLSearchParams();
    params.set("q", d.q);
    if (d.types.trim()) params.set("types", d.types.trim());
    if (d.limit.trim()) params.set("limit", d.limit.trim());
    await run("GET", `/search?${params.toString()}`);
  });

  wrap.append(card("Discover", form));
  return wrap;
}

function renderMain() {
  main.innerHTML = "";
  let section: HTMLElement;
  switch (activeSection) {
    case "auth":
      section = renderAuth();
      break;
    case "profile":
      section = renderProfile();
      break;
    case "skills":
      section = renderSkills();
      break;
    case "milestones":
      section = renderMilestones();
      break;
    case "progress":
      section = renderProgress();
      break;
    case "pods":
      section = renderPods();
      break;
    case "community":
      section = renderCommunity();
      break;
    case "audio":
      section = renderAudio();
      break;
    case "notifications":
      section = renderNotifications();
      break;
    case "reports":
      section = renderReports();
      break;
    case "search":
      section = renderSearch();
      break;
  }
  main.append(section);
}

export function mountDevTester(root: HTMLElement) {
  oauthCallback = consumeOAuthCallback();
  if (oauthCallback.pendingGoogle) {
    activeSection = "auth";
  }

  root.innerHTML = "";
  root.className = "dev-app";

  nav = document.createElement("nav");
  nav.id = "nav";
  main = document.createElement("main");
  main.id = "main";
  const aside = document.createElement("aside");
  aside.id = "log";
  aside.innerHTML = "<h2>Response log</h2>";
  logEntries = document.createElement("div");
  logEntries.id = "log-entries";
  aside.appendChild(logEntries);

  root.append(nav, main, aside);

  renderNav();
  renderMain();

  if (oauthCallback.error) {
    logResult({
      ok: false,
      status: 400,
      data: { message: oauthCallback.error },
      path: "/auth/callback",
      method: "GET",
    });
  } else if (oauthCallback.loggedIn) {
    logResult({
      ok: true,
      status: 200,
      data: {
        message: "Google login successful",
        username: getUsername(),
      },
      path: "/auth/callback",
      method: "GET",
    });
  } else if (oauthCallback.pendingGoogle) {
    logResult({
      ok: true,
      status: 200,
      data: {
        message: "Complete your Google profile below (username required)",
        email: oauthCallback.pendingGoogle.email,
      },
      path: "/auth/complete-google",
      method: "GET",
    });
  }

  return () => {
    void livekitRoom?.disconnect();
    livekitRoom = null;
    root.innerHTML = "";
  };
}
