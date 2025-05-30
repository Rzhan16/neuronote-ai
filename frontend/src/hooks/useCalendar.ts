import { useState, useEffect, useCallback } from 'react';
import axios from 'axios';
import ICAL from 'ical.js';
import { get, set } from 'idb-keyval';

interface CalendarEvent {
  id: string;
  title: string;
  start: Date;
  end: Date;
  calendar: 'google' | 'outlook';
  busy: boolean;
}

interface CalendarState {
  events: CalendarEvent[];
  loading: boolean;
  error: Error | null;
}

const CACHE_KEY = 'calendar_events';
const CACHE_DURATION = 60 * 60 * 1000; // 1 hour in milliseconds

async function fetchGoogleCalendar(token: string): Promise<CalendarEvent[]> {
  const response = await axios.get('https://www.googleapis.com/calendar/v3/calendars/primary/events', {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    params: {
      timeMin: new Date().toISOString(),
      timeMax: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(), // 30 days
      singleEvents: true,
      orderBy: 'startTime',
    },
  });

  return response.data.items.map((event: any) => ({
    id: event.id,
    title: event.summary,
    start: new Date(event.start.dateTime || event.start.date),
    end: new Date(event.end.dateTime || event.end.date),
    calendar: 'google' as const,
    busy: event.transparency !== 'transparent',
  }));
}

async function fetchOutlookCalendar(token: string): Promise<CalendarEvent[]> {
  const response = await axios.get('https://outlook.office.com/api/v2.0/me/calendarview', {
    headers: {
      Authorization: `Bearer ${token}`,
      Prefer: 'outlook.timezone="UTC"',
    },
    params: {
      startDateTime: new Date().toISOString(),
      endDateTime: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
    },
  });

  return response.data.value.map((event: any) => ({
    id: event.Id,
    title: event.Subject,
    start: new Date(event.Start.DateTime),
    end: new Date(event.End.DateTime),
    calendar: 'outlook' as const,
    busy: event.ShowAs === 'Busy',
  }));
}

async function fetchCalDAVCalendar(url: string, username: string, password: string): Promise<CalendarEvent[]> {
  const response = await axios.get(url, {
    auth: {
      username,
      password,
    },
    headers: {
      'Content-Type': 'text/calendar',
    },
  });

  const jcalData = ICAL.parse(response.data);
  const comp = new ICAL.Component(jcalData);
  const vevents = comp.getAllSubcomponents('vevent');

  return vevents.map((vevent) => {
    const event = new ICAL.Event(vevent);
    return {
      id: event.uid,
      title: event.summary,
      start: event.startDate.toJSDate(),
      end: event.endDate.toJSDate(),
      calendar: 'outlook' as const, // or determine from URL
      busy: !event.component.getFirstPropertyValue('transp') || 
            event.component.getFirstPropertyValue('transp') !== 'TRANSPARENT',
    };
  });
}

export function useCalendar() {
  const [state, setState] = useState<CalendarState>({
    events: [],
    loading: true,
    error: null,
  });

  const fetchAndCacheEvents = useCallback(async () => {
    setState(prev => ({ ...prev, loading: true, error: null }));
    try {
      // Get cached data first
      const cached = await get(CACHE_KEY);
      if (cached) {
        const { events, timestamp } = cached;
        if (Date.now() - timestamp < CACHE_DURATION) {
          setState({
            events: events.map((e: any) => ({
              ...e,
              start: new Date(e.start),
              end: new Date(e.end),
            })),
            loading: false,
            error: null,
          });
          return;
        }
      }

      // Fetch new data
      const [googleEvents, outlookEvents] = await Promise.all([
        // Replace with your actual auth tokens/credentials
        fetchGoogleCalendar(import.meta.env.VITE_GOOGLE_CALENDAR_TOKEN || ''),
        fetchOutlookCalendar(import.meta.env.VITE_OUTLOOK_CALENDAR_TOKEN || ''),
      ]);

      const allEvents = [...googleEvents, ...outlookEvents].sort(
        (a, b) => a.start.getTime() - b.start.getTime()
      );

      // Cache the results
      await set(CACHE_KEY, {
        events: allEvents,
        timestamp: Date.now(),
      });

      setState({
        events: allEvents,
        loading: false,
        error: null,
      });
    } catch (error) {
      console.error('Failed to fetch calendar events:', error);
      setState(prev => ({
        ...prev,
        loading: false,
        error: error as Error,
      }));
    }
  }, []);

  useEffect(() => {
    fetchAndCacheEvents();
  }, [fetchAndCacheEvents]);

  const refresh = useCallback(() => {
    fetchAndCacheEvents();
  }, [fetchAndCacheEvents]);

  return {
    ...state,
    refresh,
  };
} 