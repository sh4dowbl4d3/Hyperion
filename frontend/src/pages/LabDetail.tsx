import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import {
  ArrowLeft,
  Lightbulb,
  SealCheck,
  Skull,
  Target,
} from "@phosphor-icons/react";
import { getLab } from "../services/labService";
import type { LabDetail } from "../types/lab";
import { LabPlayground } from "../components/LabPlayground";

export default function LabDetailPage() {
  const { slug = "" } = useParams();
  const [lab, setLab] = useState<LabDetail | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [revealedHints, setRevealedHints] = useState(0);

  useEffect(() => {
    const controller = new AbortController();
    getLab(slug)
      .then(setLab)
      .catch(() => {
        if (!controller.signal.aborted) setError("Could not load this lab.");
      });
    return () => controller.abort();
  }, [slug]);

  function reloadLab() {
    getLab(slug).then(setLab).catch(() => {});
  }

  if (error) {
    return (
      <div className="border border-[#ff5555] border-l-4 border-l-[#ff5555] bg-[#161614] p-6 text-[#eeeeee]">
        <p className="flex items-center gap-2.5 text-sm tracking-[0.02em]">
          <span className="size-2 shrink-0 bg-[#ff5555]" aria-hidden />
          {error}
        </p>
        <div className="mt-4">
          <BackLink />
        </div>
      </div>
    );
  }

  if (!lab) {
    return (
      <div>
        <BackLink />
        <div className="mt-6 h-72 animate-pulse border border-[#2a2a26] bg-[#0d0d0b]" />
      </div>
    );
  }

  const completed = lab.status === "completed";

  return (
    <div>
      <BackLink />

      <div className="mt-6 flex flex-wrap items-start justify-between gap-4">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <span className="rounded-none border border-[#333333] bg-[#161614] px-2 py-0.5 font-mono text-[11px] uppercase tracking-[0.04em] text-[#8a8a6f]">
              {lab.category}
            </span>
            <span className="rounded-none border border-[#333333] bg-[#161614] px-2 py-0.5 text-[11px] uppercase tracking-[0.04em] text-[#777766]">
              {lab.difficulty}
            </span>
          </div>
          <h1 className="mt-3 text-[28px] font-normal uppercase leading-tight tracking-[0.04em] text-[#eeeeee] sm:text-[40px]">
            {lab.name}
          </h1>
        </div>
        <div className="flex items-center gap-3">
          {completed && (
            <span className="inline-flex items-center gap-1.5 border border-[#55ff55] bg-[#161614] px-3 py-1.5 text-caption font-medium uppercase tracking-[0.04em] text-[#55ff55]">
              <SealCheck size={14} weight="fill" aria-hidden />
              Completed
            </span>
          )}
          <span className="flex items-center gap-1.5 text-sm font-medium text-[#ffa133]">
            <span className="size-1.5 bg-[#ffa133]" aria-hidden />+{lab.xp} XP
          </span>
        </div>
      </div>

      <p className="mt-4 max-w-[65ch] text-body-lg leading-relaxed tracking-[0.02em] text-[#8a8a6f]">
        {lab.description}
      </p>

      {/* Objective — terminal plate */}
      <section className="mt-8 border border-[#2a2a26] bg-[#0d0d0b] p-7 sm:p-9">
        <h2 className="flex items-center gap-2.5 text-heading-sm font-medium tracking-[0.04em] text-[#eeeeee]">
          <Target size={22} weight="duotone" className="text-[#ffa133]" aria-hidden />
          Objective
        </h2>
        <p className="mt-4 max-w-[70ch] text-body-lg leading-relaxed tracking-[0.02em] text-[#8a8a6f]">
          {lab.objective}
        </p>
      </section>

      <LabPlayground slug={slug} onComplete={reloadLab} />

      {/* Hints */}
      <section className="mt-6 border border-[#2a2a26] bg-[#0d0d0b] p-7 sm:p-9">
        <h2 className="flex items-center gap-2.5 text-heading-sm font-medium tracking-[0.04em] text-[#eeeeee]">
          <Lightbulb size={22} weight="duotone" className="text-[#ffa133]" aria-hidden />
          Hints
        </h2>
        <ol className="mt-5 flex flex-col gap-3">
          {lab.hints.slice(0, revealedHints).map((hint, index) => (
            <li
              key={index}
              className="border border-[#2a2a26] border-l-4 border-l-[#ffa133] bg-[#161614] px-5 py-3.5 text-body leading-relaxed tracking-[0.02em] text-[#8a8a6f]"
            >
              <span className="mr-2 font-semibold uppercase text-[#eeeeee]">Hint {index + 1}</span>
              {hint}
            </li>
          ))}
        </ol>
        {revealedHints < lab.hints.length ? (
          <button
            onClick={() => setRevealedHints((n) => n + 1)}
            className="mt-5 border border-[#8a8a6f] bg-transparent px-5 py-2 text-body-sm font-medium uppercase tracking-[0.04em] text-[#8a8a6f] transition-colors hover:bg-[#8a8a6f] hover:text-black active:bg-[#ffa133]"
          >
            Reveal hint {revealedHints + 1} of {lab.hints.length}
          </button>
        ) : (
          revealedHints > 0 && (
            <p className="mt-5 text-sm tracking-[0.02em] text-[#777766]">
              All hints revealed — proceed with attack payload.
            </p>
          )
        )}
      </section>

      {completed && lab.vulnerability_type ? (
        <section className="mt-6 border border-[#2a2a26] border-l-4 border-l-[#55ff55] bg-[#0d0d0b] p-7 sm:p-9">
          <h2 className="flex items-center gap-2.5 text-heading-sm font-medium tracking-[0.04em] text-[#55ff55]">
            <Skull size={22} weight="duotone" aria-hidden />
            Vulnerability Revealed
          </h2>
          <p className="mt-4 inline-block border border-[#333333] bg-[#161614] px-3 py-1.5 font-mono text-sm tracking-[0.04em] text-[#eeeeee]">
            {lab.vulnerability_type}
          </p>
          <p className="mt-4 max-w-[70ch] text-body leading-relaxed tracking-[0.02em] text-[#8a8a6f]">
            You earned this reveal by completing the lab. The secure reference implementation is
            available on every target endpoint via its{" "}
            <code className="font-semibold text-[#eeeeee]">-safe</code> twin.
          </p>
        </section>
      ) : (
        !error && (
          <p className="mt-8 text-sm tracking-[0.02em] text-[#777766]">
            The vulnerability class behind this lab is disclosed only after completion.
          </p>
        )
      )}
    </div>
  );
}

function BackLink() {
  return (
    <Link
      to="/labs"
      className="inline-flex items-center gap-1.5 text-sm font-medium tracking-[0.02em] text-[#ffa133] underline decoration-[#ffa133] underline-offset-4 transition-colors hover:text-[#eeeeee]"
    >
      <ArrowLeft size={15} aria-hidden />
      All labs
    </Link>
  );
}
