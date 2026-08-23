import { useState } from "react";
import type { FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import { isApiError } from "../lib/api";
import { useAuth } from "../hooks/useAuth";
import { AuthLayout } from "../layouts/AuthLayout";
import { Button } from "../components/ui/Button";
import { TextField } from "../components/ui/TextField";
import { FormAlert } from "../components/ui/FormAlert";

const friendlyError: Record<string, string> = {
  email_taken: "That email already has an account. Try signing in.",
  invalid_email: "Enter a valid email address.",
  weak_password: "Password must be at least 10 characters.",
  network_error: "Cannot reach the server. Is the backend running?",
};

export default function Register() {
  const { register } = useAuth();
  const navigate = useNavigate();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [fieldErrors, setFieldErrors] = useState<{ email?: string; password?: string; confirm?: string }>({});
  const [formError, setFormError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormError(null);
    setFieldErrors({});

    if (!email.trim()) {
      setFieldErrors((prev) => ({ ...prev, email: "Email is required." }));
      return;
    }
    if (password.length < 10) {
      setFieldErrors((prev) => ({ ...prev, password: "Use at least 10 characters." }));
      return;
    }
    if (password !== confirm) {
      setFieldErrors((prev) => ({ ...prev, confirm: "Passwords do not match." }));
      return;
    }

    setSubmitting(true);
    try {
      await register(email, password);
      navigate("/dashboard", { replace: true });
    } catch (cause) {
      setFormError(
        isApiError(cause)
          ? friendlyError[cause.code] ?? cause.message
          : "Something went wrong. Please try again.",
      );
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <AuthLayout>
      <h2 className="text-heading-sm font-medium tracking-tight text-ink-black">Create account</h2>
      <p className="mt-2 mb-8 text-[15px] text-stone-gray">Start training in under a minute.</p>

      {formError && (
        <div className="mb-6">
          <FormAlert>{formError}</FormAlert>
        </div>
      )}

      <form onSubmit={handleSubmit} noValidate className="flex flex-col gap-5">
        <TextField
          label="Email"
          type="email"
          autoComplete="email"
          placeholder="you@example.com"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          error={fieldErrors.email}
        />
        <TextField
          label="Password"
          type="password"
          autoComplete="new-password"
          placeholder="At least 10 characters"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          error={fieldErrors.password}
        />
        <TextField
          label="Confirm password"
          type="password"
          autoComplete="new-password"
          placeholder="Repeat your password"
          value={confirm}
          onChange={(e) => setConfirm(e.target.value)}
          error={fieldErrors.confirm}
        />
        <Button type="submit" loading={submitting} className="mt-1 w-full">
          Create account
        </Button>
      </form>

      <p className="mt-8 text-sm text-stone-gray">
        Already have an account?{" "}
        <Link
          to="/login"
          className="font-medium text-ink-black underline decoration-stone-gray underline-offset-4 hover:decoration-ink-black"
        >
          Sign in
        </Link>
      </p>
    </AuthLayout>
  );
}
