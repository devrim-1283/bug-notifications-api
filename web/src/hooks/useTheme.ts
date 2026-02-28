import { useState, useEffect, useCallback } from 'react';

export type ThemeChoice = 'light' | 'dark' | 'system';

const STORAGE_KEY = 'theme';

function getSystemTheme(): 'light' | 'dark' {
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function loadChoice(): ThemeChoice {
  const stored = localStorage.getItem(STORAGE_KEY);
  if (stored === 'light' || stored === 'dark' || stored === 'system') return stored;
  return 'system';
}

function applyTheme(resolved: 'light' | 'dark') {
  document.documentElement.setAttribute('data-theme', resolved);
}

export function useTheme() {
  const [choice, setChoiceState] = useState<ThemeChoice>(loadChoice);
  const [resolved, setResolved] = useState<'light' | 'dark'>(
    () => (loadChoice() === 'system' ? getSystemTheme() : loadChoice()) as 'light' | 'dark'
  );

  const resolve = useCallback((c: ThemeChoice) => {
    return c === 'system' ? getSystemTheme() : c;
  }, []);

  const setChoice = useCallback((c: ThemeChoice) => {
    localStorage.setItem(STORAGE_KEY, c);
    setChoiceState(c);
    const r = c === 'system' ? getSystemTheme() : c;
    setResolved(r);
    applyTheme(r);
  }, []);

  // Apply on mount
  useEffect(() => {
    applyTheme(resolve(choice));
  }, []);

  // Listen for system theme changes when choice is 'system'
  useEffect(() => {
    if (choice !== 'system') return;

    const mq = window.matchMedia('(prefers-color-scheme: dark)');
    function onChange() {
      const r = getSystemTheme();
      setResolved(r);
      applyTheme(r);
    }
    mq.addEventListener('change', onChange);
    return () => mq.removeEventListener('change', onChange);
  }, [choice]);

  return { choice, resolved, setChoice };
}
