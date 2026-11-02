import { useId } from "react";

interface RulerProps {
  percent: number;
  tone?: "brand" | "good" | "warn" | "danger";
  labels?: boolean;
  className?: string;
  ariaLabel?: string;
}

const TICKS = Array.from({ length: 101 }, (_, index) => index * 10);

export function Ruler({ percent, tone = "brand", labels = false, className = "", ariaLabel }: RulerProps) {
  const clipId = useId().replace(/:/g, "");
  const value = Math.min(100, Math.max(0, percent));
  return (
    <div className={`z-ruler ${className}`} dir="ltr">
      <svg
        viewBox="0 0 1000 32"
        preserveAspectRatio="none"
        role="img"
        aria-label={ariaLabel}
        aria-hidden={ariaLabel ? undefined : true}
      >
        <defs>
          <clipPath id={clipId}>
            <rect x="0" y="0" width={value * 10} height="32" />
          </clipPath>
        </defs>
        <g className="z-ruler-base">
          {TICKS.map((tick) => {
            const long = tick % 100 === 0;
            const medium = tick % 50 === 0;
            const height = long ? 14 : medium ? 10 : 6;
            return (
              <line
                key={tick}
                x1={tick}
                x2={tick}
                y1={32 - height}
                y2={32}
                strokeWidth={long ? 1.6 : 1}
              />
            );
          })}
        </g>
        <g className={`z-ruler-accent z-ruler-${tone}`} clipPath={`url(#${clipId})`}>
          {TICKS.map((tick) => {
            const long = tick % 100 === 0;
            const medium = tick % 50 === 0;
            const height = long ? 14 : medium ? 10 : 6;
            return (
              <line
                key={tick}
                x1={tick}
                x2={tick}
                y1={32 - height}
                y2={32}
                strokeWidth={long ? 1.8 : 1.2}
              />
            );
          })}
        </g>
        <line
          className={`z-ruler-marker z-ruler-${tone}`}
          x1={value * 10}
          x2={value * 10}
          y1="0"
          y2="32"
          strokeWidth="2.5"
        />
        <circle className={`z-ruler-dot z-ruler-${tone}`} cx={value * 10} cy="1.5" r="3.5" />
      </svg>
      {labels && (
        <div className="z-ruler-labels" aria-hidden="true">
          <span>1×</span>
          <span>2×</span>
        </div>
      )}
    </div>
  );
}
