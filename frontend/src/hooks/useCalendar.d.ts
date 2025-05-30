interface CalendarEvent {
    id: string;
    title: string;
    start: Date;
    end: Date;
    calendar: 'google' | 'outlook';
    busy: boolean;
}
export declare function useCalendar(): {
    refresh: () => void;
    events: CalendarEvent[];
    loading: boolean;
    error: Error | null;
};
export {};
