import Link from "next/link";
import {
  ArrowRight,
  Flame,
  MessageCircle,
  Route,
  Users,
} from "lucide-react";
import { buttonVariants } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { Reveal } from "@/components/marketing/reveal";

export function Hero() {
  return (
    <section className="relative overflow-hidden border-b border-border/60">
      <div className="relative mx-auto grid max-w-6xl gap-12 px-4 pb-20 pt-20 sm:px-6 lg:grid-cols-2 lg:items-center lg:pb-24 lg:pt-24">
        <Reveal className="space-y-6">
          <Badge variant="secondary" className="rounded-full px-3 py-1">
            Gym buddies for skill growth
          </Badge>
          <h1 className="font-serif text-4xl leading-tight tracking-tight sm:text-5xl lg:text-6xl">
            Learn with structure.
            <span className="block text-primary">Stay accountable together.</span>
          </h1>
          <p className="max-w-xl text-lg text-muted-foreground">
            Taggy combines roadmaps, small pods, community chat, and progress
            tracking so you finish what you start — not alone.
          </p>
          <div className="flex flex-wrap gap-3">
            <Link
              href="/register"
              className={cn(buttonVariants({ size: "lg" }), "gap-2")}
            >
              Start learning free
              <ArrowRight className="size-4" />
            </Link>
            <Link
              href="/login"
              className={cn(buttonVariants({ size: "lg", variant: "outline" }))}
            >
              I have an account
            </Link>
          </div>
        </Reveal>

        <div className="grid gap-4 sm:grid-cols-2">
          <Reveal delay={80}>
            <StatCard
              icon={<Route className="size-5 text-primary" />}
              title="Roadmaps"
              text="Follow milestones in order and track completion."
            />
          </Reveal>
          <Reveal delay={160}>
            <StatCard
              icon={<Users className="size-5 text-primary" />}
              title="Pods"
              text="Small groups for daily accountability."
            />
          </Reveal>
          <Reveal delay={240}>
            <StatCard
              icon={<MessageCircle className="size-5 text-primary" />}
              title="Community"
              text="Skill channels and pod chat in one place."
            />
          </Reveal>
          <Reveal delay={320}>
            <StatCard
              icon={<Flame className="size-5 text-primary" />}
              title="Streaks"
              text="Log study sessions and keep momentum."
            />
          </Reveal>
        </div>
      </div>
    </section>
  );
}

function StatCard({
  icon,
  title,
  text,
}: {
  icon: React.ReactNode;
  title: string;
  text: string;
}) {
  return (
    <div className="rounded-xl border border-border/70 bg-card/55 p-5 shadow-sm backdrop-blur-md">
      <div className="mb-3">{icon}</div>
      <h3 className="font-medium">{title}</h3>
      <p className="mt-1 text-sm text-muted-foreground">{text}</p>
    </div>
  );
}
