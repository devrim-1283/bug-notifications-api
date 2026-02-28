import { useState } from 'react';
import { useI18n } from '../i18n';

interface Props {
  fullName: string;
  phone: string;
  email: string;
  onFullNameChange: (v: string) => void;
  onPhoneChange: (v: string) => void;
  onEmailChange: (v: string) => void;
}

export function ContactSection({
  fullName,
  phone,
  email,
  onFullNameChange,
  onPhoneChange,
  onEmailChange,
}: Props) {
  const { t } = useI18n();
  const [open, setOpen] = useState(false);

  return (
    <div className="contact-section">
      <button
        type="button"
        className={`contact-toggle${open ? ' active' : ''}`}
        onClick={() => setOpen(!open)}
      >
        <i className={`fa-solid ${open ? 'fa-chevron-up' : 'fa-envelope'}`} />
        {t.contactToggle}
      </button>
      {open && (
        <div className="contact-fields">
          <div className="field">
            <label>{t.labelFullName}</label>
            <input
              type="text"
              maxLength={200}
              value={fullName}
              onChange={(e) => onFullNameChange(e.target.value)}
            />
          </div>
          <div className="field">
            <label>{t.labelPhone}</label>
            <input
              type="tel"
              maxLength={30}
              value={phone}
              onChange={(e) => onPhoneChange(e.target.value)}
            />
          </div>
          <div className="field">
            <label>{t.labelEmail}</label>
            <input
              type="email"
              maxLength={200}
              value={email}
              onChange={(e) => onEmailChange(e.target.value)}
            />
          </div>
        </div>
      )}
    </div>
  );
}
