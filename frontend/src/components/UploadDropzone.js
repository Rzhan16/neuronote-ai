import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useCallback } from 'react';
import { useDropzone } from 'react-dropzone';
import { cn } from '@/lib/utils';
export function UploadDropzone({ onUploadSuccess, onUploadError, onUploadProgress }) {
    const onDrop = useCallback(async (acceptedFiles) => {
        try {
            const file = acceptedFiles[0];
            if (!file)
                return;
            // Block empty files
            if (file.size === 0) {
                onUploadError(new Error('File is empty. Please select a non-empty file.'));
                return;
            }
            // Patch the file type if missing or incorrect
            let patchedFile = file;
            if (!file.type || file.type === 'application/octet-stream') {
                let type = '';
                if (file.name.endsWith('.pdf'))
                    type = 'application/pdf';
                else if (file.name.endsWith('.txt'))
                    type = 'text/plain';
                else if (file.name.endsWith('.png'))
                    type = 'image/png';
                else if (file.name.endsWith('.jpg') || file.name.endsWith('.jpeg'))
                    type = 'image/jpeg';
                if (type) {
                    patchedFile = new File([file], file.name, { type });
                }
            }
            const formData = new FormData();
            formData.append('file', patchedFile);
            // Use fetch for upload
            const xhr = new XMLHttpRequest();
            xhr.open('POST', '/api/notes/upload', true);
            xhr.setRequestHeader('X-User-ID', 'test-user-id');
            xhr.upload.onprogress = (event) => {
                if (event.lengthComputable) {
                    const progress = (event.loaded / event.total) * 100;
                    onUploadProgress(progress);
                }
            };
            xhr.onload = () => {
                if (xhr.status === 200) {
                    const response = JSON.parse(xhr.responseText);
                    onUploadSuccess(response);
                }
                else {
                    onUploadError(new Error(`Upload failed with status ${xhr.status}: ${xhr.responseText}`));
                }
            };
            xhr.onerror = () => {
                onUploadError(new Error('Upload failed'));
            };
            xhr.send(formData);
        }
        catch (error) {
            onUploadError(error instanceof Error ? error : new Error('Upload failed'));
        }
    }, [onUploadSuccess, onUploadError, onUploadProgress]);
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
    return (_jsxs("div", { ...getRootProps(), className: cn('border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-colors', isDragActive
            ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
            : 'border-gray-300 dark:border-gray-700 hover:border-primary-500 dark:hover:border-primary-500'), children: [_jsx("input", { ...getInputProps() }), _jsxs("div", { className: "space-y-4", children: [_jsx("div", { className: "mx-auto h-12 w-12 text-gray-400", children: _jsx("svg", { className: "h-full w-full", fill: "none", viewBox: "0 0 24 24", stroke: "currentColor", strokeWidth: 2, children: _jsx("path", { strokeLinecap: "round", strokeLinejoin: "round", d: "M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" }) }) }), _jsxs("div", { className: "text-gray-600 dark:text-gray-400", children: [_jsx("p", { className: "text-base font-medium", children: isDragActive ? 'Drop your file here' : 'Drag & drop your file here' }), _jsx("p", { className: "text-sm mt-1", children: "or click to browse" }), _jsx("p", { className: "text-xs mt-2", children: "Supports PDF, images, and text files" })] })] })] }));
}
export default UploadDropzone;
