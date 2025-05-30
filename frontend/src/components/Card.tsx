import React from 'react';
import { cn } from '@/lib/utils';

interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  title?: string;
  description?: string;
  footer?: React.ReactNode;
  children?: React.ReactNode;
}

export function Card({ title, description, footer, className, children, ...props }: CardProps) {
  return (
    <div
      className={cn(
        'bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700',
        'transition-all duration-200 hover:shadow-md',
        className
      )}
      {...props}
    >
      {(title || description) && (
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          {title && (
            <h3 className="text-lg font-medium leading-6 text-gray-900 dark:text-white">
              {title}
            </h3>
          )}
          {description && (
            <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
              {description}
            </p>
          )}
        </div>
      )}
      <div className="px-6 py-4">{children}</div>
      {footer && (
        <div className="px-6 py-4 bg-gray-50 dark:bg-gray-800/50 rounded-b-lg border-t border-gray-200 dark:border-gray-700">
          {footer}
        </div>
      )}
    </div>
  );
} 