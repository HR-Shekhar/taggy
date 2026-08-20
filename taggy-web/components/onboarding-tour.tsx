"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useLayoutEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { usePathname, useRouter } from "next/navigation";
import { ArrowRight } from "lucide-react";
import { useAuth } from "@/lib/auth";
import { listMyPods, listMySkills } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export const ONBOARDING_DONE_KEY = "taggy_onboarding_done";
export const ONBOARDING_STEP_KEY = "taggy_onboarding_step";

export type TourStepId =
  | "home_overview"
  | "skills_join_or_request"
  | "skills_waiting"
  | "pods_join_or_create"
  | "pods_waiting"
  | "pod_features"
  | "community_leaderboards"
  | "done";

const STEP_ORDER: TourStepId[] = [
  "home_overview",
  "skills_join_or_request",
  "skills_waiting",
  "pods_join_or_create",
  "pods_waiting",
  "pod_features",
  "community_leaderboards",
  "done",
];

type StepDef = {
  id: TourStepId;
  title: string;
  body: string;
  /** Preferred page for this tip */
  href?: string;
  /** data-tour target to spotlight */
  target: string;
  ctaLabel?: string;
  /** If true, Next advances; else CTA navigates first */
  nextAdvances?: boolean;
};

const STEPS: Record<TourStepId, StepDef> = {
  home_overview: {
    id: "home_overview",
    title: "Welcome to your Home",
    body: "This is your hub: continue a skill, check your streak and study time, search Taggy, jump into pods, and see unread alerts. Use the sidebar anytime to move around.",
    href: "/home",
    target: "home-main",
    ctaLabel: "Next: pick a skill",
    nextAdvances: true,
  },
  skills_join_or_request: {
    id: "skills_join_or_request",
    title: "Join a skill or request a roadmap",
    body: "Browse the catalog and join a skill you want to learn. Prefer something new? Request a roadmap for your skill — Taggy will generate a followable path.",
    href: "/skills",
    target: "skills-page",
    ctaLabel: "Got it",
    nextAdvances: true,
  },
  skills_waiting: {
    id: "skills_waiting",
    title: "Enroll to continue",
    body: "Join a skill from the list, or submit a request and wait until it’s approved. This tip advances automatically once you have at least one skill.",
    href: "/skills",
    target: "nav-skills",
    ctaLabel: "I’ve joined — continue",
    nextAdvances: true,
  },
  pods_join_or_create: {
    id: "pods_join_or_create",
    title: "Find or create a pod",
    body: "Pods are small accountability groups for a skill. Join an open pod or create your own so you can chat, study live, and stay consistent together.",
    href: "/pods",
    target: "pods-page",
    ctaLabel: "Got it",
    nextAdvances: true,
  },
  pods_waiting: {
    id: "pods_waiting",
    title: "Join or create a pod",
    body: "Request to join a pod (or create one). Once you’re accepted, we’ll show you what’s inside. Advances automatically when you have an accepted pod.",
    href: "/pods",
    target: "nav-pods",
    ctaLabel: "I’m in a pod — continue",
    nextAdvances: true,
  },
  pod_features: {
    id: "pod_features",
    title: "Inside your pod",
    body: "Here you get group chat, live audio rooms, member roles, join requests (if you lead), and pod quizzes/leaderboards. Stay connected — audio and chat can keep running while you browse.",
    target: "pod-workspace",
    ctaLabel: "Next: community",
    nextAdvances: true,
  },
  community_leaderboards: {
    id: "community_leaderboards",
    title: "Community & leaderboards",
    body: "Open Community for skill-wide channels. Leaderboards on Home, Progress, and inside pods show how you stack up — great motivation with your group.",
    href: "/community",
    target: "community-page",
    ctaLabel: "Finish tour",
    nextAdvances: true,
  },
  done: {
    id: "done",
    title: "You’re set",
    body: "You know the path: skill → roadmap → pod → community. Explore Progress to mark topics done, and watch Notifications for updates.",
    href: "/home",
    target: "nav-home",
    ctaLabel: "Start learning",
    nextAdvances: true,
  },
};

type OnboardingContextValue = {
  active: boolean;
  stepId: TourStepId | null;
  skip: () => void;
  next: () => void;
  goToStep: (id: TourStepId) => void;
};

const OnboardingContext = createContext<OnboardingContextValue | null>(null);

export function useOnboarding() {
  return useContext(OnboardingContext);
}

function readDone() {
  if (typeof window === "undefined") return true;
  return window.localStorage.getItem(ONBOARDING_DONE_KEY) === "1";
}

function readStep(): TourStepId {
  if (typeof window === "undefined") return "home_overview";
  const raw = window.localStorage.getItem(ONBOARDING_STEP_KEY);
  if (raw && STEP_ORDER.includes(raw as TourStepId)) return raw as TourStepId;
  return "home_overview";
}

function writeStep(id: TourStepId) {
  window.localStorage.setItem(ONBOARDING_STEP_KEY, id);
}

function markDone() {
  window.localStorage.setItem(ONBOARDING_DONE_KEY, "1");
  window.localStorage.removeItem(ONBOARDING_STEP_KEY);
  window.dispatchEvent(new Event("taggy-onboarding-done"));
}

function stepIndex(id: TourStepId) {
  return STEP_ORDER.indexOf(id);
}

type SpotlightRect = { top: number; left: number; width: number; height: number };

function useTargetRect(target: string | null, active: boolean) {
  const [rect, setRect] = useState<SpotlightRect | null>(null);

  const measure = useCallback(() => {
    if (!target || !active) {
      setRect(null);
      return;
    }
    const el = document.querySelector(`[data-tour="${target}"]`);
    if (!el) {
      setRect(null);
      return;
    }
    const r = el.getBoundingClientRect();
    setRect({
      top: r.top,
      left: r.left,
      width: r.width,
      height: r.height,
    });
  }, [target, active]);

  useLayoutEffect(() => {
    measure();
    if (!active) return;
    window.addEventListener("resize", measure);
    window.addEventListener("scroll", measure, true);
    const t = window.setInterval(measure, 500);
    return () => {
      window.removeEventListener("resize", measure);
      window.removeEventListener("scroll", measure, true);
      window.clearInterval(t);
    };
  }, [measure, active]);

  return rect;
}

function Coachmark({
  step,
  stepNum,
  stepTotal,
  onNext,
  onSkip,
  onGoThere,
  needsNavigate,
}: {
  step: StepDef;
  stepNum: number;
  stepTotal: number;
  onNext: () => void;
  onSkip: () => void;
  onGoThere: () => void;
  needsNavigate: boolean;
}) {
  const rect = useTargetRect(step.target, true);
  const pad = 8;

  const tipStyle = useMemo(() => {
    if (!rect) {
      return { bottom: 24, right: 24 } as const;
    }
    const spaceBelow = window.innerHeight - (rect.top + rect.height);
    const preferBelow = spaceBelow > 220;
    const left = Math.min(
      Math.max(16, rect.left),
      window.innerWidth - 340
    );
    if (preferBelow) {
      return {
        top: rect.top + rect.height + pad + 12,
        left,
      } as const;
    }
    return {
      bottom: Math.max(16, window.innerHeight - rect.top + 12),
      left,
    } as const;
  }, [rect]);

  return (
    <div className="pointer-events-none fixed inset-0 z-[55]">
      {rect ? (
        <div
          aria-hidden
          className="pointer-events-none absolute rounded-xl ring-2 ring-primary ring-offset-2 ring-offset-background transition-all duration-200"
          style={{
            top: rect.top - pad,
            left: rect.left - pad,
            width: rect.width + pad * 2,
            height: rect.height + pad * 2,
            boxShadow: "0 0 0 9999px rgb(0 0 0 / 0.28)",
          }}
        />
      ) : null}

      <div
        className="pointer-events-auto absolute w-[min(22rem,calc(100vw-2rem))] rounded-xl border border-border bg-card p-4 shadow-xl"
        style={tipStyle}
        role="dialog"
        aria-labelledby="tour-title"
      >
        <p className="text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
          Tour · {stepNum}/{stepTotal}
        </p>
        <h2 id="tour-title" className="mt-1 font-serif text-lg leading-snug">
          {step.title}
        </h2>
        <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
          {step.body}
        </p>
        <div className="mt-4 flex flex-wrap items-center gap-2">
          {needsNavigate ? (
            <Button type="button" size="sm" className="gap-1.5" onClick={onGoThere}>
              Take me there
              <ArrowRight className="size-3.5" />
            </Button>
          ) : (
            <Button type="button" size="sm" className="gap-1.5" onClick={onNext}>
              {step.ctaLabel ?? "Next"}
              <ArrowRight className="size-3.5" />
            </Button>
          )}
          {!needsNavigate && step.href && step.id !== "done" ? (
            <Button type="button" size="sm" variant="outline" onClick={onNext}>
              Next
            </Button>
          ) : null}
          <Button type="button" size="sm" variant="ghost" onClick={onSkip}>
            Skip tour
          </Button>
        </div>
      </div>
    </div>
  );
}

export function OnboardingProvider({ children }: { children: ReactNode }) {
  const { username } = useAuth();
  const pathname = usePathname();
  const router = useRouter();
  const [active, setActive] = useState(false);
  const [stepId, setStepId] = useState<TourStepId | null>(null);
  const [acceptedPodSlug, setAcceptedPodSlug] = useState<string | null>(null);
  const [skillCount, setSkillCount] = useState(0);
  const [podCount, setPodCount] = useState(0);

  useEffect(() => {
    if (readDone()) {
      setActive(false);
      setStepId(null);
      return;
    }
    const s = readStep();
    setStepId(s);
    setActive(true);
  }, []);

  const goToStep = useCallback((id: TourStepId) => {
    writeStep(id);
    setStepId(id);
    setActive(true);
  }, []);

  const skip = useCallback(() => {
    markDone();
    setActive(false);
    setStepId(null);
  }, []);

  const advanceFrom = useCallback(
    (current: TourStepId) => {
      const i = stepIndex(current);
      if (i < 0 || i >= STEP_ORDER.length - 1) {
        markDone();
        setActive(false);
        setStepId(null);
        return;
      }
      const nextId = STEP_ORDER[i + 1];
      if (nextId === "done") {
        goToStep("done");
        return;
      }
      goToStep(nextId);
    },
    [goToStep]
  );

  const next = useCallback(() => {
    if (!stepId) return;
    if (stepId === "done") {
      skip();
      return;
    }
    advanceFrom(stepId);
  }, [stepId, advanceFrom, skip]);

  // Poll enrollment so waiting steps can auto-advance
  useEffect(() => {
    if (!active || !username) return;

    let cancelled = false;
    async function refresh() {
      const [sk, pd] = await Promise.all([
        listMySkills(username!),
        listMyPods(username!),
      ]);
      if (cancelled) return;
      const skills = sk.ok ? sk.data ?? [] : [];
      setSkillCount(skills.length);

      const pods = pd.ok
        ? Array.isArray(pd.data)
          ? pd.data
          : ((pd.data as { pods?: { pod_slug?: string; slug?: string; status?: string }[] })
              ?.pods ?? [])
        : [];
      const accepted = pods.filter(
        (p) => (p.status ?? "ACCEPTED") === "ACCEPTED"
      );
      setPodCount(accepted.length);
      const first = accepted[0];
      setAcceptedPodSlug(first?.pod_slug ?? first?.slug ?? null);
    }

    void refresh();
    const t = window.setInterval(() => void refresh(), 4000);
    return () => {
      cancelled = true;
      window.clearInterval(t);
    };
  }, [active, username]);

  // Auto-advance waiting steps
  useEffect(() => {
    if (!active || !stepId) return;
    if (stepId === "skills_waiting" && skillCount > 0) {
      goToStep("pods_join_or_create");
    }
    if (stepId === "pods_waiting" && podCount > 0) {
      goToStep("pod_features");
    }
  }, [active, stepId, skillCount, podCount, goToStep]);

  // Skip waiting if user already has progress when landing on those steps
  useEffect(() => {
    if (!active || !stepId) return;
    if (stepId === "skills_join_or_request" && skillCount > 0) {
      // Still show join tip briefly is ok; auto-skip waiting later
    }
    if (stepId === "skills_waiting" && skillCount > 0) {
      goToStep("pods_join_or_create");
    }
    if (
      (stepId === "pods_join_or_create" || stepId === "pods_waiting") &&
      podCount > 0 &&
      stepId === "pods_waiting"
    ) {
      goToStep("pod_features");
    }
  }, [active, stepId, skillCount, podCount, goToStep]);

  const step = stepId ? STEPS[stepId] : null;

  const onPreferredRoute = useMemo(() => {
    if (!step) return true;
    if (step.id === "pod_features") {
      return /^\/pods\/[^/]+/.test(pathname);
    }
    if (step.id === "skills_waiting") {
      return pathname.startsWith("/skills");
    }
    if (step.id === "pods_waiting") {
      return pathname.startsWith("/pods");
    }
    if (!step.href) return true;
    if (step.href === "/home") return pathname === "/home";
    return pathname === step.href || pathname.startsWith(step.href + "/");
  }, [step, pathname]);

  const goThere = useCallback(() => {
    if (!step) return;
    if (step.id === "pod_features") {
      if (acceptedPodSlug) {
        router.push(`/pods/${acceptedPodSlug}`);
      } else {
        router.push("/pods");
      }
      return;
    }
    if (step.href) router.push(step.href);
  }, [step, acceptedPodSlug, router]);

  // When advancing into pod_features, navigate into a pod if possible
  useEffect(() => {
    if (!active || stepId !== "pod_features") return;
    if (!/^\/pods\/[^/]+/.test(pathname) && acceptedPodSlug) {
      router.push(`/pods/${acceptedPodSlug}`);
    }
  }, [active, stepId, pathname, acceptedPodSlug, router]);

  // When advancing into community, nudge to /community
  useEffect(() => {
    if (!active || stepId !== "community_leaderboards") return;
    if (!pathname.startsWith("/community")) {
      // don't force-nav; tip offers Take me there if off-route
    }
  }, [active, stepId, pathname]);

  const value = useMemo(
    () => ({
      active,
      stepId,
      skip,
      next,
      goToStep,
    }),
    [active, stepId, skip, next, goToStep]
  );

  const tipStepTotal = STEP_ORDER.length;
  const tipStepNum = stepId ? stepIndex(stepId) + 1 : 0;

  return (
    <OnboardingContext.Provider value={value}>
      {children}
      {active && step ? (
        <Coachmark
          step={step}
          stepNum={tipStepNum}
          stepTotal={tipStepTotal}
          onSkip={skip}
          onNext={() => {
            if (step.id === "done") {
              skip();
              if (step.href) router.push(step.href);
              return;
            }
            // After join/request tips, move to waiting; after features go community
            if (step.id === "skills_join_or_request") {
              if (skillCount > 0) goToStep("pods_join_or_create");
              else goToStep("skills_waiting");
              return;
            }
            if (step.id === "pods_join_or_create") {
              if (podCount > 0) goToStep("pod_features");
              else goToStep("pods_waiting");
              return;
            }
            if (step.id === "home_overview") {
              goToStep("skills_join_or_request");
              router.push("/skills");
              return;
            }
            if (step.id === "pod_features") {
              goToStep("community_leaderboards");
              router.push("/community");
              return;
            }
            if (step.id === "community_leaderboards") {
              goToStep("done");
              return;
            }
            next();
          }}
          onGoThere={goThere}
          needsNavigate={!onPreferredRoute}
        />
      ) : null}
    </OnboardingContext.Provider>
  );
}

/** @deprecated checklist removed — tour owns guidance */
export function OnboardingChecklist() {
  return null;
}

/** Mounted tip host kept for layout compatibility */
export function OnboardingTour() {
  return null;
}

export function TourTarget({
  id,
  className,
  children,
}: {
  id: string;
  className?: string;
  children: ReactNode;
}) {
  return (
    <div data-tour={id} className={cn(className)}>
      {children}
    </div>
  );
}
