import React, { useState, useEffect } from 'react';
import { Card } from '@/components/Card';
import { Sidebar } from '@/components/Sidebar';
import { UploadDropzone } from '@/components/UploadDropzone';
import { cn } from '@/lib/utils';
import {
  ChevronRightIcon,
  ChevronLeftIcon,
  DocumentTextIcon,
} from '@heroicons/react/24/outline';

interface Note {
  id: string;
  title: string;
  summary: string;
  quiz_cards: Array<{ id: string; question: string; answer: string }>;
  created_at: string;
  updated_at: string;
}

interface StudyBlock {
  id: string;
  status: 'scheduled' | 'completed';
  start_time: string;
  end_time: string;
}

export function Dashboard() {
  const [notes, setNotes] = useState<Note[]>([]);
  const [studyBlocks, setStudyBlocks] = useState<StudyBlock[]>([]);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [uploadProgress, setUploadProgress] = useState<number>(0);
  const [darkMode, setDarkMode] = useState(false);
  const [activeSection, setActiveSection] = useState<'upload' | 'notes' | 'questions' | 'study'>('notes');
  const [currentQuestionIndex, setCurrentQuestionIndex] = useState(0);

  useEffect(() => {
    if (darkMode) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [darkMode]);

  const fetchNotes = async () => {
    try {
      const res = await fetch('/api/notes', {
        headers: { 'X-User-ID': 'test-user-id' },
      });
      if (!res.ok) throw new Error(`Failed to fetch notes: ${res.statusText}`);
      const data = await res.json();
      setNotes(data);
      setLoadError(null);
    } catch (err) {
      console.error('Failed to fetch notes:', err);
      setLoadError('Failed to load notes. Please try again later.');
    }
  };

  const fetchStudyBlocks = async () => {
    try {
      const res = await fetch('/api/study-blocks', {
        headers: { 'X-User-ID': 'test-user-id' },
      });
      if (!res.ok) throw new Error(`Failed to fetch study blocks: ${res.statusText}`);
      const data = await res.json();
      setStudyBlocks(data);
      setLoadError(null);
    } catch (err) {
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
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-500"></div>
          <p className="mt-4 text-gray-600 dark:text-gray-400">Loading your study dashboard...</p>
        </div>
      </div>
    );
  }

  if (loadError) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center">
        <Card className="max-w-md mx-auto">
          <div className="text-center">
            <div className="text-red-500 mb-4">
              <DocumentTextIcon className="h-12 w-12 mx-auto" />
            </div>
            <p className="text-xl font-semibold text-gray-900 dark:text-white mb-2" data-testid="error-message">
              {loadError}
            </p>
            <button
              onClick={() => window.location.reload()}
              className="mt-4 px-4 py-2 bg-primary-500 text-white rounded-md hover:bg-primary-600 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2"
            >
              Retry
            </button>
          </div>
        </Card>
      </div>
    );
  }

  const renderContent = () => {
    switch (activeSection) {
      case 'upload':
        return (
          <Card title="Upload New Note" className="max-w-2xl mx-auto">
            {uploadError && (
              <div className="mb-4 p-4 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 rounded-md" data-testid="upload-error">
                {uploadError}
              </div>
            )}
            {isUploading && (
              <div className="mb-4">
                <div className="h-2 bg-gray-200 rounded-full">
                  <div
                    className="h-2 bg-primary-600 rounded-full transition-all duration-300"
                    style={{ width: `${uploadProgress}%` }}
                    data-testid="upload-progress"
                  />
                </div>
                <p className="text-sm text-gray-500 mt-2">Processing your note... {uploadProgress}%</p>
              </div>
            )}
            <UploadDropzone
              onUploadSuccess={handleUploadSuccess}
              onUploadError={(error) => {
                console.error('Upload failed:', error);
                setIsUploading(false);
                setUploadError(error.message);
              }}
              onUploadProgress={(progress) => {
                setUploadProgress(progress);
              }}
            />
          </Card>
        );

      case 'notes':
        return (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {Array.isArray(notes) && notes.length > 0 ? (
              notes.map((note) => (
                <Card
                  key={note.id}
                  title={note.title}
                  className="h-full"
                  footer={
                    <div className="flex justify-between items-center">
                      <span className="text-sm text-gray-500">
                        {new Date(note.created_at).toLocaleDateString()}
                      </span>
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-primary-100 text-primary-800 dark:bg-primary-800/20 dark:text-primary-300">
                        {note.quiz_cards.length} cards
                      </span>
                    </div>
                  }
                >
                  <p className="text-gray-600 dark:text-gray-300 line-clamp-3">{note.summary}</p>
                </Card>
              ))
            ) : (
              <div className="col-span-full text-center text-gray-500 dark:text-gray-400 py-8">
                No notes found. Upload a note to get started.
              </div>
            )}
          </div>
        );

      case 'questions':
        if (!notes.length || !notes[0].quiz_cards.length) {
          return (
            <Card className="text-center py-12">
              <DocumentTextIcon className="h-12 w-12 mx-auto text-gray-400" />
              <h3 className="mt-2 text-sm font-medium text-gray-900 dark:text-white">No questions yet</h3>
              <p className="mt-1 text-sm text-gray-500">Upload a note to generate study questions.</p>
            </Card>
          );
        }

        {
          const currentNote = notes[0];
          const currentCard = currentNote.quiz_cards[currentQuestionIndex];

          return (
            <Card className="max-w-2xl mx-auto">
              <div className="flex items-center justify-between mb-4">
                <button
                  onClick={() => setCurrentQuestionIndex((i) => Math.max(0, i - 1))}
                  disabled={currentQuestionIndex === 0}
                  className="p-2 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-50"
                >
                  <ChevronLeftIcon className="h-6 w-6" />
                </button>
                <span className="text-sm text-gray-500">
                  Question {currentQuestionIndex + 1} of {currentNote.quiz_cards.length}
                </span>
                <button
                  onClick={() => setCurrentQuestionIndex((i) => Math.min(currentNote.quiz_cards.length - 1, i + 1))}
                  disabled={currentQuestionIndex === currentNote.quiz_cards.length - 1}
                  className="p-2 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-50"
                >
                  <ChevronRightIcon className="h-6 w-6" />
                </button>
              </div>
              <div className="space-y-4">
                <div className="p-6 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
                  <h4 className="font-medium text-gray-900 dark:text-white mb-2">Question:</h4>
                  <p className="text-gray-600 dark:text-gray-300">{currentCard.question}</p>
                </div>
                <div className="p-6 bg-primary-50 dark:bg-primary-900/20 rounded-lg">
                  <h4 className="font-medium text-gray-900 dark:text-white mb-2">Answer:</h4>
                  <p className="text-gray-600 dark:text-gray-300">{currentCard.answer}</p>
                </div>
              </div>
            </Card>
          );
        }

      case 'study':
        return (
          <Card title="Study Progress" className="max-w-2xl mx-auto">
            <div className="space-y-4">
              {Array.isArray(studyBlocks) && studyBlocks.length > 0 ? (
                studyBlocks.map((block) => (
                  <div
                    key={block.id}
                    className={cn(
                      'p-4 rounded-lg border',
                      block.status === 'completed'
                        ? 'bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800'
                        : 'bg-gray-50 dark:bg-gray-800/50 border-gray-200 dark:border-gray-700'
                    )}
                  >
                    <div className="flex justify-between items-center">
                      <div>
                        <p className="font-medium text-gray-900 dark:text-white">
                          Study Session {new Date(block.start_time).toLocaleDateString()}
                        </p>
                        <p className="text-sm text-gray-500">
                          {new Date(block.start_time).toLocaleTimeString()} -{' '}
                          {new Date(block.end_time).toLocaleTimeString()}
                        </p>
                      </div>
                      <span
                        className={cn(
                          'px-2.5 py-0.5 rounded-full text-xs font-medium',
                          block.status === 'completed'
                            ? 'bg-green-100 text-green-800 dark:bg-green-800/20 dark:text-green-300'
                            : 'bg-gray-100 text-gray-800 dark:bg-gray-800/20 dark:text-gray-300'
                        )}
                      >
                        {block.status}
                      </span>
                    </div>
                  </div>
                ))
              ) : (
                <div className="text-center text-gray-500 dark:text-gray-400 py-8">
                  No study blocks found. Complete a study session to see progress.
                </div>
              )}
            </div>
          </Card>
        );
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <Sidebar
        darkMode={darkMode}
        onDarkModeToggle={() => setDarkMode(!darkMode)}
        activeSection={activeSection}
        onSectionChange={setActiveSection}
      />
      <main className="pl-18">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <h1 className="text-2xl font-semibold text-gray-900 dark:text-white mb-8">
            {activeSection === 'upload' && 'Upload New Note'}
            {activeSection === 'notes' && 'Your Notes'}
            {activeSection === 'questions' && 'Study Questions'}
            {activeSection === 'study' && 'Study Progress'}
          </h1>
          {renderContent()}
        </div>
      </main>
    </div>
  );
} 