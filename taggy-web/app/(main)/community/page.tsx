"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { Headphones, MessageCircle, ArrowRight } from "lucide-react";
import { useAuth } from "@/lib/auth";
import { listMyPods, listMySkills, type MySkill } from "@/lib/api";
import { Empty, PageHeader, PageSkeleton, Section } from "@/components/app-ui";
import { EmptyArtChat, EmptyArtPods } from "@/components/empty-art";
import { toastApiError } from "@/lib/toast";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type MyPod = {
  pod_slug?: string;
  slug?: string;
  pod_name?: string;
  name?: string;
  skill_slug?: string;
  status?: string;
};

export default function CommunityHubPage() {
  const { username } = useAuth();
  const [skills, setSkills] = useState<MySkill[]>([]);
  const [pods, setPods] = useState<MyPod[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!username) return;
    (async () => {
      setLoading(true);
      const [sk, pd] = await Promise.all([
        listMySkills(username),
        listMyPods(username),
      ]);
      if (!sk.ok) toastApiError(sk);
      else setSkills(sk.data ?? []);
      if (pd.ok) {
        setPods(
          Array.isArray(pd.data)
            ? (pd.data as MyPod[])
            : ((pd.data as { pods?: MyPod[] })?.pods ?? [])
        );
      }
      setLoading(false);
    })();
  }, [username]);

  if (loading) return <PageSkeleton variant="list" />;

  return (
    <div className="space-y-8" data-tour="community-page">
      <PageHeader
        title="Community & audio"
        description="Jump into skill chat channels or open your pod for live rooms."
      />
      <div className="grid gap-8 lg:grid-cols-2">
        <Section title="Skill communities">
          {skills.length === 0 ? (
            <Empty
              art={<EmptyArtChat />}
              title="Join a skill first"
              description="Community chat opens for skills you're enrolled in."
              action={
                <Link href="/skills" className={cn(buttonVariants())}>
                  Browse skills
                </Link>
              }
            />
          ) : (
            <ul className="divide-y divide-border overflow-hidden rounded-xl border border-border bg-card">
              {skills.map((s) => (
                <li key={s.skill_slug}>
                  <Link
                    href={`/community/${s.skill_slug}`}
                    className="flex items-center justify-between gap-3 px-4 py-3.5 transition-colors hover:bg-muted/40"
                  >
                    <div className="min-w-0">
                      <div className="font-medium">{s.skill_name}</div>
                      <p className="text-sm text-foreground/75">
                        Chat · channels · audio
                      </p>
                    </div>
                    <span className="inline-flex items-center gap-1 text-sm text-primary">
                      <MessageCircle className="size-3.5" />
                      Open
                      <ArrowRight className="size-3.5" />
                    </span>
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </Section>

        <Section title="Pod rooms">
          {pods.length === 0 ? (
            <Empty
              art={<EmptyArtPods />}
              title="No pods yet"
              description="Find an accountability group to unlock pod chat and audio."
              action={
                <Link href="/pods" className={cn(buttonVariants())}>
                  Find a pod
                </Link>
              }
            />
          ) : (
            <ul className="divide-y divide-border overflow-hidden rounded-xl border border-border bg-card">
              {pods.map((p) => {
                const slug = p.pod_slug ?? p.slug ?? "";
                const name = p.pod_name ?? p.name ?? slug;
                return (
                  <li key={slug}>
                    <Link
                      href={`/pods/${slug}`}
                      className="flex items-center justify-between gap-3 px-4 py-3.5 transition-colors hover:bg-muted/40"
                    >
                      <div className="min-w-0">
                        <div className="font-medium">{name}</div>
                        <p className="text-sm text-foreground/75">
                          {p.skill_slug} · {p.status ?? "ACTIVE"}
                        </p>
                      </div>
                      <span className="inline-flex items-center gap-1 text-sm text-primary">
                        <Headphones className="size-3.5" />
                        Open
                      </span>
                    </Link>
                  </li>
                );
              })}
            </ul>
          )}
        </Section>
      </div>
    </div>
  );
}
