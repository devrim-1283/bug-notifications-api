import Turnstile from 'react-turnstile';
import { getConfig } from '../config';

interface Props {
  onVerify: (token: string) => void;
  onExpire: () => void;
  onError: () => void;
}

export function TurnstileWidget({ onVerify, onExpire, onError }: Props) {
  const config = getConfig();

  if (!config.turnstileSiteKey) {
    return (
      <div className="turnstile-wrap">
        <p style={{ color: 'var(--error)', fontSize: '0.75rem' }}>
          Turnstile not configured
        </p>
      </div>
    );
  }

  return (
    <div className="turnstile-wrap">
      <Turnstile
        sitekey={config.turnstileSiteKey}
        theme="light"
        onVerify={onVerify}
        onExpire={onExpire}
        onError={onError}
      />
    </div>
  );
}
