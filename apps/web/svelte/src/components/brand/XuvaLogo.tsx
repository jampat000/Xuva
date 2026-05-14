import { useId } from "react";

interface XuvaLogoProps {
  size?: number;
  className?: string;
  title?: string;
}

const STOPS = ["#A78BFA", "#7C3AED", "#DB2777"] as const;

/**
 * Xuva mark — "Slipstream"
 *
 * A play triangle whose back edge splits into two diverging tails forming an X
 * on the left; the right collapses into the universal play apex.
 */
export function XuvaLogo({ size = 40, className, title = "Xuva" }: XuvaLogoProps) {
  const uid = useId().replace(/:/g, "");
  const gMain = `xuva-main-${uid}`;
  const gShadow = `xuva-shadow-${uid}`;

  const TOP = "M 4 4 L 16 4 L 60 32 L 30 32 Z";
  const BOT = "M 4 60 L 16 60 L 60 32 L 30 32 Z";

  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 64 64"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      role="img"
      aria-label={title}
    >
      <title>{title}</title>
      <defs>
        <linearGradient id={gMain} x1="4" y1="32" x2="60" y2="32" gradientUnits="userSpaceOnUse">
          <stop offset="0%" stopColor={STOPS[0]} />
          <stop offset="55%" stopColor={STOPS[1]} />
          <stop offset="100%" stopColor={STOPS[2]} />
        </linearGradient>
        <linearGradient id={gShadow} x1="0" y1="0" x2="0" y2="64" gradientUnits="userSpaceOnUse">
          <stop offset="0%" stopColor="#000000" stopOpacity="0.35" />
          <stop offset="100%" stopColor="#000000" stopOpacity="0" />
        </linearGradient>
      </defs>

      <g transform="translate(0.6, 1.2)" opacity="0.5">
        <path d={TOP} fill="#000" />
        <path d={BOT} fill="#000" />
      </g>

      <path d={TOP} fill={`url(#${gMain})`} />
      <path d={BOT} fill={`url(#${gMain})`} />

      <path
        d="M 6 4.6 L 14 4.6 L 58.5 31 L 56 32 Z"
        fill="#FFFFFF"
        opacity="0.32"
      />
      <path d="M 30 32 L 33 30.2 L 33 33.8 Z" fill="#FFFFFF" opacity="0.55" />
      <path d={BOT} fill={`url(#${gShadow})`} />
    </svg>
  );
}

export function XuvaAppIcon({
  size = 96,
  className,
}: Omit<XuvaLogoProps, "title">) {
  const uid = useId().replace(/:/g, "");
  const bgId = `xuva-tile-${uid}`;
  const radius = size * 0.235;

  return (
    <div
      className={className}
      style={{
        width: size,
        height: size,
        borderRadius: radius,
        position: "relative",
        overflow: "hidden",
        boxShadow:
          "inset 0 1px 0 rgba(255,255,255,0.10), inset 0 -1px 0 rgba(0,0,0,0.4), 0 18px 50px -16px rgba(0,0,0,0.75)",
      }}
    >
      <svg
        width={size}
        height={size}
        viewBox="0 0 100 100"
        style={{ position: "absolute", inset: 0 }}
      >
        <defs>
          <radialGradient id={bgId} cx="32%" cy="38%" r="80%">
            <stop offset="0%" stopColor={STOPS[2]} stopOpacity="0.55" />
            <stop offset="60%" stopColor="#0a0a0d" stopOpacity="1" />
            <stop offset="100%" stopColor="#050507" stopOpacity="1" />
          </radialGradient>
        </defs>
        <rect width="100" height="100" fill="#0a0a0d" />
        <rect width="100" height="100" fill={`url(#${bgId})`} />
      </svg>
      <div
        style={{
          position: "absolute",
          inset: 0,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
        }}
      >
        <XuvaLogo size={size * 0.6} />
      </div>
    </div>
  );
}
