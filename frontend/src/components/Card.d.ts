import React from 'react';
interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
    title?: string;
    description?: string;
    footer?: React.ReactNode;
    children?: React.ReactNode;
}
export declare function Card({ title, description, footer, className, children, ...props }: CardProps): import("react/jsx-runtime").JSX.Element;
export {};
