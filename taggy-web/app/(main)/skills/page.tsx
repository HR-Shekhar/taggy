"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { ArrowRight, BookOpen, Route } from "lucide-react";
import {
  apiErrorMessage,
  createSkillRequest,
  isFreeSkillLimitError,
  joinSkill,
  listMySkills,
  listSkills,
  type MySkill,
  type SimilarSkill,
} from "@/lib/api";
import { useAuth } from "@/lib/auth";
import {
  Empty,
  ErrorBox,
  GenerationWaitNote,
  PageHeader,
  PageSkeleton,
  Section,
} from "@/components/app-ui";
import { EmptyArtSkills } from "@/components/empty-art";
import { PremiumUpgradePrompt } from "@/components/premium-upgrade";
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
import { cn } from "@/lib/utils";

type CatalogSkill = {
  name: string;
  slug: string;
  description?: string | null;
};

export default function SkillsPage() {
  const { username } = useAuth();
  const [skills, setSkills] = useState<CatalogSkill[]>([]);
  const [mine, setMine] = useState<MySkill[]>([]);
  const [loading, setLoading] = useState(true);
  const [pageError, setPageError] = useState<string | null>(null);
  const [showUpgrade, setShowUpgrade] = useState(false);
  const [busySlug, setBusySlug] = useState<string | null>(null);

  const [reqName, setReqName] = useState("");
  const [reqDesc, setReqDesc] = useState("");
  const [similar, setSimilar] = useState<SimilarSkill[] | null>(null);
  const [reqBusy, setReqBusy] = useState(false);

  async function load() {
    setLoading(true);
    setPageError(null);
    const [catalog, enrolled] = await Promise.all([
      listSkills(),
      username ? listMySkills(username) : Promise.resolve(null),
    ]);
    setLoading(false);
    if (!catalog.ok) {
      const message = apiErrorMessage(catalog);
      setPageError(message);
      return;
    }
    setSkills(catalog.data ?? []);
    if (enrolled && enrolled.ok) setMine(enrolled.data ?? []);
  }

  useEffect(() => {
    void load();
  }, [username]);

  const enrolledBySlug = new Map(mine.map((s) => [s.skill_slug, s]));

  async function submitSkillRequest(force: boolean) {
    setReqBusy(true);
    setSimilar(null);
    const res = await createSkillRequest({
      name: reqName.trim(),
      description: reqDesc.trim() || undefined,
      force,
    });
    setReqBusy(false);
    if (!res.ok) {
      toastApiError(res, "Couldn't submit request");
      return;
    }
    if (res.data?.requires_confirm) {
      setSimilar(res.data.similar ?? []);
      return;
    }
    const count = res.data?.request?.draft_milestones?.length ?? 0;
    const status = res.data?.request?.status ?? "";
    setReqName("");
    setReqDesc("");
    setSimilar(null);
    if (status === "GENERATING") {
      toastSuccess(
        "AI is drafting your roadmap in the background (this can take several minutes). Check My requests or notifications when it's ready.",
        "Generating roadmap"
      );
      return;
    }
    toastSuccess(
      count > 0
        ? `Your skill request was drafted with ${count} milestones (topics + subtopics) and is pending admin review.`
        : "Your skill request was submitted and is pending admin review.",
      "Roadmap draft ready"
    );
  }

  if (loading) return <PageSkeleton variant="list" />;

  return (
    <div className="space-y-8" data-tour="skills-page">
      <PageHeader
        title="Skills"
        description="Pick a skill, follow a versioned roadmap, and track milestones."
      />
      {pageError ? (
        <ErrorBox message={pageError} title="Couldn't load skills" />
      ) : null}

      {mine.length > 0 && (
        <Section title="Your skills">
          <ul className="divide-y divide-border overflow-hidden rounded-xl border border-border bg-card">
            {mine.map((s) => (
              <li key={s.skill_slug}>
                <Link
                  href={`/skills/${s.skill_slug}`}
                  className="flex items-center gap-4 px-4 py-3.5 transition-colors hover:bg-muted/40"
                >
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <p className="truncate font-medium">{s.skill_name}</p>
                      <Badge variant="secondary">
                        {Math.round(s.completion_percent)}%
                      </Badge>
                    </div>
                    <p className="mt-0.5 text-sm text-muted-foreground">
                      Roadmap v{s.roadmap_version_number} · {s.completed_count}/
                      {s.milestone_count} milestones
                    </p>
                    <div className="mt-2 h-1.5 overflow-hidden rounded-full bg-muted">
                      <div
                        className="h-full rounded-full bg-primary"
                        style={{
                          width: `${Math.min(100, s.completion_percent)}%`,
                        }}
                      />
                    </div>
                  </div>
                  <ArrowRight className="size-4 shrink-0 text-muted-foreground" />
                </Link>
              </li>
            ))}
          </ul>
        </Section>
      )}

      <Section title="Browse catalog">
        {showUpgrade ? <PremiumUpgradePrompt message="limit" /> : null}
        {skills.length === 0 ? (
          <Empty
            art={<EmptyArtSkills />}
            title="No skills yet"
            description="Request a new skill below to get started."
          />
        ) : (
          <div className="grid gap-3 sm:grid-cols-2" data-tour="skills-catalog">
            {skills.map((s) => {
              const enrolled = enrolledBySlug.get(s.slug);
              return (
                <Card key={s.slug}>
                  <CardHeader>
                    <div className="flex items-start gap-3">
                      <span className="inline-flex size-9 items-center justify-center rounded-lg bg-secondary text-secondary-foreground">
                        <BookOpen className="size-4 text-primary" />
                      </span>
                      <div className="min-w-0 flex-1">
                        <CardTitle className="font-serif text-lg">
                          {s.name}
                        </CardTitle>
                        <CardDescription className="line-clamp-2">
                          {s.description ||
                            "Structured roadmap with ordered milestones."}
                        </CardDescription>
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent className="flex flex-wrap items-center gap-2">
                    {enrolled ? (
                      <Badge variant="outline" className="gap-1">
                        <Route className="size-3" />
                        Enrolled · v{enrolled.roadmap_version_number}
                      </Badge>
                    ) : null}
                    <Link
                      href={`/skills/${s.slug}`}
                      className={cn(
                        buttonVariants({ variant: "outline", size: "sm" })
                      )}
                    >
                      View
                    </Link>
                    {!enrolled && (
                      <button
                        type="button"
                        className={cn(buttonVariants({ size: "sm" }))}
                        disabled={busySlug === s.slug}
                        onClick={async () => {
                          setBusySlug(s.slug);
                          setShowUpgrade(false);
                          const result = await joinSkill(s.slug);
                          setBusySlug(null);
                          if (!result.ok) {
                            if (isFreeSkillLimitError(result)) {
                              setShowUpgrade(true);
                            } else {
                              toastApiError(result, "Couldn't join skill");
                            }
                          } else void load();
                        }}
                      >
                        {busySlug === s.slug ? "Joining…" : "Join"}
                      </button>
                    )}
                  </CardContent>
                </Card>
              );
            })}
          </div>
        )}
      </Section>

      <Section title="Request a new skill">
        <div data-tour="skills-request">
          <Card>
            <CardHeader>
              <CardTitle className="font-serif text-lg">Suggest a skill</CardTitle>
              <CardDescription>
                AI drafts a course-style outline for admin review. Generation may
                take a few minutes.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <input
                className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
                placeholder="Skill name"
                value={reqName}
                onChange={(e) => {
                  setReqName(e.target.value);
                }}
                disabled={reqBusy}
              />
              <textarea
                className="min-h-20 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
                placeholder="Short description (optional)"
                value={reqDesc}
                onChange={(e) => setReqDesc(e.target.value)}
                disabled={reqBusy}
              />
              <GenerationWaitNote active={reqBusy} />
              {similar && similar.length > 0 ? (
                <div className="space-y-2 rounded-lg border border-border bg-muted/40 p-3">
                  <p className="text-sm font-medium">
                    Similar skills already exist
                  </p>
                  <p className="text-sm text-muted-foreground">
                    Review these first, or confirm if you still want a new skill.
                  </p>
                  <ul className="space-y-1 text-sm">
                    {similar.map((s) => (
                      <li key={s.slug}>
                        <Link href={`/skills/${s.slug}`} className="underline">
                          {s.name}
                        </Link>
                        <span className="text-muted-foreground">
                          {" "}
                          · score {s.score.toFixed(2)}
                        </span>
                      </li>
                    ))}
                  </ul>
                  <Button
                    size="sm"
                    variant="outline"
                    disabled={reqBusy || reqName.trim().length < 3}
                    onClick={() => void submitSkillRequest(true)}
                  >
                    Submit anyway
                  </Button>
                </div>
              ) : null}
              <div className="flex flex-wrap gap-2">
                <Button
                  disabled={reqBusy || reqName.trim().length < 3}
                  onClick={() => void submitSkillRequest(false)}
                >
                  {reqBusy ? "Generating…" : "Check & submit"}
                </Button>
                <Link
                  href="/requests"
                  className={cn(buttonVariants({ variant: "ghost", size: "sm" }))}
                >
                  View my requests
                </Link>
              </div>
            </CardContent>
          </Card>
        </div>
      </Section>
    </div>
  );
}
