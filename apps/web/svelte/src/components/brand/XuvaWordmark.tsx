import { useId } from "react";
import { XuvaLogo } from "./XuvaLogo";

interface XuvaWordmarkProps {
  size?: number;
  className?: string;
  showMark?: boolean;
}

const STOPS = ["#A78BFA", "#7C3AED", "#DB2777"] as const;

export function XuvaWordmark({ size = 36, className, showMark = true }: XuvaWordmarkProps) {
  const uid = useId().replace(/:/g, "");
  const gradId = `xuva-word-${uid}`;

  const h = size;
  const xW = h * 0.72;
  const uvaSize = h * 0.78;

  return (
    <div
      className={className}
      style={{ display: "inline-flex", alignItems: "center", gap: h * 0.32 }}
    >
      {showMark && <XuvaLogo size={h * 1.05} />}

      <div style={{ display: "inline-flex", alignItems: "baseline", gap: h * 0.04 }}>
        <svg width={xW} height={h} viewBox="0 0 72 100" style={{ display: "block" }}>
          <defs>
            <linearGradient id={gradId} x1="0" y1="50" x2="72" y2="50" gradientUnits="userSpaceOnUse">
              <stop offset="0%" stopColor={STOPS[0]} />
              <stop offset="55%" stopColor={STOPS[1]} />
              <stop offset="100%" stopColor={STOPS[2]} />
            </linearGradient>
          </defs>
          <path d="M 4 96 L 20 96 L 68 4 L 52 4 Z" fill={`url(#${gradId})`} />
          <path d="M 4 4 L 18 4 L 68 96 L 54 96 Z" fill={`url(#${gradId})`} opacity="0.88" />
        </svg>

        <span
          style={{
            color: "#FFFFFF",
            fontSize: uvaSize,
            lineHeight: 1,
            fontWeight: 600,
            letterSpacing: "-0.045em",
            fontFamily:
              "'Space Grotesk', 'Inter', ui-sans-serif, system-ui, -apple-system, sans-serif",
          }}
        >
          uva
        </span>
      </div>
    </div>
  );
}
