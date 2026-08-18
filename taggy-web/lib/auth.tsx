"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import {
  clearSession,
  clearReadNotifications,
  getAvatarUrl,
  getIsAdmin,
  getProfile,
  getSubscription,
  getTokens,
  getUsername,
  login as apiLogin,
  logout as apiLogout,
  setAvatarUrl,
  setIsAdmin,
  setSubscription,
  setTokens,
  setUsername,
  type TokenPair,
} from "@/lib/api";
import { displayFirstName } from "@/lib/utils";

type AuthState = {
  ready: boolean;
  username: string | null;
  displayName: string;
  isAdmin: boolean;
  subscription: string;
  avatarUrl: string | null;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<string | null>;
  acceptTokens: (pair: TokenPair) => void;
  logout: () => Promise<void>;
  refreshLocal: () => void;
  setSubscriptionState: (subscription: string) => void;
  setAvatarUrlState: (url: string | null) => void;
};

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [ready, setReady] = useState(false);
  const [username, setUsernameState] = useState<string | null>(null);
  const [displayName, setDisplayName] = useState("");
  const [isAdmin, setIsAdminState] = useState(false);
  const [subscription, setSubscriptionState] = useState("FREE");
  const [avatarUrl, setAvatarUrlStateInternal] = useState<string | null>(null);

  const refreshLocal = useCallback(() => {
    const { access } = getTokens();
    const u = getUsername();
    setUsernameState(access && u ? u : null);
    setDisplayName(access && u ? displayFirstName(null, u) : "");
    setIsAdminState(Boolean(access && u && getIsAdmin()));
    setSubscriptionState(access && u ? getSubscription() : "FREE");
    setAvatarUrlStateInternal(access && u ? getAvatarUrl() : null);
  }, []);

  useEffect(() => {
    refreshLocal();
    setReady(true);
    const { access } = getTokens();
    const u = getUsername();
    if (access && u) {
      void getProfile(u).then((res) => {
        if (res.ok && typeof res.data?.is_admin === "boolean") {
          setIsAdmin(Boolean(res.data.is_admin));
          setIsAdminState(Boolean(res.data.is_admin));
        }
        if (res.ok && typeof res.data?.subscription === "string") {
          setSubscription(String(res.data.subscription));
          setSubscriptionState(String(res.data.subscription));
        }
        if (res.ok) {
          const pic =
            typeof res.data?.profile_picture_url === "string"
              ? String(res.data.profile_picture_url)
              : null;
          setAvatarUrl(pic);
          setAvatarUrlStateInternal(pic);
          const name =
            typeof res.data?.name === "string" ? String(res.data.name) : null;
          setDisplayName(displayFirstName(name, u));
        }
      });
    }
  }, [refreshLocal]);

  const acceptTokens = useCallback((pair: TokenPair) => {
    setTokens(pair.access_token, pair.refresh_token);
    setUsername(pair.username);
    setUsernameState(pair.username);
    setDisplayName(displayFirstName(null, pair.username));
    const admin = Boolean(pair.is_admin);
    setIsAdmin(admin);
    setIsAdminState(admin);
    const sub = pair.subscription || "FREE";
    setSubscription(sub);
    setSubscriptionState(sub);
    void clearReadNotifications(pair.username);
    void getProfile(pair.username).then((res) => {
      if (!res.ok) return;
      const pic =
        typeof res.data?.profile_picture_url === "string"
          ? String(res.data.profile_picture_url)
          : null;
      setAvatarUrl(pic);
      setAvatarUrlStateInternal(pic);
      const name =
        typeof res.data?.name === "string" ? String(res.data.name) : null;
      setDisplayName(displayFirstName(name, pair.username));
    });
  }, []);

  const login = useCallback(
    async (email: string, password: string) => {
      const result = await apiLogin(email, password);
      if (!result.ok || !result.data?.access_token) {
        const msg =
          (result.data as { message?: string })?.message ??
          `Login failed (${result.status})`;
        return msg;
      }
      acceptTokens(result.data);
      return null;
    },
    [acceptTokens]
  );

  const logout = useCallback(async () => {
    await apiLogout();
    clearSession();
    setUsernameState(null);
    setDisplayName("");
    setIsAdminState(false);
    setSubscriptionState("FREE");
    setAvatarUrlStateInternal(null);
  }, []);

  const setSubscriptionStateCb = useCallback((sub: string) => {
    setSubscription(sub);
    setSubscriptionState(sub);
  }, []);

  const setAvatarUrlState = useCallback((url: string | null) => {
    setAvatarUrl(url);
    setAvatarUrlStateInternal(url);
  }, []);

  const value = useMemo(
    () => ({
      ready,
      username,
      displayName,
      isAdmin,
      subscription,
      avatarUrl,
      isAuthenticated: Boolean(username),
      login,
      acceptTokens,
      logout,
      refreshLocal,
      setSubscriptionState: setSubscriptionStateCb,
      setAvatarUrlState,
    }),
    [
      ready,
      username,
      displayName,
      isAdmin,
      subscription,
      avatarUrl,
      login,
      acceptTokens,
      logout,
      refreshLocal,
      setSubscriptionStateCb,
      setAvatarUrlState,
    ]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
