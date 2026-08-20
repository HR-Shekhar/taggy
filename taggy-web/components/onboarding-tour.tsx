"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useLayoutEffect,
  useMemo,
  useState,
  type CSSProperties,
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
  href?: string;
  /** Preferred data-tour targets (first match in DOM wins) */
  targets: string[];
  ctaLabel: string;
};

const STEPS: Record<TourStepId, StepDef> = {
  home_overview: {
    id: "home_overview",
    title: "Your Home hub",
    body: "From here you continue skills, check streaks, search, and jump into pods. Next, open Skills in the sidebar (or tap below) to join or request a roadmap.",
    href: "/home",
    targets: ["home-cta", "nav-home", "home-main"],
    ctaLabel: "Go to Skills",
  },
  skills_join_or_request: {
    id: "skills_join_or_request",
    title: "Join or request a skill",
    body: "Click Join on a catalog skill, or use Request a new skill below to generate a roadmap for something you want to learn.",
    href: "/skills",
    targets: ["skills-request", "skills-catalog", "nav-skills", "skills-page"],
    ctaLabel: "Continue",
  },
  skills_waiting: {
    id: "skills_waiting",
    title: "Enroll to keep going",
    body: "Join a skill (or wait for your request to be approved). Click a Join button on this page — this tip advances once you have a skill.",
    href: "/skills",
    targets: ["skills-catalog", "nav-skills", "skills-page"],
    ctaLabel: "I’ve joined — continue",
  },
  pods_join_or_create: {
    id: "pods_join_or_create",
    title: "Find your pod",
    body: "Click Pods in the sidebar, then join an open pod or create one for your skill so you can chat and study together.",
    href: "/pods",
    targets: ["pods-create", "pods-page", "nav-pods"],
    ctaLabel: "Continue",
  },
  pods_waiting: {
    id: "pods_waiting",
    title: "Join or create a pod",
    body: "Use Join / Create on this page. Once you’re accepted into a pod, we’ll open it and show what’s inside.",
    href: "/pods",
    targets: ["pods-create", "pods-page", "nav-pods"],
    ctaLabel: "I’m in a pod — continue",
  },
  pod_features: {
    id: "pod_features",
    title: "Inside your pod",
    body: "Use chat, start audio rooms, manage members, and try the pod quiz/leaderboard. Click around here — then continue to Community.",
    targets: ["pod-workspace", "nav-pods"],
    ctaLabel: "Go to Community",
  },
  community_leaderboards: {
    id: "community_leaderboards",
    title: "Community & leaderboards",
    body: "Click a skill community to open channels. Leaderboards also show on Home and inside pods — open Community from the sidebar anytime.",
    href: "/community",
    targets: ["community-page", "nav-community"],
    ctaLabel: "Finish tour",
  },
  done: {
    id: "done",
    title: "You’re set",
    body: "Path to remember: skill → roadmap → pod → community. Click Home when you’re ready to start learning.",
    href: "/home",
    targets: ["nav-home", "home-cta"],
    ctaLabel: "Go to Home",
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

function doneKey(username: string) {
  return `${ONBOARDING_DONE_KEY}:${username}`;
}
function stepKey(username: string) {
  return `${ONBOARDING_STEP_KEY}:${username}`;
}

function readDone(username: string) {
  if (typeof window === "undefined") return true;
  if (window.localStorage.getItem(doneKey(username)) === "1") return true;
  // Legacy global flag from earlier tour
  if (window.localStorage.getItem(ONBOARDING_DONE_KEY) === "1") return true;
  return false;
}

function readStep(username: string): TourStepId | null {
  if (typeof window === "undefined") return null;
  const raw =
    window.localStorage.getItem(stepKey(username)) ??
    window.localStorage.getItem(ONBOARDING_STEP_KEY);
  if (raw && STEP_ORDER.includes(raw as TourStepId)) return raw as TourStepId;
  return null;
}

function writeStep(username: string, id: TourStepId) {
  window.localStorage.setItem(stepKey(username), id);
}

function markDone(username: string) {
  window.localStorage.setItem(doneKey(username), "1");
  window.localStorage.removeItem(stepKey(username));
  window.localStorage.setItem(ONBOARDING_DONE_KEY, "1");
  window.localStorage.removeItem(ONBOARDING_STEP_KEY);
  window.dispatchEvent(new Event("taggy-onboarding-done"));
}

function stepIndex(id: TourStepId) {
  return STEP_ORDER.indexOf(id);
}

type SpotlightRect = {
  top: number;
  left: number;
  width: number;
  height: number;
};

function findTargetEl(targets: string[]) {
  for (const id of targets) {
    const el = document.querySelector(`[data-tour="${id}"]`);
    if (el instanceof HTMLElement) {
      const r = el.getBoundingClientRect();
      // Prefer visible, reasonably sized targets
      if (r.width > 0 && r.height > 0) return el;
    }
  }
  return null;
}

function useTargetRect(targets: string[], active: boolean) {
  const [rect, setRect] = useState<SpotlightRect | null>(null);

  const measure = useCallback(() => {
    if (!active || targets.length === 0) {
      setRect(null);
      return;
    }
    const el = findTargetEl(targets);
    if (!el) {
      setRect(null);
      return;
    }
    const r = el.getBoundingClientRect();
    // Cap spotlight size so giant page wrappers don't swallow the UI
    const maxW = Math.min(r.width, Math.min(420, window.innerWidth - 32));
    const maxH = Math.min(r.height, Math.min(280, window.innerHeight - 32));
    setRect({
      top: r.top,
      left: r.left,
      width: maxW,
      height: maxH,
    });
  }, [targets, active]);

  useLayoutEffect(() => {
    if (!active) return;
    const el = findTargetEl(targets);
    el?.scrollIntoView({ block: "nearest", inline: "nearest", behavior: "smooth" });
    measure();
    window.addEventListener("resize", measure);
    window.addEventListener("scroll", measure, true);
    const t = window.setInterval(measure, 400);
    return () => {
      window.removeEventListener("resize", measure);
      window.removeEventListener("scroll", measure, true);
      window.clearInterval(t);
    };
  }, [measure, active, targets]);

  return rect;
}

function clamp(n: number, min: number, max: number) {
  return Math.min(max, Math.max(min, n));
}

function tipPosition(rect: SpotlightRect | null): CSSProperties {
  const tipW = Math.min(352, window.innerWidth - 24);
  const tipH = 240;
  const margin = 12;

  if (!rect) {
    return {
      position: "fixed",
      right: margin,
      bottom: margin,
      width: tipW,
    };
  }

  const spaceRight = window.innerWidth - (rect.left + rect.width);
  const spaceBelow = window.innerHeight - (rect.top + rect.height);

  let top: number;
  let left: number;

  // Prefer to the right of sidebar-sized targets
  if (spaceRight > tipW + 24 && rect.width < 280) {
    left = rect.left + rect.width + 14;
    top = rect.top;
  } else if (spaceBelow > tipH + 16) {
    left = rect.left;
    top = rect.top + rect.height + 14;
  } else {
    // Place above or pinned to lower viewport
    left = rect.left;
    top = rect.top - tipH - 14;
  }

  left = clamp(left, margin, window.innerWidth - tipW - margin);
  top = clamp(top, margin, window.innerHeight - tipH - margin);

  return {
    position: "fixed",
    top,
    left,
    width: tipW,
  };
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
  const rect = useTargetRect(step.targets, true);
  const pad = 6;
  const style = useMemo(() => tipPosition(rect), [rect]);

  return (
    <div className="pointer-events-none fixed inset-0 z-[55] overflow-hidden">
      {rect ? (
        <div
          aria-hidden
          className="pointer-events-none absolute rounded-xl ring-2 ring-primary/90 ring-offset-2 ring-offset-background"
          style={{
            top: rect.top - pad,
            left: rect.left - pad,
            width: rect.width + pad * 2,
            height: rect.height + pad * 2,
            boxShadow: "0 0 0 9999px rgb(0 0 0 / 0.22)",
          }}
        />
      ) : null}

      <div
        className="pointer-events-auto z-[56] max-h-[min(70vh,22rem)] overflow-y-auto rounded-xl border border-border bg-card p-4 shadow-xl"
        style={style}
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
            <Button
              type="button"
              size="sm"
              className="gap-1.5"
              onClick={onGoThere}
            >
              Take me there
              <ArrowRight className="size-3.5" />
            </Button>
          ) : (
            <Button
              type="button"
              size="sm"
              className="gap-1.5"
              onClick={onNext}
            >
              {step.ctaLabel}
              <ArrowRight className="size-3.5" />
            </Button>
          )}
          <Button type="button" size="sm" variant="ghost" onClick={onSkip}>
            Skip tour
          </Button>
        </div>
      </div>
    </div>
  );
}

export function OnboardingProvider({ children }: { children: ReactNode }) {
  const { username, ready } = useAuth();
  const pathname = usePathname();
  const router = useRouter();
  const [active, setActive] = useState(false);
  const [stepId, setStepId] = useState<TourStepId | null>(null);
  const [acceptedPodSlug, setAcceptedPodSlug] = useState<string | null>(null);
  const [skillCount, setSkillCount] = useState(0);
  const [podCount, setPodCount] = useState(0);
  const [bootstrapped, setBootstrapped] = useState(false);

  // New users only: per-username flag; skip if they already have skills/pods
  useEffect(() => {
    if (!ready || !username) return;
    let cancelled = false;

    (async () => {
      if (readDone(username)) {
        if (!cancelled) {
          setActive(false);
          setStepId(null);
          setBootstrapped(true);
        }
        return;
      }

      const inProgress = readStep(username);
      const [sk, pd] = await Promise.all([
        listMySkills(username),
        listMyPods(username),
      ]);
      if (cancelled) return;

      const skills = sk.ok ? sk.data ?? [] : [];
      const pods = pd.ok
        ? Array.isArray(pd.data)
          ? pd.data
          : ((pd.data as { pods?: { status?: string }[] })?.pods ?? [])
        : [];
      const accepted = pods.filter(
        (p) => (p.status ?? "ACCEPTED") === "ACCEPTED"
      );

      setSkillCount(skills.length);
      setPodCount(accepted.length);
      const first = accepted[0] as
        | { pod_slug?: string; slug?: string }
        | undefined;
      setAcceptedPodSlug(first?.pod_slug ?? first?.slug ?? null);

      // Existing users (already enrolled) — never show tour unless mid-tour
      if (!inProgress && (skills.length > 0 || accepted.length > 0)) {
        markDone(username);
        setActive(false);
        setStepId(null);
        setBootstrapped(true);
        return;
      }

      const start = inProgress ?? "home_overview";
      writeStep(username, start);
      setStepId(start);
      setActive(true);
      setBootstrapped(true);
    })();

    return () => {
      cancelled = true;
    };
  }, [ready, username]);

  const goToStep = useCallback(
    (id: TourStepId) => {
      if (!username) return;
      writeStep(username, id);
      setStepId(id);
      setActive(true);
    },
    [username]
  );

  const skip = useCallback(() => {
    if (!username) return;
    markDone(username);
    setActive(false);
    setStepId(null);
  }, [username]);

  const advanceFrom = useCallback(
    (current: TourStepId) => {
      const i = stepIndex(current);
      if (i < 0 || i >= STEP_ORDER.length - 1) {
        skip();
        return;
      }
      goToStep(STEP_ORDER[i + 1]);
    },
    [goToStep, skip]
  );

  const next = useCallback(() => {
    if (!stepId) return;
    if (stepId === "done") {
      skip();
      return;
    }
    advanceFrom(stepId);
  }, [stepId, advanceFrom, skip]);

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

  useEffect(() => {
    if (!active || !stepId) return;
    if (stepId === "skills_waiting" && skillCount > 0) {
      goToStep("pods_join_or_create");
    }
    if (stepId === "pods_waiting" && podCount > 0) {
      goToStep("pod_features");
    }
  }, [active, stepId, skillCount, podCount, goToStep]);

  const step = stepId ? STEPS[stepId] : null;

  const onPreferredRoute = useMemo(() => {
    if (!step) return true;
    if (step.id === "pod_features") return /^\/pods\/[^/]+/.test(pathname);
    if (step.id === "skills_waiting") return pathname.startsWith("/skills");
    if (step.id === "pods_waiting") return pathname.startsWith("/pods");
    if (!step.href) return true;
    if (step.href === "/home") return pathname === "/home";
    return pathname === step.href || pathname.startsWith(`${step.href}/`);
  }, [step, pathname]);

  const goThere = useCallback(() => {
    if (!step) return;
    if (step.id === "pod_features") {
      router.push(acceptedPodSlug ? `/pods/${acceptedPodSlug}` : "/pods");
      return;
    }
    if (step.href) router.push(step.href);
  }, [step, acceptedPodSlug, router]);

  useEffect(() => {
    if (!active || stepId !== "pod_features") return;
    if (!/^\/pods\/[^/]+/.test(pathname) && acceptedPodSlug) {
      router.push(`/pods/${acceptedPodSlug}`);
    }
  }, [active, stepId, pathname, acceptedPodSlug, router]);

  const value = useMemo(
    () => ({ active, stepId, skip, next, goToStep }),
    [active, stepId, skip, next, goToStep]
  );

  return (
    <OnboardingContext.Provider value={value}>
      {children}
      {bootstrapped && active && step ? (
        <Coachmark
          step={step}
          stepNum={stepIndex(step.id) + 1}
          stepTotal={STEP_ORDER.length}
          onSkip={skip}
          needsNavigate={!onPreferredRoute}
          onGoThere={goThere}
          onNext={() => {
            if (step.id === "done") {
              skip();
              router.push("/home");
              return;
            }
            if (step.id === "home_overview") {
              goToStep("skills_join_or_request");
              router.push("/skills");
              return;
            }
            if (step.id === "skills_join_or_request") {
              goToStep(skillCount > 0 ? "pods_join_or_create" : "skills_waiting");
              if (skillCount > 0) router.push("/pods");
              return;
            }
            if (step.id === "pods_join_or_create") {
              goToStep(podCount > 0 ? "pod_features" : "pods_waiting");
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
        />
      ) : null}
    </OnboardingContext.Provider>
  );
}

export function OnboardingChecklist() {
  return null;
}

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
