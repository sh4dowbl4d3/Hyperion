import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { MagnifyingGlass, SealCheck } from "@phosphor-icons/react";
import { listLabs } from "../services/labService";
import type { Difficulty, LabStatus, LabSummary } from "../types/lab";



const difficultyDot: Record<Difficulty, string> = {
  easy: "bg-[#55ff55]",
  medium: "bg-[#ffa133]",
  hard: "bg-[#ff5555]",
};

function StatusPill({ status }: { status: LabStatus }) {
  if (status === "completed") {
    return (
      <span className="inline-flex items-center gap-1.5 rounded-none border border-[#55ff55] bg-[#161614] px-2.5 py-1 text-caption font-medium uppercase tracking-[0.04em] text-[#55ff55]">
        <SealCheck size={12} weight="fill" aria-hidden />
        Completed
      </span>
    );
  }
  if (status === "in_progress") {
    return (
      <span className="rounded-none border border-[#ffa133] bg-[#161614] px-2.5 py-1 text-caption uppercase tracking-[0.04em] text-[#ffa133]">
        In progress
      </span>
    );
  }
  return (
    <span className="rounded-none border border-[#333333] bg-[#161614] px-2.5 py-1 text-caption uppercase tracking-[0.04em] text-[#777766]">
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
          <div className="mb-2 text-[11px] uppercase tracking-[0.08em] text-[#ffa133]">
            // VULNERABILITY MODULES
          </div>
          <h1 className="text-[28px] font-normal uppercase leading-tight tracking-[0.04em] text-[#eeeeee] sm:text-[40px]">
            Labs Catalog
          </h1>
          <p className="mt-2 text-subheading tracking-[0.02em] text-[#8a8a6f]">
            Each lab pairs an exploitable target endpoint with its secure twin.
          </p>
        </div>
        <label className="relative">
          <MagnifyingGlass
            size={15}
            className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-[#777766]"
            aria-hidden
          />
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search labs…"
            aria-label="Search labs by name or category"
            className="h-10 w-56 rounded-none border border-[#333333] bg-black pl-9 pr-3 font-mono text-body-sm tracking-[0.02em] text-[#eeeeee] placeholder:text-[#555544] focus:border-[#ffa133] sm:w-64"
          />
        </label>
      </div>

      {error && (
        <div className="mt-8 flex items-center gap-2.5 border border-[#ff5555] border-l-4 border-l-[#ff5555] bg-[#161614] px-4 py-3 text-sm text-[#eeeeee]">
          <span className="size-2 shrink-0 bg-[#ff5555]" aria-hidden />
          {error}
        </div>
      )}

      {!error && labs === null && (
        <div className="mt-8 grid grid-cols-1 gap-6 md:grid-cols-2 xl:grid-cols-3">
          {[0, 1, 2].map((i) => (
            <div key={i} className="h-52 animate-pulse border border-[#2a2a26] bg-[#0d0d0b]" />
          ))}
        </div>
      )}

      {!error && labs !== null && visible.length === 0 && (
        <div className="mt-12 border border-dashed border-[#333333] bg-[#0d0d0b] p-14 text-center">
          <p className="text-heading-sm font-medium tracking-[0.04em] text-[#eeeeee]">
            {labs.length === 0 ? "No labs registered yet" : "No labs match your search"}
          </p>
          <p className="mt-2 text-body tracking-[0.02em] text-[#777766]">
            {labs.length === 0
              ? "The lab catalog populates automatically when modules are enabled on the server."
              : "Try a different search query or category filter."}
          </p>
        </div>
      )}

      {/* Terminal lab cards grid */}
      <div className="mt-8 grid grid-cols-1 gap-6 md:grid-cols-2 xl:grid-cols-3">
        {visible.map((lab) => (
          <Link
            key={lab.slug}
            to={`/labs/${lab.slug}`}
            className="group flex flex-col border border-[#2a2a26] bg-[#0d0d0b] p-6 transition-colors hover:border-[#8a8a6f]"
          >
            <div className="flex items-start justify-between gap-3">
              <span className="rounded-none border border-[#333333] bg-[#161614] px-2 py-0.5 font-mono text-[11px] uppercase tracking-[0.04em] text-[#8a8a6f]">
                {lab.category}
              </span>
              <span className="flex items-center gap-1.5 text-sm font-medium text-[#ffa133]">
                <span className="size-1.5 bg-[#ffa133]" aria-hidden />+{lab.xp} XP
              </span>
            </div>

            <h2 className="mt-4 text-heading-sm leading-snug tracking-[0.04em] text-[#eeeeee] group-hover:text-white">
              {lab.name}
            </h2>
            <p className="mt-2 line-clamp-3 text-body leading-relaxed tracking-[0.02em] text-[#8a8a6f]">
              {lab.description}
            </p>

            <div className="mt-auto flex items-center justify-between pt-6 border-t border-[#1a1a16]">
              <span className="flex items-center gap-2 text-xs uppercase tracking-[0.04em] text-[#777766]">
                <span className={`size-2 ${difficultyDot[lab.difficulty]}`} aria-hidden />
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
