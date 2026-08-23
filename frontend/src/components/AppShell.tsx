import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { ListChecks, SignOut, SquaresFour } from "@phosphor-icons/react";
import { useAuth } from "../hooks/useAuth";

const navItems = [
  { to: "/dashboard", label: "Dashboard", icon: SquaresFour, enabled: true },
  { to: "/labs", label: "Labs", icon: ListChecks, enabled: true },
];

export function AppShell() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  function handleLogout() {
    logout();
    navigate("/login", { replace: true });
  }

  return (
    <div className="flex min-h-[100dvh] flex-col bg-ink-950">
      <header className="flex h-16 shrink-0 items-center justify-between border-b border-ink-800 px-4 sm:px-6">
        <div className="flex items-center gap-2.5">
          <span className="font-display text-lg font-semibold tracking-tight text-fog-100">
            ModernDVWA
          </span>
          <span className="rounded border border-signal-500/40 px-1.5 py-0.5 font-mono text-[10px] font-medium uppercase tracking-widest text-signal-500">
            Labs
          </span>
        </div>
        <div className="flex items-center gap-4">
          <span className="hidden font-mono text-xs text-fog-400 sm:inline" title={user?.email}>
            {user?.email}
          </span>
          <button
            onClick={handleLogout}
            className="inline-flex items-center gap-1.5 rounded-lg px-3 py-2 text-sm text-fog-300 transition-colors hover:bg-ink-850 hover:text-fog-100"
          >
            <SignOut size={16} aria-hidden />
            Sign out
          </button>
        </div>
      </header>

      <div className="mx-auto flex w-full max-w-[1400px] flex-1 gap-8 px-4 py-8 sm:px-6 lg:py-12">
        <nav aria-label="Main navigation" className="w-52 shrink-0">
          <ul className="sticky top-12 flex flex-col gap-1">
            {navItems.map((item) => (
              <li key={item.to}>
                <NavLink
                  to={item.to}
                  className={({ isActive }) =>
                    `flex items-center gap-3 rounded-lg px-3.5 py-2.5 text-sm font-medium transition-colors ${
                      isActive
                        ? "bg-ink-850 text-fog-100"
                        : "text-fog-300 hover:bg-ink-900 hover:text-fog-100"
                    }`
                  }
                >
                  <item.icon size={18} aria-hidden />
                  {item.label}
                </NavLink>
              </li>
            ))}
          </ul>
        </nav>

        <main className="min-w-0 flex-1 pb-16">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
