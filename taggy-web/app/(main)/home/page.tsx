"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useEffect, useMemo, useState } from "react";
import {
  ArrowRight,
  Bell,
  BookOpen,
  Flame,
  MessageCircle,
  Search,
  Timer,
  Users,
} from "lucide-react";
import { useAuth } from "@/lib/auth";
import {
  getProgressSummary,
  getStreak,
  listMyPods,
  listMySkills,
  listNotifications,
  listStudySessions,
  type MyPod,
  type MySkill,
  type ProgressSummary,
} from "@/lib/api";
import { Empty, Loading, PageHeader } from "@/components/app-ui";
import { toastApiError } from "@/lib/toast";
import { CommunityLeaderboardPanel } from "@/components/community-leaderboard-panel";
import { PodQuizPanel } from "@/components/pod-quiz-panel";
import {
  ActivityDots,
  MiniBars,
  ProgressRing,
  Sparkline,
  StackedBar,
} from "@/components/charts";
import { Badge } from "@/components/ui/badge";
import { buttonVariants } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

type StudySession = {
  skill_slug: string;
  duration_minutes: number;
  studied_at: string;
};

function formatMinutes(mins: number) {
  if (mins < 60) return `${mins}m`;
  const h = Math.floor(mins / 60);
  const m = mins % 60;
  return m === 0 ? `${h}h` : `${h}h ${m}m`;
}

function dayKey(d: Date) {
  return d.toISOString().slice(0, 10);
}

function lastNDays(n: number) {
  const days: Date[] = [];
  const now = new Date();
  for (let i = n - 1; i >= 0; i--) {
    const d = new Date(now);
    d.setHours(12, 0, 0, 0);
    d.setDate(d.getDate() - i);
    days.push(d);
  }
  return days;
}

function bucketSessions(sessions: StudySession[], days: Date[]) {
  const map = new Map(days.map((d) => [dayKey(d), 0]));
  for (const s of sessions) {
    const key = dayKey(new Date(s.studied_at));
    if (map.has(key)) {
      map.set(key, (map.get(key) ?? 0) + (s.duration_minutes || 0));
    }
  }
  return days.map((d) => map.get(dayKey(d)) ?? 0);
}

export default function HomePage() {
  const { username, displayName } = useAuth();
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [skills, setSkills] = useState<MySkill[]>([]);
  const [pods, setPods] = useState<MyPod[]>([]);
  const [streak, setStreak] = useState<{
    current_streak: number;
    longest_streak: number;
  } | null>(null);
  const [summary, setSummary] = useState<ProgressSummary | null>(null);
  const [sessions, setSessions] = useState<StudySession[]>([]);
  const [unread, setUnread] = useState(0);
  const [searchQ, setSearchQ] = useState("");

  function submitGlobalSearch(e: FormEvent) {
    e.preventDefault();
    const q = searchQ.trim();
    if (!q) return;
    router.push(`/search?q=${encodeURIComponent(q)}`);
  }

  useEffect(() => {
    if (!username) {
      setLoading(false);
      return;
    }
    (async () => {
      setLoading(true);
      try {
        const [sk, pd, st, sm, nt, ss] = await Promise.all([
          listMySkills(username),
          listMyPods(username),
          getStreak(username),
          getProgressSummary(username),
          listNotifications(username, true),
          listStudySessions(username),
        ]);
        if (!sk.ok) toastApiError(sk);
        else setSkills(sk.data ?? []);
        if (pd.ok) {
          setPods(
            Array.isArray(pd.data)
              ? pd.data
              : (pd.data as { pods?: MyPod[] })?.pods ?? []
          );
        }
        if (st.ok) setStreak(st.data);
        if (sm.ok) setSummary(sm.data);
        if (ss.ok) {
          const raw = Array.isArray(ss.data)
            ? ss.data
            : (ss.data as { sessions?: StudySession[] })?.sessions ?? [];
          setSessions(raw as StudySession[]);
        }
        if (nt.ok) {
          const list = Array.isArray(nt.data)
            ? nt.data
            : (nt.data as { notifications?: unknown[] })?.notifications ?? [];
          setUnread(list.length);
        }
      } finally {
        setLoading(false);
      }
    })();
  }, [username]);

  const days7 = useMemo(() => lastNDays(7), []);
  const days14 = useMemo(() => lastNDays(14), []);
  const weekMinutes = useMemo(
    () => bucketSessions(sessions, days7),
    [sessions, days7]
  );
  const fortnightMinutes = useMemo(
    () => bucketSessions(sessions, days14),
    [sessions, days14]
  );
  const cumulative = useMemo(() => {
    let sum = 0;
    return fortnightMinutes.map((m) => {
      sum += m;
      return sum;
    });
  }, [fortnightMinutes]);
  const activeDays = weekMinutes.map((m) => m > 0);

  if (loading) return <Loading />;

  const currentStreak = summary?.current_streak ?? streak?.current_streak ?? 0;
  const longestStreak = summary?.longest_streak ?? streak?.longest_streak ?? 0;
  const weekly = summary?.weekly_minutes ?? weekMinutes.reduce((a, b) => a + b, 0);
  const monthly = summary?.monthly_minutes ?? 0;
  const total = summary?.total_minutes ?? 0;
  const streakRatio =
    longestStreak > 0 ? Math.min(1, currentStreak / longestStreak) : currentStreak > 0 ? 1 : 0;
  const weekGoal = Math.max(weekly, 150);
  const weekPct = Math.min(1, weekly / weekGoal);
  const restTotal = Math.max(total - monthly, 0);
  const restMonth = Math.max(monthly - weekly, 0);

  const dayLabels = days14.map((d) =>
    d.toLocaleDateString(undefined, { weekday: "narrow" })
  );

  const primarySkill = skills[0];
  const primarySkillSlug = primarySkill?.skill_slug ?? "";
  const acceptedPod =
    pods.find(
      (p) =>
        p.skill_slug === primarySkillSlug &&
        (p.status ?? "ACCEPTED") === "ACCEPTED"
    ) ??
    pods.find((p) => (p.status ?? "ACCEPTED") === "ACCEPTED") ??
    pods[0];
  const acceptedPodSlug = acceptedPod?.pod_slug ?? acceptedPod?.slug ?? "";
  const acceptedPodName =
    acceptedPod?.pod_name ?? acceptedPod?.name ?? acceptedPodSlug;

  return (
    <div className="space-y-6">
      <form onSubmit={submitGlobalSearch} className="relative">
        <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          value={searchQ}
          onChange={(e) => setSearchQ(e.target.value)}
          placeholder="Global search in Taggy"
          aria-label="Global search in Taggy"
          className="h-11 rounded-xl border-foreground/15 bg-background/80 pl-10 shadow-sm"
        />
      </form>

      <PageHeader
        title={displayName ? `Welcome back, ${displayName}` : "Home"}
        description="Your skills, pods, and study rhythm in one place."
      >
        <div className="flex flex-wrap gap-2">
          <Link
            href="/community"
            className={cn(buttonVariants({ variant: "outline" }), "gap-1.5")}
          >
            <MessageCircle className="size-4" />
            Community & audio
          </Link>
          <Link
            href="/skills"
            className={cn(buttonVariants({ variant: "outline" }), "gap-1.5")}
          >
            Browse skills
          </Link>
          <Link href="/progress" className={cn(buttonVariants(), "gap-1.5")}>
            Log study
            <ArrowRight className="size-4" />
          </Link>
        </div>
      </PageHeader>

      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <Card className="rounded-xl ring-1 ring-foreground/10">
          <CardHeader className="pb-0">
            <div className="flex items-center justify-between gap-2">
              <CardDescription>Current streak</CardDescription>
              <span className="inline-flex size-8 items-center justify-center rounded-lg bg-secondary text-secondary-foreground">
                <Flame className="size-4 text-primary" />
              </span>
            </div>
          </CardHeader>
          <CardContent className="flex items-center gap-4 pt-2">
            <ProgressRing value={streakRatio} size={76}>
              <div>
                <div className="font-serif text-xl leading-none">{currentStreak}</div>
                <div className="text-[10px] text-muted-foreground">days</div>
              </div>
            </ProgressRing>
            <div className="min-w-0 flex-1 space-y-2">
              <p className="text-sm text-muted-foreground">
                {longestStreak > 0 ? `Best ${longestStreak} days` : "Start today"}
              </p>
              <ActivityDots active={activeDays} />
              <p className="text-[11px] text-muted-foreground">Last 7 days</p>
            </div>
          </CardContent>
        </Card>

        <Card className="rounded-xl ring-1 ring-foreground/10">
          <CardHeader className="pb-0">
            <div className="flex items-center justify-between gap-2">
              <CardDescription>This week</CardDescription>
              <span className="inline-flex size-8 items-center justify-center rounded-lg bg-secondary text-secondary-foreground">
                <Timer className="size-4 text-primary" />
              </span>
            </div>
            <CardTitle className="font-serif text-3xl tracking-tight">
              {formatMinutes(weekly)}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <MiniBars values={weekMinutes} />
            <div className="flex items-center justify-between text-[11px] text-muted-foreground">
              <span>{Math.round(weekPct * 100)}% of weekly pace</span>
              <span>{formatMinutes(monthly)} / mo</span>
            </div>
          </CardContent>
        </Card>

        <Card className="rounded-xl ring-1 ring-foreground/10">
          <CardHeader className="pb-0">
            <div className="flex items-center justify-between gap-2">
              <CardDescription>Total studied</CardDescription>
              <span className="inline-flex size-8 items-center justify-center rounded-lg bg-secondary text-secondary-foreground">
                <BookOpen className="size-4 text-primary" />
              </span>
            </div>
            <CardTitle className="font-serif text-3xl tracking-tight">
              {formatMinutes(total)}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <Sparkline values={cumulative.length > 1 ? cumulative : [0, total]} />
            <StackedBar
              segments={[
                { value: weekly, label: "Week", className: "bg-primary" },
                { value: restMonth, label: "Month", className: "bg-primary/55" },
                { value: restTotal, label: "Earlier", className: "bg-primary/25" },
              ]}
            />
          </CardContent>
        </Card>

        <Card className="rounded-xl ring-1 ring-foreground/10">
          <CardHeader className="pb-0">
            <div className="flex items-center justify-between gap-2">
              <CardDescription>Notifications</CardDescription>
              <span className="inline-flex size-8 items-center justify-center rounded-lg bg-secondary text-secondary-foreground">
                <Bell className="size-4 text-primary" />
              </span>
            </div>
          </CardHeader>
          <CardContent className="flex items-center gap-4 pt-2">
            <ProgressRing
              value={unread > 0 ? Math.min(1, unread / 10) : 0.08}
              size={76}
            >
              <div className="font-serif text-xl leading-none">{unread}</div>
            </ProgressRing>
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground">
                {unread > 0 ? "Unread alerts" : "You're caught up"}
              </p>
              {unread > 0 ? (
                <Link
                  href="/notifications"
                  className="text-sm text-primary hover:underline"
                >
                  View inbox
                </Link>
              ) : (
                <p className="text-[11px] text-muted-foreground">
                  {skills.length} skill{skills.length === 1 ? "" : "s"} active
                </p>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <Card className="rounded-xl ring-1 ring-foreground/10">
          <CardHeader className="flex-row items-start justify-between gap-3">
            <div>
              <CardTitle className="font-serif text-lg">Pod leaderboard</CardTitle>
              <CardDescription>
                {acceptedPodSlug
                  ? `Members in ${acceptedPodName}`
                  : "Join a pod to see member quiz scores"}
              </CardDescription>
            </div>
            {acceptedPodSlug ? (
              <Link
                href={`/pods/${acceptedPodSlug}`}
                className={cn(buttonVariants({ variant: "ghost", size: "sm" }))}
              >
                Open pod
              </Link>
            ) : (
              <Link
                href="/pods"
                className={cn(buttonVariants({ variant: "ghost", size: "sm" }))}
              >
                Find pod
              </Link>
            )}
          </CardHeader>
          <CardContent>
            {acceptedPodSlug ? (
              <PodQuizPanel
                podSlug={acceptedPodSlug}
                enabled
                mode="leaderboard"
              />
            ) : (
              <Empty>
                No accepted pod yet.{" "}
                <Link href="/pods" className="text-primary hover:underline">
                  Join or create one
                </Link>
              </Empty>
            )}
          </CardContent>
        </Card>

        <Card className="rounded-xl ring-1 ring-foreground/10">
          <CardHeader className="flex-row items-start justify-between gap-3">
            <div>
              <CardTitle className="font-serif text-lg">
                Community leaderboard
              </CardTitle>
              <CardDescription>
                {primarySkillSlug
                  ? `Pod standings for ${primarySkill?.skill_name ?? primarySkillSlug}`
                  : "Join a skill to see community pod rankings"}
              </CardDescription>
            </div>
            {primarySkillSlug ? (
              <Link
                href={`/community/${primarySkillSlug}`}
                className={cn(buttonVariants({ variant: "ghost", size: "sm" }))}
              >
                Community
              </Link>
            ) : (
              <Link
                href="/skills"
                className={cn(buttonVariants({ variant: "ghost", size: "sm" }))}
              >
                Skills
              </Link>
            )}
          </CardHeader>
          <CardContent>
            {primarySkillSlug ? (
              <CommunityLeaderboardPanel skillSlug={primarySkillSlug} compact />
            ) : (
              <Empty>
                No skills yet.{" "}
                <Link href="/skills" className="text-primary hover:underline">
                  Browse skills
                </Link>
              </Empty>
            )}
          </CardContent>
        </Card>
      </div>

      <Card className="rounded-xl ring-1 ring-foreground/10">
        <CardHeader className="flex-row items-start justify-between gap-3">
          <div>
            <CardTitle className="font-serif text-lg">Study activity</CardTitle>
            <CardDescription>Minutes logged over the last 14 days</CardDescription>
          </div>
          <Link
            href="/progress"
            className={cn(buttonVariants({ variant: "ghost", size: "sm" }))}
          >
            Details
          </Link>
        </CardHeader>
        <CardContent className="space-y-3">
          <MiniBars values={fortnightMinutes} height={72} className="h-20" />
          <div className="hidden justify-between text-[10px] text-muted-foreground sm:flex">
            {dayLabels.map((label, i) => (
              <span key={i}>{label}</span>
            ))}
          </div>
          <p className="text-sm text-muted-foreground">
            {fortnightMinutes.reduce((a, b) => a + b, 0) > 0
              ? `${formatMinutes(fortnightMinutes.reduce((a, b) => a + b, 0))} across ${fortnightMinutes.filter((m) => m > 0).length} active days`
              : "No sessions yet — log a study block to light up this chart."}
          </p>
        </CardContent>
      </Card>

      <div className="grid gap-4 lg:grid-cols-2">
        <Card className="rounded-xl ring-1 ring-foreground/10">
          <CardHeader className="flex-row items-start justify-between gap-3">
            <div>
              <CardTitle className="font-serif text-lg">My skills</CardTitle>
              <CardDescription>
                Roadmaps you&apos;re actively learning
              </CardDescription>
            </div>
            <Link
              href="/skills"
              className={cn(buttonVariants({ variant: "ghost", size: "sm" }))}
            >
              All
            </Link>
          </CardHeader>
          <CardContent>
            {skills.length === 0 ? (
              <Empty>
                No skills yet.{" "}
                <Link href="/skills" className="text-primary hover:underline">
                  Browse skills
                </Link>
              </Empty>
            ) : (
              <ul className="divide-y divide-border">
                {skills.map((s) => (
                  <li
                    key={s.skill_slug}
                    className="space-y-2 py-3 first:pt-0 last:pb-0"
                  >
                    <div className="flex items-center justify-between gap-3">
                      <div className="min-w-0">
                        <Link
                          href={`/skills/${s.skill_slug}`}
                          className="truncate font-medium hover:text-primary"
                        >
                          {s.skill_name}
                        </Link>
                        <p className="text-xs text-muted-foreground">
                          Roadmap v{s.roadmap_version_number} ·{" "}
                          {s.completed_count}/{s.milestone_count} milestones
                        </p>
                      </div>
                      <Link
                        href={`/skills/${s.skill_slug}`}
                        className={cn(buttonVariants({ size: "sm" }), "gap-1")}
                      >
                        Open
                        <ArrowRight className="size-3.5" />
                      </Link>
                    </div>
                    <div className="flex flex-wrap gap-2">
                      <Link
                        href={`/community/${s.skill_slug}`}
                        className={cn(
                          buttonVariants({ variant: "outline", size: "sm" }),
                          "gap-1"
                        )}
                      >
                        <MessageCircle className="size-3.5" />
                        Chat
                      </Link>
                    </div>
                    <div className="h-1.5 overflow-hidden rounded-full bg-muted">
                      <div
                        className="h-full rounded-full bg-primary"
                        style={{
                          width: `${Math.min(100, s.completion_percent || 0)}%`,
                        }}
                      />
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>

        <Card className="rounded-xl ring-1 ring-foreground/10">
          <CardHeader className="flex-row items-start justify-between gap-3">
            <div>
              <CardTitle className="font-serif text-lg">My pods</CardTitle>
              <CardDescription>
                Small groups keeping you accountable
              </CardDescription>
            </div>
            <Link
              href="/pods"
              className={cn(buttonVariants({ variant: "ghost", size: "sm" }))}
            >
              All
            </Link>
          </CardHeader>
          <CardContent>
            {pods.length === 0 ? (
              <Empty>
                No pods yet.{" "}
                <Link href="/pods" className="text-primary hover:underline">
                  Find or create one
                </Link>
              </Empty>
            ) : (
              <ul className="divide-y divide-border">
                {pods.map((p, i) => {
                  const slug = p.slug ?? p.pod_slug ?? "";
                  const name = p.name ?? p.pod_name ?? slug;
                  return (
                    <li
                      key={slug || i}
                      className="flex items-center justify-between gap-3 py-3 first:pt-0 last:pb-0"
                    >
                      <div className="min-w-0 space-y-1">
                        <Link
                          href={`/pods/${slug}`}
                          className="flex items-center gap-2 truncate font-medium hover:text-primary"
                        >
                          <Users className="size-3.5 shrink-0 text-primary" />
                          {name}
                        </Link>
                        <div className="flex flex-wrap gap-1.5">
                          {p.status ? (
                            <Badge variant="secondary">{p.status}</Badge>
                          ) : null}
                          {p.role ? (
                            <Badge variant="outline">{p.role}</Badge>
                          ) : null}
                        </div>
                      </div>
                      <Link
                        href={`/pods/${slug}`}
                        className={cn(buttonVariants({ size: "sm" }), "gap-1")}
                      >
                        Chat & audio
                        <ArrowRight className="size-3.5" />
                      </Link>
                    </li>
                  );
                })}
              </ul>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
