import { forwardRef } from "react";
import type { ButtonHTMLAttributes, ReactNode } from "react";
import { CircleNotch } from "@phosphor-icons/react";

type ButtonVariant = "primary" | "ghost";

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: ButtonVariant;
  loading?: boolean;
  children: ReactNode;
};

const variantClasses: Record<ButtonVariant, string> = {
  primary:
    "bg-signal-500 text-ink-950 hover:bg-signal-400 active:translate-y-[1px] font-semibold",
  ghost:
    "bg-transparent text-fog-200 border border-ink-600 hover:border-fog-400 hover:text-fog-100 active:translate-y-[1px]",
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(function Button(
  { variant = "primary", loading = false, disabled, className = "", children, ...rest },
  ref,
) {
  return (
    <button
      ref={ref}
      disabled={disabled ?? loading}
      className={`inline-flex h-11 items-center justify-center gap-2 rounded-lg px-5 text-[15px] transition-colors duration-150 disabled:cursor-not-allowed disabled:opacity-55 ${variantClasses[variant]} ${className}`}
      {...rest}
    >
      {loading && <CircleNotch size={18} className="animate-spin" aria-hidden />}
      {children}
    </button>
  );
});
