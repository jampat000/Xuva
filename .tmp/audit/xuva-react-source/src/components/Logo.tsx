type Props = {
  size?: number;
  withWordmark?: boolean;
  className?: string;
};

export function Logo({ size = 32, withWordmark = true, className = "" }: Props) {
  return (
    <div className={`flex items-center gap-2.5 ${className}`}>
      <svg
        width={size}
        height={size}
        viewBox="0 0 48 48"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        aria-label="Xuva"
      >
        <defs>
          <linearGradient id="logo-grad" x1="0" y1="0" x2="48" y2="48" gradientUnits="userSpaceOnUse">
            <stop offset="0%" stopColor="oklch(0.55 0.22 285)" />
            <stop offset="55%" stopColor="oklch(0.65 0.2 280)" />
            <stop offset="100%" stopColor="oklch(0.72 0.18 265)" />
          </linearGradient>
          <linearGradient id="logo-grad-inner" x1="0" y1="0" x2="48" y2="48" gradientUnits="userSpaceOnUse">
            <stop offset="0%" stopColor="white" stopOpacity="0.95" />
            <stop offset="100%" stopColor="white" stopOpacity="0.55" />
          </linearGradient>
          <filter id="logo-glow" x="-50%" y="-50%" width="200%" height="200%">
            <feGaussianBlur stdDeviation="2" result="b" />
            <feMerge>
              <feMergeNode in="b" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
        </defs>
        <rect x="2" y="2" width="44" height="44" rx="12" fill="url(#logo-grad)" opacity="0.18" />
        <rect x="2" y="2" width="44" height="44" rx="12" stroke="url(#logo-grad)" strokeOpacity="0.5" strokeWidth="1" />
        <g filter="url(#logo-glow)">
          <path
            d="M14 12 L28 24 L14 36"
            stroke="url(#logo-grad)"
            strokeWidth="5"
            strokeLinecap="round"
            strokeLinejoin="round"
            fill="none"
          />
          <path
            d="M22 12 L36 24 L22 36"
            stroke="url(#logo-grad-inner)"
            strokeWidth="5"
            strokeLinecap="round"
            strokeLinejoin="round"
            fill="none"
          />
        </g>
      </svg>
      {withWordmark && (
        <span className="font-display text-xl font-semibold tracking-tight">
          X<span className="text-gradient">uva</span>
        </span>
      )}
    </div>
  );
}
