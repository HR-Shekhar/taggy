"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";
import { ArrowRight, Users } from "lucide-react";
import { useAuth } from "@/lib/auth";
import {
  createPod,
  listMyPods,
  listMySkills,
  listPodsBySkill,
  type MySkill,
} from "@/lib/api";
import { Empty, PageHeader, PageSkeleton } from "@/components/app-ui";
import { toastApiError } from "@/lib/toast";
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
import { cn } from "@/lib/utils";

type MyPod = {
  pod_slug?: string;
  slug?: string;
  pod_name?: string;
  name?: string;
  skill_slug?: string;
  status?: string;
  role?: string;
};

type SkillPod = {
  slug: string;
  name: string;
  accepted_count?: number;
  max_members?: number;
  owner_username?: string;
};

function PodsInner() {
  const { username } = useAuth();
  const params = useSearchParams();
  const [skills, setSkills] = useState<MySkill[]>([]);
  const [skillSlug, setSkillSlug] = useState(params.get("skill") ?? "");
  const [mine, setMine] = useState<MyPod[]>([]);
  const [bySkill, setBySkill] = useState<SkillPod[]>([]);
  const [name, setName] = useState("");
  const [podLink, setPodLink] = useState("");
  const [busy, setBusy] = useState(false);
  const [loading, setLoading] = useState(true);

  async function load() {
    if (!username) return;
    setLoading(true);
    const [sk, mp] = await Promise.all([
      listMySkills(username),
      listMyPods(username),
    ]);
    if (sk.ok) {
      const list = sk.data ?? [];
      setSkills(list);
      const next = skillSlug || list[0]?.skill_slug || "";
      setSkillSlug(next);
      if (next) {
        const pods = await listPodsBySkill(next);
        if (pods.ok) {
          setBySkill(
            Array.isArray(pods.data)
              ? (pods.data as SkillPod[])
              : ((pods.data as { pods?: SkillPod[] })?.pods ?? [])
          );
        }
      }
    }
    if (mp.ok) {
      setMine(
        Array.isArray(mp.data)
          ? (mp.data as MyPod[])
          : ((mp.data as { pods?: MyPod[] })?.pods ?? [])
      );
    }
    setLoading(false);
  }

  useEffect(() => {
    void load();
  }, [username]);

  useEffect(() => {
    if (!skillSlug) return;
    (async () => {
      const pods = await listPodsBySkill(skillSlug);
      if (pods.ok) {
        setBySkill(
          Array.isArray(pods.data)
            ? (pods.data as SkillPod[])
            : ((pods.data as { pods?: SkillPod[] })?.pods ?? [])
        );
      }
    })();
  }, [skillSlug]);

  if (loading) return <PageSkeleton variant="list" />;

  const membershipForSkill = mine.find((p) => p.skill_slug === skillSlug);
  const alreadyInPodForSkill = Boolean(membershipForSkill);

  return (
    <div className="space-y-6" data-tour="pods-page">
      <PageHeader
        title="Pods"
        description="Small accountability groups for the skill you're learning."
      />
      <Card>
        <CardHeader>
          <CardTitle className="font-serif text-lg">My memberships</CardTitle>
        </CardHeader>
        <CardContent>
          {mine.length === 0 ? (
            <Empty>No pods yet.</Empty>
          ) : (
            <ul className="divide-y divide-border">
              {mine.map((p) => {
                const podSlug = p.pod_slug ?? p.slug ?? "";
                const podName = p.pod_name ?? p.name ?? podSlug;
                return (
                  <li
                    key={podSlug}
                    className="flex items-center justify-between gap-3 py-3 first:pt-0 last:pb-0"
                  >
                    <div className="min-w-0 space-y-1">
                      <Link
                        href={`/pods/${podSlug}`}
                        className="flex items-center gap-2 font-medium hover:text-primary"
                      >
                        <Users className="size-3.5 text-primary" />
                        {podName}
                      </Link>
                      <div className="flex flex-wrap gap-1.5">
                        {p.status ? <Badge variant="secondary">{p.status}</Badge> : null}
                        {p.role ? <Badge variant="outline">{p.role}</Badge> : null}
                        {p.skill_slug ? (
                          <span className="text-xs text-muted-foreground">
                            {p.skill_slug}
                          </span>
                        ) : null}
                      </div>
                    </div>
                    <Link
                      href={`/pods/${podSlug}`}
                      className={cn(buttonVariants({ size: "sm" }), "gap-1")}
                    >
                      Open
                      <ArrowRight className="size-3.5" />
                    </Link>
                  </li>
                );
              })}
            </ul>
          )}
        </CardContent>
      </Card>

      {alreadyInPodForSkill ? (
        <Card>
          <CardHeader>
            <CardTitle className="font-serif text-lg">Your pod for this skill</CardTitle>
            <CardDescription>
              You&apos;re already in a pod for{" "}
              {skills.find((s) => s.skill_slug === skillSlug)?.skill_name ?? skillSlug}.
              Leave it first if you want to join or create another.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap items-center gap-2">
              <Label htmlFor="browse-skill">Skill</Label>
              <select
                id="browse-skill"
                className="h-8 min-w-48 rounded-lg border border-foreground/25 bg-background px-2.5 text-sm"
                value={skillSlug}
                onChange={(e) => setSkillSlug(e.target.value)}
              >
                {skills.map((s) => (
                  <option key={s.skill_slug} value={s.skill_slug}>
                    {s.skill_name}
                  </option>
                ))}
              </select>
              <Link
                href={`/pods/${membershipForSkill?.pod_slug ?? membershipForSkill?.slug}`}
                className={cn(buttonVariants(), "gap-1")}
              >
                Open your pod
                <ArrowRight className="size-3.5" />
              </Link>
            </div>
          </CardContent>
        </Card>
      ) : (
      <div className="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="font-serif text-lg">Browse by skill</CardTitle>
            <CardDescription>Find open pods for an enrolled skill.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="browse-skill">Skill</Label>
              <select
                id="browse-skill"
                className="h-8 w-full rounded-lg border border-foreground/25 bg-background px-2.5 text-sm"
                value={skillSlug}
                onChange={(e) => setSkillSlug(e.target.value)}
              >
                {skills.length === 0 && <option value="">Join a skill first</option>}
                {skills.map((s) => (
                  <option key={s.skill_slug} value={s.skill_slug}>
                    {s.skill_name}
                  </option>
                ))}
              </select>
            </div>
            {bySkill.length === 0 ? (
              <Empty>No pods for this skill.</Empty>
            ) : (
              <ul className="divide-y divide-border">
                {bySkill.map((p) => (
                  <li
                    key={p.slug}
                    className="flex items-center justify-between gap-3 py-3 first:pt-0 last:pb-0"
                  >
                    <div>
                      <Link
                        href={`/pods/${p.slug}`}
                        className="font-medium hover:text-primary"
                      >
                        {p.name}
                      </Link>
                      <p className="text-xs text-muted-foreground">
                        {p.accepted_count}/{p.max_members} · @{p.owner_username}
                      </p>
                    </div>
                    <Link
                      href={`/pods/${p.slug}`}
                      className={cn(buttonVariants({ variant: "outline", size: "sm" }))}
                    >
                      View
                    </Link>
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="font-serif text-lg">Create pod</CardTitle>
            <CardDescription>Start a small group (max 7).</CardDescription>
          </CardHeader>
          <CardContent>
            <form
              className="space-y-4"
              onSubmit={async (e) => {
                e.preventDefault();
                setBusy(true);
                const result = await createPod(skillSlug, {
                  name,
                  slug: podLink,
                });
                setBusy(false);
                if (!result.ok) {
                  toastApiError(result);
                  return;
                }
                setName("");
                setPodLink("");
                void load();
              }}
            >
              <div className="space-y-2">
                <Label htmlFor="create-skill">Skill</Label>
                <select
                  id="create-skill"
                  className="h-8 w-full rounded-lg border border-foreground/25 bg-background px-2.5 text-sm"
                  value={skillSlug}
                  onChange={(e) => setSkillSlug(e.target.value)}
                  required
                >
                  {skills.map((s) => (
                    <option key={s.skill_slug} value={s.skill_slug}>
                      {s.skill_name}
                    </option>
                  ))}
                </select>
              </div>
              <div className="space-y-2">
                <Label htmlFor="pod-name">Name</Label>
                <Input
                  id="pod-name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="Night Owls Study Crew"
                  minLength={3}
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="pod-link">Pod link</Label>
                <Input
                  id="pod-link"
                  value={podLink}
                  onChange={(e) => setPodLink(e.target.value)}
                  placeholder="night-owls-study-crew"
                  minLength={3}
                  maxLength={60}
                  pattern="[a-z0-9]+(?:-[a-z0-9]+)*"
                  title="Lowercase letters, numbers, and hyphens only"
                  required
                />
                <p className="text-xs text-muted-foreground">
                  Used in the URL: /pods/{podLink.trim() || "your-pod-link"}
                </p>
              </div>
              <Button
                type="submit"
                disabled={
                  busy ||
                  !skillSlug ||
                  name.trim().length < 3 ||
                  podLink.trim().length < 3
                }
              >
                {busy ? "Creating…" : "Create pod"}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
      )}
    </div>
  );
}

export default function PodsPage() {
  return (
    <Suspense fallback={<PageSkeleton variant="list" />}>
      <PodsInner />
    </Suspense>
  );
}
