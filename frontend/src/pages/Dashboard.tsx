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
      <div className="mb-2 text-[11px] uppercase tracking-[0.08em] text-[#ffa133]">
        // TELEMETRY OVERVIEW
      </div>
      <h1 className="text-[28px] font-normal uppercase leading-tight tracking-[0.04em] text-[#eeeeee] sm:text-display">
        Welcome back
        {user ? `, ${user.email.split("@")[0]}` : ""}
      </h1>
      <p className="mt-3 max-w-[60ch] text-subheading tracking-[0.02em] text-[#8a8a6f]">
        Track your training progress here. Open a lab to inspect its objective and
        begin vulnerability testing.
      </p>

      {/* Progress card + stats — dark terminal plates */}
      <div className="mt-8 grid grid-cols-1 gap-6 lg:grid-cols-3">
        <div className="border border-[#2a2a26] bg-[#0d0d0b] p-6">
          <p className="text-caption uppercase tracking-[0.05em] text-[#777766]">
            Overall progress
          </p>
          <div className="mt-4 flex items-end gap-2">
            <span className="text-[40px] font-medium leading-none tracking-[0.02em] text-[#eeeeee]">
              {summary ? `${pct}%` : "—"}
            </span>
          </div>
          <div className="mt-5 h-2.5 w-full border border-[#333333] bg-[#161614]">
            <div
              className="h-full bg-[#ffa133] transition-all"
              style={{ width: `${summary ? pct : 0}%` }}
            />
          </div>
          <p className="mt-3 text-sm tracking-[0.02em] text-[#777766]">
            {summary ? `${summary.completed_labs} of ${summary.total_labs} labs completed` : "—"}
          </p>
        </div>

        <StatCard label="XP EARNED" value={summary ? `${summary.xp_earned} / ${summary.xp_available}` : "—"} />
        <StatCard label="CURRENT RANK" value={rankFor(summary)} />
      </div>

      {/* Getting started card */}
      <section className="mt-8 border border-[#2a2a26] bg-[#0d0d0b] p-8">
        <h2 className="text-[22px] font-medium tracking-[0.04em] text-[#eeeeee]">
          Getting started
        </h2>
        <ol className="mt-6 flex flex-col gap-4 text-body leading-relaxed tracking-[0.02em] text-[#8a8a6f]">
          <li className="flex gap-3">
            <Step n={1} />
            <span>
              Pick a lab from the{" "}
              <Link
                to="/labs"
                className="font-medium text-[#ffa133] underline decoration-[#ffa133] underline-offset-4 hover:text-[#eeeeee]"
              >
                labs catalog
              </Link>{" "}
              and review its target objective.
            </span>
          </li>
          <li className="flex gap-3">
            <Step n={2} />
            <span>Probe the lab's target endpoints — every one has a secure twin for comparative auditing.</span>
          </li>
          <li className="flex gap-3">
            <Step n={3} />
            <span>
              All targets run on synthetic local data. Never direct these techniques at systems you do
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
    <div className="border border-[#2a2a26] bg-[#0d0d0b] p-6">
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
