import { useState, useMemo } from 'react';
import { I18nContext, detectLanguage, t } from './i18n';
import { useTheme } from './hooks/useTheme';
import type { Language, AppView } from './types';
import { Header } from './components/Header';
import { FeedbackForm } from './components/FeedbackForm';
import { SuccessScreen } from './components/SuccessScreen';

export default function App() {
  const [lang, setLang] = useState<Language>(detectLanguage);
  const [view, setView] = useState<AppView>('form');
  const theme = useTheme();

  const i18nValue = useMemo(
    () => ({ lang, setLang, t: t(lang) }),
    [lang]
  );

  return (
    <I18nContext.Provider value={i18nValue}>
      {view === 'success' && (
        <SuccessScreen onNewReport={() => setView('form')} />
      )}
      <div className="app-shell">
        <div className="main-card">
          <Header theme={theme} />
          <div className="form-body">
            {view === 'form' && (
              <FeedbackForm key={view} onSuccess={() => setView('success')} resolvedTheme={theme.resolved} />
            )}
          </div>
          <div className="footer">
            <span>
              {i18nValue.t.footerText}{' '}
              <a href="https://devrimsoft.com" target="_blank" rel="noopener noreferrer">
                DevrimSoft
              </a>.
            </span>
          </div>
        </div>
      </div>
    </I18nContext.Provider>
  );
}
