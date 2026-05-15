import { XuvaLogo } from "./XuvaLogo";

interface XuvaWordmarkProps {
  size?: number;
  className?: string;
  showMark?: boolean;
}

export function XuvaWordmark({ size = 36, className, showMark = true }: XuvaWordmarkProps) {
  const h = size;

  return (
    <div
      className={className}
      style={{ display: "inline-flex", alignItems: "center", gap: h * 0.32 }}
    >
      {showMark && <XuvaLogo size={h * 1.05} />}

      <div style={{ display: "inline-flex", alignItems: "baseline", gap: h * 0.015 }}>
        <span
          style={{
            fontFamily:
              "'Space Grotesk', 'Inter', ui-sans-serif, system-ui, -apple-system, sans-serif",
            fontSize: h,
            lineHeight: 1,
            fontWeight: 500,
            letterSpacing: "-0.035em",
            background: "linear-gradient(90deg, #A78BFA 0%, #7C3AED 55%, #DB2777 100%)",
            WebkitBackgroundClip: "text",
            backgroundClip: "text",
            color: "transparent",
          }}
        >
          X
        </span>
        <span
          style={{
            fontFamily:
              "'Space Grotesk', 'Inter', ui-sans-serif, system-ui, -apple-system, sans-serif",
            fontSize: h,
            lineHeight: 1,
            fontWeight: 500,
            letterSpacing: "-0.035em",
            color: "#FFFFFF",
          }}
        >
          uva
        </span>
      </div>
    </div>
  );
}
