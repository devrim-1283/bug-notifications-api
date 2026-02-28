import { useState, useEffect, useRef } from 'react';
import { useI18n, LANGUAGE_NAMES, LANGUAGES } from '../i18n';
import type { Language } from '../types';

import trFlag from 'flag-icons/flags/4x3/tr.svg';
import gbFlag from 'flag-icons/flags/4x3/gb.svg';
import deFlag from 'flag-icons/flags/4x3/de.svg';
import ruFlag from 'flag-icons/flags/4x3/ru.svg';
import uaFlag from 'flag-icons/flags/4x3/ua.svg';
import esFlag from 'flag-icons/flags/4x3/es.svg';

const FLAG_SVGS: Record<Language, string> = {
  tr: trFlag,
  en: gbFlag,
  de: deFlag,
  ru: ruFlag,
  uk: uaFlag,
  es: esFlag,
};

export function LanguageSelector() {
  const { lang, setLang } = useI18n();
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

  function select(l: Language) {
    setLang(l);
    setOpen(false);
  }

  return (
    <div className="lang-select" ref={ref}>
      <button className="lang-btn" type="button" onClick={() => setOpen(!open)}>
        <img className="flag" src={FLAG_SVGS[lang]} alt="" />
        <span>{lang.toUpperCase()}</span>
        <i className="fa-solid fa-chevron-down" />
      </button>
      <div className={`lang-menu${open ? ' open' : ''}`}>
        {LANGUAGES.map((l) => (
          <button
            key={l}
            type="button"
            className={`lang-item${l === lang ? ' active' : ''}`}
            onClick={() => select(l)}
          >
            <img className="flag" src={FLAG_SVGS[l]} alt="" /> {LANGUAGE_NAMES[l]}
          </button>
        ))}
      </div>
    </div>
  );
}
