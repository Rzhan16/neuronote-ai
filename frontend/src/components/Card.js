import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { cn } from '@/lib/utils';
export function Card({ title, description, footer, className, children, ...props }) {
    return (_jsxs("div", { className: cn('bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700', 'transition-all duration-200 hover:shadow-md', className), ...props, children: [(title || description) && (_jsxs("div", { className: "px-6 py-4 border-b border-gray-200 dark:border-gray-700", children: [title && (_jsx("h3", { className: "text-lg font-medium leading-6 text-gray-900 dark:text-white", children: title })), description && (_jsx("p", { className: "mt-1 text-sm text-gray-500 dark:text-gray-400", children: description }))] })), _jsx("div", { className: "px-6 py-4", children: children }), footer && (_jsx("div", { className: "px-6 py-4 bg-gray-50 dark:bg-gray-800/50 rounded-b-lg border-t border-gray-200 dark:border-gray-700", children: footer }))] }));
}
