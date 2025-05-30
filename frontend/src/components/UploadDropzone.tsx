import React, { useCallback, useState } from 'react';
import { useDropzone } from 'react-dropzone';
import { cn } from '@/lib/utils'; // Assuming shadcn/ui setup @/lib/utils

interface UploadDropzoneProps {
  onUploadSuccess?: (response: any) => void;
  onUploadError?: (error: any) => void;
}

export function UploadDropzone({ onUploadSuccess, onUploadError }: UploadDropzoneProps) {
  const [isUploading, setIsUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [error, setError] = useState<string | null>(null);

  const onDrop = useCallback(async (acceptedFiles: File[]) => {
    if (acceptedFiles.length === 0) {
      return;
    }
    const file = acceptedFiles[0];
    setIsUploading(true);
    setError(null);
    setUploadProgress(0);

    const formData = new FormData();
    formData.append('file', file);

    try {
      const xhr = new XMLHttpRequest();
      xhr.open('POST', '/api/notes/upload', true);
      // Cookies should be sent automatically by the browser if the API is on the same domain
      // or if CORS is configured correctly on the server to allow credentials.

      xhr.upload.onprogress = (event) => {
        if (event.lengthComputable) {
          const percentComplete = Math.round((event.loaded * 100) / event.total);
          setUploadProgress(percentComplete);
        }
      };

      xhr.onload = () => {
        setIsUploading(false);
        if (xhr.status >= 200 && xhr.status < 300) {
          try {
            const responseJson = JSON.parse(xhr.responseText);
            if (onUploadSuccess) {
              onUploadSuccess(responseJson);
            }
          } catch (e) {
            console.error('Failed to parse upload response:', e);
            setError('Upload succeeded but failed to parse server response.');
            if (onUploadError) {
              onUploadError(new Error('Failed to parse server response.'));
            }
          }
        } else {
          let errorMessage = `Upload failed with status: ${xhr.status}`;
          try {
            const errorResponse = JSON.parse(xhr.responseText);
            errorMessage = errorResponse.error || errorResponse.message || errorMessage;
          } catch (e) {
            // Keep default error message
          }
          setError(errorMessage);
          if (onUploadError) {
            onUploadError(new Error(errorMessage));
          }
        }
      };

      xhr.onerror = () => {
        setIsUploading(false);
        setError('Upload failed due to a network error.');
        if (onUploadError) {
          onUploadError(new Error('Network error during upload.'));
        }
      };

      xhr.send(formData);

    } catch (err) {
      setIsUploading(false);
      let message = 'An unknown error occurred during upload.';
      if (err instanceof Error) {
        message = err.message;
      }
      setError(message);
      console.error('Upload error:', err);
      if (onUploadError) {
        onUploadError(err);
      }
    }
  }, [onUploadSuccess, onUploadError]);

  const { getRootProps, getInputProps, isDragActive, isDragAccept, isDragReject } = useDropzone({
    onDrop,
    accept: {
      'image/*': ['.jpeg', '.jpg', '.png', '.gif', '.webp'],
      'audio/*': ['.mp3', '.wav', '.ogg', '.m4a'],
      'application/pdf': ['.pdf'],
    },
    multiple: false,
  });

  return (
    <div
      {...getRootProps({
        className: cn(
          'p-10 border-2 border-dashed rounded-lg text-center cursor-pointer',
          'transition-colors duration-200 ease-in-out',
          isDragActive ? 'border-blue-500 bg-blue-50' : 'border-gray-300 hover:border-gray-400',
          isDragAccept ? 'border-green-500 bg-green-50' : '',
          isDragReject ? 'border-red-500 bg-red-50' : '',
          isUploading ? 'bg-gray-100 cursor-not-allowed' : ''
        ),
      })}
    >
      <input {...getInputProps()} disabled={isUploading} />
      {isUploading ? (
        <div className="flex flex-col items-center">
          <p className="mb-2 text-lg font-semibold">Uploading...</p>
          <div className="w-full bg-gray-200 rounded-full h-2.5 dark:bg-gray-700">
            <div
              className="bg-blue-600 h-2.5 rounded-full transition-all duration-150"
              style={{ width: `${uploadProgress}%` }}
            ></div>
          </div>
          <p className="mt-1 text-sm text-gray-600">{uploadProgress}%</p>
        </div>
      ) : isDragActive ? (
        <p className="text-blue-600 font-semibold">
          {isDragAccept && 'Drop the file here!'}
          {isDragReject && 'File type not accepted'}
          {!isDragAccept && !isDragReject && 'Release to drop the file'}
        </p>
      ) : (
        <p className="text-gray-500">
          Drag & drop a file here, or click to select a file
          <br />
          <span className="text-xs text-gray-400">(Images, Audio, PDF)</span>
        </p>
      )}
      {error && (
        <p className="mt-4 text-sm text-red-600 bg-red-100 p-2 rounded">
          Error: {error}
        </p>
      )}
    </div>
  );
}

export default UploadDropzone; 