import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { MagnifyingGlass, SealCheck } from "@phosphor-icons/react";
import { listLabs } from "../services/labService";
import type { Difficulty, LabStatus, LabSummary } from "../types/lab";

const difficultyDot: Record<Difficulty, string> = {
  easy: "bg-fresh-grass",
  medium: "bg-sunshine-pop",
  hard: "bg-coral-pop",
};

function StatusPill({ status }: { status: LabStatus }) {
  if (status === "completed") {
    return (
      <span className="inline-flex items-center gap-1.5 rounded-pill bg-fresh-grass px-3 py-1 text-xs font-medium text-ink-black">
        <SealCheck size={13} weight="fill" aria-hidden />
        Completed
      </span>
    );
  }
  if (status === "in_progress") {
    return (
      <span className="rounded-pill border border-stone-gray px-3 py-1 text-xs text-ink-black/70">
        In progress
      </span>
    );
  }
  return (
    <span className="rounded-pill border border-hairline-mist px-3 py-1 text-xs text-stone-gray">
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
          <h1 className="text-[30px] font-medium leading-tight tracking-tight text-ink-black sm:text-[53px] sm:leading-[1.05]">
            Labs
          </h1>
          <p className="mt-2 text-body-lg text-stone-gray">
            Each lab pairs a vulnerable target with its secure twin.
          </p>
        </div>
        <label className="relative">
          <MagnifyingGlass
            size={16}
            className="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-stone-gray"
            aria-hidden
          />
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search labs…"
            aria-label="Search labs by name or category"
            className="h-11 w-56 rounded-pill border border-hairline-mist bg-pure-white pl-10 pr-4 text-[15px] text-ink-black placeholder:text-stone-gray focus:border-ink-black sm:w-64"
          />
        </label>
      </div>

      {error && (
        <div className="mt-8 flex items-center gap-2.5 rounded-card border border-coral-pop/50 bg-coral-pop/10 px-5 py-3.5 text-sm text-ink-black">
          <span className="size-2 shrink-0 rounded-full bg-coral-pop" aria-hidden />
          {error}
        </div>
      )}

      {!error && labs === null && (
        <div className="mt-8 grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
          {[0, 1, 2].map((i) => (
            <div key={i} className="h-52 animate-pulse rounded-card bg-pure-white" />
          ))}
        </div>
      )}

      {!error && labs !== null && visible.length === 0 && (
        <div className="mt-12 rounded-card border border-dashed border-hairline-mist p-14 text-center">
          <p className="text-[20px] font-medium tracking-tight text-ink-black">
            {labs.length === 0 ? "No labs registered yet" : "No labs match your search"}
          </p>
          <p className="mt-2 text-sm text-stone-gray">
            {labs.length === 0
              ? "The lab catalog populates automatically when modules are enabled on the server."
              : "Try a different name or category."}
          </p>
        </div>
      )}

      <div className="mt-8 grid grid-cols-1 gap-5 md:grid-cols-2 xl:grid-cols-3">
        {visible.map((lab) => (
          <Link
            key={lab.slug}
            to={`/labs/${lab.slug}`}
            className="group flex flex-col rounded-card border-2 border-transparent bg-pure-white p-6 transition-all hover:border-fresh-grass"
          >
            <div className="flex items-start justify-between gap-3">
              {/* Category chip — small radius per system */}
              <span className="rounded-[10px] bg-cream-paper px-2.5 py-1 font-mono text-[11px] uppercase tracking-wider text-stone-gray">
                {lab.category}
              </span>
              <span className="flex items-center gap-1.5 text-sm font-medium text-ink-black">
                <span className="size-2 rounded-full bg-sky-pop" aria-hidden />
                +{lab.xp} XP
              </span>
            </div>

            <h2 className="mt-4 text-[22px] font-medium leading-snug tracking-tight text-ink-black">
              {lab.name}
            </h2>
            <p className="mt-2 line-clamp-3 text-sm leading-relaxed text-stone-gray">
              {lab.description}
            </p>

            <div className="mt-auto flex items-center justify-between pt-6">
              <span className="flex items-center gap-2 text-xs capitalize text-ink-black/70">
                <span className={`size-2 rounded-full ${difficultyDot[lab.difficulty]}`} aria-hidden />
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
