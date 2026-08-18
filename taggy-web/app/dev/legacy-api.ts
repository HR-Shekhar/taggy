const API_BASE =
  process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") ?? "http://localhost:8080";

export type ApiResult<T = unknown> = {
  ok: boolean;
  status: number;
  data: T;
  path: string;
  method: string;
};

export function getTokens() {
  return {
    access: localStorage.getItem("access_token"),
    refresh: localStorage.getItem("refresh_token"),
  };
}

export function setTokens(access: string, refresh: string) {
  localStorage.setItem("access_token", access);
  localStorage.setItem("refresh_token", refresh);
}

export function clearTokens() {
  localStorage.removeItem("access_token");
  localStorage.removeItem("refresh_token");
}

export function getUsername() {
  return localStorage.getItem("username");
}

export function setUsername(username: string) {
  localStorage.setItem("username", username);
}

export function clearUsername() {
  localStorage.removeItem("username");
}

export function userBasePath(username?: string) {
  const u = username ?? getUsername();
  return u ? `/users/${u}` : null;
}

export async function request<T = unknown>(
  method: string,
  path: string,
  body?: unknown,
  useAuth = true
): Promise<ApiResult<T>> {
  const headers: Record<string, string> = {
    Accept: "application/json",
  };

  if (body !== undefined) {
    headers["Content-Type"] = "application/json";
  }

  const { access } = getTokens();
  if (useAuth && access) {
    headers.Authorization = `Bearer ${access}`;
  }

  const response = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  let data: T;
  const text = await response.text();
  if (text) {
    try {
      data = JSON.parse(text) as T;
    } catch {
      data = text as T;
    }
  } else {
    data = undefined as T;
  }

  return {
    ok: response.ok,
    status: response.status,
    data,
    path,
    method,
  };
}
