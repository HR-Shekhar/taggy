"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { useAuth } from "@/lib/auth";
import {
  createChannelAudioRoom,
  deleteMessage,
  editMessage,
  listChannelAudioRooms,
  listChannelMessages,
  listChannels,
  sendChannelMessage,
} from "@/lib/api";
import { Empty, Loading } from "@/components/app-ui";
import { BackButton } from "@/components/back-button";
import { CommunityLeaderboardPanel } from "@/components/community-leaderboard-panel";
import { toastApiError } from "@/lib/toast";
import { ReportDialog } from "@/components/report-dialog";
import {
  AudioRoomsSidebar,
  ChatSidebarSection,
  ChatWorkspace,
} from "@/components/chat-workspace";
import { Button, buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import {
  chatRoomKeyChannel,
  type ChatSocketEvent,
} from "@/lib/chat-socket";
import { usePersistentChatSocket } from "@/components/chat-connection-provider";
import type { ChatMessage } from "@/components/chat-panel";

type Channel = { id?: number; slug: string; name: string; description?: string };
type Message = ChatMessage & { id: number };
type AudioRoom = {
  id: string;
  title: string;
  status: string;
  host_username?: string;
};

export default function CommunityPage() {
  const { skillSlug } = useParams<{ skillSlug: string }>();
  const { username } = useAuth();
  const [channels, setChannels] = useState<Channel[]>([]);
  const [channelSlug, setChannelSlug] = useState("");
  const [messages, setMessages] = useState<Message[]>([]);
  const [rooms, setRooms] = useState<AudioRoom[]>([]);
  const [draft, setDraft] = useState("");
  const [replyTo, setReplyTo] = useState<Message | null>(null);
  const [roomTitle, setRoomTitle] = useState("Community call");
  const [loading, setLoading] = useState(true);
  const [reportChannelOpen, setReportChannelOpen] = useState(false);

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
    channelSlug ? chatRoomKeyChannel(skillSlug, channelSlug) : null,
    applyChatEvent
  );

  async function loadChannels() {
    setLoading(true);
    const result = await listChannels(skillSlug);
    setLoading(false);
    if (!result.ok) {
      toastApiError(result);
      return;
    }
    const list = Array.isArray(result.data)
      ? (result.data as Channel[])
      : ((result.data as { channels?: Channel[] })?.channels ?? []);
    setChannels(list);
    if (!channelSlug && list[0]) setChannelSlug(list[0].slug);
  }

  async function loadMessages(ch: string) {
    if (!ch) return;
    const result = await listChannelMessages(skillSlug, ch);
    if (!result.ok) {
      toastApiError(result);
      return;
    }
    setMessages(
      Array.isArray(result.data)
        ? (result.data as Message[])
        : ((result.data as { messages?: Message[] })?.messages ?? [])
    );
  }

  async function loadRooms(ch: string) {
    if (!ch) return;
    const result = await listChannelAudioRooms(skillSlug, ch);
    if (!result.ok) return;
    setRooms(
      Array.isArray(result.data)
        ? (result.data as AudioRoom[]).filter(
            (r) => r.status === "ACTIVE" || !r.status
          )
        : ((result.data as { rooms?: AudioRoom[] })?.rooms ?? []).filter(
            (r) => r.status === "ACTIVE" || !r.status
          )
    );
  }

  useEffect(() => {
    void loadChannels();
  }, [skillSlug]);

  useEffect(() => {
    setReplyTo(null);
    setDraft("");
    void loadMessages(channelSlug);
    void loadRooms(channelSlug);
  }, [channelSlug, skillSlug]);

  if (loading) {
    return (
      <div className="flex h-full flex-1 items-center justify-center">
        <Loading />
      </div>
    );
  }

  const activeChannel = channels.find((c) => c.slug === channelSlug);
  const from = `/community/${skillSlug}`;
  const channelTargetId =
    activeChannel?.id && Number.isFinite(activeChannel.id)
      ? activeChannel.id
      : null;

  return (
    <>
      <ReportDialog
        open={reportChannelOpen}
        onOpenChange={setReportChannelOpen}
        targetType="COMMUNITY_CHANNEL"
        targetId={channelTargetId}
        title="Report channel"
        description="Report this community channel for spam or abuse."
      />

      <ChatWorkspace
      leftLabel="Channels"
      header={
        <div className="flex min-w-0 flex-1 items-center justify-between gap-3">
          <div className="flex min-w-0 items-center gap-2.5">
            <BackButton fallbackHref="/community" variant="ghost" size="sm" />
            <div className="min-w-0">
              <h1 className="truncate font-serif text-xl tracking-tight">
                #{activeChannel?.name ?? channelSlug ?? "channel"}
              </h1>
              <p className="truncate text-xs text-foreground/75">
                {skillSlug} community
              </p>
            </div>
          </div>
          <div className="flex shrink-0 flex-wrap items-center gap-2">
            {/* Mobile channel switcher */}
            <select
              className="h-8 max-w-40 rounded-lg border border-foreground/25 bg-background px-2 text-sm md:hidden"
              value={channelSlug}
              onChange={(e) => setChannelSlug(e.target.value)}
              aria-label="Channel"
            >
              {channels.map((c) => (
                <option key={c.slug} value={c.slug}>
                  #{c.name}
                </option>
              ))}
            </select>
            <Button
              size="sm"
              variant="ghost"
              disabled={!channelTargetId}
              onClick={() => setReportChannelOpen(true)}
            >
              Report
            </Button>
            <Link
              href={`/skills/${skillSlug}`}
              className={cn(buttonVariants({ variant: "outline", size: "sm" }))}
            >
              Roadmap
            </Link>
            <Link
              href="/community"
              className={cn(buttonVariants({ variant: "ghost", size: "sm" }))}
            >
              Hub
            </Link>
          </div>
        </div>
      }
      left={
        <div className="flex min-h-0 flex-1 flex-col">
          <ChatSidebarSection title="Channels" description={skillSlug}>
            {channels.length === 0 ? (
              <Empty>No channels yet.</Empty>
            ) : (
              <ul className="space-y-0.5">
                {channels.map((c) => {
                  const active = c.slug === channelSlug;
                  return (
                    <li key={c.slug}>
                      <button
                        type="button"
                        onClick={() => setChannelSlug(c.slug)}
                        className={cn(
                          "flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm transition-colors",
                          active
                            ? "bg-secondary text-foreground"
                            : "text-foreground/75 hover:bg-muted hover:text-foreground"
                        )}
                      >
                        <span className="text-foreground/50">#</span>
                        <span className="truncate">{c.name}</span>
                      </button>
                    </li>
                  );
                })}
              </ul>
            )}
          </ChatSidebarSection>
          <ChatSidebarSection
            title="Leaderboard"
            className="border-t border-border/60"
          >
            <CommunityLeaderboardPanel skillSlug={skillSlug} />
          </ChatSidebarSection>
        </div>
      }
      messages={messages}
      empty={<Empty>No messages in this channel yet.</Empty>}
      threadKey={`${skillSlug}:${channelSlug}`}
      currentUsername={username}
      draft={draft}
      onDraftChange={setDraft}
      replyTo={replyTo}
      onClearReply={() => setReplyTo(null)}
      onReplyMessage={(m) => setReplyTo(m as Message)}
      composerDisabled={!channelSlug}
      composerPlaceholder={
        channelSlug
          ? `Message #${activeChannel?.name ?? channelSlug}`
          : "Select a channel"
      }
      onSend={async () => {
        if (!channelSlug) return;
        const replyId =
          replyTo && Number.isFinite(Number(replyTo.id))
            ? Number(replyTo.id)
            : null;
        const r = await sendChannelMessage(
          skillSlug,
          channelSlug,
          draft,
          replyId
        );
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
        setMessages((prev) =>
          prev.filter((m) => m.id !== id && String(m.id) !== String(id))
        );
      }}
      audio={
        <AudioRoomsSidebar
          rooms={rooms}
          from={from}
          canCreate={Boolean(channelSlug)}
          roomTitle={roomTitle}
          onRoomTitleChange={setRoomTitle}
          createDisabled={!channelSlug}
          onCreate={async () => {
            if (!channelSlug) return;
            const r = await createChannelAudioRoom(
              skillSlug,
              channelSlug,
              roomTitle
            );
            if (!r.ok) toastApiError(r);
            else void loadRooms(channelSlug);
          }}
        />
      }
    />
    </>
  );
}
