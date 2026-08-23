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
      <h1 className="text-[30px] font-medium leading-tight tracking-tight text-ink-black sm:text-[53px] sm:leading-[1.05]">
        Welcome back
        {user ? `, ${user.email.split("@")[0]}` : ""}
      </h1>
      <p className="mt-3 max-w-[60ch] text-body-lg text-stone-gray">
        Track your training progress here. Open a lab to read its objective and
        start hunting.
      </p>

      {/* Progress card + stats */}
      <div className="mt-10 grid grid-cols-1 gap-4 lg:grid-cols-3">
        <div className="rounded-card bg-pure-white p-6 lg:col-span-1">
          <p className="text-sm text-stone-gray">Overall progress</p>
          <div className="mt-4 flex items-end gap-2">
            <span className="text-[53px] font-medium leading-none tracking-[-2px] text-ink-black">
              {summary ? `${pct}` : "—"}
              {summary && "%"}
            </span>
          </div>
          <div className="mt-5 h-2.5 w-full rounded-pill bg-sandstone">
            <div
              className="h-full rounded-pill bg-fresh-grass transition-all"
              style={{ width: `${summary ? Math.max(pct, summary.completed_labs > 0 ? 4 : 0) : 0}%` }}
            />
          </div>
          <p className="mt-3 text-sm text-stone-gray">
            {summary ? `${summary.completed_labs} of ${summary.total_labs} labs completed` : "—"}
          </p>
        </div>

        <StatCard label="XP earned" value={summary ? `${summary.xp_earned} / ${summary.xp_available}` : "—"} />
        <StatCard label="Current rank" value={rankFor(summary)} />
      </div>

      {/* Getting started card */}
      <section className="mt-8 rounded-card bg-pure-white p-8">
        <h2 className="text-[30px] font-medium tracking-tight text-ink-black">Getting started</h2>
        <ol className="mt-6 flex flex-col gap-4 text-[15px] leading-relaxed text-ink-black/80">
          <li className="flex gap-3">
            <Step n={1} />
            <span>
              Pick a lab from the{" "}
              <Link
                to="/labs"
                className="underline decoration-stone-gray underline-offset-4 hover:decoration-ink-black"
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
  if (summary.completed_labs === 0) return "Initiate";
  if (summary.completed_labs < summary.total_labs) return "Operator";
  return "Elite";
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-card bg-pure-white p-6">
      <p className="text-sm text-stone-gray">{label}</p>
      <p className="mt-3 text-[30px] font-medium leading-none tracking-tight text-ink-black">
        {value}
      </p>
    </div>
  );
}

function Step({ n }: { n: number }) {
  const fills = ["bg-fresh-grass", "bg-sky-pop", "bg-sunshine-pop"];
  return (
    <span
      className={`mt-0.5 flex size-6 shrink-0 items-center justify-center rounded-full text-xs font-medium text-ink-black ${fills[n - 1]}`}
    >
      {n}
    </span>
  );
}
