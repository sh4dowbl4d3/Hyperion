import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import {
  ArrowLeft,
  Lightbulb,
  SealCheck,
  Skull,
  Target,
  Warning,
} from "@phosphor-icons/react";
import { getLab } from "../services/labService";
import type { LabDetail } from "../types/lab";

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
      <div className="rounded-xl border border-alert-500/40 bg-alert-500/10 p-6">
        <p className="flex items-center gap-2.5 text-sm text-fog-100">
          <Warning size={18} weight="fill" className="text-alert-500" aria-hidden />
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
        <div className="mt-6 h-64 animate-pulse rounded-xl bg-ink-850" />
      </div>
    );
  }

  const completed = lab.status === "completed";

  return (
    <div>
      <BackLink />

      <div className="mt-6 flex flex-wrap items-start justify-between gap-4">
        <div>
          <div className="flex flex-wrap items-center gap-2 font-mono text-xs uppercase tracking-wider text-fog-400">
            <span>{lab.category}</span>
            <span aria-hidden>·</span>
            <span>{lab.difficulty}</span>
          </div>
          <h1 className="mt-2 font-display text-3xl font-semibold tracking-tight text-fog-100">
            {lab.name}
          </h1>
        </div>
        <div className="flex items-center gap-3">
          {completed && (
            <span className="inline-flex items-center gap-1.5 rounded-full border border-signal-500/40 bg-signal-500/10 px-3 py-1.5 text-sm font-medium text-signal-500">
              <SealCheck size={15} weight="fill" aria-hidden />
              Completed
            </span>
          )}
          <span className="font-mono text-sm font-medium text-signal-500">+{lab.xp} XP</span>
        </div>
      </div>

      <p className="mt-4 max-w-[65ch] leading-relaxed text-fog-300">{lab.description}</p>

      <section className="mt-8 rounded-xl border border-ink-800 bg-ink-900 p-6 sm:p-8">
        <h2 className="flex items-center gap-2.5 font-display text-lg font-semibold tracking-tight text-fog-100">
          <Target size={20} className="text-signal-500" aria-hidden />
          Objective
        </h2>
        <p className="mt-3 max-w-[70ch] leading-relaxed text-fog-200">{lab.objective}</p>
      </section>

      <section className="mt-6 rounded-xl border border-ink-800 bg-ink-900 p-6 sm:p-8">
        <h2 className="flex items-center gap-2.5 font-display text-lg font-semibold tracking-tight text-fog-100">
          <Lightbulb size={20} className="text-caution-500" aria-hidden />
          Hints
        </h2>
        <ol className="mt-4 flex flex-col gap-3">
          {lab.hints.slice(0, revealedHints).map((hint, index) => (
            <li
              key={index}
              className="rounded-lg border border-ink-700 bg-ink-850 px-4 py-3 text-sm leading-relaxed text-fog-200"
            >
              <span className="mr-2 font-mono text-xs text-fog-400">Hint {index + 1}</span>
              {hint}
            </li>
          ))}
        </ol>
        {revealedHints < lab.hints.length ? (
          <button
            onClick={() => setRevealedHints((n) => n + 1)}
            className="mt-4 rounded-lg border border-ink-600 px-4 py-2 text-sm font-medium text-fog-200 transition-colors hover:border-fog-400 hover:text-fog-100"
          >
            Reveal hint {revealedHints + 1} of {lab.hints.length}
          </button>
        ) : (
          <p className="mt-4 text-sm text-fog-400">All hints revealed — you're on your own now.</p>
        )}
      </section>

      {completed && lab.vulnerability_type ? (
        <section className="mt-6 rounded-xl border border-ink-800 bg-ink-900 p-6 sm:p-8">
          <h2 className="flex items-center gap-2.5 font-display text-lg font-semibold tracking-tight text-fog-100">
            <Skull size={20} className="text-alert-500" aria-hidden />
            Vulnerability revealed
          </h2>
          <p className="mt-3 font-mono text-sm text-alert-500">{lab.vulnerability_type}</p>
          <p className="mt-3 max-w-[70ch] text-sm leading-relaxed text-fog-300">
            You earned this reveal by completing the lab. The secure reference implementation is
            available on every target endpoint via its{" "}
            <code className="font-mono text-fog-200">-safe</code> twin.
          </p>
        </section>
      ) : (
        !error && (
          <p className="mt-8 text-sm text-fog-400">
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
      className="inline-flex items-center gap-1.5 text-sm font-medium text-fog-300 transition-colors hover:text-fog-100"
    >
      <ArrowLeft size={16} aria-hidden />
      All labs
    </Link>
  );
}
