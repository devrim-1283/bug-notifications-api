import { useState, useEffect, useRef } from 'react';
import type { ThemeChoice } from '../hooks/useTheme';

interface Props {
  choice: ThemeChoice;
  resolved: 'light' | 'dark';
  setChoice: (c: ThemeChoice) => void;
}

const OPTIONS: { value: ThemeChoice; icon: string; label: string }[] = [
  { value: 'light', icon: 'fa-sun', label: 'Light' },
  { value: 'dark', icon: 'fa-moon', label: 'Dark' },
  { value: 'system', icon: 'fa-desktop', label: 'System' },
];

export function ThemeToggle({ choice, setChoice }: Props) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener('click', handleClick);
    return () => document.removeEventListener('click', handleClick);
  }, []);

  const current = OPTIONS.find((o) => o.value === choice)!;

  return (
    <div className="theme-select" ref={ref}>
      <button className="theme-btn" type="button" onClick={() => setOpen(!open)}>
        <i className={`fa-solid ${current.icon}`} />
      </button>
      <div className={`theme-menu${open ? ' open' : ''}`}>
        {OPTIONS.map((o) => (
          <button
            key={o.value}
            type="button"
            className={`theme-item${o.value === choice ? ' active' : ''}`}
            onClick={() => {
              setChoice(o.value);
              setOpen(false);
            }}
          >
            <i className={`fa-solid ${o.icon}`} />
            {o.label}
          </button>
        ))}
      </div>
    </div>
  );
}
