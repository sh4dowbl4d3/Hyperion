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

  if (error) {
    return (
      <div className="rounded-card border border-coral-pop/50 bg-coral-pop/10 p-6">
        <p className="flex items-center gap-2.5 text-sm text-ink-black">
          <span className="size-2 shrink-0 rounded-full bg-coral-pop" aria-hidden />
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
        <div className="mt-6 h-72 animate-pulse rounded-card bg-pure-white" />
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
            <span className="rounded-[10px] bg-cream-paper px-2.5 py-1 font-mono text-[11px] uppercase tracking-wider text-stone-gray">
              {lab.category}
            </span>
            <span className="rounded-[10px] bg-cream-paper px-2.5 py-1 text-[11px] capitalize text-stone-gray">
              {lab.difficulty}
            </span>
          </div>
          <h1 className="mt-3 text-[30px] font-medium leading-tight tracking-tight text-ink-black sm:text-[53px] sm:leading-[1.05]">
            {lab.name}
          </h1>
        </div>
        <div className="flex items-center gap-3">
          {completed && (
            <span className="inline-flex items-center gap-1.5 rounded-pill bg-fresh-grass px-4 py-2 text-sm font-medium text-ink-black">
              <SealCheck size={15} weight="fill" aria-hidden />
              Completed
            </span>
          )}
          <span className="flex items-center gap-1.5 text-sm font-medium text-ink-black">
            <span className="size-2 rounded-full bg-sky-pop" aria-hidden />+{lab.xp} XP
          </span>
        </div>
      </div>

      <p className="mt-4 max-w-[65ch] text-body-lg leading-relaxed text-stone-gray">
        {lab.description}
      </p>

      {/* Objective — white card on cream */}
      <section className="mt-8 rounded-card bg-pure-white p-7 sm:p-9">
        <h2 className="flex items-center gap-2.5 text-[22px] font-medium tracking-tight text-ink-black">
          <Target size={22} weight="duotone" className="text-fresh-grass" aria-hidden />
          Objective
        </h2>
        <p className="mt-4 max-w-[70ch] text-body-lg leading-relaxed text-ink-black/85">
          {lab.objective}
        </p>
      </section>

      <LabPlayground slug={slug} />

      {/* Hints */}
      <section className="mt-6 rounded-card bg-pure-white p-7 sm:p-9">
        <h2 className="flex items-center gap-2.5 text-[22px] font-medium tracking-tight text-ink-black">
          <Lightbulb size={22} weight="duotone" className="text-sunshine-pop" aria-hidden />
          Hints
        </h2>
        <ol className="mt-5 flex flex-col gap-3">
          {lab.hints.slice(0, revealedHints).map((hint, index) => (
            <li
              key={index}
              className="rounded-card bg-cream-paper px-5 py-4 text-[15px] leading-relaxed text-ink-black/85"
            >
              <span className="mr-2 font-mono text-xs text-stone-gray">Hint {index + 1}</span>
              {hint}
            </li>
          ))}
        </ol>
        {revealedHints < lab.hints.length ? (
          <button
            onClick={() => setRevealedHints((n) => n + 1)}
            className="mt-5 rounded-pill border border-hairline-mist bg-pure-white px-5 py-2.5 text-sm font-medium text-ink-black transition-colors hover:border-stone-gray"
          >
            Reveal hint {revealedHints + 1} of {lab.hints.length}
          </button>
        ) : (
          revealedHints > 0 && (
            <p className="mt-5 text-sm text-stone-gray">
              All hints revealed — you're on your own now.
            </p>
          )
        )}
      </section>

      {completed && lab.vulnerability_type ? (
        <section className="mt-6 rounded-card border-2 border-fresh-grass bg-pure-white p-7 sm:p-9">
          <h2 className="flex items-center gap-2.5 text-[22px] font-medium tracking-tight text-ink-black">
            <Skull size={22} weight="duotone" className="text-coral-pop" aria-hidden />
            Vulnerability revealed
          </h2>
          <p className="mt-4 inline-block rounded-[10px] bg-cream-paper px-3 py-1.5 font-mono text-sm text-ink-black">
            {lab.vulnerability_type}
          </p>
          <p className="mt-4 max-w-[70ch] text-[15px] leading-relaxed text-stone-gray">
            You earned this reveal by completing the lab. The secure reference implementation is
            available on every target endpoint via its{" "}
            <code className="font-mono text-ink-black">-safe</code> twin.
          </p>
        </section>
      ) : (
        !error && (
          <p className="mt-8 text-sm text-stone-gray">
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
      className="inline-flex items-center gap-1.5 text-sm font-medium text-stone-gray underline decoration-hairline-mist underline-offset-4 transition-colors hover:text-ink-black hover:decoration-stone-gray"
    >
      <ArrowLeft size={15} aria-hidden />
      All labs
    </Link>
  );
}
