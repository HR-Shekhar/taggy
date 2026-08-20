"use client";

import Link from "next/link";
import { useState, type FormEvent, type ReactNode } from "react";
import {
  Headphones,
  PanelLeftClose,
  PanelLeftOpen,
  PanelRightClose,
  PanelRightOpen,
  Send,
  X,
} from "lucide-react";
import { ChatPanel, type ChatMessage } from "@/components/chat-panel";
import { Empty } from "@/components/app-ui";
import { Badge } from "@/components/ui/badge";
import { Button, buttonVariants } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

export type AudioRoomItem = {
  id: string;
  title: string;
  status?: string;
  host_username?: string;
};

export function ChatWorkspace({
  header,
  left,
  leftLabel = "Sidebar",
  messages,
  empty,
  threadKey,
  currentUsername,
  draft,
  onDraftChange,
  onSend,
  onEditMessage,
  onDeleteMessage,
  onReplyMessage,
  replyTo,
  onClearReply,
  composerDisabled,
  composerPlaceholder = "Write a message",
  audio,
  className,
}: {
  header: ReactNode;
  left?: ReactNode;
  /** Label shown on the collapsed left rail (e.g. Channels / Members). */
  leftLabel?: string;
  messages: ChatMessage[];
  empty?: ReactNode;
  threadKey?: string;
  currentUsername?: string | null;
  draft: string;
  onDraftChange: (value: string) => void;
  onSend: () => void | Promise<void>;
  onEditMessage?: (id: number | string, content: string) => void | Promise<void>;
  onDeleteMessage?: (id: number | string) => void | Promise<void>;
  onReplyMessage?: (message: ChatMessage) => void;
  replyTo?: ChatMessage | null;
  onClearReply?: () => void;
  composerDisabled?: boolean;
  composerPlaceholder?: string;
  audio: ReactNode;
  className?: string;
}) {
  const [leftOpen, setLeftOpen] = useState(true);
  const [rightOpen, setRightOpen] = useState(true);

  return (
    <div
      className={cn(
        "flex h-full min-h-0 flex-1 flex-col overflow-hidden bg-background",
        className
      )}
    >
      <header className="flex shrink-0 items-center gap-3 border-b border-border bg-card px-4 py-3">
        <div className="flex min-w-0 flex-1 items-center gap-2">
          {left ? (
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              className="hidden shrink-0 md:inline-flex"
              aria-label={leftOpen ? `Hide ${leftLabel}` : `Show ${leftLabel}`}
              aria-pressed={leftOpen}
              onClick={() => setLeftOpen((v) => !v)}
            >
              {leftOpen ? (
                <PanelLeftClose className="size-4" />
              ) : (
                <PanelLeftOpen className="size-4" />
              )}
            </Button>
          ) : null}
          <div className="min-w-0 flex-1">{header}</div>
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            className="hidden shrink-0 lg:inline-flex"
            aria-label={rightOpen ? "Hide audio rooms" : "Show audio rooms"}
            aria-pressed={rightOpen}
            onClick={() => setRightOpen((v) => !v)}
          >
            {rightOpen ? (
              <PanelRightClose className="size-4" />
            ) : (
              <PanelRightOpen className="size-4" />
            )}
          </Button>
        </div>
      </header>

      <div className="flex min-h-0 flex-1">
        {left ? (
          <aside
            className={cn(
              "hidden shrink-0 flex-col overflow-hidden border-r border-border bg-card transition-[width] duration-300 ease-out md:flex",
              leftOpen ? "w-56 lg:w-60" : "w-11"
            )}
          >
            {leftOpen ? (
              left
            ) : (
              <button
                type="button"
                className="flex h-full w-full flex-col items-center gap-3 px-1 py-4 text-muted-foreground hover:bg-muted/50 hover:text-foreground"
                onClick={() => setLeftOpen(true)}
                aria-label={`Expand ${leftLabel}`}
              >
                <PanelLeftOpen className="size-4 shrink-0" />
                <span
                  className="text-[10px] font-medium uppercase tracking-[0.18em]"
                  style={{ writingMode: "vertical-rl", transform: "rotate(180deg)" }}
                >
                  {leftLabel}
                </span>
              </button>
            )}
          </aside>
        ) : null}

        <section className="flex min-w-0 flex-1 flex-col">
          <ChatPanel
            messages={messages}
            empty={empty ?? <Empty>No messages yet.</Empty>}
            threadKey={threadKey}
            currentUsername={currentUsername}
            onEditMessage={onEditMessage}
            onDeleteMessage={onDeleteMessage}
            onReplyMessage={onReplyMessage}
            className="min-h-0 flex-1 rounded-none border-0 bg-transparent p-4"
          />

          <form
            className="shrink-0 border-t border-border bg-card p-3"
            onSubmit={async (e: FormEvent) => {
              e.preventDefault();
              if (composerDisabled) return;
              await onSend();
            }}
          >
            {replyTo ? (
              <div className="mb-2 flex overflow-hidden rounded-lg border border-foreground/15 bg-muted/50 shadow-sm">
                <div className="w-1 shrink-0 bg-primary" />
                <div className="flex min-w-0 flex-1 items-start justify-between gap-2 px-3 py-2">
                  <div className="min-w-0">
                    <div className="text-xs font-semibold text-primary">
                      Replying to @{replyTo.author_username}
                    </div>
                    <div className="mt-0.5 truncate text-xs text-muted-foreground">
                      {replyTo.content}
                    </div>
                  </div>
                  <Button
                    type="button"
                    size="icon-xs"
                    variant="outline"
                    aria-label="Cancel reply"
                    onClick={() => onClearReply?.()}
                  >
                    <X className="size-3.5" />
                  </Button>
                </div>
              </div>
            ) : null}
            <div className="flex gap-2">
              <Input
                value={draft}
                onChange={(e) => onDraftChange(e.target.value)}
                placeholder={
                  replyTo
                    ? `Reply to @${replyTo.author_username}`
                    : composerPlaceholder
                }
                disabled={composerDisabled}
                required={!composerDisabled}
                className="h-10"
              />
              <Button
                type="submit"
                size="icon"
                className="size-10 shrink-0"
                disabled={composerDisabled}
              >
                <Send className="size-4" />
              </Button>
            </div>
          </form>
        </section>

        <aside
          className={cn(
            "hidden shrink-0 flex-col overflow-hidden border-l border-border bg-card transition-[width] duration-300 ease-out lg:flex",
            rightOpen ? "w-64 xl:w-72" : "w-11"
          )}
        >
          {rightOpen ? (
            audio
          ) : (
            <button
              type="button"
              className="flex h-full w-full flex-col items-center gap-3 px-1 py-4 text-muted-foreground hover:bg-muted/50 hover:text-foreground"
              onClick={() => setRightOpen(true)}
              aria-label="Expand audio rooms"
            >
              <Headphones className="size-4 shrink-0" />
              <span
                className="text-[10px] font-medium uppercase tracking-[0.18em]"
                style={{ writingMode: "vertical-rl" }}
              >
                Audio
              </span>
            </button>
          )}
        </aside>
      </div>

      {/* Mobile / tablet audio strip */}
      <div className="max-h-52 shrink-0 overflow-y-auto border-t border-border bg-card lg:hidden">
        {audio}
      </div>
    </div>
  );
}

export function ChatSidebarSection({
  title,
  description,
  children,
  className,
}: {
  title: string;
  description?: string;
  children: ReactNode;
  className?: string;
}) {
  return (
    <div className={cn("flex min-h-0 flex-1 flex-col", className)}>
      <div className="shrink-0 border-b border-border/60 px-3 py-3">
        <h2 className="text-xs font-medium uppercase tracking-[0.16em] text-muted-foreground">
          {title}
        </h2>
        {description ? (
          <p className="mt-1 text-xs text-muted-foreground/90">{description}</p>
        ) : null}
      </div>
      <div className="min-h-0 flex-1 overflow-y-auto p-3">{children}</div>
    </div>
  );
}

export function AudioRoomsSidebar({
  rooms,
  from,
  canCreate,
  roomTitle,
  onRoomTitleChange,
  onCreate,
  createDisabled,
  emptyHint = "No live rooms right now.",
}: {
  rooms: AudioRoomItem[];
  from: string;
  canCreate?: boolean;
  roomTitle: string;
  onRoomTitleChange: (value: string) => void;
  onCreate: () => void | Promise<void>;
  createDisabled?: boolean;
  emptyHint?: string;
}) {
  return (
    <ChatSidebarSection
      title="Audio rooms"
      description="Start or join a live call."
    >
      <div className="space-y-3">
        {canCreate ? (
          <form
            className="space-y-2"
            onSubmit={async (e) => {
              e.preventDefault();
              await onCreate();
            }}
          >
            <Input
              value={roomTitle}
              onChange={(e) => onRoomTitleChange(e.target.value)}
              placeholder="Room title"
              required
              disabled={createDisabled}
            />
            <Button type="submit" className="w-full" disabled={createDisabled}>
              Create room
            </Button>
          </form>
        ) : null}

        {rooms.length === 0 ? (
          <p className="text-sm text-muted-foreground">{emptyHint}</p>
        ) : (
          <ul className="space-y-2">
            {rooms.map((r) => (
              <li
                key={r.id}
                className="rounded-lg border border-border/70 bg-background/60 p-3"
              >
                <div className="mb-2 flex items-start justify-between gap-2">
                  <div className="min-w-0">
                    <div className="truncate font-medium">{r.title}</div>
                    {r.host_username ? (
                      <div className="text-xs text-muted-foreground">
                        @{r.host_username}
                      </div>
                    ) : null}
                  </div>
                  <Badge variant="secondary" className="shrink-0">
                    Live
                  </Badge>
                </div>
                <Link
                  href={`/audio-rooms/${r.id}?from=${encodeURIComponent(from)}`}
                  className={cn(buttonVariants({ size: "sm" }), "w-full gap-1")}
                >
                  <Headphones className="size-3.5" />
                  Join room
                </Link>
              </li>
            ))}
          </ul>
        )}
      </div>
    </ChatSidebarSection>
  );
}
