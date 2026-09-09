import { forwardRef } from "react";
import type { ButtonHTMLAttributes, ReactNode } from "react";
import { CircleNotch } from "@phosphor-icons/react";

type ButtonVariant = "primary" | "ghost" | "coral";

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: ButtonVariant;
  loading?: boolean;
  children: ReactNode;
};

const variantClasses: Record<ButtonVariant, string> = {
  // Primary action — Enamel (#eeeeee) fill with dark text
  primary:
    "bg-[#eeeeee] text-[#111111] border border-[#eeeeee] hover:bg-white hover:border-white active:bg-[#ffa133] active:border-[#ffa133] font-medium",
  // Ghost / Outline — Departure Mono Mud border and text
  ghost:
    "bg-transparent text-[#8a8a6f] border border-[#8a8a6f] hover:bg-[#222222] hover:text-[#eeeeee] hover:border-[#eeeeee] active:bg-[#eeeeee] active:text-[#111111] font-medium",
  // Accent / Amber highlight
  coral:
    "bg-[#ffa133] text-[#111111] border border-[#ffa133] hover:bg-[#e47b1a] hover:border-[#e47b1a] font-medium",
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(function Button(
  { variant = "primary", loading = false, disabled, className = "", children, ...rest },
  ref,
) {
  return (
    <button
      ref={ref}
      disabled={disabled ?? loading}
      className={`inline-flex h-11 items-center justify-center gap-2 rounded-none px-5 text-body-sm uppercase tracking-[0.04em] transition-colors duration-100 disabled:cursor-not-allowed disabled:opacity-50 disabled:border-[#333333] disabled:text-[#666655] disabled:bg-[#161614] ${variantClasses[variant]} ${className}`}
      {...rest}
    >
      {loading && <CircleNotch size={16} className="animate-spin" aria-hidden />}
      {children}
    </button>
  );
});
