import { useEffect } from 'react';
import Turnstile from 'react-turnstile';
import { getConfig } from '../config';

interface Props {
  onVerify: (token: string) => void;
  onExpire: () => void;
  onError: () => void;
  theme?: 'light' | 'dark' | 'auto';
}

export function TurnstileWidget({ onVerify, onExpire, onError, theme = 'auto' }: Props) {
  const config = getConfig();

  // When Turnstile is not configured, skip verification
  useEffect(() => {
    if (!config.turnstileSiteKey) {
      onVerify('__skip__');
    }
  }, [config.turnstileSiteKey, onVerify]);

  if (!config.turnstileSiteKey) {
    return null;
  }

  return (
    <div className="turnstile-wrap">
      <Turnstile
        sitekey={config.turnstileSiteKey}
        theme={theme}
        onVerify={onVerify}
        onExpire={onExpire}
        onError={onError}
      />
    </div>
  );
}
