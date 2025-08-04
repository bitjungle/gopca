/**
 * Custom hook for file operations
 * Consolidates file loading, validation, and data management
 */

import { useState, useCallback } from 'react';
import { main } from '../../wailsjs/go/models';
import { LoadCSV, ClearHistory } from '../../wailsjs/go/main/App';
import { handleAsync, useLoadingState } from '@gopca/ui-components';

type FileData = main.FileData;

export function useFileOperations() {
  const [fileLoaded, setFileLoaded] = useState(false);
  const [fileName, setFileName] = useState<string | null>(null);
  const [fileData, setFileData] = useState<FileData | null>(null);
  const { isLoading, withLoading } = useLoadingState();
  
  const resetFileState = useCallback(() => {
    setFileLoaded(false);
    setFileName(null);
    setFileData(null);
  }, []);
  
  const loadFile = useCallback(async (filePath: string) => {
    return withLoading(async () => {
      const result = await handleAsync(
        () => LoadCSV(filePath),
        {
          errorPrefix: 'Error loading file',
          showUserError: true
        }
      );
      
      if (result && result.data && result.data.length > 0) {
        setFileData(result);
        setFileLoaded(true);
        // Clear history when loading new file
        await ClearHistory();
        return result;
      } else {
        throw new Error('Failed to load file or file is empty');
      }
    });
  }, [withLoading]);
  
  const updateData = useCallback((
    updater: (data: FileData) => FileData
  ) => {
    if (fileData) {
      setFileData(updater(fileData));
    }
  }, [fileData]);
  
  return {
    fileLoaded,
    fileName,
    fileData,
    isLoading,
    loadFile,
    resetFileState,
    updateData,
    setFileName
  };
}