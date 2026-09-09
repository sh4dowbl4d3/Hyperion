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
      <div className="border-2 border-charcoal-ink bg-canary-banner p-6 shadow-offset">
        <p className="flex items-center gap-2.5 text-sm tracking-[0.02em] text-charcoal-ink">
          <span className="size-2 shrink-0 bg-coral-sketch" aria-hidden />
          {error}
        </p>
        <BackLink />
      </div>
    );
  }

  if (!lab) {
    return (
      <div>
        <BackLink />
        <div className="mt-6 h-72 animate-pulse border-2 border-charcoal-ink bg-frost-white" />
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
            <span className="rounded-[2px] border border-charcoal-ink bg-chalk-gray px-2 py-0.5 font-mono text-[11px] uppercase tracking-[0.02em] text-charcoal-ink">
              {lab.category}
            </span>
            <span className="rounded-[2px] border border-graphite px-2 py-0.5 text-[11px] uppercase tracking-[0.02em] text-pencil-gray">
              {lab.difficulty}
            </span>
          </div>
          <h1 className="mt-3 text-[24px] font-light uppercase leading-tight tracking-[0.02em] text-charcoal-ink sm:text-[40px]">
            {lab.name}
          </h1>
        </div>
        <div className="flex items-center gap-3">
          {completed && (
            <span className="inline-flex items-center gap-1.5 border-2 border-charcoal-ink bg-sky-crayon px-3 py-1.5 text-caption font-medium uppercase tracking-[0.02em] text-charcoal-ink">
              <SealCheck size={14} weight="fill" aria-hidden />
              Completed
            </span>
          )}
          <span className="flex items-center gap-1.5 text-sm font-medium text-charcoal-ink">
            <span className="size-1.5 rounded-full bg-duck-bill-orange" aria-hidden />+{lab.xp} XP
          </span>
        </div>
      </div>

      <p className="mt-4 max-w-[65ch] text-body-lg leading-relaxed tracking-[0.02em] text-charcoal-ink/75">
        {lab.description}
      </p>

      {/* Objective — white card with hard offset shadow */}
      <section className="mt-8 border-2 border-charcoal-ink bg-frost-white p-7 shadow-offset sm:p-9">
        <h2 className="flex items-center gap-2.5 text-heading-sm font-medium tracking-[0.02em] text-charcoal-ink">
          <Target size={22} weight="duotone" aria-hidden />
          Objective
        </h2>
        <p className="mt-4 max-w-[70ch] text-body-lg leading-relaxed tracking-[0.02em] text-charcoal-ink/85">
          {lab.objective}
        </p>
      </section>

      <LabPlayground slug={slug} onComplete={reloadLab} />

      {/* Hints */}
      <section className="mt-6 border-2 border-charcoal-ink bg-frost-white p-7 shadow-offset sm:p-9">
        <h2 className="flex items-center gap-2.5 text-heading-sm font-medium tracking-[0.02em] text-charcoal-ink">
          <Lightbulb size={22} weight="duotone" aria-hidden />
          Hints
        </h2>
        <ol className="mt-5 flex flex-col gap-3">
          {lab.hints.slice(0, revealedHints).map((hint, index) => (
            <li
              key={index}
              className="border-l-4 border-sky-crayon bg-chalk-gray px-5 py-3.5 text-body leading-relaxed tracking-[0.02em] text-charcoal-ink/85"
            >
              <span className="mr-2 font-semibold uppercase">Hint {index + 1}</span>
              {hint}
            </li>
          ))}
        </ol>
        {revealedHints < lab.hints.length ? (
          <button
            onClick={() => setRevealedHints((n) => n + 1)}
            className="mt-5 border-2 border-charcoal-ink bg-frost-white px-5 py-2 text-body-sm font-medium uppercase tracking-[0.02em] text-charcoal-ink shadow-offset transition-transform hover:translate-x-[3px] hover:translate-y-[3px] hover:shadow-none"
          >
            Reveal hint {revealedHints + 1} of {lab.hints.length}
          </button>
        ) : (
          revealedHints > 0 && (
            <p className="mt-5 text-sm tracking-[0.02em] text-pencil-gray">
              All hints revealed — you're on your own now.
            </p>
          )
        )}
      </section>

      {completed && lab.vulnerability_type ? (
        <section className="mt-6 border-2 border-charcoal-ink bg-canary-banner p-7 shadow-offset sm:p-9">
          <h2 className="flex items-center gap-2.5 text-heading-sm font-medium tracking-[0.02em] text-charcoal-ink">
            <Skull size={22} weight="duotone" aria-hidden />
            Vulnerability revealed
          </h2>
          <p className="mt-4 inline-block border border-charcoal-ink bg-frost-white px-3 py-1.5 font-mono text-sm tracking-[0.02em] text-charcoal-ink">
            {lab.vulnerability_type}
          </p>
          <p className="mt-4 max-w-[70ch] text-body leading-relaxed tracking-[0.02em] text-charcoal-ink/80">
            You earned this reveal by completing the lab. The secure reference implementation is
            available on every target endpoint via its{" "}
            <code className="font-semibold text-charcoal-ink">-safe</code> twin.
          </p>
        </section>
      ) : (
        !error && (
          <p className="mt-8 text-sm tracking-[0.02em] text-pencil-gray">
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
      className="inline-flex items-center gap-1.5 text-sm font-medium tracking-[0.02em] text-charcoal-ink underline decoration-pencil-gray underline-offset-4 transition-colors hover:decoration-charcoal-ink"
    >
      <ArrowLeft size={15} aria-hidden />
      All labs
    </Link>
  );
}
