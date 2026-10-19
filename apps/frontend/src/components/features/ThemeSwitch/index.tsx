import Segmented from '../../ui/Segmented';
import { useThemeStore, type Theme } from '../../../store/theme';

const options: { value: Theme; label: string }[] = [
  { value: 'info', label: 'Info' },
  { value: 'neon', label: 'Neon' },
];

export default function ThemeSwitch() {
  const { theme, setTheme } = useThemeStore();

  return (
    <div className="flex items-center justify-between gap-4 border-b border-border py-3.5">
      <div>
        <div className="font-medium text-text">Theme</div>
        <div className="text-sm text-text-faint">Saved on this device</div>
      </div>
      <Segmented label="Theme" value={theme} options={options} onChange={setTheme} />
    </div>
  );
}
