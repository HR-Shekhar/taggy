"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { MessageCircle } from "lucide-react";
import { useAuth } from "@/lib/auth";
import {
  acceptMember,
  getPod,
  listNotifications,
  markAllNotificationsRead,
  markNotificationRead,
  rejectMember,
} from "@/lib/api";
import { Empty, Loading, PageHeader } from "@/components/app-ui";
import { toastApiError } from "@/lib/toast";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { cn } from "@/lib/utils";

type Notification = {
  id: number;
  type: string;
  title: string;
  body: string;
  is_read: boolean;
  created_at: string;
  entity_type?: string | null;
  entity_id?: number | null;
};

type ParsedAction = {
  podSlug?: string;
  actorUsername?: string;
  href?: string;
  label?: string;
};

type RequestStatus = "PENDING" | "ACCEPTED" | "REJECTED";

function requestKey(podSlug: string, actorUsername: string) {
  return `${podSlug.toLowerCase()}:${actorUsername.toLowerCase()}`;
}

function parseNotification(n: Notification): ParsedAction {
  const body = n.body.trim();

  if (n.type === "POD_JOIN_REQUEST") {
    const m = body.match(/^(\S+)\s+requested to join\s+(\S+)\s*$/i);
    if (m) {
      return {
        actorUsername: m[1],
        podSlug: m[2],
        href: `/pods/${m[2]}`,
        label: "Open pod chat",
      };
    }
  }

  if (
    n.type === "POD_JOIN_ACCEPTED" ||
    n.type === "POD_JOIN_REJECTED" ||
    n.type === "POD_MEMBER_REMOVED"
  ) {
    const m =
      body.match(/join\s+(\S+)\s+was\s+(accepted|rejected)/i) ??
      body.match(/removed from\s+(\S+)\s*$/i);
    if (m?.[1]) {
      return {
        podSlug: m[1],
        href: `/pods/${m[1]}`,
        label: "Open pod chat",
      };
    }
  }

  if (
    n.type === "MILESTONE_COMPLETED" ||
    n.type === "MILESTONE_DUE" ||
    n.type === "ROADMAP_UPDATED"
  ) {
    const m =
      body.match(/\bon\s+(\S+)\s*$/i) ??
      body.match(/\bon\s+(\S+)\s+is due/i) ??
      body.match(/Your\s+(\S+)\s+roadmap/i);
    if (m?.[1]) {
      return {
        href: `/skills/${m[1]}`,
        label: "Open skill",
      };
    }
  }

  if (n.type === "COMMUNITY_ANNOUNCEMENT") {
    const m = body.match(/Welcome to the\s+(\S+)\s+community/i);
    if (m?.[1]) {
      return {
        href: `/community/${m[1]}`,
        label: "Open community",
      };
    }
  }

  return {};
}

function formatWhen(value: string) {
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return value;
  return d.toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

async function resolveJoinRequestStatuses(
  notifications: Notification[]
): Promise<Record<string, RequestStatus>> {
  const targets = new Map<string, { podSlug: string; actor: string }>();

  for (const n of notifications) {
    if (n.type !== "POD_JOIN_REQUEST") continue;
    const action = parseNotification(n);
    if (!action.podSlug || !action.actorUsername) continue;
    targets.set(requestKey(action.podSlug, action.actorUsername), {
      podSlug: action.podSlug,
      actor: action.actorUsername,
    });
  }

  const byPod = new Map<string, string[]>();
  for (const { podSlug, actor } of targets.values()) {
    const list = byPod.get(podSlug) ?? [];
    list.push(actor);
    byPod.set(podSlug, list);
  }

  const statuses: Record<string, RequestStatus> = {};

  await Promise.all(
    [...byPod.entries()].map(async ([podSlug, actors]) => {
      const result = await getPod(podSlug);
      if (!result.ok) {
        for (const actor of actors) {
          statuses[requestKey(podSlug, actor)] = "REJECTED";
        }
        return;
      }

      const data = result.data as {
        members?: { username: string }[];
        join_requests?: { username: string }[];
      };
      const members = new Set(
        (data.members ?? []).map((m) => m.username.toLowerCase())
      );
      const pending = new Set(
        (data.join_requests ?? []).map((m) => m.username.toLowerCase())
      );

      for (const actor of actors) {
        const key = requestKey(podSlug, actor);
        const name = actor.toLowerCase();
        if (pending.has(name)) statuses[key] = "PENDING";
        else if (members.has(name)) statuses[key] = "ACCEPTED";
        else statuses[key] = "REJECTED";
      }
    })
  );

  return statuses;
}

export default function NotificationsPage() {
  const { username } = useAuth();
  const router = useRouter();
  const [items, setItems] = useState<Notification[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [requestStatus, setRequestStatus] = useState<
    Record<string, RequestStatus>
  >({});
  const [loading, setLoading] = useState(true);
  const [unreadOnly, setUnreadOnly] = useState(false);
  const [busyId, setBusyId] = useState<number | null>(null);

  async function load(onlyUnread = unreadOnly) {
    if (!username) return;
    setLoading(true);
    const result = await listNotifications(username, onlyUnread);
    if (!result.ok) {
      setLoading(false);
      toastApiError(result);
      return;
    }
    const data = result.data as {
      notifications?: Notification[];
      unread_count?: number;
    };
    const list =
      data.notifications ??
      (Array.isArray(result.data) ? (result.data as Notification[]) : []);
    setItems(list);
    setUnreadCount(data.unread_count ?? 0);

    const statuses = await resolveJoinRequestStatuses(list);
    setRequestStatus(statuses);
    setLoading(false);
  }

  useEffect(() => {
    void load();
  }, [username]);

  async function markRead(id: number) {
    if (!username) return;
    const r = await markNotificationRead(username, id);
    if (!r.ok && r.status !== 409) {
      toastApiError(r);
    }
  }

  async function goToChat(n: Notification, href: string) {
    if (!n.is_read) await markRead(n.id);
    router.push(href);
  }

  if (loading) return <Loading />;

  return (
    <div className="space-y-6">
      <PageHeader
        title="Notifications"
        description={`${unreadCount} unread`}
      >
        <div className="flex flex-wrap gap-2">
          <Button
            variant={!unreadOnly ? "default" : "outline"}
            size="sm"
            onClick={() => {
              setUnreadOnly(false);
              void load(false);
            }}
          >
            All
          </Button>
          <Button
            variant={unreadOnly ? "default" : "outline"}
            size="sm"
            onClick={() => {
              setUnreadOnly(true);
              void load(true);
            }}
          >
            Unread
          </Button>
          <Button
            size="sm"
            variant="secondary"
            onClick={async () => {
              if (!username) return;
              const r = await markAllNotificationsRead(username);
              if (!r.ok) toastApiError(r);
              else void load();
            }}
          >
            Mark all read
          </Button>
        </div>
      </PageHeader>

      {items.length === 0 ? (
        <Empty>No notifications.</Empty>
      ) : (
        <div className="space-y-3">
          {items.map((n) => {
            const action = parseNotification(n);
            const statusKey =
              action.podSlug && action.actorUsername
                ? requestKey(action.podSlug, action.actorUsername)
                : null;
            const joinStatus =
              n.type === "POD_JOIN_REQUEST" && statusKey
                ? requestStatus[statusKey]
                : undefined;
            const isPending = joinStatus === "PENDING";
            const canModerate =
              n.type === "POD_JOIN_REQUEST" &&
              isPending &&
              Boolean(action.podSlug && action.actorUsername);

            return (
              <Card
                key={n.id}
                className={cn(
                  "rounded-xl ring-1 ring-foreground/10",
                  !n.is_read && "bg-secondary/20"
                )}
              >
                <CardHeader className="flex-row items-start justify-between gap-3 pb-2">
                  <div className="min-w-0">
                    <div className="flex flex-wrap items-center gap-2">
                      <CardTitle className="text-base">{n.title}</CardTitle>
                      {!n.is_read ? <Badge>unread</Badge> : null}
                      {joinStatus === "ACCEPTED" ? (
                        <Badge variant="secondary">Accepted</Badge>
                      ) : null}
                      {joinStatus === "REJECTED" ? (
                        <Badge variant="outline">Rejected</Badge>
                      ) : null}
                      {joinStatus === "PENDING" ? (
                        <Badge variant="outline">Pending</Badge>
                      ) : null}
                    </div>
                    <CardDescription className="mt-1 text-sm text-foreground/80">
                      {n.body}
                    </CardDescription>
                    <p className="mt-2 text-xs text-muted-foreground">
                      {n.type.replaceAll("_", " ")} · {formatWhen(n.created_at)}
                    </p>
                  </div>
                </CardHeader>
                <CardContent className="flex flex-wrap gap-2 pt-0">
                  {canModerate ? (
                    <>
                      <Button
                        size="sm"
                        disabled={busyId === n.id}
                        onClick={async () => {
                          if (!action.podSlug || !action.actorUsername) return;
                          setBusyId(n.id);
                          const r = await acceptMember(
                            action.podSlug,
                            action.actorUsername
                          );
                          if (!r.ok) {
                            setBusyId(null);
                            toastApiError(r);
                            return;
                          }
                          setRequestStatus((prev) => ({
                            ...prev,
                            [requestKey(action.podSlug!, action.actorUsername!)]:
                              "ACCEPTED",
                          }));
                          await markRead(n.id);
                          setBusyId(null);
                          void load();
                        }}
                      >
                        Accept
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        disabled={busyId === n.id}
                        onClick={async () => {
                          if (!action.podSlug || !action.actorUsername) return;
                          setBusyId(n.id);
                          const r = await rejectMember(
                            action.podSlug,
                            action.actorUsername
                          );
                          if (!r.ok) {
                            setBusyId(null);
                            toastApiError(r);
                            return;
                          }
                          setRequestStatus((prev) => ({
                            ...prev,
                            [requestKey(action.podSlug!, action.actorUsername!)]:
                              "REJECTED",
                          }));
                          await markRead(n.id);
                          setBusyId(null);
                          void load();
                        }}
                      >
                        Reject
                      </Button>
                    </>
                  ) : null}

                  {action.href ? (
                    <Button
                      size="sm"
                      variant={canModerate ? "outline" : "default"}
                      className="gap-1.5"
                      onClick={() => void goToChat(n, action.href!)}
                    >
                      <MessageCircle className="size-3.5" />
                      {action.label ?? "Open"}
                    </Button>
                  ) : null}

                  {!n.is_read ? (
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={async () => {
                        await markRead(n.id);
                        void load();
                      }}
                    >
                      Mark read
                    </Button>
                  ) : null}
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
}
