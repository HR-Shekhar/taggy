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
import {
  Empty,
  PageHeader,
  PageSkeleton,
  Section,
} from "@/components/app-ui";
import { EmptyArtPods, EmptyArtSkills } from "@/components/empty-art";
import { toastApiError } from "@/lib/toast";
import { CommunityLeaderboardPanel } from "@/components/community-leaderboard-panel";
import { PodQuizPanel } from "@/components/pod-quiz-panel";
import { ActivityDots, MiniBars } from "@/components/charts";
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
  const activeDays = weekMinutes.map((m) => m > 0);

  if (loading) return <PageSkeleton variant="dashboard" />;

  const currentStreak = summary?.current_streak ?? streak?.current_streak ?? 0;
  const longestStreak = summary?.longest_streak ?? streak?.longest_streak ?? 0;
  const weekly =
    summary?.weekly_minutes ?? weekMinutes.reduce((a, b) => a + b, 0);
  const fortnightTotal = fortnightMinutes.reduce((a, b) => a + b, 0);

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

  const primaryCta =
    skills.length > 0
      ? {
          href: `/skills/${skills[0].skill_slug}`,
          label: `Continue ${skills[0].skill_name}`,
        }
      : { href: "/skills", label: "Browse skills" };

  return (
    <div className="space-y-8" data-tour="home-main">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <PageHeader
          title={displayName ? `Welcome back, ${displayName}` : "Home"}
          description="Pick up where you left off."
        />
        <form
          onSubmit={submitGlobalSearch}
          className="relative w-full shrink-0 sm:mt-1 sm:max-w-xs"
        >
          <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={searchQ}
            onChange={(e) => setSearchQ(e.target.value)}
            placeholder="Search Taggy"
            aria-label="Search Taggy"
            className="h-10 rounded-xl border-border bg-card pl-10"
          />
        </form>
      </div>

      <div className="flex flex-wrap items-center gap-3">
        <Link href={primaryCta.href} className={cn(buttonVariants(), "gap-1.5")}>
          {primaryCta.label}
          <ArrowRight className="size-4" />
        </Link>
        <Link
          href="/progress"
          className={cn(buttonVariants({ variant: "outline" }), "gap-1.5")}
        >
          Log study
        </Link>
      </div>

      <div className="grid grid-cols-2 gap-px overflow-hidden rounded-xl border border-border bg-border sm:grid-cols-4">
        <div className="bg-card px-4 py-3">
          <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
            <Flame className="size-3.5 text-primary" />
            Streak
          </div>
          <p className="mt-1 font-serif text-2xl tabular-nums">
            {currentStreak}
            <span className="ml-1 text-sm text-muted-foreground">days</span>
          </p>
          <div className="mt-2">
            <ActivityDots active={activeDays} />
          </div>
        </div>
        <div className="bg-card px-4 py-3">
          <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
            <Timer className="size-3.5 text-primary" />
            This week
          </div>
          <p className="mt-1 font-serif text-2xl tabular-nums">
            {formatMinutes(weekly)}
          </p>
          <p className="mt-1 text-sm text-muted-foreground">
            Best streak {longestStreak}d
          </p>
        </div>
        <div className="bg-card px-4 py-3">
          <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
            <BookOpen className="size-3.5 text-primary" />
            Skills
          </div>
          <p className="mt-1 font-serif text-2xl tabular-nums">{skills.length}</p>
          <p className="mt-1 text-sm text-muted-foreground">
            {pods.length} pod{pods.length === 1 ? "" : "s"}
          </p>
        </div>
        <Link
          href="/notifications"
          className="bg-card px-4 py-3 transition-colors hover:bg-muted/40"
        >
          <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
            <Bell className="size-3.5 text-primary" />
            Unread
          </div>
          <p className="mt-1 font-serif text-2xl tabular-nums">{unread}</p>
          <p className="mt-1 text-sm text-muted-foreground">
            {unread > 0 ? "Open inbox" : "All caught up"}
          </p>
        </Link>
      </div>

      <div className="grid gap-8 lg:grid-cols-2">
        <Section
          title="Continue learning"
          action={
            <Link
              href="/skills"
              className={cn(buttonVariants({ variant: "ghost", size: "sm" }))}
            >
              All skills
            </Link>
          }
        >
          {skills.length === 0 ? (
            <Empty
              art={<EmptyArtSkills />}
              title="No skills yet"
              description="Join a skill roadmap to start tracking milestones."
              action={
                <Link href="/skills" className={cn(buttonVariants())}>
                  Browse skills
                </Link>
              }
            />
          ) : (
            <ul className="divide-y divide-border overflow-hidden rounded-xl border border-border bg-card">
              {skills.slice(0, 5).map((s) => (
                <li key={s.skill_slug}>
                  <Link
                    href={`/skills/${s.skill_slug}`}
                    className="flex items-center gap-3 px-4 py-3 transition-colors hover:bg-muted/40"
                  >
                    <div className="min-w-0 flex-1">
                      <p className="truncate font-medium">{s.skill_name}</p>
                      <p className="text-sm text-muted-foreground">
                        {s.completed_count}/{s.milestone_count} milestones
                      </p>
                      <div className="mt-2 h-1.5 overflow-hidden rounded-full bg-muted">
                        <div
                          className="h-full rounded-full bg-primary"
                          style={{
                            width: `${Math.min(100, s.completion_percent || 0)}%`,
                          }}
                        />
                      </div>
                    </div>
                    <ArrowRight className="size-4 shrink-0 text-muted-foreground" />
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </Section>

        <Section
          title="Your pods"
          action={
            <Link
              href="/pods"
              className={cn(buttonVariants({ variant: "ghost", size: "sm" }))}
            >
              All pods
            </Link>
          }
        >
          {pods.length === 0 ? (
            <Empty
              art={<EmptyArtPods />}
              title="No pods yet"
              description="Join a small group to stay accountable."
              action={
                <Link href="/pods" className={cn(buttonVariants())}>
                  Find a pod
                </Link>
              }
            />
          ) : (
            <ul className="divide-y divide-border overflow-hidden rounded-xl border border-border bg-card">
              {pods.slice(0, 5).map((p, i) => {
                const slug = p.slug ?? p.pod_slug ?? "";
                const name = p.name ?? p.pod_name ?? slug;
                return (
                  <li key={slug || i}>
                    <Link
                      href={`/pods/${slug}`}
                      className="flex items-center gap-3 px-4 py-3 transition-colors hover:bg-muted/40"
                    >
                      <Users className="size-4 shrink-0 text-primary" />
                      <div className="min-w-0 flex-1">
                        <p className="truncate font-medium">{name}</p>
                        <div className="mt-1 flex flex-wrap gap-1.5">
                          {p.status ? (
                            <Badge variant="secondary">{p.status}</Badge>
                          ) : null}
                          {p.role ? (
                            <Badge variant="outline">{p.role}</Badge>
                          ) : null}
                        </div>
                      </div>
                      <span className="flex items-center gap-1 text-sm text-muted-foreground">
                        <MessageCircle className="size-3.5" />
                        Chat
                      </span>
                    </Link>
                  </li>
                );
              })}
            </ul>
          )}
        </Section>
      </div>

      <Section
        title="Study activity"
        description={
          fortnightTotal > 0
            ? `${formatMinutes(fortnightTotal)} over the last 14 days`
            : "Log a study block to light up this chart"
        }
        action={
          <Link
            href="/progress"
            className={cn(buttonVariants({ variant: "ghost", size: "sm" }))}
          >
            Details
          </Link>
        }
      >
        <div className="rounded-xl border border-border bg-card p-4">
          <MiniBars values={fortnightMinutes} height={72} className="h-20" />
        </div>
      </Section>

      <div className="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader className="flex-row items-start justify-between gap-3">
            <div>
              <CardTitle className="font-serif text-lg">Pod leaderboard</CardTitle>
              <CardDescription>
                {acceptedPodSlug
                  ? `Members in ${acceptedPodName}`
                  : "Join a pod to see quiz scores"}
              </CardDescription>
            </div>
            <Link
              href={acceptedPodSlug ? `/pods/${acceptedPodSlug}` : "/pods"}
              className={cn(buttonVariants({ variant: "ghost", size: "sm" }))}
            >
              {acceptedPodSlug ? "Open" : "Find pod"}
            </Link>
          </CardHeader>
          <CardContent>
            {acceptedPodSlug ? (
              <PodQuizPanel
                podSlug={acceptedPodSlug}
                enabled
                mode="leaderboard"
              />
            ) : (
              <Empty
                art={<EmptyArtPods />}
                title="No pod yet"
                action={
                  <Link href="/pods" className="text-primary hover:underline">
                    Join or create one
                  </Link>
                }
              />
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex-row items-start justify-between gap-3">
            <div>
              <CardTitle className="font-serif text-lg">
                Community leaderboard
              </CardTitle>
              <CardDescription>
                {primarySkillSlug
                  ? `Standings for ${primarySkill?.skill_name ?? primarySkillSlug}`
                  : "Join a skill to see rankings"}
              </CardDescription>
            </div>
            <Link
              href={
                primarySkillSlug
                  ? `/community/${primarySkillSlug}`
                  : "/skills"
              }
              className={cn(buttonVariants({ variant: "ghost", size: "sm" }))}
            >
              {primarySkillSlug ? "Community" : "Skills"}
            </Link>
          </CardHeader>
          <CardContent>
            {primarySkillSlug ? (
              <CommunityLeaderboardPanel skillSlug={primarySkillSlug} compact />
            ) : (
              <Empty
                art={<EmptyArtSkills />}
                title="No skills yet"
                action={
                  <Link href="/skills" className="text-primary hover:underline">
                    Browse skills
                  </Link>
                }
              />
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
