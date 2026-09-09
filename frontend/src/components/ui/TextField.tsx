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
      <label htmlFor={id} className="text-body font-medium text-[#eeeeee]">
        {label}
      </label>
      <input
        id={id}
        aria-invalid={error ? true : undefined}
        aria-describedby={error ? errorId : hint ? hintId : undefined}
        className={`h-11 rounded-none border bg-black px-3 font-mono text-body tracking-[0.02em] text-[#eeeeee] placeholder:text-[#555544] transition-colors focus:border-[#ffa133] ${
          error ? "border-[#ff5555]" : "border-[#333333]"
        } ${className}`}
        {...rest}
      />
      {hint && !error && (
        <p id={hintId} className="text-caption text-[#777766]">
          {hint}
        </p>
      )}
      {error && (
        <p id={errorId} role="alert" className="text-caption font-medium text-[#ff5555]">
          <span className="mr-1 inline-block size-1.5 bg-[#ff5555] align-middle" aria-hidden />
          {error}
        </p>
      )}
    </div>
  );
}
