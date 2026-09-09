// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// See LICENSE for the full license terms.

/**
 * Keeping variable labels apart in the Circle of Correlations.
 *
 * Labels were placed radially, at a fixed multiple of each arrow tip, centred,
 * with nothing to stop two of them landing on top of each other. On Wine
 * (standard-scaled, PC1/PC2) three arrows sit within about 7° of each other and
 * their names overlap — and those are precisely the tightly correlated
 * variables a reader is trying to tell apart, so the labels fail exactly where
 * they are needed most (#830).
 *
 * Why the separation is angular rather than radial. Two labels differing only
 * in arrow length separate along the radius, and near the horizontal that is
 * the direction in which text is *widest*. Measured on the Wine pair
 * `proanthocyanins` / `total_phenols`: their anchors are 43 px apart radially,
 * against 77 px of combined half-width — they overlap. Moving them along the
 * tangent instead separates them in the direction where a label is only about
 * one line tall, so a nudge of a few degrees is enough. The labels stay on
 * their own radius, so arrow length still reads correctly.
 */

/** A label to place, positioned at its arrow tip. */
export interface CircleLabel {
  text: string;
  /** Correlation with the first displayed component. */
  x: number;
  /** Correlation with the second displayed component. */
  y: number;
}

export interface CircleLabelPlacementOptions {
  /** Where the label sits along its arrow's direction. */
  radiusFactor?: number;
  /** Rendered font size in pixels. */
  fontSizePx?: number;
  /**
   * Assumed rendered size of the square plot, in pixels.
   *
   * The only calibrated number here, and it is needed because label size is
   * fixed in pixels while positions are in data units, so the two can only be
   * compared through the rendered scale — which the trace builder does not
   * know. Everything else is exact geometry. Getting this wrong makes the
   * de-confliction slightly too eager or too shy, never wrong in kind.
   */
  plotSizePx?: number;
  /** Width of the data range on each axis; the circle plot uses -1.2..1.2. */
  axisSpan?: number;
  /**
   * Furthest a label may reach from the origin before it is clipped.
   *
   * Pushing a label clear of its own arrow costs half its width, and the
   * longest names cannot afford that: `od280/od315_of_diluted_wines` would end
   * up 1.67 from the origin against an axis that stops at 1.2, so it would be
   * cut off. Trading an overlapped arrow for an unreadable label is not a
   * trade, so a label that cannot fit outside is left where it was.
   */
  axisLimit?: number;
  /**
   * How far a label may be rotated away from its own arrow.
   *
   * A cap matters: a label that has drifted far from its arrow is no longer
   * obviously that arrow's name, which is the problem this set out to fix
   * rather than a solution to it. When the cap is reached the overlap is left
   * in place, because a wrong association is worse than a crowded one.
   */
  maxShiftDegrees?: number;
  /** Mean glyph width as a fraction of font size. */
  charWidthRatio?: number;
  /** Line height as a fraction of font size. */
  lineHeightRatio?: number;
  /**
   * Gap left between an arrow tip and the near edge of its label, in pixels.
   *
   * Labels were centred on a point at a fixed multiple of the arrow tip, which
   * put the inner half of the text back over the arrow it names -- on Wine that
   * covered a third of some arrows, and `proanthocyanins` rendered as
   * "oanthocyanins" with its first letters lost under its own arrowhead. Each
   * label is now pushed out until it clears its tip.
   */
  tipClearancePx?: number;
  /**
   * Clearance required around each label, in pixels.
   *
   * Labels that merely fail to overlap are still unreadable: without whitespace
   * between them the eye cannot tell where one name ends. Calibrated against
   * the three pairs reported as colliding on Wine in #830 — at 4px all three
   * are recognised as collisions and the next-closest pair, 10.8° away, is
   * not. Six leaves margin without reaching that pair.
   */
  paddingPx?: number;
}

export interface PlacedCircleLabel {
  x: number;
  y: number;
  /** Signed rotation applied, in degrees; zero when nothing had to move. */
  shiftedDegrees: number;
}

const DEFAULTS: Required<CircleLabelPlacementOptions> = {
  radiusFactor: 1.15,
  fontSizePx: 10,
  plotSizePx: 500,
  axisSpan: 2.4,
  axisLimit: 1.2,
  maxShiftDegrees: 14,
  charWidthRatio: 0.55,
  lineHeightRatio: 1.2,
  paddingPx: 6,
  tipClearancePx: 4
};

/** Degrees moved per relaxation step; small enough not to overshoot a fix. */
const STEP_DEGREES = 0.5;

const toRadians = (degrees: number) => (degrees * Math.PI) / 180;

interface Placement {
  index: number;
  angle: number;
  radius: number;
  halfWidth: number;
  halfHeight: number;
  shift: number;
}

/**
 * How far along its own direction a label should sit.
 *
 * Three bounds, in order of priority: never nearer than the original radial
 * placement, far enough out to clear the arrow tip, and not so far that the
 * text leaves the plot. When the last two conflict -- a long name on a long
 * arrow -- the label stays where it was and keeps its overlap, because a
 * clipped label is worse than a crowded one.
 */
function boundedRadius(
  tipRadius: number,
  reach: number,
  settings: Required<CircleLabelPlacementOptions>,
  dataPerPixel: number
): number {
  const original = tipRadius * settings.radiusFactor;
  const clearsTip = tipRadius + reach + settings.tipClearancePx * dataPerPixel;
  const fitsOnAxis = settings.axisLimit - reach;
  return Math.max(original, Math.min(clearsTip, Math.max(original, fitsOnAxis)));
}

/**
 * Returns a position for each label such that, where possible, no two overlap.
 *
 * Labels keep their own radius and move only in angle, by at most
 * `maxShiftDegrees`. The result is deterministic: the same input always yields
 * the same output, so a plot does not rearrange itself between renders.
 */
export function placeCircleLabels(
  labels: CircleLabel[],
  options: CircleLabelPlacementOptions = {}
): PlacedCircleLabel[] {
  const settings = { ...DEFAULTS, ...options };
  const dataPerPixel = settings.axisSpan / settings.plotSizePx;

  const placements: Placement[] = labels.map((label, index) => {
    const tipRadius = Math.hypot(label.x, label.y);
    const angle = Math.atan2(label.y, label.x);
    const halfWidth = ((label.text.length * settings.charWidthRatio * settings.fontSizePx)
      + 2 * settings.paddingPx) * dataPerPixel / 2;
    const halfHeight = ((settings.lineHeightRatio * settings.fontSizePx)
      + 2 * settings.paddingPx) * dataPerPixel / 2;

    // How far the label box reaches back toward the origin: the support of an
    // axis-aligned box in the inward radial direction. For an arrow pointing
    // sideways that is the label's half-width, and for one pointing up it is
    // half a line height -- which is why a vertical arrow needs almost no push
    // and a horizontal one needs a lot.
    const inwardReach = halfWidth * Math.abs(Math.cos(angle))
      + halfHeight * Math.abs(Math.sin(angle));

    return {
      index,
      angle,
      // Far enough out that the text starts beyond its own arrow, but never
      // closer than the original placement, so short names are unaffected, and
      // never so far that the label runs off the axis.
      radius: boundedRadius(tipRadius, inwardReach, settings, dataPerPixel),
      halfWidth,
      halfHeight,
      shift: 0
    };
  });

  const position = (p: Placement) => ({
    x: Math.cos(p.angle) * p.radius,
    y: Math.sin(p.angle) * p.radius
  });

  const overlaps = (a: Placement, b: Placement) => {
    const pa = position(a);
    const pb = position(b);
    return (
      Math.abs(pa.x - pb.x) < a.halfWidth + b.halfWidth
      && Math.abs(pa.y - pb.y) < a.halfHeight + b.halfHeight
    );
  };

  // Relax in small steps rather than solving in one move. A single jump sized
  // to clear the current overlap can push a label into a third one; stepping
  // lets each pass see the arrangement the previous one produced.
  const maxIterations = Math.ceil(settings.maxShiftDegrees / STEP_DEGREES) * 2;
  for (let iteration = 0; iteration < maxIterations; iteration++) {
    let moved = false;

    for (let i = 0; i < placements.length; i++) {
      for (let j = i + 1; j < placements.length; j++) {
        const a = placements[i];
        const b = placements[j];

        // A label at the origin has no direction to be rotated around, and an
        // arrow of zero length has no angle worth preserving anyway.
        if (a.radius === 0 || b.radius === 0 || !overlaps(a, b)) {
          continue;
        }

        // Push each away from the other along the circle. Which way is "away"
        // is decided by their current angular order, normalised so that a pair
        // straddling the -pi/pi wrap is still pushed apart rather than through
        // each other.
        let delta = b.angle - a.angle;
        while (delta > Math.PI) delta -= 2 * Math.PI;
        while (delta < -Math.PI) delta += 2 * Math.PI;
        const direction = delta >= 0 ? 1 : -1;

        const step = toRadians(STEP_DEGREES);
        const cap = toRadians(settings.maxShiftDegrees);

        const moveA = Math.max(-cap - a.shift, Math.min(cap - a.shift, -direction * step));
        const moveB = Math.max(-cap - b.shift, Math.min(cap - b.shift, direction * step));

        if (moveA !== 0) {
          a.angle += moveA;
          a.shift += moveA;
          moved = true;
        }
        if (moveB !== 0) {
          b.angle += moveB;
          b.shift += moveB;
          moved = true;
        }
      }
    }

    if (!moved) {
      break;
    }
  }

  // Re-apply the radial bound at the final angle. The relaxation above rotates
  // labels, and how far a box reaches along the radius depends on its
  // direction, so a radius computed before the rotation is slightly wrong
  // afterwards -- enough to leave a label just touching the arrow it was moved
  // out to clear.
  for (const p of placements) {
    const tipRadius = Math.hypot(labels[p.index].x, labels[p.index].y);
    const reach = p.halfWidth * Math.abs(Math.cos(p.angle))
      + p.halfHeight * Math.abs(Math.sin(p.angle));
    p.radius = boundedRadius(tipRadius, reach, settings, dataPerPixel);
  }

  const result: PlacedCircleLabel[] = new Array(placements.length);
  for (const p of placements) {
    const { x, y } = position(p);
    result[p.index] = { x, y, shiftedDegrees: (p.shift * 180) / Math.PI };
  }
  return result;
}
