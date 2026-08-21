"use client";

import Link from "next/link";
import { Suspense, useEffect, useMemo, useState } from "react";
import { useSearchParams } from "next/navigation";
import {
  CheckCircle2,
  ChevronDown,
  ChevronRight,
  Circle,
  Clock3,
  Flame,
  PauseCircle,
  Timer,
} from "lucide-react";
import { useAuth } from "@/lib/auth";
import {
  getProgressSummary,
  getStreak,
  listMilestones,
  listMyPods,
  listMySkills,
  listStudySessions,
  logStudySession,
  updateMilestone,
  type MyPod,
  type MySkill,
  type ProgressSummary,
} from "@/lib/api";
import { Empty, PageHeader, PageSkeleton } from "@/components/app-ui";
import { PodQuizPanel } from "@/components/pod-quiz-panel";
import { toastApiError, toastSuccess } from "@/lib/toast";
import { Badge } from "@/components/ui/badge";
import { Button, buttonVariants } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";

type StudySession = {
  skill_slug?: string;
  duration_minutes?: number;
  studied_at?: string;
  notes?: string;
};

type Streak = {
  current_streak: number;
  longest_streak: number;
  last_activity_date?: string | null;
};

type ProgressMilestone = {
  slug: string;
  title: string;
  description?: string;
  order_index: number;
  status: string;
  estimated_hours?: number | null;
  difficulty?: string | null;
  chapter?: string | null;
  kind?: string;
  postponed_until?: string | null;
  completed_at?: string | null;
};

function postponeUntilISO(daysFromNow: number) {
  const d = new Date();
  d.setUTCDate(d.getUTCDate() + daysFromNow);
  d.setUTCHours(0, 0, 0, 0);
  return d.toISOString();
}

function formatWhen(value?: string | null) {
  if (!value) return null;
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return value;
  return d.toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function StatusIcon({ status, size = "md" }: { status: string; size?: "sm" | "md" }) {
  const cls =
    size === "sm"
      ? "relative z-10 size-5 shrink-0 bg-background"
      : "relative z-10 size-8 shrink-0 bg-background";
  if (status === "COMPLETED") {
    return <CheckCircle2 className={cn(cls, "text-primary")} />;
  }
  if (status === "POSTPONED") {
    return <PauseCircle className={cn(cls, "text-muted-foreground")} />;
  }
  if (status === "IN_PROGRESS") {
    return <Clock3 className={cn(cls, "text-primary")} />;
  }
  return <Circle className={cn(cls, "text-muted-foreground")} />;
}

type ChapterGroup = {
  key: string;
  title: string;
  chapter: ProgressMilestone | null;
  topics: ProgressMilestone[];
  allComplete: boolean;
};

function groupMilestones(list: ProgressMilestone[]): ChapterGroup[] {
  const groups: ChapterGroup[] = [];
  const byTitle = new Map<string, ChapterGroup>();

  for (const m of list) {
    if (m.kind === "CHAPTER") {
      const title = m.chapter || m.title;
      const group: ChapterGroup = {
        key: m.slug,
        title,
        chapter: m,
        topics: [],
        allComplete: false,
      };
      groups.push(group);
      byTitle.set(title, group);
      continue;
    }

    const title = m.chapter || "Other";
    let group = byTitle.get(title);
    if (!group) {
      group = {
        key: `orphan-${title}`,
        title,
        chapter: null,
        topics: [],
        allComplete: false,
      };
      groups.push(group);
      byTitle.set(title, group);
    }
    group.topics.push(m);
  }

  for (const g of groups) {
    const chapterDone = !g.chapter || g.chapter.status === "COMPLETED";
    const topicsDone =
      g.topics.length === 0 || g.topics.every((t) => t.status === "COMPLETED");
    g.allComplete = chapterDone && topicsDone;
  }
  return groups;
}

/** Show the latest completed topic + all incomplete topics (keeps focus near current work). */
function visibleChapterGroups(groups: ChapterGroup[]): ChapterGroup[] {
  const incomplete = groups.filter((g) => !g.allComplete);
  const completed = groups.filter((g) => g.allComplete);
  const latestDone =
    completed.length > 0 ? [completed[completed.length - 1]] : [];
  return [...latestDone, ...incomplete];
}

function priorTopicsComplete(
  list: ProgressMilestone[],
  target: ProgressMilestone
): boolean {
  return list
    .filter(
      (m) =>
        (m.kind ?? "TOPIC") !== "CHAPTER" && m.order_index < target.order_index
    )
    .every((m) => m.status === "COMPLETED");
}

function priorChaptersComplete(groups: ChapterGroup[], groupKey: string): boolean {
  for (const g of groups) {
    if (g.key === groupKey) return true;
    if (!g.allComplete) return false;
  }
  return true;
}

function ProgressInner() {
  const { username } = useAuth();
  const params = useSearchParams();
  const [skills, setSkills] = useState<MySkill[]>([]);
  const [pods, setPods] = useState<MyPod[]>([]);
  const [sessions, setSessions] = useState<StudySession[]>([]);
  const [milestones, setMilestones] = useState<ProgressMilestone[]>([]);
  const [streak, setStreak] = useState<Streak | null>(null);
  const [summary, setSummary] = useState<ProgressSummary | null>(null);
  const [skillSlug, setSkillSlug] = useState(params.get("skill") ?? "");
  const [minutes, setMinutes] = useState("30");
  const [notes, setNotes] = useState("");
  const [busy, setBusy] = useState(false);
  const [loading, setLoading] = useState(true);
  const [expandedDone, setExpandedDone] = useState<Record<string, boolean>>({});

  async function load(opts?: { quiet?: boolean }) {
    if (!username) return;
    if (!opts?.quiet) setLoading(true);
    const [sk, se, st, sm, mp] = await Promise.all([
      listMySkills(username),
      listStudySessions(username),
      getStreak(username),
      getProgressSummary(username),
      listMyPods(username),
    ]);
    if (sk.ok) {
      const list = sk.data ?? [];
      setSkills(list);
      if (!opts?.quiet) {
        const next = skillSlug || params.get("skill") || list[0]?.skill_slug || "";
        setSkillSlug(next);
      }
    }
    if (se.ok) {
      setSessions(
        Array.isArray(se.data)
          ? (se.data as StudySession[])
          : ((se.data as { sessions?: StudySession[] })?.sessions ?? [])
      );
    }
    if (st.ok) setStreak(st.data as Streak);
    if (sm.ok) setSummary(sm.data);
    if (mp.ok) {
      setPods(
        Array.isArray(mp.data)
          ? mp.data
          : ((mp.data as { pods?: MyPod[] })?.pods ?? [])
      );
    }
    if (!opts?.quiet) setLoading(false);
  }

  async function loadMilestones(slug: string) {
    if (!username || !slug) {
      setMilestones([]);
      return;
    }
    const ms = await listMilestones(username, slug);
    if (ms.ok) setMilestones((ms.data as ProgressMilestone[]) ?? []);
    else {
      setMilestones([]);
      toastApiError(ms);
    }
  }

  function applyLocalMilestone(
    slug: string,
    patch: Partial<ProgressMilestone>
  ) {
    setMilestones((prev) =>
      prev.map((m) => (m.slug === slug ? { ...m, ...patch } : m))
    );
  }

  async function completeMilestone(m: ProgressMilestone) {
    if (!username) return;
    setBusy(true);
    const result = await updateMilestone(
      username,
      skillSlug,
      m.slug,
      "COMPLETE"
    );
    setBusy(false);
    if (!result.ok) {
      toastApiError(result);
      return;
    }
    applyLocalMilestone(m.slug, {
      status: "COMPLETED",
      completed_at: new Date().toISOString(),
      postponed_until: null,
    });
    const isChapter = m.kind === "CHAPTER";
    toastSuccess(
      isChapter
        ? `You finished “${m.title}” and all of its subtopics.`
        : `You completed “${m.title}”. Keep going on the next one.`,
      isChapter ? "Topic complete!" : "Subtopic complete!"
    );
    void load({ quiet: true });
  }

  async function postponeMilestone(m: ProgressMilestone, days: number) {
    if (!username) return;
    setBusy(true);
    const until = postponeUntilISO(days);
    const result = await updateMilestone(
      username,
      skillSlug,
      m.slug,
      "POSTPONE",
      until
    );
    setBusy(false);
    if (!result.ok) {
      toastApiError(result);
      return;
    }
    applyLocalMilestone(m.slug, {
      status: "POSTPONED",
      postponed_until: until,
    });
  }

  useEffect(() => {
    void load();
  }, [username]);

  useEffect(() => {
    void loadMilestones(skillSlug);
  }, [username, skillSlug]);

  const selectedSkill = skills.find((s) => s.skill_slug === skillSlug);
  const acceptedPod = useMemo(() => {
    return pods.find(
      (p) => p.skill_slug === skillSlug && (p.status ?? "ACCEPTED") === "ACCEPTED"
    );
  }, [pods, skillSlug]);
  const podSlug = acceptedPod?.pod_slug ?? acceptedPod?.slug ?? "";

  const chapterGroups = useMemo(
    () => groupMilestones(milestones),
    [milestones]
  );
  const roadmapGroups = useMemo(
    () => visibleChapterGroups(chapterGroups),
    [chapterGroups]
  );

  const completion = useMemo(() => {
    if (milestones.length === 0) return 0;
    const done = milestones.filter((m) => m.status === "COMPLETED").length;
    return (done / milestones.length) * 100;
  }, [milestones]);

  if (loading) return <PageSkeleton variant="dashboard" />;

  const formatMinutes = (mins: number) => {
    if (mins < 60) return `${mins}m`;
    const h = Math.floor(mins / 60);
    const m = mins % 60;
    return m ? `${h}h ${m}m` : `${h}h`;
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title="Progress"
        description="Complete milestones, take pod quizzes, and log study time."
      />
      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardDescription className="flex items-center gap-2">
              <Flame className="size-4 text-primary" /> Current streak
            </CardDescription>
            <CardTitle className="font-serif text-3xl">
              {streak?.current_streak ?? 0}
            </CardTitle>
          </CardHeader>
          <CardContent className="text-sm text-foreground/75">
            Best {streak?.longest_streak ?? 0} days
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription className="flex items-center gap-2">
              <Timer className="size-4 text-primary" /> This week
            </CardDescription>
            <CardTitle className="font-serif text-3xl">
              {formatMinutes(summary?.weekly_minutes ?? 0)}
            </CardTitle>
          </CardHeader>
          <CardContent className="text-sm text-foreground/75">
            {formatMinutes(summary?.monthly_minutes ?? 0)} this month
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Total studied</CardDescription>
            <CardTitle className="font-serif text-3xl">
              {formatMinutes(summary?.total_minutes ?? 0)}
            </CardTitle>
          </CardHeader>
          <CardContent className="text-sm text-foreground/75">
            Across all enrolled skills
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Last activity</CardDescription>
            <CardTitle className="font-serif text-xl">
              {streak?.last_activity_date ?? "—"}
            </CardTitle>
          </CardHeader>
          <CardContent className="text-sm text-foreground/75">
            {skills.length} active skill{skills.length === 1 ? "" : "s"}
          </CardContent>
        </Card>
      </div>

      <div className="flex flex-wrap items-end justify-between gap-3 rounded-xl border border-border/70 bg-card px-4 py-3">
        <div className="space-y-2 min-w-[12rem] flex-1 max-w-md">
          <Label htmlFor="progress-skill">Skill</Label>
          <select
            id="progress-skill"
            className="h-8 w-full rounded-lg border border-foreground/25 bg-background px-2.5 text-sm"
            value={skillSlug}
            onChange={(e) => {
              setSkillSlug(e.target.value);
            }}
          >
            {skills.length === 0 && (
              <option value="">Join a skill first</option>
            )}
            {skills.map((s) => (
              <option key={s.skill_slug} value={s.skill_slug}>
                {s.skill_name}
              </option>
            ))}
          </select>
        </div>
        {selectedSkill ? (
          <div className="min-w-[14rem] flex-1 space-y-2">
            <div className="flex items-center justify-between text-sm">
              <span className="text-foreground/75">
                v{selectedSkill.roadmap_version_number} ·{" "}
                {selectedSkill.completed_count}/{selectedSkill.milestone_count}{" "}
                complete
              </span>
              <span className="font-medium">{Math.round(completion)}%</span>
            </div>
            <div className="h-2 overflow-hidden rounded-full bg-muted">
              <div
                className="h-full rounded-full bg-primary"
                style={{ width: `${Math.min(100, completion)}%` }}
              />
            </div>
          </div>
        ) : null}
        {skillSlug ? (
          <Link
            href={`/skills/${skillSlug}`}
            className={cn(buttonVariants({ variant: "outline", size: "sm" }))}
          >
            Skill details
          </Link>
        ) : null}
      </div>

      <div className="grid gap-6 lg:grid-cols-2 lg:items-start">
        <Card className=" lg:min-h-[36rem]">
          <CardHeader>
            <CardTitle className="font-serif text-lg">Roadmap</CardTitle>
            <CardDescription>
              Finish subtopics first, then mark the topic complete. Showing your
              latest finished topic plus what&apos;s left.
            </CardDescription>
          </CardHeader>
          <CardContent className="max-h-[70vh] space-y-4 overflow-y-auto pr-1">
            {!skillSlug || skills.length === 0 ? (
              <Empty>Join a skill to track milestones.</Empty>
            ) : milestones.length === 0 ? (
              <Empty>No milestones on your roadmap version.</Empty>
            ) : roadmapGroups.length === 0 ? (
              <Empty>All topics complete — great work.</Empty>
            ) : (
              <div className="space-y-5">
                {roadmapGroups.map((group) => {
                  const chapter = group.chapter;
                  const topicsDone =
                    group.topics.length === 0 ||
                    group.topics.every((t) => t.status === "COMPLETED");
                  const canFinishChapter =
                    Boolean(chapter) &&
                    chapter!.status !== "COMPLETED" &&
                    topicsDone &&
                    priorChaptersComplete(chapterGroups, group.key);
                  const chapterPostponed = formatWhen(
                    chapter?.postponed_until
                  );
                  const isMinimized =
                    group.allComplete && !expandedDone[group.key];

                  if (isMinimized) {
                    return (
                      <button
                        key={group.key}
                        type="button"
                        onClick={() =>
                          setExpandedDone((prev) => ({
                            ...prev,
                            [group.key]: true,
                          }))
                        }
                        className="flex w-full items-center justify-between gap-3 rounded-xl border border-primary/20 bg-primary/5 px-3 py-2.5 text-left transition-colors hover:bg-primary/10"
                      >
                        <span className="flex min-w-0 items-center gap-2">
                          <ChevronRight className="size-4 shrink-0 text-foreground/50" />
                          <CheckCircle2 className="size-4 shrink-0 text-primary" />
                          <span className="truncate font-medium">
                            {group.title}
                          </span>
                        </span>
                        <span className="flex shrink-0 items-center gap-1.5">
                          <Badge variant="outline" className="text-[10px]">
                            {group.topics.length} subtopics
                          </Badge>
                          <Badge className="text-[10px]">Done</Badge>
                        </span>
                      </button>
                    );
                  }

                  return (
                    <section
                      key={group.key}
                      className={cn(
                        "rounded-xl border p-4",
                        group.allComplete
                          ? "border-primary/25 bg-primary/5"
                          : "border-border bg-card"
                      )}
                    >
                      <div className="flex flex-wrap items-start justify-between gap-3">
                        <div className="min-w-0 space-y-1">
                          <div className="flex items-center gap-2">
                            {group.allComplete ? (
                              <button
                                type="button"
                                aria-label="Minimize topic"
                                className="rounded-md p-0.5 text-muted-foreground hover:bg-muted hover:text-foreground"
                                onClick={() =>
                                  setExpandedDone((prev) => ({
                                    ...prev,
                                    [group.key]: false,
                                  }))
                                }
                              >
                                <ChevronDown className="size-4" />
                              </button>
                            ) : null}
                            {chapter ? (
                              <StatusIcon status={chapter.status} />
                            ) : (
                              <StatusIcon
                                status={topicsDone ? "COMPLETED" : "NOT_STARTED"}
                              />
                            )}
                            <div>
                              <p className="text-[10px] font-medium uppercase tracking-[0.16em] text-foreground/60">
                                Topic
                              </p>
                              <h3 className="font-serif text-lg leading-tight">
                                {group.title}
                              </h3>
                            </div>
                          </div>
                          {chapter?.description ? (
                            <p className="pl-10 text-sm text-foreground/75">
                              {chapter.description}
                            </p>
                          ) : null}
                          {chapterPostponed ? (
                            <p className="pl-10 text-xs text-foreground/75">
                              Postponed until {chapterPostponed}
                            </p>
                          ) : null}
                        </div>
                        <div className="flex flex-wrap items-center gap-1.5">
                          {chapter?.estimated_hours ? (
                            <Badge variant="outline">
                              {chapter.estimated_hours}h total
                            </Badge>
                          ) : null}
                          <Badge variant="outline">
                            {
                              group.topics.filter((t) => t.status === "COMPLETED")
                                .length
                            }
                            /{group.topics.length} subtopics
                          </Badge>
                          {group.allComplete ? (
                            <Badge>Done</Badge>
                          ) : chapter?.status === "COMPLETED" ? (
                            <Badge variant="secondary">Topic marked</Badge>
                          ) : null}
                        </div>
                      </div>

                      {group.topics.length > 0 ? (
                        <ul className="mt-4 space-y-2 border-l-2 border-border/70 ml-4 pl-4">
                          {group.topics.map((t) => {
                            const canComplete =
                              t.status !== "COMPLETED" &&
                              priorTopicsComplete(milestones, t);
                            const postponedLabel = formatWhen(t.postponed_until);
                            return (
                              <li
                                key={t.slug}
                                className="rounded-lg border border-border/60 bg-background/60 px-3 py-2.5"
                              >
                                <div className="flex flex-wrap items-start justify-between gap-2">
                                  <div className="flex min-w-0 items-start gap-2">
                                    <StatusIcon status={t.status} size="sm" />
                                    <div className="min-w-0">
                                      <p className="text-[10px] uppercase tracking-wide text-foreground/60">
                                        Subtopic
                                      </p>
                                      <p className="font-medium leading-snug">
                                        {t.title}
                                      </p>
                                      {t.description ? (
                                        <p className="mt-0.5 text-sm text-foreground/75">
                                          {t.description}
                                        </p>
                                      ) : null}
                                      {postponedLabel ? (
                                        <p className="mt-0.5 text-xs text-foreground/75">
                                          Postponed until {postponedLabel}
                                        </p>
                                      ) : null}
                                    </div>
                                  </div>
                                  <div className="flex shrink-0 flex-col items-end gap-1">
                                    {t.estimated_hours ? (
                                      <Badge variant="outline" className="text-[10px]">
                                        {t.estimated_hours}h
                                      </Badge>
                                    ) : null}
                                    <Badge variant="outline" className="shrink-0">
                                      {t.status.replaceAll("_", " ")}
                                    </Badge>
                                  </div>
                                </div>
                                {t.status !== "COMPLETED" && username ? (
                                  <div className="mt-2 flex flex-wrap gap-2 pl-7">
                                    {canComplete ? (
                                      <Button
                                        type="button"
                                        size="sm"
                                        disabled={busy}
                                        onClick={() => void completeMilestone(t)}
                                      >
                                        Mark complete
                                      </Button>
                                    ) : null}
                                    <Button
                                      type="button"
                                      size="sm"
                                      variant="outline"
                                      disabled={busy}
                                      onClick={() =>
                                        void postponeMilestone(t, 7)
                                      }
                                    >
                                      Postpone 7d
                                    </Button>
                                  </div>
                                ) : null}
                              </li>
                            );
                          })}
                        </ul>
                      ) : (
                        <p className="mt-3 text-sm text-foreground/75">
                          No subtopics under this topic.
                        </p>
                      )}

                      {chapter &&
                      chapter.status !== "COMPLETED" &&
                      username ? (
                        <div className="mt-4 flex flex-wrap items-center gap-2 border-t border-border/60 pt-3">
                          {canFinishChapter ? (
                            <Button
                              type="button"
                              size="sm"
                              disabled={busy}
                              onClick={() => void completeMilestone(chapter)}
                            >
                              Finish topic
                            </Button>
                          ) : (
                            <p className="text-xs text-foreground/75">
                              Complete every subtopic before finishing this
                              topic.
                            </p>
                          )}
                          <Button
                            type="button"
                            size="sm"
                            variant="ghost"
                            disabled={busy}
                            onClick={() => void postponeMilestone(chapter, 7)}
                          >
                            Postpone topic 7d
                          </Button>
                        </div>
                      ) : null}
                    </section>
                  );
                })}
              </div>
            )}
          </CardContent>
        </Card>

        <div className="flex flex-col gap-6">
          <Card>
            <CardHeader>
              <CardTitle className="font-serif text-lg">Pod quiz</CardTitle>
              <CardDescription>
                AI questions from your completed topics. Needs an accepted pod.
              </CardDescription>
            </CardHeader>
            <CardContent>
              {!skillSlug ? (
                <Empty>Select a skill first.</Empty>
              ) : !podSlug ? (
                <div className="space-y-3">
                  <p className="text-sm text-foreground/75">
                    You&apos;re not in a pod for this skill yet.
                  </p>
                  <Link
                    href={`/pods?skill=${skillSlug}`}
                    className={cn(buttonVariants({ size: "sm" }))}
                  >
                    Find or create a pod
                  </Link>
                </div>
              ) : (
                <div className="space-y-2">
                  <p className="text-xs text-foreground/75">
                    Pod:{" "}
                    <Link
                      href={`/pods/${podSlug}`}
                      className="font-medium text-foreground hover:text-primary"
                    >
                      {acceptedPod?.pod_name ?? acceptedPod?.name ?? podSlug}
                    </Link>
                  </p>
                  <PodQuizPanel
                    podSlug={podSlug}
                    enabled
                    mode="quiz"
                  />
                </div>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="font-serif text-lg">
                Log study session
              </CardTitle>
              <CardDescription>
                Minutes count toward streaks and summaries.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form
                className="grid gap-4"
                onSubmit={async (e) => {
                  e.preventDefault();
                  if (!username) return;
                  setBusy(true);
                  const result = await logStudySession(username, {
                    skill_slug: skillSlug,
                    duration_minutes: Number(minutes),
                    notes: notes || undefined,
                  });
                  setBusy(false);
                  if (!result.ok) {
                    toastApiError(result);
                    return;
                  }
                  setNotes("");
                  void load({ quiet: true });
                }}
              >
                <div className="space-y-2">
                  <Label htmlFor="minutes">Minutes</Label>
                  <Input
                    id="minutes"
                    type="number"
                    min={1}
                    value={minutes}
                    onChange={(e) => setMinutes(e.target.value)}
                    required
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="notes">Notes</Label>
                  <Textarea
                    id="notes"
                    value={notes}
                    onChange={(e) => setNotes(e.target.value)}
                    placeholder="What did you work on?"
                  />
                </div>
                <Button type="submit" disabled={busy || !skillSlug}>
                  {busy ? "Logging…" : "Log session"}
                </Button>
              </form>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="font-serif text-lg">
                Recent sessions
              </CardTitle>
            </CardHeader>
            <CardContent className="max-h-56 overflow-y-auto">
              {sessions.length === 0 ? (
                <Empty>No sessions logged.</Empty>
              ) : (
                <ul className="divide-y divide-border">
                  {sessions.map((row, i) => (
                    <li key={i} className="py-3 first:pt-0 last:pb-0">
                      <div className="font-medium">
                        {row.skill_slug} · {row.duration_minutes}m
                      </div>
                      <div className="text-sm text-foreground/75">
                        {row.studied_at}
                        {row.notes ? ` · ${row.notes}` : ""}
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

export default function ProgressPage() {
  return (
    <Suspense fallback={<PageSkeleton variant="dashboard" />}>
      <ProgressInner />
    </Suspense>
  );
}
