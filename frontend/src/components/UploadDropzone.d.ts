interface UploadDropzoneProps {
    onUploadSuccess: (response: unknown) => void;
    onUploadError: (error: Error) => void;
    onUploadProgress: (progress: number) => void;
}
export declare function UploadDropzone({ onUploadSuccess, onUploadError, onUploadProgress }: UploadDropzoneProps): import("react/jsx-runtime").JSX.Element;
export default UploadDropzone;
