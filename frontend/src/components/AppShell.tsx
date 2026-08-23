import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { ListChecks, SignOut, SquaresFour } from "@phosphor-icons/react";
import { useAuth } from "../hooks/useAuth";

const navItems = [
  { to: "/dashboard", label: "Dashboard", icon: SquaresFour },
  { to: "/labs", label: "Labs", icon: ListChecks },
];

export function AppShell() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  function handleLogout() {
    logout();
    navigate("/login", { replace: true });
  }

  return (
    <div className="min-h-[100dvh] bg-cream-paper">
      {/* Floating pill navigation */}
      <div className="sticky top-4 z-20 mx-auto w-full max-w-[1200px] px-4">
        <header className="flex h-14 items-center justify-between rounded-pill bg-pure-white px-3">
          <div className="flex items-center gap-6">
            {/* Brand logo container */}
            <NavLink to="/dashboard" className="flex items-center gap-2.5 pl-2">
              <span
                aria-hidden
                className="flex size-8 items-center justify-center rounded-[10px] bg-fresh-grass"
              >
                <span className="size-2.5 rounded-full bg-ink-black" />
              </span>
              <span className="text-[17px] font-medium tracking-tight text-ink-black">
                ModernDVWA
              </span>
            </NavLink>

            <nav aria-label="Main navigation" className="hidden sm:block">
              <ul className="flex items-center gap-1">
                {navItems.map((item) => (
                  <li key={item.to}>
                    <NavLink
                      to={item.to}
                      className={({ isActive }) =>
                        `flex items-center gap-2 rounded-pill px-4 py-2 text-[15px] transition-colors ${
                          isActive
                            ? "bg-cream-paper font-medium text-ink-black"
                            : "text-ink-black/80 hover:bg-cream-paper/60 hover:text-ink-black"
                        }`
                      }
                    >
                      <item.icon size={16} aria-hidden />
                      {item.label}
                    </NavLink>
                  </li>
                ))}
              </ul>
            </nav>
          </div>

          <div className="flex items-center gap-2">
            <span
              className="hidden max-w-[180px] truncate text-sm text-stone-gray md:inline"
              title={user?.email}
            >
              {user?.email}
            </span>
            <button
              onClick={handleLogout}
              aria-label="Sign out"
              title="Sign out"
              className="flex size-10 items-center justify-center rounded-full border border-hairline-mist text-ink-black transition-colors hover:border-stone-gray"
            >
              <SignOut size={15} aria-hidden />
            </button>
          </div>
        </header>

        {/* Mobile nav row (below the pill) */}
        <nav aria-label="Mobile navigation" className="mt-2 flex gap-1 rounded-pill bg-pure-white p-1.5 sm:hidden">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
                `flex flex-1 items-center justify-center gap-2 rounded-pill py-2 text-sm transition-colors ${
                  isActive ? "bg-cream-paper font-medium text-ink-black" : "text-ink-black/70"
                }`
              }
            >
              <item.icon size={15} aria-hidden />
              {item.label}
            </NavLink>
          ))}
        </nav>
      </div>

      <main className="mx-auto w-full max-w-[1200px] px-4 pb-24 pt-10 sm:px-6">
        <Outlet />
      </main>

      {/* Sunshine footer accent band */}
      <footer className="bg-sunshine-pop px-6 py-6 text-center">
        <p className="text-sm font-medium text-ink-black">
          LOCAL ENVIRONMENT · SYNTHETIC DATA · NO EXTERNAL TARGETS
        </p>
      </footer>
    </div>
  );
}
