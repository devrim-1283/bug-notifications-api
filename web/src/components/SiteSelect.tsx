import { useEffect, useRef } from 'react';
import { useI18n } from '../i18n';
import { getConfig } from '../config';

interface Props {
  value: string;
  onChange: (siteId: string) => void;
  onAutoFillUrl: (url: string) => void;
  autoDetected: boolean;
}

export function SiteSelect({
  value,
  onChange,
  onAutoFillUrl,
  autoDetected,
}: Props) {
  const { t } = useI18n();
  const config = getConfig();
  const hasAutoFilled = useRef(false);

  useEffect(() => {
    if (value && !hasAutoFilled.current && !autoDetected) {
      onAutoFillUrl(`https://${value}`);
    }
    if (value) {
      hasAutoFilled.current = true;
    }
  }, [value, onAutoFillUrl, autoDetected]);

  function handleChange(siteId: string) {
    onChange(siteId);
    if (siteId) {
      onAutoFillUrl(`https://${siteId}`);
    }
    hasAutoFilled.current = true;
  }

  return (
    <>
      <label>
        {t.labelSite} <span className="req">*</span>
      </label>
      <select
        value={value}
        onChange={(e) => handleChange(e.target.value)}
        required
      >
        <option value="">{t.siteSelectPlaceholder}</option>
        {config.sites.map((s) => (
          <option key={s} value={s}>
            {s}
          </option>
        ))}
      </select>
      {autoDetected && (
        <div className="field-note auto">{t.autoDetected}</div>
      )}
    </>
  );
}
