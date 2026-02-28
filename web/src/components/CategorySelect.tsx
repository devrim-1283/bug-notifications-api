import { useI18n } from '../i18n';
import type { Category } from '../types';

interface Props {
  value: Category | '';
  onChange: (cat: Category | '') => void;
}

export function CategorySelect({ value, onChange }: Props) {
  const { t } = useI18n();

  const categories: { value: Category; label: string }[] = [
    { value: 'design', label: t.catDesign },
    { value: 'functionality', label: t.catFunctionality },
    { value: 'performance', label: t.catPerformance },
    { value: 'content', label: t.catContent },
    { value: 'mobile', label: t.catMobile },
    { value: 'security', label: t.catSecurity },
    { value: 'other', label: t.catOther },
  ];

  return (
    <>
      <label>
        {t.labelCategory} <span className="req">*</span>
      </label>
      <select
        value={value}
        onChange={(e) => onChange(e.target.value as Category | '')}
        required
      >
        <option value="">{t.selectPlaceholder}</option>
        {categories.map((c) => (
          <option key={c.value} value={c.value}>
            {c.label}
          </option>
        ))}
      </select>
    </>
  );
}
