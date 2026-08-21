import { CheckCircle2, Users, Map, Target, Mic, MessageSquare, Bell } from "lucide-react";
import { Reveal } from "@/components/marketing/reveal";

const features = [
  {
    title: "Structured skills",
    description:
      "Pick a skill, join its community, and follow a milestone roadmap built for real outcomes.",
    icon: Map,
  },
  {
    title: "Accountability pods",
    description:
      "Create or join small pods. Request to join, get accepted, and study together.",
    icon: Users,
  },
  {
    title: "Progress you can see",
    description:
      "Log study sessions, watch streaks grow, and mark milestones complete in sequence.",
    icon: Target,
  },
  {
    title: "Live study rooms",
    description:
      "Spin up audio rooms inside pods when you want to focus together in real time.",
    icon: Mic,
  },
  {
    title: "Community channels",
    description:
      "Ask questions, share wins, and stay connected with others on the same path.",
    icon: MessageSquare,
  },
  {
    title: "Built for consistency",
    description:
      "Notifications for pod activity keep you in the loop without drowning in noise.",
    icon: Bell,
  },
];

export function Features() {
  return (
    <section id="features" className="border-b border-border/50 py-24 bg-muted/10">
      <div className="mx-auto max-w-6xl px-6 sm:px-8 lg:px-10">
        <Reveal className="mx-auto mb-16 max-w-2xl text-center">
          <h2 className="font-serif text-3xl tracking-tight sm:text-4xl lg:text-5xl">
            Everything you need to stay on track
          </h2>
          <p className="mt-4 text-lg text-foreground/70">
            Not another course catalog — a system for showing up with people who
            care.
          </p>
        </Reveal>
        <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {features.map((f, i) => (
            <Reveal key={f.title} delay={i * 60}>
              <div className="group h-full rounded-2xl border border-border/60 bg-card/45 p-6 shadow-sm backdrop-blur-md transition-all hover:shadow-md hover:border-primary/30 dark:bg-card/40">
                <div className="mb-4 inline-flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 text-primary transition-colors group-hover:bg-primary group-hover:text-primary-foreground">
                  <f.icon className="size-5" />
                </div>
                <h3 className="font-serif text-xl font-medium">{f.title}</h3>
                <p className="mt-2 text-sm leading-relaxed text-foreground/70">
                  {f.description}
                </p>
              </div>
            </Reveal>
          ))}
        </div>
      </div>
    </section>
  );
}

const steps = [
  {
    step: "01",
    title: "Choose a skill",
    text: "Browse skills like web development and join the official community.",
  },
  {
    step: "02",
    title: "Find your pod",
    text: "Join or create a small group aligned with your pace and goals.",
  },
  {
    step: "03",
    title: "Show up daily",
    text: "Log sessions, complete milestones, and keep your streak alive.",
  },
];

export function HowItWorks() {
  return (
    <section id="how-it-works" className="border-b border-border/50 py-24">
      <div className="mx-auto max-w-6xl px-6 sm:px-8 lg:px-10">
        <div className="grid gap-12 lg:grid-cols-2 lg:gap-8 items-center">
          <Reveal className="max-w-xl">
            <h2 className="font-serif text-3xl tracking-tight sm:text-4xl lg:text-5xl">
              How it works
            </h2>
            <p className="mt-4 text-lg text-foreground/70">
              Three steps from curious to consistent. We built Taggy to help you build habits that stick.
            </p>
          </Reveal>
          <ol className="grid gap-6">
            {steps.map((s, i) => (
              <Reveal key={s.step} delay={i * 100}>
                <li className="relative flex gap-6 rounded-2xl border border-border/60 bg-card/45 p-6 shadow-sm backdrop-blur-md transition-all hover:bg-card/60 dark:bg-card/40">
                  <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-full bg-primary/10 font-mono text-lg font-semibold text-primary">
                    {s.step}
                  </div>
                  <div>
                    <h3 className="font-serif text-xl font-medium">{s.title}</h3>
                    <p className="mt-2 text-sm leading-relaxed text-foreground/70">
                      {s.text}
                    </p>
                  </div>
                </li>
              </Reveal>
            ))}
          </ol>
        </div>
      </div>
    </section>
  );
}
