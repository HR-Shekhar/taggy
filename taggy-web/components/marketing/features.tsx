import { CheckCircle2 } from "lucide-react";
import { Reveal } from "@/components/marketing/reveal";

const features = [
  {
    title: "Structured skills",
    description:
      "Pick a skill, join its community, and follow a milestone roadmap built for real outcomes.",
  },
  {
    title: "Accountability pods",
    description:
      "Create or join small pods. Request to join, get accepted, and study together.",
  },
  {
    title: "Progress you can see",
    description:
      "Log study sessions, watch streaks grow, and mark milestones complete in sequence.",
  },
  {
    title: "Live study rooms",
    description:
      "Spin up audio rooms inside pods when you want to focus together in real time.",
  },
  {
    title: "Community channels",
    description:
      "Ask questions, share wins, and stay connected with others on the same path.",
  },
  {
    title: "Built for consistency",
    description:
      "Notifications for pod activity keep you in the loop without drowning in noise.",
  },
];

export function Features() {
  return (
    <section id="features" className="border-b border-border/60 py-20">
      <div className="mx-auto max-w-6xl px-4 sm:px-6">
        <Reveal className="mx-auto mb-12 max-w-2xl text-center">
          <h2 className="font-serif text-3xl tracking-tight sm:text-4xl">
            Everything you need to stay on track
          </h2>
          <p className="mt-3 text-muted-foreground">
            Taggy is not another course catalog. It is a system for showing up
            consistently with people who care.
          </p>
        </Reveal>
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {features.map((f, i) => (
            <Reveal key={f.title} delay={i * 80}>
              <div className="rounded-xl border border-border/70 bg-card/55 p-6 shadow-sm backdrop-blur-md">
                <CheckCircle2 className="mb-4 size-5 text-primary" />
                <h3 className="font-medium">{f.title}</h3>
                <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
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
    <section id="how-it-works" className="py-20">
      <div className="mx-auto max-w-6xl px-4 sm:px-6">
        <Reveal className="mb-12">
          <h2 className="font-serif text-3xl tracking-tight sm:text-4xl">
            How it works
          </h2>
          <p className="mt-3 max-w-xl text-muted-foreground">
            Three steps from curious to consistent.
          </p>
        </Reveal>
        <div className="grid gap-6 lg:grid-cols-3">
          {steps.map((s, i) => (
            <Reveal key={s.step} delay={i * 100}>
              <div className="relative rounded-xl bg-secondary/50 p-6">
                <span className="font-mono text-sm text-primary">{s.step}</span>
                <h3 className="mt-3 text-lg font-medium">{s.title}</h3>
                <p className="mt-2 text-sm text-muted-foreground">{s.text}</p>
              </div>
            </Reveal>
          ))}
        </div>
      </div>
    </section>
  );
}
