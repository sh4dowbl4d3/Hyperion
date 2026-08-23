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

  return (
    <div>
      <h1 className="font-display text-3xl font-semibold tracking-tight text-fog-100">
        Welcome back{user ? `, ${user.email.split("@")[0]}` : ""}
      </h1>
      <p className="mt-2 max-w-[65ch] leading-relaxed text-fog-300">
        Track your training progress here. Open a lab to read its objective and
        start hunting.
      </p>

      <div className="mt-10 grid grid-cols-1 gap-4 sm:grid-cols-3">
        <StatCard label="Labs completed" value={summary ? `${summary.completed_labs} / ${summary.total_labs}` : "—"} />
        <StatCard label="XP earned" value={summary ? `${summary.xp_earned} / ${summary.xp_available}` : "—"} />
        <StatCard label="Current rank" value={rankFor(summary)} />
      </div>

      <section className="mt-12 rounded-xl border border-ink-800 bg-ink-900 p-8">
        <h2 className="font-display text-lg font-semibold tracking-tight text-fog-100">
          Getting started
        </h2>
        <ol className="mt-4 flex flex-col gap-3 text-sm leading-relaxed text-fog-300">
          <li className="flex gap-3">
            <Step n={1} />
            Pick a lab from the <Link to="/labs" className="font-medium text-signal-500 hover:text-signal-400">labs page</Link> and read its objective.
          </li>
          <li className="flex gap-3">
            <Step n={2} />
            Probe the lab's target endpoints — every one has a secure twin for comparison.
          </li>
          <li className="flex gap-3">
            <Step n={3} />
            All targets are local and synthetic. Never point these techniques at systems you do not own.
          </li>
        </ol>
      </section>
    </div>
  );
}

function rankFor(summary: ProgressSummary | null): string {
  if (!summary) return "—";
  if (summary.completed_labs === 0) return "Initiate";
  if (summary.completed_labs < summary.total_labs) return "Operator";
  return "Elite";
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-ink-800 bg-ink-900 px-5 py-4">
      <p className="text-sm text-fog-400">{label}</p>
      <p className="mt-1 font-mono text-2xl font-medium text-fog-100">{value}</p>
    </div>
  );
}

function Step({ n }: { n: number }) {
  return (
    <span className="mt-0.5 flex size-5 shrink-0 items-center justify-center rounded-full border border-signal-500/50 font-mono text-[11px] text-signal-500">
      {n}
    </span>
  );
}
