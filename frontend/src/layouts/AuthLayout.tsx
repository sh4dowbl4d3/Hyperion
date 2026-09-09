import type { ReactNode } from "react";

export function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <main className="grid min-h-[100dvh] grid-cols-1 lg:grid-cols-2 bg-black text-[#8a8a6f]">
      {/* Editorial side panel on dark terminal canvas */}
      <section className="relative hidden flex-col justify-between overflow-hidden border-r border-[#2a2a26] bg-[#0d0d0b] p-12 lg:flex xl:p-16">
        <div className="relative flex items-center gap-2.5">
          <span
            aria-hidden
            className="flex size-8 items-center justify-center border border-[#ffa133] bg-[#ffa133] text-sm font-bold text-black"
          >
            ▶
          </span>
          <span className="text-body font-semibold tracking-[0.06em] text-[#eeeeee]">
            HYPERION
          </span>
          <span className="border-l border-[#333333] pl-2 text-[11px] text-[#777766]">
            v1.424
          </span>
        </div>

        <div className="relative max-w-xl">
          <div className="mb-4 text-[11px] uppercase tracking-[0.08em] text-[#ffa133]">
            // SECURITY ENVIRONMENT
          </div>
          <h1 className="text-heading-lg font-normal uppercase leading-[1.05] tracking-tight text-[#eeeeee] sm:text-display">
            BREAK MODERN
            <br />
            WEB APPS.
          </h1>
          <p className="mt-6 max-w-[46ch] text-subheading leading-relaxed tracking-[0.02em] text-[#8a8a6f]">
            Hands-on labs for the vulnerabilities that matter in today's APIs,
            authentication flows, and business logic — running entirely on your
            local machine against synthetic data.
          </p>
        </div>

        {/* Departure Mono terminal status markers */}
        <div aria-hidden className="relative flex flex-wrap items-center gap-3 text-caption">
          <span className="inline-block border border-[#333333] bg-[#161614] px-2 py-1 text-[#55ff55]">
            ● LOCAL CONTAINER
          </span>
          <span className="inline-block border border-[#333333] bg-[#161614] px-2 py-1 text-[#8a8a6f]">
            SYNTHETIC DATA
          </span>
          <span className="inline-block border border-[#333333] bg-[#161614] px-2 py-1 text-[#ffa133]">
            WCAG 2.2 AA
          </span>
        </div>
      </section>

      {/* Form column */}
      <section className="flex items-center justify-center bg-black px-5 py-14 sm:px-10">
        <div className="w-full max-w-md border border-[#2a2a26] bg-[#0d0d0b] p-8 sm:p-10">
          {children}
        </div>
      </section>
    </main>
  );
}
