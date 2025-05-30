interface SidebarProps {
    darkMode: boolean;
    onDarkModeToggle: () => void;
    activeSection: 'upload' | 'notes' | 'questions' | 'study';
    onSectionChange: (section: 'upload' | 'notes' | 'questions' | 'study') => void;
}
export declare function Sidebar({ darkMode, onDarkModeToggle, activeSection, onSectionChange }: SidebarProps): import("react/jsx-runtime").JSX.Element;
export {};
