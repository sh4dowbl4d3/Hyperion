import { NavLink, Outlet, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";

const navItems = [
  { to: "/dashboard", label: "DASHBOARD" },
  { to: "/labs", label: "LABS" },
];

export function AppShell() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const isHomepage = location.pathname === "/" || location.pathname === "/dashboard";

  function handleLogout() {
    logout();
    navigate("/login", { replace: true });
  }

  return (
    <div className="flex min-h-[100dvh] flex-col bg-cream-paper">
      {/* Flat top navigation bar */}
      <header className="border-b border-charcoal-ink bg-frost-white">
        <div className="mx-auto flex h-14 w-full max-w-[1200px] items-center justify-between px-4 sm:px-6">
          {/* Logo */}
          <NavLink to="/dashboard" className="flex items-center gap-2.5">
            <span
              aria-hidden
              className="flex size-7 items-center justify-center border-2 border-charcoal-ink bg-duck-bill-orange text-xs font-semibold"
            >
              M
            </span>
            <span className="text-body font-semibold tracking-[0.02em] text-charcoal-ink">
              Hyperion
            </span>
          </NavLink>

          {/* Center links (desktop) */}
          <nav
            aria-label="Main navigation"
            className="absolute left-1/2 hidden -translate-x-1/2 sm:block"
          >
            <ul className="flex items-center gap-1">
              {navItems.map((item) => (
                <li key={item.to}>
                  <NavLink
                    to={item.to}
                    className={({ isActive }) =>
                      `px-3 py-1.5 text-body-sm transition-colors ${
                        isActive
                          ? "bg-ice-wash font-medium text-charcoal-ink"
                          : "text-charcoal-ink hover:bg-chalk-gray"
                      }`
                    }
                  >
                    {item.label}
                  </NavLink>
                </li>
              ))}
            </ul>
          </nav>

          <div className="flex items-center gap-3">
            <span
              className="hidden max-w-[180px] truncate text-caption text-pencil-gray md:inline"
              title={user?.email}
            >
              {user?.email}
            </span>
            <button
              onClick={handleLogout}
              className="border-2 border-charcoal-ink bg-sky-crayon px-3 py-1.5 text-caption font-medium uppercase tracking-[0.02em] text-charcoal-ink transition-transform hover:translate-x-[2px] hover:translate-y-[2px]"
            >
              LOG OUT
            </button>
          </div>
        </div>
        {/* Mobile nav row */}
        <nav aria-label="Mobile navigation" className="border-t border-graphite sm:hidden">
          <ul className="mx-auto flex w-full max-w-[1200px]">
            {navItems.map((item) => (
              <li key={item.to} className="flex-1">
                <NavLink
                  to={item.to}
                  className={({ isActive }) =>
                    `block px-4 py-2.5 text-center text-body-sm ${
                      isActive ? "bg-ice-wash font-medium" : "text-charcoal-ink"
                    }`
                  }
                >
                  {item.label}
                </NavLink>
              </li>
            ))}
          </ul>
        </nav>
      </header>

      <main className="mx-auto w-full max-w-[1200px] flex-1 px-4 pb-20 pt-10 sm:px-6">
        <Outlet />
      </main>

      {/* Site footer — hidden on homepage */}
      {!isHomepage && (
        <footer className="border-t-2 border-charcoal-ink bg-frost-white">
          <div className="mx-auto flex w-full max-w-[1200px] flex-col gap-6 px-4 py-8 sm:px-6 md:flex-row md:items-center md:justify-between">
            <div className="flex flex-col gap-2">
              <div className="flex items-center gap-2.5">
                <span
                  aria-hidden
                  className="flex size-6 items-center justify-center border-2 border-charcoal-ink bg-duck-bill-orange text-[11px] font-semibold text-charcoal-ink"
                >
                  M
                </span>
                <span className="text-body font-semibold tracking-[0.02em] text-charcoal-ink">
                  Hyperion
                </span>
                <span className="border border-charcoal-ink bg-chalk-gray px-1.5 py-0.5 font-mono text-[10px] text-pencil-gray">
                  v0.1.0
                </span>
              </div>
              <p className="max-w-md text-caption leading-relaxed tracking-[0.02em] text-pencil-gray">
                Intentionally vulnerable security training runtime. Hands-on labs for modern web application vulnerabilities.
              </p>
            </div>

            <div className="flex flex-wrap items-center gap-6 sm:gap-8">
              <nav aria-label="Footer navigation">
                <ul className="flex items-center gap-4 text-body-sm font-medium uppercase tracking-[0.02em]">
                  <li>
                    <NavLink
                      to="/dashboard"
                      className="text-charcoal-ink underline decoration-graphite underline-offset-4 transition-colors hover:decoration-charcoal-ink"
                    >
                      Dashboard
                    </NavLink>
                  </li>
                  <li>
                    <NavLink
                      to="/labs"
                      className="text-charcoal-ink underline decoration-graphite underline-offset-4 transition-colors hover:decoration-charcoal-ink"
                    >
                      Labs
                    </NavLink>
                  </li>
                </ul>
              </nav>

              {/* Crayon sketch decoration swatches */}
              <div aria-hidden className="flex items-center gap-1.5">
                <span className="size-3.5 border border-charcoal-ink bg-sky-crayon" />
                <span className="size-3.5 border border-charcoal-ink bg-canary-banner" />
                <span className="size-3.5 border border-charcoal-ink bg-coral-sketch" />
                <span className="size-3.5 border border-charcoal-ink bg-mint-sketch" />
              </div>
            </div>
          </div>

          <div className="border-t border-graphite/30 bg-chalk-gray/50 py-3">
            <div className="mx-auto flex w-full max-w-[1200px] flex-col items-center justify-between gap-2 px-4 text-caption text-pencil-gray sm:flex-row sm:px-6">
              <p>© 2026 Hyperion Security. For educational and defensive research only.</p>
              <p className="font-mono text-[10px] tracking-wider text-charcoal-ink/60">
                LOCAL SANDBOX · SYNTHETIC TARGETS
              </p>
            </div>
          </div>
        </footer>
      )}
    </div>
  );
}
