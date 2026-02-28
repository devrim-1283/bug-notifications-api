import { useI18n } from '../i18n';

interface Props {
  onNewReport: () => void;
}

export function SuccessScreen({ onNewReport }: Props) {
  const { t } = useI18n();

  return (
    <div className="success-overlay visible">
      <div className="success-content">
        <div className="success-check">
          <i className="fa-solid fa-check" />
        </div>
        <h2>{t.successTitle}</h2>
        <p>{t.successText}</p>
        <button type="button" className="btn-new" onClick={onNewReport}>
          <i className="fa-solid fa-plus" /> {t.newReport}
        </button>
      </div>
    </div>
  );
}
