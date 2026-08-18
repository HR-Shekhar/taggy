"use client";

import { useEffect, useRef, useState } from "react";
import { Check, Flag, Pencil, Reply, Trash2, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ReportDialog } from "@/components/report-dialog";
import { cn } from "@/lib/utils";

export type ChatMessage = {
  id: number | string;
  author_username: string;
  content: string;
  created_at: string;
  edited_at?: string | null;
  reply_to_message_id?: number | null;
  reply_to_content?: string | null;
  reply_to_author_username?: string | null;
};

function formatTime(value: string) {
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return value;
  return d.toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

function truncateReply(content: string, max = 90) {
  const t = content.trim();
  if (t.length <= max) return t;
  return `${t.slice(0, max).trimEnd()}…`;
}

export function ChatPanel({
  messages,
  empty,
  className,
  threadKey,
  currentUsername,
  onEditMessage,
  onDeleteMessage,
  onReplyMessage,
  allowReport = true,
}: {
  messages: ChatMessage[];
  empty?: React.ReactNode;
  className?: string;
  /** Reset animation state when the conversation changes (pod / channel). */
  threadKey?: string;
  currentUsername?: string | null;
  onEditMessage?: (id: number | string, content: string) => void | Promise<void>;
  onDeleteMessage?: (id: number | string) => void | Promise<void>;
  onReplyMessage?: (message: ChatMessage) => void;
  allowReport?: boolean;
}) {
  const endRef = useRef<HTMLDivElement | null>(null);
  const listRef = useRef<HTMLDivElement | null>(null);
  const seenIdsRef = useRef<Set<string>>(new Set());
  const hydratedRef = useRef(false);
  const threadRef = useRef(threadKey);
  const [animatingIds, setAnimatingIds] = useState<Set<string>>(() => new Set());
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editDraft, setEditDraft] = useState("");
  const [busyId, setBusyId] = useState<string | null>(null);
  const [reportTargetId, setReportTargetId] = useState<number | null>(null);
  const [flashId, setFlashId] = useState<string | null>(null);

  useEffect(() => {
    if (threadRef.current !== threadKey) {
      threadRef.current = threadKey;
      seenIdsRef.current = new Set();
      hydratedRef.current = false;
      setAnimatingIds(new Set());
      setEditingId(null);
      setFlashId(null);
    }

    const ids = messages.map((m) => String(m.id));

    if (!hydratedRef.current) {
      for (const id of ids) seenIdsRef.current.add(id);
      hydratedRef.current = true;
      endRef.current?.scrollIntoView({ behavior: "auto", block: "end" });
      return;
    }

    const fresh: string[] = [];
    for (const id of ids) {
      if (!seenIdsRef.current.has(id)) {
        fresh.push(id);
        seenIdsRef.current.add(id);
      }
    }

    if (fresh.length === 0) return;

    setAnimatingIds(new Set(fresh));
    endRef.current?.scrollIntoView({ behavior: "smooth", block: "end" });

    const timer = window.setTimeout(() => {
      setAnimatingIds(new Set());
    }, 520);
    return () => window.clearTimeout(timer);
  }, [messages, threadKey]);

  function jumpToMessage(targetId: number | string) {
    const el = listRef.current?.querySelector(
      `[data-message-id="${CSS.escape(String(targetId))}"]`
    );
    if (!(el instanceof HTMLElement)) return;
    el.scrollIntoView({ behavior: "smooth", block: "center" });
    setFlashId(String(targetId));
    window.setTimeout(() => setFlashId(null), 1400);
  }

  if (messages.length === 0) {
    return (
      <div
        className={cn(
          "flex min-h-0 flex-1 items-center justify-center p-6 text-center",
          className
        )}
      >
        {empty}
      </div>
    );
  }

  const me = currentUsername?.toLowerCase() ?? "";
  const canManage = Boolean(onEditMessage || onDeleteMessage);

  return (
    <>
      <ReportDialog
        open={reportTargetId != null}
        onOpenChange={(open) => {
          if (!open) setReportTargetId(null);
        }}
        targetType="MESSAGE"
        targetId={reportTargetId}
        title="Report message"
        description="Report this message for spam, harassment, or other abuse."
      />

      <div
        ref={listRef}
        className={cn(
          "flex min-h-0 flex-1 flex-col gap-2.5 overflow-y-auto scroll-smooth p-3",
          className
        )}
      >
        {messages.map((m) => {
          const id = String(m.id);
          const isMine = Boolean(me) && m.author_username.toLowerCase() === me;
          const isNew = animatingIds.has(id);
          const isEditing = editingId === id;
          const isFlash = flashId === id;
          const numericId = typeof m.id === "number" ? m.id : Number(m.id);
          const canReport =
            allowReport &&
            !isMine &&
            Number.isFinite(numericId) &&
            numericId > 0;
          const canReply =
            Boolean(onReplyMessage) &&
            Number.isFinite(numericId) &&
            numericId > 0;
          const showActions =
            !isEditing && ((isMine && canManage) || canReport || canReply);

          return (
            <div
              key={m.id}
              data-message-id={id}
              className={cn(
                "group/msg relative flex w-full",
                isMine ? "justify-end" : "justify-start"
              )}
            >
              <div
                className={cn(
                  "relative max-w-[min(32rem,85%)] rounded-2xl px-3.5 py-2 shadow-sm ring-1 backdrop-blur-sm transition-[box-shadow,background-color] duration-300",
                  isMine
                    ? "rounded-br-md bg-primary/15 ring-primary/20"
                    : "rounded-bl-md bg-background/80 ring-foreground/8",
                  isFlash && "ring-2 ring-primary/70 shadow-[0_0_0_4px] shadow-primary/15",
                  isNew &&
                    (isMine ? "chat-msg-enter-right" : "chat-msg-enter-left")
                )}
              >
                {showActions ? (
                  <div
                    className={cn(
                      "absolute -top-3 z-10 flex items-center gap-0.5 rounded-lg border border-foreground/15 bg-card/95 p-0.5 shadow-md backdrop-blur-md",
                      "opacity-0 transition-opacity duration-150 group-hover/msg:opacity-100 group-focus-within/msg:opacity-100",
                      isMine ? "right-2" : "left-2"
                    )}
                  >
                    {canReply ? (
                      <Button
                        type="button"
                        size="icon-xs"
                        variant="ghost"
                        aria-label="Reply to message"
                        onClick={() => onReplyMessage?.(m)}
                      >
                        <Reply className="size-3" />
                      </Button>
                    ) : null}
                    {isMine && onEditMessage ? (
                      <Button
                        type="button"
                        size="icon-xs"
                        variant="ghost"
                        disabled={busyId === id}
                        aria-label="Edit message"
                        onClick={() => {
                          setEditingId(id);
                          setEditDraft(m.content);
                        }}
                      >
                        <Pencil className="size-3" />
                      </Button>
                    ) : null}
                    {isMine && onDeleteMessage ? (
                      <Button
                        type="button"
                        size="icon-xs"
                        variant="ghost"
                        className="text-destructive hover:text-destructive"
                        disabled={busyId === id}
                        aria-label="Delete message"
                        onClick={async () => {
                          if (
                            !window.confirm(
                              "Delete this message? This cannot be undone."
                            )
                          ) {
                            return;
                          }
                          setBusyId(id);
                          await onDeleteMessage(m.id);
                          setBusyId(null);
                        }}
                      >
                        <Trash2 className="size-3" />
                      </Button>
                    ) : null}
                    {canReport ? (
                      <Button
                        type="button"
                        size="icon-xs"
                        variant="ghost"
                        aria-label="Report message"
                        onClick={() => setReportTargetId(numericId)}
                      >
                        <Flag className="size-3" />
                      </Button>
                    ) : null}
                  </div>
                ) : null}

                <div
                  className={cn(
                    "mb-0.5 flex items-center gap-2 text-[11px] text-muted-foreground",
                    isMine && "justify-end"
                  )}
                >
                  <span>
                    {isMine ? "You" : `@${m.author_username}`}
                    <span className="opacity-70">
                      {" "}
                      · {formatTime(m.created_at)}
                      {m.edited_at ? " · edited" : ""}
                    </span>
                  </span>
                </div>

                {m.reply_to_message_id ? (
                  <button
                    type="button"
                    onClick={() => jumpToMessage(m.reply_to_message_id!)}
                    className={cn(
                      "mb-2 flex w-full cursor-pointer gap-0 overflow-hidden rounded-md text-left transition-colors",
                      "bg-black/5 hover:bg-black/10 dark:bg-white/5 dark:hover:bg-white/10"
                    )}
                    aria-label={`Jump to reply from @${m.reply_to_author_username ?? "user"}`}
                  >
                    <span className="w-1 shrink-0 self-stretch bg-primary" />
                    <span className="min-w-0 flex-1 px-2.5 py-1.5">
                      <span className="block text-[11px] font-semibold text-primary">
                        @{m.reply_to_author_username ?? "unknown"}
                      </span>
                      <span className="mt-0.5 block line-clamp-2 text-[11px] leading-snug text-muted-foreground">
                        {truncateReply(m.reply_to_content ?? "Original message")}
                      </span>
                    </span>
                  </button>
                ) : null}

                {isEditing ? (
                  <form
                    className="mt-1 space-y-2"
                    onSubmit={async (e) => {
                      e.preventDefault();
                      if (!onEditMessage || !editDraft.trim()) return;
                      setBusyId(id);
                      await onEditMessage(m.id, editDraft.trim());
                      setBusyId(null);
                      setEditingId(null);
                    }}
                  >
                    <Input
                      value={editDraft}
                      onChange={(e) => setEditDraft(e.target.value)}
                      autoFocus
                      required
                    />
                    <div className="flex justify-end gap-1">
                      <Button
                        type="button"
                        size="icon-xs"
                        variant="ghost"
                        disabled={busyId === id}
                        onClick={() => setEditingId(null)}
                      >
                        <X className="size-3.5" />
                      </Button>
                      <Button
                        type="submit"
                        size="icon-xs"
                        disabled={busyId === id || !editDraft.trim()}
                      >
                        <Check className="size-3.5" />
                      </Button>
                    </div>
                  </form>
                ) : (
                  <div className="text-sm leading-relaxed whitespace-pre-wrap break-words">
                    {m.content}
                  </div>
                )}
              </div>
            </div>
          );
        })}
        <div ref={endRef} />
      </div>
    </>
  );
}
