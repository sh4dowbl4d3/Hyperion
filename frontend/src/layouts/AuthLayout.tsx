import type { ReactNode } from "react";

export function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <main className="grid min-h-[100dvh] grid-cols-1 lg:grid-cols-2">
      <section className="relative hidden flex-col justify-between overflow-hidden border-r border-ink-800 bg-ink-900 p-12 lg:flex xl:p-16">
        <div
          aria-hidden
          className="pointer-events-none absolute inset-0 opacity-40"
          style={{
            backgroundImage:
              "linear-gradient(rgba(23,184,119,0.06) 1px, transparent 1px), linear-gradient(90deg, rgba(23,184,119,0.06) 1px, transparent 1px)",
            backgroundSize: "44px 44px",
          }}
        />
        <div className="relative flex items-center gap-3">
          <span className="font-display text-xl font-semibold tracking-tight text-fog-100">
            ModernDVWA
          </span>
        </div>

        <div className="relative max-w-md">
          <h1 className="font-display text-[2.6rem] font-semibold leading-[1.08] tracking-tight text-fog-100">
            Break modern web apps.
            <br />
            <span className="italic">Learn</span> how to defend them.
          </h1>
          <p className="mt-5 max-w-[46ch] leading-relaxed text-fog-300">
            Hands-on labs for the vulnerabilities that matter in today's APIs,
            authentication flows and business logic — running entirely on your
            machine against synthetic data.
          </p>
        </div>

        <p className="relative font-mono text-xs tracking-wide text-fog-400">
          LOCAL ENVIRONMENT · SYNTHETIC DATA · NO EXTERNAL TARGETS
        </p>
      </section>

      <section className="flex items-center justify-center px-5 py-14 sm:px-10">
        <div className="w-full max-w-sm">{children}</div>
      </section>
    </main>
  );
}
