import { useEffect, useState } from "react";
import { Link } from "@tanstack/react-router";
import { Search, Settings, Menu, Bell } from "lucide-react";
import { Logo } from "./Logo";

const NAV = [
  { label: "Home", to: "/" as const },
  { label: "Movies", to: "/movies" as const },
  { label: "Series", to: "/series" as const },
];

export function Header() {
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 12);
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  return (
    <header
      className={`fixed inset-x-0 top-0 z-50 transition-all duration-300 ${
        scrolled
          ? "border-b border-border bg-background/70 backdrop-blur-xl"
          : "bg-gradient-to-b from-background/80 to-transparent"
      }`}
    >
      <div className="flex h-16 items-center gap-6 px-6 md:h-18 md:px-12 lg:px-20">
        <button
          type="button"
          className="flex h-9 w-9 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-surface hover:text-foreground lg:hidden"
          aria-label="Menu"
        >
          <Menu className="h-5 w-5" />
        </button>

        <Link to="/" className="shrink-0">
          <Logo />
        </Link>

        <nav className="hidden items-center gap-1 lg:flex">
          {NAV.map((item) => (
            <Link
              key={item.label}
              to={item.to}
              activeOptions={{ exact: true }}
              className="relative rounded-full px-4 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
              activeProps={{
                className:
                  "relative rounded-full px-4 py-1.5 text-sm font-medium text-foreground bg-surface",
              }}
            >
              {item.label}
            </Link>
          ))}
        </nav>

        <div className="flex flex-1 items-center justify-end gap-2">
          <div className="relative hidden w-full max-w-md md:block">
            <Search className="pointer-events-none absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <input
              type="search"
              placeholder="Search movies, shows, people…"
              className="h-10 w-full rounded-full border border-border bg-surface/60 pl-11 pr-4 text-sm text-foreground outline-none ring-0 transition-all placeholder:text-muted-foreground/70 focus:border-primary/60 focus:bg-surface focus:shadow-glow"
            />
            <kbd className="pointer-events-none absolute right-3 top-1/2 hidden -translate-y-1/2 rounded border border-border bg-background/60 px-1.5 py-0.5 text-[10px] text-muted-foreground lg:inline-block">
              ⌘K
            </kbd>
          </div>

          <button
            type="button"
            aria-label="Search"
            className="flex h-9 w-9 items-center justify-center rounded-full text-muted-foreground transition-colors hover:bg-surface hover:text-foreground md:hidden"
          >
            <Search className="h-5 w-5" />
          </button>
          <button
            type="button"
            aria-label="Notifications"
            className="relative hidden h-9 w-9 items-center justify-center rounded-full text-muted-foreground transition-colors hover:bg-surface hover:text-foreground sm:flex"
          >
            <Bell className="h-5 w-5" />
            <span className="absolute right-2 top-2 h-1.5 w-1.5 rounded-full bg-accent shadow-glow" />
          </button>
          <button
            type="button"
            aria-label="Settings"
            className="hidden h-9 w-9 items-center justify-center rounded-full text-muted-foreground transition-colors hover:bg-surface hover:text-foreground sm:flex"
          >
            <Settings className="h-5 w-5" />
          </button>
          <button
            type="button"
            className="flex h-9 w-9 items-center justify-center rounded-full bg-gradient-primary text-sm font-semibold text-white shadow-glow ring-1 ring-white/20"
            aria-label="Profile"
          >
            A
          </button>
        </div>
      </div>
    </header>
  );
}
