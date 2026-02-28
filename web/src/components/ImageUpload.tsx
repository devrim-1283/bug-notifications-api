import { useRef, useState, useCallback } from 'react';
import { useI18n } from '../i18n';

const MAX_IMAGES = 5;
const MAX_SIZE = 5 * 1024 * 1024;
const ALLOWED_TYPES = ['image/png', 'image/jpeg', 'image/webp', 'image/gif'];

interface Props {
  files: File[];
  onChange: (files: File[]) => void;
}

export function ImageUpload({ files, onChange }: Props) {
  const { t } = useI18n();
  const inputRef = useRef<HTMLInputElement>(null);
  const [dragOver, setDragOver] = useState(false);

  const addFiles = useCallback(
    (newFiles: FileList | File[]) => {
      const current = [...files];
      for (const f of Array.from(newFiles)) {
        if (current.length >= MAX_IMAGES) {
          alert(t.maxImages);
          break;
        }
        if (ALLOWED_TYPES.includes(f.type) && f.size <= MAX_SIZE) {
          current.push(f);
        }
      }
      onChange(current);
    },
    [files, onChange, t.maxImages]
  );

  function removeFile(index: number) {
    const next = files.filter((_, i) => i !== index);
    onChange(next);
  }

  function handleDrop(e: React.DragEvent) {
    e.preventDefault();
    setDragOver(false);
    addFiles(e.dataTransfer.files);
  }

  return (
    <div className="images-section">
      <div className="images-title">
        <i className="fa-solid fa-image" /> {t.labelImages}
      </div>
      <div
        className={`drop-zone${dragOver ? ' drag-over' : ''}`}
        onDragOver={(e) => {
          e.preventDefault();
          setDragOver(true);
        }}
        onDragLeave={() => setDragOver(false)}
        onDrop={handleDrop}
      >
        <input
          ref={inputRef}
          type="file"
          multiple
          accept="image/png,image/jpeg,image/webp,image/gif"
          onChange={(e) => {
            if (e.target.files) addFiles(e.target.files);
            e.target.value = '';
          }}
        />
        <i className="fa-solid fa-cloud-arrow-up" />
        <p>{t.dropText}</p>
      </div>
      {files.length > 0 && (
        <div className="preview-grid">
          {files.map((f, i) => (
            <PreviewItem
              key={`${f.name}-${i}`}
              file={f}
              onRemove={() => removeFile(i)}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function PreviewItem({
  file,
  onRemove,
}: {
  file: File;
  onRemove: () => void;
}) {
  const [src, setSrc] = useState('');

  if (!src) {
    const url = URL.createObjectURL(file);
    setSrc(url);
  }

  return (
    <div className="preview-item">
      <img
        src={src}
        alt=""
        onLoad={() => {
          if (src) URL.revokeObjectURL(src);
        }}
      />
      <button type="button" onClick={onRemove}>
        <i className="fa-solid fa-xmark" />
      </button>
    </div>
  );
}
