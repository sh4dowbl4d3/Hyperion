import { useAuth } from "../hooks/useAuth";

export default function Dashboard() {
  const { user } = useAuth();

  return (
    <div>
      <h1 className="font-display text-3xl font-semibold tracking-tight text-fog-100">
        Welcome back{user ? `, ${user.email.split("@")[0]}` : ""}
      </h1>
      <p className="mt-2 max-w-[65ch] leading-relaxed text-fog-300">
        Your training dashboard is being prepared. Vulnerability labs will
        appear here with objectives, hints and XP once the lab engine ships.
      </p>

      <div className="mt-10 grid grid-cols-1 gap-4 sm:grid-cols-3">
        <StatCard label="Labs completed" value="0 / 6" />
        <StatCard label="XP earned" value="0" />
        <StatCard label="Current rank" value="Initiate" />
      </div>

      <section className="mt-12 rounded-xl border border-ink-800 bg-ink-900 p-8">
        <h2 className="font-display text-lg font-semibold tracking-tight text-fog-100">
          Getting started
        </h2>
        <ol className="mt-4 flex flex-col gap-3 text-sm leading-relaxed text-fog-300">
          <li className="flex gap-3">
            <Step n={1} />
            Confirm you can reach the API — your session is active if this page loaded.
          </li>
          <li className="flex gap-3">
            <Step n={2} />
            Lab modules unlock progressively; each pairs a vulnerable endpoint with its secure twin.
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
