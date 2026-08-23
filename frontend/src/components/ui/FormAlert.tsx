import { WarningCircle } from "@phosphor-icons/react";
import type { ReactNode } from "react";

export function FormAlert({ children }: { children: ReactNode }) {
  return (
    <div
      role="alert"
      className="flex items-start gap-2.5 rounded-[2px] border-2 border-charcoal-ink bg-canary-banner px-4 py-3 text-sm text-charcoal-ink"
    >
      <WarningCircle
        size={18}
        weight="fill"
        className="mt-0.5 shrink-0 text-charcoal-ink"
        aria-hidden
      />
      <span>{children}</span>
    </div>
  );
}
