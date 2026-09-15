/**
 * How a point is named when the data carries no row names.
 *
 * A CSV without a row-name column leaves RowNames nil, so every chart falls back
 * to naming points by position. That fallback was written out at fourteen
 * separate places and had already drifted: twelve used a zero-based index and
 * two used a one-based one, so the same unnamed sample appeared as "Sample 5" in
 * a scores plot and "Sample 6" in a contribution plot.
 *
 * One based is the correct half of that disagreement. GoCSV numbers its rows
 * from one in the position gutter, so the first row of a file is row 1 there and
 * must not be "Sample 0" here -- the suite would be off by one against itself,
 * and a user cross-checking a plot against the grid would land on the wrong row.
 *
 * Duplicating the expression is what let it drift, so it lives here instead.
 */
export function sampleLabel(sampleNames: string[] | undefined, index: number): string {
    const name = sampleNames?.[index];
    return name !== undefined && name !== '' ? name : `Sample ${index + 1}`;
}
