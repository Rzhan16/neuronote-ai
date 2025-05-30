import React, { useCallback } from 'react';
import { useDropzone } from 'react-dropzone';
import { cn } from '@/lib/utils';

interface UploadDropzoneProps {
  onUploadSuccess: (response: unknown) => void;
  onUploadError: (error: Error) => void;
  onUploadProgress: (progress: number) => void;
}

export function UploadDropzone({ onUploadSuccess, onUploadError, onUploadProgress }: UploadDropzoneProps) {
  const onDrop = useCallback(async (acceptedFiles: File[]) => {
    try {
      const file = acceptedFiles[0];
      if (!file) return;

      // Block empty files
      if (file.size === 0) {
        onUploadError(new Error('File is empty. Please select a non-empty file.'));
        return;
      }

      // Patch the file type if missing or incorrect
      let patchedFile = file;
      if (!file.type || file.type === 'application/octet-stream') {
        let type = '';
        if (file.name.endsWith('.pdf')) type = 'application/pdf';
        else if (file.name.endsWith('.txt')) type = 'text/plain';
        else if (file.name.endsWith('.png')) type = 'image/png';
        else if (file.name.endsWith('.jpg') || file.name.endsWith('.jpeg')) type = 'image/jpeg';
        if (type) {
          patchedFile = new File([file], file.name, { type });
        }
      }

      const formData = new FormData();
      formData.append('file', patchedFile);

      // Use fetch for upload
      const response = await fetch('/api/notes/upload', {
        method: 'POST',
        headers: {
          'X-User-ID': 'test-user-id',
        },
        body: formData,
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Upload failed with status ${response.status}: ${errorText}`);
      }

      const result = await response.json();
      onUploadSuccess(result);
    } catch (error) {
      onUploadError(error instanceof Error ? error : new Error('Upload failed'));
    }
  }, [onUploadSuccess, onUploadError]);

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: {
      'application/pdf': ['.pdf'],
      'image/png': ['.png'],
      'image/jpeg': ['.jpg', '.jpeg'],
      'text/plain': ['.txt'],
    },
    maxFiles: 1,
  });

  return (
    <div
      {...getRootProps()}
      className={cn(
        'border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-colors',
        isDragActive
          ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
          : 'border-gray-300 dark:border-gray-700 hover:border-primary-500 dark:hover:border-primary-500'
      )}
    >
      <input {...getInputProps()} />
      <div className="space-y-4">
        <div className="mx-auto h-12 w-12 text-gray-400">
          <svg
            className="h-full w-full"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={2}
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
            />
          </svg>
        </div>
        <div className="text-gray-600 dark:text-gray-400">
          <p className="text-base font-medium">
            {isDragActive ? 'Drop your file here' : 'Drag & drop your file here'}
          </p>
          <p className="text-sm mt-1">or click to browse</p>
          <p className="text-xs mt-2">Supports PDF, images, and text files</p>
        </div>
      </div>
    </div>
  );
}

export default UploadDropzone;