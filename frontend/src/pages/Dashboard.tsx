import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { getProgress } from "../services/labService";
import type { ProgressSummary } from "../types/lab";

export default function Dashboard() {
  const [summary, setSummary] = useState<ProgressSummary | null>(null);

  useEffect(() => {
    const controller = new AbortController();
    getProgress()
      .then((res) => setSummary(res.summary))
      .catch(() => {
        if (!controller.signal.aborted) setSummary(null);
      });
    return () => controller.abort();
  }, []);

  const pct =
    summary && summary.total_labs > 0
      ? Math.round((summary.completed_labs / summary.total_labs) * 100)
      : 0;

  return (
    <div className="flex flex-col items-center justify-center text-center max-w-4xl mx-auto">
      {/* Centered Telemetry & Welcome Header */}
      <div className="flex flex-col items-center max-w-2xl">
        <div className="mb-2 text-[11px] uppercase tracking-[0.1em] text-[#ffa133]">
          // SECURITY TRAINING ENVIRONMENT
        </div>
        <h1 className="text-[32px] font-normal uppercase leading-tight tracking-[0.04em] text-[#eeeeee] sm:text-[44px]">
          Welcome to Hyperion
        </h1>
        <p className="mt-3 text-subheading tracking-[0.02em] text-[#8a8a6f]">
          Interactive security training labs. Probe vulnerable endpoints,
          exploit flaws, and compare with secure reference implementations.
        </p>
      </div>

      {/* Progress card + stats — center-aligned grid */}
      <div className="mt-8 grid w-full grid-cols-1 gap-6 sm:grid-cols-3">
        <div className="flex flex-col items-center justify-between border border-[#2a2a26] bg-[#0d0d0b] p-6 text-center">
          <p className="text-caption uppercase tracking-[0.05em] text-[#777766]">
            Overall progress
          </p>
          <div className="my-3">
            <span className="text-[40px] font-medium leading-none tracking-[0.02em] text-[#eeeeee]">
              {summary ? `${pct}%` : "—"}
            </span>
          </div>
          <div className="h-2 w-full border border-[#333333] bg-[#161614]">
            <div
              className="h-full bg-[#ffa133] transition-all"
              style={{ width: `${summary ? pct : 0}%` }}
            />
          </div>
          <p className="mt-3 text-xs tracking-[0.02em] text-[#777766]">
            {summary ? `${summary.completed_labs} of ${summary.total_labs} labs completed` : "—"}
          </p>
        </div>

        <StatCard label="XP EARNED" value={summary ? `${summary.xp_earned} / ${summary.xp_available}` : "—"} />
        <StatCard label="CURRENT RANK" value={rankFor(summary)} />
      </div>

      {/* Getting started card — center-aligned container */}
      <section className="mt-8 w-full border border-[#2a2a26] bg-[#0d0d0b] p-8 sm:p-10 text-center">
        <h2 className="text-[22px] font-medium tracking-[0.04em] text-[#eeeeee]">
          Getting Started
        </h2>
        <div className="mt-6 flex flex-col gap-4 text-body leading-relaxed tracking-[0.02em] text-[#8a8a6f] max-w-xl mx-auto text-left">
          <div className="flex items-start gap-3">
            <Step n={1} />
            <span>
              Pick a lab from the{" "}
              <Link
                to="/labs"
                className="font-medium text-[#ffa133] underline decoration-[#ffa133] underline-offset-4 hover:text-[#eeeeee]"
              >
                labs catalog
              </Link>{" "}
              and inspect the target objectives.
            </span>
          </div>
          <div className="flex items-start gap-3">
            <Step n={2} />
            <span>Probe endpoints in the interactive playground without leaving your browser.</span>
          </div>
          <div className="flex items-start gap-3">
            <Step n={3} />
            <span>
              All targets run on synthetic data in isolated local containers.
            </span>
          </div>
        </div>

        <div className="mt-8 flex justify-center">
          <Link
            to="/labs"
            className="inline-flex h-11 items-center justify-center border border-[#ffa133] bg-[#ffa133] px-6 text-body-sm font-medium uppercase tracking-[0.06em] text-black transition-colors hover:bg-[#e47b1a] hover:border-[#e47b1a]"
          >
            EXPLORE LABS CATALOG →
          </Link>
        </div>
      </section>
    </div>
  );
}

function rankFor(summary: ProgressSummary | null): string {
  if (!summary) return "—";
  if (summary.completed_labs === 0) return "INITIATE";
  if (summary.completed_labs < summary.total_labs) return "OPERATOR";
  return "ELITE";
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex flex-col items-center justify-center border border-[#2a2a26] bg-[#0d0d0b] p-6 text-center">
      <p className="text-caption uppercase tracking-[0.05em] text-[#777766]">{label}</p>
      <p className="mt-4 text-heading-sm font-medium leading-none tracking-[0.02em] text-[#eeeeee]">
        {value}
      </p>
    </div>
  );
}

function Step({ n }: { n: number }) {
  return (
    <span
      className="mt-0.5 flex size-6 shrink-0 items-center justify-center border border-[#ffa133] bg-[#161614] text-xs font-bold text-[#ffa133]"
    >
      {n}
    </span>
  );
}
