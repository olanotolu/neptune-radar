import { useEffect, useLayoutEffect, useState, useCallback } from "react";

const STORAGE_KEY = "neptune-onboarded";

type Step = {
  selector: string;
  title: string;
  description: string;
};

const STEPS: Step[] = [
  {
    selector: ".live-indicator",
    title: "Your Radar",
    description: "This is your Radar. It watches Instagram for newly engaged couples.",
  },
  {
    selector: '[data-testid="nav-today"]',
    title: "Celebrate First",
    description: "Your celebrate-first queue. When a couple is ready, start here.",
  },
  {
    selector: '[data-testid="nav-work"]',
    title: "Prospect Board",
    description: "The prospect board. Review and approve couples for outreach.",
  },
  {
    selector: '[data-testid="nav-map"]',
    title: "Coverage Map",
    description: "Coverage map. 50 states + DC of wedding sources.",
  },
  {
    selector: ".app-search",
    title: "Search Anything",
    description: "Search anything. Couples, leads, cases — all here.",
  },
];

export function isOnboarded(): boolean {
  return localStorage.getItem(STORAGE_KEY) === "1";
}

export function resetOnboarding(): void {
  localStorage.removeItem(STORAGE_KEY);
}

type Rect = { top: number; left: number; width: number; height: number };

export function OnboardingTour({ onClose }: { onClose: () => void }) {
  const [step, setStep] = useState(0);
  const [targetRect, setTargetRect] = useState<Rect | null>(null);

  const finish = useCallback(() => {
    localStorage.setItem(STORAGE_KEY, "1");
    onClose();
  }, [onClose]);

  const measure = useCallback(() => {
    const el = document.querySelector(STEPS[step].selector) as HTMLElement | null;
    if (!el) return;
    const r = el.getBoundingClientRect();
    setTargetRect({ top: r.top, left: r.left, width: r.width, height: r.height });
  }, [step]);

  useLayoutEffect(() => {
    measure();
  }, [measure]);

  // Re-measure on resize / scroll while the tour is active.
  useEffect(() => {
    const handler = () => measure();
    window.addEventListener("resize", handler);
    window.addEventListener("scroll", handler, true);
    return () => {
      window.removeEventListener("resize", handler);
      window.removeEventListener("scroll", handler, true);
    };
  }, [measure]);

  const current = STEPS[step];
  const isLast = step === STEPS.length - 1;

  // Tooltip positioning: place below the target by default, flip above if
  // there isn't enough room. Clamp horizontally so it never overflows.
  const tooltipStyle: React.CSSProperties = (() => {
    if (!targetRect) return { opacity: 0 };
    const tooltipW = 320;
    const gap = 12;
    const below = targetRect.top + targetRect.height + gap;
    const flipUp = below + 160 > window.innerHeight;
    const top = flipUp ? Math.max(gap, targetRect.top - 160 - gap) : below;
    const left = Math.min(
      Math.max(gap, targetRect.left + targetRect.width / 2 - tooltipW / 2),
      window.innerWidth - tooltipW - gap,
    );
    return { top, left, width: tooltipW };
  })();

  return (
    <div className="onboarding-overlay" role="dialog" aria-label="Welcome tour">
      {/* Backdrop with a cutout for the target element */}
      {targetRect && (
        <div
          className="onboarding-highlight"
          style={{
            top: targetRect.top - 4,
            left: targetRect.left - 4,
            width: targetRect.width + 8,
            height: targetRect.height + 8,
          }}
        />
      )}
      <div className="onboarding-tooltip" style={tooltipStyle}>
        <div className="onboarding-tooltip__progress">
          {step + 1} / {STEPS.length}
        </div>
        <h3 className="onboarding-tooltip__title">{current.title}</h3>
        <p className="onboarding-tooltip__desc">{current.description}</p>
        <div className="onboarding-tooltip__actions">
          <button type="button" className="onboarding-tooltip__skip" onClick={finish}>
            Skip tour
          </button>
          <button
            type="button"
            className="btn btn--primary btn--sm"
            onClick={() => (isLast ? finish() : setStep((s) => s + 1))}
          >
            {isLast ? "Done" : "Next"}
          </button>
        </div>
      </div>
    </div>
  );
}
