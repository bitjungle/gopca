import { useState, useEffect } from 'react';
import { CalculateEllipses } from '../../wailsjs/go/main/App';
import { EllipseParams } from '../types';

interface UseEllipsesParams {
  scores?: number[][];
  groupLabels?: string[];
  xComponent: number;
  yComponent: number;
  enabled: boolean;
}

interface UseEllipsesResult {
  ellipses90: Record<string, EllipseParams> | undefined;
  ellipses95: Record<string, EllipseParams> | undefined;
  ellipses99: Record<string, EllipseParams> | undefined;
  isLoading: boolean;
  error: string | null;
}

export function useEllipses({
  scores,
  groupLabels,
  xComponent,
  yComponent,
  enabled
}: UseEllipsesParams): UseEllipsesResult {
  const [ellipses90, setEllipses90] = useState<Record<string, EllipseParams> | undefined>();
  const [ellipses95, setEllipses95] = useState<Record<string, EllipseParams> | undefined>();
  const [ellipses99, setEllipses99] = useState<Record<string, EllipseParams> | undefined>();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!enabled || !scores || !groupLabels || groupLabels.length === 0) {
      setEllipses90(undefined);
      setEllipses95(undefined);
      setEllipses99(undefined);
      return;
    }

    const calculateEllipses = async () => {
      setIsLoading(true);
      setError(null);

      try {
        const response = await CalculateEllipses({
          scores,
          groupLabels,
          xComponent,
          yComponent
        });

        if (response.success) {
          setEllipses90(response.groupEllipses90);
          setEllipses95(response.groupEllipses95);
          setEllipses99(response.groupEllipses99);
        } else {
          setError(response.error || 'Failed to calculate ellipses');
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to calculate ellipses');
      } finally {
        setIsLoading(false);
      }
    };

    calculateEllipses();
  }, [scores, groupLabels, xComponent, yComponent, enabled]);

  return { ellipses90, ellipses95, ellipses99, isLoading, error };
}