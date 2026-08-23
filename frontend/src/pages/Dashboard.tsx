import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { getProgress } from "../services/labService";
import type { ProgressSummary } from "../types/lab";

export default function Dashboard() {
  const { user } = useAuth();
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
    <div>
      <h1 className="text-[24px] font-light uppercase leading-tight tracking-[0.02em] text-charcoal-ink sm:text-display">
        Welcome back
        {user ? `, ${user.email.split("@")[0]}` : ""}
      </h1>
      <p className="mt-3 max-w-[60ch] text-subheading tracking-[0.02em] text-charcoal-ink/75">
        Track your training progress here. Open a lab to read its objective and
        start hunting.
      </p>

      {/* Progress card + stats — white cards with hard offset shadows */}
      <div className="mt-10 grid grid-cols-1 gap-6 lg:grid-cols-3">
        <div className="border-2 border-charcoal-ink bg-frost-white p-6 shadow-offset">
          <p className="text-caption uppercase tracking-[0.02em] text-pencil-gray">
            Overall progress
          </p>
          <div className="mt-4 flex items-end gap-2">
            <span className="text-[40px] font-medium leading-none tracking-[0.02em] text-charcoal-ink">
              {summary ? `${pct}%` : "—"}
            </span>
          </div>
          <div className="mt-5 h-3 w-full border-2 border-charcoal-ink bg-chalk-gray">
            <div
              className="h-full bg-sky-crayon transition-all"
              style={{ width: `${summary ? pct : 0}%` }}
            />
          </div>
          <p className="mt-3 text-sm tracking-[0.02em] text-charcoal-ink/70">
            {summary ? `${summary.completed_labs} of ${summary.total_labs} labs completed` : "—"}
          </p>
        </div>

        <StatCard label="XP EARNED" value={summary ? `${summary.xp_earned} / ${summary.xp_available}` : "—"} />
        <StatCard label="CURRENT RANK" value={rankFor(summary)} />
      </div>

      {/* Getting started card */}
      <section className="mt-8 border-2 border-charcoal-ink bg-frost-white p-8 shadow-offset">
        <h2 className="text-[24px] font-medium tracking-[0.02em] text-charcoal-ink">
          Getting started
        </h2>
        <ol className="mt-6 flex flex-col gap-4 text-body leading-relaxed tracking-[0.02em] text-charcoal-ink/85">
          <li className="flex gap-3">
            <Step n={1} />
            <span>
              Pick a lab from the{" "}
              <Link
                to="/labs"
                className="font-medium underline decoration-sky-crayon decoration-2 underline-offset-4 hover:decoration-charcoal-ink"
              >
                labs page
              </Link>{" "}
              and read its objective.
            </span>
          </li>
          <li className="flex gap-3">
            <Step n={2} />
            <span>Probe the lab's target endpoints — every one has a secure twin for comparison.</span>
          </li>
          <li className="flex gap-3">
            <Step n={3} />
            <span>
              All targets are local and synthetic. Never point these techniques at systems you do
              not own.
            </span>
          </li>
        </ol>
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
    <div className="border-2 border-charcoal-ink bg-frost-white p-6 shadow-offset">
      <p className="text-caption uppercase tracking-[0.02em] text-pencil-gray">{label}</p>
      <p className="mt-4 text-heading-sm font-medium leading-none tracking-[0.02em] text-charcoal-ink">
        {value}
      </p>
    </div>
  );
}

function Step({ n }: { n: number }) {
  const fills = ["bg-mint-sketch", "bg-sky-crayon", "bg-canary-banner"];
  return (
    <span
      className={`mt-0.5 flex size-6 shrink-0 items-center justify-center border-2 border-charcoal-ink text-xs font-semibold text-charcoal-ink ${fills[n - 1]}`}
    >
      {n}
    </span>
  );
}
