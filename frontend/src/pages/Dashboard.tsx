import React, { useState, useEffect } from 'react';
import { PieChart, Pie, Cell, ResponsiveContainer, Legend } from 'recharts';
import * as Dialog from '@radix-ui/react-dialog';
import { UploadDropzone } from '@/components/UploadDropzone';
import { cn } from '@/lib/utils';

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
  note_id: string;
  start_time: string;
  end_time: string;
  status: 'scheduled' | 'completed';
}

export function Dashboard() {
  const [notes, setNotes] = useState<Note[]>([]);
  const [studyBlocks, setStudyBlocks] = useState<StudyBlock[]>([]);
  const [isUploading, setIsUploading] = useState(false);

  useEffect(() => {
    // Fetch notes
    fetch('/api/notes')
      .then(res => res.json())
      .then(data => setNotes(data))
      .catch(err => console.error('Failed to fetch notes:', err));

    // Fetch study blocks
    fetch('/api/schedule')
      .then(res => res.json())
      .then(data => setStudyBlocks(data))
      .catch(err => console.error('Failed to fetch study blocks:', err));
  }, []);

  // Calculate stats for pie chart
  const studyStats = studyBlocks.reduce(
    (acc, block) => {
      acc[block.status]++;
      return acc;
    },
    { scheduled: 0, completed: 0 }
  );

  const pieData = [
    { name: 'Scheduled', value: studyStats.scheduled },
    { name: 'Completed', value: studyStats.completed },
  ];

  const COLORS = ['#0088FE', '#00C49F'];

  const handleUploadSuccess = (response: any) => {
    setIsUploading(false);
    // Refresh notes list
    fetch('/api/notes')
      .then(res => res.json())
      .then(data => setNotes(data))
      .catch(err => console.error('Failed to fetch notes:', err));
  };

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold">Your Notes</h1>
        <Dialog.Root>
          <Dialog.Trigger asChild>
            <button className="fixed bottom-8 right-8 bg-blue-600 text-white rounded-full p-4 shadow-lg hover:bg-blue-700 transition-colors">
              <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
            </button>
          </Dialog.Trigger>
          <Dialog.Portal>
            <Dialog.Overlay className="fixed inset-0 bg-black/50" />
            <Dialog.Content className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-white rounded-lg p-6 w-full max-w-lg">
              <Dialog.Title className="text-xl font-semibold mb-4">Upload Note</Dialog.Title>
              <UploadDropzone
                onUploadSuccess={handleUploadSuccess}
                onUploadError={(error) => {
                  console.error('Upload failed:', error);
                  setIsUploading(false);
                }}
              />
              <Dialog.Close asChild>
                <button className="absolute top-4 right-4 text-gray-400 hover:text-gray-600">
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </Dialog.Close>
            </Dialog.Content>
          </Dialog.Portal>
        </Dialog.Root>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
        <div className="md:col-span-2">
          <div className="bg-white rounded-lg shadow-md">
            {notes.length === 0 ? (
              <div className="p-8 text-center text-gray-500">
                No notes yet. Click the + button to add your first note!
              </div>
            ) : (
              <ul className="divide-y divide-gray-200">
                {notes.map(note => (
                  <li key={note.id} className="p-6 hover:bg-gray-50 transition-colors">
                    <div className="flex justify-between items-start">
                      <div>
                        <h3 className="text-lg font-semibold text-gray-900">{note.title}</h3>
                        <p className="mt-1 text-sm text-gray-600 line-clamp-2">{note.summary}</p>
                      </div>
                      <div className="flex items-center space-x-2">
                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                          {note.quiz_cards.length} cards
                        </span>
                        <span className="text-xs text-gray-500">
                          {new Date(note.created_at).toLocaleDateString()}
                        </span>
                      </div>
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-md p-6">
          <h2 className="text-xl font-semibold mb-4">Study Progress</h2>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={pieData}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  outerRadius={80}
                  fill="#8884d8"
                  dataKey="value"
                >
                  {pieData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                  ))}
                </Pie>
                <Legend verticalAlign="bottom" height={36} />
              </PieChart>
            </ResponsiveContainer>
          </div>
          <div className="mt-4 grid grid-cols-2 gap-4 text-center">
            <div className="bg-blue-50 rounded-lg p-3">
              <div className="text-2xl font-bold text-blue-600">{studyStats.scheduled}</div>
              <div className="text-sm text-blue-800">Scheduled</div>
            </div>
            <div className="bg-green-50 rounded-lg p-3">
              <div className="text-2xl font-bold text-green-600">{studyStats.completed}</div>
              <div className="text-sm text-green-800">Completed</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
} 