import { WarningCircle } from "@phosphor-icons/react";
import type { ReactNode } from "react";

export function FormAlert({ children }: { children: ReactNode }) {
  return (
    <div
      role="alert"
      className="flex items-start gap-2.5 rounded-none border border-[#333333] border-l-4 border-l-[#ff5555] bg-[#161614] px-4 py-3 text-sm text-[#eeeeee]"
    >
      <WarningCircle
        size={18}
        weight="fill"
        className="mt-0.5 shrink-0 text-[#ff5555]"
        aria-hidden
      />
      <span>{children}</span>
    </div>
  );
}
