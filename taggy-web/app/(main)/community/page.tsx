"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { Headphones, MessageCircle, ArrowRight } from "lucide-react";
import { useAuth } from "@/lib/auth";
import { listMyPods, listMySkills, type MySkill } from "@/lib/api";
import { Empty, Loading, PageHeader } from "@/components/app-ui";
import { toastApiError } from "@/lib/toast";
import { buttonVariants } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
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

  if (loading) return <Loading />;

  return (
    <div className="space-y-6">
      <PageHeader
        title="Community & audio"
        description="Jump into skill chat channels or open your pod for live rooms."
      />
      <div className="grid gap-4 lg:grid-cols-2">
        <Card className="rounded-xl ring-1 ring-foreground/10">
          <CardHeader>
            <CardTitle className="font-serif text-lg">Skill communities</CardTitle>
            <CardDescription>
              Channel chat and community audio for skills you joined.
            </CardDescription>
          </CardHeader>
          <CardContent>
            {skills.length === 0 ? (
              <Empty>
                Join a skill first.{" "}
                <Link href="/skills" className="text-primary hover:underline">
                  Browse skills
                </Link>
              </Empty>
            ) : (
              <ul className="divide-y divide-border">
                {skills.map((s) => (
                  <li
                    key={s.skill_slug}
                    className="flex items-center justify-between gap-3 py-3 first:pt-0 last:pb-0"
                  >
                    <div className="min-w-0">
                      <div className="font-medium">{s.skill_name}</div>
                      <p className="text-xs text-muted-foreground">
                        Chat · channels · audio rooms
                      </p>
                    </div>
                    <Link
                      href={`/community/${s.skill_slug}`}
                      className={cn(buttonVariants({ size: "sm" }), "gap-1")}
                    >
                      <MessageCircle className="size-3.5" />
                      Open chat
                      <ArrowRight className="size-3.5" />
                    </Link>
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>

        <Card className="rounded-xl ring-1 ring-foreground/10">
          <CardHeader>
            <CardTitle className="font-serif text-lg">Pod rooms</CardTitle>
            <CardDescription>
              Pod chat and LiveKit audio for your accountability groups.
            </CardDescription>
          </CardHeader>
          <CardContent>
            {pods.length === 0 ? (
              <Empty>
                You&apos;re not in a pod yet.{" "}
                <Link href="/pods" className="text-primary hover:underline">
                  Find a pod
                </Link>
              </Empty>
            ) : (
              <ul className="divide-y divide-border">
                {pods.map((p) => {
                  const slug = p.pod_slug ?? p.slug ?? "";
                  const name = p.pod_name ?? p.name ?? slug;
                  return (
                    <li
                      key={slug}
                      className="flex items-center justify-between gap-3 py-3 first:pt-0 last:pb-0"
                    >
                      <div className="min-w-0">
                        <div className="font-medium">{name}</div>
                        <p className="text-xs text-muted-foreground">
                          {p.skill_slug} · {p.status ?? "ACTIVE"}
                        </p>
                      </div>
                      <Link
                        href={`/pods/${slug}`}
                        className={cn(buttonVariants({ size: "sm" }), "gap-1")}
                      >
                        <Headphones className="size-3.5" />
                        Chat & audio
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
