import { useId } from "react";
import type { InputHTMLAttributes } from "react";

type TextFieldProps = InputHTMLAttributes<HTMLInputElement> & {
  label: string;
  error?: string;
  hint?: string;
};

export function TextField({ label, error, hint, className = "", ...rest }: TextFieldProps) {
  const id = useId();
  const errorId = `${id}-error`;
  const hintId = `${id}-hint`;

  return (
    <div className="flex flex-col gap-2">
      <label htmlFor={id} className="text-sm font-medium text-fog-200">
        {label}
      </label>
      <input
        id={id}
        aria-invalid={error ? true : undefined}
        aria-describedby={error ? errorId : hint ? hintId : undefined}
        className={`h-11 rounded-lg border bg-ink-900 px-3.5 text-[15px] text-fog-100 placeholder:text-fog-400/70 transition-colors duration-150 focus:border-signal-500 ${
          error ? "border-alert-500" : "border-ink-600"
        } ${className}`}
        {...rest}
      />
      {hint && !error && (
        <p id={hintId} className="text-sm text-fog-400">
          {hint}
        </p>
      )}
      {error && (
        <p id={errorId} role="alert" className="text-sm text-alert-500">
          {error}
        </p>
      )}
    </div>
  );
}
