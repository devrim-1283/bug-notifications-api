import { useI18n } from '../i18n';
import { LanguageSelector } from './LanguageSelector';
import { ThemeToggle } from './ThemeToggle';
import type { ThemeChoice } from '../hooks/useTheme';

interface Props {
  theme: {
    choice: ThemeChoice;
    resolved: 'light' | 'dark';
    setChoice: (c: ThemeChoice) => void;
  };
}

export function Header({ theme }: Props) {
  const { t } = useI18n();

  return (
    <div className="header">
      <div className="brand">
        <div className="brand-icon">
          <i className="fa-solid fa-comment-dots" />
        </div>
        <div className="brand-text">
          <h1>{t.pageTitle}</h1>
          <small>{t.pageSubtitle}</small>
        </div>
      </div>
      <div className="header-actions">
        <ThemeToggle
          choice={theme.choice}
          resolved={theme.resolved}
          setChoice={theme.setChoice}
        />
        <LanguageSelector />
      </div>
    </div>
  );
}
