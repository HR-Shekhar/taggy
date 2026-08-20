"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useAuth } from "@/lib/auth";
import {
  acceptMember,
  apiErrorMessage,
  createPodAudioRoom,
  deleteMessage,
  editMessage,
  getPod,
  joinPod,
  leavePod,
  listPodAudioRooms,
  listPodMessages,
  rejectMember,
  removeMember,
  sendPodMessage,
  setMemberRole,
} from "@/lib/api";
import { Empty, ErrorBox, Loading } from "@/components/app-ui";
import { PodQuizPanel } from "@/components/pod-quiz-panel";
import { toastApiError } from "@/lib/toast";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { ReportDialog } from "@/components/report-dialog";
import {
  AudioRoomsSidebar,
  ChatSidebarSection,
  ChatWorkspace,
} from "@/components/chat-workspace";
import { Badge } from "@/components/ui/badge";
import { Button, buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import {
  chatRoomKeyPod,
  type ChatSocketEvent,
} from "@/lib/chat-socket";
import { usePersistentChatSocket } from "@/components/chat-connection-provider";
import type { ChatMessage } from "@/components/chat-panel";

type Member = {
  username: string;
  name: string;
  role: string;
  status?: string;
  joined_at?: string;
};
type Message = ChatMessage & {
  id: number;
  author_name?: string;
};
type AudioRoom = {
  id: string;
  title: string;
  status: string;
  host_username: string;
};

function isAcceptedMember(members: Member[], username: string | null) {
  if (!username) return false;
  return members.some((m) => m.username === username);
}

export default function PodDetailPage() {
  const { podSlug } = useParams<{ podSlug: string }>();
  const router = useRouter();
  const { username } = useAuth();
  const [pod, setPod] = useState<Record<string, unknown> | null>(null);
  const [members, setMembers] = useState<Member[]>([]);
  const [joinRequests, setJoinRequests] = useState<Member[]>([]);
  const [messages, setMessages] = useState<Message[]>([]);
  const [rooms, setRooms] = useState<AudioRoom[]>([]);
  const [draft, setDraft] = useState("");
  const [replyTo, setReplyTo] = useState<Message | null>(null);
  const [roomTitle, setRoomTitle] = useState("Focus session");
  const [loadError, setLoadError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [leaveOpen, setLeaveOpen] = useState(false);
  const [reportPodOpen, setReportPodOpen] = useState(false);

  function applyChatEvent(event: ChatSocketEvent) {
    if (event.type === "message.created") {
      setMessages((prev) => {
        if (prev.some((m) => m.id === event.message.id)) return prev;
        return [...prev, event.message as Message];
      });
      return;
    }
    if (event.type === "message.updated") {
      setMessages((prev) =>
        prev.map((m) =>
          m.id === event.message.id ? ({ ...m, ...event.message } as Message) : m
        )
      );
      return;
    }
    if (event.type === "message.deleted") {
      setMessages((prev) => prev.filter((m) => m.id !== event.message_id));
    }
  }

  usePersistentChatSocket(
    isAcceptedMember(members, username) ? chatRoomKeyPod(podSlug) : null,
    applyChatEvent
  );

  async function load(opts?: { quiet?: boolean }) {
    if (!opts?.quiet) {
      setLoading(true);
      setLoadError(null);
    }
    const [detail, msgs, audio] = await Promise.all([
      getPod(podSlug),
      listPodMessages(podSlug),
      listPodAudioRooms(podSlug),
    ]);
    if (!detail.ok) {
      const message = apiErrorMessage(detail);
      setLoadError(message);
      toastApiError(detail);
      setLoading(false);
      return;
    }
    const data = detail.data as {
      pod?: Record<string, unknown>;
      members?: Member[];
      join_requests?: Member[];
    };
    setPod(data.pod ?? (detail.data as Record<string, unknown>));
    setMembers(data.members ?? []);
    setJoinRequests(data.join_requests ?? []);
    if (msgs.ok) {
      setMessages(
        Array.isArray(msgs.data)
          ? (msgs.data as Message[])
          : ((msgs.data as { messages?: Message[] })?.messages ?? [])
      );
    }
    if (audio.ok) {
      const list = Array.isArray(audio.data)
        ? (audio.data as AudioRoom[])
        : ((audio.data as { rooms?: AudioRoom[] })?.rooms ?? []);
      setRooms(list.filter((r) => r.status === "ACTIVE" || !r.status));
    }
    setLoading(false);
  }

  useEffect(() => {
    void load();
  }, [podSlug]);

  if (loading) {
    return (
      <div className="flex h-full flex-1 items-center justify-center">
        <Loading />
      </div>
    );
  }
  if (!pod) {
    return (
      <div className="flex h-full flex-1 items-center justify-center p-6">
        <ErrorBox message={loadError ?? "Pod not found"} />
      </div>
    );
  }

  const isMember = members.some((m) => m.username === username);
  const isOwner = members.some(
    (m) => m.username === username && m.role === "OWNER"
  );
  const from = `/pods/${podSlug}`;

  return (
    <div className="flex h-full min-h-0 flex-1 flex-col" data-tour="pod-workspace">
      <ConfirmDialog
        open={leaveOpen}
        onOpenChange={setLeaveOpen}
        title="Leave this pod?"
        description="You'll lose access to pod chat and audio rooms. You can join another pod for this skill afterward."
        confirmLabel="Leave pod"
        destructive
        busy={busy}
        onConfirm={async () => {
          setBusy(true);
          const r = await leavePod(podSlug);
          setBusy(false);
          if (!r.ok) {
            toastApiError(r);
            setLeaveOpen(false);
            return;
          }
          setLeaveOpen(false);
          router.push("/pods");
        }}
      />

      <ReportDialog
        open={reportPodOpen}
        onOpenChange={setReportPodOpen}
        targetType="POD"
        targetId={Number(pod.id) || null}
        title="Report pod"
        description="Report this pod for spam, harassment, or other abuse."
      />

      <ChatWorkspace
        leftLabel="Members"
        header={
          <div className="flex min-w-0 flex-1 items-center justify-between gap-3">
            <div className="min-w-0">
              <h1 className="truncate font-serif text-xl tracking-tight">
                {String(pod.name ?? podSlug)}
              </h1>
      <p className="truncate text-xs text-muted-foreground">
                {String(pod.skill_name ?? pod.skill_slug ?? "")}
                {" · "}
                @{String(pod.owner_username ?? "")}
              </p>
            </div>
            <div className="flex shrink-0 flex-wrap items-center gap-2">
              <Button
                size="sm"
                variant="ghost"
                onClick={() => setReportPodOpen(true)}
              >
                Report
              </Button>
              <Link
                href="/pods"
                className={cn(buttonVariants({ variant: "ghost", size: "sm" }))}
              >
                All pods
              </Link>
              {!isMember ? (
                <Button
                  size="sm"
                  disabled={busy}
                  onClick={async () => {
                    setBusy(true);
                    const r = await joinPod(podSlug);
                    setBusy(false);
                    if (!r.ok) toastApiError(r);
                    else void load({ quiet: true });
                  }}
                >
                  Request join
                </Button>
              ) : (
                <Button
                  size="sm"
                  variant="ghost"
                  disabled={busy}
                  onClick={() => setLeaveOpen(true)}
                >
                  Leave
                </Button>
              )}
            </div>
          </div>
        }
        left={
          <div className="flex min-h-0 flex-1 flex-col">
            <ChatSidebarSection title="Members">
              <ul className="space-y-1">
                {members.map((m) => (
                  <li
                    key={m.username}
                    className="rounded-lg px-2 py-1.5 hover:bg-muted/60"
                  >
                    <div className="flex items-center justify-between gap-2">
                      <span className="truncate text-sm font-medium">
                        @{m.username}
                      </span>
                      <Badge variant="outline" className="shrink-0 text-[10px]">
                        {m.role}
                      </Badge>
                    </div>
                    {isOwner && m.username !== username ? (
                      <div className="mt-1.5 flex flex-wrap gap-1">
                        <Button
                          size="xs"
                          variant="outline"
                          onClick={async () => {
                            const r = await setMemberRole(
                              podSlug,
                              m.username,
                              "ADMIN"
                            );
                            if (!r.ok) toastApiError(r);
                            else void load({ quiet: true });
                          }}
                        >
                          Admin
                        </Button>
                        <Button
                          size="xs"
                          variant="destructive"
                          onClick={async () => {
                            const r = await removeMember(podSlug, m.username);
                            if (!r.ok) toastApiError(r);
                            else void load({ quiet: true });
                          }}
                        >
                          Remove
                        </Button>
                      </div>
                    ) : null}
                  </li>
                ))}
              </ul>
            </ChatSidebarSection>

            <ChatSidebarSection
              title="Leaderboard"
              className="border-t border-border/60"
            >
              <div className="space-y-3">
                <PodQuizPanel
                  podSlug={podSlug}
                  enabled={isMember}
                  mode="leaderboard"
                />
                {isMember ? (
                  <Link
                    href={`/progress?skill=${String(pod.skill_slug ?? "")}`}
                    className={cn(
                      buttonVariants({ variant: "outline", size: "sm" }),
                      "w-full"
                    )}
                  >
                    Take quiz on Progress
                  </Link>
                ) : null}
              </div>
            </ChatSidebarSection>

            {isOwner ? (
              <ChatSidebarSection
                title="Join requests"
                className="border-t border-border/60"
              >
                {joinRequests.length === 0 ? (
                  <p className="text-sm text-muted-foreground">
                    No pending requests.
                  </p>
                ) : (
                  <ul className="space-y-2">
                    {joinRequests.map((m) => (
                      <li
                        key={m.username}
                        className="rounded-lg border border-border/60 bg-background/50 p-2"
                      >
                        <div className="mb-2 truncate text-sm font-medium">
                          @{m.username}
                        </div>
                        <div className="flex gap-1">
                          <Button
                            size="xs"
                            disabled={busy}
                            onClick={async () => {
                              setBusy(true);
                              const r = await acceptMember(podSlug, m.username);
                              setBusy(false);
                              if (!r.ok) toastApiError(r);
                              else void load({ quiet: true });
                            }}
                          >
                            Accept
                          </Button>
                          <Button
                            size="xs"
                            variant="outline"
                            disabled={busy}
                            onClick={async () => {
                              setBusy(true);
                              const r = await rejectMember(podSlug, m.username);
                              setBusy(false);
                              if (!r.ok) toastApiError(r);
                              else void load({ quiet: true });
                            }}
                          >
                            Reject
                          </Button>
                        </div>
                      </li>
                    ))}
                  </ul>
                )}
              </ChatSidebarSection>
            ) : null}
          </div>
        }
        messages={messages}
        empty={
          <Empty>
            {isMember
              ? "No messages yet — say hello to your pod."
              : "Join the pod to participate in chat."}
          </Empty>
        }
        threadKey={podSlug}
        currentUsername={username}
        draft={draft}
        onDraftChange={setDraft}
        replyTo={replyTo}
        onClearReply={() => setReplyTo(null)}
        onReplyMessage={(m) => setReplyTo(m as Message)}
        composerDisabled={!isMember}
        composerPlaceholder={
          isMember ? "Message the pod" : "Join the pod to chat"
        }
        onSend={async () => {
          const replyId =
            replyTo && Number.isFinite(Number(replyTo.id))
              ? Number(replyTo.id)
              : null;
          const r = await sendPodMessage(podSlug, draft, replyId);
          if (!r.ok) {
            toastApiError(r);
            return;
          }
          const created = r.data as Message;
          if (created?.id) {
            setMessages((prev) =>
              prev.some((m) => m.id === created.id) ? prev : [...prev, created]
            );
          }
          setDraft("");
          setReplyTo(null);
        }}
        onEditMessage={async (id, content) => {
          const r = await editMessage(id, content);
          if (!r.ok) {
            toastApiError(r);
            return;
          }
          const updated = r.data as Message;
          if (updated?.id) {
            setMessages((prev) =>
              prev.map((m) => (m.id === updated.id ? { ...m, ...updated } : m))
            );
          }
        }}
        onDeleteMessage={async (id) => {
          const r = await deleteMessage(id);
          if (!r.ok) {
            toastApiError(r);
            return;
          }
          setMessages((prev) => prev.filter((m) => m.id !== id && String(m.id) !== String(id)));
        }}
        audio={
          <AudioRoomsSidebar
            rooms={rooms}
            from={from}
            canCreate={isMember}
            roomTitle={roomTitle}
            onRoomTitleChange={setRoomTitle}
            createDisabled={!isMember}
            emptyHint={
              isMember
                ? "No live rooms right now."
                : "Join the pod to use audio rooms."
            }
            onCreate={async () => {
              const r = await createPodAudioRoom(podSlug, roomTitle);
                              if (!r.ok) toastApiError(r);
              else void load({ quiet: true });
            }}
          />
        }
      />
    </div>
  );
}
