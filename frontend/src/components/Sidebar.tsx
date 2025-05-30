import React from 'react';
import { cn } from '@/lib/utils';
import {
  DocumentPlusIcon,
  DocumentTextIcon,
  QuestionMarkCircleIcon,
  ClockIcon,
  SunIcon,
  MoonIcon,
} from '@heroicons/react/24/outline';

interface SidebarProps {
  darkMode: boolean;
  onDarkModeToggle: () => void;
  activeSection: 'upload' | 'notes' | 'questions' | 'study';
  onSectionChange: (section: 'upload' | 'notes' | 'questions' | 'study') => void;
}

export function Sidebar({ darkMode, onDarkModeToggle, activeSection, onSectionChange }: SidebarProps) {
  const navItems = [
    { id: 'upload', icon: DocumentPlusIcon, label: 'Upload' },
    { id: 'notes', icon: DocumentTextIcon, label: 'Notes' },
    { id: 'questions', icon: QuestionMarkCircleIcon, label: 'Questions' },
    { id: 'study', icon: ClockIcon, label: 'Study Plan' },
  ] as const;

  return (
    <div className="fixed inset-y-0 left-0 w-18 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 z-30">
      <div className="h-full flex flex-col items-center py-6">
        <div className="flex-1 flex flex-col items-center space-y-6">
          {navItems.map((item) => {
            const Icon = item.icon;
            return (
              <button
                key={item.id}
                onClick={() => onSectionChange(item.id)}
                className={cn(
                  'p-3 rounded-xl transition-all duration-200',
                  'hover:bg-gray-100 dark:hover:bg-gray-700',
                  'group flex flex-col items-center',
                  activeSection === item.id
                    ? 'text-primary-600 dark:text-primary-400 bg-primary-50 dark:bg-primary-900/20'
                    : 'text-gray-500 dark:text-gray-400'
                )}
              >
                <Icon className="h-6 w-6" />
                <span className="mt-1 text-xs font-medium">{item.label}</span>
              </button>
            );
          })}
        </div>
        <button
          onClick={onDarkModeToggle}
          className="p-3 text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors"
          aria-label="Toggle dark mode"
        >
          {darkMode ? <SunIcon className="h-6 w-6" /> : <MoonIcon className="h-6 w-6" />}
        </button>
      </div>
    </div>
  );
} 