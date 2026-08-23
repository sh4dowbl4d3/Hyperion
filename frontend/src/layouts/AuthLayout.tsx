import type { ReactNode } from "react";

export function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <main className="grid min-h-[100dvh] grid-cols-1 lg:grid-cols-2">
      {/* Editorial side panel on cream canvas */}
      <section className="relative hidden flex-col justify-between overflow-hidden border-r-2 border-charcoal-ink bg-cream-paper p-12 lg:flex xl:p-16">
        <div className="relative flex items-center gap-2.5">
          <span
            aria-hidden
            className="flex size-8 items-center justify-center border-2 border-charcoal-ink bg-duck-bill-orange text-sm font-semibold text-charcoal-ink"
          >
            M
          </span>
          <span className="text-body font-semibold tracking-[0.02em] text-charcoal-ink">
            ModernDVWA
          </span>
        </div>

        <div className="relative max-w-xl">
          <h1 className="text-heading-lg font-light uppercase leading-[1.05] tracking-[0.02em] text-charcoal-ink sm:text-display">
            Break modern
            <br />
            web apps.
          </h1>
          <p className="mt-6 max-w-[46ch] text-subheading leading-relaxed tracking-[0.02em] text-charcoal-ink/80">
            Hands-on labs for the vulnerabilities that matter in today's APIs,
            authentication flows and business logic — running entirely on your
            machine against synthetic data.
          </p>
        </div>

        {/* Crayon sketch decorations */}
        <div aria-hidden className="relative flex items-center gap-3">
          <span className="size-7 border-2 border-charcoal-ink bg-sky-crayon" />
          <span className="size-7 border-2 border-charcoal-ink bg-canary-banner" />
          <span className="size-7 border-2 border-charcoal-ink bg-coral-sketch" />
          <span className="size-7 border-2 border-charcoal-ink bg-mint-sketch" />
        </div>
      </section>

      {/* Form column */}
      <section className="flex items-center justify-center bg-frost-white px-5 py-14 sm:px-10">
        <div className="w-full max-w-md">{children}</div>
      </section>
    </main>
  );
}
