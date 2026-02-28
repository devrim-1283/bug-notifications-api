import { useI18n } from '../i18n';
import type { ReportType } from '../types';

interface Props {
  value: ReportType;
  onChange: (type: ReportType) => void;
}

export function ReportTypeToggle({ value, onChange }: Props) {
  const { t } = useI18n();

  return (
    <div className="type-toggle">
      <button
        type="button"
        className={value === 'bug' ? 'active' : ''}
        onClick={() => onChange('bug')}
      >
        <i className="fa-solid fa-bug" /> {t.typeBug}
      </button>
      <button
        type="button"
        className={value === 'request' ? 'active' : ''}
        onClick={() => onChange('request')}
      >
        <i className="fa-solid fa-lightbulb" /> {t.typeRequest}
      </button>
    </div>
  );
}
