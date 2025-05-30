import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useState, useEffect } from 'react';
import { Card } from '@/components/Card';
import { Sidebar } from '@/components/Sidebar';
import { UploadDropzone } from '@/components/UploadDropzone';
import { cn } from '@/lib/utils';
import { ChevronRightIcon, ChevronLeftIcon, DocumentTextIcon, } from '@heroicons/react/24/outline';
export function Dashboard() {
    const [notes, setNotes] = useState([]);
    const [studyBlocks, setStudyBlocks] = useState([]);
    const [isUploading, setIsUploading] = useState(false);
    const [uploadError, setUploadError] = useState(null);
    const [isLoading, setIsLoading] = useState(true);
    const [loadError, setLoadError] = useState(null);
    const [uploadProgress, setUploadProgress] = useState(0);
    const [darkMode, setDarkMode] = useState(false);
    const [activeSection, setActiveSection] = useState('notes');
    const [currentQuestionIndex, setCurrentQuestionIndex] = useState(0);
    useEffect(() => {
        if (darkMode) {
            document.documentElement.classList.add('dark');
        }
        else {
            document.documentElement.classList.remove('dark');
        }
    }, [darkMode]);
    const fetchNotes = async () => {
        try {
            const res = await fetch('/api/notes', {
                headers: { 'X-User-ID': 'test-user-id' },
            });
            if (!res.ok)
                throw new Error(`Failed to fetch notes: ${res.statusText}`);
            const data = await res.json();
            setNotes(data);
            setLoadError(null);
        }
        catch (err) {
            console.error('Failed to fetch notes:', err);
            setLoadError('Failed to load notes. Please try again later.');
        }
    };
    const fetchStudyBlocks = async () => {
        try {
            const res = await fetch('/api/study-blocks', {
                headers: { 'X-User-ID': 'test-user-id' },
            });
            if (!res.ok)
                throw new Error(`Failed to fetch study blocks: ${res.statusText}`);
            const data = await res.json();
            setStudyBlocks(data);
            setLoadError(null);
        }
        catch (err) {
            console.error('Failed to fetch study blocks:', err);
            setLoadError('Failed to load study progress. Please try again later.');
        }
    };
    useEffect(() => {
        const loadData = async () => {
            setIsLoading(true);
            await Promise.all([fetchNotes(), fetchStudyBlocks()]);
            setIsLoading(false);
        };
        loadData();
    }, []);
    const handleUploadSuccess = async () => {
        setIsUploading(false);
        setUploadError(null);
        setUploadProgress(100);
        await fetchNotes();
        await fetchStudyBlocks();
    };
    if (isLoading) {
        return (_jsx("div", { className: "min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center", children: _jsxs("div", { className: "text-center", children: [_jsx("div", { className: "animate-spin rounded-full h-12 w-12 border-b-2 border-primary-500" }), _jsx("p", { className: "mt-4 text-gray-600 dark:text-gray-400", children: "Loading your study dashboard..." })] }) }));
    }
    if (loadError) {
        return (_jsx("div", { className: "min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center", children: _jsx(Card, { className: "max-w-md mx-auto", children: _jsxs("div", { className: "text-center", children: [_jsx("div", { className: "text-red-500 mb-4", children: _jsx(DocumentTextIcon, { className: "h-12 w-12 mx-auto" }) }), _jsx("p", { className: "text-xl font-semibold text-gray-900 dark:text-white mb-2", "data-testid": "error-message", children: loadError }), _jsx("button", { onClick: () => window.location.reload(), className: "mt-4 px-4 py-2 bg-primary-500 text-white rounded-md hover:bg-primary-600 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2", children: "Retry" })] }) }) }));
    }
    const renderContent = () => {
        switch (activeSection) {
            case 'upload':
                return (_jsxs(Card, { title: "Upload New Note", className: "max-w-2xl mx-auto", children: [uploadError && (_jsx("div", { className: "mb-4 p-4 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 rounded-md", "data-testid": "upload-error", children: uploadError })), isUploading && (_jsxs("div", { className: "mb-4", children: [_jsx("div", { className: "h-2 bg-gray-200 rounded-full", children: _jsx("div", { className: "h-2 bg-primary-600 rounded-full transition-all duration-300", style: { width: `${uploadProgress}%` }, "data-testid": "upload-progress" }) }), _jsxs("p", { className: "text-sm text-gray-500 mt-2", children: ["Processing your note... ", uploadProgress, "%"] })] })), _jsx(UploadDropzone, { onUploadSuccess: handleUploadSuccess, onUploadError: (error) => {
                                console.error('Upload failed:', error);
                                setIsUploading(false);
                                setUploadError(error.message);
                            }, onUploadProgress: (progress) => {
                                setUploadProgress(progress);
                            } })] }));
            case 'notes':
                return (_jsx("div", { className: "grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6", children: Array.isArray(notes) && notes.length > 0 ? (notes.map((note) => (_jsx(Card, { title: note.title, className: "h-full", footer: _jsxs("div", { className: "flex justify-between items-center", children: [_jsx("span", { className: "text-sm text-gray-500", children: new Date(note.created_at).toLocaleDateString() }), _jsxs("span", { className: "inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-primary-100 text-primary-800 dark:bg-primary-800/20 dark:text-primary-300", children: [note.quiz_cards.length, " cards"] })] }), children: _jsx("p", { className: "text-gray-600 dark:text-gray-300 line-clamp-3", children: note.summary }) }, note.id)))) : (_jsx("div", { className: "col-span-full text-center text-gray-500 dark:text-gray-400 py-8", children: "No notes found. Upload a note to get started." })) }));
            case 'questions':
                if (!notes.length || !notes[0].quiz_cards.length) {
                    return (_jsxs(Card, { className: "text-center py-12", children: [_jsx(DocumentTextIcon, { className: "h-12 w-12 mx-auto text-gray-400" }), _jsx("h3", { className: "mt-2 text-sm font-medium text-gray-900 dark:text-white", children: "No questions yet" }), _jsx("p", { className: "mt-1 text-sm text-gray-500", children: "Upload a note to generate study questions." })] }));
                }
                {
                    const currentNote = notes[0];
                    const currentCard = currentNote.quiz_cards[currentQuestionIndex];
                    return (_jsxs(Card, { className: "max-w-2xl mx-auto", children: [_jsxs("div", { className: "flex items-center justify-between mb-4", children: [_jsx("button", { onClick: () => setCurrentQuestionIndex((i) => Math.max(0, i - 1)), disabled: currentQuestionIndex === 0, className: "p-2 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-50", children: _jsx(ChevronLeftIcon, { className: "h-6 w-6" }) }), _jsxs("span", { className: "text-sm text-gray-500", children: ["Question ", currentQuestionIndex + 1, " of ", currentNote.quiz_cards.length] }), _jsx("button", { onClick: () => setCurrentQuestionIndex((i) => Math.min(currentNote.quiz_cards.length - 1, i + 1)), disabled: currentQuestionIndex === currentNote.quiz_cards.length - 1, className: "p-2 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-50", children: _jsx(ChevronRightIcon, { className: "h-6 w-6" }) })] }), _jsxs("div", { className: "space-y-4", children: [_jsxs("div", { className: "p-6 bg-gray-50 dark:bg-gray-800/50 rounded-lg", children: [_jsx("h4", { className: "font-medium text-gray-900 dark:text-white mb-2", children: "Question:" }), _jsx("p", { className: "text-gray-600 dark:text-gray-300", children: currentCard.question })] }), _jsxs("div", { className: "p-6 bg-primary-50 dark:bg-primary-900/20 rounded-lg", children: [_jsx("h4", { className: "font-medium text-gray-900 dark:text-white mb-2", children: "Answer:" }), _jsx("p", { className: "text-gray-600 dark:text-gray-300", children: currentCard.answer })] })] })] }));
                }
            case 'study':
                return (_jsx(Card, { title: "Study Progress", className: "max-w-2xl mx-auto", children: _jsx("div", { className: "space-y-4", children: Array.isArray(studyBlocks) && studyBlocks.length > 0 ? (studyBlocks.map((block) => (_jsx("div", { className: cn('p-4 rounded-lg border', block.status === 'completed'
                                ? 'bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800'
                                : 'bg-gray-50 dark:bg-gray-800/50 border-gray-200 dark:border-gray-700'), children: _jsxs("div", { className: "flex justify-between items-center", children: [_jsxs("div", { children: [_jsxs("p", { className: "font-medium text-gray-900 dark:text-white", children: ["Study Session ", new Date(block.start_time).toLocaleDateString()] }), _jsxs("p", { className: "text-sm text-gray-500", children: [new Date(block.start_time).toLocaleTimeString(), " -", ' ', new Date(block.end_time).toLocaleTimeString()] })] }), _jsx("span", { className: cn('px-2.5 py-0.5 rounded-full text-xs font-medium', block.status === 'completed'
                                            ? 'bg-green-100 text-green-800 dark:bg-green-800/20 dark:text-green-300'
                                            : 'bg-gray-100 text-gray-800 dark:bg-gray-800/20 dark:text-gray-300'), children: block.status })] }) }, block.id)))) : (_jsx("div", { className: "text-center text-gray-500 dark:text-gray-400 py-8", children: "No study blocks found. Complete a study session to see progress." })) }) }));
        }
    };
    return (_jsxs("div", { className: "min-h-screen bg-gray-50 dark:bg-gray-900", children: [_jsx(Sidebar, { darkMode: darkMode, onDarkModeToggle: () => setDarkMode(!darkMode), activeSection: activeSection, onSectionChange: setActiveSection }), _jsx("main", { className: "pl-18", children: _jsxs("div", { className: "max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8", children: [_jsxs("h1", { className: "text-2xl font-semibold text-gray-900 dark:text-white mb-8", children: [activeSection === 'upload' && 'Upload New Note', activeSection === 'notes' && 'Your Notes', activeSection === 'questions' && 'Study Questions', activeSection === 'study' && 'Study Progress'] }), renderContent()] }) })] }));
}
