import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { MagnifyingGlass, SealCheck } from "@phosphor-icons/react";
import { listLabs } from "../services/labService";
import type { Difficulty, LabStatus, LabSummary } from "../types/lab";

/* Rainbow outline palette — distributed across lab tiles like pencils in a cup */
const rainbowBorders = [
  "border-coral-sketch",
  "border-peach-sketch",
  "border-mint-sketch",
  "border-lilac-sketch",
  "border-lime-sketch",
  "border-periwinkle-sketch",
  "border-slate-sketch",
  "border-marigold-sketch",
];

const difficultyDot: Record<Difficulty, string> = {
  easy: "bg-mint-sketch",
  medium: "bg-canary-banner",
  hard: "bg-coral-sketch",
};

function StatusPill({ status }: { status: LabStatus }) {
  if (status === "completed") {
    return (
      <span className="inline-flex items-center gap-1.5 rounded-[2px] border border-charcoal-ink bg-sky-crayon px-2.5 py-1 text-caption font-medium uppercase tracking-[0.02em] text-charcoal-ink">
        <SealCheck size={12} weight="fill" aria-hidden />
        Completed
      </span>
    );
  }
  if (status === "in_progress") {
    return (
      <span className="rounded-[2px] border border-pencil-gray px-2.5 py-1 text-caption uppercase tracking-[0.02em] text-charcoal-ink/70">
        In progress
      </span>
    );
  }
  return (
    <span className="rounded-[2px] border border-graphite px-2.5 py-1 text-caption uppercase tracking-[0.02em] text-pencil-gray">
      Not started
    </span>
  );
}

export default function Labs() {
  const [labs, setLabs] = useState<LabSummary[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [query, setQuery] = useState("");

  useEffect(() => {
    const controller = new AbortController();
    listLabs()
      .then(setLabs)
      .catch(() => {
        if (!controller.signal.aborted) setError("Could not load labs. Is the backend running?");
      });
    return () => controller.abort();
  }, []);

  const visible = (labs ?? []).filter(
    (lab) =>
      lab.name.toLowerCase().includes(query.toLowerCase()) ||
      lab.category.toLowerCase().includes(query.toLowerCase()),
  );

  return (
    <div>
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-[24px] font-light uppercase leading-tight tracking-[0.02em] text-charcoal-ink sm:text-[40px]">
            Labs
          </h1>
          <p className="mt-2 text-subheading tracking-[0.02em] text-charcoal-ink/75">
            Each lab pairs a vulnerable target with its secure twin.
          </p>
        </div>
        <label className="relative">
          <MagnifyingGlass
            size={15}
            className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-pencil-gray"
            aria-hidden
          />
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search labs…"
            aria-label="Search labs by name or category"
            className="h-10 w-56 rounded-[2px] border-2 border-charcoal-ink bg-frost-white pl-9 pr-3 text-body-sm tracking-[0.02em] text-charcoal-ink placeholder:text-pencil-gray focus:border-sky-crayon sm:w-64"
          />
        </label>
      </div>

      {error && (
        <div className="mt-8 flex items-center gap-2.5 border-2 border-charcoal-ink bg-canary-banner px-4 py-3 text-sm tracking-[0.02em] text-charcoal-ink">
          <span className="size-2 shrink-0 bg-coral-sketch" aria-hidden />
          {error}
        </div>
      )}

      {!error && labs === null && (
        <div className="mt-8 grid grid-cols-1 gap-6 md:grid-cols-2 xl:grid-cols-3">
          {[0, 1, 2].map((i) => (
            <div key={i} className="h-52 animate-pulse border-2 border-charcoal-ink bg-frost-white" />
          ))}
        </div>
      )}

      {!error && labs !== null && visible.length === 0 && (
        <div className="mt-12 border-2 border-dashed border-graphite p-14 text-center">
          <p className="text-heading-sm font-medium tracking-[0.02em] text-charcoal-ink">
            {labs.length === 0 ? "No labs registered yet" : "No labs match your search"}
          </p>
          <p className="mt-2 text-body tracking-[0.02em] text-pencil-gray">
            {labs.length === 0
              ? "The lab catalog populates automatically when modules are enabled on the server."
              : "Try a different name or category."}
          </p>
        </div>
      )}

      {/* Rainbow-bordered card grid */}
      <div className="mt-8 grid grid-cols-1 gap-7 md:grid-cols-2 xl:grid-cols-3">
        {visible.map((lab, i) => (
          <Link
            key={lab.slug}
            to={`/labs/${lab.slug}`}
            className={`group flex flex-col border-2 ${rainbowBorders[i % rainbowBorders.length]} bg-frost-white p-6 shadow-offset transition-transform hover:-translate-y-1`}
          >
            <div className="flex items-start justify-between gap-3">
              <span className="rounded-[2px] border border-charcoal-ink bg-chalk-gray px-2 py-0.5 font-mono text-[11px] uppercase tracking-[0.02em] text-charcoal-ink">
                {lab.category}
              </span>
              <span className="flex items-center gap-1.5 text-sm font-medium text-charcoal-ink">
                <span className="size-1.5 rounded-full bg-duck-bill-orange" aria-hidden />+{lab.xp} XP
              </span>
            </div>

            <h2 className="mt-4 text-heading-sm leading-snug tracking-[0.02em] text-charcoal-ink">
              {lab.name}
            </h2>
            <p className="mt-2 line-clamp-3 text-body leading-relaxed tracking-[0.02em] text-charcoal-ink/70">
              {lab.description}
            </p>

            <div className="mt-auto flex items-center justify-between pt-6">
              <span className="flex items-center gap-2 text-xs uppercase tracking-[0.02em] text-charcoal-ink/70">
                <span className={`size-2 ${difficultyDot[lab.difficulty]} border border-charcoal-ink`} aria-hidden />
                {lab.difficulty}
              </span>
              <StatusPill status={lab.status} />
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}
