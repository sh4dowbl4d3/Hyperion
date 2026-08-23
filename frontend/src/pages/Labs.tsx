import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { MagnifyingGlass, SealCheck, Warning } from "@phosphor-icons/react";
import { listLabs } from "../services/labService";
import type { Difficulty, LabStatus, LabSummary } from "../types/lab";

const difficultyDot: Record<Difficulty, string> = {
  easy: "bg-signal-500",
  medium: "bg-caution-500",
  hard: "bg-alert-500",
};

function StatusPill({ status }: { status: LabStatus }) {
  if (status === "completed") {
    return (
      <span className="inline-flex items-center gap-1 rounded-full border border-signal-500/40 bg-signal-500/10 px-2.5 py-1 text-xs font-medium text-signal-500">
        <SealCheck size={13} weight="fill" aria-hidden />
        Completed
      </span>
    );
  }
  if (status === "in_progress") {
    return (
      <span className="rounded-full border border-fog-400/40 px-2.5 py-1 text-xs text-fog-200">
        In progress
      </span>
    );
  }
  return (
    <span className="rounded-full border border-ink-600 px-2.5 py-1 text-xs text-fog-400">
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
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="font-display text-3xl font-semibold tracking-tight text-fog-100">Labs</h1>
          <p className="mt-1 text-fog-300">Each lab pairs a vulnerable target with its secure twin.</p>
        </div>
        <label className="relative">
          <MagnifyingGlass
            size={16}
            className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-fog-400"
            aria-hidden
          />
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search labs…"
            aria-label="Search labs by name or category"
            className="h-10 w-56 rounded-lg border border-ink-600 bg-ink-900 pl-9 pr-3 text-sm text-fog-100 placeholder:text-fog-400/70 focus:border-signal-500 sm:w-64"
          />
        </label>
      </div>

      {error && (
        <div className="mt-8 flex items-center gap-2.5 rounded-lg border border-alert-500/40 bg-alert-500/10 px-4 py-3 text-sm text-fog-100">
          <Warning size={18} weight="fill" className="text-alert-500" aria-hidden />
          {error}
        </div>
      )}

      {!error && labs === null && (
        <div className="mt-8 grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
          {[0, 1, 2].map((i) => (
            <div key={i} className="h-44 animate-pulse rounded-xl bg-ink-850" />
          ))}
        </div>
      )}

      {!error && labs !== null && visible.length === 0 && (
        <div className="mt-12 rounded-xl border border-dashed border-ink-600 p-12 text-center">
          <p className="font-display text-lg font-medium text-fog-200">
            {labs.length === 0 ? "No labs registered yet" : "No labs match your search"}
          </p>
          <p className="mt-2 text-sm text-fog-400">
            {labs.length === 0
              ? "The lab catalog populates automatically when modules are enabled on the server."
              : "Try a different name or category."}
          </p>
        </div>
      )}

      <div className="mt-8 grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
        {visible.map((lab) => (
          <Link
            key={lab.slug}
            to={`/labs/${lab.slug}`}
            className="group flex flex-col rounded-xl border border-ink-800 bg-ink-900 p-6 transition-colors hover:border-signal-500/50"
          >
            <div className="flex items-start justify-between gap-3">
              <span className="font-mono text-xs uppercase tracking-wider text-fog-400">
                {lab.category}
              </span>
              <span className="font-mono text-sm font-medium text-signal-500">+{lab.xp} XP</span>
            </div>

            <h2 className="mt-3 font-display text-lg font-semibold tracking-tight text-fog-100 group-hover:text-white">
              {lab.name}
            </h2>
            <p className="mt-2 line-clamp-3 text-sm leading-relaxed text-fog-300">
              {lab.description}
            </p>

            <div className="mt-auto flex items-center justify-between pt-5">
              <span className="flex items-center gap-2 text-xs capitalize text-fog-300">
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
